import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let teamBizKey: string;
let poolBizKey: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
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

test.describe('RBAC — Item Pool Review (TC-038)', () => {
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
      body: JSON.stringify({ name: `RBAC Pool Team ${runId}`, code: randomCode() }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    teamBizKey = extractBizKey(parseData(teamRes.body));

    // 4. Create PM and member users
    const users = [
      { username: `e2e-pm-pool-${runId}`, displayName: 'PM Pool', keyHolder: 'pm' },
      { username: `e2e-member-pool-${runId}`, displayName: 'Member Pool', keyHolder: 'member' },
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

      const token = await loginWithRetry(u.username, initialPassword);
      if (u.keyHolder === 'pm') pmToken = token;
      if (u.keyHolder === 'member') memberToken = token;
    }

    // 5. Add PM and member to team
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-pool-${runId}`, roleKey: memberRoleKey }),
    });
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-member-pool-${runId}`, roleKey: memberRoleKey }),
    });

    // Transfer PM role
    const transferRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();

    // 6. Create item pool entry
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: 'Test Pool Item Review' }),
    });
    expect(poolRes.status === 200 || poolRes.status === 201).toBeTruthy();
    poolBizKey = extractBizKey(parseData(poolRes.body));
  });

  // Traceability: TC-038 → Story 6 / AC-3
  test('TC-038: 拥有 item_pool:review 权限审核事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}/reject`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ reason: '测试拒绝' }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });
});
