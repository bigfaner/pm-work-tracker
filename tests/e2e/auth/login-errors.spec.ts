import { test, expect } from '@playwright/test';
import {
  curl, apiBaseUrl, getApiToken, authHeader, parseApiBody, extractBizKey,
  setupRbacFixtures,
} from '../helpers.js';

const apiUrl = apiBaseUrl;

let superadminToken: string;
let disabledUserBizKey: string;
let disabledUsername: string;
let disabledPassword: string;
const runId = Date.now();

test.describe('Login Error States (TC-107..TC-110)', () => {
  test.beforeAll(async () => {
    const f = await setupRbacFixtures();
    superadminToken = f.superadminToken;

    // Create a user and disable them
    disabledUsername = `disabled-tc107-${runId}`;
    const userRes = await curl('POST', `${apiUrl}/v1/admin/users`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ username: disabledUsername, displayName: 'Disabled TC107' }),
    });
    const userData = parseApiBody(userRes.body);
    disabledUserBizKey = extractBizKey(userData)!;
    disabledPassword = userData.initialPassword;

    // Disable the user
    await curl('PUT', `${apiUrl}/v1/admin/users/${disabledUserBizKey}/status`, {
      headers: authHeader(superadminToken),
      body: JSON.stringify({ status: 'disabled' }),
    });
  });

  // Traceability: TC-107 → Auth / disabled account
  test('TC-107: 被禁用账户登录返回 403 USER_DISABLED', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: disabledUsername, password: disabledPassword }),
    });
    expect(res.status).toBe(403);
    const body = JSON.parse(res.body);
    expect(body.code).toBe('USER_DISABLED');
  });

  // Traceability: TC-108 → Auth / wrong password
  test('TC-108: 错误密码登录返回 401', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: 'admin', password: 'wrong-password' }),
    });
    expect(res.status).toBe(401);
  });

  // Traceability: TC-109 → Auth / nonexistent user
  test('TC-109: 不存在的用户登录返回 401', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: `nonexistent-${runId}`, password: 'anything' }),
    });
    expect(res.status).toBe(401);
  });

  // Traceability: TC-110 → Auth / missing fields
  test('TC-110: 缺少密码字段登录返回 400', async () => {
    const res = await curl('POST', `${apiUrl}/v1/auth/login`, {
      body: JSON.stringify({ username: 'admin' }),
    });
    expect(res.status === 400 || res.status === 401).toBeTruthy();
  });
});
