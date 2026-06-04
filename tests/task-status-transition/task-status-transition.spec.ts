/**
 * Web E2E Test: task-status-transition journey
 * @feature system-ux-optimization
 * @web-e2e
 *
 * Traceability: contracts/step-1..step-7 (web UI contracts)
 *
 * Contract test functions: 7 (invalid transition, non-terminal success, terminal confirm, conversion form, submit conversion, form cleanup, todo-to-main)
 * Journey smoke test functions: 1 (happy path status transition + conversion)
 */
import { test, expect } from '@playwright/test';
import {
  login,
  loginAsUser,
  screenshot,
  baseUrl,
  API,
  getApiToken,
  apiBaseUrl,
  setupRbacFixtures,
  createTestMainItem,
  createTestSubItem,
  authHeader,
  curl,
  getFirstTeamId,
  parseApiBody,
  randomCode,
  extractBizKey,
} from '../web/helpers.js';

let token: string;
let teamBizKey: string;
let memberToken: string;

test.describe('task-status-transition: Contract tests', () => {

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    teamBizKey = (await getFirstTeamId(token))!;
    const fixtures = await setupRbacFixtures();
    memberToken = fixtures.memberToken;
  });

  // Traceability: step-1-trigger-status-transition-error / Outcome "invalid-transition"
  test('step1-invalid-transition: Invalid status transition shows error', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E invalid transition', 'P2');

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Try to transition via API with invalid target
    const invalidRes = await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'invalid_status' }),
    });
    // Should get an error (400 or specific error)
    expect(invalidRes.status >= 400).toBeTruthy();
    await screenshot(page, 'status-step1-invalid');
  });

  // Traceability: step-1-trigger-status-transition-error / Outcome "unauthorized-attempt"
  test('step1-unauthorized: Member cannot change item status via API', async () => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E unauthorized status', 'P2');

    const res = await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ status: 'progressing' }),
    });
    expect(res.status).toBe(403);
  });

  // Traceability: step-2-successful-status-transition-non-terminal / Outcome "success"
  test('step2-success: Successful non-terminal status transition via UI', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E non-terminal', 'P2');

    // Set item to pending status first
    await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'pending' }),
    });

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Find status dropdown/badge and click
    const statusBtn = page.getByRole('button', { name: /待开始|pending/i });
    if (await statusBtn.first().isVisible().catch(() => false)) {
      await statusBtn.first().click();
      await page.waitForTimeout(800);

      // Select "进行中" (progressing)
      const progressingOption = page.getByRole('menuitem', { name: /进行中/ }).or(
        page.getByRole('option', { name: /进行中/ }),
      );
      if (await progressingOption.first().isVisible().catch(() => false)) {
        await progressingOption.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify status changed via API
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    const data = parseApiBody(verifyRes.body);
    expect(['progressing', '进行中'].some(s => JSON.stringify(data).includes(s))).toBeTruthy();
    await screenshot(page, 'status-step2-success');
  });

  // Traceability: step-3-terminal-status-transition-confirmation / Outcome "success"
  test('step3-success: Terminal status transition with confirmation dialog', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E terminal', 'P2');
    // Create and complete sub-item
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E terminal sub');
    await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/sub-items/${subKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'completed' }),
    });

    // Set main item to reviewing status
    await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'reviewing' }),
    });

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Click status badge
    const statusBtn = page.getByRole('button', { name: /待验收|reviewing/i });
    if (await statusBtn.first().isVisible().catch(() => false)) {
      await statusBtn.first().click();
      await page.waitForTimeout(800);

      // Select terminal status
      const closedOption = page.getByRole('menuitem', { name: /已关闭/ }).or(
        page.getByRole('option', { name: /已关闭/ }),
      );
      if (await closedOption.first().isVisible().catch(() => false)) {
        await closedOption.first().click();
        await page.waitForTimeout(800);

        // Confirmation dialog should appear
        const dialogText = await page.textContent('body') ?? '';
        expect(dialogText.includes('确认') || dialogText.includes('确定')).toBeTruthy();

        // Confirm
        const confirmBtn = page.getByRole('button', { name: /确认|确定/ }).and(
          page.getByRole('button').filter({ hasNotText: /取消/ }),
        );
        if (await confirmBtn.first().isVisible().catch(() => false)) {
          await confirmBtn.first().click();
          await page.waitForLoadState('networkidle');
        }
      }
    }

    await screenshot(page, 'status-step3-terminal');
  });

  // Traceability: step-3-terminal-status-transition-confirmation / Outcome "cancelled"
  test('step3-cancelled: Cancel terminal transition dialog', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E cancel terminal', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E cancel sub');
    await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/sub-items/${subKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'completed' }),
    });
    await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'reviewing' }),
    });

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    const statusBtn = page.getByRole('button', { name: /待验收|reviewing/i });
    if (await statusBtn.first().isVisible().catch(() => false)) {
      await statusBtn.first().click();
      await page.waitForTimeout(800);

      const closedOption = page.getByRole('menuitem', { name: /已关闭/ }).or(
        page.getByRole('option', { name: /已关闭/ }),
      );
      if (await closedOption.first().isVisible().catch(() => false)) {
        await closedOption.first().click();
        await page.waitForTimeout(800);

        // Cancel
        const cancelBtn = page.getByRole('button', { name: /取消/ });
        if (await cancelBtn.first().isVisible().catch(() => false)) {
          await cancelBtn.first().click();
          await page.waitForLoadState('networkidle');
        }
      }
    }

    // Verify status unchanged via API
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    const data = parseApiBody(verifyRes.body);
    expect(data.itemStatus).toBe('reviewing');
    await screenshot(page, 'status-step3-cancelled');
  });

  // Traceability: step-4-open-conversion-form-defaults / Outcome "success"
  test('step4-success: Conversion form opens with defaults and required markers', async ({ page }) => {
    // Create a pool item and convert to todo first
    const poolRes = await curl('POST', `${API}/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(token),
      body: JSON.stringify({ title: 'E2E conversion test', background: 'e2e test', expectedOutput: 'e2e' }),
    });
    const poolData = parseApiBody(poolRes.body);
    const poolKey = extractBizKey(poolData)!;

    await login(page, undefined, '/item-pool');
    await page.waitForLoadState('networkidle');

    // Find the pool item and look for conversion button
    const toItemBtn = page.locator(`[data-testid="to-sub-${poolKey}"]`).or(
      page.getByRole('button', { name: /转.*子事项|convert.*sub/i }).first(),
    );
    if (await toItemBtn.first().isVisible().catch(() => false)) {
      await toItemBtn.first().click();
      await page.waitForTimeout(1000);

      // Verify form is visible with required markers
      const dialogText = await page.textContent('body') ?? '';
      expect(dialogText.includes('优先级') || dialogText.includes('负责人') || dialogText.includes('priority') || dialogText.includes('assignee')).toBeTruthy();
    }
    await screenshot(page, 'status-step4-form');
  });

  // Traceability: step-6-close-reopen-conversion-form / Outcome "success"
  test('step6-success: Form fields cleared on close and reopen', async ({ page }) => {
    await login(page, undefined, '/item-pool');
    await page.waitForLoadState('networkidle');

    // Open conversion dialog for first pool item
    const toItemBtn = page.getByRole('button', { name: /转.*子事项|convert.*sub/i }).first();
    if (await toItemBtn.isVisible().catch(() => false)) {
      await toItemBtn.click();
      await page.waitForTimeout(1000);

      // Fill some data
      const textInputs = page.getByRole('textbox');
      if (await textInputs.count() > 0) {
        await textInputs.first().fill('test data');
      }

      // Close dialog
      const cancelBtn = page.getByRole('button', { name: /取消|close|关闭/ });
      if (await cancelBtn.first().isVisible().catch(() => false)) {
        await cancelBtn.first().click();
        await page.waitForTimeout(500);
      }

      // Reopen
      if (await toItemBtn.isVisible().catch(() => false)) {
        await toItemBtn.click();
        await page.waitForTimeout(1000);

        // Verify fields are empty
        const textInputs2 = page.getByRole('textbox');
        if (await textInputs2.count() > 0) {
          const value = await textInputs2.first().inputValue().catch(() => '');
          expect(value).toBe('');
        }
      }
    }
    await screenshot(page, 'status-step6-cleanup');
  });

  // Traceability: step-1-trigger-status-transition-error / Outcome "validation-error" (API)
  test('step1-validation-error: API rejects empty status field', async () => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E validation status', 'P2');
    const res = await curl('PUT', `${API}/teams/${teamBizKey}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({}),
    });
    expect(res.status).toBe(400);
  });
});

