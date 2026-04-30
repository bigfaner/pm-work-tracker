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

*Last updated: 2026-04-30*
