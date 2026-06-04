/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey: task-status-transition
 * Contract: step-1-trigger-status-transition-error
 *
 * Traceability: Step 1 — API outcomes for status transition errors
 *   Covers: unauthorized-attempt, validation-error
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
let mainItemBizKey: string;
const runId = Date.now();

describe('API: Status transition errors', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = await createTestTeam(f.superadminToken, `e2e-status-err-${runId}`);
    mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'StatusErr Main', 'P0',
    );
  });

  // ── Outcome "unauthorized-attempt" ──────────────────────────────────
  // Traceability: TC-API-STATUS-001 -> Step 1 Outcome "unauthorized-attempt"
  test('TC-API-STATUS-001: member without change_status permission gets 403', async () => {
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({ status: 'in_progress' }),
      },
    );
    expect(res.status).toBe(403);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Outcome "validation-error" ──────────────────────────────────────
  // Traceability: TC-API-STATUS-002 -> Step 1 Outcome "validation-error"
  test('TC-API-STATUS-002: request with missing status field returns 400', async () => {
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({}),
      },
    );
    // Should be 400 (validation error) or 422
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });

  // ── Outcome "invalid-transition" (API variant) ──────────────────────
  // Traceability: TC-API-STATUS-003 -> Step 1 Outcome "invalid-transition"
  test('TC-API-STATUS-003: transition to invalid target status returns error', async () => {
    // Attempt transition to a status that is not valid from the current state
    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({ status: 'non_existent_status' }),
      },
    );
    expect(res.status).toBeGreaterThanOrEqual(400);
    expect(res.status).toBeLessThan(500);

    const body = JSON.parse(res.body);
    expect(body.code).toBeDefined();
  });
});
