# Maintenance Guide

This document explains the automated dependency update workflow for Bluefin Firehose.

## Renovate Bot

Bluefin Firehose uses [Renovate Bot](https://renovatebot.com/) for automated dependency updates. Renovate runs on a schedule and creates pull requests when new versions of dependencies are available.

### How It Works

**Schedule**: Renovate runs weekly on Mondays between 2-4 AM UTC.

**What It Updates**:
- Go dependencies (`go.mod`)
- npm dependencies (`package.json`)
- GitHub Actions workflows (`.github/workflows/*.yml`)

**PR Behavior**:
- **Patch updates** (1.2.3 → 1.2.4): Auto-merged after CI passes
- **Minor updates** (1.2.0 → 1.3.0): Auto-merged after CI passes
- **Major updates** (1.0.0 → 2.0.0): Requires manual review and merge
- **Security updates**: Created immediately (ignores schedule) and auto-merged if patch/minor

### Auto-Merge Rules

Renovate automatically merges PRs when ALL conditions are met:
1. Update is patch or minor version (not major)
2. All CI checks pass (build, tests, pipeline)
3. Dependency is not pinned or frozen
4. No merge conflicts

Major version updates require manual review to assess breaking changes.

### PR Grouping

Related updates are grouped to reduce PR noise:
- **Go dependencies**: All non-major Go updates in one PR
- **npm dependencies**: All non-major npm updates in one PR
- **Security patches**: Grouped by severity (high, medium, low)

### Configuration

Renovate configuration is located at **`.github/renovate.json`**.

Key settings:
```json
{
  "schedule": ["before 4am on Monday"],
  "automerge": true,
  "automergeType": "pr",
  "vulnerabilityAlerts": {
    "enabled": true,
    "schedule": ["at any time"]
  }
}
```

### Monitoring

**Check Renovate Status**:
- Visit the [Dependency Dashboard](../../issues) issue created by Renovate
- Shows pending updates, rate limits, and errors

**Review Merged PRs**:
- Check [Closed Pull Requests](../../pulls?q=is%3Apr+is%3Aclosed+author%3Aapp%2Frenovate)
- Filter by `author:app/renovate` to see Renovate's work

**Manual Trigger**:
- Add a comment to the Dependency Dashboard issue: `@renovatebot rebase`
- Forces Renovate to check for updates immediately

## Dependabot Alerts

GitHub's Dependabot scans for security vulnerabilities and creates alerts.

**Where to Find Alerts**:
- Repository → Security tab → Dependabot alerts
- Or visit: https://github.com/castrojo/bluefin-releases/security/dependabot

**Resolution Workflow**:
1. Dependabot detects vulnerable dependency
2. Renovate creates PR with security fix (usually within 1 hour)
3. PR auto-merges if patch/minor update
4. Dependabot alert auto-closes after merge
5. Manual intervention only needed for major updates

**Alert Resolution Time**:
- Patch/minor fixes: ~1-2 hours (Renovate + CI + auto-merge)
- Major updates: Requires manual review (no SLA)

## Manual Dependency Updates

If you need to update dependencies manually (not recommended - let Renovate handle it):

### Go Dependencies

```bash
# Update specific dependency
go get github.com/example/package@latest
go mod tidy

# Update all dependencies
go get -u ./...
go mod tidy

# Verify
go run cmd/bluefin-releases/main.go
```

### npm Dependencies

```bash
# Update specific dependency
npm update package-name

# Update all dependencies
npm update

# Verify
npm run build
npm run preview
```

### After Manual Updates

1. Run full build: `npm run build`
2. Test locally: `npm run preview`
3. Commit with detailed message explaining why manual update was needed
4. Push and verify CI passes

## Troubleshooting

### Renovate Not Creating PRs

**Check Rate Limits**:
- Renovate may hit GitHub API rate limits
- View status in Dependency Dashboard issue
- Wait for rate limit reset (shown in dashboard)

**Check Configuration**:
```bash
# Validate renovate.json syntax
npx --yes renovate-config-validator .github/renovate.json
```

**Force Refresh**:
- Comment on Dependency Dashboard: `@renovatebot rebase`

### Auto-Merge Not Working

**Common Causes**:
1. CI checks failed → Fix test/build errors
2. Merge conflicts → Rebase PR manually
3. Major version update → Requires manual review (by design)
4. Branch protection rules → Check repo settings

**Debug**:
- Check PR checks tab for failed CI
- Review Renovate's PR description for warnings
- Check repo settings → Branches → main protection rules

### Dependabot Alerts Not Resolving

**Timeline**:
- Renovate typically creates fix PR within 1 hour of alert
- Auto-merge happens within 15 minutes of CI passing
- Dependabot alert closes within 1-2 hours of merge

**If Stuck**:
1. Check if Renovate created PR (search for "security" label)
2. If no PR, manually update: `go get package@fixed-version`
3. If PR exists but not merged, check CI failures
4. If PR merged but alert open, wait 2-4 hours for GitHub sync

### Build Failures After Update

**Immediate Actions**:
1. Check CI logs for specific error
2. Review PR diff for breaking changes
3. Check dependency's changelog for migration notes

**Resolution**:
1. If breaking change, revert PR and pin old version temporarily
2. Create issue to track migration work
3. Update code to handle breaking changes
4. Re-run update

## Best Practices

**DO**:
- ✅ Let Renovate handle routine updates automatically
- ✅ Review major version update PRs carefully
- ✅ Keep CI checks fast (auto-merge depends on them)
- ✅ Monitor Dependency Dashboard weekly

**DON'T**:
- ❌ Manually update dependencies that Renovate tracks
- ❌ Disable Renovate without good reason
- ❌ Ignore major version update PRs for months
- ❌ Merge failing CI checks just to unblock Renovate

## Resources

- [Renovate Documentation](https://docs.renovatebot.com/)
- [GitHub Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [Project AGENTS.md](../AGENTS.md) - Development workflow guide
- [Dependency Baseline](deps-before.txt) - Pre-update snapshot (2025-02-08)
- [Dependency After](deps-after.txt) - Post-update snapshot (2025-02-08)
