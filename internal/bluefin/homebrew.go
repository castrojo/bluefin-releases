package bluefin

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

// HomebrewFormula represents metadata from Homebrew API
type HomebrewFormula struct {
	Name       string   `json:"name"`
	FullName   string   `json:"full_name"`
	Tap        string   `json:"tap"`
	Desc       string   `json:"desc"`
	License    string   `json:"license"`
	Homepage   string   `json:"homepage"`
	Versions   Versions `json:"versions"`
	URLs       URLs     `json:"urls"`
	Deprecated bool     `json:"deprecated"`
	Disabled   bool     `json:"disabled"`
	Bottle     *Bottle  `json:"bottle,omitempty"`
}

type Versions struct {
	Stable string `json:"stable"`
	Head   string `json:"head,omitempty"`
}

type URLs struct {
	Stable URLInfo `json:"stable"`
	Head   URLInfo `json:"head,omitempty"`
}

type URLInfo struct {
	URL string `json:"url"`
}

type Bottle struct {
	Stable BottleStable `json:"stable"`
}

type BottleStable struct {
	Files map[string]interface{} `json:"files"`
}

// FetchHomebrewPackages fetches Homebrew packages from Brewfiles and enriches with metadata
// Returns a slice of App structs compatible with the existing models.
func FetchHomebrewPackages() ([]models.App, error) {
	log.Println("Fetching Bluefin Homebrew packages...")

	// Step 1: Parse Brewfiles to get package names
	packageNames, err := FetchHomebrewList()
	if err != nil {
		return nil, fmt.Errorf("fetch homebrew list: %w", err)
	}

	log.Printf("Fetching metadata for %d Homebrew packages...", len(packageNames))

	// Step 2: Fetch metadata for each package (with concurrency)
	apps := make([]models.App, 0, len(packageNames))
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent requests

	for _, pkgName := range packageNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			app, err := fetchHomebrewPackageMetadata(name)
			if err != nil {
				log.Printf("⚠️  Failed to fetch metadata for %s: %v", name, err)
				return
			}

			if app != nil {
				mu.Lock()
				apps = append(apps, *app)
				mu.Unlock()
			}
		}(pkgName)
	}

	wg.Wait()
	close(semaphore)

	log.Printf("✅ Successfully fetched metadata for %d Homebrew packages", len(apps))
	return apps, nil
}

// fetchHomebrewPackageMetadata fetches metadata for a single Homebrew package
func fetchHomebrewPackageMetadata(packageName string) (*models.App, error) {
	// Check if it's a custom tap package (contains "/")
	if strings.Contains(packageName, "/") {
		// For custom tap packages, we'll create a minimal entry
		// since they're not in the main Homebrew API
		return createMinimalHomebrewApp(packageName), nil
	}

	// Fetch from Homebrew API
	url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", packageName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Package not found in homebrew-core, treat as custom tap
		return createMinimalHomebrewApp(packageName), nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var formula HomebrewFormula
	if err := json.Unmarshal(body, &formula); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Skip deprecated or disabled packages
	if formula.Deprecated || formula.Disabled {
		log.Printf("  Skipping deprecated/disabled package: %s", packageName)
		return nil, nil
	}

	// Check if Linux-compatible (has Linux bottles)
	if !isLinuxCompatible(formula) {
		log.Printf("  Skipping non-Linux package: %s", packageName)
		return nil, nil
	}

	// Convert to App model
	return convertHomebrewFormulaToApp(formula), nil
}

// isLinuxCompatible checks if a formula has Linux bottles
func isLinuxCompatible(formula HomebrewFormula) bool {
	if formula.Bottle == nil {
		// No bottles means source-only (still might be Linux compatible)
		return true
	}

	// Check for Linux-specific bottles
	for filename := range formula.Bottle.Stable.Files {
		if strings.Contains(filename, "linux") {
			return true
		}
	}

	return false
}

// convertHomebrewFormulaToApp converts a Homebrew formula to our App model
func convertHomebrewFormulaToApp(formula HomebrewFormula) *models.App {
	// Clean up the name - remove "homebrew-" prefix if present
	cleanName := strings.TrimPrefix(formula.Name, "homebrew-")

	app := &models.App{
		ID:          fmt.Sprintf("homebrew-%s", formula.Name),
		Name:        cleanName,
		Summary:     formula.Desc,
		Description: formula.Desc,
		Version:     formula.Versions.Stable,
		PackageType: "homebrew",
		FetchedAt:   time.Now(),
		HomebrewInfo: &models.HomebrewInfo{
			Formula:  formula.Name,
			FullName: formula.FullName,
			Tap:      formula.Tap,
			Homepage: formula.Homepage,
			Versions: []string{formula.Versions.Stable},
		},
	}

	// Extract GitHub URL for source repo
	if formula.URLs.Stable.URL != "" {
		sourceRepo := extractGitHubRepoFromURL(formula.URLs.Stable.URL)
		if sourceRepo != nil {
			app.SourceRepo = sourceRepo
		}
	} else if formula.Homepage != "" && strings.Contains(formula.Homepage, "github.com") {
		sourceRepo := extractGitHubRepoFromURL(formula.Homepage)
		if sourceRepo != nil {
			app.SourceRepo = sourceRepo
		}
	}

	return app
}

