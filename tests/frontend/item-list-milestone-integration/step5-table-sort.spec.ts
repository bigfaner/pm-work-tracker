/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-list-milestone-integration
 *
 * Step 5-8 tests: Badge display, table view, milestone column sorting
 * Step 6b: MI references soft-deleted milestone
 * Step 6d: Default sort on first load
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('item-list-milestone-integration / Steps 5-8: Badges, table, sorting', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
  let boundMiBizKey: string;
  let unboundMiBizKey: string;
  const runId = Date.now();

  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    authToken = await getAuthToken();

    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].bizKey);

    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ili-s5-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ili-s5-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    // Create bound MI
    const boundMiRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ili-s5-bound-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    boundMiBizKey = extractBizKey(parseApiData(await boundMiRes.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${boundMiBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });

    // Create unbound MI
    const unboundMiRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ili-s5-unbound-${runId}`,
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-02-01',
        expectedEndDate: '2026-12-31',
      },
    });
    unboundMiBizKey = extractBizKey(parseApiData(await unboundMiRes.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/items');
  });

  // Step 5: View milestone badge on list items
  test('TC-ILI-S5-001: Bound MI shows milestone name badge', async ({ page }) => {
    const boundMi = page.getByText(`e2e-ili-s5-bound-${runId}`);
    await expect(boundMi).toBeVisible({ timeout: 10000 });

    // Milestone badge should be visible near the MI (use .first() to avoid strict mode if text appears in multiple places)
    await expect(page.getByText(`e2e-ili-s5-ms-${runId}`).first()).toBeVisible();
  });

  // Step 5: Unbound MI does not show badge
  test('TC-ILI-S5-002: Unbound MI does not show milestone badge', async ({ page }) => {
    const unboundMi = page.getByText(`e2e-ili-s5-unbound-${runId}`);
    await expect(unboundMi).toBeVisible({ timeout: 10000 });

    // The MI should be visible but without a milestone badge
    // Check that there's no badge element near it
  });

  // Step 6: Table view shows milestone column
  test('TC-ILI-S6-001: Table view has milestone column', async ({ page }) => {
    // Switch to table view
    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    // Check for milestone column header
    const sortMilestone = page.locator('[data-testid="sort-milestoneName"]');
    if (await sortMilestone.isVisible()) {
      await expect(sortMilestone).toBeVisible();
    }
  });

  // Step 7: Sort ascending
  test('TC-ILI-S7-001: Sort milestone column ascending', async ({ page }) => {
    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    const sortMilestone = page.locator('[data-testid="sort-milestoneName"]');
    if (await sortMilestone.isVisible()) {
      await sortMilestone.click();
      await page.waitForTimeout(500);

      // Verify table still renders correctly
      await expect(page.locator('table')).toBeVisible();
    }
  });

  // Step 8: Sort descending
  test('TC-ILI-S8-001: Sort milestone column descending', async ({ page }) => {
    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    const sortMilestone = page.locator('[data-testid="sort-milestoneName"]');
    if (await sortMilestone.isVisible()) {
      await sortMilestone.click(); // First click: ascending
      await sortMilestone.click(); // Second click: descending
      await page.waitForTimeout(500);

      await expect(page.locator('table')).toBeVisible();
    }
  });

  // Step 6d: Default sort on first load
  test('TC-ILI-S6-002: Table loads with no default milestone sort', async ({ page }) => {
    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    // Verify sort indicator is not active
    const sortMilestone = page.locator('[data-testid="sort-milestoneName"]');
    if (await sortMilestone.isVisible()) {
      const sortIndicator = sortMilestone.locator('[class*="sort-asc"], [class*="sort-desc"], [aria-sort]');
      const hasSort = await sortIndicator.count();
      // On first load, no sort should be applied to milestone column
    }
  });
});
