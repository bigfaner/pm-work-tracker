import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import { writeFileSync, readFileSync, mkdtempSync, chmodSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { updatePackageJson } from './update-package-json.js';

function writeTempPkg(content: object): string {
  const dir = mkdtempSync(join(tmpdir(), 'update-pkg-'));
  const pkgPath = join(dir, 'package.json');
  writeFileSync(pkgPath, JSON.stringify(content, null, 2) + '\n', 'utf8');
  return pkgPath;
}

function readPkg(pkgPath: string): Record<string, unknown> {
  return JSON.parse(readFileSync(pkgPath, 'utf8'));
}

const BASE_PKG = {
  name: 'e2e-tests',
  private: true,
  type: 'module',
  scripts: {
    'test:cli': 'node --import tsx/esm --test cli/startup/cli.spec.ts',
    'test:api': 'node --import tsx/esm --test api/main-items/api.spec.ts',
    test: 'npm run test:cli && npm run test:api',
  },
};

describe('updatePackageJson', () => {
  // Traceability: TC-UPJ-001 → merge new API paths into test:api
  test('TC-UPJ-001: merges new API paths into existing test:api script', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: ['api/bizkey-unification/api.spec.ts'], ui: [], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok(pkg.scripts['test:api'].includes('api/main-items/api.spec.ts'));
    assert.ok(pkg.scripts['test:api'].includes('api/bizkey-unification/api.spec.ts'));
  });

  // Traceability: TC-UPJ-002 → merge new CLI paths into test:cli
  test('TC-UPJ-002: merges new CLI paths into existing test:cli script', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: [], ui: [], cli: ['cli/config-yaml/cli.spec.ts'] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok(pkg.scripts['test:cli'].includes('cli/startup/cli.spec.ts'));
    assert.ok(pkg.scripts['test:cli'].includes('cli/config-yaml/cli.spec.ts'));
  });

  // Traceability: TC-UPJ-003 → add test:ui when it doesn't exist
  test('TC-UPJ-003: adds test:ui script when it does not exist and ui paths are provided', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: [], ui: ['ui/improve-ui/ui.spec.ts'], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok('test:ui' in pkg.scripts);
    assert.ok((pkg.scripts['test:ui'] as string).includes('ui/improve-ui/ui.spec.ts'));
  });

  // Traceability: TC-UPJ-004 → update test script to include npm run test:ui
  test('TC-UPJ-004: updates test script to include npm run test:ui when UI specs added', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: [], ui: ['ui/improve-ui/ui.spec.ts'], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok((pkg.scripts['test'] as string).includes('npm run test:ui'));
  });

  // Traceability: TC-UPJ-005 → does not duplicate existing paths
  test('TC-UPJ-005: does not duplicate paths already present in existing scripts', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: ['api/main-items/api.spec.ts'], ui: [], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    const parts = (pkg.scripts['test:api'] as string).split(' && ');
    const count = parts.filter((p) => p.includes('api/main-items/api.spec.ts')).length;
    assert.equal(count, 1);
  });

  // Traceability: TC-UPJ-006 → does not add npm run test:ui twice
  test('TC-UPJ-006: does not add npm run test:ui to test script if already present', () => {
    const pkgWithUi = {
      ...BASE_PKG,
      scripts: {
        ...BASE_PKG.scripts,
        'test:ui': 'node --import tsx/esm --test ui/existing/ui.spec.ts',
        test: 'npm run test:cli && npm run test:api && npm run test:ui',
      },
    };
    const pkgPath = writeTempPkg(pkgWithUi);
    updatePackageJson({ api: [], ui: ['ui/new-feature/ui.spec.ts'], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    const testScript = pkg.scripts['test'] as string;
    const occurrences = testScript.split('npm run test:ui').length - 1;
    assert.equal(occurrences, 1);
  });

  // Traceability: TC-UPJ-007 → merge into existing test:ui
  test('TC-UPJ-007: merges new UI paths into existing test:ui script', () => {
    const pkgWithUi = {
      ...BASE_PKG,
      scripts: {
        ...BASE_PKG.scripts,
        'test:ui': 'node --import tsx/esm --test ui/existing/ui.spec.ts',
        test: 'npm run test:cli && npm run test:api && npm run test:ui',
      },
    };
    const pkgPath = writeTempPkg(pkgWithUi);
    updatePackageJson({ api: [], ui: ['ui/new-feature/ui.spec.ts'], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok((pkg.scripts['test:ui'] as string).includes('ui/existing/ui.spec.ts'));
    assert.ok((pkg.scripts['test:ui'] as string).includes('ui/new-feature/ui.spec.ts'));
  });

  // Traceability: TC-UPJ-008 → write failure throws ERR_PACKAGE_JSON_WRITE
  test('TC-UPJ-008: throws Error with code ERR_PACKAGE_JSON_WRITE on write failure', () => {
    const nonExistentPath = '/nonexistent/path/package.json';
    // Create a temp pkg to read from, but write to a bad path
    // We need to trick the function — use a path that exists for read but not for write
    // The simplest approach: pass a path that doesn't exist at all (read will also fail)
    // Instead, test by passing a read-only directory path
    // Actually the function reads then writes to the same path, so we test write failure
    // by making the file read-only after writing it
    const pkgPath = writeTempPkg(BASE_PKG);
    // Make the file read-only
    chmodSync(pkgPath, 0o444);
    try {
      assert.throws(
        () => updatePackageJson({ api: ['api/new/api.spec.ts'], ui: [], cli: [] }, pkgPath),
        (err: unknown) => {
          assert.ok(err instanceof Error);
          assert.equal((err as Error & { code: string }).code, 'ERR_PACKAGE_JSON_WRITE');
          return true;
        }
      );
    } finally {
      chmodSync(pkgPath, 0o644);
    }
  });

  // Traceability: TC-UPJ-009 → no-op when all specPaths are empty
  test('TC-UPJ-009: does not modify scripts when all specPaths are empty', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson({ api: [], ui: [], cli: [] }, pkgPath);
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.equal(pkg.scripts['test:api'], BASE_PKG.scripts['test:api']);
    assert.equal(pkg.scripts['test:cli'], BASE_PKG.scripts['test:cli']);
    assert.equal(pkg.scripts['test'], BASE_PKG.scripts['test']);
    assert.ok(!('test:ui' in pkg.scripts));
  });

  // Traceability: TC-UPJ-010 → multiple new paths merged at once
  test('TC-UPJ-010: merges multiple new paths across api, ui, and cli in one call', () => {
    const pkgPath = writeTempPkg(BASE_PKG);
    updatePackageJson(
      {
        api: ['api/feature-a/api.spec.ts', 'api/feature-b/api.spec.ts'],
        ui: ['ui/feature-a/ui.spec.ts'],
        cli: ['cli/feature-a/cli.spec.ts'],
      },
      pkgPath
    );
    const pkg = readPkg(pkgPath) as { scripts: Record<string, string> };
    assert.ok((pkg.scripts['test:api'] as string).includes('api/feature-a/api.spec.ts'));
    assert.ok((pkg.scripts['test:api'] as string).includes('api/feature-b/api.spec.ts'));
    assert.ok((pkg.scripts['test:ui'] as string).includes('ui/feature-a/ui.spec.ts'));
    assert.ok((pkg.scripts['test:cli'] as string).includes('cli/feature-a/cli.spec.ts'));
    assert.ok((pkg.scripts['test'] as string).includes('npm run test:ui'));
  });
});
