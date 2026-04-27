---
feature: "db-dialect-compat"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-27"
---

# Test Cases: db-dialect-compat

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 5     |
| CLI  | 6     |
| **Total** | **11** |

---

## UI Test Cases

No UI test cases. PRD sections 5.1-5.3 explicitly state this feature has no UI changes.

---

## API Test Cases

## TC-001: Convert item-pool entry to main item returns 200 on MySQL
- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/item-pool
- **Test ID**: api/item-pool/convert-to-main-returns-200-on-mysql
- **Pre-conditions**: Local MySQL 8.0 instance with schema imported; app started in MySQL mode (`database.driver: mysql`); team and item-pool entry exist
- **Steps**:
  1. Send `POST /api/v1/teams/:id/item-pool/:id/convert-to-main` with valid auth
- **Expected**: Response status 200; main item created with code in format `{teamCode}-{seq:05d}` (e.g. `TEAM-00042`)
- **Priority**: P0

## TC-002: Main item code increments sequentially on MySQL
- **Source**: Story 1 / AC-1 (main item numbering rule)
- **Type**: API
- **Target**: api/item-pool
- **Test ID**: api/item-pool/main-item-code-increments-sequentially-on-mysql
- **Pre-conditions**: Local MySQL 8.0 instance with schema imported; app started in MySQL mode; team exists with code prefix (e.g. `TEAM`)
- **Steps**:
  1. Convert first item-pool entry to main item, record code (e.g. `TEAM-00001`)
  2. Convert second item-pool entry to main item, record code
  3. Convert third item-pool entry to main item, record code
- **Expected**: Codes are strictly sequential with no gaps: `TEAM-00001`, `TEAM-00002`, `TEAM-00003`
- **Priority**: P0

## TC-003: Sub item code increments sequentially on MySQL
- **Source**: Story 1 / AC-1 (sub item numbering rule)
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/sub-item-code-increments-sequentially-on-mysql
- **Pre-conditions**: Local MySQL 8.0 instance with schema imported; app started in MySQL mode; main item exists (e.g. `TEAM-00042`)
- **Steps**:
  1. Create first sub item under the main item, record code
  2. Create second sub item, record code
  3. Create third sub item, record code
- **Expected**: Sub item codes follow format `{mainCode}-{seq:02d}` and are sequential: `TEAM-00042-01`, `TEAM-00042-02`, `TEAM-00042-03`
- **Priority**: P0

## TC-004: Convert item-pool entry returns 200 on SQLite
- **Source**: Story 5 / AC-1
- **Type**: API
- **Target**: api/item-pool
- **Test ID**: api/item-pool/convert-to-main-returns-200-on-sqlite
- **Pre-conditions**: App started in SQLite mode (default config); team and item-pool entry exist
- **Steps**:
  1. Send `POST /api/v1/teams/:id/item-pool/:id/convert-to-main` with valid auth
- **Expected**: Response status 200; main item created with correct sequential code. No regression from dialect changes.
- **Priority**: P0

## TC-005: All existing tests pass on SQLite after dialect changes
- **Source**: Story 5 / AC-1
- **Type**: API
- **Target**: api/regression
- **Test ID**: api/regression/all-existing-tests-pass-on-sqlite
- **Pre-conditions**: App in SQLite mode (default config); dialect module implemented
- **Steps**:
  1. Run `go test ./internal/... ./config/... ./cmd/...`
- **Expected**: All tests pass with no failures; zero regressions from dialect abstraction
- **Priority**: P0

---

## CLI Test Cases

## TC-006: Git commit blocked when repo layer contains hardcoded SUBSTR
- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/lint-staged
- **Test ID**: cli/lint-staged/commit-blocked-on-hardcoded-substr
- **Pre-conditions**: Developer has modified a `.go` file under `backend/internal/repository/` containing raw `SUBSTR(` in a string literal
- **Steps**:
  1. Stage the modified file
  2. Run `git commit` to trigger lint-staged
- **Expected**: Commit is blocked with error message advising use of `dialect` package for dialect-safe SQL generation
- **Priority**: P0

