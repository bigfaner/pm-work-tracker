import { execSync } from 'node:child_process';
import { writeFileSync, unlinkSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const DEFAULT_TIMEOUT = 15000;
const PROJECT_ROOT = join(__dirname, '..', '..', '..', '..');
const BACKEND_DIR = join(PROJECT_ROOT, 'backend');
const CONFIG_PATH = join(BACKEND_DIR, 'config.yaml');

export const serverBin = process.env.E2E_SERVER_BIN ?? join(PROJECT_ROOT, 'bin', 'windows-amd64', 'pm-work-tracker.exe');

export function writeConfig(yaml: string): void {
  writeFileSync(CONFIG_PATH, yaml, 'utf-8');
}

export function removeConfig(): void {
  if (existsSync(CONFIG_PATH)) {
    unlinkSync(CONFIG_PATH);
  }
}

export function runServer(env?: Record<string, string>) {
  try {
    const stdout = execSync(`"${serverBin}" --dev 2>&1`, {
      encoding: 'utf-8',
      timeout: DEFAULT_TIMEOUT,
      cwd: BACKEND_DIR,
      env: { ...process.env, ...env },
    });
    return { stdout, stderr: '', exitCode: 0 };
  } catch (e: any) {
    return { stdout: e.stdout ?? '', stderr: e.stderr ?? '', exitCode: e.status ?? 1 };
  }
}
