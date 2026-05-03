import { test, expect } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId } from '../helpers.js';

// ── Section 1: Page Load ──────────────────────────────────────────────────────

test.describe('Team Management - Page Load', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('page renders with table headers', async ({ page }) => {
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('th', { hasText: '团队名称' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Code' })).toBeVisible();
    await expect(page.locator('th', { hasText: '操作' })).toBeVisible();
  });

  test('create team button is visible', async ({ page }) => {
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button', { hasText: '创建团队' })).toBeVisible();
  });
});

// ── Section 2: Create Team ────────────────────────────────────────────────────

test.describe('Team Management - Create Team', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking create team opens dialog', async ({ page }) => {
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建团队' }).click();
    await expect(page.locator('input[placeholder="请输入团队名称"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('input[placeholder="如 FEAT、CORE"]')).toBeVisible();
  });

  test('submit disabled when name or code is empty', async ({ page }) => {
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建团队' }).click();
    await expect(page.locator('input[placeholder="请输入团队名称"]')).toBeVisible({ timeout: 5000 });
    const submitBtn = page.locator('button', { hasText: '确认创建' });
    await expect(submitBtn).toBeDisabled();
  });

  test('invalid code shows validation error', async ({ page }) => {
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建团队' }).click();
    await page.locator('input[placeholder="请输入团队名称"]').fill('Test Team');
    const codeInput = page.locator('input[placeholder="如 FEAT、CORE"]');
    await codeInput.fill('1');
    await codeInput.blur();
    await expect(page.locator('text=CODE须为 2~6 位英文字母')).toBeVisible({ timeout: 3000 });
  });

  test('creates team successfully and shows it in table', async ({ page }) => {
    const uniqueCode = `T${Date.now().toString(36).slice(-4).toUpperCase().replace(/\d/g, 'X')}`;
    const teamName = `E2E Team ${uniqueCode}`;

    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建团队' }).click();
    await page.locator('input[placeholder="请输入团队名称"]').fill(teamName);
    await page.locator('input[placeholder="如 FEAT、CORE"]').fill(uniqueCode);
    await page.locator('button', { hasText: '确认创建' }).click();

    await expect(page.locator('input[placeholder="请输入团队名称"]')).not.toBeVisible({ timeout: 5000 });
    await expect(page.locator(`text=${teamName}`)).toBeVisible({ timeout: 5000 });

    const listRes = await fetch(`${API}/teams`, { headers: { Authorization: `Bearer ${authToken}` } });
    const listData = await listRes.json();
    const raw = listData.data ?? listData;
    const teams = Array.isArray(raw) ? raw : (raw?.items ?? []);
    const created = teams.find((t: any) => t.name === teamName);
    if (created) {
      await fetch(`${API}/teams/${created.id || created.ID}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
    }
  });
});

// ── Section 3: Add Member ─────────────────────────────────────────────────────

test.describe('Team Management - Add Member', () => {
  let authToken: string;
  let teamId: string | null = null;
  let testUserId: string;
  let testUserDisplay: string;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
    // Create a non-member user for search tests
    const username = `e2e_search_${Date.now()}`;
    const res = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, displayName: 'E2E搜索用户' }),
    });
    const data = await res.json();
    testUserId = String(data.data?.id || data.id);
    testUserDisplay = 'E2E搜索用户';
  });

  test.afterAll(async () => {
    if (testUserId) {
      try {
        await fetch(`${API}/admin/users/${testUserId}`, { method: 'DELETE', headers: { Authorization: `Bearer ${authToken}` } });
      } catch { /* best effort */ }
    }
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('each team row has an add member button', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    const btn = page.locator('button[data-testid^="add-member-btn-"]').first();
    await expect(btn).toBeVisible();
    await expect(btn).toContainText('添加成员');
  });

  test('clicking add member opens dialog with search and role fields', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="add-member-btn-${teamId}"]`).click();
    await expect(page.locator('[data-testid="add-member-user-search"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('[data-testid="add-member-role-select"]')).toBeVisible();
    await expect(page.locator('[data-testid="add-member-submit-btn"]')).toBeDisabled();
  });

  test('searching users shows dropdown results', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="add-member-btn-${teamId}"]`).click();
    const searchInput = page.locator('[data-testid="add-member-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });
    await searchInput.fill(testUserDisplay);
    const dropdown = page.locator('[data-testid="add-member-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });
    const options = dropdown.locator('button[data-testid^="add-member-user-option-"]');
    expect(await options.count()).toBeGreaterThan(0);
  });

  test('selecting user enables submit button', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="add-member-btn-${teamId}"]`).click();
    const searchInput = page.locator('[data-testid="add-member-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });
    await searchInput.fill(testUserDisplay);
    const dropdown = page.locator('[data-testid="add-member-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });
    await dropdown.locator('button[data-testid^="add-member-user-option-"]').first().click();
    await expect(dropdown).not.toBeVisible();
    await expect(page.locator('[data-testid="add-member-submit-btn"]')).toBeEnabled({ timeout: 3000 });
  });

  test('role select does not include pm option', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="add-member-btn-${teamId}"]`).click();
    await expect(page.locator('[data-testid="add-member-role-select"]')).toBeVisible({ timeout: 5000 });
    await page.locator('[data-testid="add-member-role-select"]').click({ force: true });
    const pmOption = page.locator('[role="option"]', { hasText: /^pm$/ });
    await expect(pmOption).not.toBeVisible({ timeout: 2000 }).catch(() => {});
    expect(await pmOption.count()).toBe(0);
  });

  test('cancelling dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams`);
    await expect(page.locator('[data-testid="team-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="add-member-btn-${teamId}"]`).click();
    await expect(page.locator('[data-testid="add-member-user-search"]')).toBeVisible({ timeout: 5000 });
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-testid="add-member-user-search"]')).not.toBeVisible({ timeout: 3000 });
  });
});