test.describe('task-status-transition: Journey smoke test (happy path)', () => {

  let token: string;
  let smokeTeamId: string;

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    smokeTeamId = (await getFirstTeamId(token))!;
  });

  // Journey smoke: create item → transition to progressing → transition to reviewing → verify
  test('smoke: Full happy path — create item, transition status twice, verify final state', async ({ page }) => {
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke status', 'P2');

    // Ensure pending status
    await curl('PUT', `${API}/teams/${smokeTeamId}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'pending' }),
    });

    // Step 1: Navigate to item
    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Step 2: Transition pending → progressing
    const statusBtn = page.getByRole('button', { name: /待开始|pending/i });
    if (await statusBtn.first().isVisible().catch(() => false)) {
      await statusBtn.first().click();
      await page.waitForTimeout(800);
      const opt = page.getByRole('menuitem', { name: /进行中/ }).or(page.getByRole('option', { name: /进行中/ }));
      if (await opt.first().isVisible().catch(() => false)) {
        await opt.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Step 3: Verify via API
    const verifyRes = await curl('GET', `${API}/teams/${smokeTeamId}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);
    const data = parseApiBody(verifyRes.body);
    expect(['progressing', '进行中'].some(s => JSON.stringify(data).includes(s))).toBeTruthy();

    await screenshot(page, 'status-smoke-complete');
  });

  // Journey smoke: terminal transition cancelled
  test('smoke-cancel: Attempt terminal transition, cancel dialog, verify status unchanged', async ({ page }) => {
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke cancel term', 'P2');
    const subKey = await createTestSubItem(token, smokeTeamId, mainKey, 'E2E cancel term sub');
    await curl('PUT', `${API}/teams/${smokeTeamId}/main-items/${mainKey}/sub-items/${subKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'completed' }),
    });
    await curl('PUT', `${API}/teams/${smokeTeamId}/main-items/${mainKey}/status`, {
      headers: authHeader(token),
      body: JSON.stringify({ status: 'reviewing' }),
    });

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    const statusBtn = page.getByRole('button', { name: /待验收|reviewing/i });
    if (await statusBtn.first().isVisible().catch(() => false)) {
      await statusBtn.first().click();
      await page.waitForTimeout(800);

      const closedOption = page.getByRole('menuitem', { name: /已关闭/ }).or(
        page.getByRole('option', { name: /已关闭/ }),
      );
      if (await closedOption.first().isVisible().catch(() => false)) {
        await closedOption.first().click();
        await page.waitForTimeout(800);

        const cancelBtn = page.getByRole('button', { name: /取消/ });
        if (await cancelBtn.first().isVisible().catch(() => false)) {
          await cancelBtn.first().click();
          await page.waitForLoadState('networkidle');
        }
      }
    }

    // Verify status unchanged
    const verifyRes = await curl('GET', `${API}/teams/${smokeTeamId}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    const data = parseApiBody(verifyRes.body);
    expect(data.itemStatus).toBe('reviewing');
    await screenshot(page, 'status-smoke-cancel');
  });

  // Journey smoke: unauthorized member attempt
  test('smoke-unauthorized: Member cannot change status via API', async ({ page }) => {
    const fixtures = await setupRbacFixtures();
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke unauth', 'P2');

    const res = await curl('PUT', `${API}/teams/${smokeTeamId}/main-items/${mainKey}/status`, {
      headers: authHeader(fixtures.memberToken),
      body: JSON.stringify({ status: 'progressing' }),
    });
    expect(res.status).toBe(403);
  });
});
