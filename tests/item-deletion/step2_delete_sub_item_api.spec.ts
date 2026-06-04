/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey: item-deletion
 * Contract: step-2-delete-individual-sub-item
 *
 * Traceability: Step 2 — API outcomes for individual sub-item delete
 *   Covers: success (delete-last-sub-item), unauthorized, not-found
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

describe('API: Delete individual sub-item', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = await createTestTeam(f.superadminToken, `e2e-del-sub-${runId}`);
  });

  // ── Outcome "delete-last-sub-item" ──────────────────────────────────
  // Traceability: TC-API-DEL-003 -> Step 2 Outcome "delete-last-sub-item"
  test('TC-API-DEL-003: deleting last sub-item leaves main item with zero sub-items', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'DelLast Main', 'P1',
    );
    const subItemBizKey = await createTestSubItem(
      f.superadminToken, teamBizKey, mainItemBizKey, 'Only Sub',
    );

    // Delete the sub-item
    const delRes = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`,
      { headers: authHeader(f.pmToken) },
    );
    expect(delRes.status).toBe(200);

    // Verify sub-item list is now empty
    const listRes = await curl(
      'GET',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`,
      { headers: authHeader(f.pmToken) },
    );
    expect(listRes.status).toBe(200);
    const data = parseApiBody(listRes.body);
    const items = data?.items ?? data ?? [];
    expect(items.length).toBe(0);
  });

  // ── Outcome "unauthorized" ──────────────────────────────────────────
  // Traceability: TC-API-DEL-004 -> Step 2 Outcome "unauthorized"
  test('TC-API-DEL-004: member without sub_item:delete gets 403', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'Unauthorized Main', 'P2',
    );
    const subItemBizKey = await createTestSubItem(
      f.superadminToken, teamBizKey, mainItemBizKey, 'Unauthorized Sub',
    );

    const res = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`,
      { headers: authHeader(f.memberToken) },
    );
    expect(res.status).toBe(403);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Outcome "not-found" ─────────────────────────────────────────────
  // Traceability: TC-API-DEL-005 -> Step 2 Outcome "not-found"
  test('TC-API-DEL-005: delete non-existent sub-item returns 404', async () => {
    const fakeSubKey = '9999999999999999';
    const res = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${fakeSubKey}`,
      { headers: authHeader(f.pmToken) },
    );
    expect(res.status).toBe(404);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });
});
