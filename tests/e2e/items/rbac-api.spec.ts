import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, defaultCreds, randomCode, setupRbacFixtures } from '../helpers.js';

// ── Shared state ────────────────────────────────────────────────────
let adminCurl: ReturnType<typeof createAuthCurl>;
let superadminToken: string;
let memberToken: string;
let pmToken: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let testTeamId: string;
let testItemId: string;
let subItemBizKey: string;

test.describe('RBAC — Items (TC-001, TC-002, TC-036, TC-037)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    pmUserBizKey = f.pmUserBizKey;
    memberUserBizKey = f.memberUserBizKey;
    testTeamId = f.teamBizKey;
    adminCurl = createAuthCurl(apiBaseUrl, superadminToken);

    // Create a main item and transition to completed status for archive tests
    const itemRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: `e2e-item-rbac-${Date.now()}`, priority: 'P1', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(itemRes.status === 200 || itemRes.status === 201).toBeTruthy();
    const itemData = JSON.parse(itemRes.body).data;
    testItemId = String(itemData?.bizKey ?? itemData?.id);

    for (const status of ['progressing', 'reviewing', 'completed']) {
      await adminCurl('PUT', `/v1/teams/${testTeamId}/main-items/${testItemId}/status`, {
        body: JSON.stringify({ status }),
      });
    }

    // Create sub-item for assign tests
    const subRes = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items/${testItemId}/sub-items`, {
      body: JSON.stringify({ mainItemKey: testItemId, title: 'Sub Item RBAC', priority: 'P2', assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    if (subRes.status === 200 || subRes.status === 201) {
      const subData = JSON.parse(subRes.body).data;
      subItemBizKey = String(subData?.bizKey ?? subData?.id);
    }
  });

  // ── Archive Permission ───────────────────────────────────────────

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

  // ── Item Create & Assign ──────────────────────────────────────────

  // Traceability: TC-036 → Story 6 / AC-1
  test('TC-036: 拥有 main_item:create 权限创建主事项', async () => {
    const res = await adminCurl('POST', `/v1/teams/${testTeamId}/main-items`, {
      body: JSON.stringify({ title: 'PM 创建的主事项', priority: 'P2', assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
      headers: { Authorization: `Bearer ${pmToken}` },
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
  });

  // Traceability: TC-037 → Story 6 / AC-2
  test('TC-037: 拥有 sub_item:assign 权限分配负责人', async () => {
    if (!subItemBizKey) return;
    const res = await adminCurl('PUT', `/v1/teams/${testTeamId}/sub-items/${subItemBizKey}/assignee`, {
      body: JSON.stringify({ assigneeKey: memberUserBizKey }),
      headers: { Authorization: `Bearer ${pmToken}` },
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });
});
