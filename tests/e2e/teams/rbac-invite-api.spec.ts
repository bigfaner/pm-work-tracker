import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let teamBizKey: string;
let pmUserBizKey: string;
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

test.describe('RBAC — Team Invite & Create (TC-034, TC-035)', () => {
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
    const memberRole = roles.find((r) => r.roleName === 'member');
    expect(memberRole).toBeTruthy();
    memberRoleKey = memberRole!.bizKey;

    // 3. Create test team
    const teamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `RBAC Invite Team ${runId}`, code: randomCode() }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    teamBizKey = extractBizKey(parseData(teamRes.body));

    // 4. Create PM user
    const pmUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-inv-${runId}`, displayName: 'PM Invite' }),
    });
    expect(pmUserRes.status === 200 || pmUserRes.status === 201).toBeTruthy();
    const pmData = parseData(pmUserRes.body);
    pmUserBizKey = extractBizKey(pmData);
    pmToken = await loginWithRetry(`e2e-pm-inv-${runId}`, pmData.initialPassword);

    // 5. Add PM to team as member, then transfer PM role
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-inv-${runId}`, roleKey: memberRoleKey }),
    });
    const transferRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();

    // 6. Create noPerms user for invite target
    const noPermsRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-noperms-inv-${runId}`, displayName: 'NoPerms Invite' }),
    });
    expect(noPermsRes.status === 200 || noPermsRes.status === 201).toBeTruthy();
  });

  // ── Story 2: PM 在邀请成员时指定角色 ──

  // Traceability: TC-034 → Story 2 / AC-2
  test('TC-034: 邀请用户加入团队并分配角色', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ username: `e2e-noperms-inv-${runId}`, roleKey: memberRoleKey }),
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
});
