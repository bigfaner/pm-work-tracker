import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from '../../helpers.js';

// Helpers to create test fixtures dynamically
async function createTeam(authCurl: ReturnType<typeof createAuthCurl>): Promise<string> {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
  const res = await authCurl('POST', '/v1/teams', {
    body: JSON.stringify({ name: `e2e-${code}`, code }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = JSON.parse(res.body);
  const bizKey = data.data?.bizKey ?? data.data?.id;
  expect(bizKey).toBeTruthy();
  return String(bizKey);
}

async function getPMRoleBizKey(authCurl: ReturnType<typeof createAuthCurl>): Promise<string> {
  const res = await authCurl('GET', '/v1/admin/roles');
  expect(res.status).toBe(200);
  const data = JSON.parse(res.body);
  const roles: any[] = data.data?.items ?? data.data ?? [];
  const pm = roles.find((r: any) => r.roleName === 'pm' || r.name === 'pm');
  expect(pm).toBeTruthy();
  return String(pm.bizKey ?? pm.id);
}

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  test.beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Main item response contains snowflake teamKey, not small uint', async () => {
    const teamBizKey = await createTeam(authCurl);
    const today = new Date().toISOString().split('T')[0];
    const future = new Date(Date.now() + 30 * 86400000).toISOString().split('T')[0];

    // Get admin user bizKey
    const meRes = await authCurl('GET', '/v1/me/permissions');
    const adminBizKey = '1'; // admin is always user 1 in seed

    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/main-items`, {
      body: JSON.stringify({
        title: 'TC-001 test item',
        priority: 'P2',
        assigneeKey: adminBizKey,
        startDate: today,
        expectedEndDate: future,
      }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    const teamKey = item.teamKey ?? item.team_key;
    expect(teamKey).toBeTruthy();
    // teamKey must be the snowflake bizKey (large number), not a small uint like 1, 2, 3
    expect(
      String(teamKey),
    ).toBe(teamBizKey);
  });

  // Traceability: TC-002 → Story 2 / AC-1
  test('TC-002: Invite member with PM role bizKey returns ErrCannotAssignPMRole', async () => {
    const teamBizKey = await createTeam(authCurl);
    const pmRoleBizKey = await getPMRoleBizKey(authCurl);

    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/members`, {
      body: JSON.stringify({
        username: 'admin', // admin is already PM, but PM role assignment should be rejected first
        roleKey: pmRoleBizKey,
      }),
    });
    expect(
      res.status === 400 || res.status === 422,
    ).toBeTruthy();
    const body = res.body.toLowerCase();
    expect(
      body.includes('pm') || body.includes('cannot') || body.includes('role'),
    ).toBeTruthy();
  });

  // Traceability: TC-003 → Spec Section 5.5 — bizKey 校验规则 (rule 1)
  test('TC-003: Non-numeric teamId in URL returns 400', async () => {
    const res = await authCurl('GET', '/v1/teams/abc/main-items');
    expect(res.status).toBe(400);
  });

  // Traceability: TC-004 → Spec Section 5.5 — bizKey 校验规则 (rule 2)
  test('TC-004: Non-positive teamId in URL returns 400', async () => {
    const resZero = await authCurl('GET', '/v1/teams/0/main-items');
    expect(resZero.status).toBe(400);

    const resNeg = await authCurl('GET', '/v1/teams/-1/main-items');
    expect(resNeg.status).toBe(400);
  });

  // Traceability: TC-005 → Spec Section 5.5 — bizKey 校验规则 (rule 3)
  test('TC-005: Non-existent teamId returns 404', async () => {
    const res = await authCurl('GET', '/v1/teams/999999999999999999/main-items');
    expect(res.status).toBe(404);
  });
});
