import { test, expect } from '@playwright/test';
import {
  curl,
  apiBaseUrl,
  getApiToken,
  createAuthCurl,
  randomCode,
  parseApiBody,
  extractBizKey,
  createTestTeam,
  authHeader,
} from '../../helpers.js';

// ── Fact Table (verified from source) ───────────────────────────────
// API_PREFIX: /v1                    (router.go:85)
// MAP_CREATE: POST /v1/teams/:teamId/milestone-maps   (router.go:158)
// MAP_CREATE_REQ: {mapName, mapDesc?}                  (dto/milestone_dto.go:5)
// MAP_VO: {bizKey, teamKey, mapName, mapDesc, mapStatus, statusName, milestoneCount, itemCount, overallProgress, createTime, dbUpdateTime}
// MS_CREATE: POST /v1/teams/:teamId/milestone-maps/:mapId/milestones (router.go:167)
// MS_CREATE_REQ: {milestoneName, expectedEndDate}      (dto/milestone_dto.go:22)
// MS_VO: {bizKey, teamKey, milestoneMapKey, milestoneName, expectedEndDate, milestoneStatus, statusName, completion, itemCount, createTime, dbUpdateTime}
// STATUS_CHANGE_REQ: {status: string}                   (dto/item_dto.go:110)
// MAP_TRANSITIONS: planning->[reviewed], reviewed->[ready,planning], ready->[executing,reviewed], executing->[completed,ready]
// MS_TRANSITIONS: not_started->[in_progress,cancelled], in_progress->[completed,cancelled], completed->[cancelled]

// ── Helpers ─────────────────────────────────────────────────────────

async function createMilestoneMap(
  authCurl: ReturnType<typeof createAuthCurl>,
  teamBizKey: string,
  name: string,
  description?: string,
): Promise<string> {
  const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
    body: JSON.stringify({ mapName: name, mapDesc: description ?? '' }),
  });
  expect(res.status).toBe(201);
  const data = parseApiBody(res.body);
  const bizKey = extractBizKey(data);
  expect(bizKey).toBeTruthy();
  return bizKey!;
}

