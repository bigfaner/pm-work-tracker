/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 1: Create milestone map via list page
 * Contracts: step-1-create-milestone-map.md
 *
 * NOTE: Contract eval score below target (665/1000). Review with extra scrutiny.
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-map-lifecycle / Step 1: Create milestone map via list page', () => {
  let authToken: string;
  let teamId: string;
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

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, '/milestones');
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "success"
  test('TC-MML-S1-001: Create milestone map with valid data shows new card in list', async ({ page }) => {
    const mapName = `e2e-mml-s1-${runId}`;

    // Click create button
    await page.locator('[data-testid="create-map-btn"]').click();

    // Fill form
    await page.getByPlaceholder('请输入里程碑图名称').fill(mapName);

    // Select owner (first member)
    await page.getByText('选择负责人').click();
    await page.locator('[role="option"]').first().click();

    // Submit
    await page.getByRole('button', { name: '确认' }).click();

    // Verify: new card appears in list
    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await expect(card).toBeVisible({ timeout: 10000 });

    // Verify: status badge shows "规划中"
    await expect(card.getByText('规划中')).toBeVisible();
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "success" with dates
  test('TC-MML-S1-002: Create milestone map with plan dates', async ({ page }) => {
    const mapName = `e2e-mml-s1-dated-${runId}`;

    await page.locator('[data-testid="create-map-btn"]').click();
    await page.getByPlaceholder('请输入里程碑图名称').fill(mapName);

    // Select owner
    await page.getByText('选择负责人').click();
    await page.locator('[role="option"]').first().click();

    // Set plan dates (DateInput renders <input type="date">, locate by label text)
    const dialog = page.getByRole('dialog');
    await dialog.locator('label', { hasText: '计划开始时间' }).locator('..').locator('input[type="date"]').fill('2026-01-01');
    await dialog.locator('label', { hasText: '计划完成时间' }).locator('..').locator('input[type="date"]').fill('2026-12-31');

    await page.getByRole('button', { name: '确认' }).click();

    const card = page.locator(`[data-testid^="milestone-map-card-"]`).filter({ hasText: mapName });
    await expect(card).toBeVisible({ timeout: 10000 });
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-missing-name"
  test('TC-MML-S1-003: Create milestone map without name - submit button disabled', async ({ page }) => {
    await page.locator('[data-testid="create-map-btn"]').click();

    // Leave name empty, select owner
    await page.getByText('选择负责人').click();
    await page.locator('[role="option"]').first().click();

    // Submit button should be disabled since name is required
    const submitBtn = page.getByRole('button', { name: '确认' });
    await expect(submitBtn).toBeDisabled();
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-missing-owner"
  test('TC-MML-S1-004: Create milestone map without owner - submit button disabled', async ({ page }) => {
    await page.locator('[data-testid="create-map-btn"]').click();

    // Fill name only, leave owner empty
    await page.getByPlaceholder('请输入里程碑图名称').fill(`e2e-noowner-${runId}`);

    // Submit button should be disabled since owner is required
    const submitBtn = page.getByRole('button', { name: '确认' });
    await expect(submitBtn).toBeDisabled();
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-invalid-date-range"
  test('TC-MML-S1-005: Create milestone map with invalid date range shows validation error', async ({ page }) => {
    await page.locator('[data-testid="create-map-btn"]').click();

    await page.getByPlaceholder('请输入里程碑图名称').fill(`e2e-baddate-${runId}`);
    await page.getByText('选择负责人').click();
    await page.locator('[role="option"]').first().click();

    // Set invalid date range: end before start (DateInput renders <input type="date">)
    const dialog = page.getByRole('dialog');
    await dialog.locator('label', { hasText: '计划开始时间' }).locator('..').locator('input[type="date"]').fill('2026-12-31');
    await dialog.locator('label', { hasText: '计划完成时间' }).locator('..').locator('input[type="date"]').fill('2026-01-01');

    // Verify date validation error appears
    await expect(page.getByText(/计划完成时间不得早于计划开始时间/)).toBeVisible();
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "unauthorized-api"
  test('TC-MML-S1-006: Create milestone map API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      data: { mapName: 'unauth-map', assigneeBizKey: '1' },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
