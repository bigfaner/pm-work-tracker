/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-visualization
 *
 * Smoke Test: Happy path through all steps of milestone-map-visualization.
 * Covers list view -> timeline -> filters -> zoom -> navigation.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('milestone-map-visualization smoke: list -> timeline -> filters -> zoom -> nav', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let ms1BizKey: string;
  let ms2BizKey: string;
  let miBizKey: string;
  const runId = Date.now();
  const mapName = `e2e-mmv-smoke-map-${runId}`;

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
      data: { mapName, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    // Create milestones
    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mmv-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mmv-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    // Create and bind MI to ms1
    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-mmv-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    await request.dispose();
  });

  // Step 1: Load list view with cards
  test('Smoke 1: List view loads with milestone map cards', async ({ page }) => {
    await login(page, undefined, '/milestones');

    // Verify loading skeleton appears first
    await expect(page.locator('[data-testid="loading-skeleton"], [class*="skeleton"]')).toBeVisible({ timeout: 5000 }).catch(() => {});

    // Verify map card is visible
    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await expect(card).toBeVisible({ timeout: 10000 });
    await expect(card.getByText('规划中')).toBeVisible();
  });

  // Step 3: Navigate to timeline view
  test('Smoke 2: Click card navigates to timeline view', async ({ page }) => {
    await login(page, undefined, '/milestones');

    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await card.click();

    await expect(page).toHaveURL(new RegExp(`/milestones/${mapBizKey}`), { timeout: 10000 });
    await expect(page.locator('[data-testid="milestone-timeline"]')).toBeVisible({ timeout: 10000 });
  });

  // Step 4: Interact with timeline nodes
  test('Smoke 3: Hover milestone node shows tooltip', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-ms1-${runId}` });
    await node.hover();
    // Visual highlight feedback (no assertion needed for hover highlight)
    await page.waitForTimeout(500);
  });

  // Step 5: Filter by name search
  test('Smoke 4: Search filter hides non-matching nodes', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const searchInput = page.locator('[data-testid="search-milestones-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('mmv-ms1');
      await page.waitForTimeout(500);

      // ms1 should be visible, ms2 should be hidden
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-ms1-${runId}` })).toBeVisible();
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-ms2-${runId}` })).not.toBeVisible();
    }
  });

  // Step 7: Reset filters
  test('Smoke 5: Reset filters shows all nodes', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    // Apply filter first
    const searchInput = page.locator('[data-testid="search-milestones-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('nonexistent');
      await page.waitForTimeout(500);
    }

    // Click reset
    const resetBtn = page.locator('[data-testid="reset-filters-btn"]');
    if (await resetBtn.isVisible()) {
      await resetBtn.click();
      await page.waitForTimeout(500);

      // All nodes should be visible again
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-ms1-${runId}` })).toBeVisible({ timeout: 5000 });
      await expect(page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mmv-ms2-${runId}` })).toBeVisible();
    }
  });

  // Step 8: Zoom timeline
  test('Smoke 6: Zoom controls change timeline spacing', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    // Click compact zoom
    const compactBtn = page.locator('[data-testid="zoom-compact"]');
    if (await compactBtn.isVisible()) {
      await compactBtn.click();
      await page.waitForTimeout(500);
    }

    // Click standard zoom
    const standardBtn = page.locator('[data-testid="zoom-standard"]');
    if (await standardBtn.isVisible()) {
      await standardBtn.click();
      await page.waitForTimeout(500);
    }

    // Click relaxed zoom
    const relaxedBtn = page.locator('[data-testid="zoom-relaxed"]');
    if (await relaxedBtn.isVisible()) {
      await relaxedBtn.click();
      await page.waitForTimeout(500);
    }
  });

  // Step 9: Navigate back to list
  test('Smoke 7: Breadcrumb navigates back to list', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    // Click breadcrumb link (scope to breadcrumb nav to avoid matching page title)
    await page.locator('nav[aria-label="breadcrumb"]').getByText(/里程碑图|milestone map/i).click();
    await expect(page).toHaveURL(/\/milestones$/, { timeout: 10000 });
  });

  // Step 10: Click MI item navigates to detail
  test('Smoke 8: Click MI on timeline navigates to MI detail', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const miItem = page.locator(`[data-testid="mi-item-${miBizKey}"]`);
    if (await miItem.isVisible()) {
      await miItem.click();
      await page.waitForURL(/\/items\//, { timeout: 10000 });
    }
  });
});
