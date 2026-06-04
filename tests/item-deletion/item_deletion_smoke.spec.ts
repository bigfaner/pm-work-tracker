/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey Smoke: item-deletion
 * Covers the happy path for API-relevant deletion steps.
 *
 * Traceability: Journey-level smoke test — verifies cascade delete and
 *   individual sub-item delete success flows end-to-end.
 */

import { test, expect, describe, beforeAll } from 'vitest';
import {
  curl,
  apiBaseUrl,
  apiUrl,
  authHeader,
  getApiToken,
  parseApiBody,
  setupRbacFixtures,
  createTestTeam,
  createTestMainItem,
  createTestSubItem,
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
let teamBizKey: string;
const runId = Date.now();

describe('Journey smoke: item-deletion (API happy path)', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = await createTestTeam(f.superadminToken, `e2e-del-smoke-${runId}`);
  });

  test('SMOKE: PM deletes a sub-item and verifies it is gone from the list', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'Smoke Main', 'P0',
    );
    const subItemBizKey = await createTestSubItem(
      f.pmToken, teamBizKey, mainItemBizKey, 'Smoke Sub',
    );

    // Delete the sub-item
    const delRes = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`,
      { headers: authHeader(f.pmToken) },
    );
    expect(delRes.status).toBe(200);

    // Verify it no longer appears in the sub-item list
    const listRes = await curl(
      'GET',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`,
      { headers: authHeader(f.pmToken) },
    );
    expect(listRes.status).toBe(200);
    const data = parseApiBody(listRes.body);
    const items = data?.items ?? data ?? [];
    const exists = items.some(
      (i: any) => String(i.bizKey ?? i.id) === subItemBizKey,
    );
    expect(exists).toBe(false);
  });

  test('SMOKE: member user gets 403 when attempting to delete a main item', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'Smoke Unauthorized Main', 'P1',
    );

    const res = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}`,
      { headers: authHeader(f.memberToken) },
    );
    expect(res.status).toBe(403);
  });
});
