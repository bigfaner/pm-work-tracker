---
feature: "config-yaml"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-19"
---

# Test Cases: config-yaml

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 2     |
| CLI  | 20    |
| **Total** | **22** |

---

## CLI Test Cases

### Config Loading (Story 1, Spec 配置域与字段)

## TC-001: Load config from config.yaml
- **Source**: Story 1 / AC-1
- **Type**: CLI
- **Pre-conditions**: config.yaml exists in project root with valid settings
- **Steps**:
  1. Create config.yaml with server.port, database.driver, auth.jwt_secret, etc.
  2. Start application
  3. Check application logs for loaded config values
- **Expected**: Application reads YAML config and loads all fields grouped by domain; no errors
- **Priority**: P0

## TC-002: Use hardcoded defaults when config.yaml missing
- **Source**: Spec 流程说明 — "若未找到，使用硬编码默认值"
- **Type**: CLI
- **Pre-conditions**: config.yaml does not exist; auth.jwt_secret is set via env var
- **Steps**:
  1. Delete or rename config.yaml
  2. Set AUTH_JWT_SECRET env var (>= 32 bytes)
  3. Start application
- **Expected**: Application uses hardcoded defaults for all fields; starts successfully
- **Priority**: P0

## TC-003: Malformed YAML file causes startup failure
- **Source**: Spec 流程说明 — YAML parsing
- **Type**: CLI
- **Pre-conditions**: config.yaml contains invalid YAML syntax
- **Steps**:
  1. Create config.yaml with broken syntax (e.g., unclosed bracket)
  2. Start application
- **Expected**: Application fails to start with clear error indicating YAML parse failure
- **Priority**: P0

### Environment Variable Override (Story 2, Spec 环境变量覆盖规则)

## TC-004: Env var overrides YAML value
- **Source**: Story 2 / AC-1
- **Type**: CLI
- **Pre-conditions**: config.yaml has auth.jwt_secret set to a weak value; env var AUTH_JWT_SECRET is set to >= 32 bytes
- **Steps**:
  1. Set AUTH_JWT_SECRET env var to a 32+ byte value
  2. Start application
  3. Verify the env var value is used
- **Expected**: Application uses env var value, overriding the YAML config
- **Priority**: P0

## TC-005: Nested field env var naming convention
- **Source**: Spec 环境变量覆盖规则 — "嵌套字段用下划线连接"
- **Type**: CLI
- **Pre-conditions**: config.yaml exists
- **Steps**:
  1. Set AUTH_INITIAL_ADMIN_USERNAME env var
  2. Start application
  3. Verify initial admin username from env var is used
- **Expected**: Nested path `auth.initial_admin.username` correctly maps to `AUTH_INITIAL_ADMIN_USERNAME`
- **Priority**: P0

## TC-006: Array field env var with comma separation
- **Source**: Spec 环境变量覆盖规则 — "数组字段用逗号分隔"
- **Type**: CLI
- **Pre-conditions**: config.yaml exists
- **Steps**:
  1. Set CORS_ORIGINS="http://a.com,http://b.com"
  2. Start application
  3. Verify CORS origins are parsed as array ["http://a.com", "http://b.com"]
- **Expected**: Comma-separated env var correctly parsed into string array
- **Priority**: P1

## TC-007: Config priority — env var > YAML > defaults
- **Source**: Spec 环境变量覆盖规则 — "配置优先级"
- **Type**: CLI
- **Pre-conditions**: config.yaml sets server.port to "9090"
- **Steps**:
  1. Set SERVER_PORT env var to "7070"
  2. Start application
  3. Check which port is used
  4. Remove env var, restart
  5. Check which port is used
- **Expected**: With env var: uses 7070 (env wins); without env var: uses 9090 (YAML wins); without YAML: uses 8080 (default)
- **Priority**: P0

### Validation (Story 3, Spec 验证规则)

## TC-008: JWT secret too short — startup rejected
- **Source**: Story 3 / AC-1, Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: auth.jwt_secret is less than 32 bytes (via YAML or env)
- **Steps**:
  1. Set auth.jwt_secret to "short" (< 32 bytes)
  2. Start application
- **Expected**: Application refuses to start; outputs "auth.jwt_secret must be at least 32 bytes"
- **Priority**: P0

## TC-009: Invalid port — startup rejected
- **Source**: Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: server.port is outside 1024-65535
- **Steps**:
  1. Set server.port to "80" (< 1024)
  2. Start application
  3. Repeat with port "99999" (> 65535)
- **Expected**: Application refuses to start; outputs "server.port must be between 1024 and 65535"
- **Priority**: P0

## TC-010: Invalid database driver — startup rejected
- **Source**: Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: database.driver is not "sqlite" or "mysql"
- **Steps**:
  1. Set database.driver to "postgres"
  2. Start application
- **Expected**: Application refuses to start; outputs "database.driver must be sqlite or mysql"
- **Priority**: P0

## TC-011: Invalid logging level — startup rejected
- **Source**: Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: logging.level is not one of debug/info/warn/error
- **Steps**:
  1. Set logging.level to "verbose"
  2. Start application
- **Expected**: Application refuses to start; outputs "logging.level must be one of: debug, info, warn, error"
- **Priority**: P1

## TC-012: Invalid logging format — startup rejected
- **Source**: Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: logging.format is not "json" or "text"
- **Steps**:
  1. Set logging.format to "xml"
  2. Start application
- **Expected**: Application refuses to start; outputs "logging.format must be json or text"
- **Priority**: P1

