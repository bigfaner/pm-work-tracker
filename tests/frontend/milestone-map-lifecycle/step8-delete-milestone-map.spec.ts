/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 8: Delete milestone map
 * Contracts: step-8-delete-milestone-map.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 8: Delete milestone map', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKeyPlanning: string;
  let mapBizKeyExecuting: string;
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

    // Create map in planning status (deletable)
    const mapRes1 = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s8-del-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKeyPlanning = extractBizKey(parseApiData(await mapRes1.json())) ?? '';

    // Create map in executing status (non-deletable)
    const mapRes2 = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s8-exec-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKeyExecuting = extractBizKey(parseApiData(await mapRes2.json())) ?? '';

    for (const status of ['reviewed', 'ready', 'executing']) {
      await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKeyExecuting}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    await request.dispose();
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "success"
  test('TC-MML-S8-001: Delete planning milestone map redirects to list', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKeyPlanning}`);

    // Click delete button
    await page.locator('[data-testid="delete-map-btn"]').click();

    // Confirm deletion
    await page.getByRole('button', { name: '确认删除' }).click();

    // Verify redirect to milestones list page
    await expect(page).toHaveURL(/\/milestones$/, { timeout: 10000 });

    // Verify card is no longer in the list
    const deletedCard = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: `e2e-s8-del-${runId}` });
    await expect(deletedCard).not.toBeVisible({ timeout: 5000 });
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "non-deletable-status"
  test('TC-MML-S8-002: Executing milestone map does not show delete button', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKeyExecuting}`);

    // Delete button should not be visible
    await expect(page.locator('[data-testid="delete-map-btn"]')).not.toBeVisible();
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "unauthorized-api"
  test('TC-MML-S8-003: Delete milestone map API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.delete(`/v1/teams/${teamId}/milestone-maps/999999999999999999`);

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
