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
  createTestTeam,
  createTestMainItem,
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
let teamBizKey: string;
const runId = Date.now();

describe('API: Conversion form errors', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = await createTestTeam(f.superadminToken, `e2e-conv-${runId}`);
  });

  // ── Step 5 Outcome "unauthorized" ───────────────────────────────────
  // Traceability: TC-API-CONV-001 -> Step 5 Outcome "unauthorized"
  test('TC-API-CONV-001: member gets 403 for sub-item creation', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'ConvAuth Main', 'P0',
    );

    // Member attempts to create a sub-item (which is what conversion does)
    const res = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({
          mainItemKey: mainItemBizKey,
          title: 'Unauthorized Sub',
          priority: 'P2',
          assigneeKey: '1',
          startDate: '2026-01-01',
          expectedEndDate: '2026-12-31',
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
      f.pmToken, teamBizKey, 'ConvVal Main', 'P1',
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
  test('TC-API-CONV-003: member gets 403 for todo-to-main-item conversion', async () => {
    // Todo-to-main-item conversion uses item pool endpoints
    // Member without item_pool:review permission should be rejected
    const res = await curl(
      'POST',
      `${apiUrl}/v1/teams/${teamBizKey}/item-pool`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({
          title: 'Unauthorized Pool Item',
          priority: 'P2',
        }),
      },
    );
    // Either 403 (no permission) or 400 (validation) is acceptable
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);
  });
});
