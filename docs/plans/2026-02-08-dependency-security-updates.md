# Dependency Security Updates Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 7 Dependabot security alerts by updating Go dependencies and configure Renovate bot for automated future updates

**Architecture:** Surgical dependency update (golang.org/x/net and golang.org/x/oauth2) followed by Renovate bot configuration for automated maintenance

**Tech Stack:** Go modules, GitHub Renovate Bot

---

## Task 1: Capture Baseline

**Files:**
- Create: `docs/deps-before.txt`

**Step 1: Capture current dependency versions**

Run:
```bash
go list -m all > docs/deps-before.txt
```

Expected: File created with all current Go module versions

**Step 2: Verify current pipeline works**

Run:
```bash
go run cmd/bluefin-releases/main.go
```

Expected: Pipeline completes successfully with "✅ Pipeline complete" message

**Step 3: Verify current build works**

Run:
```bash
go build ./...
```

Expected: No errors, builds successfully

**Step 4: Commit baseline**

Run:
```bash
git add docs/deps-before.txt
git commit -m "docs: capture dependency baseline before security updates"
```

Expected: Commit created successfully

---

## Task 2: Update Go Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Update golang.org/x/net to latest**

Run:
```bash
go get -u golang.org/x/net@latest
```

Expected: Output shows version update (e.g., "golang.org/x/net v0.4.0 => v0.33.0")

**Step 2: Update golang.org/x/oauth2 to latest**

Run:
```bash
go get -u golang.org/x/oauth2@latest
```

Expected: Output shows version update

**Step 3: Clean up dependencies**

Run:
```bash
go mod tidy
```

Expected: Removes unused dependencies, updates go.sum

**Step 4: Verify build succeeds**

Run:
```bash
go build ./...
```

Expected: No errors, builds successfully

**Step 5: Capture updated dependencies**

Run:
```bash
go list -m all > docs/deps-after.txt
```

Expected: File created with updated versions

---

## Task 3: Verify Pipeline Functionality

**Files:**
- Reference: `cmd/bluefin-releases/main.go`
- Reference: `src/data/apps.json`

**Step 1: Run full pipeline**

Run:
```bash
go run cmd/bluefin-releases/main.go
```

Expected: Pipeline completes successfully, outputs JSON with 130 packages

**Step 2: Verify JSON output is valid**

Run:
```bash
jq '.apps | length' src/data/apps.json
```

Expected: Output shows "130"

**Step 3: Check for any new warnings or errors**

Review the pipeline output for:
- ❌ No new error messages
- ⚠️ No new warning messages
- ✅ Same success metrics as before

**Step 4: Verify frontend build**

Run:
```bash
npm run build
```

Expected: Build completes successfully with "✓ Completed in XXms"

---

## Task 4: Commit Dependency Updates

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Create: `docs/deps-after.txt`

**Step 1: Compare dependency changes**

Run:
```bash
diff docs/deps-before.txt docs/deps-after.txt
```

Expected: Shows updated versions of golang.org/x/net and golang.org/x/oauth2 (and their transitive deps)

**Step 2: Stage all dependency files**

Run:
```bash
git add go.mod go.sum docs/deps-after.txt
```

Expected: Files staged for commit

**Step 3: Commit with security context**

Run:
```bash
git commit -m "security: update golang.org/x/net and golang.org/x/oauth2

Fixes 7 Dependabot security alerts:
- golang.org/x/net: v0.4.0 -> v0.33.0+ (6 vulnerabilities)
  - CVE: HTTP/2 rapid reset DoS
  - CVE: HPACK decoder DoS
  - CVE: XSS in HTML parser
  - CVE: HTTP proxy bypass via IPv6 Zone IDs
  - CVE: Stream cancellation attack
  - CVE: Improper text node rendering
- golang.org/x/oauth2: updated to latest (1 vulnerability)
  - CVE: Input validation issue

Impact: Low risk - these affect HTTP/2 server-side code, but this
project is a CLI tool (HTTP client only). Updated for security
hygiene and to remove GitHub warnings.

Verification:
- go build ./... succeeds
- Pipeline runs successfully (130 packages)
- Frontend builds without errors
- No functional regressions detected"
```

Expected: Commit created with detailed security context

---

## Task 5: Configure Renovate Bot

**Files:**
- Create: `.github/renovate.json`

**Step 1: Create Renovate configuration**

