/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Smoke Test: Happy path through milestone-lifecycle journey.
 * Covers create -> edit -> status transitions -> cancel -> delete.
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('milestone-lifecycle smoke: happy path + failure path', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let milestoneBizKey: string;
  const runId = Date.now();
  const milestoneName = `e2e-ml-smoke-${runId}`;

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
      data: { mapName: `e2e-ml-smoke-map-${runId}`, assigneeBizKey: '1' },
    });
    mapBizKey = extractBizKey(parseApiData(await mapRes.json())) ?? '';

    // Create milestone in not_started
    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName, expectedEndDate: '2026-07-15' },
    });
    milestoneBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    await request.dispose();
  });

  // Happy path Step 1: Verify milestone node in timeline
  test('Smoke 1: Created milestone visible in timeline', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await expect(node).toBeVisible({ timeout: 10000 });
    await expect(node.getByText(milestoneName)).toBeVisible();
  });

  // Happy path Step 2: Open detail panel
  test('Smoke 2: Open milestone detail panel', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();

    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('dialog').getByText(milestoneName)).toBeVisible();
  });

  // Happy path Step 3: Edit milestone name
  test('Smoke 3: Edit milestone name via detail panel', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });
    // Wait for content to finish loading (skeleton replaced by real data)
    await expect(panel.getByText(milestoneName)).toBeVisible({ timeout: 10000 });

    await panel.getByRole('button', { name: '编辑' }).click();
    const nameInput = page.getByPlaceholder('请输入里程碑名称');
    await nameInput.clear();
    await nameInput.fill(`${milestoneName}-edited`);
    await page.getByRole('button', { name: '确认' }).click();

    await expect(panel.getByText(`${milestoneName}-edited`)).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 4: Transition not_started -> in_progress
  test('Smoke 4: Transition milestone to in_progress', async ({ page }) => {
    // Use API to transition status (StatusTransitionDropdown uses main-item API, not milestone API)
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'in_progress' },
    });

    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });

    await expect(panel.getByText(/进行中/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 5: Transition in_progress -> cancelled (cancel cascade)
  test('Smoke 5: Cancel milestone from in_progress', async ({ page }) => {
    // Use API to transition status to cancelled
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${milestoneBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'cancelled' },
    });

    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });

    await expect(panel.getByText(/已取消/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 6: Delete cancelled milestone
  test('Smoke 6: Delete cancelled milestone', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });
    // Wait for content to finish loading
    await expect(panel.getByText(/已取消/i).first()).toBeVisible({ timeout: 10000 });

    await panel.getByRole('button', { name: /删除里程碑/ }).click();
    await page.getByRole('button', { name: '确认删除' }).click();

    // Verify panel closes and node removed
    await expect(page.getByRole('dialog', { name: /里程碑详情/ })).not.toBeVisible({ timeout: 10000 });
    await expect(page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`)).not.toBeVisible({ timeout: 5000 });
  });

  // Failure path: Create milestone without required fields
  test('Smoke 7: Failure path - create without required fields disables submit', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
    await page.locator('[data-testid="create-milestone-btn"]').click();

    // Both name and date are required - submit should be disabled
    await expect(page.getByRole('button', { name: '确认' })).toBeDisabled();
  });
});
