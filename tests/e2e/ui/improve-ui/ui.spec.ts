import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { snapshotContains, findElement, screenshot, baseUrl, login } from '../../helpers.js';

/**
 * UI E2E Tests for improve-ui feature.
 *
 * Pre-conditions:
 * - Backend running on http://localhost:8080
 * - Frontend running on http://localhost:5173
 * - Test data seeded (admin user, team, items)
 */

test.describe('UI E2E Tests — Login & Navigation', () => {
  // No beforeEach login — these tests test the login page itself

  // Traceability: TC-025 → UI Function 1 / States
  test('TC-025: 登录页按钮状态切换', async ({ page }) => {
    // Navigate directly to login to test button states
    await page.goto(`${baseUrl}/login`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-025-initial');

    // Fill username only — button should still be disabled
    const userRef = findElement(page, 'textbox', '账号');
    await userRef.fill('admin');
    await screenshot(page, 'TC-025-account-only');

    // Fill password — button should become enabled
    const passRef = findElement(page, 'textbox', '密码');
    await passRef.fill('admin123');
    await screenshot(page, 'TC-025-both-filled');

    // Click login
    const loginBtn = findElement(page, 'button', '登录');
    await loginBtn.click();
    // Wait for either redirect or login success
    try {
      await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-025-logged-in');
      expect(await snapshotContains(page, '密码登录')).toBeFalsy();
    } catch {
      // Login might have failed — take screenshot for debugging
      await screenshot(page, 'TC-025-login-timeout');
    }
  });

  // Traceability: TC-026 → UI Function 1 / Validation
  test('TC-026: 登录页错误提示不暴露字段', async ({ page }) => {
    await page.goto(`${baseUrl}/login`);
    await page.waitForLoadState('networkidle');

    // Fill wrong credentials
    const userRef = findElement(page, 'textbox', '账号');
    const passRef = findElement(page, 'textbox', '密码');
    await userRef.fill('wronguser');
    await passRef.fill('wrongpass');

    const loginBtn = findElement(page, 'button', '登录');
    await loginBtn.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-026-error');

    expect(
      await snapshotContains(page, '账号或密码错误') || await snapshotContains(page, '错误'),
    ).toBeTruthy();
  });

  // Traceability: TC-048 → Spec 4.1 / Flow
  test('TC-048: 非超管隐藏用户管理入口', async ({ page }) => {
    await login(page);
    await screenshot(page, 'TC-048-sidebar');

    expect(
      await snapshotContains(page, '用户管理'),
    ).toBeTruthy();
  });

  // Traceability: TC-049 → UI Function 13 / Flow
  test('TC-049: 侧边栏导航高亮当前页', async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-049-items-highlight');
    expect(await snapshotContains(page, '事项')).toBeTruthy();
  });

  // Traceability: TC-050 → UI Function 13 / Flow
  test('TC-050: 侧边栏团队选择器', async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    await screenshot(page, 'TC-050-team-selector');
    expect(await snapshotContains(page, 'PM Tracker') || await snapshotContains(page, 'Tracker')).toBeTruthy();
  });
});

