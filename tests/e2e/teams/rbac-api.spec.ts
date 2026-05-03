import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds, randomCode, setupRbacFixtures } from '../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;
let pmToken: string;
let memberToken: string;
let pmUserBizKey: string;
let testTeamId: string;
const RUN_ID = Date.now();

test.describe('RBAC — Teams (TC-003, TC-004, TC-034, TC-035)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    pmUserBizKey = f.pmUserBizKey;
    testTeamId = f.teamBizKey;
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);
  });

  // ── Archive Permission ───────────────────────────────────────────

  // Traceability: TC-003 → Story 2 / AC-1
  test('TC-003: archive endpoint: superadmin→200, pm→200, member→403', async () => {
    // Create and complete an item for archive testing
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-archive-${RUN_ID}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    const itemData = JSON.parse(itemRes.body).data;
    const itemId = String(itemData?.bizKey ?? itemData?.id);

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

  // ── Team Invite Permission ────────────────────────────────────────

  // Traceability: TC-004 → Story 2 / AC-2
  test('TC-004: team invite endpoint: superadmin→200, pm→200, member→403', async () => {
    const pmCurl = createAuthCurl(apiBaseUrl, pmToken);
    const memberCurl = createAuthCurl(apiBaseUrl, memberToken);

    // Create a fresh user to invite
    const newUserRes = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `invite-tc004-${RUN_ID}`, displayName: 'Invite TC004' }),
    });
    const newUserData = JSON.parse(newUserRes.body).data;
    const newUserBizKey = String(newUserData?.bizKey ?? newUserData?.id);

    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    const rolesBody = JSON.parse(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesBody.data?.items ?? rolesBody.data ?? rolesBody;
    const memberRole = roles.find((r) => r.roleName === 'member');
    const memberRoleKey = memberRole?.bizKey ?? '';

    const path = `/v1/teams/${testTeamId}/members`;
    const body = JSON.stringify({ username: `invite-tc004-${RUN_ID}`, roleKey: memberRoleKey });

    const superadminRes = await adminCurl('POST', path, { body });
    expect(superadminRes.status === 200 || superadminRes.status === 201).toBeTruthy();

    // PM should also be able to invite
    const newUser2Res = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `invite-tc004b-${RUN_ID}`, displayName: 'Invite TC004b' }),
    });
    if (newUser2Res.status === 200 || newUser2Res.status === 201) {
      const pmRes = await pmCurl('POST', path, {
        body: JSON.stringify({ username: `invite-tc004b-${RUN_ID}`, roleKey: memberRoleKey }),
      });
      expect(pmRes.status === 200 || pmRes.status === 201).toBeTruthy();
    }

    // Member should NOT be able to invite
    const memberRes = await memberCurl('POST', path, {
      body: JSON.stringify({ username: `invite-tc004c-${RUN_ID}`, roleKey: memberRoleKey }),
    });
    expect(memberRes.status).toBe(403);
  });

  // ── Team Invite & Create ──────────────────────────────────────────

  // Traceability: TC-034 → Story 2 / AC-2
  test('TC-034: 邀请用户加入团队并分配角色', async () => {
    const noPermsRes = await adminCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username: `noperms-tc034-${RUN_ID}`, displayName: 'NoPerms TC034' }),
    });
    expect(noPermsRes.status === 200 || noPermsRes.status === 201).toBeTruthy();

    const rolesRes = await adminCurl('GET', '/v1/admin/roles');
    const rolesBody = JSON.parse(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesBody.data?.items ?? rolesBody.data ?? rolesBody;
    const memberRole = roles.find((r) => r.roleName === 'member');
    const memberRoleKey = memberRole?.bizKey ?? '';

    const res = await adminCurl('POST', `/v1/teams/${testTeamId}/members`, {
      body: JSON.stringify({ username: `noperms-tc034-${RUN_ID}`, roleKey: memberRoleKey }),
      headers: { Authorization: `Bearer ${pmToken}` },
    });
    expect(res.status).toBe(200);
  });

  // Traceability: TC-035 → Story 5 / AC-1
  test('TC-035: 拥有 team:create 权限创建团队成功', async () => {
    const res = await adminCurl('POST', '/v1/teams', {
      body: JSON.stringify({ name: `PM 创建的团队 ${Date.now()}`, code: randomCode() }),
      headers: { Authorization: `Bearer ${pmToken}` },
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });
});
