import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from './helpers.js';

const __dirname = resolve(fileURLToPath(import.meta.url), '..');
const PROJECT_ROOT = resolve(__dirname, '..', '..', '..', '..', '..');

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let testBizKey: string;
  let softDeleteBizKey: string;
  let TEAM_BIZ_KEY: string;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);

    // Resolve team bizKey: use env var, or pick the first available team
    if (process.env.TEAM_ID) {
      TEAM_BIZ_KEY = process.env.TEAM_ID;
    } else {
      const teamsRes = await authCurl('GET', '/v1/teams');
      const teamsData = JSON.parse(teamsRes.body);
      const teams = teamsData.data?.items ?? [];
      const team = teams.find((t: any) => t.bizKey && t.bizKey !== '0') || teams[0];
      if (!team?.bizKey) throw new Error('No team found. Create a team first or set TEAM_ID env var.');
      TEAM_BIZ_KEY = team.bizKey;
    }

    // Create test data: two main items
    const makeItem = (title: string) => JSON.stringify({
      title,
      priority: 'P1',
      assigneeKey: '1',
      startDate: '2026-01-01',
      expectedEndDate: '2026-12-31',
    });

    const res1 = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items`, { body: makeItem('E2E Schema Test Item') });
    assert.ok(res1.status === 200 || res1.status === 201, `Create test item failed: ${res1.body}`);
    // Extract bizKey as string to preserve int64 precision (JSON.parse loses precision)
    const bk1 = res1.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    testBizKey = bk1 ? bk1[1] : '';

    const res2 = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items`, { body: makeItem('E2E Soft Delete Target') });
    assert.ok(res2.status === 200 || res2.status === 201, `Create delete target failed: ${res2.body}`);
    const bk2 = res2.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    softDeleteBizKey = bk2 ? bk2[1] : '';
  });

  // ── Authenticated Tests (use shared auth) ───────────────────────

  // Traceability: TC-002 → Story 2 / AC-1
  test('TC-002: Archive API 通过 bizKey 路径参数正确归档记录', async () => {
    // Transition item to "completed" so it can be archived:
    // pending → progressing → reviewing → completed
    const transitions = ['progressing', 'reviewing', 'completed'];
    for (const status of transitions) {
      const statusRes = await authCurl('PUT', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}/status`, {
        body: JSON.stringify({ status }),
      });
      assert.ok(statusRes.status === 200, `Status change to ${status} failed: ${statusRes.status} ${statusRes.body}`);
    }

    // Archive the item via POST /archive with bizKey path param
    const archiveRes = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}/archive`);
    assert.ok(
      archiveRes.status === 200,
      `Expected 200, got ${archiveRes.status}: ${archiveRes.body}`,
    );

    // Verify record is archived by checking archivedAt in GET response
    const getRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}`);
    assert.equal(getRes.status, 200, `GET archived item should still return 200, got ${getRes.status}`);
    const archivedAtMatch = getRes.body.match(/"archivedAt"\s*:\s*"(\d{4}-\d{2}-\d{2})/);
    assert.ok(archivedAtMatch, `Archived item should have archivedAt set, body: ${getRes.body.slice(0, 200)}`);
  });

  // Traceability: TC-003 → Story 3 / AC-1
  test('TC-003: 已归档记录不出现在列表 API 响应中', async () => {
    // Step 1: List should not contain the archived record
    const listRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    assert.equal(listRes.status, 200, `List should return 200, got ${listRes.status}`);
    // Extract bizKey values via regex to avoid int64 precision loss
    const bizKeyMatches = listRes.body.matchAll(/"bizKey"\s*:\s*"(\d+)"/g);
    const listedBizKeys = [...bizKeyMatches].map((m) => m[1]);
    const archivedInList = listedBizKeys.includes(softDeleteBizKey);
    assert.ok(!archivedInList, `Archived record (bizKey=${softDeleteBizKey}) should not appear in list`);

    // Step 2: Detail endpoint still returns the archived item (with archivedAt set)
    const detailRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}`);
    assert.equal(detailRes.status, 200, `GET archived item should return 200, got ${detailRes.status}`);
    const archivedAtMatch = detailRes.body.match(/"archivedAt"\s*:\s*"(\d{4}-\d{2}-\d{2})/);
    assert.ok(archivedAtMatch, `Archived item should have archivedAt set in detail response`);
  });

  // Traceability: TC-004 → Story 4 / AC-1
  test('TC-004: 资源 API 响应使用新字段名', async () => {
    // Step 1: Main items list — verify new field names
    const mainRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    assert.equal(mainRes.status, 200, `Main items list failed: ${mainRes.status}`);
    const mainData = JSON.parse(mainRes.body);
    const mainItems: any[] = mainData.data?.items ?? [];
    if (mainItems.length > 0) {
      const item = mainItems[0];
      assert.ok('bizKey' in item, `Should contain bizKey, got: ${Object.keys(item).join(', ')}`);
      assert.ok('createTime' in item, `Should contain createTime`);
      assert.ok('dbUpdateTime' in item, `Should contain dbUpdateTime`);
      assert.ok(!('createdAt' in item), `Should NOT contain createdAt`);
      assert.ok(!('updatedAt' in item), `Should NOT contain updatedAt`);
      assert.ok(!('deletedAt' in item), `Should NOT contain deletedAt`);
    }

    // Step 2: Item pools list — verify new field names
    const poolRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/item-pool`);
    assert.equal(poolRes.status, 200, `Item pools list failed: ${poolRes.status}`);
    const poolData = JSON.parse(poolRes.body);
    const pools: any[] = poolData.data?.items ?? [];
    if (pools.length > 0) {
      assert.ok('bizKey' in pools[0], 'Pool should contain bizKey');
      assert.ok('createTime' in pools[0], 'Pool should contain createTime');
      assert.ok(!('createdAt' in pools[0]), 'Pool should NOT contain createdAt');
    }
  });

  // Traceability: TC-005 → Story 4 / AC-1; Story 6 / AC-1
  test('TC-005: 资源 API 响应包含 bizKey 且不含 id', async () => {
    const res = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    assert.equal(res.status, 200, `List failed: ${res.status}`);
    const data = JSON.parse(res.body);
    const items: any[] = data.data?.items ?? [];
    assert.ok(items.length > 0, 'Should have at least one item');

    for (const item of items) {
      // bizKey should exist and be a string (int64 serialized as string to preserve precision)
      assert.ok('bizKey' in item, `Item should contain bizKey, got: ${Object.keys(item).join(', ')}`);
      assert.ok(typeof item.bizKey === 'string', `bizKey should be a string, got ${typeof item.bizKey}`);
      assert.ok(/^\d+$/.test(item.bizKey), `bizKey should be a numeric string, got ${item.bizKey}`);

      // id should NOT be exposed (json:"-" on model)
      assert.ok(!('id' in item), `Item should NOT expose id field, got id=${item.id}`);
    }
  });

  // Traceability: TC-006 → Story 6 / AC-3
  test('TC-006: 后端通过 bizKey 路径参数正确定位记录', async () => {
    assert.ok(testBizKey, 'Test bizKey should be available from before hook');

    const res = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${testBizKey}`);
    assert.equal(res.status, 200, `GET by bizKey failed: ${res.status} ${res.body}`);

    // Extract bizKey via regex to avoid int64 precision loss from JSON.parse
    const bkMatch = res.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    const responseBizKey = bkMatch ? bkMatch[1] : '';
    assert.equal(responseBizKey, testBizKey, `Response bizKey should match path parameter`);
    assert.ok(res.body.includes('"title"'), 'Response should contain title field');
  });

  // ── Public Tests (no auth needed — code inspection) ──────────────

  // Traceability: TC-007 → Story 6 / AC-2
  test('TC-007: 前端 API 模块使用 bizKey 构造请求路径', async () => {
    const apiFile = resolve(PROJECT_ROOT, 'frontend', 'src', 'api', 'mainItems.ts');
    assert.ok(existsSync(apiFile), `Frontend API module should exist: ${apiFile}`);

    const content = readFileSync(apiFile, 'utf-8');

    // Step 1: Verify bizKey is used in path construction
    assert.ok(content.includes('bizKey'), 'API module should use bizKey in path construction');

    // Verify id is NOT used as path parameter
    const idPathPattern = /`[^`]*\/\$\{[^}]*\.id\}/;
    assert.ok(!idPathPattern.test(content), 'API module should NOT use item.id for path params');
  });
});
