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
 * Traceability: milestone-map-lifecycle / Step 5: Transition from ready to executing
 * Contracts: step-5-transition-ready-to-executing.md
 */

describe('milestone-map-lifecycle / Step 5: ready -> executing', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s5-${runId}`);

    // Create map and transition to ready
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s5-map-${runId}`, assigneeBizKey: '1' }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed 2 milestones before advancing map status
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s5-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s5-ms2-${runId}`, expectedEndDate: '2026-12-31' }),
    });

    // planning -> reviewed -> ready
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'ready' }),
    });
  });

  // Traceability: milestone-map-lifecycle / Step 5 / Outcome "success"
  test('TC-MML-S5-001: Transition from ready to executing returns 200', async () => {
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'executing' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('executing');
  });

  // Traceability: milestone-map-lifecycle / Step 5 / Outcome "unauthorized-api"
  test('TC-MML-S5-002: Transition without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'executing' }),
    });

    expect(res.status).toBe(401);
  });
});
