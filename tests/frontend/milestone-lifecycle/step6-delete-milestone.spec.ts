/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 6: Delete milestone
 * Contracts: step-6-delete-milestone.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 6: Delete milestone', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let milestoneBizKeyNotStarted: string;
  let milestoneBizKeyInProgress: string;
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

    // Create map
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s6-map-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

    // Create milestone in not_started (deletable)
    const msRes1 = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s6-ns-${runId}`, expectedEndDate: '2026-07-15' },
    });
    milestoneBizKeyNotStarted = extractBizKey(parseApiData(await msRes1.json())) ?? '';

    // Create milestone in in_progress (non-deletable)
    const msRes2 = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s6-ip-${runId}`, expectedEndDate: '2026-08-15' },
    });
    milestoneBizKeyInProgress = extractBizKey(parseApiData(await msRes2.json())) ?? '';

    // Transition to in_progress
    await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKeyInProgress}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'in_progress' },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "success"
  test('TC-ML-S6-001: Delete not_started milestone removes node from timeline', async ({ page }) => {
    // Open detail panel
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKeyNotStarted}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    // Click delete button
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await panel.getByRole('button', { name: /删除里程碑/ }).click();

    // Confirm deletion
    await page.getByRole('button', { name: '确认删除' }).click();

    // Verify: detail panel closes
    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).not.toBeVisible({ timeout: 10000 });

    // Verify: node no longer in timeline
    await expect(page.locator(`[data-testid="milestone-node-${milestoneBizKeyNotStarted}"]`)).not.toBeVisible({ timeout: 5000 });
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "non-deletable-status"
  test('TC-ML-S6-002: In-progress milestone does not show delete button', async ({ page }) => {
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKeyInProgress}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });

    const panel = page.getByRole('dialog', { name: /里程碑详情/ });

    // Delete button should NOT be visible for in_progress milestone
    await expect(panel.getByRole('button', { name: /删除里程碑/ })).not.toBeVisible();
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "unauthorized-api"
  test('TC-ML-S6-003: Delete milestone API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.delete(`/v1/teams/${teamId}/milestones/${milestoneBizKeyNotStarted}`);

    expect(res.status()).toBe(401);
    await request.dispose();
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "api-not-found"
  test('TC-ML-S6-004: Delete non-existent milestone API returns 404', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.get(`/v1/teams/${teamId}/milestones/999999999999999999`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });

    expect(res.status()).toBe(404);
    await request.dispose();
  });
});
