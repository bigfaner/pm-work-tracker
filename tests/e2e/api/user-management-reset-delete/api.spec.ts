import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl } from './helpers.js';

const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

let superadminToken: string;
let pmToken: string;
let targetUserId: number;
let ownUserId: number;
const runId = Date.now();

function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

function parseResponse(body: string): { code: number; data: any; message?: string } {
  return JSON.parse(body);
}

async function login(username: string, password: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
    body: JSON.stringify({ username, password }),
  });
  assert.equal(res.status, 200, `Login as ${username} should succeed`);
  const data = parseResponse(res.body);
  assert.equal(data.code, 0, `Login response code should be 0`);
  return data.data.token;
}

async function createTestUser(token: string, username: string, displayName: string): Promise<number> {
  const res = await curl('POST', `${apiUrl}/api/v1/admin/users`, {
    headers: authHeader(token),
    body: JSON.stringify({ username, displayName }),
  });
  assert.ok(res.status === 200 || res.status === 201, `Create user ${username}: ${res.status} ${res.body}`);
  const data = parseResponse(res.body);
  return data.data.id;
}

describe('API E2E Tests — User Management Reset Password & Delete', () => {
  before(async () => {
    // 1. Login as seeded super admin
    superadminToken = await login('admin', 'admin123');

    // 2. Get own user ID (admin user)
    const meRes = await curl('GET', `${apiUrl}/api/v1/me`, {
      headers: authHeader(superadminToken),
    });
    if (meRes.status === 200) {
      const meData = parseResponse(meRes.body);
      ownUserId = meData.data?.id ?? meData.data?.userId ?? 1;
    } else {
      ownUserId = 1;
    }

    // 3. Create a target test user for reset/delete operations
    targetUserId = await createTestUser(
      superadminToken,
      `e2e-target-${runId}`,
      'E2E Target User',
    );

    // 4. Login as target user to get a non-super-admin token
    const targetLoginRes = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
      body: JSON.stringify({ username: `e2e-target-${runId}`, password: '' }),
    });
    // Note: initial password comes from create response; we need it
    // For now, create a PM user via another approach
    // Create a second user and use it as non-super-admin
    const pmUserId = await createTestUser(
      superadminToken,
      `e2e-pm-${runId}`,
      'E2E PM User',
    );

    // Login as PM user using the initial password from create response
    const createRes = await curl('POST', `${apiUrl}/api/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `e2e-pm2-${runId}`, displayName: 'E2E PM2 User' }),
    });
    if (createRes.status === 200 || createRes.status === 201) {
      const createData = parseResponse(createRes.body);
      if (createData.data?.initialPassword) {
        try {
          pmToken = await login(`e2e-pm2-${runId}`, createData.data.initialPassword);
        } catch {
          pmToken = 'invalid-token';
        }
      } else {
        pmToken = 'invalid-token';
      }
    } else {
      pmToken = 'invalid-token';
    }
  });

  // ── Reset Password API ──

  // Traceability: TC-017 → Story 1 / AC-1, API Handbook - Reset Password
  test('TC-017: Reset password with valid request returns 200', async () => {
    // Create a fresh user for this test
    const userId = await createTestUser(
      superadminToken,
      `e2e-reset-target-${runId}`,
      'Reset Target',
    );

    const res = await curl('PUT', `${apiUrl}/api/v1/admin/users/${userId}/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'NewPass123' }),
    });

    assert.equal(res.status, 200, `Reset password should return 200, got ${res.status}`);
    const data = parseResponse(res.body);
    assert.equal(data.code, 0, 'Response code should be 0');
    assert.ok(data.data, 'Response should contain data');
    assert.ok(data.data.bizKey !== undefined || data.data.username !== undefined, 'Data should contain bizKey or username');
  });

  // Traceability: TC-018 → API Handbook - Reset Password Error Responses, PRD Spec Security
  test('TC-018: Reset password without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/admin/users/999/password`, {
      body: JSON.stringify({ newPassword: 'NewPass123' }),
    });

    assert.equal(res.status, 401, `Should return 401, got ${res.status}`);
  });

  // Traceability: TC-019 → Story 5 / AC-1, API Handbook - Reset Password Error Responses
  test('TC-019: Reset password by non-super-admin returns 403', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/admin/users/${targetUserId}/password`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ newPassword: 'NewPass123' }),
    });

    assert.equal(res.status, 403, `Non-super-admin should get 403, got ${res.status}`);
  });

  // Traceability: TC-020 → PRD Spec Section 5.3 Validation Rules, API Handbook
  test('TC-020: Reset password with weak password returns 400', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/admin/users/${targetUserId}/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'abc' }),
    });

    assert.equal(res.status, 400, `Weak password should return 400, got ${res.status}`);
    const data = parseResponse(res.body);
    assert.ok(data.code !== 0, 'Error response should have non-zero code');
  });

  // Traceability: TC-021 → API Handbook - Reset Password Error Responses
  test('TC-021: Reset password for non-existent user returns 404', async () => {
    const res = await curl('PUT', `${apiUrl}/api/v1/admin/users/999999/password`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPassword: 'NewPass123' }),
    });

    assert.equal(res.status, 404, `Non-existent user should return 404, got ${res.status}`);
  });

  // ── Delete User API ──

  // Traceability: TC-022 → Story 3 / AC-1, API Handbook - Delete User
  test('TC-022: Delete user with valid request returns 200', async () => {
    // Create a fresh user for deletion
    const userId = await createTestUser(
      superadminToken,
      `e2e-delete-target-${runId}`,
      'Delete Target',
    );

    const res = await curl('DELETE', `${apiUrl}/api/v1/admin/users/${userId}`, {
      headers: authHeader(superadminToken),
    });

    assert.equal(res.status, 200, `Delete user should return 200, got ${res.status}`);
    const data = parseResponse(res.body);
    assert.equal(data.code, 0, 'Response code should be 0');
  });

  // Traceability: TC-023 → API Handbook - Delete User Error Responses
  test('TC-023: Delete user without auth returns 401', async () => {
    const res = await curl('DELETE', `${apiUrl}/api/v1/admin/users/999`);

    assert.equal(res.status, 401, `Should return 401, got ${res.status}`);
  });

  // Traceability: TC-024 → Story 5 / AC-1, API Handbook - Delete User Error Responses
  test('TC-024: Delete user by non-super-admin returns 403', async () => {
    const res = await curl('DELETE', `${apiUrl}/api/v1/admin/users/${targetUserId}`, {
      headers: authHeader(pmToken),
    });

    assert.equal(res.status, 403, `Non-super-admin should get 403, got ${res.status}`);
  });

  // Traceability: TC-025 → Story 4, API Handbook - Delete User Error Responses
  test('TC-025: Delete self returns 422', async () => {
    const res = await curl('DELETE', `${apiUrl}/api/v1/admin/users/${ownUserId}`, {
      headers: authHeader(superadminToken),
    });

    assert.equal(res.status, 422, `Self-delete should return 422, got ${res.status}`);
    const data = parseResponse(res.body);
    assert.ok(data.code !== 0, 'Error response should have non-zero code');
  });

  // Traceability: TC-026 → Story 3 / AC-2, API Handbook - Delete User Error Responses
  test('TC-026: Delete non-existent user returns 404', async () => {
    const res = await curl('DELETE', `${apiUrl}/api/v1/admin/users/999999`, {
      headers: authHeader(superadminToken),
    });

    assert.equal(res.status, 404, `Non-existent user should return 404, got ${res.status}`);
  });
});
