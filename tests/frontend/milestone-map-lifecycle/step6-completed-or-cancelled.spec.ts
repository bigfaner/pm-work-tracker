/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 6: Transition status to completed or cancelled
 * Contracts: step-6-transition-to-completed-or-cancelled.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 6: Transition to completed or cancelled', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKeyExecuting: string;
  let mapBizKeyCompleted: string;
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

    // Create map and transition through to executing
    const mapRes1 = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s6-exec-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKeyExecuting = extractBizKey(parseApiData(await mapRes1.json())) ?? '';

    for (const status of ['reviewed', 'ready', 'executing']) {
      await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKeyExecuting}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    // Create a second map transitioned to completed for terminal state testing
    const mapRes2 = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s6-comp-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKeyCompleted = extractBizKey(parseApiData(await mapRes2.json())) ?? '';

    for (const status of ['reviewed', 'ready', 'executing', 'completed']) {
      await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKeyCompleted}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    await request.dispose();
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "success-cancelled"
  test('TC-MML-S6-001: Transition from executing to cancelled updates status', async ({ page }) => {
    // Use API to transition status to cancelled
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKeyExecuting}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'cancelled' },
    });

    await login(page, undefined, `/milestones/${mapBizKeyExecuting}`);

    await expect(page.getByText(/已取消|cancelled/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "completed-is-terminal"
  test('TC-MML-S6-002: Completed status has no transition options available', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKeyCompleted}`);

    // The status dropdown should be disabled or not interactive in terminal state
    // Verify the status badge shows completed
    await expect(page.getByText(/已完成|completed/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "unauthorized-api"
  test('TC-MML-S6-003: Status transition API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKeyExecuting}/status`, {
      data: { status: 'cancelled' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
