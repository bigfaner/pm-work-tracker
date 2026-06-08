/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 2: Edit milestone map information
 * Contracts: step-2-edit-milestone-map.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 2: Edit milestone map information', () => {
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

    // Create a milestone map for editing
    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-edit-s2-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "success"
  test('TC-MML-S2-001: Edit milestone map name updates detail page', async ({ page }) => {
    const newName = `e2e-edited-${runId}`;

    // Click edit button
    await page.getByRole('button', { name: '编辑' }).click();

    // Clear and fill new name
    const nameInput = page.getByPlaceholder('请输入里程碑图名称');
    await nameInput.clear();
    await nameInput.fill(newName);

    // Submit
    await page.getByRole('button', { name: '确认' }).click();

    // Verify: breadcrumb and title show new name
    await expect(page.getByText(newName).first()).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "no-changes"
  test('TC-MML-S2-002: Edit with no changes closes dialog as no-op', async ({ page }) => {
    // Open edit dialog
    await page.getByRole('button', { name: '编辑' }).click();

    // Wait for dialog
    await expect(page.getByRole('dialog')).toBeVisible();

    // Click confirm without changes
    await page.getByRole('button', { name: '确认' }).click();

    // Dialog should close
    await expect(page.getByRole('dialog')).not.toBeVisible({ timeout: 5000 });
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "validation-error-invalid-date-range"
  test('TC-MML-S2-003: Edit with invalid date range shows validation error', async ({ page }) => {
    await page.getByRole('button', { name: '编辑' }).click();

    await expect(page.getByRole('dialog')).toBeVisible();

    // Set invalid date range
    await page.getByPlaceholder(/计划开始时间|开始/).first().fill('2026-12-31');
    await page.getByPlaceholder(/计划完成时间|结束/).last().fill('2026-01-01');

    await expect(page.getByText(/计划完成时间不得早于计划开始时间/)).toBeVisible();
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "unauthorized-api"
  test('TC-MML-S2-004: Edit milestone map API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}`, {
      data: { mapName: 'unauth-edit', assigneeBizKey: '1' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
