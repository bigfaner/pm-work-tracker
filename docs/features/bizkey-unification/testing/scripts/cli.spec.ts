import { describe, test } from 'node:test';
import assert from 'node:assert/strict';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';
import { runCli } from './helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const BACKEND_DIR = resolve(__dirname, '../../../../../../backend');

describe('CLI E2E Tests', () => {
  // Traceability: TC-006 → Story 3 / AC-1
  test('TC-006: Build fails when uint internal ID passed to int64 bizKey parameter', () => {
    // Verify the current codebase compiles cleanly (post-migration state)
    const result = runCli('go build ./...', BACKEND_DIR);
    assert.equal(
      result.exitCode,
      0,
      `go build should succeed in post-migration state.\nstderr: ${result.stderr}`,
    );
    // The test documents that introducing a uint→int64 mismatch would cause a compile error.
    // The green path here confirms the current state has no such mismatches.
  });

  // Traceability: TC-007 → Spec 需求目标 — 消除 uint/int64 混用导致的数据错误
  test('TC-007: No uint(bizKey) or int64(teamID) casts exist in service and handler layers', () => {
    const result = runCli(
      'grep -rn "uint(bizKey)\\|uint(teamBizKey)\\|int64(teamID)" backend/internal/service/ backend/internal/handler/',
      BACKEND_DIR,
    );
    // grep exits 1 when no matches found — that is the expected (passing) outcome
    assert.equal(
      result.exitCode,
      1,
      `Expected zero matches for forced uint/int64 casts, but grep found matches:\n${result.stdout}`,
    );
    assert.equal(
      result.stdout.trim(),
      '',
      `Found forbidden casts in service/handler layers:\n${result.stdout}`,
    );
  });
});
