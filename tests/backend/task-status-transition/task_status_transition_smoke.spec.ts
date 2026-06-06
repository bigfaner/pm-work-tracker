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
  createTestMainItem,
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
let teamBizKey: string;

describe('Journey smoke: task-status-transition (API happy path)', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
    teamBizKey = f.teamBizKey;
  });

  test('SMOKE: Superadmin creates item, transitions status successfully', async () => {
    // Step 2: Create and transition to non-terminal status
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'Smoke Status Main', 'P0',
    );

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
    expect(body.data).toBeDefined();
  });

  test('SMOKE: PM gets 403 when attempting status transition (no main_item:change_status)', async () => {
    const mainItemBizKey = await createTestMainItem(
      f.superadminToken, teamBizKey, 'Smoke PM Main', 'P1',
    );

    const res = await curl(
      'PUT',
      `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`,
      {
        headers: authHeader(f.pmToken),
        body: JSON.stringify({ status: 'progressing' }),
      },
    );
    expect(res.status).toBe(403);
  });
});
