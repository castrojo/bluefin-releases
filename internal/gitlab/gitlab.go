package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/castrojo/bluefin-releases/internal/models"
)

// GitLabRelease represents a release from GitLab API v4
type GitLabRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleasedAt  time.Time `json:"released_at"`
	CreatedAt   time.Time `json:"created_at"`
	Links       struct {
		Self string `json:"self"`
	} `json:"_links"`
}

// EnrichWithGitLabReleases fetches GitLab releases for apps with GitLab repos
// and adds them to the app's release list (prioritizing actual source changelogs)
func EnrichWithGitLabReleases(apps []models.App) []models.App {
	// Check if GitLab token is available (optional)
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		log.Println("⚠️  No GITLAB_TOKEN found, using public API (lower rate limits)")
	}

	ctx := context.Background()

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	enrichedApps := make([]models.App, len(apps))
	copy(enrichedApps, apps)

	// Process apps with GitLab repos in parallel
	for i := range enrichedApps {
		app := &enrichedApps[i]
		if app.SourceRepo == nil || app.SourceRepo.Type != "gitlab" || app.SourceRepo.URL == "" {
			continue
		}

		wg.Add(1)
		go func(app *models.App) {
			defer wg.Done()

			releases, err := fetchGitLabReleases(ctx, token, app.SourceRepo.URL, app.SourceRepo.Owner, app.SourceRepo.Repo)
			if err != nil {
				log.Printf("⚠️  Failed to fetch GitLab releases for %s: %v",
					app.SourceRepo.URL, err)
				return
			}

			mu.Lock()
			// Prepend GitLab releases (they are from actual source, so prioritize them)
			app.Releases = append(releases, app.Releases...)
			log.Printf("✅ Added %d GitLab releases for %s", len(releases), app.ID)
			mu.Unlock()

			// Rate limiting: GitLab has a rate limit of 600 requests/15 minutes for unauthenticated
			// and higher limits for authenticated. This conservative sleep helps avoid hitting limits.
			time.Sleep(500 * time.Millisecond)
		}(app)
	}

	wg.Wait()
	return enrichedApps
}

// fetchGitLabReleases fetches the latest releases from a GitLab repository
// Supports both gitlab.com and self-hosted GitLab instances (like gitlab.gnome.org)
func fetchGitLabReleases(ctx context.Context, token, repoURL, owner, repo string) ([]models.Release, error) {
	// Parse the repository URL to extract the GitLab host and project path
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("parse repo URL: %w", err)
	}

	gitlabHost := parsedURL.Host
	if gitlabHost == "" {
		gitlabHost = "gitlab.com"
	}

	// Build the project path (owner/repo)
	projectPath := owner + "/" + repo
	if owner == "" || repo == "" {
		// Try to extract from URL path
		pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathParts) >= 2 {
			projectPath = strings.Join(pathParts, "/")
		} else {
			return nil, fmt.Errorf("invalid project path in URL: %s", repoURL)
		}
	}

	// URL-encode the project path (GitLab requires this)
	encodedPath := url.PathEscape(projectPath)

	// Build the API URL
	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/releases?per_page=5", gitlabHost, encodedPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add authorization header if token is provided
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 404 {
		// No releases found, not an error
		return []models.Release{}, nil
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var gitlabReleases []GitLabRelease
	if err := json.NewDecoder(resp.Body).Decode(&gitlabReleases); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Convert to models.Release
	var releases []models.Release
	for _, gr := range gitlabReleases {
		if gr.TagName == "" {
			continue
		}

		date := gr.ReleasedAt
		if date.IsZero() {
			date = gr.CreatedAt
		}
		if date.IsZero() {
			date = time.Now()
		}

		title := gr.TagName
		if gr.Name != "" {
			title = gr.Name
		}

		description := gr.Description

		// Build release URL
		releaseURL := fmt.Sprintf("%s/-/releases/%s", strings.TrimSuffix(repoURL, ".git"), gr.TagName)

		releases = append(releases, models.Release{
			Version:     gr.TagName,
			Date:        date,
			Title:       title,
			Description: description,
			URL:         releaseURL,
			Type:        "gitlab-release",
		})
	}

	return releases, nil
}