## TC-013: Invalid duration format — startup rejected
- **Source**: Spec 验证规则
- **Type**: CLI
- **Pre-conditions**: A duration field contains invalid format
- **Steps**:
  1. Set server.read_timeout to "abc" (not a valid Go duration)
  2. Start application
- **Expected**: Application refuses to start; outputs "invalid duration format for server.read_timeout"
- **Priority**: P1

## TC-014: All validations pass — startup succeeds
- **Source**: Spec 流程说明 — "验证通过 → 使用配置初始化"
- **Type**: CLI
- **Pre-conditions**: All config values are valid
- **Steps**:
  1. Create valid config.yaml with all required fields
  2. Start application
- **Expected**: Application starts successfully; all components initialized with config values
- **Priority**: P0

### Initial Admin (Story 4, Spec 数据需求)

## TC-015: First startup creates initial admin
- **Source**: Story 4 / AC-1
- **Type**: CLI
- **Pre-conditions**: Database is empty; config.yaml has auth.initial_admin.username and password both non-empty
- **Steps**:
  1. Set auth.initial_admin.username and auth.initial_admin.password in config
  2. Start application (fresh database)
  3. Check database for admin user
- **Expected**: Admin user created with specified username; password is hashed (not plaintext)
- **Priority**: P0

## TC-016: Subsequent startups skip admin creation
- **Source**: Story 4 / AC-1 — "应用再次启动（用户已存在）"
- **Type**: CLI
- **Pre-conditions**: Admin user already exists in database
- **Steps**:
  1. Start application again (admin user already exists)
  2. Check admin user count in database
- **Expected**: No duplicate admin created; existing user unchanged; idempotent operation
- **Priority**: P0

## TC-017: Empty initial_admin config skips creation
- **Source**: Spec Auth 域 — "空则跳过"
- **Type**: CLI
- **Pre-conditions**: auth.initial_admin.username is empty in config
- **Steps**:
  1. Set auth.initial_admin.username to ""
  2. Start application
- **Expected**: No admin user created; no error; application starts normally
- **Priority**: P1

### Connection Pool (Story 5, Spec Database 域)

## TC-018: Connection pool config applied to database
- **Source**: Story 5 / AC-1
- **Type**: CLI
- **Pre-conditions**: config.yaml sets database.max_open_conns to 20
- **Steps**:
  1. Set database.max_open_conns: 20, max_idle_conns: 5, conn_max_lifetime: 1h
  2. Start application
  3. Verify database connection pool settings
- **Expected**: Connection pool configured with max_open=20, max_idle=5, max_lifetime=1h
- **Priority**: P0

## TC-019: Connection pool defaults when not configured
- **Source**: Spec Database 域 — 默认值
- **Type**: CLI
- **Pre-conditions**: config.yaml does not specify connection pool fields
- **Steps**:
  1. Start application without pool config
  2. Verify connection pool uses defaults
- **Expected**: Defaults applied: max_open=10, max_idle=5, conn_max_lifetime=1h
- **Priority**: P1

---

## API Test Cases

## TC-020: Application starts and serves HTTP on configured port
- **Source**: Spec Server 域
- **Type**: API
- **Pre-conditions**: Valid config with server.port set
- **Steps**:
  1. Set server.port to "9090"
  2. Start application
  3. Send HTTP request to localhost:9090
- **Expected**: Application responds on port 9090; not on default port 8080
- **Priority**: P0

## TC-021: CORS origins from config applied to responses
- **Source**: Spec CORS 域
- **Type**: API
- **Pre-conditions**: config.yaml sets cors.origins to ["http://localhost:3000"]
- **Steps**:
  1. Configure CORS origins
  2. Start application
  3. Send preflight OPTIONS request with Origin: http://localhost:3000
  4. Send preflight OPTIONS request with Origin: http://evil.com
- **Expected**: Allowed origin receives Access-Control-Allow-Origin header; disallowed origin does not
- **Priority**: P1

---

## Traceability

| TC ID | Source | Type | Priority |
|-------|--------|------|----------|
| TC-001 | Story 1 / AC-1 | CLI | P0 |
| TC-002 | Spec 流程说明 | CLI | P0 |
| TC-003 | Spec 流程说明 | CLI | P0 |
| TC-004 | Story 2 / AC-1 | CLI | P0 |
| TC-005 | Spec 环境变量覆盖规则 | CLI | P0 |
| TC-006 | Spec 环境变量覆盖规则 | CLI | P1 |
| TC-007 | Spec 环境变量覆盖规则 | CLI | P0 |
| TC-008 | Story 3 / AC-1, Spec 验证规则 | CLI | P0 |
| TC-009 | Spec 验证规则 | CLI | P0 |
| TC-010 | Spec 验证规则 | CLI | P0 |
| TC-011 | Spec 验证规则 | CLI | P1 |
| TC-012 | Spec 验证规则 | CLI | P1 |
| TC-013 | Spec 验证规则 | CLI | P1 |
| TC-014 | Spec 流程说明 | CLI | P0 |
| TC-015 | Story 4 / AC-1 | CLI | P0 |
| TC-016 | Story 4 / AC-1 | CLI | P0 |
| TC-017 | Spec Auth 域 | CLI | P1 |
| TC-018 | Story 5 / AC-1 | CLI | P0 |
| TC-019 | Spec Database 域 | CLI | P1 |
| TC-020 | Spec Server 域 | API | P0 |
| TC-021 | Spec CORS 域 | API | P1 |
