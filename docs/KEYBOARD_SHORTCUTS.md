# Keyboard Shortcuts Implementation Guide

## Overview

This document explains how keyboard shortcuts are implemented in the Bluefin Releases dashboard. Understanding this architecture will help you maintain, extend, and debug keyboard navigation features.

## Architecture

Keyboard shortcuts are implemented using a **functional programming approach** with clear separation between pure logic and side effects.

### Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `src/scripts/keyboard-nav.ts` | Main keyboard navigation logic | 267 |
| `src/layouts/BaseLayout.astro` | Initialization and wiring | ~400 |
| `src/components/KeyboardHelp.astro` | Help modal UI | ~200 |
| `e2e/keyboard-shortcuts.spec.ts` | E2E tests (Playwright) | ~350 |
| `src/scripts/keyboard-nav.test.ts` | Unit tests (Jest) | ~130 |

### Design Principles

1. **Pure functions** for logic (testable, no side effects)
2. **Separation of concerns** (logic vs. DOM manipulation)
3. **Context awareness** (disabled when typing in inputs)
4. **Screen reader friendly** (live region announcements)
5. **Smooth scrolling** (automatic scroll-into-view with header offset)

## Keyboard Shortcuts Reference

| Key | Action | Implementation |
|-----|--------|----------------|
| `j` | Next app | `focusCard(state.focusedIndex + 1)` |
| `k` | Previous app | `focusCard(state.focusedIndex - 1)` |
| `o` / `Enter` | Open focused app | Opens Flathub/GitHub/Homebrew link in new tab |
| `/` / `s` | Focus search | `searchInput.focus()` and `searchInput.select()` |
| `?` | Show help | Opens keyboard help modal |
| `Esc` | Context-aware | Closes modal / blurs search / clears focus |
| `t` | Toggle theme | Calls `toggleTheme()` function |
| `Space` | Page down | `window.scrollBy({top: window.innerHeight})` |
| `Shift+Space` | Page up | `window.scrollBy({top: -window.innerHeight})` |
| `h` | Scroll to top | `window.scrollTo({top: 0})` |

## How It Works

### 1. Initialization (`initKeyboardNav`)

Called in `BaseLayout.astro` on `DOMContentLoaded`:

```typescript
initKeyboardNav('.release-card', '.search-bar input');
```

**What it does:**
- Takes CSS selector for navigable items (cards)
- Queries DOM for all matching elements
- Stores state: `{ focusedIndex: -1, items: Element[] }`
- Attaches global `keydown` event listener
- Logs initialization count for debugging

**State management:**
```typescript
let state = {
  focusedIndex: -1,        // Currently focused card (-1 = none)
  items: Element[]         // Array of all navigable cards
};
```

### 2. Event Handling (`handleKeyPress`)

On every keypress, the handler:

1. **Checks context**: Is user typing in an input field?
   ```typescript
   if (isTypingContext(event.target)) return; // Skip shortcuts
   ```

2. **Maps key to action**: Uses switch statement
   ```typescript
   switch (event.key) {
     case 'j': // Next
     case 'k': // Previous
     case '/': // Search
     // ... etc
   }
   ```

3. **Updates state**: Calculates new `focusedIndex`
4. **Calls side-effect function**: `focusCard()`, `openFocused()`, etc.

### 3. Navigation (`focusCard`)

When focusing a card:

```typescript
function focusCard(index: number) {
  // 1. Remove focus from all cards
  clearAllFocus(state.items);
  
  // 2. Add focus class to target card
  state.items[index].classList.add('kbd-focused');
  
  // 3. Scroll card into view (with header offset)
  scrollToItem(state.items[index]);
  
  // 4. Announce to screen reader
  announceToScreenReader(`Focused: ${getCardTitle(state.items[index])}`);
}
```

### 4. Context Awareness (`isTypingContext`)

Shortcuts are **disabled** when user is typing:

```typescript
function isTypingContext(element: EventTarget | null): boolean {
  if (!element || !(element instanceof HTMLElement)) return false;
  
  const tagName = element.tagName.toLowerCase();
  return (
    tagName === 'input' ||
    tagName === 'textarea' ||
    element.isContentEditable ||
    element.closest('.search-bar') !== null
  );
}
```

This prevents `j` from navigating while typing "javascript" in search.

## CSS Selector Requirements

**CRITICAL**: The selector passed to `initKeyboardNav` MUST match actual card elements.

### ✅ Correct Example

```typescript
// Selector matches actual cards in DOM
initKeyboardNav('.grouped-release-card', '.search-bar input');
```

```html
<article class="grouped-release-card release-card">
  <!-- Card content -->
</article>
```

Result: `console.log` shows "Initialized with 40 items" ✅

### ❌ Wrong Example

```typescript
// Selector doesn't match - missing 'grouped-' prefix
initKeyboardNav('.release-card', '.search-bar input');
```

```html
<article class="grouped-release-card">
  <!-- Selector doesn't match! -->
</article>
```

