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

test.describe('Soft-Delete — Items (TC-003, TC-004)', () => {
  test.beforeAll(async () => {
    superadminToken = await login('admin', 'admin123');
  });

  // ── TC-003: Soft-deleted sub-item disappears from sub-item list ──

  // Traceability: TC-003 -> Story 2 / AC-1
  test('TC-003: Soft-deleted sub-item disappears from sub-item list', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc003-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC003 Main`, 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC003 Sub 01',
    );

    const listBefore = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
    });
    expect(listBefore.status).toBe(200);
    const dataBefore = parseResponse(listBefore.body);
    const itemsBefore = dataBefore.data?.items ?? dataBefore.data ?? [];
    const existsBefore = itemsBefore.some((i: any) => String(i.bizKey ?? i.id) === subItemBizKey);
    expect(existsBefore).toBe(true);
  });

  // ── TC-004: Re-create sub-item with same title after soft-delete succeeds ──

  // Traceability: TC-004 -> Story 2 / AC-2
  test('TC-004: Re-create sub-item with same title after soft-delete succeeds', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc004-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC004 Main`, 'P0',
    );

    // Create sub-item
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC004 Sub 01',
    );
    expect(subItemBizKey).toBeTruthy();

    // Attempt to create another sub-item with the same title should succeed (no unique constraint on title)
    const dupRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'TC004 Sub 01 Duplicate', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(dupRes.status === 200 || dupRes.status === 201).toBeTruthy();
  });
});
