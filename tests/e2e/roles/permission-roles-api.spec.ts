import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds } from '../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;
let customRoleToken: string;

// Fixture IDs (bizKeys)
let testTeamId: string;
let testItemId: string;
let customRoleBizKey: string;
let emptyRoleBizKey: string;

// Unique suffix to avoid collisions across runs
const RUN_ID = Date.now();

test.describe('Permission — Custom Roles (TC-005, TC-006)', () => {
  test.beforeAll(async () => {
    // ── 1. Admin (superadmin) auth ──────────────────────────────────
    superadminToken = await getApiToken(apiBaseUrl, defaultCreds);
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // ── 2. Create test team ─────────────────────────────────────────
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
    const teamRes = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `e2e-team-roles-${RUN_ID}`, code }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    const teamData = JSON.parse(teamRes.body).data;
    testTeamId = String(teamData?.bizKey ?? teamData?.id ?? JSON.parse(teamRes.body).id);

    // ── 3. Create a main item and transition to completed status for archive tests ──
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-roles-${RUN_ID}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(itemRes.status === 200 || itemRes.status === 201).toBeTruthy();
    const itemData = JSON.parse(itemRes.body).data;
    testItemId = String(itemData?.bizKey ?? itemData?.id ?? JSON.parse(itemRes.body).id);

    // Transition item: pending → progressing → reviewing → completed
    for (const status of ['progressing', 'reviewing', 'completed']) {
      const statusRes = await adminCurl('PUT', `/v1/teams/${testTeamId}/main-items/${testItemId}/status`, {
        body: JSON.stringify({ status }),
      });
      expect(statusRes.status === 200 || statusRes.status === 204).toBeTruthy();
    }

    // ── 4. Create custom role (partial permissions) ─────────────────
    const customRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({
        name: `custom-roles-${RUN_ID}`,
        permissionCodes: ['main_item:read', 'progress:read'],
      }),
    });
    expect(customRoleRes.status === 200 || customRoleRes.status === 201).toBeTruthy();
    const customRoleData = JSON.parse(customRoleRes.body).data;
    customRoleBizKey = String(customRoleData?.bizKey ?? customRoleData?.id ?? JSON.parse(customRoleRes.body).id);

    // ── 5. Create empty role (no permissions) ───────────────────────
    const emptyRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({ name: `empty-roles-${RUN_ID}`, permissionCodes: ['team:read'] }),
    });
    expect(emptyRoleRes.status === 200 || emptyRoleRes.status === 201).toBeTruthy();
    const emptyRoleData = JSON.parse(emptyRoleRes.body).data;
    emptyRoleBizKey = String(emptyRoleData?.bizKey ?? emptyRoleData?.id ?? JSON.parse(emptyRoleRes.body).id);

    // ── 6. Create test user with custom role and get token ─────────
    const createRes = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `custom-roles-${RUN_ID}`, displayName: 'Custom Role User' }),
    });
    expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();
    const userData = JSON.parse(createRes.body).data;
    const initialPassword = userData?.initialPassword;
    customRoleToken = await getApiToken(apiBaseUrl, { username: `custom-roles-${RUN_ID}`, password: initialPassword });

    // ── 7. Assign user to the team ─────────────────────────
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `custom-roles-${RUN_ID}`, roleKey: customRoleBizKey }),
    }).catch(() => {});
  });

  test.afterAll(async () => {
    // Best-effort cleanup — failures here do not fail the suite
    await adminCurl('DELETE', `/v1/admin/roles/${customRoleBizKey}`).catch(() => {});
    await adminCurl('DELETE', `/v1/admin/roles/${emptyRoleBizKey}`).catch(() => {});
  });

  // ── Custom Role ──────────────────────────────────────────────────

  // Traceability: TC-005 → Story 3 / AC-1
  test('TC-005: Custom role with partial permissions allows read — GET /main-items returns 200', async () => {
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('GET', `/v1/teams/${testTeamId}/main-items`);
    expect(res.status).toBe(200);
  });

  // Traceability: TC-006 → Story 3 / AC-2
  test('TC-006: Custom role without create permission denies write — POST /main-items returns 403', async () => {
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: 'should-be-denied', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status).toBe(403);
  });
});
