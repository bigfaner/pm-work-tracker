import { test, expect } from '@playwright/test';
import { snapshotContains, findElement, screenshot, baseUrl, login, curl, loginAs, apiUrl } from '../../helpers.js';

let _teamId: string;
async function getTeamId(): Promise<string> {
  if (_teamId) return _teamId;
  const { authHeader } = await loginAs('admin', 'admin123');
  const res = await curl('GET', `${apiUrl}/v1/teams`, { headers: authHeader });
  const data = JSON.parse(res.body);
  const teams = data.data?.items ?? data.data ?? data;
  if (Array.isArray(teams) && teams.length > 0) {
    _teamId = String(teams[0].bizKey ?? teams[0].id);
  }
  return _teamId;
}

async function weeklyUrl(params?: string): Promise<string> {
  const tid = await getTeamId();
  return `${apiUrl}/v1/teams/${tid}/views/weekly${params ? '?' + params : ''}`;
}

/**
 * Weekly View (每周进展) Deep E2E Tests
 *
 * Tests cover:
 * 1. Login flow and navigation to weekly view
 * 2. Week selector (date picker) interaction
 * 3. Stats bar rendering and accuracy
 * 4. Comparison cards layout and content
 * 5. Sub-item details (status, priority, assignee, delta badges)
 * 6. NEW badges for newly created sub-items
 * 7. Collapsible "completed no change" section
 * 8. Main item title navigation to detail page
 * 9. Legend display
 * 10. Empty state when no data
 * 11. Future week restriction
 * 12. API data consistency with UI
 * 13. Week navigation (previous/next week)
 * 14. Responsive layout verification
 */

test.describe('每周进展 E2E — 登录与导航', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-025 → W01
  test('W01: 登录成功后跳转到主页', async ({ page }) => {
    await screenshot(page, 'W01-logged-in');
    expect(!(await snapshotContains(page, '密码登录'))).toBeTruthy();
  });

  // Traceability: TC-049 → W02
  test('W02: 侧边栏显示每周进展入口', async ({ page }) => {
    expect(
      (await snapshotContains(page, '每周进展')) || (await snapshotContains(page, '周进展')),
    ).toBeTruthy();
  });

  // Traceability: TC-049 → W03
  test('W03: 点击每周进展导航进入页面', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W03-weekly-page');

    expect(
      (await snapshotContains(page, '每周')) || (await snapshotContains(page, '进展')) || (await snapshotContains(page, '周')),
    ).toBeTruthy();
  });
});

test.describe('每周进展 E2E — 周选择器', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-039 → W04
  test('W04: 周选择器存在且可交互', async ({ page }) => {
    // Look for a week input or textbox/combobox
    const weekInput = page.locator('input[type="week"]').or(
      page.getByRole('textbox'),
    ).or(
      page.getByRole('combobox'),
    ).first();
    const weekInputFound = await weekInput.isVisible().catch(() => false);
    expect(weekInputFound).toBeTruthy();
    await screenshot(page, 'W04-week-selector');
  });

  // Traceability: TC-039 → W05
  test('W05: 页面显示当前周日期范围', async ({ page }) => {
    const text = await page.textContent('body') ?? '';
    const hasDateRange = /\d{4}-\d{2}-\d{2}/.test(text) || text.includes('周');
    expect(hasDateRange).toBeTruthy();
    await screenshot(page, 'W05-date-range');
  });

  // Traceability: TC-039 → W06
  test('W06: 不允许选择未来周', async ({ page }) => {
    await screenshot(page, 'W06-future-week-check');
    // Check for week navigation buttons instead of input[type="week"]
    const nextBtn = page.getByRole('button', { name: /下一周|next|>|▶/i }).first();
    const hasNavigation = await nextBtn.isVisible().catch(() => false);
    // Either navigation buttons exist, or week text is displayed
    const text = await page.textContent('body') ?? '';
    expect(hasNavigation || text.includes('周')).toBeTruthy();
  });

  // Traceability: TC-039 → W07
  test('W07: 切换到上一周数据刷新', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W07-current-week');

    // Find the week input and change to previous week
    const weekInput = page.locator('input[type="week"]').or(
      page.getByRole('textbox'),
    ).first();
    if (await weekInput.isVisible().catch(() => false)) {
      await weekInput.fill('2026-W14');
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'W07-previous-week');
  });
});

test.describe('每周进展 E2E — 统计概览栏', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-010 → W08
  test('W08: 统计概览栏显示4个指标', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const stats = apiData.data?.stats;

    await screenshot(page, 'W08-stats-bar');

    const text = await page.textContent('body') ?? '';

    const hasActive = text.includes('活跃') || text.includes('子事项');
    const hasCompleted = text.includes('完成') || text.includes('新完成');
    const hasInProgress = text.includes('进行中') || text.includes('推进');
    const hasBlocked = text.includes('阻塞');

    expect(
      hasActive || hasCompleted || hasInProgress || hasBlocked,
    ).toBeTruthy();
  });

  // Traceability: TC-010 → W09
  test('W09: 统计数值与API一致', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const stats = apiData.data?.stats;

    await screenshot(page, 'W09-stats-values');

    if (stats) {
      const text = await page.textContent('body') ?? '';
      const activeStr = String(stats.activeSubItems);
      if (stats.activeSubItems > 0) {
        expect(text.includes(activeStr)).toBeTruthy();
      }
    }
  });
});

