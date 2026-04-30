import { readFileSync } from 'node:fs';

export interface ValidationError {
  filePath: string;
  line: number;
  code: 'EXTERNAL_IMPORT' | 'MISSING_TRACEABILITY' | 'STALE_IMPORT_PATH';
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
}

// External test framework package names
const EXTERNAL_FRAMEWORKS = ['mocha', 'jest', 'chai', 'jasmine', 'vitest'];

// Matches: from 'mocha' / from "mocha" (any external framework)
const EXTERNAL_IMPORT_RE = new RegExp(
  `from\\s+['"](?:${EXTERNAL_FRAMEWORKS.join('|')})['"]`
);

// Matches a bare describe( or it( call — not preceded by a node:test import on the same line
const BARE_DESCRIBE_RE = /(?:^|[^a-zA-Z0-9_$])describe\s*\(/;
const BARE_IT_RE = /(?:^|[^a-zA-Z0-9_$])it\s*\(/;

// Matches: from '...testing/scripts/...' or from "...testing/scripts/..."
const STALE_IMPORT_RE = /from\s+['"][^'"]*testing\/scripts\//;

// Matches a test( call at the start of a statement (not inside a string)
const TEST_CALL_RE = /(?:^|[^a-zA-Z0-9_$])test\s*\(/;

// Matches a traceability comment
const TRACEABILITY_RE = /\/\/\s*Traceability:\s*TC-/;

export function validateSpec(filePath: string): ValidationResult {
  const source = readFileSync(filePath, 'utf8');
  const lines = source.split('\n');
  const errors: ValidationError[] = [];

  // Determine if the file imports from node:test (allows describe/it)
  const hasNodeTestImport = lines.some((line) => /from\s+['"]node:test['"]/.test(line));

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const lineNum = i + 1;

    // Check EXTERNAL_IMPORT: framework package imports
    if (EXTERNAL_IMPORT_RE.test(line)) {
      const matched = EXTERNAL_FRAMEWORKS.find((fw) =>
        new RegExp(`from\\s+['"]${fw}['"]`).test(line)
      );
      errors.push({
        filePath,
        line: lineNum,
        code: 'EXTERNAL_IMPORT',
        message: `External test framework import detected: '${matched}' at line ${lineNum}`,
      });
    }

    // Check EXTERNAL_IMPORT: bare describe() / it() without node:test import
    if (!hasNodeTestImport) {
      if (BARE_DESCRIBE_RE.test(line)) {
        errors.push({
          filePath,
          line: lineNum,
          code: 'EXTERNAL_IMPORT',
          message: `Bare describe() call without 'node:test' import at line ${lineNum}`,
        });
      }
      if (BARE_IT_RE.test(line)) {
        errors.push({
          filePath,
          line: lineNum,
          code: 'EXTERNAL_IMPORT',
          message: `Bare it() call without 'node:test' import at line ${lineNum}`,
        });
      }
    }

    // Check STALE_IMPORT_PATH
    if (STALE_IMPORT_RE.test(line)) {
      errors.push({
        filePath,
        line: lineNum,
        code: 'STALE_IMPORT_PATH',
        message: `Import from 'testing/scripts/' path detected at line ${lineNum}. Use '../../helpers.js' instead.`,
      });
    }

    // Check MISSING_TRACEABILITY: for each test() call, preceding 3 lines must have traceability comment
    if (TEST_CALL_RE.test(line)) {
      const start = Math.max(0, i - 3);
      const precedingLines = lines.slice(start, i);
      const hasTraceability = precedingLines.some((l) => TRACEABILITY_RE.test(l));
      if (!hasTraceability) {
        errors.push({
          filePath,
          line: lineNum,
          code: 'MISSING_TRACEABILITY',
          message: `test() at line ${lineNum} is missing a '// Traceability: TC-' comment in the preceding 3 lines`,
        });
      }
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}
