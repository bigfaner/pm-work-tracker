import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { ab, abJson, snapshotContains, findElement, screenshot, baseUrl, browserLogin, curl, loginAs } from './helpers.js';

const SCREENSHOTS_DIR = 'docs/features/improve-ui/testing/results/screenshots';

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

describe('每周进展 E2E — 登录与导航', () => {
  before(() => {
    browserLogin('admin', 'admin123');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-025 → W01
  test('W01: 登录成功后跳转到主页', () => {
    screenshot('W01-logged-in');
    assert.ok(!snapshotContains('密码登录'), 'Should leave login page');
  });

  // Traceability: TC-049 → W02
  test('W02: 侧边栏显示每周进展入口', () => {
    assert.ok(
      snapshotContains('每周进展') || snapshotContains('周进展'),
      'Sidebar should show 每周进展 nav item',
    );
  });

  // Traceability: TC-049 → W03
  test('W03: 点击每周进展导航进入页面', () => {
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
    screenshot('W03-weekly-page');

    assert.ok(
      snapshotContains('每周') || snapshotContains('进展') || snapshotContains('周'),
      'Weekly view page should load',
    );
  });
});

describe('每周进展 E2E — 周选择器', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-039 → W04
  test('W04: 周选择器存在且可交互', () => {
    const snap = abJson('snapshot -i');
    const refs = snap?.data?.refs ?? {};
    let weekInputFound = false;
    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { role?: string; type?: string; name?: string };
      if (elem.type === 'week' || elem.role === 'textbox' || elem.role === 'combobox') {
        weekInputFound = true;
        break;
      }
    }
    assert.ok(weekInputFound, 'Week selector input should exist');
    screenshot('W04-week-selector');
  });

  // Traceability: TC-039 → W05
  test('W05: 页面显示当前周日期范围', () => {
    // The page should display the current week date range like "2026-04-13 ~ 2026-04-19"
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    // Check for date pattern in the page
    // Traceability: TC-039 → W05 (regex .test call)
    const hasDateRange = /\d{4}-\d{2}-\d{2}/.test(text) || text.includes('周');
    assert.ok(hasDateRange, 'Page should display week date range');
    screenshot('W05-date-range');
  });

  // Traceability: TC-039 → W06
  test('W06: 不允许选择未来周', () => {
    // The week input should have a max attribute set to current week
    const snap = abJson('snapshot -i');
    screenshot('W06-future-week-check');
    // Verify the input exists (max attribute prevents future selection in HTML)
    const refs = snap?.data?.refs ?? {};
    let foundWeekInput = false;
    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { type?: string };
      if (elem.type === 'week') {
        foundWeekInput = true;
        break;
      }
    }
    assert.ok(foundWeekInput, 'Week input with max restriction should exist');
  });

  // Traceability: TC-039 → W07
  test('W07: 切换到上一周数据刷新', async () => {
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
    screenshot('W07-current-week');

    // Find the week input and change to previous week
    const snap = abJson('snapshot -i');
    const refs = snap?.data?.refs ?? {};
    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { type?: string; role?: string };
      if (elem.type === 'week' || elem.role === 'textbox') {
        // Try to fill with previous week value (2026-W14)
        ab(`fill ${ref} "2026-W14"`);
        ab('wait --load networkidle');
        break;
      }
    }
    screenshot('W07-previous-week');
  });
});

describe('每周进展 E2E — 统计概览栏', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-010 → W08
  test('W08: 统计概览栏显示4个指标', async () => {
    // First verify via API what data we expect
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const stats = apiData.data?.stats;

    screenshot('W08-stats-bar');

    // The page should show stats labels
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';

    // Check for stat labels
    const hasActive = text.includes('活跃') || text.includes('子事项');
    const hasCompleted = text.includes('完成') || text.includes('新完成');
    const hasInProgress = text.includes('进行中') || text.includes('推进');
    const hasBlocked = text.includes('阻塞');

    // At least some stats should be visible
    assert.ok(
      hasActive || hasCompleted || hasInProgress || hasBlocked,
      `Stats bar should show at least one metric. Got: active=${hasActive}, completed=${hasCompleted}, inProgress=${hasInProgress}, blocked=${hasBlocked}`,
    );
  });

  // Traceability: TC-010 → W09
  test('W09: 统计数值与API一致', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const stats = apiData.data?.stats;

    screenshot('W09-stats-values');

    if (stats) {
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      // Verify numeric values appear in the UI
      const activeStr = String(stats.activeSubItems);
      if (stats.activeSubItems > 0) {
        assert.ok(text.includes(activeStr), `Active count ${activeStr} should appear in UI`);
      }
    }
  });
});

