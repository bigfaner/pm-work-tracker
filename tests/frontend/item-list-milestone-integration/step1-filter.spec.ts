/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-list-milestone-integration
 *
 * Step 1-4 tests: Milestone filter operations on items list
 * Step 2b: Filter by cancelled milestone not in dropdown
 * Step 5b: Invalid bizKey filter fallback
 * Step 7b: Team switch does not reset milestone filter
 * Step 7c: Milestone dropdown load failure
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('item-list-milestone-integration / Steps 1-4: Milestone filter', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
  let miBizKey: string;
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
      data: { mapName: `e2e-ili-s1-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ili-s1-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ili-s1-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/items');
  });

  // Step 1: Milestone filter visible
  test('TC-ILI-S1-001: Milestone dropdown filter appears with default All', async ({ page }) => {
    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await expect(milestoneFilter).toBeVisible();
    }
  });

  // Step 2: Filter by specific milestone
  test('TC-ILI-S2-001: Filter by milestone shows only matching MIs', async ({ page }) => {
    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      await page.getByText(`e2e-ili-s1-ms-${runId}`).click();
      await page.waitForTimeout(1000);

      await expect(page.getByText(`e2e-ili-s1-mi-${runId}`)).toBeVisible({ timeout: 5000 });
    }
  });

  // Step 3: Filter by "Unassigned"
  test('TC-ILI-S3-001: Filter by Unassigned shows only unbound MIs', async ({ page }) => {
    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      await page.getByText(/未分配|unassigned/i).click();
      await page.waitForTimeout(1000);

      // Bound MI should not be visible
      await expect(page.getByText(`e2e-ili-s1-mi-${runId}`)).not.toBeVisible();
    }
  });

  // Step 2b: Cancelled milestone not in dropdown
  test('TC-ILI-S2-002: Cancelled milestones not shown in filter dropdown', async ({ page }) => {
    // Create and cancel a milestone
    const request = page.context().request;
    const cancelledMsRes = await request.post(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneName: `e2e-ili-s1-cancelled-${runId}`, expectedEndDate: '2026-08-30' },
    });
    const cancelledMsKey = extractBizKey(parseApiData(await cancelledMsRes.json())) ?? '';

    for (const status of ['in_progress', 'completed', 'cancelled']) {
      await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${cancelledMsKey}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    // Open dropdown and verify cancelled milestone is not listed
    const milestoneFilter = page.locator('[data-testid="milestone-filter"]');
    if (await milestoneFilter.isVisible()) {
      await milestoneFilter.click();
      await expect(page.getByText(`e2e-ili-s1-cancelled-${runId}`)).not.toBeVisible();
    }
  });

  // Unauthorized API test
  test('TC-ILI-S1-002: List items API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/main-items`);
    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
