import { execSync } from 'node:child_process';
import { writeFileSync, unlinkSync, existsSync, mkdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

const DEFAULT_TIMEOUT = 15000;
const PROJECT_ROOT = join(__dirname, '..', '..', '..', '..', '..');
export const serverBin = process.env.E2E_SERVER_BIN ?? join(PROJECT_ROOT, 'server');
export const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';
const CONFIG_PATH = join(PROJECT_ROOT, 'config.test.yaml');

export interface CliResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export interface CurlResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
}

export function runCli(cmd: string, env?: Record<string, string>): CliResult {
  try {
    const stdout = execSync(cmd, {
      encoding: 'utf-8',
      timeout: DEFAULT_TIMEOUT,
      cwd: PROJECT_ROOT,
      env: { ...process.env, ...env },
    });
    return { stdout, stderr: '', exitCode: 0 };
  } catch (e: any) {
    return {
      stdout: e.stdout ?? '',
      stderr: e.stderr ?? '',
      exitCode: e.status ?? 1,
    };
  }
}

/** Write a config.yaml for testing. */
export function writeConfig(yaml: string): void {
  writeFileSync(CONFIG_PATH, yaml, 'utf-8');
}

/** Remove test config.yaml. */
export function removeConfig(): void {
  if (existsSync(CONFIG_PATH)) {
    unlinkSync(CONFIG_PATH);
  }
}

/** Run the server binary (expecting it to exit with error for validation tests). */
export function runServer(env?: Record<string, string>): CliResult {
  return runCli(`${serverBin} --config config.test.yaml 2>&1`, env);
}

export async function curl(
  method: string,
  url: string,
  opts?: {
    body?: string;
    headers?: Record<string, string>;
    timeout?: number;
  },
): Promise<CurlResponse> {
  const controller = new AbortController();
  const timeout = setTimeout(
    () => controller.abort(),
    opts?.timeout ?? 10000,
  );

  try {
    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...opts?.headers,
      },
      body: opts?.body,
      signal: controller.signal,
    });

    const headers: Record<string, string> = {};
    res.headers.forEach((v, k) => { headers[k] = v; });

    return {
      status: res.status,
      headers,
      body: await res.text(),
    };
  } finally {
    clearTimeout(timeout);
  }
}
