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

test.describe('API E2E Tests — Users', () => {
  let adminAuth: { authHeader: Record<string, string>; token: string };

  test.beforeAll(async () => {
    adminAuth = await loginAs('admin', 'admin123');
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
});
