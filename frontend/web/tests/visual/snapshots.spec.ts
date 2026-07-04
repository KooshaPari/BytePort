import { test, expect } from '@playwright/test';

/**
 * Visual-regression snapshots for the BytePort splash + empty-state surfaces.
 *
 * Snapshots intentionally cover the surfaces that end-users see at:
 * - desktop boot (splash)
 * - any empty list (no-data, no-results, error, mascot)
 *
 * These are the user-visible brand surfaces. Any drift here is a regression.
 */

test.describe('BytePort splash', () => {
  test('desktop-light matches', async ({ page }) => {
    await page.goto('/__visual/splash');
    await expect(page).toHaveScreenshot('splash-desktop-light.png', { fullPage: false });
  });

  test('desktop-dark matches', async ({ page }) => {
    await page.goto('/__visual/splash');
    await expect(page).toHaveScreenshot('splash-desktop-dark.png', { fullPage: false });
  });

  test('mobile-light matches', async ({ page }) => {
    await page.goto('/__visual/splash');
    await expect(page).toHaveScreenshot('splash-mobile-light.png', { fullPage: false });
  });
});

test.describe('BytePort empty states', () => {
  for (const variant of ['no-data', 'no-results', 'error', 'mascot'] as const) {
    test(`${variant} matches on desktop-light`, async ({ page }) => {
      await page.goto(`/__visual/empty-state?variant=${variant}`);
      await expect(page).toHaveScreenshot(`empty-${variant}-desktop-light.png`);
    });

    test(`${variant} matches on mobile-light`, async ({ page }) => {
      await page.goto(`/__visual/empty-state?variant=${variant}`);
      await expect(page).toHaveScreenshot(`empty-${variant}-mobile-light.png`);
    });
  }
});

test.describe('BytePort mascot', () => {
  test('mascot idle frame (mid-breath)', async ({ page }) => {
    await page.goto('/__visual/mascot');
    // wait one full breath cycle so SMIL is mid-state, not at frame-0
    await page.waitForTimeout(2100);
    await expect(page).toHaveScreenshot('mascot-idle.png');
  });
});