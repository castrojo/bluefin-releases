# Dependency Security Updates - Design Document

**Date:** 2026-02-08  
**Status:** Ready for Implementation

## Problem Statement

GitHub Dependabot has identified **7 security vulnerabilities** (2 high, 5 medium severity) in Go dependencies:

- **golang.org/x/net v0.4.0** - 6 vulnerabilities (HTTP/2 DoS, HPACK decoder, XSS, etc.)
- **golang.org/x/oauth2 v0.27.0** - 1 vulnerability (input validation)

**Impact Analysis:**
While the high-severity alerts (#6, #7) affect HTTP/2 **server-side** code and this project is a **CLI tool** (HTTP client only), updating is still warranted for:
- Security hygiene and best practices
- Removing GitHub security warnings
- Future-proofing if architecture changes
- Getting bug fixes and performance improvements

## Goals

1. **Immediate:** Fix all 7 Dependabot security alerts via surgical dependency updates
2. **Long-term:** Automate future dependency updates using Renovate Bot

## Solution Architecture

### Phase 1: Surgical Dependency Update

**Approach:** Update only `golang.org/x/net` and `golang.org/x/oauth2` to latest versions

**Why surgical vs. full update:**
- Lower risk of breaking changes
- Directly addresses security alerts
- Faster to test and validate
- Can update other deps separately later

**Update Strategy:**
```bash
go get -u golang.org/x/net@latest
go get -u golang.org/x/oauth2@latest
go mod tidy
```

This updates the transitive dependencies to versions that fix the vulnerabilities. Go's module system ensures compatibility constraints are respected.

**Verification:**
- Go build succeeds (`go build ./...`)
- Pipeline runs successfully
- Frontend builds (`npm run build`)
- No functional regressions

### Phase 2: Renovate Bot Configuration

**Why Renovate over Dependabot:**
- More flexible configuration (grouping, scheduling, automerge)
- Better monorepo support (Go + npm in one repo)
- Smarter update strategies (respect semantic versioning)
- Can auto-merge minor/patch updates
- Single bot for both Go and npm dependencies

**Configuration Strategy:**
Create `.github/renovate.json` with:

1. **Go Dependencies:**
   - Group minor/patch updates together
   - Separate PRs for major updates
   - Auto-merge patch updates after CI passes
   - Weekly schedule to avoid PR spam

2. **npm Dependencies:**
   - Group by update type (devDependencies vs dependencies)
   - Auto-merge patch updates
   - Require approval for major updates

3. **Security Updates:**
   - Separate PR for security vulnerabilities
   - Higher priority, immediate scheduling
   - Pin SHA in PR description for audit trail

**Renovate Features Used:**
- **Grouping:** Reduce PR noise by batching related updates
- **Scheduling:** `schedule: ["before 6am on Monday"]` - updates once per week
- **Auto-merge:** Patch/minor updates auto-merge if CI passes
- **Vulnerability alerts:** Security updates get immediate PRs

## Implementation Steps

### Task 1: Update Go Dependencies
1. Update `golang.org/x/net` to latest
2. Update `golang.org/x/oauth2` to latest
3. Run `go mod tidy`
4. Verify build succeeds
5. Run pipeline test
6. Commit with security context

### Task 2: Configure Renovate Bot
1. Create `.github/renovate.json`
2. Configure Go module updates
3. Configure npm updates
4. Set auto-merge rules
5. Commit configuration

### Task 3: Enable Renovate in Repository
1. Install Renovate GitHub App
2. Grant repository access
3. Verify onboarding PR appears
4. Review and merge onboarding PR
5. Verify Renovate starts monitoring

### Task 4: Validation
1. Check Dependabot alerts are resolved
2. Verify Renovate onboarding succeeded
3. Wait for first scheduled Renovate run
4. Document maintenance workflow

## Testing Strategy

**Before Update:**
- Capture current dependency versions (`go list -m all > deps-before.txt`)
- Verify pipeline works (`go run cmd/bluefin-releases/main.go`)
- Verify frontend builds (`npm run build`)

**After Update:**
- Compare dependency changes (`go list -m all > deps-after.txt && diff deps-before.txt deps-after.txt`)
- Verify pipeline still works
- Verify frontend still builds
- Check for any deprecation warnings

**Regression Testing:**
- Pipeline fetches all 130 packages successfully
- Frontend renders all package types
- RSS feeds generate correctly
- No new build errors or warnings

## Success Criteria

✅ All 7 Dependabot alerts are resolved  
✅ `go build ./...` succeeds without errors  
✅ Pipeline runs successfully and generates valid `apps.json`  
✅ Frontend builds and deploys without issues  
✅ Renovate bot is configured and creating PRs  
✅ Documentation updated with maintenance workflow  

## Rollback Plan

If updates cause issues:
1. `git revert <commit-hash>` - revert the dependency update commit
2. `git push` - restore previous working state
3. Investigate specific breaking changes
4. Pin problematic dependency to previous version in `go.mod`

## Future Considerations

- Consider setting up `govulncheck` in CI to catch vulnerabilities earlier
- Evaluate enabling Renovate's "lockfile maintenance" for npm
- Monitor Renovate PR volume and adjust grouping/scheduling if needed

## References

- Dependabot Alerts: https://github.com/castrojo/bluefin-releases/security/dependabot
- Renovate Docs: https://docs.renovatebot.com/
- Go Modules: https://go.dev/ref/mod
- golang.org/x/net CVE fixes: Patched in v0.17.0+
