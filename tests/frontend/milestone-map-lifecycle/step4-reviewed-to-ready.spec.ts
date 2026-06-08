/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 4: Transition status from reviewed to ready
 * Contracts: step-4-transition-reviewed-to-ready.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 4: Transition reviewed to ready', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
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

    // Create map and transition to reviewed
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s4-trans-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'reviewed' },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-map-lifecycle / Step 4 / Outcome "success"
  test('TC-MML-S4-001: Transition from reviewed to ready updates status badge', async ({ page }) => {
    // Use API to transition status
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'ready' },
    });

    await page.reload();

    await expect(page.getByText(/待实施|ready/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 4 / Outcome "unauthorized-api"
  test('TC-MML-S4-002: Status transition API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      data: { status: 'ready' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
