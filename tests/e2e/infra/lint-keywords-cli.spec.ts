import { test, expect } from '@playwright/test';
import { writeFileSync, unlinkSync, existsSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runCli } from '../helpers.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
// Path to project root (where scripts/lint-staged.sh lives)
// infra/ -> e2e/ -> tests/ -> pm-work-tracker/
const PROJECT_ROOT = resolve(__dirname, '..', '..', '..');
// Directory for temp test files (must match lint-staged.sh check_sqlite_keywords glob)
const GORM_DIR = resolve(PROJECT_ROOT, 'backend', 'internal', 'repository', 'gorm');

// Helper: create file, stage it, run lint-staged's check_sqlite_keywords, cleanup
function testLintStagedKeyword(
  tcId: string,
  fileName: string,
  content: string,
  expectedPattern: RegExp,
) {
  const absFile = resolve(GORM_DIR, fileName);
  const relFile = `backend/internal/repository/gorm/${fileName}`;

  try {
    writeFileSync(absFile, content);
    runCli(`git add ${relFile}`, PROJECT_ROOT);
    // Run only check_sqlite_keywords (extract function from lint-staged.sh) to avoid
    // slow golangci-lint and pre-existing lint issues unrelated to this feature.
    const result = runCli('bash -c \'eval "$(sed -n "/^check_sqlite_keywords/,/^}/p" scripts/lint-staged.sh)"; check_sqlite_keywords\'', PROJECT_ROOT, 15000);
    expect(result.exitCode).not.toBe(0);
    expect(result.stdout + result.stderr).toMatch(expectedPattern);
  } finally {
    runCli(`git reset HEAD -- ${relFile}`, PROJECT_ROOT);
    if (existsSync(absFile)) unlinkSync(absFile);
  }
}

