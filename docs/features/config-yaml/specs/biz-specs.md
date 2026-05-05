---
feature: "config-yaml"
generated: "2026-05-04"
status: draft
---

# Business Rules: config-yaml

## Configuration Management

### BIZ-001: Config Priority Chain

**Rule**: Configuration values follow a strict priority chain: environment variables > config.yaml > hardcoded defaults. Higher priority always wins.

**Context**: Production deployments need to override file-based config (e.g., secrets via CI/CD env injection) without modifying the YAML file. Local development relies on YAML defaults.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section "环境变量覆盖规则"

### BIZ-002: Startup Validation Principle

**Rule**: All configuration must be validated at application startup. Invalid configuration causes immediate exit with a descriptive error message. No partial startup with bad config.

**Context**: Running with invalid config (wrong port, missing JWT secret) causes runtime failures that are hard to diagnose. Fail-fast at startup is cheaper than debugging production incidents.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section "验证规则"

### BIZ-003: Env Variable Naming Convention

**Rule**: Environment variable names follow the YAML path converted to uppercase with underscores. Nested fields are joined with underscores. Arrays use comma-separated values.

**Context**: Consistent naming makes it easy to infer the env var name from the YAML structure without looking up documentation.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section "环境变量覆盖规则"

Examples:
- `auth.jwt_secret` → `AUTH_JWT_SECRET`
- `auth.initial_admin.username` → `AUTH_INITIAL_ADMIN_USERNAME`
- `cors.origins` → `CORS_ORIGINS=http://a.com,http://b.com`

## Bootstrap Operations

### BIZ-004: Seed Operation Idempotency

**Rule**: Seed/bootstrap operations (like creating initial admin) must be idempotent: skip if target already exists, warn but don't block on failure.

**Context**: Production deployments may restart multiple times. Non-idempotent seeds would fail on restart or create duplicates.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section "数据需求"

## Security

### BIZ-005: Sensitive Config via Env Vars

**Rule**: Secrets (JWT secret, admin password) must be settable via environment variables. Config files containing secrets must be gitignored. Only template files (`.example`) are committed.

**Context**: Prevents accidental secret leaks through version control. CI/CD pipelines inject secrets via env vars.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section "安全性需求"
