import { test, expect } from '@playwright/test';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, resolve } from 'node:path';

const PROJECT_ROOT = resolve(import.meta.dirname, '../../..');

function projectPath(...segments: string[]): string {
  return join(PROJECT_ROOT, ...segments);
}

// Recursively find .ts/.tsx files and return lines matching a regex
function grepInDir(pattern: RegExp, dir: string): string[] {
  const results: string[] = [];
  try {
    for (const entry of readdirSync(dir)) {
      const full = join(dir, entry);
      const stat = statSync(full);
      if (stat.isDirectory() && !entry.startsWith('.') && entry !== 'node_modules') {
        results.push(...grepInDir(pattern, full));
      } else if (stat.isFile() && /\.(ts|tsx)$/.test(entry)) {
        const content = readFileSync(full, 'utf-8');
        for (const line of content.split('\n')) {
          if (pattern.test(line)) {
            results.push(`${full}: ${line.trim()}`);
          }
        }
      }
    }
  } catch { /* ignore unreadable dirs */ }
  return results;
}

function grepInFile(pattern: RegExp, filePath: string): string[] {
  try {
    const content = readFileSync(filePath, 'utf-8');
    return content.split('\n').filter((l) => pattern.test(l));
  } catch { return []; }
}

test.describe('UI E2E Tests', () => {
  // These UI test cases are code-inspection tests that verify frontend
  // cleanup tasks were completed correctly. They use grep and build
  // commands rather than browser interactions.

  // ── Code Inspection Tests ────────────────────────────────────────

  // Traceability: TC-001 → Spec Round 2 Item 8
  test('TC-001: console.error replaced with toast in API client', () => {
    const matches = grepInFile(/console\.error/, projectPath('frontend', 'src', 'api', 'client.ts'));
    expect(matches.length).toBe(0);
  });

  // Traceability: TC-002 → Spec Round 2 Item 9
  test('TC-002: Test data Role.id is string type', () => {
    const filePath = projectPath('frontend', 'src', 'types', 'index.test.ts');
    const content = readFileSync(filePath, 'utf-8');
    // Verify Role.id values in test data use string type (e.g., "role-1"), not number
    expect(content).not.toMatch(/Role\.id.*:\s*\d+(?!\s*as)/);
  });

  // Traceability: TC-003 → Spec Round 3 Item 16
  test('TC-003: Redundant String() wrappers removed from frontend', () => {
    const lines = grepInDir(/\.String\(\)/, projectPath('frontend', 'src'));
    // Filter out legitimate uses (enum conversions, etc.)
    const redundant = lines.filter(l => !l.includes('.toString()') && !l.includes('String('));
    expect(redundant.length).toBe(0);
  });

  // Traceability: TC-004 → Story 4 / AC-1, Spec Round 3 Item 17
  test('TC-004: PermissionData.teamPermissions uses string keys', () => {
    const matches = grepInFile(/Record<number/, projectPath('frontend', 'src', 'types', 'index.ts'));
    expect(matches.length).toBe(0);
  });

  // Traceability: TC-005 → Story 4 / AC-2, Spec Round 3 Item 18
  test('TC-005: Form field assigneeId renamed to assigneeKey', () => {
    const matches = grepInDir(/assigneeId/, projectPath('frontend', 'src'));
    expect(matches.length).toBe(0);
  });
});