Create file `.github/renovate.json` with content:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended"
  ],
  "schedule": [
    "before 6am on Monday"
  ],
  "packageRules": [
    {
      "description": "Auto-merge Go patch and minor updates",
      "matchDatasources": ["go"],
      "matchUpdateTypes": ["patch", "minor"],
      "automerge": true
    },
    {
      "description": "Group Go major updates",
      "matchDatasources": ["go"],
      "matchUpdateTypes": ["major"],
      "groupName": "Go major dependencies"
    },
    {
      "description": "Auto-merge npm patch updates",
      "matchDatasources": ["npm"],
      "matchDepTypes": ["devDependencies"],
      "matchUpdateTypes": ["patch"],
      "automerge": true
    },
    {
      "description": "Group npm minor updates",
      "matchDatasources": ["npm"],
      "matchUpdateTypes": ["minor"],
      "groupName": "npm minor updates"
    },
    {
      "description": "Separate PR for npm major updates",
      "matchDatasources": ["npm"],
      "matchUpdateTypes": ["major"],
      "groupName": "npm major updates"
    },
    {
      "description": "Priority for security updates",
      "matchDatasources": ["go", "npm"],
      "vulnerabilityAlerts": {
        "enabled": true
      },
      "prPriority": 10,
      "schedule": ["at any time"]
    }
  ],
  "prConcurrentLimit": 3,
  "prHourlyLimit": 2,
  "labels": ["dependencies"],
  "commitMessagePrefix": "chore(deps):",
  "rebaseWhen": "conflicted"
}
```

**Step 2: Verify JSON is valid**

Run:
```bash
jq empty .github/renovate.json && echo "Valid JSON" || echo "Invalid JSON"
```

Expected: Output shows "Valid JSON"

**Step 3: Add documentation comment**

Add comment at top of `.github/renovate.json`:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "_comment": "Renovate Bot Configuration - Automates dependency updates for Go and npm",
  "extends": [
    "config:recommended"
  ],
  ...
}
```

---

## Task 6: Commit Renovate Configuration

**Files:**
- Create: `.github/renovate.json`

**Step 1: Stage Renovate config**

Run:
```bash
git add .github/renovate.json
```

Expected: File staged for commit

**Step 2: Commit Renovate configuration**

Run:
```bash
git commit -m "chore: configure Renovate bot for automated dependency updates

Add Renovate configuration with:
- Weekly schedule (Mondays before 6am)
- Auto-merge for Go patch/minor updates
- Auto-merge for npm patch updates  
- Grouped major updates for easier review
- Priority handling for security vulnerabilities
- Rate limiting (3 concurrent, 2/hour) to avoid PR spam

Benefits:
- Automated security patches
- Reduced maintenance burden
- Consistent dependency updates
- Better than Dependabot for monorepo (Go + npm)"
```

Expected: Commit created successfully

---

## Task 7: Enable Renovate in Repository

**Files:**
- External: GitHub repository settings

**Step 1: Push commits to GitHub**

Run:
```bash
git push origin main
```

Expected: Both commits pushed successfully

**Step 2: Install Renovate GitHub App**

1. Go to: https://github.com/apps/renovate
2. Click "Configure"
3. Select "castrojo/bluefin-releases" repository
4. Grant access permissions (read/write for code, PRs, issues)
5. Click "Install"

Expected: Renovate app installed and has repository access

**Step 3: Wait for onboarding PR**

Within 5-10 minutes, Renovate should:
1. Detect `.github/renovate.json`
2. Validate configuration
3. Create "Configure Renovate" onboarding PR

Expected: PR appears in repository with title "Configure Renovate"

**Step 4: Review onboarding PR**

The onboarding PR will show:
- Detected dependencies (Go modules, npm packages)
- Proposed update schedule
- Configuration validation results

Review and approve if configuration looks correct.

**Step 5: Merge onboarding PR**

Run (or use GitHub UI):
```bash
gh pr merge --auto --squash
```

Expected: Onboarding PR merged, Renovate starts monitoring

---

## Task 8: Validation and Documentation

**Files:**
- Modify: `README.md` (add Renovate badge)
- Create: `docs/MAINTENANCE.md`

**Step 1: Verify Dependabot alerts resolved**

Run:
```bash
gh api repos/castrojo/bluefin-releases/dependabot/alerts --jq '[.[] | select(.state == "open")] | length'
```

Expected: Output shows "0" (all alerts resolved)

**Step 2: Verify Renovate is active**

Run:
```bash
gh api repos/castrojo/bluefin-releases/installation --jq '.[] | select(.app_slug == "renovate")'
```

Expected: Shows Renovate installation details

**Step 3: Add Renovate badge to README**

Add to `README.md` after existing badges:

```markdown
[![Renovate enabled](https://img.shields.io/badge/renovate-enabled-brightgreen.svg)](https://renovatebot.com/)
```

**Step 4: Create maintenance documentation**

Create `docs/MAINTENANCE.md`:

