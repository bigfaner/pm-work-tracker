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

async function createTestUser(token: string, username: string, displayName: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/admin/users`, {
    headers: authHeader(token),
    body: JSON.stringify({ username, displayName }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.id);
}

async function softDeleteUser(token: string, userBizKey: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${userBizKey}`, {
    headers: authHeader(token),
  });
  expect(res.status === 200).toBeTruthy();
}

async function createTestTeam(token: string, name: string): Promise<string> {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
  const res = await curl('POST', `${apiUrl}/v1/teams`, {
    headers: authHeader(token),
    body: JSON.stringify({ name, code }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.teamKey ?? data.data.id);
}

async function createTestMainItem(token: string, teamBizKey: string, title: string, priority: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
    headers: authHeader(token),
    body: JSON.stringify({ title, priority, assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.id);
}

async function createTestSubItem(token: string, teamBizKey: string, mainItemBizKey: string, title: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
    headers: authHeader(token),
    body: JSON.stringify({ mainItemKey: mainItemBizKey, title, priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.id);
}

test.describe('Soft-Delete — Users (TC-005, TC-006)', () => {
  test.beforeAll(async () => {
    superadminToken = await login('admin', 'admin123');
  });

  // ── TC-005: FindByID returns NotFound for soft-deleted User ──

  // Traceability: TC-005 -> Story 3 / AC-1
  test('TC-005: FindByID returns NotFound for soft-deleted User', async () => {
    const userBizKey = await createTestUser(
      superadminToken,
      `e2e-deleted-user-tc005-${runId}`,
      'TC005 Deleted User',
    );
    await softDeleteUser(superadminToken, userBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/users/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
  });

  // ── TC-006: FindByID returns record for non-soft-deletable ProgressRecord ──

  // Traceability: TC-006 -> Story 3 / AC-2
  test('TC-006: Progress records are accessible (non-soft-deletable entity)', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc006-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC006 Main`, 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC006 Sub 01',
    );

    // Append a progress record
    const appendRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 50, note: 'TC006 progress' }),
    });
    expect(
      appendRes.status === 200 || appendRes.status === 201,
    ).toBeTruthy();

    // Verify progress list is accessible
    const listRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
    });
    expect(listRes.status).toBe(200);
  });
});
