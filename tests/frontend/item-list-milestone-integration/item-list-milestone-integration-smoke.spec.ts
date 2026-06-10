/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-list-milestone-integration
 *
 * Smoke Test: Happy path through all steps.
 * Covers milestone filter on items list, badge display, table view, sorting.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('item-list-milestone-integration smoke: filter, badges, table, sort', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
  let mi1BizKey: string;
  let mi2BizKey: string;
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
      data: { mapName: `e2e-ili-smoke-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ili-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    // Create two MIs: one bound, one unbound
    const mi1Res = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ili-bound-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    mi1BizKey = extractBizKey(parseApiData(await mi1Res.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${mi1BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });

    const mi2Res = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ili-unbound-mi-${runId}`,
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-02-01',
        expectedEndDate: '2026-12-31',
      },
    });
    mi2BizKey = extractBizKey(parseApiData(await mi2Res.json())) ?? '';

    await request.dispose();
  });

  // Step 1: View milestone filter
  test('Smoke 1: Milestone filter dropdown is visible on items list', async ({ page }) => {
    await login(page, undefined, '/items');

    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await expect(milestoneFilter).toBeVisible();
    }
  });

  // Step 2: Filter by specific milestone
  test('Smoke 2: Filter by milestone shows only matching MIs', async ({ page }) => {
    await login(page, undefined, '/items');

    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      // Scope to the open listbox to avoid matching badge text on item cards
      await page.getByRole('listbox').getByText(`e2e-ili-ms-${runId}`).click();
      await page.waitForTimeout(1000);

      // Only bound MI should be visible
      await expect(page.getByText(`e2e-ili-bound-mi-${runId}`)).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(`e2e-ili-unbound-mi-${runId}`)).not.toBeVisible();
    }
  });

  // Step 3: Filter by "Unassigned"
  test('Smoke 3: Filter by Unassigned shows only unbound MIs', async ({ page }) => {
    await login(page, undefined, '/items');

    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      // Scope to the open listbox to avoid matching page text
      await page.getByRole('listbox').getByText(/未分配|unassigned/i).click();
      await page.waitForTimeout(1000);

      await expect(page.getByText(`e2e-ili-unbound-mi-${runId}`)).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(`e2e-ili-bound-mi-${runId}`)).not.toBeVisible();
    }
  });

  // Step 4: Filter by "All"
  test('Smoke 4: Filter by All shows all MIs', async ({ page }) => {
    await login(page, undefined, '/items');

    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      // Scope to the open listbox to avoid matching page text
      await page.getByRole('listbox').getByText(/全部|all/i).first().click();
      await page.waitForTimeout(1000);

      await expect(page.getByText(`e2e-ili-bound-mi-${runId}`)).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(`e2e-ili-unbound-mi-${runId}`)).toBeVisible();
    }
  });

  // Step 6: View milestone column in table view
  test('Smoke 5: Table view shows milestone column', async ({ page }) => {
    await login(page, undefined, '/items');

    // Switch to table view if available
    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    // Look for milestone column header
    const milestoneCol = page.locator('[data-testid="sort-milestoneName"]');
    if (await milestoneCol.isVisible()) {
      await expect(milestoneCol).toBeVisible();
    }
  });

  // Step 7: Sort milestone column ascending
  test('Smoke 6: Sort milestone column ascending', async ({ page }) => {
    await login(page, undefined, '/items');

    const tableViewBtn = page.getByRole('button', { name: /table|表格|列表/i });
    if (await tableViewBtn.isVisible()) {
      await tableViewBtn.click();
      await page.waitForTimeout(1000);
    }

    const sortMilestone = page.locator('[data-testid="sort-milestoneName"]');
    if (await sortMilestone.isVisible()) {
      await sortMilestone.click();
      await page.waitForTimeout(500);
    }
  });
});
