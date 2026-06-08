/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 3: Transition status from not_started to in_progress
 * Contracts: step-3-transition-to-in-progress.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 3: Transition not_started to in_progress', () => {
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

    // Create map + milestone in not_started status
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s3-map-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s3-ms-${runId}`, expectedEndDate: '2026-07-15' },
    });
    milestoneBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 3 / Outcome "success"
  test('TC-ML-S3-001: Transition from not_started to in_progress', async ({ page }) => {
    // Open detail panel for milestone
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    // Click status dropdown and select in_progress
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    const statusDropdown = panel.locator('button').filter({ hasText: /未开始/i }).first();
    await statusDropdown.click();

    await page.getByRole('menuitem', { name: /进行中/i }).click();

    // Verify status changed
    await expect(panel.getByText(/进行中/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-lifecycle / Step 3 / Outcome "unauthorized-api"
  test('TC-ML-S3-002: Milestone status transition API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      data: { status: 'in_progress' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