// extractGitHubRepoFromURL extracts GitHub owner/repo from a URL
func extractGitHubRepoFromURL(urlStr string) *models.SourceRepo {
	// Match patterns like:
	// - https://github.com/owner/repo
	// - https://github.com/owner/repo.git
	// - https://github.com/owner/repo/archive/...
	re := regexp.MustCompile(`github\.com[/:]([^/]+)/([^/\.]+)`)
	matches := re.FindStringSubmatch(urlStr)
	if len(matches) >= 3 {
		return &models.SourceRepo{
			Type:  "github",
			URL:   fmt.Sprintf("https://github.com/%s/%s", matches[1], matches[2]),
			Owner: matches[1],
			Repo:  matches[2],
		}
	}
	return nil
}

// createMinimalHomebrewApp creates a minimal App entry for custom tap packages
func createMinimalHomebrewApp(packageName string) *models.App {
	// Clean up the name - remove "homebrew-" prefix if present
	cleanName := strings.TrimPrefix(packageName, "homebrew-")
	// For tap packages with "/", use the package name after the "/"
	if strings.Contains(cleanName, "/") {
		parts := strings.Split(cleanName, "/")
		cleanName = parts[len(parts)-1]
	}

	return &models.App{
		ID:          fmt.Sprintf("homebrew-%s", strings.ReplaceAll(packageName, "/", "-")),
		Name:        cleanName,
		Summary:     fmt.Sprintf("Homebrew package: %s", cleanName),
		PackageType: "homebrew",
		FetchedAt:   time.Now(),
		HomebrewInfo: &models.HomebrewInfo{
			Formula: packageName,
		},
	}
}

// FetchHomebrewList fetches the list of Homebrew packages that Bluefin includes
// by parsing the Brewfiles from projectbluefin/common repository.
// Returns a slice of Homebrew package names (e.g., "bat", "gh").
// Supports GITHUB_TOKEN environment variable for API rate limits.
func FetchHomebrewList() ([]string, error) {
	log.Println("Fetching Bluefin Homebrew package list from Brewfiles...")

	var allPackages []string

	// List of Brewfiles containing Homebrew package definitions
	brewfiles := []string{
		"system_files/shared/usr/share/ublue-os/homebrew/cli.Brewfile",
		"system_files/shared/usr/share/ublue-os/homebrew/ai-tools.Brewfile",
		"system_files/shared/usr/share/ublue-os/homebrew/k8s-tools.Brewfile",
		"system_files/shared/usr/share/ublue-os/homebrew/ide.Brewfile",
		// Skip fonts, artwork, and experimental for now (too many, less relevant for release tracking)
	}

	for _, brewfile := range brewfiles {
		log.Printf("  Fetching %s...", brewfile)

		content, err := fetchRawFile(BluefinCommonOwner, BluefinCommonRepo, BluefinCommonBranch, brewfile)
		if err != nil {
			log.Printf("⚠️  Failed to fetch %s: %v", brewfile, err)
			continue // Skip this file, but continue with others
		}

		packages := parseHomebrewBrewfile(content)
		log.Printf("  Found %d Homebrew packages in %s", len(packages), brewfile)

		allPackages = append(allPackages, packages...)
	}

	// Deduplicate package names
	allPackages = deduplicate(allPackages)

	log.Printf("✅ Total Homebrew packages: %d", len(allPackages))
	return allPackages, nil
}

// parseHomebrewBrewfile parses a Brewfile and extracts Homebrew package names
// Matches lines like: brew "package-name"
// Ignores tap lines like: tap "owner/repo"
func parseHomebrewBrewfile(content []byte) []string {
	var packages []string

	// Regex pattern: brew "package-name"
	// Note: We ignore tap lines, only extract brew package names
	re := regexp.MustCompile(`brew\s+"([^"]+)"`)

	matches := re.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			packageName := string(match[1])
			packages = append(packages, packageName)
		}
	}

	return packages
}
