import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../helpers.js';

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

test.describe('API E2E Tests — Items', () => {
  let adminAuth: { authHeader: Record<string, string>; token: string };
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
