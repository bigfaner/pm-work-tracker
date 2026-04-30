# Known E2E Test Failures

This file documents test failures that could not be resolved at graduation time.
Format: feature slug | test ID | reason | owner

---

## api-permission-test-coverage

| Test ID | Reason | Owner |
|---------|--------|-------|
| all (TC-001 to TC-012) | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with seeded admin/pm/member roles and a running Go server. | fanhuifeng |

**Notes**: TC-011 and TC-012 additionally require a Go backend at `/Users/fanhuifeng/Projects/Go/pm-work-tracker-2/backend` with `TestPermissionCodeCoverage` test. The path is hardcoded in the spec and may need updating.

---

## soft-delete-consistency

| Test ID | Reason | Owner |
|---------|--------|-------|
| all (TC-001 to TC-021) | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with admin credentials. | fanhuifeng |

**Notes**: TC-003, TC-004, TC-017, TC-019, TC-021 have partial coverage — full soft-delete verification requires direct DB manipulation (no public API endpoint for soft-deleting sub-items or main items). These tests verify the behavioral contract via available API endpoints only.

---

---

## bizkey-unification

| Test ID | Reason | Owner |
|---------|--------|-------|
| TC-001 to TC-005 (api) | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with seeded admin credentials and at least one team. | fanhuifeng |

**Notes**: TC-006 and TC-007 (cli) are pure build/grep checks and pass without a live server.

---

## config-yaml

| Test ID | Reason | Owner |
|---------|--------|-------|
| TC-001 to TC-019 (cli) | Require a compiled server binary at `./server` in project root. Binary not present at graduation time. Tests also require the server to start and print startup logs to stdout/stderr. | fanhuifeng |
| TC-020, TC-021 (api) | Backend server not running at graduation time (http://localhost:8080 unreachable). | fanhuifeng |

**Notes**: TC-011 (db-dialect-compat cli) is guarded by `MYSQL_HOST` env var and skipped when MySQL is unavailable.

---

## db-dialect-compat (re-graduated)

| Test ID | Reason | Owner |
|---------|--------|-------|
| TC-001 to TC-004 (api) | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with at least one team and item-pool entries. | fanhuifeng |
| TC-011 (cli) | Requires MySQL 8.0 instance. Skipped unless `MYSQL_HOST` env var is set. | fanhuifeng |

**Notes**: TC-006 to TC-010 (cli lint-staged tests) pass without a live server — they only require git and bash.

---

## improve-ui

| Test ID | Reason | Owner |
|---------|--------|-------|
| all API (TC-053 to TC-056) | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with seeded admin credentials and at least one team. | fanhuifeng |
| all UI (TC-001 to TC-052) | Backend and frontend servers not running at graduation time. UI tests require agent-browser CLI, live frontend on http://localhost:5173, and live backend on http://localhost:8080. | fanhuifeng |
| weekly-view UI (W01 to W30) | Same as above — requires live stack and agent-browser CLI. | fanhuifeng |

---

## schema-alignment-cleanup

| Test ID | Reason | Owner |
|---------|--------|-------|
| all API | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with admin credentials. | fanhuifeng |
| all UI | Backend and frontend servers not running at graduation time. UI tests require agent-browser CLI. | fanhuifeng |

---

## status-flow-optimization

| Test ID | Reason | Owner |
|---------|--------|-------|
| all API | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with seeded PM/executor tokens and at least one team. | fanhuifeng |
| all UI | Backend and frontend servers not running at graduation time. UI tests require agent-browser CLI. | fanhuifeng |

---

## user-management-reset-delete

| Test ID | Reason | Owner |
|---------|--------|-------|
| all API | Backend server not running at graduation time (http://localhost:8080 unreachable). Tests require live backend with superadmin credentials. | fanhuifeng |
| all UI | Backend and frontend servers not running at graduation time. UI tests require agent-browser CLI. | fanhuifeng |

---

*Last updated: 2026-04-30*