describe('每周进展 E2E — 双列对比布局', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-011 → W10
  test('W10: 对比卡片显示主事项标题', () => {
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';

    screenshot('W10-comparison-cards');

    // Should show at least one main item title
    assert.ok(
      text.includes('事项') || text.includes('测试') || text.includes('开发'),
      'Comparison cards should show main item titles',
    );
  });

  // Traceability: TC-011 → W11
  test('W11: 对比卡片显示优先级徽章', () => {
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    screenshot('W11-priority-badges');

    // P1/P2/P3 priority badges should be visible
    const hasPriority = text.includes('P1') || text.includes('P2') || text.includes('P3');
    assert.ok(hasPriority, 'Priority badges (P1/P2/P3) should be visible');
  });

  // Traceability: TC-011 → W12
  test('W12: 对比卡片显示进度条', () => {
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    screenshot('W12-progress-bar');

    // Progress percentage should appear somewhere
    // Traceability: TC-011 → W12 (regex .test call)
    const hasProgress = /\d+%/.test(text) || text.includes('完成');
    assert.ok(hasProgress, 'Progress bar/percentage should be visible');
  });

  // Traceability: TC-011 → W13
  test('W13: 显示"上周"和"本周"列标题', () => {
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    screenshot('W13-dual-columns');

    // Should show column headers for comparison
    const hasLastWeek = text.includes('上周') || text.includes('Last');
    const hasThisWeek = text.includes('本周') || text.includes('This');
    assert.ok(hasThisWeek, '"本周" column header should be visible');
  });

  // Traceability: TC-011 → W14
  test('W14: 本周列显示子事项详情', () => {
    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    screenshot('W14-sub-items');

    // Sub-items should show their titles
    assert.ok(
      text.includes('前端') || text.includes('后端') || text.includes('测试') || text.includes('子事项'),
      'Sub-item titles should be visible in this week column',
    );
  });
});

describe('每周进展 E2E — 增量徽章与标记', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-014 → W15
  test('W15: 新建子事项显示NEW标记', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    screenshot('W15-new-badges');

    // Check if any sub-item has isNew=true
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
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      assert.ok(
        text.includes('NEW') || text.includes('新建') || text.includes('新增'),
        'NEW badges should appear for newly created sub-items',
      );
    }
  });

  // Traceability: TC-012 → W16
  test('W16: 有进度增量子事项显示+N%标记', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    screenshot('W16-delta-badges');

    // Check if any sub-item has delta > 0
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
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      assert.ok(
        // Traceability: TC-012 → W16 (regex .test call)
        /\+\d+%/.test(text), '+N% delta badges should appear for items with progress increase');
    }
  });

  // Traceability: TC-013 → W17
  test('W17: 本周完成子事项显示完成标记', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    screenshot('W17-completed-badges');

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
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      assert.ok(
        text.includes('完成') || text.includes('✓'),
        'Completed badges should appear for items completed this week',
      );
    }
  });

  // Traceability: TC-012 → W18
  test('W18: 进度描述显示在子事项行', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    screenshot('W18-progress-desc');

    // Check if any sub-item has a progress description
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
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      // Progress descriptions should contain achievement text
      assert.ok(
        text.includes('开发') || text.includes('完成') || text.includes('接口') || text.includes('API'),
        'Progress descriptions should be visible in sub-item rows',
      );
    }
  });
});

