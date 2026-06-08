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
 * Traceability: milestone-lifecycle / Step 6: Delete milestone
 * Contracts: step-6-delete-milestone.md
 */

describe('milestone-lifecycle / Step 6: Delete milestone', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s6-${runId}`);
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "success"
  test('TC-ML-S6-001: Delete milestone in not_started status returns 200', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s6-del-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s6-del-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.message).toBe('deleted');

    // Verify milestone is gone (GET returns 404)
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    expect(getRes.status).toBe(404);
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "delete-cancelled"
  test('TC-ML-S6-002: Delete milestone in cancelled status returns 200', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s6-delc-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s6-delc-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    // Cancel the milestone first
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });

    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    expect(res.status).toBe(200);
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "api-not-found"
  test('TC-ML-S6-003: Delete non-existent milestone returns 404', async () => {
    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestones/999999999999999999`);

    expect(res.status).toBe(404);
  });

  // Traceability: milestone-lifecycle / Step 6 / Outcome "unauthorized-api"
  test('TC-ML-S6-004: Delete milestone without auth returns 401', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/milestones/999999999999999999`);

    expect(res.status).toBe(401);
  });
});
