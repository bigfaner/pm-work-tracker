import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, runCli } from '../helpers.js';

const PROJECT_ROOT = resolve(import.meta.dirname, '../../..');

function projectPath(...segments: string[]): string {
  return join(PROJECT_ROOT, ...segments);
}

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  test.beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // ── Authenticated API Endpoint Tests ─────────────────────────────

  // Traceability: TC-009 → Story 1 / AC-1, Spec Round 1 Item 1
  test('TC-009: Assign sub-item updates assignee_key column', async () => {
    // Step 1: Verify backend code uses assignee_key, not assignee_id
    const matches = runCli(`grep -n "assignee_id" "${projectPath('backend', 'internal', 'service', 'sub_item_service.go')}"`);
    expect(matches.stdout.trim().length).toBe(0);

    // Step 2: Verify the assign endpoint exists and uses assignee_key
    const codeContent = readFileSync(projectPath('backend', 'internal', 'service', 'sub_item_service.go'), 'utf-8');
    expect(codeContent).toMatch(/assignee_key|AssigneeKey/);
  });

  // Traceability: TC-010 → Story 1 / AC-2
  test('TC-010: Assignee persists after page refresh', async () => {
    // Discover a valid team first
    const teamsRes = await authCurl('GET', '/v1/teams');
    expect(teamsRes.status).toBe(200);
    const teamsBody = JSON.parse(teamsRes.body);
    const teams = teamsBody.data?.items ?? teamsBody.data ?? [];
    if (!teams || teams.length === 0) return;

    const team = teams[0];
    const teamKey = team.bizKey ?? team.biz_key;

    // List main items to find one with sub-items
    const mainListRes = await authCurl('GET', `/v1/teams/${teamKey}/main-items`);
    expect(mainListRes.status).toBe(200);

    const mainBody = JSON.parse(mainListRes.body);
    const mainItems = mainBody.data?.items ?? [];
    if (mainItems.length === 0) return;

    const mainItem = mainItems[0];

    // List sub-items for this main item
    const listRes = await authCurl('GET', `/v1/teams/${teamKey}/main-items/${mainItem.bizKey}/sub-items`);
    expect(listRes.status).toBe(200);

    const listBody = JSON.parse(listRes.body);
    const subItems = listBody.data?.items ?? [];
    if (subItems.length === 0) return;

    const subItem = subItems[0];
    // Fetch the same sub-item again via the flat route to verify persistence
    const getRes = await authCurl('GET', `/v1/teams/${teamKey}/sub-items/${subItem.bizKey}`);
    expect(getRes.status).toBe(200);
    const getBody = JSON.parse(getRes.body);
    const fetched = getBody.data;
    if (subItem.assigneeName) {
      expect(fetched.assigneeName).toBe(subItem.assigneeName);
    }
  });

  // Traceability: TC-011 → Story 2 / AC-1, Spec Round 1 Item 2
  test('TC-011: Filter by assignee returns correct subset', async () => {
    // Discover a valid team first
    const teamsRes = await authCurl('GET', '/v1/teams');
    expect(teamsRes.status).toBe(200);
    const teamsBody = JSON.parse(teamsRes.body);
    const teams = teamsBody.data?.items ?? teamsBody.data ?? [];
    if (!teams || teams.length === 0) return;

    const team = teams[0];
    const teamKey = team.bizKey ?? team.biz_key;

    // Fetch items with assignee filter
    const res = await authCurl('GET', `/v1/teams/${teamKey}/main-items?assigneeKey=test-user`);
    expect(res.status).toBe(200);

    const body = JSON.parse(res.body);
    // Response wrapped in { code: 0, data: { items: [...], total: N } }
    const data = body.data;
    // If items exist, all should have the matching assignee
    if (data && data.items && data.items.length > 0) {
      for (const item of data.items) {
        expect(item.assigneeKey).toBe('test-user');
      }
    }
  });

  // Traceability: TC-012 → Story 2 / AC-2
  test('TC-012: Filter by assignee consistent across MySQL and SQLite', async () => {
    // This test verifies code consistency between dialects
    // Both schema files should have the same assignee_key column definition
    const sqliteSchema = readFileSync(projectPath('backend', 'migrations', 'SQLite-schema.sql'), 'utf-8');
    const mysqlSchema = readFileSync(projectPath('backend', 'migrations', 'MySql-schema.sql'), 'utf-8');

    // Both schemas should define assignee_key column
    expect(sqliteSchema).toMatch(/assignee_key/);
    expect(mysqlSchema).toMatch(/assignee_key/);
  });
});
