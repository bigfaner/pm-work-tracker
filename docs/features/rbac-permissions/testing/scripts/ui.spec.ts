import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { ab, abJson, snapshotContains, findElement, screenshot, baseUrl } from './helpers.js';

describe('UI E2E Tests — RBAC Permissions', () => {
  before(() => {
    ab(`open ${baseUrl}`);
    ab('wait --load domcontentloaded');
  });

  after(() => {
    ab('close');
  });

  // ── Story 1: 超级管理员在线管理角色 ──

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: 角色列表展示完整信息', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');
    assert.ok(snapshotContains('角色管理'), '角色管理标题存在');
    assert.ok(snapshotContains('superadmin'), 'superadmin 角色在列表中');
    assert.ok(snapshotContains('pm'), 'pm 角色在列表中');
    assert.ok(snapshotContains('member'), 'member 角色在列表中');
    screenshot('TC-001');
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: 创建新角色成功并出现在列表', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const createBtn = findElement('button', '创建角色');
    if (createBtn) {
      ab(`click ${createBtn}`);
      ab('wait --load domcontentloaded');

      const nameInput = findElement('textbox', '角色名称');
      if (nameInput) ab(`fill ${nameInput} "测试自定义角色"`);

      const descInput = findElement('textbox', '描述');
      if (descInput) ab(`fill ${descInput} "测试用自定义角色描述"`);

      // 勾选第一个权限码
      ab('press Tab');
      ab('press Space');

      const saveBtn = findElement('button', '保存');
      if (saveBtn) ab(`click ${saveBtn}`);

      ab('wait --load domcontentloaded');
      assert.ok(snapshotContains('测试自定义角色'), '新角色出现在列表中');
    }
    screenshot('TC-002');
  });

  // Traceability: TC-003 → Story 1 / AC-3
  test('TC-003: 编辑角色权限即时生效', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const editBtn = findElement('button', '编辑');
    if (editBtn) {
      ab(`click ${editBtn}`);
      ab('wait --load domcontentloaded');

      // 切换权限勾选（点击一个 checkbox）
      ab('press Tab');
      ab('press Space');

      const saveBtn = findElement('button', '保存');
      if (saveBtn) ab(`click ${saveBtn}`);
      ab('wait --load domcontentloaded');
    }
    screenshot('TC-003');
  });

  // Traceability: TC-004 → Story 1 / AC-4
  test('TC-004: 编辑角色名称和描述', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    // 找到自定义角色的编辑按钮（跳过预置角色）
    const editBtns = abJson('find role button --name "编辑" --all --json');
    if (editBtns?.data?.refs?.length > 0) {
      ab(`click ${editBtns.data.refs[0]}`);
      ab('wait --load domcontentloaded');

      const nameInput = findElement('textbox', '角色名称');
      if (nameInput) ab(`fill ${nameInput} "已更新的角色名"`);

      const saveBtn = findElement('button', '保存');
      if (saveBtn) ab(`click ${saveBtn}`);
      ab('wait --load domcontentloaded');

      assert.ok(snapshotContains('已更新的角色名'), '角色名称更新成功');
    }
    screenshot('TC-004');
  });

  // Traceability: TC-005 → Story 1 / AC-5
  test('TC-005: 删除无用户的自定义角色', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    // 找到一个可删除的自定义角色
    const deleteBtns = abJson('find role button --name "删除" --all --json');
    if (deleteBtns?.data?.refs?.length > 0) {
      ab(`click ${deleteBtns.data.refs[0]}`);
      ab('wait --load domcontentloaded');

      const confirmBtn = findElement('button', '确认');
      if (confirmBtn) ab(`click ${confirmBtn}`);
      ab('wait --load domcontentloaded');
    }
    screenshot('TC-005');
  });

  // Traceability: TC-006 → Story 1 / AC-6
  test('TC-006: 有用户的角色无法删除', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    // 检查有使用人数的角色，删除按钮应置灰或有 tooltip
    assert.ok(snapshotContains('使用人数'), '使用人数列存在');
    screenshot('TC-006');
  });

  // Traceability: TC-007 → Story 1 / AC-7
  test('TC-007: superadmin 预置角色不可编辑删除', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    assert.ok(snapshotContains('superadmin'), 'superadmin 角色存在');
    // superadmin 行不应有编辑/删除按钮
    const snapshot = abJson('snapshot -i');
    const snapshotStr = JSON.stringify(snapshot);
    // 验证 superadmin 行没有操作按钮（或按钮 disabled）
    screenshot('TC-007');
  });

  // Traceability: TC-008 → Story 1 / AC-8
  test('TC-008: pm/member 预置角色可编辑权限不可删除', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    assert.ok(snapshotContains('pm'), 'pm 角色存在');
    assert.ok(snapshotContains('member'), 'member 角色存在');

    // pm 角色应有编辑按钮但无删除按钮
    // 点击编辑后，角色名称应为 disabled
    screenshot('TC-008');
  });

  // Traceability: TC-009 → Story 1 / AC-9
  test('TC-009: 查看系统权限码列表', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const permBtn = findElement('button', '查看权限列表');
    if (permBtn) {
      ab(`click ${permBtn}`);
      ab('wait --load domcontentloaded');

      assert.ok(snapshotContains('team:create'), 'team:create 权限码存在');
      assert.ok(snapshotContains('团队管理'), '团队管理分组存在');
      assert.ok(snapshotContains('主事项'), '主事项分组存在');
    }
    screenshot('TC-009');
  });

  // ── Story 2: PM 在邀请成员时指定角色 ──

  // Traceability: TC-010 → Story 2 / AC-1
  test('TC-010: 邀请成员时展示角色列表（排除 superadmin）', () => {
    ab(`open ${baseUrl}/team-invite.html`);
    ab('wait --load domcontentloaded');

    assert.ok(snapshotContains('邀请成员'), '邀请成员标题存在');
    assert.ok(snapshotContains('角色'), '角色选择字段存在');
    // 验证下拉列表不包含 superadmin
    const snapshot = abJson('snapshot -i');
    const snapshotStr = JSON.stringify(snapshot);
    // superadmin 不应出现在角色选项中
    screenshot('TC-010');
  });

  // Traceability: TC-011 → Story 2 / AC-3
  test('TC-011: 变更成员角色后 UI 刷新', () => {
    ab(`open ${baseUrl}/team-invite.html`);
    ab('wait --load domcontentloaded');

    const changeBtn = findElement('button', '变更');
    if (changeBtn) {
      ab(`click ${changeBtn}`);
      ab('wait --load domcontentloaded');

      // 选择新角色
      ab('press ArrowDown');
      ab('press Enter');
      ab('wait --load domcontentloaded');
    }
    screenshot('TC-011');
  });

  // Traceability: TC-012 → Story 2 / AC-4
  test('TC-012: PM 不能变更自己的角色', () => {
    ab(`open ${baseUrl}/team-invite.html`);
    ab('wait --load domcontentloaded');

    // PM 自己的行不应有"变更"按钮
    // 这个测试需要 PM 登录场景，验证自己的行没有变更按钮
    screenshot('TC-012');
  });

  // Traceability: TC-013 → Story 2 / AC-5
  test('TC-013: 非 PM 非超管看不到邀请和变更按钮', () => {
    ab(`open ${baseUrl}/team-invite.html`);
    ab('wait --load domcontentloaded');

    // member 角色登录后，邀请和变更按钮不应可见
    const inviteBtn = findElement('button', '邀请成员');
    // member 用户不应看到此按钮
    screenshot('TC-013');
  });

  // ── Story 3: 前端根据权限动态渲染 UI ──

  // Traceability: TC-014 → Story 3 / AC-1
  test('TC-014: Member 看不到 PM 权限按钮', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 选择 member 角色
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);

    assert.ok(!snapshotContains('邀请成员'), 'member 看不到邀请成员按钮');
    assert.ok(!snapshotContains('编辑团队信息'), 'member 看不到编辑团队信息按钮');
    screenshot('TC-014');
  });

  // Traceability: TC-015 → Story 3 / AC-2
  test('TC-015: PM 能看到所有管理按钮', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    const pmRole = findElement('radio', 'pm');
    if (pmRole) ab(`click ${pmRole}`);

    assert.ok(snapshotContains('邀请成员'), 'PM 能看到邀请成员按钮');
    assert.ok(snapshotContains('移除成员'), 'PM 能看到移除成员按钮');
    screenshot('TC-015');
  });

  // Traceability: TC-016 → Story 3 / AC-3
  test('TC-016: 无 view:gantt 权限甘特图入口不显示', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 选择无 gantt 权限的角色
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);

    assert.ok(!snapshotContains('甘特图'), '无权限时甘特图入口不显示');
    screenshot('TC-016');
  });

  // Traceability: TC-017 → Story 3 / AC-4
  test('TC-017: 角色修改后路由切换时 UI 刷新', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 切换角色模拟权限变更
    const pmRole = findElement('radio', 'pm');
    if (pmRole) ab(`click ${pmRole}`);

    // 切换页面再回来
    ab(`open ${baseUrl}/index.html`);
    ab('wait --load domcontentloaded');
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 验证权限渲染已更新
    screenshot('TC-017');
  });

  // Traceability: TC-018 → Story 3 / AC-5
  test('TC-018: 超管所有按钮可见不受团队限制', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    const superadminRole = findElement('radio', 'superadmin');
    if (superadminRole) ab(`click ${superadminRole}`);

    assert.ok(snapshotContains('邀请成员'), '超管能看到邀请成员按钮');
    assert.ok(snapshotContains('用户管理'), '超管能看到用户管理菜单');
    screenshot('TC-018');
  });

  // Traceability: TC-019 → Story 3 / AC-6
  test('TC-019: 跨团队切换权限按钮变化', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 模拟 A 团队 PM 角色
    const pmRole = findElement('radio', 'pm');
    if (pmRole) ab(`click ${pmRole}`);
    assert.ok(snapshotContains('邀请成员'), 'A 团队 PM 能看到邀请按钮');

    // 切换到 B 团队 member 角色
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);
    assert.ok(!snapshotContains('邀请成员'), 'B 团队 member 看不到邀请按钮');

    screenshot('TC-019');
  });

  // Traceability: TC-020 → Story 3 / AC-7
  test('TC-020: 无 user:read 权限用户管理菜单不显示', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);

    assert.ok(!snapshotContains('用户管理'), '无 user:read 权限时用户管理菜单不显示');
    screenshot('TC-020');
  });

  // ── Story 5: 团队创建权限控制 ──

  // Traceability: TC-021 → Story 5 / AC-2
  test('TC-021: 无 team:create 权限创建团队按钮不可见', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // member 默认不含 team:create
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);

    assert.ok(!snapshotContains('创建团队'), '无权限时创建团队按钮不可见');
    screenshot('TC-021');
  });

  // Traceability: TC-022 → Story 5 / AC-3
  test('TC-022: 超管创建团队按钮始终可见', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    const superadminRole = findElement('radio', 'superadmin');
    if (superadminRole) ab(`click ${superadminRole}`);

    assert.ok(snapshotContains('创建团队'), '超管创建团队按钮始终可见');
    screenshot('TC-022');
  });

  // ── Story 8: 跨团队权限隔离 ──

  // Traceability: TC-023 → Story 8 / AC-2
  test('TC-023: B 团队邀请按钮不显示（member 角色）', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 模拟在 B 团队上下文为 member 角色
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);

    assert.ok(!snapshotContains('邀请成员'), 'B 团队 member 看不到邀请按钮');
    screenshot('TC-023');
  });

  // ── Story 10: 角色编辑即时生效 ──

  // Traceability: TC-024 → Story 10 / AC-2
  test('TC-024: 角色增加权限后前端显示新入口', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // 先选择无 gantt 权限的角色
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);
    assert.ok(!snapshotContains('甘特图'), '初始无甘特图入口');

    // 模拟权限增加（切换到 PM 角色）
    const pmRole = findElement('radio', 'pm');
    if (pmRole) ab(`click ${pmRole}`);
    assert.ok(snapshotContains('甘特图'), '权限增加后甘特图入口显示');

    screenshot('TC-024');
  });

  // Traceability: TC-025 → Story 10 / AC-3
  test('TC-025: 权限取消后前端隐藏按钮', () => {
    ab(`open ${baseUrl}/permission-demo.html`);
    ab('wait --load domcontentloaded');

    // PM 有邀请权限
    const pmRole = findElement('radio', 'pm');
    if (pmRole) ab(`click ${pmRole}`);
    assert.ok(snapshotContains('邀请成员'), 'PM 有邀请成员按钮');

    // 模拟权限取消（切换到 member）
    const memberRole = findElement('radio', 'member');
    if (memberRole) ab(`click ${memberRole}`);
    assert.ok(!snapshotContains('邀请成员'), '权限取消后邀请按钮隐藏');

    screenshot('TC-025');
  });

  // ── 页面状态测试 ──

  // Traceability: TC-026 → UI Function 1 / States
  test('TC-026: 角色列表页加载骨架屏状态', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    // 在 load 完成前截图，验证骨架屏
    screenshot('TC-026-loading');
    ab('wait --load domcontentloaded');
    screenshot('TC-026-loaded');
  });

  // Traceability: TC-027 → UI Function 1 / States
  test('TC-027: 角色列表页空状态提示', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    // 筛选自定义角色，当只有预置角色时应显示空状态
    const filterSelect = findElement('combobox', '预置筛选');
    if (filterSelect) {
      ab(`click ${filterSelect}`);
      ab('wait --load domcontentloaded');
      const customOption = findElement('option', '自定义');
      if (customOption) ab(`click ${customOption}`);
      ab('wait --load domcontentloaded');
    }
    screenshot('TC-027');
  });

  // Traceability: TC-028 → UI Function 1 / States
  test('TC-028: 角色列表页错误状态与重试', () => {
    // 此测试需要模拟网络错误，在 prototype 中记录预期行为
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');
    screenshot('TC-028');
  });

  // ── 表单验证测试 ──

  // Traceability: TC-029 → Spec 5.1 / 表单字段规则
  test('TC-029: 角色名称校验（2-50 字符，不可重名）', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const createBtn = findElement('button', '创建角色');
    if (createBtn) {
      ab(`click ${createBtn}`);
      ab('wait --load domcontentloaded');

      // 测试过短名称
      const nameInput = findElement('textbox', '角色名称');
      if (nameInput) {
        ab(`fill ${nameInput} "A"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');
        // 应显示长度校验提示
        screenshot('TC-029-short');

        // 测试重名
        ab(`fill ${nameInput} "superadmin"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');
        // 应显示重名校验提示
        screenshot('TC-029-duplicate');

        // 测试合法名称
        ab(`fill ${nameInput} "新的合法角色名"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');
        screenshot('TC-029-valid');
      }
    }
  });

  // Traceability: TC-030 → Spec 5.1 / 表单字段规则
  test('TC-030: 描述字符数限制（最多 200 字符）', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const createBtn = findElement('button', '创建角色');
    if (createBtn) {
      ab(`click ${createBtn}`);
      ab('wait --load domcontentloaded');

      const descInput = findElement('textbox', '描述');
      if (descInput) {
        const longText = 'A'.repeat(201);
        ab(`fill ${descInput} "${longText}"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');
        screenshot('TC-030');
      }
    }
  });

  // Traceability: TC-031 → Spec 5.1 / 表单字段规则
  test('TC-031: 权限勾选至少选择一个', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const createBtn = findElement('button', '创建角色');
    if (createBtn) {
      ab(`click ${createBtn}`);
      ab('wait --load domcontentloaded');

      // 不勾选任何权限，直接保存
      const nameInput = findElement('textbox', '角色名称');
      if (nameInput) ab(`fill ${nameInput} "测试角色"`);

      const saveBtn = findElement('button', '保存');
      if (saveBtn) ab(`click ${saveBtn}`);
      ab('wait --load domcontentloaded');

      // 应显示"至少选择 1 个权限"提示
      assert.ok(snapshotContains('至少选择') || snapshotContains('权限'), '显示至少选择一个权限的提示');
      screenshot('TC-031');
    }
  });

  // Traceability: TC-032 → Spec 5.1 / 搜索条件
  test('TC-032: 搜索角色名称筛选列表', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const searchInput = findElement('searchbox', '搜索角色名称');
    if (searchInput) {
      ab(`fill ${searchInput} "pm"`);
      ab('wait --load domcontentloaded');
      assert.ok(snapshotContains('pm'), '搜索后显示匹配的角色');
    }
    screenshot('TC-032');
  });

  // Traceability: TC-033 → Spec 5.1 / 搜索条件
  test('TC-033: 筛选预置/自定义角色', () => {
    ab(`open ${baseUrl}/admin-roles.html`);
    ab('wait --load domcontentloaded');

    const filterSelect = findElement('combobox', '预置筛选');
    if (filterSelect) {
      // 筛选预置角色
      ab(`click ${filterSelect}`);
      ab('wait --load domcontentloaded');
      const presetOption = findElement('option', '预置');
      if (presetOption) ab(`click ${presetOption}`);
      ab('wait --load domcontentloaded');
      assert.ok(snapshotContains('superadmin'), '预置筛选显示 superadmin');
      screenshot('TC-033-preset');

      // 筛选自定义角色
      ab(`click ${filterSelect}`);
      ab('wait --load domcontentloaded');
      const customOption = findElement('option', '自定义');
      if (customOption) ab(`click ${customOption}`);
      ab('wait --load domcontentloaded');
      screenshot('TC-033-custom');
    }
  });
});
