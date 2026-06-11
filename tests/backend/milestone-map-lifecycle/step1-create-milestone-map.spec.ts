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
 * Traceability: milestone-map-lifecycle / Step 1: Create milestone map via list page
 * Contracts: step-1-create-milestone-map.md
 */

describe('milestone-map-lifecycle / Step 1: Create milestone map', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `mml-s1-${runId}`);
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "success"
  test('TC-MML-S1-001: Create milestone map with valid data returns 201', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `mml-map-${runId}`,
        assigneeBizKey: '1',
      }),
    });

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.bizKey).toBeTruthy();
    expect(data.mapStatus).toBe('planning');
    expect(data.mapName).toBe(`mml-map-${runId}`);
    expect(data.teamKey).toBe(teamBizKey);
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "success" with dates
  test('TC-MML-S1-002: Create milestone map with plan dates returns 201', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `mml-map-dated-${runId}`,
        assigneeBizKey: '1',
        planStartDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.bizKey).toBeTruthy();
    expect(data.planStartDate).toBeTruthy();
    expect(data.expectedEndDate).toBeTruthy();
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-missing-name"
  test('TC-MML-S1-003: Create milestone map without name returns validation error', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        assigneeBizKey: '1',
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-name-too-long"
  test('TC-MML-S1-004: Create milestone map with name exceeding 100 chars returns validation error', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: 'X'.repeat(101),
        assigneeBizKey: '1',
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "validation-error-missing-owner"
  test('TC-MML-S1-005: Create milestone map without assigneeBizKey returns validation error', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `mml-noowner-${runId}`,
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-map-lifecycle / Step 1 / Outcome "unauthorized-api"
  test('TC-MML-S1-006: Create milestone map without auth returns 401', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: 'unauth-map',
        assigneeBizKey: '1',
      }),
    });

    expect(res.status).toBe(401);
  });
});
