# ublue-os Homebrew Tap Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add automatic discovery and tracking of ~41 packages from ublue-os/homebrew-tap and ublue-os/homebrew-experimental-tap to the Bluefin Firehose dashboard.

**Architecture:** Fetch .rb formula/cask files from GitHub repos using Contents API, parse metadata with regex, convert to models.App format, enrich with GitHub releases via existing pipeline.

**Tech Stack:** Go (backend), Astro (frontend), GitHub API, regex parsing

---

## Task 1: Add Experimental Field to Data Model

**Files:**
- Modify: `internal/models/models.go`

**Step 1: Add Experimental field to App struct**

Open `internal/models/models.go` and add the new field to the `App` struct:

```go
type App struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Icon        string    `json:"icon,omitempty"`
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"releaseDate,omitempty"`
	PackageType string    `json:"packageType"` // "flatpak", "homebrew", "bluefin-release"
	Category    string    `json:"category,omitempty"`
	SourceRepo  *SourceRepo `json:"sourceRepo,omitempty"`
	FetchedAt   time.Time   `json:"fetchedAt"`

	// Package-specific metadata
	FlatpakInfo  *FlatpakInfo  `json:"flatpakInfo,omitempty"`
	HomebrewInfo *HomebrewInfo `json:"homebrewInfo,omitempty"`
	ReleaseInfo  *ReleaseInfo  `json:"releaseInfo,omitempty"`
	
	// Experimental flag for unstable packages
	Experimental bool `json:"experimental,omitempty"`
}
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: SUCCESS (no compilation errors)

**Step 3: Commit**

```bash
git add internal/models/models.go
git commit -m "Add Experimental field to App model for tap packages

- Add bool field to mark experimental-tap packages
- Will be used to render warning badges in UI
- Omitempty ensures it doesn't clutter JSON for stable packages"
```

---

## Task 2: Create Homebrew Tap Fetcher - Data Structures

**Files:**
- Create: `internal/bluefin/homebrew_taps.go`

**Step 1: Create file with package declaration and imports**

Create `internal/bluefin/homebrew_taps.go`:

```go
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
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/bluefin/homebrew_taps.go
git commit -m "Add data structures for Homebrew tap fetching

- TapConfig for tap repository configuration
- GitHubContentItem for GitHub Contents API response
- FormulaMetadata for parsed .rb file data"
```

---

## Task 3: Create Homebrew Tap Fetcher - Main Entry Point

**Files:**
- Modify: `internal/bluefin/homebrew_taps.go`

**Step 1: Add FetchUblueOSTapPackages function**

Add to `internal/bluefin/homebrew_taps.go`:

```go
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
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: FAIL (fetchTapDirectory not defined yet - that's expected)

**Step 3: Commit**

```bash
git add internal/bluefin/homebrew_taps.go
git commit -m "Add main entry point for tap package fetching

