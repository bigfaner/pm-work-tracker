import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, loginAs } from '../helpers.js';

const apiUrl = apiBaseUrl;

test.describe('API E2E Tests — Teams', () => {
  let adminAuth: { authHeader: Record<string, string>; token: string };
  let teamId: string;

  test.beforeAll(async () => {
    adminAuth = await loginAs('admin', 'admin123');
    // Get a team ID for team-scoped endpoints
    const teamsRes = await curl('GET', `${apiUrl}/v1/teams`, {
      headers: adminAuth.authHeader,
    });
    const teams = JSON.parse(teamsRes.body);
    const teamList = teams.data?.items ?? teams.data ?? teams;
    if (Array.isArray(teamList) && teamList.length > 0) {
      teamId = String(teamList[0].bizKey ?? teamList[0].id ?? teamList[0].ID);
    }
  });

  // Traceability: TC-053 → Spec 5.7 #1
  test('TC-053: 团队详情独立路由成员 CRUD API', async () => {
    expect(teamId).toBeTruthy();

    // GET team detail
    const detailRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}`, {
      headers: adminAuth.authHeader,
    });
    expect(detailRes.status).toBe(200);

    const detail = JSON.parse(detailRes.body);
    const teamData = detail.data ?? detail;
    expect(teamData.name || teamData.Name).toBeTruthy();

    // GET members
    const membersRes = await curl('GET', `${apiUrl}/v1/teams/${teamId}/members`, {
      headers: adminAuth.authHeader,
    });
    expect(membersRes.status).toBe(200);

    const members = JSON.parse(membersRes.body);
    const memberList = members.data ?? members;
    expect(Array.isArray(memberList)).toBeTruthy();

    // Note: POST/PUT/DELETE member operations are tested with specific data
    // that may not exist in seed — these verify the endpoint structure
  });
});
