---
status: "completed"
started: "2026-04-28 01:09"
completed: "2026-04-28 01:14"
time_spent: "~5m"
---

# Task Record: T-test-2 Generate e2e Test Scripts

## Summary
Generated e2e test scripts (helpers.ts + api.spec.ts) covering all 21 API test cases from TC-001 through TC-021. Tests use node:test + node:assert/strict + fetch, following established project patterns. Covers roles, users, teams, members, main items, sub-items, item pools, permissions, and schema unique index verification.

## Changes

### Files Created
- docs/features/soft-delete-consistency/testing/scripts/helpers.ts
- docs/features/soft-delete-consistency/testing/scripts/api.spec.ts

### Files Modified
无

### Key Decisions
- helpers.ts is identical to other feature test scripts (curl wrapper, no external deps)
- TC-003/TC-004/TC-019/TC-021 include inline notes explaining that full soft-delete verification requires direct DB manipulation since no public API endpoint triggers sub-item soft-delete
- TC-006 verifies ProgressRecord accessibility as a proxy for the non-soft-deletable entity behavior
- TC-007/TC-008/TC-011 verify permission exclusion indirectly via member list since test user password is unknown
- All test cases are API-only, consistent with backend-only feature having no UI changes

## Test Results
- **Passed**: 0
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] All 21 test cases from test-cases.md have corresponding test scripts
- [x] Scripts use node:test + node:assert/strict (zero external test frameworks)
- [x] Scripts use fetch for HTTP requests via helpers.ts curl wrapper
- [x] Each test has traceability comment mapping back to TC ID and PRD source
- [x] helpers.ts follows same pattern as other feature test scripts

## Notes
无
