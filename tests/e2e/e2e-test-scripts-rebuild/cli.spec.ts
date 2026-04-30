import { test, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync, readdirSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { runCli } from '../helpers.js';

const ROOT = resolve(new URL('../..', import.meta.url).pathname);
const E2E_DIR = join(ROOT, 'tests/e2e');
const GRADUATED_DIR = join(E2E_DIR, '.graduated');
const KNOWN_FAILURES = join(E2E_DIR, 'KNOWN_FAILURES.md');
const VALIDATE_SPEC = join(ROOT, 'docs/features/e2e-test-scripts-rebuild/testing/scripts/validate-spec.ts');

// Traceability: TC-001 → Story 1 / AC-1
test('TC-001: npm test exits 0 after all features graduated', () => {
  const result = runCli('npm test', E2E_DIR, 120000);
  // All failures must be documented in KNOWN_FAILURES.md — exit 0 is the target
  assert.equal(result.exitCode, 0,
    `npm test failed with exit code ${result.exitCode}.\nstdout: ${result.stdout}\nstderr: ${result.stderr}`);
  // Verify per-file output present
  assert.match(result.stdout, /passed|failed|skipped/i,
    'Expected per-file test counts in stdout');
});

// Traceability: TC-002 → Story 2 / AC-1
test('TC-002: graduation marker and spec exist for api-permission-test-coverage', () => {
  const marker = join(GRADUATED_DIR, 'api-permission-test-coverage');
  assert.ok(existsSync(marker),
    `Graduation marker missing: ${marker}`);
  const content = readFileSync(marker, 'utf-8').trim();
  // Marker must contain ISO 8601 timestamp
  assert.match(content, /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/,
    `Marker content is not an ISO 8601 timestamp: "${content}"`);
  const spec = join(E2E_DIR, 'api/api-permission-test-coverage/api.spec.ts');
  assert.ok(existsSync(spec),
    `Graduated spec missing: ${spec}`);
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

  assert.deepEqual(violations, [],
    `Graduated specs with stale testing/scripts/ imports:\n${violations.join('\n')}`);
});

// Traceability: TC-004 → Story 3 / AC-1
test('TC-004: KNOWN_FAILURES.md exists and has required fields', () => {
  assert.ok(existsSync(KNOWN_FAILURES),
    `KNOWN_FAILURES.md missing at ${KNOWN_FAILURES}`);
  const content = readFileSync(KNOWN_FAILURES, 'utf-8');
  // Must have at least the header row with required columns
  assert.match(content, /Test ID/,
    'KNOWN_FAILURES.md missing "Test ID" column header');
  assert.match(content, /Description/,
    'KNOWN_FAILURES.md missing "Description" column header');
  assert.match(content, /Reason/,
    'KNOWN_FAILURES.md missing "Reason" column header');
  assert.match(content, /Owner/,
    'KNOWN_FAILURES.md missing "Owner" column header');
});

// Traceability: TC-005 → Story 2 / AC-1 + Spec §5.1
test('TC-005: validate-spec detects external imports and missing traceability', () => {
  assert.ok(existsSync(VALIDATE_SPEC),
    `validate-spec.ts not found at ${VALIDATE_SPEC}`);

  // Run the validator's unit tests to confirm it works correctly
  const result = runCli(
    `node --import tsx/esm --test ${VALIDATE_SPEC}`,
    ROOT,
    30000
  );
  assert.equal(result.exitCode, 0,
    `validate-spec.ts unit tests failed:\n${result.stdout}\n${result.stderr}`);
  assert.match(result.stdout, /passed/i,
    'Expected "passed" in validate-spec test output');
});

// Traceability: TC-006 → Story 1 / AC-1 + Spec §5.4
test('TC-006: package.json includes all graduated spec paths', () => {
  const pkgPath = join(E2E_DIR, 'package.json');
  assert.ok(existsSync(pkgPath), `package.json missing at ${pkgPath}`);

  const pkg = JSON.parse(readFileSync(pkgPath, 'utf-8'));
  const scripts: Record<string, string> = pkg.scripts ?? {};

  // test:api must exist
  assert.ok(scripts['test:api'],
    'package.json missing "test:api" script');

  // test:cli must exist
  assert.ok(scripts['test:cli'],
    'package.json missing "test:cli" script');

  // test must chain test:cli and test:api
  assert.ok(scripts['test'],
    'package.json missing "test" script');
  assert.match(scripts['test'], /test:cli/,
    '"test" script must include "test:cli"');
  assert.match(scripts['test'], /test:api/,
    '"test" script must include "test:api"');

  // Verify all spec paths in scripts point to existing files
  const allScripts = Object.values(scripts).join(' ');
  const specPaths = allScripts.match(/\S+\.spec\.ts/g) ?? [];
  const missing: string[] = [];
  for (const rel of specPaths) {
    const abs = join(E2E_DIR, rel);
    if (!existsSync(abs)) missing.push(rel);
  }
  assert.deepEqual(missing, [],
    `Spec paths in package.json point to non-existent files:\n${missing.join('\n')}`);
});
