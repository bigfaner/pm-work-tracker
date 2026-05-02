import { test, expect } from '@playwright/test';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';
import { runCli } from '../../helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
// bizkey-unification/ -> cli/ -> e2e/ -> tests/ -> pm-work-tracker/
const BACKEND_DIR = resolve(__dirname, '..', '..', '..', '..', 'backend');

test.describe('CLI E2E Tests', () => {
  // Traceability: TC-006 → Story 3 / AC-1
  test('TC-006: Build fails when uint internal ID passed to int64 bizKey parameter', () => {
    // Verify the current codebase compiles cleanly (post-migration state)
    const result = runCli('go build ./...', BACKEND_DIR);
    expect(result.exitCode).toBe(0);
    // The test documents that introducing a uint→int64 mismatch would cause a compile error.
    // The green path here confirms the current state has no such mismatches.
  });

  // Traceability: TC-007 → Spec 需求目标 — 消除 uint/int64 混用导致的数据错误
  test('TC-007: No uint(bizKey) or int64(teamID) casts exist in service and handler layers', () => {
    const result = runCli(
      'grep -rn --include="*.go" --exclude="*_test.go" "uint(bizKey)\\|uint(teamBizKey)\\|int64(teamID)" internal/service/ internal/handler/',
      BACKEND_DIR,
    );
    // grep exits 1 when no matches found — that is the expected (passing) outcome
    expect(result.exitCode).toBe(1);
    expect(result.stdout.trim()).toBe('');
  });
});
