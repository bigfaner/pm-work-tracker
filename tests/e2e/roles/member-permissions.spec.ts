import { test, expect } from '@playwright/test';
import { BASE, API, getAuthToken, getFirstTeamId, getRoleKey, createUser, deleteUser, addUserToTeam, loginAsUser } from '../helpers.js';

test.describe('团队成员权限测试', () => {
  let memberToken: string;
  let teamId: string;
  let testUsername: string;
  let testUserBizKey: string;

  test.beforeAll(async () => {
    const adminToken = await getAuthToken();

    testUsername = `test_member_${Date.now()}`;
    const user = await createUser(adminToken, testUsername, 'Test Member');
    testUserBizKey = user.userId;
    const initialPassword = user.initialPassword!;

    memberToken = await (await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: testUsername, password: initialPassword }),
    })).json().then((j: any) => j.data.token);

    teamId = (await getFirstTeamId(adminToken))!;
    const memberRoleKey = (await getRoleKey(adminToken, 'member'))!;
    await addUserToTeam(adminToken, teamId!, testUsername, memberRoleKey);
  });

  test.afterAll(async () => {
    if (!testUserBizKey) return;
    const adminToken = await getAuthToken();
    await deleteUser(adminToken, testUserBizKey);
  });

  test('成员可以访问事项清单页面', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false });
    await page.goto(`${BASE}/items`);
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问待办事项页面', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false });
    await page.goto(`${BASE}/item-pool`);
    await expect(page.locator('[data-testid="item-pool-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问周视图页面', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false });
    await page.goto(`${BASE}/weekly`);
    await expect(page.locator('[data-testid="weekly-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问团队管理页面', async ({ page }) => {
    await loginAsUser(page, memberToken, { isSuperAdmin: false });
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=团队名称')).toBeVisible({ timeout: 5000 });

    await expect(page.locator('button', { hasText: '解散团队' })).not.toBeVisible({ timeout: 3000 });
    await expect(page.locator('button', { hasText: '添加成员' })).not.toBeVisible({ timeout: 3000 });
  });
});
