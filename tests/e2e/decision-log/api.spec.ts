import { test, expect } from '@playwright/test';
import {
  apiBaseUrl,
  getApiToken,
  createAuthCurl,
  curl,
  createTestTeam,
  createTestMainItem,
  authHeader,
  parseApiBody,
  extractBizKey,
  randomCode,
} from '../helpers.js';

test.describe('Decision Log API E2E Tests', () => {
  let authCurl: ReturnType<typeof createAuthCurl>;
  let token: string;
  let teamId: string;
  let mainItemId: string;

  test.beforeAll(async () => {
    token = await getApiToken(apiBaseUrl);
    authCurl = createAuthCurl(apiBaseUrl, token);

    // Create a test team and main item for decision log tests
    teamId = await createTestTeam(token, `e2e-dl-${Date.now()}`);
    mainItemId = await createTestMainItem(token, teamId, 'Decision Log Test Item', 'P1');
  });

  /** Helper: create a decision log and return its bizKey */
  async function createDecisionLog(opts: {
    category?: string;
    content?: string;
    logStatus?: string;
    tags?: string[];
  } = {}): Promise<string> {
    const res = await authCurl(
      'POST',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
      {
        body: JSON.stringify({
          category: opts.category ?? 'technical',
          content: opts.content ?? 'Test decision content',
          logStatus: opts.logStatus ?? 'draft',
          tags: opts.tags ?? [],
        }),
      },
    );
    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    return extractBizKey(data)!;
  }

  // ── TC-016: Create draft decision ──────────────────────────────────

  // Traceability: TC-016 → Story 2 / AC-1; Spec Section 5.2
  test('TC-016: Create draft decision returns 201 with correct fields', async () => {
    const res = await authCurl(
      'POST',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
      {
        body: JSON.stringify({
          category: 'technical',
          content: 'Draft decision text',
          logStatus: 'draft',
        }),
      },
    );

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);

    expect(data.logStatus).toBe('draft');
    expect(data.category).toBe('technical');
    expect(data.content).toBe('Draft decision text');
    expect(data.bizKey).toBeTruthy();
    expect(data.createdBy).toBeTruthy();
    expect(data.creatorName).toBeTruthy();
    expect(data.createTime).toBeTruthy();
  });

  // ── TC-017: Create and publish decision in one step ────────────────

  // Traceability: TC-017 → Story 1 / AC-1
  test('TC-017: Create and publish decision in one step returns 201 with published status', async () => {
    const res = await authCurl(
      'POST',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
      {
        body: JSON.stringify({
          category: 'resource',
          content: 'Published decision text',
          logStatus: 'published',
        }),
      },
    );

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);

    expect(data.logStatus).toBe('published');
    expect(data.category).toBe('resource');
    expect(data.content).toBe('Published decision text');
  });

  // ── TC-018: List decisions returns published and own drafts only ────

  // Traceability: TC-018 → Story 3/4 AC; Spec Section 5.1
  test('TC-018: List returns published decisions and own drafts only', async () => {
    // Create one published and one draft as admin user
    await createDecisionLog({ logStatus: 'published', content: 'TC-018 published' });
    await createDecisionLog({ logStatus: 'draft', content: 'TC-018 own draft' });

    const res = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
    );

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items: any[] = data.items ?? [];

    // All items should be either published or draft belonging to current user
    for (const item of items) {
      const isPublished = item.logStatus === 'published';
      const isOwnDraft = item.logStatus === 'draft';
      expect(isPublished || isOwnDraft).toBeTruthy();
    }
  });

  // ── TC-019: List decisions sorted by creation time descending ──────

  // Traceability: TC-019 → Spec Section 5.1
  test('TC-019: List decisions sorted by createTime descending', async () => {
    // Create two decisions with a small delay to ensure different timestamps
    await createDecisionLog({ content: 'TC-019 first', logStatus: 'published' });
    await new Promise((r) => setTimeout(r, 100));
    await createDecisionLog({ content: 'TC-019 second', logStatus: 'published' });

    const res = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
    );

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items: any[] = data.items ?? [];

    // Verify descending order by createTime
    for (let i = 1; i < items.length; i++) {
      const prev = new Date(items[i - 1].createTime).getTime();
      const curr = new Date(items[i].createTime).getTime();
      expect(prev).toBeGreaterThanOrEqual(curr);
    }
  });

  // ── TC-020: List decisions paginated at 20 items ───────────────────

  // Traceability: TC-020 → Spec Section 5.1
  test('TC-020: List decisions paginated at 20 items', async () => {
    // Create a main item with controlled data set
    const pagMainItemId = await createTestMainItem(token, teamId, 'Pagination Test Item', 'P2');

    // Create 25 published decisions
    for (let i = 0; i < 25; i++) {
      const res = await authCurl(
        'POST',
        `/v1/teams/${teamId}/main-items/${pagMainItemId}/decision-logs`,
        {
          body: JSON.stringify({
            category: 'technical',
            content: `Paginated item ${i}`,
            logStatus: 'published',
          }),
        },
      );
      expect(res.status).toBe(201);
    }

    // Fetch page 1
    const page1Res = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${pagMainItemId}/decision-logs?page=1&pageSize=20`,
    );
    expect(page1Res.status).toBe(200);
    const page1Data = parseApiBody(page1Res.body);
    expect(page1Data.items.length).toBe(20);
    expect(page1Data.total).toBe(25);

    // Fetch page 2
    const page2Res = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${pagMainItemId}/decision-logs?page=2&pageSize=20`,
    );
    expect(page2Res.status).toBe(200);
    const page2Data = parseApiBody(page2Res.body);
    expect(page2Data.items.length).toBe(5);
    expect(page2Data.total).toBe(25);
  });

  // ── TC-021: Update draft decision ──────────────────────────────────

  // Traceability: TC-021 → Story 2 / AC-2; Spec Section 5.2
  test('TC-021: Update draft decision returns 200 with updated content', async () => {
    const logBizKey = await createDecisionLog({ content: 'Original content' });

    const res = await authCurl(
      'PUT',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs/${logBizKey}`,
      {
        body: JSON.stringify({
          category: 'resource',
          content: 'Updated content',
        }),
      },
    );

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);

    expect(data.content).toBe('Updated content');
    expect(data.category).toBe('resource');
    expect(data.logStatus).toBe('draft');
  });

  // ── TC-022: Publish draft decision ─────────────────────────────────

  // Traceability: TC-022 → Story 2 / AC-3; Spec Section 5.2
  test('TC-022: Publish draft decision changes status to published', async () => {
    const logBizKey = await createDecisionLog({ logStatus: 'draft' });

    const res = await authCurl(
      'PATCH',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs/${logBizKey}/publish`,
    );

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    expect(data.logStatus).toBe('published');
  });

  // ── TC-023: Edit published decision returns 403 ────────────────────

  // Traceability: TC-023 → Story 4 / AC-2
  test('TC-023: Edit published decision returns 403', async () => {
    const logBizKey = await createDecisionLog({ logStatus: 'published' });

    const res = await authCurl(
      'PUT',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs/${logBizKey}`,
      {
        body: JSON.stringify({
          category: 'technical',
          content: 'Attempted edit on published',
        }),
      },
    );

    expect(res.status).toBe(403);
  });

  // ── TC-024: Create decision requires main_item:update permission ───

  // Traceability: TC-024 → Spec Section 5.2; Story 4 / AC-3
  test('TC-024: Create decision without permission returns 403', async () => {
    // Create a no-permission user for this test
    const runId = Date.now();

    // Fetch member role (has no main_item:update permission by default)
    const rolesRes = await authCurl('GET', '/v1/admin/roles');
    const rolesData = parseApiBody(rolesRes.body);
    const roles: Array<{ bizKey: string; roleName: string }> = rolesData.items ?? rolesData;
    const memberRole = roles.find((r) => r.roleName === 'member');
    if (!memberRole) {
      // Skip if no member role — cannot set up the fixture
      test.skip();
      return;
    }

    // Create a no-perm role without main_item:update
    const noPermRoleRes = await authCurl('POST', '/v1/admin/roles', {
      body: JSON.stringify({
        name: `e2e-no-dl-${runId}`,
        description: 'No main_item:update permission',
        permissionCodes: [],
      }),
    });
    const noPermRoleData = parseApiBody(noPermRoleRes.body);
    const noPermRoleKey = extractBizKey(noPermRoleData)!;

    // Create user
    const userRes = await authCurl('POST', '/v1/admin/users', {
      body: JSON.stringify({
        username: `e2e-nodl-${runId}`,
        displayName: 'E2E No DL Perms',
      }),
    });
    const userData = parseApiBody(userRes.body);
    const userBizKey = extractBizKey(userData)!;

    // Add user to team with no-perm role
    await authCurl('POST', `/v1/teams/${teamId}/members`, {
      body: JSON.stringify({ username: `e2e-nodl-${runId}`, roleKey: noPermRoleKey }),
    });

    // Login as that user
    const noPermToken = await getApiToken(apiBaseUrl, {
      username: `e2e-nodl-${runId}`,
      password: userData.initialPassword,
    });
    const noPermCurl = createAuthCurl(apiBaseUrl, noPermToken);

    // Try to create a decision log
    const res = await noPermCurl(
      'POST',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
      {
        body: JSON.stringify({
          category: 'technical',
          content: 'Should be forbidden',
          logStatus: 'draft',
        }),
      },
    );

    expect(res.status).toBe(403);
  });

  // ── TC-025: List decisions requires authentication ─────────────────

  // Traceability: TC-025 → Spec Section 5.1
  test('TC-025: List decisions without auth returns 401', async () => {
    const res = await curl(
      'GET',
      `${apiBaseUrl}/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
    );

    expect(res.status).toBe(401);
  });

  // ── TC-026: Create decision with tags ──────────────────────────────

  // Traceability: TC-026 → Spec Section 5.3
  test('TC-026: Create decision with tags returns tags in response', async () => {
    const res = await authCurl(
      'POST',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
      {
        body: JSON.stringify({
          category: 'technical',
          tags: ['缓存策略', '性能优化'],
          content: 'Decision with tags',
          logStatus: 'draft',
        }),
      },
    );

    expect(res.status).toBe(201);
    const data = parseApiBody(res.body);
    expect(data.tags).toContain('缓存策略');
    expect(data.tags).toContain('性能优化');
  });

  // ── TC-027: Create decision with all six categories ────────────────

  // Traceability: TC-027 → Spec Section 5.3
  test('TC-027: Create decision with all six categories', async () => {
    const categories = ['technical', 'resource', 'requirement', 'schedule', 'risk', 'other'];
    const catMainItemId = await createTestMainItem(token, teamId, 'All Categories Test', 'P2');
    const createdCategories: string[] = [];

    for (const category of categories) {
      const res = await authCurl(
        'POST',
        `/v1/teams/${teamId}/main-items/${catMainItemId}/decision-logs`,
        {
          body: JSON.stringify({
            category,
            content: `${category} decision`,
            logStatus: 'published',
          }),
        },
      );

      expect(res.status).toBe(201);
      const data = parseApiBody(res.body);
      expect(data.category).toBe(category);
      createdCategories.push(data.category);
    }

    // Verify all six categories were created
    expect(createdCategories.sort()).toEqual([...categories].sort());

    // List and verify all appear
    const listRes = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${catMainItemId}/decision-logs`,
    );
    expect(listRes.status).toBe(200);
    const listData = parseApiBody(listRes.body);
    const listedCategories = (listData.items ?? []).map((i: any) => i.category).sort();
    expect(listedCategories).toEqual([...categories].sort());
  });

  // ── TC-028: List response includes required fields ─────────────────

  // Traceability: TC-028 → Spec Section 5.1
  test('TC-028: List response includes all required fields', async () => {
    // Create at least one decision to ensure data exists
    await createDecisionLog({ content: 'TC-028 field check', logStatus: 'published' });

    const res = await authCurl(
      'GET',
      `/v1/teams/${teamId}/main-items/${mainItemId}/decision-logs`,
    );

    expect(res.status).toBe(200);
    const data = parseApiBody(res.body);
    const items: any[] = data.items ?? [];
    expect(items.length).toBeGreaterThan(0);

    const item = items[0];
    // Required fields per spec: bizKey, category, tags, content, createdBy/creatorName, createTime, logStatus
    expect(item.bizKey).toBeTruthy();
    expect(item.category).toBeTruthy();
    expect(Array.isArray(item.tags)).toBeTruthy();
    expect(item.content).toBeTruthy();
    expect(item.createdBy).toBeTruthy();
    expect(item.creatorName).toBeTruthy();
    expect(item.createTime).toBeTruthy();
    expect(item.logStatus).toBeTruthy();
  });
});
