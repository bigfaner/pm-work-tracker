import { test, expect } from '@playwright/test';
import {
  curl, apiBaseUrl, getApiToken, authHeader, parseApiBody, extractBizKey,
  randomCode, setupRbacFixtures,
} from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let teamBizKey: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
let poolBizKey: string;
let memberRoleKey: string;
const runId = Date.now();

const parseData = parseApiBody;

test.describe('Untested Endpoint Coverage (TC-082..TC-093)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    teamBizKey = f.teamBizKey;
    pmUserBizKey = f.pmUserBizKey;
    memberUserBizKey = f.memberUserBizKey;
    memberRoleKey = f.memberRoleKey;

    // Create main item + sub-item + progress
    const itemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        title: 'Endpoint Main', priority: 'P1',
        assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    mainItemBizKey = extractBizKey(parseData(itemRes.body))!;

    const subRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        mainItemKey: mainItemBizKey, title: 'Endpoint Sub', priority: 'P2',
        assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    subItemBizKey = extractBizKey(parseData(subRes.body))!;

    // Create pool item
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: 'Endpoint Pool Item' }),
    });
    poolBizKey = extractBizKey(parseData(poolRes.body))!;
  });

  // ── POST /v1/auth/logout ─────────────────────────────────────────

  test('TC-082: POST /v1/auth/logout returns 200', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/logout`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
  });

  // ── GET /v1/teams/:teamId/search-users ────────────────────────────

  test('TC-083: GET search-users returns matching users', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/search-users?keyword=admin`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-084: GET search-users requires team:invite permission', async () => {
    // Create a no-perm user not in team
    const tmpRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `noperm-tc084-${runId}`, displayName: 'NoPerms TC084' }),
    });
    const tmpData = parseData(tmpRes.body);
    const tmpToken = await getApiToken(apiBaseUrl, { username: `noperm-tc084-${runId}`, password: tmpData.initialPassword });

    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/search-users?keyword=admin`, {
      headers: authHeader(tmpToken),
    });
    // Not a team member → 403
    expect(res.status === 403 || res.status === 404).toBeTruthy();
  });

  // ── PUT /v1/teams/:teamId/members/:userId/role ────────────────────

  test('TC-085: PUT members/:userId/role changes role successfully', async () => {
    // Create a custom role for testing
    const roleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `role-tc085-${runId}`, permissionCodes: ['team:read'] }),
    });
    const customRoleKey = extractBizKey(parseData(roleRes.body))!;

    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/members/${memberUserBizKey}/role`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ roleKey: customRoleKey }),
    });
    expect(res.status).toBe(200);

    // Restore member role
    await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/members/${memberUserBizKey}/role`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ roleKey: memberRoleKey }),
    });
  });

  // ── GET available-transitions (main item) ─────────────────────────

  test('TC-086: GET main-items available-transitions returns valid list', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/available-transitions`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    // Main item starts in 'pending' → should have transitioning options
    const transitions: string[] = Array.isArray(data) ? data : (data?.transitions ?? data?.items ?? []);
    expect(transitions.length).toBeGreaterThan(0);
  });

  // ── GET available-transitions (sub item) ──────────────────────────

  test('TC-087: GET sub-items available-transitions returns valid list', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/available-transitions`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    const transitions: string[] = Array.isArray(data) ? data : (data?.transitions ?? data?.items ?? []);
    expect(transitions.length).toBeGreaterThan(0);
  });

  // ── GET /v1/teams/:teamId/item-pool/:poolId ───────────────────────

  test('TC-088: GET item-pool/:poolId returns single pool item', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(extractBizKey(data)).toBe(poolBizKey);
  });

  // ── PUT /v1/teams/:teamId/item-pool/:poolId ───────────────────────

  test('TC-089: PUT item-pool/:poolId updates pool item title', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: `Updated Pool ${runId}` }),
    });
    expect(res.status).toBe(200);
  });

  // ── POST /v1/teams/:teamId/item-pool/:poolId/convert-to-main ─────

  test('TC-090: POST item-pool convert-to-main creates main item', async () => {
    // Create a fresh pool item for conversion
    const freshPoolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: 'Convert to Main Test' }),
    });
    const freshPoolBizKey = extractBizKey(parseData(freshPoolRes.body))!;

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${freshPoolBizKey}/convert-to-main`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // ── POST /v1/teams/:teamId/item-pool/:poolId/assign (convert to sub) ─

  test('TC-091: POST item-pool assign creates sub-item under main', async () => {
    const freshPoolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: 'Assign to Sub Test' }),
    });
    const freshPoolBizKey = extractBizKey(parseData(freshPoolRes.body))!;

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${freshPoolBizKey}/assign`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // ── GET /v1/admin/users/:userId ───────────────────────────────────

  test('TC-092: GET admin/users/:userId returns user details', async () => {
    const res = await curl('GET', `${apiUrl}/v1/admin/users/${pmUserBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(extractBizKey(data)).toBe(pmUserBizKey);
  });

  // ── PUT /v1/admin/users/:userId/status ────────────────────────────

  test('TC-093: PUT admin/users/:userId/status toggles user status', async () => {
    // Create temp user for toggle test
    const tmpRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `toggle-tc093-${runId}`, displayName: 'Toggle TC093' }),
    });
    const tmpData = parseData(tmpRes.body);
    const tmpBizKey = extractBizKey(tmpData)!;

    const res = await curl('PUT', `${apiUrl}/v1/admin/users/${tmpBizKey}/status`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ status: 'disabled' }),
    });
    expect(res.status).toBe(200);
  });
});
