import { test as base, Page } from '@playwright/test';

/**
 * Test credentials for the acme tenant.
 * Must match a real seeded user in the backend.
 */
export const TEST_USER = {
  code: 'admin',
  password: 'admin123',
  tenant: 'acme',
} as const;

/**
 * Authenticated fixture: every test that uses this fixture starts on the
 * dashboard with a valid session.
 *
 * Logs in via the actual /login form so we exercise the same flow as a real user.
 */
export const test = base.extend<{ authenticatedPage: Page }>({
  authenticatedPage: async ({ page }, use) => {
    await page.goto('/login?tenant=acme');
    await page.locator('#code').fill(TEST_USER.code);
    await page.locator('#password').fill(TEST_USER.password);
    await page.locator('button[type="submit"]').click();

    // Wait for the post-login redirect to settle.
    // Auth service stores the token; router navigates to /dashboard.
    await page.waitForURL(/\/dashboard/, { timeout: 10_000 });

    await use(page);
  },
});

export { expect } from '@playwright/test';
