import { test, expect } from '@playwright/test';

test.describe('RSS Feeds', () => {
  const baseUrl = process.env.BASE_URL || 'http://localhost:4321/bluefin-releases';
  
  const rssFeeds = [
    { path: '/rss.xml', name: 'All Releases', shouldHaveItems: true },
    { path: '/rss/flatpak.xml', name: 'Flatpak Releases', shouldHaveItems: true },
    { path: '/rss/homebrew.xml', name: 'Homebrew Releases', shouldHaveItems: false }, // Currently no releases
    { path: '/rss/os.xml', name: 'OS Releases', shouldHaveItems: true },
    { path: '/rss/verified.xml', name: 'Verified Apps', shouldHaveItems: false }, // Currently no verified apps
  ];

  test.describe('Feed Accessibility', () => {
    for (const feed of rssFeeds) {
      test(`${feed.name} feed (${feed.path}) returns HTTP 200`, async ({ page }) => {
        const response = await page.goto(`${baseUrl}${feed.path}`);
        expect(response).not.toBeNull();
        expect(response!.status()).toBe(200);
      });
    }
  });

  test.describe('XML Structure Validation', () => {
    for (const feed of rssFeeds) {
      test(`${feed.name} feed has valid RSS XML structure`, async ({ page }) => {
        const response = await page.goto(`${baseUrl}${feed.path}`);
        expect(response).not.toBeNull();
        
        const contentType = response!.headers()['content-type'];
        expect(contentType).toContain('xml');
        
        const content = await response!.text();
        
        // Check for basic RSS structure (using regex to handle xmlns attributes)
        expect(content).toContain('<?xml version="1.0"');
        expect(content).toMatch(/<rss version="2\.0"(\s+xmlns:[^>]*)?>/);
        expect(content).toContain('<channel>');
        expect(content).toContain('<title>');
        expect(content).toContain('<description>');
        expect(content).toContain('<link>');
        expect(content).toContain('</channel>');
        expect(content).toContain('</rss>');
      });
    }
  });

  test.describe('Feed Content Validation', () => {
    for (const feed of rssFeeds) {
      test(`${feed.name} feed has expected title`, async ({ page }) => {
        const response = await page.goto(`${baseUrl}${feed.path}`);
        const content = await response!.text();
        
        expect(content).toContain('Bluefin Firehose');
      });

      if (feed.shouldHaveItems) {
        test(`${feed.name} feed has items`, async ({ page }) => {
          const response = await page.goto(`${baseUrl}${feed.path}`);
          const content = await response!.text();
          
          // Check for at least one <item> tag
          const itemMatches = content.match(/<item>/g);
          expect(itemMatches).not.toBeNull();
          expect(itemMatches!.length).toBeGreaterThan(0);
        });

        test(`${feed.name} feed items have required fields`, async ({ page }) => {
          const response = await page.goto(`${baseUrl}${feed.path}`);
          const content = await response!.text();
          
          // If there are items, they should have title, link, pubDate, and description
          if (content.includes('<item>')) {
            expect(content).toContain('<title>');
            expect(content).toContain('<link>');
            expect(content).toContain('<pubDate>');
            expect(content).toContain('<description>');
          }
        });
      } else {
        test(`${feed.name} feed is empty (expected)`, async ({ page }) => {
          const response = await page.goto(`${baseUrl}${feed.path}`);
          const content = await response!.text();
          
          // Should have channel but no items
          expect(content).toContain('<channel>');
          expect(content).not.toContain('<item>');
        });
      }
    }
  });

  test.describe('Feed Discovery', () => {
    test('main page has RSS feed links in HTML head', async ({ page }) => {
      await page.goto(`${baseUrl}/`);
      
      // Check for RSS feed link tags in HTML head
      const rssLinks = await page.locator('link[type="application/rss+xml"]').all();
      expect(rssLinks.length).toBeGreaterThanOrEqual(5);
      
      // Verify each feed is declared
      const titles = await Promise.all(rssLinks.map(link => link.getAttribute('title')));
      expect(titles).toContain('Bluefin Firehose - All Releases');
      expect(titles).toContain('Bluefin Firehose - Flatpak Releases');
      expect(titles).toContain('Bluefin Firehose - Homebrew Releases');
      expect(titles).toContain('Bluefin Firehose - OS Releases');
      expect(titles).toContain('Bluefin Firehose - Verified Apps Only');
    });

    test('main page footer has RSS feed links', async ({ page }) => {
      await page.goto(`${baseUrl}/`);
      
      // Check for RSS feed links in footer
      const allFeedLink = await page.locator('a[href*="rss.xml"]:has-text("RSS Feed (All)")');
      await expect(allFeedLink).toBeVisible();
      
      const flatpakFeedLink = await page.locator('a[href*="rss/flatpak.xml"]:has-text("RSS Feed (Flathub)")');
      await expect(flatpakFeedLink).toBeVisible();
      
      const homebrewFeedLink = await page.locator('a[href*="rss/homebrew.xml"]:has-text("RSS Feed (Homebrew)")');
      await expect(homebrewFeedLink).toBeVisible();
      
      const osFeedLink = await page.locator('a[href*="rss/os.xml"]:has-text("RSS Feed (OS)")');
      await expect(osFeedLink).toBeVisible();
      
      const verifiedFeedLink = await page.locator('a[href*="rss/verified.xml"]:has-text("RSS Feed (Verified)")');
      await expect(verifiedFeedLink).toBeVisible();
    });
  });

  test.describe('Date Validation', () => {
    test('All Releases feed has valid RFC 822 dates', async ({ page }) => {
      const response = await page.goto(`${baseUrl}/rss.xml`);
      const content = await response!.text();
      
      // Extract pubDate values
      const pubDateRegex = /<pubDate>([^<]+)<\/pubDate>/g;
      const matches = [...content.matchAll(pubDateRegex)];
      
      if (matches.length > 0) {
        // Check first pubDate is a valid date
        const firstDate = matches[0][1];
        const parsedDate = new Date(firstDate);
        expect(parsedDate.toString()).not.toBe('Invalid Date');
      }
    });
  });
});
