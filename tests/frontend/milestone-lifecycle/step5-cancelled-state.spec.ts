/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 5: Cancelled state interactions
 * Contracts: step-5-cancelled-state.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 5: Cancelled state interactions', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let milestoneBizKey: string;
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

    // Create map + milestone, then cancel it
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s5-map-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s5-cancelled-${runId}`, expectedEndDate: '2026-07-15' },
    });
    milestoneBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    // Cancel the milestone
    await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'cancelled' },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 5 / Outcome "panel-muted-appearance"
  test('TC-ML-S5-001: Cancelled milestone panel shows muted appearance', async ({ page }) => {
    // Open detail panel for cancelled milestone
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();

    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });
    // Wait for content to load
    await expect(panel.locator('h2')).toBeVisible({ timeout: 10000 });

    // Name should have strikethrough
    const nameEl = panel.locator('h2');
    await expect(nameEl).toHaveCSS('text-decoration-line', 'line-through');

    // Related MI section should be hidden for cancelled milestones
    await expect(panel.getByText(/关联事项/)).not.toBeVisible();
  });

  // Traceability: milestone-lifecycle / Step 5 / Outcome "panel-muted-appearance" - delete visible
  test('TC-ML-S5-002: Cancelled milestone shows delete button', async ({ page }) => {
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    const panel = page.getByRole('dialog', { name: /里程碑详情/ });

    // Delete button should be visible for cancelled milestone
    await expect(panel.getByRole('button', { name: /删除里程碑/ })).toBeVisible();
  });

  // Traceability: milestone-lifecycle / Step 5 / Outcome "unauthorized-api"
  test('TC-ML-S5-003: Cancelled milestone API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/milestones/${milestoneBizKey}`);

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
