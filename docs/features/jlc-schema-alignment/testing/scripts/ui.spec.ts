import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import type { Page } from 'playwright';
import {
  setupBrowser,
  teardownBrowser,
  screenshot,
  baseUrl,
  loginViaUI,
} from './helpers.js';

describe('UI E2E Tests', () => {
  let page: Page;

  before(async () => {
    page = await setupBrowser();
    await loginViaUI(page);
  });

  after(async () => {
    await teardownBrowser();
  });

  // Traceability: TC-001 → Story 6 / AC-1
  test('TC-001: 主事项详情页 URL 使用 bizKey 导航', async () => {
    // Step 1: Navigate to items list page
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    // Select team via layout combobox (L-002) if no items are visible
    const noTeamMsg = page.getByText('请先选择团队');
    if (await noTeamMsg.isVisible()) {
      await page.getByRole('combobox', { name: /团队/i }).click();
      // Pick first team option from dropdown
      const firstOption = page.getByRole('option').first();
      await firstOption.click();
      await page.waitForLoadState('networkidle');
    }

    // Verify items list heading is visible (E-005)
    await page.getByRole('heading', { name: '事项清单', level: 1 }).waitFor({ state: 'visible' });

    // Click "明细" tab to show table view (if on summary view)
    const detailBtn = page.getByRole('button', { name: '明细' });
    if (await detailBtn.isVisible()) {
      await detailBtn.click();
      await page.waitForLoadState('networkidle');
    }

    // Step 2: Click first main item row link to enter detail page
    // The item title is a <Link> inside a table row
    const firstItemLink = page.locator('table tbody tr').first().locator('a').first();
    await firstItemLink.waitFor({ state: 'visible', timeout: 5000 }).catch(() => null);

    // Fallback: click any link inside the table body
    const tableLink = page.locator('table tbody a').first();
    await tableLink.waitFor({ state: 'visible', timeout: 5000 });
    await tableLink.click();

    // Wait for detail page navigation
    await page.waitForURL(/\/items\/\d+/, { timeout: 10000 });

    // Expected: URL should contain a bizKey (snowflake ID, ~16 digits, not short auto-increment)
    const currentUrl = page.url();
    const urlMatch = currentUrl.match(/\/items\/(\d+)/);
    assert.ok(urlMatch, `URL should match /items/{bizKey} pattern, got: ${currentUrl}`);
    const bizKey = urlMatch![1];
    assert.ok(
      bizKey.length >= 10,
      `bizKey should be a long snowflake ID (>=10 digits), got: ${bizKey} (${bizKey.length} digits)`,
    );

    // Verify detail page breadcrumb (E-021) is visible
    await page.getByRole('navigation', { name: 'Breadcrumb' }).waitFor({ state: 'visible' });

    await screenshot(page, 'TC-001');
  });
});
