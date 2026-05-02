import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds } from '../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;
let memberToken: string;

// Fixture IDs (bizKeys)
let testTeamId: string;
let testItemId: string;

// Unique suffix to avoid collisions across runs
const RUN_ID = Date.now();

test.describe('Permission — Archive (TC-001, TC-002)', () => {
  test.beforeAll(async () => {
    // ── 1. Admin (superadmin) auth ──────────────────────────────────
    superadminToken = await getApiToken(apiBaseUrl, defaultCreds);
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // ── 2. Create test team ─────────────────────────────────────────
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
    const teamRes = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `e2e-team-arch-${RUN_ID}`, code }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    const teamData = JSON.parse(teamRes.body).data;
    testTeamId = String(teamData?.bizKey ?? teamData?.id ?? JSON.parse(teamRes.body).id);

    // ── 3. Create a main item and transition to completed status for archive tests ──
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-arch-${RUN_ID}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
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

    // ── 4. Create member user ───────────────────────────────────────
    const createRes = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `member-arch-${RUN_ID}`, displayName: 'Member Archive' }),
    });
    expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();
    const userData = JSON.parse(createRes.body).data;
    const initialPassword = userData?.initialPassword;
    memberToken = await getApiToken(apiBaseUrl, { username: `member-arch-${RUN_ID}`, password: initialPassword });

    // ── 5. Fetch preset role bizKeys and add member to team ──────────
    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    expect(rolesRes.status).toBe(200);
    const rolesBody = JSON.parse(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesBody.data?.items ?? rolesBody.data ?? rolesBody;
    const memberRole = roles.find((r) => r.roleName === 'member');
    const memberRoleKey = memberRole?.bizKey ?? '';

    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `member-arch-${RUN_ID}`, roleKey: memberRoleKey }),
    }).catch(() => {});
  });

  // ── Permission Middleware ────────────────────────────────────────

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Permission injection grants access — archive endpoint returns 200 for user with main_item:archive', async () => {
    const res = await adminCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    expect(res.status).toBe(200);
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: Empty permission injection denies access — archive endpoint returns 403 for user without main_item:archive', async () => {
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const res = await memberCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    expect(res.status).toBe(403);
  });
});
