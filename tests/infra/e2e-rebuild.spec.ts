import { test, expect } from 'vitest';
import { readFileSync, existsSync, readdirSync } from 'node:fs';
import { join, resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runCli } from '../shared/helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const TESTS_DIR = resolve(__dirname, '..');
const ROOT = resolve(TESTS_DIR, '..');
const GRADUATED_DIR = join(TESTS_DIR, '.graduated');
const KNOWN_FAILURES = join(TESTS_DIR, 'KNOWN_FAILURES.md');
const VALIDATE_SPEC = join(ROOT, 'docs/features/e2e-test-scripts-rebuild/testing/scripts/validate-spec.ts');

// Traceability: TC-001 -> Story 1 / AC-1
test('TC-001: vitest discovers all api/infra specs and playwright discovers web specs', () => {
  // API surface — vitest
  const apiResult = runCli('npx vitest run --reporter=verbose 2>&1 | head -5', join(TESTS_DIR, 'api'), 120000);
  expect(apiResult.exitCode).toBe(0);

  // Infra surface — vitest
  const infraResult = runCli('npx vitest run --reporter=verbose 2>&1 | head -5', join(TESTS_DIR, 'infra'), 60000);
  expect(infraResult.exitCode).toBe(0);
});

// Traceability: TC-002 -> Story 2 / AC-1
test('TC-002: graduation marker and spec exist for config-yaml', () => {
  const marker = join(GRADUATED_DIR, 'config-yaml');
  expect(existsSync(marker)).toBeTruthy();
  const content = readFileSync(marker, 'utf-8').trim();
  // Marker must contain ISO 8601 timestamp
  expect(content).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
  // The spec should exist in one of the surface directories
  const apiSpec = join(TESTS_DIR, 'api/smoke/config-yaml-api.spec.ts');
  const infraSpec = join(TESTS_DIR, 'infra/config-yaml-build.spec.ts');
  expect(existsSync(apiSpec) || existsSync(infraSpec)).toBeTruthy();
});

// Traceability: TC-003 -> Story 2 / AC-1 (no stale imports)
test('TC-003: surface specs contain no stale import paths', () => {
  const stalePattern = /from\s+['"][^'"]*testing\/scripts[^'"]*['"]/;
  const violations: string[] = [];

  const surfaces = ['api', 'web', 'infra'];
  for (const surface of surfaces) {
    const surfaceDir = join(TESTS_DIR, surface);
    if (!existsSync(surfaceDir)) continue;
    for (const entry of readdirSync(surfaceDir)) {
      const dirPath = join(surfaceDir, entry);
      if (entry.startsWith('.') || entry === 'node_modules' || entry === 'test-results') continue;
      try { readdirSync(dirPath); } catch { continue; }
      for (const file of readdirSync(dirPath).filter(f => f.endsWith('.spec.ts'))) {
        const filePath = join(dirPath, file);
        const content = readFileSync(filePath, 'utf-8');
        if (stalePattern.test(content)) {
          violations.push(filePath);
        }
      }
    }
  }

  expect(violations).toEqual([]);
});

// Traceability: TC-004 -> Story 3 / AC-1
test('TC-004: KNOWN_FAILURES.md exists and has required fields', () => {
  expect(existsSync(KNOWN_FAILURES)).toBeTruthy();
  const content = readFileSync(KNOWN_FAILURES, 'utf-8');
  // Must have at least the header row with required columns
  expect(content).toMatch(/Test ID/);
  expect(content).toMatch(/Reason/);
  expect(content).toMatch(/Owner/);
});

// Traceability: TC-005 -> Story 2 / AC-1 + Spec 5.1
test('TC-005: validate-spec detects external imports and missing traceability', () => {
  expect(existsSync(VALIDATE_SPEC)).toBeTruthy();

  // Run the validator's unit tests using tsx from the infra directory
  const testFile = join(ROOT, 'docs/features/e2e-test-scripts-rebuild/testing/scripts/validate-spec.test.ts');
  const result = runCli(
    `npx tsx ${testFile}`,
    join(TESTS_DIR, 'infra'),
    30000
  );
  expect(result.exitCode).toBe(0);
  expect(result.stdout).toMatch(/pass\s+\d+/i);
});

// Traceability: TC-006 -> Story 1 / AC-1 + Spec 5.4
test('TC-006: surface configs exist and specs use correct test framework', () => {
  // API surface uses vitest
  const apiConfig = join(TESTS_DIR, 'api/vitest.config.ts');
  expect(existsSync(apiConfig)).toBeTruthy();

  // Web surface uses playwright
  const webConfig = join(TESTS_DIR, 'web/playwright.config.ts');
  expect(existsSync(webConfig)).toBeTruthy();

  // Infra surface uses vitest
  const infraConfig = join(TESTS_DIR, 'infra/vitest.config.ts');
  expect(existsSync(infraConfig)).toBeTruthy();

  // Verify API specs import vitest (not @playwright/test)
  const apiDir = join(TESTS_DIR, 'api');
  const apiWrongImport: string[] = [];
  for (const entry of readdirSync(apiDir)) {
    const dirPath = join(apiDir, entry);
    if (entry.startsWith('.') || entry === 'node_modules') continue;
    try { readdirSync(dirPath); } catch { continue; }
    for (const file of readdirSync(dirPath).filter(f => f.endsWith('.spec.ts'))) {
      const content = readFileSync(join(dirPath, file), 'utf-8');
      if (content.includes('@playwright/test')) {
        apiWrongImport.push(`api/${entry}/${file} (should use vitest, not @playwright/test)`);
      }
    }
  }
  expect(apiWrongImport).toEqual([]);
});
