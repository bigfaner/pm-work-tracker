/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey: item-deletion
 * Contract: step-1-delete-main-item-cascade
 *
 * Traceability: Step 1 — API outcomes for main item cascade delete
 *   Covers: unauthorized, not-found (API-surface outcomes)
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
  createTestMainItem,
  createTestSubItem,
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
let teamBizKey: string;
let mainItemBizKey: string;

describe('API: Delete main item (cascade)', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = f.teamBizKey;
    mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'DelCascade Main', 'P0',
    );
    // Create sub-items to verify cascade
    await createTestSubItem(f.pmToken, teamBizKey, mainItemBizKey, 'Sub 1');
    await createTestSubItem(f.pmToken, teamBizKey, mainItemBizKey, 'Sub 2');
  });

  // ── Outcome "unauthorized" ──────────────────────────────────────────
  // Traceability: TC-API-DEL-001 -> Step 1 Outcome "unauthorized"
  test('TC-API-DEL-001: member without delete permission gets 403', async () => {
    const res = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}`,
      { headers: authHeader(f.memberToken) },
    );
    expect(res.status).toBe(403);

    const body = JSON.parse(res.body);
    // Verify error code is present
    expect(body.code).toBeDefined();
    expect(typeof body.code).toBe('string');
  });

  // ── Outcome "not-found" ─────────────────────────────────────────────
  // Traceability: TC-API-DEL-002 -> Step 1 Outcome "not-found"
  test('TC-API-DEL-002: delete non-existent main item returns 404', async () => {
    const fakeBizKey = '9999999999999999';
    const res = await curl(
      'DELETE',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${fakeBizKey}`,
      { headers: authHeader(f.pmToken) },
    );
    expect(res.status).toBe(404);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });
});