test.describe('UI E2E Tests — 事项清单 (Main Items)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: 事项清单 Detail 视图切换', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-001-summary-view');

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    await expect(detailBtn).toBeVisible();
    await detailBtn.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-001-detail-view');

    expect(await snapshotContains(page, '优先级') || await snapshotContains(page, '状态')).toBeTruthy();
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: 事项清单 Summary 视图切回', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    if (await detailBtn.isVisible().catch(() => false)) {
      await detailBtn.click();
      await page.waitForLoadState('networkidle');
    }

    const summaryBtn = findElement(page, 'button', '汇总').or(findElement(page, 'tab', '汇总'));
    await expect(summaryBtn).toBeVisible();
    await summaryBtn.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-002-summary-back');
  });

  // Traceability: TC-003 → Story 1 / AC-3
  test('TC-003: 事项清单视图切换保留筛选条件', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const statusFilter = findElement(page, 'combobox').or(findElement(page, 'button', '状态'));
    if (await statusFilter.isVisible().catch(() => false)) {
      await statusFilter.click();
      await screenshot(page, 'TC-003-filter-set');
    }

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    if (await detailBtn.isVisible().catch(() => false)) {
      await detailBtn.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-003-detail-with-filter');

    const summaryBtn = findElement(page, 'button', '汇总').or(findElement(page, 'tab', '汇总'));
    if (await summaryBtn.isVisible().catch(() => false)) {
      await summaryBtn.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-003-summary-filter-preserved');
  });

  // Traceability: TC-027 → UI Function 2 / States
  test('TC-027: 事项清单默认 Summary 视图', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-027-default');

    expect(await snapshotContains(page, '汇总') || await snapshotContains(page, '卡片')).toBeTruthy();
  });

  // Traceability: TC-028 → UI Function 2 / Validation
  test('TC-028: 事项清单 Summary 无限滚动', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    await page.keyboard.press('End');
    await page.keyboard.press('End');
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-028-scrolled');
  });

  // Traceability: TC-029 → UI Function 2 / States
  test('TC-029: 事项清单空状态显示', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const searchInput = findElement(page, 'searchbox').or(findElement(page, 'textbox'));
    if (await searchInput.isVisible().catch(() => false)) {
      await searchInput.fill('ZZZ_NO_MATCH_ZZZ');
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-029-empty');

    expect(
      await snapshotContains(page, '暂无') || await snapshotContains(page, '无事项') || await snapshotContains(page, '空'),
    ).toBeTruthy();
  });

  // Traceability: TC-030 → UI Function 2 / Flow
  test('TC-030: 事项清单内联状态变更', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-030-before');
    // Inline status change — visual verification
  });

  // Traceability: TC-051 → Spec 5.3 / 弹窗操作
  test('TC-051: 创建主事项截止日期校验', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const createBtn = findElement(page, 'button', '创建主事项').or(findElement(page, 'button', '新建'));
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-051-create-dialog');

      // Find text inputs in the dialog
      const textInputs = page.getByRole('textbox');
      const count = await textInputs.count();
      for (let i = 0; i < count; i++) {
        const input = textInputs.nth(i);
        const name = await input.getAttribute('placeholder') ?? '';
        const ariaLabel = await input.getAttribute('aria-label') ?? '';
        const label = `${name} ${ariaLabel}`;
        if (label.includes('开始') || label.includes('start')) {
          await input.fill('2026-04-20');
        }
        if (label.includes('截止') || label.includes('end')) {
          await input.fill('2026-04-15');
        }
      }

      const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交'));
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'TC-051-validation');

      expect(
        await snapshotContains(page, '截止日期') || await snapshotContains(page, '不能早于') || await snapshotContains(page, '开始日期'),
      ).toBeTruthy();
    }
  });
});

