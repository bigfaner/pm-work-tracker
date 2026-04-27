---
status: "completed"
started: "2026-04-27 01:36"
completed: "2026-04-27 01:40"
time_spent: "~4m"
---

# Task Record: T-test-2 Generate e2e Test Scripts

## Summary
Generated executable TypeScript e2e test scripts from test cases. Created api.spec.ts (4 tests: TC-001 to TC-004) and cli.spec.ts (6 tests: TC-006 to TC-011). TC-005 excluded from scripts as it runs via Go test runner. No UI tests generated (feature has no UI changes). All scripts use node:test + node:assert, fetch for API, child_process for CLI.

## Changes

### Files Created
- docs/features/db-dialect-compat/testing/scripts/helpers.ts
- docs/features/db-dialect-compat/testing/scripts/api.spec.ts
- docs/features/db-dialect-compat/testing/scripts/cli.spec.ts
- docs/features/db-dialect-compat/testing/scripts/package.json
- docs/features/db-dialect-compat/testing/scripts/tsconfig.json

### Files Modified
无

### Key Decisions
- TC-005 (go test regression) excluded from scripts: it runs via Go test runner, not TypeScript e2e scripts
- API tests use shared auth via getApiToken + createAuthCurl pattern (all 4 API tests are auth-required)
- CLI lint-staged tests (TC-006 to TC-009) create temp files, stage, attempt commit, verify blocked, then cleanup
- TC-010 uses --no-verify to avoid polluting git history while testing that clean code passes lint
- TC-011 uses direct MySQL CLI queries to verify RBAC initialization and HasColumn behavior

## Test Results
- **Passed**: 0
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] testing/scripts/package.json created
- [x] testing/scripts/helpers.ts created
- [x] At least one spec file generated (ui.spec.ts / api.spec.ts / cli.spec.ts)
- [x] Each test() includes traceability comment // Traceability: TC-NNN → {PRD Source}

## Notes
无
