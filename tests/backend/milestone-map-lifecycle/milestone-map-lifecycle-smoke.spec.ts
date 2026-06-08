import { test, expect, describe, beforeAll } from 'vitest';
import {
  apiBaseUrl, getApiToken, parseApiBody,
  createAuthCurl, extractBizKey, createTestTeam,
} from '../../shared/helpers.js';

/**
 * @api-functional
 * @feature milestone-map
 * Journey: milestone-map-lifecycle
 *
 * Smoke test: Happy path through all steps of the milestone-map-lifecycle journey.
 * Traceability: milestone-map-lifecycle journey (all steps success outcomes)
 */

describe('milestone-map-lifecycle / Smoke: Happy path + error path', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-smoke-${runId}`);
  });

  test('TC-MML-SMOKE: Full milestone-map lifecycle happy path', async () => {
    // Step 1: Create milestone map
    const createRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `mml-smoke-${runId}`,
        assigneeBizKey: '1',
        planStartDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });
    expect(createRes.status).toBe(201);
    const mapBizKey = extractBizKey(parseApiBody(createRes.body))!;
    expect(parseApiBody(createRes.body).mapStatus).toBe('planning');

    // Step 2: Edit milestone map
    const editRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({
        mapName: `mml-smoke-edited-${runId}`,
        mapDesc: 'Updated description',
      }),
    });
    expect(editRes.status).toBe(200);
    expect(parseApiBody(editRes.body).mapName).toBe(`mml-smoke-edited-${runId}`);

    // Step 3: planning -> reviewed
    const revRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });
    expect(revRes.status).toBe(200);
    expect(parseApiBody(revRes.body).mapStatus).toBe('reviewed');

    // Step 4: reviewed -> ready
    const readyRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'ready' }),
    });
    expect(readyRes.status).toBe(200);
    expect(parseApiBody(readyRes.body).mapStatus).toBe('ready');

    // Step 5: ready -> executing
    const execRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'executing' }),
    });
    expect(execRes.status).toBe(200);
    expect(parseApiBody(execRes.body).mapStatus).toBe('executing');

    // Step 6: executing -> completed
    const compRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });
    expect(compRes.status).toBe(200);
    expect(parseApiBody(compRes.body).mapStatus).toBe('completed');

    // Verify completed is terminal
    const transRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/available-transitions`);
    expect(transRes.status).toBe(200);
    expect(parseApiBody(transRes.body).transitions).toEqual([]);
  });

  test('TC-MML-SMOKE-ERR: Error path - rollback and cancel', async () => {
    // Create map and advance to reviewed, then rollback
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-smoke-err-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // planning -> reviewed
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });

    // Step 7: Rollback reviewed -> planning
    const rollbackRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'planning' }),
    });
    expect(rollbackRes.status).toBe(200);
    expect(parseApiBody(rollbackRes.body).mapStatus).toBe('planning');

    // Advance to executing, then cancel
    for (const s of ['reviewed', 'ready', 'executing']) {
      await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
        body: JSON.stringify({ status: s }),
      });
    }

    // Cancel from executing
    const cancelRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });
    expect(cancelRes.status).toBe(200);
    expect(parseApiBody(cancelRes.body).mapStatus).toBe('cancelled');

    // Verify cancelled is terminal
    const transRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/available-transitions`);
    expect(transRes.status).toBe(200);
    expect(parseApiBody(transRes.body).transitions).toEqual([]);
  });
});
