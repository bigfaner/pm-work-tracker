import { execSync } from 'node:child_process';
import { readFileSync, mkdirSync, existsSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parse as parseYaml } from 'yaml';
import type { Page, Locator } from '@playwright/test';

const __dirname = dirname(fileURLToPath(import.meta.url));

// ── Config ─────────────────────────────────────────────────────────
const _configPath = findConfigPath();

function findConfigPath(): string {
  // Allow explicit override via environment variable
  const envPath = process.env.E2E_CONFIG_PATH;
  if (envPath && existsSync(envPath)) return resolve(envPath);

  let dir = __dirname;
  for (let i = 0; i < 10; i++) {
    const candidate = resolve(dir, 'tests', 'e2e', 'config.yaml');
    if (existsSync(candidate)) return candidate;
    const parent = resolve(dir, '..');
    if (parent === dir) break;
    dir = parent;
  }
  throw new Error(`tests/e2e/config.yaml not found. Searched upward from ${__dirname}. Set E2E_CONFIG_PATH or run /gen-sitemap first.`);
}

// Screenshots go to <helpers-dir>/../results/screenshots
const SCREENSHOTS_DIR = join(__dirname, '..', 'results', 'screenshots');

interface E2EConfig {
  baseUrl?: string;
  apiBaseUrl?: string;
  timeout?: number | string;
  username?: string;
  password?: string;
  loginLocators?: { usernameField?: string; passwordField?: string; submitButton?: string };
}

function readConfig(): E2EConfig {
  return parseYaml(readFileSync(findConfigPath(), 'utf-8'));
}

const _config = readConfig();

function toNumber(val: unknown, fallback: number): number {
  if (typeof val === 'number' && Number.isFinite(val)) return val;
  if (typeof val === 'string') {
    const n = parseInt(val, 10);
    return Number.isFinite(n) ? n : fallback;
  }
  return fallback;
}

export const baseUrl = _config.baseUrl ?? 'http://localhost:5174';
export const apiBaseUrl = _config.apiBaseUrl ?? 'http://localhost:8083';
export const apiUrl = apiBaseUrl;
const DEFAULT_TIMEOUT = toNumber(_config.timeout, 30000);

// ── Evidence ───────────────────────────────────────────────────────
export async function screenshot(page: Page, tcId: string): Promise<string> {
  if (!existsSync(SCREENSHOTS_DIR)) mkdirSync(SCREENSHOTS_DIR, { recursive: true });
  const path = join(SCREENSHOTS_DIR, `${tcId}.png`);
  await page.screenshot({ path, fullPage: true });
  return path;
}

// ── HTTP ───────────────────────────────────────────────────────────
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

// ── Auth ────────────────────────────────────────────────────────────
export interface UICredentials {
  username: string;
  password: string;
}

export const defaultCreds: UICredentials = {
  username: _config.username ?? 'admin',
  password: _config.password ?? 'admin123',
};

const _loginLocators = _config.loginLocators;

// ── Token caching (one login per test run) ──────────────────────────
let cachedToken: string | null = null;
let cachedTokenExpiry = 0;

