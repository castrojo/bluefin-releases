# Bluefin Firehose Transformation Plan

**Status:** Phase 1 of 20 Complete ✅  
**Started:** 2026-02-02  
**Last Updated:** 2026-02-02 21:06 EST

---

## 🎯 Project Vision

Transform **Flatpak Firehose** → **Bluefin Firehose**

**Before:**
- Tracks 250 recently updated Flathub apps
- Generic Flathub branding
- Flatpak-only support

**After:**
- Tracks 42 curated Flatpak apps from Project Bluefin
- Tracks Homebrew packages from Project Bluefin
- Bluefin branding and color scheme
- Core/DX app set filtering
- Flatpak/Homebrew source filtering
- Featured app banner

---

## 📊 Progress Overview

- **Total Phases:** 20
- **Completed:** 1 (5%)
- **In Progress:** 1 (Phase 2)
- **Ready:** 5
- **Blocked:** 13

---

## 🗺️ Complete Roadmap

### ✅ COMPLETED

#### Phase 1: Layout Update (flatpak-firehose-b4d) - DONE
- **Status:** Closed
- **Commit:** `ec8c26a`
- **Changes:**
  - Changed `.apps-grid` from multi-column grid to single-column flex layout
  - Set AppCard `width: 100%`
  - Each app card now takes full width (CNCF firehose style)
- **Files Modified:**
  - `src/pages/index.astro`
  - `src/components/AppCard.astro`

---

### 🚧 IN PROGRESS

#### Phase 2: Download and integrate Bluefin logo (flatpak-firehose-lzi)
- **Status:** In Progress
- **Priority:** P1
- **Blocks:** Phase 3
- **Steps:**
  1. Download SVG logo from `https://projectbluefin.io/favicons/favicon.svg`
  2. Save to `public/bluefin-logo.svg`
  3. Update header in `src/pages/index.astro`:
     - Replace 🔥 emoji with `<img src="/bluefin-logo.svg" alt="Bluefin" />`
     - Add appropriate sizing (height: 32px or 40px)
     - Maintain responsive behavior
  4. Test in dev server

**Logo SVG Content:**
```xml
<svg xmlns="http://www.w3.org/2000/svg" xml:space="preserve" width="256" height="256">
  <path d="M255 128a127 127 0 0 1-127 127A127 127 0 0 1 1 128 127 127 0 0 1 128 1a127 127 0 0 1 127 127" style="fill:#6c7ae9;fill-rule:evenodd;stroke-width:2.20571"/>
  <path d="M130 66.266..." style="fill:#000;..."/>
</svg>
```

---

### 🔜 READY TO START

#### Phase 3: Extract Bluefin color palette and update theme (flatpak-firehose-ybf)
- **Status:** Open
- **Priority:** P1
- **Depends On:** Phase 2
- **Blocks:** Phase 15

**Bluefin Colors Identified:**
- **Primary:** `#6c7ae9` (Bluefin blue)
- **Theme Color:** `#4285f4` (from meta tag)

**Steps:**
1. Fetch `https://projectbluefin.io` and inspect CSS
2. Extract full color palette
3. Update CSS variables in `src/pages/index.astro`:
   ```css
   :root {
     --color-accent-emphasis: #6c7ae9;  /* Bluefin primary */
     --color-text-link: #6c7ae9;
     /* Update other colors as needed */
   }
   
   [data-theme="dark"] {
     --color-accent-emphasis: #8890ff;  /* Lighter for dark mode */
     /* ... */
   }
   ```
4. Test both light and dark themes
5. Verify accessibility (WCAG AA contrast ratios)

---

#### Phase 4: Create Bluefin Flatpak list fetcher (flatpak-firehose-15l)
- **Status:** Open
- **Priority:** P1
- **Blocks:** Phase 5

**Steps:**
1. Create `internal/bluefin/flatpaks.go`
2. Implement fetcher:
   ```go
   package bluefin
   
   import (
       "encoding/base64"
       "encoding/json"
       "fmt"
       "net/http"
       "os"
       "regexp"
   )
   
   const (
       BluefinCommonRepo = "projectbluefin/common"
       SystemFlatpaksPath = "system_files/bluefin/usr/share/ublue-os/homebrew/system-flatpaks.Brewfile"
       SystemDxFlatpaksPath = "system_files/bluefin/usr/share/ublue-os/homebrew/system-dx-flatpaks.Brewfile"
   )
   
   type FlatpakList struct {
       CoreApps []string
       DxApps   []string
   }
   
   func FetchBluefinFlatpaks() (*FlatpakList, error) {
       // Fetch both Brewfiles from GitHub API
       // Support GITHUB_TOKEN env var
       // Parse: ^flatpak "([^"]+)"$
       // Return CoreApps (37) and DxApps (5)
   }
   ```
