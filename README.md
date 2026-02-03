# Bluefin Firehose

A unified update dashboard for Bluefin OS releases, Flatpak applications, and Homebrew packages with real-time changelogs from source repositories.

## Overview

Bluefin Firehose aggregates release information from three sources into one streamlined dashboard:
- **Bluefin OS releases** from the ublue-os/bluefin repository
- **Flatpak applications** curated by Project Bluefin (42 apps from system Brewfiles)
- **Homebrew packages** from Bluefin's CLI, AI, K8s, and IDE tool collections (44 packages)

This creates a comprehensive view of all software updates relevant to Bluefin OS users.

Inspired by [castrojo/firehose](https://github.com/castrojo/firehose), with an emphasis on curated content specific to the Bluefin project.

## Features

- **Multi-Source Tracking**: Bluefin OS + Flatpak apps + Homebrew packages in one feed
- **Real Changelogs**: Fetches release notes from actual source repositories (not packaging repos)
- **Smart Filtering**: Filter by package type (Flatpak/Homebrew/OS), app set (Core/DX), category, verification status
- **Keyboard Navigation**: Full vim-style keyboard shortcuts (j/k, /, o, t, ?)
- **Theme Toggle**: Light/dark mode with system preference detection
- **Inline Changelogs**: Latest 3 releases shown directly (no collapsing)
- **GitHub Integration**: Optional GitHub API token for enhanced release data
- **GitLab Integration**: Fetches releases from GitLab repos (gitlab.com and self-hosted)
- **Static Site**: Fast, deployable to GitHub Pages or any static host
- **Daily Updates**: Automated builds via GitHub Actions

## Quick Start

### Prerequisites

- Go 1.23 or later
- Node.js 20 or later
- (Optional) `GITHUB_TOKEN` for enhanced GitHub API access
  - Without token: Uses Flathub/Homebrew metadata only
  - With token: Fetches rich release notes from 49+ GitHub repositories
  - Get your token at: https://github.com/settings/tokens (read-only access is sufficient)
- (Optional) `GITLAB_TOKEN` for enhanced GitLab API access
  - Without token: Uses public GitLab API (lower rate limits)
  - With token: Higher rate limits for GitLab release fetching
  - Supports gitlab.com and self-hosted GitLab (e.g., gitlab.gnome.org)
  - Enriches 3 GNOME apps: File Roller, Sushi, Firmware
  - Get your token at: https://gitlab.com/-/profile/personal_access_tokens (read_api scope)

### Local Development

```bash
# Install dependencies
npm install
go mod download

# Build everything (Go pipeline + Astro)
npm run build

# Preview the site
npm run preview

# Development mode (requires pre-generated data)
npm run dev
```

### Running the Pipeline Manually

```bash
# Without API tokens (uses Flathub/Homebrew metadata only)
go run cmd/bluefin-releases/main.go

# With GitHub integration (fetches release notes from GitHub repos)
export GITHUB_TOKEN=your_github_token
go run cmd/bluefin-releases/main.go

# With GitLab integration (fetches release notes from GitLab repos)
export GITLAB_TOKEN=your_gitlab_token
go run cmd/bluefin-releases/main.go

# With both GitHub and GitLab integration (recommended)
export GITHUB_TOKEN=your_github_token
export GITLAB_TOKEN=your_gitlab_token
go run cmd/bluefin-releases/main.go

# Or in one line:
GITHUB_TOKEN=your_github_token GITLAB_TOKEN=your_gitlab_token go run cmd/bluefin-releases/main.go
```

**Notes:**
- **GitHub token** enables rich release notes for 49+ apps with GitHub repos
- **GitLab token** enables release notes for 3 GNOME apps (File Roller, Sushi, Firmware) hosted on gitlab.gnome.org
- Both tokens are optional but recommended for complete release data

## Architecture

**Hybrid Stack:** Go backend (data aggregation) + Astro frontend (static site generation)

### Go Backend (`cmd/bluefin-releases/main.go`)

The data pipeline runs in three parallel phases:

1. **Bluefin OS Releases** (`internal/bluefin/releases.go`)
   - Fetches releases from ublue-os/bluefin GitHub repository
   - Parses release notes for version info and changelogs
   - Converts to unified App format

2. **Flatpak Applications** (`internal/bluefin/flatpak.go`)
   - Reads curated app lists from Bluefin's system Brewfiles:
     - `system-flatpaks.Brewfile` (Core apps)
     - `system-dx-flatpaks.Brewfile` (DX/Developer apps)
   - Fetches metadata from Flathub API
   - Enriches with GitHub repo detection

3. **Homebrew Packages** (`internal/bluefin/homebrew.go`)
   - Reads package lists from Bluefin's Homebrew Brewfiles:
     - `cli.Brewfile` (CLI tools)
     - `ai-tools.Brewfile` (AI/ML tools)
     - `k8s-tools.Brewfile` (Kubernetes tools)
     - `ide.Brewfile` (IDE tools)
   - Fetches metadata from Homebrew formulae API
   - Filters for Linux-compatible packages
   - Extracts GitHub repos for release tracking

4. **GitHub Enrichment** (`internal/github/github.go`)
   - Fetches actual release notes from detected GitHub repos
   - Rate-limited and concurrent (respects GitHub API limits)
   - Falls back gracefully when token unavailable

5. **GitLab Enrichment** (`internal/gitlab/gitlab.go`)
   - Fetches actual release notes from detected GitLab repos
   - Supports both gitlab.com and self-hosted GitLab instances
   - Rate-limited and concurrent (respects GitLab API limits)
   - Falls back to public API when token unavailable

**Output:** `src/data/apps.json` (96 packages total)

### Astro Frontend (`src/pages/index.astro`)

Static site generator that:
- Imports JSON data from Go pipeline
- Renders responsive app cards (`src/components/AppCard.astro`)
- Implements filters (`src/components/FilterBar.astro`)
- Adds search and keyboard navigation
- Generates static HTML for deployment

## Project Structure

```
bluefin-releases/
├── cmd/
│   └── bluefin-releases/
│       └── main.go              # Pipeline orchestration
├── internal/
│   ├── models/
│   │   └── models.go            # Unified data structures
│   ├── bluefin/
│   │   ├── flatpak.go           # Bluefin Flatpak fetcher
│   │   ├── homebrew.go          # Bluefin Homebrew fetcher
│   │   └── releases.go          # Bluefin OS releases fetcher
│   ├── flathub/
│   │   └── flathub.go           # Flathub API client
│   ├── github/
│   │   └── github.go            # GitHub API client
│   └── gitlab/
│       └── gitlab.go            # GitLab API client
├── src/
│   ├── pages/
│   │   └── index.astro          # Main page
│   ├── components/
│   │   ├── AppCard.astro        # App display cards
│   │   ├── FilterBar.astro      # Filtering controls
│   │   ├── SearchBar.astro      # Search input
│   │   ├── ThemeToggle.astro    # Dark/light mode
│   │   └── KeyboardHelp.astro   # Keyboard shortcuts modal
│   └── data/
│       └── apps.json            # Generated by pipeline
├── .github/
│   └── workflows/
│       └── deploy.yml           # Automated deployment
├── go.mod                       # Go dependencies
├── package.json                 # Node dependencies
└── astro.config.mjs             # Astro config
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Go Data Pipeline                        │
├─────────────────────────────────────────────────────────────┤
│  1. Fetch Bluefin OS releases (ublue-os/bluefin)          │
│  2. Fetch Bluefin Flatpak apps (from Brewfiles)           │
│  3. Fetch Bluefin Homebrew packages (from Brewfiles)      │
│  4. Enrich with GitHub releases (parallel, rate-limited)   │
│  5. Enrich with GitLab releases (parallel, rate-limited)   │
│  6. Output unified JSON → src/data/apps.json               │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    Astro Static Site                        │
├─────────────────────────────────────────────────────────────┤
│  1. Import apps.json                                        │
│  2. Render app cards with changelogs                       │
│  3. Add filters, search, keyboard nav                      │
│  4. Generate static HTML                                    │
└─────────────────────────────────────────────────────────────┘
                           ↓
                    GitHub Pages
```

## Deployment

### GitHub Actions

Automated deployment runs:
- **Schedule**: Every 6 hours (0:00, 6:00, 12:00, 18:00 UTC)
- **Manual**: Via workflow_dispatch
- **On Push**: When pushing to main branch

Workflow (`.github/workflows/deploy.yml`):
1. Checkout code with submodules (projectbluefin/common for Brewfiles)
2. Run Go pipeline with `GITHUB_TOKEN`
3. Build Astro static site
4. Deploy to GitHub Pages

### Setup GitHub Pages

1. Repository Settings → Pages
2. Source: "GitHub Actions"
3. Workflow deploys automatically

### Environment Variables

**GitHub Actions (automatic):**
- `GITHUB_TOKEN`: Auto-provided by GitHub Actions for API access (already configured in workflow)
- `BASE_URL`: Set in `astro.config.mjs` for GitHub Pages path

**Local Development (optional):**
- `GITHUB_TOKEN`: Set manually for enhanced release data
  - Get token: https://github.com/settings/tokens (read-only access)
  - Export: `export GITHUB_TOKEN=your_token_here`
  - Impact: Enables rich changelogs for 10+ apps with GitHub repos

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` | Move down (next app) |
| `k` | Move up (previous app) |
| `/` or `s` | Focus search |
| `o` or `Enter` | Open focused app link |
| `t` | Toggle theme (light/dark) |
| `h` | Scroll to top |
| `Space` | Page down |
| `Shift+Space` | Page up |
| `?` | Show keyboard help |
| `Esc` | Close modals / blur search |

## Package Sources

### Flatpak Apps (42 total)

Curated from Bluefin's system Brewfiles:
- **Core**: `system_files/bluefin/usr/share/ublue-os/homebrew/system-flatpaks.Brewfile`
- **DX**: `system_files/bluefin/usr/share/ublue-os/homebrew/system-dx-flatpaks.Brewfile`

### Homebrew Packages (44 total)

From Bluefin's Homebrew tool collections:
- **CLI Tools**: `system_files/shared/usr/share/ublue-os/homebrew/cli.Brewfile`
- **AI/ML Tools**: `system_files/shared/usr/share/ublue-os/homebrew/ai-tools.Brewfile`
- **K8s Tools**: `system_files/shared/usr/share/ublue-os/homebrew/k8s-tools.Brewfile`
- **IDE Tools**: `system_files/shared/usr/share/ublue-os/homebrew/ide.Brewfile`

### Bluefin OS Releases (10 total)

Latest releases from:
- **Repository**: `ublue-os/bluefin`
- **Includes**: Stable and GTS streams

## Performance

Typical build times:
- **Flatpak fetch**: ~600-800ms (42 apps, parallel)
- **Homebrew fetch**: ~200-300ms (44 packages, parallel)
- **Bluefin OS fetch**: ~300ms (10 releases)
- **GitHub enrichment**: ~10-20s with token (rate-limited)
- **Astro build**: ~600ms
- **Total pipeline**: ~1-2s (no GitHub) or ~20-30s (with GitHub)

## Credits

- **Design Inspiration**: [castrojo/firehose](https://github.com/castrojo/firehose)
- **Bluefin OS**: [ublue-os/bluefin](https://github.com/ublue-os/bluefin)
- **Data Sources**: Flathub, Homebrew, GitHub
- **Built With**: Go, Astro, TypeScript

## License

MIT
