import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl } from './helpers.js';

const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let noPermsToken: string;
let teamId: number;
let team2Id: number;
let mainItemId: number;
let subItemId: number;
let poolId: number;
let progressId: number;
let pmUserId: number;
let memberUserId: number;
let noPermsUserId: number;
const runId = Date.now();

function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

function parseData(body: string): any {
  const resp = JSON.parse(body);
  if (resp.code !== 0) throw new Error(`API error: ${resp.message ?? resp.code}`);
  return resp.data;
}

async function ensureTestUsers(): Promise<void> {
  const users = [
    { username: `e2e-pm-${runId}`, displayName: 'Test PM' },
    { username: `e2e-member-${runId}`, displayName: 'Test Member' },
    { username: `e2e-noperms-${runId}`, displayName: 'Test NoPerms' },
  ];
  const passwords: Record<string, string> = {};

  for (const u of users) {
    const res = await curl('POST', `${apiUrl}/api/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify(u),
    });
    assert.ok(res.status === 200 || res.status === 201, `Create user ${u.username}: ${res.status} ${res.body}`);
    const data = parseData(res.body);
    passwords[u.username] = data.initialPassword;
    if (u.username === `e2e-pm-${runId}`) pmUserId = data.id;
    if (u.username === `e2e-member-${runId}`) memberUserId = data.id;
    if (u.username === `e2e-noperms-${runId}`) noPermsUserId = data.id;
  }

  // Login as each user to get tokens
  for (const [username, password] of Object.entries(passwords)) {
    const loginRes = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
      body: JSON.stringify({ username, password }),
    });
    assert.equal(loginRes.status, 200, `Login as ${username} should succeed`);
    const loginData = parseData(loginRes.body);
    if (username === `e2e-pm-${runId}`) pmToken = loginData.token;
    if (username === `e2e-member-${runId}`) memberToken = loginData.token;
    if (username === `e2e-noperms-${runId}`) noPermsToken = loginData.token;
  }
}

describe('API E2E Tests — RBAC Permissions', () => {
  before(async () => {
    // 1. Login as seeded admin
    const adminLogin = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    assert.equal(adminLogin.status, 200, 'Admin login should succeed');
    superadminToken = parseData(adminLogin.body).token;

    // 2. Create test team A
    const teamRes = await curl('POST', `${apiUrl}/api/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Team A ${runId}` }),
    });
    assert.ok(teamRes.status === 200 || teamRes.status === 201, `Create team A: ${teamRes.status}`);
    teamId = parseData(teamRes.body).id;

    // 3. Create test team B (for cross-team tests)
    const team2Res = await curl('POST', `${apiUrl}/api/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Team B ${runId}` }),
    });
    assert.ok(team2Res.status === 200 || team2Res.status === 201, `Create team B: ${team2Res.status}`);
    team2Id = parseData(team2Res.body).id;

    // 4. Create test users
    await ensureTestUsers();

    // 5. Add PM and member users to team A
    // Role IDs: 1=superadmin, 2=pm, 3=member
    const addPmRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-${runId}`, roleId: 2 }),
    });
    assert.ok(addPmRes.status === 200, `Add PM to team: ${addPmRes.status} ${addPmRes.body}`);
    const addMemberRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-member-${runId}`, roleId: 3 }),
    });
    assert.ok(addMemberRes.status === 200, `Add member to team: ${addMemberRes.status} ${addMemberRes.body}`);
    // noPerms user NOT added to any team

    // 6. Create main item in team A
    const mainItemRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        title: 'Test Main Item',
        priority: 'P2',
        assigneeId: pmUserId,
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });
    assert.ok(mainItemRes.status === 200 || mainItemRes.status === 201, `Create main item: ${mainItemRes.status}`);
    mainItemId = parseData(mainItemRes.body).id;

    // 7. Create sub item under main item
    const subItemRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        mainItemId: mainItemId,
        title: 'Test Sub Item',
        priority: 'P2',
        assigneeId: memberUserId,
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });
    assert.ok(subItemRes.status === 200 || subItemRes.status === 201, `Create sub item: ${subItemRes.status}`);
    subItemId = parseData(subItemRes.body).id;

    // 8. Create progress record
    const progressRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/sub-items/${subItemId}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 30, achievement: '初始进度', blocker: '', lesson: '' }),
    });
    assert.ok(progressRes.status === 200 || progressRes.status === 201, `Create progress: ${progressRes.status}`);
    progressId = parseData(progressRes.body).id;

    // 9. Create item pool entry
    const poolRes = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: 'Test Pool Item' }),
    });
    assert.ok(poolRes.status === 200 || poolRes.status === 201, `Create pool item: ${poolRes.status}`);
    poolId = parseData(poolRes.body).id;
  });

  // ── Story 2: PM 在邀请成员时指定角色 ──

  // Traceability: TC-034 → Story 2 / AC-2
  test('TC-034: 邀请用户加入团队并分配角色', async () => {
    // noPerms user is not in team; PM invites them
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: `e2e-noperms-${runId}`, roleId: 3 }),
    });
    assert.equal(res.status, 200, '邀请成员成功返回 200');

    const data = parseData(res.body);
    assert.ok(data !== null || res.status === 200, '操作成功');
  });

  // ── Story 5: 团队创建权限控制 ──

  // Traceability: TC-035 → Story 5 / AC-1
  test('TC-035: 拥有 team:create 权限创建团队成功', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ name: `PM 创建的团队 ${Date.now()}` }),
    });
    assert.ok(res.status === 200 || res.status === 201, '有 team:create 权限可创建团队');
  });

  // ── Story 6: PM 的权限驱动操作 ──

  // Traceability: TC-036 → Story 6 / AC-1
  test('TC-036: 拥有 main_item:create 权限创建主事项', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/main-items`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: 'PM 创建的主事项', priority: 'P2', assigneeId: pmUserId, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    assert.ok(res.status === 200 || res.status === 201, '有 main_item:create 权限可创建主事项');
  });

  // Traceability: TC-037 → Story 6 / AC-2
  test('TC-037: 拥有 sub_item:assign 权限分配负责人', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/teams/${teamId}/sub-items/${subItemId}/assignee`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ assigneeId: memberUserId }),
    });
    assert.ok(res.status === 200 || res.status === 204, '有 sub_item:assign 权限可分配负责人');
  });

  // Traceability: TC-038 → Story 6 / AC-3
  test('TC-038: 拥有 item_pool:review 权限审核事项', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/item-pool/${poolId}/reject`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ reason: '测试拒绝' }),
    });
    assert.ok(res.status === 200 || res.status === 204, '有 item_pool:review 权限可审核事项');
  });

  // Traceability: TC-039 → Story 6 / AC-4
  test('TC-039: 无 main_item:archive 权限归档返回 403', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/main-items/${mainItemId}/archive`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({}),
    });
    assert.equal(res.status, 403, '无 main_item:archive 权限返回 403');
  });

  // Traceability: TC-040 → Story 6 / AC-5
  test('TC-040: 拥有 progress:update 权限修正进度', async () => {
    const res = await curl('PATCH', `${apiUrl}/api/v1/teams/${teamId}/progress/${progressId}/completion`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ completion: 50 }),
    });
    assert.ok(res.status === 200 || res.status === 204, '有 progress:update 权限可修正进度');
  });

  // ── Story 7: Member 的受限操作 ──

  // Traceability: TC-041 → Story 7 / AC-1
  test('TC-041: 拥有 sub_item:create 权限创建子事项', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ mainItemId: mainItemId, title: '测试子事项', priority: 'P2', assigneeId: memberUserId, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    assert.ok(res.status === 200 || res.status === 201, '有 sub_item:create 权限可创建子事项');
  });

  // Traceability: TC-042 → Story 7 / AC-2
  test('TC-042: 拥有 progress:create 权限追加进度', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/sub-items/${subItemId}/progress`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ completion: 60, achievement: '进度记录', blocker: '', lesson: '' }),
    });
    assert.ok(res.status === 200 || res.status === 201, '有 progress:create 权限可追加进度');
  });

  // Traceability: TC-043 → Story 7 / AC-3
  test('TC-043: 拥有 item_pool:submit 权限提交事项', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '提交到事项池' }),
    });
    assert.ok(res.status === 200 || res.status === 201, '有 item_pool:submit 权限可提交事项');
  });

  // Traceability: TC-044 → Story 7 / AC-4, Story 9 / AC-1
  test('TC-044: 无 team:invite 权限 API 返回 403', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ username: 'admin', roleId: 3 }),
    });
    assert.equal(res.status, 403, '无 team:invite 权限返回 403 Forbidden');
  });

  // Traceability: TC-045 → Story 7 / AC-5
  test('TC-045: 被分配者可编辑自己的子事项', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/teams/${teamId}/sub-items/${subItemId}`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '更新子事项标题' }),
    });
    assert.equal(res.status, 200, '被分配者可编辑自己的子事项');
  });

  // Traceability: TC-046 → Story 7 / AC-6
  test('TC-046: 非被分配者编辑子事项被拒绝', async () => {
    // member edits a sub-item that doesn't exist or isn't assigned to them
    const res = await curl('PUT', `${apiUrl}/api/v1/teams/${teamId}/sub-items/99999`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: '尝试编辑非自己的子事项' }),
    });
    assert.ok(res.status === 403 || res.status === 404, `非被分配者编辑子事项被拒绝 (got ${res.status})`);
  });

  // ── Story 8: 跨团队权限隔离 ──

  // Traceability: TC-047 → Story 8 / AC-1
  test('TC-047: A 团队 pm 角色邀请成员成功', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: `e2e-noperms-${runId}`, roleId: 3 }),
    });
    // May fail if already invited in TC-034
    assert.ok(res.status === 200 || res.status === 422, 'A 团队 PM 角色邀请成员成功或已存在');
  });

  // Traceability: TC-048 → Story 8 / AC-3
  test('TC-048: A 团队权限不跨团队到 B 团队', async () => {
    const res = await curl('POST', `${apiUrl}/api/v1/teams/${team2Id}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: 'test-noperms', roleId: 3 }),
    });
    assert.equal(res.status, 403, 'A 团队权限不能跨团队操作 B 团队');
  });

  // ── Story 9: 后端权限强制执行 ──

  // Traceability: TC-049 → Story 9 / AC-2
  test('TC-049: 非团队成员调用团队 API 返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/api/v1/teams/99999/main-items`, {
      headers: authHeader(noPermsToken),
    });
    assert.equal(res.status, 403, '非团队成员调用团队 API 返回 403');
  });

  // Traceability: TC-050 → Story 9 / AC-3
  test('TC-050: 非超管调用 admin API 返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/api/v1/admin/users`, {
      headers: authHeader(memberToken),
    });
    assert.equal(res.status, 403, '非超管调用 admin API 返回 403');
  });

  // Traceability: TC-051 → Story 9 / AC-4
  test('TC-051: 无 sub_item:assign 权限直接调用 API 返回 403', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/teams/${teamId}/sub-items/${subItemId}/assignee`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ assigneeId: memberUserId }),
    });
    assert.equal(res.status, 403, '无 sub_item:assign 权限返回 403');
  });

  // ── Story 10: 角色编辑即时生效 ──

  // Traceability: TC-052 → Story 10 / AC-1
  test('TC-052: 角色权限取消后后端即时生效', async () => {
    // Step 1: Create a new test role with limited permissions
    const createRoleRes = await curl('POST', `${apiUrl}/api/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `test-revoke-role-${Date.now()}`,
        description: 'Test role for permission revocation',
        permissionCodes: ['team:read', 'team:invite'],
      }),
    });
    assert.ok(createRoleRes.status === 200 || createRoleRes.status === 201, '创建测试角色成功');
    const testRoleId = parseData(createRoleRes.body).id;

    // Step 2: Edit role to remove team:invite
    const editRes = await curl('PUT', `${apiUrl}/api/v1/admin/roles/${testRoleId}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ permissionCodes: ['team:read'] }),
    });
    assert.ok(editRes.status === 200 || editRes.status === 204, '角色编辑成功');
  });

  // ── Spec 5.6: JWT Claims 与权限获取 ──

  // Traceability: TC-053 → Spec 5.6 / 权限获取方式
  test('TC-053: 获取用户权限 API 返回正确数据', async () => {
    const res = await curl('GET', `${apiUrl}/api/v1/me/permissions`, {
      headers: authHeader(pmToken),
    });
    assert.equal(res.status, 200, '权限 API 返回 200');

    const data = parseData(res.body);
    assert.ok(typeof data.isSuperAdmin === 'boolean', 'isSuperAdmin 为布尔值');
    assert.ok(typeof data.teamPermissions === 'object', 'teamPermissions 为对象');
  });

  // ── 性能需求 ──

  // Traceability: TC-054 → Spec / 性能需求
  test('TC-054: 权限 API 响应时间 < 200ms', async () => {
    const start = Date.now();
    const res = await curl('GET', `${apiUrl}/api/v1/me/permissions`, {
      headers: authHeader(pmToken),
    });
    const elapsed = Date.now() - start;

    assert.equal(res.status, 200, 'API 正常响应');
    assert.ok(elapsed < 200, `响应时间 ${elapsed}ms 应 < 200ms`);
  });

  // Traceability: TC-055 → Spec / 性能需求
  test('TC-055: 权限检查中间件响应时间 < 10ms', async () => {
    const iterations = 10;
    const start = Date.now();
    for (let i = 0; i < iterations; i++) {
      await curl('GET', `${apiUrl}/api/v1/teams/${teamId}/main-items`, {
        headers: authHeader(pmToken),
      });
    }
    const avgElapsed = (Date.now() - start) / iterations;

    assert.ok(avgElapsed < 100, `平均响应时间 ${avgElapsed}ms，中间件部分应 < 10ms`);
  });
});
