import { type Page } from '@playwright/test';

export const BASE = 'http://localhost:5173';
export const API = 'http://localhost:8080/v1';

let cachedToken: string | null = null;
let cachedTokenExpiry = 0;
let cachedTeamId: string | null = null;

export async function getAuthToken(): Promise<string> {
  if (cachedToken && Date.now() < cachedTokenExpiry) {
    return cachedToken;
  }
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const json = await res.json();
  const token = json.data?.token || json.token;
  if (!token) throw new Error('Login failed: no token in response');
  cachedToken = token;
  cachedTokenExpiry = Date.now() + 23 * 60 * 60 * 1000; // JWT ~24h, cache 23h
  return token;
}

export async function getFirstTeamId(token: string): Promise<string | null> {
  if (cachedTeamId != null) return cachedTeamId;
  const res = await fetch(`${API}/teams`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = await res.json();
  const data = json.data ?? json;
  const list = Array.isArray(data) ? data : (data?.items ?? []);
  if (list.length > 0) {
    cachedTeamId = String(list[0].bizKey ?? list[0].id ?? list[0].ID);
  }
  return cachedTeamId;
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
  const role = list.find((r: any) => r.name === roleName);
  return role ? String(role.id) : null;
}

/** Clear cached token so next getAuthToken() fetches a fresh one. */
export function invalidateAuthCache(): void {
  cachedToken = null;
  cachedTokenExpiry = 0;
}

export function extractBizKey(data: any): string | null {
  if (!data) return null;
  const val = data.bizKey ?? data.id ?? data.ID ?? data.data?.bizKey ?? data.data?.id;
  return val != null ? String(val) : null;
}

export async function login(page: Page): Promise<void> {
  const token = await getAuthToken();

  await page.goto(`${BASE}/login`);
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

  await page.goto(`${BASE}/items`);
  await page.waitForURL(/\/items/, { timeout: 10000 });

  // Wait for AppLayout to fetch permissions and persist them to localStorage.
  await page.waitForFunction(() => {
    try {
      const raw = localStorage.getItem('auth-storage');
      if (!raw) return false;
      const parsed = JSON.parse(raw);
      return parsed?.state?.permissions !== null && parsed?.state?.permissions !== undefined;
    } catch { return false; }
  }, { timeout: 10000 });

  // Wait for AppLayout to auto-select a team (currentTeamId persisted to localStorage)
  await page.waitForFunction(() => {
    try {
      const raw = localStorage.getItem('team-storage');
      if (!raw) return false;
      const parsed = JSON.parse(raw);
      return parsed?.state?.currentTeamId != null;
    } catch { return false; }
  }, { timeout: 5000 }).catch(() => {});
}

export async function navTo(page: Page, path: string) {
  const link = page.locator(`[data-testid="sidebar"] a[href="${path}"]`);
  await link.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
  await link.click();
  await page.waitForTimeout(1500);
}

/**
 * Login as a specific user by setting localStorage directly.
 * @param page Playwright page
 * @param token JWT token for the user
 * @param user User object with isSuperAdmin flag
 */
export async function loginAs(page: Page, token: string, user: { isSuperAdmin: boolean }): Promise<void> {
  await page.goto(`${BASE}/login`);
  await page.evaluate((t) => {
    localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        token: t,
        user: { isSuperAdmin: false },
        isAuthenticated: true,
        isSuperAdmin: false,
        permissions: null,
        permissionsLoadedAt: null,
        _hasHydrated: true,
      },
      version: 0,
    }));
  }, token);

  await page.goto(`${BASE}/items`);
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
}

/**
 * Get auth token for a specific user.
 */
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

/**
 * Create a new user via admin API and return user info.
 */
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

/**
 * Delete a user via admin API.
 */
export async function deleteUser(token: string, userId: string): Promise<void> {
  await fetch(`${API}/admin/users/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

/**
 * Add a user to a team.
 */
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

/**
 * Remove a user from a team.
 */
export async function removeUserFromTeam(token: string, teamId: string, userId: string): Promise<void> {
  await fetch(`${API}/teams/${teamId}/members/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}
