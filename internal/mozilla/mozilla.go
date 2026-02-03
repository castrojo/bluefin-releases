package mozilla

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/castrojo/bluefin-releases/internal/markdown"
	"github.com/castrojo/bluefin-releases/internal/models"
)

// EnrichWithMozillaReleases fetches release notes for Firefox and Thunderbird
func EnrichWithMozillaReleases(apps []models.App) []models.App {
	log.Println("Enriching Mozilla products with release notes...")

	enrichedApps := make([]models.App, len(apps))
	copy(enrichedApps, apps)

	for i := range enrichedApps {
		app := &enrichedApps[i]

		// Check if this is Firefox or Thunderbird
		if app.ID == "org.mozilla.firefox" {
			if releases, err := fetchFirefoxReleases(); err == nil {
				// Replace the single Flathub release with actual Firefox releases
				app.Releases = releases
				log.Printf("✅ Added %d Firefox releases", len(releases))
			} else {
				log.Printf("⚠️  Failed to fetch Firefox releases: %v", err)
			}
		} else if app.ID == "org.mozilla.Thunderbird" {
			if releases, err := fetchThunderbirdReleases(); err == nil {
				// Replace the single Flathub release with actual Thunderbird releases
				app.Releases = releases
				log.Printf("✅ Added %d Thunderbird releases", len(releases))
			} else {
				log.Printf("⚠️  Failed to fetch Thunderbird releases: %v", err)
			}
		}
	}

	return enrichedApps
}

// fetchFirefoxReleases fetches the latest Firefox release notes
func fetchFirefoxReleases() ([]models.Release, error) {
	// First, get the latest version
	resp, err := http.Get("https://product-details.mozilla.org/1.0/firefox_versions.json")
	if err != nil {
		return nil, fmt.Errorf("fetch version info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read version info: %w", err)
	}

	// Extract version (simple regex since we just need LATEST_FIREFOX_VERSION)
	versionRe := regexp.MustCompile(`"LATEST_FIREFOX_VERSION":\s*"([^"]+)"`)
	matches := versionRe.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find latest version")
	}

	version := matches[1]

	// Fetch the release notes page
	releaseNotesURL := fmt.Sprintf("https://www.mozilla.org/en-US/firefox/%s/releasenotes/", version)
	resp, err = http.Get(releaseNotesURL)
	if err != nil {
		return nil, fmt.Errorf("fetch release notes: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read release notes: %w", err)
	}

	html := string(body)

	// Extract release notes content
	description := extractFirefoxReleaseNotes(html)

	// Parse release date from page if available
	dateStr := extractReleaseDate(html)
	releaseDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("January 2, 2006", dateStr); err == nil {
			releaseDate = parsed
		}
	}

	return []models.Release{
		{
			Version:     version,
			Date:        releaseDate,
			Title:       fmt.Sprintf("Firefox %s", version),
			Description: description,
			URL:         releaseNotesURL,
			Type:        "mozilla-release",
		},
	}, nil
}

// fetchThunderbirdReleases fetches the latest Thunderbird release notes
func fetchThunderbirdReleases() ([]models.Release, error) {
	// Thunderbird uses a similar structure but different API
	resp, err := http.Get("https://product-details.mozilla.org/1.0/thunderbird_versions.json")
	if err != nil {
		return nil, fmt.Errorf("fetch version info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read version info: %w", err)
	}

	// Extract version
	versionRe := regexp.MustCompile(`"LATEST_THUNDERBIRD_VERSION":\s*"([^"]+)"`)
	matches := versionRe.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find latest version")
	}

	version := matches[1]

	// Fetch the release notes page
	releaseNotesURL := fmt.Sprintf("https://www.thunderbird.net/en-US/thunderbird/%s/releasenotes/", version)
	resp, err = http.Get(releaseNotesURL)
	if err != nil {
		return nil, fmt.Errorf("fetch release notes: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read release notes: %w", err)
	}

	html := string(body)

	// Extract release notes content
	description := extractThunderbirdReleaseNotes(html)

	// Parse release date
	dateStr := extractReleaseDate(html)
	releaseDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("January 2, 2006", dateStr); err == nil {
			releaseDate = parsed
		}
	}

	return []models.Release{
		{
			Version:     version,
			Date:        releaseDate,
			Title:       fmt.Sprintf("Thunderbird %s", version),
			Description: description,
			URL:         releaseNotesURL,
			Type:        "mozilla-release",
		},
	}, nil
}

