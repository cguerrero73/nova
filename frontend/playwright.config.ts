import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for Nova E2E tests.
 *
 * - Backend: `make dev` from /backend (Go server on port 4000)
 * - Frontend: `npm start` from /frontend (Angular dev server on port 4200)
 * - Tests run against the real backend; no HTTP mocking.
 *
 * To run: `npm run test:e2e`
 * To debug: `npm run test:e2e:debug`
 * To open UI mode: `npm run test:e2e:ui`
 */
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 5_000 },

  // Run tests in files in parallel
  fullyParallel: true,
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env['CI'] !== undefined,
  // Retry on CI only
  retries: process.env['CI'] !== undefined ? 2 : 0,
  // Opt out of parallel tests on CI
  workers: process.env['CI'] !== undefined ? 1 : undefined,

  reporter: process.env['CI'] !== undefined ? 'github' : 'html',

  use: {
    baseURL: 'http://localhost:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Start backend and frontend before tests
  webServer: [
    {
      command: 'make dev',
      cwd: '../backend',
      port: 4000,
      reuseExistingServer: !process.env['CI'] !== undefined,
      timeout: 60_000,
      stdout: 'ignore',
      stderr: 'pipe',
    },
    {
      command: 'npm start -- --port 4200',
      port: 4200,
      reuseExistingServer: !process.env['CI'] !== undefined,
      timeout: 60_000,
      stdout: 'ignore',
      stderr: 'pipe',
    },
  ],
});
