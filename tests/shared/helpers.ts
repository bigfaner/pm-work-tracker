import { execSync } from 'node:child_process';
import { readFileSync, mkdirSync, existsSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parse as parseYaml } from 'yaml';

const __dirname = dirname(fileURLToPath(import.meta.url));

// ── Config ─────────────────────────────────────────────────────────
const _configPath = findConfigPath();

function findConfigPath(): string {
  // Allow explicit override via environment variable
  const envPath = process.env.E2E_CONFIG_PATH;
  if (envPath && existsSync(envPath)) return resolve(envPath);

  let dir = __dirname;
  for (let i = 0; i < 10; i++) {
    const candidate = resolve(dir, 'tests', 'shared', 'config.yaml');
    if (existsSync(candidate)) return candidate;
    const parent = resolve(dir, '..');
    if (parent === dir) break;
    dir = parent;
  }
  throw new Error(`tests/shared/config.yaml not found. Searched upward from ${__dirname}. Set E2E_CONFIG_PATH or run /gen-sitemap first.`);
}

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

export const baseUrl = _config.baseUrl ?? 'http://localhost:5173';
export const apiBaseUrl = _config.apiBaseUrl ?? 'http://localhost:8080';
export const apiUrl = apiBaseUrl;
export const BASE = baseUrl;
export const API = apiBaseUrl + '/v1';
const DEFAULT_TIMEOUT = toNumber(_config.timeout, 30000);

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

export const _loginLocators = _config.loginLocators;

// ── Token caching (multi-account, shared across all helpers) ───────
const _tokenCache = new Map<string, { token: string; expiry: number }>();
const TOKEN_TTL = 23 * 60 * 60 * 1000; // 23 hours

function cacheKey(username: string, password: string): string {
  return `${username}:${password}`;
}

function getCachedToken(key: string): string | null {
  const entry = _tokenCache.get(key);
  if (entry && Date.now() < entry.expiry) return entry.token;
  _tokenCache.delete(key);
  return null;
}

function setCachedToken(key: string, token: string): void {
  _tokenCache.set(key, { token, expiry: Date.now() + TOKEN_TTL });
}

