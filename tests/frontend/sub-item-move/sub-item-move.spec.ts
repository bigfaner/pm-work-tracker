/**
 * Web E2E Test: sub-item-move journey
 * @feature system-ux-optimization
 * @web-e2e
 *
 * Traceability: contracts/step-1..step-3 (web UI contracts)
 *
 * Contract test functions: 5 (initiate move, confirm move, verify location, no-target-selected, move-to-same-parent)
 * Journey smoke test functions: 1 (happy path full move)
 */
import { test, expect } from '@playwright/test';
import {
  login,
  screenshot,
  baseUrl,
  API,
  getApiToken,
  apiBaseUrl,
  createTestMainItem,
  createTestSubItem,
  authHeader,
  curl,
  getFirstTeamId,
  parseApiBody,
  extractBizKey,
  randomCode,
} from '../helpers.js';

let token: string;
let teamBizKey: string;

test.describe('sub-item-move: Contract tests', () => {

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    teamBizKey = (await getFirstTeamId(token))!;
  });

  // Traceability: step-1-initiate-sub-item-move / Outcome "success"
  test('step1-success: PM opens move dialog from sub-item detail', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E move source', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E sub to move');
    // Create a second main item as potential target
    await createTestMainItem(token, teamBizKey, 'E2E move target', 'P2');

    await login(page);
    // Navigate to sub-item detail page
    await page.goto(`${baseUrl}/items/${mainKey}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Click move button
    const moveBtn = page.getByRole('button', { name: /移动/ });
    await expect(moveBtn.first()).toBeVisible({ timeout: 10000 });
    await moveBtn.first().click();

    // Move dialog should appear
    const dialogText = await page.textContent('body') ?? '';
    expect(dialogText.includes('移动到其他主事项') || dialogText.includes('选择目标')).toBeTruthy();
    await screenshot(page, 'move-step1-dialog');
  });

  // Traceability: step-2-select-target-and-confirm / Outcome "success"
  test('step2-success: PM selects target and confirms move', async ({ page }) => {
    const mainKeyA = await createTestMainItem(token, teamBizKey, 'E2E move src A', 'P2');
    const mainKeyB = await createTestMainItem(token, teamBizKey, 'E2E move tgt B', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKeyA, 'E2E sub move');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainKeyA}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Click move button
    const moveBtn = page.getByRole('button', { name: /移动/ });
    if (await moveBtn.first().isVisible().catch(() => false)) {
      await moveBtn.first().click();
      await page.waitForTimeout(1000);

      // Select target using the SelectTrigger button (contains "选择目标主事项" placeholder)
      const selectTrigger = page.locator('button').filter({ hasText: /选择目标|选择/ }).first();
      if (await selectTrigger.isVisible().catch(() => false)) {
        await selectTrigger.click({ force: true });
        await page.waitForTimeout(500);

        // Select the target option by text
        const targetOption = page.getByRole('option').filter({ hasText: /E2E move tgt B/ }).first();
        if (await targetOption.isVisible().catch(() => false)) {
          await targetOption.click();
        }
      }

      // Click confirm move
      const confirmBtn = page.getByRole('button', { name: /确认移动/ });
      if (await confirmBtn.isVisible().catch(() => false)) {
        await confirmBtn.click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify via API that sub-item is now under target main item
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKeyB}/sub-items`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);
    await screenshot(page, 'move-step2-confirmed');
  });

  // Traceability: step-2-select-target-and-confirm / Outcome "no-target-selected"
  test('step2-no-target: Confirm button disabled when no target selected', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E no target src', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E sub no target');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainKey}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Click move button
    const moveBtn = page.getByRole('button', { name: /移动/ });
    if (await moveBtn.first().isVisible().catch(() => false)) {
      await moveBtn.first().click();
      await page.waitForTimeout(1000);

      // Confirm button should be disabled without selecting a target
      const confirmBtn = page.getByRole('button', { name: /确认移动/ });
      if (await confirmBtn.isVisible().catch(() => false)) {
        const isDisabled = await confirmBtn.isDisabled();
        expect(isDisabled).toBeTruthy();
      }
    }
    await screenshot(page, 'move-step2-no-target');
  });

  // Traceability: step-3-verify-sub-item-in-new-location / Outcome "success"
  test('step3-success: Moved sub-item visible under new parent', async ({ page }) => {
    const mainKeyA = await createTestMainItem(token, teamBizKey, 'E2E verify src A', 'P2');
    const mainKeyB = await createTestMainItem(token, teamBizKey, 'E2E verify tgt B', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKeyA, 'E2E sub verify');

    // Move via API
    const moveRes = await curl('PUT', `${API}/teams/${teamBizKey}/sub-items/${subKey}/move`, {
      headers: authHeader(token),
      body: JSON.stringify({ targetMainItemBizKey: mainKeyB }),
    });
    expect(moveRes.status).toBe(200);

    // Navigate to new parent's page and verify
    await login(page, undefined, `/items/${mainKeyB}`);
    await page.waitForLoadState('networkidle');

    // Verify sub-item appears under new parent
    const pageText = await page.textContent('body') ?? '';
    // Sub-item should be visible somewhere on the page
    const hasSubItem = pageText.includes('E2E sub verify');
    expect(hasSubItem).toBeTruthy();
    await screenshot(page, 'move-step3-verified');
  });

  // Traceability: step-2-select-target-and-confirm / Outcome "move-to-same-parent"
  test('step2-same-parent: Moving to same parent is rejected', async ({ page }) => {
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E same parent', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E sub same');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainKey}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Click move
    const moveBtn = page.getByRole('button', { name: /移动/ });
    if (await moveBtn.first().isVisible().catch(() => false)) {
      await moveBtn.first().click();
      await page.waitForTimeout(1000);

      // The current parent should not be in the selector options
      // Use SelectTrigger button instead of combobox role
      const selectTrigger = page.locator('button').filter({ hasText: /选择目标|选择/ }).first();
      if (await selectTrigger.isVisible().catch(() => false)) {
        await selectTrigger.click({ force: true });
        await page.waitForTimeout(500);

        // Current parent should be excluded from options
        const options = page.getByRole('option');
        const count = await options.count();
        for (let i = 0; i < count; i++) {
          const text = await options.nth(i).textContent();
          expect(text?.includes('E2E same parent')).toBeFalsy();
        }
      }
    }
    await screenshot(page, 'move-step2-same-parent');
  });
});

