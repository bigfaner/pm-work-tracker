import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
const runId = Date.now();

function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

function parseResponse(body: string): { code: number; data: any; message?: string } {
  return JSON.parse(body);
}

async function login(username: string, password: string): Promise<string> {
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username, password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    expect(data.code).toBe(0);
    return data.data.token;
  }
  throw new Error('Login failed after retries: rate limited');
}

async function createTestRole(token: string, name: string, permissionCodes: string[]): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/admin/roles`, {
    headers: authHeader(token),
    body: JSON.stringify({ name, permissionCodes }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.id);
}

async function softDeleteRole(token: string, roleId: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/roles/${roleId}`, {
    headers: authHeader(token),
  });
  expect(res.status === 200).toBeTruthy();
}

test.describe('Soft-Delete — Roles (TC-001, TC-002)', () => {
  test.beforeAll(async () => {
    superadminToken = await login('admin', 'admin123');
  });

  // ── TC-001: Deleted role excluded from role list API response ──

  // Traceability: TC-001 -> Story 1 / AC-1
  test('TC-001: Deleted role excluded from role list API response', async () => {
    const roleBizKey = await createTestRole(
      superadminToken,
      `e2e-deleted-role-${runId}`,
      ['team:read'],
    );
    await softDeleteRole(superadminToken, roleBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const found = (data.data?.items ?? data.data ?? []).some(
      (r: any) => String(r.bizKey ?? r.id) === roleBizKey || r.name === `e2e-deleted-role-${runId}`,
    );
    expect(found).toBe(false);
  });

  // ── TC-002: Deleted role returns 404 when accessed by bizKey ──

  // Traceability: TC-002 -> Story 1 / AC-2
  test('TC-002: Deleted role returns 404 when accessed by ID', async () => {
    const roleBizKey = await createTestRole(
      superadminToken,
      `e2e-bizkey-role-${runId}`,
      ['team:read'],
    );
    await softDeleteRole(superadminToken, roleBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/roles/${roleBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
  });
});