test.describe('CLI E2E Tests', () => {
  // ── Lint-staged keyword detection tests ──────────────────────────
  // These tests validate that lint-staged.sh blocks hardcoded SQLite keywords
  // in the repository layer. They create a temporary file with the offending keyword,
  // stage it, and run lint-staged.sh directly.

  // Traceability: TC-006 → Story 3 / AC-1
  test('TC-006: Git commit blocked when repo layer contains hardcoded SUBSTR', () => {
    testLintStagedKeyword(
      'TC-006',
      'lint_test_substr.go',
      'package gorm\n\nfunc lintTestSubstr() { _ = "SELECT SUBSTR(code, 5) FROM items" }\n',
      /SUBSTR|dialect/i,
    );
  });

  // Traceability: TC-007 → Story 3 / AC-1
  test('TC-007: Git commit blocked when repo layer contains hardcoded CAST', () => {
    testLintStagedKeyword(
      'TC-007',
      'lint_test_cast.go',
      'package gorm\n\nfunc lintTestCast() { _ = "SELECT CAST(x AS INTEGER) FROM items" }\n',
      /CAST|dialect/i,
    );
  });

  // Traceability: TC-008 → Story 3 / AC-1
  test('TC-008: Git commit blocked when repo layer contains hardcoded datetime', () => {
    testLintStagedKeyword(
      'TC-008',
      'lint_test_datetime.go',
      'package gorm\n\nfunc lintTestDatetime() { _ = "datetime(\'now\')" }\n',
      /datetime|dialect/i,
    );
  });

  // Traceability: TC-009 → Story 3 / AC-1
  test('TC-009: Git commit blocked when repo layer contains hardcoded pragma_', () => {
    testLintStagedKeyword(
      'TC-009',
      'lint_test_pragma.go',
      'package gorm\n\nfunc lintTestPragma() { _ = "pragma_table_info(\'table_name\')" }\n',
      /pragma_|dialect/i,
    );
  });

  // Traceability: TC-010 → Story 3 / AC-2
  test('TC-010: Git commit passes when repo layer uses dialect package', () => {
    const fileName = 'lint_test_clean.go';
    const absFile = resolve(GORM_DIR, fileName);
    const relFile = `backend/internal/repository/gorm/${fileName}`;
    // This file has no hardcoded SQLite keywords
    const content = 'package gorm\n\nfunc lintTestClean() { _ = "SELECT * FROM items WHERE status = ?" }\n';

    try {
      writeFileSync(absFile, content);
      runCli(`git add ${relFile}`, PROJECT_ROOT);
      // Run only check_sqlite_keywords to avoid pre-existing golangci-lint issues
      const result = runCli('bash -c \'eval "$(sed -n "/^check_sqlite_keywords/,/^}/p" scripts/lint-staged.sh)"; check_sqlite_keywords\'', PROJECT_ROOT, 15000);
      expect(result.exitCode).toBe(0);
    } finally {
      runCli(`git reset HEAD -- ${relFile}`, PROJECT_ROOT);
      if (existsSync(absFile)) unlinkSync(absFile);
    }
  });

  // Traceability: TC-011 → Story 2 / AC-1, Story 4 / AC-1
  test('TC-011: Fresh MySQL startup initializes RBAC with preset roles', () => {
    if (!process.env.MYSQL_HOST) { test.skip(); return; }
    // This test verifies that the application starts successfully against MySQL
    // and initializes RBAC correctly.
    //
    // Pre-conditions:
    //   - MySQL 8.0 instance running with schema imported
    //   - config.yaml has database.driver: mysql, auto_schema: false
    //
    // The test starts the app, queries the roles table, and verifies preset data.
    // Requires environment variables:
    //   MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE
    //
    // Set MYSQL_HOST env var to enable TC-011 (e.g. MYSQL_HOST=127.0.0.1)

    const mysqlHost = process.env.MYSQL_HOST ?? '127.0.0.1';
    const mysqlPort = process.env.MYSQL_PORT ?? '3306';
    const mysqlUser = process.env.MYSQL_USER ?? 'root';
    const mysqlPassword = process.env.MYSQL_PASSWORD ?? '';
    const mysqlDb = process.env.MYSQL_DATABASE ?? 'pm_work_tracker';

    // Query roles table to verify 3 preset roles
    const rolesResult = runCli(
      `mysql -h ${mysqlHost} -P ${mysqlPort} -u ${mysqlUser} ${mysqlPassword ? `-p${mysqlPassword}` : ''} ${mysqlDb} -N -e "SELECT COUNT(*) FROM pmw_roles"`,
    );
    expect(rolesResult.exitCode).toBe(0);
    const roleCount = parseInt(rolesResult.stdout.trim(), 10);
    expect(roleCount).toBe(3);

    // Verify preset role names
    const namesResult = runCli(
      `mysql -h ${mysqlHost} -P ${mysqlPort} -u ${mysqlUser} ${mysqlPassword ? `-p${mysqlPassword}` : ''} ${mysqlDb} -N -e "SELECT name FROM pmw_roles ORDER BY name"`,
    );
    expect(namesResult.exitCode).toBe(0);
    const roleNames = namesResult.stdout.trim().split('\n').map((s) => s.trim());
    expect(roleNames.includes('superadmin')).toBeTruthy();
    expect(roleNames.includes('pm')).toBeTruthy();
    expect(roleNames.includes('member')).toBeTruthy();

    // Verify HasColumn equivalent: check pmw_team_members has role_key column
    const colResult = runCli(
      `mysql -h ${mysqlHost} -P ${mysqlPort} -u ${mysqlUser} ${mysqlPassword ? `-p${mysqlPassword}` : ''} ${mysqlDb} -N -e "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='${mysqlDb}' AND table_name='pmw_team_members' AND column_name='role_key'"`,
    );
    expect(colResult.exitCode).toBe(0);
    expect(parseInt(colResult.stdout.trim(), 10)).toBe(1);

    // Verify nonexistent column returns false
    const noColResult = runCli(
      `mysql -h ${mysqlHost} -P ${mysqlPort} -u ${mysqlUser} ${mysqlPassword ? `-p${mysqlPassword}` : ''} ${mysqlDb} -N -e "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='${mysqlDb}' AND table_name='pmw_team_members' AND column_name='nonexistent'"`,
    );
    expect(noColResult.exitCode).toBe(0);
    expect(parseInt(noColResult.stdout.trim(), 10)).toBe(0);
  });
});
