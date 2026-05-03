import { test, expect } from '@playwright/test';
import {
  curl, apiBaseUrl, getApiToken, authHeader, parseApiBody, extractBizKey,
  randomCode, setupRbacFixtures,
} from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let pmToken: string;
let memberToken: string;
let pmUserBizKey: string;
let memberUserBizKey: string;
let teamBizKey: string;
let memberRoleKey: string;
let mainItemBizKey: string;
let subItemBizKey: string;
let progressBizKey: string;
const runId = Date.now();

const parseData = parseApiBody;

test.describe('Permission Granularity — boundaries (TC-056..TC-081)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;
    pmToken = f.pmToken;
    memberToken = f.memberToken;
    pmUserBizKey = f.pmUserBizKey;
    memberUserBizKey = f.memberUserBizKey;
    teamBizKey = f.teamBizKey;
    memberRoleKey = f.memberRoleKey;

    // Create main item + sub-item for sub-item permission tests
    const itemRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        title: 'Granularity Main', priority: 'P1',
        assigneeKey: pmUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    mainItemBizKey = extractBizKey(parseData(itemRes.body))!;

    const subRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({
        mainItemKey: mainItemBizKey, title: 'Granularity Sub', priority: 'P2',
        assigneeKey: memberUserBizKey, startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    subItemBizKey = extractBizKey(parseData(subRes.body))!;

    // Create progress record for progress:read/update tests
    const progRes = await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ completion: 30, achievement: 'init', blocker: '', lesson: '' }),
    });
    if (progRes.status === 200 || progRes.status === 201) {
      progressBizKey = extractBizKey(parseData(progRes.body))!;
    }
  });

  // ── team:update / team:delete ──────────────────────────────────────

  test('TC-056: pm 拥有 team:update 可更新团队名称', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ name: `updated-by-pm-${runId}` }),
    });
    expect(res.status).toBe(200);
  });

  test('TC-057: member 无 team:update 更新团队名称返回 403', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ name: `denied-${runId}` }),
    });
    expect(res.status).toBe(403);
  });

  test('TC-058: member 无 team:delete 解散团队返回 403', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });

  // ── team:remove ────────────────────────────────────────────────────

  test('TC-059: pm 拥有 team:remove 可移除成员', async () => {
    // Create a temp user, add to team, then remove
    const tmpUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `remove-tc059-${runId}`, displayName: 'Remove TC059' }),
    });
    const tmpData = parseData(tmpUserRes.body);
    const tmpBizKey = extractBizKey(tmpData)!;

    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `remove-tc059-${runId}`, roleKey: memberRoleKey }),
    });

    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/${tmpBizKey}`, {
      headers: authHeader(pmToken),
    });
    expect(res.status === 200 || res.status === 204).toBeTruthy();
  });

  test('TC-060: member 无 team:remove 移除成员返回 403', async () => {
    const res = await curl('DELETE', `${apiUrl}/v1/teams/${teamBizKey}/members/99999`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });

  // ── team:transfer ──────────────────────────────────────────────────

  test('TC-061: pm 拥有 team:transfer 可转让 PM', async () => {
    // Create a temp user, add to team, transfer PM to them, then transfer back
    const tmpUserRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `transfer-tc061-${runId}`, displayName: 'Transfer TC061' }),
    });
    const tmpData = parseData(tmpUserRes.body);
    const tmpBizKey = extractBizKey(tmpData)!;

    await curl('POST', `${apiUrl}/v1/teams/${teamBizKey}/members`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: `transfer-tc061-${runId}`, roleKey: memberRoleKey }),
    });

    // Transfer PM to temp user
    const transferRes = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ newPmUserKey: tmpBizKey }),
    });
    expect(transferRes.status).toBe(200);

    // Transfer back using superadmin to restore state
    await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ newPmUserKey: pmUserBizKey }),
    });
  });

  test('TC-062: member 无 team:transfer 转让 PM 返回 403', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/pm`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ newPmUserKey: memberUserBizKey }),
    });
    expect(res.status).toBe(403);
  });

  // ── main_item:change_status (not granted to pm or member) ──────────

  test('TC-063: pm 无 main_item:change_status 返回 403', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/status`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ status: 'progressing' }),
    });
    expect(res.status).toBe(403);
  });

  // ── sub_item:change_status ─────────────────────────────────────────

  test('TC-064: pm 拥有 sub_item:change_status 更改子事项状态', async () => {
    // Sub-item is in 'progressing' (auto-transitioned by progress creation); valid: progressing→pausing
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/status`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ status: 'pausing' }),
    });
    expect(res.status).toBe(200);
  });

  test('TC-065: member 拥有 sub_item:change_status 更改子事项状态', async () => {
    // Sub-item is in 'pausing' after TC-064; valid: pausing→progressing
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/status`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ status: 'progressing' }),
    });
    expect(res.status).toBe(200);
  });

  // ── view:gantt (pm has it, member does not) ────────────────────────

  test('TC-066: pm 拥有 view:gantt 可查看甘特图', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/gantt`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-067: member 无 view:gantt 返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/gantt`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });

  // ── report:export (both pm and member have it) ─────────────────────

  test('TC-068: member 拥有 report:export 可预览周报', async () => {
    // weekStart is required and must be a Monday
    const weekStart = '2026-04-27';
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/reports/weekly/preview?weekStart=${weekStart}`, {
      headers: authHeader(memberToken),
    });
    // 200 = got data, 422 = no data for that week (both prove permission passed)
    expect(res.status === 200 || res.status === 422).toBeTruthy();
    expect(res.status).not.toBe(403);
  });

  // ── user:read (pm has, member does not) ────────────────────────────

  test('TC-069: pm 拥有 user:read 可查看用户列表', async () => {
    const res = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-070: member 无 user:read 查看用户列表返回 403', async () => {
    const res = await curl('GET', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(403);
  });

  // ── user:manage_role (neither pm nor member has it) ────────────────

  test('TC-071: pm 无 user:manage_role 创建角色返回 403', async () => {
    const res = await curl('POST', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ name: `denied-role-${runId}`, permissionCodes: ['team:read'] }),
    });
    expect(res.status).toBe(403);
  });

  // ── main_item:update (pm has, member has) ──────────────────────────

  test('TC-072: pm 拥有 main_item:update 可编辑主事项', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}`, {
      headers: authHeader(pmToken),
      body: JSON.stringify({ title: `updated-by-pm-${runId}` }),
    });
    expect(res.status).toBe(200);
  });

  test('TC-073: member 拥有 main_item:update 可编辑主事项', async () => {
    const res = await curl('PUT', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}`, {
      headers: authHeader(memberToken),
      body: JSON.stringify({ title: `updated-by-member-${runId}` }),
    });
    expect(res.status).toBe(200);
  });

  // ── team:read (pm has, member has) ─────────────────────────────────

  test('TC-074: pm 拥有 team:read 可查看团队详情', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-075: member 拥有 team:read 可查看团队详情', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(200);
  });

  // ── sub_item:read (pm has, member has) ─────────────────────────────

  test('TC-076: pm 拥有 sub_item:read 可查看子事项列表', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/main-items/${mainItemBizKey}/sub-items`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-077: member 拥有 sub_item:read 可查看子事项详情', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(200);
  });

  // ── progress:read (pm has, member has) ─────────────────────────────

  test('TC-078: pm 拥有 progress:read 可查看进度记录', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(pmToken),
    });
    expect(res.status).toBe(200);
  });

  test('TC-079: member 拥有 progress:read 可查看进度记录', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/sub-items/${subItemBizKey}/progress`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(200);
  });

  // ── view:weekly (pm has, member has) ───────────────────────────────

  test('TC-080: member 拥有 view:weekly 可查看周视图', async () => {
    const weekStart = '2026-04-27';
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/weekly?weekStart=${weekStart}`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(200);
  });

  // ── view:table (pm has, member has) ────────────────────────────────

  test('TC-081: member 拥有 view:table 可查看表格视图', async () => {
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamBizKey}/views/table`, {
      headers: authHeader(memberToken),
    });
    expect(res.status).toBe(200);
  });
});
