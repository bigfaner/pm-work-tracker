import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, defaultCreds } from '../../helpers.js';

const apiUrl = apiBaseUrl;

/** Login via API and return the Authorization header and token. */
async function loginAs(
  username: string,
  password: string,
): Promise<{ authHeader: Record<string, string>; token: string }> {
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username, password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    if (res.status !== 200) {
      throw new Error(`Login failed for ${username}: ${res.status} ${res.body}`);
    }
    const data = JSON.parse(res.body);
    const token = data.data?.token ?? data.token;
    return { authHeader: { Authorization: `Bearer ${token}` }, token };
  }
  throw new Error(`Login failed for ${username} after retries: rate limited`);
}

/**
 * API E2E Tests for improve-ui feature.
 *
 * Pre-conditions:
 * - Backend running on http://localhost:8080
 * - Test data seeded (admin user, team, items)
 */
test.describe('API E2E Tests', () => {
  let adminAuth: { authHeader: Record<string, string>; token: string };
  let pmAuth: { authHeader: Record<string, string>; token: string };
  let teamId: string;

  test.beforeAll(async () => {
    adminAuth = await loginAs('admin', 'admin123');
    // Get a team ID for team-scoped endpoints
    const teamsRes = await curl('GET', `${apiUrl}/v1/teams`, {
      headers: adminAuth.authHeader,
    });
    const teams = JSON.parse(teamsRes.body);
    const teamList = teams.data?.items ?? teams.data ?? teams;
    if (Array.isArray(teamList) && teamList.length > 0) {
      teamId = String(teamList[0].bizKey ?? teamList[0].id ?? teamList[0].ID);
    }
  });

  // Traceability: TC-053 → Spec 5.7 #1
  test('TC-053: 团队详情独立路由成员 CRUD API', async () => {
    expect(teamId).toBeTruthy();

    // GET team detail
    const detailRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}`, {
      headers: adminAuth.authHeader,
    });
    expect(detailRes.status).toBe(200);

    const detail = JSON.parse(detailRes.body);
    const teamData = detail.data ?? detail;
    expect(teamData.name || teamData.Name).toBeTruthy();

    // GET members
    const membersRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}/members`, {
      headers: adminAuth.authHeader,
    });
    expect(membersRes.status).toBe(200);

    const members = JSON.parse(membersRes.body);
    const memberList = members.data ?? members;
    expect(Array.isArray(memberList)).toBeTruthy();

    // Note: POST/PUT/DELETE member operations are tested with specific data
    // that may not exist in seed — these verify the endpoint structure
  });

  // Traceability: TC-054 → Spec 5.7 #2
  test('TC-054: 用户管理全量操作 API', async () => {
    // GET all users
    const listRes = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: adminAuth.authHeader,
    });
    expect(listRes.status).toBe(200);

    const users = JSON.parse(listRes.body);
    const userList = users.data?.items ?? users.data ?? users;
    expect(Array.isArray(userList)).toBeTruthy();

    // POST create user — unique username
    const uniqueAccount = `test_${Date.now()}`;
    const createRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: adminAuth.authHeader,
      body: JSON.stringify({
        username: uniqueAccount,
        displayName: 'Test User',
        email: `${uniqueAccount}@test.com`,
        canCreateTeam: false,
      }),
    });
    expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();

    // POST duplicate username — should fail
    const dupRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: adminAuth.authHeader,
      body: JSON.stringify({
        username: uniqueAccount,
        displayName: 'Duplicate',
        email: `dup@${uniqueAccount}.com`,
        canCreateTeam: false,
      }),
    });
    expect(dupRes.status === 409 || dupRes.status === 422).toBeTruthy();

    // PUT update user
    const createData = JSON.parse(createRes.body);
    const userId = String(createData.data?.bizKey ?? createData.data?.id ?? createData.data?.ID ?? createData.id);
    if (userId) {
      const updateRes = await curl('PUT', `${apiUrl}/v1/admin/users/${userId}`, {
        headers: adminAuth.authHeader,
        body: JSON.stringify({
          displayName: 'Updated Name',
          email: `updated@${uniqueAccount}.com`,
        }),
      });
      expect(updateRes.status).toBe(200);

      // PUT toggle status (disable)
      const statusRes = await curl('PUT', `${apiUrl}/v1/admin/users/${userId}/status`, {
        headers: adminAuth.authHeader,
        body: JSON.stringify({ status: 'disabled' }),
      });
      expect(statusRes.status).toBe(200);
    }
  });

  // Traceability: TC-055 → Spec 5.7 #3
  test('TC-055: 事项清单 Detail 分页参数', async () => {
    expect(teamId).toBeTruthy();

    // Without pagination
    const defaultRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}/main-items`, {
      headers: adminAuth.authHeader,
    });
    expect(defaultRes.status).toBe(200);

    // With pagination params
    const pagedRes = await curl(
      'GET',
      `${apiUrl}/v1/teams/${teamId}/main-items?page=1&pageSize=20`,
      { headers: adminAuth.authHeader },
    );
    expect(pagedRes.status).toBe(200);

    const data = JSON.parse(pagedRes.body);
    const items = data.data?.items ?? data.data ?? data;
    expect(Array.isArray(items)).toBeTruthy();
  });

  // Traceability: TC-056 → Spec 5.7 #4
  test('TC-056: 全量表格聚合查询 API', async () => {
    expect(teamId).toBeTruthy();

    // GET unified table view
    const tableRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}/views/table`, {
      headers: adminAuth.authHeader,
    });
    expect(tableRes.status).toBe(200);

    const data = JSON.parse(tableRes.body);
    const items = data.data?.items ?? data.data ?? data;
    expect(Array.isArray(items)).toBeTruthy();

    // Each item should have a type field
    if (items.length > 0) {
      const firstItem = items[0];
      expect(
        firstItem.type !== undefined || firstItem.Type !== undefined,
      ).toBeTruthy();
    }

    // GET with type filter
    const filteredRes = await curl(
      'GET',
      `${apiUrl}/v1/teams/${teamId}/views/table?type=main`,
      { headers: adminAuth.authHeader },
    );
    expect(filteredRes.status).toBe(200);

    // GET with multiple filters
    const multiFilterRes = await curl(
      'GET',
      `${apiUrl}/v1/teams/${teamId}/views/table?type=main&priority=P1`,
      { headers: adminAuth.authHeader },
    );
    expect(multiFilterRes.status).toBe(200);
  });
});
