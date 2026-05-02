import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let teamBizKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
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

test.describe('RBAC — Item Create & Assign (TC-036, TC-037)', () => {
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
      body: JSON.stringify({ name: `RBAC Items Team ${runId}`, code: randomCode() }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    teamBizKey = extractBizKey(parseData(teamRes.body));

    // 4. Create test users
    const users = [
      { username: `e2e-pm-ia-${runId}`, displayName: 'PM IA', keyHolder: 'pm' },
      { username: `e2e-member-ia-${runId}`, displayName: 'Member IA', keyHolder: 'member' },
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

    // 5. Add PM and member users to team
    const addPmRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm-ia-${runId}`, roleKey: memberRoleKey }),
    });
    expect(addPmRes.status === 200 || addPmRes.status === 201).toBeTruthy();

    const addMemberRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-member-ia-${runId}`, roleKey: memberRoleKey }),
    });
    expect(addMemberRes.status === 200 || addMemberRes.status === 201).toBeTruthy();

    // Transfer PM role
    const transferRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();

    // 6. Create main item in team
    const mainItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        title: 'Test Main Item IA',
        priority: 'P2',
        assigneeKey: pmUserBizKey,
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });
    expect(mainItemRes.status === 200 || mainItemRes.status === 201).toBeTruthy();
    mainItemBizKey = extractBizKey(parseData(mainItemRes.body));

    // 7. Create sub item under main item
    const subItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'Test Sub Item IA', priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(subItemRes.status === 200 || subItemRes.status === 201).toBeTruthy();
    subItemBizKey = extractBizKey(parseData(subItemRes.body));
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
});
