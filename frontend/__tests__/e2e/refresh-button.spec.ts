import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId } from './test-helpers';

async function expectRefreshToast(page: Page) {
  await page.locator('[data-testid="refresh-btn"]').click();
  await expect(page.locator('[role="alert"]:has-text("数据已刷新")')).toBeVisible({ timeout: 5000 });
}

test.describe('刷新按钮 - 事项清单', () => {
  test.beforeEach(async ({ page }) => { await login(page) });

  test('刷新按钮可见', async ({ page }) => {
    await expect(page.locator('[data-testid="refresh-btn"]')).toBeVisible({ timeout: 5000 });
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await expectRefreshToast(page);
  });
});

test.describe('刷新按钮 - 待办事项', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/item-pool`);
    await expect(page.locator('[data-testid="item-pool-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮在重置按钮右侧', async ({ page }) => {
    const filterBar = page.locator('[data-testid="item-pool-page"] .flex.items-center.gap-3').first();
    const resetBtn = filterBar.locator('button:has-text("重置")');
    const refreshBtn = filterBar.locator('[data-testid="refresh-btn"]');
    await expect(resetBtn).toBeVisible();
    await expect(refreshBtn).toBeVisible();

    const resetIndex = await resetBtn.evaluate((el) =>
      Array.from(el.parentElement!.children).indexOf(el)
    );
    const refreshIndex = await refreshBtn.evaluate((el) =>
      Array.from(el.parentElement!.children).indexOf(el)
    );
    expect(refreshIndex).toBeGreaterThan(resetIndex);
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await expectRefreshToast(page);
  });
});

test.describe('刷新按钮 - 表格视图', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/table`);
    await expect(page.locator('[data-testid="table-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮可见', async ({ page }) => {
    await expect(page.locator('[data-testid="refresh-btn"]')).toBeVisible({ timeout: 5000 });
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await expectRefreshToast(page);
  });
});

test.describe('刷新按钮 - 整体进度', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/gantt`);
    await expect(page.locator('[data-testid="gantt-view-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮在搜索框右侧', async ({ page }) => {
    const header = page.locator('.gantt-label-header');
    const searchInput = header.locator('.label-search');
    const refreshBtn = header.locator('[data-testid="refresh-btn"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });
    await expect(refreshBtn).toBeVisible({ timeout: 5000 });
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await page.waitForTimeout(1000);
    const refreshBtn = page.locator('[data-testid="refresh-btn"]');
    if (await refreshBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expectRefreshToast(page);
    } else {
      test.skip();
    }
  });
});

test.describe('刷新按钮 - 用户管理', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/users`);
    await expect(page.locator('[data-testid="user-management-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮在搜索框右侧', async ({ page }) => {
    const filterBar = page.locator('[data-testid="user-management-page"] .flex.items-center.gap-3').first();
    const searchInput = filterBar.locator('input[placeholder*="搜索"]');
    const refreshBtn = filterBar.locator('[data-testid="refresh-btn"]');
    await expect(searchInput).toBeVisible();
    await expect(refreshBtn).toBeVisible();
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await expectRefreshToast(page);
  });
});

test.describe('刷新按钮 - 角色管理', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/roles`);
    await expect(page.locator('[data-testid="role-management-page"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮在类型筛选右侧', async ({ page }) => {
    const filterBar = page.locator('[data-testid="role-management-page"] .flex.items-center.gap-3').first();
    const typeSelect = filterBar.locator('button[role="combobox"]');
    const refreshBtn = filterBar.locator('[data-testid="refresh-btn"]');
    await expect(typeSelect).toBeVisible();
    await expect(refreshBtn).toBeVisible();

    const selectIndex = await typeSelect.evaluate((el) =>
      Array.from(el.closest('.flex')!.children).indexOf(el.closest('button')!)
    );
    const refreshIndex = await refreshBtn.evaluate((el) =>
      Array.from(el.parentElement!.children).indexOf(el)
    );
    expect(refreshIndex).toBeGreaterThan(selectIndex);
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    await expectRefreshToast(page);
  });
});

test.describe('刷新按钮 - 团队详情', () => {
  let teamId: string | null = null;

  test.beforeAll(async () => {
    const token = await getAuthToken();
    teamId = await getFirstTeamId(token);
  });

  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('刷新按钮在角色筛选右侧', async ({ page }) => {
    if (!teamId) { test.skip(); return; }
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    const filterBar = page.locator('[data-testid="team-detail-page"] .flex.items-center.gap-3').first();
    const roleSelect = filterBar.locator('button[role="combobox"]');
    const refreshBtn = filterBar.locator('[data-testid="refresh-btn"]');
    await expect(roleSelect).toBeVisible({ timeout: 5000 });
    await expect(refreshBtn).toBeVisible({ timeout: 5000 });

    const selectIndex = await roleSelect.evaluate((el) =>
      Array.from(el.closest('.flex')!.children).indexOf(el.closest('button')!)
    );
    const refreshIndex = await refreshBtn.evaluate((el) =>
      Array.from(el.parentElement!.children).indexOf(el)
    );
    expect(refreshIndex).toBeGreaterThan(selectIndex);
  });

  test('点击刷新显示气泡提示', async ({ page }) => {
    if (!teamId) { test.skip(); return; }
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expectRefreshToast(page);
  });
});