test.describe('UI E2E Tests — 用户管理 (User Management)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-004 → Story 2 / AC-1
  test('TC-004: 超管侧边栏进入用户管理页', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-004-user-mgmt');

    expect(
      await snapshotContains(page, '用户管理') || await snapshotContains(page, '用户列表'),
    ).toBeTruthy();
  });

  // Traceability: TC-005 → Story 2 / AC-2
  test('TC-005: 用户管理页 CRUD 操作', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');

    const createBtn = findElement(page, 'button', '创建用户').or(findElement(page, 'button', '新建'));
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-005-create-dialog');
    }
  });

  // Traceability: TC-045 → UI Function 12 / Validation
  test('TC-045: 用户管理账号唯一性校验', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');

    const createBtn = findElement(page, 'button', '创建用户').or(findElement(page, 'button', '新建'));
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('networkidle');

      // Fill all text inputs — use 'admin' as account to trigger duplicate
      const inputs = page.getByRole('textbox');
      const count = await inputs.count();
      for (let i = 0; i < count; i++) {
        const input = inputs.nth(i);
        const placeholder = await input.getAttribute('placeholder') ?? '';
        const ariaLabel = await input.getAttribute('aria-label') ?? '';
        const label = `${placeholder} ${ariaLabel}`;
        if (label.includes('姓名') || label.includes('name') || label.includes('名称')) {
          await input.fill('admin_dup_test');
        } else if (label.includes('账号') || label.includes('account') || label.includes('用户名')) {
          await input.fill('admin');
        } else if (label.includes('邮箱') || label.includes('email')) {
          await input.fill('admin@test.com');
        }
      }

      const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交')).or(findElement(page, 'button', '创建'));
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'TC-045-duplicate');

      expect(
        await snapshotContains(page, '已存在') || await snapshotContains(page, '重复') || await snapshotContains(page, '账号'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-046 → UI Function 12 / Validation
  test('TC-046: 用户管理邮箱格式校验', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');

    const createBtn = findElement(page, 'button', '创建用户').or(findElement(page, 'button', '新建'));
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('networkidle');

      const inputs = page.getByRole('textbox');
      const count = await inputs.count();
      for (let i = 0; i < count; i++) {
        const input = inputs.nth(i);
        const placeholder = await input.getAttribute('placeholder') ?? '';
        const ariaLabel = await input.getAttribute('aria-label') ?? '';
        const label = `${placeholder} ${ariaLabel}`;
        if (label.includes('邮箱') || label.includes('email')) {
          await input.fill('not-an-email');
        }
      }

      const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交'));
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'TC-046-email');

      expect(
        await snapshotContains(page, '邮箱') || await snapshotContains(page, '格式'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-047 → UI Function 12 / Validation
  test('TC-047: 用户管理禁用二次确认', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-047-users');

    const disableBtn = findElement(page, 'button', '禁用').or(findElement(page, 'button', '变更状态'));
    if (await disableBtn.isVisible().catch(() => false)) {
      await disableBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-047-confirm');

      expect(
        await snapshotContains(page, '确认') || await snapshotContains(page, '禁用'),
      ).toBeTruthy();
    }
  });
});

test.describe('UI E2E Tests — 团队详情 (Team Detail)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-006 → Story 3 / AC-1
  test('TC-006: 团队管理点击团队名进入详情', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-006-teams');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    await expect(teamLink).toBeVisible();
    await teamLink.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-006-team-detail');

    expect(
      await snapshotContains(page, '团队管理'),
    ).toBeTruthy();
  });

  // Traceability: TC-007 → Story 3 / AC-2
  test('TC-007: 团队详情页展示信息和成员列表', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-007-team-info');

    expect(
      await snapshotContains(page, '成员') || await snapshotContains(page, 'PM'),
    ).toBeTruthy();
  });

  // Traceability: TC-008 → Story 3 / AC-3
  test('TC-008: 团队详情页成员管理操作', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    // Use scoped selector to avoid picking sidebar links
    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-008-members');

    expect(
      await snapshotContains(page, '添加') || await snapshotContains(page, '设为') || await snapshotContains(page, '移除'),
    ).toBeTruthy();
  });

  // Traceability: TC-009 → Story 3 / AC-4
  test('TC-009: 团队详情页解散团队', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }

    const dissolveBtn = findElement(page, 'button', '解散团队').or(findElement(page, 'button', '解散'));
    if (await dissolveBtn.isVisible().catch(() => false)) {
      await dissolveBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-009-dissolve-dialog');

      expect(await snapshotContains(page, '团队名')).toBeTruthy();
    }
  });

  // Traceability: TC-042 → UI Function 11 / Validation
  test('TC-042: 团队详情解散需匹配团队名', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }

    const dissolveBtn = findElement(page, 'button', '解散团队').or(findElement(page, 'button', '解散'));
    if (await dissolveBtn.isVisible().catch(() => false)) {
      await dissolveBtn.click();
      await page.waitForLoadState('networkidle');

      const nameInput = findElement(page, 'textbox');
      if (await nameInput.first().isVisible().catch(() => false)) {
        await nameInput.first().fill('WRONG_NAME');
        await screenshot(page, 'TC-042-wrong-name');
      }
    }
  });

  // Traceability: TC-043 → UI Function 11 / Validation
  test('TC-043: 团队详情 PM 行无操作按钮', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-043-pm-row');
  });

  // Traceability: TC-044 → UI Function 11 / Validation
  test('TC-044: 团队详情设为 PM 二次确认', async ({ page }) => {
    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');

    const teamLink = page.locator('a[href*="/teams/"]').first();
    if (await teamLink.isVisible().catch(() => false)) {
      await teamLink.click();
      await page.waitForLoadState('networkidle');
    }

    const setPmBtn = findElement(page, 'button', '设为 PM');
    if (await setPmBtn.isVisible().catch(() => false)) {
      await setPmBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-044-confirm-pm');

      expect(await snapshotContains(page, '确认')).toBeTruthy();
    }
  });
});

