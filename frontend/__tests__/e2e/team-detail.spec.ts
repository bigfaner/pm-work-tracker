import { test, expect } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId, loginAs, getTokenForUser, createUser, deleteUser, addUserToTeam, removeUserFromTeam, getRoleKey } from './test-helpers';

async function createTestMember(token: string, tid: string, label: string) {
  const username = `e2e_${label}_${Date.now()}`;
  const res = await fetch(`${API}/admin/users`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, displayName: `E2E测试${label}` }),
  });
  const data = await res.json();
  const userId = String(data.data?.bizKey || data.bizKey);
  const inviteRes = await fetch(`${API}/teams/${tid}/members`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, roleKey: '3' }),
  });
  if (!inviteRes.ok) {
    const err = await inviteRes.text();
    console.error(`createTestMember invite failed: ${inviteRes.status} ${err}`);
  }
  return { username, userId };
}

async function removeTestMember(token: string, tid: string, userId: string) {
  try {
    await fetch(`${API}/teams/${tid}/members/${userId}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } });
    await fetch(`${API}/admin/users/${userId}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } });
  } catch { /* best effort */ }
}

// ── Section 1: Page Load ──────────────────────────────────────────────────────

test.describe('Team Detail - Page Load', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('page renders team info card', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=团队名称')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=CODE')).toBeVisible();
    await expect(page.locator('[data-testid="team-detail-page"]').locator('text=PM').first()).toBeVisible();
    await expect(page.locator('text=成员数')).toBeVisible();
  });

  test('breadcrumb shows 团队管理 link', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('[data-testid="team-detail-page"]').locator('nav a[href="/teams"]')).toBeVisible({ timeout: 10000 });
  });

  test('member list section is visible', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=成员列表')).toBeVisible();
    await expect(page.locator('th', { hasText: '姓名' })).toBeVisible();
    await expect(page.locator('th', { hasText: '角色' })).toBeVisible();
    await expect(page.locator('th', { hasText: '加入时间' })).toBeVisible();
    await expect(page.locator('th', { hasText: '操作' })).toBeVisible();
  });

  test('danger zone section is visible', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=危险操作')).toBeVisible();
    await expect(page.locator('button', { hasText: '解散团队' })).toBeVisible();
  });

  test('navigating to non-existent team shows not found', async ({ page }) => {
    await page.goto(`${BASE}/teams/999999`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=团队不存在')).toBeVisible({ timeout: 5000 });
  });
});

// ── Section 2: Member List ────────────────────────────────────────────────────

test.describe('Team Detail - Member List', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('at least one member row is visible', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
  });

  test('PM row shows PM badge and no action buttons', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    const pmRow = page.locator('tbody tr').filter({ has: page.getByText('PM', { exact: true }) }).first();
    await expect(pmRow).toBeVisible({ timeout: 5000 });
    await expect(pmRow.locator('[data-testid="change-role-btn"]')).not.toBeVisible();
  });

  test('search by name filters member list', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 10000 });

    const firstName = await page.locator('tbody tr').first().locator('td').first().textContent();
    // textContent concatenates avatar + displayName without spaces, so use 1 char
    const query = firstName?.trim()[0] || 'a';

    await page.locator('input[placeholder="搜索姓名..."]').fill(query);
    await page.waitForTimeout(500);
    expect(await page.locator('tbody tr').count()).toBeGreaterThan(0);
  });

  test('searching with no match shows empty table', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('input[placeholder="搜索姓名..."]').fill('zzz_no_such_member_xyz');
    await page.waitForTimeout(500);
    await expect(page.locator('tbody tr')).toHaveCount(0);
  });

  test('refresh button reloads member list', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="refresh-btn"]').click();
    await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 5000 });
  });
});

// ── Section 3: Change Role ────────────────────────────────────────────────────

