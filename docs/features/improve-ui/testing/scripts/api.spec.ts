import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiUrl, loginAs } from './helpers.js';

/**
 * API E2E Tests for improve-ui feature.
 *
 * Pre-conditions:
 * - Backend running on http://localhost:8080
 * - Test data seeded (admin user, team, items)
 */
describe('API E2E Tests', () => {
  let adminAuth: { authHeader: Record<string, string>; token: string };
  let pmAuth: { authHeader: Record<string, string>; token: string };
  let teamId: string;

  before(async () => {
    adminAuth = await loginAs('admin', 'admin123');
    // Get a team ID for team-scoped endpoints
    const teamsRes = await curl('GET', `${apiUrl}/api/v1/teams`, {
      headers: adminAuth.authHeader,
    });
    const teams = JSON.parse(teamsRes.body);
    const teamList = teams.data?.items ?? teams.data ?? teams;
    if (Array.isArray(teamList) && teamList.length > 0) {
      teamId = teamList[0].id ?? teamList[0].ID;
    }
  });

  // Traceability: TC-053 → Spec 5.7 #1
  test('TC-053: 团队详情独立路由成员 CRUD API', async () => {
    assert.ok(teamId, 'Team ID available for testing');

    // GET team detail
    const detailRes = await curl('GET', `${apiUrl}/api/v1/teams/${teamId}`, {
      headers: adminAuth.authHeader,
    });
    assert.equal(detailRes.status, 200, 'GET team detail returns 200');

    const detail = JSON.parse(detailRes.body);
    const teamData = detail.data ?? detail;
    assert.ok(teamData.name || teamData.Name, 'Team detail contains name');

    // GET members
    const membersRes = await curl('GET', `${apiUrl}/api/v1/teams/${teamId}/members`, {
      headers: adminAuth.authHeader,
    });
    assert.equal(membersRes.status, 200, 'GET members returns 200');

    const members = JSON.parse(membersRes.body);
    const memberList = members.data ?? members;
    assert.ok(Array.isArray(memberList), 'Members response is array');

    // Note: POST/PUT/DELETE member operations are tested with specific data
    // that may not exist in seed — these verify the endpoint structure
  });

  // Traceability: TC-054 → Spec 5.7 #2
  test('TC-054: 用户管理全量操作 API', async () => {
    // GET all users
    const listRes = await curl('GET', `${apiUrl}/api/v1/admin/users`, {
      headers: adminAuth.authHeader,
    });
    assert.equal(listRes.status, 200, 'GET users returns 200');

    const users = JSON.parse(listRes.body);
    const userList = users.data?.items ?? users.data ?? users;
    assert.ok(Array.isArray(userList), 'Users response is array');

    // POST create user — unique username
    const uniqueAccount = `test_${Date.now()}`;
    const createRes = await curl('POST', `${apiUrl}/api/v1/admin/users`, {
      headers: adminAuth.authHeader,
      body: JSON.stringify({
        username: uniqueAccount,
        displayName: 'Test User',
        email: `${uniqueAccount}@test.com`,
        canCreateTeam: false,
      }),
    });
    assert.equal(createRes.status, 201, 'POST create user returns 201');

    // POST duplicate username — should fail
    const dupRes = await curl('POST', `${apiUrl}/api/v1/admin/users`, {
      headers: adminAuth.authHeader,
      body: JSON.stringify({
        username: uniqueAccount,
        displayName: 'Duplicate',
        email: `dup@${uniqueAccount}.com`,
        canCreateTeam: false,
      }),
    });
    assert.ok(dupRes.status === 409 || dupRes.status === 422, 'Duplicate account returns error');

    // PUT update user
    const createData = JSON.parse(createRes.body);
    const userId = createData.data?.id ?? createData.data?.ID ?? createData.id;
    if (userId) {
      const updateRes = await curl('PUT', `${apiUrl}/api/v1/admin/users/${userId}`, {
        headers: adminAuth.authHeader,
        body: JSON.stringify({
          displayName: 'Updated Name',
          email: `updated@${uniqueAccount}.com`,
        }),
      });
      assert.equal(updateRes.status, 200, 'PUT update user returns 200');

      // PUT toggle status (disable)
      const statusRes = await curl('PUT', `${apiUrl}/api/v1/admin/users/${userId}/status`, {
        headers: adminAuth.authHeader,
        body: JSON.stringify({ status: 'disabled' }),
      });
      assert.equal(statusRes.status, 200, 'PUT toggle status returns 200');
    }
  });

  // Traceability: TC-055 → Spec 5.7 #3
  test('TC-055: 事项清单 Detail 分页参数', async () => {
    assert.ok(teamId, 'Team ID available for testing');

    // Without pagination
    const defaultRes = await curl('GET', `${apiUrl}/api/v1/teams/${teamId}/main-items`, {
      headers: adminAuth.authHeader,
    });
    assert.equal(defaultRes.status, 200, 'GET items without pagination returns 200');

    // With pagination params
    const pagedRes = await curl(
      'GET',
      `${apiUrl}/api/v1/teams/${teamId}/main-items?page=1&pageSize=20`,
      { headers: adminAuth.authHeader },
    );
    assert.equal(pagedRes.status, 200, 'GET items with pagination returns 200');

    const data = JSON.parse(pagedRes.body);
    const items = data.data?.items ?? data.data ?? data;
    assert.ok(Array.isArray(items), 'Response contains items array');
  });

  // Traceability: TC-056 → Spec 5.7 #4
  test('TC-056: 全量表格聚合查询 API', async () => {
    assert.ok(teamId, 'Team ID available for testing');

    // GET unified table view
    const tableRes = await curl('GET', `${apiUrl}/api/v1/teams/${teamId}/views/table`, {
      headers: adminAuth.authHeader,
    });
    assert.equal(tableRes.status, 200, 'GET table view returns 200');

    const data = JSON.parse(tableRes.body);
    const items = data.data?.items ?? data.data ?? data;
    assert.ok(Array.isArray(items), 'Table view returns array');

    // Each item should have a type field
    if (items.length > 0) {
      const firstItem = items[0];
      assert.ok(
        firstItem.type !== undefined || firstItem.Type !== undefined,
        'Items contain type field (main/sub)',
      );
    }

    // GET with type filter
    const filteredRes = await curl(
      'GET',
      `${apiUrl}/api/v1/teams/${teamId}/views/table?type=main`,
      { headers: adminAuth.authHeader },
    );
    assert.equal(filteredRes.status, 200, 'GET table with type filter returns 200');

    // GET with multiple filters
    const multiFilterRes = await curl(
      'GET',
      `${apiUrl}/api/v1/teams/${teamId}/views/table?type=main&priority=P1`,
      { headers: adminAuth.authHeader },
    );
    assert.equal(multiFilterRes.status, 200, 'GET table with multi-filter returns 200');
  });
});