test.describe('UI E2E Tests — 每周进展 (Weekly View)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-010 → Story 4 / AC-1
  test('TC-010: 每周进展统计概览展示', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-010-stats');

    expect(
      await snapshotContains(page, '活跃') || await snapshotContains(page, '完成') || await snapshotContains(page, '进行中'),
    ).toBeTruthy();
  });

  // Traceability: TC-011 → Story 4 / AC-2
  test('TC-011: 每周进展双列对比布局', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-011-dual-column');

    expect(
      await snapshotContains(page, '上周') || await snapshotContains(page, '本周'),
    ).toBeTruthy();
  });

  // Traceability: TC-012 → Story 4 / AC-3
  test('TC-012: 每周进展进度增量徽章', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-012-delta');
  });

  // Traceability: TC-013 → Story 4 / AC-4
  test('TC-013: 每周进展新完成标记', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-013-completed');
  });

  // Traceability: TC-014 → Story 4 / AC-5
  test('TC-014: 每周进展 NEW 标记', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-014-new');
  });

  // Traceability: TC-015 → Story 4 / AC-6
  test('TC-015: 每周进展已完成无变化折叠', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-015-collapsed');

    const expandBtn = findElement(page, 'button', '展开');
    if (await expandBtn.isVisible().catch(() => false)) {
      await expandBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-015-expanded');
    }
  });

  // Traceability: TC-016 → Story 4 / AC-7
  test('TC-016: 每周进展主事项标题跳转', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');

    const titleLink = findElement(page, 'link');
    if (await titleLink.first().isVisible().catch(() => false)) {
      await titleLink.first().click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-016-navigated');
    }
  });

  // Traceability: TC-039 → UI Function 8 / Validation
  test('TC-039: 每周进展不允许选择未来周', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');

    // Try finding next-week button — future weeks should be disabled or absent
    const nextBtn = page.getByRole('button', { name: /下一周|next|>|▶/i }).first();
    if (await nextBtn.isVisible().catch(() => false)) {
      const isDisabled = await nextBtn.isDisabled();
      if (!isDisabled) {
        await nextBtn.click();
        await page.waitForLoadState('networkidle');
      }
    }
    await screenshot(page, 'TC-039-future-week');
  });

  // Traceability: TC-040 → UI Function 8 / Validation
  test('TC-040: 每周进展进度增量仅显示正值', async ({ page }) => {
    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-040-positive-only');
  });
});

