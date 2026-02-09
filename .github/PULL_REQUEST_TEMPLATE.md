# Pull Request

## Description

<!-- Briefly describe what this PR does -->

## Changes

<!-- List the main changes made in this PR -->

- 
- 
- 

## Testing

<!-- Describe how you tested your changes -->

---

## Keyboard Shortcuts Testing Checklist

**⚠️ Required if your changes touch keyboard navigation, search, filters, cards, or modals**

Before merging changes that affect keyboard shortcuts, test these scenarios in your browser:

### Navigation Shortcuts

- [ ] **`j`** - Moves focus to next card (visual indicator appears)
- [ ] **`k`** - Moves focus to previous card
- [ ] Focused card scrolls into view automatically
- [ ] Focus stays at boundaries (doesn't wrap past first/last card)

### Action Shortcuts

- [ ] **`o`** - Opens focused card's link in new tab
- [ ] **`Enter`** - Also opens focused card's link in new tab
- [ ] **`/`** - Focuses search input
- [ ] **`s`** - Also focuses search input
- [ ] **`?`** - Opens keyboard help modal
- [ ] **`Esc`** - Closes keyboard help modal (when open)
- [ ] **`Esc`** - Blurs search input (when focused)
- [ ] **`Esc`** - Clears card focus (when a card is focused)

### Theme & Scroll Shortcuts

- [ ] **`t`** - Toggles between light/dark theme
- [ ] **`Space`** - Scrolls page down
- [ ] **`Shift+Space`** - Scrolls page up
- [ ] **`h`** - Scrolls to top of page

### Context Awareness

- [ ] Shortcuts are **disabled** when typing in search bar
- [ ] Shortcuts **work again** after pressing `Esc` to blur search
- [ ] Shortcuts work after using filters
- [ ] Shortcuts work after expanding/collapsing older releases
- [ ] Shortcuts work on initial page load

### Screen Reader Announcements

- [ ] Navigating with `j`/`k` announces focused app name to screen readers
- [ ] Live region (`#kbd-live-region`) updates without page refresh
- [ ] No duplicate announcements

### Console Checks

- [ ] **No JavaScript errors** in browser console (F12)
- [ ] Console shows `[KeyboardNav] Initialized with N items` (where N > 0)
- [ ] No warnings about missing DOM elements
- [ ] No 404s for missing resources

---

## Automated Tests

<!-- Mark which tests you ran -->

- [ ] Unit tests pass: `npm test`
- [ ] E2E tests pass: `npm run test:e2e`
- [ ] Build succeeds: `npm run build`
- [ ] Preview works: `npm run preview`

---

## Additional Notes

<!-- Any other context, screenshots, or information reviewers should know -->
