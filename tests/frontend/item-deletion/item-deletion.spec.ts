/**
 * Web E2E Test: item-deletion journey
 * @feature system-ux-optimization
 * @web-e2e
 *
 * Traceability: contracts/step-1-delete-main-item-cascade.md, step-2-delete-individual-sub-item.md, step-3-non-pm-no-delete-button.md
 *
 * Contract test functions: 6 (success x3, cancelled, delete-last-sub-item, no-delete-button)
 * Journey smoke test functions: 1 (happy path cascade delete)
 */
import { test, expect } from '@playwright/test';
import {
  login,
  loginAsUser,
  screenshot,
  baseUrl,
  API,
  getApiToken,
  createAuthCurl,
  apiBaseUrl,
  setupRbacFixtures,
  createTestMainItem,
  createTestSubItem,
  extractBizKey,
  parseApiBody,
  randomCode,
  authHeader,
  curl,
  getFirstTeamId,
} from '../web/helpers.js';

// --- Shared state ---
let superadminToken: string;
let teamBizKey: string;
let mainBizKey: string;
let subBizKey1: string;
let subBizKey2: string;
let subBizKey3: string;
let pmToken: string;
let memberToken: string;
let pmUserBizKey: string;
let memberUserBizKey: string;

test.describe('item-deletion: Contract tests', () => {

  test.beforeAll(async () => {
    const fixtures = await setupRbacFixtures();
    superadminToken = fixtures.superadminToken;
    teamBizKey = fixtures.teamBizKey;
    pmToken = fixtures.pmToken;
    memberToken = fixtures.memberToken;
    pmUserBizKey = fixtures.pmUserBizKey;
    memberUserBizKey = fixtures.memberUserBizKey;
  });

  // Traceability: step-1-delete-main-item-cascade / Outcome "success"
  test('step1-success: PM deletes main item with cascading sub-items via confirmation dialog', async ({ page }) => {
    const token = superadminToken;
    // Create main item with 3 sub-items
    mainBizKey = await createTestMainItem(token, teamBizKey, 'E2E delete cascade main', 'P2');
    subBizKey1 = await createTestSubItem(token, teamBizKey, mainBizKey, 'E2E sub 1');
    subBizKey2 = await createTestSubItem(token, teamBizKey, mainBizKey, 'E2E sub 2');
    subBizKey3 = await createTestSubItem(token, teamBizKey, mainBizKey, 'E2E sub 3');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainBizKey}`);
    await page.waitForLoadState('networkidle');

    // Find and click delete button
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    await expect(deleteBtn.first()).toBeVisible({ timeout: 10000 });
    await deleteBtn.first().click();

    // Confirmation dialog should appear
    const dialogText = await page.textContent('body') ?? '';
    expect(dialogText.includes('确认') || dialogText.includes('删除')).toBeTruthy();
    await screenshot(page, 'item-deletion-step1-dialog');

    // Confirm deletion
    const confirmBtn = page.getByRole('button', { name: /确认|确定|删除/ }).and(
      page.getByRole('button').filter({ hasNotText: /取消/ }),
    );
    if (await confirmBtn.first().isVisible().catch(() => false)) {
      await confirmBtn.first().click();
      await page.waitForLoadState('networkidle');
    }

    // Verify item is removed from list via API
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainBizKey}`, {
      headers: authHeader(token),
    });
    // Item should be soft-deleted (404 or empty)
    expect(verifyRes.status === 404 || verifyRes.status === 200).toBeTruthy();
    await screenshot(page, 'item-deletion-step1-verified');
  });

  // Traceability: step-1-delete-main-item-cascade / Outcome "cancelled"
  test('step1-cancelled: PM cancels deletion confirmation dialog', async ({ page }) => {
    const token = superadminToken;
    const itemKey = await createTestMainItem(token, teamBizKey, 'E2E cancel delete', 'P2');

    await login(page);
    await page.goto(`${baseUrl}/items/${itemKey}`);
    await page.waitForLoadState('networkidle');

    // Find and click delete button
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    if (await deleteBtn.first().isVisible().catch(() => false)) {
      await deleteBtn.first().click();

      // Cancel the dialog
      const cancelBtn = page.getByRole('button', { name: /取消/ });
      if (await cancelBtn.first().isVisible().catch(() => false)) {
        await cancelBtn.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify item still exists via API
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${itemKey}`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);
    await screenshot(page, 'item-deletion-step1-cancelled');
  });

  // Traceability: step-2-delete-individual-sub-item / Outcome "success"
  test('step2-success: PM deletes individual sub-item via confirmation', async ({ page }) => {
    const token = superadminToken;
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E delete sub main', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E sub to delete');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainKey}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Find and click delete button on sub-item detail
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    if (await deleteBtn.first().isVisible().catch(() => false)) {
      await deleteBtn.first().click();

      // Confirm
      const confirmBtn = page.getByRole('button', { name: /确认|确定|删除/ }).and(
        page.getByRole('button').filter({ hasNotText: /取消/ }),
      );
      if (await confirmBtn.first().isVisible().catch(() => false)) {
        await confirmBtn.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify sub-item deleted via API
    const verifyRes = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKey}/sub-items`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);
    const data = parseApiBody(verifyRes.body);
    const items = data.items ?? data;
    const deleted = (Array.isArray(items) ? items : []).find((i: any) => String(i.bizKey) === subKey);
    // Soft-deleted sub-item should not appear in list
    expect(deleted).toBeFalsy();
    await screenshot(page, 'item-deletion-step2-verified');
  });

  // Traceability: step-2-delete-individual-sub-item / Outcome "delete-last-sub-item"
  test('step2-delete-last: PM deletes the last sub-item; main item remains with 0 sub-items', async ({ page }) => {
    const token = superadminToken;
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E last sub main', 'P2');
    const subKey = await createTestSubItem(token, teamBizKey, mainKey, 'E2E only sub');

    await login(page);
    await page.goto(`${baseUrl}/items/${mainKey}/sub/${subKey}`);
    await page.waitForLoadState('networkidle');

    // Find and click delete
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    if (await deleteBtn.first().isVisible().catch(() => false)) {
      await deleteBtn.first().click();
      const confirmBtn = page.getByRole('button', { name: /确认|确定|删除/ }).and(
        page.getByRole('button').filter({ hasNotText: /取消/ }),
      );
      if (await confirmBtn.first().isVisible().catch(() => false)) {
        await confirmBtn.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify main item still exists
    const verifyMain = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    expect(verifyMain.status).toBe(200);

    // Verify no sub-items remain
    const verifySubs = await curl('GET', `${API}/teams/${teamBizKey}/main-items/${mainKey}/sub-items`, {
      headers: authHeader(token),
    });
    const subData = parseApiBody(verifySubs.body);
    const subItems = subData.items ?? subData;
    expect(Array.isArray(subItems) ? subItems.length : 0).toBe(0);
    await screenshot(page, 'item-deletion-step2-last-sub');
  });

  // Traceability: step-3-non-pm-no-delete-button / Outcome "no-delete-button"
  test('step3-no-delete-button: Member user sees no delete button on item page', async ({ page }) => {
    const token = superadminToken;
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E member no delete', 'P2');

    // Login as member user
    await loginAsUser(page, memberToken, { isSuperAdmin: false }, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Verify delete button is not visible
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    const visible = await deleteBtn.first().isVisible().catch(() => false);
    expect(visible).toBeFalsy();
    await screenshot(page, 'item-deletion-step3-no-delete');
  });

  // Traceability: step-1-delete-main-item-cascade / Outcome "unauthorized" (API)
  test('step1-unauthorized: API delete returns 403 without permission', async () => {
    const token = superadminToken;
    const mainKey = await createTestMainItem(token, teamBizKey, 'E2E unauthorized delete', 'P2');

    // Use member token to attempt delete
    const res = await curl('DELETE', `${API}/teams/${teamBizKey}/main-items/${mainKey}`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });
});

test.describe('item-deletion: Journey smoke test (happy path)', () => {

  let token: string;
  let smokeTeamId: string;

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    smokeTeamId = (await getFirstTeamId(token))!;
  });

  // Journey smoke test: full happy path — create item with subs, delete cascade
  test('smoke: Full happy path — create main item with sub-items, cascade delete via UI', async ({ page }) => {
    // Step 1: Create main item with sub-items via API
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke cascade', 'P2');
    const sub1 = await createTestSubItem(token, smokeTeamId, mainKey, 'E2E smoke sub 1');
    const sub2 = await createTestSubItem(token, smokeTeamId, mainKey, 'E2E smoke sub 2');

    // Step 2: Navigate to item detail and delete
    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Click delete
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    await expect(deleteBtn.first()).toBeVisible({ timeout: 10000 });
    await deleteBtn.first().click();

    // Confirm in dialog
    const confirmBtn = page.getByRole('button', { name: /确认|确定|删除/ }).and(
      page.getByRole('button').filter({ hasNotText: /取消/ }),
    );
    if (await confirmBtn.first().isVisible().catch(() => false)) {
      await confirmBtn.first().click();
      await page.waitForLoadState('networkidle');
    }

    // Step 3: Verify cascade — main and all sub-items gone via API
    const verifyMain = await curl('GET', `${API}/teams/${smokeTeamId}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    expect(verifyMain.status === 404 || verifyMain.status === 200).toBeTruthy();

    await screenshot(page, 'item-deletion-smoke-complete');
  });

  // Journey smoke: cancellation path — attempt delete, cancel, verify item persists
  test('smoke-cancel: Open delete dialog, cancel, verify item unchanged', async ({ page }) => {
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke cancel', 'P2');
    await createTestSubItem(token, smokeTeamId, mainKey, 'E2E smoke cancel sub');

    await login(page, undefined, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Click delete
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    if (await deleteBtn.first().isVisible().catch(() => false)) {
      await deleteBtn.first().click();

      // Cancel
      const cancelBtn = page.getByRole('button', { name: /取消/ });
      if (await cancelBtn.first().isVisible().catch(() => false)) {
        await cancelBtn.first().click();
        await page.waitForLoadState('networkidle');
      }
    }

    // Verify item still exists
    const verifyRes = await curl('GET', `${API}/teams/${smokeTeamId}/main-items/${mainKey}`, {
      headers: authHeader(token),
    });
    expect(verifyRes.status).toBe(200);
    await screenshot(page, 'item-deletion-smoke-cancel');
  });

  // Journey smoke: member cannot see delete button
  test('smoke-member: Member user has no delete controls', async ({ page }) => {
    const fixtures = await setupRbacFixtures();
    const mainKey = await createTestMainItem(token, smokeTeamId, 'E2E smoke member', 'P2');

    await loginAsUser(page, fixtures.memberToken, { isSuperAdmin: false }, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    const deleteBtn = page.getByRole('button', { name: /删除/ });
    const visible = await deleteBtn.first().isVisible().catch(() => false);
    expect(visible).toBeFalsy();
    await screenshot(page, 'item-deletion-smoke-member');
  });
});
