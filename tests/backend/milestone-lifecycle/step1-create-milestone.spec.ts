import { test, expect, describe, beforeAll, afterAll } from 'vitest';
import {
  curl, apiBaseUrl, apiUrl, getApiToken, authHeader, parseApiBody,
  createAuthCurl, randomCode, extractBizKey, createTestTeam,
} from '../../shared/helpers.js';

/**
 * @api-functional
 * @feature milestone-map
 * Journey: milestone-lifecycle
 *
 * Traceability: milestone-lifecycle / Step 1: Create milestone in timeline view
 * Contracts: step-1-create-milestone.md
 */

describe('milestone-lifecycle / Step 1: Create milestone', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let token: string;
  let teamBizKey: string;
  let mapBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s1-${runId}`);

    // Create a parent milestone map in planning status
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `ml-s1-map-${runId}`,
        assigneeBizKey: '1',
      }),
    });
    expect(mapRes.status).toBe(201);
    mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "success"
  test('TC-ML-S1-001: Create milestone with valid data returns 201', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: `ml-milestone-${runId}`,
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.bizKey).toBeTruthy();
    expect(data.milestoneStatus).toBe('not_started');
    expect(data.milestoneName).toBe(`ml-milestone-${runId}`);
    expect(data.milestoneMapKey).toBe(mapBizKey);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "validation-error-missing-name"
  test('TC-ML-S1-002: Create milestone without name returns validation error', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "validation-error-name-too-long"
  test('TC-ML-S1-003: Create milestone with name exceeding 100 chars returns validation error', async () => {
    const longName = 'A'.repeat(101);
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: longName,
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "validation-error-missing-date"
  test('TC-ML-S1-004: Create milestone without expectedEndDate returns validation error', async () => {
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: `ml-nodate-${runId}`,
      }),
    });

    expect(res.status).toBe(400);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "unauthorized-api"
  test('TC-ML-S1-005: Create milestone without auth returns 401', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: 'unauth-test',
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(401);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "terminal-parent-map"
  test('TC-ML-S1-006: Create milestone under terminal parent map returns error', async () => {
    // Create a map and transition it to completed (terminal)
    const terminalMapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({
        mapName: `ml-terminal-${runId}`,
        assigneeBizKey: '1',
      }),
    });
    const terminalMapKey = extractBizKey(parseApiBody(terminalMapRes.body))!;

    // Transition: planning -> reviewed -> ready -> executing -> completed
    for (const status of ['reviewed', 'ready', 'executing', 'completed']) {
      const transRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/milestone-maps/${terminalMapKey}/status`, {
        body: JSON.stringify({ status }),
      });
      expect(transRes.status).toBe(200);
    }

    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${terminalMapKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: `ml-terminal-ms-${runId}`,
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).not.toBe(201);
    expect(res.status).toBeGreaterThanOrEqual(400);
  });

  // Traceability: milestone-lifecycle / Step 1 / Outcome "duplicate-name"
  test('TC-ML-S1-007: Create milestone with duplicate name in same map returns 409', async () => {
    const dupName = `ml-dup-${runId}`;
    // Create first milestone
    const first = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: dupName,
        expectedEndDate: '2026-12-31',
      }),
    });
    expect(first.status).toBe(201);

    // Try creating duplicate
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({
        milestoneName: dupName,
        expectedEndDate: '2026-12-31',
      }),
    });

    expect(res.status).toBe(409);
    const body = JSON.parse(res.body);
    expect(body.code).not.toBe(0);
  });
});
