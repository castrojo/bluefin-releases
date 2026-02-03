package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/castrojo/bluefin-releases/internal/bluefin"
	"github.com/castrojo/bluefin-releases/internal/flathub"
	"github.com/castrojo/bluefin-releases/internal/github"
	"github.com/castrojo/bluefin-releases/internal/models"
)

const version = "1.0.0"

func main() {
	// Parse command-line flags
	legacyMode := flag.Bool("legacy", false, "Use legacy mode (fetch recently updated apps instead of Bluefin list)")
	flag.Parse()

	startTime := time.Now()

	log.Printf("Bluefin Releases Pipeline v%s", version)
	if *legacyMode {
		log.Println("Running in LEGACY mode (recently updated apps)")
	} else {
		log.Println("Running in BLUEFIN mode (curated app list)")
	}
	log.Println("Starting data aggregation...")

	// Step 1: Fetch Flathub apps and enrich with details
	var results *models.FetchResults
	flathubStart := time.Now()

	if *legacyMode {
		// Legacy mode: fetch recently updated apps
		log.Println("Fetching recently updated Flathub apps...")
		results = flathub.FetchAllApps()
	} else {
		// Bluefin mode: fetch specific apps from Bluefin Brewfiles
		log.Println("Fetching Bluefin app list...")
		appIDs, err := bluefin.FetchFlatpakList()
		if err != nil {
			log.Fatalf("Failed to fetch Bluefin app list: %v", err)
		}
		log.Printf("Fetching %d Bluefin-curated apps from Flathub...", len(appIDs))
		results = flathub.FetchAllApps(appIDs...)
	}

	flathubDuration := time.Since(flathubStart)
	log.Printf("Fetched and enriched %d apps in %s", len(results.Apps), flathubDuration)

	// Step 2: Enrich with GitHub releases (from actual source repos)
	log.Println("Enriching with GitHub releases from source repositories...")
	githubStart := time.Now()
	enrichedApps := github.EnrichWithGitHubReleases(results.Apps)
	githubDuration := time.Since(githubStart)
	log.Printf("GitHub enrichment complete in %s", githubDuration)

	// Step 3: Collect statistics
	appsWithGitHubRepo := 0
	appsWithChangelogs := 0
	totalReleases := 0

	for _, app := range enrichedApps {
		if app.SourceRepo != nil && app.SourceRepo.Type == "github" {
			appsWithGitHubRepo++
		}
		if len(app.Releases) > 0 {
			appsWithChangelogs++
			totalReleases += len(app.Releases)
		}
	}

	log.Printf("Apps with GitHub repos: %d", appsWithGitHubRepo)
	log.Printf("Apps with changelogs: %d", appsWithChangelogs)
	log.Printf("Total releases: %d", totalReleases)

	// Step 4: Build output structure
	buildDuration := time.Since(startTime)
	output := &models.OutputData{
		Metadata: models.Metadata{
			SchemaVersion: "1.0.0",
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
			GeneratedBy:   fmt.Sprintf("bluefin-releases v%s", version),
			BuildDuration: buildDuration.String(),
			Stats: models.Stats{
				AppsTotal:          len(enrichedApps),
				AppsWithGitHubRepo: appsWithGitHubRepo,
				AppsWithChangelogs: appsWithChangelogs,
				TotalReleases:      totalReleases,
			},
			Performance: models.Performance{
				FlathubFetchDuration: flathubDuration.String(),
				DetailsFetchDuration: flathubDuration.String(), // Combined in FetchAllApps
				GitHubFetchDuration:  githubDuration.String(),
				OutputDuration:       "0s", // Will be updated
			},
		},
		Apps: enrichedApps,
	}

	// Step 5: Write output JSON
	log.Println("Writing output JSON...")
	outputStart := time.Now()
	outputPath := "src/data/apps.json"
	if err := output.WriteJSON(outputPath); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}
	outputDuration := time.Since(outputStart)
	output.Metadata.Performance.OutputDuration = outputDuration.String()

	// Log final summary
	log.Printf("✅ Pipeline complete in %s", buildDuration)
	log.Printf("📊 Output: %s", outputPath)

	// Write summary as JSON for GitHub Actions
	summary := map[string]interface{}{
		"success":             true,
		"duration":            buildDuration.String(),
		"apps_total":          len(enrichedApps),
		"apps_with_github":    appsWithGitHubRepo,
		"apps_with_changelog": appsWithChangelogs,
		"total_releases":      totalReleases,
	}
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(summaryJSON))
}
