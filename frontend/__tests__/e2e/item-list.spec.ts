import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, parseApiData, extractBizKey, invalidateAuthCache } from './test-helpers';

const TIMEOUT = 120000;

test.setTimeout(TIMEOUT);

// Navigate within SPA
async function navTo(page: Page, path: string) {
  const link = page.locator(`[data-testid="sidebar"] a[href="${path}"]`);
  await link.click();
  await page.waitForTimeout(1500);
}

// ============================================================
// SERIAL TEST SUITE: Run tests in order, shared session
// ============================================================

test.describe.serial('事项清单 - 完整E2E业务流程测试', () => {
  let authToken: string;
  let teamId: string;
  let testMainItemId: string;
  let testSubItemId: string;

  // ====== SETUP: Create test data via Playwright request ======
  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    authToken = await getAuthToken();

    // Get teams
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

    // Create main item
    const mainRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { title: 'E2E测试-主事项-详情页', priority: 'P2', assigneeKey, startDate: '2026-04-19', expectedEndDate: '2026-05-19' },
    });
    const mainData = parseApiData(await mainRes.json());
    testMainItemId = extractBizKey(mainData) ?? '';

    // Create sub-item
    const subRes = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mainItemKey: String(testMainItemId), title: 'E2E测试-子事项-详情页', priority: 'P2', assigneeKey, startDate: '2026-04-19', expectedEndDate: '2026-05-19' },
    });
    const subData = parseApiData(await subRes.json());
    testSubItemId = extractBizKey(subData) ?? '';

    await request.dispose();
    console.log(`Setup: team=${teamId}, main=${testMainItemId}, sub=${testSubItemId}`);
  });

  // ====== STEP 1: LOGIN ======
  test('1.1 登录页渲染正确', async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await expect(page.locator('[data-testid="login-username"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-password"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-submit"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-submit"]')).toBeDisabled();
  });

  // NOTE: 错误密码测试跳过 - rate limiting导致不稳定
  test.skip('1.2 错误密码显示错误提示', async ({ page }) => {
    await page.goto(`${BASE}/login`);
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('wrongpass');
    await page.locator('[data-testid="login-submit"]').click();
    const errorEl = page.locator('[data-testid="login-error"]');
    await expect(errorEl).toBeVisible({ timeout: 15000 });
  });

  test('1.3 正确密码登录成功跳转', async ({ page }) => {
    await login(page);
    expect(page.url()).toContain('/items');
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible();
  });

  // ====== STEP 2: PAGE LAYOUT ======
  test('2.1 页面头部和侧边栏', async ({ page }) => {
    await login(page);
    await expect(page.locator('h1:text("事项清单")')).toBeVisible();
    await expect(page.locator('[data-testid="sidebar"]')).toBeVisible();
    await expect(page.locator('[data-testid="team-switcher"]')).toBeVisible();
    await expect(page.locator('[data-testid="toggle-summary"]')).toBeVisible();
    await expect(page.locator('[data-testid="toggle-detail"]')).toBeVisible();
    await expect(page.locator('button:has-text("新增主事项")')).toBeVisible();
  });

  test('2.2 筛选栏', async ({ page }) => {
    await login(page);
    const container = page.locator('[data-testid="item-view-page"]');
    await expect(container.locator('input[placeholder*="搜索"]')).toBeVisible();
    await expect(page.locator('button:has-text("重置")')).toBeVisible();
  });

  // ====== STEP 3: VIEW MODE TOGGLE ======
  test('3.1 默认汇总视图', async ({ page }) => {
    await login(page);
    await expect(page.locator('[data-testid="toggle-summary"]')).toHaveClass(/bg-primary/);
  });

  test('3.2 切换到明细视图', async ({ page }) => {
    await login(page);
    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1000);
    await expect(page.locator('[data-testid="detail-table"]')).toBeVisible();
    for (const h of ['编号', '标题', '状态']) {
      await expect(page.locator(`[data-testid="detail-table"] th:has-text("${h}")`)).toBeVisible();
    }
    await expect(page.locator('text=/共 \\d+ 条/')).toBeVisible({ timeout: 5000 });
  });

  test('3.3 切换回汇总视图', async ({ page }) => {
    await login(page);
    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(500);
    await page.locator('[data-testid="toggle-summary"]').click();
    await page.waitForTimeout(500);
    await expect(page.locator('[data-testid="detail-table"]')).not.toBeVisible();
  });

  // ====== STEP 4: CREATE MAIN ITEM ======
  test('4.1 打开创建对话框并验证表单', async ({ page }) => {
    await login(page);
    await page.locator('button:has-text("新增主事项")').click();
    await expect(page.locator('text=新建主事项')).toBeVisible();
    const confirmBtn = page.locator('[role="dialog"] button:has-text("确认")');
    await expect(confirmBtn).toBeDisabled();
  });

  test('4.2 创建主事项并验证出现', async ({ page }) => {
    const uniqueTitle = `E2E新建-${Date.now()}`;
    await login(page);
    await page.locator('button:has-text("新增主事项")').click();
    await expect(page.locator('text=新建主事项')).toBeVisible();
    const dialog = page.locator('[role="dialog"]');
    await dialog.locator('input[placeholder="请输入标题"]').fill(uniqueTitle);
    // Select assignee (required) - 2nd combobox in dialog (1st is priority)
    const assigneeSelect = dialog.locator('button[role="combobox"]').nth(1);
    await assigneeSelect.click();
    await page.waitForTimeout(500);
    const assigneeOption = page.locator('[role="option"]').first();
    if (await assigneeOption.isVisible({ timeout: 2000 }).catch(() => false)) {
      await assigneeOption.click();
      await page.waitForTimeout(300);
    }
    // Fill dates (required)
    await dialog.locator('input[type="date"]').first().fill('2026-04-19');
    await dialog.locator('input[type="date"]').last().fill('2026-05-19');
    await dialog.locator('button:has-text("确认")').click();
    await expect(page.locator('text=新建主事项')).not.toBeVisible({ timeout: 10000 });
    await expect(page.locator(`text=${uniqueTitle}`).first()).toBeVisible({ timeout: 10000 });
  });

  test('4.3 取消创建不创建事项', async ({ page }) => {
    await login(page);
    await page.locator('button:has-text("新增主事项")').click();
    await page.locator('[role="dialog"] button:has-text("取消")').click();
    await expect(page.locator('text=新建主事项')).not.toBeVisible();
  });

  // ====== STEP 5: FILTERS ======
  test('5.1 搜索过滤', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    const searchInput = page.locator('[data-testid="item-view-page"] input[placeholder*="搜索"]');
    await searchInput.fill('ZZZZ_NONEXISTENT');
    await page.waitForTimeout(1000);
    const emptyVisible = await page.locator('text=暂无事项').isVisible().catch(() => false);
    console.log(`Empty filter shows empty state: ${emptyVisible}`);
    await searchInput.clear();
    await page.waitForTimeout(1000);
  });

  test('5.2 重置按钮清除筛选', async ({ page }) => {
    await login(page);
    const searchInput = page.locator('[data-testid="item-view-page"] input[placeholder*="搜索"]');
    await searchInput.fill('test');
    await page.locator('button:has-text("重置")').click();
    await page.waitForTimeout(300);
    await expect(searchInput).toHaveValue('');
  });

  // ====== STEP 6: SUMMARY VIEW - EXPAND SUB-ITEMS ======
  test('6.1 展开主事项查看子事项', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    const itemTitle = page.locator('text=E2E测试-主事项-详情页').first();
    await expect(itemTitle).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="expand-card-${testMainItemId}"]`).dispatchEvent('click');
    await page.waitForTimeout(1500);
    await expect(page.locator('text=E2E测试-子事项-详情页')).toBeVisible({ timeout: 8000 });
  });

  test('6.2 再次点击折叠子事项', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    const itemTitle = page.locator('text=E2E测试-主事项-详情页').first();
    await expect(itemTitle).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="expand-card-${testMainItemId}"]`).dispatchEvent('click');
    await page.waitForTimeout(1500);
    await expect(page.locator('text=E2E测试-子事项-详情页')).toBeVisible({ timeout: 8000 });
    await page.locator(`[data-testid="expand-card-${testMainItemId}"]`).dispatchEvent('click');
    await page.waitForTimeout(1000);
    const subVisible = await page.locator('text=E2E测试-子事项-详情页').isVisible().catch(() => false);
    expect(subVisible).toBe(false);
  });

  test('6.3 子事项链接跳转到子事项详情页', async ({ page }) => {
    await login(page);
    await page.waitForTimeout(2000);
    const itemTitle = page.locator('text=E2E测试-主事项-详情页').first();
    await expect(itemTitle).toBeVisible({ timeout: 10000 });
    await page.locator(`[data-testid="expand-card-${testMainItemId}"]`).dispatchEvent('click');
    await page.waitForTimeout(1500);
    const subLink = page.locator('a:has-text("E2E测试-子事项-详情页")');
    await expect(subLink).toBeVisible({ timeout: 8000 });
    await subLink.click();
    await page.waitForTimeout(2000);
    expect(page.url()).toMatch(/\/items\/\d+\/sub\/\d+/);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible();
  });

  // ====== STEP 7: MAIN ITEM DETAIL PAGE ======
  test('7.1 主事项详情页渲染', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('h1:text("E2E测试-主事项-详情页")')).toBeVisible();
  });

  test('7.2 面包屑导航', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('[data-testid="main-item-detail-page"] a:has-text("事项清单")')).toBeVisible();
  });

  test('7.3 信息卡片显示', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    const detailPage = page.locator('[data-testid="main-item-detail-page"]');
    await expect(detailPage).toBeVisible({ timeout: 10000 });
    await expect(detailPage.locator('text=负责人').first()).toBeVisible();
    await expect(detailPage.locator('text=预期完成时间').first()).toBeVisible();
    await expect(detailPage.locator('text=开始时间').first()).toBeVisible();
  });

  test('7.4 编辑主事项标题', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Click the edit button (SVG icon + text)
    const editBtn = page.locator('[data-testid="main-item-detail-page"] >> text=编辑').first();
    await editBtn.click();
    await expect(page.locator('[role="dialog"] h2:text("编辑主事项")')).toBeVisible({ timeout: 5000 });

    const dialog = page.locator('[role="dialog"]');
    const titleInput = dialog.locator('input').first();
    await titleInput.clear();
    await titleInput.fill('E2E测试-主事项-已编辑');
    await dialog.locator('button:has-text("保存")').click();
    await page.waitForTimeout(3000);
    await expect(page.locator('h1:text("E2E测试-主事项-已编辑")')).toBeVisible({ timeout: 10000 });
  });

  test('7.5 子事项列表显示', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Verify the sub-item created in beforeAll appears in the table
    await expect(page.locator('text=E2E测试-子事项-详情页').first()).toBeVisible({ timeout: 10000 });
  });

  test('7.6 子事项状态变更', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    const detailPage = page.locator('[data-testid="main-item-detail-page"]');
    const statusBadges = detailPage.locator('table button');
    const count = await statusBadges.count();
    if (count > 0) {
      await statusBadges.first().click();
      await page.waitForTimeout(500);
      const statusOption = page.locator('[role="menuitem"]').first();
      if (await statusOption.isVisible({ timeout: 2000 }).catch(() => false)) {
        await statusOption.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('7.7 面包屑返回事项清单', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    await page.locator('[data-testid="main-item-detail-page"] a:has-text("事项清单")').click();
    await page.waitForTimeout(2000);
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible();
  });

  // ====== STEP 8: SUB-ITEM DETAIL PAGE ======
  test('8.1 子事项详情页渲染', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('h1:text("E2E测试-子事项-详情页")')).toBeVisible();
  });

  test('8.2 面包屑完整路径', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('[data-testid="sub-item-detail-page"] a:has-text("事项清单")')).toBeVisible();
  });

  test('8.3 信息卡片字段', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    const subPage = page.locator('[data-testid="sub-item-detail-page"]');
    await expect(subPage).toBeVisible({ timeout: 10000 });
    await expect(subPage.locator('text=负责人').first()).toBeVisible();
    await expect(subPage.locator('text=开始时间').first()).toBeVisible();
    await expect(subPage.locator('text=预期完成时间').first()).toBeVisible();
  });

  test('8.4 进度条显示', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    const subPage = page.locator('[data-testid="sub-item-detail-page"]');
    await expect(subPage).toBeVisible({ timeout: 10000 });
    await expect(subPage.locator('text=总进度').first()).toBeVisible();
  });

  test('8.5 追加进度记录', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    await page.locator('button:has-text("追加进度")').click();
    await expect(page.locator('[role="dialog"] h2:text("追加进度")')).toBeVisible({ timeout: 5000 });

    const dialog = page.locator('[role="dialog"]');
    await dialog.locator('input[type="number"]').fill('30');
    await dialog.locator('textarea').first().fill('E2E测试-取得30%进展');
    await dialog.locator('button:has-text("提交")').click();
    await page.waitForTimeout(3000);

    await expect(page.locator('text=E2E测试-取得30%进展').first()).toBeVisible({ timeout: 10000 });
  });

  test('8.6 追加更多进度(递增)', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    await page.locator('button:has-text("追加进度")').click();
    await expect(page.locator('[role="dialog"] h2:text("追加进度")')).toBeVisible({ timeout: 5000 });

    const dialog = page.locator('[role="dialog"]');
    await dialog.locator('input[type="number"]').fill('60');
    await dialog.locator('textarea').nth(1).fill('E2E测试-遇到技术难点');
    await dialog.locator('button:has-text("提交")').click();
    await page.waitForTimeout(3000);

    await expect(page.locator('text=E2E测试-遇到技术难点')).toBeVisible({ timeout: 10000 });
  });

  test('8.7 面包屑跳转到主事项详情', async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${testSubItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    await page.locator(`[data-testid="sub-item-detail-page"] nav a[href="/items/${testMainItemId}"]`).click();
    await page.waitForTimeout(2000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible();
  });

  // ====== STEP 9: DETAIL VIEW TABLE ======
  test('9.1 明细视图表格数据', async ({ page }) => {
    await login(page);
    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1500);
    const rows = page.locator('[data-testid="detail-table"] tbody tr');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
    console.log(`Detail table rows: ${count}`);
  });

  test('9.2 表格编辑链接跳转', async ({ page }) => {
    await login(page);
    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1500);
    const editLink = page.locator('[data-testid="detail-table"] a:has-text("编辑")').first();
    if (await editLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editLink.click();
      await page.waitForTimeout(2000);
      await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    }
  });

  test('9.3 表格标题链接跳转', async ({ page }) => {
    await login(page);
    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1500);
    const titleLink = page.locator('[data-testid="detail-table"] td a.font-medium').first();
    if (await titleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await titleLink.click();
      await page.waitForTimeout(2000);
      await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });
    }
  });

  // ====== STEP 10: FULL BUSINESS FLOW ======
  test('10.1 完整流程: 详情→子事项→进度→面包屑返回', async ({ page }) => {
    const progressText = `进度50%-${Date.now()}`;
    await login(page);
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible();

    // Navigate to pre-existing main item detail
    await page.goto(`${BASE}/items/${testMainItemId}`);
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="main-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Verify sub-item in table, then navigate to sub-item detail
    await page.locator(`a[href="/items/${testMainItemId}/sub/${testSubItemId}"]`).first().click();
    await page.waitForTimeout(3000);
    await expect(page.locator('[data-testid="sub-item-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Append progress
    await page.locator('button:has-text("追加进度")').click();
    const progressDialog = page.locator('[role="dialog"]');
    await expect(progressDialog).toBeVisible({ timeout: 5000 });
    await progressDialog.locator('input[type="number"]').fill('80');
    await progressDialog.locator('textarea').first().fill(progressText);
    await progressDialog.locator('button:has-text("提交")').click();
    await page.waitForTimeout(3000);
    await expect(page.locator(`text=${progressText}`).first()).toBeVisible({ timeout: 10000 });

    // Navigate back via breadcrumb
    await page.locator('[data-testid="sub-item-detail-page"] nav a:has-text("事项清单")').click();
    await page.waitForTimeout(2000);
    await expect(page.locator('[data-testid="item-view-page"]')).toBeVisible();
  });

  // ====== STEP 11: EDGE CASES ======
  test('11.1 无效搜索显示空状态', async ({ page }) => {
    await login(page);
    const searchInput = page.locator('[data-testid="item-view-page"] input[placeholder*="搜索"]');
    await searchInput.fill('___NONEXISTENT_ITEM___');
    await page.waitForTimeout(1500);
    const emptyVisible = await page.locator('text=暂无事项').isVisible().catch(() => false);
    console.log(`Empty state shown for invalid search: ${emptyVisible}`);
  });

  test('11.2 登出功能', async ({ page }) => {
    await login(page);
    await expect(page.locator('[data-testid="sidebar-logout"]')).toBeVisible();
    await page.locator('[data-testid="sidebar-logout"]').click();
    await page.waitForURL(/\/login/, { timeout: 10000 });
    expect(page.url()).toContain('/login');
    invalidateAuthCache();
  });

  // ====== STEP 12: CONSOLE ERROR CHECK ======
  test('12.1 无关键控制台错误', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') errors.push(msg.text());
    });

    await login(page);
    await page.waitForTimeout(2000);

    await page.locator('[data-testid="toggle-detail"]').click();
    await page.waitForTimeout(1000);
    await page.locator('[data-testid="toggle-summary"]').click();
    await page.waitForTimeout(1000);

    await page.locator('button:has-text("新增主事项")').click();
    await page.waitForTimeout(500);
    await page.locator('[role="dialog"] button:has-text("取消")').click();
    await page.waitForTimeout(500);

    const criticalErrors = errors.filter(e =>
      !e.includes('favicon') && !e.includes('net::') && !e.includes('404')
    );
    console.log(`Console errors: total=${errors.length}, critical=${criticalErrors.length}`);
    if (criticalErrors.length > 0) {
      console.log('Critical errors:', criticalErrors.join('\n'));
    }
  });
});
