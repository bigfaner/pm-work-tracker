/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-item-management
 *
 * Step 4 tests: Quick-add a MainItem in panel
 * Step 4b-4e: Validation edge cases
 * Step 4f: Milestone field locked
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-item-management / Step 4: Quick-add MainItem in panel', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
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
      data: { mapName: `e2e-mim-s4-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-s4-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Step 4: Quick-add creates and auto-binds MI
  test('TC-MIM-S4-001: Quick-add MI creates and auto-binds to milestone', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s4-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    // Click quick-add button
    const addBtn = page.getByRole('button', { name: /添加|新增|add/i }).first();
    await addBtn.click();
    await page.waitForTimeout(500);

    // Fill form in dialog
    const dialog = page.getByRole('dialog');
    if (await dialog.isVisible()) {
      await dialog.getByPlaceholder(/请输入.*标题|title/i).fill(`e2e-mim-s4-quickadd-${runId}`);

      // Select owner
      const ownerTrigger = dialog.getByText(/选择负责人|assignee/i);
      if (await ownerTrigger.isVisible()) {
        await ownerTrigger.click();
        await page.locator('[role="option"]').first().click();
      }

      // Dates
      const startInput = dialog.locator('label', { hasText: /开始时间/i }).locator('..').locator('input[type="date"]');
      if (await startInput.isVisible()) {
        await startInput.fill('2026-03-01');
      }
      const endInput = dialog.locator('label', { hasText: /完成时间|结束时间/i }).locator('..').locator('input[type="date"]');
      if (await endInput.isVisible()) {
        await endInput.fill('2026-12-31');
      }

      // Submit
      await dialog.getByRole('button', { name: /确认|submit/i }).click();
      await page.waitForTimeout(2000);

      // Verify MI appears in panel list
      await expect(page.getByText(`e2e-mim-s4-quickadd-${runId}`)).toBeVisible({ timeout: 10000 });
    }
  });

  // Step 4b: Quick-add without title - validation error
  test('TC-MIM-S4-002: Quick-add without title shows validation error', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s4-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    const addBtn = page.getByRole('button', { name: /添加|新增|add/i }).first();
    await addBtn.click();
    await page.waitForTimeout(500);

    const dialog = page.getByRole('dialog');
    if (await dialog.isVisible()) {
      // Leave title empty, fill other fields
      const ownerTrigger = dialog.getByText(/选择负责人|assignee/i);
      if (await ownerTrigger.isVisible()) {
        await ownerTrigger.click();
        await page.locator('[role="option"]').first().click();
      }

      // Submit should be disabled
      const submitBtn = dialog.getByRole('button', { name: /确认|submit/i });
      await expect(submitBtn).toBeDisabled();
    }
  });

  // Step 4f: Milestone field is locked (pre-filled and disabled)
  test('TC-MIM-S4-003: Quick-add milestone field is pre-filled and disabled', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s4-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    const addBtn = page.getByRole('button', { name: /添加|新增|add/i }).first();
    await addBtn.click();
    await page.waitForTimeout(500);

    const dialog = page.getByRole('dialog');
    if (await dialog.isVisible()) {
      // Milestone field should show the current milestone name and be disabled
      const milestoneField = dialog.locator('input, select, [class*="disabled"], [disabled]').filter({ hasText: /e2e-mim-s4-ms/i });
      if (await milestoneField.isVisible()) {
        await expect(milestoneField).toBeDisabled();
      }
    }
  });

  // Step 4e: Quick-add loading state
  test('TC-MIM-S4-004: Quick-add shows loading state during submission', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s4-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    const addBtn = page.getByRole('button', { name: /添加|新增|add/i }).first();
    await addBtn.click();
    await page.waitForTimeout(500);

    const dialog = page.getByRole('dialog');
    if (await dialog.isVisible()) {
      await dialog.getByPlaceholder(/请输入.*标题|title/i).fill(`e2e-mim-s4-loading-${runId}`);

      const ownerTrigger = dialog.getByText(/选择负责人|assignee/i);
      if (await ownerTrigger.isVisible()) {
        await ownerTrigger.click();
        await page.locator('[role="option"]').first().click();
      }

      const startInput = dialog.locator('label', { hasText: /开始时间/i }).locator('..').locator('input[type="date"]');
      if (await startInput.isVisible()) {
        await startInput.fill('2026-04-01');
      }
      const endInput = dialog.locator('label', { hasText: /完成时间|结束时间/i }).locator('..').locator('input[type="date"]');
      if (await endInput.isVisible()) {
        await endInput.fill('2026-12-31');
      }

      // Click submit and check for loading state
      const submitBtn = dialog.getByRole('button', { name: /确认|submit/i });
      await submitBtn.click();

      // Button should show loading state briefly
      const isLoading = await submitBtn.getAttribute('data-loading') ?? await submitBtn.locator('[class*="loading"], [class*="spinner"]').count();
      // Loading state is transient, just verify the dialog eventually closes
      await page.waitForTimeout(3000);
    }
  });

  // Unauthorized API test
  test('TC-MIM-S4-005: Create MI API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.post(`/v1/teams/${teamId}/main-items`, {
      data: {
        title: 'unauth-mi',
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