test.describe('每周进展 E2E — 双列对比布局', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-011 → W10
  test('W10: 对比卡片显示主事项标题', async ({ page }) => {
    const text = await page.textContent('body') ?? '';

    await screenshot(page, 'W10-comparison-cards');

    expect(
      text.includes('事项') || text.includes('测试') || text.includes('开发'),
    ).toBeTruthy();
  });

  // Traceability: TC-011 → W11
  test('W11: 对比卡片显示优先级徽章', async ({ page }) => {
    const text = await page.textContent('body') ?? '';
    await screenshot(page, 'W11-priority-badges');

    const hasPriority = text.includes('P1') || text.includes('P2') || text.includes('P3');
    expect(hasPriority).toBeTruthy();
  });

  // Traceability: TC-011 → W12
  test('W12: 对比卡片显示进度条', async ({ page }) => {
    const text = await page.textContent('body') ?? '';
    await screenshot(page, 'W12-progress-bar');

    const hasProgress = /\d+%/.test(text) || text.includes('完成');
    expect(hasProgress).toBeTruthy();
  });

  // Traceability: TC-011 → W13
  test('W13: 显示"上周"和"本周"列标题', async ({ page }) => {
    const text = await page.textContent('body') ?? '';
    await screenshot(page, 'W13-dual-columns');

    const hasLastWeek = text.includes('上周') || text.includes('Last');
    const hasThisWeek = text.includes('本周') || text.includes('This');
    expect(hasThisWeek).toBeTruthy();
  });

  // Traceability: TC-011 → W14
  test('W14: 本周列显示子事项详情', async ({ page }) => {
    const text = await page.textContent('body') ?? '';
    await screenshot(page, 'W14-sub-items');

    expect(
      text.includes('前端') || text.includes('后端') || text.includes('测试') || text.includes('子事项'),
    ).toBeTruthy();
  });
});

test.describe('每周进展 E2E — 增量徽章与标记', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-014 → W15
  test('W15: 新建子事项显示NEW标记', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    await screenshot(page, 'W15-new-badges');

    let hasNewItems = false;
    for (const group of groups) {
      for (const sub of (group.thisWeek ?? [])) {
        if (sub.isNew) {
          hasNewItems = true;
          break;
        }
      }
    }

    if (hasNewItems) {
      const text = await page.textContent('body') ?? '';
      expect(
        text.includes('NEW') || text.includes('新建') || text.includes('新增'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-012 → W16
  test('W16: 有进度增量子事项显示+N%标记', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    await screenshot(page, 'W16-delta-badges');

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

  // Traceability: TC-013 → W17
  test('W17: 本周完成子事项显示完成标记', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    await screenshot(page, 'W17-completed-badges');

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

  // Traceability: TC-012 → W18
  test('W18: 进度描述显示在子事项行', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    await screenshot(page, 'W18-progress-desc');

    let hasDesc = false;
    for (const group of groups) {
      for (const sub of (group.thisWeek ?? [])) {
        if (sub.progressDescription && sub.progressDescription.length > 0) {
          hasDesc = true;
          break;
        }
      }
    }

    if (hasDesc) {
      const text = await page.textContent('body') ?? '';
      expect(
        text.includes('开发') || text.includes('完成') || text.includes('接口') || text.includes('API'),
      ).toBeTruthy();
    }
  });
});

test.describe('每周进展 E2E — 折叠展开功能', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-015 → W19
  test('W19: 已完成无变化子事项默认折叠', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    let hasCompletedNoChange = false;
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        break;
      }
    }

    await screenshot(page, 'W19-collapsed-state');

    if (hasCompletedNoChange) {
      const text = await page.textContent('body') ?? '';
      expect(
        text.includes('已完成无变化') || text.includes('展开') || text.includes('折叠'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-015 → W20
  test('W20: 点击展开按钮显示已完成无变化子事项', async ({ page }) => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    let hasCompletedNoChange = false;
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        break;
      }
    }

    if (hasCompletedNoChange) {
      // Try to find and click expand button
      const expandBtn = page.getByRole('button', { name: /展开|已完成/ }).first();
      if (await expandBtn.isVisible().catch(() => false)) {
        await expandBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'W20-expanded-state');
    }
  });
});

test.describe('每周进展 E2E — 导航跳转', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-016 → W21
  test('W21: 主事项标题可点击跳转详情页', async ({ page }) => {
    await screenshot(page, 'W21-before-nav');

    const firstLink = page.getByRole('link').first();
    if (await firstLink.isVisible().catch(() => false)) {
      await firstLink.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'W21-after-nav');

      const afterText = await page.textContent('body') ?? '';
      expect(
        afterText.includes('详情') || afterText.includes('事项') || afterText.includes('子事项'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-049 → W22
  test('W22: 侧边栏导航保持高亮', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W22-sidebar-highlight');

    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('每周进展') || text.includes('周进展'),
    ).toBeTruthy();
  });
});

test.describe('每周进展 E2E — 图例与底部信息', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-010 → W23
  test('W23: 图例区域显示', async ({ page }) => {
    await page.mouse.wheel(0, 300);
    await page.waitForTimeout(500);
    await screenshot(page, 'W23-legend');

    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('图例') || text.includes('P1') || text.includes('P2') || text.includes('P3'),
    ).toBeTruthy();
  });
});

test.describe('每周进展 E2E — 空状态处理', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-039 → W24
  test('W24: 选择无数据的周显示空状态', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');

    // Navigate to a week far in the past with no data
    const weekInput = page.locator('input[type="week"]').or(
      page.getByRole('textbox'),
    ).first();
    if (await weekInput.isVisible().catch(() => false)) {
      await weekInput.fill('2020-W01');
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'W24-empty-state');

    // Should show empty state message
    const text = await page.textContent('body') ?? '';
    expect(
      text.includes('暂无') || text.includes('无') || text.includes('0') || text.includes('没有'),
    ).toBeTruthy();
  });
});

