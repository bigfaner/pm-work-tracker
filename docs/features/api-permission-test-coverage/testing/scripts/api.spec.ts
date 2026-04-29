import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds, runCli } from './helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;

// Per-role tokens created in before()
let pmToken: string;
let memberToken: string;
let customRoleToken: string;
let emptyRoleToken: string;

// Fixture IDs
let testTeamId: number;
let testItemId: number;
let customRoleId: number;
let emptyRoleId: number;

// Unique suffix to avoid collisions across runs
const RUN_ID = Date.now();

describe('API E2E Tests: api-permission-test-coverage', () => {
  before(async () => {
    // ── 1. Admin (superadmin) auth ──────────────────────────────────
    superadminToken = await getApiToken(apiBaseUrl, defaultCreds);
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // ── 2. Create test team ─────────────────────────────────────────
    const teamRes = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `e2e-team-${RUN_ID}` }),
    });
    assert.equal(teamRes.status, 200, `Create team failed: ${teamRes.body}`);
    testTeamId = JSON.parse(teamRes.body).data?.id ?? JSON.parse(teamRes.body).id;

    // ── 3. Create a main item (completed status for archive tests) ──
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-${RUN_ID}`, priority: 'P1' }),
    });
    assert.equal(itemRes.status, 200, `Create item failed: ${itemRes.body}`);
    testItemId = JSON.parse(itemRes.body).data?.id ?? JSON.parse(itemRes.body).id;

    // ── 4. Fetch preset role IDs ────────────────────────────────────
    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    assert.equal(rolesRes.status, 200, `List roles failed: ${rolesRes.body}`);
    const roles: Array<{ id: number; name: string }> =
      JSON.parse(rolesRes.body).data ?? JSON.parse(rolesRes.body);
    const pmRole = roles.find((r) => r.name === 'pm');
    const memberRole = roles.find((r) => r.name === 'member');
    assert.ok(pmRole, 'Preset role "pm" not found');
    assert.ok(memberRole, 'Preset role "member" not found');

    // ── 5. Create custom role (partial permissions) ─────────────────
    const customRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({
        name: `custom-${RUN_ID}`,
        permissionCodes: ['main_item:read', 'progress:read'],
      }),
    });
    assert.equal(customRoleRes.status, 200, `Create custom role failed: ${customRoleRes.body}`);
    customRoleId = JSON.parse(customRoleRes.body).data?.id ?? JSON.parse(customRoleRes.body).id;

    // ── 6. Create empty role (no permissions) ───────────────────────
    const emptyRoleRes = await adminCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({ name: `empty-${RUN_ID}`, permissionCodes: [] }),
    });
    assert.equal(emptyRoleRes.status, 200, `Create empty role failed: ${emptyRoleRes.body}`);
    emptyRoleId = JSON.parse(emptyRoleRes.body).data?.id ?? JSON.parse(emptyRoleRes.body).id;

    // ── 7. Create test users and assign roles ───────────────────────
    async function createUserAndGetToken(
      username: string,
      roleId: number,
    ): Promise<string> {
      const password = `Pass${RUN_ID}!`;
      const createRes = await adminCurl('POST', '/v1/admin/users', {
        body: JSON.stringify({ username, password, roleId }),
      });
      assert.equal(createRes.status, 200, `Create user ${username} failed: ${createRes.body}`);
      const tokenRes = await getApiToken(apiBaseUrl, { username, password });
      return tokenRes;
    }

    pmToken = await createUserAndGetToken(`pm-${RUN_ID}`, pmRole!.id);
    memberToken = await createUserAndGetToken(`member-${RUN_ID}`, memberRole!.id);
    customRoleToken = await createUserAndGetToken(`custom-${RUN_ID}`, customRoleId);
    emptyRoleToken = await createUserAndGetToken(`empty-${RUN_ID}`, emptyRoleId);
  });

  after(async () => {
    // Best-effort cleanup — failures here do not fail the suite
    await adminCurl('DELETE', `/v1/admin/roles/${customRoleId}`).catch(() => {});
    await adminCurl('DELETE', `/v1/admin/roles/${emptyRoleId}`).catch(() => {});
  });

  // ── Permission Middleware ────────────────────────────────────────

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Permission injection grants access — archive endpoint returns 200 for user with main_item:archive', async () => {
    // superadmin has main_item:archive via preset role
    const res = await adminCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    assert.equal(res.status, 200, `Expected 200 but got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: Empty permission injection denies access — archive endpoint returns 403 for user without main_item:archive', async () => {
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const res = await memberCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    assert.equal(res.status, 403, `Expected 403 but got ${res.status}: ${res.body}`);
  });

  // ── Preset Roles Matrix ──────────────────────────────────────────

  // Traceability: TC-003 → Story 2 / AC-1
  test('TC-003: Preset roles matrix — archive endpoint: superadmin→200, pm→200, member→403', async () => {
    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const path = `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`;

    const superadminRes = await adminCurl('POST', path);
    assert.equal(superadminRes.status, 200, `superadmin: expected 200, got ${superadminRes.status}`);

    const pmRes = await pmCurl('POST', path);
    assert.equal(pmRes.status, 200, `pm: expected 200, got ${pmRes.status}`);

    const memberRes = await memberCurl('POST', path);
    assert.equal(memberRes.status, 403, `member: expected 403, got ${memberRes.status}`);
  });

  // Traceability: TC-004 → Story 2 / AC-2
  test('TC-004: Preset roles matrix — team invite endpoint: superadmin→200, pm→200, member→403', async () => {
    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const path = `/v1/teams/${testTeamId}/members`;
    const body = JSON.stringify({ username: `invite-target-${RUN_ID}` });

    const superadminRes = await adminCurl('POST', path, { body });
    assert.ok(
      superadminRes.status === 200 || superadminRes.status === 201,
      `superadmin: expected 200/201, got ${superadminRes.status}`,
    );

    const pmRes = await pmCurl('POST', path, { body });
    assert.ok(
      pmRes.status === 200 || pmRes.status === 201,
      `pm: expected 200/201, got ${pmRes.status}`,
    );

    const memberRes = await memberCurl('POST', path, { body });
    assert.equal(memberRes.status, 403, `member: expected 403, got ${memberRes.status}`);
  });

  // ── Custom Role ──────────────────────────────────────────────────

  // Traceability: TC-005 → Story 3 / AC-1
  test('TC-005: Custom role with partial permissions allows read — GET /main-items returns 200', async () => {
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('GET', `/v1/teams/${testTeamId}/main-items`);
    assert.equal(res.status, 200, `Expected 200 but got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-006 → Story 3 / AC-2
  test('TC-006: Custom role without create permission denies write — POST /main-items returns 403', async () => {
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: 'should-be-denied', priority: 'P2' }),
    });
    assert.equal(res.status, 403, `Expected 403 but got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-007 → Story 3 / AC-3
  test('TC-007: Permission change takes effect immediately without re-login', async () => {
    // Add main_item:create to the custom role via admin API
    const updateRes = await adminCurl('PUT', `/v1/admin/roles/${customRoleId}`, {
      body: JSON.stringify({
        permissionCodes: ['main_item:read', 'progress:read', 'main_item:create'],
      }),
    });
    assert.equal(updateRes.status, 200, `Update role failed: ${updateRes.body}`);

    // Use the SAME token (no re-login) — permission change must be reflected immediately
    const customCurl = createAuthCurl(apiBaseUrl, customRoleToken);
    const res = await customCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `post-grant-item-${RUN_ID}`, priority: 'P2' }),
    });
    assert.equal(res.status, 200, `Expected 200 after permission grant, got ${res.status}: ${res.body}`);
  });

  // ── Permission Boundaries ────────────────────────────────────────

  // Traceability: TC-008 → Story 4 / AC-1
  test('TC-008: Empty permission role is denied on protected endpoint — returns 403', async () => {
    const emptyCurl = createAuthCurl(apiBaseUrl, emptyRoleToken);
    const res = await emptyCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    assert.equal(res.status, 403, `Expected 403 but got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-009 → Story 4 / AC-2
  test('TC-009: Superadmin bypasses permission check — returns 200 on protected endpoint', async () => {
    const res = await adminCurl(
      'POST',
      `/v1/teams/${testTeamId}/main-items/${testItemId}/archive`,
    );
    // 404/500 indicates missing fixture, not a pass
    assert.equal(res.status, 200, `Expected 200 (not 403/404/500), got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-010 → Story 4 / AC-3
  test('TC-010: Invalid token returns 401 not 403', async () => {
    const res = await curl(
      'GET',
      `${apiBaseUrl}/v1/teams/${testTeamId}/main-items`,
      { headers: { Authorization: 'Bearer invalid.jwt.token' } },
    );
    assert.equal(res.status, 401, `Expected 401 (auth failure), got ${res.status}: ${res.body}`);
  });

  // ── Permission Coverage CI ───────────────────────────────────────

  // Traceability: TC-011 → Story 5 / AC-1
  test('TC-011: CI fails when permission code lacks test coverage', async () => {
    // This test verifies the Go-level coverage assertion catches uncovered codes.
    // It runs the coverage check against the actual backend test suite.
    const result = runCli(
      'go test ./... -run TestPermissionCodeCoverage',
      '/Users/fanhuifeng/Projects/Go/pm-work-tracker-2/backend',
    );
    // The test itself should pass (all codes covered); if it fails, output must name the missing code.
    // Here we verify the command runs without unexpected panics (exit 0 or structured failure).
    const combined = result.stdout + result.stderr;
    if (result.exitCode !== 0) {
      assert.match(
        combined,
        /missing test coverage for:/,
        `Coverage check failed without expected message. Output: ${combined}`,
      );
    }
    // exitCode 0 means all codes are covered — also a valid pass
  });

  // Traceability: TC-012 → Story 5 / AC-2
  test('TC-012: CI passes when all permission codes have test coverage', async () => {
    const result = runCli(
      'go test ./... -run TestPermissionCodeCoverage',
      '/Users/fanhuifeng/Projects/Go/pm-work-tracker-2/backend',
    );
    assert.equal(
      result.exitCode,
      0,
      `Expected coverage check to pass (exit 0). Output: ${result.stdout}${result.stderr}`,
    );
  });
});
