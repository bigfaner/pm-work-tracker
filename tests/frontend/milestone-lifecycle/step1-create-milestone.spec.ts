/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 1: Create milestone in timeline view
 * Contracts: step-1-create-milestone.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 1: Create milestone in timeline view', () => {
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

    // Create a milestone map for this journey
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s1-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "success"
  test('TC-ML-S1-001: Create milestone with valid data shows node in timeline', async ({ page }) => {
    const milestoneName = `e2e-ml-${runId}`;

    // Click create milestone button
    await page.locator('[data-testid="create-milestone-btn"]').click();

    // Fill form
    await page.getByPlaceholder('请输入里程碑名称').fill(milestoneName);
    await page.getByPlaceholder(/计划完成时间/).fill('2026-06-30');

    // Submit
    await page.getByRole('button', { name: '确认' }).click();

    // Verify: milestone node appears in timeline
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: milestoneName });
    await expect(node).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "validation-error-missing-name"
  test('TC-ML-S1-002: Create milestone without name - submit disabled', async ({ page }) => {
    await page.locator('[data-testid="create-milestone-btn"]').click();

    // Fill only date, leave name empty
    await page.getByPlaceholder(/计划完成时间/).fill('2026-06-30');

    // Submit should be disabled
    await expect(page.getByRole('button', { name: '确认' })).toBeDisabled();
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "validation-error-missing-date"
  test('TC-ML-S1-003: Create milestone without date - submit disabled', async ({ page }) => {
    await page.locator('[data-testid="create-milestone-btn"]').click();

    // Fill only name, leave date empty
    await page.getByPlaceholder('请输入里程碑名称').fill(`e2e-nodate-${runId}`);

    // Submit should be disabled
    await expect(page.getByRole('button', { name: '确认' })).toBeDisabled();
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "unauthorized-api"
  test('TC-ML-S1-004: Create milestone API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      data: { milestoneName: 'unauth-ms', expectedEndDate: '2026-06-30' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