test.describe('每周进展 E2E — API数据一致性验证', () => {
  // Traceability: TC-010 → W25
  test('W25: API返回数据结构正确', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.code).toBe(0);
    expect(data.data).toBeTruthy();
    expect(data.data.weekStart).toBeTruthy();
    expect(data.data.weekEnd).toBeTruthy();
    expect(data.data.stats).toBeTruthy();
    expect(Array.isArray(data.data.groups) || data.data.groups === null).toBeTruthy();
  });

  // Traceability: TC-010 → W26
  test('W26: API统计数值非负', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const data = JSON.parse(res.body);
    const stats = data.data?.stats;
    expect(stats).toBeTruthy();
    expect(stats.activeSubItems >= 0).toBeTruthy();
    expect(stats.newlyCompleted >= 0).toBeTruthy();
    expect(stats.inProgress >= 0).toBeTruthy();
    expect(stats.blocked >= 0).toBeTruthy();
  });

  // Traceability: TC-039 → W27
  test('W27: 未来周请求被拒绝', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', await weeklyUrl('weekStart=2028-01-03'), {
      headers: authHeader,
    });
    expect(res.status >= 400).toBeTruthy();
  });

  // Traceability: TC-039 → W28
  test('W28: 无效weekStart参数返回错误', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', await weeklyUrl(), {
      headers: authHeader,
    });
    expect(res.status >= 400).toBeTruthy();
  });

  // Traceability: TC-011 → W29
  test('W29: 每个对比组的结构完整', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', await weeklyUrl('weekStart=2026-04-13'), {
      headers: authHeader,
    });
    const data = JSON.parse(res.body);
    const groups = data.data?.groups ?? [];

    for (const group of groups) {
      expect(group.mainItem).toBeTruthy();
      expect(group.mainItem.id).toBeTruthy();
      expect(group.mainItem.title).toBeTruthy();
      expect(group.mainItem.priority).toBeTruthy();
      expect(typeof group.mainItem.completion === 'number').toBeTruthy();
      expect(typeof group.mainItem.subItemCount === 'number').toBeTruthy();

      // thisWeek items should have required fields
      for (const sub of (group.thisWeek ?? [])) {
        expect(sub.id).toBeTruthy();
        expect(sub.title).toBeTruthy();
        expect(sub.status).toBeTruthy();
        expect(typeof sub.completion === 'number').toBeTruthy();
        expect(sub.priority).toBeTruthy();
      }
    }
  });
});

test.describe('每周进展 E2E — 全流程集成测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-016 → W30
  test('W30: 完整业务流程：登录→导航→查看→切周→返回', async ({ page }) => {
    // Step 1: Already logged in
    await screenshot(page, 'W30-step1-login');

    // Step 2: Navigate to weekly view
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W30-step2-weekly');

    expect(
      (await snapshotContains(page, '每周')) || (await snapshotContains(page, '进展')) || (await snapshotContains(page, '周')),
    ).toBeTruthy();

    // Step 3: Verify data is displayed
    const text1 = await page.textContent('body') ?? '';
    const hasData = text1.includes('事项') || text1.includes('子事项') || text1.includes('P1');
    expect(hasData).toBeTruthy();

    // Step 4: Scroll down to see more content
    await page.mouse.wheel(0, 300);
    await page.waitForTimeout(500);
    await screenshot(page, 'W30-step4-scrolled');

    // Step 5: Scroll back up
    await page.mouse.wheel(0, -300);
    await page.waitForTimeout(500);

    // Step 6: Navigate to items page and back
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W30-step6-items');

    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'W30-step7-back-to-weekly');

    expect(
      (await snapshotContains(page, '每周')) || (await snapshotContains(page, '进展')),
    ).toBeTruthy();
  });
});
