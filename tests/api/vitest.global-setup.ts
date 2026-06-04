import { execSync, spawn, ChildProcess } from 'node:child_process';
import { writeFileSync, mkdirSync, rmSync, existsSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createServer } from 'node:net';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROJECT_ROOT = resolve(__dirname, '..', '..');
const BACKEND_DIR = resolve(PROJECT_ROOT, 'backend');

let serverProcess: ChildProcess | null = null;
let tempDir: string | null = null;

function findAvailablePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const server = createServer();
    server.listen(0, () => {
      const addr = server.address();
      if (addr && typeof addr === 'object') {
        const port = addr.port;
        server.close(() => resolve(port));
      } else {
        server.close(() => reject(new Error('Failed to get port')));
      }
    });
    server.on('error', reject);
  });
}

function waitForServer(url: string, timeoutMs = 30000): Promise<void> {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    const check = async () => {
      if (Date.now() - start > timeoutMs) {
        reject(new Error(`Server at ${url} did not respond within ${timeoutMs}ms`));
        return;
      }
      try {
        const res = await fetch(url);
        if (res.ok) {
          resolve();
          return;
        }
      } catch {
        // Not ready yet
      }
      setTimeout(check, 500);
    };
    check();
  });
}

export default async function setup() {
  const port = await findAvailablePort();
  const apiBaseUrl = `http://localhost:${port}`;

  // Create temp directory for config and DB
  tempDir = resolve(PROJECT_ROOT, '.tmp', `api-test-${port}`);
  mkdirSync(tempDir, { recursive: true });

  // Write backend config pointing to this temp dir
  const configContent = `server:
  port: "${port}"
  gin_mode: "test"
  base_path: ""
  read_timeout: "30s"
  write_timeout: "30s"
  max_body_size: 10485760

database:
  driver: "sqlite"
  path: "${tempDir}/test.db"
  auto_schema: true
  max_open_conns: 5
  max_idle_conns: 2
  conn_max_lifetime: "1h"

auth:
  jwt_secret: "test-secret-that-is-at-least-32-bytes!!"
  jwt_expiry: "24h"
  initial_admin:
    username: "admin"
    password: "admin123"

cors:
  origins:
    - "http://localhost:5173"

logging:
  level: "warn"
  format: "json"
`;
  const configPath = resolve(tempDir, 'config.yaml');
  writeFileSync(configPath, configContent);

  // Copy migrations directory so the server can find it
  const migrationsDir = resolve(tempDir, 'migrations');
  mkdirSync(migrationsDir, { recursive: true });
  execSync(`cp ${resolve(BACKEND_DIR, 'migrations', 'SQLite-schema.sql')} ${migrationsDir}/`, {
    cwd: PROJECT_ROOT,
  });

  // Write E2E config for the test helpers to pick up
  const e2eConfigContent = `baseUrl: http://localhost:5173
apiBaseUrl: ${apiBaseUrl}
username: 'admin'
password: 'admin123'
`;
  const e2eConfigPath = resolve(tempDir, 'e2e-config.yaml');
  writeFileSync(e2eConfigPath, e2eConfigContent);

  // Set env var BEFORE tests import helpers (globalSetup runs before test files)
  process.env.E2E_CONFIG_PATH = e2eConfigPath;

  // Build and start the backend server
  const binPath = resolve(tempDir, 'test-server');
  execSync(`go build -o ${binPath} ./cmd/server/`, {
    cwd: BACKEND_DIR,
    stdio: 'pipe',
    timeout: 60000,
  });

  serverProcess = spawn(binPath, ['-config', configPath, '-dev'], {
    cwd: BACKEND_DIR,
    stdio: 'pipe',
    env: { ...process.env },
  });

  serverProcess.stdout?.on('data', (data: Buffer) => {
    // Log server output for debugging
    const msg = data.toString().trim();
    if (msg) console.log('[backend]', msg);
  });

  serverProcess.stderr?.on('data', (data: Buffer) => {
    const msg = data.toString().trim();
    if (msg) console.error('[backend]', msg);
  });

  serverProcess.on('error', (err) => {
    console.error('Failed to start backend server:', err);
  });

  // Wait for server to be ready
  await waitForServer(`${apiBaseUrl}/health`);

  return () => {
    if (serverProcess && !serverProcess.killed) {
      serverProcess.kill('SIGTERM');
      serverProcess = null;
    }
    if (tempDir && existsSync(tempDir)) {
      rmSync(tempDir, { recursive: true, force: true });
      tempDir = null;
    }
  };
}
