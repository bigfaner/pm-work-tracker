import { type Page } from '@playwright/test';

export const BASE = 'http://localhost:5173';
export const API = 'http://localhost:8080/v1';

let cachedTeamId: string | null = null;

export async function getAuthToken(): Promise<string> {
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const json = await res.json();
  const token = json.data?.token || json.token;
  if (!token) throw new Error('Login failed: no token in response');
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
