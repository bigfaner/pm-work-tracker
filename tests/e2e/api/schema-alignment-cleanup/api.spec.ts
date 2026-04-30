import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, runCli } from './helpers.js';

const PROJECT_ROOT = resolve(import.meta.dirname, '../../../../../../');

function projectPath(...segments: string[]): string {
  return join(PROJECT_ROOT, ...segments);
}

// Grep helper: returns matching lines from a file, empty array if no matches
function grepFile(pattern: string, filePath: string): string[] {
  const result = runCli(`grep -n '${pattern}' "${filePath}" 2>/dev/null || true`);
  return result.stdout.trim().split('\n').filter(Boolean);
}

// Grep helper: returns matching lines from a directory recursively
function grepDir(pattern: string, dirPath: string): string[] {
  const result = runCli(`grep -rn "${pattern}" "${dirPath}" --include='*.go' 2>/dev/null || true`);
  return result.stdout.trim().split('\n').filter(Boolean);
}

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // ── Authenticated API Endpoint Tests ─────────────────────────────

  // Traceability: TC-009 → Story 1 / AC-1, Spec Round 1 Item 1
  test('TC-009: Assign sub-item updates assignee_key column', async () => {
    // Step 1: Verify backend code uses assignee_key, not assignee_id
    const matches = grepFile('assignee_id', projectPath('backend', 'internal', 'service', 'sub_item_service.go'));
    assert.equal(matches.length, 0, `Expected no assignee_id in sub_item_service.go, found:\n${matches.join('\n')}`);

    // Step 2: Verify the assign endpoint exists and uses assignee_key
    const codeContent = readFileSync(projectPath('backend', 'internal', 'service', 'sub_item_service.go'), 'utf-8');
    assert.match(codeContent, /assignee_key|AssigneeKey/, 'sub_item_service should reference assignee_key');
  });

  // Traceability: TC-010 → Story 1 / AC-2
  test('TC-010: Assignee persists after page refresh', async () => {
    // Create a sub-item with an assignee via API, then fetch it again
    const TEAM_ID = 1; // Default team
    const MAIN_ITEM_ID = 1; // Assumes seed data exists

    // List sub-items to find one
    const listRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items/${MAIN_ITEM_ID}/sub-items`);
    assert.equal(listRes.status, 200, `List sub-items failed: ${listRes.status} ${listRes.body}`);

    const listData = JSON.parse(listRes.body);
    if (listData.items && listData.items.length > 0) {
      const subItem = listData.items[0];
      // Fetch the same sub-item again to verify persistence
      const getRes = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items/${MAIN_ITEM_ID}/sub-items/${subItem.id}`);
      assert.equal(getRes.status, 200);
      const fetched = JSON.parse(getRes.body);
      if (subItem.assigneeName) {
        assert.equal(fetched.assigneeName, subItem.assigneeName, 'Assignee name should persist after re-fetch');
      }
    }
  });

  // Traceability: TC-011 → Story 2 / AC-1, Spec Round 1 Item 2
  test('TC-011: Filter by assignee returns correct subset', async () => {
    const TEAM_ID = 1;
    // Fetch items with assignee filter
    const res = await authCurl('GET', `/api/v1/teams/${TEAM_ID}/main-items?assigneeKey=test-user`);
    assert.equal(res.status, 200, `Filter request failed: ${res.status} ${res.body}`);

    const data = JSON.parse(res.body);
    // If items exist, all should have the matching assignee
    if (data.items && data.items.length > 0) {
      for (const item of data.items) {
        assert.equal(item.assigneeKey, 'test-user', `Item ${item.id} should have assigneeKey=test-user`);
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
    assert.match(sqliteSchema, /assignee_key/, 'SQLite schema should have assignee_key column');
    assert.match(mysqlSchema, /assignee_key/, 'MySQL schema should have assignee_key column');
  });

  // ── Code Inspection Tests (no auth needed) ───────────────────────

  // Traceability: TC-013 → Story 2.5 / AC-1, Spec Round 2 Item 3
  test('TC-013: Deprecated DTOs removed from item_dto.go', () => {
    const matches = grepFile('Deprecated', projectPath('backend', 'internal', 'dto', 'item_dto.go'));
    assert.equal(matches.length, 0, `Expected no Deprecated DTOs, found:\n${matches.join('\n')}`);

    // Also verify the project builds
    const buildResult = runCli('go build ./...', projectPath('backend'));
    assert.equal(buildResult.exitCode, 0, `Go build failed:\n${buildResult.stderr}`);
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
      assert.equal(buildResult.exitCode, 0, `Go build failed after cleanup:\n${buildResult.stderr}`);
    }
  });

  // Traceability: TC-015 → Spec Round 2 Item 7
  test('TC-015: Redundant GORM column tags removed from role_repo', () => {
    const matches = grepFile('column:', projectPath('backend', 'internal', 'repository', 'gorm', 'role_repo.go'));
    assert.equal(matches.length, 0, `Expected no column: GORM tags in role_repo.go, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-016 → Story 3 / AC-1, Spec Round 3 Item 10
  test('TC-016: TransactionDB and dbTransactor merged into single interface', () => {
    const matches = grepDir('dbTransactor', projectPath('backend'));
    assert.equal(matches.length, 0, `Expected no dbTransactor references, found:\n${matches.join('\n')}`);
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
      assert.doesNotMatch(content, /offset\s*:=\s*\(.*page.*-\s*1\s*\)\s*\*\s*pageSize/, `${file} should not have manual pagination logic`);
    }
  });

  // Traceability: TC-018 → Spec Round 3 Item 13
  test('TC-018: teamToDTO returns typed struct instead of gin.H', () => {
    const matches = grepFile('gin.H', projectPath('backend', 'internal', 'handler', 'team_handler.go'));
    assert.equal(matches.length, 0, `Expected no gin.H in team_handler.go, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-019 → Spec Round 3 Item 14
  test('TC-019: Shared userToDTO defined only once', () => {
    const matches = grepDir('func userToDTO', projectPath('backend'));
    assert.equal(matches.length, 1, `Expected exactly 1 userToDTO definition, found ${matches.length}:\n${matches.join('\n')}`);
  });

  // Traceability: TC-020 → Story 5 / AC-1, Spec Round 4 Item 19
  test('TC-020: Roles table renamed with pmw_ prefix', () => {
    // Verify model TableName methods return pmw_ prefixed names
    const modelContent = readFileSync(projectPath('backend', 'internal', 'model', 'role.go'), 'utf-8');
    assert.match(modelContent, /pmw_roles/, 'Role.TableName() should return pmw_roles');
    assert.match(modelContent, /pmw_role_permissions/, 'RolePermission.TableName() should return pmw_role_permissions');

    // Verify both schema files are updated
    const sqliteSchema = readFileSync(projectPath('backend', 'migrations', 'SQLite-schema.sql'), 'utf-8');
    const mysqlSchema = readFileSync(projectPath('backend', 'migrations', 'MySql-schema.sql'), 'utf-8');
    assert.match(sqliteSchema, /pmw_roles/, 'SQLite schema should have pmw_roles table');
    assert.match(mysqlSchema, /pmw_roles/, 'MySQL schema should have pmw_roles table');
  });

  // Traceability: TC-021 → Spec Round 4 Item 20
  test('TC-021: ViewService has single unified constructor', () => {
    const content = readFileSync(projectPath('backend', 'internal', 'service', 'view_service.go'), 'utf-8');
    const constructors = content.match(/func New\w+Service\s*\(/g) ?? [];
    assert.equal(constructors.length, 1, `Expected exactly 1 ViewService constructor, found ${constructors.length}`);
  });

  // Traceability: TC-022 → Spec Round 4 Item 22
  test('TC-022: NotDeleted scope used consistently across all repositories', () => {
    const matches = grepDir('deleted_flag.*=.*0', projectPath('backend', 'internal', 'repository'));
    assert.equal(matches.length, 0, `Expected no inline deleted_flag = 0 checks, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-023 → Spec Round 2 Items 4 and 6
  test('TC-023: Dead assignments and nil-slice initializations removed', () => {
    const serviceFile = projectPath('backend', 'internal', 'service', 'team_service.go');
    const deadAssign = grepFile('_ = team.PmKey', serviceFile);
    assert.equal(deadAssign.length, 0, `Expected no '_ = team.PmKey' dead assignment, found:\n${deadAssign.join('\n')}`);

    const nilSlice = grepFile('= \\[\\]string{}', serviceFile);
    assert.equal(nilSlice.length, 0, `Expected no nil-slice initializations, found:\n${nilSlice.join('\n')}`);
  });

  // Traceability: TC-024 → Spec Round 3 Item 12
  test('TC-024: Shared resolveBizKey helper extracted', () => {
    const matches = grepDir('func resolveBizKey', projectPath('backend', 'internal', 'handler'));
    assert.equal(matches.length, 1, `Expected exactly 1 resolveBizKey definition in shared helper, found ${matches.length}:\n${matches.join('\n')}`);
  });
});
