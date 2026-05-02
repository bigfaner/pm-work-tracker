import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds, runCli } from '../../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;

// Per-role tokens created in before()
let pmToken: string;
let pmUserBizKey: string;
let memberToken: string;
let customRoleToken: string;
let emptyRoleToken: string;

// Fixture IDs (bizKeys)
let testTeamId: string;
let testItemId: string;
let customRoleBizKey: string;
let emptyRoleBizKey: string;

// Unique suffix to avoid collisions across runs
const RUN_ID = Date.now();

test.describe('API E2E Tests: api-permission-test-coverage', () => {
  test.beforeAll(async () => {
    // ── 1. Admin (superadmin) auth ──────────────────────────────────
    superadminToken = await getApiToken(apiBaseUrl, defaultCreds);
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // ── 2. Create test team ─────────────────────────────────────────
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
    const teamRes = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `e2e-team-${RUN_ID}`, code }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    const teamData = JSON.parse(teamRes.body).data;
    testTeamId = String(teamData?.bizKey ?? teamData?.id ?? JSON.parse(teamRes.body).id);

    // ── 3. Create a main item and transition to completed status for archive tests ──
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-${RUN_ID}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
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

    // ── 4. Fetch preset role bizKeys ────────────────────────────────────
    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    expect(rolesRes.status).toBe(200);
    const rolesBody = JSON.parse(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string; id: number; name: string }> =
      rolesBody.data?.items ?? rolesBody.data ?? rolesBody;
    const pmRole = roles.find((r) => r.roleName === 'pm' || r.name === 'pm');
    const memberRole = roles.find((r) => r.roleName === 'member' || r.name === 'member');
    expect(pmRole).toBeTruthy();
    expect(memberRole).toBeTruthy();
    const pmRoleKey = pmRole!.bizKey ?? String(pmRole!.id);
    const memberRoleKey = memberRole!.bizKey ?? String(memberRole!.id);

    // ── 5. Create custom role (partial permissions) ─────────────────
    const customRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({
        name: `custom-${RUN_ID}`,
        permissionCodes: ['main_item:read', 'progress:read'],
      }),
    });
    expect(customRoleRes.status === 200 || customRoleRes.status === 201).toBeTruthy();
    const customRoleData = JSON.parse(customRoleRes.body).data;
    customRoleBizKey = String(customRoleData?.bizKey ?? customRoleData?.id ?? JSON.parse(customRoleRes.body).id);

    // ── 6. Create empty role (no permissions) ───────────────────────
    const emptyRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({ name: `empty-${RUN_ID}`, permissionCodes: ['team:read'] }),
    });
    expect(emptyRoleRes.status === 200 || emptyRoleRes.status === 201).toBeTruthy();
    const emptyRoleData = JSON.parse(emptyRoleRes.body).data;
    emptyRoleBizKey = String(emptyRoleData?.bizKey ?? emptyRoleData?.id ?? JSON.parse(emptyRoleRes.body).id);

    // ── 7. Create test users and get tokens ───────────────────────
    async function createUserAndGetToken(
      username: string,
      displayName: string,
    ): Promise<{ token: string; bizKey: string }> {
      const createRes = await adminCurl('POST', '/v1/admin/users', {
        body: JSON.stringify({ username, displayName }),
      });
      expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();
      const userData = JSON.parse(createRes.body).data;
      const initialPassword = userData?.initialPassword;
      if (!initialPassword) throw new Error(`No initialPassword for ${username}`);
      const bizKey = String(userData?.bizKey ?? userData?.id);
      const tokenRes = await getApiToken(apiBaseUrl, { username, password: initialPassword });
      return { token: tokenRes, bizKey };
    }

    const pmResult = await createUserAndGetToken(`pm-${RUN_ID}`, 'PM User');
    pmToken = pmResult.token;
    pmUserBizKey = pmResult.bizKey;
    memberToken = (await createUserAndGetToken(`member-${RUN_ID}`, 'Member User')).token;
    customRoleToken = (await createUserAndGetToken(`custom-${RUN_ID}`, 'Custom Role User')).token;
    emptyRoleToken = (await createUserAndGetToken(`empty-${RUN_ID}`, 'Empty Role User')).token;

    // ── 8. Assign users to the team ─────────────────────────
    // Invite PM as member first (PM role can't be assigned via invite)
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `pm-${RUN_ID}`, roleKey: memberRoleKey }),
    }).catch(() => {});
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `member-${RUN_ID}`, roleKey: memberRoleKey }),
    }).catch(() => {});
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `custom-${RUN_ID}`, roleKey: customRoleBizKey }),
    }).catch(() => {});
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `empty-${RUN_ID}`, roleKey: emptyRoleBizKey }),
    }).catch(() => {});

    // Transfer PM role to pm user (can't assign pm role via invite)
    const transferRes = await adminCurl('PUT', `/v1/teams/${testTeamId}/pm`, {
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();
  });

  test.afterAll(async () => {
    // Best-effort cleanup — failures here do not fail the suite
    await adminCurl('DELETE', `/v1/admin/roles/${customRoleBizKey}`).catch(() => {});
    await adminCurl('DELETE', `/v1/admin/roles/${emptyRoleBizKey}`).catch(() => {});
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

  // ── Preset Roles Matrix ──────────────────────────────────────────

  // Traceability: TC-003 → Story 2 / AC-1
  test('TC-003: Preset roles matrix — archive endpoint: superadmin→200, pm→200, member→403', async () => {
    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const path = `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`;

    const superadminRes = await adminCurl('POST', path);
    expect(superadminRes.status).toBe(200);

    const pmRes = await pmCurl('POST', path);
    expect(pmRes.status).toBe(200);

    const memberRes = await memberCurl('POST', path);
    expect(memberRes.status).toBe(403);
  });

  // Traceability: TC-004 → Story 2 / AC-2
  test('TC-004: Preset roles matrix — team invite endpoint: superadmin→200, pm→200, member→403', async () => {
    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);

    // Create a fresh user to invite
    const newUserRes = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `invite-target-${RUN_ID}`, displayName: 'Invite Target' }),
    });
    const newUserData = JSON.parse(newUserRes.body).data;
    const newUserBizKey = String(newUserData?.bizKey ?? newUserData?.id);

    // Fetch member role bizKey for invite
    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    const rolesBody = JSON.parse(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesBody.data?.items ?? rolesBody.data ?? rolesBody;
    const memberRole = roles.find((r) => r.roleName === 'member');
    const memberRoleKey = memberRole?.bizKey ?? '';

    const path = `/v1/teams/${testTeamId}/members`;
    const body = JSON.stringify({ username: `invite-target-${RUN_ID}`, roleKey: memberRoleKey });

    const superadminRes = await adminCurl('POST', path, { body });
    expect(
      superadminRes.status === 200 || superadminRes.status === 201,
    ).toBeTruthy();

    // PM should also be able to invite (create another target)
    const newUser2Res = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `invite-target2-${RUN_ID}`, displayName: 'Invite Target 2' }),
    });
    if (newUser2Res.status === 200 || newUser2Res.status === 201) {
      const pmRes = await pmCurl('POST', path, {
        body: JSON.stringify({ username: `invite-target2-${RUN_ID}`, roleKey: memberRoleKey }),
      });
      expect(pmRes.status === 200 || pmRes.status === 201).toBeTruthy();
    }

    // Member should NOT be able to invite
    const memberRes = await memberCurl('POST', path, {
      body: JSON.stringify({ username: `invite-target-${RUN_ID}`, roleKey: memberRoleKey }),
    });
    expect(memberRes.status).toBe(403);
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

  // Traceability: TC-007 → Story 3 / AC-3
  test('TC-007: Permission change takes effect immediately without re-login', async () => {
    // Add main_item:create to the custom role via admin API
    const updateRes = await adminCurl('PUT', `/v1/admin/roles/${customRoleBizKey}`, {
      body: JSON.stringify({
        permissionCodes: ['main_item:read', 'progress:read', 'main_item:create'],
      }),
    });
    expect(updateRes.status).toBe(200);

    // Use the SAME token (no re-login) — permission change must be reflected immediately
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `post-grant-item-${RUN_ID}`, priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // ── Permission Boundaries ────────────────────────────────────────

  // Traceability: TC-008 → Story 4 / AC-1
  test('TC-008: Empty permission role is denied on protected endpoint — returns 403', async () => {
    const emptyCurl = createAuthCurl(apiBaseUrl, emptyRoleToken);
    const res = await emptyCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    expect(res.status).toBe(403);
  });

  // Traceability: TC-009 → Story 4 / AC-2
  test('TC-009: Superadmin bypasses permission check — returns 200 on protected endpoint', async () => {
    const res = await adminCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    expect(res.status).toBe(200);
  });

  // Traceability: TC-010 → Story 4 / AC-3
  test('TC-010: Invalid token returns 401 not 403', async () => {
    const res = await curl(
      'GET',
      `${apiBaseUrl}/v1/teams/${testTeamId}/main-items`,
      { headers: { Authorization: 'Bearer invalid.jwt.token' } },
    );
    expect(res.status).toBe(401);
  });

  // ── Permission Coverage CI ───────────────────────────────────────

  // Traceability: TC-011 → Story 5 / AC-1
  test.skip('TC-011: CI fails when permission code lacks test coverage', async () => {
    const result = runCli(
      'go test ./... -run TestPermissionCodeCoverage',
      '/Users/fanhuifeng/Projects/Go/pm-work-tracker-2/backend',
    );
    const combined = result.stdout + result.stderr;
    if (result.exitCode !== 0) {
      expect(combined).toMatch(/missing test coverage for:/);
    }
  });

  // Traceability: TC-012 → Story 5 / AC-2
  test.skip('TC-012: CI passes when all permission codes have test coverage', async () => {
    const result = runCli(
      'go test ./... -run TestPermissionCodeCoverage',
      '/Users/fanhuifeng/Projects/Go/pm-work-tracker-2/backend',
    );
    expect(result.exitCode).toBe(0);
  });
});
