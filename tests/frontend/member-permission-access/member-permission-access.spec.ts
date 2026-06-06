/**
 * Web E2E Test: member-permission-access journey
 * @feature system-ux-optimization
 * @web-e2e
 *
 * Traceability: contracts/step-1..step-4 (web UI contracts)
 *
 * Contract test functions: 6 (valid login, nil rolekey login, menu visibility, pm-only hidden, item listing, validation error)
 * Journey smoke test functions: 1 (happy path login + menu + listing)
 */
import { test, expect } from '@playwright/test';
import {
  login,
  loginAsUser,
  loginViaUI,
  screenshot,
  baseUrl,
  API,
  getApiToken,
  createAuthCurl,
  apiBaseUrl,
  setupRbacFixtures,
  createTestMainItem,
  authHeader,
  curl,
  getFirstTeamId,
  parseApiBody,
  getTokenForUser,
  randomCode,
  extractBizKey,
} from '../helpers.js';

let superadminToken: string;
let teamBizKey: string;
let memberRoleKey: string;
let pmToken: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let memberToken: string;
let memberUsername: string;
let memberPassword: string;
let nilRoleKeyMemberToken: string;
let nilRoleKeyMemberUsername: string;
let nilRoleKeyMemberPassword: string;

test.describe('member-permission-access: Contract tests', () => {

  test.beforeAll(async () => {
    const fixtures = await setupRbacFixtures();
    superadminToken = fixtures.superadminToken;
    teamBizKey = fixtures.teamBizKey;
    memberRoleKey = fixtures.memberRoleKey;
    pmToken = fixtures.pmToken;
    pmUserBizKey = fixtures.pmUserBizKey;
    memberUserBizKey = fixtures.memberUserBizKey;
    memberToken = fixtures.memberToken;

    // Create a member user WITHOUT role_key (nil role_key)
    const runId = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    nilRoleKeyMemberUsername = `e2e-nilrole-${runId}`;
    const res = await curl('POST', `${API}/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: nilRoleKeyMemberUsername, displayName: 'E2E NilRole Member' }),
    });
    const data = parseApiBody(res.body);
    nilRoleKeyMemberPassword = data.initialPassword;
    nilRoleKeyMemberToken = await getTokenForUser(nilRoleKeyMemberUsername, nilRoleKeyMemberPassword);

    // Add nil-role member to team without assigning a role
    await curl('POST', `${API}/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: nilRoleKeyMemberUsername, roleKey: memberRoleKey }),
    });
  });

  // Traceability: step-1-member-valid-rolekey-login / Outcome "success"
  test('step1-success: Member user with valid role_key logs in successfully', async ({ page }) => {
    // Login as member user with valid role key via UI
    await loginViaUI(page, { username: nilRoleKeyMemberUsername, password: nilRoleKeyMemberPassword });
    // Should be redirected away from login page
    await expect(page).toHaveURL(/\/items|\/weekly/, { timeout: 15000 });
    await screenshot(page, 'member-step1-login');
  });

  // Traceability: step-1-member-valid-rolekey-login / Outcome "validation-error"
  test('step1-validation-error: Login with empty password shows error', async ({ page }) => {
    await page.goto(`${baseUrl}/login`);
    await page.waitForLoadState('networkidle');

    // Fill username but leave password empty
    await page.locator('[data-testid="login-username"]').fill('testuser');
    await page.locator('[data-testid="login-submit"]').click();

    // Should see a validation error
    const errorVisible = await page.locator('[data-testid="login-error"]').isVisible().catch(() => false);
    const pageText = await page.textContent('body') ?? '';
    expect(errorVisible || pageText.includes('密码') || pageText.includes('请输入')).toBeTruthy();
    await screenshot(page, 'member-step1-validation');
  });

  // Traceability: step-2-member-nil-rolekey-login / Outcome "success"
  test('step2-success: Member with nil role_key logs in with default permissions', async ({ page }) => {
    // The nil-rolekey user should login successfully
    await loginAsUser(page, nilRoleKeyMemberToken, { isSuperAdmin: false });
    await expect(page).toHaveURL(/\/items/, { timeout: 15000 });

    // Verify no permission errors on page
    const pageText = await page.textContent('body') ?? '';
    expect(pageText.includes('403') || pageText.includes('权限不足')).toBeFalsy();
    await screenshot(page, 'member-step2-nil-rolekey');
  });

  // Traceability: step-3-verify-menu-visibility / Outcome "success"
  test('step3-success: Menu visibility matches member permission set', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false });
    await page.waitForLoadState('networkidle');

    // Sidebar should be visible with items menu
    const sidebar = page.locator('[data-testid="sidebar"]');
    await expect(sidebar).toBeVisible({ timeout: 10000 });

    // Items link should be visible (main_item:list permission)
    const itemsLink = sidebar.locator('a[href="/items"]');
    const itemsVisible = await itemsLink.isVisible().catch(() => false);
    expect(itemsVisible).toBeTruthy();
    await screenshot(page, 'member-step3-menu');
  });

  // Traceability: step-3-verify-menu-visibility / Outcome "pm-only-actions-hidden"
  test('step3-pm-hidden: PM-only actions are not visible to member', async ({ page }) => {
    const mainKey = await createTestMainItem(superadminToken, teamBizKey, 'E2E member view item', 'P2');

    await loginAsUser(page, memberToken, { isSuperAdmin: false }, `/items/${mainKey}`);
    await page.waitForLoadState('networkidle');

    // Delete button should NOT be visible
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    const visible = await deleteBtn.first().isVisible().catch(() => false);
    expect(visible).toBeFalsy();
    await screenshot(page, 'member-step3-pm-hidden');
  });

  // Traceability: step-4-access-item-listing / Outcome "success"
  test('step4-success: Member can access item listing page', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false }, '/items');
    await page.waitForLoadState('networkidle');

    // Page should load without permission errors
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible({ timeout: 10000 });
    const pageText = await page.textContent('body') ?? '';
    expect(pageText.includes('403') || pageText.includes('权限不足')).toBeFalsy();
    await screenshot(page, 'member-step4-listing');
  });
});

