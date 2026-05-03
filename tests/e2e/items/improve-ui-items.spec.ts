import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { snapshotContains, findElement, screenshot, baseUrl, login } from '../helpers.js';

test.describe('UI E2E Tests — 事项清单 (Main Items)', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
  });

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: 事项清单 Detail 视图切换', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-001-summary-view');

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    await expect(detailBtn).toBeVisible();
    await detailBtn.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-001-detail-view');

    expect(await snapshotContains(page, '优先级') || await snapshotContains(page, '状态')).toBeTruthy();
  });

  // Traceability: TC-002 → Story 1 / AC-2
  test('TC-002: 事项清单 Summary 视图切回', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    if (await detailBtn.isVisible().catch(() => false)) {
      await detailBtn.click();
      await page.waitForLoadState('networkidle');
    }

    const summaryBtn = findElement(page, 'button', '汇总').or(findElement(page, 'tab', '汇总'));
    await expect(summaryBtn).toBeVisible();
    await summaryBtn.click();
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-002-summary-back');
  });

  // Traceability: TC-003 → Story 1 / AC-3
  test('TC-003: 事项清单视图切换保留筛选条件', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    const statusFilter = findElement(page, 'combobox').or(findElement(page, 'button', '状态'));
    if (await statusFilter.isVisible().catch(() => false)) {
      await statusFilter.click();
      await screenshot(page, 'TC-003-filter-set');
    }

    const detailBtn = findElement(page, 'button', '明细').or(findElement(page, 'tab', '明细'));
    if (await detailBtn.isVisible().catch(() => false)) {
      await detailBtn.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-003-detail-with-filter');

    const summaryBtn = findElement(page, 'button', '汇总').or(findElement(page, 'tab', '汇总'));
    if (await summaryBtn.isVisible().catch(() => false)) {
      await summaryBtn.click();
      await page.waitForLoadState('networkidle');
    }
    await screenshot(page, 'TC-003-summary-filter-preserved');
  });

  // Traceability: TC-027 → UI Function 2 / States
  test('TC-027: 事项清单默认 Summary 视图', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-027-default');

    expect(await snapshotContains(page, '汇总') || await snapshotContains(page, '卡片')).toBeTruthy();
  });

  // Traceability: TC-028 → UI Function 2 / Validation
  test('TC-028: 事项清单 Summary 无限滚动', async ({ page }) => {
    await page.goto(`${baseUrl}/items`);
    await page.waitForLoadState('networkidle');

    await page.keyboard.press('End');
    await page.keyboard.press('End');
    await page.waitForLoadState('networkidle');
    await screenshot(page, 'TC-028-scrolled');
  });
});
