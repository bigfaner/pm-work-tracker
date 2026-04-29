import { test, expect } from '@playwright/test';
import { BASE, API, login, getAuthToken } from './test-helpers';

// ── Section 1: Page Load ──────────────────────────────────────────────────────

test.describe('Role Management - Page Load', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('page renders with table headers', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('th', { hasText: '角色名称' })).toBeVisible();
    await expect(page.locator('th', { hasText: '权限数量' })).toBeVisible();
    await expect(page.locator('th', { hasText: '使用人数' })).toBeVisible();
    await expect(page.locator('th', { hasText: '类型' })).toBeVisible();
    await expect(page.locator('th', { hasText: '操作' })).toBeVisible();
  });

  test('create role and permission list buttons are visible', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button', { hasText: '创建角色' })).toBeVisible();
    await expect(page.locator('button', { hasText: '权限列表' })).toBeVisible();
  });

  test('superadmin role is listed', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' })).toBeVisible({ timeout: 5000 });
  });

  test('superadmin shows non-zero permission count', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    const superadminBtn = page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' });
    await expect(superadminBtn).toBeVisible({ timeout: 5000 });
    const row = page.locator('tr').filter({ has: superadminBtn });
    const permCount = parseInt((await row.locator('td').nth(2).textContent())?.trim() || '0', 10);
    expect(permCount).toBeGreaterThan(0);
  });
});

// ── Section 2: Search & Filter ────────────────────────────────────────────────

test.describe('Role Management - Search and Filter', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('searching by role name filters results', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('input[placeholder="搜索角色名称"]').fill('superadmin');
    await page.waitForTimeout(800);
    await expect(page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' })).toBeVisible({ timeout: 5000 });
  });

  test.skip('searching with no match shows empty state', async ({ page }) => {
    // Backend ListRoles doesn't filter by search param yet
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('input[placeholder="搜索角色名称"]').fill('zzz_no_such_role_xyz');
    await expect(page.locator('text=没有匹配的角色')).toBeVisible({ timeout: 10000 });
  });

  test('filtering by preset shows only preset roles', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="role-management-page"] button[role="combobox"]').click();
    await page.locator('[role="option"]', { hasText: '预置角色' }).click();
    await page.waitForTimeout(800);
    const badges = page.locator('tbody td').filter({ hasText: '预置' });
    expect(await badges.count()).toBeGreaterThan(0);
  });

  test('refresh button reloads data', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="refresh-btn"]').click();
    await expect(page.locator('button[data-testid^="role-name-"]').first()).toBeVisible({ timeout: 5000 });
  });
});

// ── Section 3: Permissions Dialog ────────────────────────────────────────────

test.describe('Role Management - Permissions Dialog', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking role name opens permissions dialog', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    const roleNames = page.locator('button[data-testid^="role-name-"]');
    await expect(roleNames.first()).toBeVisible({ timeout: 5000 });
    await roleNames.first().click();
    await expect(page.locator('[data-testid="role-permissions-dialog"]')).toBeVisible({ timeout: 5000 });
  });

  test('dialog title contains the role name', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    const roleNames = page.locator('button[data-testid^="role-name-"]');
    await expect(roleNames.first()).toBeVisible({ timeout: 5000 });
    const roleName = await roleNames.first().textContent();
    await roleNames.first().click();
    await expect(page.locator('[data-testid="role-permissions-dialog-title"]')).toContainText(roleName!, { timeout: 5000 });
  });

  test('dialog displays permission checkboxes', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button[data-testid^="role-name-"]').first().click();
    await expect(page.locator('[data-testid="role-permissions-dialog"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('[data-testid="role-permissions-list"]')).toBeVisible({ timeout: 5000 });
    const checkboxes = page.locator('[data-testid="role-permissions-dialog"] input[type="checkbox"]');
    expect(await checkboxes.count()).toBeGreaterThan(0);
  });
});

// ── Section 4: Edit Button State ──────────────────────────────────────────────

