import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import { writeFileSync, unlinkSync, mkdtempSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { validateSpec } from './validate-spec.js';

// Helper: write a temp spec file and return its path
function writeTempSpec(content: string): string {
  const dir = mkdtempSync(join(tmpdir(), 'validate-spec-'));
  const filePath = join(dir, 'test.spec.ts');
  writeFileSync(filePath, content, 'utf8');
  return filePath;
}

const CONFORMANT_SPEC = `
import { test, describe } from 'node:test';
import assert from 'node:assert/strict';

describe('Suite', () => {
  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: does something', async () => {
    assert.ok(true);
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: does another thing', async () => {
    assert.ok(true);
  });
});
`;

describe('validateSpec', () => {
  // Traceability: TC-VAL-001 → happy path
  test('TC-VAL-001: returns valid=true and empty errors for a conformant spec', () => {
    const filePath = writeTempSpec(CONFORMANT_SPEC);
    const result = validateSpec(filePath);
    assert.equal(result.valid, true);
    assert.deepEqual(result.errors, []);
  });

  // Traceability: TC-VAL-002 → EXTERNAL_IMPORT from mocha
  test('TC-VAL-002: detects EXTERNAL_IMPORT for mocha import', () => {
    const spec = `import { describe, it } from 'mocha';\ntest('x', () => {});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    const codes = result.errors.map((e) => e.code);
    assert.ok(codes.includes('EXTERNAL_IMPORT'), `Expected EXTERNAL_IMPORT, got: ${JSON.stringify(codes)}`);
  });

  // Traceability: TC-VAL-003 → EXTERNAL_IMPORT from jest
  test('TC-VAL-003: detects EXTERNAL_IMPORT for jest import', () => {
    const spec = `import { expect } from 'jest';\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-004 → EXTERNAL_IMPORT from chai
  test('TC-VAL-004: detects EXTERNAL_IMPORT for chai import', () => {
    const spec = `import { expect } from 'chai';\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-005 → EXTERNAL_IMPORT from jasmine
  test('TC-VAL-005: detects EXTERNAL_IMPORT for jasmine import', () => {
    const spec = `import jasmine from 'jasmine';\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-006 → EXTERNAL_IMPORT from vitest
  test('TC-VAL-006: detects EXTERNAL_IMPORT for vitest import', () => {
    const spec = `import { describe } from 'vitest';\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-007 → EXTERNAL_IMPORT bare describe() without node:test
  test('TC-VAL-007: detects EXTERNAL_IMPORT for bare describe() without node:test import', () => {
    const spec = `describe('suite', () => {\n  it('test', () => {});\n});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-008 → EXTERNAL_IMPORT bare it() without node:test
  test('TC-VAL-008: detects EXTERNAL_IMPORT for bare it() without node:test import', () => {
    const spec = `it('does something', () => {});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'EXTERNAL_IMPORT'));
  });

  // Traceability: TC-VAL-009 → describe() with node:test is allowed
  test('TC-VAL-009: allows describe() when imported from node:test', () => {
    const spec = `import { describe, test } from 'node:test';\nimport assert from 'node:assert/strict';\n\ndescribe('suite', () => {\n  // Traceability: TC-001 → Story 1 / AC-1\n  test('TC-001: ok', () => { assert.ok(true); });\n});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, true);
    assert.deepEqual(result.errors, []);
  });

  // Traceability: TC-VAL-010 → MISSING_TRACEABILITY
  test('TC-VAL-010: detects MISSING_TRACEABILITY when test() has no traceability comment', () => {
    const spec = `import { test } from 'node:test';\nimport assert from 'node:assert/strict';\n\ntest('TC-001: missing traceability', () => {\n  assert.ok(true);\n});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'MISSING_TRACEABILITY'));
  });

  // Traceability: TC-VAL-011 → MISSING_TRACEABILITY with comment too far away
  test('TC-VAL-011: detects MISSING_TRACEABILITY when traceability comment is more than 3 lines before test()', () => {
    const spec = `import { test } from 'node:test';\nimport assert from 'node:assert/strict';\n\n// Traceability: TC-001 → Story 1 / AC-1\n\n\n\ntest('TC-001: too far', () => {\n  assert.ok(true);\n});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'MISSING_TRACEABILITY'));
  });

  // Traceability: TC-VAL-012 → STALE_IMPORT_PATH
  test('TC-VAL-012: detects STALE_IMPORT_PATH for imports from testing/scripts/', () => {
    const spec = `import { helper } from '../../testing/scripts/helpers.js';\nimport { test } from 'node:test';\n\n// Traceability: TC-001 → Story 1 / AC-1\ntest('TC-001: ok', () => {});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    assert.ok(result.errors.some((e) => e.code === 'STALE_IMPORT_PATH'));
  });

  // Traceability: TC-VAL-013 → error includes filePath and line number
  test('TC-VAL-013: errors include correct filePath and line number', () => {
    const spec = `import { expect } from 'chai';\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    const err = result.errors[0];
    assert.equal(err.filePath, filePath);
    assert.equal(err.line, 1);
  });

  // Traceability: TC-VAL-014 → multiple errors reported together
  test('TC-VAL-014: reports multiple error codes in a single result', () => {
    const spec = `import { expect } from 'chai';\nimport { helper } from '../../testing/scripts/helpers.js';\ntest('no traceability', () => {});\n`;
    const filePath = writeTempSpec(spec);
    const result = validateSpec(filePath);
    assert.equal(result.valid, false);
    const codes = result.errors.map((e) => e.code);
    assert.ok(codes.includes('EXTERNAL_IMPORT'));
    assert.ok(codes.includes('STALE_IMPORT_PATH'));
    assert.ok(codes.includes('MISSING_TRACEABILITY'));
  });
});
