import { test, expect, Page } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/api/v1';
const TIMEOUT = 60000;

test.setTimeout(TIMEOUT);

function parseApiData(resp: any): any {
  return resp.data !== undefined ? resp.data : resp;
}

async function login(page: Page) {
  for (let attempt = 0; attempt < 3; attempt++) {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    try {
      await page.waitForURL(/\/items/, { timeout: 10000 });
      await page.waitForTimeout(1000);
      return;
    } catch {
      if (attempt < 2) await page.waitForTimeout(3000);
      else throw new Error('Login failed after 3 attempts');
    }
  }
}

test.describe.serial('事项清单 Bug修复验证', () => {
  let authToken: string;
  let teamId: string;
  let itemA: string;
  let itemB: string;
  let subItemA1: string;

  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const loginRes = await request.post('/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    });
    const loginData = parseApiData(await loginRes.json());
    authToken = loginData.token;

    const teamsRes = await request.get('/api/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsData = parseApiData(await teamsRes.json());
    teamId = String(teamsData[0].id);

    // Create main item with dates (use snake_case field names for backend)
    const mainARes = await request.post(`/api/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: 'E2E修复测试-带日期',
        priority: 'P1',
        start_date: '2026-04-01',
        expected_end_date: '2026-04-30',
      },
    });
    const mainAData = parseApiData(await mainARes.json());
    itemA = String(mainAData.id);

    // Create sub-item (backend DTO requires main_item_id, snake_case fields)
    const subRes = await request.post(`/api/v1/teams/${teamId}/main-items/${itemA}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        main_item_id: Number(itemA),
        title: '子事项-已有',
        priority: 'P1',
        assignee_id: Number(loginData.user.id),
      },
    });
    const subData = parseApiData(await subRes.json());
    subItemA1 = String(subData.id);

    // Create main item B (no dates)
    const mainBRes = await request.post(`/api/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { title: 'E2E修复测试-无日期', priority: 'P2' },
    });
    const mainBData = parseApiData(await mainBRes.json());
    itemB = String(mainBData.id);

    await request.dispose();
    console.log(`Setup: team=${teamId}, itemA=${itemA}, itemB=${itemB}, subA1=${subItemA1}`);
  });

  test('显示主事项的计划时间周期', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });
    // Date format uses formatDate which replaces - with /
    await expect(page.locator('text=2026/04/01').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=2026/04/30').first()).toBeVisible({ timeout: 5000 });
  });

  test('展开单个主事项不影响其他事项', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Click the card row to expand item A
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.click();
    await page.waitForTimeout(3000);

    // Item A shows sub-items
    await expect(page.locator('text=子事项-已有').first()).toBeVisible({ timeout: 10000 });

    // Item B should NOT show expanded content
    const bCard = page.locator(`:text("E2E修复测试-无日期")`).first().locator('..').locator('..');
    await expect(bCard.locator('text=暂无子事项')).not.toBeVisible();
  });

  test('编辑按钮正确跳转到详情页', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Find and click edit button for item A (button is inside the card row)
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.locator('button:has-text("编辑")').first().click();
    await page.waitForTimeout(2000);

    await expect(page).toHaveURL(new RegExp(`/items/${itemA}`), { timeout: 10000 });
  });

  test('新增子事项按钮打开表单对话框', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Click "新增子事项" button (visible in the card row)
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.locator('button:has-text("新增子事项")').first().click();
    await page.waitForTimeout(1000);

    // Dialog should appear with form
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.locator('h2:text("新增子事项")')).toBeVisible();
    await expect(dialog.locator('input[placeholder="请输入子事项标题"]')).toBeVisible();

    // Close dialog
    await dialog.locator('button:has-text("取消")').click();
  });

  test('通过对话框创建子事项', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Click "新增子事项"
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.locator('button:has-text("新增子事项")').first().click();
    await page.waitForTimeout(1000);

    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Fill form
    const subTitle = `子事项-新建-${Date.now()}`;
    await dialog.locator('input[placeholder="请输入子事项标题"]').fill(subTitle);

    // Select an assignee (required) - click the trigger showing "不指定"
    await dialog.locator('button:has-text("不指定")').click();
    await page.waitForTimeout(500);
    // Select a real member option (not "不指定") - options are in a portal
    const memberOption = page.locator('[role="option"]').filter({ hasNotText: '不指定' }).first();
    await memberOption.click();
    await page.waitForTimeout(500);

    // Submit (note: assignee may be required)
    await dialog.locator('button:has-text("确认")').click();
    await page.waitForTimeout(3000);

    // Dialog should close
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 5000 });

    // Expand item A and verify new sub-item appears
    const itemACard2 = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard2.click();
    await page.waitForTimeout(3000);
    await expect(page.locator(`text=${subTitle}`).first()).toBeVisible({ timeout: 10000 });
  });

  test('再次点击折叠已展开的事项', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Click card to expand
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.click();
    await page.waitForTimeout(3000);
    await expect(page.locator('text=子事项-已有').first()).toBeVisible({ timeout: 10000 });

    // Click again to collapse
    await itemACard.click();
    await page.waitForTimeout(1000);
    await expect(page.locator('text=子事项-已有').first()).not.toBeVisible({ timeout: 5000 });
  });

  test('明细视图中编辑按钮跳转正常', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1000);
    await expect(page.locator('[data-testid="detail-table"]')).toBeVisible();

    const tableRow = page.locator(`tr:has-text("E2E修复测试-带日期")`).first();
    await tableRow.locator('button:has-text("编辑")').click();
    await page.waitForTimeout(2000);

    await expect(page).toHaveURL(new RegExp(`/items/${itemA}$`), { timeout: 10000 });
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });
  });
});
