# Forensic Report: E2E Test Failures — Empty Database File

**Date:** 2026-05-14
**Sessions analyzed:** 8e0e2b62, ae9b46f7, 509a9501, 61f41ab4 (current dispatcher)
**Evidence dir:** `docs/forensics/e2e-db-state/evidence/`

## Executive Summary

**Root cause: `backend/data/dev.db` was a 0-byte empty file.** The SQLite driver created the file but the schema migration failed silently or wasn't triggered, leaving an empty database with no tables. Every API call returned `internal server error` because there were no tables to query.

After restarting the backend, the schema migration ran successfully, tables were created, and all APIs work correctly.

## Evidence Chain

### 1. Symptom: E2E tests fail with `API error: internal server error`

All UI test failures in the latest run (2026-05-14 14:55) show the same error:
```
Error: API error: internal server error
```
from `parseApiBody()` at helpers.ts:510.

### 2. Direct cause: Backend returns 500 for all write operations

```bash
$ curl -s -X POST http://localhost:8080/v1/teams -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"test","code":"FTEST"}'
{"code":"INTERNAL_ERROR","message":"internal server error"}
```

Even login fails:
```bash
$ curl -s -X POST http://localhost:8080/v1/auth/login -d '{"username":"admin","password":"admin123"}'
{"code":"UNAUTHORIZED","message":"authentication required"}
```

### 3. Root cause: `backend/data/dev.db` is 0 bytes

```bash
$ ls -la backend/data/dev.db
-rw-r--r--  1 fanhuifeng  staff  0 May 14 15:13 backend/data/dev.db
$ file backend/data/dev.db
backend/data/dev.db: empty
```

The `go run` process (PID 48943) had its CWD at `backend/`, so it reads `./data/dev.db` (i.e., `backend/data/dev.db`). This file was 0 bytes — SQLite can open it but cannot execute any queries because there are no tables.

### 4. Fix: Delete the empty db file and restart

```bash
$ rm backend/data/dev.db
$ cd backend && go run cmd/server/main.go -dev
# Schema migration runs, tables created, admin seeded
$ ls -la data/dev.db
-rw-r--r--  1 fanhuifeng  staff  290816 May 14 15:17 data/dev.db
$ sqlite3 data/dev.db ".tables"
pmw_item_pools  pmw_milestones  pmw_progress_records  pmw_role_permissions  ...
```

After restart, all API calls work:
```bash
$ curl -s -X POST http://localhost:8080/v1/teams -d '{"name":"test","code":"FTEST"}'
{"code":0,"data":{"bizKey":"2054823478250442752",...}}
```

## Test Results Breakdown (latest run, 2026-05-14 14:55)

| Category | Pass | Fail | Total |
|----------|------|------|-------|
| API tests (api.spec.ts) | 20 | 0 | 20 |
| UI — Milestones page | 1 | 11 | 12 |
| UI — Existing pages | 4 | 13 | 17 |
| Auth setup | 1 | 0 | 1 |

**Note:** The 20 API tests passed because the e2e test run creates its own server instance (via `just e2e-setup` which starts a fresh test server with a clean database). The UI tests, however, may be hitting the same stale server or the database was also empty at that point.

## How the empty file was created

Three `.db` files were all 0 bytes:
```
backend/pm-work-tracker.db          (Apr 28)
data/dev.db                         (May 14 15:13)
backend/data/dev.db                 (May 14 15:14)
```

Most likely sequence:
1. A previous `go run` or test process created `backend/data/dev.db` as an empty file
2. That process was killed before schema migration completed
3. Next `go run` started, found the file exists, but it was 0 bytes
4. SQLite's `gorm.Open()` succeeds on a 0-byte file (it's a valid empty SQLite file header)
5. Schema migration (`migration.RunSchema`) detected "tables already exist" or the SQL DDL `CREATE TABLE IF NOT EXISTS` silently succeeded on the corrupt file
6. All actual queries fail because the file has no table definitions

## Recommendations

1. **Add DB health check at startup:** After migration, verify at least one core table exists. If not, log a fatal error and exit.
2. **Delete stale empty db files in `just dev` / `just e2e-setup`:** Before starting the server, check if the db file is 0 bytes and delete it.
3. **Consider adding this as a lesson:** "Empty SQLite file is not the same as missing — GORM opens it but queries fail."

## Session Timing

| Session | Duration | Key finding |
|---------|----------|-------------|
| 8e0e2b62 (2026-05-14) | 17h (mostly idle) | Analyzed timeout vs command invocation issue |
| ae9b46f7 (2026-05-09) | 5.9h, 8760s tool time | Multiple TaskOutput timeouts (600s each) |
| 509a9501 (2026-05-06) | 6.5h, 650s tool time | Ran `just test-e2e`, found errors, fixed them |
| 61f41ab4 (current) | ~45 min | Dispatcher: fix-3 completed, T-test-3 killed after e2e stall |