async function createMilestone(
  authCurl: ReturnType<typeof createAuthCurl>,
  teamBizKey: string,
  mapBizKey: string,
  name: string,
  expectedEndDate: string,
): Promise<string> {
  const res = await authCurl(
    'POST',
    `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
    {
      body: JSON.stringify({ milestoneName: name, expectedEndDate }),
    },
  );
  expect(res.status).toBe(201);
  const data = parseApiBody(res.body);
  const bizKey = extractBizKey(data);
  expect(bizKey).toBeTruthy();
  return bizKey!;
}

// ── Test Suite ──────────────────────────────────────────────────────

test.describe('Milestone Map API', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let token: string;

  test.beforeAll(async () => {
    try {
      token = await getApiToken(apiBaseUrl);
      if (!token) throw new Error('Token is undefined after getApiToken');
      authCurl = createAuthCurl(apiBaseUrl, token);
    } catch (e) {
      console.error('beforeAll: auth failed', e);
      throw e;
    }

    try {
      teamBizKey = await createTestTeam(token, `e2e-msmap-${randomCode()}`);
      if (!teamBizKey) throw new Error('teamBizKey is undefined after createTestTeam');
    } catch (e) {
      console.error('beforeAll: team creation failed', e);
      throw e;
    }
  });

  // ── MilestoneMap CRUD ───────────────────────────────────────────

  // Traceability: TC-053 → Story 1 / AC-1, PRD Spec Related Changes #1
  test('TC-053: API Create milestone map', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: 'Test Map', mapDesc: 'desc' }),
    });
    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.mapName).toBe('Test Map');
    expect(data.mapDesc).toBe('desc');
    expect(data.mapStatus).toBe('planning');
    expect(data.bizKey).toBeTruthy();
  });

  // Traceability: TC-054 → Story 1 / AC-2, Story 1 / AC-3
  test('TC-054: API Create milestone map validation errors', async () => {
    // Empty name
    const res1 = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: '', mapDesc: '' }),
    });
    expect(res1.status).toBe(400);

    // Name exceeding 100 chars
    const longName = 'x'.repeat(101);
    const res2 = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: longName, mapDesc: '' }),
    });
    expect(res2.status).toBe(400);
  });

  // Traceability: TC-055 → Story 8 / AC-1, PRD Spec Related Changes #1
  test('TC-055: API List milestone maps', async () => {
    // Create 3 maps to ensure list has items
    await createMilestoneMap(authCurl, teamBizKey, 'List Test Map 1');
    await createMilestoneMap(authCurl, teamBizKey, 'List Test Map 2');
    await createMilestoneMap(authCurl, teamBizKey, 'List Test Map 3');

    const res = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps`);
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items = data.items ?? data;
    expect(Array.isArray(items)).toBeTruthy();
    expect(items.length).toBeGreaterThanOrEqual(3);
    const first = items[0];
    expect(first.mapName).toBeTruthy();
    expect(first.mapStatus).toBeTruthy();
    expect(first.bizKey).toBeTruthy();
  });

  // Traceability: TC-056 → PRD Spec Related Changes #1
  test('TC-056: API Get milestone map by ID', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Get Test Map');

    const res = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.bizKey).toBe(mapBizKey);
    expect(data.mapName).toBe('Get Test Map');
    expect(data.mapStatus).toBe('planning');
  });

  // Traceability: TC-057 → Story 2 / AC-1
  test('TC-057: API Update milestone map', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Original Name');

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({ mapName: 'Updated Name', mapDesc: 'Updated desc' }),
    });
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapName).toBe('Updated Name');
    expect(data.mapDesc).toBe('Updated desc');
  });

  // Traceability: TC-058 → PRD Spec Related Changes #1
  test('TC-058: API Delete milestone map', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Delete Test Map');

    const delRes = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    expect(delRes.status).toBe(200);

    // Subsequent GET returns 404
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    expect(getRes.status).toBe(404);
  });

  // Traceability: TC-059 → Story 3 / AC-1, Story 3 / AC-2, Story 3 / AC-3
  test('TC-059: API Change milestone map status', async () => {
    // Valid transition: planning -> reviewed
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Status Test Map');
    const res1 = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`,
      { body: JSON.stringify({ status: 'reviewed' }) },
    );
    expect(res1.status).toBe(200);
    const data1 = parseApiBody(res1.body);
    expect(data1.mapStatus).toBe('reviewed');

    // Invalid transition: planning -> completed (skip intermediate)
    const mapBizKey2 = await createMilestoneMap(authCurl, teamBizKey, 'Invalid Transition Map');
    const res2 = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey2}/status`,
      { body: JSON.stringify({ status: 'completed' }) },
    );
    expect(res2.status).toBe(400);

    // Completed map: no available transitions (verified via transitions endpoint below)
  });

  // Traceability: TC-060 → Story 3 / AC-1, Story 3 / AC-2, Story 3 / AC-3
  test('TC-060: API Get available transitions for milestone map', async () => {
    // Planning map -> should return ["reviewed"]
    const planningMapKey = await createMilestoneMap(authCurl, teamBizKey, 'Transitions Planning');
    const res1 = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/available-transitions`,
    );
    expect(res1.status).toBe(200);
    const data1 = parseApiBody(res1.body);
    const transitions1 = data1.transitions ?? data1;
    expect(transitions1).toEqual(['reviewed']);

    // Completed map -> should return empty array
    // Transition: planning -> reviewed -> ready -> executing -> completed
    let res = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/status`,
      { body: JSON.stringify({ status: 'reviewed' }) },
    );
    expect(res.status).toBe(200);
    res = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/status`,
      { body: JSON.stringify({ status: 'ready' }) },
    );
    expect(res.status).toBe(200);
    res = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/status`,
      { body: JSON.stringify({ status: 'executing' }) },
    );
    expect(res.status).toBe(200);
    res = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/status`,
      { body: JSON.stringify({ status: 'completed' }) },
    );
    expect(res.status).toBe(200);

    const res2 = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestone-maps/${planningMapKey}/available-transitions`,
    );
    expect(res2.status).toBe(200);
    const data2 = parseApiBody(res2.body);
    const transitions2 = data2.transitions ?? data2;
    expect(transitions2).toEqual([]);
  });

  // ── Milestone CRUD ──────────────────────────────────────────────

  // Traceability: TC-061 → Story 4a / AC-1
  test('TC-061: API Create milestone', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Milestone Parent Map');

    const res = await authCurl(
      'POST',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
      { body: JSON.stringify({ milestoneName: 'Phase 1', expectedEndDate: '2026-06-30' }) },
    );
    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.milestoneName).toBe('Phase 1');
    expect(data.milestoneStatus).toBe('not_started');
    expect(data.bizKey).toBeTruthy();
  });

  // Traceability: TC-062 → Story 4a / AC-2, Story 4a / AC-3
  test('TC-062: API Create milestone validation errors', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Validation Map');

    // Empty name
    const res1 = await authCurl(
      'POST',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
      { body: JSON.stringify({ milestoneName: '', expectedEndDate: '2026-06-30' }) },
    );
    expect(res1.status).toBe(400);

    // Name exceeding 100 chars
    const longName = 'x'.repeat(101);
    const res2 = await authCurl(
      'POST',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
      { body: JSON.stringify({ milestoneName: longName, expectedEndDate: '2026-06-30' }) },
    );
    expect(res2.status).toBe(400);
  });

  // Traceability: TC-063 → PRD Spec Related Changes #2
  test('TC-063: API List milestones by map', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'List By Map');
    await createMilestone(authCurl, teamBizKey, mapBizKey, 'MS 1', '2026-07-01');
    await createMilestone(authCurl, teamBizKey, mapBizKey, 'MS 2', '2026-08-01');

    const res = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
    );
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items = data.items ?? data;
    expect(Array.isArray(items)).toBeTruthy();
    expect(items.length).toBeGreaterThanOrEqual(2);
  });

  // Traceability: TC-064 → PRD Spec Related Changes #2, Story 6 / AC-3
  test('TC-064: API List milestones by team', async () => {
    const map1 = await createMilestoneMap(authCurl, teamBizKey, 'Team Map 1');
    const map2 = await createMilestoneMap(authCurl, teamBizKey, 'Team Map 2');
    await createMilestone(authCurl, teamBizKey, map1, 'Team MS 1', '2026-07-01');
    await createMilestone(authCurl, teamBizKey, map2, 'Team MS 2', '2026-08-01');

    const res = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones`);
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items = data.items ?? data;
    expect(Array.isArray(items)).toBeTruthy();
    expect(items.length).toBeGreaterThanOrEqual(2);
  });

  // Traceability: TC-065 → PRD Spec Related Changes #2
  test('TC-065: API Get milestone by ID', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Get MS Map');
    const msBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Get Test MS', '2026-09-01');

    const res = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.bizKey).toBe(msBizKey);
    expect(data.milestoneName).toBe('Get Test MS');
    expect(data.milestoneStatus).toBe('not_started');
  });

  // Traceability: TC-066 → Story 4b / AC-1
  test('TC-066: API Update milestone', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Update MS Map');
    const msBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Original MS', '2026-09-01');

    // Need dbUpdateTime for optimistic concurrency
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    const existing = parseApiBody(getRes.body);

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`, {
      body: JSON.stringify({
        milestoneName: 'Updated Phase',
        expectedEndDate: '2026-07-31',
        dbUpdateTime: existing.dbUpdateTime,
      }),
    });
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneName).toBe('Updated Phase');
  });

  // Traceability: TC-067 → Story 4c / AC-1
  test('TC-067: API Delete milestone unbinds associated MIs', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Delete MS Map');
    const msBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Delete Target MS', '2026-10-01');

    // Create an MI bound to this milestone
    const miRes = await authCurl('POST', `/v1/teams/${teamBizKey}/main-items`, {
      body: JSON.stringify({
        title: 'Bound MI for delete test',
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
        milestoneKey: msBizKey,
      }),
    });
    expect(miRes.status).toBe(201);
    const miData = parseApiBody(miRes.body);
    const miBizKey = extractBizKey(miData);

    // Delete the milestone
    const delRes = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    expect(delRes.status).toBe(200);

    // Verify MI's milestone_key is cleared
    const miGetRes = await authCurl('GET', `/v1/teams/${teamBizKey}/main-items/${miBizKey}`);
    expect(miGetRes.status).toBe(200);
    const miUpdated = parseApiBody(miGetRes.body);
    expect(miUpdated.milestoneKey).toBeNull();
  });

  // Traceability: TC-068 → Story 5 / AC-1, Story 5 / AC-2, Story 5 / AC-3, Story 5 / AC-4
  test('TC-068: API Change milestone status', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'MS Status Map');

    // Valid transition: not_started -> in_progress
    const ms1BizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Status MS 1', '2026-11-01');
    const res1 = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${ms1BizKey}/status`,
      { body: JSON.stringify({ status: 'in_progress' }) },
    );
    expect(res1.status).toBe(200);
    const data1 = parseApiBody(res1.body);
    expect(data1.milestoneStatus).toBe('in_progress');

    // Invalid transition: cancelled -> any (cancelled is terminal, no outgoing transitions)
    // First transition to cancelled
    const ms2BizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Status MS 2', '2026-11-15');
    await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`,
      { body: JSON.stringify({ status: 'cancelled' }) },
    );
    const res2 = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`,
      { body: JSON.stringify({ status: 'in_progress' }) },
    );
    expect(res2.status).toBe(400);

    // Cancellation auto-unbinds MIs
    const ms3BizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Status MS 3', '2026-12-01');
    // Transition to in_progress first (not_started -> in_progress)
    await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${ms3BizKey}/status`,
      { body: JSON.stringify({ status: 'in_progress' }) },
    );
    // Create an MI bound to this milestone
    const miRes = await authCurl('POST', `/v1/teams/${teamBizKey}/main-items`, {
      body: JSON.stringify({
        title: 'Bound MI for cancel test',
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
        milestoneKey: ms3BizKey,
      }),
    });
    expect(miRes.status).toBe(201);
    const miData = parseApiBody(miRes.body);
    const miBizKey = extractBizKey(miData);

    // Cancel the milestone (in_progress -> cancelled)
    const res3 = await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${ms3BizKey}/status`,
      { body: JSON.stringify({ status: 'cancelled' }) },
    );
    expect(res3.status).toBe(200);

    // Verify MI is unbound
    const miGetRes = await authCurl('GET', `/v1/teams/${teamBizKey}/main-items/${miBizKey}`);
    expect(miGetRes.status).toBe(200);
    const miUpdated = parseApiBody(miGetRes.body);
    expect(miUpdated.milestoneKey).toBeNull();
  });

  // Traceability: TC-069 → Story 5 / AC-2, Story 5 / AC-3
  test('TC-069: API Get available transitions for milestone', async () => {
    const mapBizKey = await createMilestoneMap(authCurl, teamBizKey, 'Transitions MS Map');

    // not_started -> ["in_progress", "cancelled"]
    const nsBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'NS MS', '2026-07-01');
    const res1 = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestones/${nsBizKey}/available-transitions`,
    );
    expect(res1.status).toBe(200);
    const data1 = parseApiBody(res1.body);
    const trans1 = data1.transitions ?? data1;
    expect(trans1.sort()).toEqual(['cancelled', 'in_progress']);

    // completed -> ["cancelled"]
    const compBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Comp MS', '2026-08-01');
    await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${compBizKey}/status`,
      { body: JSON.stringify({ status: 'in_progress' }) },
    );
    await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${compBizKey}/status`,
      { body: JSON.stringify({ status: 'completed' }) },
    );
    const res2 = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestones/${compBizKey}/available-transitions`,
    );
    expect(res2.status).toBe(200);
    const data2 = parseApiBody(res2.body);
    const trans2 = data2.transitions ?? data2;
    expect(trans2).toEqual(['cancelled']);

    // cancelled -> [] (terminal)
    const canBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, 'Can MS', '2026-09-01');
    await authCurl(
      'PUT',
      `/v1/teams/${teamBizKey}/milestones/${canBizKey}/status`,
      { body: JSON.stringify({ status: 'cancelled' }) },
    );
    const res3 = await authCurl(
      'GET',
      `/v1/teams/${teamBizKey}/milestones/${canBizKey}/available-transitions`,
    );
    expect(res3.status).toBe(200);
    const data3 = parseApiBody(res3.body);
    const trans3 = data3.transitions ?? data3;
    expect(trans3).toEqual([]);
  });

  // ── Permission Tests ────────────────────────────────────────────

  // Traceability: TC-070 → PRD Spec Security Requirements
  test('TC-070: API milestone operations without permission return 403', async () => {
    // Create a team with a member who does NOT have milestone permissions
    // Use the member role which lacks milestone:xxx permissions
    const superToken = await getApiToken(apiBaseUrl);
    const superCurl = createAuthCurl(apiBaseUrl, superToken);

    // Create a separate team and add a member
    const permTeamKey = await createTestTeam(superToken, `e2e-perm-${randomCode()}`);

    // Create a user with member role (no milestone permissions)
    const runId = Date.now();
    const username = `e2e-noperm-${runId}`;
    const userRes = await superCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({ username, displayName: 'NoPerms User' }),
    });
    expect(userRes.status).toBe(200);
    const userData = parseApiBody(userRes.body);
    const userBizKey = extractBizKey(userData)!;

    // Get member role key
    const rolesRes = await superCurl('GET', '/v1/admin/roles');
    const rolesData = parseApiBody(rolesRes.body);
    const roles = rolesData.items ?? rolesData;
    const memberRole = roles.find((r: any) => r.roleName === 'member');
    expect(memberRole).toBeTruthy();

    // Add user to team with member role
    await superCurl('POST', `/v1/teams/${permTeamKey}/members`, {
      body: JSON.stringify({ username, roleKey: memberRole.bizKey }),
    });

    // Login as the no-permission user
    const noPermToken = await getApiToken(apiBaseUrl, {
      username,
      password: userData.initialPassword,
    });
    const noPermCurl = createAuthCurl(apiBaseUrl, noPermToken);

    // POST create without milestone:create
    const createRes = await noPermCurl('POST', `/v1/teams/${permTeamKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: 'NoPerm Map', mapDesc: '' }),
    });
    expect(createRes.status).toBe(403);

    // GET without milestone:read (member role typically has read, test with explicit denial if applicable)
    // Note: member role may have read permission, so test update/delete which they definitely lack
    const mapBizKey = await createMilestoneMap(superCurl, permTeamKey, 'Existing Map');

    // PUT update without milestone:update
    const updateRes = await noPermCurl('PUT', `/v1/teams/${permTeamKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({ mapName: 'Hacked Name' }),
    });
    expect(updateRes.status).toBe(403);

    // DELETE without milestone:delete
    const delRes = await noPermCurl('DELETE', `/v1/teams/${permTeamKey}/milestone-maps/${mapBizKey}`);
    expect(delRes.status).toBe(403);
  });
});
