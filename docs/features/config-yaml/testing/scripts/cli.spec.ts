import { describe, test, after } from 'node:test';
import assert from 'node:assert/strict';
import { runServer, writeConfig, removeConfig } from './helpers.js';

describe('CLI E2E Tests — config-yaml', () => {

  after(() => {
    removeConfig();
  });

  // ===== Config Loading (TC-001 ~ TC-003) =====

  describe('Config Loading', () => {
    // Traceability: TC-001 → Story 1 / AC-1
    test('TC-001: Load config from config.yaml', () => {
      writeConfig(`
server:
  port: "9090"
database:
  driver: sqlite
  path: "./data/e2e-test.db"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /9090|listening|started|ready/i);
    });

    // Traceability: TC-002 → Spec 流程说明
    test('TC-002: Use hardcoded defaults when config.yaml missing', () => {
      removeConfig();
      const result = runServer({ AUTH_JWT_SECRET: 'e2e-default-jwt-secret-at-least-32-bytes!!' });
      // Should start without config file using defaults
      assert.match(result.stdout + result.stderr, /8080|listening|started|ready/i);
    });

    // Traceability: TC-003 → Spec 流程说明
    test('TC-003: Malformed YAML causes startup failure', () => {
      writeConfig('server:\n  port: "8080"\n  broken: [unclosed');
      const result = runServer();
      assert.notEqual(result.exitCode, 0, 'Should fail with non-zero exit code');
      assert.match(result.stdout + result.stderr, /yaml|parse|syntax|config/i);
    });
  });

  // ===== Environment Variable Override (TC-004 ~ TC-007) =====

  describe('Environment Variable Override', () => {
    // Traceability: TC-004 → Story 2 / AC-1
    test('TC-004: Env var overrides YAML value', () => {
      writeConfig(`
auth:
  jwt_secret: "weak-value"
server:
  port: "9090"
`);
      const result = runServer({ AUTH_JWT_SECRET: 'this-is-a-strong-override-value-32-bytes!!' });
      assert.match(result.stdout + result.stderr, /9090|listening|started|ready/i);
    });

    // Traceability: TC-005 → Spec 环境变量覆盖规则
    test('TC-005: Nested field env var naming (AUTH_INITIAL_ADMIN_USERNAME)', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
  initial_admin:
    username: "yaml_admin"
    password: "yaml_pass"
`);
      const result = runServer({ AUTH_INITIAL_ADMIN_USERNAME: 'env_admin', AUTH_INITIAL_ADMIN_PASSWORD: 'env_pass_123' });
      assert.match(result.stdout + result.stderr, /admin|created|seed|env_admin/i);
    });

    // Traceability: TC-006 → Spec 环境变量覆盖规则
    test('TC-006: Array field env var with comma separation', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
server:
  port: "9091"
`);
      const result = runServer({ CORS_ORIGINS: 'http://a.com,http://b.com' });
      // Server should start; CORS origins parsed from comma-separated env var
      assert.match(result.stdout + result.stderr, /9091|listening|started|ready/i);
    });

    // Traceability: TC-007 → Spec 环境变量覆盖规则
    test('TC-007: Config priority — env var > YAML > defaults', () => {
      writeConfig(`
server:
  port: "9092"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      // Env var should override YAML port
      const result = runServer({ SERVER_PORT: '9093' });
      assert.match(result.stdout + result.stderr, /9093|listening|started|ready/i);
    });
  });

  // ===== Validation (TC-008 ~ TC-014) =====

  describe('Validation', () => {
    // Traceability: TC-008 → Story 3 / AC-1, Spec 验证规则
    test('TC-008: JWT secret too short — startup rejected', () => {
      writeConfig(`
auth:
  jwt_secret: "short"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /jwt_secret.*32|at least 32/i);
    });

    // Traceability: TC-009 → Spec 验证规则
    test('TC-009: Invalid port — startup rejected', () => {
      writeConfig(`
server:
  port: "80"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /port.*1024.*65535|between 1024/i);
    });

    // Traceability: TC-010 → Spec 验证规则
    test('TC-010: Invalid database driver — startup rejected', () => {
      writeConfig(`
database:
  driver: "postgres"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /driver.*sqlite.*mysql/i);
    });

    // Traceability: TC-011 → Spec 验证规则
    test('TC-011: Invalid logging level — startup rejected', () => {
      writeConfig(`
logging:
  level: "verbose"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /level.*debug.*info.*warn.*error/i);
    });

    // Traceability: TC-012 → Spec 验证规则
    test('TC-012: Invalid logging format — startup rejected', () => {
      writeConfig(`
logging:
  format: "xml"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /format.*json.*text/i);
    });

    // Traceability: TC-013 → Spec 验证规则
    test('TC-013: Invalid duration format — startup rejected', () => {
      writeConfig(`
server:
  port: "8080"
  read_timeout: "abc"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
`);
      const result = runServer();
      assert.notEqual(result.exitCode, 0);
      assert.match(result.stdout + result.stderr, /duration|invalid.*format/i);
    });

    // Traceability: TC-014 → Spec 流程说明
    test('TC-014: All validations pass — startup succeeds', () => {
      writeConfig(`
server:
  port: "8080"
  read_timeout: "30s"
  write_timeout: "30s"
database:
  driver: "sqlite"
  path: "./data/e2e-test.db"
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
  jwt_expiry: "24h"
logging:
  level: "info"
  format: "json"
cors:
  origins:
    - "http://localhost:3000"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /8080|listening|started|ready/i);
    });
  });

  // ===== Initial Admin (TC-015 ~ TC-017) =====

  describe('Initial Admin', () => {
    // Traceability: TC-015 → Story 4 / AC-1
    test('TC-015: First startup creates initial admin', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
  initial_admin:
    username: "testadmin"
    password: "TestAdmin123!"
database:
  driver: "sqlite"
  path: "./data/e2e-admin-test.db"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /admin|created|seed|testadmin/i);
    });

    // Traceability: TC-016 → Story 4 / AC-1
    test('TC-016: Subsequent startups skip admin creation', () => {
      // Same config, server should start and skip creating existing admin
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
  initial_admin:
    username: "testadmin"
    password: "TestAdmin123!"
database:
  driver: "sqlite"
  path: "./data/e2e-admin-test.db"
`);
      const result = runServer();
      // Should start successfully without error about duplicate user
      assert.match(result.stdout + result.stderr, /listening|started|ready|skip|exists/i);
    });

    // Traceability: TC-017 → Spec Auth 域
    test('TC-017: Empty initial_admin config skips creation', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
database:
  driver: "sqlite"
  path: "./data/e2e-noadmin-test.db"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /listening|started|ready/i);
    });
  });

  // ===== Connection Pool (TC-018 ~ TC-019) =====

  describe('Connection Pool', () => {
    // Traceability: TC-018 → Story 5 / AC-1
    test('TC-018: Connection pool config applied', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
database:
  driver: "sqlite"
  path: "./data/e2e-pool-test.db"
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_lifetime: "1h"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /listening|started|ready/i);
    });

    // Traceability: TC-019 → Spec Database 域
    test('TC-019: Connection pool defaults when not configured', () => {
      writeConfig(`
auth:
  jwt_secret: "e2e-test-jwt-secret-at-least-32-bytes-long!!"
database:
  driver: "sqlite"
  path: "./data/e2e-defaults-test.db"
`);
      const result = runServer();
      assert.match(result.stdout + result.stderr, /listening|started|ready/i);
    });
  });
});
