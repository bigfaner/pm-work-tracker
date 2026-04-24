import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, parseApiData } from './test-helpers';

const TIMEOUT = 120000;

test.setTimeout(TIMEOUT);

test.describe.serial('进度追加 - 自动状态流转', () => {
  let authToken: string;
  let teamId: string;
  let testMainItemId: string;
  let subItem1Id: string; // for first-progress -> progressing test
  let subItem2Id: string; // for 100% -> completed test

  // ====== SETUP ======
  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    // Login
    authToken = await getAuthToken();

    // Get teams
    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].id);

    // Create main item
    const mainRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { title: 'E2E自动状态测试-主事项', priority: 'P2', assigneeId: 1, startDate: '2026-04-19', expectedEndDate: '2026-05-19' },
    });
    const mainData = parseApiData(await mainRes.json());
    testMainItemId = String(mainData.id);

    // Create sub-item 1 (for first-progress test)
    const sub1Res = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mainItemId: Number(testMainItemId), title: 'E2E自动状态-首次进度', priority: 'P2', assigneeId: 1 },
    });
    const sub1Data = parseApiData(await sub1Res.json());
    subItem1Id = String(sub1Data.id);

    // Create sub-item 2 (for 100% completion test)
    const sub2Res = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mainItemId: Number(testMainItemId), title: 'E2E自动状态-100%完成', priority: 'P2', assigneeId: 1 },
    });
    const sub2Data = parseApiData(await sub2Res.json());
    subItem2Id = String(sub2Data.id);

    // Transition sub-item 2 to progressing first (so 100% can reach completed)
    await request.put(`/v1/teams/${teamId}/sub-items/${subItem2Id}/status`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { status: 'progressing' },
    });

    await request.dispose();
    console.log(`Setup: team=${teamId}, main=${testMainItemId}, sub1=${subItem1Id}, sub2=${subItem2Id}`);
  });

  // ====== TEST 1: First progress on pending sub-item -> progressing ======
  test('首次追加进度，pending 自动变为 progressing', async ({ page }) => {
    await login(page);

    // Navigate to sub-item detail page
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${subItem1Id}`);
    await page.waitForTimeout(2000);

    // Verify initial status is "待开始"
    const statusBadge = page.locator('[data-testid="sub-item-detail-page"]').locator('text=待开始').first();
    await expect(statusBadge).toBeVisible({ timeout: 5000 });

    // Open append progress dialog
    const appendBtn = page.locator('button:has-text("追加进度")');
    await expect(appendBtn).toBeVisible();
    await appendBtn.click();
    await page.waitForTimeout(500);

    // Fill in progress
    const completionInput = page.locator('.dialog input[type="number"], [role="dialog"] input[type="number"]').first();
    await expect(completionInput).toBeVisible();
    await completionInput.fill('30');

    // Fill in achievement
    const achievementTextarea = page.locator('[role="dialog"] textarea').first();
    await achievementTextarea.fill('完成了初步设计');

    // Submit
    const submitBtn = page.locator('[role="dialog"] button:has-text("提交")');
    await submitBtn.click();
    await page.waitForTimeout(2000);

    // Verify status changed to "进行中"
    const newStatusBadge = page.locator('[data-testid="sub-item-detail-page"]').locator('text=进行中').first();
    await expect(newStatusBadge).toBeVisible({ timeout: 5000 });
  });

  // ====== TEST 2: 100% completion on progressing sub-item -> completed ======
  test('追加100%进度，progressing 自动变为 completed', async ({ page }) => {
    await login(page);

    // Navigate to sub-item detail page
    await page.goto(`${BASE}/items/${testMainItemId}/sub/${subItem2Id}`);
    await page.waitForTimeout(2000);

    // Verify initial status is "进行中"
    const statusBadge = page.locator('[data-testid="sub-item-detail-page"]').locator('text=进行中').first();
    await expect(statusBadge).toBeVisible({ timeout: 5000 });

    // Open append progress dialog
    const appendBtn = page.locator('button:has-text("追加进度")');
    await expect(appendBtn).toBeVisible();
    await appendBtn.click();
    await page.waitForTimeout(500);

    // Fill in progress = 100
    const completionInput = page.locator('[role="dialog"] input[type="number"]').first();
    await expect(completionInput).toBeVisible();
    await completionInput.fill('100');

    // Fill in achievement
    const achievementTextarea = page.locator('[role="dialog"] textarea').first();
    await achievementTextarea.fill('全部完成');

    // Submit
    const submitBtn = page.locator('[role="dialog"] button:has-text("提交")');
    await submitBtn.click();
    await page.waitForTimeout(2000);

    // Verify status changed to "已完成"
    const newStatusBadge = page.locator('[data-testid="sub-item-detail-page"]').locator('text=已完成').first();
    await expect(newStatusBadge).toBeVisible({ timeout: 5000 });
  });

  // ====== TEST 3: Verify API-level auto-transition for pending -> 100% ======
  test('API: pending 子事项首次追加100%直接变为 completed', async ({ request }) => {
    // Create a new sub-item via API
    const subRes = await request.post(`/v1/teams/${teamId}/main-items/${testMainItemId}/sub-items`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { mainItemId: Number(testMainItemId), title: 'API自动状态-pending直接100%', priority: 'P2', assigneeId: 1 },
    });
    const subData = parseApiData(await subRes.json());
    const subId = subData.id;

    // Verify initial status is pending
    expect(subData.status).toBe('pending');

    // Append progress with 100%
    const progressRes = await request.post(`/v1/teams/${teamId}/sub-items/${subId}/progress`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { completion: 100, achievement: '直接完成' },
    });
    expect(progressRes.ok()).toBeTruthy();

    // Fetch sub-item and verify status is completed
    const fetchRes = await request.get(`/v1/teams/${teamId}/sub-items/${subId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const fetchData = parseApiData(await fetchRes.json());
    expect(fetchData.status).toBe('completed');
    expect(fetchData.completion).toBe(100);
  });

});
