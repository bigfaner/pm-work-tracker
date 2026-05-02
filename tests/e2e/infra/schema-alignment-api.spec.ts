import { test, expect } from '@playwright/test';
import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from '../helpers.js';

const __dirname = resolve(fileURLToPath(import.meta.url), '..');
const PROJECT_ROOT = resolve(__dirname, '..', '..', '..');

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let TEAM_BIZ_KEY: string;

  test.beforeAll(async () => {
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
  });

  // Traceability: TC-004 → Story 4 / AC-1
  test('TC-004: 资源 API 响应使用新字段名', async () => {
    // Step 1: Main items list — verify new field names
    const mainRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    expect(mainRes.status).toBe(200);
    const mainData = JSON.parse(mainRes.body);
    const mainItems: any[] = mainData.data?.items ?? [];
    if (mainItems.length > 0) {
      const item = mainItems[0];
      expect('bizKey' in item).toBeTruthy();
      expect('createTime' in item).toBeTruthy();
      expect('dbUpdateTime' in item).toBeTruthy();
      expect(!('createdAt' in item)).toBeTruthy();
      expect(!('updatedAt' in item)).toBeTruthy();
      expect(!('deletedAt' in item)).toBeTruthy();
    }

    // Step 2: Item pools list — verify new field names
    const poolRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/item-pool`);
    expect(poolRes.status).toBe(200);
    const poolData = JSON.parse(poolRes.body);
    const pools: any[] = poolData.data?.items ?? [];
    if (pools.length > 0) {
      expect('bizKey' in pools[0]).toBeTruthy();
      expect('createTime' in pools[0]).toBeTruthy();
      expect(!('createdAt' in pools[0])).toBeTruthy();
    }
  });

  // Traceability: TC-005 → Story 4 / AC-1; Story 6 / AC-1
  test('TC-005: 资源 API 响应包含 bizKey 且不含 id', async () => {
    const res = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    const items: any[] = data.data?.items ?? [];
    expect(items.length > 0).toBeTruthy();

    for (const item of items) {
      // bizKey should exist and be a string (int64 serialized as string to preserve precision)
      expect('bizKey' in item).toBeTruthy();
      expect(typeof item.bizKey === 'string').toBeTruthy();
      expect(/^\d+$/.test(item.bizKey)).toBeTruthy();

      // id should NOT be exposed (json:"-" on model)
      expect(!('id' in item)).toBeTruthy();
    }
  });

  // Traceability: TC-006 → Story 6 / AC-3
  test('TC-006: 后端通过 bizKey 路径参数正确定位记录', async () => {
    // Create a test item to get a fresh bizKey
    const makeItem = JSON.stringify({
      title: 'E2E TC-006 test item',
      priority: 'P1',
      assigneeKey: '1',
      startDate: '2026-01-01',
      expectedEndDate: '2026-12-31',
    });
    const createRes = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items`, { body: makeItem });
    expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();
    const bkMatch = createRes.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    const testBizKey = bkMatch ? bkMatch[1] : '';
    expect(testBizKey).toBeTruthy();

    const res = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${testBizKey}`);
    expect(res.status).toBe(200);

    // Extract bizKey via regex to avoid int64 precision loss from JSON.parse
    const bkMatch2 = res.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    const responseBizKey = bkMatch2 ? bkMatch2[1] : '';
    expect(responseBizKey).toBe(testBizKey);
    expect(res.body.includes('"title"')).toBeTruthy();
  });

  // ── Public Tests (no auth needed — code inspection) ──────────────

  // Traceability: TC-007 → Story 6 / AC-2
  test('TC-007: 前端 API 模块使用 bizKey 构造请求路径', async () => {
    const apiFile = resolve(PROJECT_ROOT, 'frontend', 'src', 'api', 'mainItems.ts');
    expect(existsSync(apiFile)).toBeTruthy();

    const content = readFileSync(apiFile, 'utf-8');

    // Step 1: Verify bizKey is used in path construction
    expect(content.includes('bizKey')).toBeTruthy();

    // Verify id is NOT used as path parameter
    const idPathPattern = /`[^`]*\/\$\{[^}]*\.id\}/;
    expect(!idPathPattern.test(content)).toBeTruthy();
  });
});
