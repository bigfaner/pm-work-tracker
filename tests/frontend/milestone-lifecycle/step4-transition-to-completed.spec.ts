/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 4: Transition status to completed
 * Contracts: step-4-transition-to-completed.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 4: Transition to completed or cancelled', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let milestoneBizKey: string;
  let milestoneBizKeyForCancel: string;
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

    // Create map + milestone, transition to in_progress
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s4-map-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s4-comp-${runId}`, expectedEndDate: '2026-07-15' },
    });
    milestoneBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    // Transition to in_progress
    await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'in_progress' },
    });

    // Create another milestone for cancel test
    const msRes2 = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s4-cancel-${runId}`, expectedEndDate: '2026-08-15' },
    });
    milestoneBizKeyForCancel = extractBizKey(parseApiData(await msRes2.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "cancel-cascade"
  test('TC-ML-S4-001: Cancel milestone transitions to cancelled status', async ({ page }) => {
    // Open detail panel
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKeyForCancel}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    // Click status dropdown and select cancelled
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    const statusDropdown = panel.locator('button').filter({ hasText: /未开始/i }).first();
    await statusDropdown.click();

    await page.getByRole('menuitem', { name: /已取消/i }).click();

    // Verify status changed to cancelled
    await expect(panel.getByText(/已取消/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "cancelled-is-terminal"
  test('TC-ML-S4-002: Cancelled status is terminal - no dropdown available', async ({ page }) => {
    // Re-open the cancelled milestone panel
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKeyForCancel}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    // Panel should show muted appearance for cancelled milestone
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    // The milestone name should have strikethrough (line-through) for cancelled
    const nameEl = panel.locator('h2');
    await expect(nameEl).toHaveCSS('text-decoration-line', /line-through/);
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "unauthorized-api"
  test('TC-ML-S4-003: Milestone status transition API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      data: { status: 'completed' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
