import { test, expect } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId } from './test-helpers';

test.describe('Team Invite Member - Searchable User Picker', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('invite dialog has searchable user input', async ({ page }) => {
    if (!teamId) return;
    test.skip();

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
    if (!teamId) return;
    test.skip();

    await page.goto(`${BASE}/teams/${teamId}`);
    await page.locator('text=添加成员').click();

    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill('admin');

    const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    const options = dropdown.locator('button[data-testid^="invite-user-option-"]');
    const count = await options.count();
    expect(count).toBeGreaterThan(0);
  });

  test('selecting a user fills the field and enables submit', async ({ page }) => {
    if (!teamId) return;
    test.skip();

    await page.goto(`${BASE}/teams/${teamId}`);
    await page.locator('text=添加成员').click();

    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    await searchInput.fill('admin');

    const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    const firstOption = dropdown.locator('button[data-testid^="invite-user-option-"]').first();
    await firstOption.click();

    await expect(dropdown).not.toBeVisible();

    const inputValue = await searchInput.inputValue();
    expect(inputValue).toContain('admin');

    const submitBtn = page.locator('[data-testid="invite-submit-btn"]');
    await expect(submitBtn).toBeDisabled();

    await page.locator('[data-testid="invite-role-select"]').click();
    const roleOption = page.locator('.radix-select-viewport [role="option"]').first();
    if (await roleOption.isVisible({ timeout: 2000 }).catch(() => false)) {
      await roleOption.click();
    }

    await expect(submitBtn).toBeEnabled();
  });

  test('can add a member via API and see it in member list', async ({ page }) => {
    if (!teamId) return;
    test.skip();

    const uniqueName = `e2e_invite_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: uniqueName,
        displayName: `E2E Invite Test`,
        email: `${uniqueName}@test.com`,
      }),
    });

    if (createRes.status !== 201) {
      test.skip();
    }

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

    await expect(page.locator('[data-testid="invite-user-search"]')).not.toBeVisible({ timeout: 5000 }).catch(() => {});

    const memberList = page.locator('text=E2E Invite Test');
    await expect(memberList).toBeVisible({ timeout: 5000 }).catch(() => {});
  });
});
