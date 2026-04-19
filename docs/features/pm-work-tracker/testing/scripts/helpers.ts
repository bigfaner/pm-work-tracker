import { execSync } from 'node:child_process';
import { writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

export const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';

export interface CurlResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
}

export async function curl(
  method: string,
  url: string,
  opts?: {
    body?: string;
    headers?: Record<string, string>;
    timeout?: number;
  },
): Promise<CurlResponse> {
  const controller = new AbortController();
  const timeout = setTimeout(
    () => controller.abort(),
    opts?.timeout ?? 10000,
  );

  try {
    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...opts?.headers,
      },
      body: opts?.body,
      signal: controller.signal,
    });

    const headers: Record<string, string> = {};
    res.headers.forEach((v, k) => { headers[k] = v; });

    return {
      status: res.status,
      headers,
      body: await res.text(),
    };
  } finally {
    clearTimeout(timeout);
  }
}

/** Login via API and return the Authorization header. */
export async function loginAs(
  account: string,
  password: string,
): Promise<{ authHeader: Record<string, string>; token: string; body: any }> {
  const res = await curl('POST', `${apiUrl}/api/v1/auth/login`, {
    body: JSON.stringify({ username: account, password }),
  });
  if (res.status !== 200) {
    throw new Error(`Login failed for ${account}: ${res.status} ${res.body}`);
  }
  const data = JSON.parse(res.body);
  const token = data.data?.token ?? data.token;
  return { authHeader: { Authorization: `Bearer ${token}` }, token, body: data.data ?? data };
}

/** Parse response body as JSON. */
export function parseBody(res: CurlResponse): any {
  return JSON.parse(res.body);
}
