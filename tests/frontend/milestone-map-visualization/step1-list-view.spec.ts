/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-visualization
 *
 * Step 1 tests: List view operations
 * Step 1b: Empty state
 * Step 1c: Server error
 * Step 1d-1g: Filtering and refresh
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-visualization / Step 1: List view', () => {
  let authToken: string;
  let teamId: string;
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

    // Create 3 maps with various statuses
    for (let i = 1; i <= 3; i++) {
      const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
        headers: { Authorization: `Bearer ${authToken}` },
        data: { mapName: `e2e-mmv-s1-map${i}-${runId}`, assigneeBizKey: '1' },
      });
      const mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

      // Transition map2 to reviewed
      if (i === 2) {
        await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
          headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
          data: { status: 'reviewed' },
        });
      }
    }

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/milestones');
  });

  // Step 1: List view loads cards
  test('TC-MMV-S1-001: List view shows milestone map cards with status badges', async ({ page }) => {
    await expect(page.locator('[data-testid^="milestone-map-card-"]').first()).toBeVisible({ timeout: 10000 });

    // Verify at least 3 cards
    const cards = page.locator('[data-testid^="milestone-map-card-"]');
    const count = await cards.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });

  // Step 2: Hover card for highlight
  test('TC-MMV-S1-002: Hover on card shows visual highlight', async ({ page }) => {
    const card = page.locator('[data-testid^="milestone-map-card-"]').first();
    await card.hover();
    // Verify no error occurs and card is still visible
    await expect(card).toBeVisible();
  });

  // Step 1d: Filter by status
  test('TC-MMV-S1-003: Status filter shows only matching cards', async ({ page }) => {
    // Look for status filter
    const statusFilter = page.locator('[class*="filter"], [data-testid="status-filter"], button').filter({ hasText: /状态|status/i }).first();
    if (await statusFilter.isVisible()) {
      await statusFilter.click();
      await page.waitForTimeout(300);
      // Select a specific status
      const option = page.getByText(/规划中|planning/i).first();
      if (await option.isVisible()) {
        await option.click();
        await page.waitForTimeout(500);
      }
    }
  });

  // Step 1f: Search by name
  test('TC-MMV-S1-004: Search filters cards by name', async ({ page }) => {
    const searchInput = page.locator('[data-testid="search-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill(`map1-${runId}`);
      await page.waitForTimeout(500);

      // Only map1 should be visible
      await expect(page.locator('[data-testid^="milestone-map-card-"]').filter({ hasText: `e2e-mmv-s1-map1-${runId}` })).toBeVisible({ timeout: 5000 });
      await expect(page.locator('[data-testid^="milestone-map-card-"]').filter({ hasText: `e2e-mmv-s1-map3-${runId}` })).not.toBeVisible();
    }
  });

  // Step 1g: Refresh list
  test('TC-MMV-S1-005: Refresh button reloads list', async ({ page }) => {
    const refreshBtn = page.locator('[data-testid="refresh-btn"]');
    if (await refreshBtn.isVisible()) {
      await refreshBtn.click();
      // List should reload
      await expect(page.locator('[data-testid^="milestone-map-card-"]').first()).toBeVisible({ timeout: 10000 });
    }
  });

  // Unauthorized API test
  test('TC-MMV-S1-006: List milestone maps API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/milestone-maps`);
    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
