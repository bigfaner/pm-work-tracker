---
status: "completed"
started: "2026-06-03 21:12"
completed: "2026-06-03 21:39"
time_spent: "~27m"
---

# Task Record: fix-3 Fix: API tests need backend server for quality gate

## Summary
Fixed API tests ECONNREFUSED by adding Vitest global setup that builds and starts the Go backend server with in-memory SQLite on a random port before tests run, and tears it down after. Also increased login rate limit from 10 to 100 req/min in test mode to prevent 429 errors when many test files login simultaneously. Fixed hardcoded localhost:8080 URL in rbac-migration.spec.ts to use shared apiBaseUrl.

## Changes

### Files Created
- tests/api/vitest.global-setup.ts

### Files Modified
- tests/api/vitest.config.ts
- tests/api/roles/rbac-migration.spec.ts
- backend/internal/handler/router.go
- backend/internal/handler/router_test.go

### Key Decisions
- Used Vitest globalSetup to start/stop Go backend server per test run rather than requiring a pre-running server
- Used random port via node:net to avoid port conflicts
- Set E2E_CONFIG_PATH env var in globalSetup so shared helpers resolve the dynamic port
- Increased rate limit to 100 req/min when gin_mode=test to accommodate sequential test suite login bursts
- Fixed rbac-migration.spec.ts to use shared apiBaseUrl instead of hardcoded localhost:8080

## Test Results
- **Tests Executed**: Yes
- **Passed**: 207
- **Failed**: 0
- **Coverage**: 100.0%

## Acceptance Criteria
- [x] Backend server started and API tests pass (cd tests/api && npx vitest run exits 0)
- [x] Quality gate passes for task 3 after fix completion

## Notes
All 26 test files pass with 207 tests total. Backend unit tests also pass (rate limit test updated to account for higher test-mode limit). Frontend tests unaffected.
