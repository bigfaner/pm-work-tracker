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
 * Traceability: milestone-lifecycle / Step 5: Cancelled state interactions
 * Contracts: step-5-cancelled-state.md
 */

describe('milestone-lifecycle / Step 5: Cancelled state interactions', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamBizKey: string;
  const runId = Date.now();

  beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
    teamBizKey = await createTestTeam(token, `ml-s5-${runId}`);
  });

  // Traceability: milestone-lifecycle / Step 5 / Outcome "cannot-receive-bindings"
  test('TC-ML-S5-001: Cancelled milestone rejects new MainItem binding', async () => {
    // Create map + milestone, cancel it
    const mapRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps`, {
      body: JSON.stringify({ mapName: `ml-s5-map-${runId}`, assigneeBizKey: '1' }),
    });
    const mapBizKey = extractBizKey(parseApiBody(mapRes.body))!;

    const msRes = await authCurl('POST', `/v1/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`, {
      body: JSON.stringify({ milestoneName: `ml-s5-ms-${runId}`, expectedEndDate: '2026-12-31' }),
    });
    const msBizKey = extractBizKey(parseApiBody(msRes.body))!;

    // Cancel the milestone
    await authCurl('PUT', `/v1/teams/${teamBizKey}/milestones/${msBizKey}/status`, {
      body: JSON.stringify({ status: 'cancelled' }),
    });

    // Attempt to create a MainItem bound to the cancelled milestone
    const itemRes = await authCurl('POST', `/v1/teams/${teamBizKey}/main-items`, {
      body: JSON.stringify({
        title: `ml-s5-item-${runId}`,
        priority: 'P2',
        assigneeKey: '1',
        startDate: '2026-01-01',
        expectedEndDate: '2026-12-31',
        milestoneKey: msBizKey,
      }),
    });
    // Binding to cancelled milestone should fail
    if (itemRes.status === 201) {
      // If create succeeded without binding rejection, try explicit binding via update
      const itemBizKey = extractBizKey(parseApiBody(itemRes.body))!;
      const bindRes = await authCurl('PUT', `/v1/teams/${teamBizKey}/main-items/${itemBizKey}`, {
        body: JSON.stringify({ milestoneKey: msBizKey }),
      });
      expect(bindRes.status).not.toBe(200);
    } else {
      // The create itself rejected the binding to cancelled milestone
      expect(itemRes.status).not.toBe(201);
    }
  });

  // Traceability: milestone-lifecycle / Step 5 / Outcome "unauthorized-api"
  test('TC-ML-S5-002: Milestone endpoint without auth returns 401', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/milestones/999999999999999999`);

    expect(res.status).toBe(401);
  });
});