3. Add tests
4. Handle errors gracefully

**GitHub API Endpoints:**
- `https://api.github.com/repos/projectbluefin/common/contents/{path}`
- Returns base64-encoded file content
- Support `GITHUB_TOKEN` in Authorization header

---

#### Phase 5: Modify Flathub fetcher to accept specific app IDs (flatpak-firehose-78w)
- **Status:** Open
- **Priority:** P1
- **Depends On:** Phase 4
- **Blocks:** Phase 6

**Steps:**
1. Update `internal/flathub/flathub.go`:
   ```go
   // OLD:
   func FetchAllApps() *models.FetchResults
   
   // NEW:
   func FetchAllApps(appIDs []string) *models.FetchResults
   // If appIDs is nil/empty, use legacy "recently-updated" behavior
   // If appIDs provided, fetch those specific apps instead
   ```
2. For specific app IDs, use: `https://flathub.org/api/v2/appstream/{app_id}`
3. Maintain existing enrichment logic (parallel fetching, rate limiting)
4. Ensure backward compatibility for testing

---

#### Phase 6: Integrate Bluefin fetcher into main pipeline (flatpak-firehose-8i3)
- **Status:** Open
- **Priority:** P1
- **Depends On:** Phase 5
- **Blocks:** Phase 7, 14, 16, 20

**Steps:**
1. Update `cmd/flatpak-firehose/main.go`:
   ```go
   func main() {
       // Add mode flag
       mode := flag.String("mode", "bluefin", "Mode: bluefin or legacy")
       flag.Parse()
       
       var appIDs []string
       if *mode == "bluefin" {
           // Fetch Bluefin app list
           list, err := bluefin.FetchBluefinFlatpaks()
           if err != nil {
               log.Fatalf("Failed to fetch Bluefin apps: %v", err)
           }
           appIDs = append(list.CoreApps, list.DxApps...)
           log.Printf("Bluefin mode: tracking %d apps", len(appIDs))
       }
       
       // Fetch apps (bluefin mode or legacy mode)
       results := flathub.FetchAllApps(appIDs)
   }
   ```
2. Update models to track Core vs DX apps
3. Test both modes
4. Default to Bluefin mode

---

#### Phase 7: Add Core/DX filter to FilterBar (flatpak-firehose-k8a)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 6

**Steps:**
1. Update `internal/models/models.go`:
   ```go
   type App struct {
       // ... existing fields
       AppSet string `json:"appSet"` // "core" or "dx"
   }
   ```
2. Update `src/components/FilterBar.astro`:
   - Add "App Set" dropdown: All, Core (37), DX (5)
3. Update `src/components/AppCard.astro`:
   - Add `data-app-set={app.appSet}` attribute
4. Update JavaScript filter logic
5. Update result count display

---

### 📦 HOMEBREW INTEGRATION TRACK

#### Phase 8: Research Homebrew Brewfile structure (flatpak-firehose-8vh)
- **Status:** Open
- **Priority:** P2
- **Blocks:** Phase 10

**Steps:**
1. Analyze Homebrew Brewfiles in projectbluefin/common:
   - `cli.Brewfile`
   - `fonts.Brewfile`
   - `ai-tools.Brewfile`
   - `k8s-tools.Brewfile`
   - Others?
2. Document Brewfile format: `brew "package-name"`
3. Research Homebrew API:
   - Formula metadata: `https://formulae.brew.sh/api/formula/{name}.json`
   - Cask metadata: `https://formulae.brew.sh/api/cask/{name}.json`
4. Identify Linux-compatible packages
5. Document release/version tracking strategy

---

#### Phase 9: Create Homebrew package fetcher (flatpak-firehose-7oo)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 10
- **Blocks:** Phase 13

**Steps:**
1. Create `internal/homebrew/homebrew.go`
2. Implement fetcher:
   ```go
   package homebrew
   
   type Package struct {
       Name        string
       Description string
       Homepage    string
       Version     string
       License     string
       // ... other fields
   }
   
   func FetchBluefinHomebrewPackages() ([]Package, error) {
       // Fetch Brewfiles from projectbluefin/common
       // Parse: ^brew "([^"]+)"$
       // Fetch metadata from formulae.brew.sh API
       // Return package list
   }
   ```
3. Support GITHUB_TOKEN for rate limits
4. Handle formula vs cask distinction
5. Filter Linux-compatible packages only

