import { test, expect } from '@playwright/test';
import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from '../helpers.js';

const __dirname = resolve(fileURLToPath(import.meta.url), '..');
const PROJECT_ROOT = resolve(__dirname, '..', '..', '..');

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let testBizKey: string;
  let softDeleteBizKey: string;
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

    // Create test data: two main items
    const makeItem = (title: string) => JSON.stringify({
      title,
      priority: 'P1',
      assigneeKey: '1',
      startDate: '2026-01-01',
      expectedEndDate: '2026-12-31',
    });

    const res1 = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items`, { body: makeItem('E2E Schema Test Item') });
    expect(res1.status === 200 || res1.status === 201).toBeTruthy();
    const bk1 = res1.body.match(/"bizKey"\s*:\s*"(\d+)"/);
    testBizKey = bk1 ? bk1[1] : '';

    const res2 = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items`, { body: makeItem('E2E Soft Delete Target') });
    expect(res2.status === 200 || res2.status === 201).toBeTruthy();
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
      expect(statusRes.status === 200).toBeTruthy();
    }

    // Archive the item via POST /archive with bizKey path param
    const archiveRes = await authCurl('POST', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}/archive`);
    expect(
      archiveRes.status === 200,
    ).toBeTruthy();

    // Verify record is archived by checking archivedAt in GET response
    const getRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}`);
    expect(getRes.status).toBe(200);
    const archivedAtMatch = getRes.body.match(/"archivedAt"\s*:\s*"(\d{4}-\d{2}-\d{2})/);
    expect(archivedAtMatch).toBeTruthy();
  });

  // Traceability: TC-003 → Story 3 / AC-1
  test('TC-003: 已归档记录不出现在列表 API 响应中', async () => {
    // Step 1: List should not contain the archived record
    const listRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items`);
    expect(listRes.status).toBe(200);
    // Extract bizKey values via regex to avoid int64 precision loss
    const bizKeyMatches = listRes.body.matchAll(/"bizKey"\s*:\s*"(\d+)"/g);
    const listedBizKeys = [...bizKeyMatches].map((m) => m[1]);
    const archivedInList = listedBizKeys.includes(softDeleteBizKey);
    expect(!archivedInList).toBeTruthy();

    // Step 2: Detail endpoint still returns the archived item (with archivedAt set)
    const detailRes = await authCurl('GET', `/v1/teams/${TEAM_BIZ_KEY}/main-items/${softDeleteBizKey}`);
    expect(detailRes.status).toBe(200);
    const archivedAtMatch = detailRes.body.match(/"archivedAt"\s*:\s*"(\d{4}-\d{2}-\d{2})/);
    expect(archivedAtMatch).toBeTruthy();
  });
});
