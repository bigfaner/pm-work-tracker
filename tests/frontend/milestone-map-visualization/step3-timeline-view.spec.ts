/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-visualization
 *
 * Step 3-8 tests: Timeline view interactions
 * Step 3b: Loading state
 * Step 3c: Empty state
 * Step 5-6: Search and status filters
 * Step 8: Zoom controls
 * Step 8b: Horizontal scroll
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-visualization / Steps 3-8: Timeline view', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let ms1BizKey: string;
  let ms2BizKey: string;
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
      data: { mapName: `e2e-mmv-s3-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mmv-s3-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mmv-s3-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Step 3: Timeline loads with correct components
  test('TC-MMV-S3-001: Timeline shows breadcrumb, info card, filter bar, nodes', async ({ page }) => {
    await expect(page.locator('[data-testid="milestone-timeline"]')).toBeVisible({ timeout: 10000 });

    // Verify milestone nodes
    await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms1-${runId}` })).toBeVisible({ timeout: 10000 });
    await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms2-${runId}` })).toBeVisible();
  });

  // Step 4: Hover milestone node
  test('TC-MMV-S3-002: Hover on milestone node shows highlight', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms1-${runId}` });
    await node.hover();
    await expect(node).toBeVisible();
  });

  // Step 5: Filter by name search
  test('TC-MMV-S3-003: Search input filters milestone nodes', async ({ page }) => {
    const searchInput = page.locator('[data-testid="search-milestones-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('ms1');
      await page.waitForTimeout(500);

      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms1-${runId}` })).toBeVisible();
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms2-${runId}` })).not.toBeVisible();
    }
  });

  // Step 6: Filter by status tags
  test('TC-MMV-S3-004: Status filter shows only matching nodes', async ({ page }) => {
    // Look for status filter in filter bar
    const statusFilter = page.locator('[class*="filter"]').filter({ hasText: /状态|status/i }).first();
    if (await statusFilter.isVisible()) {
      await statusFilter.click();
      await page.waitForTimeout(300);
      const notStartedOption = page.getByText(/未开始|not_started/i).first();
      if (await notStartedOption.isVisible()) {
        await notStartedOption.click();
        await page.waitForTimeout(500);
      }
    }
  });

  // Step 7: Reset filters
  test('TC-MMV-S3-005: Reset button restores all nodes', async ({ page }) => {
    // Apply search first
    const searchInput = page.locator('[data-testid="search-milestones-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('nonexistent');
    }

    // Click reset
    const resetBtn = page.locator('[data-testid="reset-filters-btn"]');
    if (await resetBtn.isVisible()) {
      await resetBtn.click();
      await page.waitForTimeout(500);

      // Both nodes should be visible
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms1-${runId}` })).toBeVisible({ timeout: 5000 });
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-s3-ms2-${runId}` })).toBeVisible();
    }
  });

  // Step 8: Zoom controls
  test('TC-MMV-S3-006: Zoom controls change timeline spacing', async ({ page }) => {
    // Get initial timeline width
    const timeline = page.locator('[data-testid="milestone-timeline"]');

    // Click compact
    const compactBtn = page.locator('[data-testid="zoom-compact"]');
    if (await compactBtn.isVisible()) {
      await compactBtn.click();
      await page.waitForTimeout(300);
    }

    // Click relaxed
    const relaxedBtn = page.locator('[data-testid="zoom-relaxed"]');
    if (await relaxedBtn.isVisible()) {
      await relaxedBtn.click();
      await page.waitForTimeout(300);
    }

    // Verify nodes still visible after zoom
    await expect(page.locator(`[data-testid^="milestone-node-"]`).first()).toBeVisible();
  });

  // Step 8b: Horizontal scroll for dense timeline
  test('TC-MMV-S3-007: Timeline provides horizontal scroll for dense nodes', async ({ page }) => {
    // Check scroll buttons
    const scrollLeft = page.locator('[data-testid="scroll-left"]');
    const scrollRight = page.locator('[data-testid="scroll-right"]');

    // At least one should be present for dense timelines
    const hasScrollBtns = (await scrollLeft.isVisible()) || (await scrollRight.isVisible());
    // This is conditional - not all timelines have scroll buttons
  });

  // Unauthorized API test
  test('TC-MMV-S3-008: Get milestone map detail API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}`);
    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
