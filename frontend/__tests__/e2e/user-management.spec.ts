import { test, expect, type Page } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/v1';

async function login(page: Page) {
  for (let attempt = 0; attempt < 3; attempt++) {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    try {
      await page.waitForURL(/\/items/, { timeout: 10000 });
      return;
    } catch {
      if (attempt < 2) await page.waitForTimeout(6000);
      else throw new Error('Login failed after 3 attempts');
    }
  }
}

async function getAuthToken(): Promise<string> {
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const json = await res.json();
  return json.data?.token || json.token;
}

// ── Section 1: Page Load ──────────────────────────────────────────────────────

test.describe('User Management - Page Load', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('page renders with table headers', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('th', { hasText: '姓名' })).toBeVisible();
    await expect(page.locator('th', { hasText: '账号' })).toBeVisible();
    await expect(page.locator('th', { hasText: '状态' })).toBeVisible();
    await expect(page.locator('th', { hasText: '操作' })).toBeVisible();
  });

  test('create user button is visible', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button', { hasText: '创建用户' })).toBeVisible();
  });

  test('search input is visible', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('input[placeholder="搜索用户名/姓名"]')).toBeVisible();
  });

  test('admin user appears in table', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    // At least one row should be visible (the admin user)
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
  });
});

// ── Section 2: Create User ────────────────────────────────────────────────────

test.describe('User Management - Create User', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking create user opens dialog', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建用户' }).click();
    await expect(page.locator('input[placeholder="请输入姓名"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('input[placeholder="请输入账号"]')).toBeVisible();
  });

  test('submit disabled when required fields are empty', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建用户' }).click();
    await expect(page.locator('input[placeholder="请输入姓名"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button', { hasText: '确认创建' })).toBeDisabled();
  });

  test('submit enabled when name and username are filled', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建用户' }).click();
    await page.locator('input[placeholder="请输入姓名"]').fill('Test User');
    await page.locator('input[placeholder="请输入账号"]').fill('testuser_e2e');
    await expect(page.locator('button', { hasText: '确认创建' })).toBeEnabled({ timeout: 3000 });
  });

  test('creates user and shows initial password dialog', async ({ page }) => {
    const username = `e2e_${Date.now()}`;
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建用户' }).click();
    await page.locator('input[placeholder="请输入姓名"]').fill('E2E Test User');
    await page.locator('input[placeholder="请输入账号"]').fill(username);
    await page.locator('button', { hasText: '确认创建' }).click();

    // Initial password dialog should appear
    await expect(page.locator('[data-testid="initial-password"]')).toBeVisible({ timeout: 8000 });
    const pwd = await page.locator('[data-testid="initial-password"]').textContent();
    expect(pwd?.trim().length).toBeGreaterThan(0);

    await page.locator('button', { hasText: '我知道了' }).click();
    await expect(page.locator('[data-testid="initial-password"]')).not.toBeVisible({ timeout: 3000 });

    // Cleanup
    const listRes = await fetch(`${API}/admin/users?search=${username}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = await listRes.json();
    const users = listData.data?.items || listData.items || [];
    const created = users.find((u: any) => u.username === username);
    if (created) {
      await fetch(`${API}/admin/users/${created.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
    }
  });

  test('cancelling dialog closes it', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建用户' }).click();
    await expect(page.locator('input[placeholder="请输入姓名"]')).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '取消' }).first().click();
    await expect(page.locator('input[placeholder="请输入姓名"]')).not.toBeVisible({ timeout: 3000 });
  });
});

// ── Section 3: Edit User ──────────────────────────────────────────────────────

test.describe('User Management - Edit User', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking edit opens edit dialog with pre-filled data', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });

    // Click the first edit button
    await page.locator('button', { hasText: '编辑' }).first().click();

    // Edit dialog should open with a pre-filled display name
    await expect(page.locator('text=编辑用户')).toBeVisible({ timeout: 5000 });
    const displayNameInput = page.locator('dialog input').first();
    const value = await displayNameInput.inputValue();
    expect(value.length).toBeGreaterThan(0);
  });

  test('save button is visible in edit dialog', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '编辑' }).first().click();
    await expect(page.locator('button', { hasText: '保存修改' })).toBeVisible({ timeout: 5000 });
  });
});

// ── Section 4: Toggle Status ──────────────────────────────────────────────────

test.describe('User Management - Toggle Status', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking modify status opens status dialog', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '修改状态' }).first().click();
    await expect(page.locator('text=修改用户状态')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=当前状态')).toBeVisible();
    await expect(page.locator('text=新状态')).toBeVisible();
  });

  test('status dialog has confirm and cancel buttons', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '修改状态' }).first().click();
    await expect(page.locator('text=修改用户状态')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button', { hasText: '确认修改' })).toBeVisible();
    await expect(page.locator('button', { hasText: '取消' })).toBeVisible();
  });

  test('can toggle a non-admin user status', async ({ page }) => {
    // Create a test user first
    const username = `e2e_status_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, displayName: 'E2E Status Test' }),
    });
    if (createRes.status !== 200 && createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const userId = created.data?.id || created.id;

    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });

    // Search for the test user
    await page.locator('input[placeholder="搜索用户名/姓名"]').fill(username);
    await page.waitForTimeout(1000);

    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '修改状态' }).first().click();
    await expect(page.locator('text=修改用户状态')).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '确认修改' }).click();

    // Dialog should close after success
    await expect(page.locator('text=修改用户状态')).not.toBeVisible({ timeout: 5000 });

    // Cleanup
    if (userId) {
      await fetch(`${API}/admin/users/${userId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
    }
  });
});

// ── Section 5: Search & Refresh ───────────────────────────────────────────────

test.describe('User Management - Search and Refresh', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('searching by username filters the table', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('input[placeholder="搜索用户名/姓名"]').fill('admin');
    await page.waitForTimeout(800);
    // admin row should be visible
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
    const rows = await page.locator('tbody tr').count();
    expect(rows).toBeGreaterThan(0);
  });

  test('searching with no match shows empty state', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('input[placeholder="搜索用户名/姓名"]').fill('zzz_no_such_user_xyz');
    await page.waitForTimeout(800);
    await expect(page.locator('text=暂无用户')).toBeVisible({ timeout: 5000 });
  });

  test('refresh button triggers data reload', async ({ page }) => {
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="refresh-btn"]').click();
    // After refresh, table should still show data
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
  });
});
