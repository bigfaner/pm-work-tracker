/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-milestone-binding
 *
 * Smoke Test: Happy path through all steps of item-milestone-binding.
 * Covers open dialog -> bind -> rebind -> unbind -> save with no changes.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('item-milestone-binding smoke: happy path through bind/rebind/unbind', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let ms1BizKey: string;
  let ms2BizKey: string;
  let miBizKey: string;
  const runId = Date.now();

  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    authToken = await getAuthToken();

    // Get team
    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].bizKey);

    // Create a milestone map
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-imb-smoke-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    // Create two milestones
    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-imb-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-imb-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    // Create a MainItem (not in terminal state)
    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-imb-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';

    await request.dispose();
  });

  // Step 1: Open edit dialog and verify milestone selector
  test('Smoke 1: Open edit dialog shows milestone selector', async ({ page }) => {
    await login(page, undefined, '/items');

    // Find the MI and open its edit dialog
    const miCard = page.locator(`[data-testid^="mi-card-"], [data-testid^="main-item-"]`).filter({ hasText: `e2e-imb-mi-${runId}` });
    // Fallback: use text search if no specific data-testid
    await page.getByText(`e2e-imb-mi-${runId}`).first().click();

    // Wait for item detail or edit mode
    await page.waitForTimeout(1000);

    // Click edit button
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();

    // Verify dialog opens with milestone dropdown
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/里程碑|milestone/i)).toBeVisible();
  });

  // Step 2: Bind unassigned MI to milestone
  test('Smoke 2: Bind MI to milestone via edit dialog', async ({ page }) => {
    await login(page, undefined, '/items');

    // Use API to verify MI is unbound
    const request = page.context().request;
    const miRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const miData = parseApiData(await miRes.json());
    expect(miData.milestoneKey).toBeFalsy();

    // Navigate to the item and open edit dialog
    await page.goto(`http://127.0.0.1:8080/items`);
    await page.getByText(`e2e-imb-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    // Select milestone from dropdown
    const dialog = page.getByRole('dialog');
    await dialog.getByText(/未分配|unassigned/i).click();
    await page.getByText(`e2e-imb-ms1-${runId}`).click();

    // Submit
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify binding via API
    await page.waitForTimeout(1000);
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms1BizKey);
  });

  // Step 3: Rebind MI to a different milestone
  test('Smoke 3: Rebind MI to different milestone', async ({ page }) => {
    await login(page, undefined, '/items');

    await page.getByText(`e2e-imb-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    const dialog = page.getByRole('dialog');
    // Change to ms2
    await dialog.getByText(`e2e-imb-ms1-${runId}`).click();
    await page.getByText(`e2e-imb-ms2-${runId}`).click();
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify rebind via API
    await page.waitForTimeout(1000);
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms2BizKey);
  });

  // Step 4: Unbind MI from milestone
  test('Smoke 4: Unbind MI from milestone', async ({ page }) => {
    await login(page, undefined, '/items');

    await page.getByText(`e2e-imb-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    const dialog = page.getByRole('dialog');
    // Select "Unassigned"
    await dialog.getByText(`e2e-imb-ms2-${runId}`).click();
    await page.getByText(/未分配|unassigned/i).click();
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify unbind via API
    await page.waitForTimeout(1000);
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(verifyData.milestoneKey).toBeFalsy();
  });

  // Step 5: Save with no changes preserves assignment
  test('Smoke 5: Save with no changes preserves assignment', async ({ page }) => {
    // First bind via API
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    await login(page, undefined, '/items');

    await page.getByText(`e2e-imb-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    // Save without changing milestone
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify still bound to ms1
    await page.waitForTimeout(1000);
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms1BizKey);
  });
});