Result: `console.log` shows "Initialized with 0 items" ❌  
Symptoms: j/k keys do nothing

### When Adding New Card Types

You have three options:

1. **Use same CSS class** (recommended)
   ```html
   <article class="release-card">
   <article class="release-card variant-minimal">
   ```

2. **Add multiple classes**
   ```html
   <article class="grouped-release-card release-card">
   ```

3. **Update selector in `BaseLayout.astro`**
   ```typescript
   initKeyboardNav('.grouped-release-card, .release-card, .new-card-type', '.search-bar input');
   ```

## Adding New Shortcuts

Follow these steps:

### 1. Add Key Handler

Edit `src/scripts/keyboard-nav.ts`:

```typescript
function handleKeyPress(event: KeyboardEvent, state, searchInput, updateState) {
  if (isTypingContext(event.target)) return;
  
  switch (event.key) {
    // ... existing cases
    
    case 'n': // NEW: Next page
      event.preventDefault();
      goToNextPage();
      break;
  }
}
```

### 2. Implement Side Effect Function

```typescript
function goToNextPage() {
  // Your implementation
  console.log('[KeyboardNav] Going to next page');
}
```

### 3. Document in Help Modal

Edit `src/components/KeyboardHelp.astro`:

```html
<tr>
  <td class="shortcut-key"><kbd>n</kbd></td>
  <td class="shortcut-description">Next page</td>
</tr>
```

### 4. Add E2E Test

Edit `e2e/keyboard-shortcuts.spec.ts`:

```typescript
test('n key goes to next page', async ({ page }) => {
  const currentUrl = page.url();
  await page.keyboard.press('n');
  
  const newUrl = page.url();
  expect(newUrl).not.toBe(currentUrl);
});
```

### 5. Update PR Template

Edit `.github/PULL_REQUEST_TEMPLATE.md`:

```markdown
- [ ] **`n`** - Goes to next page
```

## Testing Shortcuts

### Unit Tests (Pure Functions)

```bash
npm test                 # Run all Jest tests
npm run test:watch       # Watch mode
npm run test:coverage    # With coverage report
```

Tests pure functions only:
- `isTypingContext()`
- `getNextIndex()`
- `getPrevIndex()`
- `getAnnouncementText()`

Coverage target: **>80%** for pure functions

### E2E Tests (Full Experience)

```bash
npm run test:e2e         # Run all Playwright tests
npm run test:e2e:ui      # Open Playwright UI (interactive)
npm run test:e2e:debug   # Debug mode (step through)
```

Tests all 10 shortcuts in real browser:
- Card navigation (j/k)
- Search focus (/,s)
- Modal (?, Esc)
- Theme toggle (t)
- App opening (o, Enter)
- Scrolling (Space, Shift+Space, h)
- Context awareness

**Requirements:**
- Must run `npm run build` first (generates `dist/`)
- Playwright auto-starts preview server on port 4321
- Tests run against actual built site

### Manual Testing

1. Start dev server:
   ```bash
   npm run dev
   ```

2. Open browser: `http://localhost:4321/bluefin-releases`

3. Open console (F12) and verify:
   ```
   [KeyboardNav] Initialized with 40 items
   ```

4. Test each shortcut manually using PR template checklist

## Common Pitfalls

### 1. Wrong Selector (Most Common)

**Symptom**: j/k keys do nothing

**Diagnosis**:
```javascript
// In browser console
document.querySelectorAll('.grouped-release-card').length
// Should return > 0
```

**Fix**: Update selector in `BaseLayout.astro` to match actual card class

### 2. Timing Issues

**Symptom**: Shortcuts work sometimes, not other times

**Cause**: Cards not rendered when `initKeyboardNav` runs

**Fix**: Use `DOMContentLoaded` or `defer` attribute on script:
```html
<script type="module" defer>
  import { initKeyboardNav } from './scripts/keyboard-nav';
  document.addEventListener('DOMContentLoaded', () => {
    initKeyboardNav('.release-card', '.search-bar input');
  });
</script>
```

### 3. Dynamically Added Cards

**Symptom**: Shortcuts work initially, break after filtering

**Cause**: Filters hide/show cards, but `state.items` is stale

**Fix**: Call refresh after DOM updates:
```javascript
// After filtering
(window as any).keyboardNavRefresh();
```

Current implementation already handles this in `FilterBar.astro`.

### 4. Z-Index Issues

**Symptom**: Focus indicator hidden behind other elements

**Fix**: Ensure `.kbd-focused` has appropriate z-index:
```css
.kbd-focused {
  position: relative;
  z-index: 10;
  outline: 2px solid var(--accent-color);
}
```

### 5. Scroll Not Working

**Symptom**: Card focuses but doesn't scroll into view

**Cause**: Header offset not accounted for

**Fix**: Already handled in `scrollToItem()`:
```typescript
const headerHeight = document.querySelector('.site-header')?.clientHeight || 0;
const targetY = card.getBoundingClientRect().top + window.pageYOffset - headerHeight - 16;
```

