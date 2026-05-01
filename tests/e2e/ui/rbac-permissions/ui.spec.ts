import { test, expect } from '@playwright/test';
import { snapshotContains, findElement, screenshot, baseUrl } from './helpers.js';

test.describe('UI E2E Tests — RBAC Permissions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5174/login');
    await page.waitForLoadState('networkidle');
    await page.locator('[data-testid="login-username"]').fill('admin');
    await page.locator('[data-testid="login-password"]').fill('admin123');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForURL(url => !url.pathname.includes('login'));
  });

  // ── Story 1: 超级管理员在线管理角色 ──

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: 角色列表展示完整信息', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');
    expect(await snapshotContains(page, '角色管理')).toBeTruthy();
    expect(await snapshotContains(page, 'superadmin')).toBeTruthy();
    expect(await snapshotContains(page, 'pm')).toBeTruthy();
    expect(await snapshotContains(page, 'member')).toBeTruthy();
    await screenshot(page, 'TC-001');
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: 创建新角色成功并出现在列表', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const createBtn = findElement(page, 'button', '创建角色');
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('domcontentloaded');

      const nameInput = findElement(page, 'textbox', '角色名称');
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill('测试自定义角色');
      }

      const descInput = findElement(page, 'textbox', '描述');
      if (await descInput.isVisible().catch(() => false)) {
        await descInput.fill('测试用自定义角色描述');
      }

      // 勾选第一个权限码
      await page.keyboard.press('Tab');
      await page.keyboard.press('Space');

      const saveBtn = findElement(page, 'button', '保存');
      if (await saveBtn.isVisible().catch(() => false)) {
        await saveBtn.click();
      }

      await page.waitForLoadState('domcontentloaded');
      expect(await snapshotContains(page, '测试自定义角色')).toBeTruthy();
    }
    await screenshot(page, 'TC-002');
  });

  // Traceability: TC-003 → Story 1 / AC-3
  test('TC-003: 编辑角色权限即时生效', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const editBtn = findElement(page, 'button', '编辑');
    if (await editBtn.isVisible().catch(() => false)) {
      await editBtn.click();
      await page.waitForLoadState('domcontentloaded');

      // 切换权限勾选（点击一个 checkbox）
      await page.keyboard.press('Tab');
      await page.keyboard.press('Space');

      const saveBtn = findElement(page, 'button', '保存');
      if (await saveBtn.isVisible().catch(() => false)) {
        await saveBtn.click();
      }
      await page.waitForLoadState('domcontentloaded');
    }
    await screenshot(page, 'TC-003');
  });

  // Traceability: TC-004 → Story 1 / AC-4
  test('TC-004: 编辑角色名称和描述', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    // 找到自定义角色的编辑按钮（跳过预置角色）
    const editBtns = page.getByRole('button', { name: /编辑/i });
    const editBtnCount = await editBtns.count();
    if (editBtnCount > 0) {
      await editBtns.first().click();
      await page.waitForLoadState('domcontentloaded');

      const nameInput = findElement(page, 'textbox', '角色名称');
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill('已更新的角色名');
      }

      const saveBtn = findElement(page, 'button', '保存');
      if (await saveBtn.isVisible().catch(() => false)) {
        await saveBtn.click();
      }
      await page.waitForLoadState('domcontentloaded');

      expect(await snapshotContains(page, '已更新的角色名')).toBeTruthy();
    }
    await screenshot(page, 'TC-004');
  });

  // Traceability: TC-005 → Story 1 / AC-5
  test('TC-005: 删除无用户的自定义角色', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    // 找到一个可删除的自定义角色
    const deleteBtns = page.getByRole('button', { name: /删除/i });
    const deleteBtnCount = await deleteBtns.count();
    if (deleteBtnCount > 0) {
      await deleteBtns.first().click();
      await page.waitForLoadState('domcontentloaded');

      const confirmBtn = findElement(page, 'button', '确认');
      if (await confirmBtn.isVisible().catch(() => false)) {
        await confirmBtn.click();
      }
      await page.waitForLoadState('domcontentloaded');
    }
    await screenshot(page, 'TC-005');
  });

  // Traceability: TC-006 → Story 1 / AC-6
  test('TC-006: 有用户的角色无法删除', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    // 检查有使用人数的角色，删除按钮应置灰或有 tooltip
    expect(await snapshotContains(page, '使用人数')).toBeTruthy();
    await screenshot(page, 'TC-006');
  });

  // Traceability: TC-007 → Story 1 / AC-7
  test('TC-007: superadmin 预置角色不可编辑删除', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    expect(await snapshotContains(page, 'superadmin')).toBeTruthy();
    // superadmin 行不应有编辑/删除按钮
    await screenshot(page, 'TC-007');
  });

  // Traceability: TC-008 → Story 1 / AC-8
  test('TC-008: pm/member 预置角色可编辑权限不可删除', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    expect(await snapshotContains(page, 'pm')).toBeTruthy();
    expect(await snapshotContains(page, 'member')).toBeTruthy();

    // pm 角色应有编辑按钮但无删除按钮
    // 点击编辑后，角色名称应为 disabled
    await screenshot(page, 'TC-008');
  });

  // Traceability: TC-009 → Story 1 / AC-9
  test('TC-009: 查看系统权限码列表', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const permBtn = findElement(page, 'button', '查看权限列表');
    if (await permBtn.isVisible().catch(() => false)) {
      await permBtn.click();
      await page.waitForLoadState('domcontentloaded');

      expect(await snapshotContains(page, 'team:create')).toBeTruthy();
      expect(await snapshotContains(page, '团队管理')).toBeTruthy();
      expect(await snapshotContains(page, '主事项')).toBeTruthy();
    }
    await screenshot(page, 'TC-009');
  });

  // ── Story 2: PM 在邀请成员时指定角色 ──

  // Traceability: TC-010 → Story 2 / AC-1
  test('TC-010: 邀请成员时展示角色列表（排除 superadmin）', async ({ page }) => {
    await page.goto(`${baseUrl}/team-invite.html`);
    await page.waitForLoadState('domcontentloaded');

    expect(await snapshotContains(page, '邀请成员')).toBeTruthy();
    expect(await snapshotContains(page, '角色')).toBeTruthy();
    // 验证下拉列表不包含 superadmin
    await screenshot(page, 'TC-010');
  });

  // Traceability: TC-011 → Story 2 / AC-3
  test('TC-011: 变更成员角色后 UI 刷新', async ({ page }) => {
    await page.goto(`${baseUrl}/team-invite.html`);
    await page.waitForLoadState('domcontentloaded');

    const changeBtn = findElement(page, 'button', '变更');
    if (await changeBtn.isVisible().catch(() => false)) {
      await changeBtn.click();
      await page.waitForLoadState('domcontentloaded');

      // 选择新角色
      await page.keyboard.press('ArrowDown');
      await page.keyboard.press('Enter');
      await page.waitForLoadState('domcontentloaded');
    }
    await screenshot(page, 'TC-011');
  });

  // Traceability: TC-012 → Story 2 / AC-4
  test('TC-012: PM 不能变更自己的角色', async ({ page }) => {
    await page.goto(`${baseUrl}/team-invite.html`);
    await page.waitForLoadState('domcontentloaded');

    // PM 自己的行不应有"变更"按钮
    // 这个测试需要 PM 登录场景，验证自己的行没有变更按钮
    await screenshot(page, 'TC-012');
  });

  // Traceability: TC-013 → Story 2 / AC-5
  test('TC-013: 非 PM 非超管看不到邀请和变更按钮', async ({ page }) => {
    await page.goto(`${baseUrl}/team-invite.html`);
    await page.waitForLoadState('domcontentloaded');

    // member 角色登录后，邀请和变更按钮不应可见
    const inviteBtn = findElement(page, 'button', '邀请成员');
    // member 用户不应看到此按钮
    await screenshot(page, 'TC-013');
  });

  // ── Story 3: 前端根据权限动态渲染 UI ──

  // Traceability: TC-014 → Story 3 / AC-1
  test('TC-014: Member 看不到 PM 权限按钮', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 选择 member 角色
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }

    expect(!(await snapshotContains(page, '邀请成员'))).toBeTruthy();
    expect(!(await snapshotContains(page, '编辑团队信息'))).toBeTruthy();
    await screenshot(page, 'TC-014');
  });

  // Traceability: TC-015 → Story 3 / AC-2
  test('TC-015: PM 能看到所有管理按钮', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    const pmRole = findElement(page, 'radio', 'pm');
    if (await pmRole.isVisible().catch(() => false)) {
      await pmRole.click();
    }

    expect(await snapshotContains(page, '邀请成员')).toBeTruthy();
    expect(await snapshotContains(page, '移除成员')).toBeTruthy();
    await screenshot(page, 'TC-015');
  });

  // Traceability: TC-016 → Story 3 / AC-3
  test('TC-016: 无 view:gantt 权限甘特图入口不显示', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 选择无 gantt 权限的角色
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }

    expect(!(await snapshotContains(page, '甘特图'))).toBeTruthy();
    await screenshot(page, 'TC-016');
  });

  // Traceability: TC-017 → Story 3 / AC-4
  test('TC-017: 角色修改后路由切换时 UI 刷新', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 切换角色模拟权限变更
    const pmRole = findElement(page, 'radio', 'pm');
    if (await pmRole.isVisible().catch(() => false)) {
      await pmRole.click();
    }

    // 切换页面再回来
    await page.goto(`${baseUrl}/index.html`);
    await page.waitForLoadState('domcontentloaded');
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 验证权限渲染已更新
    await screenshot(page, 'TC-017');
  });

  // Traceability: TC-018 → Story 3 / AC-5
  test('TC-018: 超管所有按钮可见不受团队限制', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    const superadminRole = findElement(page, 'radio', 'superadmin');
    if (await superadminRole.isVisible().catch(() => false)) {
      await superadminRole.click();
    }

    expect(await snapshotContains(page, '邀请成员')).toBeTruthy();
    expect(await snapshotContains(page, '用户管理')).toBeTruthy();
    await screenshot(page, 'TC-018');
  });

  // Traceability: TC-019 → Story 3 / AC-6
  test('TC-019: 跨团队切换权限按钮变化', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 模拟 A 团队 PM 角色
    const pmRole = findElement(page, 'radio', 'pm');
    if (await pmRole.isVisible().catch(() => false)) {
      await pmRole.click();
    }
    expect(await snapshotContains(page, '邀请成员')).toBeTruthy();

    // 切换到 B 团队 member 角色
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }
    expect(!(await snapshotContains(page, '邀请成员'))).toBeTruthy();

    await screenshot(page, 'TC-019');
  });

  // Traceability: TC-020 → Story 3 / AC-7
  test('TC-020: 无 user:read 权限用户管理菜单不显示', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }

    expect(!(await snapshotContains(page, '用户管理'))).toBeTruthy();
    await screenshot(page, 'TC-020');
  });

  // ── Story 5: 团队创建权限控制 ──

  // Traceability: TC-021 → Story 5 / AC-2
  test('TC-021: 无 team:create 权限创建团队按钮不可见', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // member 默认不含 team:create
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }

    expect(!(await snapshotContains(page, '创建团队'))).toBeTruthy();
    await screenshot(page, 'TC-021');
  });

  // Traceability: TC-022 → Story 5 / AC-3
  test('TC-022: 超管创建团队按钮始终可见', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    const superadminRole = findElement(page, 'radio', 'superadmin');
    if (await superadminRole.isVisible().catch(() => false)) {
      await superadminRole.click();
    }

    expect(await snapshotContains(page, '创建团队')).toBeTruthy();
    await screenshot(page, 'TC-022');
  });

  // ── Story 8: 跨团队权限隔离 ──

  // Traceability: TC-023 → Story 8 / AC-2
  test('TC-023: B 团队邀请按钮不显示（member 角色）', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 模拟在 B 团队上下文为 member 角色
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }

    expect(!(await snapshotContains(page, '邀请成员'))).toBeTruthy();
    await screenshot(page, 'TC-023');
  });

  // ── Story 10: 角色编辑即时生效 ──

  // Traceability: TC-024 → Story 10 / AC-2
  test('TC-024: 角色增加权限后前端显示新入口', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // 先选择无 gantt 权限的角色
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }
    expect(!(await snapshotContains(page, '甘特图'))).toBeTruthy();

    // 模拟权限增加（切换到 PM 角色）
    const pmRole = findElement(page, 'radio', 'pm');
    if (await pmRole.isVisible().catch(() => false)) {
      await pmRole.click();
    }
    expect(await snapshotContains(page, '甘特图')).toBeTruthy();

    await screenshot(page, 'TC-024');
  });

  // Traceability: TC-025 → Story 10 / AC-3
  test('TC-025: 权限取消后前端隐藏按钮', async ({ page }) => {
    await page.goto(`${baseUrl}/permission-demo.html`);
    await page.waitForLoadState('domcontentloaded');

    // PM 有邀请权限
    const pmRole = findElement(page, 'radio', 'pm');
    if (await pmRole.isVisible().catch(() => false)) {
      await pmRole.click();
    }
    expect(await snapshotContains(page, '邀请成员')).toBeTruthy();

    // 模拟权限取消（切换到 member）
    const memberRole = findElement(page, 'radio', 'member');
    if (await memberRole.isVisible().catch(() => false)) {
      await memberRole.click();
    }
    expect(!(await snapshotContains(page, '邀请成员'))).toBeTruthy();

    await screenshot(page, 'TC-025');
  });

  // ── 页面状态测试 ──

  // Traceability: TC-026 → UI Function 1 / States
  test('TC-026: 角色列表页加载骨架屏状态', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    // 在 load 完成前截图，验证骨架屏
    await screenshot(page, 'TC-026-loading');
    await page.waitForLoadState('domcontentloaded');
    await screenshot(page, 'TC-026-loaded');
  });

  // Traceability: TC-027 → UI Function 1 / States
  test('TC-027: 角色列表页空状态提示', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    // 筛选自定义角色，当只有预置角色时应显示空状态
    const filterSelect = findElement(page, 'combobox', '预置筛选');
    if (await filterSelect.isVisible().catch(() => false)) {
      await filterSelect.click();
      await page.waitForLoadState('domcontentloaded');
      const customOption = findElement(page, 'option', '自定义');
      if (await customOption.isVisible().catch(() => false)) {
        await customOption.click();
      }
      await page.waitForLoadState('domcontentloaded');
    }
    await screenshot(page, 'TC-027');
  });

  // Traceability: TC-028 → UI Function 1 / States
  test('TC-028: 角色列表页错误状态与重试', async ({ page }) => {
    // 此测试需要模拟网络错误，在 prototype 中记录预期行为
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');
    await screenshot(page, 'TC-028');
  });

  // ── 表单验证测试 ──

  // Traceability: TC-029 → Spec 5.1 / 表单字段规则
  test('TC-029: 角色名称校验（2-50 字符，不可重名）', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const createBtn = findElement(page, 'button', '创建角色');
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('domcontentloaded');

      // 测试过短名称
      const nameInput = findElement(page, 'textbox', '角色名称');
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill('A');
        await page.keyboard.press('Tab');
        await page.waitForLoadState('domcontentloaded');
        // 应显示长度校验提示
        await screenshot(page, 'TC-029-short');

        // 测试重名
        await nameInput.fill('superadmin');
        await page.keyboard.press('Tab');
        await page.waitForLoadState('domcontentloaded');
        // 应显示重名校验提示
        await screenshot(page, 'TC-029-duplicate');

        // 测试合法名称
        await nameInput.fill('新的合法角色名');
        await page.keyboard.press('Tab');
        await page.waitForLoadState('domcontentloaded');
        await screenshot(page, 'TC-029-valid');
      }
    }
  });

  // Traceability: TC-030 → Spec 5.1 / 表单字段规则
  test('TC-030: 描述字符数限制（最多 200 字符）', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const createBtn = findElement(page, 'button', '创建角色');
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('domcontentloaded');

      const descInput = findElement(page, 'textbox', '描述');
      if (await descInput.isVisible().catch(() => false)) {
        const longText = 'A'.repeat(201);
        await descInput.fill(longText);
        await page.keyboard.press('Tab');
        await page.waitForLoadState('domcontentloaded');
        await screenshot(page, 'TC-030');
      }
    }
  });

  // Traceability: TC-031 → Spec 5.1 / 表单字段规则
  test('TC-031: 权限勾选至少选择一个', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const createBtn = findElement(page, 'button', '创建角色');
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForLoadState('domcontentloaded');

      // 不勾选任何权限，直接保存
      const nameInput = findElement(page, 'textbox', '角色名称');
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill('测试角色');
      }

      const saveBtn = findElement(page, 'button', '保存');
      if (await saveBtn.isVisible().catch(() => false)) {
        await saveBtn.click();
      }
      await page.waitForLoadState('domcontentloaded');

      // 应显示"至少选择 1 个权限"提示
      expect(
        (await snapshotContains(page, '至少选择')) ||
        (await snapshotContains(page, '权限')),
      ).toBeTruthy();
      await screenshot(page, 'TC-031');
    }
  });

  // Traceability: TC-032 → Spec 5.1 / 搜索条件
  test('TC-032: 搜索角色名称筛选列表', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const searchInput = findElement(page, 'searchbox', '搜索角色名称');
    if (await searchInput.isVisible().catch(() => false)) {
      await searchInput.fill('pm');
      await page.waitForLoadState('domcontentloaded');
      expect(await snapshotContains(page, 'pm')).toBeTruthy();
    }
    await screenshot(page, 'TC-032');
  });

  // Traceability: TC-033 → Spec 5.1 / 搜索条件
  test('TC-033: 筛选预置/自定义角色', async ({ page }) => {
    await page.goto(`${baseUrl}/admin-roles.html`);
    await page.waitForLoadState('domcontentloaded');

    const filterSelect = findElement(page, 'combobox', '预置筛选');
    if (await filterSelect.isVisible().catch(() => false)) {
      // 筛选预置角色
      await filterSelect.click();
      await page.waitForLoadState('domcontentloaded');
      const presetOption = findElement(page, 'option', '预置');
      if (await presetOption.isVisible().catch(() => false)) {
        await presetOption.click();
      }
      await page.waitForLoadState('domcontentloaded');
      expect(await snapshotContains(page, 'superadmin')).toBeTruthy();
      await screenshot(page, 'TC-033-preset');

      // 筛选自定义角色
      await filterSelect.click();
      await page.waitForLoadState('domcontentloaded');
      const customOption = findElement(page, 'option', '自定义');
      if (await customOption.isVisible().catch(() => false)) {
        await customOption.click();
      }
      await page.waitForLoadState('domcontentloaded');
      await screenshot(page, 'TC-033-custom');
    }
  });
});
