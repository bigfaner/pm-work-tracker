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
 * Traceability: milestone-map-lifecycle / Step 7: Rollback status from reviewed back to planning
 * Contracts: step-7-rollback-status.md
 */

describe('milestone-map-lifecycle / Step 7: Rollback reviewed -> planning', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s7-${runId}`);

    // Create map and transition to reviewed
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s7-map-${runId}`, assigneeBizKey: '1' }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed 2 milestones (diverse statuses)
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s7-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    const ms2Res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s7-ms2-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const ms2BizKey = extractBizKey(parseApiBody(ms2Res.body))!;
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });
  });

  // Traceability: milestone-map-lifecycle / Step 7 / Outcome "success"
  test('TC-MML-S7-001: Rollback from reviewed to planning returns 200', async () => {
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'planning' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('planning');
  });

  // Traceability: milestone-map-lifecycle / Step 7 / Outcome "unauthorized-api"
  test('TC-MML-S7-002: Rollback without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'planning' }),
    });

    expect(res.status).toBe(401);
  });
});
