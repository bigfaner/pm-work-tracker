---
created: "2026-05-14"
tags: [local-dev-deployment, testing]
---

# Empty SQLite File Causes Silent API Failures

## Problem

Backend returns `{"code":"INTERNAL_ERROR","message":"internal server error"}` for all API calls, including login. No useful error message — just internal server error on every endpoint.

## Root Cause

Causal chain:
1. **Symptom**: All API endpoints return 500 internal server error
2. **Direct cause**: SQLite database file exists but has 0 bytes — no tables, no data
3. **Root cause**: A previous process (e.g. `go run` or test) created the file (`data/dev.db`) but was killed before schema migration completed. Next startup found the file exists, GORM opened it successfully (0-byte is a valid empty SQLite file), but `CREATE TABLE IF NOT EXISTS` may have silently passed while subsequent queries failed
4. **Trigger condition**: Stale `.db` files from interrupted process lifecycles

**Why it's hard to diagnose**: GORM's `gorm.Open()` does NOT error on a 0-byte SQLite file. The file appears valid. Only actual queries fail.

## Solution

Delete the empty file and restart the server:
```bash
rm backend/data/dev.db
cd backend && go run cmd/server/main.go -dev
# Schema migration runs, tables created, admin seeded
```

## Reusable Pattern

**When backend returns INTERNAL_ERROR on all endpoints:**
1. Check the database file size: `ls -la backend/data/dev.db`
2. If 0 bytes — delete it and restart
3. If file exists but errors persist: `sqlite3 backend/data/dev.db ".tables"` to verify tables exist

**Prevention**: Add a startup health check that verifies at least one core table exists after migration. If the DB file is empty/corrupt, log a fatal error and exit instead of serving broken requests.

## Example

```bash
# Diagnosis
$ ls -la backend/data/dev.db
-rw-r--r--  1 fanhuifeng  staff  0 May 14 15:13 backend/data/dev.db   # 0 bytes!

# Fix
$ rm backend/data/dev.db
$ cd backend && go run cmd/server/main.go -dev
# Output: schema migration runs, tables created

$ ls -la data/dev.db
-rw-r--r--  1 fanhuifeng  staff  290816 May 14 15:17 data/dev.db   # populated
```

## Related Files

- `backend/cmd/server/main.go` — startup sequence with schema migration
- `backend/config.yaml` — `database.path: ./data/dev.db`
- `backend/internal/migration/runner.go` — `RunSchema()` executes DDL
