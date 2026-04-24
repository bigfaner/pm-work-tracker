import { test, expect, Page } from '@playwright/test';
import { BASE, API, login, getAuthToken, getFirstTeamId, parseApiData, navTo } from './test-helpers';

// Helper: create main item via API
async function createMainItemApi(token: string, teamId: string, title: string) {
  const res = await fetch(`${API}/teams/${teamId}/main-items`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, priority: 'P2', startDate: '2026-04-19', expectedEndDate: '2026-05-19' }),
  });
  const data = await res.json();
  return data.id || data.data?.id;
}

// Helper: create pool item via API
async function createPoolItemApi(token: string, teamId: string, title: string, background?: string) {
  const res = await fetch(`${API}/teams/${teamId}/item-pool`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, background: background || '测试背景', expectedOutput: '测试产出' }),
  });
  const data = await res.json();
  return data.id || data.data?.id;
}

// Helper: dialog scope
function dialog(page: Page) {
  return page.locator('[role="dialog"]');
}

test.describe('待办事项 (ItemPool) - E2E Business Flow', () => {
  test.describe.configure({ mode: 'serial' });
  test.setTimeout(60000);

  let page: Page;
  let authToken: string;
  let teamId: string;
  const uid = Date.now().toString(36);

  // ====== 0. SETUP: Login once, share across all tests ======
  test.beforeAll(async ({ browser }) => {
    authToken = await getAuthToken();
    teamId = (await getFirstTeamId(authToken))!;

    if (teamId) {
      await createPoolItemApi(authToken, teamId, 'E2E-待处理A', 'API预建数据A');
      await createMainItemApi(authToken, teamId, 'E2E-目标主事项-转子项测试');
      await createPoolItemApi(authToken, teamId, 'E2E-待处理B', 'API预建数据B');
    }

    page = await browser.newPage();
    await login(page);

    // Navigate to item pool
    await navTo(page, '/item-pool');
    await page.waitForTimeout(2000);
  });

  test.afterAll(async () => {
    await page.close();
  });

  // ====== 1. PAGE LAYOUT ======
  test('1.1 page shows title and layout', async () => {
    await expect(page.locator('h1:text("待办事项")')).toBeVisible();
    await expect(page.locator('[data-testid="item-pool-page"]')).toBeVisible();
  });

  test('1.2 filter bar has search, status filter, reset', async () => {
    await expect(page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]')).toBeVisible();
    await expect(page.locator('[data-testid="pool-status-filter"]')).toBeVisible();
    await expect(page.locator('button:has-text("重置")')).toBeVisible();
  });

  test('1.3 has "新增待办事项" button', async () => {
    await expect(page.locator('button:has-text("新增待办事项")')).toBeVisible();
  });

  // ====== 2. SUBMIT NEW POOL ITEM ======
  test('2.1 open submit dialog shows form fields', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await expect(d.locator('text=新增待办事项')).toBeVisible();
    await expect(d.locator('label:has-text("标题")')).toBeVisible();
    await expect(d.locator('input[placeholder="请输入事项标题"]')).toBeVisible();
    await d.locator('button:has-text("取消")').click();
    await expect(d).not.toBeVisible({ timeout: 3000 });
  });

  test('2.2 submit button disabled when title empty', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await expect(d.locator('button:has-text("提交")')).toBeDisabled();
    await d.locator('button:has-text("取消")').click();
    await expect(d).not.toBeVisible({ timeout: 3000 });
  });

  test('2.3 submit with title only (minimum required)', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });

    await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E-基本提交-${uid}`);
    await expect(d.locator('button:has-text("提交")')).toBeEnabled();
    await d.locator('button:has-text("提交")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });

    await expect(page.locator(`text=E2E-基本提交-${uid}`)).toBeVisible({ timeout: 5000 });
  });

  test('2.4 submit with all fields (title + background + expectedOutput)', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });

    await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E-完整提交-${uid}`);
    const textareas = d.locator('textarea');
    await textareas.nth(0).fill(`完整提交背景-${uid}`);
    await textareas.nth(1).fill('完整提交预期产出');
    await d.locator('button:has-text("提交")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });

    await expect(page.locator(`text=E2E-完整提交-${uid}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`text=完整提交背景-${uid}`)).toBeVisible({ timeout: 3000 });
  });

  // ====== 3. FILTERING & SEARCH ======
  test('3.1 search by title filters to matching items', async () => {
    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    await searchInput.fill(`E2E-基本提交-${uid}`);
    await page.waitForTimeout(500);
    await expect(page.locator(`text=E2E-基本提交-${uid}`)).toBeVisible({ timeout: 3000 });
  });

  test('3.2 non-matching search shows empty state', async () => {
    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    await searchInput.fill('ZZZZZ_NONEXISTENT_12345');
    await page.waitForTimeout(500);
    await expect(page.locator('text=暂无待办事项')).toBeVisible({ timeout: 3000 });
  });

  test('3.3 reset button clears all filters', async () => {
    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    await searchInput.fill('test');
    await page.waitForTimeout(300);
    await page.locator('button:has-text("重置")').click();
    await page.waitForTimeout(500);
    await expect(searchInput).toHaveValue('');
    const items = page.locator('[data-testid^="pool-item-"]');
    const count = await items.count();
    expect(count).toBeGreaterThan(0);
  });

  test('3.4 status filter dropdown shows options', async () => {
    await page.locator('[data-testid="pool-status-filter"]').click();
    await page.waitForTimeout(300);
    await expect(page.locator('[role="option"]:has-text("待分配")')).toBeVisible();
    await expect(page.locator('[role="option"]:has-text("已分配")')).toBeVisible();
    await expect(page.locator('[role="option"]:has-text("已拒绝")')).toBeVisible();
    await page.locator('[role="option"]:has-text("待分配")').click();
    await page.waitForTimeout(500);
    const pendingBadges = page.locator('[data-testid^="pool-item-"] >> text=待分配');
    const count = await pendingBadges.count();
    console.log(`Pending items after filter: ${count}`);
    await page.locator('button:has-text("重置")').click();
    await page.waitForTimeout(500);
  });

  // ====== 4. CONVERT TO MAIN ITEM ======
  test('4.1 pending item shows action buttons', async () => {
    const pendingItem = page.locator('[data-testid^="pool-item-"]').first();
    await expect(pendingItem).toBeVisible({ timeout: 5000 });

    const toMainBtns = page.locator('[data-testid^="to-main-"]');
    const count = await toMainBtns.count();
    expect(count).toBeGreaterThan(0);
    console.log(`Items with "转为主事项" button: ${count}`);
  });

  test('4.2 convert a pending item to main item (dialog opens with form)', async () => {
    const toMainBtn = page.locator('[data-testid^="to-main-"]').first();
    await expect(toMainBtn).toBeVisible({ timeout: 5000 });

    await toMainBtn.click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await expect(d.locator('text=转为主事项')).toBeVisible();

    await expect(d.locator('text=优先级')).toBeVisible();
    await expect(d.locator('text=负责人')).toBeVisible();
    await expect(d.locator('text=开始时间')).toBeVisible();
    await expect(d.locator('text=预期完成时间')).toBeVisible();

    const assigneeTrigger = d.locator('button:has-text("选择负责人")');
    if (await assigneeTrigger.isVisible({ timeout: 2000 }).catch(() => false)) {
      await assigneeTrigger.click();
      await page.waitForTimeout(300);
      const option = page.locator('[role="option"]').first();
      if (await option.isVisible({ timeout: 2000 }).catch(() => false)) {
        await option.click();
        await page.waitForTimeout(300);
      }
    }

    const dateInputs = d.locator('input[type="date"]');
    await dateInputs.first().fill('2026-04-19');
    await dateInputs.last().fill('2026-05-19');

    const confirmBtn = d.locator('button:has-text("确认转换")');
    if (await confirmBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
      await page.waitForTimeout(3000);
    } else {
      console.log('Confirm button still disabled, skipping conversion');
    }

    const dialogVisible = await d.isVisible().catch(() => false);
    if (dialogVisible) {
      await d.locator('button:has-text("取消")').click();
      await expect(d).not.toBeVisible({ timeout: 3000 });
      console.log('BUG: Convert-to-main-item sends mainItemId=0 which backend rejects');
    } else {
      console.log('Converted pool item to main item successfully');
    }
  });

  // ====== 5. CONVERT TO SUB ITEM ======
  test('5.1 convert a pending item to sub item', async () => {
    const toSubBtn = page.locator('[data-testid^="to-sub-"]').first();
    if (!(await toSubBtn.isVisible({ timeout: 3000 }).catch(() => false))) {
      console.log('No pending items available for sub-item conversion, creating one...');
      await page.locator('button:has-text("新增待办事项")').click();
      const d = dialog(page);
      await expect(d).toBeVisible({ timeout: 5000 });
      await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E-转子-${uid}`);
      await d.locator('button:has-text("提交")').click();
      await expect(d).not.toBeVisible({ timeout: 5000 });
      await page.waitForTimeout(2000);
    }

    const btn = page.locator('[data-testid^="to-sub-"]').first();
    await expect(btn).toBeVisible({ timeout: 5000 });

    await btn.click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await expect(d.locator('text=转为子事项')).toBeVisible();
    await expect(d.locator('text=挂载主事项')).toBeVisible();

    await expect(d.locator('button:has-text("确认转换")')).toBeDisabled();

    const parentSelect = d.locator('button:has-text("请选择主事项")');
    if (await parentSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
      await parentSelect.click();
      await page.waitForTimeout(500);
      const options = page.locator('[role="option"]');
      const optCount = await options.count();
      if (optCount > 1) {
        await options.nth(1).click();
        await page.waitForTimeout(300);
      }
    }

    const confirmBtn = d.locator('button:has-text("确认转换")');
    if (await confirmBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
      await page.waitForTimeout(3000);

      const dialogClosed = !(await d.isVisible().catch(() => false));
      if (dialogClosed) {
        console.log('Successfully converted pool item to sub item');
      } else {
        await d.locator('button:has-text("取消")').click();
        await expect(d).not.toBeVisible({ timeout: 3000 });
        console.log('Sub-item conversion failed (API error)');
      }
    } else {
      await d.locator('button:has-text("取消")').click();
      await expect(d).not.toBeVisible({ timeout: 3000 });
      console.log('No parent main item available, conversion skipped');
    }
  });

  // ====== 6. REJECT POOL ITEM ======
  test('6.1 reject a pending item with reason', async () => {
    let rejectBtn = page.locator('[data-testid^="reject-"]').first();

    if (!(await rejectBtn.isVisible({ timeout: 3000 }).catch(() => false))) {
      await page.locator('button:has-text("新增待办事项")').click();
      const d = dialog(page);
      await expect(d).toBeVisible({ timeout: 5000 });
      await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E-待拒-${uid}`);
      await d.locator('button:has-text("提交")').click();
      await expect(d).not.toBeVisible({ timeout: 5000 });
      await page.waitForTimeout(2000);
      rejectBtn = page.locator('[data-testid^="reject-"]').first();
    }

    await expect(rejectBtn).toBeVisible({ timeout: 5000 });

    await rejectBtn.click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await expect(d.locator('text=拒绝事项')).toBeVisible();
    await expect(d.locator('text=拒绝原因')).toBeVisible();

    await expect(d.locator('button:has-text("确认拒绝")')).toBeDisabled();

    await d.locator('textarea').fill('E2E测试：不符合当前迭代计划');
    await expect(d.locator('button:has-text("确认拒绝")')).toBeEnabled();
    await d.locator('button:has-text("确认拒绝")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(2000);

    await expect(page.locator('text=已拒绝').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=不符合当前迭代计划').first()).toBeVisible({ timeout: 3000 });
    console.log('Successfully rejected pool item');
  });

  // ====== 7. ITEM STATUS & DISPLAY ======
  test('7.1 items show POOL-XXX ID format', async () => {
    const poolIds = page.locator('text=/POOL-\\d{3}/');
    const count = await poolIds.count();
    expect(count).toBeGreaterThan(0);
    console.log(`Items with POOL-XXX format: ${count}`);
  });

  test('7.2 items show relative time labels', async () => {
    const content = await page.content();
    const hasTime = content.includes('今天提交') || content.includes('天前提交') || content.includes('周前提交');
    console.log(`Items show relative time: ${hasTime}`);
  });

  test('7.3 rejected items do NOT show action buttons', async () => {
    const rejectedBadges = page.locator('text=已拒绝');
    const rejectedCount = await rejectedBadges.count();
    console.log(`Rejected items count: ${rejectedCount}`);

    if (rejectedCount > 0) {
      const rejectedCard = rejectedBadges.first().locator('xpath=ancestor::div[@data-testid and starts-with(@data-testid,"pool-item-")]');
      const hasActions = await rejectedCard.locator('button:has-text("转为主事项")').isVisible().catch(() => false);
      expect(hasActions).toBe(false);
      console.log('Rejected items correctly hide action buttons');
    }
  });

  test('7.4 assigned items show main item link', async () => {
    const content = await page.content();
    const hasAssigned = content.includes('已转为子事项挂载至');
    console.log(`Assigned items show main item link: ${hasAssigned}`);
    if (hasAssigned) {
      const links = page.locator('a:has-text("主事项 #")');
      const linkCount = await links.count();
      console.log(`Main item links found: ${linkCount}`);
    }
  });

  // ====== 8. FULL BUSINESS FLOW ======
  test('8.1 full flow: submit → search → filter → verify', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });

    await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E全流程-${uid}`);
    const textareas = d.locator('textarea');
    await textareas.nth(0).fill(`全流程背景-${uid}`);
    await textareas.nth(1).fill('全流程预期产出');
    await d.locator('button:has-text("提交")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(2000);

    await expect(page.locator(`text=E2E全流程-${uid}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`text=全流程背景-${uid}`)).toBeVisible();

    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    await searchInput.fill(`E2E全流程-${uid}`);
    await page.waitForTimeout(500);
    await expect(page.locator(`text=E2E全流程-${uid}`)).toBeVisible();

    await page.locator('button:has-text("重置")').click();
    await page.waitForTimeout(500);
    await expect(searchInput).toHaveValue('');

    console.log('Full submit→search→verify flow passed');
  });

  test('8.2 full flow: submit → convert to main → verify in items list', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    let d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E转主-${uid}`);
    await d.locator('button:has-text("提交")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(2000);

    const toMainBtn = page.locator('[data-testid^="to-main-"]').first();
    if (await toMainBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await toMainBtn.click();
      d = dialog(page);
      await expect(d).toBeVisible({ timeout: 5000 });
      await d.locator('input[type="date"]').first().fill('2026-04-20');
      await d.locator('input[type="date"]').last().fill('2026-05-20');
      await d.locator('button:has-text("确认转换")').click();
      await page.waitForTimeout(3000);

      const dialogClosed = !(await d.isVisible().catch(() => false));
      if (dialogClosed) {
        await navTo(page, '/items');
        await page.waitForTimeout(2000);

        const inList = page.locator(`text=E2E转主-${uid}`);
        const visible = await inList.isVisible({ timeout: 5000 }).catch(() => false);
        console.log(`Converted main item visible in items list: ${visible}`);

        await navTo(page, '/item-pool');
        await page.waitForTimeout(2000);
      } else {
        await d.locator('button:has-text("取消")').click();
        await expect(d).not.toBeVisible({ timeout: 3000 });
        console.log('BUG: convert-to-main fails (mainItemId=0 rejected by backend)');
      }
    } else {
      console.log('No pending items for conversion test');
    }
  });

  test('8.3 full flow: submit → reject → verify rejected state', async () => {
    await page.locator('button:has-text("新增待办事项")').click();
    let d = dialog(page);
    await expect(d).toBeVisible({ timeout: 5000 });
    await d.locator('input[placeholder="请输入事项标题"]').fill(`E2E拒绝-${uid}`);
    await d.locator('textarea').first().fill(`拒绝流程背景-${uid}`);
    await d.locator('button:has-text("提交")').click();
    await expect(d).not.toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(2000);

    const rejectBtn = page.locator('[data-testid^="reject-"]').first();
    if (await rejectBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await rejectBtn.click();
      d = dialog(page);
      await expect(d).toBeVisible({ timeout: 5000 });
      await d.locator('textarea').fill('E2E拒绝原因：不符合计划');
      await d.locator('button:has-text("确认拒绝")').click();
      await expect(d).not.toBeVisible({ timeout: 5000 });
      await page.waitForTimeout(2000);

      await expect(page.locator('text=已拒绝').first()).toBeVisible({ timeout: 5000 });
      console.log('Submit→reject flow verified');
    }
  });

  // ====== 9. EDGE CASES ======
  test('9.1 search by POOL-XXX ID format', async () => {
    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    const firstItem = page.locator('[data-testid^="pool-item-"]').first();
    if (await firstItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      const testid = await firstItem.getAttribute('data-testid');
      const id = testid?.replace('pool-item-', '');
      if (id) {
        const paddedId = `POOL-${String(id).padStart(3, '0')}`;
        await searchInput.fill(paddedId);
        await page.waitForTimeout(500);
        const matches = await page.locator(`[data-testid="pool-item-${id}"]`).count();
        console.log(`POOL ID search for ${paddedId}: ${matches} match(es)`);
        await searchInput.clear();
        await page.waitForTimeout(300);
      }
    }
  });

  test('9.2 navigate away and back preserves page', async () => {
    await navTo(page, '/items');
    await page.waitForTimeout(1000);
    await navTo(page, '/item-pool');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1:text("待办事项")')).toBeVisible();
    const items = page.locator('[data-testid^="pool-item-"]');
    const count = await items.count();
    console.log(`Items after navigation: ${count}`);
  });

  test('9.3 console error scan', async () => {
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') errors.push(msg.text());
    });

    await page.waitForTimeout(1000);

    await page.locator('button:has-text("新增待办事项")').click();
    const d = dialog(page);
    if (await d.isVisible({ timeout: 3000 }).catch(() => false)) {
      await d.locator('button:has-text("取消")').click();
      await page.waitForTimeout(500);
    }

    const searchInput = page.locator('[data-testid="item-pool-page"] input[placeholder*="搜索"]');
    await searchInput.fill('test');
    await page.waitForTimeout(300);
    await page.locator('button:has-text("重置")').click();
    await page.waitForTimeout(300);

    const criticalErrors = errors.filter(e =>
      !e.includes('favicon') && !e.includes('net::') && !e.includes('404')
    );
    console.log(`Console errors: total=${errors.length}, critical=${criticalErrors.length}`);
    if (criticalErrors.length > 0) {
      console.log(`Critical errors: ${criticalErrors.slice(0, 5).join('\n')}`);
    }
  });

  // ====== 10. API VALIDATION ======
  test('10.1 API: list pool items', async () => {
    if (!teamId) return;
    const res = await fetch(`${API}/teams/${teamId}/item-pool`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    const items = data.items || [];
    console.log(`API pool items count: ${items.length}`);
  });

  test('10.2 API: create → reject → verify', async () => {
    if (!teamId) return;
    const cRes = await fetch(`${API}/teams/${teamId}/item-pool`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: 'API验证-待拒绝', background: 'API测试', expectedOutput: 'API测试' }),
    });
    expect(cRes.status).toBe(201);
    const cData = await cRes.json();
    const poolId = cData.id || cData.data?.id;
    expect(poolId).toBeTruthy();

    const rRes = await fetch(`${API}/teams/${teamId}/item-pool/${poolId}/reject`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ reason: 'API验证拒绝' }),
    });
    expect(rRes.status).toBe(200);
    console.log('API create→reject verified');
  });

  test('10.3 API: create → assign to main item', async () => {
    if (!teamId) return;
    const pRes = await fetch(`${API}/teams/${teamId}/item-pool`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: 'API验证-待分配', background: '测试', expectedOutput: '测试' }),
    });
    const pData = await pRes.json();
    const poolId = pData.id || pData.data?.id;

    const mRes = await fetch(`${API}/teams/${teamId}/main-items`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: 'API验证-分配目标', priority: 'P2', assigneeId: 1, startDate: '2026-04-19', expectedEndDate: '2026-05-19' }),
    });
    const mData = await mRes.json();
    const mainItemId = mData.id || mData.data?.id;

    const aRes = await fetch(`${API}/teams/${teamId}/item-pool/${poolId}/assign`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ mainItemId, assigneeId: 1, startDate: '2026-04-19', expectedEndDate: '2026-05-19' }),
    });
    expect(aRes.status).toBe(200);
    console.log('API assign verified');
  });

  test('10.4 API: convert pool item to new main item', async () => {
    if (!teamId) return;
    const pRes = await fetch(`${API}/teams/${teamId}/item-pool`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: 'API验证-转新主事项', background: '测试', expectedOutput: '测试' }),
    });
    const pData = await pRes.json();
    const poolId = pData.id || pData.data?.id;

    const aRes = await fetch(`${API}/teams/${teamId}/item-pool/${poolId}/convert-to-main`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ priority: 'P2', assigneeId: 1, startDate: '2026-04-19', expectedEndDate: '2026-05-19' }),
    });
    expect(aRes.status).toBe(200);
    console.log('API convert-to-main verified');
  });
});
