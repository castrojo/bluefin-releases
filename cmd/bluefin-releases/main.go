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

	// Step 1: Fetch Flatpak apps and enrich with details
	var flatpakApps []models.App
	flathubStart := time.Now()

	if *legacyMode {
		// Legacy mode: fetch recently updated apps
		log.Println("Fetching recently updated Flathub apps...")
		results := flathub.FetchAllApps()
		flatpakApps = results.Apps
	} else {
		// Bluefin mode: fetch specific apps from Bluefin Brewfiles
		log.Println("Fetching Bluefin app list...")
		appSetInfos, err := bluefin.FetchFlatpakListWithAppSets()
		if err != nil {
			log.Fatalf("Failed to fetch Bluefin app list: %v", err)
		}

		// Create app set map for lookup
		appSetMap := make(map[string]string)
		appIDs := make([]string, len(appSetInfos))
		for i, info := range appSetInfos {
			appIDs[i] = info.AppID
			appSetMap[info.AppID] = info.AppSet
		}

		log.Printf("Fetching %d Bluefin-curated Flatpak apps from Flathub...", len(appIDs))
		results := flathub.FetchAllApps(appIDs...)
		flatpakApps = results.Apps

		// Add app set information to each app
		for i := range flatpakApps {
			if appSet, ok := appSetMap[flatpakApps[i].ID]; ok {
				flatpakApps[i].AppSet = appSet
			}
		}
	}

	flathubDuration := time.Since(flathubStart)
	log.Printf("Fetched and enriched %d Flatpak apps in %s", len(flatpakApps), flathubDuration)

	// Step 2: Fetch Homebrew packages (Bluefin mode only)
	var homebrewApps []models.App
	homebrewDuration := time.Duration(0)

	if !*legacyMode {
		log.Println("Fetching Homebrew packages...")
		homebrewStart := time.Now()

		var err error
		homebrewApps, err = bluefin.FetchHomebrewPackages()
		if err != nil {
			log.Printf("⚠️  Failed to fetch Homebrew packages: %v", err)
		} else {
			homebrewDuration = time.Since(homebrewStart)
			log.Printf("Fetched %d Homebrew packages in %s", len(homebrewApps), homebrewDuration)
		}
	}

	// Step 3: Merge Flatpak and Homebrew apps
	allApps := append(flatpakApps, homebrewApps...)
	log.Printf("Total apps: %d (%d Flatpak + %d Homebrew)", len(allApps), len(flatpakApps), len(homebrewApps))

	// Step 4: Enrich with GitHub releases (from actual source repos)
	log.Println("Enriching with GitHub releases from source repositories...")
	githubStart := time.Now()
	enrichedApps := github.EnrichWithGitHubReleases(allApps)
	githubDuration := time.Since(githubStart)
	log.Printf("GitHub enrichment complete in %s", githubDuration)

	// Step 5: Sort by update date (Flatpak apps have updatedAt, Homebrew may not)
	// For now, just use the order they come in (Flatpak first, then Homebrew)
	// Future: could sort by latest release date

	// Step 6: Collect statistics
	appsWithGitHubRepo := 0
	appsWithChangelogs := 0
	totalReleases := 0
	flatpakCount := 0
	homebrewCount := 0

	for _, app := range enrichedApps {
		if app.SourceRepo != nil && app.SourceRepo.Type == "github" {
			appsWithGitHubRepo++
		}
		if len(app.Releases) > 0 {
			appsWithChangelogs++
			totalReleases += len(app.Releases)
		}
		if app.PackageType == "flatpak" {
			flatpakCount++
		} else if app.PackageType == "homebrew" {
			homebrewCount++
		}
	}

	log.Printf("Apps with GitHub repos: %d", appsWithGitHubRepo)
	log.Printf("Apps with changelogs: %d", appsWithChangelogs)
	log.Printf("Total releases: %d", totalReleases)

	// Step 7: Build output structure
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

	// Step 8: Write output JSON
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
	log.Printf("📦 Packages: %d Flatpak + %d Homebrew = %d total", flatpakCount, homebrewCount, len(enrichedApps))

	// Write summary as JSON for GitHub Actions
	summary := map[string]interface{}{
		"success":             true,
		"duration":            buildDuration.String(),
		"apps_total":          len(enrichedApps),
		"flatpak_count":       flatpakCount,
		"homebrew_count":      homebrewCount,
		"apps_with_github":    appsWithGitHubRepo,
		"apps_with_changelog": appsWithChangelogs,
		"total_releases":      totalReleases,
	}
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(summaryJSON))
}
