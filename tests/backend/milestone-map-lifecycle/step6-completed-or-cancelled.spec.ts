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
 * Traceability: milestone-map-lifecycle / Step 6: Transition to completed or cancelled
 * Contracts: step-6-transition-to-completed-or-cancelled.md
 */

async function createMapToExecuting(
  authCurl: ReturnType<typeof createAuthCurl>,
  teamBizKey: string,
  name: string,
): Promise<string> {
  const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
    body: JSON.stringify({ mapName: name, assigneeBizKey: '1' }),
  });
  const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;
  for (const s of ['reviewed', 'ready', 'executing']) {
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: s }),
    });
  }
  return mapBizKey;
}

describe('milestone-map-lifecycle / Step 6: completed or cancelled', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s6-${runId}`);
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "success-completed"
  test('TC-MML-S6-001: Transition from executing to completed returns 200', async () => {
    const mapBizKey = await createMapToExecuting(authCurl, teamBizKey, `mml-s6-comp-${runId}`);

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('completed');
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "success-cancelled"
  test('TC-MML-S6-002: Transition from executing to cancelled returns 200', async () => {
    const mapBizKey = await createMapToExecuting(authCurl, teamBizKey, `mml-s6-cancel-${runId}`);

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapStatus).toBe('cancelled');
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "completed-is-terminal"
  test('TC-MML-S6-003: Completed map has no available transitions', async () => {
    const mapBizKey = await createMapToExecuting(authCurl, teamBizKey, `mml-s6-terminal-${runId}`);

    // Transition to completed
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    // Check available transitions
    const transRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/available-transitions`);
    expect(transRes.status).toBe(200);
    const transData = parseApiBody(transRes.body);
    expect(transData.transitions).toEqual([]);
  });

  // Traceability: milestone-map-lifecycle / Step 6 / Outcome "unauthorized-api"
  test('TC-MML-S6-004: Transition without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/999999999999999999/status`, {
      body: JSON.stringify({ status: 'completed' }),
    });

    expect(res.status).toBe(401);
  });
});
