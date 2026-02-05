package bluefin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/castrojo/bluefin-releases/internal/markdown"
	"github.com/castrojo/bluefin-releases/internal/models"
)

const (
	// GitHub repository for Bluefin OS releases
	BluefinOSOwner  = "ublue-os"
	BluefinOSRepo   = "bluefin"
	BluefinLTSRepo  = "bluefin-lts"
	BluefinImageURL = "ghcr.io/ublue-os/bluefin"
)

// GitHubRelease represents a GitHub release from the API
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// FetchBluefinReleases fetches the latest Bluefin OS releases from GitHub
// Returns a slice of Release structs compatible with the existing models.
// Supports GITHUB_TOKEN environment variable for API rate limits.
func FetchBluefinReleases() ([]models.Release, error) {
	log.Println("Fetching Bluefin OS releases from GitHub...")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=10", BluefinOSOwner, BluefinOSRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	// GitHub API requires a User-Agent header
	req.Header.Set("User-Agent", "bluefin-releases")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded (403) - consider setting GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var githubReleases []GitHubRelease
	if err := json.Unmarshal(body, &githubReleases); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Convert GitHub releases to our Release model
	var releases []models.Release
	for _, ghRelease := range githubReleases {
		// Skip draft and pre-releases
		if ghRelease.Draft || ghRelease.Prerelease {
			continue
		}

		release := models.Release{
			Version:     ghRelease.TagName,
			Date:        ghRelease.PublishedAt,
			Title:       ghRelease.Name,
			Description: parseReleaseNotes(ghRelease.Body),
			URL:         ghRelease.HTMLURL,
			Type:        "bluefin-os-release",
		}

		releases = append(releases, release)
	}

	log.Printf("✅ Fetched %d Bluefin OS releases", len(releases))
	return releases, nil
}

// parseReleaseNotes formats release notes for display
// Converts markdown to HTML for proper rendering in the UI
func parseReleaseNotes(body string) string {
	return markdown.ToHTML(body)
}

// FetchBluefinOSApps fetches Bluefin OS releases and converts them to App objects
// for integration with the unified dashboard. Returns only the latest release for each stream.
func FetchBluefinOSApps() ([]models.App, error) {
	log.Println("Fetching Bluefin OS releases as Apps...")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=10", BluefinOSOwner, BluefinOSRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	req.Header.Set("User-Agent", "bluefin-releases")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded (403) - consider setting GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var githubReleases []GitHubRelease
	if err := json.Unmarshal(body, &githubReleases); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Track latest release for each stream
	latestByStream := make(map[string]*GitHubRelease)

	// Convert GitHub releases to App objects, keeping only latest per stream
	for i := range githubReleases {
		ghRelease := &githubReleases[i]

		// Skip draft and pre-releases
		if ghRelease.Draft || ghRelease.Prerelease {
			continue
		}

		// Parse OS-specific information to get stream
		osInfo := parseOSInfo(*ghRelease)
		stream := osInfo.Stream

		// Only keep the latest release for each stream
		if existing, ok := latestByStream[stream]; !ok || ghRelease.PublishedAt.After(existing.PublishedAt) {
			latestByStream[stream] = ghRelease
		}
	}

	// Convert the latest releases to App objects
	var apps []models.App
	for _, ghRelease := range latestByStream {
		// Parse OS-specific information
		osInfo := parseOSInfo(*ghRelease)

		// Create App object for this OS release
		app := models.App{
			ID:          fmt.Sprintf("bluefin-os-%s", osInfo.Stream),
			Name:        formatOSName(osInfo),
			Summary:     extractSummary(*ghRelease, osInfo),
			Description: ghRelease.Body,
			Icon:        "https://avatars.githubusercontent.com/u/120078124?s=200&v=4", // Bluefin logo from GitHub
			Version:     ghRelease.TagName,
			ReleaseDate: ghRelease.PublishedAt.Format(time.RFC3339),
			UpdatedAt:   ghRelease.PublishedAt.Format(time.RFC3339),
			FlathubURL:  ghRelease.HTMLURL, // Link to GitHub release
			SourceRepo: &models.SourceRepo{
				Type:  "github",
				URL:   fmt.Sprintf("https://github.com/%s/%s", BluefinOSOwner, BluefinOSRepo),
				Owner: BluefinOSOwner,
				Repo:  BluefinOSRepo,
			},
			FetchedAt:   time.Now(),
			PackageType: "os",
			OSInfo:      osInfo,
			Releases: []models.Release{
				{
					Version:     ghRelease.TagName,
					Date:        ghRelease.PublishedAt,
					Title:       ghRelease.Name,
					Description: parseReleaseNotes(ghRelease.Body),
					URL:         ghRelease.HTMLURL,
					Type:        "bluefin-os-release",
				},
			},
		}

		apps = append(apps, app)
	}

	log.Printf("✅ Fetched %d unique Bluefin OS streams (latest of each)", len(apps))
	return apps, nil
}