## TC-007: Git commit blocked when repo layer contains hardcoded CAST
- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/lint-staged
- **Test ID**: cli/lint-staged/commit-blocked-on-hardcoded-cast
- **Pre-conditions**: Developer has modified a `.go` file under `backend/internal/repository/` containing raw `CAST(` in a string literal
- **Steps**:
  1. Stage the modified file
  2. Run `git commit` to trigger lint-staged
- **Expected**: Commit is blocked with error message advising use of `dialect` package
- **Priority**: P0

## TC-008: Git commit blocked when repo layer contains hardcoded datetime
- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/lint-staged
- **Test ID**: cli/lint-staged/commit-blocked-on-hardcoded-datetime
- **Pre-conditions**: Developer has modified a `.go` file under `backend/internal/repository/` containing raw `datetime(` in a string literal
- **Steps**:
  1. Stage the modified file
  2. Run `git commit` to trigger lint-staged
- **Expected**: Commit is blocked with error message advising use of `dialect` package
- **Priority**: P0

## TC-009: Git commit blocked when repo layer contains hardcoded pragma_
- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/lint-staged
- **Test ID**: cli/lint-staged/commit-blocked-on-hardcoded-pragma
- **Pre-conditions**: Developer has modified a `.go` file under `backend/internal/repository/` containing raw `pragma_` in a string literal
- **Steps**:
  1. Stage the modified file
  2. Run `git commit` to trigger lint-staged
- **Expected**: Commit is blocked with error message advising use of `dialect` package
- **Priority**: P0

## TC-010: Git commit passes when repo layer uses dialect package
- **Source**: Story 3 / AC-2
- **Type**: CLI
- **Target**: cli/lint-staged
- **Test ID**: cli/lint-staged/commit-passes-with-dialect-package
- **Pre-conditions**: Developer has modified a `.go` file under `backend/internal/repository/` that uses `dialect.CastInt()`, `dialect.Substr()`, or `dialect.Now()` instead of hardcoded SQLite keywords
- **Steps**:
  1. Stage the modified file
  2. Run `git commit` to trigger lint-staged
- **Expected**: Commit succeeds with no errors (no false positive)
- **Priority**: P0

## TC-011: Fresh MySQL startup initializes RBAC with preset roles
- **Source**: Story 2 / AC-1, Story 4 / AC-1
- **Type**: CLI
- **Target**: cli/startup
- **Test ID**: cli/startup/fresh-mysql-startup-initializes-rbac
- **Pre-conditions**: Local MySQL 8.0 instance with schema imported but no migration records; app configured with `database.driver: mysql`, `auto_schema: false`
- **Steps**:
  1. Start the application
  2. Wait for startup to complete
  3. Query `roles` table for count
  4. Query `role_permissions` table for permission codes
  5. Verify `HasColumn(db, 'pmw_team_members', 'role_key')` returns `true`
  6. Verify `HasColumn(db, 'pmw_team_members', 'nonexistent')` returns `false`
- **Expected**: App starts without SQL syntax errors; `roles` table contains 3 preset roles (superadmin, pm, member); `role_permissions` contains corresponding permission codes; HasColumn returns correct boolean values
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | API | api/item-pool | P0 |
| TC-002 | Story 1 / AC-1 | API | api/item-pool | P0 |
| TC-003 | Story 1 / AC-1 | API | api/main-items | P0 |
| TC-004 | Story 5 / AC-1 | API | api/item-pool | P0 |
| TC-005 | Story 5 / AC-1 | API | api/regression | P0 |
| TC-006 | Story 3 / AC-1 | CLI | cli/lint-staged | P0 |
| TC-007 | Story 3 / AC-1 | CLI | cli/lint-staged | P0 |
| TC-008 | Story 3 / AC-1 | CLI | cli/lint-staged | P0 |
| TC-009 | Story 3 / AC-1 | CLI | cli/lint-staged | P0 |
| TC-010 | Story 3 / AC-2 | CLI | cli/lint-staged | P0 |
| TC-011 | Story 2 / AC-1, Story 4 / AC-1 | CLI | cli/startup | P0 |
