import { test, expect } from '@playwright/test';

const SMOKE_ROUTES = ['/', '/login', '/signup'];

for (const route of SMOKE_ROUTES) {
	test(`route ${route} renders successfully`, async ({ page }) => {
		await page.goto(route);
		await expect(page.locator('body')).toBeVisible();
	});
}
