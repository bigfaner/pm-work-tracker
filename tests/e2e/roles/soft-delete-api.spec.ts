import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, authHeader, getApiToken, parseApiBody, createTestRole, softDeleteRole } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
const runId = Date.now();

test.describe('Soft-Delete — Roles (TC-001, TC-002)', () => {
  test.beforeAll(async () => {
    superadminToken = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
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
    const data = parseApiBody(res.body);
    const found = (data?.items ?? data ?? []).some(
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
