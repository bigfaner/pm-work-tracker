import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds } from '../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;
let pmToken: string;
let memberToken: string;
let pmUserBizKey: string;

// Fixture IDs (bizKeys)
let testTeamId: string;

// Unique suffix to avoid collisions across runs
const RUN_ID = Date.now();

test.describe('Permission — Team Invite (TC-003, TC-004)', () => {
  test.beforeAll(async () => {
    // ── 1. Admin (superadmin) auth ──────────────────────────────────
    superadminToken = await getApiToken(apiBaseUrl, defaultCreds);
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // ── 2. Create test team ─────────────────────────────────────────
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const code = Array.from({ length: 5 }, () => letters[Math.floor(Math.random() * 26)]).join('');
    const teamRes = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `e2e-team-inv-${RUN_ID}`, code }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    const teamData = JSON.parse(teamRes.body).data;
    testTeamId = String(teamData?.bizKey ?? teamData?.id ?? JSON.parse(teamRes.body).id);

    // ── 3. Fetch preset role bizKeys ────────────────────────────────────
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

    // ── 4. Create test users and get tokens ───────────────────────
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

    const pmResult = await createUserAndGetToken(`pm-inv-${RUN_ID}`, 'PM Invite User');
    pmToken = pmResult.token;
    pmUserBizKey = pmResult.bizKey;
    memberToken = (await createUserAndGetToken(`member-inv-${RUN_ID}`, 'Member Invite User')).token;

    // ── 5. Assign users to the team ─────────────────────────
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `pm-inv-${RUN_ID}`, roleKey: memberRoleKey }),
    }).catch(() => {});
    await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `member-inv-${RUN_ID}`, roleKey: memberRoleKey }),
    }).catch(() => {});

    // Transfer PM role to pm user
    const transferRes = await adminCurl('PUT', `/v1/teams/${testTeamId}/pm`, {
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
    expect(transferRes.status === 200 || transferRes.status === 204).toBeTruthy();
  });

  // ── Preset Roles Matrix ──────────────────────────────────────────

  // Traceability: TC-003 → Story 2 / AC-1
  test('TC-003: Preset roles matrix — archive endpoint: superadmin→200, pm→200, member→403', async () => {
    // Need a completed item to archive
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-tc003-${RUN_ID}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    const itemData = JSON.parse(itemRes.body).data;
    const itemId = String(itemData?.bizKey ?? itemData?.id ?? JSON.parse(itemRes.body).id);

    for (const status of ['progressing', 'reviewing', 'completed']) {
      await adminCurl('PUT', `/v1/teams/${testTeamId}/main-items/${itemId}/status`, {
        body: JSON.stringify({ status }),
      });
    }

    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);
    const path = `/v1/teams/${testTeamId}/main-items/${itemId}/archive`;

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
});
