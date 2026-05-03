import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, createAuthCurl, randomCode } from '../helpers.js';

// Helpers to create test fixtures dynamically
async function createTeam(authCurl: ReturnType<typeof createAuthCurl>): Promise<string> {
  const code = randomCode();
  const res = await authCurl('POST', '/v1/teams', {
    body: JSON.stringify({ name: `e2e-${code}`, code }),
  });
  expect(res.status === 200 || res.status === 201).toBeTruthy();
  const data = JSON.parse(res.body);
  const bizKey = data.data?.bizKey ?? data.data?.id;
  expect(bizKey).toBeTruthy();
  return String(bizKey);
}

test.describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  test.beforeAll(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Main item response contains snowflake teamKey, not small uint', async () => {
    const teamBizKey = await createTeam(authCurl);
    const today = new Date().toISOString().split('T')[0];
    const future = new Date(Date.now() + 30 * 86400000).toISOString().split('T')[0];

    // Get admin user bizKey
    const meRes = await authCurl('GET', '/v1/me/permissions');
    const adminBizKey = '1'; // admin is always user 1 in seed

    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/main-items`, {
      body: JSON.stringify({
        title: 'TC-001 test item',
        priority: 'P2',
        assigneeKey: adminBizKey,
        startDate: today,
        expectedEndDate: future,
      }),
    });
    expect(res.status === 200 || res.status === 201).toBeTruthy();
    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    const teamKey = item.teamKey ?? item.team_key;
    expect(teamKey).toBeTruthy();
    // teamKey must be the snowflake bizKey (large number), not a small uint like 1, 2, 3
    expect(
      String(teamKey),
    ).toBe(teamBizKey);
  });
});
