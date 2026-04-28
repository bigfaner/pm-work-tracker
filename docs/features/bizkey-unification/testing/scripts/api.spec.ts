import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiBaseUrl, getApiToken, createAuthCurl } from './helpers.js';

describe('API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;

  before(async () => {
    const token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);
  });

  // Traceability: TC-001 → Story 1 / AC-1
  test('TC-001: Progress record stores correct snowflake team_key', async () => {
    // Pre-condition: team with bizKey 123456789012345678 must exist and user must be a member
    const teamBizKey = '123456789012345678';
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/progress`, {
      body: JSON.stringify({
        weekStart: new Date().toISOString().split('T')[0],
        content: 'test progress entry',
      }),
    });
    // Accept 200 or 201; reject 500 (which would indicate internal ID was used)
    assert.ok(
      res.status === 200 || res.status === 201,
      `Expected 200/201 but got ${res.status}: ${res.body}`,
    );
    const data = JSON.parse(res.body);
    // The returned record's teamKey (or team_key) must equal the snowflake bizKey, not a small uint
    const teamKey = data.teamKey ?? data.team_key ?? data.data?.teamKey ?? data.data?.team_key;
    if (teamKey !== undefined) {
      assert.equal(String(teamKey), teamBizKey, 'team_key must be the snowflake bizKey, not internal uint ID');
    }
  });

  // Traceability: TC-002 → Story 2 / AC-1
  test('TC-002: Invite member with PM role bizKey returns ErrCannotAssignPMRole', async () => {
    // Pre-condition: PM role exists with bizKey 987654321098765432; caller is team PM
    const teamBizKey = '123456789012345678';
    const pmRoleBizKey = 987654321098765432;
    const res = await authCurl('POST', `/v1/teams/${teamBizKey}/members`, {
      body: JSON.stringify({
        username: 'testuser',
        roleBizKey: pmRoleBizKey,
      }),
    });
    assert.ok(
      res.status === 400 || res.status === 422,
      `Expected 400/422 for PM role assignment but got ${res.status}: ${res.body}`,
    );
    const body = res.body.toLowerCase();
    assert.ok(
      body.includes('pm') || body.includes('cannot') || body.includes('role'),
      `Response should indicate PM role assignment is forbidden: ${res.body}`,
    );
  });

  // Traceability: TC-003 → Spec Section 5.5 — bizKey 校验规则 (rule 1)
  test('TC-003: Non-numeric teamId in URL returns 400', async () => {
    const res = await authCurl('GET', '/v1/teams/abc/main-items');
    assert.equal(res.status, 400, `Expected 400 for non-numeric teamId but got ${res.status}: ${res.body}`);
  });

  // Traceability: TC-004 → Spec Section 5.5 — bizKey 校验规则 (rule 2)
  test('TC-004: Non-positive teamId in URL returns 400', async () => {
    const resZero = await authCurl('GET', '/v1/teams/0/main-items');
    assert.equal(resZero.status, 400, `Expected 400 for teamId=0 but got ${resZero.status}: ${resZero.body}`);

    const resNeg = await authCurl('GET', '/v1/teams/-1/main-items');
    assert.equal(resNeg.status, 400, `Expected 400 for teamId=-1 but got ${resNeg.status}: ${resNeg.body}`);
  });

  // Traceability: TC-005 → Spec Section 5.5 — bizKey 校验规则 (rule 3)
  test('TC-005: Non-existent teamId returns 404', async () => {
    const res = await authCurl('GET', '/v1/teams/999999999999999999/main-items');
    assert.equal(res.status, 404, `Expected 404 for non-existent teamId but got ${res.status}: ${res.body}`);
  });
});
