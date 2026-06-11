import { test, expect, describe, beforeAll } from 'vitest';
import {
  curl, apiBaseUrl, apiUrl, getApiToken, parseApiBody,
  createAuthCurl, extractBizKey, createTestTeam,
} from '../../shared/helpers.js';

/**
 * @api-functional
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 2: Edit milestone information
 * Contracts: step-2-edit-milestone.md
 */

async function createMilestone(
  authCurl: ReturnType<typeof createAuthCurl>,
  teamBizKey: string,
  mapBizKey: string,
  name: string,
): Promise<string> {
  const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
    body: JSON.stringify({
      milestoneName: name,
      expectedEndDate: '2026-12-31',
    }),
  });
  expect(res.status).toBe(201);
  return extractBizKey(parseApiBody(res.body))!;
}

describe('milestone-lifecycle / Step 2: Edit milestone', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  let milestoneBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s2-${runId}`);

    // Create a parent milestone map
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `ml-s2-map-${runId}`,
        assigneeBizKey: '1',
      }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    milestoneBizKey = await createMilestone(authCurl, teamBizKey, mapBizKey, `ml-s2-ms-${runId}`);
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "success"
  test('TC-ML-S2-001: Edit milestone name and date returns 200 with updated fields', async () => {
    const newName = `ml-s2-edited-${runId}`;
    const newDate = '2026-11-30';

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}`, {
      body: JSON.stringify({
        milestoneName: newName,
        expectedEndDate: newDate,
      }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneName).toBe(newName);
    expect(data.expectedEndDate).toContain('2026-11-30');
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "no-changes"
  test('TC-ML-S2-002: Edit milestone with same values returns 200 (no-op)', async () => {
    // Fetch current values first
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}`);
    const current = parseApiBody(getRes.body);

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}`, {
      body: JSON.stringify({
        milestoneName: current.milestoneName,
      }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneName).toBe(current.milestoneName);
  });

  // Traceability: milestone-lifecycle / Step 2 / Outcome "unauthorized-api"
  test('TC-ML-S2-003: Edit milestone without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}`, {
      body: JSON.stringify({ milestoneName: 'hacked' }),
    });

    expect(res.status).toBe(401);
  });
});
