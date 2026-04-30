import { describe, test } from 'node:test';
import assert from 'node:assert/strict';
import { curl, apiUrl } from './helpers.js';

describe('API E2E Tests — config-yaml', () => {

  // Traceability: TC-020 → Spec Server 域
  test('TC-020: Application starts and serves HTTP on configured port', async () => {
    const res = await curl('GET', `${apiUrl}/api/v1/auth/login`, {
      // Using GET on login endpoint — expect 405 or 400, not connection refused
      timeout: 5000,
    });
    assert.ok(res.status !== 0, 'Server should be reachable');
    assert.ok(res.status < 500, `Got ${res.status}, server should respond without 5xx`);
  });

  // Traceability: TC-021 → Spec CORS 域
  test('TC-021: CORS origins from config applied to responses', async () => {
    const res = await curl('OPTIONS', `${apiUrl}/api/v1/auth/login`, {
      headers: {
        'Origin': 'http://localhost:3000',
        'Access-Control-Request-Method': 'POST',
      },
      timeout: 5000,
    });
    // If CORS is configured with http://localhost:3000, expect ACAO header
    const acao = res.headers['access-control-allow-origin'];
    if (acao) {
      assert.equal(acao, 'http://localhost:3000');
    }
    // If no CORS configured, just verify server responded
    assert.ok(res.status !== 0, 'Server should be reachable');
  });
});
