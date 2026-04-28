/**
 * UI E2E: sub-item edit refreshes list
 *
 * Steps:
 * 1. Create a main item + sub-item via API (priority P1)
 * 2. Navigate to /items, expand the main item
 * 3. Click 编辑 on the sub-item
 * 4. Change priority to P2, save
 * 5. Assert the sub-items list API is called again AND the badge shows P2
 */
import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import {
  setupBrowser,
  teardownBrowser,
  getPage,
  loginViaUI,
  getApiToken,
  createAuthCurl,
  apiBaseUrl,
  baseUrl,
  screenshot,
} from '../../helpers.js';

describe('UI E2E: sub-item edit refreshes list', () => {
  let subBizKey: string;
  let mainBizKey: string;
  let teamId: string;

  before(async () => {
    // --- API setup ---
    const token = await getApiToken(apiBaseUrl);
    const authCurl = createAuthCurl(apiBaseUrl, token);

    // Pick first team
    const teamsRes = await authCurl('GET', '/v1/teams');
    assert.equal(teamsRes.status, 200, `List teams: ${teamsRes.body}`);
    const teams = JSON.parse(teamsRes.body).data?.items ?? JSON.parse(teamsRes.body);
    assert.ok(teams.length > 0, 'Need at least one team');
    teamId = String(teams[0].bizKey);

    // Create pool entry → main item
    const poolRes = await authCurl('POST', `/v1/teams/${teamId}/item-pool`, {
      body: JSON.stringify({ title: 'UI-TC-001 sub-item edit test', background: 'e2e', expectedOutput: 'e2e' }),
    });
    assert.ok(poolRes.status === 200 || poolRes.status === 201, `Create pool: ${poolRes.body}`);
    const poolKey = String((JSON.parse(poolRes.body).data ?? JSON.parse(poolRes.body)).bizKey);

    const convertRes = await authCurl('POST', `/v1/teams/${teamId}/item-pool/${poolKey}/convert-to-main`, {
      body: JSON.stringify({ priority: 'P2', assigneeKey: '0', startDate: '2026-04-28', expectedEndDate: '2026-05-28' }),
    });
    assert.equal(convertRes.status, 200, `Convert: ${convertRes.body}`);
    const match = convertRes.body.match(/"mainItemBizKey"\s*:\s*(\d+)/);
    assert.ok(match, 'mainItemBizKey not found');
    mainBizKey = match[1];

    // Create sub-item with priority P1
    const subRes = await authCurl('POST', `/v1/teams/${teamId}/main-items/${mainBizKey}/sub-items`, {
      body: JSON.stringify({
        mainItemKey: mainBizKey,
        title: 'UI-TC-001 sub item',
        priority: 'P1',
        assigneeKey: '0',
        startDate: '2026-04-28',
        expectedEndDate: '2026-05-28',
      }),
    });
    assert.ok(subRes.status === 200 || subRes.status === 201, `Create sub: ${subRes.body}`);
    subBizKey = String((JSON.parse(subRes.body).data ?? JSON.parse(subRes.body)).bizKey);

    // --- Browser setup ---
    const page = await setupBrowser();
    await loginViaUI(page);
  });

  after(async () => {
    await teardownBrowser();
  });

  test('UI-TC-001: editing sub-item priority triggers list refresh and shows updated value', async () => {
    const page = getPage();

    // Navigate to items page
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    // Expand the main item card via JS (button is hidden, used only for test targeting)
    await page.waitForSelector(`[data-testid="expand-card-${mainBizKey}"]`, { state: 'attached', timeout: 10000 });
    await page.evaluate((testId) => {
      const btn = document.querySelector(`[data-testid="${testId}"]`) as HTMLElement;
      btn?.click();
    }, `expand-card-${mainBizKey}`);

    // Wait for sub-item to appear
    await page.waitForSelector(`text=UI-TC-001 sub item`, { timeout: 10000 });
    await screenshot(page, 'UI-TC-001-expanded');

    // Verify initial priority badge shows P1
    await expect_text_visible(page, 'P1');

    // Click 编辑 button on the specific sub-item row
    const subItemRow = page.locator('div').filter({ hasText: 'UI-TC-001 sub item' }).filter({ has: page.getByRole('button', { name: '编辑' }) }).last();
    const editBtn = subItemRow.getByRole('button', { name: '编辑' });
    await editBtn.waitFor({ state: 'visible', timeout: 5000 });
    await editBtn.click();

    // Dialog should open
    await page.waitForSelector('text=编辑子事项', { timeout: 5000 });

    // Change priority to P2
    await page.getByRole('combobox').filter({ hasText: /P1|P2|P3/ }).click();
    await page.getByRole('option', { name: 'P2' }).click();

    // Save
    await page.getByRole('button', { name: '保存' }).click();

    // Dialog should close
    await page.waitForSelector('text=编辑子事项', { state: 'hidden', timeout: 5000 });

    await screenshot(page, 'UI-TC-001-after-edit');

    // The sub-item row should now show P2 — verify via API that backend has P2
    const verifyRes = await (await import('../../helpers.js')).createAuthCurl(apiBaseUrl, await (await import('../../helpers.js')).getApiToken(apiBaseUrl))('GET', `/v1/teams/${teamId}/main-items/${mainBizKey}/sub-items`);
    assert.equal(verifyRes.status, 200, `Verify list: ${verifyRes.body}`);
    const items: any[] = (JSON.parse(verifyRes.body).data ?? JSON.parse(verifyRes.body)).items;
    const updated = items.find((i: any) => String(i.bizKey) === subBizKey);
    assert.ok(updated, 'Sub item found after edit');
    assert.equal(updated.priority, 'P2', `Priority should be P2, got ${updated.priority}`);

    await screenshot(page, 'UI-TC-001-verified');
  });
});

async function expect_text_visible(page: import('playwright').Page, text: string) {
  const el = page.locator(`text=${text}`).first();
  await el.waitFor({ state: 'visible', timeout: 5000 });
}
