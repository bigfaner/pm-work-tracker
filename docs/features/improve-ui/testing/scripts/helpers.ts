import { execSync } from 'node:child_process';
import { writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

const DEFAULT_TIMEOUT = 30000;
const SCREENSHOTS_DIR = join(__dirname, '..', 'results', 'screenshots');

export const baseUrl = process.env.E2E_BASE_URL ?? 'http://localhost:5173';
export const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

export function ab(cmd: string): string {
  try {
    return execSync(`agent-browser ${cmd}`, {
      encoding: 'utf-8',
      timeout: DEFAULT_TIMEOUT,
    });
  } catch (e: any) {
    // Windows: agent-browser may print output but not exit cleanly.
    // If stdout has content, treat it as success.
    const stdout = (e.stdout ?? '') as string;
    if (stdout.trim()) return stdout;
    throw e;
  }
}

export function abJson(cmd: string): any {
  const raw = ab(`${cmd} --json`);
  try {
    return JSON.parse(raw);
  } catch {
    throw new Error(`Failed to parse agent-browser JSON output: ${raw.slice(0, 200)}`);
  }
}

export function snapshotContains(text: string): boolean {
  const result = abJson('snapshot');
  return result?.data?.snapshot?.includes(text) ?? false;
}

export function findElement(role: string, name?: string): string | null {
  // find role is unreliable; use snapshot-based discovery instead
  try {
    const snap = abJson('snapshot -i');
    const refs = snap?.data?.refs ?? {};
    for (const [ref, el] of Object.entries(refs)) {
      const elem = el as { role?: string; name?: string };
      if (elem.role === role) {
        if (!name || elem.name === name) {
          return ref;
        }
      }
    }
    return null;
  } catch {
    return null;
  }
}

export function screenshot(tcId: string): string {
  if (!existsSync(SCREENSHOTS_DIR)) {
    mkdirSync(SCREENSHOTS_DIR, { recursive: true });
  }
  const path = join(SCREENSHOTS_DIR, `${tcId}.png`);
  ab(`screenshot "${path}"`);
  return path;
}

export interface CurlResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
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

export interface CliResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export function runCli(cmd: string, cwd?: string): CliResult {
  try {
    const stdout = execSync(cmd, {
      encoding: 'utf-8',
      timeout: DEFAULT_TIMEOUT,
      cwd: cwd ?? process.cwd(),
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

/** Login via API and return the Authorization header. */
export async function loginAs(
  username: string,
  password: string,
): Promise<{ authHeader: Record<string, string>; token: string }> {
  const res = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
    body: JSON.stringify({ username, password }),
  });
  if (res.status !== 200) {
    throw new Error(`Login failed for ${username}: ${res.status} ${res.body}`);
  }
  const data = JSON.parse(res.body);
  const token = data.data?.token ?? data.token;
  return { authHeader: { Authorization: `Bearer ${token}` }, token };
}

/** Login via agent-browser UI. Opens login page, fills credentials, submits. */
export function browserLogin(username: string, password: string): void {
  ab(`open ${baseUrl}/login`);
  ab('wait --load networkidle');

  // Find inputs by role and name
  const userRef = findElement('textbox', '账号');
  const passRef = findElement('textbox', '密码');

  if (!userRef || !passRef) {
    throw new Error(`Login form elements not found. user=${userRef} pass=${passRef}`);
  }

  ab(`fill ${userRef} "${username}"`);
  ab(`fill ${passRef} "${password}"`);

  // Click login button
  const loginBtn = findElement('button', '登录');
  if (loginBtn) {
    ab(`click ${loginBtn}`);
  }
  ab('wait --load networkidle');
}