// extractFirefoxReleaseNotes extracts and formats release notes from Firefox HTML
func extractFirefoxReleaseNotes(html string) string {
	var sections []string

	// Extract "New" section
	if newSection := extractSection(html, "New"); newSection != "" {
		sections = append(sections, "## New\n\n"+newSection)
	}

	// Extract "Fixed" section
	if fixedSection := extractSection(html, "Fixed"); fixedSection != "" {
		sections = append(sections, "## Fixed\n\n"+fixedSection)
	}

	// Extract "Changed" section
	if changedSection := extractSection(html, "Changed"); changedSection != "" {
		sections = append(sections, "## Changed\n\n"+changedSection)
	}

	// Extract "Enterprise" section
	if enterpriseSection := extractSection(html, "Enterprise"); enterpriseSection != "" {
		sections = append(sections, "## Enterprise\n\n"+enterpriseSection)
	}

	// Extract "Developer" section
	if devSection := extractSection(html, "Developer"); devSection != "" {
		sections = append(sections, "## Developer\n\n"+devSection)
	}

	if len(sections) == 0 {
		return "<p>Release notes available at the source link.</p>"
	}

	markdownContent := strings.Join(sections, "\n\n")
	return markdown.ToHTML(markdownContent)
}

// extractThunderbirdReleaseNotes extracts and formats release notes from Thunderbird HTML
func extractThunderbirdReleaseNotes(html string) string {
	// Thunderbird has similar structure to Firefox
	return extractFirefoxReleaseNotes(html)
}

// extractSection extracts a specific section from the release notes HTML
func extractSection(html, sectionName string) string {
	// Mozilla's structure: <h3>SectionName</h3> ... <ul> ... </ul>
	// The h3 and ul are in different divs, so we need to match more flexibly

	// Find the section starting from the h3
	sectionStartRe := regexp.MustCompile(fmt.Sprintf(`<h3>%s</h3>`, regexp.QuoteMeta(sectionName)))
	startIdx := sectionStartRe.FindStringIndex(html)
	if startIdx == nil {
		return ""
	}

	// Find the next <ul> after this h3 (use (?s) to match across newlines)
	remaining := html[startIdx[1]:]
	ulRe := regexp.MustCompile(`(?s)<ul>(.*?)</ul>`)
	ulMatches := ulRe.FindStringSubmatch(remaining)
	if len(ulMatches) < 2 {
		return ""
	}

	listHTML := ulMatches[1]

	// Extract list items with release-note class (use (?s) for multiline matching)
	itemRe := regexp.MustCompile(`(?s)<li[^>]*class="release-note"[^>]*>.*?<div class="release-note-content">(.*?)</div>`)
	itemMatches := itemRe.FindAllStringSubmatch(listHTML, -1)

	var items []string
	for _, match := range itemMatches {
		if len(match) > 1 {
			content := match[1]
			// Clean up HTML
			content = cleanHTML(content)
			if content != "" {
				items = append(items, "- "+content)
			}
		}
	}

	return strings.Join(items, "\n")
}

// extractReleaseDate extracts the release date from the HTML
func extractReleaseDate(html string) string {
	dateRe := regexp.MustCompile(`<time[^>]*datetime="([^"]+)"`)
	matches := dateRe.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// cleanHTML removes HTML tags and cleans up text
func cleanHTML(html string) string {
	// Remove line breaks
	text := strings.ReplaceAll(html, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Convert <code> tags to backticks
	text = regexp.MustCompile(`<code[^>]*>`).ReplaceAllString(text, "`")
	text = regexp.MustCompile(`</code>`).ReplaceAllString(text, "`")

	// Convert links to markdown
	linkRe := regexp.MustCompile(`<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	text = linkRe.ReplaceAllString(text, "[$2]($1)")

	// Remove remaining HTML tags
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}
