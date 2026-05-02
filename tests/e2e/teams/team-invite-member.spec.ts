import { test, expect } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId } from '../helpers.js';

test.describe('Team Invite Member - Searchable User Picker', () => {
  let authToken: string;
  let teamId: string | null;
  let searchUsername: string;
  let searchUserBizKey: string;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);

    // Create a user that is NOT in the team, so it appears in the invite search
    searchUsername = `e2e_invite_search_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: searchUsername, displayName: 'E2E Invite Search' }),
    });
    const json = await createRes.json();
    searchUserBizKey = String(json.data.bizKey);
  });

  test.afterAll(async () => {
    if (!searchUserBizKey) return;
    await fetch(`${API}/admin/users/${searchUserBizKey}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authToken}` },
    });
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('invite dialog has searchable user input', async ({ page }) => {
    if (!teamId) { test.skip(); return; }

    await page.goto(`${BASE}/teams/${teamId}`);
    await page.locator('text=添加成员').click();

    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    const roleSelect = page.locator('[data-testid="invite-role-select"]');
    await expect(roleSelect).toBeVisible();

    const submitBtn = page.locator('[data-testid="invite-submit-btn"]');
    await expect(submitBtn).toBeDisabled();
  });

  test('searching users shows dropdown with results', async ({ page }) => {
    if (!teamId) { test.skip(); return; }

    await page.goto(`${BASE}/teams/${teamId}`);
    await page.locator('text=添加成员').click();

    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(searchUsername);

    const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    const options = dropdown.locator('button[data-testid^="invite-user-option-"]');
    const count = await options.count();
    expect(count).toBeGreaterThan(0);
  });

  test('selecting a user fills the field and enables submit', async ({ page }) => {
    if (!teamId) { test.skip(); return; }

    await page.goto(`${BASE}/teams/${teamId}`);
    await page.locator('text=添加成员').click();

    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill(searchUsername);

    const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    const firstOption = dropdown.locator('button[data-testid^="invite-user-option-"]').first();
    await firstOption.click();

    await expect(dropdown).not.toBeVisible();

    const inputValue = await searchInput.inputValue();
    expect(inputValue.length).toBeGreaterThan(0);

    const submitBtn = page.locator('[data-testid="invite-submit-btn"]');

    await page.locator('[data-testid="invite-role-select"]').click();
    const roleOption = page.locator('[role="option"]').first();
    if (await roleOption.isVisible({ timeout: 2000 }).catch(() => false)) {
      await roleOption.click();
    }

    await expect(submitBtn).toBeEnabled();
  });

  test('can add a member via API and see it in member list', async ({ page }) => {
    if (!teamId) { test.skip(); return; }

    // Create a fresh user to add to the team
    const uniqueName = `e2e_invite_add_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: uniqueName, displayName: 'E2E Invite Test' }),
    });
    if (createRes.status !== 200 && createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const newUserBizKey = String(created.data.bizKey);

    try {
      await page.goto(`${BASE}/teams/${teamId}`);
      await page.locator('text=添加成员').click();

      const searchInput = page.locator('[data-testid="invite-user-search"]');
      await expect(searchInput).toBeVisible({ timeout: 5000 });

      await searchInput.fill(uniqueName);

      const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
      await expect(dropdown).toBeVisible({ timeout: 5000 });

      const matchingOption = dropdown.locator(`text=${uniqueName}`).first();
      await matchingOption.click();

      await page.locator('[data-testid="invite-role-select"]').click();
      const roleOption = page.locator('[role="option"]').first();
      if (await roleOption.isVisible({ timeout: 2000 }).catch(() => false)) {
        await roleOption.click();
      }

      const submitBtn = page.locator('[data-testid="invite-submit-btn"]');
      await expect(submitBtn).toBeEnabled();
      await submitBtn.click();

      await page.waitForTimeout(2000);

      const memberList = page.locator('text=E2E Invite Test');
      await expect(memberList).toBeVisible({ timeout: 5000 });
    } finally {
      await fetch(`${API}/admin/users/${newUserBizKey}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
    }
  });
});
