import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import type { Page } from 'playwright';
import {
  setupBrowser,
  teardownBrowser,
  screenshot,
  baseUrl,
  loginViaUI,
} from './helpers.js';

describe('UI E2E Tests', () => {
  let page: Page;

  before(async () => {
    page = await setupBrowser();
    await loginViaUI(page);
  });

  after(async () => {
    await teardownBrowser();
  });

  // Traceability: TC-001 → Story 6 / AC-1
  test('TC-001: 主事项详情页 URL 使用 bizKey 导航', async () => {
    // Step 1: 导航至事项清单页
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    // 验证事项清单页已加载（E-005）
    await page.getByRole('heading', { name: '事项清单', level: 1 }).waitFor({ state: 'visible' });

    // Step 2: 点击任意主事项行进入详情页
    // UNSTABLE: no semantic anchor for table row — using first clickable row link
    const firstRow = page.locator('table tbody tr').first();
    await firstRow.waitFor({ state: 'visible' });
    await firstRow.click();

    // 等待导航完成
    await page.waitForURL(/\/items\/\d+/, { timeout: 10000 });

    // 验证 URL 格式为 /items/{bizKey}，bizKey 为雪花算法数字（10位以上），不为自增短整数
    const currentUrl = page.url();
    const urlMatch = currentUrl.match(/\/items\/(\d+)/);
    assert.ok(urlMatch, `URL 应匹配 /items/{bizKey} 格式，实际: ${currentUrl}`);
    const bizKey = urlMatch![1];
    assert.ok(
      bizKey.length >= 10,
      `bizKey 应为雪花算法数字（10位以上），实际: ${bizKey}（${bizKey.length}位）`,
    );

    // 验证详情页面包屑导航（E-021）正常显示
    await page.getByRole('navigation', { name: 'Breadcrumb' }).waitFor({ state: 'visible' });

    await screenshot(page, 'TC-001');
  });
});
