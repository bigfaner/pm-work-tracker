import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/v1';

test.describe('团队成员权限测试', () => {
  let memberToken: string;
  let teamId: string;

  test.beforeAll(async () => {
    // 登录获取 token
    const res = await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'test', password: 'abcd1234' }),
    });
    const json = await res.json();
    memberToken = json.data.token;
    const testUserBizKey = json.data.user.bizKey;

    // 获取管理员 token，用于分配角色
    const adminRes = await fetch(`${API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    const adminToken = (await adminRes.json()).data.token;

    // 获取团队 ID
    const teamsRes = await fetch(`${API}/teams`, {
      headers: { Authorization: `Bearer ${memberToken}` },
    });
    const teamsJson = await teamsRes.json();
    teamId = String(teamsJson.data.items[0].bizKey);

    // 获取 member 角色的 bizKey
    const rolesRes = await fetch(`${API}/admin/roles`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    const rolesJson = await rolesRes.json();
    const memberRole = rolesJson.data.items.find((r: any) => r.roleName === 'member');
    const memberRoleKey = String(memberRole.bizKey);

    // 确保 test 用户在团队中拥有 member 角色（先加入，再更新角色）
    const addRes = await fetch(`${API}/teams/${teamId}/members`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${adminToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'test', roleKey: memberRoleKey }),
    });
    // If already a member (409 or similar), update role instead
    if (!addRes.ok) {
      await fetch(`${API}/teams/${teamId}/members/${testUserBizKey}/role`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${adminToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ roleKey: memberRoleKey }),
      });
    }
  });

  async function loginAsMember(page: any, token: string) {
    await page.goto(`${BASE}/login`);
    await page.evaluate((t) => {
      localStorage.setItem('auth-storage', JSON.stringify({
        state: {
          token: t,
          user: { isSuperAdmin: false },
          isAuthenticated: true,
          isSuperAdmin: false,
          permissions: null,
          permissionsLoadedAt: null,
          _hasHydrated: true,
        },
        version: 0,
      }));
    }, token);

    await page.goto(`${BASE}/items`);
    await page.waitForURL(/\/items/, { timeout: 10000 });
    
    // 等待权限加载
    await page.waitForFunction(() => {
      try {
        const raw = localStorage.getItem('auth-storage');
        if (!raw) return false;
        const parsed = JSON.parse(raw);
        return parsed?.state?.permissions !== null;
      } catch { return false; }
    }, { timeout: 10000 });
  }

  test('成员可以访问事项清单页面', async ({ page }) => {
    await loginAsMember(page, memberToken);
    await page.goto(`${BASE}/items`);
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问待办事项页面', async ({ page }) => {
    await loginAsMember(page, memberToken);
    await page.goto(`${BASE}/item-pool`);
    await expect(page.locator('[data-testid="item-pool-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问周视图页面', async ({ page }) => {
    await loginAsMember(page, memberToken);
    await page.goto(`${BASE}/weekly`);
    await expect(page.locator('[data-testid="weekly-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('成员可以访问团队管理页面', async ({ page }) => {
    await loginAsMember(page, memberToken);
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=团队名称')).toBeVisible({ timeout: 5000 });
    
    // 验证成员不能看到 PM 专属按钮
    await expect(page.locator('button', { hasText: '解散团队' })).not.toBeVisible({ timeout: 3000 });
    await expect(page.locator('button', { hasText: '添加成员' })).not.toBeVisible({ timeout: 3000 });
  });
});
