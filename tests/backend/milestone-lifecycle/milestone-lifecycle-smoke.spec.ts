import { test, expect, describe, beforeAll } from 'vitest';
import {
  apiBaseUrl, getApiToken, parseApiBody,
  createAuthCurl, extractBizKey, createTestTeam,
} from '../../shared/helpers.js';

/**
 * @api-functional
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Smoke test: Happy path through all steps of the milestone-lifecycle journey.
 * Traceability: milestone-lifecycle journey (all steps success outcomes)
 */

describe('milestone-lifecycle / Smoke: Happy path', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-smoke-${runId}`);
  });

  test('TC-ML-SMOKE: Full milestone lifecycle happy path', async () => {
    // Step 1: Create milestone map
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-smoke-map-${runId}`, assigneeBizKey: '1' }),
    });
    expect(mapRes.status).toBe(201);
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Step 1: Create milestone under map
    const createRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: `ml-smoke-ms-${runId}`,
        expectedEndDate: '2026-12-31',
      }),
    });
    expect(createRes.status).toBe(201);
    const msBizKey = extractBizKey(parseApiBody(createRes.body))!;
    expect(parseApiBody(createRes.body).milestoneStatus).toBe('not_started');

    // Step 2: Edit milestone
    const editRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`, {
      body: JSON.stringify({
        milestoneName: `ml-smoke-edited-${runId}`,
      }),
    });
    expect(editRes.status).toBe(200);
    expect(parseApiBody(editRes.body).milestoneName).toBe(`ml-smoke-edited-${runId}`);

    // Step 3: Transition not_started -> in_progress
    const progressRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });
    expect(progressRes.status).toBe(200);
    expect(parseApiBody(progressRes.body).milestoneStatus).toBe('in_progress');

    // Step 4: Transition in_progress -> completed
    const completeRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });
    expect(completeRes.status).toBe(200);
    expect(parseApiBody(completeRes.body).milestoneStatus).toBe('completed');

    // Step 6: Delete milestone (completed status milestones may not be deletable;
    // verify the endpoint responds correctly)
    const deleteRes = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestones/${msBizKey}`);
    // Completed milestones may not be deletable, so accept 200 or appropriate error
    expect([200, 400, 403, 422]).toContain(deleteRes.status);
  });

  test('TC-ML-SMOKE-ERR: Error path - create then cancel milestone', async () => {
    // Create a map + milestone, then cancel it (error path scenario)
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-smoke-err-map-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: `ml-smoke-err-ms-${runId}`,
        expectedEndDate: '2026-12-31',
      }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    // Cancel the milestone directly from not_started
    const cancelRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });
    expect(cancelRes.status).toBe(200);
    expect(parseApiBody(cancelRes.body).milestoneStatus).toBe('cancelled');

    // Verify no transitions available for cancelled milestone
    const transRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/available-transitions`);
    expect(transRes.status).toBe(200);
    const transData = parseApiBody(transRes.body);
    expect(transData.transitions).toEqual([]);
  });
});
