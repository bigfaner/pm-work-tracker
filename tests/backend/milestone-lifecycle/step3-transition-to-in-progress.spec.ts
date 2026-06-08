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
 * Traceability: milestone-lifecycle / Step 3: Transition status from not_started to in_progress
 * Contracts: step-3-transition-to-in-progress.md
 */

describe('milestone-lifecycle / Step 3: Transition to in_progress', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let milestoneBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s3-${runId}`);

    // Create milestone map + milestone in not_started status
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s3-map-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s3-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    milestoneBizKey = extractBizKey(parseApiBody(msRes.body))!;
  });

  // Traceability: milestone-lifecycle / Step 3 / Outcome "success"
  test('TC-ML-S3-001: Transition from not_started to in_progress returns 200', async () => {
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneStatus).toBe('in_progress');
    expect(data.bizKey).toBe(milestoneBizKey);
  });

  // Traceability: milestone-lifecycle / Step 3 / Outcome "unauthorized-api"
  test('TC-ML-S3-002: Transition status without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestones/${milestoneBizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

    expect(res.status).toBe(401);
  });
});
