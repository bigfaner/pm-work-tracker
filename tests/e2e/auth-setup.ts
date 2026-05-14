import { test as setup, expect } from '@playwright/test';
import { baseUrl, ensureAuthState } from './helpers.js';

setup('authenticate', async ({ page }) => {
  await ensureAuthState(page);
  await page.goto(`${baseUrl}/`);
  await expect(page).toHaveURL(/(?!.*login)/);
  // Save browser storage state for authenticated project reuse
  await page.context().storageState({ path: 'results/.auth/state.json' });
});
