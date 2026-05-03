import { readFileSync, writeFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

export interface SpecPaths {
  api: string[];
  ui: string[];
  cli: string[];
}

const __dirname = dirname(fileURLToPath(import.meta.url));
const DEFAULT_PKG_PATH = resolve(__dirname, '../../../../../tests/e2e/package.json');

const NODE_TEST_PREFIX = 'node --import tsx/esm --test';

function parseSpecsFromScript(script: string): string[] {
  if (!script) return [];
  return script
    .split(' && ')
    .map((part) => part.trim())
    .filter((part) => part.startsWith(NODE_TEST_PREFIX))
    .map((part) => part.slice(NODE_TEST_PREFIX.length).trim())
    .filter(Boolean);
}

function buildScript(paths: string[]): string {
  return paths.map((p) => `${NODE_TEST_PREFIX} ${p}`).join(' && ');
}

function mergePaths(existing: string[], newPaths: string[]): string[] {
  const seen = new Set(existing);
  const result = [...existing];
  for (const p of newPaths) {
    if (!seen.has(p)) {
      seen.add(p);
      result.push(p);
    }
  }
  return result;
}

export function updatePackageJson(specPaths: SpecPaths, pkgPath: string = DEFAULT_PKG_PATH): void {
  const raw = readFileSync(pkgPath, 'utf8');
  const pkg = JSON.parse(raw) as { scripts: Record<string, string> };

  // Merge API paths
  if (specPaths.api.length > 0) {
    const existing = parseSpecsFromScript(pkg.scripts['test:api'] ?? '');
    pkg.scripts['test:api'] = buildScript(mergePaths(existing, specPaths.api));
  }

  // Merge CLI paths
  if (specPaths.cli.length > 0) {
    const existing = parseSpecsFromScript(pkg.scripts['test:cli'] ?? '');
    pkg.scripts['test:cli'] = buildScript(mergePaths(existing, specPaths.cli));
  }

  // Merge UI paths and update test script
  if (specPaths.ui.length > 0) {
    const existing = parseSpecsFromScript(pkg.scripts['test:ui'] ?? '');
    pkg.scripts['test:ui'] = buildScript(mergePaths(existing, specPaths.ui));

    const testScript = pkg.scripts['test'] ?? '';
    if (!testScript.includes('npm run test:ui')) {
      pkg.scripts['test'] = testScript
        ? testScript + ' && npm run test:ui'
        : 'npm run test:ui';
    }
  }

  try {
    writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + '\n', 'utf8');
  } catch (err) {
    const error = new Error(
      `Failed to write ${pkgPath}: ${(err as Error).message}`
    ) as Error & { code: string };
    error.code = 'ERR_PACKAGE_JSON_WRITE';
    throw error;
  }
}
