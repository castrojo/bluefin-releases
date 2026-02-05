package rss

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/castrojo/bluefin-releases/internal/models"
	"github.com/mmcdole/gofeed"
)

// Parser wraps gofeed parser with custom configuration
type Parser struct {
	parser     *gofeed.Parser
	httpClient *http.Client
}

// NewParser creates a new RSS parser with custom HTTP client
func NewParser(timeout time.Duration) *Parser {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	parser := gofeed.NewParser()
	parser.Client = httpClient

	return &Parser{
		parser:     parser,
		httpClient: httpClient,
	}
}

// FetchAndParse fetches and parses an RSS feed from the given URL
func (p *Parser) FetchAndParse(ctx context.Context, url string) (*gofeed.Feed, error) {
	feed, err := p.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("parse RSS feed: %w", err)
	}
	return feed, nil
}

// ConvertToReleases converts RSS feed items to Release structs
func ConvertToReleases(feed *gofeed.Feed, releaseType string) []models.Release {
	releases := make([]models.Release, 0, len(feed.Items))

	for _, item := range feed.Items {
		release := models.Release{
			Version:     extractVersion(item),
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			Type:        releaseType,
		}

		// Parse date (RSS uses Published, Atom uses Updated)
		if item.PublishedParsed != nil {
			release.Date = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			release.Date = *item.UpdatedParsed
		} else {
			release.Date = time.Now()
		}

		releases = append(releases, release)
	}

	return releases
}

// extractVersion tries to extract version from RSS item
func extractVersion(item *gofeed.Item) string {
	// Try to extract version from title (e.g., "v1.2.3 - Release Title")
	if item.Title != "" {
		// Look for version patterns in title
		// This is a simple heuristic - may need refinement
		for _, word := range []string{item.Title} {
			if len(word) > 0 && (word[0] == 'v' || word[0] == 'V') {
				return word
			}
		}
	}

	// Fallback to using GUID or custom fields
	if item.GUID != "" {
		return item.GUID
	}

	return "unknown"
}

// FetchGitHubReleases fetches releases from a GitHub repository RSS feed
func (p *Parser) FetchGitHubReleases(ctx context.Context, owner, repo string) ([]models.Release, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/releases.atom", owner, repo)

	feed, err := p.FetchAndParse(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch GitHub releases: %w", err)
	}

	return ConvertToReleases(feed, "github-release"), nil
}

// FetchFlathubRSS fetches app updates from Flathub RSS feed (if available)
func (p *Parser) FetchFlathubRSS(ctx context.Context) (*gofeed.Feed, error) {
	// Note: Flathub doesn't have a global RSS feed, but individual apps might
	// This is a placeholder for testing
	url := "https://flathub.org/feeds/recently-updated.xml"

	feed, err := p.FetchAndParse(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch Flathub RSS: %w", err)
	}

	return feed, nil
}
