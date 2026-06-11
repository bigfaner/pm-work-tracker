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
 * Traceability: milestone-map-lifecycle / Step 3: Transition from planning to reviewed
 * Contracts: step-3-transition-planning-to-reviewed.md
 */

describe('milestone-map-lifecycle / Step 3: planning -> reviewed', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s3-${runId}`);

    // Create map in planning status
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s3-map-${runId}`, assigneeBizKey: '1' }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed milestones (not_started and in_progress for diversity)
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s3-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    const ms2Res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s3-ms2-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const ms2BizKey = extractBizKey(parseApiBody(ms2Res.body))!;
    // Transition one milestone to in_progress
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });
  });

  // Traceability: milestone-map-lifecycle / Step 3 / Outcome "success"
  test('TC-MML-S3-001: Transition from planning to reviewed returns 200', async () => {
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('reviewed');
    expect(data.bizKey).toBe(mapBizKey);
  });

  // Traceability: milestone-map-lifecycle / Step 3 / Outcome "unauthorized-api"
  test('TC-MML-S3-002: Transition without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });

    expect(res.status).toBe(401);
  });
});