---

#### Phase 10: Update data models for Homebrew (flatpak-firehose-uze)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 8
- **Blocks:** Phase 9, 12

**Steps:**
1. Update `internal/models/models.go`:
   ```go
   type App struct {
       // ... existing fields
       PackageType string `json:"packageType"` // "flatpak" or "homebrew"
       
       // Homebrew-specific fields
       HomebrewName     string `json:"homebrewName,omitempty"`
       HomebrewFormula  string `json:"homebrewFormula,omitempty"`
       HomebrewHomepage string `json:"homebrewHomepage,omitempty"`
   }
   ```
2. Ensure JSON marshaling works for both types
3. Update FetchResults to handle mixed types
4. Add type guards where needed

---

#### Phase 11: Add Flatpak/Homebrew source filter (flatpak-firehose-efc)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 12

**Steps:**
1. Update `src/components/FilterBar.astro`:
   - Add "Package Type" dropdown: All, Flatpak, Homebrew
2. Update `src/components/AppCard.astro`:
   - Add `data-package-type={app.packageType}` attribute
3. Update JavaScript filter logic
4. Update stats to show counts separately:
   - "X Flatpak Apps"
   - "X Homebrew Packages"
   - "X Total Packages"

---

#### Phase 12: Update AppCard for Homebrew (flatpak-firehose-9vj)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 10
- **Blocks:** Phase 11

**Steps:**
1. Update `src/components/AppCard.astro`:
   - Show package type badge: "Flatpak" or "Homebrew"
   - For Flatpak: link to Flathub
   - For Homebrew: link to formulae.brew.sh or homepage
   - Display Homebrew-specific metadata (formula, tap, etc.)
   - Maintain consistent design
2. Update styles for badges
3. Test both package types

---

#### Phase 13: Integrate Homebrew into pipeline (flatpak-firehose-cif)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 9
- **Blocks:** Phase 17, 18, 20

**Steps:**
1. Update `cmd/flatpak-firehose/main.go`:
   ```go
   func main() {
       // Fetch Bluefin Flatpaks
       flatpaks := bluefin.FetchBluefinFlatpaks()
       flatpakApps := flathub.FetchAllApps(flatpaks)
       
       // Fetch Homebrew packages
       homebrewPkgs := homebrew.FetchBluefinHomebrewPackages()
       
       // Merge results
       allApps := mergeApps(flatpakApps, homebrewPkgs)
       
       // Sort by update date
       sort.Slice(allApps, func(i, j int) bool {
           return allApps[i].UpdatedAt > allApps[j].UpdatedAt
       })
       
       // Output JSON
       writeJSON(allApps)
   }
   ```
2. Test with both package types
3. Verify sorting and display

---

### 🎨 FEATURES & POLISH

#### Phase 14: Create Featured App Banner (flatpak-firehose-14s)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 6

**Steps:**
1. Create `src/components/FeaturedAppBanner.astro`
2. Implement selection algorithm:
   ```typescript
   function selectFeaturedApp(apps: App[]): App {
       // Filter: verified apps with >1000 installs
       const eligible = apps.filter(app => 
           app.isVerified && 
           (app.installsLastMonth || 0) > 1000
       );
       
       // Deterministic by day (same app all day)
       const dayOfYear = Math.floor(Date.now() / 86400000);
       const index = dayOfYear % eligible.length;
       return eligible[index];
   }
   ```
3. Design with Bluefin gradient styling:
   ```css
   .featured-banner {
       background: linear-gradient(135deg, #6c7ae9 0%, #4285f4 100%);
   }
   ```
4. Display: icon, name, summary, "View on Flathub" CTA
5. Add to sidebar above FilterBar

---

### 📝 BRANDING & DOCUMENTATION

#### Phase 15: Update all text and metadata (flatpak-firehose-sgh)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 3

**Changes:**
1. `src/pages/index.astro`:
   - Title: "Flatpak Firehose" → "Bluefin Firehose"
   - Subtitle: "Recently updated applications from Flathub" → "Applications and packages shipped with Project Bluefin"
   - Meta description
   - About box text
   - Footer text
   - GitHub repository links
2. All components: update any hardcoded text

---

#### Phase 16: Update package.json and go.mod (flatpak-firehose-er7)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 6

**Steps:**
1. Update `package.json`:
   ```json
   {
     "name": "bluefin-firehose",
     "description": "Feed reader for Project Bluefin applications and packages",
     "repository": "github:castrojo/bluefin-firehose"
   }
   ```
2. Update `go.mod`:
   ```go
   module github.com/castrojo/bluefin-firehose
   ```
