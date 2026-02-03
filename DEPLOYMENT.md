# Deployment Configuration

## Overview

This project is deployed to GitHub Pages using GitHub Actions with automated builds every 6 hours.

**Live Site**: https://castrojo.github.io/bluefin-releases/

## Automated Deployment Schedule

The site automatically rebuilds and deploys:

- **Every 6 hours**: 0:00, 6:00, 12:00, 18:00 UTC (via cron schedule)
- **On push to main**: Immediate deployment on code changes
- **Manual trigger**: Via workflow_dispatch in GitHub Actions UI

### Cron Schedule Details

```yaml
schedule:
  - cron: '0 */6 * * *'
```

This runs at:
- **00:00 UTC** (12:00 AM)
- **06:00 UTC** (6:00 AM)
- **12:00 UTC** (12:00 PM)
- **18:00 UTC** (6:00 PM)

## GitHub Pages Configuration

### Current Settings

- **Source**: GitHub Actions (build_type: workflow)
- **Branch**: main (source branch)
- **Public**: Yes
- **HTTPS**: Enforced
- **Custom 404**: Not configured
- **Custom Domain**: Not configured

### How to Verify Configuration

```bash
# Check Pages status
gh api repos/castrojo/bluefin-releases/pages

# View recent deployments
gh run list --repo castrojo/bluefin-releases --limit 5

# Check site accessibility
curl -I https://castrojo.github.io/bluefin-releases/
```

## Workflow Architecture

### Build Process

The `.github/workflows/deploy.yml` workflow consists of two jobs:

#### 1. Build Job
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - Checkout repository
      - Setup Go 1.21
      - Setup Node.js 20
      - Install npm dependencies
      - Download Go dependencies
      - Run Go pipeline (fetch Flathub/GitHub data)
      - Build Astro site
      - Upload artifact for deployment
```

#### 2. Deploy Job
```yaml
jobs:
  deploy:
    needs: build
    environment: github-pages
    steps:
      - Deploy to GitHub Pages
```

### Permissions

The workflow requires these permissions:
- `contents: read` - Read repository contents
- `pages: write` - Write to GitHub Pages
- `id-token: write` - Generate deployment token

### Concurrency Control

```yaml
concurrency:
  group: "pages"
  cancel-in-progress: false
```

This ensures only one deployment runs at a time and prevents race conditions.

## Environment Variables

### Automatic Variables

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
  - Used for GitHub API authentication
  - Prevents rate limiting
  - No manual configuration needed

### Optional Variables

Currently, no additional secrets or environment variables are required.

## Testing the Deployment

### Manual Workflow Trigger

1. Go to: https://github.com/castrojo/bluefin-releases/actions
2. Select "Build and Deploy" workflow
3. Click "Run workflow" button
4. Select branch: main
5. Click "Run workflow"

### Verify Deployment

```bash
# Check latest workflow run
gh run list --repo castrojo/bluefin-releases --limit 1

# View workflow logs
gh run view <run-id> --log

# Test site accessibility
curl -s https://castrojo.github.io/bluefin-releases/ | grep -o '<title>.*</title>'
```

### Expected Results

- **Build time**: ~1-2 minutes total
  - Go pipeline: ~15-30s (depends on API response times)
  - Astro build: ~1-2s
  - Upload/Deploy: ~30-45s
- **Status**: Green checkmark in Actions tab
- **Site**: Accessible at https://castrojo.github.io/bluefin-releases/
- **HTTP Status**: 200 OK
- **Content**: Updated Flathub app data

## Monitoring

### Checking Scheduled Runs

```bash
# View recent workflow runs
gh run list --repo castrojo/bluefin-releases --workflow="Build and Deploy"

# Check for failures
gh run list --repo castrojo/bluefin-releases --status failure
```

### Troubleshooting Failed Deployments

1. **Check workflow logs**:
   ```bash
   gh run view <run-id> --log
   ```

2. **Common issues**:
   - API rate limiting (GitHub/Flathub)
   - Network timeouts during data fetch
   - Build failures in Go or Astro

3. **Manual recovery**:
   - Trigger manual workflow run
   - Check API status (GitHub, Flathub)
   - Review recent commits for breaking changes

## Performance Metrics

### Build Statistics

- **Total build time**: 55-60 seconds average
- **Go pipeline**: 15-30 seconds
  - Flathub API fetch: 1-2s
  - Details enrichment: 5-10s
  - GitHub releases fetch: 10-20s
- **Astro build**: 1-2 seconds
- **Upload/Deploy**: 30-45 seconds

### Deployment Frequency

- **Automated**: 4 times per day (every 6 hours)
- **Manual**: On-demand via workflow_dispatch
- **Code changes**: Immediate on push to main

## Configuration Management

### Modifying Schedule

To change the cron schedule, edit `.github/workflows/deploy.yml`:

```yaml
schedule:
  - cron: '0 */6 * * *'  # Current: Every 6 hours
```

Examples:
- Every hour: `'0 * * * *'`
- Every 12 hours: `'0 */12 * * *'`
- Daily at 6 AM: `'0 6 * * *'`
- Twice daily: `'0 6,18 * * *'`

### Modifying Build Process

Key files to modify:
- **Workflow**: `.github/workflows/deploy.yml`
- **Go pipeline**: `cmd/bluefin-releases/main.go`
- **Astro config**: `astro.config.mjs`

## GitHub Pages Setup (Reference)

### Initial Setup Steps

1. **Enable GitHub Pages**:
   - Go to: Settings → Pages
   - Source: GitHub Actions
   - Save

2. **Verify Permissions**:
   - Settings → Actions → General
   - Workflow permissions: Read and write permissions
   - Allow GitHub Actions to create and approve pull requests: ✓

3. **First Deployment**:
   - Push to main or trigger workflow manually
   - Check Actions tab for deployment status
   - Visit site URL when deployment completes

### Current Configuration

✅ GitHub Pages is properly configured:
- Build type: `workflow` (uses GitHub Actions)
- Status: Active
- URL: https://castrojo.github.io/bluefin-releases/
- Public: Yes
- HTTPS: Enforced

## Maintenance

### Regular Checks

- Review workflow runs for failures
- Monitor API rate limits
- Check site accessibility
- Verify data freshness (last update timestamp)

### Updates

When updating dependencies:
1. Go modules: `go get -u && go mod tidy`
2. npm packages: `npm update`
3. Test locally: `npm run build && npm run preview`
4. Commit and push changes
5. Monitor deployment

## Security

### API Tokens

- `GITHUB_TOKEN` is automatically provided and scoped per workflow run
- No need to store or manage tokens manually
- Token has minimal permissions required for operation

### Best Practices

- Keep dependencies updated
- Review workflow logs regularly
- Monitor for suspicious activity
- Use workflow concurrency controls

## Support

### Resources

- **GitHub Actions Docs**: https://docs.github.com/en/actions
- **GitHub Pages Docs**: https://docs.github.com/en/pages
- **Cron Expression Guide**: https://crontab.guru

### Contact

For issues or questions:
- Open an issue in the repository
- Check existing workflow runs for patterns
- Review deployment logs