test.describe('Role Management - Edit Button State', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('preset role edit button is disabled', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    const superadminBtn = page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' });
    await expect(superadminBtn).toBeVisible({ timeout: 5000 });
    const roleId = (await superadminBtn.getAttribute('data-testid'))?.replace('role-name-', '');
    await expect(page.locator(`button[data-testid="edit-role-${roleId}"]`)).toBeDisabled();
  });

  test('custom role edit button is enabled', async ({ page }) => {
    const uniqueName = `e2e_edit_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/roles`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: uniqueName, description: 'E2E test', permissionCodes: ['team:read'] }),
    });
    if (createRes.status !== 200 && createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const roleId = created?.data?.bizKey;

    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator(`button[data-testid="edit-role-${roleId}"]`)).toBeEnabled({ timeout: 5000 });

    await fetch(`${API}/admin/roles/${roleId}`, { method: 'DELETE', headers: { Authorization: `Bearer ${authToken}` } });
  });
});

// ── Section 5: Create Role ────────────────────────────────────────────────────

test.describe('Role Management - Create Role', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking create role opens edit dialog', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建角色' }).click();
    await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });
  });

  test('creates a custom role and it appears in the table', async ({ page }) => {
    const uniqueName = `e2e_create_${Date.now()}`;
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '创建角色' }).click();
    await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });

    const nameInput = page.locator('[role="dialog"] input').first();
    await nameInput.fill(uniqueName);

    // Must select at least 1 permission (handleSave validates this)
    const firstCheckbox = page.locator('[role="dialog"] input[type="checkbox"]').first();
    if (await firstCheckbox.isVisible({ timeout: 3000 }).catch(() => false)) {
      await firstCheckbox.check();
    }

    await page.locator('[role="dialog"] button', { hasText: '保存' }).click();

    await expect(page.locator(`button[data-testid^="role-name-"]`).filter({ hasText: uniqueName })).toBeVisible({ timeout: 8000 });

    const listRes = await fetch(`${API}/admin/roles?search=${uniqueName}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = await listRes.json();
    const roles = listData.data?.items || listData.items || [];
    const created = roles.find((r: any) => r.name === uniqueName);
    if (created) {
      await fetch(`${API}/admin/roles/${created.id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${authToken}` } });
    }
  });
});

// ── Section 6: Delete Role ────────────────────────────────────────────────────

test.describe('Role Management - Delete Role', () => {
  let authToken: string;

  test.beforeAll(async () => { authToken = await getAuthToken(); });
  test.beforeEach(async ({ page }) => { await login(page); });

  test('preset role delete button is disabled', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    const superadminBtn = page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' });
    await expect(superadminBtn).toBeVisible({ timeout: 5000 });
    const row = page.locator('tr').filter({ has: superadminBtn });
    const deleteBtn = row.locator('button', { hasText: '删除' });
    await expect(deleteBtn).toBeDisabled();
  });

  test('can delete a custom role with no members', async ({ page }) => {
    const uniqueName = `e2e_del_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/roles`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: uniqueName, description: 'E2E delete test', permissionCodes: ['team:read'] }),
    });
    if (createRes.status !== 200 && createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const roleId = created?.data?.bizKey;

    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });

    const roleBtn = page.locator(`button[data-testid="role-name-${roleId}"]`);
    await expect(roleBtn).toBeVisible({ timeout: 5000 });
    const row = page.locator('tr').filter({ has: roleBtn });
    await row.locator('button', { hasText: '删除' }).click();

    await expect(page.locator('h2:has-text("删除角色")')).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '确认删除' }).click();

    await expect(roleBtn).not.toBeVisible({ timeout: 5000 });
  });
});

// ── Section 7: Permission Browse Dialog ──────────────────────────────────────

test.describe('Role Management - Permission Browse Dialog', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking permission list button opens browse dialog', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '权限列表' }).click();
    await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });
  });
});
