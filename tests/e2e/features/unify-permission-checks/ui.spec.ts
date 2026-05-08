import { test, expect } from '@playwright/test';
import { runCli } from '../../helpers.js';

// ── TC-001 & TC-002: Build and test verification (no browser needed) ──

// Traceability: TC-001 → Story 5 / AC-1
test('TC-001: TypeScript compilation passes after isSuperAdmin removal', () => {
  const result = runCli('npx tsc --noEmit', 'ui');
  expect(result.exitCode).toBe(0);
  expect(result.stderr).not.toContain('isSuperAdmin');
});

// Traceability: TC-002 → Story 5 / AC-1
test('TC-002: Frontend tests pass after isSuperAdmin removal', () => {
  const result = runCli('npx vitest run', 'ui', 60000);
  expect(result.exitCode).toBe(0);
});