// FetchBluefinLTSApps fetches Bluefin LTS releases from the bluefin-lts repository
func FetchBluefinLTSApps() ([]models.App, error) {
	log.Println("Fetching Bluefin LTS releases as Apps...")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=10", BluefinOSOwner, BluefinLTSRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	req.Header.Set("User-Agent", "bluefin-releases")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded (403) - consider setting GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var githubReleases []GitHubRelease
	if err := json.Unmarshal(body, &githubReleases); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Get the latest LTS release
	var latestRelease *GitHubRelease
	for i := range githubReleases {
		ghRelease := &githubReleases[i]

		// Skip draft and pre-releases
		if ghRelease.Draft || ghRelease.Prerelease {
			continue
		}

		if latestRelease == nil || ghRelease.PublishedAt.After(latestRelease.PublishedAt) {
			latestRelease = ghRelease
		}
	}

	var apps []models.App
	if latestRelease != nil {
		// Parse LTS-specific information
		osInfo := parseLTSInfo(*latestRelease)

		// Create App object for this OS release
		app := models.App{
			ID:          "bluefin-os-lts",
			Name:        formatOSName(osInfo),
			Summary:     extractSummary(*latestRelease, osInfo),
			Description: latestRelease.Body,
			Icon:        "https://avatars.githubusercontent.com/u/120078124?s=200&v=4", // Bluefin logo from GitHub
			Version:     latestRelease.TagName,
			ReleaseDate: latestRelease.PublishedAt.Format(time.RFC3339),
			UpdatedAt:   latestRelease.PublishedAt.Format(time.RFC3339),
			FlathubURL:  latestRelease.HTMLURL, // Link to GitHub release
			SourceRepo: &models.SourceRepo{
				Type:  "github",
				URL:   fmt.Sprintf("https://github.com/%s/%s", BluefinOSOwner, BluefinLTSRepo),
				Owner: BluefinOSOwner,
				Repo:  BluefinLTSRepo,
			},
			FetchedAt:   time.Now(),
			PackageType: "os",
			OSInfo:      osInfo,
			Releases: []models.Release{
				{
					Version:     latestRelease.TagName,
					Date:        latestRelease.PublishedAt,
					Title:       latestRelease.Name,
					Description: parseReleaseNotes(latestRelease.Body),
					URL:         latestRelease.HTMLURL,
					Type:        "bluefin-os-release",
				},
			},
		}

		apps = append(apps, app)
	}

	log.Printf("✅ Fetched %d Bluefin LTS release(s)", len(apps))
	return apps, nil
}

// parseOSInfo extracts OS-specific information from release data
func parseOSInfo(release GitHubRelease) *models.OSInfo {
	// Parse tag name (e.g., "stable-20260203" or "gts-20260203")
	parts := strings.Split(release.TagName, "-")
	stream := "stable"
	buildNumber := release.TagName
	if len(parts) >= 2 {
		stream = parts[0]
		buildNumber = parts[1]
	}

	// Parse release name to extract Fedora version and commit
	// Format: "stable-20260203: Stable (F43.20260203, #4132884)"
	fedoraVersion := ""
	commitHash := ""
	if nameMatch := regexp.MustCompile(`F(\d+)\.\d+`).FindStringSubmatch(release.Name); len(nameMatch) > 1 {
		fedoraVersion = nameMatch[1]
	}
	if commitMatch := regexp.MustCompile(`#([a-f0-9]+)`).FindStringSubmatch(release.Name); len(commitMatch) > 1 {
		commitHash = commitMatch[1]
	}

	// Extract major package versions from changelog
	kernelVersion := extractPackageVersion(release.Body, "Kernel")
	gnomeVersion := extractPackageVersion(release.Body, "Gnome")
	mesaVersion := extractPackageVersion(release.Body, "Mesa")

	// Extract other major packages
	majorPackages := make(map[string]string)
	if podmanVer := extractPackageVersion(release.Body, "Podman"); podmanVer != "" {
		majorPackages["Podman"] = podmanVer
	}
	if nvidiaVer := extractPackageVersion(release.Body, "Nvidia"); nvidiaVer != "" {
		majorPackages["Nvidia"] = nvidiaVer
	}
	if dockerVer := extractPackageVersion(release.Body, "Docker"); dockerVer != "" {
		majorPackages["Docker"] = dockerVer
	}
	if incusVer := extractPackageVersion(release.Body, "Incus"); incusVer != "" {
		majorPackages["Incus"] = incusVer
	}

	return &models.OSInfo{
		Stream:        stream,
		FedoraVersion: fedoraVersion,
		BuildNumber:   buildNumber,
		CommitHash:    commitHash,
		ImageName:     fmt.Sprintf("%s:%s", BluefinImageURL, stream),
		KernelVersion: kernelVersion,
		GnomeVersion:  gnomeVersion,
		MesaVersion:   mesaVersion,
		MajorPackages: majorPackages,
	}
}

