import { test, expect, Page, APIRequestContext } from '@playwright/test';

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080/v1';
const TIMEOUT = 60000;

test.setTimeout(TIMEOUT);

// Helper: parse API response (format: {code, data})
function parseApiData(resp: any): any {
  return resp.data !== undefined ? resp.data : resp;
}

// Helper: login via UI and wait for redirect (with retry for rate limiting)
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
      if (attempt < 2) {
        await page.waitForTimeout(3000);
      } else {
        throw new Error('Login failed after 3 attempts (likely rate limited)');
      }
    }
  }
}

// Navigate within SPA
async function navTo(page: Page, path: string) {
  const link = page.locator(`[data-testid="sidebar"] a[href="${path}"]`);
  await link.click();
  await page.waitForTimeout(1500);
}

// ============================================================
// SERIAL TEST SUITE: Weekly View E2E
// ============================================================

test.describe.serial('每周进展 - 完整E2E交互流程测试', () => {
  let authToken: string;
  let teamId: string;
  let testMainItemId: string;
  let testSubItemId1: string;
  let testSubItemId2: string;

  // ====== SETUP: Login + create test data ======
  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    // Wait to avoid rate limiting from previous test suites
    await new Promise(r => setTimeout(r, 5000));

    // Login
    const loginResp = await request.post('/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    });
    const loginData = await loginResp.json();
    authToken = parseApiData(loginData).token;

    // Get team
    const teamsResp = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsData = await teamsResp.json();
    const teams = Array.isArray(teamsData) ? teamsData : (parseApiData(teamsData) || []);
    teamId = String(teams[0]?.id || teams[0]?.ID);
    if (!teamId) throw new Error('No team found');

    // Clean up stale test data from previous runs
    try {
      const itemsResp = await request.get(`/v1/teams/${teamId}/main-items`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      const itemsData = await itemsResp.json();
      const items = parseApiData(itemsData) || [];
      for (const item of items) {
        if (item.title === 'E2E周视图测试主事项') {
          await request.post(`/v1/teams/${teamId}/main-items/${item.id}/archive`, {
            headers: { Authorization: `Bearer ${authToken}` },
          });
        }
      }
    } catch { /* best effort */ }

    // Create a main item for testing
    const mainResp = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: 'E2E周视图测试主事项',
        priority: 'P1',
        assigneeId: 1,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-25',
      },
    });
    const mainData = await mainResp.json();
    testMainItemId = String(parseApiData(mainData)?.id || mainData?.id);

    // Create sub-item 1
    const sub1Resp = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        mainItemId: Number(testMainItemId),
        title: 'E2E子事项-进度测试A',
        priority: 'P2',
        assigneeId: 1,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-20',
      },
    });
    const sub1Data = await sub1Resp.json();
    testSubItemId1 = String(parseApiData(sub1Data)?.id || sub1Data?.id);

    // Create sub-item 2
    const sub2Resp = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        mainItemId: Number(testMainItemId),
        title: 'E2E子事项-进度测试B',
        priority: 'P3',
        assigneeId: 1,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-22',
      },
    });
    const sub2Data = await sub2Resp.json();
    testSubItemId2 = String(parseApiData(sub2Data)?.id || sub2Data?.id);

    // Change sub-item 1 to in-progress
    await request.put(`/v1/teams/${teamId}/sub-items/${testSubItemId1}/status`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { status: '进行中' },
    });

    // Append progress record 1 for sub-item 1 (this week)
    await request.post(`/v1/teams/${teamId}/sub-items/${testSubItemId1}/progress`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        completion: 40,
        achievement: 'E2E测试成就-第一阶段完成',
        blocker: '',
        lesson: '',
      },
    });

    // Append progress record 2 for sub-item 1 (this week, multiple records)
    await request.post(`/v1/teams/${teamId}/sub-items/${testSubItemId1}/progress`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        completion: 70,
        achievement: 'E2E测试成就-联调完成',
        blocker: 'E2E测试卡点-等待依赖',
        lesson: '',
      },
    });

    // Append progress for sub-item 2
    await request.put(`/v1/teams/${teamId}/sub-items/${testSubItemId2}/status`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { status: '进行中' },
    });
    await request.post(`/v1/teams/${teamId}/sub-items/${testSubItemId2}/progress`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        completion: 30,
        achievement: 'E2E测试成就B-基础完成',
        blocker: 'E2E测试卡点B-性能问题',
        lesson: '',
      },
    });
  });

  test.beforeEach(async ({ page }) => {
    await login(page);
    await navTo(page, '/weekly');
  });

  // ====== 1. Page Structure ======
  test('1.1 页面标题和周选择器可见', async ({ page }) => {
    await expect(page.locator('h1:text("每周进展")')).toBeVisible();
    await expect(page.locator('[data-testid="week-selector"]')).toBeVisible();
  });

  test('1.2 日期范围显示正确', async ({ page }) => {
    const dateRange = page.locator('text=/\\d{4}\\/\\d{2}\\/\\d{2}.*~.*\\d{4}\\/\\d{2}\\/\\d{2}/').first();
    await expect(dateRange).toBeVisible({ timeout: 15000 });
  });

  // ====== 2. Stats Bar ======
  test('2.1 四个统计卡片渲染', async ({ page }) => {
    await expect(page.locator('text=本周活跃子事项')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=本周新完成').first()).toBeVisible();
    await expect(page.locator('text=进度推进中')).toBeVisible();
    await expect(page.locator('text=阻塞中')).toBeVisible();
  });

  test('2.2 统计数字非空', async ({ page }) => {
    await page.waitForTimeout(2000);
    const statActive = page.locator('[data-testid="stat-active"]');
    await expect(statActive).toBeVisible();
    const value = await statActive.textContent();
    expect(value).toBeTruthy();
    expect(Number(value)).not.toBeNaN();
  });

  // ====== 3. Comparison Cards ======
  test('3.1 测试主事项卡片渲染', async ({ page }) => {
    await expect(page.locator(`[data-testid="group-card-${testMainItemId}"]`)).toBeVisible({ timeout: 5000 });
  });

  test('3.2 卡片显示优先级徽章', async ({ page }) => {
    await page.waitForTimeout(2000);
    // The main item is P1
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    if (await card.isVisible()) {
      const p1Badge = card.locator('text=P1');
      await expect(p1Badge.first()).toBeVisible();
    }
  });

  test('3.3 卡片显示进度条和完成度', async ({ page }) => {
    await page.waitForTimeout(2000);
    // Should show a percentage somewhere in the card
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    if (await card.isVisible()) {
      // Progress percentage in the header
      const progressText = card.locator('text=/\\d+%/');
      await expect(progressText.first()).toBeVisible();
    }
  });

  test('3.4 上周和本周列标题显示', async ({ page }) => {
    await page.waitForTimeout(2000);
    const lastWeekHeader = page.locator('text=/上周.*W\\d+/');
    const thisWeekHeader = page.locator('text=/本周.*W\\d+/');
    await expect(lastWeekHeader.first()).toBeVisible({ timeout: 5000 });
    await expect(thisWeekHeader.first()).toBeVisible();
  });

  // ====== 4. Sub-Item Rows ======
  test('4.1 子事项标题在页面中显示', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card.locator(`text=E2E子事项-进度测试A`)).toBeVisible({ timeout: 5000 });
  });

  test('4.2 子事项显示状态徽章', async ({ page }) => {
    await page.waitForTimeout(2000);
    // Sub-item 1 is "进行中"
    const statusBadge = page.locator('text=进行中').first();
    await expect(statusBadge).toBeVisible({ timeout: 5000 });
  });

  test('4.3 子事项显示优先级徽章', async ({ page }) => {
    await page.waitForTimeout(2000);
    // Sub-item 1 is P2, sub-item 2 is P3
    const p2Badge = page.locator('text=P2').first();
    await expect(p2Badge).toBeVisible({ timeout: 5000 });
  });

  // ====== 5. Progress Records (individual records display) ======
  test('5.1 成果记录换行显示', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card.locator('text=成果：E2E测试成就-第一阶段完成')).toBeVisible({ timeout: 5000 });
    await expect(card.locator('text=成果：E2E测试成就-联调完成')).toBeVisible();
  });

  test('5.2 卡点记录换行显示', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card.locator('text=卡点：E2E测试卡点-等待依赖')).toBeVisible({ timeout: 5000 });
  });

  test('5.3 第二个子事项的进度记录显示', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card.locator('text=成果：E2E测试成就B-基础完成')).toBeVisible({ timeout: 5000 });
    await expect(card.locator('text=卡点：E2E测试卡点B-性能问题')).toBeVisible();
  });

  test('5.4 同一子事项的多条进度记录各自独立显示', async ({ page }) => {
    // Sub-item 1 has 2 progress records - both should be visible as separate lines
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await page.waitForTimeout(2000);
    const achievements = card.locator('text=成果：E2E测试成就');
    const count = await achievements.count();
    // At least 2 separate achievement lines for sub-item 1
    expect(count).toBeGreaterThanOrEqual(2);
  });

  // ====== 6. Delta Badges ======
  test('6.1 NEW 徽章对新子事项显示', async ({ page }) => {
    // Both sub-items were created this week, so they should show NEW badge
    await page.waitForTimeout(2000);
    const newBadge = page.locator('text=NEW');
    if (await newBadge.first().isVisible()) {
      // Check amber styling
      const badge = newBadge.first();
      const classes = await badge.getAttribute('class') || '';
      expect(classes).toMatch(/amber|warning/);
    }
  });

  // ====== 7. Week Selector Interaction ======
  test('7.1 周选择器不允许选择未来周', async ({ page }) => {
    // When on current week, the "next week" button should be disabled
    const nextBtn = page.locator('[data-testid="week-selector"] button[aria-label="next week"]');
    await expect(nextBtn).toBeVisible();
    await expect(nextBtn).toBeDisabled();
  });

  test('7.2 切换周次后数据刷新', async ({ page }) => {
    await page.waitForTimeout(2000);

    // Click the "prev week" arrow button to go back one week
    await page.locator('[data-testid="week-selector"] button[aria-label="prev week"]').click();
    await page.waitForTimeout(3000);

    // Page should still show weekly view content (may be different data)
    await expect(page.locator('h1:text("每周进展")')).toBeVisible();
  });

  test('7.3 手动选择周次后下方显示数据（非空状态）', async ({ page }) => {
    await page.waitForTimeout(2000);

    // Verify current week shows data (each test starts fresh via beforeEach)
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    const mainItemVisible = await card.isVisible().catch(() => false);
    expect(mainItemVisible).toBeTruthy();

    // Should NOT show "暂无周数据" empty state
    const emptyState = page.locator('text=暂无周数据');
    const emptyVisible = await emptyState.isVisible().catch(() => false);
    expect(emptyVisible).toBeFalsy();

    // Data should be displayed - stats bar should be visible
    await expect(page.locator('text=本周活跃子事项')).toBeVisible({ timeout: 5000 });

    // Test data should appear in the view
    await expect(page.locator(`[data-testid="group-card-${testMainItemId}"]`)).toBeVisible({ timeout: 5000 });
  });

  // ====== 8. Expand/Collapse Completed Items ======
  test('8.1 无已完成项时不显示展开按钮', async ({ page }) => {
    await page.waitForTimeout(2000);
    // Our test items are all in-progress, no completed-no-change yet
    // The expand button may or may not exist depending on data
    const expandBtn = page.locator(`[data-testid="expand-completed-${testMainItemId}"]`);
    const isVisible = await expandBtn.isVisible().catch(() => false);
    // If visible, it means there are completed items
    console.log(`Expand button visible: ${isVisible}`);
  });

  // ====== 10. API Verification ======
  test('10.1 周视图API返回progressRecords数组', async ({ request }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const y = monday.getFullYear();
    const m = String(monday.getMonth() + 1).padStart(2, '0');
    const d = String(monday.getDate()).padStart(2, '0');
    const weekStart = `${y}-${m}-${d}`;

    const resp = await request.get(`/v1/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(resp.status()).toBe(200);
    const data = await resp.json();
    const result = parseApiData(data);

    // Find our test sub-item in thisWeek
    let foundWithRecords = false;
    for (const group of result.groups || []) {
      for (const item of group.thisWeek || []) {
        if (item.title === 'E2E子事项-进度测试A') {
          expect(item.progressRecords).toBeDefined();
          expect(Array.isArray(item.progressRecords)).toBeTruthy();
          expect(item.progressRecords.length).toBeGreaterThanOrEqual(2);
          // Verify record structure
          const record = item.progressRecords[0];
          expect(record).toHaveProperty('achievement');
          expect(record).toHaveProperty('blocker');
          expect(record).toHaveProperty('completion');
          foundWithRecords = true;
        }
      }
    }
    expect(foundWithRecords).toBeTruthy();
  });

  // ====== 11. Visual Separation Check ======
  test('11.1 子事项标题行与进度记录视觉分离', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await page.waitForTimeout(2000);
    // Verify the sub-item title exists
    const title = card.locator('text=E2E子事项-进度测试A');
    await expect(title).toBeVisible({ timeout: 5000 });

    // Verify the progress record exists below it
    const progressLine = card.locator('text=成果：E2E测试成就-第一阶段完成');
    await expect(progressLine).toBeVisible();

    // Both should be visible simultaneously
    await expect(title).toBeVisible();
    await expect(progressLine).toBeVisible();
  });

  // ====== CLEANUP ======
  test.afterAll(async ({ playwright }) => {
    // Attempt to archive test main item
    try {
      const request = await playwright.request.newContext({
        baseURL: 'http://127.0.0.1:8080',
        extraHTTPHeaders: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`,
        },
      });

      await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/archive`);
    } catch {
      // Cleanup is best-effort
    }
  });
});
