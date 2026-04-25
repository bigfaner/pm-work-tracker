import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from './helpers.js';

// Test team and resource IDs — adjust to match your test environment
const TEAM_ID = 1;
// A known bizKey for a main item in the test DB (update before running)
const KNOWN_BIZ_KEY = process.env.TEST_BIZ_KEY ?? '1234567890123456';

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // Traceability: TC-002 → Story 2 / AC-1
  test('TC-002: SoftDelete API 调用设置 deleted_flag 和 deleted_time', async () => {
    // Step 1: 发送 DELETE 请求软删除主事项
    const res = await authCurl('DELETE', `/api/v1/teams/${TEAM_ID}/main-items/${KNOWN_BIZ_KEY}`);
    assert.ok(
      res.status === 200 || res.status === 204,
      `期望状态码 200 或 204，实际: ${res.status}，响应: ${res.body}`,
    );
    // Note: Step 2 (DB verification) requires direct DB access — verify via GET returning 404
    const getRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items/${KNOWN_BIZ_KEY}`);
    assert.ok(
      getRes.status === 404 || getRes.status === 400,
      `软删后 GET 应返回 404，实际: ${getRes.status}`,
    );
  });

  // Traceability: TC-003 → Story 3 / AC-1
  test('TC-003: 已软删记录不出现在列表和详情 API 响应中', async () => {
    // Step 1: 获取列表，验证已软删记录不在其中
    const listRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items`);
    assert.equal(listRes.status, 200, `列表接口应返回 200，实际: ${listRes.status}`);
    const listData = JSON.parse(listRes.body);
    const items: any[] = listData.items ?? listData.data ?? listData ?? [];
    const found = items.find((item: any) => String(item.bizKey) === KNOWN_BIZ_KEY);
    assert.ok(!found, `已软删记录 bizKey=${KNOWN_BIZ_KEY} 不应出现在列表中`);

    // Step 2: 详情接口应返回 404 或业务错误
    const detailRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items/${KNOWN_BIZ_KEY}`);
    assert.ok(
      detailRes.status === 404 || detailRes.status === 400,
      `已软删记录详情应返回 404，实际: ${detailRes.status}`,
    );
  });

  // Traceability: TC-004 → Story 4 / AC-1
  test('TC-004: 资源 API 响应使用新字段名', async () => {
    // Step 1: 主事项列表
    const mainRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items`);
    assert.equal(mainRes.status, 200, `主事项列表应返回 200，实际: ${mainRes.status}`);
    const mainData = JSON.parse(mainRes.body);
    const mainItems: any[] = mainData.items ?? mainData.data ?? [];
    if (mainItems.length > 0) {
      const item = mainItems[0];
      assert.ok('itemStatus' in item, `主事项应含 itemStatus 字段，实际字段: ${Object.keys(item).join(', ')}`);
      assert.ok(!('status' in item), `主事项不应含旧 status 字段`);
      assert.ok('createTime' in item, `主事项应含 createTime 字段`);
      assert.ok(!('createdAt' in item), `主事项不应含旧 createdAt 字段`);
      assert.ok('dbUpdateTime' in item, `主事项应含 dbUpdateTime 字段`);
      assert.ok(!('updatedAt' in item), `主事项不应含旧 updatedAt 字段`);
      assert.ok(!('deletedAt' in item), `主事项不应含 deletedAt 字段`);
    }

    // Step 2: 待办事项列表
    const poolRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/item-pools`);
    assert.equal(poolRes.status, 200, `待办事项列表应返回 200，实际: ${poolRes.status}`);
    const poolData = JSON.parse(poolRes.body);
    const poolItems: any[] = poolData.items ?? poolData.data ?? [];
    if (poolItems.length > 0) {
      const pool = poolItems[0];
      assert.ok('poolStatus' in pool, `待办事项应含 poolStatus 字段`);
      assert.ok(!('status' in pool), `待办事项不应含旧 status 字段`);
    }
  });

  // Traceability: TC-005 → Story 4 / AC-1; Story 6 / AC-1
  test('TC-005: 资源 API 响应包含 bizKey 且不含 id', async () => {
    // Step 1 & 2: 获取主事项列表，检查字段
    const res = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items`);
    assert.equal(res.status, 200, `主事项列表应返回 200，实际: ${res.status}`);
    const data = JSON.parse(res.body);
    const items: any[] = data.items ?? data.data ?? [];
    assert.ok(items.length > 0, '列表应至少包含一条记录');
    for (const item of items) {
      assert.ok('bizKey' in item, `每个资源对象应含 bizKey 字段，实际字段: ${Object.keys(item).join(', ')}`);
      assert.ok(!('id' in item), `每个资源对象不应含 id 字段（json:"-"）`);
    }
  });

  // Traceability: TC-006 → Story 6 / AC-3
  test('TC-006: 后端通过 bizKey 路径参数正确定位记录', async () => {
    // 先获取一个有效的 bizKey
    const listRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items`);
    assert.equal(listRes.status, 200);
    const listData = JSON.parse(listRes.body);
    const items: any[] = listData.items ?? listData.data ?? [];
    assert.ok(items.length > 0, '列表应至少包含一条记录以供测试');
    const bizKey = items[0].bizKey;

    // Step 1: 通过 bizKey 获取详情
    const res = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items/${bizKey}`);
    assert.equal(res.status, 200, `GET by bizKey 应返回 200，实际: ${res.status}`);

    // Step 2: 验证响应体
    const item = JSON.parse(res.body);
    const returnedBizKey = item.bizKey ?? item.data?.bizKey;
    assert.equal(
      String(returnedBizKey),
      String(bizKey),
      `响应体 bizKey 应与路径参数一致，期望: ${bizKey}，实际: ${returnedBizKey}`,
    );
  });

  // Traceability: TC-007 → Story 6 / AC-2
  test('TC-007: 前端 API 模块使用 bizKey 构造请求路径', async () => {
    // This test verifies the frontend source code uses bizKey in path construction.
    // It reads the frontend API module and checks for bizKey usage patterns.
    const { readFileSync, existsSync } = await import('node:fs');
    const { resolve } = await import('node:path');
    const { fileURLToPath } = await import('node:url');
    const __dirname2 = resolve(fileURLToPath(import.meta.url), '..', '..', '..', '..', '..', '..');
    const apiFile = resolve(__dirname2, 'frontend', 'src', 'api', 'mainItems.ts');

    assert.ok(existsSync(apiFile), `前端 API 模块应存在: ${apiFile}`);
    const content = readFileSync(apiFile, 'utf-8');

    // Verify bizKey is used in path construction
    assert.ok(
      content.includes('bizKey'),
      '前端 API 模块应使用 bizKey 构造请求路径',
    );
    // Verify id is NOT used as path param (item.id pattern)
    const idPathPattern = /`[^`]*\/\$\{[^}]*\.id\}/;
    assert.ok(
      !idPathPattern.test(content),
      '前端 API 模块不应使用 item.id 构造路径参数',
    );
  });
});
