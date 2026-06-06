/**
 * @feature system-ux-optimization
 * @api-functional
 *
 * API Functional Test — Journey Smoke: team-and-progress-visibility
 * Covers the happy path (success outcomes) for the API-relevant steps.
 *
 * Traceability: Journey-level smoke test — verifies Step 4 success flow end-to-end.
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
  type RbacFixtures,
} from '../../shared/helpers.js';

let f: RbacFixtures;
const runId = Date.now();

describe('Journey smoke: team-and-progress-visibility (API happy path)', () => {
  beforeAll(async () => {
    f = await setupRbacFixtures();
  });

  test('SMOKE: PM user can list teams and sees their team', async () => {
    // Step 4: List teams via API with PM user
    const res = await curl('GET', `${apiUrl}/v1/teams`, {
      headers: authHeader(f.pmToken),
    });
    expect(res.status).toBe(200);

    const data = parseApiBody(res.body);
    const items = data?.items ?? data ?? [];
    expect(Array.isArray(items)).toBe(true);
    // PM user should see at least the team they were added to
    expect(items.length).toBeGreaterThan(0);

    // Verify the returned team has expected structure
    const team = items[0];
    expect(team.bizKey ?? team.id).toBeTruthy();
    expect(typeof (team.bizKey ?? team.id)).toBe('string');
  });

  test('SMOKE: unauthenticated request to list teams is rejected', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams`);
    expect(res.status).toBe(401);
  });
});
