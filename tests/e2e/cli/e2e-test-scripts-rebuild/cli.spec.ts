import { test, expect } from '@playwright/test';
import { readFileSync, existsSync, readdirSync } from 'node:fs';
import { join, resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runCli } from '../../helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const E2E_DIR = resolve(__dirname, '../..');
const ROOT = resolve(E2E_DIR, '../..');
const GRADUATED_DIR = join(E2E_DIR, '.graduated');
const KNOWN_FAILURES = join(E2E_DIR, 'KNOWN_FAILURES.md');
const VALIDATE_SPEC = join(ROOT, 'docs/features/e2e-test-scripts-rebuild/testing/scripts/validate-spec.ts');

// Traceability: TC-001 → Story 1 / AC-1
test('TC-001: npx playwright test discovers all graduated specs', () => {
  const result = runCli('npx playwright test --list', E2E_DIR, 120000);
  expect(result.exitCode).toBe(0);
  // Should list tests from graduated features
  expect(result.stdout).toMatch(/\d+ tests?/i);
});

// Traceability: TC-002 → Story 2 / AC-1
test('TC-002: graduation marker and spec exist for api-permission-test-coverage', () => {
  const marker = join(GRADUATED_DIR, 'api-permission-test-coverage');
  expect(existsSync(marker)).toBeTruthy();
  const content = readFileSync(marker, 'utf-8').trim();
  // Marker must contain ISO 8601 timestamp
  expect(content).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
  const spec = join(E2E_DIR, 'api/api-permission-test-coverage/api.spec.ts');
  expect(existsSync(spec)).toBeTruthy();
});

// Traceability: TC-003 → Story 2 / AC-1 (no stale imports)
test('TC-003: graduated specs contain no stale import paths', () => {
  const stalePattern = /from\s+['"][^'"]*testing\/scripts[^'"]*['"]/;
  const specDirs = ['api', 'ui', 'cli'];
  const violations: string[] = [];

  for (const dir of specDirs) {
    const dirPath = join(E2E_DIR, dir);
    if (!existsSync(dirPath)) continue;
    for (const slug of readdirSync(dirPath)) {
      const slugPath = join(dirPath, slug);
      for (const file of readdirSync(slugPath).filter(f => f.endsWith('.spec.ts'))) {
        const filePath = join(slugPath, file);
        const content = readFileSync(filePath, 'utf-8');
        if (stalePattern.test(content)) {
          violations.push(filePath);
        }
      }
    }
  }

  expect(violations).toEqual([]);
});

// Traceability: TC-004 → Story 3 / AC-1
test('TC-004: KNOWN_FAILURES.md exists and has required fields', () => {
  expect(existsSync(KNOWN_FAILURES)).toBeTruthy();
  const content = readFileSync(KNOWN_FAILURES, 'utf-8');
  // Must have at least the header row with required columns
  expect(content).toMatch(/Test ID/);
  expect(content).toMatch(/Reason/);
  expect(content).toMatch(/Owner/);
});

// Traceability: TC-005 → Story 2 / AC-1 + Spec §5.1
test('TC-005: validate-spec detects external imports and missing traceability', () => {
  expect(existsSync(VALIDATE_SPEC)).toBeTruthy();

  // Run the validator's unit tests from tests/e2e where tsx is available
  const testFile = join(ROOT, 'docs/features/e2e-test-scripts-rebuild/testing/scripts/validate-spec.test.ts');
  const result = runCli(
    `npx --prefix tests/e2e tsx ${testFile}`,
    ROOT,
    30000
  );
  expect(result.exitCode).toBe(0);
  expect(result.stdout).toMatch(/pass\s+\d+/i);
});

// Traceability: TC-006 → Story 1 / AC-1 + Spec §5.4
test('TC-006: playwright.config.ts exists and all graduated specs are discoverable', () => {
  const configPath = join(E2E_DIR, 'playwright.config.ts');
  expect(existsSync(configPath)).toBeTruthy();

  const pkgPath = join(E2E_DIR, 'package.json');
  const pkg = JSON.parse(readFileSync(pkgPath, 'utf-8'));
  expect(pkg.devDependencies['@playwright/test']).toBeTruthy();

  // Verify all spec files under graduated markers are discoverable
  const specDirs = ['api', 'ui', 'cli'];
  const missing: string[] = [];

  for (const dir of specDirs) {
    const dirPath = join(E2E_DIR, dir);
    if (!existsSync(dirPath)) continue;
    for (const slug of readdirSync(dirPath)) {
      const marker = join(GRADUATED_DIR, slug);
      if (!existsSync(marker)) continue;
      const slugPath = join(dirPath, slug);
      if (!existsSync(slugPath)) continue;
      const specs = readdirSync(slugPath).filter(f => f.endsWith('.spec.ts'));
      if (specs.length === 0) missing.push(`${dir}/${slug}/ (no spec files)`);
      for (const spec of specs) {
        const content = readFileSync(join(slugPath, spec), 'utf-8');
        if (!content.includes('@playwright/test')) {
          missing.push(`${dir}/${slug}/${spec} (missing @playwright/test import)`);
        }
      }
    }
  }

  expect(missing).toEqual([]);
});
