# RSS Library Proof of Concept

This directory contains a proof-of-concept implementation for evaluating Go RSS libraries for the Bluefin Releases project.

## Quick Start

```bash
# Build the POC
go build -o rss-poc cmd/rss-poc/main.go

# Run the POC
./rss-poc

# View results
cat rss-poc-output.json
```

## What It Does

The POC demonstrates:

1. **Fetches GitHub releases via RSS** (Atom feed from ublue-os/bluefin)
2. **Fetches same releases via GitHub API** (current approach)
3. **Compares performance** (latency, memory usage)
4. **Tests error handling** (invalid feeds)
5. **Exports sample data** to JSON

## Files

```
cmd/rss-poc/main.go       # Benchmark and demo application
internal/rss/rss.go       # RSS parser wrapper using gofeed
rss-poc-output.json       # Generated output (sample data)
docs/go-rss-library-evaluation.md  # Full evaluation document
```

## Architecture

### Parser (`internal/rss/rss.go`)

```go
// Create parser with timeout
parser := rss.NewParser(30 * time.Second)

// Fetch any RSS/Atom feed
feed, err := parser.FetchAndParse(ctx, "https://example.com/feed.xml")

// Convert to models.Release structs
releases := rss.ConvertToReleases(feed, "github-release")

// Convenience function for GitHub
releases, err := parser.FetchGitHubReleases(ctx, "owner", "repo")
```

### Benchmark Results

Based on 5 runs with ublue-os/bluefin (10 releases):

```
RSS (gofeed):     215ms ± 2ms
GitHub REST API:  102ms ± 1ms

Result: API is 2.1x faster than RSS
Memory: ~16MB RSS (for both RSS + API fetch)
```

## Integration Examples

### Option 1: RSS as Fallback

```go
// Try API first, fall back to RSS
releases, err := github.FetchReleasesAPI(owner, repo)
if err != nil {
    log.Printf("API failed, trying RSS: %v", err)
    releases, err = rss.FetchGitHubReleases(ctx, owner, repo)
}
```

### Option 2: RSS as Primary

```go
// Use RSS for sources with good feeds
parser := rss.NewParser(30 * time.Second)
releases, err := parser.FetchGitHubReleases(ctx, owner, repo)
```

### Option 3: Parallel Fetch (Best of Both)

```go
// Fetch both, use whichever succeeds first
var releases []models.Release
errg, ctx := errgroup.WithContext(context.Background())

// RSS fetch
errg.Go(func() error {
    r, err := rss.FetchGitHubReleases(ctx, owner, repo)
    if err == nil {
        releases = r
    }
    return err
})

// API fetch
errg.Go(func() error {
    r, err := github.FetchReleasesAPI(ctx, owner, repo)
    if err == nil {
        releases = r
    }
    return err
})

errg.Wait()
```

## Known Limitations

1. **Version Extraction**: GitHub RSS uses GUID format like:
   ```
   tag:github.com,2008:Repository/611397346/stable-20260203
   ```
   Current implementation extracts this as-is. May need refinement to extract just the version tag.

2. **Performance**: RSS is ~2x slower than GitHub API (215ms vs 102ms). Acceptable for 6-hour refresh cycle, but not if sub-100ms is required.

3. **Feed Availability**: Not all sources provide RSS feeds:
   - ✅ GitHub: Atom feeds for releases, tags, commits
   - ❌ Flathub: No RSS feed (uses REST API only)
   - ❌ Homebrew: No RSS feed (uses formulae API)
   - ✅ Mozilla: RSS feeds available for Firefox, Thunderbird

## Dependencies Added

```
github.com/mmcdole/gofeed v1.3.0
  ├── github.com/PuerkitoBio/goquery v1.8.0
  ├── github.com/json-iterator/go v1.1.12
  └── github.com/mmcdole/goxpp v1.1.1
```

Total: ~6 transitive dependencies, all well-maintained.

## Testing

```bash
# Run POC
go run cmd/rss-poc/main.go

# Run multiple times for benchmark
for i in {1..10}; do go run cmd/rss-poc/main.go | grep "Performance"; done

# Memory profiling
/usr/bin/time -v ./rss-poc

# Compare with current pipeline
go run cmd/bluefin-releases/main.go  # Current approach
./rss-poc                             # RSS approach
```

## Evaluation Summary

**Recommendation:** Use `mmcdole/gofeed` if RSS parsing is needed.

**When to use RSS:**
- ✅ Source has comprehensive RSS feeds
- ✅ Want to reduce API rate limits
- ✅ No authentication available/needed
- ✅ ~100ms extra latency acceptable

**When to use API:**
- ✅ Need rich metadata not in RSS
- ✅ Performance critical (<100ms)
- ✅ Source provides structured JSON
- ✅ Current approach already works well

For full evaluation details, see: `docs/go-rss-library-evaluation.md`

---

**Next Steps:** Decide whether to integrate RSS support into main pipeline or keep as reference implementation for future use.
