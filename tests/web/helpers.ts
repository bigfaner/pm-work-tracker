// Re-export all shared (pure Node.js) helpers
export * from '../shared/helpers.js';

import type { Page, Locator } from '@playwright/test';
import { baseUrl, API, defaultCreds, getAuthToken, _loginLocators, invalidateAuthCache } from '../shared/helpers.js';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { existsSync, mkdirSync } from 'node:fs';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Screenshots go to <helpers-dir>/../results/screenshots
const SCREENSHOTS_DIR = join(__dirname, '..', 'results', 'screenshots');

// ── Evidence ───────────────────────────────────────────────────────
export async function screenshot(page: Page, tcId: string): Promise<string> {
  if (!existsSync(SCREENSHOTS_DIR)) mkdirSync(SCREENSHOTS_DIR, { recursive: true });
  const path = join(SCREENSHOTS_DIR, `${tcId}.png`);
  await page.screenshot({ path, fullPage: true });
  return path;
}

/** Inject cached token and navigate to targetPath */
export async function login(page: Page, creds: typeof defaultCreds = defaultCreds, targetPath = '/items'): Promise<void> {
  const token = await getAuthToken(creds);
  const authStorage = JSON.stringify({
    state: {
      token,
      user: { isSuperAdmin: true },
      isAuthenticated: true,
      isSuperAdmin: true,
      permissions: null,
      permissionsLoadedAt: null,
      _hasHydrated: true,
    },
    version: 0,
  });
  // Navigate to /items first to let the app boot and hydrate Zustand.
  // We inject auth after the page loads (not via addInitScript) to avoid
  // re-injection on subsequent navigations which causes Zustand rehydration storms.
  await page.goto(`${baseUrl}/items`);
  await page.evaluate((storage) => {
    localStorage.setItem('auth-storage', storage);
  }, authStorage);
  // Reload so Zustand picks up the injected auth from localStorage
  await page.reload();
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
  // Navigate to actual target if different from /items
  if (targetPath !== '/items') {
    await page.goto(`${baseUrl}${targetPath}`);
  }
}

/** Legacy UI login — kept for login-specific tests (TC-025, TC-026) */
export async function loginViaUI(page: Page, creds: typeof defaultCreds = defaultCreds): Promise<void> {
  const DEFAULT_TIMEOUT = 30000;
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
  const DEFAULT_TIMEOUT = 30000;
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

export async function navTo(page: Page, path: string) {
  const link = page.locator(`[data-testid="sidebar"] a[href="${path}"]`);
  await link.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
  await link.click();
  await page.waitForTimeout(1500);
}

/** Login as a specific user via UI (sets localStorage). Different from loginAs (API-only). */
export async function loginAsUser(page: Page, token: string, user: { isSuperAdmin: boolean }, targetPath = '/items'): Promise<void> {
  const authStorage = JSON.stringify({
    state: {
      token,
      user: { isSuperAdmin: user.isSuperAdmin },
      isAuthenticated: true,
      isSuperAdmin: user.isSuperAdmin,
      permissions: null,
      permissionsLoadedAt: null,
      _hasHydrated: true,
    },
    version: 0,
  });

  await page.goto(`${baseUrl}/items`);
  await page.evaluate((storage) => {
    localStorage.setItem('auth-storage', storage);
  }, authStorage);
  await page.reload();
  await page.waitForURL(/\/items/, { timeout: 10000 });
  await page.waitForFunction(() => {
    try {
      const raw = localStorage.getItem('auth-storage');
      if (!raw) return false;
      const parsed = JSON.parse(raw);
      return parsed?.state?.permissions !== null && parsed?.state?.permissions !== undefined;
    } catch { return false; }
  }, { timeout: 10000 });

  if (targetPath !== '/items') {
    await page.goto(`${baseUrl}${targetPath}`);
  }
}
