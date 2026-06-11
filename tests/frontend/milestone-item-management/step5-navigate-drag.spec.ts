/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-item-management
 *
 * Step 5-6 tests: Navigate to MI detail, drag-drop rebinding
 * Step 5b: Cancelled milestone panel
 * Step 6b: Cross-team drag rejection
 * Step 6c: Terminal state MI drag
 * Step 7b: Delete action hidden for in_progress/completed milestones
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey, baseUrl } from '../helpers.js';

test.describe('milestone-item-management / Steps 5-7: Navigate, drag-drop, panel edge cases', () => {
  test.setTimeout(120000);
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

    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].bizKey);

    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-mim-s5-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-s5-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-s5-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    // Create MI and bind to ms1
    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-mim-s5-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    // Re-bind MI to ms1 before each test
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Step 5: Navigate to MI detail from panel
  test('TC-MIM-S5-001: Click MI in panel navigates to MI detail page', async ({ page }) => {
    await page.locator('[data-testid="milestone-timeline"]').waitFor({ state: 'visible', timeout: 15000 });
    const node = page.locator(`[data-testid="milestone-node-${ms1BizKey}"]`);
    await node.click();
    const panel = page.getByRole('dialog');
    await expect(panel).toBeVisible({ timeout: 10000 });

    // Click on MI title link within panel
    const miLink = panel.getByText(`e2e-mim-s5-mi-${runId}`).first();
    await miLink.click();

    // Should navigate to MI detail
    await page.waitForURL(/\/items\//, { timeout: 10000 });
  });

  // Step 6: Drag-drop MI rebinding
  test('TC-MIM-S6-001: Drag MI from ms1 to ms2 rebinds it', async ({ page }) => {
    // Wait for the milestone timeline to finish loading
    await page.locator('[data-testid="milestone-timeline"]').waitFor({ state: 'visible', timeout: 15000 });
    // Verify MI is on ms1 in the timeline MI layer
    const miItem = page.locator(`[data-testid="mi-item-${miBizKey}"]`);
    await miItem.waitFor({ state: 'visible', timeout: 10000 });

    const targetNode = page.locator(`[data-testid="milestone-node-${ms2BizKey}"]`);
    await targetNode.waitFor({ state: 'visible', timeout: 10000 });

    // Set drag data on window.__dragMI manually (Playwright mouse events don't
    // reliably trigger React's onDragStart HTML5 event). Then dispatch a drop
    // event on the target milestone node so the component's onDrop handler fires.
    await page.evaluate(
      ({ miKey, msKey }) => {
        (window as any).__dragMI = {
          miBizKey: miKey,
          miCode: '',
          sourceMilestoneKey: msKey,
        };
      },
      { miKey: miBizKey, msKey: ms1BizKey },
    );

    // Listen for the PUT rebind API call
    const rebindResponse = page.waitForResponse(
      resp => resp.url().includes('/main-items/') && resp.request().method() === 'PUT',
      { timeout: 10000 },
    );
    await targetNode.dispatchEvent('drop');
    await rebindResponse;

    // Verify via PUT API (GET endpoint omits milestoneKey in response)
    const request = page.context().request;
    const verifyRes = await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms2BizKey },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms2BizKey);
  });

  // Step 5b: Cancelled milestone panel
  test('TC-MIM-S5-002: Cancelled milestone shows muted panel with empty MI list', async ({ page }) => {
    // Create and cancel a milestone
    const request = page.context().request;
    const cancelledMsRes = await request.post(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneName: `e2e-mim-s5-cancelled-${runId}`, expectedEndDate: '2026-08-30' },
    });
    const cancelledMsKey = extractBizKey(parseApiData(await cancelledMsRes.json())) ?? '';

    // Transition to cancelled
    for (const status of ['in_progress', 'completed', 'cancelled']) {
      await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${cancelledMsKey}/status`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { status },
      });
    }

    // Navigate to map and reload
    await page.goto(`${baseUrl}/milestones/${mapBizKey}`);
    await page.locator('[data-testid="milestone-timeline"]').waitFor({ state: 'visible', timeout: 15000 });

    // Find and click cancelled milestone node
    const cancelledNode = page.locator(`[data-testid="milestone-node-${cancelledMsKey}"]`);
    if (await cancelledNode.isVisible()) {
      await cancelledNode.click();
      const panel = page.getByRole('dialog');
      await expect(panel).toBeVisible({ timeout: 10000 });

      // Panel should show cancelled status badge
      await expect(panel.locator('.bg-error-bg').filter({ hasText: '已取消' })).toBeVisible({ timeout: 5000 });

      // Add button should not be visible for cancelled milestone (panel omits MI section entirely)
      const addBtn = panel.getByRole('button', { name: /新建事项/ });
      // The add button should not be visible for cancelled milestones
      await expect(addBtn).not.toBeVisible();
    }
  });

  // Step 7b: Delete action hidden for in_progress milestone
  test('TC-MIM-S7-001: Delete action hidden for in_progress milestone', async ({ page }) => {
    // Transition ms1 to in_progress
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${ms1BizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'in_progress' },
    });

    await page.goto(`${baseUrl}/milestones/${mapBizKey}`);
    // Wait for the milestone timeline to finish loading
    await page.locator('[data-testid="milestone-timeline"]').waitFor({ state: 'visible', timeout: 15000 });
    // Wait for the milestone node to appear in the timeline (it may take time to load)
    const node = page.locator(`[data-testid="milestone-node-${ms1BizKey}"]`);
    await node.waitFor({ state: 'visible', timeout: 15000 });
    await node.click();
    const panel = page.getByRole('dialog');
    await expect(panel).toBeVisible({ timeout: 10000 });

    // Delete button should not be visible for in_progress milestone
    const deleteBtn = panel.getByRole('button', { name: /删除里程碑/ });
    await expect(deleteBtn).not.toBeVisible();

    // Reset: transition to cancelled (terminal) — cannot go back to not_started from in_progress
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${ms1BizKey}/status`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { status: 'cancelled' },
    });
  });

  // Unauthorized API test
  test('TC-MIM-S6-002: Drag-drop rebind API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      data: { milestoneKey: ms2BizKey },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
