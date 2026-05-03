import { test, expect } from '@playwright/test';
import { curl, authHeader, extractBizKey, getApiToken, apiBaseUrl } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let targetUserBizKey: string;
let ownUserBizKey: string;
const runId = Date.now();

async function createTestUser(token: string, username: string, displayName: string): Promise<{ bizKey: string; initialPassword: string }> {
  const res = await curl('POST', `${apiUrl}/v1/admin/users`, {
    headers: authHeader(token),
    body: JSON.stringify({ username, displayName }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = JSON.parse(res.body);
  return { bizKey: extractBizKey(data.data)!, initialPassword: data.data.initialPassword ?? '' };
}

test.describe('API E2E Tests — User Management Reset Password & Delete', () => {
  test.beforeAll(async () => {
    // 1. Login as seeded super admin
    superadminToken = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });

    // 2. Get own user bizKey (admin user) via admin user list
    const usersRes = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
    });
    expect(usersRes.status).toBe(200);
    const usersBody = JSON.parse(usersRes.body);
    const usersList: any[] = usersBody.data?.items ?? usersBody.data ?? (Array.isArray(usersBody) ? usersBody : []);
    const admin = usersList.find((u: any) => u.isSuperAdmin);
    expect(admin).toBeTruthy();
    ownUserBizKey = String(admin.bizKey);

    // 3. Create a target test user for reset/delete operations
    const target = await createTestUser(
      superadminToken,
      `e2e-target-${runId}`,
      'E2E Target User',
    );
    targetUserBizKey = target.bizKey;

    // 4. Create a non-super-admin user for permission tests
    const pm2 = await createTestUser(
      superadminToken,
      `e2e-pm2-${runId}`,
      'E2E PM2 User',
    );
    if (pm2.initialPassword) {
      try {
        pmToken = await getApiToken(apiBaseUrl, { username: `e2e-pm2-${runId}`, password: pm2.initialPassword });
      } catch {
        pmToken = 'invalid-token';
      }
    } else {
      pmToken = 'invalid-token';
    }
  });

  // ── Reset Password API ──

  // Traceability: TC-017 → Story 1 / AC-1, API Handbook - Reset Password
  test('TC-017: Reset password with valid request returns 200', async () => {
    const { bizKey: userId } = await createTestUser(
      superadminToken,
      `e2e-reset-target-${runId}`,
      'Reset Target',
    );

    const res = await curl('PUT', `${apiUrl}/v1/admin/users/${userId}/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'NewPass123456' }),
    });

    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.code).toBe(0);
    expect(data.data).toBeTruthy();
    expect(data.data.bizKey !== undefined || data.data.username !== undefined).toBeTruthy();
  });

  // Traceability: TC-018 → API Handbook - Reset Password Error Responses, PRD Spec Security
  test('TC-018: Reset password without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/admin/users/999/password`, {
      body: JSON.stringify({ newPassword: 'NewPass123456' }),
    });

    expect(res.status).toBe(401);
  });

  // Traceability: TC-019 → Story 5 / AC-1, API Handbook - Reset Password Error Responses
  test('TC-019: Reset password by non-super-admin returns 403', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/admin/users/${targetUserBizKey}/password`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ newPassword: 'NewPass123456' }),
    });

    expect(res.status).toBe(403);
  });

  // Traceability: TC-020 → PRD Spec Section 5.3 Validation Rules, API Handbook
  test('TC-020: Reset password with weak password returns 400', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/admin/users/${targetUserBizKey}/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'abc' }),
    });

    expect(res.status).toBe(400);
    const data = JSON.parse(res.body);
    expect(data.code !== 0).toBeTruthy();
  });

  // Traceability: TC-021 → API Handbook - Reset Password Error Responses
  test('TC-021: Reset password for non-existent user returns 404', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/admin/users/999999/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'NewPass123456' }),
    });

    expect(res.status).toBe(404);
  });

  // ── Delete User API ──

  // Traceability: TC-022 → Story 3 / AC-1, API Handbook - Delete User
  test('TC-022: Delete user with valid request returns 200', async () => {
    const { bizKey: userId } = await createTestUser(
      superadminToken,
      `e2e-delete-target-${runId}`,
      'Delete Target',
    );

    const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${userId}`, {
      headers: authHeader(superadminToken),
    });

    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.code).toBe(0);
  });

  // Traceability: TC-023 → API Handbook - Delete User Error Responses
  test('TC-023: Delete user without auth returns 401', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/admin/users/999`);

    expect(res.status).toBe(401);
  });

  // Traceability: TC-024 → Story 5 / AC-1, API Handbook - Delete User Error Responses
  test('TC-024: Delete user by non-super-admin returns 403', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${targetUserBizKey}`, {
      headers: authHeader(pmToken),
    });

    expect(res.status).toBe(403);
  });

  // Traceability: TC-025 → Story 4, API Handbook - Delete User Error Responses
  test('TC-025: Delete self returns 422', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${ownUserBizKey}`, {
      headers: authHeader(superadminToken),
    });

    expect(res.status).toBe(422);
    const data = JSON.parse(res.body);
    expect(data.code !== 0).toBeTruthy();
  });

  // Traceability: TC-026 → Story 3 / AC-2, API Handbook - Delete User Error Responses
  test('TC-026: Delete non-existent user returns 404', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/admin/users/999999`, {
      headers: authHeader(superadminToken),
    });

    expect(res.status).toBe(404);
  });
});
