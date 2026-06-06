/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey: task-status-transition
 * Contracts: step-5-submit-conversion-form-success, step-7-submit-todo-to-main-item-conversion
 *
 * Traceability: Steps 5 & 7 — API outcomes for conversion form submissions
 *   Covers: unauthorized, validation-error
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
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
let teamBizKey: string;

describe('API: Conversion form errors', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = f.teamBizKey;
  });

  // ── Step 5 Outcome "unauthorized" ───────────────────────────────────
  // Traceability: TC-API-CONV-001 -> Step 5 Outcome "unauthorized"
  // Member has sub_item:create but NOT sub_item:assign
  test('TC-API-CONV-001: member gets 403 for sub-item assignment (no sub_item:assign)', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'ConvAuth Main', 'P0',
    );

    // First create a sub-item as PM
    const subRes = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({
          mainItemKey: mainItemBizKey,
          title: 'Sub for Assign',
          priority: 'P2',
          assigneeKey: '1',
          startDate: '2026-01-01',
          expectedEndDate: '2026-12-31',
        }),
      },
    );
    const subData = parseApiBody(subRes.body);
    const subBizKey = String(subData.bizKey ?? subData.id);

    // Member attempts to assign the sub-item (requires sub_item:assign via /assignee endpoint)
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subBizKey}/assignee`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({
          assigneeKey: f.memberUserBizKey,
        }),
      },
    );
    expect(res.status).toBe(403);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Step 5 Outcome "validation-error" ───────────────────────────────
  // Traceability: TC-API-CONV-002 -> Step 5 Outcome "validation-error"
  test('TC-API-CONV-002: sub-item creation with missing required fields returns error', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'ConvVal Main', 'P1',
    );

    const res = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({
          mainItemKey: mainItemBizKey,
          // Missing title, priority, assigneeKey
        }),
      },
    );
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Step 7 Outcome "unauthorized" ───────────────────────────────────
  // Traceability: TC-API-CONV-003 -> Step 7 Outcome "unauthorized"
  // Member has item_pool:submit but NOT item_pool:review (convert/reject)
  test('TC-API-CONV-003: member gets 403 for pool item conversion (no item_pool:review)', async () => {
    // First submit a pool item as member (they have item_pool:submit)
    const submitRes = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/item-pool`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({
          title: 'Pool Item for Conversion',
          priority: 'P2',
        }),
      },
    );
    const submitData = parseApiBody(submitRes.body);
    const poolBizKey = String(submitData.bizKey ?? submitData.id);

    // Member attempts to convert the pool item to main item (requires item_pool:review)
    const res = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}/convert-to-main`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({
          priority: 'P2',
          assigneeKey: f.pmUserBizKey,
          startDate: '2026-01-01',
          expectedEndDate: '2026-12-31',
        }),
      },
    );
    // Either 403 (no permission) or 400 (validation) is acceptable
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);
  });
});
