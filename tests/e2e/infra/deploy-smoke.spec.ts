import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl, getApiToken, authHeader, parseApiBody, extractBizKey, randomCode } from '../helpers.js';

const apiUrl = apiBaseUrl;

test.describe('Single-Binary Deploy Smoke (TC-072..TC-078)', () => {

  // ── Health Endpoint ────────────────────────────────────────────────

  // Traceability: TC-072 → Spec /health endpoint
  test('TC-072: /health endpoint returns 200 with status ok', async () => {
    const res = await curl('GET', `${apiUrl}/health`);
    expect(res.status).toBe(200);
    const body = JSON.parse(res.body);
    expect(body.status).toBe('ok');
  });

  // Traceability: TC-073 → Spec /health endpoint
  test('TC-073: /health requires no authentication', async () => {
    // No Authorization header — should still return 200
    const res = await curl('GET', `${apiUrl}/health`);
    expect(res.status).toBe(200);
  });

  // ── API Startup Integrity ──────────────────────────────────────────

  // Traceability: TC-074 → Spec startup sequence
  test('TC-074: auth/login endpoint is reachable after startup', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.data?.token ?? data.token).toBeTruthy();
  });

  // Traceability: TC-075 → Spec RBAC migration
  test('TC-075: preset roles exist after migration', async () => {
    const token = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
    const res = await curl('GET', `${apiUrl}/v1/admin/roles`, {
      headers: authHeader(token),
    });
    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const roles: Array<{ roleName: string }> = data.items ?? data;
    const names = roles.map((r) => r.roleName);
    expect(names).toContain('superadmin');
    expect(names).toContain('pm');
    expect(names).toContain('member');
  });

  // Traceability: TC-076 → Spec admin seed
  test('TC-076: initial admin user exists and can authenticate', async () => {
    const token = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });
    const res = await curl('GET', `${apiUrl}/v1/me/permissions`, {
      headers: authHeader(token),
    });
    expect(res.status).toBe(200);
    const data = JSON.parse(res.body);
    expect(data.data?.isSuperAdmin ?? data.isSuperAdmin).toBe(true);
  });

  // ── Basic CRUD Smoke ───────────────────────────────────────────────

  // Traceability: TC-077 → Spec full stack CRUD
  test('TC-077: create team → create item → create sub-item → full CRUD smoke', async () => {
    const token = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });

    // Create team
    const teamRes = await curl('POST', `${apiUrl}/v1/teams`, {
      headers: authHeader(token),
      body: JSON.stringify({ name: `smoke-team-${Date.now()}`, code: randomCode() }),
    });
    expect(teamRes.status === 200 || teamRes.status === 201).toBeTruthy();
    const teamKey = extractBizKey(parseApiBody(teamRes.body))!;

    // Create main item
    const itemRes = await curl('POST', `${apiUrl}/v1/teams/${teamKey}/main-items`, {
      headers: authHeader(token),
      body: JSON.stringify({
        title: 'Smoke Test Item', priority: 'P1', assigneeKey: '1',
        startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    expect(itemRes.status === 200 || itemRes.status === 201).toBeTruthy();
    const itemKey = extractBizKey(parseApiBody(itemRes.body))!;

    // Read main item
    const readRes = await curl('GET', `${apiUrl}/v1/teams/${teamKey}/main-items/${itemKey}`, {
      headers: authHeader(token),
    });
    expect(readRes.status).toBe(200);

    // Create sub-item
    const subRes = await curl('POST', `${apiUrl}/v1/teams/${teamKey}/main-items/${itemKey}/sub-items`, {
      headers: authHeader(token),
      body: JSON.stringify({
        mainItemKey: itemKey, title: 'Smoke Sub', priority: 'P2', assigneeKey: '1',
        startDate: '2026-01-01', expectedEndDate: '2026-12-31',
      }),
    });
    expect(subRes.status === 200 || subRes.status === 201).toBeTruthy();
    const subKey = extractBizKey(parseApiBody(subRes.body))!;

    // Create progress
    const progRes = await curl('POST', `${apiUrl}/v1/teams/${teamKey}/sub-items/${subKey}/progress`, {
      headers: authHeader(token),
      body: JSON.stringify({ completion: 50, achievement: 'smoke', blocker: '', lesson: '' }),
    });
    expect(progRes.status === 200 || progRes.status === 201).toBeTruthy();

    // Submit to item pool
    const poolRes = await curl('POST', `${apiUrl}/v1/teams/${teamKey}/item-pool`, {
      headers: authHeader(token),
      body: JSON.stringify({ title: 'Smoke Pool Item' }),
    });
    expect(poolRes.status === 200 || poolRes.status === 201).toBeTruthy();
  });

  // ── View Endpoints ─────────────────────────────────────────────────

  // Traceability: TC-078 → Spec views
  test('TC-078: weekly view endpoint returns 200 for team with data', async () => {
    const token = await getApiToken(apiBaseUrl, { username: 'admin', password: 'admin123' });

    // Get first team
    const teamsRes = await curl('GET', `${apiUrl}/v1/teams`, { headers: authHeader(token) });
    const teamsData = parseApiBody(teamsRes.body);
    const teams: Array<{ bizKey: string }> = teamsData.items ?? teamsData;
    if (teams.length === 0) return;

    const teamKey = teams[0].bizKey;
    const weekStart = '2026-04-27';
    const res = await curl('GET', `${apiUrl}/v1/teams/${teamKey}/views/weekly?weekStart=${weekStart}`, {
      headers: authHeader(token),
    });
    expect(res.status).toBe(200);
  });
});
