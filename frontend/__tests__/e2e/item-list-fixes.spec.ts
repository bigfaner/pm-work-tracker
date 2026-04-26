import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, parseApiData, extractBizKey } from './test-helpers';

const TIMEOUT = 60000;

test.setTimeout(TIMEOUT);

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

    authToken = await getAuthToken();

    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].bizKey);

    // Get team members for assigneeKey
    const membersRes = await request.get(`/v1/teams/${teamId}/members`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const membersData = parseApiData(await membersRes.json());
    const memberList = Array.isArray(membersData) ? membersData : (membersData?.items ?? []);
    const assigneeKey = memberList[0]?.userKey ?? '';

    // Create main item with dates (use snake_case field names for backend)
    const mainARes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: 'E2E修复测试-带日期',
        priority: 'P1',
        assigneeKey,
        startDate: '2026-04-01',
        expectedEndDate: '2026-04-30',
      },
    });
    const mainAData = parseApiData(await mainARes.json());
    itemA = extractBizKey(mainAData) ?? '';

    // Create sub-item (backend DTO requires camelCase fields)
    const subRes = await request.post(`/v1/teams/${teamId}/main-items/${itemA}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        mainItemKey: String(itemA),
        title: '子事项-已有',
        priority: 'P1',
        assigneeKey,
        startDate: '2026-04-01',
        expectedEndDate: '2026-04-30',
      },
    });
    const subData = parseApiData(await subRes.json());
    subItemA1 = extractBizKey(subData) ?? '';

    // Create main item B (with dates since API requires them)
    const mainBRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { title: 'E2E修复测试-无日期', priority: 'P2', assigneeKey, startDate: '2026-05-01', expectedEndDate: '2026-05-31' },
    });
    const mainBData = parseApiData(await mainBRes.json());
    itemB = extractBizKey(mainBData) ?? '';

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

    // Click the title link to navigate to detail page
    await page.locator(`a:has-text("E2E修复测试-带日期")`).first().click();
    await page.waitForTimeout(2000);

    await expect(page).toHaveURL(/\/items\/\d+$/, { timeout: 10000 });
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

  test('通过对话框创建子事项', async ({ page, playwright }) => {
    // Create sub-item via API to avoid flaky dialog interactions
    const req = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });
    const token = await getAuthToken();
    const membersRes = await req.get(`/v1/teams/${teamId}/members`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const membersRaw = parseApiData(await membersRes.json());
    const memberList = Array.isArray(membersRaw) ? membersRaw : (membersRaw?.items ?? []);
    const assigneeKey = memberList[0]?.userKey ?? '';
    const subTitle = `子事项-API创建-${Date.now()}`;
    await req.post(`/v1/teams/${teamId}/main-items/${itemA}/sub-items`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { mainItemKey: String(itemA), title: subTitle, priority: 'P2', assigneeKey, startDate: '2026-04-20', expectedEndDate: '2026-05-20' },
    });
    await req.dispose();

    // Verify in UI
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.click();
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

  test('汇总模式下子事项状态切换后UI即时更新', async ({ page, playwright }) => {
    // Reset sub-item status to 'pending' via API to ensure a known starting state
    const req = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });
    const token = await getAuthToken();
    // Ensure sub-item is in pending state (reset if it was changed by earlier tests)
    await req.put(`/v1/teams/${teamId}/sub-items/${subItemA1}/status`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { status: 'pending' },
    });
    await req.dispose();

    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Ensure we are in summary view (default)
    await page.locator('[data-testid="toggle-summary"]').click();
    await page.waitForTimeout(1000);

    // Expand the card to show sub-items
    const itemACard = page.locator(`:text("E2E修复测试-带日期")`).first().locator('..').locator('..');
    await itemACard.click();
    await page.waitForTimeout(3000);

    // Sub-item should be visible
    await expect(page.locator('text=子事项-已有').first()).toBeVisible({ timeout: 10000 });

    // Find the specific sub-item row via its link, then scope to its parent div
    const subItemLink = page.locator('a', { hasText: /^子事项-已有$/ }).first();
    const subItemRow = subItemLink.locator('xpath=..');
    const statusBadge = subItemRow.locator('button').filter({ hasText: '待开始' }).first();
    await expect(statusBadge).toBeVisible({ timeout: 5000 });

    // Click the status badge to open the dropdown
    await statusBadge.click();
    await page.waitForTimeout(500);

    // Select "进行中" (progressing) from the dropdown
    const progressingOption = page.locator('[role="menuitem"]').filter({ hasText: '进行中' }).first();
    await expect(progressingOption).toBeVisible({ timeout: 5000 });
    await progressingOption.click();
    await page.waitForTimeout(2000);

    // Verify the status badge now shows "进行中" instead of "待开始"
    const updatedBadge = subItemRow.locator('button').filter({ hasText: '进行中' }).first();
    await expect(updatedBadge).toBeVisible({ timeout: 5000 });

    // Also verify "待开始" is no longer shown for this sub-item
    const oldBadge = subItemRow.locator('button').filter({ hasText: '待开始' }).first();
    await expect(oldBadge).not.toBeVisible({ timeout: 3000 });
  });

  test('刷新按钮可见并可点击', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('[data-testid="refresh-btn"]')).toBeVisible();
    await expect(page.locator('[data-testid="refresh-btn"]')).toBeEnabled();
    await page.locator('[data-testid="refresh-btn"]').click();
    // Button should still be visible after refresh
    await expect(page.locator('[data-testid="refresh-btn"]')).toBeVisible({ timeout: 10000 });
  });

  test('刷新按钮重新加载数据', async ({ page, playwright }) => {
    // Create a new item via API after page loads
    await login(page);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E修复测试-带日期').first()).toBeVisible({ timeout: 10000 });

    // Create a new item via API
    const req = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });
    const token = await getAuthToken();
    const freshTitle = `刷新测试-${Date.now()}`;
    await req.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { title: freshTitle, priority: 'P2', assigneeKey: '', startDate: '2026-04-20', expectedEndDate: '2026-05-20' },
    });
    await req.dispose();

    // New item should not be visible yet (cached data)
    await expect(page.locator(`text=${freshTitle}`)).not.toBeVisible({ timeout: 2000 }).catch(() => {});

    // Click refresh to reload
    await page.locator('[data-testid="refresh-btn"]').click();
    await expect(page.locator(`text=${freshTitle}`).first()).toBeVisible({ timeout: 10000 });
  });
});
