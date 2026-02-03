package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/castrojo/bluefin-releases/internal/models"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// EnrichWithGitHubReleases fetches GitHub releases for apps with GitHub repos
// and adds them to the app's release list (prioritizing actual source changelogs)
func EnrichWithGitHubReleases(apps []models.App) []models.App {
	// Check if GitHub token is available
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("⚠️  No GITHUB_TOKEN found, skipping GitHub release fetching")
		return apps
	}

	// Create GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	enrichedApps := make([]models.App, len(apps))
	copy(enrichedApps, apps)

	// Process apps with GitHub repos in parallel
	for i := range enrichedApps {
		app := &enrichedApps[i]
		if app.SourceRepo == nil || app.SourceRepo.Type != "github" || app.SourceRepo.Owner == "" || app.SourceRepo.Repo == "" {
			continue
		}

		wg.Add(1)
		go func(app *models.App) {
			defer wg.Done()

			releases, err := fetchGitHubReleases(ctx, client, app.SourceRepo.Owner, app.SourceRepo.Repo)
			if err != nil {
				log.Printf("⚠️  Failed to fetch GitHub releases for %s/%s: %v",
					app.SourceRepo.Owner, app.SourceRepo.Repo, err)
				return
			}

			mu.Lock()
			// Prepend GitHub releases (they are from actual source, so prioritize them)
			app.Releases = append(releases, app.Releases...)
			log.Printf("✅ Added %d GitHub releases for %s", len(releases), app.ID)
			mu.Unlock()

			// Rate limiting: GitHub has a rate limit of 60 requests/hour for unauthenticated
			// and 5000/hour for authenticated. This conservative sleep helps avoid hitting limits.
			// In production with GITHUB_TOKEN, this is overly conservative but safe.
			time.Sleep(500 * time.Millisecond)
		}(app)
	}

	wg.Wait()
	return enrichedApps
}

// fetchGitHubReleases fetches the latest releases from a GitHub repository
func fetchGitHubReleases(ctx context.Context, client *github.Client, owner, repo string) ([]models.Release, error) {
	// Fetch up to 5 latest releases
	opts := &github.ListOptions{PerPage: 5}
	githubReleases, _, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}

	var releases []models.Release
	for _, gr := range githubReleases {
		if gr.TagName == nil {
			continue
		}

		date := time.Now()
		if gr.PublishedAt != nil {
			date = gr.PublishedAt.Time
		}

		title := *gr.TagName
		if gr.Name != nil && *gr.Name != "" {
			title = *gr.Name
		}

		description := ""
		if gr.Body != nil {
			description = *gr.Body
		}

		url := ""
		if gr.HTMLURL != nil {
			url = *gr.HTMLURL
		}

		releases = append(releases, models.Release{
			Version:     *gr.TagName,
			Date:        date,
			Title:       title,
			Description: description,
			URL:         url,
			Type:        "github-release",
		})
	}

	return releases, nil
}
