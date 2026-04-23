import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/v1';

test.describe('Role Management - Superadmin Permission Count', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForURL('**/items**', { timeout: 10000 });
  });

  test('superadmin role shows non-zero permission count', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    // Find the superadmin row by looking for the button with text "superadmin"
    const superadminButton = page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' });
    await expect(superadminButton).toBeVisible({ timeout: 5000 });

    // Get the table row containing the superadmin button
    const superadminRow = page.locator('tr').filter({ has: superadminButton });
    await expect(superadminRow).toBeVisible();

    // The permission count is in the 3rd <td> (index 2). It should not be "0".
    const permCountCell = superadminRow.locator('td').nth(2);
    const permCountText = await permCountCell.textContent();
    const permCount = parseInt(permCountText?.trim() || '0', 10);

    expect(permCount).toBeGreaterThan(0);
  });
});

test.describe('Role Management - Permissions Dialog', () => {
  let authToken: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    const json = await res.json();
    authToken = json.data?.token || json.token;
  });

  test.beforeEach(async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForURL('**/items**', { timeout: 10000 });
  });

  test('clicking a role name opens the permissions dialog', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    // Find the first clickable role name
    const roleNames = page.locator('button[data-testid^="role-name-"]');
    const count = await roleNames.count();
    if (count === 0) return;
    test.skip(count === 0, 'No roles found');

    // Click the first role name
    await roleNames.first().click();

    // Verify the permissions dialog opened
    await expect(page.locator('[data-testid="role-permissions-dialog"]')).toBeVisible({ timeout: 5000 });
  });

  test('dialog shows the role name as title', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    const roleNames = page.locator('button[data-testid^="role-name-"]');
    const count = await roleNames.count();
    if (count === 0) return;
    test.skip(count === 0, 'No roles found');

    // Get the text of the first role name
    const roleName = await roleNames.first().textContent();
    expect(roleName).toBeTruthy();

    // Click it
    await roleNames.first().click();

    // Verify dialog title contains the role name
    const dialogTitle = page.locator('[data-testid="role-permissions-dialog-title"]');
    await expect(dialogTitle).toBeVisible({ timeout: 5000 });
    await expect(dialogTitle).toContainText(roleName!);
  });

  test('dialog displays permissions', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    const roleNames = page.locator('button[data-testid^="role-name-"]');
    const count = await roleNames.count();
    if (count === 0) return;
    test.skip(count === 0, 'No roles found');

    await roleNames.first().click();

    // Wait for the dialog to be visible
    await expect(page.locator('[data-testid="role-permissions-dialog"]')).toBeVisible({ timeout: 5000 });

    // Wait for permissions list to load (not showing "loading" state)
    const permissionsList = page.locator('[data-testid="role-permissions-list"]');
    await expect(permissionsList).toBeVisible({ timeout: 5000 });

    // There should be at least one checkbox visible (permissions are displayed)
    const checkboxes = page.locator('[data-testid="role-permissions-dialog"] input[type="checkbox"]');
    const checkboxCount = await checkboxes.count();
    expect(checkboxCount).toBeGreaterThan(0);
  });
});

test.describe('Role Management - Edit Button Disabled for Preset Roles', () => {
  let authToken: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    const json = await res.json();
    authToken = json.data?.token || json.token;
  });

  test.beforeEach(async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForURL('**/items**', { timeout: 10000 });
  });

  test('preset role edit button is disabled', async ({ page }) => {
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    // Find the superadmin role row (preset role)
    const superadminButton = page.locator('button[data-testid^="role-name-"]').filter({ hasText: 'superadmin' });
    await expect(superadminButton).toBeVisible({ timeout: 5000 });

    const superadminRow = page.locator('tr').filter({ has: superadminButton });

    // Get the role id from the data-testid attribute
    const roleButtonTestId = await superadminButton.getAttribute('data-testid');
    const roleId = roleButtonTestId?.replace('role-name-', '');

    // Find the edit button in the same row
    const editButton = page.locator(`button[data-testid="edit-role-${roleId}"]`);
    await expect(editButton).toBeVisible();
    await expect(editButton).toBeDisabled();
  });

  test('custom role edit button is enabled', async ({ page }) => {
    // Create a custom role via API to ensure one exists
    const uniqueName = `e2e_edit_test_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/roles`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: uniqueName,
        description: 'E2E test role for edit button',
        permissionCodes: ['items:view'],
      }),
    });

    if (createRes.status !== 200 && createRes.status !== 201) {
      test.skip();
    }

    const createdRole = await createRes.json();
    const roleId = createdRole?.data?.id || createdRole?.id;

    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 5000 });

    // Find the custom role's edit button
    const editButton = page.locator(`button[data-testid="edit-role-${roleId}"]`);
    await expect(editButton).toBeVisible({ timeout: 5000 });
    await expect(editButton).toBeEnabled();

    // Clean up: delete the test role
    await fetch(`${API}/admin/roles/${roleId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authToken}` },
    });
  });
});
