import { test as setup, expect } from '@playwright/test';
import { baseUrl, ensureAuthState } from './helpers.js';

setup('authenticate', async ({ page }) => {
  await ensureAuthState(page);
  await page.goto(`${baseUrl}/`);
  await expect(page).toHaveURL(/(?!.*login)/);
});