3. Update all import statements across Go files
4. Run `go mod tidy`
5. Test build

---

#### Phase 17: Rewrite README.md (flatpak-firehose-a93)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 13

**Content:**
- Project vision and scope
- Bluefin-specific focus (42 apps + Homebrew)
- Complete architecture diagram
- Full feature list
- Build instructions with GITHUB_TOKEN setup
- Deployment guide
- Contributing section
- Credits to firehose inspiration

---

#### Phase 18: Rewrite AGENTS.md (flatpak-firehose-1gi)
- **Status:** Open
- **Priority:** P2
- **Depends On:** Phase 13

**Content:**
- Bluefin Firehose architecture
- Component descriptions (Bluefin fetcher, Homebrew fetcher)
- Data flow diagrams
- Common development tasks
- Debugging tips
- Testing strategies
- Keep "Landing the Plane" section
- Keep bd workflow section

---

### 🔍 FINALIZATION

#### Phase 19: Dependency management and cleanup (flatpak-firehose-hts)
- **Status:** Open
- **Priority:** P2

**Steps:**
1. Review all phase dependencies: `bd show <id>`
2. Close any remaining superseded issues
3. Verify all 20 phases are properly linked
4. Final beads sync: `bd sync`

---

#### Phase 20: Final testing and validation (flatpak-firehose-e2k)
- **Status:** Open
- **Priority:** P1
- **Depends On:** Phase 13

**Test Checklist:**
- [ ] Build succeeds: `npm run build`
- [ ] All 42 Flatpak apps load correctly
- [ ] Homebrew packages load correctly
- [ ] Core/DX filter works (37 Core, 5 DX)
- [ ] Flatpak/Homebrew filter works
- [ ] Category filter works
- [ ] Verification filter works
- [ ] Date filter works
- [ ] Search works (app names)
- [ ] Keyboard navigation works (j/k/o/?/t)
- [ ] Theme toggle works (light/dark)
- [ ] Featured app banner rotates daily
- [ ] Responsive design (desktop/tablet/mobile)
- [ ] GITHUB_TOKEN support works
- [ ] GitHub Actions workflow succeeds
- [ ] Deployment to GitHub Pages works
- [ ] All links functional
- [ ] Accessibility (WCAG AA)

---

## 🏗️ Architecture Diagram

### Current (Phase 1)

```
┌─────────────────────────────────────────────────────────┐
│ Flathub API                                             │
│ /api/v2/collection/recently-updated → 250 apps         │
└────────────────┬────────────────────────────────────────┘
                 │
                 v
┌─────────────────────────────────────────────────────────┐
│ Go Pipeline (cmd/flatpak-firehose/main.go)             │
│ ├── Limit to 50 apps                                    │
│ ├── Enrich with Flathub details                         │
│ ├── Fetch GitHub releases (if GITHUB_TOKEN)             │
│ └── Output: src/data/apps.json                          │
└────────────────┬────────────────────────────────────────┘
                 │
                 v
┌─────────────────────────────────────────────────────────┐
│ Astro Frontend (src/pages/index.astro)                 │
│ ├── Import apps.json                                    │
│ ├── Render single-column layout                         │
│ ├── Client-side filters (category, verification, date)  │
│ ├── Keyboard navigation                                 │
│ └── Static HTML → GitHub Pages                          │
└─────────────────────────────────────────────────────────┘
```

### Target (After Phase 13)

```
┌──────────────────────────────────────────────────────────┐
│ GitHub API: projectbluefin/common                        │
│ ├── system-flatpaks.Brewfile (37 apps)                  │
│ ├── system-dx-flatpaks.Brewfile (5 apps)                │
│ └── cli.Brewfile, fonts.Brewfile, etc. (Homebrew)       │
└────────────┬─────────────────────────────────────────────┘
             │
             v
┌────────────────────────────────────────────────────────┐
│ internal/bluefin/flatpaks.go                           │
│ Parse 42 Flatpak app IDs (Core + DX)                  │
└────────────┬───────────────────────────────────────────┘
             │
             v
┌────────────────────────────────────────────────────────┐
│ Flathub API                                            │
│ Fetch specific apps by ID (not "recently-updated")    │
└────────────┬───────────────────────────────────────────┘
             │
             ├─────────────┐
             │             │
             v             v
┌────────────────────┐  ┌────────────────────┐
│ internal/flathub/  │  │ internal/homebrew/ │
│ flathub.go         │  │ homebrew.go        │
│ Enrich Flatpak     │  │ Fetch Homebrew     │
│ apps with details  │  │ package metadata   │
└────────────┬───────┘  └────────┬───────────┘
             │                   │
             └─────────┬─────────┘
                       v
┌────────────────────────────────────────────────────────┐
│ cmd/flatpak-firehose/main.go                           │
│ ├── Merge Flatpak + Homebrew data                      │
│ ├── Fetch GitHub releases (if GITHUB_TOKEN)            │
│ ├── Sort by update date                                │
│ └── Output: src/data/apps.json                         │
└────────────┬───────────────────────────────────────────┘
             │
             v
┌────────────────────────────────────────────────────────┐
│ Astro Frontend (src/pages/index.astro)                │
│ ├── Import apps.json (Flatpak + Homebrew)             │
│ ├── Render single-column layout                        │
│ ├── Filters: Core/DX, Flatpak/Homebrew, category, etc.│
│ ├── Featured app banner (daily rotation)               │
│ ├── Keyboard navigation                                │
│ └── Static HTML → GitHub Pages                         │
└────────────────────────────────────────────────────────┘
```

