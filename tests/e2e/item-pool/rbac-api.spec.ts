import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, authHeader, parseApiBody, extractBizKey, setupRbacFixtures } from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let teamBizKey: string;
let poolBizKey: string;

const parseData = parseApiBody;

test.describe('RBAC — Item Pool Review (TC-038)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    teamBizKey = f.teamBizKey;

    // Create item pool entry
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: 'Test Pool Item Review' }),
    });
    expect(poolRes.status === 200 || poolRes.status === 201).toBeTruthy();
    poolBizKey = extractBizKey(parseData(poolRes.body))!;
  });

  // Traceability: TC-038 → Story 6 / AC-3
  test('TC-038: 拥有 item_pool:review 权限审核事项', async () => {
    const res = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/item-pool/${poolBizKey}/reject`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ reason: '测试拒绝' }),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });
});
