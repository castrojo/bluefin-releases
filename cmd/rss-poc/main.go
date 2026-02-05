package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/castrojo/bluefin-releases/internal/models"
	"github.com/castrojo/bluefin-releases/internal/rss"
	"github.com/google/go-github/v57/github"
)

func main() {
	fmt.Println("=== Go RSS Library Evaluation - Proof of Concept ===\n")

	// Test 1: Fetch GitHub releases via RSS (using gofeed)
	fmt.Println("Test 1: Fetching GitHub releases via RSS (ublue-os/bluefin)")
	rssStart := time.Now()
	releases, err := fetchGitHubReleasesViaRSS()
	rssDuration := time.Since(rssStart)

	if err != nil {
		log.Printf("Error fetching via RSS: %v\n", err)
	} else {
		fmt.Printf("✓ Fetched %d releases via RSS in %v\n", len(releases), rssDuration)
		if len(releases) > 0 {
			fmt.Printf("  Latest: %s - %s (%s)\n", releases[0].Version, releases[0].Title, releases[0].Date.Format("2006-01-02"))
		}
	}

	fmt.Println()

	// Test 2: Fetch GitHub releases via API (current approach)
	fmt.Println("Test 2: Fetching GitHub releases via API (ublue-os/bluefin)")
	apiStart := time.Now()
	apiReleases, err := fetchGitHubReleasesViaAPI()
	apiDuration := time.Since(apiStart)

	if err != nil {
		log.Printf("Error fetching via API: %v\n", err)
	} else {
		fmt.Printf("✓ Fetched %d releases via API in %v\n", len(apiReleases), apiDuration)
		if len(apiReleases) > 0 {
			fmt.Printf("  Latest: %s - %s (%s)\n",
				*apiReleases[0].TagName,
				*apiReleases[0].Name,
				apiReleases[0].PublishedAt.Format("2006-01-02"))
		}
	}

	fmt.Println()

	// Performance comparison
	fmt.Println("=== Performance Comparison ===")
	fmt.Printf("RSS:  %v\n", rssDuration)
	fmt.Printf("API:  %v\n", apiDuration)

	speedup := float64(apiDuration) / float64(rssDuration)
	if rssDuration < apiDuration {
		fmt.Printf("RSS is %.2fx faster than API\n", speedup)
	} else {
		fmt.Printf("API is %.2fx faster than RSS\n", 1.0/speedup)
	}

	fmt.Println()

	// Test 3: Error handling
	fmt.Println("Test 3: Error handling (invalid feed)")
	parser := rss.NewParser(10 * time.Second)
	ctx := context.Background()
	_, err = parser.FetchAndParse(ctx, "https://example.com/nonexistent.xml")
	if err != nil {
		fmt.Printf("✓ Error handled gracefully: %v\n", err)
	}

	fmt.Println()

	// Test 4: Export sample data
	fmt.Println("Test 4: Exporting sample RSS data to JSON")
	if len(releases) > 0 {
		exportData := map[string]interface{}{
			"releases":      releases[:min(5, len(releases))],
			"rssDuration":   rssDuration.String(),
			"apiDuration":   apiDuration.String(),
			"speedupFactor": speedup,
		}

		data, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			log.Printf("Error marshaling JSON: %v\n", err)
		} else {
			err = os.WriteFile("rss-poc-output.json", data, 0644)
			if err != nil {
				log.Printf("Error writing file: %v\n", err)
			} else {
				fmt.Println("✓ Sample data exported to rss-poc-output.json")
			}
		}
	}

	fmt.Println("\n=== Proof of Concept Complete ===")
}

func fetchGitHubReleasesViaRSS() ([]models.Release, error) {
	parser := rss.NewParser(30 * time.Second)
	ctx := context.Background()

	return parser.FetchGitHubReleases(ctx, "ublue-os", "bluefin")
}

func fetchGitHubReleasesViaAPI() ([]*github.RepositoryRelease, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	opts := &github.ListOptions{PerPage: 10}
	releases, _, err := client.Repositories.ListReleases(ctx, "ublue-os", "bluefin", opts)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}

	return releases, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
