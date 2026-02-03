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

	"github.com/castrojo/bluefin-releases/internal/models"
)

const (
	FlathubAPIBase = "https://flathub.org/api/v2"
)

// FetchAllApps fetches apps and enriches with details.
// If appIDs is provided, fetches only those specific apps.
// Otherwise, fetches recently updated apps.
// Follows the pattern of feeds.FetchAllFeeds from firehose
func FetchAllApps(appIDs ...string) *models.FetchResults {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allApps []models.App
	)

	var flathubApps []models.FlathubApp

	// Step 1: Fetch list of apps (either specific IDs or recently updated)
	if len(appIDs) > 0 {
		// Fetch specific app IDs
		log.Printf("Fetching %d specific apps from Flathub...", len(appIDs))
		for _, appID := range appIDs {
			// Create a FlathubApp stub with just the ID
			// The enrichApp function will fetch full details
			flathubApps = append(flathubApps, models.FlathubApp{
				AppID: appID,
			})
		}
	} else {
		// Fetch recently updated apps (original behavior)
		log.Println("Fetching recently updated apps from Flathub...")
		var err error
		flathubApps, err = FetchRecentlyUpdated()
		if err != nil {
			log.Fatalf("Failed to fetch apps: %v", err)
		}
		log.Printf("Fetched %d recently updated apps", len(flathubApps))
	}

	// Step 2: Fetch details for each app in parallel
	// Limit to 50 apps only for recently-updated to avoid timeouts
	// For specific app IDs, fetch all of them
	appsToFetch := flathubApps
	if len(appIDs) == 0 && len(appsToFetch) > 50 {
		// Only limit when fetching recently updated apps
		appsToFetch = appsToFetch[:50]
		log.Printf("Limited to first 50 apps to avoid timeouts")
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

	// Fetch detailed information first (needed for apps with only ID)
	details, err := FetchAppDetails(flathubApp.AppID)
	if err != nil {
		log.Printf("⚠️  Failed to fetch details for %s: %v", flathubApp.AppID, err)
		// Return minimal app with just ID and URL
		return models.App{
			ID:         flathubApp.AppID,
			FlathubURL: fmt.Sprintf("https://flathub.org/apps/%s", flathubApp.AppID),
			FetchedAt:  fetchedAt,
		}
	}

	// Use details to fill in missing data from collection API
	name := flathubApp.Name
	if name == "" && details != nil {
		name = details.Name
	}

	summary := flathubApp.Summary
	if summary == "" && details != nil {
		summary = details.Summary
	}

	description := flathubApp.Description
	if description == "" && details != nil {
		description = details.Description
	}

	icon := flathubApp.Icon
	if icon == "" && details != nil {
		icon = details.Icon
	}

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

	// Create base app from collection data (with fallbacks from details)
	app := models.App{
		ID:                flathubApp.AppID,
		Name:              name,
		Summary:           summary,
		Description:       description,
		DeveloperName:     flathubApp.DeveloperName,
		Icon:              icon,
		ProjectLicense:    flathubApp.ProjectLicense,
		Categories:        categories,
		UpdatedAt:         updatedAt,
		FlathubURL:        fmt.Sprintf("https://flathub.org/apps/%s", flathubApp.AppID),
		FetchedAt:         fetchedAt,
		InstallsLastMonth: flathubApp.InstallsLastMonth,
		FavoritesCount:    flathubApp.FavoritesCount,
		IsVerified:        flathubApp.VerificationVerified,
		VerificationInfo:  verificationInfo,
		PackageType:       "flatpak", // All apps from Flathub are Flatpaks
	}

	if details != nil {
		// Extract source repository
		sourceRepo := ExtractSourceRepo(details)
		if sourceRepo != nil {
			app.SourceRepo = sourceRepo
		}

		// Convert Flathub releases to our format - only keep the latest one
		if len(details.Releases) > 0 {
			app.Releases = ConvertFlathubReleases(details.Releases[:1]) // Only take the first (latest) release
			// Set current version and release date from first release
			app.Version = details.Releases[0].Version
			app.ReleaseDate = details.Releases[0].Date
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
