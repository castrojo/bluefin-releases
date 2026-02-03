package flathub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/castrojo/flatpak-firehose/internal/models"
)

const (
	FlathubAPIBase = "https://flathub.org/api/v2"
)

// FetchAllApps fetches recently updated apps and enriches with details
// Follows the pattern of feeds.FetchAllFeeds from firehose
func FetchAllApps() *models.FetchResults {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allApps []models.App
	)

	// Step 1: Fetch list of recently updated apps
	log.Println("Fetching recently updated apps from Flathub...")
	flathubApps, err := FetchRecentlyUpdated()
	if err != nil {
		log.Fatalf("Failed to fetch apps: %v", err)
	}
	log.Printf("Fetched %d recently updated apps", len(flathubApps))

	// Step 2: Fetch details for each app in parallel (limit to first 50 to avoid timeouts)
	appsToFetch := flathubApps
	if len(appsToFetch) > 50 {
		appsToFetch = appsToFetch[:50]
	}

	for _, flathubApp := range appsToFetch {
		wg.Add(1)
		go func(fa models.FlathubApp) {
			defer wg.Done()

			appStart := time.Now()
			app := enrichApp(fa)

			log.Printf("✅ Processed %s in %s", app.ID, time.Since(appStart))

			mu.Lock()
			allApps = append(allApps, app)
			mu.Unlock()
		}(flathubApp)
	}

	wg.Wait()

	return &models.FetchResults{
		Apps: allApps,
	}
}

// enrichApp fetches details and enriches a single app
func enrichApp(flathubApp models.FlathubApp) models.App {
	fetchedAt := time.Now().UTC()

	// Build categories array from main and sub categories
	categories := []string{}
	// MainCategories is now a StringOrArray, append all elements
	categories = append(categories, flathubApp.MainCategories...)
	categories = append(categories, flathubApp.SubCategories...)

	// Convert Unix timestamp to string
	updatedAt := ""
	if flathubApp.UpdatedAt > 0 {
		updatedAt = time.Unix(flathubApp.UpdatedAt, 0).UTC().Format(time.RFC3339)
	}

	// Build verification info
	var verificationInfo *models.Verification
	if flathubApp.VerificationVerified {
		verificationInfo = &models.Verification{
			Method: flathubApp.VerificationMethod,
		}
		if flathubApp.VerificationLoginName != nil {
			verificationInfo.LoginName = flathubApp.VerificationLoginName
		}
		if flathubApp.VerificationWebsite != nil {
			verificationInfo.Website = flathubApp.VerificationWebsite
		}
	}

	// Create base app from collection data
	app := models.App{
		ID:                flathubApp.AppID,
		Name:              flathubApp.Name,
		Summary:           flathubApp.Summary,
		Description:       flathubApp.Description,
		DeveloperName:     flathubApp.DeveloperName,
		Icon:              flathubApp.Icon,
		ProjectLicense:    flathubApp.ProjectLicense,
		Categories:        categories,
		UpdatedAt:         updatedAt,
		FlathubURL:        fmt.Sprintf("https://flathub.org/apps/%s", flathubApp.AppID),
		FetchedAt:         fetchedAt,
		InstallsLastMonth: flathubApp.InstallsLastMonth,
		FavoritesCount:    flathubApp.FavoritesCount,
		IsVerified:        flathubApp.VerificationVerified,
		VerificationInfo:  verificationInfo,
	}

	// Fetch detailed information for releases and source repo
	details, err := FetchAppDetails(flathubApp.AppID)
	if err != nil {
		log.Printf("⚠️  Failed to fetch details for %s: %v", flathubApp.AppID, err)
		return app
	}

	if details != nil {
		// Extract source repository
		sourceRepo := ExtractSourceRepo(details)
		if sourceRepo != nil {
			app.SourceRepo = sourceRepo
		}

		// Convert Flathub releases to our format
		if len(details.Releases) > 0 {
			app.Releases = ConvertFlathubReleases(details.Releases)
			// Set current version and release date from first release
			if len(details.Releases) > 0 {
				app.Version = details.Releases[0].Version
				app.ReleaseDate = details.Releases[0].Date
			}
		}
	}

	// Add small delay to avoid rate limiting
	time.Sleep(100 * time.Millisecond)

	return app
}

// FetchRecentlyUpdated fetches the list of recently updated apps from Flathub (using JSON collection API)
func FetchRecentlyUpdated() ([]models.FlathubApp, error) {
	url := fmt.Sprintf("%s/collection/recently-updated", FlathubAPIBase)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch recently updated: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var collectionResp models.FlathubCollectionResponse
	if err := json.Unmarshal(body, &collectionResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return collectionResp.Hits, nil
}

// FetchAppDetails fetches detailed information for a specific app
func FetchAppDetails(appID string) (*models.FlathubAppDetails, error) {
	url := fmt.Sprintf("%s/appstream/%s", FlathubAPIBase, appID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch app details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // App not found, not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var details models.FlathubAppDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &details, nil
}

// ExtractSourceRepo extracts source repository information from app details
func ExtractSourceRepo(details *models.FlathubAppDetails) *models.SourceRepo {
	if details == nil || details.URLs == nil {
		return nil
	}

	// Priority: homepage, bugtracker, then any other URL
	var repoURL string
	if homepage, ok := details.URLs["homepage"]; ok {
		repoURL = homepage
	} else if bugtracker, ok := details.URLs["bugtracker"]; ok {
		repoURL = bugtracker
	} else {
		// Take first available URL
		for _, url := range details.URLs {
			repoURL = url
			break
		}
	}

	if repoURL == "" {
		return nil
	}

	// Check if it's a GitHub URL
	if strings.Contains(repoURL, "github.com") {
		return extractGitHubRepo(repoURL)
	}

	// Check if it's a GitLab URL
	if strings.Contains(repoURL, "gitlab.com") {
		return &models.SourceRepo{
			Type: "gitlab",
			URL:  repoURL,
		}
	}

	// Other repository
	return &models.SourceRepo{
		Type: "other",
		URL:  repoURL,
	}
}

// extractGitHubRepo extracts owner/repo from a GitHub URL
func extractGitHubRepo(url string) *models.SourceRepo {
	// Match github.com/owner/repo patterns
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/\s?#]+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) < 3 {
		return &models.SourceRepo{
			Type: "github",
			URL:  url,
		}
	}

	owner := matches[1]
	repo := strings.TrimSuffix(matches[2], ".git")

	return &models.SourceRepo{
		Type:  "github",
		URL:   url,
		Owner: owner,
		Repo:  repo,
	}
}

// ConvertFlathubReleases converts Flathub releases to our Release format
func ConvertFlathubReleases(releases []models.FlathubReleaseEntry) []models.Release {
	var result []models.Release

	for _, release := range releases {
		// Parse date
		date, err := time.Parse("2006-01-02", release.Date)
		if err != nil {
			// Try timestamp format
			date, err = time.Parse(time.RFC3339, release.Date)
			if err != nil {
				// Default to now if parsing fails
				date = time.Now()
			}
		}

		result = append(result, models.Release{
			Version:     release.Version,
			Date:        date,
			Title:       fmt.Sprintf("Version %s", release.Version),
			Description: release.Description,
			Type:        "appstream",
		})
	}

	return result
}