export async function getAuthToken(creds: UICredentials = defaultCreds): Promise<string> {
  if (cachedToken && Date.now() < cachedTokenExpiry) {
    return cachedToken;
  }
  const api = _config.apiBaseUrl ?? 'http://localhost:8080';
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await fetch(`${api}/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: creds.username, password: creds.password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    if (res.status !== 200) throw new Error(`Auth failed: ${res.status}`);
    const json = await res.json();
    const token = json.data?.token ?? json.token;
    if (!token) throw new Error(`No token in auth response`);
    cachedToken = token;
    cachedTokenExpiry = Date.now() + 23 * 60 * 60 * 1000;
    return token;
  }
  throw new Error('Auth failed after retries: rate limited');
}

/** Inject cached token into localStorage and navigate to /items */
export async function login(page: Page, creds: UICredentials = defaultCreds): Promise<void> {
  const token = await getAuthToken(creds);
  await page.goto(`${baseUrl}/login`);
  await page.evaluate((t) => {
    localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        token: t,
        user: { isSuperAdmin: true },
        isAuthenticated: true,
        isSuperAdmin: true,
        permissions: null,
        permissionsLoadedAt: null,
        _hasHydrated: true,
      },
      version: 0,
    }));
  }, token);
  await page.goto(`${baseUrl}/items`);
  await page.waitForURL(/\/items/, { timeout: 10000 });
  // Wait for permissions to load
  await page.waitForFunction(() => {
    try {
      const raw = localStorage.getItem('auth-storage');
      if (!raw) return false;
      const parsed = JSON.parse(raw);
      return parsed?.state?.permissions !== null && parsed?.state?.permissions !== undefined;
    } catch { return false; }
  }, { timeout: 10000 });
  // Wait for team to be auto-selected
  await page.waitForFunction(() => {
    try {
      const raw = localStorage.getItem('team-storage');
      if (!raw) return false;
      const parsed = JSON.parse(raw);
      return parsed?.state?.currentTeamId != null;
    } catch { return false; }
  }, { timeout: 5000 }).catch(() => {});
}

/** Legacy UI login — kept for login-specific tests (TC-025, TC-026) */
export async function loginViaUI(page: Page, creds: UICredentials = defaultCreds): Promise<void> {
  await page.goto(`${baseUrl}/login`);
  await page.waitForLoadState('networkidle');
  const uPat = new RegExp(_loginLocators?.usernameField ?? 'username|email', 'i');
  const pPat = new RegExp(_loginLocators?.passwordField ?? 'password', 'i');
  const bPat = new RegExp(_loginLocators?.submitButton ?? 'login|sign in|submit', 'i');
  await page.getByRole('textbox', { name: uPat }).fill(creds.username);
  await page.getByRole('textbox', { name: pPat }).fill(creds.password);
  await page.getByRole('button', { name: bPat }).click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: DEFAULT_TIMEOUT });
}

export async function getApiToken(apiBaseUrl: string, creds: UICredentials = defaultCreds): Promise<string> {
  // Auth endpoint: POST /v1/auth/login (matches backend router)
  // Retry on 429 (rate limit) with exponential backoff
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await curl('POST', `${apiBaseUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: creds.username, password: creds.password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    if (res.status !== 200) throw new Error(`Auth failed: ${res.status} ${res.body}`);
    const data = JSON.parse(res.body);
    const token = data.token ?? data.access_token ?? data.data?.token;
    if (!token) throw new Error(`No token in auth response. Keys: ${Object.keys(data).join(', ')}`);
    return token;
  }
  throw new Error('Auth failed after retries: rate limited');
}

export function createAuthCurl(
  apiBaseUrl: string,
  token: string,
): (method: string, path: string, opts?: { body?: string; headers?: Record<string, string>; timeout?: number }) => Promise<CurlResponse> {
  return (method, path, opts) =>
    curl(method, new URL(path, apiBaseUrl).toString(), {
      ...opts,
      headers: { Authorization: `Bearer ${token}`, ...opts?.headers },
    });
}

// ── CLI ────────────────────────────────────────────────────────────
export interface CliResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export function runCli(cmd: string, cwd?: string, timeout?: number): CliResult {
  try {
    const stdout = execSync(cmd, {
      encoding: 'utf-8',
      timeout: timeout ?? DEFAULT_TIMEOUT,
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

// ── UI helpers ────────────────────────────────────────────────────────

/** Check if text is visible anywhere on the page */
export async function snapshotContains(page: Page, text: string): Promise<boolean> {
  return page.getByText(text).first().isVisible().catch(() => false);
}

/** Find a single element by ARIA role and optional accessible name */
export function findElement(page: Page, role: string, name?: string): Locator {
  if (name) {
    return page.getByRole(role as any, { name: new RegExp(name, 'i') });
  }
  return page.getByRole(role as any);
}

/** Find all elements matching a role and optional name */
export function findElements(page: Page, role: string, name?: string): Locator[] {
  const loc = name
    ? page.getByRole(role as any, { name: new RegExp(name, 'i') })
    : page.getByRole(role as any);
  return [loc];
}

/** Login via browser UI — kept for login-specific tests (TC-025, TC-026) */
export async function browserLogin(page: Page, username: string, password: string): Promise<void> {
  await page.goto(`${baseUrl}/login`);
  await page.waitForLoadState('networkidle');
  await page.locator('[data-testid="login-username"]').or(
    page.getByRole('textbox', { name: /账号/i }),
  ).fill(username);
  await page.locator('[data-testid="login-password"]').or(
    page.getByRole('textbox', { name: /密码/i }),
  ).fill(password);
  await page.locator('[data-testid="login-submit"]').or(
    page.getByRole('button', { name: /登录/i }),
  ).click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: DEFAULT_TIMEOUT });
}

/** Login via API and return {authHeader, token}. Retries on 429. */
export async function loginAs(
  username: string,
  password: string,
): Promise<{ authHeader: Record<string, string>; token: string }> {
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await curl('POST', `${apiBaseUrl}/v1/auth/login`, {
      body: JSON.stringify({ username, password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    if (res.status !== 200) {
      throw new Error(`Login failed for ${username}: ${res.status} ${res.body}`);
    }
    const data = JSON.parse(res.body);
    const token = data.data?.token ?? data.token;
    return { authHeader: { Authorization: `Bearer ${token}` }, token };
  }
  throw new Error(`Login failed for ${username} after retries: rate limited`);
}
