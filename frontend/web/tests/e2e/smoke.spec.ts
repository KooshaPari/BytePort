import { test, expect } from '@playwright/test';

test('homepage renders and has expected title', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/.+/);
});

test('login route loads successfully', async ({ page }) => {
  await page.goto('/login');
  await expect(page.locator('body')).toBeVisible();
});

test('signup route loads successfully', async ({ page }) => {
  await page.goto('/signup');
  await expect(page.locator('body')).toBeVisible();
});
