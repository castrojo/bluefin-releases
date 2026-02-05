# Bluefin Releases - Testing Plan

This document provides a comprehensive testing checklist to ensure refactoring work doesn't break functionality.

## Pre-Refactoring Baseline

**Established:** February 5, 2026

### Build Metrics
- **Build command:** `npm run build`
- **Build status:** ✅ Success
- **Build time:** ~7 seconds (pipeline: 5.9s, Astro: 0.7s)
- **Build output size:** 608KB
- **Total packages:** 89 (42 Flatpak + 44 Homebrew + 3 OS)
- **Total releases:** 127

### Code Metrics (Before Refactoring)
- `src/pages/index.astro`: 857 lines
- Total inline CSS: ~1,200 lines across components
- Total inline JavaScript: ~600 lines across components
- Total codebase: ~3,458 lines

---

## Testing Checklist

### 1. Visual Baseline (Screenshot Testing)

Take screenshots of key pages/states BEFORE refactoring:
- [ ] Homepage (default view)
- [ ] Homepage with filters applied (OS/Flatpak/Homebrew)
- [ ] Homepage with search active
- [ ] Light theme
- [ ] Dark theme
- [ ] Mobile viewport (375px)
- [ ] Tablet viewport (768px)
- [ ] Desktop viewport (1920px)

**Storage:** `tests/screenshots/baseline/`

**Note:** Manual screenshots recommended. Store in baseline/ directory with descriptive names like:
- `homepage-default-light-desktop.png`
- `homepage-filter-flatpak-dark-mobile.png`

---

### 2. Functional Testing Checklist

#### Theme System
- [ ] Theme persists on page reload
- [ ] Theme toggle button works
- [ ] No FOUC (Flash of Unstyled Content)
- [ ] System theme detection works on first visit

#### Filter System
- [ ] 'All' filter shows all items
- [ ] 'OS Releases' filter shows only OS items
- [ ] 'Flatpak Apps' filter shows only Flatpak items
- [ ] 'Homebrew Packages' filter shows only Homebrew items
- [ ] Active filter has correct styling
- [ ] Filter count badges are accurate
- [ ] URL updates with filter parameter (?filter=flatpak)
- [ ] Direct URL with filter parameter works

#### Search System
- [ ] Search input filters results in real-time
- [ ] Search is case-insensitive
- [ ] Search matches app names, descriptions, categories
- [ ] Clearing search restores all filtered items
- [ ] Search works with filters combined
- [ ] Search icon/clear button appears/disappears correctly

#### Keyboard Navigation
- [ ] Tab key cycles through focusable elements
- [ ] Arrow keys navigate between app cards
- [ ] Enter key on card opens details
- [ ] Escape key clears focus/search
- [ ] Keyboard shortcuts documented in UI work

#### App Cards
- [ ] All app metadata displays correctly (name, version, date, description)
- [ ] App icons load correctly (with fallback)
- [ ] External links open in new tabs
- [ ] GitHub badges/links work
- [ ] Flathub badges/links work (for Flatpaks)
- [ ] Release notes expand/collapse
- [ ] Hover states work correctly

#### Recent Releases Section
- [ ] Shows 10 most recent releases
- [ ] Sorted by date (newest first)
- [ ] Release cards display correctly
- [ ] Links work correctly

#### Featured App Banner
- [ ] Banner displays on homepage
- [ ] Featured app data loads correctly
- [ ] Banner styling correct in light/dark themes

#### RSS Feed
- [ ] RSS feed generates correctly (/rss.xml)
- [ ] Feed contains all recent releases
- [ ] Feed validates (https://validator.w3.org/feed/)

#### Performance
- [ ] Page load time < 2s (production build)
- [ ] No JavaScript errors in console
- [ ] No CSS rendering issues
- [ ] Images load and display correctly

#### Responsive Design
- [ ] Mobile (375px): Single column, filters stack, search usable
- [ ] Tablet (768px): Two columns, filters horizontal
- [ ] Desktop (1920px): Three columns, full layout
- [ ] No horizontal scrolling on any viewport

---

### 3. Build & Deployment Testing

- [ ] `npm run build` succeeds with no errors
- [ ] `npm run build` produces expected output in `dist/`
- [ ] `dist/` size is reasonable (~608KB baseline, or smaller with optimizations)
- [ ] `npm run preview` serves site correctly
- [ ] All assets load from correct paths (`/bluefin-releases/` base)
- [ ] GitHub Pages deployment preview works

---

### 4. Code Quality Checks

- [ ] TypeScript compilation succeeds (`npx tsc --noEmit`)
- [ ] No TypeScript errors in editor
- [ ] No unused imports or variables
- [ ] Astro check passes (`npx astro check`)

---

### 5. Regression Testing After Each Phase

#### After Priority 1 tasks (env.d.ts, SEO, Images)
- [ ] Run full functional checklist
- [ ] Compare screenshots (especially images)
- [ ] Verify no console errors
- [ ] Check build output size

#### After Priority 2 tasks (BaseLayout, CSS extraction, Scripts refactoring)
- [ ] Run full functional checklist
- [ ] Compare screenshots (should be identical)
- [ ] Verify all interactive features work
- [ ] Performance check (should be same or better)

---

### 6. Testing Workflow

#### Before Starting Refactoring
1. Document current state:
   - [x] Take baseline screenshots
   - [x] Record page load time
   - [x] Record build output size
   - [ ] Run functional checklist (all should pass)
2. [ ] Commit baseline documentation

#### After Each Task
1. `npm run build` (must succeed)
2. `npm run preview`
3. Run relevant subset of functional tests
4. Visual comparison
5. Check for console errors
6. If any test fails: FIX before marking complete

#### After All Refactoring Complete
1. Run FULL functional checklist
2. Screenshot comparison (all viewports, both themes)
3. Performance comparison
4. Build output comparison
5. Deploy to preview environment
6. Final sign-off

---

### 7. Acceptance Criteria

**BEFORE closing bluefin-releases-sc8 epic:**
- [ ] ALL functional tests pass
- [ ] Visual regression = ZERO (screenshots identical or intentionally improved)
- [ ] No new console errors
- [ ] Build succeeds
- [ ] Performance same or better
- [ ] All 6 sub-issues closed with verification

---

## Quick Verification Commands

```bash
# Build the site
npm run build

# Preview the site
npm run preview

# Check TypeScript
npx tsc --noEmit

# Run Astro check
npx astro check

# Check build output size
du -sh dist/

# Check for console errors (open browser DevTools after preview)
```

---

## Notes

- **Zero visual regression:** The goal is for the site to look IDENTICAL after refactoring
- **All functionality preserved:** Every interactive feature must work exactly as before
- **Performance maintained or improved:** Build times and output size should not increase
- **Code quality improved:** TypeScript strict mode, better organization, maintainability

---

## Baseline Results

### Initial Build Test (Feb 5, 2026)
- ✅ Build succeeded
- ✅ Pipeline completed in 5.9s
- ✅ Astro build completed in 0.7s
- ✅ Output size: 608KB
- ✅ No build errors
- ✅ RSS feeds generated correctly

**Ready for refactoring work to begin.**
