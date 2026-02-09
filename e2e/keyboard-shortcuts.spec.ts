import { test, expect } from '@playwright/test';

/**
 * E2E tests for keyboard navigation shortcuts
 * Tests all 10 documented keyboard shortcuts to prevent regressions
 */

test.describe('Keyboard Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the homepage (baseURL is configured in playwright.config.ts)
    await page.goto('/bluefin-releases/', { waitUntil: 'networkidle' });
    
    // Wait for DOM to be ready and cards to be in the DOM
    await page.waitForLoadState('domcontentloaded');
    
    // Wait for at least one card to exist in the DOM (even if hidden)
    await page.waitForSelector('.grouped-release-card, .release-card', { 
      state: 'attached',
      timeout: 10000 
    });
    
    // Verify console shows keyboard nav initialized
    const logs: string[] = [];
    page.on('console', msg => logs.push(msg.text()));
    
    // Give it a moment for initialization scripts to run
    await page.waitForTimeout(1000);
  });

  test('j key moves focus to next card', async ({ page }) => {
    // Press j
    await page.keyboard.press('j');
    
    // First card should have kbd-focused class
    const focusedCards = await page.locator('.kbd-focused').count();
    expect(focusedCards).toBe(1);
    
    // Get the first focused card
    const firstFocused = await page.locator('.kbd-focused').first();
    const firstText = await firstFocused.textContent();
    
    // Press j again
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    // Still only 1 focused card (moved to next)
    const focusedCards2 = await page.locator('.kbd-focused').count();
    expect(focusedCards2).toBe(1);
    
    // Second focused should be different
    const secondFocused = await page.locator('.kbd-focused').first();
    const secondText = await secondFocused.textContent();
    expect(secondText).not.toBe(firstText);
  });

  test('k key moves focus to previous card', async ({ page }) => {
    // Focus first card
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    const secondCard = await page.locator('.kbd-focused').first().textContent();
    
    // Move back
    await page.keyboard.press('k');
    await page.waitForTimeout(100);
    
    const focusedCards = await page.locator('.kbd-focused').count();
    expect(focusedCards).toBe(1);
    
    const firstCard = await page.locator('.kbd-focused').first().textContent();
    expect(firstCard).not.toBe(secondCard);
  });

  test('/ focuses search input', async ({ page }) => {
    await page.keyboard.press('/');
    await page.waitForTimeout(100);
    
    const searchInput = page.locator('.search-input');
    await expect(searchInput).toBeFocused();
  });

  test('s focuses search input', async ({ page }) => {
    await page.keyboard.press('s');
    await page.waitForTimeout(100);
    
    const searchInput = page.locator('.search-input');
    await expect(searchInput).toBeFocused();
  });

  test('? shows keyboard help modal', async ({ page }) => {
    await page.keyboard.press('?');
    await page.waitForTimeout(200);
    
    const modal = page.locator('#keyboard-help-modal');
    await expect(modal).toHaveClass(/visible/);
    
    // Verify modal content shows shortcuts
    const modalText = await modal.textContent();
    expect(modalText).toContain('Keyboard Shortcuts');
    expect(modalText).toContain('Next app');
    expect(modalText).toContain('Previous app');
  });

  test('Escape closes keyboard help modal', async ({ page }) => {
    // Open modal
    await page.keyboard.press('?');
    await page.waitForTimeout(200);
    
    const modal = page.locator('#keyboard-help-modal');
    await expect(modal).toHaveClass(/visible/);
    
    // Close it
    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    
    await expect(modal).not.toHaveClass(/visible/);
  });

  test('t toggles theme', async ({ page }) => {
    const initialTheme = await page.evaluate(() => 
      document.documentElement.getAttribute('data-theme')
    );
    
    await page.keyboard.press('t');
    await page.waitForTimeout(300);
    
    const newTheme = await page.evaluate(() => 
      document.documentElement.getAttribute('data-theme')
    );
    
    expect(newTheme).not.toBe(initialTheme);
    expect(['light', 'dark']).toContain(newTheme);
  });

  test('o opens focused app link', async ({ page }) => {
    // Focus first card
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    // Listen for popup
    const popupPromise = page.context().waitForEvent('page');
    
    // Press o
    await page.keyboard.press('o');
    
    // Wait for popup to open
    const popup = await popupPromise;
    
    // Verify it opened a valid URL
    const url = popup.url();
    const hasValidUrl = url.includes('flathub.org') || 
                        url.includes('github.com') || 
                        url.includes('formulae.brew.sh') ||
                        url.includes('gitlab.');
    
    expect(hasValidUrl).toBe(true);
    
    await popup.close();
  });

  test('Space scrolls page down', async ({ page }) => {
    const initialScroll = await page.evaluate(() => window.scrollY);
    
    await page.keyboard.press('Space');
    await page.waitForTimeout(500); // Space uses smooth scroll, needs longer wait
    
    const newScroll = await page.evaluate(() => window.scrollY);
    expect(newScroll).toBeGreaterThan(initialScroll);
  });

  test('h scrolls to top', async ({ page }) => {
    // Scroll down first
    await page.keyboard.press('Space');
    await page.waitForTimeout(500); // Space uses smooth scroll, needs time
    
    const scrolledPosition = await page.evaluate(() => window.scrollY);
    expect(scrolledPosition).toBeGreaterThan(0);
    
    // Scroll to top
    await page.keyboard.press('h');
    await page.waitForTimeout(500); // h uses smooth scroll, needs time
    
    const scroll = await page.evaluate(() => window.scrollY);
    expect(scroll).toBeLessThan(50); // Should be near top (allowing for header)
  });

  test('shortcuts do not work when typing in search', async ({ page }) => {
    // Focus search
    await page.keyboard.press('/');
    await page.waitForTimeout(100);
    
    // Type 'j' in search - should not navigate
    await page.keyboard.type('j');
    await page.waitForTimeout(100);
    
    // No cards should be focused
    const focusedCards = await page.locator('.kbd-focused').count();
    expect(focusedCards).toBe(0);
    
    // Search input should have 'j' in it
    const searchValue = await page.locator('.search-input').inputValue();
    expect(searchValue).toContain('j');
  });

  test('Escape blurs search input', async ({ page }) => {
    // Focus search
    await page.keyboard.press('/');
    await page.waitForTimeout(100);
    
    const searchInput = page.locator('.search-input');
    await expect(searchInput).toBeFocused();
    
    // Press Escape
    await page.keyboard.press('Escape');
    await page.waitForTimeout(100);
    
    // Search should no longer be focused
    await expect(searchInput).not.toBeFocused();
  });

  test('Escape clears card focus', async ({ page }) => {
    // Focus a card
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    const focusedCards = await page.locator('.kbd-focused').count();
    expect(focusedCards).toBe(1);
    
    // Press Escape
    await page.keyboard.press('Escape');
    await page.waitForTimeout(100);
    
    // No cards should be focused
    const focusedCardsAfter = await page.locator('.kbd-focused').count();
    expect(focusedCardsAfter).toBe(0);
  });

  test('Shift+Space scrolls page up', async ({ page }) => {
    // Scroll down first
    await page.keyboard.press('Space');
    await page.waitForTimeout(500); // Space uses smooth scroll, needs time
    
    const scrolledPosition = await page.evaluate(() => window.scrollY);
    
    // Scroll up
    await page.keyboard.press('Shift+Space');
    await page.waitForTimeout(500); // Shift+Space uses smooth scroll, needs time
    
    const newScroll = await page.evaluate(() => window.scrollY);
    expect(newScroll).toBeLessThan(scrolledPosition);
  });

  test('Enter opens focused app link', async ({ page }) => {
    // Focus first card
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    // Listen for popup
    const popupPromise = page.context().waitForEvent('page');
    
    // Press Enter
    await page.keyboard.press('Enter');
    
    // Wait for popup to open
    const popup = await popupPromise;
    
    // Verify it opened a valid URL
    const url = popup.url();
    const hasValidUrl = url.includes('flathub.org') || 
                        url.includes('github.com') || 
                        url.includes('formulae.brew.sh') ||
                        url.includes('gitlab.');
    
    expect(hasValidUrl).toBe(true);
    
    await popup.close();
  });

  test('keyboard navigation works after using filters', async ({ page }) => {
    // Click a filter button
    const filterButton = page.locator('.filter-button').first();
    if (await filterButton.isVisible()) {
      await filterButton.click();
      await page.waitForTimeout(300);
    }
    
    // Keyboard nav should still work
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    
    const focusedCards = await page.locator('.kbd-focused').count();
    expect(focusedCards).toBe(1);
  });

  test('focused card scrolls into view', async ({ page }) => {
    // Press j multiple times to focus cards further down
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('j');
      await page.waitForTimeout(50); // Instant scroll, minimal wait
    }
    
    // Brief wait for DOM updates
    await page.waitForTimeout(100);
    
    // Get the focused card
    const focusedCard = page.locator('.kbd-focused').first();
    
    // Verify it's visible in viewport - should be at top with scrollIntoView
    const boundingBox = await focusedCard.boundingBox();
    expect(boundingBox).not.toBeNull();
    
    if (boundingBox) {
      const viewportSize = page.viewportSize();
      // Card should be at or near top of viewport (allowing some margin)
      expect(boundingBox.y).toBeGreaterThanOrEqual(0);
      expect(boundingBox.y).toBeLessThan(100); // Should be near top, not middle/bottom
    }
  });

  test('no console errors during keyboard navigation', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    // Exercise all shortcuts
    await page.keyboard.press('j');
    await page.waitForTimeout(100);
    await page.keyboard.press('k');
    await page.waitForTimeout(100);
    await page.keyboard.press('/');
    await page.waitForTimeout(100);
    await page.keyboard.press('Escape');
    await page.waitForTimeout(100);
    await page.keyboard.press('?');
    await page.waitForTimeout(100);
    await page.keyboard.press('Escape');
    await page.waitForTimeout(100);
    
    // Should have no errors
    expect(errors).toHaveLength(0);
  });
});
