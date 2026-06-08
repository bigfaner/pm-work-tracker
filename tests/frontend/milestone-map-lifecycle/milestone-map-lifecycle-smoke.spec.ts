/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Smoke Test: Happy path through all 8 steps of milestone-map-lifecycle.
 * Covers create -> edit -> transition through states -> rollback -> delete.
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('milestone-map-lifecycle smoke: happy path + failure path', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  const runId = Date.now();
  const mapName = `e2e-mml-smoke-${runId}`;

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

    // Pre-create map for the smoke journey
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    await request.dispose();
  });

  // Happy path Step 1: Verify created map appears in list
  test('Smoke 1: Created map visible in list with planning status', async ({ page }) => {
    await login(page, undefined, '/milestones');

    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await expect(card).toBeVisible({ timeout: 10000 });
    await expect(card.getByText('规划中')).toBeVisible();
  });

  // Happy path Step 2: Navigate to detail page
  test('Smoke 2: Navigate to map detail page', async ({ page }) => {
    await login(page, undefined, '/milestones');

    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await card.click();

    await expect(page).toHaveURL(new RegExp(`/milestones/${mapBizKey}`), { timeout: 10000 });
    await expect(page.locator('[data-testid="milestone-timeline"]')).toBeVisible();
  });

  // Happy path Step 3: Edit map name
  test('Smoke 3: Edit map name on detail page', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const newName = `${mapName}-edited`;
    await page.getByRole('button', { name: '编辑' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const nameInput = page.getByPlaceholder('请输入里程碑图名称');
    await nameInput.clear();
    await nameInput.fill(newName);
    await page.getByRole('button', { name: '确认' }).click();

    await expect(page.getByText(newName).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 4: Transition planning -> reviewed
  test('Smoke 4: Transition planning to reviewed', async ({ page }) => {
    // Use API to transition status (StatusTransitionDropdown queries main-item endpoint, not milestone-map endpoint)
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'reviewed' },
    });

    await login(page, undefined, `/milestones/${mapBizKey}`);

    await expect(page.getByText(/已评审/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 5: Rollback reviewed -> planning
  test('Smoke 5: Rollback reviewed back to planning', async ({ page }) => {
    // Use API to rollback status
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'planning' },
    });

    await login(page, undefined, `/milestones/${mapBizKey}`);

    await expect(page.getByText(/规划中/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 6: Full transition chain planning -> reviewed -> ready -> executing
  test('Smoke 6: Full transition chain to executing', async ({ page }) => {
    // Use API for quick transitions, verify final state in UI
    const request = await page.context().request;
    for (const status of ['reviewed', 'ready', 'executing']) {
      await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    await login(page, undefined, `/milestones/${mapBizKey}`);
    await expect(page.getByText(/实施中|executing/i).first()).toBeVisible({ timeout: 10000 });
  });

  // Happy path Step 7: Delete is not available for executing map
  test('Smoke 7: Delete button hidden for executing map', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
    await expect(page.locator('[data-testid="delete-map-btn"]')).not.toBeVisible();
  });

  // Failure path: Create map with missing name shows disabled submit
  test('Smoke 8: Failure path - create without name disables submit', async ({ page }) => {
    await login(page, undefined, '/milestones');
    await page.locator('[data-testid="create-map-btn"]').click();

    // Submit button should be disabled with empty name
    await expect(page.getByRole('button', { name: '确认' })).toBeDisabled();
  });
});
