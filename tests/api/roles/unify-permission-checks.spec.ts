import { test, expect, describe, beforeAll } from 'vitest';
import {
  curl,
  apiBaseUrl,
  getApiToken,
  authHeader,
  parseApiBody,
  extractBizKey,
  randomCode,
  setupRbacFixtures,
} from '../../shared/helpers.js';

const apiUrl = apiBaseUrl;
const runId = Date.now();

// ── Shared state ────────────────────────────────────────────────────
let superadminToken: string;
let pmToken: string;
let memberToken: string;
let customToken: string;
let noPermsToken: string;
let teamBizKey: string;
let team2BizKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
let itemPoolBizKey: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let customUserBizKey: string;
let noPermsUserBizKey: string;
let memberRoleKey: string;
let customRoleKey: string;
let noPermsRoleKey: string;

const parseData = parseApiBody;

describe('Unify Permission Checks — API Tests (TC-004..TC-039)', () => {
  beforeAll(async () => {
    const f = await setupRbacFixtures({ noPerms: true });
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    noPermsToken = f.noPermsToken!;
    pmUserBizKey = f.pmUserBizKey;
    memberUserBizKey = f.memberUserBizKey;
    noPermsUserBizKey = f.noPermsUserBizKey!;
    teamBizKey = f.teamBizKey;
    memberRoleKey = f.memberRoleKey;

    // Create team B for cross-team tests (TC-022, TC-023)
    const team2Res = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `CrossTeam B ${runId}`, code: randomCode() }),
    });
    expect(team2Res.status === 200 || team2Res.status === 201).toBeTruthy();
    team2BizKey = extractBizKey(parseData(team2Res.body))!;

    // Create main item in team A
    const mainItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'Main Item PermTest', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(mainItemRes.status === 200 || mainItemRes.status === 201).toBeTruthy();
    mainItemBizKey = extractBizKey(parseData(mainItemRes.body))!;

    // Create sub-item assigned to PM user (not the custom-role user)
    const subItemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'Sub Item PermTest', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(subItemRes.status === 200 || subItemRes.status === 201).toBeTruthy();
    subItemBizKey = extractBizKey(parseData(subItemRes.body))!;

    // Create item pool entry for review test (TC-027)
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'Pool Entry PermTest', description: 'test' }),
    });
    expect(poolRes.status === 200 || poolRes.status === 201).toBeTruthy();
    itemPoolBizKey = extractBizKey(parseData(poolRes.body))!;

    // Create a custom role with sub_item:update + sub_item:change_status + progress:create
    const customRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `custom-perm-${runId}`,
        permissionCodes: ['sub_item:update', 'sub_item:change_status', 'progress:create', 'sub_item:read', 'main_item:read'],
      }),
    });
    expect(customRoleRes.status === 200 || customRoleRes.status === 201).toBeTruthy();
    customRoleKey = extractBizKey(parseData(customRoleRes.body))!;

    // Create a custom-role user and add to team
    const customUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-perm-${runId}`, displayName: 'Custom Perm User' }),
    });
    expect(customUserRes.status === 200 || customUserRes.status === 201).toBeTruthy();
    const customUserData = parseData(customUserRes.body);
    customUserBizKey = extractBizKey(customUserData)!;
    customToken = await getApiToken(apiBaseUrl, { username: `custom-perm-${runId}`, password: customUserData.initialPassword });

    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `custom-perm-${runId}`, roleKey: customRoleKey }),
    });

    // Create a no-perms role (sub_item:view only) for negative tests
    const noPermsRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        name: `noperm-role-${runId}`,
        permissionCodes: ['sub_item:read', 'main_item:read'],
      }),
    });
    expect(noPermsRoleRes.status === 200 || noPermsRoleRes.status === 201).toBeTruthy();
    noPermsRoleKey = extractBizKey(parseData(noPermsRoleRes.body))!;

    // Change noPerms user's role to noPerms role (user already in team via setupRbacFixtures)
    await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/members/${noPermsUserBizKey}/role`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ roleKey: noPermsRoleKey }),
    });
  });

  // ── Story 1: Custom role with sub_item:update ──────────────────────

  // Traceability: TC-004 → Story 1 / AC-1
  test('TC-004: Custom role with sub_item:update edits non-assigned sub-item', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(customToken),
      body: JSON.stringify({ title: 'updated-title' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.title).toBe('updated-title');
  });

  // Traceability: TC-005 → Story 1 / AC-2
  test('TC-005: Custom role without sub_item:update gets 403 on edit', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ title: 'hack' }),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('ERR_FORBIDDEN');
  });

  // ── Story 2: Custom role with sub_item:change_status ───────────────

  // Traceability: TC-006 → Story 2 / AC-1
  test('TC-006: Custom role with sub_item:change_status changes non-assigned sub-item status', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/status`, {
      headers: authHeader(customToken),
      body: JSON.stringify({ status: 'progressing' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.subItem.itemStatus).toBe('progressing');
  });

  // Traceability: TC-007 → Story 2 / AC-2
  test('TC-007: Custom role without sub_item:change_status gets 403 on status change', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/status`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ status: 'closed' }),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('ERR_FORBIDDEN');
  });

  // ── Story 3: SuperAdmin team management (AC 3a) ────────────────────

  // Traceability: TC-008 → Story 3 / AC 3a
  test('TC-008: SuperAdmin creates team (201)', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `tc008-team-${runId}`, code: randomCode() }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = parseData(res.body);
    expect(data.name).toContain('tc008-team');
  });

  // Traceability: TC-009 → Story 3 / AC 3a
  test('TC-009: SuperAdmin updates team (200)', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: 'updated-team-tc009' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.name).toBe('updated-team-tc009');
  });

  // Traceability: TC-010 → Story 3 / AC 3a
  test('TC-010: SuperAdmin invites member (200)', async () => {
    const newUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc010-user-${runId}`, displayName: 'TC010 Invitee' }),
    });
    expect(newUserRes.status === 200 || newUserRes.status === 201).toBeTruthy();
    const newUserData = parseData(newUserRes.body);
    const newUserBizKey = extractBizKey(newUserData)!;

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc010-user-${runId}`, roleKey: memberRoleKey }),
    });
    expect(res.status).toBe(200);
  });

  // Traceability: TC-011 → Story 3 / AC 3a
  test('TC-011: SuperAdmin modifies member role (200)', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/members/${memberUserBizKey}/role`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ roleKey: memberRoleKey }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });

  // Traceability: TC-012 → Story 3 / AC 3a
  test('TC-012: SuperAdmin removes member (200)', async () => {
    // Create a removable user
    const tmpUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc012-user-${runId}`, displayName: 'TC012 Remove' }),
    });
    expect(tmpUserRes.status === 200 || tmpUserRes.status === 201).toBeTruthy();
    const tmpUserData = parseData(tmpUserRes.body);
    const tmpUserBizKey = extractBizKey(tmpUserData)!;

    // Add to team first
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc012-user-${runId}`, roleKey: memberRoleKey }),
    });

    // Remove
    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/${tmpUserBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
  });

  // Traceability: TC-013 → Story 3 / AC 3a
  test('TC-013: SuperAdmin transfers PM (200)', async () => {
    // Create a fresh throwaway team for PM transfer to avoid state pollution
    const tmpTeamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `tc013-transfer-${runId}`, code: randomCode() }),
    });
    expect(tmpTeamRes.status === 200 || tmpTeamRes.status === 201).toBeTruthy();
    const tmpTeamKey = extractBizKey(parseData(tmpTeamRes.body))!;

    // Create fresh users for the transfer test
    const tmpPmRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc013-pm-${runId}`, displayName: 'TC013 PM' }),
    });
    expect(tmpPmRes.status === 200 || tmpPmRes.status === 201).toBeTruthy();
    const tmpPmData = parseData(tmpPmRes.body);
    const tmpPmKey = extractBizKey(tmpPmData)!;

    const tmpMemberRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc013-member-${runId}`, displayName: 'TC013 Member' }),
    });
    expect(tmpMemberRes.status === 200 || tmpMemberRes.status === 201).toBeTruthy();
    const tmpMemberData = parseData(tmpMemberRes.body);
    const tmpMemberKey = extractBizKey(tmpMemberData)!;

    // Add both users to the throwaway team
    await curl('POST', `${apiUrl}/v1/teams/${tmpTeamKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc013-pm-${runId}`, roleKey: memberRoleKey }),
    });
    await curl('POST', `${apiUrl}/v1/teams/${tmpTeamKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc013-member-${runId}`, roleKey: memberRoleKey }),
    });

    // Transfer PM to the pm user first
    const transferToPm = await curl('PUT', `${apiUrl}/v1/teams/${tmpTeamKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: tmpPmKey }),
    });
    expect(transferToPm.status === 200 || transferToPm.status === 204).toBeTruthy();

    // Transfer PM to member user
    const transferToMember = await curl('PUT', `${apiUrl}/v1/teams/${tmpTeamKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: tmpMemberKey }),
    });
    expect(transferToMember.status === 200 || transferToMember.status === 204).toBeTruthy();
  });

  // Traceability: TC-014 → Story 3 / AC 3a
  test('TC-014: SuperAdmin disbands team (200)', async () => {
    // Create a throwaway team
    const tmpTeamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `tc014-team-${runId}`, code: randomCode() }),
    });
    const tmpTeamKey = extractBizKey(parseData(tmpTeamRes.body))!;
    const tmpTeamData = parseData(tmpTeamRes.body);

    const res = await curl('DELETE', `${apiUrl}/v1/teams/${tmpTeamKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ confirmName: tmpTeamData.name }),
    });
    expect(res.status).toBe(200);

    // Verify team is gone
    const getRes = await curl('GET', `${apiUrl}/v1/teams/${tmpTeamKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(getRes.status === 404 || getRes.status === 500).toBeTruthy();
  });

  // ── Story 3: SuperAdmin item management (AC 3b) ────────────────────

  // Traceability: TC-015 → Story 3 / AC 3b
  test('TC-015: SuperAdmin creates main item (201)', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'tc015-test-item', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = parseData(res.body);
    expect(data.title).toBe('tc015-test-item');
  });

  // Traceability: TC-016 → Story 3 / AC 3b
  test('TC-016: SuperAdmin edits main item (200)', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'edited-item-tc016' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.title).toBe('edited-item-tc016');
  });

  // Traceability: TC-017 → Story 3 / AC 3b
  test('TC-017: SuperAdmin archives main item (200)', async () => {
    // Create a fresh main item and transition to completed for archive
    const itemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: `tc017-archive-${runId}`, priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    const itemKey = extractBizKey(parseData(itemRes.body))!;

    for (const status of ['progressing', 'reviewing', 'completed']) {
      await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${itemKey}/status`, {
        headers: authHeader(superadminToken),
        body: JSON.stringify({ status }),
      });
    }

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${itemKey}/archive`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    // Archive endpoint returns data: null on success
  });

  // Traceability: TC-018 → Story 3 / AC 3b
  test('TC-018: SuperAdmin creates sub-item (201)', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'tc018-sub-1', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = parseData(res.body);
    expect(data.title).toBe('tc018-sub-1');
  });

  // Traceability: TC-019 → Story 3 / AC 3b
  test('TC-019: SuperAdmin edits sub-item (200)', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'edited-sub-tc019' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.title).toBe('edited-sub-tc019');
  });

  // Traceability: TC-020 → Story 3 / AC 3b
  test('TC-020: SuperAdmin assigns sub-item (200)', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/assignee`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ assigneeKey: memberUserBizKey }),
    });
    expect(res.status).toBe(200);
  });

  // Traceability: TC-021 → Story 3 / AC 3b
  test('TC-021: SuperAdmin changes sub-item status (200)', async () => {
    // Create a fresh sub-item to avoid state pollution from TC-006
    const freshSubRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'tc021-status-sub', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(freshSubRes.status === 200 || freshSubRes.status === 201).toBeTruthy();
    const freshSubKey = extractBizKey(parseData(freshSubRes.body))!;

    // Transition: pending -> progressing -> blocking
    await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${freshSubKey}/status`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ status: 'progressing' }),
    });

    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${freshSubKey}/status`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ status: 'blocking' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.subItem.itemStatus).toBe('blocking');
  });

  // ── Story 4: Cross-team access ─────────────────────────────────────

  // Traceability: TC-022 → Story 4 / AC-1
  test('TC-022: SuperAdmin accesses non-member team resources (200)', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${team2BizKey}/main-items`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(Array.isArray(data.items ?? data)).toBeTruthy();
  });

  // Traceability: TC-023 → Story 4 / AC-2
  test('TC-023: Non-member user without SuperAdmin gets 403 on cross-team access', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${team2BizKey}/main-items`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('NOT_TEAM_MEMBER');
  });

  // ── Story 6: PM team management via permission codes ───────────────

  // Traceability: TC-024 → Story 6 / AC-1
  test('TC-024: PM team management operations succeed via permission codes', async () => {
    // Create a fresh throwaway team for PM management tests to avoid state pollution
    const tmpTeamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `tc024-pm-mgmt-${runId}`, code: randomCode() }),
    });
    expect(tmpTeamRes.status === 200 || tmpTeamRes.status === 201).toBeTruthy();
    const tmpTeamKey = extractBizKey(parseData(tmpTeamRes.body))!;

    // Create a fresh PM user for this team
    const tmpPmRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc024-pm-${runId}`, displayName: 'TC024 PM' }),
    });
    expect(tmpPmRes.status === 200 || tmpPmRes.status === 201).toBeTruthy();
    const tmpPmData = parseData(tmpPmRes.body);
    const tmpPmKey = extractBizKey(tmpPmData)!;
    const tmpPmToken = await getApiToken(apiBaseUrl, { username: `tc024-pm-${runId}`, password: tmpPmData.initialPassword });

    // Add PM user to team and transfer PM role
    await curl('POST', `${apiUrl}/v1/teams/${tmpTeamKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc024-pm-${runId}`, roleKey: memberRoleKey }),
    });
    await curl('PUT', `${apiUrl}/v1/teams/${tmpTeamKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: tmpPmKey }),
    });

    // Create a fresh user for invite test
    const newUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `tc024-newuser-${runId}`, displayName: 'TC024 NewUser' }),
    });
    expect(newUserRes.status === 200 || newUserRes.status === 201).toBeTruthy();
    const newUserData = parseData(newUserRes.body);
    const newUserBizKey = extractBizKey(newUserData)!;

    // PM invites member
    const inviteRes = await curl('POST', `${apiUrl}/v1/teams/${tmpTeamKey}/members`, {
      headers: authHeader(tmpPmToken),
      body: JSON.stringify({ username: `tc024-newuser-${runId}`, roleKey: memberRoleKey }),
    });
    expect(inviteRes.status === 200 || inviteRes.status === 201).toBeTruthy();

    // PM removes member
    const removeRes = await curl('DELETE', `${apiUrl}/v1/teams/${tmpTeamKey}/members/${newUserBizKey}`, {
      headers: authHeader(tmpPmToken),
    });
    expect(removeRes.status).toBe(200);
  });

  // ── Story 7: Progress permission ────────────────────────────────────

  // Traceability: TC-025 → Story 7 / AC-1
  test('TC-025: Custom role with progress:create adds progress to non-assigned sub-item', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(customToken),
      body: JSON.stringify({ completion: 50, achievement: 'test progress' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = parseData(res.body);
    expect(data.completion).toBe(50);
  });

  // Traceability: TC-033 → Story 7 / AC-2
  test('TC-033: Custom role without progress:create gets 403 on progress add', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ completion: 10, achievement: 'unauthorized' }),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('ERR_FORBIDDEN');
  });

  // ── Story 3: SuperAdmin item pool and views (AC 3c) ────────────────

  // Traceability: TC-026 → Story 3 / AC 3c
  test('TC-026: SuperAdmin submits item pool entry (201)', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'todo-entry-tc026', description: 'test' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = parseData(res.body);
    expect(data.title).toBe('todo-entry-tc026');
  });

  // Traceability: TC-027 → Story 3 / AC 3c
  test('TC-027: SuperAdmin reviews (rejects) item pool entry (200)', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${itemPoolBizKey}/reject`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ reason: 'not needed' }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.poolStatus).toBe('rejected');
  });

  // Traceability: TC-028 → Story 3 / AC 3c
  test('TC-028: SuperAdmin views weekly report (200)', async () => {
    // Use a known Monday as weekStart (2026-05-04 is a Monday)
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/weekly?weekStart=2026-05-04`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data).toBeDefined();
  });

  // Traceability: TC-029 → Story 3 / AC 3c
  test('TC-029: SuperAdmin views gantt chart (200)', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/gantt`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data).toBeDefined();
  });

  // Traceability: TC-030 → Story 3 / AC 3c
  test('TC-030: SuperAdmin views table view (200)', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/table`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data).toBeDefined();
  });

  // Traceability: TC-031 → Story 3 / AC 3c
  test('TC-031: SuperAdmin exports table report (200)', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/table/export`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    expect(res.headers['content-type']).toBeDefined();
  });

  // Traceability: TC-032 → Story 3 / AC 3c
  test('TC-032: SuperAdmin views user list (200)', async () => {
    const res = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    const users = data.items ?? data;
    expect(Array.isArray(users)).toBeTruthy();
    expect(users.length).toBeGreaterThanOrEqual(1);
  });

  // ── Edge cases ─────────────────────────────────────────────────────

  // Traceability: TC-034 → Story 3 / edge case
  test('TC-034: User with read-only role gets 403 on team-scoped write endpoint', async () => {
    // Create a custom role with only read permissions (no create/update)
    const readOnlyRoleRes = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ name: `readonly-role-${runId}`, permissionCodes: ['main_item:read'] }),
    });
    expect(readOnlyRoleRes.status === 200 || readOnlyRoleRes.status === 201).toBeTruthy();
    const readOnlyRoleKey = extractBizKey(parseData(readOnlyRoleRes.body))!;

    // Create a user with the read-only role
    const readOnlyUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `readonly-user-${runId}`, displayName: 'Read-Only User' }),
    });
    const readOnlyUserData = parseData(readOnlyUserRes.body);
    const readOnlyToken = await getApiToken(apiBaseUrl, { username: `readonly-user-${runId}`, password: readOnlyUserData.initialPassword });

    // Add user to team with read-only role
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `readonly-user-${runId}`, roleKey: readOnlyRoleKey }),
    });

    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(readOnlyToken),
      body: JSON.stringify({ title: 'edge-case', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('ERR_FORBIDDEN');
  });

  // Traceability: TC-035 → Story 3 / edge case
  test('TC-035: SuperAdmin requests non-existent team resource returns 404', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/999999/main-items`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('TEAM_NOT_FOUND');
  });

  // Traceability: TC-038 → Story 1 / boundary case
  test('TC-038: POST main-item with missing required fields returns 400', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: '' }),
    });
    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('VALIDATION_ERROR');
  });

  // Traceability: TC-039 → Story 5 / boundary case
  test('TC-039: PUT sub-item with leftover isSuperAdmin field in body is ignored', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(customToken),
      body: JSON.stringify({ title: 'updated-title-ignores-sa', isSuperAdmin: true }),
    });
    expect(res.status).toBe(200);
    const data = parseData(res.body);
    expect(data.title).toBe('updated-title-ignores-sa');
  });

  // ── Integration: Full middleware-handler chain ─────────────────────

  // Traceability: TC-036 → Story 4 / AC-1 + Story 3
  test('TC-036: Full chain — SuperAdmin cross-team access flows through middleware to handler', async () => {
    // Read: GET main-items from team2
    const readRes = await curl('GET', `${apiUrl}/v1/teams/${team2BizKey}/main-items`, {
      headers: authHeader(superadminToken),
    });
    expect(readRes.status).toBe(200);

    // Write: edit main item in team2 (create one first)
    const itemRes = await curl('POST', `${apiUrl}/v1/teams/${team2BizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'cross-team-item', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(itemRes.status === 200 || itemRes.status === 201).toBeTruthy();
    const crossItemKey = extractBizKey(parseData(itemRes.body))!;

    const editRes = await curl('PUT', `${apiUrl}/v1/teams/${team2BizKey}/main-items/${crossItemKey}`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ title: 'cross-team-edit' }),
    });
    expect(editRes.status).toBe(200);
    expect(parseData(editRes.body).title).toBe('cross-team-edit');

    // Create sub-item
    const subRes = await curl('POST', `${apiUrl}/v1/teams/${team2BizKey}/main-items/${crossItemKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: crossItemKey, title: 'cross-team-sub', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(subRes.status === 200 || subRes.status === 201).toBeTruthy();
  });

  // Traceability: TC-037 → Story 1 / AC-2 + Story 7 / AC-2
  test('TC-037: Full chain — Custom role permission denial propagates through middleware to 403', async () => {
    // Step 2: read succeeds (user has main_item:read)
    const readRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(noPermsToken),
    });
    expect(readRes.status).toBe(200);

    // Step 3: sub_item:update denied
    const updateRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ title: 'blocked' }),
    });
    expect(updateRes.status).toBe(403);
    expect(JSON.parse(updateRes.body).code).toBe('ERR_FORBIDDEN');

    // Step 4: progress:create denied
    const progressRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(noPermsToken),
      body: JSON.stringify({ completion: 10, achievement: 'blocked' }),
    });
    expect(progressRes.status).toBe(403);
    expect(JSON.parse(progressRes.body).code).toBe('ERR_FORBIDDEN');
  });
});