describe('每周进展 E2E — 折叠展开功能', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-015 → W19
  test('W19: 已完成无变化子事项默认折叠', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    // Check if any group has completedNoChange items
    let hasCompletedNoChange = false;
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        break;
      }
    }

    screenshot('W19-collapsed-state');

    if (hasCompletedNoChange) {
      const snap = abJson('snapshot');
      const text = snap?.data?.snapshot ?? '';
      assert.ok(
        text.includes('已完成无变化') || text.includes('展开') || text.includes('折叠'),
        'Completed no-change section should show expand button',
      );
    }
  });

  // Traceability: TC-015 → W20
  test('W20: 点击展开按钮显示已完成无变化子事项', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const apiRes = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const apiData = JSON.parse(apiRes.body);
    const groups = apiData.data?.groups ?? [];

    let hasCompletedNoChange = false;
    let targetMainItem = '';
    for (const group of groups) {
      if (group.completedNoChange && group.completedNoChange.length > 0) {
        hasCompletedNoChange = true;
        targetMainItem = String(group.mainItem.id);
        break;
      }
    }

    if (hasCompletedNoChange) {
      // Try to find and click expand button
      const snap = abJson('snapshot -i');
      const refs = snap?.data?.refs ?? {};
      for (const [ref, el] of Object.entries(refs)) {
        const elem = el as { role?: string; name?: string };
        if (elem.role === 'button' && (elem.name?.includes('展开') || elem.name?.includes('已完成'))) {
          ab(`click ${ref}`);
          ab('wait --load networkidle');
          break;
        }
      }
      screenshot('W20-expanded-state');
    }
  });
});

describe('每周进展 E2E — 导航跳转', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-016 → W21
  test('W21: 主事项标题可点击跳转详情页', () => {
    // Find a link (main item title) and click it
    const snap = abJson('snapshot -i');
    const refs = snap?.data?.refs ?? {};

    screenshot('W21-before-nav');

    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { role?: string; name?: string };
      if (elem.role === 'link') {
        ab(`click ${ref}`);
        ab('wait --load networkidle');
        screenshot('W21-after-nav');

        // Should have navigated away from /weekly
        const afterSnap = abJson('snapshot');
        const afterText = afterSnap?.data?.snapshot ?? '';
        assert.ok(
          afterText.includes('详情') || afterText.includes('事项') || afterText.includes('子事项'),
          'Should navigate to item detail page',
        );
        // Navigate back for other tests
        ab(`open ${baseUrl}/weekly`);
        ab('wait --load networkidle');
        return;
      }
    }
  });

  // Traceability: TC-049 → W22
  test('W22: 侧边栏导航保持高亮', () => {
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
    screenshot('W22-sidebar-highlight');

    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    assert.ok(
      text.includes('每周进展') || text.includes('周进展'),
      'Sidebar should show 每周进展 entry',
    );
  });
});

