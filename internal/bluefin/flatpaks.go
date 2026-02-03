package bluefin

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

const (
	// GitHub repository containing Bluefin Brewfiles
	BluefinCommonOwner  = "projectbluefin"
	BluefinCommonRepo   = "common"
	BluefinCommonBranch = "main"
)

// AppSetInfo contains app ID and its app set classification
type AppSetInfo struct {
	AppID  string
	AppSet string // "core" or "dx"
}

// FetchFlatpakList fetches the list of Flatpak app IDs that Bluefin ships with
// by parsing the Brewfiles from projectbluefin/common repository.
// Returns a slice of Flatpak app IDs (e.g., "org.gnome.Calculator").
// Supports GITHUB_TOKEN environment variable for API rate limits.
func FetchFlatpakList() ([]string, error) {
	appSetInfos, err := FetchFlatpakListWithAppSets()
	if err != nil {
		return nil, err
	}

	// Extract just the app IDs for backward compatibility
	appIDs := make([]string, len(appSetInfos))
	for i, info := range appSetInfos {
		appIDs[i] = info.AppID
	}

	return appIDs, nil
}

// FetchFlatpakListWithAppSets fetches the list of Flatpak app IDs with app set classification
// Returns a slice of AppSetInfo containing app IDs and their app set (core/dx).
func FetchFlatpakListWithAppSets() ([]AppSetInfo, error) {
	log.Println("Fetching Bluefin Flatpak list from Brewfiles...")

	var allAppSetInfos []AppSetInfo

	// Map of Brewfiles to their app set classification
	brewfiles := map[string]string{
		"system_files/bluefin/usr/share/ublue-os/homebrew/system-flatpaks.Brewfile":    "core",
		"system_files/bluefin/usr/share/ublue-os/homebrew/system-dx-flatpaks.Brewfile": "dx",
	}

	for brewfile, appSet := range brewfiles {
		log.Printf("  Fetching %s (%s apps)...", brewfile, appSet)

		content, err := fetchRawFile(BluefinCommonOwner, BluefinCommonRepo, BluefinCommonBranch, brewfile)
		if err != nil {
			log.Printf("⚠️  Failed to fetch %s: %v", brewfile, err)
			continue // Skip this file, but continue with others
		}

		appIDs := parseFlatpakBrewfile(content)
		log.Printf("  Found %d Flatpak app IDs in %s", len(appIDs), brewfile)

		for _, appID := range appIDs {
			allAppSetInfos = append(allAppSetInfos, AppSetInfo{
				AppID:  appID,
				AppSet: appSet,
			})
		}
	}

	// Count by app set
	coreCount := 0
	dxCount := 0
	for _, info := range allAppSetInfos {
		if info.AppSet == "core" {
			coreCount++
		} else if info.AppSet == "dx" {
			dxCount++
		}
	}

	log.Printf("✅ Total Flatpak app IDs: %d (Core: %d, DX: %d)", len(allAppSetInfos), coreCount, dxCount)
	return allAppSetInfos, nil
}

// fetchRawFile fetches a raw file from GitHub using raw.githubusercontent.com
// Supports optional GITHUB_TOKEN for authentication (helps with rate limits)
func fetchRawFile(owner, repo, branch, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add GitHub token if available (optional, helps with rate limits)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file not found (404): %s", path)
	}

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

	return body, nil
}

// parseFlatpakBrewfile parses a Brewfile and extracts Flatpak app IDs
// Matches lines like: flatpak "org.gnome.Calculator"
func parseFlatpakBrewfile(content []byte) []string {
	var appIDs []string

	// Regex pattern: flatpak "app.id.here"
	re := regexp.MustCompile(`flatpak\s+"([^"]+)"`)

	matches := re.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			appID := string(match[1])
			appIDs = append(appIDs, appID)
		}
	}

	return appIDs
}

// deduplicate removes duplicate strings from a slice
func deduplicate(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
