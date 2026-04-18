---
title: config.yaml Configuration Management
status: draft
created: 2026-04-18
---

# Proposal: config.yaml Configuration Management

## Problem

Current backend configuration relies solely on environment variables (`os.Getenv`), with 7 flat fields in `config.Config`. This works but has limitations:

- No structure or grouping — all vars are flat, making it hard to understand config relationships
- No new config categories — adding logging, connection pools, or auth policy requires more env vars
- No defaults file — developers must read `.env.example` and manually set each var
- Sensitive values mixed with regular config — no clear separation

## Proposed Solution

Introduce `config.yaml` as the primary configuration source, organized by functional domain. Environment variables override specific fields for deployment flexibility.

### Configuration Priority

```
env vars > config.yaml > hardcoded defaults
```

- `config.yaml` provides structured defaults for development
- Environment variables override any field (for production / secrets)
- Backward compatible: existing env var names (`JWT_SECRET`, `DB_DRIVER`, etc.) continue to work

### YAML Structure

```yaml
server:
  port: "8080"
  gin_mode: ""              # "release" for production
  read_timeout: 30s         # HTTP read timeout
  write_timeout: 30s        # HTTP write timeout
  max_body_size: 10485760   # 10MB in bytes

database:
  driver: sqlite            # "sqlite" or "mysql"
  path: ./data/dev.db       # sqlite path
  url: ""                   # mysql DSN
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 1h

auth:
  jwt_secret: ""            # MUST set via env or yaml; min 32 bytes
  jwt_expiry: 24h
  initial_admin:
    username: admin
    password: admin123      # SHOULD override via env in production

cors:
  origins: []               # e.g. ["http://localhost:3000"]

logging:
  level: info               # debug, info, warn, error
  format: json              # json or text
```

### Environment Variable Override

Two override mechanisms:

1. **Backward-compatible env vars** — existing names still work:
   - `JWT_SECRET` → `auth.jwt_secret`
   - `DB_DRIVER` → `database.driver`
   - `DB_PATH` → `database.path`
   - `DATABASE_URL` → `database.url`
   - `CORS_ORIGINS` → `cors.origins` (comma-separated)
   - `PORT` → `server.port`
   - `GIN_MODE` → `server.gin_mode`

2. **Dotted env prefix** — for new fields without legacy names:
   - `CFG_SERVER_READ_TIMEOUT` → `server.read_timeout`
   - `CFG_DATABASE_MAX_OPEN_CONNS` → `database.max_open_conns`
   - `CFG_AUTH_JWT_EXPIRY` → `auth.jwt_expiry`
   - `CFG_LOGGING_LEVEL` → `logging.level`

Priority within env vars: backward-compatible names take precedence over `CFG_` prefixed names.

### New Configuration Categories

| Category | New Fields | Purpose |
|----------|-----------|---------|
| Server | `read_timeout`, `write_timeout`, `max_body_size` | HTTP server hardening |
| Database | `max_open_conns`, `max_idle_conns`, `conn_max_lifetime` | Connection pool tuning |
| Auth | `jwt_expiry`, `initial_admin.username`, `initial_admin.password` | Token policy + bootstrap admin |
| Logging | `level`, `format` | Observability control |

### Validation

- `auth.jwt_secret` must be at least 32 bytes (same as current)
- Duration fields (`read_timeout`, `conn_max_lifetime`, `jwt_expiry`) parsed from Go duration strings (`30s`, `1h`, `24h`)
- `database.driver` must be `sqlite` or `mysql`
- `server.port` must be a valid port number
- `logging.level` must be one of `debug`, `info`, `warn`, `error`

## Scope

### In Scope

- Replace `config.Config` struct with categorized YAML-backed struct
- YAML parsing with env var override (both backward-compatible and `CFG_` prefix)
- Validation with clear error messages
- `config.yaml` example file with documented defaults
- Update `cmd/server/main.go` to use new config
- Database connection pool configuration wired into `config.InitDB()`
- Initial admin account seeding on first startup

### Out of Scope

- Frontend configuration (stays with Vite `.env`)
- Multi-environment profile layering
- Hot config reload
- Config file encryption
- Config API endpoints (read/write config at runtime)

## Impact

- **Breaking change**: `Config` struct changes shape. All consumers updated.
- **Backward compatible**: Existing env vars still work without `config.yaml`.
- **New dependency**: YAML parsing library needed.