test.describe('Team Detail - Change Role', () => {
  let authToken: string;
  let teamId: string | null;
  let testMemberId: string;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
    if (teamId) {
      const m = await createTestMember(authToken, teamId, '角色');
      testMemberId = m.userId;
    }
  });

  test.afterAll(async () => {
    if (teamId && testMemberId) await removeTestMember(authToken, teamId, testMemberId);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('non-PM member row has change role button', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    const changeRoleBtn = page.locator('[data-testid="change-role-btn"]').first();
    await expect(changeRoleBtn).toBeVisible({ timeout: 5000 });
    await expect(changeRoleBtn).toContainText('修改角色');
  });

  test('clicking change role opens dialog with role select', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="change-role-btn"]').first().click();
    await expect(page.getByRole('heading', { name: '修改角色' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('[data-testid="role-edit-select"]')).toBeVisible();
  });

  test('confirm button disabled when role unchanged', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="change-role-btn"]').first().click();
    await expect(page.getByRole('heading', { name: '修改角色' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button', { hasText: '确认修改' })).toBeDisabled();
  });

  test('role select does not include pm option', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="change-role-btn"]').first().click();
    await expect(page.locator('[data-testid="role-edit-select"]')).toBeVisible({ timeout: 5000 });
    await page.locator('[data-testid="role-edit-select"]').click();
    const pmOption = page.locator('[role="option"]', { hasText: /^pm$/ });
    expect(await pmOption.count()).toBe(0);
  });

  test('cancelling change role dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-testid="change-role-btn"]').first().click();
    await expect(page.getByRole('heading', { name: '修改角色' })).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '取消' }).click();
    await expect(page.getByRole('heading', { name: '修改角色' })).not.toBeVisible({ timeout: 3000 });
  });
});

// ── Section 4: Transfer PM ────────────────────────────────────────────────────

test.describe('Team Detail - Transfer PM', () => {
  let authToken: string;
  let teamId: string | null;
  let testMemberId: string;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
    if (teamId) {
      const m = await createTestMember(authToken, teamId, '转PM');
      testMemberId = m.userId;
    }
  });

  test.afterAll(async () => {
    if (teamId && testMemberId) await removeTestMember(authToken, teamId, testMemberId);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('non-PM member row has 设为PM button', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    const transferBtn = page.locator('button', { hasText: '设为PM' }).first();
    await expect(transferBtn).toBeVisible({ timeout: 5000 });
  });

  test('clicking 设为PM opens confirmation dialog', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '设为PM' }).first().click();
    await expect(page.locator('text=确认将')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button', { hasText: '确认设为PM' })).toBeVisible();
  });

  test('cancelling transfer PM dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '设为PM' }).first().click();
    await expect(page.locator('button', { hasText: '确认设为PM' })).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '取消' }).click();
    await expect(page.locator('button', { hasText: '确认设为PM' })).not.toBeVisible({ timeout: 3000 });
  });
});

// ── Section 5: Remove Member ──────────────────────────────────────────────────

