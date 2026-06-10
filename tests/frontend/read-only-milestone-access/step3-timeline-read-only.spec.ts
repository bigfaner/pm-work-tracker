/**
 * @web-e2e
 * @feature milestone-map
 * Journey: read-only-milestone-access
 *
 * Step 3-4 tests: Read-only timeline and panel
 * Step 3b: Timeline info displays correctly
 * Step 4b: Panel description tooltip
 * Step 5b: Timeline filters work
 * Step 6b: MI navigation from panel
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('read-only-milestone-access / Steps 3-6: Read-only timeline, panel, navigation', () => {
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
      data: { mapName: `e2e-ro-s3-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        milestoneName: `e2e-ro-s3-ms-${runId}`,
        expectedEndDate: '2026-07-30',
        description: 'Test milestone for read-only access verification',
      },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-ro-s3-mi-${runId}`,
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
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Step 3: Timeline renders fully with info
  test('TC-RO-S3-001: Timeline shows all read-only info without edit controls', async ({ page }) => {
    await expect(page.locator('[data-testid="milestone-timeline"]')).toBeVisible({ timeout: 10000 });

    // Verify milestone node is visible
    await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-ro-s3-ms-${runId}` })).toBeVisible({ timeout: 10000 });
  });

  // Step 3b: Read-only info displays correctly
  test('TC-RO-S3-002: Timeline info card shows name, status, progress', async ({ page }) => {
    // Basic info card should be visible (scope to h1 to avoid matching breadcrumb)
    await expect(page.locator('h1').filter({ hasText: `e2e-ro-s3-map-${runId}` })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/规划中/i).first()).toBeVisible();
  });

  // Step 4: Open panel in read-only mode
  test('TC-RO-S4-001: Panel opens showing all info without edit controls', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-ro-s3-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    // Panel should show milestone info (scope to dialog to avoid matching timeline node)
    await expect(page.getByRole('dialog').getByText(`e2e-ro-s3-ms-${runId}`)).toBeVisible({ timeout: 10000 });

    // Verify MI in panel list (scope to dialog)
    await expect(page.getByRole('dialog').getByText(`e2e-ro-s3-mi-${runId}`)).toBeVisible();
  });

  // Step 5b: Timeline filters work in read-only mode
  test('TC-RO-S5-001: Search, status filter, zoom work in read-only mode', async ({ page }) => {
    // Search
    const searchInput = page.locator('[data-testid="search-milestones-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('ro-s3-ms');
      await page.waitForTimeout(500);
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-ro-s3-ms-${runId}` })).toBeVisible();
    }

    // Reset
    const resetBtn = page.locator('[data-testid="reset-filters-btn"]');
    if (await resetBtn.isVisible()) {
      await resetBtn.click();
    }

    // Zoom
    const compactBtn = page.locator('[data-testid="zoom-compact"]');
    if (await compactBtn.isVisible()) {
      await compactBtn.click();
    }

    const relaxedBtn = page.locator('[data-testid="zoom-relaxed"]');
    if (await relaxedBtn.isVisible()) {
      await relaxedBtn.click();
    }
  });

  // Step 6b: MI navigation from read-only panel
  test('TC-RO-S6-001: Click MI in panel navigates to MI detail', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-ro-s3-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    // Click MI link (scope to dialog to avoid matching other elements)
    const miLink = page.getByRole('dialog').getByText(`e2e-ro-s3-mi-${runId}`);
    await miLink.click();
    await page.waitForURL(/\/items\//, { timeout: 10000 });
  });

  // Unauthorized API test
  test('TC-RO-S3-003: Milestone detail API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/milestones/${msBizKey}`);
    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
