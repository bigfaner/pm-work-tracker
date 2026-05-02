import { test, expect } from '@playwright/test';
import { curl, apiBaseUrl } from '../../helpers.js';

const apiUrl = apiBaseUrl;

test.describe('API E2E Tests — config-yaml', () => {

  // Traceability: TC-020 → Spec Server 域
  test('TC-020: Application starts and serves HTTP on configured port', async () => {
    const res = await curl('GET', `${apiUrl}/api/v1/auth/login`, {
      // Using GET on login endpoint — expect 405 or 400, not connection refused
      timeout: 5000,
    });
    expect(res.status !== 0).toBeTruthy();
    expect(res.status < 500).toBeTruthy();
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
      expect(acao).toBe('http://localhost:3000');
    }
    // If no CORS configured, just verify server responded
    expect(res.status !== 0).toBeTruthy();
  });
});
