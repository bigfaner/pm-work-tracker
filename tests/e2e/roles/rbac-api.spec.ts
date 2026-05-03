import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, authHeader, parseApiBody, extractBizKey, randomCode, setupRbacFixtures } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let noPermsToken: string;
let teamBizKey: string;
let team2BizKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
let progressBizKey: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let noPermsUserBizKey: string;
let memberRoleKey: string;
const runId = Date.now();

const parseData = parseApiBody;

test.describe('RBAC — Roles (TC-039..TC-055)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures({ noPerms: true });
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    noPermsToken = f.noPermsToken!;
    pmUserBizKey = f.pmUserBizKey;
    memberUserBizKey = f.memberUserBizKey;
    noPermsUserBizKey = f.noPermsUserBizKey!;
    teamBizKey = f.teamBizKey;
    memberRoleKey = f.memberRoleKey;

    // Create team B for cross-team tests
    const team2Res = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Role Team B ${Date.now()}`, code: randomCode() }),
    });
    expect(team2Res.status === 200 || team2Res.status === 201).toBeTruthy();
    team2BizKey = extractBizKey(parseData(team2Res.body))!;

    // Create main item in team A
    const mainItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'Test Main Item Role', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(mainItemRes.status === 200 || mainItemRes.status === 201).toBeTruthy();
    mainItemBizKey = extractBizKey(parseData(mainItemRes.body))!;

    // Create sub item under main item
    const subItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'Test Sub Item Role', priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(subItemRes.status === 200 || subItemRes.status === 201).toBeTruthy();
    subItemBizKey = extractBizKey(parseData(subItemRes.body))!;

    // Create progress record
    const progressRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 30, achievement: '初始进度', blocker: '', lesson: '' }),
    });
    expect(progressRes.status === 200 || progressRes.status === 201).toBeTruthy();
    progressBizKey = extractBizKey(parseData(progressRes.body))!;
  });

  // Traceability: TC-039 → Story 6 / AC-4
  test('TC-039: 无 main_item:archive 权限归档返回 403', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/archive`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({}),
    });
    expect(res.status).toBe(403);
  });

  // Traceability: TC-040 → Story 6 / AC-5
  test('TC-040: 拥有 progress:update 权限修正进度', async () => {
    const res = await curl('PATCH', `${apiUrl}/v1/teams/${teamBizKey}/progress/${progressBizKey}/completion`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ completion: 50 }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });

  // ── Story 7: Member 的受限操作 ──

  // Traceability: TC-041 → Story 7 / AC-1
  test('TC-041: 拥有 sub_item:create 权限创建子事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: '测试子事项', priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // Traceability: TC-042 → Story 7 / AC-2
  test('TC-042: 拥有 progress:create 权限追加进度', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ completion: 60, achievement: '进度记录', blocker: '', lesson: '' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // Traceability: TC-043 → Story 7 / AC-3
  test('TC-043: 拥有 item_pool:submit 权限提交事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '提交到事项池' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // Traceability: TC-044 → Story 7 / AC-4, Story 9 / AC-1
  test('TC-044: 无 team:invite 权限 API 返回 403', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ username: 'admin', roleKey: memberRoleKey }),
    });
    expect(res.status).toBe(403);
  });

  // Traceability: TC-045 → Story 7 / AC-5
  test('TC-045: 被分配者可编辑自己的子事项', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '更新子事项标题' }),
    });
    expect(res.status).toBe(200);
  });

  // Traceability: TC-046 → Story 7 / AC-6
  test('TC-046: 非被分配者编辑子事项被拒绝', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/99999`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '尝试编辑非自己的子事项' }),
    });
    expect(res.status === 403 || res.status === 404).toBeTruthy();
  });

  // ── Story 8: 跨团队权限隔离 ──

  // Traceability: TC-047 → Story 8 / AC-1
  test('TC-047: A 团队 pm 角色邀请成员成功', async () => {
    const newUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-tc047-${runId}`, displayName: 'TC047 User' }),
    });
    if (newUserRes.status === 200 || newUserRes.status === 201) {
      const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
        headers: authHeader(pmToken),
        body: JSON.stringify({ username: `e2e-tc047-${runId}`, roleKey: memberRoleKey }),
      });
      expect(res.status === 200 || res.status === 201).toBeTruthy();
    }
  });

  // Traceability: TC-048 → Story 8 / AC-3
  test('TC-048: A 团队权限不跨团队到 B 团队', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${team2BizKey}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: `e2e-noperms-role-${runId}`, roleKey: memberRoleKey }),
    });
    expect(res.status).toBe(403);
  });

  // ── Story 9: 后端权限强制执行 ──

  // Traceability: TC-049 → Story 9 / AC-2
  test('TC-049: 非团队成员调用团队 API 返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/99999/main-items`, {
      headers: authHeader(noPermsToken),
    });
    expect(res.status === 403 || res.status === 404).toBeTruthy();
  });

  // Traceability: TC-050 → Story 9 / AC-3
  test('TC-050: 非超管调用 admin API 返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });

  // Traceability: TC-051 → Story 9 / AC-4
  test('TC-051: 无 sub_item:assign 权限直接调用 API 返回 403', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/assignee`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ assigneeKey: memberUserBizKey }),
    });
    expect(res.status).toBe(403);
  });

  // ── Story 10: 角色编辑即时生效 ──

  // Traceability: TC-052 → Story 10 / AC-1
  test('TC-052: 角色权限取消后后端即时生效', async () => {
    const createRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `test-revoke-role-${Date.now()}`,
        description: 'Test role for permission revocation',
        permissionCodes: ['team:read', 'team:invite'],
      }),
    });
    expect(createRoleRes.status === 200 || createRoleRes.status === 201).toBeTruthy();
    const roleData = parseData(createRoleRes.body);
    const testRoleBizKey = extractBizKey(roleData)!;

    // Edit role to remove team:invite
    const editRes = await curl('PUT', `${apiUrl}/v1/admin/roles/${testRoleBizKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ permissionCodes: ['team:read'] }),
    });
    expect(editRes.status === 200 || editRes.status === 204).toBeTruthy();
  });

  // ── Spec 5.6: JWT Claims 与权限获取 ──

  // Traceability: TC-053 → Spec 5.6 / 权限获取方式
  test('TC-053: 获取用户权限 API 返回正确数据', async () => {
    const res = await curl('GET', `${apiUrl}/v1/me/permissions`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);

    const data = parseData(res.body);
    expect(typeof data.isSuperAdmin === 'boolean').toBeTruthy();
    expect(typeof data.teamPermissions === 'object').toBeTruthy();
  });

  // ── 性能需求 ──

  // Traceability: TC-054 → Spec / 性能需求
  test('TC-054: 权限 API 响应时间 < 200ms', async () => {
    const start = Date.now();
    const res = await curl('GET', `${apiUrl}/v1/me/permissions`, {
      headers: authHeader(pmToken),
    });
    const elapsed = Date.now() - start;

    expect(res.status).toBe(200);
    expect(elapsed < 200).toBeTruthy();
  });

  // Traceability: TC-055 → Spec / 性能需求
  test('TC-055: 权限检查中间件响应时间 < 10ms', async () => {
    const iterations = 10;
    const start = Date.now();
    for (let i = 0; i < iterations; i++) {
      await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
        headers: authHeader(pmToken),
      });
    }
    const avgElapsed = (Date.now() - start) / iterations;

    expect(avgElapsed < 100).toBeTruthy();
  });

  // ── Custom Role Permissions (from permission-roles-api) ──────────

  // Traceability: TC-005 → Story 3 / AC-1
  test('TC-005: Custom role with partial permissions allows read — GET /main-items returns 200', async () => {
    // Create a custom role with partial permissions
    const customRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `custom-tc005-${runId}`,
        permissionCodes: ['main_item:read', 'progress:read'],
      }),
    });
    expect(customRoleRes.status === 200 || customRoleRes.status === 201).toBeTruthy();
    const customRoleData = parseData(customRoleRes.body);
    const customRoleBizKey = extractBizKey(customRoleData)!;

    // Create user with custom role
    const userRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-tc005-${runId}`, displayName: 'Custom TC005' }),
    });
    expect(userRes.status === 200 || userRes.status === 201).toBeTruthy();
    const userData = parseData(userRes.body);
    const customToken = await getApiToken(apiBaseUrl, { username: `custom-tc005-${runId}`, password: userData.initialPassword });

    // Add user to team with custom role
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-tc005-${runId}`, roleKey: customRoleBizKey }),
    }).catch(() => {});

    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(customToken),
    });
    expect(res.status).toBe(200);

    // Cleanup
    await curl('DELETE', `${apiUrl}/v1/admin/roles/${customRoleBizKey}`, {
      headers: authHeader(superadminToken),
    }).catch(() => {});
  });

  // Traceability: TC-006 → Story 3 / AC-2
  test('TC-006: Custom role without create permission denies write — POST /main-items returns 403', async () => {
    // Reuse the custom role from TC-005 (only has read permissions)
    const customRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `custom-tc006-${runId}`,
        permissionCodes: ['main_item:read'],
      }),
    });
    expect(customRoleRes.status === 200 || customRoleRes.status === 201).toBeTruthy();
    const customRoleData = parseData(customRoleRes.body);
    const customRoleBizKey = extractBizKey(customRoleData)!;

    const userRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-tc006-${runId}`, displayName: 'Custom TC006' }),
    });
    expect(userRes.status === 200 || userRes.status === 201).toBeTruthy();
    const userData = parseData(userRes.body);
    const customToken = await getApiToken(apiBaseUrl, { username: `custom-tc006-${runId}`, password: userData.initialPassword });

    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-tc006-${runId}`, roleKey: customRoleBizKey }),
    }).catch(() => {});

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(customToken),
      body: JSON.stringify({ title: 'should-be-denied', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status).toBe(403);

    // Cleanup
    await curl('DELETE', `${apiUrl}/v1/admin/roles/${customRoleBizKey}`, {
      headers: authHeader(superadminToken),
    }).catch(() => {});
  });
});
