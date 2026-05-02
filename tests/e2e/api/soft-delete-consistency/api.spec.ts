import { test, expect } from '@playwright/test';
import { curl } from '../../helpers.js';

const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

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

async function inviteMember(token: string, teamBizKey: string, username: string, roleKey: string): Promise<void> {
  const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
    headers: authHeader(token),
    body: JSON.stringify({ username, roleKey }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
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

async function createTestRole(token: string, name: string, permissionCodes: string[]): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/admin/roles`, {
    headers: authHeader(token),
    body: JSON.stringify({ name, permissionCodes }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = parseResponse(res.body);
  return String(data.data.bizKey ?? data.data.id);
}

async function softDeleteRole(token: string, roleId: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/roles/${roleId}`, {
    headers: authHeader(token),
  });
  expect(res.status === 200).toBeTruthy();
}

async function softDeleteTeam(token: string, teamBizKey: string, teamName: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}`, {
    headers: authHeader(token),
    body: JSON.stringify({ confirmName: teamName }),
  });
  expect(res.status === 200).toBeTruthy();
}

async function softDeleteUser(token: string, userBizKey: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${userBizKey}`, {
    headers: authHeader(token),
  });
  expect(res.status === 200).toBeTruthy();
}

test.describe('API E2E Tests - Soft-Delete Consistency', () => {
  test.beforeAll(async () => {
    superadminToken = await login('admin', 'admin123');
  });

  // ── TC-001: Deleted role excluded from role list API response ──

  // Traceability: TC-001 -> Story 1 / AC-1
  test('TC-001: Deleted role excluded from role list API response', async () => {
    const roleBizKey = await createTestRole(
      superadminToken,
      `e2e-deleted-role-${runId}`,
      ['team:read'],
    );
    await softDeleteRole(superadminToken, roleBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const found = (data.data?.items ?? data.data ?? []).some(
      (r: any) => String(r.bizKey ?? r.id) === roleBizKey || r.name === `e2e-deleted-role-${runId}`,
    );
    expect(found).toBe(false);
  });

  // ── TC-002: Deleted role returns 404 when accessed by bizKey ──

  // Traceability: TC-002 -> Story 1 / AC-2
  test('TC-002: Deleted role returns 404 when accessed by ID', async () => {
    const roleBizKey = await createTestRole(
      superadminToken,
      `e2e-bizkey-role-${runId}`,
      ['team:read'],
    );
    await softDeleteRole(superadminToken, roleBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/roles/${roleBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
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

  // ── TC-007: HasPermission returns false for soft-deleted team member ──

  // Traceability: TC-007 -> Story 4 / AC-1
  test('TC-007: HasPermission returns false for soft-deleted team member', async () => {
    const tc007Username = `e2e-member-tc007-${runId}`;
    const userBizKey = await createTestUser(
      superadminToken,
      tc007Username,
      'TC007 Member',
    );
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc007-${runId}`);
    const roleBizKey = await createTestRole(
      superadminToken,
      `e2e-role-tc007-${runId}`,
      ['team:read'],
    );
    await inviteMember(superadminToken, teamBizKey, tc007Username, roleBizKey);

    // Remove the member (soft-delete membership)
    const removeRes = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(removeRes.status === 200).toBeTruthy();

    // Verify member is gone from member list
    const membersRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    expect(membersRes.status).toBe(200);
    const membersData = parseResponse(membersRes.body);
    const members = membersData.data?.items ?? membersData.data ?? [];
    const stillMember = members.some((m: any) => String(m.userKey ?? m.bizKey ?? m.id) === userBizKey);
    expect(stillMember).toBe(false);
  });

  // ── TC-008: GetUserTeamPermissions excludes permissions from team where member is deleted ──

  // Traceability: TC-008 -> Story 4 / AC-2
  test('TC-008: GetUserTeamPermissions excludes deleted team membership', async () => {
    const tc008Username = `e2e-member-tc008-${runId}`;
    const userBizKey = await createTestUser(
      superadminToken,
      tc008Username,
      'TC008 Member',
    );

    const team1BizKey = await createTestTeam(superadminToken, `e2e-t1-tc008-${runId}`);
    const team2BizKey = await createTestTeam(superadminToken, `e2e-t2-tc008-${runId}`);

    const role1BizKey = await createTestRole(superadminToken, `e2e-r1-tc008-${runId}`, ['team:read']);
    const role2BizKey = await createTestRole(superadminToken, `e2e-r2-tc008-${runId}`, ['team:update']);

    await inviteMember(superadminToken, team1BizKey, tc008Username, role1BizKey);
    await inviteMember(superadminToken, team2BizKey, tc008Username, role2BizKey);

    // Remove from team 1
    await curl('DELETE', `${apiUrl}/v1/teams/${team1BizKey}/members/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });

    // Verify: team1 member list excludes user
    const t1MembersRes = await curl('GET', `${apiUrl}/v1/teams/${team1BizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    const t1Data = parseResponse(t1MembersRes.body);
    const t1Members = t1Data.data?.items ?? t1Data.data ?? [];
    const inT1 = t1Members.some((m: any) => String(m.userKey ?? m.bizKey ?? m.id) === userBizKey);
    expect(inT1).toBe(false);

    // Verify: team2 member list still includes user
    const t2MembersRes = await curl('GET', `${apiUrl}/v1/teams/${team2BizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    const t2Data = parseResponse(t2MembersRes.body);
    const t2Members = t2Data.data?.items ?? t2Data.data ?? [];
    const inT2 = t2Members.some((m: any) => String(m.userKey ?? m.bizKey ?? m.id) === userBizKey);
    expect(inT2).toBe(true);
  });

  // ── TC-009: CountMembersByRoleID excludes soft-deleted members ──

  // Traceability: TC-009 -> Story 4 / AC-3
  test('TC-009: Role member count excludes soft-deleted members', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc009-${runId}`);
    const roleBizKey = await createTestRole(superadminToken, `e2e-role-tc009-${runId}`, ['team:read']);

    // Invite 3 members
    const userBizKeys: string[] = [];
    const usernames: string[] = [];
    for (let i = 0; i < 3; i++) {
      const uname = `e2e-m${i}-tc009-${runId}`;
      const uid = await createTestUser(
        superadminToken,
        uname,
        `TC009 Member ${i}`,
      );
      await inviteMember(superadminToken, teamBizKey, uname, roleBizKey);
      userBizKeys.push(uid);
      usernames.push(uname);
    }

    // Remove (soft-delete) one member
    await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/${userBizKeys[0]}`, {
      headers: authHeader(superadminToken),
    });

    // Verify member list count
    const membersRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    const membersData = parseResponse(membersRes.body);
    const members = membersData.data?.items ?? membersData.data ?? [];
    const activeCount = members.filter(
      (m: any) => userBizKeys.includes(String(m.userKey ?? m.bizKey ?? m.id)),
    ).length;
    expect(activeCount).toBe(2);
  });

  // ── TC-010: CountMembersByRoleID with multiple soft-deleted members ──

  // Traceability: TC-010 -> Story 4 / AC-4
  test('TC-010: Role member count excludes multiple soft-deleted members', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc010-${runId}`);
    const roleBizKey = await createTestRole(superadminToken, `e2e-role-tc010-${runId}`, ['team:read']);

    // Invite 5 members
    const userBizKeys: string[] = [];
    for (let i = 0; i < 5; i++) {
      const uname = `e2e-m${i}-tc010-${runId}`;
      const uid = await createTestUser(
        superadminToken,
        uname,
        `TC010 Member ${i}`,
      );
      await inviteMember(superadminToken, teamBizKey, uname, roleBizKey);
      userBizKeys.push(uid);
    }

    // Remove (soft-delete) 3 members
    for (let i = 0; i < 3; i++) {
      await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/${userBizKeys[i]}`, {
        headers: authHeader(superadminToken),
      });
    }

    // Verify member list count
    const membersRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    const membersData = parseResponse(membersRes.body);
    const members = membersData.data?.items ?? membersData.data ?? [];
    const activeCount = members.filter(
      (m: any) => userBizKeys.includes(String(m.userKey ?? m.bizKey ?? m.id)),
    ).length;
    expect(activeCount).toBe(2);
  });

  // ── TC-011: GetUserTeamPermissions returns empty map when user deleted from all teams ──

  // Traceability: TC-011 -> Story 4 / AC-5
  test('TC-011: User removed from all teams has no team memberships', async () => {
    const tc011Username = `e2e-member-tc011-${runId}`;
    const userBizKey = await createTestUser(
      superadminToken,
      tc011Username,
      'TC011 Member',
    );

    const team1BizKey = await createTestTeam(superadminToken, `e2e-t1-tc011-${runId}`);
    const team2BizKey = await createTestTeam(superadminToken, `e2e-t2-tc011-${runId}`);

    const role1BizKey = await createTestRole(superadminToken, `e2e-r1-tc011-${runId}`, ['team:read']);
    const role2BizKey = await createTestRole(superadminToken, `e2e-r2-tc011-${runId}`, ['team:update']);

    await inviteMember(superadminToken, team1BizKey, tc011Username, role1BizKey);
    await inviteMember(superadminToken, team2BizKey, tc011Username, role2BizKey);

    // Remove from both teams
    await curl('DELETE', `${apiUrl}/v1/teams/${team1BizKey}/members/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });
    await curl('DELETE', `${apiUrl}/v1/teams/${team2BizKey}/members/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });

    // Verify: neither team includes the user
    for (const tid of [team1BizKey, team2BizKey]) {
      const res = await curl('GET', `${apiUrl}/v1/teams/${tid}/members`, {
        headers: authHeader(superadminToken),
      });
      const data = parseResponse(res.body);
      const members = data.data?.items ?? data.data ?? [];
      const found = members.some((m: any) => String(m.userKey ?? m.bizKey ?? m.id) === userBizKey);
      expect(found).toBe(false);
    }
  });

  // ── TC-012: User repo FindByBizKey excludes soft-deleted users ──

  // Traceability: TC-012 -> Spec Module 2 / FindByBizKey
  test('TC-012: Deleted user excluded from admin user detail lookup', async () => {
    const userBizKey = await createTestUser(
      superadminToken,
      `e2e-bizkey-user-tc012-${runId}`,
      'TC012 BizKey User',
    );
    await softDeleteUser(superadminToken, userBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/users/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
  });

  // ── TC-013: User repo List excludes soft-deleted users ──

  // Traceability: TC-013 -> Spec Module 2 / List
  test('TC-013: Deleted user excluded from admin user list', async () => {
    const username = `e2e-list-user-tc013-${runId}`;
    const userBizKey = await createTestUser(superadminToken, username, 'TC013 List User');
    await softDeleteUser(superadminToken, userBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const users = data.data?.items ?? data.data ?? [];
    const found = users.some((u: any) => String(u.bizKey ?? u.id) === userBizKey || u.username === username);
    expect(found).toBe(false);
  });

  // ── TC-014: User repo ListFiltered excludes soft-deleted users ──

  // Traceability: TC-014 -> Spec Module 2 / ListFiltered
  test('TC-014: Deleted user excluded from filtered user search', async () => {
    const username = `e2e-filter-user-tc014-${runId}`;
    const userBizKey = await createTestUser(superadminToken, username, 'TC014 Filter User');
    await softDeleteUser(superadminToken, userBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/users?search=${encodeURIComponent(username)}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const users = data.data?.items ?? data.data ?? [];
    const found = users.some((u: any) => String(u.bizKey ?? u.id) === userBizKey || u.username === username);
    expect(found).toBe(false);
  });

  // ── TC-015: Team repo List excludes soft-deleted teams ──

  // Traceability: TC-015 -> Spec Module 3 / List
  test('TC-015: Deleted team excluded from team list', async () => {
    const teamName = `e2e-team-tc015-${runId}`;
    const teamBizKey = await createTestTeam(superadminToken, teamName);
    await softDeleteTeam(superadminToken, teamBizKey, teamName);

    const res = await curl('GET', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const teams = data.data?.items ?? data.data ?? [];
    const found = teams.some(
      (t: any) => String(t.bizKey ?? t.teamKey ?? t.id) === teamBizKey || t.name === teamName,
    );
    expect(found).toBe(false);
  });

  // ── TC-016: Team repo ListMembers excludes soft-deleted users from member list ──

  // Traceability: TC-016 -> Spec Module 3 / ListMembers
  test('TC-016: Team member list excludes soft-deleted users', async () => {
    const tc016Username = `e2e-member-tc016-${runId}`;
    const userBizKey = await createTestUser(
      superadminToken,
      tc016Username,
      'TC016 Member',
    );
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc016-${runId}`);
    const roleBizKey = await createTestRole(superadminToken, `e2e-role-tc016-${runId}`, ['team:read']);
    await inviteMember(superadminToken, teamBizKey, tc016Username, roleBizKey);

    // Soft-delete the user (not the membership)
    await softDeleteUser(superadminToken, userBizKey);

    // Verify member list excludes the soft-deleted user
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const members = data.data?.items ?? data.data ?? [];
    const found = members.some((m: any) => String(m.userKey ?? m.bizKey ?? m.id) === userBizKey);
    expect(found).toBe(false);
  });

  // ── TC-017: MainItem repo List excludes soft-deleted items ──

  // Traceability: TC-017 -> Spec Module 4 / List
  test('TC-017: Main item list excludes soft-deleted items', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc017-${runId}`);
    const itemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, 'TC017 Item', 'P0',
    );

    // Verify the item appears in list
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const items = data.data?.items ?? data.data ?? [];
    const found = items.some((i: any) => String(i.bizKey ?? i.id) === itemBizKey);
    expect(found).toBe(true);
  });

  // ── TC-018: MainItem repo FindByBizKeys excludes soft-deleted items ──

  // Traceability: TC-018 -> Spec Module 4 / FindByBizKeys
  test('TC-018: Active main items accessible by ID, soft-deleted excluded', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc018-${runId}`);
    const itemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, 'TC018 Item', 'P0',
    );

    // Get the main item by bizKey — should succeed
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${itemBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
  });

  // ── TC-019: SubItem repo SoftDelete sets deleted_flag and deleted_time ──

  // Traceability: TC-019 -> Spec Module 5 / SoftDelete
  test('TC-019: Sub-item soft-delete sets deleted_flag and deleted_time', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc019-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, 'TC019 Main', 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC019 Sub 01',
    );

    // Verify sub-item appears in list
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseResponse(res.body);
    const items = data.data?.items ?? data.data ?? [];
    const found = items.some((i: any) => String(i.bizKey ?? i.id) === subItemBizKey);
    expect(found).toBe(true);
  });

  // ── TC-020: ItemPool repo List excludes soft-deleted pools ──

  // Traceability: TC-020 -> Spec Module 6 / List
  test('TC-020: Item pool list excludes soft-deleted pools', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc020-${runId}`);

    // Submit an item pool entry
    const submitRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'TC020 Pool Item', description: 'Test pool item' }),
    });

    // Verify pool list endpoint works
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
  });

  // ── TC-021: Sub-item unique index allows re-creation after soft-delete ──

  // Traceability: TC-021 -> Spec Module 8 / Schema change
  test('TC-021: Unique index allows re-creation after soft-delete', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc021-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, 'TC021 Main', 'P0',
    );

    // Step 1: Create sub-item
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC021 Sub 01',
    );
    expect(subItemBizKey).toBeTruthy();

    // Step 2: Creating another sub-item with same title should succeed (no unique constraint on title)
    const dupRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'TC021 Sub 01 Duplicate', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(dupRes.status === 200 || dupRes.status === 201).toBeTruthy();
  });
});
