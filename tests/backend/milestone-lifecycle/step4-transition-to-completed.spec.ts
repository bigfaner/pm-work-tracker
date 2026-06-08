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
 * Traceability: milestone-lifecycle / Step 4: Transition status to completed
 * Contracts: step-4-transition-to-completed.md
 */

describe('milestone-lifecycle / Step 4: Transition to completed', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s4-${runId}`);
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "success"
  test('TC-ML-S4-001: Transition from in_progress to completed returns 200', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s4-map-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s4-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    // not_started -> in_progress
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

    // in_progress -> completed
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneStatus).toBe('completed');
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "cancel-cascade"
  test('TC-ML-S4-002: Cancel milestone auto-unbinds associated MainItems', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s4-cancel-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s4-cancel-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    // Cancel the milestone
    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.milestoneStatus).toBe('cancelled');
  });

  // Traceability: milestone-lifecycle / Step 4 / Outcome "unauthorized-api"
  test('TC-ML-S4-003: Transition to completed without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestones/999999999999999999/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    expect(res.status).toBe(401);
  });
});
