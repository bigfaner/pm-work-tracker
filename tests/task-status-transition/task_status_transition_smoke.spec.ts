/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey Smoke: task-status-transition
 * Covers the happy path for API-relevant status transition steps.
 *
 * Traceability: Journey-level smoke test — verifies successful non-terminal
 *   status transition end-to-end.
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

describe('Journey smoke: task-status-transition (API happy path)', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = await createTestTeam(f.superadminToken, `e2e-status-smoke-${runId}`);
  });

  test('SMOKE: PM creates item, transitions status successfully', async () => {
    // Step 2: Create and transition to non-terminal status
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'Smoke Status Main', 'P0',
    );

    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({ status: 'in_progress' }),
      },
    );
    expect(res.status).toBe(200);

    const body = JSON.parse(res.body);
    expect(body.code).toBe(0);
    expect(body.data).toBeDefined();
  });

  test('SMOKE: member gets 403 when attempting status transition', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.pmToken, teamBizKey, 'Smoke Member Main', 'P1',
    );

    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.memberToken),
        body: JSON.stringify({ status: 'in_progress' }),
      },
    );
    expect(res.status).toBe(403);
  });
});