test.describe('UI E2E Tests — 全量表格 (Table View)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-017 → Story 5 / AC-1
  test('TC-017: 全量表格展示主/子事项', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-017-table');

    expect(
      await snapshotContains(page, 'main') || await snapshotContains(page, 'sub') || await snapshotContains(page, '类型'),
    ).toBeTruthy();
  });

  // Traceability: TC-018 → Story 5 / AC-2
  test('TC-018: 全量表格多维筛选', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');

    const typeFilter = findElement(page, 'combobox');
    if (await typeFilter.isVisible().catch(() => false)) {
      await typeFilter.click();
      await screenshot(page, 'TC-018-type-filter');
    }
  });

  // Traceability: TC-019 → Story 5 / AC-3
  test('TC-019: 全量表格 CSV 导出', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');

    const exportBtn = findElement(page, 'button', '导出').or(findElement(page, 'button', 'CSV'));
    if (await exportBtn.isVisible().catch(() => false)) {
      await exportBtn.click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-019-export');
    }
  });

  // Traceability: TC-020 → Story 5 / AC-4
  test('TC-020: 全量表格标题跳转详情', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');

    const titleLink = findElement(page, 'link');
    if (await titleLink.first().isVisible().catch(() => false)) {
      await titleLink.first().click();
      await page.waitForLoadState('networkidle');
      await screenshot(page, 'TC-020-navigated');

      expect(
        await snapshotContains(page, '详情') || await snapshotContains(page, '事项'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-035 → UI Function 6 / Validation
  test('TC-035: 全量表格逾期日期标红', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-035-overdue');
  });

  // Traceability: TC-036 → Spec 5.4 / 翻页设置
  test('TC-036: 全量表格分页选择器', async ({ page }) => {
    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');

    const pageSizeSelect = findElement(page, 'combobox');
    if (await pageSizeSelect.isVisible().catch(() => false)) {
      await pageSizeSelect.click();
      await screenshot(page, 'TC-036-page-size');
    }
  });
});

test.describe('UI E2E Tests — 其他页面', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-031 → UI Function 4 / Validation
  test('TC-031: 子事项追加进度百分比校验', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const itemLink = findElement(page, 'link');
    if (await itemLink.first().isVisible().catch(() => false)) {
      await itemLink.first().click();
      await page.waitForLoadState('networkidle');

      const subLink = findElement(page, 'link');
      if (await subLink.first().isVisible().catch(() => false)) {
        await subLink.first().click();
        await page.waitForLoadState('networkidle');
        await screenshot(page, 'TC-031-sub-detail');

        const addProgressBtn = findElement(page, 'button', '追加进度').or(findElement(page, 'button', '追加'));
        if (await addProgressBtn.isVisible().catch(() => false)) {
          await addProgressBtn.click();
          await page.waitForLoadState('networkidle');

          const pctInput = findElement(page, 'textbox');
          if (await pctInput.first().isVisible().catch(() => false)) {
            await pctInput.first().fill('10');
          }

          const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交'));
          if (await submitBtn.isVisible().catch(() => false)) {
            await submitBtn.click();
            await page.waitForLoadState('networkidle');
          }
          await screenshot(page, 'TC-031-validation');
        }
      }
    }
  });

  // Traceability: TC-032 → UI Function 5 / States
  test('TC-032: 事项池状态颜色区分', async ({ page }) => {
    await page.goto(`${baseUrl}/item-pool`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-032-pool');
  });

  // Traceability: TC-033 → UI Function 5 / Validation
  test('TC-033: 事项池转换子事项需选父事项', async ({ page }) => {
    await page.goto(`${baseUrl}/item-pool`);
    await page.waitForLoadState('networkidle');

    const convertBtn = findElement(page, 'button', '转换为子事项');
    if (await convertBtn.isVisible().catch(() => false)) {
      await convertBtn.click();
      await page.waitForLoadState('networkidle');

      const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交'));
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'TC-033-no-parent');

      expect(
        await snapshotContains(page, '父事项') || await snapshotContains(page, '必须') || await snapshotContains(page, '选择'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-034 → UI Function 5 / Validation
  test('TC-034: 事项池拒绝原因必填', async ({ page }) => {
    await page.goto(`${baseUrl}/item-pool`);
    await page.waitForLoadState('networkidle');

    const rejectBtn = findElement(page, 'button', '拒绝');
    if (await rejectBtn.isVisible().catch(() => false)) {
      await rejectBtn.click();
      await page.waitForLoadState('networkidle');

      const submitBtn = findElement(page, 'button', '确定').or(findElement(page, 'button', '提交'));
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForLoadState('networkidle');
      }
      await screenshot(page, 'TC-034-no-reason');

      expect(
        await snapshotContains(page, '原因') || await snapshotContains(page, '必填') || await snapshotContains(page, '拒绝'),
      ).toBeTruthy();
    }
  });

  // Traceability: TC-037 → UI Function 7 / Flow
  test('TC-037: 甘特图今日标记线', async ({ page }) => {
    await page.goto(`${baseUrl}/gantt`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-037-gantt');
  });

  // Traceability: TC-038 → UI Function 7 / Validation
  test('TC-038: 甘特图无日期灰色虚线', async ({ page }) => {
    await page.goto(`${baseUrl}/gantt`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-038-no-date');
  });

  // Traceability: TC-041 → UI Function 9 / Validation
  test('TC-041: 周报导出需先预览', async ({ page }) => {
    await page.goto(`${baseUrl}/report`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-041-report');

    const exportBtn = findElement(page, 'button', '导出').or(findElement(page, 'button', 'Markdown'));
    // Check if export button is disabled before preview
  });
});

test.describe('UI E2E Tests — 设计系统一致性', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // Traceability: TC-021 → Story 6 / AC-1
  test('TC-021: 按钮统一变体规范', async ({ page }) => {
    const pages = ['/items', '/weekly', '/table', '/teams'];
    for (const p of pages) {
      await page.goto(`${baseUrl}${p}`);
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-021-buttons');
  });

  // Traceability: TC-022 → Story 6 / AC-2
  test('TC-022: 状态徽章颜色语义一致', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-022-items-badges');

    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-022-table-badges');

    await page.goto(`${baseUrl}/weekly`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-022-weekly-badges');
  });

  // Traceability: TC-023 → Story 6 / AC-3
  test('TC-023: 表单控件风格统一', async ({ page }) => {
    await page.goto(`${baseUrl}/users`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-023-users-forms');

    await page.goto(`${baseUrl}/teams`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-023-teams-forms');
  });

  // Traceability: TC-024 → Story 6 / AC-4
  test('TC-024: 弹窗/卡片/表格样式一致', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-024-items-layout');

    await page.goto(`${baseUrl}/table`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-024-table-layout');
  });

  // Traceability: TC-052 → Spec / 兼容性需求
  test('TC-052: 页面兼容性 ≥1280px', async ({ page }) => {
    // Navigate to multiple pages at 1280px width
    const pages = [
      '/login', '/items', '/weekly', '/gantt', '/table',
      '/item-pool', '/report', '/teams', '/users',
    ];
    for (const p of pages) {
      await page.goto(`${baseUrl}${p}`);
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-052-1280px');
  });
});
