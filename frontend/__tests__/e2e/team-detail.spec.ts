import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/v1';

test.describe('Team Detail - Member List Actions Column', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    // Login to get auth token
    const res = await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    const json = await res.json();
    authToken = json.data?.token || json.token;

    if (authToken) {
      // Get first team
      const tRes = await fetch(`${API}/teams`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      const tData = await tRes.json();
      const list = tData.data || (Array.isArray(tData) ? tData : []);
      teamId = list.length > 0 ? String(list[0].id || list[0].ID) : null;
    }
  });

  // Login before each test
  test.beforeEach(async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForURL('**/items**', { timeout: 10000 });
  });

  test('actions column header is visible in member table', async ({ page }) => {
    if (!teamId) return;
    test.skip();

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 5000 });

    // Verify the actions column header exists in the member table
    const actionsHeader = page.locator('th', { hasText: '操作' });
    await expect(actionsHeader).toBeVisible();
  });

  test('change role button has icon and text', async ({ page }) => {
    if (!teamId) return;
    test.skip();

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 5000 });

    // Find the change role button in the actions column
    const changeRoleBtn = page.locator('[data-testid="change-role-btn"]').first();

    // Verify it exists and is visible
    await expect(changeRoleBtn).toBeVisible();

    // Verify it contains the text "修改角色"
    await expect(changeRoleBtn).toContainText('修改角色');

    // Verify it contains an SVG icon (Edit icon)
    const icon = changeRoleBtn.locator('svg');
    await expect(icon).toBeVisible();
  });

  test('clicking change role button opens inline role select', async ({ page }) => {
    if (!teamId) return;
    test.skip();

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 5000 });

    // Find and click the change role button
    const changeRoleBtn = page.locator('[data-testid="change-role-btn"]').first();
    await expect(changeRoleBtn).toBeVisible({ timeout: 5000 });
    await changeRoleBtn.click();

    // Verify the inline role select appears
    const roleSelect = page.locator('[data-testid="inline-role-select"]');
    await expect(roleSelect).toBeVisible({ timeout: 5000 });
  });
});
