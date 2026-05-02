import { test, expect } from '@playwright/test';
import { apiBaseUrl, getApiToken, createAuthCurl } from '../helpers.js';

/** Required request body for convert-to-main endpoint */
const convertBody = {
  priority: 'P1',
  assigneeKey: '0',
  startDate: '2026-04-27',
  expectedEndDate: '2026-05-27',
};

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamId: string;

  test.beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);

    // Find or create a team for testing
    teamId = process.env.E2E_TEAM_ID ?? '';
    if (!teamId) {
      // List teams and pick the first one
      const listRes = await authCurl('GET', '/v1/teams');
      expect(listRes.status).toBe(200);
      const listData = JSON.parse(listRes.body);
      const teams = listData.data?.items ?? listData;
      if (teams.length > 0) {
        teamId = String(teams[0].bizKey);
      } else {
        const createRes = await authCurl('POST', '/v1/teams', {
          body: JSON.stringify({ name: 'E2E Test Team', code: 'ETEST', description: 'Auto-created for e2e tests' }),
        });
        expect(createRes.status).toBe(201);
        const created = JSON.parse(createRes.body);
        teamId = String((created.data ?? created).bizKey);
      }
    }
  });

  /** Submit a pool entry and return its bizKey */
  async function submitPool(title: string): Promise<string> {
    const res = await authCurl('POST', `/v1/teams/${teamId}/item-pool`, {
      body: JSON.stringify({ title, background: 'e2e test', expectedOutput: 'e2e test' }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = JSON.parse(res.body);
    const entry = data.data ?? data;
    return String(entry.bizKey);
  }

  /** Convert a pool entry to main item, return the main item bizKey as a string */
  async function convertToMain(poolBizKey: string): Promise<string> {
    const res = await authCurl('POST', `/v1/teams/${teamId}/item-pool/${poolBizKey}/convert-to-main`, {
      body: JSON.stringify(convertBody),
    });
    expect(res.status).toBe(200);
    // mainItemBizKey is a JSON number that exceeds JS safe integer range.
    // Extract it as a string from the raw body to avoid precision loss.
    const match = res.body.match(/"mainItemBizKey"\s*:\s*(\d+)/);
    expect(match).toBeTruthy();
    return match![1];
  }

  /** Fetch a main item and return its code */
  async function getMainItemCode(mainItemBizKey: string): Promise<string> {
    const res = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}`);
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    expect(item.code).toBeTruthy();
    return String(item.code);
  }

  // ── Authenticated Tests (use shared auth) ───────────────────────

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Convert item-pool entry to main item returns 200 on MySQL', async () => {
    const poolId = await submitPool('E2E TC-001 convert test');
    const mainItemBizKey = await convertToMain(poolId);
    const code = await getMainItemCode(mainItemBizKey);

    // Code format: {teamCode}-{seq:05d} e.g. TEAM-00042
    expect(code).toMatch(/^[A-Z]+-\d{5}$/);
  });
});