```markdown
# Maintenance Guide

## Automated Dependency Updates

This project uses **Renovate Bot** for automated dependency updates.

### How It Works

- **Schedule:** Updates run weekly (Mondays before 6am UTC)
- **Auto-merge:** Patch and minor updates merge automatically if CI passes
- **Manual Review:** Major updates require manual review and approval
- **Security:** Vulnerability fixes get immediate PRs (no schedule wait)

### Renovate PR Types

1. **Patch Updates** (auto-merged)
   - Example: `v1.2.3` → `v1.2.4`
   - Go and npm patch updates merge automatically

2. **Minor Updates** (grouped, auto-merged for Go)
   - Example: `v1.2.0` → `v1.3.0`
   - Go: Auto-merged if CI passes
   - npm: Grouped PR, requires approval

3. **Major Updates** (grouped, manual review)
   - Example: `v1.0.0` → `v2.0.0`
   - Always requires manual review
   - May contain breaking changes

4. **Security Updates** (immediate, high priority)
   - CVE fixes get immediate PRs regardless of schedule
   - Should be reviewed and merged ASAP

### Configuration

Renovate config: `.github/renovate.json`

To adjust behavior:
- Change schedule: Modify `schedule` field
- Change auto-merge rules: Modify `packageRules[].automerge`
- Change grouping: Modify `packageRules[].groupName`

### Troubleshooting

**Renovate not creating PRs:**
- Check: https://github.com/castrojo/bluefin-releases/settings/installations
- Verify Renovate app has access
- Check Renovate logs: https://app.renovatebot.com/dashboard

**Too many PRs:**
- Increase grouping in `.github/renovate.json`
- Adjust `prConcurrentLimit` and `prHourlyLimit`

**Auto-merge not working:**
- Verify GitHub Actions CI is passing
- Check branch protection rules don't block auto-merge
- Review Renovate logs for errors
```

**Step 5: Commit documentation**

Run:
```bash
git add README.md docs/MAINTENANCE.md
git commit -m "docs: add Renovate badge and maintenance guide

Add Renovate enabled badge to README and create maintenance
documentation explaining automated dependency update workflow."
```

Expected: Commit created successfully

**Step 6: Push final commits**

Run:
```bash
git push origin main
```

Expected: All commits pushed successfully

---

## Task 9: Final Verification

**Step 1: Verify all Dependabot alerts resolved**

Visit: https://github.com/castrojo/bluefin-releases/security/dependabot

Expected: Shows "0 open alerts" or "No open security advisories"

**Step 2: Verify Renovate onboarding succeeded**

Visit: https://github.com/castrojo/bluefin-releases/pulls

Expected: Renovate onboarding PR is merged

**Step 3: Check for first Renovate dependency PRs**

After first scheduled run (next Monday), check for:
- Grouped dependency update PRs
- Auto-merged patch updates in commit history

**Step 4: Print completion summary**

Output summary:
```
✅ Dependency Security Updates Complete

Security Fixes:
- Updated golang.org/x/net: v0.4.0 → v0.33.0+
- Updated golang.org/x/oauth2 to latest
- Resolved all 7 Dependabot alerts (2 high, 5 medium)

Automation:
- ✅ Renovate Bot configured and enabled
- ✅ Auto-merge for patch/minor updates
- ✅ Weekly update schedule (Mondays)
- ✅ Priority handling for security vulnerabilities

Verification:
- ✅ Pipeline runs successfully
- ✅ Frontend builds without errors
- ✅ All dependencies updated cleanly
- ✅ Documentation updated

Next Steps:
- Monitor first Renovate PR (next Monday)
- Review and merge any major update PRs
- Adjust Renovate config if needed

GitHub Actions: https://github.com/castrojo/bluefin-releases/actions
Dependabot: https://github.com/castrojo/bluefin-releases/security/dependabot
Renovate Dashboard: https://app.renovatebot.com/dashboard
```

---

## Success Criteria Checklist

Before considering this plan complete, verify:

- [ ] `go build ./...` succeeds without errors
- [ ] Pipeline runs and generates valid apps.json (130 packages)
- [ ] Frontend builds successfully (`npm run build`)
- [ ] All 7 Dependabot alerts are resolved
- [ ] `.github/renovate.json` is committed
- [ ] Renovate bot is installed on repository
- [ ] Renovate onboarding PR is merged
- [ ] README.md has Renovate badge
- [ ] `docs/MAINTENANCE.md` is created with workflow documentation
- [ ] All commits are pushed to main
- [ ] GitHub Actions deployment succeeds

## Rollback Procedure

If issues arise after deployment:

1. **Revert dependency updates:**
   ```bash
   git revert <commit-hash>  # Revert Task 4 commit
   git push origin main
   ```

2. **Disable Renovate (if causing issues):**
   - Go to: https://github.com/castrojo/bluefin-releases/settings/installations
   - Find Renovate, click "Configure"
   - Click "Suspend" to temporarily disable

3. **Investigate and fix:**
   - Review specific errors
   - Check if breaking changes in dependency updates
   - Pin problematic dependencies in go.mod
   - Re-enable Renovate after fix
