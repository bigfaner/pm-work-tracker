import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from './helpers.js';

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // ── Authenticated Tests (use shared auth) ───────────────────────

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Convert item-pool entry to main item returns 200 on MySQL', async () => {
    // Pre-conditions: team and item-pool entry must exist in MySQL.
    // This test assumes teamId=1 and a valid item-pool entry exist.
    // Adjust teamId and poolId to match your MySQL test data.
    const teamId = parseInt(process.env.E2E_TEAM_ID ?? '1', 10);
    const poolId = parseInt(process.env.E2E_POOL_ID ?? '1', 10);

    const res = await authCurl('POST', `/api/v1/teams/${teamId}/item-pool/${poolId}/convert-to-main`);
    assert.equal(res.status, 200, `Expected 200, got ${res.status}: ${res.body}`);

    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    assert.ok(item.code, 'Response contains main item code');
    // Code format: {teamCode}-{seq:05d} e.g. TEAM-00042
    assert.match(item.code, /^[A-Z]+-\d{5}$/, `Code "${item.code}" matches format {{teamCode}}-{{seq:05d}}`);
  });

  // Traceability: TC-002 → Story 1 / AC-1 (main item numbering rule)
  test('TC-002: Main item code increments sequentially on MySQL', async () => {
    const teamId = parseInt(process.env.E2E_TEAM_ID ?? '1', 10);
    const codes: string[] = [];

    // Create 3 item-pool entries and convert them to main items
    // Assumes 3 item-pool entries exist with IDs 1,2,3 (or set via env)
    const poolIds = [
      parseInt(process.env.E2E_POOL_ID_1 ?? '1', 10),
      parseInt(process.env.E2E_POOL_ID_2 ?? '2', 10),
      parseInt(process.env.E2E_POOL_ID_3 ?? '3', 10),
    ];

    for (const poolId of poolIds) {
      const res = await authCurl('POST', `/api/v1/teams/${teamId}/item-pool/${poolId}/convert-to-main`);
      assert.equal(res.status, 200, `Convert pool ${poolId} failed: ${res.status} ${res.body}`);
      const data = JSON.parse(res.body);
      const item = data.data ?? data;
      codes.push(item.code);
    }

    // Extract numeric sequences and verify strict increment
    const seqs = codes.map((c) => parseInt(c.split('-').pop() ?? '', 10));
    for (let i = 1; i < seqs.length; i++) {
      assert.equal(seqs[i], seqs[i - 1] + 1, `Code ${codes[i]} should be sequential after ${codes[i - 1]}`);
    }
  });

  // Traceability: TC-003 → Story 1 / AC-1 (sub item numbering rule)
  test('TC-003: Sub item code increments sequentially on MySQL', async () => {
    const teamId = parseInt(process.env.E2E_TEAM_ID ?? '1', 10);
    const mainItemId = parseInt(process.env.E2E_MAIN_ITEM_ID ?? '1', 10);
    const codes: string[] = [];

    // Create 3 sub items under the main item
    for (let i = 0; i < 3; i++) {
      const res = await authCurl('POST', `/api/v1/teams/${teamId}/main-items/${mainItemId}/sub-items`, {
        body: JSON.stringify({ title: `Sub item test ${i + 1}`, priority: 'P1' }),
      });
      assert.equal(res.status, 200, `Create sub item ${i + 1} failed: ${res.status} ${res.body}`);
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
    const teamId = parseInt(process.env.E2E_TEAM_ID ?? '1', 10);
    const poolId = parseInt(process.env.E2E_POOL_ID ?? '1', 10);

    const res = await authCurl('POST', `/api/v1/teams/${teamId}/item-pool/${poolId}/convert-to-main`);
    assert.equal(res.status, 200, `Expected 200, got ${res.status}: ${res.body}`);

    const data = JSON.parse(res.body);
    const item = data.data ?? data;
    assert.ok(item.code, 'Response contains main item code');
    assert.match(item.code, /^[A-Z]+-\d{5}$/, `Code "${item.code}" matches expected format`);
  });
});
