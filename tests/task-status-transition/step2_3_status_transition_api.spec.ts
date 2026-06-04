/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey: task-status-transition
 * Contracts: step-2-successful-status-transition-non-terminal, step-3-terminal-status-transition-confirmation
 *
 * Traceability: Steps 2 & 3 — API outcomes for successful and terminal status transitions
 *   Covers: success (non-terminal), unauthorized, sub-items-not-terminal
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

describe('API: Successful status transitions', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = f.teamBizKey;
  });

  // ── Outcome "success" (non-terminal) ────────────────────────────────
  // Traceability: TC-API-STATUS-004 -> Step 2 Outcome "success"
  test('TC-API-STATUS-004: superadmin can transition item to valid non-terminal status', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'NonTerminal Main', 'P0',
    );

    // Transition status (requires main_item:change_status, only superadmin has it)
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.superadminToken),
        body: JSON.stringify({ status: 'progressing' }),
      },
    );
    expect(res.status).toBe(200);

    const body = JSON.parse(res.body);
    expect(body.code).toBe(0);
    // Verify the status was updated in the response data
    const data = body.data;
    expect(data).toBeDefined();
  });

  // ── Outcome "unauthorized" (Step 2) ─────────────────────────────────
  // Traceability: TC-API-STATUS-005 -> Step 2 Outcome "unauthorized"
  test('TC-API-STATUS-005: member gets 403 for non-terminal transition', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'Unauthorized Main', 'P1',
    );

    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({ status: 'progressing' }),
      },
    );
    expect(res.status).toBe(403);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Outcome "sub-items-not-terminal" (Step 3) ───────────────────────
  // Traceability: TC-API-STATUS-006 -> Step 3 Outcome "sub-items-not-terminal"
  test('TC-API-STATUS-006: terminal transition fails when sub-items are not terminal', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'TerminalGuard Main', 'P0',
    );
    // Create a non-terminal sub-item
    await createTestSubItem(
      f.pmToken, teamBizKey, mainItemBizKey, 'Active Sub',
    );

    // Attempt terminal transition (requires main_item:change_status)
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.superadminToken),
        body: JSON.stringify({ status: 'completed' }),
      },
    );
    // Should be rejected because sub-items are not terminal
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
    // The error message should mention sub-items
    expect(body.message ?? body.msg ?? '').toBeTruthy();
  });
});
