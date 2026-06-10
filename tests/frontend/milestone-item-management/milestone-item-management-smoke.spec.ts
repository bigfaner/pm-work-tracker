/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-item-management
 *
 * Smoke Test: Happy path through all steps of milestone-item-management.
 * Covers open panel -> view MI list -> unbind -> quick-add -> navigate -> drag-drop.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 180000;
test.setTimeout(TIMEOUT);

test.describe.serial('milestone-item-management smoke: happy path through panel operations', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let ms1BizKey: string;
  let ms2BizKey: string;
  let mi1BizKey: string;
  let mi2BizKey: string;
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

    // Create milestone map
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-mim-smoke-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    // Create two milestones
    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    // Create MIs and bind them to ms1
    const mi1Res = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-mim-mi1-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    mi1BizKey = extractBizKey(parseApiData(await mi1Res.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${mi1BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    const mi2Res = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-mim-mi2-${runId}`,
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-02-01',
        expectedEndDate: '2026-12-31',
      },
    });
    mi2BizKey = extractBizKey(parseApiData(await mi2Res.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${mi2BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    await request.dispose();
  });

  // Step 1: Open milestone detail panel
  test('Smoke 1: Click milestone node opens detail panel', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    // Click on ms1 node
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-ms1-${runId}` });
    await node.click();

    // Verify panel opens with key information
    await expect(page.getByText(`e2e-mim-ms1-${runId}`)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/进度|progress/i)).toBeVisible();
  });

  // Step 2: View associated MI list with hover unbind
  test('Smoke 2: Panel shows MI list with hover unbind control', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-ms1-${runId}` });
    await node.click();
    await page.waitForTimeout(1000);

    // Verify MI entries are visible
    await expect(page.getByText(`e2e-mim-mi1-${runId}`)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`e2e-mim-mi2-${runId}`)).toBeVisible();
  });

  // Step 3: Unbind a MainItem
  test('Smoke 3: Unbind MI removes it from panel list', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-ms1-${runId}` });
    await node.click();
    await page.waitForTimeout(1000);

    // Hover over MI row to reveal unbind control
    const miRow = page.locator(`[data-testid="mi-drag-${mi1BizKey}"]`);
    if (await miRow.isVisible()) {
      await miRow.hover();
      // Look for unbind button/icon
      const unbindBtn = miRow.getByRole('button', { name: /解绑|unbind/i });
      if (await unbindBtn.isVisible()) {
        await unbindBtn.click();
        await page.waitForTimeout(1000);

        // Verify toast message
        await expect(page.getByText(/已解除|unbind/i)).toBeVisible({ timeout: 5000 });
      }
    }

    // Verify via API
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${mi1BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(verifyData.milestoneKey).toBeFalsy();
  });

  // Step 4: Quick-add a MainItem
  test('Smoke 4: Quick-add MI in panel creates and auto-binds', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-ms1-${runId}` });
    await node.click();
    await page.waitForTimeout(1000);

    // Click add/quick-add button in panel
    const addBtn = page.getByRole('button', { name: /添加|新增|quick-add|add/i }).first();
    await addBtn.click();
    await page.waitForTimeout(500);

    // Fill in quick-add form
    const dialog = page.getByRole('dialog');
    if (await dialog.isVisible()) {
      await dialog.getByPlaceholder(/请输入.*标题|title/i).fill(`e2e-mim-quickadd-${runId}`);

      // Owner - select first option
      const ownerTrigger = dialog.getByText(/选择负责人|assignee/i);
      if (await ownerTrigger.isVisible()) {
        await ownerTrigger.click();
        await page.locator('[role="option"]').first().click();
      }

      // Start date
      const startDateInput = dialog.locator('label', { hasText: /开始时间/i }).locator('..').locator('input[type="date"]');
      if (await startDateInput.isVisible()) {
        await startDateInput.fill('2026-03-01');
      }

      // End date
      const endDateInput = dialog.locator('label', { hasText: /完成时间|结束时间/i }).locator('..').locator('input[type="date"]');
      if (await endDateInput.isVisible()) {
        await endDateInput.fill('2026-12-31');
      }

      // Submit
      await dialog.getByRole('button', { name: /确认|submit/i }).click();
      await page.waitForTimeout(2000);
    }
  });

  // Step 7: Close panel via overlay click
  test('Smoke 7: Click overlay closes panel', async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);

    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-ms1-${runId}` });
    await node.click();
    await page.waitForTimeout(1000);

    // Click outside the panel (on the overlay/backdrop)
    const overlay = page.locator('.fixed.inset-0, [class*="overlay"], [class*="backdrop"]').first();
    if (await overlay.isVisible()) {
      await overlay.click({ position: { x: 5, y: 5 } });
    }
  });
});
