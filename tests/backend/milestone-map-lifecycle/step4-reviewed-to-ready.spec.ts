import { test, expect, describe, beforeAll } from 'vitest';
import {
  curl, apiBaseUrl, apiUrl, getApiToken, parseApiBody,
  createAuthCurl, extractBizKey, createTestTeam,
} from '../../shared/helpers.js';

/**
 * @api-functional
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Traceability: milestone-map-lifecycle / Step 4: Transition from reviewed to ready
 * Contracts: step-4-transition-reviewed-to-ready.md
 */

describe('milestone-map-lifecycle / Step 4: reviewed -> ready', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s4-${runId}`);

    // Create map and transition to reviewed
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s4-map-${runId}`, assigneeBizKey: '1' }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed 2 milestones before transitioning map status
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s4-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s4-ms2-${runId}`, expectedEndDate: '2026-12-31' }),
    });

    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });
  });

  // Traceability: milestone-map-lifecycle / Step 4 / Outcome "success"
  test('TC-MML-S4-001: Transition from reviewed to ready returns 200', async () => {
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'ready' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('ready');
  });

  // Traceability: milestone-map-lifecycle / Step 4 / Outcome "unauthorized-api"
  test('TC-MML-S4-002: Transition without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'ready' }),
    });

    expect(res.status).toBe(401);
  });
});
