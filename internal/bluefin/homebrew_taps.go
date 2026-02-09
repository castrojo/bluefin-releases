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
	"sync"
	"time"

	"github.com/castrojo/bluefin-releases/internal/models"
)

// TapConfig defines a Homebrew tap repository to fetch from
type TapConfig struct {
	Owner        string
	Repo         string
	Experimental bool
}

// GitHubContentItem represents a file in GitHub Contents API response
type GitHubContentItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// FormulaMetadata holds parsed metadata from .rb files
type FormulaMetadata struct {
	Description string
	Homepage    string
	Version     string
	GitHubRepo  string // owner/repo format
}

// FetchUblueOSTapPackages fetches packages from ublue-os Homebrew taps
// Discovers packages dynamically from GitHub repositories
func FetchUblueOSTapPackages() ([]models.App, error) {
	log.Println("Fetching ublue-os tap packages...")

	var allApps []models.App
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Define taps to fetch from
	taps := []TapConfig{
		{Owner: "ublue-os", Repo: "homebrew-tap", Experimental: false},
		{Owner: "ublue-os", Repo: "homebrew-experimental-tap", Experimental: true},
	}

	for _, tap := range taps {
		wg.Add(1)
		go func(t TapConfig) {
			defer wg.Done()

			// Fetch formulae from /Formula directory
			formulae, err := fetchTapDirectory(t.Owner, t.Repo, "Formula", "formula", t.Experimental)
			if err != nil {
				log.Printf("⚠️  Failed to fetch formulae from %s/%s: %v", t.Owner, t.Repo, err)
			} else {
				mu.Lock()
				allApps = append(allApps, formulae...)
				mu.Unlock()
				log.Printf("  ✅ Fetched %d formulae from %s/%s", len(formulae), t.Owner, t.Repo)
			}

			// Fetch casks from /Casks directory
			casks, err := fetchTapDirectory(t.Owner, t.Repo, "Casks", "cask", t.Experimental)
			if err != nil {
				log.Printf("⚠️  Failed to fetch casks from %s/%s: %v", t.Owner, t.Repo, err)
			} else {
				mu.Lock()
				allApps = append(allApps, casks...)
				mu.Unlock()
				log.Printf("  ✅ Fetched %d casks from %s/%s", len(casks), t.Owner, t.Repo)
			}
		}(tap)
	}

	wg.Wait()

	log.Printf("✅ Successfully discovered %d ublue-os tap packages", len(allApps))
	return allApps, nil
}

// fetchTapDirectory lists .rb files from a GitHub repo directory and parses them
func fetchTapDirectory(owner, repo, directory, pkgType string, experimental bool) ([]models.App, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, directory)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Use GITHUB_TOKEN if available for rate limiting
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Directory doesn't exist (some taps may not have Formula or Casks)
		log.Printf("  Directory %s/%s/%s not found (may not exist)", owner, repo, directory)
		return []models.App{}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var files []GitHubContentItem
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var apps []models.App
	for _, file := range files {
		if !strings.HasSuffix(file.Name, ".rb") {
			continue
		}

		// Extract package name (remove .rb extension)
		pkgName := strings.TrimSuffix(file.Name, ".rb")

		// Parse the .rb file
		app, err := parseTapPackage(owner, repo, directory, file.Name, pkgName, pkgType, experimental)
		if err != nil {
			log.Printf("⚠️  Failed to parse %s/%s: %v", directory, file.Name, err)
			continue
		}

		apps = append(apps, app)
	}

	return apps, nil
}