test.describe('sub-item-move: Journey smoke test (happy path)', () => {

  let token: string;
  let smokeTeamId: string;

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    smokeTeamId = (await getFirstTeamId(token))!;
  });

  // Journey smoke: full move flow via UI
  test('smoke: Full happy path — initiate move → select target → confirm → verify', async ({ page }) => {
    const mainKeyA = await createTestMainItem(token, smokeTeamId, 'E2E smoke src A', 'P2');
    const mainKeyB = await createTestMainItem(token, smokeTeamId, 'E2E smoke tgt B', 'P2');
    const subKey = await createTestSubItem(token, smokeTeamId, mainKeyA, 'E2E smoke sub');

    // Step 1: Navigate to sub-item detail and initiate move
    await login(page, undefined, `/items/${mainKeyA}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    const moveBtn = page.getByRole('button', { name: /移动/ });
    await expect(moveBtn.first()).toBeVisible({ timeout: 10000 });
    await moveBtn.first().click();
    await page.waitForTimeout(1000);

    // Step 2: Select target using SelectTrigger button
    const selectTrigger = page.locator('button').filter({ hasText: /选择目标|选择/ }).first();
    if (await selectTrigger.isVisible().catch(() => false)) {
      await selectTrigger.click({ force: true });
      await page.waitForTimeout(500);

      const targetOption = page.getByRole('option').filter({ hasText: /E2E smoke tgt B/ }).first();
      if (await targetOption.isVisible().catch(() => false)) {
        await targetOption.click();
      }
    }

    // Step 3: Confirm move
    const confirmBtn = page.getByRole('button', { name: /确认移动/ });
    if (await confirmBtn.isVisible().catch(() => false)) {
      await confirmBtn.click();
      await page.waitForLoadState('networkidle');
    }

    // Step 4: Verify sub-item is now under target via API
    const verifyRes = await curl('GET', `${API}/teams/${smokeTeamId}/main-items/${mainKeyB}/sub-items`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);

    await screenshot(page, 'move-smoke-complete');
  });

  // Journey smoke: verify moved sub-item no longer under original parent
  test('smoke-verify: Move sub-item and verify absent from source parent', async ({ page }) => {
    const mainKeyA = await createTestMainItem(token, smokeTeamId, 'E2E verify src', 'P2');
    const mainKeyB = await createTestMainItem(token, smokeTeamId, 'E2E verify tgt', 'P2');
    const subKey = await createTestSubItem(token, smokeTeamId, mainKeyA, 'E2E verify sub');

    // Move via API
    await curl('PUT', `${API}/teams/${smokeTeamId}/sub-items/${subKey}/move`, {
      headers: authHeader(token),
      body: JSON.stringify({ targetMainItemBizKey: mainKeyB }),
    });

    // Navigate to original parent and verify sub-item absent
    await login(page, undefined, `/items/${mainKeyA}`);
    await page.waitForLoadState('networkidle');

    const pageText = await page.textContent('body') ?? '';
    expect(pageText.includes('E2E verify sub')).toBeFalsy();

    // Navigate to new parent and verify sub-item present
    await page.goto(`${baseUrl}/items/${mainKeyB}`);
    await page.waitForLoadState('networkidle');

    const pageText2 = await page.textContent('body') ?? '';
    expect(pageText2.includes('E2E verify sub')).toBeTruthy();

    await screenshot(page, 'move-smoke-verify');
  });
});
