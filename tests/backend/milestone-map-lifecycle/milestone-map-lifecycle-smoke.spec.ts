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

  test('TC-MML-SMOKE: Full milestone-map lifecycle happy path with milestones', async () => {
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

    // Seed milestones inside the map (different statuses)
    const ms1Res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-smoke-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    expect(ms1Res.status).toBe(201);
    const ms1BizKey = extractBizKey(parseApiBody(ms1Res.body))!;
    expect(parseApiBody(ms1Res.body).milestoneStatus).toBe('not_started');

    const ms2Res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-smoke-ms2-${runId}`, expectedEndDate: '2026-09-30' }),
    });
    expect(ms2Res.status).toBe(201);
    const ms2BizKey = extractBizKey(parseApiBody(ms2Res.body))!;

    const ms3Res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-smoke-ms3-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    expect(ms3Res.status).toBe(201);
    const ms3BizKey = extractBizKey(parseApiBody(ms3Res.body))!;

    // Transition ms1 to in_progress (diverse statuses)
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms1BizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

    // Verify map contains all 3 milestones
    const listRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`);
    expect(listRes.status).toBe(200);
    const milestones = parseApiBody(listRes.body);
    const items = Array.isArray(milestones) ? milestones : (milestones?.items ?? []);
    expect(items.length).toBeGreaterThanOrEqual(3);

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

    // Transition ms2 to in_progress, ms2 to completed (diverse milestone statuses in executing map)
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms2BizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    // BR-2: All milestones must be terminal before map can transition to completed
    // Transition ms1 (in_progress) -> completed
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms1BizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });
    // Transition ms3 (not_started) -> cancelled
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${ms3BizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });

    // Step 6: executing -> completed
    const compRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });
    expect(compRes.status).toBe(200);
    expect(parseApiBody(compRes.body).mapStatus).toBe('completed');

    // Verify milestones still accessible after map completed
    const msAfterComp = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`);
    expect(msAfterComp.status).toBe(200);
    const msItemsAfter = parseApiBody(msAfterComp.body);
    const msListAfter = Array.isArray(msItemsAfter) ? msItemsAfter : (msItemsAfter?.items ?? []);
    expect(msListAfter.length).toBeGreaterThanOrEqual(3);

    // Verify completed is terminal
    const transRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/available-transitions`);
    expect(transRes.status).toBe(200);
    expect(parseApiBody(transRes.body).transitions).toEqual([]);
  });

  test('TC-MML-SMOKE-ERR: Error path - rollback and cancel with milestones', async () => {
    // Create map and seed milestones, then advance to reviewed, then rollback
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-smoke-err-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed 2 milestones in the map
    const msA = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-err-msA-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    expect(msA.status).toBe(201);
    const msABizKey = extractBizKey(parseApiBody(msA.body))!;

    const msB = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-err-msB-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    expect(msB.status).toBe(201);
    const msBBizKey = extractBizKey(parseApiBody(msB.body))!;

    // Transition msA to in_progress (diverse statuses before rollback)
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msABizKey}/status`, {
      body: JSON.stringify({ status: 'in_progress' }),
    });

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

    // Milestones should still be present after rollback
    const msAfterRollback = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`);
    expect(msAfterRollback.status).toBe(200);
    const msListRollback = parseApiBody(msAfterRollback.body);
    const msItemsRollback = Array.isArray(msListRollback) ? msListRollback : (msListRollback?.items ?? []);
    expect(msItemsRollback.length).toBeGreaterThanOrEqual(2);

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
