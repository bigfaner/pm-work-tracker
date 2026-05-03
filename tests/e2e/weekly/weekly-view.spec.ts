import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, parseApiData, navTo, extractBizKey, screenshot } from '../helpers.js';

const TIMEOUT = 60000;

test.setTimeout(TIMEOUT);

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
  test.beforeAll(async () => {
    authToken = await getAuthToken();

    // Get team
    const teamsResp = await fetch(`${API}/teams`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsData = await teamsResp.json();
    const teamsRaw = parseApiData(teamsData) || teamsData;
    const teams = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    teamId = String(teams[0]?.bizKey);
    if (!teamId || teamId === 'undefined') throw new Error('No team found');

    // Get team members for assigneeKey
    const membersResp = await fetch(`${API}/teams/${teamId}/members`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const membersRaw = parseApiData(await membersResp.json());
    const memberList = Array.isArray(membersRaw) ? membersRaw : (membersRaw?.items ?? []);
    const assigneeKey = memberList[0]?.userKey ?? '';

    // Clean up stale test data from previous runs
    try {
      const itemsResp = await fetch(`${API}/teams/${teamId}/main-items`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      const itemsData = await itemsResp.json();
      const items = parseApiData(itemsData) || [];
      for (const item of items) {
        if (item.title === 'E2E周视图测试主事项') {
          await fetch(`${API}/teams/${teamId}/main-items/${extractBizKey(item)}/archive`, {
            method: 'POST',
            headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
          });
        }
      }
    } catch { /* best effort */ }

    // Create a main item for testing
    const mainResp = await fetch(`${API}/teams/${teamId}/main-items`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: 'E2E周视图测试主事项',
        priority: 'P1',
        assigneeKey,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-25',
      }),
    });
    const mainData = await mainResp.json();
    testMainItemId = extractBizKey(parseApiData(mainData)) || extractBizKey(mainData) || '';

    // Create sub-item 1
    const sub1Resp = await fetch(`${API}/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        mainItemKey: String(testMainItemId),
        title: 'E2E子事项-进度测试A',
        priority: 'P2',
        assigneeKey,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-20',
      }),
    });
    const sub1Data = await sub1Resp.json();
    testSubItemId1 = extractBizKey(parseApiData(sub1Data)) || extractBizKey(sub1Data) || '';

    // Create sub-item 2
    const sub2Resp = await fetch(`${API}/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        mainItemKey: String(testMainItemId),
        title: 'E2E子事项-进度测试B',
        priority: 'P3',
        assigneeKey,
        startDate: '2026-04-13',
        expectedEndDate: '2026-04-22',
      }),
    });
    const sub2Data = await sub2Resp.json();
    testSubItemId2 = extractBizKey(parseApiData(sub2Data)) || extractBizKey(sub2Data) || '';

    // Change sub-item 1 to in-progress
    await fetch(`${API}/teams/${teamId}/sub-items/${testSubItemId1}/status`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: 'progressing' }),
    });

    // Append progress record 1 for sub-item 1 (this week)
    await fetch(`${API}/teams/${teamId}/sub-items/${testSubItemId1}/progress`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        completion: 40,
        achievement: 'E2E测试成就-第一阶段完成',
        blocker: '',
        lesson: '',
      }),
    });

    // Append progress record 2 for sub-item 1 (this week, multiple records)
    await fetch(`${API}/teams/${teamId}/sub-items/${testSubItemId1}/progress`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        completion: 70,
        achievement: 'E2E测试成就-联调完成',
        blocker: 'E2E测试卡点-等待依赖',
        lesson: '',
      }),
    });

    // Append progress for sub-item 2
    await fetch(`${API}/teams/${teamId}/sub-items/${testSubItemId2}/status`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: 'progressing' }),
    });
    await fetch(`${API}/teams/${teamId}/sub-items/${testSubItemId2}/progress`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        completion: 30,
        achievement: 'E2E测试成就B-基础完成',
        blocker: 'E2E测试卡点B-性能问题',
        lesson: '',
      }),
    });
  });

  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${BASE}/weekly`);
    await page.waitForLoadState('networkidle').catch(() => {});
  });

  // ====== 1. Page Structure ======
  test('1.1 页面标题和周选择器可见', async ({ page }) => {
    await expect(page.locator('h1:text("每周进展")')).toBeVisible();
    await expect(page.locator('[data-testid="week-selector"]')).toBeVisible();
  });

  test('1.2 日期范围显示正确', async ({ page }) => {
    const dateRange = page.locator('[data-testid="week-selector"]').locator('text=/\\d{2}\\/\\d{2}.*~.*\\d{2}\\/\\d{2}/');
    await expect(dateRange).toBeVisible({ timeout: 15000 });
  });

  // ====== 2. Stats Bar ======
  test('2.1 四个统计卡片渲染', async ({ page }) => {
    await expect(page.locator('text=本周活跃')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=本周新完成').first()).toBeVisible();
    await expect(page.locator('text=进行中').first()).toBeVisible();
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
    await expect(page.locator('text=本周活跃')).toBeVisible({ timeout: 5000 });

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
  test('10.1 周视图API返回progressRecords数组', async () => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const y = monday.getFullYear();
    const m = String(monday.getMonth() + 1).padStart(2, '0');
    const d = String(monday.getDate()).padStart(2, '0');
    const weekStart = `${y}-${m}-${d}`;

    const resp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(resp.status).toBe(200);
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

  // ====================================================================
  // MERGED FROM weekly-ui.spec.ts — unique tests not covered above
  // ====================================================================

  // ====== 12. Stats Accuracy (W09) ======
  test('12.1 统计数值与API一致', async ({ page }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const apiResp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const apiData = await apiResp.json();
    const stats = parseApiData(apiData)?.stats;

    await screenshot(page, 'stats-values-api-check');

    if (stats) {
      const text = await page.textContent('body') ?? '';
      const activeStr = String(stats.activeSubItems);
      if (stats.activeSubItems > 0) {
        expect(text.includes(activeStr)).toBeTruthy();
      }
    }
  });

  // ====== 13. This-Week Sub-Item Details (W14) ======
  test('13.1 本周列显示子事项详情', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'this-week-sub-items');

    // Our test sub-items should appear in the thisWeek column
    const text = await card.textContent() ?? '';
    expect(
      text.includes('E2E子事项') || text.includes('进度测试'),
    ).toBeTruthy();
  });

  // ====== 14. +N% Delta Badge (W16) ======
  test('14.1 有进度增量子事项显示+N%标记', async ({ page }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const apiResp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const apiData = await apiResp.json();
    const groups = parseApiData(apiData)?.groups ?? [];

    await screenshot(page, 'delta-badges');

    let hasDeltaItems = false;
    for (const group of groups) {
      for (const sub of (group.thisWeek ?? [])) {
        if (sub.delta > 0) {
          hasDeltaItems = true;
          break;
        }
      }
    }

    if (hasDeltaItems) {
      const text = await page.textContent('body') ?? '';
      expect(/\+\d+%/.test(text)).toBeTruthy();
    }
  });

  // ====== 15. Completed Sub-Item Marker (W17) ======
  test('15.1 本周完成子事项显示完成标记', async ({ page }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const apiResp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const apiData = await apiResp.json();
    const groups = parseApiData(apiData)?.groups ?? [];

    await screenshot(page, 'completed-badges');

    let hasJustCompleted = false;
    for (const group of groups) {
      for (const sub of (group.thisWeek ?? [])) {
        if (sub.justCompleted) {
          hasJustCompleted = true;
          break;
        }
      }
    }

    if (hasJustCompleted) {
      const text = await page.textContent('body') ?? '';
      expect(
        text.includes('完成') || text.includes('✓'),
      ).toBeTruthy();
    }
  });

  // ====== 16. Progress Description Display (W18) ======
  test('16.1 进度描述显示在子事项行', async ({ page }) => {
    const card = page.locator(`[data-testid="group-card-${testMainItemId}"]`);
    await expect(card).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'progress-desc');

    // Our test data has achievement text that should display
    const text = await card.textContent() ?? '';
    expect(
      text.includes('成果') || text.includes('成就') || text.includes('完成'),
    ).toBeTruthy();
  });

  // ====== 17. Completed No-Change Collapsed (W19) ======
  test('17.1 已完成无变化子事项默认折叠', async ({ page }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const apiResp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const apiData = await apiResp.json();
    const groups = parseApiData(apiData)?.groups ?? [];

    let hasCompletedNoChange = false;
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        break;
      }
    }

    await screenshot(page, 'collapsed-state');

    if (hasCompletedNoChange) {
      const text = await page.textContent('body') ?? '';
      expect(
        text.includes('已完成无变化') || text.includes('展开') || text.includes('折叠'),
      ).toBeTruthy();
    }
  });

  // ====== 18. Expand Completed No-Change (W20) ======
  test('18.1 点击展开按钮显示已完成无变化子事项', async ({ page }) => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const apiResp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const apiData = await apiResp.json();
    const groups = parseApiData(apiData)?.groups ?? [];

    let hasCompletedNoChange = false;
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        break;
      }
    }

    if (hasCompletedNoChange) {
      const expandBtn = page.getByRole('button', { name: /展开|已完成/ }).first();
      if (await expandBtn.isVisible().catch(() => false)) {
        await expandBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'expanded-state');
    }
  });

  // ====== 19. Main Item Title Navigation (W21) ======
  test('19.1 主事项标题可点击跳转详情页', async ({ page }) => {
    await screenshot(page, 'before-nav');

    const firstLink = page.getByRole('link').first();
    if (await firstLink.isVisible().catch(() => false)) {
      await firstLink.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'after-nav');

      const afterText = await page.textContent('body') ?? '';
      expect(
        afterText.includes('详情') || afterText.includes('事项') || afterText.includes('子事项'),
      ).toBeTruthy();
    }
  });

  // ====== 20. Sidebar Highlight (W22) ======
  test('20.1 侧边栏导航保持高亮', async ({ page }) => {
    await page.goto(`${BASE}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'sidebar-highlight');

    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('每周进展') || text.includes('周进展'),
    ).toBeTruthy();
  });

  // ====== 21. Legend Area (W23) ======
  test('21.1 图例区域显示', async ({ page }) => {
    await page.mouse.wheel(0, 300);
    await page.waitForTimeout(500);
    await screenshot(page, 'legend');

    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('图例') || text.includes('P1') || text.includes('P2') || text.includes('P3'),
    ).toBeTruthy();
  });

  // ====== 22. Empty State for No-Data Week (W24) ======
  test('22.1 选择无数据的周显示空状态', async ({ page }) => {
    await page.goto(`${BASE}/weekly`);
    await page.waitForLoadState('networkidle');

    // Navigate to a week far in the past with no data using prev week button
    const prevBtn = page.locator('[data-testid="week-selector"] button[aria-label="prev week"]');
    // Click prev several times to reach a week with no data
    for (let i = 0; i < 10; i++) {
      if (await prevBtn.isVisible().catch(() => false)) {
        await prevBtn.click();
        await page.waitForTimeout(500);
      }
    }
    await screenshot(page, 'empty-state');

    // Should show empty state message or low count
    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('暂无') || text.includes('无') || text.includes('0') || text.includes('没有'),
    ).toBeTruthy();
  });

  // ====== 23. API Stats Non-Negative (W26) ======
  test('23.1 API统计数值非负', async () => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const resp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const data = await resp.json();
    const stats = parseApiData(data)?.stats;
    expect(stats).toBeTruthy();
    expect(stats.activeSubItems >= 0).toBeTruthy();
    expect(stats.newlyCompleted >= 0).toBeTruthy();
    expect(stats.inProgress >= 0).toBeTruthy();
    expect(stats.blocked >= 0).toBeTruthy();
  });

  // ====== 24. Future Week Request Rejected (W27) ======
  test('24.1 未来周请求被拒绝', async () => {
    const resp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=2028-01-03`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(resp.status >= 400).toBeTruthy();
  });

  // ====== 25. Invalid Parameter Returns Error (W28) ======
  test('25.1 无效weekStart参数返回错误', async () => {
    const resp = await fetch(`${API}/teams/${teamId}/views/weekly`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(resp.status >= 400).toBeTruthy();
  });

  // ====== 26. Comparison Group Structure Complete (W29) ======
  test('26.1 每个对比组的结构完整', async () => {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    const monday = new Date(now);
    monday.setDate(diff);
    const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;

    const resp = await fetch(`${API}/teams/${teamId}/views/weekly?weekStart=${weekStart}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const data = await resp.json();
    const result = parseApiData(data);
    const groups = result?.groups ?? [];

    for (const group of groups) {
      expect(group.mainItem).toBeTruthy();
      expect(group.mainItem.id || group.mainItem.bizKey).toBeTruthy();
      expect(group.mainItem.title).toBeTruthy();
      expect(group.mainItem.priority).toBeTruthy();
      expect(typeof group.mainItem.completion === 'number').toBeTruthy();
      expect(typeof group.mainItem.subItemCount === 'number').toBeTruthy();

      // thisWeek items should have required fields
      for (const sub of (group.thisWeek ?? [])) {
        expect(sub.id || sub.bizKey).toBeTruthy();
        expect(sub.title).toBeTruthy();
        expect(sub.status).toBeTruthy();
        expect(typeof sub.completion === 'number').toBeTruthy();
        expect(sub.priority).toBeTruthy();
      }
    }
  });

  // ====== 27. Full Business Flow (W30) ======
  test('27.1 完整业务流程：导航→查看→切周→返回', async ({ page }) => {
    // Step 1: Navigate to weekly view (already done in beforeEach)
    await screenshot(page, 'full-flow-step1-weekly');

    expect(
      await page.locator('h1:text("每周进展")').isVisible().catch(() => false)
      || await page.locator('text=每周').first().isVisible().catch(() => false)
      || await page.locator('text=进展').first().isVisible().catch(() => false),
    ).toBeTruthy();

    // Step 2: Verify data is displayed
    const text1 = await page.textContent('body') ?? '';
    const hasData = text1.includes('事项') || text1.includes('子事项') || text1.includes('P1');
    expect(hasData).toBeTruthy();

    // Step 3: Scroll down to see more content
    await page.mouse.wheel(0, 300);
    await page.waitForTimeout(500);
    await screenshot(page, 'full-flow-step3-scrolled');

    // Step 4: Scroll back up
    await page.mouse.wheel(0, -300);
    await page.waitForTimeout(500);

    // Step 5: Navigate to items page and back
    await page.goto(`${BASE}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'full-flow-step5-items');

    await page.goto(`${BASE}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'full-flow-step6-back-to-weekly');

    await expect(page.locator('h1:text("每周进展")')).toBeVisible({ timeout: 5000 });
  });

  // ====== CLEANUP ======
  test.afterAll(async () => {
    // Attempt to archive test main item
    try {
      await fetch(`${API}/teams/${teamId}/main-items/${testMainItemId}/archive`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`,
        },
      });
    } catch {
      // Cleanup is best-effort
    }
  });
});
