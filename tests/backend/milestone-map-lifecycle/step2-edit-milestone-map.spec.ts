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
 * Traceability: milestone-map-lifecycle / Step 2: Edit milestone map information
 * Contracts: step-2-edit-milestone-map.md
 */

describe('milestone-map-lifecycle / Step 2: Edit milestone map', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s2-${runId}`);

    // Create a milestone map in planning status
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `mml-s2-map-${runId}`, assigneeBizKey: '1' }),
    });
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    // Seed 2 milestones so the map has content
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s2-ms1-${runId}`, expectedEndDate: '2026-06-30' }),
    });
    await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `mml-s2-ms2-${runId}`, expectedEndDate: '2026-12-31' }),
    });
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "success"
  test('TC-MML-S2-001: Edit milestone map name and desc returns 200 with updated fields', async () => {
    const newName = `mml-s2-edited-${runId}`;
    const newDesc = 'Updated description';

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({
        mapName: newName,
        mapDesc: newDesc,
      }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapName).toBe(newName);
    expect(data.mapDesc).toBe(newDesc);
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "no-changes"
  test('TC-MML-S2-002: Edit milestone map with same values returns 200', async () => {
    const getRes = await authCurl('GET', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`);
    const current = parseApiBody(getRes.body);

    const res = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({
        mapName: current.mapName,
      }),
    });

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.mapName).toBe(current.mapName);
  });

  // Traceability: milestone-map-lifecycle / Step 2 / Outcome "unauthorized-api"
  test('TC-MML-S2-003: Edit milestone map without auth returns 401', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}`, {
      body: JSON.stringify({ mapName: 'hacked' }),
    });

    expect(res.status).toBe(401);
  });
});
