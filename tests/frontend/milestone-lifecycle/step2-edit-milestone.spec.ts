/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 2: Edit milestone information
 * Contracts: step-2-edit-milestone.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-lifecycle / Step 2: Edit milestone information', () => {
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

    // Create map + milestone
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-ml-s2-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-ml-s2-ms-${runId}`, expectedEndDate: '2026-07-15' },
    });
    const msData = parseApiData(await msRes.json());
    milestoneBizKey = extractBizKey(msData) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "success"
  test('TC-ML-S2-001: Edit milestone name updates detail panel', async ({ page }) => {
    // Click on the milestone node to open detail panel
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();

    // Wait for detail panel to open and content to load
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });
    // Wait for content to finish loading (skeleton replaced by real data)
    await expect(panel.locator('h2')).toBeVisible({ timeout: 10000 });

    // Click edit button in the panel
    await panel.getByRole('button', { name: '编辑里程碑' }).click();

    // Edit name in dialog
    const nameInput = page.getByPlaceholder('请输入里程碑名称');
    await nameInput.clear();
    await nameInput.fill(`e2e-edited-ml-${runId}`);

    await page.getByRole('button', { name: '确认' }).click();

    // Verify updated name in panel
    await expect(panel.getByText(`e2e-edited-ml-${runId}`)).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "no-changes"
  test('TC-ML-S2-002: Edit with no changes closes dialog', async ({ page }) => {
    const node = page.locator(`[data-testid="milestone-node-${milestoneBizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog', { name: /里程碑详情/ });
    await expect(panel).toBeVisible({ timeout: 10000 });
    // Wait for content to finish loading
    await expect(panel.locator('h2')).toBeVisible({ timeout: 10000 });

    await panel.getByRole('button', { name: '编辑里程碑' }).click();

    // Dialog opens, click confirm without changes
    await page.getByRole('button', { name: '确认' }).click();

    // Dialog should close
    const editDialogTitle = page.getByRole('dialog').getByText('编辑里程碑');
    await expect(editDialogTitle).not.toBeVisible({ timeout: 5000 });
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "unauthorized-api"
  test('TC-ML-S2-003: Edit milestone API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestones/${milestoneBizKey}`, {
      data: { milestoneName: 'unauth-edit' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