## Debugging Guide

### Enable Debug Logging

Already enabled! Check browser console for:
```
[KeyboardNav] Initialized with 40 items
[KeyboardNav] Refreshed, now tracking 38 items
Focused: Firefox 147.0.3
```

### Check Initialization

```javascript
// Browser console
console.log(window.keyboardNavRefresh);
// Should show function, not undefined
```

If `undefined`, script didn't load or init failed.

### Check Card Count

```javascript
// Browser console
document.querySelectorAll('.grouped-release-card').length
// Should match "[KeyboardNav] Initialized with N items"
```

If mismatch, selector is wrong.

### Check Focus Class

```javascript
// After pressing 'j'
document.querySelectorAll('.kbd-focused').length
// Should be 1 (or 0 if nothing focused)
```

### Check Event Listener

```javascript
// Browser console
document.addEventListener('keydown', (e) => {
  console.log('Key pressed:', e.key, 'Target:', e.target);
});
```

Then press `j` - should log "Key pressed: j Target: <body>"

## Performance Considerations

### Initialization Cost

- **DOM Query**: `O(n)` where n = number of cards (~40-100)
- **Event Listener**: Single global listener (efficient)
- **Memory**: Stores references to all cards (~1-2 KB)

### Navigation Cost

- **Focus Update**: `O(n)` to clear all focus classes
- **Scroll**: Native smooth scroll (GPU-accelerated)
- **Screen Reader**: Single `textContent` update

### Optimization Opportunities

1. **Use event delegation** instead of storing card references
2. **Cache header height** instead of querying on every scroll
3. **Debounce rapid key presses** (e.g., holding `j`)

Current implementation is fast enough (no noticeable lag with 100 cards).

## Browser Compatibility

Tested and working on:
- ✅ Chrome 120+ (Chromium)
- ✅ Firefox 120+
- ✅ Safari 17+
- ✅ Edge 120+

Requires:
- ES2020 features (arrow functions, template literals)
- `Element.classList`
- `Element.scrollIntoView`
- `window.scrollTo` with options

All features supported by target browsers (2023+).

## Screen Reader Support

### Live Region

```html
<div id="kbd-live-region" 
     aria-live="polite" 
     aria-atomic="true" 
     class="sr-only">
  <!-- Announcements inserted here -->
</div>
```

- `aria-live="polite"`: Announces when user pauses
- `aria-atomic="true"`: Reads entire region, not just changes
- `.sr-only`: Visually hidden, but read by screen readers

### Announcements

Format: `"Focused: [App Name] [Version]"`

Examples:
- "Focused: Firefox 147.0.3"
- "Focused: Bluefin stable-20260203"

Extracted from card's `<h2>` or `<h3>` element.

## Related Documentation

- [Playwright Tests](../e2e/keyboard-shortcuts.spec.ts) - E2E test suite
- [Jest Tests](../src/scripts/keyboard-nav.test.ts) - Unit test suite
- [PR Template](../.github/PULL_REQUEST_TEMPLATE.md) - Manual testing checklist
- [CI Workflow](../.github/workflows/ci.yml) - Automated testing in CI

## Troubleshooting Checklist

If shortcuts aren't working:

- [ ] Open browser console - any errors?
- [ ] Check init message: "Initialized with N items" (N > 0)?
- [ ] Run: `document.querySelectorAll('.grouped-release-card').length`
- [ ] Does number match init message?
- [ ] Press `j` - does card get `.kbd-focused` class?
- [ ] Is card scrolling into view?
- [ ] Are you typing in search when pressing keys?
- [ ] Did you rebuild after code changes? (`npm run build`)

Still stuck? Check:
1. Git blame to see recent changes to `keyboard-nav.ts`
2. Run E2E tests: `npm run test:e2e`
3. Compare with last working commit

## Future Enhancements

Potential improvements (not yet implemented):

1. **Vim-style navigation**: `gg` (top), `G` (bottom), `Ctrl+d/u` (half-page)
2. **Jump to letter**: Press `f` + letter to jump to first app starting with that letter
3. **Search within results**: `/` opens search, `n` for next match, `N` for previous
4. **Bookmark cards**: `m` to mark, `'` to jump to mark
5. **Preview on focus**: Show quick preview popup when card is focused
6. **Keyboard configuration**: Let users rebind keys via settings
7. **Tour mode**: Guided walkthrough of keyboard shortcuts for new users

## Contributing

When modifying keyboard navigation:

1. **Read this doc first** 🧐
2. **Write tests** (unit + E2E)
3. **Update documentation** (if adding features)
4. **Test manually** (use PR template checklist)
5. **Check console** (no errors/warnings)

Questions? Check existing issues or ask in PRs!

---

**Last Updated**: 2026-02-09  
**Maintainer**: Jorge O. Castro  
**Status**: ✅ Production-ready with comprehensive test coverage
