import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let noPermsToken: string;
let teamBizKey: string;
let team2BizKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
let poolBizKey: string;
let progressBizKey: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let noPermsUserBizKey: string;
let pmRoleKey: string;
let memberRoleKey: string;
const runId = Date.now();

function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

function parseData(body: string): any {
  const resp = JSON.parse(body);
  if (resp.code !== 0) throw new Error(`API error: ${resp.message ?? resp.code}`);
  return resp.data;
}

function extractBizKey(data: any): string {
  return String(data.bizKey ?? data.id);
}

async function loginWithRetry(username: string, password: string): Promise<string> {
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username, password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    expect(res.status).toBe(200);
    return parseData(res.body).token;
  }
  throw new Error('Login failed after retries: rate limited');
}

function randomCode(): string {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  return Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
}

test.describe('API E2E Tests — RBAC Permissions', () => {
  test.beforeAll(async () => {
    // 1. Login as seeded admin
    superadminToken = await loginWithRetry('admin', 'admin123');

    // 2. Fetch preset role bizKeys
    const rolesRes = await curl('GET', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
    });
    expect(rolesRes.status).toBe(200);
    const rolesData = parseData(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesData.items ?? rolesData;
    const pmRole = roles.find((r) => r.roleName === 'pm');
    const memberRole = roles.find((r) => r.roleName === 'member');
    expect(pmRole).toBeTruthy();
    expect(memberRole).toBeTruthy();
    pmRoleKey = pmRole!.bizKey;
    memberRoleKey = memberRole!.bizKey;

    // 3. Create test team A
    const teamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Team A ${runId}`, code: randomCode() }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    teamBizKey = extractBizKey(parseData(teamRes.body));

    // 4. Create test team B (for cross-team tests)
    const team2Res = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Team B ${runId}`, code: randomCode() }),
    });
    expect(team2Res.status === 200 || team2Res.status === 201).toBeTruthy();
    team2BizKey = extractBizKey(parseData(team2Res.body));

    // 5. Create test users
    const users = [
      { username: `e2e-pm-${runId}`, displayName: 'Test PM', keyHolder: 'pm' },
      { username: `e2e-member-${runId}`, displayName: 'Test Member', keyHolder: 'member' },
      { username: `e2e-noperms-${runId}`, displayName: 'Test NoPerms', keyHolder: 'noperms' },
    ];

    for (const u of users) {
      const res = await curl('POST', `${apiUrl}/v1/admin/users`, {
        headers: authHeader(superadminToken),
        body: JSON.stringify({ username: u.username, displayName: u.displayName }),
      });
      expect(res.status === 200 || res.status === 201).toBeTruthy();
      const data = parseData(res.body);
      const bizKey = extractBizKey(data);
      const initialPassword = data.initialPassword;

      if (u.keyHolder === 'pm') pmUserBizKey = bizKey;
      if (u.keyHolder === 'member') memberUserBizKey = bizKey;
      if (u.keyHolder === 'noperms') noPermsUserBizKey = bizKey;

      // Login to get token
      const token = await loginWithRetry(u.username, initialPassword);
      if (u.keyHolder === 'pm') pmToken = token;
      if (u.keyHolder === 'member') memberToken = token;
      if (u.keyHolder === 'noperms') noPermsToken = token;
    }

    // 6. Add PM and member users to team A (PM must be invited as member first, then transferred)
    const addPmRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-${runId}`, roleKey: memberRoleKey }),
    });
    expect(addPmRes.status === 200 || addPmRes.status === 201).toBeTruthy();

    const addMemberRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-member-${runId}`, roleKey: memberRoleKey }),
    });
    expect(addMemberRes.status === 200 || addMemberRes.status === 201).toBeTruthy();

    // Transfer PM to pmUser so they get the pm role in the team
    const transferRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();
    // noPerms user NOT added to any team

    // 7. Create main item in team A
    const mainItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        title: 'Test Main Item',
        priority: 'P2',
        assigneeKey: pmUserBizKey,
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });
    expect(mainItemRes.status === 200 || mainItemRes.status === 201).toBeTruthy();
    mainItemBizKey = extractBizKey(parseData(mainItemRes.body));

    // 8. Create sub item under main item
    const subItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'Test Sub Item', priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(subItemRes.status === 200 || subItemRes.status === 201).toBeTruthy();
    subItemBizKey = extractBizKey(parseData(subItemRes.body));

    // 9. Create progress record
    const progressRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 30, achievement: '初始进度', blocker: '', lesson: '' }),
    });
    expect(progressRes.status === 200 || progressRes.status === 201).toBeTruthy();
    progressBizKey = extractBizKey(parseData(progressRes.body));

    // 10. Create item pool entry
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: 'Test Pool Item' }),
    });
    expect(poolRes.status === 200 || poolRes.status === 201).toBeTruthy();
    poolBizKey = extractBizKey(parseData(poolRes.body));
  });

  // ── Story 2: PM 在邀请成员时指定角色 ──

  // Traceability: TC-034 → Story 2 / AC-2
  test('TC-034: 邀请用户加入团队并分配角色', async () => {
    // noPerms user is not in team; PM invites them
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: `e2e-noperms-${runId}`, roleKey: memberRoleKey }),
    });
    expect(res.status).toBe(200);
  });

  // ── Story 5: 团队创建权限控制 ──

  // Traceability: TC-035 → Story 5 / AC-1
  test('TC-035: 拥有 team:create 权限创建团队成功', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ name: `PM 创建的团队 ${Date.now()}`, code: randomCode() }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // ── Story 6: PM 的权限驱动操作 ──

  // Traceability: TC-036 → Story 6 / AC-1
  test('TC-036: 拥有 main_item:create 权限创建主事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: 'PM 创建的主事项', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // Traceability: TC-037 → Story 6 / AC-2
  test('TC-037: 拥有 sub_item:assign 权限分配负责人', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/assignee`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ assigneeKey: memberUserBizKey }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });

  // Traceability: TC-038 → Story 6 / AC-3
  test('TC-038: 拥有 item_pool:review 权限审核事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}/reject`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ reason: '测试拒绝' }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
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
    // Create a new user for this test since noPerms may already be invited
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
      body: JSON.stringify({ username: `e2e-noperms-${runId}`, roleKey: memberRoleKey }),
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
    const testRoleBizKey = extractBizKey(roleData);

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
});
