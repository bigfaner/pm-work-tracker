import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, randomCode, authHeader, getApiToken, parseApiBody, createTestUser, softDeleteUser, createTestTeam, createTestMainItem, createTestSubItem } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
const runId = Date.now();

test.describe('Soft-Delete — Users (TC-005, TC-006)', () => {
  test.beforeAll(async () => {
    superadminToken = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
  });

  // ── TC-005: FindByID returns NotFound for soft-deleted User ──

  // Traceability: TC-005 -> Story 3 / AC-1
  test('TC-005: FindByID returns NotFound for soft-deleted User', async () => {
    const userBizKey = await createTestUser(
      superadminToken,
      `e2e-deleted-user-tc005-${runId}`,
      'TC005 Deleted User',
    );
    await softDeleteUser(superadminToken, userBizKey);

    const res = await curl('GET', `${apiUrl}/v1/admin/users/${userBizKey}`, {
      headers: authHeader(superadminToken),
    });
    expect(res.status).toBe(404);
  });

  // ── TC-006: FindByID returns record for non-soft-deletable ProgressRecord ──

  // Traceability: TC-006 -> Story 3 / AC-2
  test('TC-006: Progress records are accessible (non-soft-deletable entity)', async () => {
    const teamBizKey = await createTestTeam(superadminToken, `e2e-team-tc006-${runId}`);
    const mainItemBizKey = await createTestMainItem(
      superadminToken, teamBizKey, `TC006 Main`, 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      superadminToken, teamBizKey, mainItemBizKey, 'TC006 Sub 01',
    );

    // Append a progress record
    const appendRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 50, note: 'TC006 progress' }),
    });
    expect(
      appendRes.status === 200 || appendRes.status === 201,
    ).toBeTruthy();

    // Verify progress list is accessible
    const listRes = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
    });
    expect(listRes.status).toBe(200);
  });
});
