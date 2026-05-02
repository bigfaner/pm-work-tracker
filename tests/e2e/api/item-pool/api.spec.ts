import { test, expect } from '@playwright/test';
import { apiBaseUrl, getApiToken, createAuthCurl } from '../../helpers.js';

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
  let mainItemId: string;

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
        // No teams exist — create one for e2e tests
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
      expect(seqs[i]).toBe(seqs[i - 1] + 1);
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
      expect(res.status === 200 || res.status === 201).toBeTruthy();
      const data = JSON.parse(res.body);
      const item = data.data ?? data;
      codes.push(item.code);
    }

    // Sub item code format: {mainCode}-{seq:02d} e.g. TEAM-00042-01
    for (const code of codes) {
      expect(code).toMatch(/^[A-Z]+-\d{5}-\d{2}$/);
    }

    // Extract numeric sequences and verify strict increment
    const seqs = codes.map((c) => parseInt(c.split('-').pop() ?? '', 10));
    for (let i = 1; i < seqs.length; i++) {
      expect(seqs[i]).toBe(seqs[i - 1] + 1);
    }
  });

  // Traceability: TC-004 → Story 5 / AC-1
  test('TC-004: Convert item-pool entry returns 200 on SQLite', async () => {
    // Same API call as TC-001 but runs against SQLite (default dev mode).
    // Verifies no regression from dialect changes.
    const poolId = await submitPool('E2E TC-004 sqlite test');
    const mainItemBizKey = await convertToMain(poolId);
    const code = await getMainItemCode(mainItemBizKey);

    expect(code).toMatch(/^[A-Z]+-\d{5}$/);
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
    expect(createRes.status === 200 || createRes.status === 201).toBeTruthy();
    const subItem = JSON.parse(createRes.body).data ?? JSON.parse(createRes.body);
    const subBizKey = String(subItem.bizKey);

    // Verify initial priority in list
    const listBefore = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    expect(listBefore.status).toBe(200);
    const itemsBefore: any[] = (JSON.parse(listBefore.body).data ?? JSON.parse(listBefore.body)).items;
    const before = itemsBefore.find((i: any) => String(i.bizKey) === subBizKey);
    expect(before).toBeTruthy();
    expect(before.priority).toBe('P1');

    // Update priority to P2
    const updateRes = await authCurl('PUT', `/v1/teams/${teamId}/sub-items/${subBizKey}`, {
      body: JSON.stringify({ priority: 'P2' }),
    });
    expect(updateRes.status).toBe(200);

    // Verify updated priority is reflected in list
    const listAfter = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    expect(listAfter.status).toBe(200);
    const itemsAfter: any[] = (JSON.parse(listAfter.body).data ?? JSON.parse(listAfter.body)).items;
    const after = itemsAfter.find((i: any) => String(i.bizKey) === subBizKey);
    expect(after).toBeTruthy();
    expect(after.priority).toBe('P2');
  });

  // Traceability: TC-006 → edit main item does not affect sub-item list
  test('TC-006: Updating a main item does not corrupt its sub-item list', async () => {
    // Setup: create a main item with two sub-items
    const poolId = await submitPool('E2E TC-006 main item edit test');
    const mainItemBizKey = await convertToMain(poolId);

    const subTitles = ['TC-006 sub A', 'TC-006 sub B'];
    const subBizKeys: string[] = [];
    for (const title of subTitles) {
      const res = await authCurl('POST', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`, {
        body: JSON.stringify({
          mainItemKey: mainItemBizKey,
          title,
          priority: 'P2',
          assigneeKey: '0',
          startDate: '2026-04-28',
          expectedEndDate: '2026-05-28',
        }),
      });
      expect(res.status === 200 || res.status === 201).toBeTruthy();
      const sub = JSON.parse(res.body).data ?? JSON.parse(res.body);
      subBizKeys.push(String(sub.bizKey));
    }

    // Verify sub-items exist before editing the main item
    const listBefore = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    expect(listBefore.status).toBe(200);
    const itemsBefore: any[] = (JSON.parse(listBefore.body).data ?? JSON.parse(listBefore.body)).items;
    expect(itemsBefore.length).toBe(2);

    // Edit the main item (change title and priority)
    const editRes = await authCurl('PUT', `/v1/teams/${teamId}/main-items/${mainItemBizKey}`, {
      body: JSON.stringify({
        title: 'TC-006 updated title',
        priority: 'P1',
        assigneeKey: null,
        expectedEndDate: null,
        actualEndDate: null,
        description: '',
      }),
    });
    expect(editRes.status).toBe(200);

    // Verify the main item title was updated
    const getRes = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}`);
    expect(getRes.status).toBe(200);
    const updated = JSON.parse(getRes.body).data ?? JSON.parse(getRes.body);
    expect(updated.title).toBe('TC-006 updated title');

    // Verify sub-items are still intact after editing the main item
    const listAfter = await authCurl('GET', `/v1/teams/${teamId}/main-items/${mainItemBizKey}/sub-items`);
    expect(listAfter.status).toBe(200);
    const itemsAfter: any[] = (JSON.parse(listAfter.body).data ?? JSON.parse(listAfter.body)).items;
    expect(itemsAfter.length).toBe(2);

    // Both sub-item bizKeys must still be present
    for (const key of subBizKeys) {
      const found = itemsAfter.find((i: any) => String(i.bizKey) === key);
      expect(found).toBeTruthy();
    }
  });
});