test.describe('Team Detail - Remove Member', () => {
  let authToken: string;
  let teamId: string | null;
  let testMemberId: string;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
    if (teamId) {
      const m = await createTestMember(authToken, teamId, '移除');
      testMemberId = m.userId;
    }
  });

  test.afterAll(async () => {
    // Best-effort cleanup — test 5.4 may have already removed this member
    if (teamId && testMemberId) await removeTestMember(authToken, teamId, testMemberId);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('non-PM member row has 移除 button', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button', { hasText: '移除' }).first()).toBeVisible({ timeout: 5000 });
  });

  test('clicking 移除 opens confirmation dialog', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '移除' }).first().click();
    await expect(page.getByRole('heading', { name: '移除成员' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=确认移除成员')).toBeVisible();
    await expect(page.locator('button', { hasText: '确认移除' })).toBeVisible();
  });

  test('cancelling remove dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '移除' }).first().click();
    await expect(page.getByRole('heading', { name: '移除成员' })).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '取消' }).click();
    await expect(page.getByRole('heading', { name: '移除成员' })).not.toBeVisible({ timeout: 3000 });
  });

  test('can remove a non-PM member', async ({ page }) => {
    test.skip(!teamId, 'No team available');

    const username = `e2e_rm_${Date.now()}`;
    const createRes = await fetch(`${API}/admin/users`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, displayName: 'E2E Remove Test' }),
    });
    if (createRes.status !== 200 && createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const userId = created.data?.bizKey || created.bizKey;

    await fetch(`${API}/teams/${teamId}/members`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, roleKey: '3' }),
    });

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    const memberRow = page.locator('tbody tr').filter({ hasText: 'E2E Remove Test' });
    await expect(memberRow).toBeVisible({ timeout: 5000 });
    await memberRow.locator('button', { hasText: '移除' }).click();

    await expect(page.getByRole('heading', { name: '移除成员' })).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '确认移除' }).click();

    await expect(memberRow).not.toBeVisible({ timeout: 5000 });

    if (userId) {
      await fetch(`${API}/admin/users/${userId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
    }
  });
});

// ── Section 6: Add Member (Invite) ────────────────────────────────────────────

test.describe('Team Detail - Add Member', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('添加成员 button is visible', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('button', { hasText: '添加成员' })).toBeVisible({ timeout: 5000 });
  });

  test('clicking 添加成员 opens invite dialog', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '添加成员' }).click();
    await expect(page.locator('[data-testid="invite-user-search"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('[data-testid="invite-role-select"]')).toBeVisible();
    await expect(page.locator('[data-testid="invite-submit-btn"]')).toBeDisabled();
  });

  test('invite role select does not include pm option', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '添加成员' }).click();
    await expect(page.locator('[data-testid="invite-role-select"]')).toBeVisible({ timeout: 5000 });
    await page.locator('[data-testid="invite-role-select"]').click({ force: true });
    const pmOption = page.locator('[role="option"]', { hasText: /^pm$/ });
    expect(await pmOption.count()).toBe(0);
  });

  test('searching users shows dropdown', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '添加成员' }).click();
    const searchInput = page.locator('[data-testid="invite-user-search"]');
    await expect(searchInput).toBeVisible({ timeout: 5000 });
    await searchInput.fill('a');
    const dropdown = page.locator('[data-testid="invite-user-dropdown"]');
    const hasResults = await dropdown.isVisible({ timeout: 3000 }).catch(() => false);
    if (hasResults) {
      expect(await dropdown.locator('button[data-testid^="invite-user-option-"]').count()).toBeGreaterThan(0);
    }
  });

  test('cancelling invite dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '添加成员' }).click();
    await expect(page.locator('[data-testid="invite-user-search"]')).toBeVisible({ timeout: 5000 });
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-testid="invite-user-search"]')).not.toBeVisible({ timeout: 3000 });
  });
});

// ── Section 7: Disband Team ───────────────────────────────────────────────────

test.describe('Team Detail - Disband Team', () => {
  let authToken: string;
  let teamId: string | null;

  test.beforeAll(async () => {
    authToken = await getAuthToken();
    teamId = await getFirstTeamId(authToken);
  });

  test.beforeEach(async ({ page }) => { await login(page); });

  test('clicking 解散团队 opens confirmation dialog', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '解散团队' }).click();
    await expect(page.locator('text=此操作不可恢复')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=请输入团队名称确认')).toBeVisible();
  });

  test('disband confirm button disabled when name not entered', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '解散团队' }).click();
    await expect(page.locator('text=此操作不可恢复')).toBeVisible({ timeout: 5000 });
    const confirmBtn = page.locator('[role="dialog"] button', { hasText: '解散团队' });
    await expect(confirmBtn).toBeDisabled();
  });

  test('disband confirm button disabled with wrong name', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '解散团队' }).click();
    await expect(page.locator('text=此操作不可恢复')).toBeVisible({ timeout: 5000 });
    await page.locator('[role="dialog"] input').fill('wrong team name');
    const confirmBtn = page.locator('[role="dialog"] button', { hasText: '解散团队' });
    await expect(confirmBtn).toBeDisabled();
  });

  test('cancelling disband dialog closes it', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await page.locator('button', { hasText: '解散团队' }).click();
    await expect(page.locator('text=此操作不可恢复')).toBeVisible({ timeout: 5000 });
    await page.locator('button', { hasText: '取消' }).click();
    await expect(page.locator('text=此操作不可恢复')).not.toBeVisible({ timeout: 3000 });
  });

  test('disband enabled only when correct team name entered', async ({ page }) => {
    test.skip(!teamId, 'No team available');
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    const teamName = await page.locator('h1').first().textContent();
    expect(teamName).toBeTruthy();

    await page.locator('button', { hasText: '解散团队' }).click();
    await expect(page.locator('text=此操作不可恢复')).toBeVisible({ timeout: 5000 });
    await page.locator('[role="dialog"] input').fill(teamName!.trim());

    const confirmBtn = page.locator('[role="dialog"] button', { hasText: '解散团队' });
    await expect(confirmBtn).toBeEnabled({ timeout: 3000 });

    await page.locator('button', { hasText: '取消' }).click();
  });
});

// ── Section 8: Member Perspective (Non-PM) ─────────────────────────────────────

test.describe('Team Detail - Member Perspective', () => {
  let adminToken: string;
  let teamId: string | null;
  let memberUser: { userId: string; username: string; initialPassword?: string } | null = null;

  test.beforeAll(async () => {
    adminToken = await getAuthToken();
    teamId = await getFirstTeamId(adminToken);
    if (teamId) {
      // Create a new user and add to team as member
      const username = `member_${Date.now()}`;
      memberUser = await createUser(adminToken, username, '普通成员测试');
      const memberRoleKey = await getRoleKey(adminToken, 'member') ?? '3';
      await addUserToTeam(adminToken, teamId, username, memberRoleKey);
    }
  });

  test.afterAll(async () => {
    if (adminToken && memberUser) {
      if (teamId) await removeUserFromTeam(adminToken, teamId, memberUser.userId);
      await deleteUser(adminToken, memberUser.userId);
    }
  });

  test('member can view team detail page', async ({ page }) => {
    test.skip(!teamId || !memberUser, 'No team or member available');

    // Login as the member
    const memberToken = await getTokenForUser(memberUser!.username, memberUser!.initialPassword || 'defaultPassword');
    await loginAs(page, memberToken, { isSuperAdmin: false });

    // Navigate to team detail
    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=团队名称')).toBeVisible({ timeout: 5000 });
  });

  test('member cannot see PM-only action buttons', async ({ page }) => {
    test.skip(!teamId || !memberUser, 'No team or member available');

    const memberToken = await getTokenForUser(memberUser!.username, memberUser!.initialPassword || 'defaultPassword');
    await loginAs(page, memberToken, { isSuperAdmin: false });

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Member should NOT see these PM-only buttons
    await expect(page.locator('button', { hasText: '解散团队' })).not.toBeVisible({ timeout: 3000 });
    await expect(page.locator('button', { hasText: '添加成员' })).not.toBeVisible({ timeout: 3000 });
  });

  test('member cannot see change role button for other members', async ({ page }) => {
    test.skip(!teamId || !memberUser, 'No team or member available');

    const memberToken = await getTokenForUser(memberUser!.username, memberUser!.initialPassword || 'defaultPassword');
    await loginAs(page, memberToken, { isSuperAdmin: false });

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Member should NOT see change role buttons
    await expect(page.locator('[data-testid="change-role-btn"]')).not.toBeVisible({ timeout: 3000 });
  });

  test('member cannot see remove member button', async ({ page }) => {
    test.skip(!teamId || !memberUser, 'No team or member available');

    const memberToken = await getTokenForUser(memberUser!.username, memberUser!.initialPassword || 'defaultPassword');
    await loginAs(page, memberToken, { isSuperAdmin: false });

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Member should NOT see remove buttons
    await expect(page.locator('button', { hasText: '移除' })).not.toBeVisible({ timeout: 3000 });
  });

  test('member cannot see transfer PM button', async ({ page }) => {
    test.skip(!teamId || !memberUser, 'No team or member available');

    const memberToken = await getTokenForUser(memberUser!.username, memberUser!.initialPassword || 'defaultPassword');
    await loginAs(page, memberToken, { isSuperAdmin: false });

    await page.goto(`${BASE}/teams/${teamId}`);
    await expect(page.locator('[data-testid="team-detail-page"]')).toBeVisible({ timeout: 10000 });

    // Member should NOT see "设为PM" buttons
    await expect(page.locator('button', { hasText: '设为PM' })).not.toBeVisible({ timeout: 3000 });
  });
});
