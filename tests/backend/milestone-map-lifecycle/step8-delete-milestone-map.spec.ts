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
 * Traceability: milestone-map-lifecycle / Step 8: Delete milestone map
 * Contracts: step-8-delete-milestone-map.md
 */

describe('milestone-map-lifecycle / Step 8: Delete milestone map', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s8-${runId}`);
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "success"
  test('TC-MML-S8-001: Delete milestone map in planning status returns 200', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s8-del-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.message).toBe('deleted');

    // Verify map is gone
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    expect(getRes.status).toBe(404);
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "success" - delete in reviewed status
  test('TC-MML-S8-002: Delete milestone map in reviewed status returns 200', async () => {
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s8-delrev-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Transition to reviewed
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`, {
      body: JSON.stringify({ status: 'reviewed' }),
    });

    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    expect(res.status).toBe(200);
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "api-not-found"
  test('TC-MML-S8-003: Delete non-existent milestone map returns 404', async () => {
    const res = await authCurl('DELETE', `/v1/teams/${teamBizKey}/milestone-maps/999999999999999999`);

    expect(res.status).toBe(404);
  });

  // Traceability: milestone-map-lifecycle / Step 8 / Outcome "unauthorized-api"
  test('TC-MML-S8-004: Delete milestone map without auth returns 401', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/999999999999999999`);

    expect(res.status).toBe(401);
  });
});
