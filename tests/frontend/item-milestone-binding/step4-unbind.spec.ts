/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-milestone-binding
 *
 * Step 4-5 tests: Unbind MI and save with no changes
 * Step 4b: Unbind triggers completion recalculation
 * Step 2e: Server validation error
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('item-milestone-binding / Steps 4-5: Unbind and save unchanged', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
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
      data: { mapName: `e2e-imb-s4-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-imb-s4-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-imb-s4-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';

    // Bind MI to milestone
    await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/items');
  });

  // Step 4: Unbind MI from milestone
  test('TC-IMB-S4-001: Unbind MI from milestone clears milestone_key', async ({ page }) => {
    await page.getByText(`e2e-imb-s4-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    const dialog = page.getByRole('dialog');
    const milestoneTrigger = dialog.locator('[class*="trigger"], button').filter({ hasText: /e2e-imb-s4-ms/i }).first();
    await milestoneTrigger.click();
    await page.getByText(/未分配|unassigned/i).click();
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify unbind
    await page.waitForTimeout(1000);
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(verifyData.milestoneKey).toBeFalsy();

    // Re-bind for subsequent tests
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });
  });

  // Step 4b: Unbind triggers completion recalculation
  test('TC-IMB-S4-002: Unbind triggers milestone completion recalculation', async ({ page }) => {
    const request = page.context().request;

    // Get milestone completion before unbind
    const msBeforeRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${msBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const msBefore = parseApiData(await msBeforeRes.json());

    // Unbind via API
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: null },
    });

    // Get milestone completion after unbind
    const msAfterRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${msBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const msAfter = parseApiData(await msAfterRes.json());

    // Completion percentage should have changed (itemCount decreased)
    expect(msAfter.itemCount ?? msAfter.miCount).toBeLessThan(msBefore.itemCount ?? msBefore.miCount ?? 1);

    // Re-bind for subsequent tests
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });
  });

  // Step 5: Save with no changes preserves assignment
  test('TC-IMB-S5-001: Save without changes preserves milestone assignment', async ({ page }) => {
    await page.getByText(`e2e-imb-s4-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    // Save without modifying
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify milestone still bound
    await page.waitForTimeout(1000);
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(msBizKey);
  });

  // Step 2e: Server validation error (e.g. milestone deleted)
  test('TC-IMB-S2-005: Server validation error on bind to deleted milestone', async ({ page }) => {
    // Create and delete a milestone via API
    const request = page.context().request;
    const tempMsRes = await request.post(`http://127.0.0.1:8080/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneName: `e2e-imb-s4-tempms-${runId}`, expectedEndDate: '2026-09-30' },
    });
    const tempMsBizKey = extractBizKey(parseApiData(await tempMsRes.json())) ?? '';

    // Delete the milestone
    await request.delete(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${tempMsBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });

    // Try to bind MI to the deleted milestone
    const bindRes = await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: tempMsBizKey },
    });

    const bindData = await bindRes.json();
    expect(bindData.code === 0).toBeFalsy();
  });
});
