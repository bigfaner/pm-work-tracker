import { execSync } from 'node:child_process';
import { mkdirSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

const DEFAULT_TIMEOUT = 60000;
const SCREENSHOTS_DIR = join(__dirname, '..', 'results', 'screenshots');

export const baseUrl = process.env.E2E_BASE_URL ?? 'http://localhost:3456';
export const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

// Auth helpers — set via env or use defaults for local dev
export const PM_TOKEN = process.env.E2E_PM_TOKEN ?? '';
export const EXECUTOR_TOKEN = process.env.E2E_EXECUTOR_TOKEN ?? '';
export const TEAM_ID = process.env.E2E_TEAM_ID ?? '1';

export function ab(cmd: string): string {
  return execSync(`npx agent-browser ${cmd}`, {
    encoding: 'utf-8',
    timeout: DEFAULT_TIMEOUT,
  });
}

export function abJson(cmd: string): any {
  const raw = ab(`${cmd} --json`);
  try {
    return JSON.parse(raw);
  } catch {
    throw new Error(`Failed to parse agent-browser JSON output: ${raw.slice(0, 200)}`);
  }
}

/** Convert snapshot refs object to array with ref field included */
export function getElements(snap: any): Array<{ ref: string; name: string; role: string; [k: string]: any }> {
  const refs = snap?.data?.refs ?? {};
  return Object.entries(refs).map(([ref, el]: [string, any]) => ({ ref: `@${ref}`, ...el }));
}

export function snapshotContains(text: string): boolean {
  const result = abJson('snapshot');
  return result?.data?.snapshot?.includes(text) ?? false;
}

export function findElement(role: string, name?: string): string | null {
  const cmd = name
    ? `find role ${role} --name "${name}" --json`
    : `find role ${role} --json`;
  try {
    const result = abJson(cmd);
    return result?.data?.ref ?? result?.ref ?? null;
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
  json<T = any>(): T;
}

export async function curl(
  method: string,
  url: string,
  opts?: {
    body?: unknown;
    token?: string;
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
        ...(opts?.token ? { Authorization: `Bearer ${opts.token}` } : {}),
      },
      body: opts?.body !== undefined ? JSON.stringify(opts.body) : undefined,
      signal: controller.signal,
    });

    const headers: Record<string, string> = {};
    res.headers.forEach((v, k) => { headers[k] = v; });
    const bodyText = await res.text();

    return {
      status: res.status,
      headers,
      body: bodyText,
      json<T = any>(): T {
        return JSON.parse(bodyText) as T;
      },
    };
  } finally {
    clearTimeout(timeout);
  }
}

/** Convenience: build team-scoped API URL */
export function teamUrl(path: string): string {
  return `${apiUrl}/api/v1/teams/${TEAM_ID}${path}`;
}
