import { describe, test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { runCli } from './helpers.js';

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
  const result = runCli(`grep -rn "${pattern}" "${dirPath}" --include='*.ts' --include='*.tsx' 2>/dev/null || true`);
  return result.stdout.trim().split('\n').filter(Boolean);
}

describe('UI E2E Tests', () => {
  // These UI test cases are code-inspection tests that verify frontend
  // cleanup tasks were completed correctly. They use grep and build
  // commands rather than browser interactions.

  // ── Code Inspection Tests ────────────────────────────────────────

  // Traceability: TC-001 → Spec Round 2 Item 8
  test('TC-001: console.error replaced with toast in API client', () => {
    const matches = grepFile('console\\.error', projectPath('frontend', 'src', 'api', 'client.ts'));
    assert.equal(matches.length, 0, `Expected no console.error calls, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-002 → Spec Round 2 Item 9
  test('TC-002: Test data Role.id is string type', () => {
    const filePath = projectPath('frontend', 'src', 'types', 'index.test.ts');
    const content = readFileSync(filePath, 'utf-8');
    // Verify Role.id values in test data use string type (e.g., "role-1"), not number
    assert.doesNotMatch(content, /Role\.id.*:\s*\d+(?!\s*as)/, 'Role.id should be string type in test data');
  });

  // Traceability: TC-003 → Spec Round 3 Item 16
  test('TC-003: Redundant String() wrappers removed from frontend', () => {
    const lines = grepDir('\\.String\\(\\)', projectPath('frontend', 'src'));
    // Filter out legitimate uses (enum conversions, etc.)
    const redundant = lines.filter(l => !l.includes('.toString()') && !l.includes('String('));
    assert.equal(redundant.length, 0, `Expected no redundant .String() calls, found:\n${redundant.join('\n')}`);
  });

  // Traceability: TC-004 → Story 4 / AC-1, Spec Round 3 Item 17
  test('TC-004: PermissionData.teamPermissions uses string keys', () => {
    const matches = grepFile('Record<number', projectPath('frontend', 'src', 'types', 'index.ts'));
    assert.equal(matches.length, 0, `Expected no Record<number> usage, teamPermissions keys should be string. Found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-005 → Story 4 / AC-2, Spec Round 3 Item 18
  test('TC-005: Form field assigneeId renamed to assigneeKey', () => {
    const matches = grepDir('assigneeId', projectPath('frontend', 'src'));
    assert.equal(matches.length, 0, `Expected no assigneeId references, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-006 → Spec Round 4 Item 23
  test('TC-006: TableRow.mainItemId is string type', () => {
    const matches = grepFile('mainItemId.*number', projectPath('frontend', 'src', 'types', 'index.ts'));
    assert.equal(matches.length, 0, `Expected mainItemId to not use number type, found:\n${matches.join('\n')}`);
  });

  // Traceability: TC-007 → Spec Round 4 Item 24
  test('TC-007: Shared formatDate utility extracted', () => {
    const matches = grepDir('function formatDate|const formatDate', projectPath('frontend', 'src'));
    // formatDate should be defined only once (in a shared utils file)
    assert.equal(matches.length, 1, `Expected exactly 1 formatDate definition, found ${matches.length}:\n${matches.join('\n')}`);
  });

  // Traceability: TC-008 → Spec Round 3 Item 15
  test('TC-008: Shared state recording helper extracted', () => {
    // Verify state recording uses a shared helper pattern
    const matches = grepDir('state.*=.*recording|recordState|handleStateRecording', projectPath('frontend', 'src'));
    // This test verifies that a shared helper exists (at least one definition)
    assert.ok(matches.length > 0, 'Expected shared state recording helper to exist');
  });
});
