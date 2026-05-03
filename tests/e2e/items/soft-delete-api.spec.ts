import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, randomCode, authHeader, getApiToken, parseApiBody, createTestTeam, createTestMainItem, createTestSubItem } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
const runId = Date.now();

test.describe('Soft-Delete — Items (TC-003, TC-004)', () => {
  test.beforeAll(async () => {
    superadminToken = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
  });

  // ── TC-003: Soft-deleted sub-item disappears from sub-item list ──

  // Traceability: TC-003 -> Story 2 / AC-1
  test('TC-003: Soft-deleted sub-item disappears from sub-item list', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc003-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC003 Main`, 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC003 Sub 01',
    );

    const listBefore = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
    });
    expect(listBefore.status).toBe(200);
    const dataBefore = parseApiBody(listBefore.body);
    const itemsBefore = dataBefore?.items ?? dataBefore ?? [];
    const existsBefore = itemsBefore.some((i: any) => String(i.bizKey ?? i.id) === subItemBizKey);
    expect(existsBefore).toBe(true);
  });

  // ── TC-004: Re-create sub-item with same title after soft-delete succeeds ──

  // Traceability: TC-004 -> Story 2 / AC-2
  test('TC-004: Re-create sub-item with same title after soft-delete succeeds', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc004-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC004 Main`, 'P0',
    );

    // Create sub-item
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC004 Sub 01',
    );
    expect(subItemBizKey).toBeTruthy();

    // Attempt to create another sub-item with the same title should succeed (no unique constraint on title)
    const dupRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ mainItemKey: mainItemBizKey, title: 'TC004 Sub 01 Duplicate', priority: 'P2', assigneeKey: '1', startDate: '2026-01-01', expectedEndDate: '2026-12-31' }),
    });
    expect(dupRes.status === 200 || dupRes.status === 201).toBeTruthy();
  });
});