async function fetchLoginToken(apiBase: string, username: string, password: string): Promise<string> {
  for (let attempt = 0; attempt < 5; attempt++) {
    const res = await fetch(`${apiBase}/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (res.status === 429) {
      await new Promise((r) => setTimeout(r, 2000 * (attempt + 1)));
      continue;
    }
    if (res.status !== 200) throw new Error(`Auth failed for ${username}: ${res.status}`);
    const json = await res.json();
    const token = json.data?.token ?? json.token ?? json.access_token;
    if (!token) throw new Error(`No token in auth response`);
    return token;
  }
  throw new Error(`Auth failed for ${username} after retries: rate limited`);
}

export async function getAuthToken(creds: UICredentials = defaultCreds): Promise<string> {
  const key = cacheKey(creds.username, creds.password);
  const cached = getCachedToken(key);
  if (cached) return cached;
  const api = _config.apiBaseUrl ?? 'http://localhost:8080';
  const token = await fetchLoginToken(api, creds.username, creds.password);
  setCachedToken(key, token);
  return token;
}

export async function getApiToken(apiBaseUrl: string, creds: UICredentials = defaultCreds): Promise<string> {
  const key = cacheKey(creds.username, creds.password);
  const cached = getCachedToken(key);
  if (cached) return cached;
  const token = await fetchLoginToken(apiBaseUrl, creds.username, creds.password);
  setCachedToken(key, token);
  return token;
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

/** Login via API and return {authHeader, token}. Uses shared token cache. */
export async function loginAs(
  username: string,
  password: string,
): Promise<{ authHeader: Record<string, string>; token: string }> {
  const key = cacheKey(username, password);
  let token = getCachedToken(key);
  if (!token) {
    token = await fetchLoginToken(apiBaseUrl, username, password);
    setCachedToken(key, token);
  }
  return { authHeader: { Authorization: `Bearer ${token}` }, token };
}

// ── Frontend compatibility functions ────────────────────────────────

let _cachedTeamId: string | null = null;

export function invalidateAuthCache(): void {
  _tokenCache.clear();
  _cachedTeamId = null;
}

export async function getFirstTeamId(token: string): Promise<string | null> {
  if (_cachedTeamId != null) return _cachedTeamId;
  const res = await fetch(`${API}/teams`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = await res.json();
  const data = json.data ?? json;
  const list = Array.isArray(data) ? data : (data?.items ?? []);
  if (list.length > 0) {
    _cachedTeamId = String(list[0].bizKey ?? list[0].id ?? list[0].ID);
  }
  return _cachedTeamId;
}

export function parseApiData(resp: any): any {
  return resp.data !== undefined ? resp.data : resp;
}

export async function getFirstMemberKey(token: string, teamId: string): Promise<string | null> {
  const res = await fetch(`${API}/teams/${teamId}/members`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = await res.json();
  const data = parseApiData(json);
  const list = Array.isArray(data) ? data : (data?.items ?? []);
  if (list.length > 0) {
    return String(list[0].userKey ?? list[0].userId ?? list[0].id);
  }
  return null;
}

export async function getRoleKey(token: string, roleName: string): Promise<string | null> {
  const res = await fetch(`${API}/admin/roles`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = await res.json();
  const data = parseApiData(json);
  const list = Array.isArray(data) ? data : (data?.items ?? []);
  const role = list.find((r: any) => r.roleName === roleName);
  return role ? String(role.bizKey) : null;
}

export function extractBizKey(data: any): string | null {
  if (!data) return null;
  const val = data.bizKey ?? data.id ?? data.ID ?? data.data?.bizKey ?? data.data?.id;
  return val != null ? String(val) : null;
}

export async function getTokenForUser(username: string, password: string): Promise<string> {
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  const json = await res.json();
  const token = json.data?.token || json.token;
  if (!token) throw new Error(`Login failed for ${username}: ${JSON.stringify(json)}`);
  return token;
}

export async function createUser(token: string, username: string, displayName: string): Promise<{ userId: string; username: string; initialPassword?: string }> {
  const res = await fetch(`${API}/admin/users`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, displayName }),
  });
  const json = await res.json();
  if (json.code !== 0) throw new Error(`Failed to create user: ${JSON.stringify(json)}`);
  return {
    userId: String(json.data?.bizKey),
    username,
    initialPassword: json.data?.initialPassword,
  };
}

export async function deleteUser(token: string, userId: string): Promise<void> {
  await fetch(`${API}/admin/users/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function addUserToTeam(token: string, teamId: string, username: string, roleKey: string): Promise<void> {
  const res = await fetch(`${API}/teams/${teamId}/members`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, roleKey }),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Failed to add user to team: ${text}`);
  }
}

export async function removeUserFromTeam(token: string, teamId: string, userId: string): Promise<void> {
  await fetch(`${API}/teams/${teamId}/members/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

// ── Shared RBAC helpers ────────────────────────────────────────────

export function randomCode(length = 5): string {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  return Array.from({ length }, () => letters[Math.floor(Math.random() * 26)]).join('');
}

export function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

/** Parse a curl response body string: check code === 0, return data. */
export function parseApiBody(body: string): any {
  const resp = JSON.parse(body);
  if (resp.code !== 0) throw new Error(`API error: ${resp.message ?? resp.code}`);
  return resp.data;
}

// ── RBAC fixture setup ─────────────────────────────────────────────

export interface RbacFixtures {
  superadminToken: string;
  pmToken: string;
  memberToken: string;
  pmUserBizKey: string;
  memberUserBizKey: string;
  teamBizKey: string;
  memberRoleKey: string;
  noPermsUsername?: string;
}

/**
 * Shared beforeAll fixture for RBAC tests.
 * Creates team + PM/member users with tokens. Returns all IDs.
 */
export async function setupRbacFixtures(extra?: { noPerms?: boolean }): Promise<RbacFixtures & { noPermsToken?: string; noPermsUserBizKey?: string }> {
  const superadminToken = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
  const runId = Date.now();

  // Fetch preset role bizKeys
  const rolesRes = await curl('GET', `${apiUrl}/v1/admin/roles`, { headers: authHeader(superadminToken) });
  const rolesData = parseApiBody(rolesRes.body);
  const roles: Array<{ bizKey: string; roleName: string }> = rolesData.items ?? rolesData;
  const memberRole = roles.find((r) => r.roleName === 'member');
  if (!memberRole) throw new Error('member role not found');
  const memberRoleKey = memberRole.bizKey;
  const pmRole = roles.find((r) => r.roleName === 'pm');
  if (!pmRole) throw new Error('pm role not found');
  const pmRoleKey = pmRole.bizKey;

  // Create test team
  const teamRes = await curl('POST', `${apiUrl}/v1/teams`, {
    headers: authHeader(superadminToken),
    body: JSON.stringify({ name: `e2e-rbac-${runId}`, code: randomCode() }),
  });
  const teamBizKey = extractBizKey(parseApiBody(teamRes.body))!;

  // Create users helper
  async function makeUser(username: string, displayName: string) {
    const res = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username, displayName }),
    });
    const data = parseApiBody(res.body);
    const bizKey = extractBizKey(data)!;
    const token = await getApiToken(apiBaseUrl, { username, password: data.initialPassword });
    return { bizKey, token };
  }

  const pm = await makeUser(`e2e-pm-${runId}`, 'E2E PM');
  const member = await makeUser(`e2e-member-${runId}`, 'E2E Member');

  // Add users to team
  await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
    headers: authHeader(superadminToken),
    body: JSON.stringify({ username: `e2e-pm-${runId}`, roleKey: memberRoleKey }),
  });
  await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
    headers: authHeader(superadminToken),
    body: JSON.stringify({ username: `e2e-member-${runId}`, roleKey: memberRoleKey }),
  });

  // Transfer PM role
  await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
    headers: authHeader(superadminToken),
    body: JSON.stringify({ newPmUserKey: pm.bizKey }),
  });

  const base: RbacFixtures = {
    superadminToken,
    pmToken: pm.token,
    memberToken: member.token,
    pmUserBizKey: pm.bizKey,
    memberUserBizKey: member.bizKey,
    teamBizKey,
    memberRoleKey,
  };

  if (extra?.noPerms) {
    const noPermsUsername = `e2e-noperms-${runId}`;
    const noPerms = await makeUser(noPermsUsername, 'E2E NoPerms');
    // Add noPerms user to team with member role so test can later change role
    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: noPermsUsername, roleKey: memberRoleKey }),
    });
    return { ...base, noPermsToken: noPerms.token, noPermsUserBizKey: noPerms.bizKey, noPermsUsername };
  }

  return base;
}

