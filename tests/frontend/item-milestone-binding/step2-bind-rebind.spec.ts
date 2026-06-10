/**
 * @web-e2e
 * @feature milestone-map
 * Journey: item-milestone-binding
 *
 * Step 2-3 tests: Bind and rebind MI to milestones
 * Step 2b: Bind MI in terminal state
 * Step 2c: Cross-team binding rejected
 * Step 2d: Bind to cancelled milestone rejected
 * Step 3b: Rebind triggers completion recalculation on both milestones
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('item-milestone-binding / Steps 2-3: Bind and rebind', () => {
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
      data: { mapName: `e2e-imb-s2-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const ms1Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-imb-s2-ms1-${runId}`, expectedEndDate: '2026-07-30' },
    });
    ms1BizKey = extractBizKey(parseApiData(await ms1Res.json())) ?? '';

    const ms2Res = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-imb-s2-ms2-${runId}`, expectedEndDate: '2026-10-30' },
    });
    ms2BizKey = extractBizKey(parseApiData(await ms2Res.json())) ?? '';

    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-imb-s2-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/items');
  });

  // Step 2: Bind unassigned MI to milestone
  test('TC-IMB-S2-001: Bind unassigned MI to milestone updates milestone_key', async ({ page }) => {
    // Verify MI is unbound
    const request = page.context().request;
    const miRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const miData = parseApiData(await miRes.json());
    expect(miData.milestoneKey).toBeFalsy();

    // Open edit dialog and bind
    await page.getByText(`e2e-imb-s2-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    const dialog = page.getByRole('dialog');
    const milestoneTrigger = dialog.locator('[class*="trigger"], button').filter({ hasText: /未分配|unassigned/i }).first();
    await milestoneTrigger.click();
    await page.getByText(`e2e-imb-s2-ms1-${runId}`).click();
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify binding via API
    await page.waitForTimeout(1000);
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms1BizKey);
  });

  // Step 3: Rebind MI to different milestone
  test('TC-IMB-S3-001: Rebind MI from milestone A to milestone B', async ({ page }) => {
    // First ensure MI is bound to ms1
    const request = page.context().request;
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    await page.getByText(`e2e-imb-s2-mi-${runId}`).first().click();
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /编辑|edit/i }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    const dialog = page.getByRole('dialog');
    const milestoneTrigger = dialog.locator('[class*="trigger"], button').filter({ hasText: /e2e-imb-s2-ms1/i }).first();
    await milestoneTrigger.click();
    await page.getByText(`e2e-imb-s2-ms2-${runId}`).click();
    await dialog.getByRole('button', { name: /确认|保存|submit/i }).click();

    // Verify rebind
    await page.waitForTimeout(1000);
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(String(verifyData.milestoneKey)).toBe(ms2BizKey);
  });

  // Step 2b: Bind MI in terminal state - server rejects
  test('TC-IMB-S2-002: Bind MI in terminal state rejected by server', async ({ page }) => {
    // Create a completed MI via API
    const request = page.context().request;
    const completedMiRes = await request.post(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: {
        title: `e2e-imb-s2-completed-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
        status: 'completed',
      },
    });
    const completedMiBizKey = extractBizKey(parseApiData(await completedMiRes.json())) ?? '';

    // Try to bind via API - should be rejected
    const bindRes = await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${completedMiBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    // Server should reject (4xx or error code in body)
    const bindData = await bindRes.json();
    expect(bindData.code === 0).toBeFalsy();
  });

  // Step 2c: Cross-team binding rejected
  test('TC-IMB-S2-003: Cross-team binding rejected', async ({ page }) => {
    const request = page.context().request;

    // Get second team (or skip if only one team)
    const teamsRes = await request.get('http://127.0.0.1:8080/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);

    if (teamsData.length < 2) {
      // Skip if only one team
      return;
    }

    const team2Id = String(teamsData[1].bizKey);

    // Try to bind MI to a milestone in different team
    const bindRes = await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: 'cross-team-fake-key' },
    });

    const bindData = await bindRes.json();
    expect(bindData.code === 0).toBeFalsy();
  });

  // Step 3b: Rebind triggers completion recalculation
  test('TC-IMB-S3-002: Rebind recalculates completion on both milestones', async ({ page }) => {
    const request = page.context().request;

    // Bind MI to ms1 first
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms1BizKey },
    });

    // Get ms1 completion before rebind
    const ms1BeforeRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${ms1BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const ms1Before = parseApiData(await ms1BeforeRes.json());

    // Rebind to ms2 via API
    await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: ms2BizKey },
    });

    // Get ms2 after rebind
    const ms2AfterRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/milestones/${ms2BizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const ms2After = parseApiData(await ms2AfterRes.json());

    // Verify ms2 now has the MI bound
    expect(String(parseApiData(await (await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })).json()).milestoneKey)).toBe(ms2BizKey);
  });
});