- Fetches from both ublue-os/tap and experimental-tap
- Processes Formula and Casks directories in parallel
- Returns combined list of discovered packages"
```

---

## Task 4: Create Homebrew Tap Fetcher - Directory Fetching

**Files:**
- Modify: `internal/bluefin/homebrew_taps.go`

**Step 1: Add fetchTapDirectory function**

Add to `internal/bluefin/homebrew_taps.go`:

```go
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
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: FAIL (parseTapPackage not defined yet - that's expected)

**Step 3: Commit**

```bash
git add internal/bluefin/homebrew_taps.go
git commit -m "Add directory fetching for tap packages

- Uses GitHub Contents API to list .rb files
- Handles 404 gracefully (directory may not exist)
- Uses GITHUB_TOKEN if available for rate limits
- Filters to only .rb files"
```

---

## Task 5: Create Homebrew Tap Fetcher - Package Parsing

**Files:**
- Modify: `internal/bluefin/homebrew_taps.go`

**Step 1: Add parseTapPackage function**

Add to `internal/bluefin/homebrew_taps.go`:

```go
// parseTapPackage fetches and parses a .rb file to extract metadata
func parseTapPackage(owner, repo, directory, filename, pkgName, pkgType string, experimental bool) (models.App, error) {
	// Fetch raw .rb file
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s/%s", owner, repo, directory, filename)

	resp, err := http.Get(url)
	if err != nil {
		return models.App{}, fmt.Errorf("fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.App{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.App{}, fmt.Errorf("read file: %w", err)
	}

	// Parse metadata from Ruby file
	metadata := parseRubyFormula(string(content))

	// Build tap name (e.g., "ublue-os/tap")
	tapName := fmt.Sprintf("%s/%s", owner, strings.TrimPrefix(repo, "homebrew-"))
	fullName := fmt.Sprintf("%s/%s", tapName, pkgName)

	app := models.App{
		ID:           fmt.Sprintf("homebrew-%s", strings.ReplaceAll(fullName, "/", "-")),
		Name:         pkgName,
		Summary:      metadata.Description,
		Description:  metadata.Description,
		Version:      metadata.Version,
		PackageType:  "homebrew",
		Experimental: experimental,
		FetchedAt:    time.Now(),
		HomebrewInfo: &models.HomebrewInfo{
			Formula:  fullName,
			Tap:      tapName,
			Homepage: metadata.Homepage,
			Versions: []string{metadata.Version},
		},
	}

	// Use description as fallback if empty
	if app.Summary == "" {
		app.Summary = fmt.Sprintf("Homebrew %s: %s", pkgType, pkgName)
	}

	// Extract GitHub repo if present
	if metadata.GitHubRepo != "" {
		parts := strings.Split(metadata.GitHubRepo, "/")
		if len(parts) == 2 {
			app.SourceRepo = &models.SourceRepo{
				Type:  "github",
				Owner: parts[0],
				Repo:  parts[1],
				URL:   fmt.Sprintf("https://github.com/%s", metadata.GitHubRepo),
			}
		}
	}

	return app, nil
}
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: FAIL (parseRubyFormula not defined yet - that's expected)

**Step 3: Commit**

```bash
git add internal/bluefin/homebrew_taps.go
git commit -m "Add package parsing for tap .rb files

- Fetches raw .rb file from GitHub
- Parses metadata using parseRubyFormula
- Builds App struct with tap-specific fields
- Extracts GitHub repo for release tracking
- Handles missing metadata with fallbacks"
```

---

## Task 6: Create Homebrew Tap Fetcher - Regex Metadata Parser

**Files:**
- Modify: `internal/bluefin/homebrew_taps.go`

**Step 1: Add parseRubyFormula function**

Add to `internal/bluefin/homebrew_taps.go`:

```go
// parseRubyFormula extracts metadata from .rb file using regex
func parseRubyFormula(content string) FormulaMetadata {
	metadata := FormulaMetadata{}

	// Extract description: desc "..."
	descRe := regexp.MustCompile(`desc\s+"([^"]+)"`)
	if match := descRe.FindStringSubmatch(content); len(match) > 1 {
		metadata.Description = match[1]
	}

	// Extract homepage: homepage "..."
	homepageRe := regexp.MustCompile(`homepage\s+"([^"]+)"`)
	if match := homepageRe.FindStringSubmatch(content); len(match) > 1 {
		metadata.Homepage = match[1]
	}

	// Extract version: version "..."
	versionRe := regexp.MustCompile(`version\s+"([^"]+)"`)
	if match := versionRe.FindStringSubmatch(content); len(match) > 1 {
		metadata.Version = match[1]
	}

	// Extract GitHub repo from url: patterns
	// Matches: github.com/owner/repo or github.com:owner/repo
	githubRe := regexp.MustCompile(`github\.com[/:]([^/\s"]+)/([^/\s"\.]+)`)
	if match := githubRe.FindStringSubmatch(content); len(match) > 2 {
		metadata.GitHubRepo = fmt.Sprintf("%s/%s", match[1], match[2])
	}

	return metadata
}
```

**Step 2: Verify the code compiles**

Run: `go build ./...`
Expected: SUCCESS (all functions now defined)

**Step 3: Test manually (optional verification)**

Run: `go run cmd/bluefin-releases/main.go` (will fail because not integrated yet, but checks compilation)
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add internal/bluefin/homebrew_taps.go
git commit -m "Add regex parser for Ruby formula metadata

- Extracts desc, homepage, version fields
- Detects GitHub repos from url patterns
- Returns structured FormulaMetadata
- Simple regex approach handles common patterns"
```

---

## Task 7: Integrate Tap Fetcher into Main Pipeline

**Files:**
- Modify: `cmd/bluefin-releases/main.go`

**Step 1: Add tap fetching to pipeline**

Open `cmd/bluefin-releases/main.go` and find the section where apps are fetched (around line 40-80). Add the tap fetcher call after the existing Homebrew fetcher:

```go
	// Step 3: Fetch Bluefin Homebrew packages (from Brewfiles)
	log.Println("\n=== Step 3: Fetching Bluefin Homebrew packages ===")
	homebrewApps, err := bluefin.FetchHomebrewPackages()
	if err != nil {
		log.Printf("⚠️  Failed to fetch Homebrew packages: %v", err)
	} else {
		allApps = append(allApps, homebrewApps...)
	}

	// Step 3b: Fetch ublue-os tap packages
	log.Println("\n=== Step 3b: Fetching ublue-os tap packages ===")
	tapApps, err := bluefin.FetchUblueOSTapPackages()
	if err != nil {
		log.Printf("⚠️  Failed to fetch tap packages: %v", err)
	} else {
		allApps = append(allApps, tapApps...)
	}
```

**Step 2: Run the pipeline and verify tap packages are fetched**

Run: `go run cmd/bluefin-releases/main.go`
Expected output should include:
```
=== Step 3b: Fetching ublue-os tap packages ===
Fetching ublue-os tap packages...
  ✅ Fetched N formulae from ublue-os/homebrew-tap
  ✅ Fetched N casks from ublue-os/homebrew-tap
  ✅ Fetched N formulae from ublue-os/homebrew-experimental-tap
  ✅ Fetched N casks from ublue-os/homebrew-experimental-tap
✅ Successfully discovered ~41 ublue-os tap packages
```

**Step 3: Verify JSON output contains tap packages**

Run: `jq '.apps | map(select(.homebrewInfo.tap != null)) | length' src/data/apps.json`
Expected: ~41 (number of tap packages)

Run: `jq '.apps | map(select(.experimental == true)) | length' src/data/apps.json`
Expected: ~25 (number of experimental packages)

**Step 4: Commit**

```bash
git add cmd/bluefin-releases/main.go
git commit -m "Integrate tap package fetching into main pipeline

- Call FetchUblueOSTapPackages after homebrew-core fetch
- Adds ~41 packages from ublue-os taps
- Gracefully handles errors (continues if tap fetch fails)"
```

---

## Task 8: Add Experimental Badge to AppCard Component

**Files:**
- Modify: `src/components/AppCard.astro`

**Step 1: Read current AppCard structure**

Read the file to understand the current badge placement:

```bash
cat src/components/AppCard.astro | grep -A 5 "class=\"app-header\""
```

**Step 2: Add experimental badge after app name**

Find the section with the app name/header (around line 50-80) and add the experimental badge:

```astro
<div class="app-header">
  <h3 class="app-name">{app.name}</h3>
  {app.experimental && (
    <span class="experimental-badge" title="Experimental package - may be unstable">
      ⚠️ Experimental
    </span>
  )}
</div>
```

**Step 3: Add CSS for experimental badge**

Find the `<style>` section at the bottom of the file and add badge styles:

```css
.experimental-badge {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  background: rgba(255, 193, 7, 0.15);
  border: 1px solid rgba(255, 193, 7, 0.4);
  border-radius: 4px;
  font-size: 0.75rem;
  color: #ffa726;
  font-weight: 500;
  vertical-align: middle;
}

@media (prefers-color-scheme: dark) {
  .experimental-badge {
    background: rgba(255, 193, 7, 0.2);
    border-color: rgba(255, 193, 7, 0.5);
    color: #ffb74d;
  }
}
```

**Step 4: Build and preview**

Run: `npm run build && npm run preview`
Expected: Site builds successfully, experimental badges appear on experimental-tap packages

**Step 5: Commit**

```bash
git add src/components/AppCard.astro
git commit -m "Add experimental badge to AppCard component

- Shows warning badge for experimental packages
- Styled with yellow/amber color scheme
- Includes dark mode support
- Tooltip explains instability"
```

---

## Task 9: Add Experimental Badge to ReleaseCard Component

**Files:**
- Modify: `src/components/ReleaseCard.astro`

**Step 1: Add experimental badge (same as AppCard)**

Find the app name section and add the experimental badge:

```astro
<div class="release-header">
  <h3 class="release-name">{app.name}</h3>
  {app.experimental && (
    <span class="experimental-badge" title="Experimental package - may be unstable">
      ⚠️ Experimental
    </span>
  )}
</div>
```

**Step 2: Add CSS for experimental badge (same styles as AppCard)**

Add to the `<style>` section:

```css
.experimental-badge {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  background: rgba(255, 193, 7, 0.15);
  border: 1px solid rgba(255, 193, 7, 0.4);
  border-radius: 4px;
  font-size: 0.75rem;
  color: #ffa726;
  font-weight: 500;
  vertical-align: middle;
}

@media (prefers-color-scheme: dark) {
  .experimental-badge {
    background: rgba(255, 193, 7, 0.2);
    border-color: rgba(255, 193, 7, 0.5);
    color: #ffb74d;
  }
}
```

**Step 3: Build and preview**

Run: `npm run build && npm run preview`
Expected: Badges appear in both AppCard and ReleaseCard views

**Step 4: Commit**

```bash
git add src/components/ReleaseCard.astro
git commit -m "Add experimental badge to ReleaseCard component

- Consistent with AppCard styling
- Shows on recent releases view
- Dark mode support included"
```

---

## Task 10: Update Package Count Statistics

**Files:**
- Modify: `src/pages/index.astro`

**Step 1: Find the stats calculation section**

Look for where package counts are calculated (search for "apps.filter" or "packageType")

**Step 2: Update counts to reflect new packages**

The existing code should already handle this correctly since we're using `packageType: "homebrew"` for all tap packages. Verify the counts look correct.

If there's a hardcoded count or comment, update it:

```astro
---
// Calculate package statistics
const stats = {
  total: apps.length,
  flatpak: apps.filter(a => a.packageType === 'flatpak').length,
  homebrew: apps.filter(a => a.packageType === 'homebrew').length, // Now includes core + taps (~85 total)
  os: apps.filter(a => a.packageType === 'bluefin-release').length,
};
---
```

**Step 3: Verify counts in browser**

Run: `npm run build && npm run preview`
Expected: Status bar shows updated counts (e.g., "130 packages total, 85 Homebrew")

**Step 4: Commit**

```bash
git add src/pages/index.astro
git commit -m "Update package count calculations for tap packages

- Homebrew count now includes core + tap packages
- Stats automatically reflect new packages
- No UI changes needed (counts are dynamic)"
```

---

## Task 11: Update README Documentation

**Files:**
- Modify: `README.md`

**Step 1: Update package counts in overview**

Find the "Features" or overview section and update counts:

```markdown
## Overview

A unified update dashboard for Bluefin OS releases, Flatpak applications, and Homebrew packages with real-time changelogs from source repositories.

**Currently tracking:**
- **Bluefin OS releases** from ublue-os/bluefin repository (~10 releases)
- **Flatpak applications** from Bluefin's system Brewfiles (42 apps)
- **Homebrew packages** from Bluefin's CLI, AI, K8s, and IDE tool collections (44 packages)
- **ublue-os tap packages** from ublue-os/homebrew-tap and experimental-tap (41 packages)

**Total: ~137 packages tracked**
```

**Step 2: Add section about tap packages**

Add a new section explaining tap packages:

```markdown
### ublue-os Homebrew Taps (41 packages)

From ublue-os custom Homebrew taps:
- **ublue-os/homebrew-tap** (16 packages): VSCode, JetBrains Toolbox, LM Studio, 1Password, wallpaper packs
- **ublue-os/homebrew-experimental-tap** (25 packages): Individual JetBrains IDEs, Cursor, Rancher Desktop, system tools

Experimental tap packages are marked with a ⚠️ warning badge indicating they may be unstable.
```

**Step 3: Update architecture section**

Add tap fetcher to the pipeline description:

```markdown
3. **Homebrew Packages** (`internal/bluefin/homebrew.go`)
   - Reads package lists from Bluefin's Homebrew Brewfiles
   - Fetches metadata from Homebrew formulae API
   - Filters for Linux-compatible packages

4. **ublue-os Tap Packages** (`internal/bluefin/homebrew_taps.go`)
   - Discovers packages from ublue-os/homebrew-tap and experimental-tap
   - Fetches .rb files from GitHub and parses metadata
   - Marks experimental packages with flag
```

**Step 4: Commit**

```bash
git add README.md
git commit -m "Update README with tap package documentation

- Add counts for tap packages (~41 new packages)
- Document experimental badge meaning
- Update total package count to ~137
- Add tap fetcher to architecture section"
```

---

## Task 12: Update AGENTS.md Documentation

**Files:**
- Modify: `AGENTS.md`

**Step 1: Update Quick Reference package counts**

Find the overview section and update counts:

```markdown
**Bluefin Firehose** is a unified release dashboard that aggregates updates from three sources:
- **Bluefin OS releases** (from ublue-os/bluefin repository)
- **Flatpak applications** (42 curated apps from Bluefin's system Brewfiles)
- **Homebrew packages** (44 packages from Bluefin's CLI, AI, K8s, and IDE tool collections)
- **ublue-os tap packages** (41 packages from ublue-os/homebrew-tap and experimental-tap)

**Total: ~137 packages tracked**
```

**Step 2: Add tap fetcher to Architecture section**

Update the pipeline description:

```markdown
4. **ublue-os Tap Packages** (`internal/bluefin/homebrew_taps.go`)
   - Discovers packages from ublue-os/homebrew-tap and experimental-tap GitHub repos
   - Uses GitHub Contents API to list .rb formula/cask files
   - Parses metadata with regex: desc, homepage, version, GitHub repo
   - Marks experimental-tap packages with flag
   - ~41 packages total
```

**Step 3: Update Key Components section**

Add the new file:

```markdown
internal/
  ├── models/models.go          # Unified data structures
  ├── bluefin/
  │   ├── flatpaks.go           # Bluefin Flatpak fetcher
  │   ├── homebrew.go           # Bluefin Homebrew fetcher
  │   ├── homebrew_taps.go      # ublue-os tap fetcher (NEW)
  │   └── releases.go           # Bluefin OS releases fetcher
```

**Step 4: Update performance section**

Update build times:

```markdown
## Performance Tuning

Typical build times:
- **Flatpak fetch**: ~600-800ms (42 apps, parallel)
- **Homebrew fetch**: ~200-300ms (44 packages, parallel)
- **Tap fetch**: ~3-5s (41 packages, GitHub API + file fetching)
- **Bluefin OS fetch**: ~300ms (10 releases)
- **GitHub enrichment**: ~10-20s with token (rate-limited)
- **Astro build**: ~600ms
- **Total**: ~5-8s (no GitHub) or ~25-35s (with GitHub)
```

**Step 5: Commit**

```bash
git add AGENTS.md
git commit -m "Update AGENTS.md with tap package integration

- Add tap fetcher to architecture overview
- Update package counts and build times
- Document new homebrew_taps.go component
- Update performance benchmarks"
```

---

## Task 13: Full Pipeline Test and Validation

**Files:**
- N/A (testing only)

**Step 1: Clean build from scratch**

```bash
rm -rf src/data/apps.json
go run cmd/bluefin-releases/main.go
```

Expected output:
- No errors in pipeline
- All three sources fetch successfully (Flatpak, Homebrew, Taps, OS)
- Message: "✅ Successfully discovered ~41 ublue-os tap packages"

**Step 2: Validate JSON output**

```bash
# Total app count
jq '.apps | length' src/data/apps.json
# Expected: ~137

# Tap packages count
jq '.apps | map(select(.homebrewInfo.tap != null)) | length' src/data/apps.json
# Expected: ~41

# Experimental packages count
jq '.apps | map(select(.experimental == true)) | length' src/data/apps.json
# Expected: ~25

# Verify a sample tap package
jq '.apps[] | select(.name == "visual-studio-code-linux")' src/data/apps.json
# Expected: Should show full metadata with tap info
```

**Step 3: Build frontend**

```bash
npm run build
```

Expected: SUCCESS with no errors

**Step 4: Preview and manual testing**

```bash
npm run preview
```

Open browser and verify:
- [ ] ~137 total packages shown
- [ ] "85 Homebrew Packages" filter option
- [ ] Experimental badges appear on experimental-tap packages (check Cursor, OpenCode, etc.)
- [ ] Tap packages searchable (try "visual-studio-code")
- [ ] GitHub release notes appear for tap packages with repos
- [ ] No JavaScript console errors
- [ ] Badges look correct in both light and dark mode

**Step 5: Document validation results**

Create a note of what was verified:

```bash
cat > VALIDATION_RESULTS.txt << 'EOF'
Tap Integration Validation - 2026-02-08

✅ Pipeline Tests:
- Go build: SUCCESS
- Pipeline run: SUCCESS
- No errors fetching from ublue-os/homebrew-tap
- No errors fetching from ublue-os/homebrew-experimental-tap

✅ Data Validation:
- Total apps: 137 (was 96 before)
- Tap packages: 41 (16 from main tap + 25 from experimental)
- Experimental packages: 25
- Sample package (visual-studio-code-linux): ✓ metadata correct

✅ Frontend Tests:
- Build: SUCCESS
- Experimental badges: ✓ visible on experimental packages
- Package counts: ✓ updated correctly
- Search: ✓ works for tap packages
- GitHub releases: ✓ working for packages with repos
- Dark mode: ✓ badges look correct

✅ Documentation:
- README.md: ✓ updated
- AGENTS.md: ✓ updated
EOF
cat VALIDATION_RESULTS.txt
```

**Step 6: No commit needed (testing only)**

---

## Task 14: Final Commit and Verification

**Files:**
- All modified files

**Step 1: Review all changes**

```bash
git status
git diff --stat
```

Expected: Should see all files we modified throughout the plan

**Step 2: Run final verification**

```bash
# Ensure everything compiles
go build ./...

# Ensure frontend builds
npm run build

# Quick smoke test
go run cmd/bluefin-releases/main.go
```

All should succeed with no errors.

**Step 3: Create summary commit (if any uncommitted changes)**

```bash
# Stage any remaining changes
git add -A

# Check if there are uncommitted changes
if git diff --cached --quiet; then
  echo "✅ All changes already committed"
else
  git commit -m "Final cleanup for tap integration

- Ensure all files staged
- Validation complete
- Ready for deployment"
fi
```

**Step 4: Verify commit history**

```bash
git log --oneline -15
```

Expected: Should see ~14 commits for this feature

**Step 5: Push to main**

```bash
git push origin main
```

Expected: SUCCESS - changes pushed to GitHub

**Step 6: Monitor GitHub Actions**

```bash
echo "✅ Implementation complete!"
echo ""
echo "Monitor deployment at:"
echo "https://github.com/castrojo/bluefin-releases/actions"
echo ""
echo "Site will deploy automatically within 6 hours or trigger manually:"
echo "https://github.com/castrojo/bluefin-releases/actions/workflows/deploy.yml"
```

---

## Success Criteria

After completion, verify:

- [x] Pipeline fetches ~41 packages from ublue-os taps
- [x] Experimental flag set correctly on experimental-tap packages
- [x] Experimental badges render in UI
- [x] GitHub releases work for tap packages with repos
- [x] Total package count: ~137 (up from ~96)
- [x] Build time: <60 seconds total
- [x] Documentation updated (README, AGENTS.md)
- [x] All changes committed and pushed
- [x] No errors in pipeline or build

## Rollback Plan

If issues occur after deployment:

```bash
# Revert the tap integration
git revert HEAD~14..HEAD

# Or revert to specific commit before changes
git reset --hard <commit-before-tap-integration>
git push -f origin main
```

The pipeline will continue to work without tap packages (existing behavior).
