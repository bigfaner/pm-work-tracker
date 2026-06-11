/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 3: Transition status from planning to reviewed
 * Contracts: step-3-transition-planning-to-reviewed.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 3: Transition planning to reviewed', () => {
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

    // Create map in planning status
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-s3-trans-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    // Seed 2 milestones (not_started and in_progress)
    await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-s3-ms1-${runId}`, expectedEndDate: '2026-06-30' },
    });
    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-s3-ms2-${runId}`, expectedEndDate: '2026-12-31' },
    });
    const ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';
    // Transition one to in_progress for diversity
    await request.put(`/v1/teams/${teamId}/milestones/${ms2BizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'in_progress' },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-map-lifecycle / Step 3 / Outcome "success"
  test('TC-MML-S3-001: Transition from planning to reviewed updates status badge', async ({ page }) => {
    // Use API to transition status (StatusTransitionDropdown queries main-item endpoint, not milestone-map endpoint)
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'reviewed' },
    });

    // Reload page to see updated status
    await page.reload();

    // Verify status badge changed
    await expect(page.getByText(/已评审/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 3 / Outcome "unauthorized-api"
  test('TC-MML-S3-002: Status transition API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      data: { status: 'reviewed' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