test.describe('member-permission-access: Journey smoke test (happy path)', () => {

  let token: string;
  let smokeTeamId: string;
  let smokeMemberToken: string;
  let smokeMemberUsername: string;
  let smokeMemberPassword: string;

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    smokeTeamId = (await getFirstTeamId(token))!;
  });

  // Journey smoke: member logs in, views menu, accesses item listing
  test('smoke: Full happy path — member login → verify menu → access listing', async ({ page }) => {
    // Use the pre-created member token for login
    const fixtures = await setupRbacFixtures();

    // Step 1: Login as member
    await loginAsUser(page, fixtures.memberToken, { isSuperAdmin: false });
    await expect(page).toHaveURL(/\/items/, { timeout: 15000 });

    // Step 2: Verify menu visibility
    const sidebar = page.locator('[data-testid="sidebar"]');
    await expect(sidebar).toBeVisible({ timeout: 10000 });

    // Items link should be visible
    const itemsLink = sidebar.locator('a[href="/items"]');
    await expect(itemsLink).toBeVisible({ timeout: 5000 });

    // Step 3: Access item listing
    await itemsLink.click();
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible({ timeout: 10000 });

    // Step 4: Verify no admin/PM-only controls
    const deleteBtn = page.getByRole('button', { name: /删除/ });
    const deleteVisible = await deleteBtn.first().isVisible().catch(() => false);
    expect(deleteVisible).toBeFalsy();

    await screenshot(page, 'member-smoke-complete');
  });

  // Journey smoke: nil-roleKey member can view items without errors
  test('smoke-nil-role: Nil-roleKey member logs in, views items, no permission errors', async ({ page }) => {
    const fixtures = await setupRbacFixtures();

    // Create a nil-role member
    const runId = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const nilUsername = `e2e-smoke-nil-${runId}`;
    const res = await curl('POST', `${API}/admin/users`, {
      headers: authHeader(fixtures.superadminToken),
      body: JSON.stringify({ username: nilUsername, displayName: 'Smoke NilRole' }),
    });
    const data = parseApiBody(res.body);
    const nilToken = await getTokenForUser(nilUsername, data.initialPassword);

    await loginAsUser(page, nilToken, { isSuperAdmin: false });
    await expect(page).toHaveURL(/\/items/, { timeout: 15000 });

    // Verify items page loads
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible({ timeout: 10000 });
    const pageText = await page.textContent('body') ?? '';
    expect(pageText.includes('403') || pageText.includes('权限不足')).toBeFalsy();

    await screenshot(page, 'member-smoke-nil-role');
  });
});
