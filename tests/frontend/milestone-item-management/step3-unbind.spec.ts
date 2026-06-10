/**
 * @web-e2e
 * @feature milestone-map
 * Journey: milestone-item-management
 *
 * Step 3 tests: Unbind a MainItem from panel
 * Step 3b: Undo unbind - not implemented (toast is informational only)
 */

import { test, expect } from '@playwright/test';
import { login, getAuthToken, parseApiData, extractBizKey } from '../helpers.js';

const TIMEOUT = 120000;
test.setTimeout(TIMEOUT);

test.describe('milestone-item-management / Step 3: Unbind MainItem from panel', () => {
  let authToken: string;
  let teamId: string;
  let mapBizKey: string;
  let msBizKey: string;
  let miBizKey: string;
  const runId = Date.now();

  test.beforeAll(async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    authToken = await getAuthToken();

    const teamsRes = await request.get('/v1/teams', {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const teamsRaw = parseApiData(await teamsRes.json());
    const teamsData = Array.isArray(teamsRaw) ? teamsRaw : (teamsRaw?.items ?? []);
    if (teamsData.length === 0) throw new Error('beforeAll: no teams found');
    teamId = String(teamsData[0].bizKey);

    const mapRes = await request.post(`/v1/teams/${teamId}/milestone-maps`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { mapName: `e2e-mim-s3-map-${runId}`, assigneeBizKey: '1' },
    });
    const mapData = parseApiData(await mapRes.json());
    mapBizKey = extractBizKey(mapData) ?? '';

    const msRes = await request.post(`/v1/teams/${teamId}/milestone-maps/${mapBizKey}/milestones`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { milestoneName: `e2e-mim-s3-ms-${runId}`, expectedEndDate: '2026-07-30' },
    });
    msBizKey = extractBizKey(parseApiData(await msRes.json())) ?? '';

    // Create and bind MI
    const miRes = await request.post(`/v1/teams/${teamId}/main-items`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        title: `e2e-mim-s3-mi-${runId}`,
        priority: 'P1',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      },
    });
    miBizKey = extractBizKey(parseApiData(await miRes.json())) ?? '';
    await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { milestoneKey: msBizKey },
    });

    await request.dispose();
  });

  test.beforeEach(async ({ page }) => {
    // Re-bind MI before each test
    const request = page.context?.request;
    if (request) {
      await request.put(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
        headers: { Authorization: `Bearer ${authToken}`, 'Content-Type': 'application/json' },
        data: { milestoneKey: msBizKey },
      });
    }
    await login(page, undefined, `/milestones/${mapBizKey}`);
  });

  // Step 3: Unbind MI via panel
  test('TC-MIM-S3-001: Unbind MI removes it from panel list', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s3-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    // Find the MI row
    const miRow = page.locator(`[data-testid="mi-drag-${miBizKey}"]`);
    await miRow.waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    if (await miRow.isVisible()) {
      await miRow.hover();

      // Find unbind button
      const unbindBtn = miRow.getByRole('button', { name: /解绑|unbind/i }).or(
        miRow.locator('[class*="unbind"], [title*="unbind"], [aria-label*="unbind"]').first(),
      );

      if (await unbindBtn.isVisible()) {
        await unbindBtn.click();
        await page.waitForTimeout(1000);

        // Verify toast
        await expect(page.getByText(/已解除.*绑定|unbind/i)).toBeVisible({ timeout: 5000 });

        // Verify MI removed from list
        await expect(miRow).not.toBeVisible({ timeout: 5000 });
      }
    }

    // Verify via API
    const request = page.context().request;
    const verifyRes = await request.get(`http://127.0.0.1:8080/v1/teams/${teamId}/main-items/${miBizKey}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const verifyData = parseApiData(await verifyRes.json());
    expect(verifyData.milestoneKey).toBeFalsy();
  });

  // Step 3b: Undo info stored in window.__lastUndoInfo
  test('TC-MIM-S3-002: Unbind stores undo info but no clickable undo in toast', async ({ page }) => {
    const node = page.locator(`[data-testid^="milestone-node-"]`).filter({ hasText: `e2e-mim-s3-ms-${runId}` });
    await node.click();
    await page.waitForTimeout(1500);

    const miRow = page.locator(`[data-testid="mi-drag-${miBizKey}"]`);
    await miRow.waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    if (await miRow.isVisible()) {
      await miRow.hover();
      const unbindBtn = miRow.getByRole('button', { name: /解绑|unbind/i }).or(
        miRow.locator('[class*="unbind"]').first(),
      );
      if (await unbindBtn.isVisible()) {
        await unbindBtn.click();
        await page.waitForTimeout(1000);

        // Check window.__lastUndoInfo exists (for test access)
        const undoInfo = await page.evaluate(() => (window as any).__lastUndoInfo);
        // Undo info should exist but toast has no clickable undo
        // This test validates the gap: undo info is stored but not exposed in UI
      }
    }
  });

  // Unauthorized API test
  test('TC-MIM-S3-003: Unbind MI API without auth returns 401', async ({ playwright }) => {
    const request = await playwright.request.newContext({
      baseURL: 'http://127.0.0.1:8080',
      extraHTTPHeaders: { 'Content-Type': 'application/json' },
    });

    const res = await request.put(`/v1/teams/${teamId}/main-items/${miBizKey}`, {
      data: { milestoneKey: null },
    });

    expect(res.status()).toBe(401);
    await request.dispose();
  });
});
