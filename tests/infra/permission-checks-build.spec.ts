import { test, expect } from 'vitest';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runCli } from '../shared/helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROJECT_ROOT = resolve(__dirname, '..', '..');
const FRONTEND_DIR = resolve(PROJECT_ROOT, 'frontend');

// ── TC-001 & TC-002: Build and test verification (no browser needed) ──

// Traceability: TC-001 → Story 5 / AC-1
test('TC-001: TypeScript compilation passes after isSuperAdmin removal', () => {
  const result = runCli('npx tsc --noEmit', FRONTEND_DIR);
  expect(result.exitCode).toBe(0);
  expect(result.stderr).not.toContain('isSuperAdmin');
});

// Traceability: TC-002 → Story 5 / AC-1
test('TC-002: Frontend tests pass after isSuperAdmin removal', () => {
  const result = runCli('npx vitest run', FRONTEND_DIR, 60000);
  expect(result.exitCode).toBe(0);
});
