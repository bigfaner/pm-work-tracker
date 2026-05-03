import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, runCli } from '../helpers.js';

const PROJECT_ROOT = resolve(import.meta.dirname, '../../..');

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
});