---

## 🎯 Key Decisions Made

### 1. Data Source Strategy
**Decision:** Fetch Brewfiles fresh on every build (no caching)  
**Rationale:** Ensure we always track the exact current state of Bluefin

### 2. Scope
**Decision:** Track only 42 default Flatpak apps + Homebrew packages  
**Rationale:** Bluefin users care about *their* apps, not all of Flathub

### 3. Repository Name
**Decision:** Full rename to `bluefin-firehose` in Phase 16  
**Rationale:** This is a complete rebrand, not just a fork

### 4. Branding
**Decision:** Extract colors/design from projectbluefin.io  
**Rationale:** Maintain visual consistency with Bluefin ecosystem

### 5. Deployment
**Decision:** Keep GitHub Pages (existing setup)  
**Rationale:** Works well, no need for infrastructure changes

### 6. Layout
**Decision:** Single-column "one app per row" style  
**Rationale:** User specifically requested CNCF firehose layout

### 7. Homebrew Integration
**Decision:** Full dual-package support (Flatpak + Homebrew)  
**Rationale:** User wants to track "both" package types

### 8. GITHUB_TOKEN
**Decision:** Support but don't require  
**Rationale:** Works without (60/hour), better with (5000/hour)

---

## 📦 Beads Issue Summary

### By Status
- **Closed:** 1
- **In Progress:** 1
- **Open (Ready):** 5
- **Open (Blocked):** 13
- **Total:** 20

### By Priority
- **P0:** 0
- **P1:** 7 (critical path)
- **P2:** 13 (features & polish)

### Dependencies
- **Phase 1** → Phase 2 → Phase 3 → Phase 15
- **Phase 4** → Phase 5 → Phase 6 → {Phase 7, 14, 16, 20}
- **Phase 8** → Phase 10 → {Phase 9, 12}
- **Phase 9** → Phase 13
- **Phase 12** → Phase 11
- **Phase 13** → {Phase 17, 18, 20}

### Critical Path
1. Phase 1 ✅
2. Phase 2 (in progress)
3. Phase 3
4. Phase 4
5. Phase 5
6. Phase 6
7. Phase 20 (testing)

---

## 🚀 Quick Start for Next Session

### Resume Phase 2 (Logo Integration)

```bash
# Check status
bd show flatpak-firehose-lzi

# Download logo
curl -o public/bluefin-logo.svg https://projectbluefin.io/favicons/favicon.svg

# Edit src/pages/index.astro (replace fire emoji with logo)
# Build and test
npm run build
npm run preview

# Commit
git add public/bluefin-logo.svg src/pages/index.astro
git commit -m "Phase 2: Integrate Bluefin logo"

# Close issue
bd close flatpak-firehose-lzi

# Move to Phase 3
bd update flatpak-firehose-ybf --status=in_progress
```

### View All Ready Work

```bash
bd ready
```

### View Dependency Graph

```bash
bd show flatpak-firehose-8i3  # Phase 6 has most dependents
```

---

## 📚 Additional Resources

- **Project Bluefin:** https://projectbluefin.io
- **Brewfiles Location:** https://github.com/projectbluefin/common/tree/main/system_files/bluefin/usr/share/ublue-os/homebrew
- **Flathub API Docs:** https://flathub.org/api/v2/docs
- **Homebrew Formula API:** https://formulae.brew.sh/docs/api/
- **CNCF Firehose (inspiration):** https://castrojo.github.io/firehose/

---

**Next Steps:** Complete Phase 2 (logo), then Phase 3 (colors), then Phase 4 (Bluefin fetcher)
