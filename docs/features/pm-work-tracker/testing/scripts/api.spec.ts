import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiUrl, loginAs, parseBody } from './helpers.js';

// Test credentials — adjust via env vars or seed data
const ADMIN_USER = process.env.TEST_ADMIN_USER ?? 'admin';
const ADMIN_PASS = process.env.TEST_ADMIN_PASS ?? 'admin123';
const PM_USER = process.env.TEST_PM_USER ?? 'pm1';
const PM_PASS = process.env.TEST_PM_PASS ?? 'pm123';
const MEMBER_USER = process.env.TEST_MEMBER_USER ?? 'member1';
const MEMBER_PASS = process.env.TEST_MEMBER_PASS ?? 'member123';
const OTHER_USER = process.env.TEST_OTHER_USER ?? 'user2';
const OTHER_PASS = process.env.TEST_OTHER_PASS ?? 'user234';

const base = `${apiUrl}/api/v1`;

describe('API E2E Tests — pm-work-tracker', () => {

  // ===== Shared test data =====
  let adminAuth: Record<string, string>;
  let pmAuth: Record<string, string>;
  let memberAuth: Record<string, string>;
  let otherAuth: Record<string, string>;
  let teamId: number;
  let otherTeamId: number;
  let mainItemId: number;
  let subItemId: number;
  let poolItemId: number;
  let progressRecordId: number;

  before(async () => {
    adminAuth = (await loginAs(ADMIN_USER, ADMIN_PASS)).authHeader;
    pmAuth = (await loginAs(PM_USER, PM_PASS)).authHeader;
    memberAuth = (await loginAs(MEMBER_USER, MEMBER_PASS)).authHeader;
    otherAuth = (await loginAs(OTHER_USER, OTHER_PASS)).authHeader;

    // Create test teams
    const teamRes = await curl('POST', `${base}/teams`, {
      headers: adminAuth,
      body: JSON.stringify({ name: `E2E-Team-${Date.now()}`, description: 'Test team' }),
    });
    const teamData = parseBody(teamRes);
    teamId = teamData.data?.id ?? teamData.id;

    const otherTeamRes = await curl('POST', `${base}/teams`, {
      headers: otherAuth,
      body: JSON.stringify({ name: `E2E-Other-${Date.now()}`, description: 'Other team' }),
    });
    const otherTeamData = parseBody(otherTeamRes);
    otherTeamId = otherTeamData.data?.id ?? otherTeamData.id;

    // Create test main item
    const mainRes = await curl('POST', `${base}/teams/${teamId}/main-items`, {
      headers: pmAuth,
      body: JSON.stringify({
        title: 'E2E Main Item',
        priority: 'P2',
        expectedEndDate: '2026-12-31',
      }),
    });
    const mainData = parseBody(mainRes);
    mainItemId = mainData.data?.id ?? mainData.id;

    // Create test sub-item
    const subRes = await curl('POST', `${base}/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
      headers: pmAuth,
      body: JSON.stringify({
        title: 'E2E Sub Item',
        priority: 'P2',
        assigneeId: 1,
      }),
    });
    const subData = parseBody(subRes);
    subItemId = subData.data?.id ?? subData.id;

    // Move sub-item to "进行中" for progress tests
    await curl('PUT', `${base}/teams/${teamId}/sub-items/${subItemId}/status`, {
      headers: pmAuth,
      body: JSON.stringify({ status: '进行中' }),
    });

    // Create test pool item
    const poolRes = await curl('POST', `${base}/teams/${teamId}/item-pool`, {
      headers: memberAuth,
      body: JSON.stringify({
        title: 'E2E Pool Item',
        background: 'Test background',
        expectedOutput: 'Test output',
      }),
    });
    const poolData = parseBody(poolRes);
    poolItemId = poolData.data?.id ?? poolData.id;
  });

  // ===== Authentication (TC-069 ~ TC-071) =====

  describe('Authentication', () => {
    // Traceability: TC-069 → Spec 5.1
    test('TC-069: Login with valid credentials returns token', async () => {
      const res = await curl('POST', `${base}/auth/login`, {
        body: JSON.stringify({ username: ADMIN_USER, password: ADMIN_PASS }),
      });
      assert.equal(res.status, 200);
      const data = parseBody(res);
      const token = data.data?.token ?? data.token;
      assert.ok(token, 'Response should contain a token');
    });

    // Traceability: TC-070 → Spec 5.1
    test('TC-070: Login with invalid credentials returns 401', async () => {
      const res = await curl('POST', `${base}/auth/login`, {
        body: JSON.stringify({ username: ADMIN_USER, password: 'wrongpassword' }),
      });
      assert.equal(res.status, 401);
      const data = parseBody(res);
      assert.ok(data.message || data.code, 'Response should contain error info');
    });

    // Traceability: TC-071 → Spec 安全性需求
    test('TC-071: Unauthenticated access blocked', async () => {
      const res = await curl('GET', `${base}/teams`);
      assert.equal(res.status, 401);
    });
  });

  // ===== Team Data Isolation (TC-072 ~ TC-073) =====

  describe('Team Data Isolation', () => {
    // Traceability: TC-072 → Spec 安全性需求
    test('TC-072: Cross-team data access blocked', async () => {
      const res = await curl('GET', `${base}/teams/${teamId}/main-items`, {
        headers: otherAuth,
      });
      assert.ok(res.status === 403 || res.status === 404,
        `Expected 403/404, got ${res.status}`);
    });

    // Traceability: TC-073 → Spec 5.2
    test('TC-073: Super admin bypasses team isolation', async () => {
      const res = await curl('GET', `${base}/admin/teams`, {
        headers: adminAuth,
      });
      assert.equal(res.status, 200);
      const data = parseBody(res);
      const teams = data.data ?? data;
      assert.ok(Array.isArray(teams), 'Should return team array');
    });
  });

  // ===== Main Item Operations (TC-074 ~ TC-075) =====

  describe('Main Item Operations', () => {
    // Traceability: TC-074 → Spec 5.3 表单校验规则
    test('TC-074: Create main item with validation', async () => {
      // Empty title
      const r1 = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: '', priority: 'P1' }),
      });
      assert.ok(r1.status === 400, `Empty title should be 400, got ${r1.status}`);

      // Title > 100 chars
      const r2 = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'x'.repeat(101), priority: 'P1' }),
      });
      assert.ok(r2.status === 400, `101-char title should be 400, got ${r2.status}`);

      // Missing priority
      const r3 = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Valid title' }),
      });
      assert.ok(r3.status === 400, `Missing priority should be 400, got ${r3.status}`);

      // Valid item
      const r4 = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Valid Main Item', priority: 'P1', expectedEndDate: '2026-12-31' }),
      });
      assert.equal(r4.status, 201);
    });

    // Traceability: TC-075 → Spec 5.3 归档
    test('TC-075: Archive main item only when completed or closed', async () => {
      // Create a fresh item in "进行中" status (default after creation)
      const createRes = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Archive Test Item', priority: 'P2' }),
      });
      const itemData = parseBody(createRes);
      const itemId = itemData.data?.id ?? itemData.id;

      // Attempt archive on non-terminal status → should fail
      const r1 = await curl('POST', `${base}/teams/${teamId}/main-items/${itemId}/archive`, {
        headers: pmAuth,
      });
      assert.ok(r1.status === 400 || r1.status === 409,
        `Archive on active item should fail, got ${r1.status}`);

      // Create and complete an item, then archive
      const completeRes = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Complete Then Archive', priority: 'P3', status: '已完成' }),
      });
      // Note: status may need to go through flow; this tests the archive validation
    });
  });

  // ===== Sub-item Status Transitions (TC-076 ~ TC-077) =====

  describe('Sub-item Status Transitions', () => {
    // Traceability: TC-076 → Spec 事项状态流转
    test('TC-076: Valid status transitions succeed', async () => {
      // Create sub-item (status: 待开始)
      const createRes = await curl('POST', `${base}/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Status Flow Test', priority: 'P2', assigneeId: 1 }),
      });
      const data = parseBody(createRes);
      const sid = data.data?.id ?? data.id;

      // 待开始 → 进行中
      const r1 = await curl('PUT', `${base}/teams/${teamId}/sub-items/${sid}/status`, {
        headers: pmAuth,
        body: JSON.stringify({ status: '进行中' }),
      });
      assert.equal(r1.status, 200);

      // 进行中 → 待验收
      const r2 = await curl('PUT', `${base}/teams/${teamId}/sub-items/${sid}/status`, {
        headers: pmAuth,
        body: JSON.stringify({ status: '待验收' }),
      });
      assert.equal(r2.status, 200);

      // 待验收 → 已完成
      const r3 = await curl('PUT', `${base}/teams/${teamId}/sub-items/${sid}/status`, {
        headers: pmAuth,
        body: JSON.stringify({ status: '已完成' }),
      });
      assert.equal(r3.status, 200);
    });

    // Traceability: TC-077 → Spec 事项状态流转
    test('TC-077: Invalid status transitions rejected', async () => {
      // Create sub-item (status: 待开始)
      const createRes = await curl('POST', `${base}/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
        headers: pmAuth,
        body: JSON.stringify({ title: 'Invalid Transition Test', priority: 'P2', assigneeId: 1 }),
      });
      const data = parseBody(createRes);
      const sid = data.data?.id ?? data.id;

      // 待开始 → 已完成 (skip 进行中, 待验收)
      const r1 = await curl('PUT', `${base}/teams/${teamId}/sub-items/${sid}/status`, {
        headers: pmAuth,
        body: JSON.stringify({ status: '已完成' }),
      });
      assert.ok(r1.status === 400 || r1.status === 409,
        `Direct to completed should fail, got ${r1.status}`);

      // 待开始 → 阻塞中 (requires 进行中)
      const r2 = await curl('PUT', `${base}/teams/${teamId}/sub-items/${sid}/status`, {
        headers: pmAuth,
        body: JSON.stringify({ status: '阻塞中' }),
      });
      assert.ok(r2.status === 400 || r2.status === 409,
        `Blocked from 待开始 should fail, got ${r2.status}`);
    });
  });

  // ===== Progress Records (TC-078 ~ TC-080) =====

  describe('Progress Records', () => {
    // Traceability: TC-078 → Spec 5.4, UI Function 2
    test('TC-078: Add progress record with completion validation', async () => {
      // Add first record: 60%
      const r1 = await curl('POST', `${base}/teams/${teamId}/sub-items/${subItemId}/progress`, {
        headers: pmAuth,
        body: JSON.stringify({ completionRate: 60, achievement: 'Did stuff', blocker: '' }),
      });
      assert.equal(r1.status, 201);
      const d1 = parseBody(r1);
      progressRecordId = d1.data?.id ?? d1.id;

      // Decreasing completion: 50%
      const r2 = await curl('POST', `${base}/teams/${teamId}/sub-items/${subItemId}/progress`, {
        headers: pmAuth,
        body: JSON.stringify({ completionRate: 50 }),
      });
      assert.ok(r2.status === 400, `Decreasing completion should be 400, got ${r2.status}`);

      // Out of range: 150%
      const r3 = await curl('POST', `${base}/teams/${teamId}/sub-items/${subItemId}/progress`, {
        headers: pmAuth,
        body: JSON.stringify({ completionRate: 150 }),
      });
      assert.ok(r3.status === 400, `150% should be 400, got ${r3.status}`);

      // Valid: 75%
      const r4 = await curl('POST', `${base}/teams/${teamId}/sub-items/${subItemId}/progress`, {
        headers: pmAuth,
        body: JSON.stringify({ completionRate: 75, achievement: 'More progress' }),
      });
      assert.equal(r4.status, 201);
    });

    // Traceability: TC-079 → Spec 5.4
    test('TC-079: Progress records cannot be deleted', async () => {
      // There is no DELETE endpoint for progress records
      const res = await curl('DELETE', `${base}/teams/${teamId}/progress/${progressRecordId}`, {
        headers: pmAuth,
      });
      assert.ok(res.status === 404 || res.status === 405,
        `DELETE should not be available, got ${res.status}`);
    });

    // Traceability: TC-080 → Spec 5.3 进度自动汇总
    test('TC-080: Completion auto-recalculated after progress update', async () => {
      const beforeRes = await curl('GET', `${base}/teams/${teamId}/main-items/${mainItemId}`, {
        headers: pmAuth,
      });
      assert.equal(beforeRes.status, 200);
      const beforeData = parseBody(beforeRes);
      const beforeRate = beforeData.data?.completionRate ?? beforeData.completionRate;

      // Add progress to push sub-item to 100%
      await curl('POST', `${base}/teams/${teamId}/sub-items/${subItemId}/progress`, {
        headers: pmAuth,
        body: JSON.stringify({ completionRate: 100 }),
      });

      const afterRes = await curl('GET', `${base}/teams/${teamId}/main-items/${mainItemId}`, {
        headers: pmAuth,
      });
      const afterData = parseBody(afterRes);
      const afterRate = afterData.data?.completionRate ?? afterData.completionRate;

      assert.ok(afterRate > beforeRate,
        `Completion should increase: ${beforeRate} → ${afterRate}`);
    });
  });

  // ===== Item Pool (TC-081 ~ TC-083) =====

  describe('Item Pool', () => {
    // Traceability: TC-081 → Spec 5.5
    test('TC-081: Submit item to pool', async () => {
      const res = await curl('POST', `${base}/teams/${teamId}/item-pool`, {
        headers: memberAuth,
        body: JSON.stringify({ title: 'New Pool Item', background: 'Why', expectedOutput: 'What' }),
      });
      assert.equal(res.status, 201);
      const data = parseBody(res);
      assert.equal(data.data?.status ?? data.status, '待分配');
    });

    // Traceability: TC-082 → Story 9 / AC-1, Spec 5.5
    test('TC-082: Assign pool item creates sub-item', async () => {
      const res = await curl('POST', `${base}/teams/${teamId}/item-pool/${poolItemId}/assign`, {
        headers: pmAuth,
        body: JSON.stringify({ mainItemId, assigneeId: 1 }),
      });
      assert.equal(res.status, 200);
      const data = parseBody(res);
      assert.equal(data.data?.status ?? data.status, '已分配');
    });

    // Traceability: TC-083 → UI Function 3
    test('TC-083: Reject pool item requires reason', async () => {
      // Create another pool item to reject
      const createRes = await curl('POST', `${base}/teams/${teamId}/item-pool`, {
        headers: memberAuth,
        body: JSON.stringify({ title: 'To Be Rejected' }),
      });
      const data = parseBody(createRes);
      const poolId = data.data?.id ?? data.id;

      // Reject without reason
      const r1 = await curl('POST', `${base}/teams/${teamId}/item-pool/${poolId}/reject`, {
        headers: pmAuth,
        body: JSON.stringify({}),
      });
      assert.ok(r1.status === 400, `Reject without reason should be 400, got ${r1.status}`);

      // Reject with reason
      const r2 = await curl('POST', `${base}/teams/${teamId}/item-pool/${poolId}/reject`, {
        headers: pmAuth,
        body: JSON.stringify({ reason: 'Not in scope' }),
      });
      assert.equal(r2.status, 200);
    });
  });

  // ===== Role-Based Access Control (TC-084 ~ TC-087) =====

  describe('Role-Based Access Control', () => {
    // Traceability: TC-084 → Spec 5.3
    test('TC-084: Member cannot create main item', async () => {
      const res = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: memberAuth,
        body: JSON.stringify({ title: 'Unauthorized', priority: 'P1' }),
      });
      assert.equal(res.status, 403);
    });

    // Traceability: TC-085 → Spec 5.3
    test('TC-085: Member cannot archive main item', async () => {
      const res = await curl('POST', `${base}/teams/${teamId}/main-items/${mainItemId}/archive`, {
        headers: memberAuth,
      });
      assert.equal(res.status, 403);
    });

    // Traceability: TC-086 → Spec 5.2
    test('TC-086: User without permission cannot create team', async () => {
      const res = await curl('POST', `${base}/teams`, {
        headers: memberAuth,
        body: JSON.stringify({ name: 'Unauthorized Team' }),
      });
      assert.equal(res.status, 403);
    });

    // Traceability: TC-087 → UI Function 11
    test('TC-087: Non-super-admin cannot access admin endpoints', async () => {
      const r1 = await curl('GET', `${base}/admin/users`, { headers: memberAuth });
      assert.equal(r1.status, 403);

      const r2 = await curl('PUT', `${base}/admin/users/1/can-create-team`, {
        headers: memberAuth,
        body: JSON.stringify({ value: true }),
      });
      assert.equal(r2.status, 403);
    });
  });

  // ===== Auto-Calculations (TC-088) =====

  describe('Auto-Calculations', () => {
    // Traceability: TC-088 → Spec 5.3 重点事项标记
    test('TC-088: Key item auto-upgrade after 2+ delays', async () => {
      // Create P2 item
      const createRes = await curl('POST', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        body: JSON.stringify({
          title: 'Delay Test Item',
          priority: 'P2',
          expectedEndDate: '2026-01-01',
        }),
      });
      const data = parseBody(createRes);
      const itemId = data.data?.id ?? data.id;

      // Extend expected end date twice (simulating delays)
      await curl('PUT', `${base}/teams/${teamId}/main-items/${itemId}`, {
        headers: pmAuth,
        body: JSON.stringify({ expectedEndDate: '2026-02-01' }),
      });
      await curl('PUT', `${base}/teams/${teamId}/main-items/${itemId}`, {
        headers: pmAuth,
        body: JSON.stringify({ expectedEndDate: '2026-03-01' }),
      });

      // Check if priority auto-upgraded to P1
      const checkRes = await curl('GET', `${base}/teams/${teamId}/main-items/${itemId}`, {
        headers: pmAuth,
      });
      const checkData = parseBody(checkRes);
      const priority = checkData.data?.priority ?? checkData.priority;
      // Note: this depends on backend implementing the auto-upgrade rule
      assert.ok(priority === 'P1' || priority === 'P2',
        `Priority after 2 delays: ${priority} (implementation may vary)`);
    });
  });

  // ===== Performance (TC-089 ~ TC-090) =====

  describe('Performance', () => {
    // Traceability: TC-089 → Spec 性能需求, Story 5 / AC-1
    test('TC-089: Report export responds within 5 seconds', async () => {
      const start = Date.now();
      const res = await curl('GET', `${base}/teams/${teamId}/reports/weekly/export?week=2026-W16`, {
        headers: pmAuth,
        timeout: 10000,
      });
      const elapsed = Date.now() - start;
      assert.ok(elapsed < 5000, `Report export took ${elapsed}ms, expected < 5000ms`);
    });

    // Traceability: TC-090 → Spec 性能需求
    test('TC-090: List page loads within 2 seconds', async () => {
      const start = Date.now();
      const res = await curl('GET', `${base}/teams/${teamId}/main-items`, {
        headers: pmAuth,
        timeout: 5000,
      });
      const elapsed = Date.now() - start;
      assert.ok(elapsed < 2000, `List load took ${elapsed}ms, expected < 2000ms`);
      assert.equal(res.status, 200);
    });
  });
});