// parseLTSInfo extracts LTS-specific information from release data
func parseLTSInfo(release GitHubRelease) *models.OSInfo {
	// Parse tag name (e.g., "lts-20260203")
	parts := strings.Split(release.TagName, "-")
	buildNumber := release.TagName
	if len(parts) >= 2 {
		buildNumber = parts[1]
	}

	// Parse release name to extract CentOS version and commit
	// Format: "bluefin-lts LTS: 20251223 (c10s, #087b221)"
	centosVersion := ""
	commitHash := ""
	if nameMatch := regexp.MustCompile(`c(\d+)s`).FindStringSubmatch(release.Name); len(nameMatch) > 1 {
		centosVersion = nameMatch[1]
	}
	if commitMatch := regexp.MustCompile(`#([a-f0-9]+)`).FindStringSubmatch(release.Name); len(commitMatch) > 1 {
		commitHash = commitMatch[1]
	}

	// Extract major package versions from changelog
	kernelVersion := extractPackageVersion(release.Body, "Kernel")
	gnomeVersion := extractPackageVersion(release.Body, "Gnome")
	mesaVersion := extractPackageVersion(release.Body, "Mesa")

	// Extract other major packages
	majorPackages := make(map[string]string)
	if podmanVer := extractPackageVersion(release.Body, "Podman"); podmanVer != "" {
		majorPackages["Podman"] = podmanVer
	}
	if nvidiaVer := extractPackageVersion(release.Body, "Nvidia"); nvidiaVer != "" {
		majorPackages["Nvidia"] = nvidiaVer
	}
	if dockerVer := extractPackageVersion(release.Body, "Docker"); dockerVer != "" {
		majorPackages["Docker"] = dockerVer
	}
	if incusVer := extractPackageVersion(release.Body, "Incus"); incusVer != "" {
		majorPackages["Incus"] = incusVer
	}

	return &models.OSInfo{
		Stream:        "lts",
		CentOSVersion: centosVersion,
		BuildNumber:   buildNumber,
		CommitHash:    commitHash,
		ImageName:     fmt.Sprintf("%s:lts", BluefinImageURL),
		KernelVersion: kernelVersion,
		GnomeVersion:  gnomeVersion,
		MesaVersion:   mesaVersion,
		MajorPackages: majorPackages,
	}
}

// extractPackageVersion extracts a package version from the release body
// Looks for lines like "| **Kernel** | 6.17.12-300 |"
func extractPackageVersion(body, packageName string) string {
	// Pattern: | **PackageName** | version |
	pattern := fmt.Sprintf(`\|\s*\*\*%s\*\*\s*\|\s*([^\|]+)\s*\|`, regexp.QuoteMeta(packageName))
	re := regexp.MustCompile(pattern)
	if match := re.FindStringSubmatch(body); len(match) > 1 {
		// Clean up version string (remove arrows and extra whitespace)
		version := strings.TrimSpace(match[1])
		// If there's an arrow (version change), take the second version
		if strings.Contains(version, "➡️") {
			parts := strings.Split(version, "➡️")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		return version
	}
	return ""
}

// formatOSName creates a display name for the OS release
func formatOSName(osInfo *models.OSInfo) string {
	switch osInfo.Stream {
	case "gts":
		return "Bluefin GTS"
	case "lts":
		return "Bluefin LTS"
	default:
		return "Bluefin"
	}
}

// extractSummary creates a concise summary for the OS release
func extractSummary(release GitHubRelease, osInfo *models.OSInfo) string {
	var streamName string
	var baseOS string

	switch osInfo.Stream {
	case "gts":
		streamName = "GTS (General-Term Support)"
		baseOS = fmt.Sprintf("Fedora %s", osInfo.FedoraVersion)
	case "lts":
		streamName = "LTS (Long-Term Support)"
		baseOS = fmt.Sprintf("CentOS Stream %s", osInfo.CentOSVersion)
	default:
		streamName = "Stable"
		baseOS = fmt.Sprintf("Fedora %s", osInfo.FedoraVersion)
	}

	summary := fmt.Sprintf("%s release based on %s", streamName, baseOS)

	if osInfo.KernelVersion != "" {
		summary += fmt.Sprintf(" with Kernel %s", osInfo.KernelVersion)
	}

	return summary
}