// ── Soft-delete test helpers ───────────────────────────────────────

export async function createTestTeam(token: string, name: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/teams`, {
    headers: authHeader(token),
    body: JSON.stringify({ name, code: randomCode() }),
  });
  const data = parseApiBody(res.body);
  return String(data.bizKey ?? data.teamKey ?? data.id);
}

export async function createTestMainItem(token: string, teamBizKey: string, title: string, priority: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
    headers: authHeader(token),
    body: JSON.stringify({ title, priority, assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
  });
  const data = parseApiBody(res.body);
  return String(data.bizKey ?? data.id);
}

export async function createTestSubItem(token: string, teamBizKey: string, mainItemBizKey: string, title: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
    headers: authHeader(token),
    body: JSON.stringify({ mainItemKey: mainItemBizKey, title, priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
  });
  const data = parseApiBody(res.body);
  return String(data.bizKey ?? data.id);
}

export async function createTestRole(token: string, name: string, permissionCodes: string[]): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/admin/roles`, {
    headers: authHeader(token),
    body: JSON.stringify({ name, permissionCodes }),
  });
  const data = parseApiBody(res.body);
  return String(data.bizKey ?? data.id);
}

export async function softDeleteRole(token: string, roleId: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/roles/${roleId}`, { headers: authHeader(token) });
  if (res.status !== 200) throw new Error(`softDeleteRole failed: ${res.status}`);
}

export async function createTestUser(token: string, username: string, displayName: string): Promise<string> {
  const res = await curl('POST', `${apiUrl}/v1/admin/users`, {
    headers: authHeader(token),
    body: JSON.stringify({ username, displayName }),
  });
  const data = parseApiBody(res.body);
  return String(data.bizKey ?? data.id);
}

export async function softDeleteUser(token: string, userBizKey: string): Promise<void> {
  const res = await curl('DELETE', `${apiUrl}/v1/admin/users/${userBizKey}`, { headers: authHeader(token) });
  if (res.status !== 200) throw new Error(`softDeleteUser failed: ${res.status}`);
}
