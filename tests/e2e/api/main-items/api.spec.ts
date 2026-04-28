import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { apiBaseUrl, getApiToken, createAuthCurl } from './helpers.js';

/** Required request body for convert-to-main endpoint */
const convertBody = {
  priority: 'P1',
  assigneeKey: '0',
  startDate: '2026-04-27',
  expectedEndDate: '2026-05-27',
};

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let teamId: string;
  let mainItemId: string;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);

    // Find or create a team for testing
    teamId = process.env.E2E_TEAM_ID ?? '';
    if (!teamId) {
      // List teams and pick the first one
      const listRes = await authCurl('GET', '/v1/teams');
      assert.equal(listRes.status, 200, `List teams failed: ${listRes.status} ${listRes.body}`);
      const listData = JSON.parse(listRes.body);
      const teams = listData.data?.items ?? listData;
      if (teams.length > 0) {
        teamId = String(teams[0].bizKey);
      } else {
        // No teams exist — create one for e2e tests
        const createRes = await authCurl('POST', '/v1/teams', {
          body: JSON.stringify({ name: 'E2E Test Team', code: 'ETEST', description: 'Auto-created for e2e tests' }),
        });
        assert.equal(createRes.status, 201, `Create team failed: ${createRes.status} ${createRes.body}`);
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
    assert.ok(res.status === 200 || res.status === 201, `Submit pool failed: ${res.status} ${res.body}`);
    const data = JSON.parse(res.body);
    const entry = data.data ?? data;
    return String(entry.bizKey);
  }

  /** Convert a pool entry to main item, return the main item bizKey as a string */
  async function convertToMain(poolBizKey: string): Promise<string> {
    const res = await authCurl('POST', `/v1/teams/${teamId}/item-pool/${poolBizKey}/convert-to-main`, {
      body: JSON.stringify(convertBody),
    });
    assert.equal(res.status, 200, `Convert failed: ${res.status} ${res.body}`);
    // mainItemBizKey is a JSON number that exceeds JS safe integer range.
    // Extract it as a string from the raw body to avoid precision loss.
    const match = res.body.match(/"mainItemBizKey"\s*:\s*(\d+)/);
    assert.ok(match, `mainItemBizKey not found in response: ${res.body}`);
    return match[1];
  }

  /** Fetch a main item and return its code */
  async function getMainItemCode(mainItemBizKey: string): Promise<string> {
    const res = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}`);
    assert.equal(res.status, 200, `Get main item failed: ${res.status} ${res.body}`);
    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    assert.ok(item.code, 'Main item has code');
    return String(item.code);
  }

  // ── Authenticated Tests (use shared auth) ───────────────────────

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Convert item-pool entry to main item returns 200 on MySQL', async () => {
    const poolId = await submitPool('E2E TC-001 convert test');
    const mainItemBizKey = await convertToMain(poolId);
    const code = await getMainItemCode(mainItemBizKey);

    // Code format: {teamCode}-{seq:05d} e.g. TEAM-00042
    assert.match(code, /^[A-Z]+-\d{5}$/, `Code "${code}" matches format {{teamCode}}-{{seq:05d}}`);

    // Save main item ID for TC-003
    mainItemId = mainItemBizKey;
  });

  // Traceability: TC-002 → Story 1 / AC-1 (main item numbering rule)
  test('TC-002: Main item code increments sequentially on MySQL', async () => {
    const codes: string[] = [];

    // Submit 3 pool entries and convert them to main items
    for (let i = 0; i < 3; i++) {
      const poolId = await submitPool(`E2E TC-002 seq test ${i + 1}`);
      const mainItemBizKey = await convertToMain(poolId);
      const code = await getMainItemCode(mainItemBizKey);
      codes.push(code);
    }

    // Extract numeric sequences and verify strict increment
    const seqs = codes.map((c) => parseInt(c.split('-').pop() ?? '', 10));
    for (let i = 1; i < seqs.length; i++) {
      assert.equal(seqs[i], seqs[i - 1] + 1, `Code ${codes[i]} should be sequential after ${codes[i - 1]}`);
    }
  });

  // Traceability: TC-003 → Story 1 / AC-1 (sub item numbering rule)
  test('TC-003: Sub item code increments sequentially on MySQL', async () => {
    // If TC-001 didn't set mainItemId, create a main item first
    if (!mainItemId) {
      const poolId = await submitPool('E2E TC-003 parent');
      mainItemId = await convertToMain(poolId);
    }

    const codes: string[] = [];

    // Create 3 sub items under the main item
    for (let i = 0; i < 3; i++) {
      const res = await authCurl('POST', `/v1/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
        body: JSON.stringify({
          mainItemKey: mainItemId,
          title: `Sub item test ${i + 1}`,
          priority: 'P1',
          assigneeKey: '0',
          startDate: '2026-04-27',
          expectedEndDate: '2026-05-27',
        }),
      });
      assert.ok(res.status === 200 || res.status === 201, `Create sub item ${i + 1} failed: ${res.status} ${res.body}`);
      const data = JSON.parse(res.body);
      const item = data.data ?? data;
      codes.push(item.code);
    }

    // Sub item code format: {mainCode}-{seq:02d} e.g. TEAM-00042-01
    for (const code of codes) {
      assert.match(code, /^[A-Z]+-\d{5}-\d{2}$/, `Sub code "${code}" matches format {{mainCode}}-{{seq:02d}}`);
    }

    // Extract numeric sequences and verify strict increment
    const seqs = codes.map((c) => parseInt(c.split('-').pop() ?? '', 10));
    for (let i = 1; i < seqs.length; i++) {
      assert.equal(seqs[i], seqs[i - 1] + 1, `Sub code ${codes[i]} should be sequential after ${codes[i - 1]}`);
    }
  });

  // Traceability: TC-004 → Story 5 / AC-1
  test('TC-004: Convert item-pool entry returns 200 on SQLite', async () => {
    // Same API call as TC-001 but runs against SQLite (default dev mode).
    // Verifies no regression from dialect changes.
    const poolId = await submitPool('E2E TC-004 sqlite test');
    const mainItemBizKey = await convertToMain(poolId);
    const code = await getMainItemCode(mainItemBizKey);

    assert.match(code, /^[A-Z]+-\d{5}$/, `Code "${code}" matches expected format`);
  });

  // Traceability: TC-005 → sub-item edit refresh
  test('TC-005: Update sub-item priority is reflected in list', async () => {
    // Setup: create a main item and a sub-item under it
    const poolId = await submitPool('E2E TC-005 sub-item edit test');
    const mainItemBizKey = await convertToMain(poolId);

    const createRes = await authCurl('POST', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`, {
      body: JSON.stringify({
        mainItemKey: mainItemBizKey,
        title: 'TC-005 sub item',
        priority: 'P1',
        assigneeKey: '0',
        startDate: '2026-04-28',
        expectedEndDate: '2026-05-28',
      }),
    });
    assert.ok(createRes.status === 200 || createRes.status === 201, `Create sub item failed: ${createRes.status} ${createRes.body}`);
    const subItem = JSON.parse(createRes.body).data ?? JSON.parse(createRes.body);
    const subBizKey = String(subItem.bizKey);

    // Verify initial priority in list
    const listBefore = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    assert.equal(listBefore.status, 200, `List sub items failed: ${listBefore.status} ${listBefore.body}`);
    const itemsBefore: any[] = (JSON.parse(listBefore.body).data ?? JSON.parse(listBefore.body)).items;
    const before = itemsBefore.find((i: any) => String(i.bizKey) === subBizKey);
    assert.ok(before, 'Sub item found in list before update');
    assert.equal(before.priority, 'P1', `Priority before update should be P1, got ${before.priority}`);

    // Update priority to P2
    const updateRes = await authCurl('PUT', `/v1/teams/${teamId}/sub-items/${subBizKey}`, {
      body: JSON.stringify({ priority: 'P2' }),
    });
    assert.equal(updateRes.status, 200, `Update sub item failed: ${updateRes.status} ${updateRes.body}`);

    // Verify updated priority is reflected in list
    const listAfter = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    assert.equal(listAfter.status, 200, `List sub items after update failed: ${listAfter.status} ${listAfter.body}`);
    const itemsAfter: any[] = (JSON.parse(listAfter.body).data ?? JSON.parse(listAfter.body)).items;
    const after = itemsAfter.find((i: any) => String(i.bizKey) === subBizKey);
    assert.ok(after, 'Sub item found in list after update');
    assert.equal(after.priority, 'P2', `Priority after update should be P2, got ${after.priority}`);
  });
});