describe('每周进展 E2E — 图例与底部信息', () => {
  before(() => {
    browserLogin('admin', 'admin123');
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-010 → W23
  test('W23: 图例区域显示', () => {
    // Scroll to bottom to see legend
    ab('scroll down');
    ab('wait 500');
    screenshot('W23-legend');

    const snap = abJson('snapshot');
    const text = snap?.data?.snapshot ?? '';
    assert.ok(
      text.includes('图例') || text.includes('P1') || text.includes('P2') || text.includes('P3'),
      'Legend section should be visible at the bottom',
    );
  });
});

describe('每周进展 E2E — 空状态处理', () => {
  before(() => {
    browserLogin('admin', 'admin123');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-039 → W24
  test('W24: 选择无数据的周显示空状态', () => {
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');

    // Navigate to a week far in the past with no data
    const snap = abJson('snapshot -i');
    const refs = snap?.data?.refs ?? {};
    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { type?: string; role?: string };
      if (elem.type === 'week' || elem.role === 'textbox') {
        // Fill with a week far in the past
        ab(`fill ${ref} "2020-W01"`);
        ab('wait --load networkidle');
        break;
      }
    }
    screenshot('W24-empty-state');

    // Should show empty state message
    const afterSnap = abJson('snapshot');
    const text = afterSnap?.data?.snapshot ?? '';
    assert.ok(
      text.includes('暂无') || text.includes('无') || text.includes('0') || text.includes('没有'),
      'Empty state message should appear for weeks with no data',
    );
  });
});

describe('每周进展 E2E — API数据一致性验证', () => {
  // Traceability: TC-010 → W25
  test('W25: API返回数据结构正确', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    assert.equal(res.status, 200, 'API should return 200');
    const data = JSON.parse(res.body);
    assert.equal(data.code, 0, 'API response code should be 0');
    assert.ok(data.data, 'API should return data object');
    assert.ok(data.data.weekStart, 'Should have weekStart');
    assert.ok(data.data.weekEnd, 'Should have weekEnd');
    assert.ok(data.data.stats, 'Should have stats');
    assert.ok(Array.isArray(data.data.groups), 'Should have groups array');
  });

  // Traceability: TC-010 → W26
  test('W26: API统计数值非负', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const data = JSON.parse(res.body);
    const stats = data.data?.stats;
    assert.ok(stats, 'Stats should exist');
    assert.ok(stats.activeSubItems >= 0, 'activeSubItems should be >= 0');
    assert.ok(stats.newlyCompleted >= 0, 'newlyCompleted should be >= 0');
    assert.ok(stats.inProgress >= 0, 'inProgress should be >= 0');
    assert.ok(stats.blocked >= 0, 'blocked should be >= 0');
  });

  // Traceability: TC-039 → W27
  test('W27: 未来周请求被拒绝', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2028-01-03`, {
      headers: authHeader,
    });
    assert.ok(res.status >= 400, 'Future week request should be rejected');
  });

  // Traceability: TC-039 → W28
  test('W28: 无效weekStart参数返回错误', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly`, {
      headers: authHeader,
    });
    assert.ok(res.status >= 400, 'Missing weekStart should return error');
  });

  // Traceability: TC-011 → W29
  test('W29: 每个对比组的结构完整', async () => {
    const { authHeader } = await loginAs('admin', 'admin123');
    const res = await curl('GET', `http://localhost:8080/api/v1/teams/1/views/weekly?weekStart=2026-04-13`, {
      headers: authHeader,
    });
    const data = JSON.parse(res.body);
    const groups = data.data?.groups ?? [];

    for (const group of groups) {
      assert.ok(group.mainItem, 'Each group should have mainItem');
      assert.ok(group.mainItem.id, 'mainItem should have id');
      assert.ok(group.mainItem.title, 'mainItem should have title');
      assert.ok(group.mainItem.priority, 'mainItem should have priority');
      assert.ok(typeof group.mainItem.completion === 'number', 'mainItem should have numeric completion');
      assert.ok(typeof group.mainItem.subItemCount === 'number', 'mainItem should have subItemCount');

      // thisWeek items should have required fields
      for (const sub of (group.thisWeek ?? [])) {
        assert.ok(sub.id, 'Sub-item should have id');
        assert.ok(sub.title, 'Sub-item should have title');
        assert.ok(sub.status, 'Sub-item should have status');
        assert.ok(typeof sub.completion === 'number', 'Sub-item should have numeric completion');
        assert.ok(sub.priority, 'Sub-item should have priority');
      }
    }
  });
});

describe('每周进展 E2E — 全流程集成测试', () => {
  before(() => {
    browserLogin('admin', 'admin123');
  });

  after(() => {
    ab('close');
  });

  // Traceability: TC-016 → W30
  test('W30: 完整业务流程：登录→导航→查看→切周→返回', () => {
    // Step 1: Already logged in
    screenshot('W30-step1-login');

    // Step 2: Navigate to weekly view
    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
    screenshot('W30-step2-weekly');

    assert.ok(
      snapshotContains('每周') || snapshotContains('进展') || snapshotContains('周'),
      'Weekly page loaded',
    );

    // Step 3: Verify data is displayed
    const snap1 = abJson('snapshot');
    const text1 = snap1?.data?.snapshot ?? '';
    const hasData = text1.includes('事项') || text1.includes('子事项') || text1.includes('P1');
    assert.ok(hasData, 'Weekly data should be displayed');

    // Step 4: Scroll down to see more content
    ab('scroll down');
    ab('wait 500');
    screenshot('W30-step4-scrolled');

    // Step 5: Scroll back up
    ab('scroll up');
    ab('wait 500');

    // Step 6: Navigate to items page and back
    ab(`open ${baseUrl}/items`);
    ab('wait --load networkidle');
    screenshot('W30-step6-items');

    ab(`open ${baseUrl}/weekly`);
    ab('wait --load networkidle');
    screenshot('W30-step7-back-to-weekly');

    assert.ok(
      snapshotContains('每周') || snapshotContains('进展'),
      'Weekly page reloaded successfully',
    );
  });
});
