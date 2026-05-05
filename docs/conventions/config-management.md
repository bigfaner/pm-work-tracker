---
scope: backend
source: feature/config-yaml BIZ-001–005
---

# Configuration Management Rules

## BIZ-config-001: Config Priority Chain

Configuration values follow a strict priority chain:

**env vars > config.yaml > hardcoded defaults**

Higher priority always wins. Production deployments override file-based config via CI/CD env injection; local development relies on YAML defaults.

## BIZ-config-002: Startup Validation Principle

All configuration must be validated at application startup. Invalid configuration causes immediate exit with a descriptive error message. No partial startup with bad config.

**Why**: Running with invalid config (wrong port, missing JWT secret) causes runtime failures that are hard to diagnose. Fail-fast at startup is cheaper than debugging production incidents.

## BIZ-config-003: Env Variable Naming Convention

Environment variable names follow the YAML path converted to uppercase with underscores:

- `auth.jwt_secret` → `AUTH_JWT_SECRET`
- `auth.initial_admin.username` → `AUTH_INITIAL_ADMIN_USERNAME`
- `cors.origins` → `CORS_ORIGINS=http://a.com,http://b.com` (comma-separated for arrays)

Nested fields are joined with underscores. Arrays use comma-separated values.

## BIZ-config-004: Seed Operation Idempotency

Seed/bootstrap operations (like creating initial admin) must be idempotent: skip if target already exists, warn but don't block on failure.

**Why**: Production deployments may restart multiple times. Non-idempotent seeds would fail on restart or create duplicates.

## BIZ-config-005: Sensitive Config via Env Vars

Secrets (JWT secret, admin password) must be settable via environment variables. Config files containing secrets must be gitignored. Only template files (`.example`) are committed.

**Why**: Prevents accidental secret leaks through version control. CI/CD pipelines inject secrets via env vars.
