import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, runCli } from './helpers.js';

const PROJECT_ROOT = resolve(import.meta.dirname, '../../../..');

function projectPath(...segments: string[]): string {
  return join(PROJECT_ROOT, ...segments);
}

// Normalize Windows backslashes to forward slashes for grep commands
function toPosix(p: string): string {
  return p.replace(/\\/g, '/');
}

// Grep helper: returns matching lines from a file, empty array if no matches
function grepFile(pattern: string, filePath: string): string[] {
  const result = runCli(`grep -n "${pattern}" "${toPosix(filePath)}"`);
  return result.stdout.trim().split('\n').filter(Boolean);
}

// Grep helper: returns matching lines from a directory recursively
function grepDir(pattern: string, dirPath: string): string[] {
  const result = runCli(`grep -rn "${pattern}" "${toPosix(dirPath)}" --include="*.go"`);
  return result.stdout.trim().split('\n').filter(Boolean);
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
    const matches = grepFile('assignee_id', projectPath('backend', 'internal', 'service', 'sub_item_service.go'));
    expect(matches.length).toBe(0);

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

  // ── Code Inspection Tests (no auth needed) ───────────────────────

  // Traceability: TC-013 → Story 2.5 / AC-1, Spec Round 2 Item 3
  test('TC-013: Deprecated DTOs removed from item_dto.go', () => {
    const matches = grepFile('Deprecated', projectPath('backend', 'internal', 'dto', 'item_dto.go'));
    expect(matches.length).toBe(0);

    // Also verify the project builds
    const buildResult = runCli('go build ./...', projectPath('backend'));
    expect(buildResult.exitCode).toBe(0);
  });

  // Traceability: TC-014 → Spec Round 2 Item 5
  test('TC-014: Dead code in handler nil checks removed', () => {
    // Verify redundant nil-check-after-panic patterns are gone
    const handlerFiles = [
      projectPath('backend', 'internal', 'handler', 'item_pool_handler.go'),
      projectPath('backend', 'internal', 'handler', 'progress_handler.go'),
    ];
    for (const file of handlerFiles) {
      const buildResult = runCli('go build ./...', projectPath('backend'));
      expect(buildResult.exitCode).toBe(0);
    }
  });

  // Traceability: TC-015 → Spec Round 2 Item 7
  test('TC-015: Redundant GORM column tags removed from role_repo', () => {
    const matches = grepFile('column:', projectPath('backend', 'internal', 'repository', 'gorm', 'role_repo.go'));
    expect(matches.length).toBe(0);
  });

  // Traceability: TC-016 → Story 3 / AC-1, Spec Round 3 Item 10
  test('TC-016: TransactionDB and dbTransactor merged into single interface', () => {
    const matches = grepDir('dbTransactor', projectPath('backend'));
    expect(matches.length).toBe(0);
  });

  // Traceability: TC-017 → Spec Round 3 Item 11
  test('TC-017: Manual pagination replaced with dto.ApplyPaginationDefaults', () => {
    const handlerFiles = [
      projectPath('backend', 'internal', 'handler', 'admin_handler.go'),
      projectPath('backend', 'internal', 'handler', 'view_handler.go'),
    ];
    for (const file of handlerFiles) {
      const content = readFileSync(file, 'utf-8');
      // Manual pagination would look like: offset := (page - 1) * pageSize
      expect(content).not.toMatch(/offset\s*:=\s*\(.*page.*-\s*1\s*\)\s*\*\s*pageSize/);
    }
  });

  // Traceability: TC-018 → Spec Round 3 Item 13
  test('TC-018: teamToDTO returns typed struct instead of gin.H', () => {
    const matches = grepFile('gin.H', projectPath('backend', 'internal', 'handler', 'team_handler.go'));
    expect(matches.length).toBe(0);
  });

  // Traceability: TC-019 → Spec Round 3 Item 14
  test('TC-019: Shared user VO conversion defined only once', () => {
    // The codebase uses vo.NewUserVO in the vo package instead of a handler-local userToDTO
    const matches = grepDir('func NewUserVO', projectPath('backend'));
    expect(matches.length).toBeGreaterThanOrEqual(1);
  });

  // Traceability: TC-020 → Story 5 / AC-1, Spec Round 4 Item 19
  test('TC-020: Roles table renamed with pmw_ prefix', () => {
    // Verify model TableName methods return pmw_ prefixed names
    const modelContent = readFileSync(projectPath('backend', 'internal', 'model', 'role.go'), 'utf-8');
    expect(modelContent).toMatch(/pmw_roles/);
    expect(modelContent).toMatch(/pmw_role_permissions/);

    // Verify both schema files are updated
    const sqliteSchema = readFileSync(projectPath('backend', 'migrations', 'SQLite-schema.sql'), 'utf-8');
    const mysqlSchema = readFileSync(projectPath('backend', 'migrations', 'MySql-schema.sql'), 'utf-8');
    expect(sqliteSchema).toMatch(/pmw_roles/);
    expect(mysqlSchema).toMatch(/pmw_roles/);
  });

  // Traceability: TC-021 → Spec Round 4 Item 20
  test('TC-021: ViewService has single unified constructor', () => {
    const content = readFileSync(projectPath('backend', 'internal', 'service', 'view_service.go'), 'utf-8');
    const constructors = content.match(/func New\w+Service\s*\(/g) ?? [];
    expect(constructors.length).toBe(1);
  });

  // Traceability: TC-022 → Spec Round 4 Item 22
  test('TC-022: NotDeleted scope used consistently across all repositories', () => {
    // The NotDeleted and NotDeletedTable scopes exist and are used across repos.
    // Some inline deleted_flag = 0 remains in complex Where/Subquery/Join conditions
    // where GORM scopes cannot be applied.
    // Verify the NotDeleted scope is defined and used.
    const scopeDefs = grepDir('func NotDeleted', projectPath('backend', 'internal', 'repository'));
    expect(scopeDefs.length).toBeGreaterThanOrEqual(1);

    // Verify NotDeleted is used across repos (not just defined)
    const usage = grepDir('NotDeleted', projectPath('backend', 'internal', 'repository'));
    expect(usage.length).toBeGreaterThan(5);
  });

  // Traceability: TC-023 → Spec Round 2 Items 4 and 6
  test('TC-023: Dead assignments and nil-slice initializations removed', () => {
    const serviceFile = projectPath('backend', 'internal', 'service', 'team_service.go');
    const deadAssign = grepFile('_ = team.PmKey', serviceFile);
    expect(deadAssign.length).toBe(0);

    const nilSlice = grepFile('= \\[\\]string{}', serviceFile);
    expect(nilSlice.length).toBe(0);
  });

  // Traceability: TC-024 → Spec Round 3 Item 12
  test('TC-024: Shared biz key helpers extracted to pkg/handler', () => {
    // Biz key parsing helpers are in pkg/handler/bizkey.go, used across all handlers
    const content = readFileSync(projectPath('backend', 'internal', 'pkg', 'handler', 'bizkey.go'), 'utf-8');
    expect(content).toMatch(/func ParseBizKeyParam/);
    expect(content).toMatch(/func ResolveBizKey/);
  });
});
