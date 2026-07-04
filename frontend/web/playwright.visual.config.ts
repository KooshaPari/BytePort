import { defineConfig, devices } from '@playwright/test';

/**
 * Visual-regression suite for BytePort UI surface (L60).
 *
 * Spec: snapshot the rendered DOM at known viewport sizes + themes.
 * Snapshots land under `frontend/web/tests/visual/__snapshots__/`.
 * CI: gated by `npm run test:visual`. Local dev: `npx playwright test tests/visual`.
 *
 * Anti-flake:
 * - `animations: 'disabled'` stops CSS transitions + animations during render.
 * - `mask` strips dynamic content (timestamps, IDs) so snapshots are deterministic.
 */
export default defineConfig({
  testDir: './tests/visual',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI ? 'github' : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'desktop-light',
      use: { ...devices['Desktop Chrome'], colorScheme: 'light' },
    },
    {
      name: 'desktop-dark',
      use: { ...devices['Desktop Chrome'], colorScheme: 'dark' },
    },
    {
      name: 'mobile-light',
      use: { ...devices['Pixel 7'], colorScheme: 'light' },
    },
    {
      name: 'mobile-dark',
      use: { ...devices['Pixel 7'], colorScheme: 'dark' },
    },
  ],
  expect: {
    toHaveScreenshot: {
      animations: 'disabled',
      caret: 'hide',
      scale: 'device',
      maxDiffPixelRatio: 0.01,
    },
  },
  webServer: process.env.PLAYWRIGHT_BASE_URL
    ? undefined
    : {
        command: 'npm run build && npm run preview -- --port 4173',
        url: 'http://127.0.0.1:4173',
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      },
});