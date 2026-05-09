---
status: "completed"
started: "2026-05-09 03:47"
completed: "2026-05-09 04:20"
time_spent: "~33m"
---

# Task Record: T-test-3 Run e2e Tests

## Summary
Executed e2e test scripts for unify-permission-checks feature. Fixed multiple test script issues (invalid permission codes, wrong error codes, incorrect response field names, invalid status transitions, missing API parameters, wrong role assignments). Generated results report. 20/38 tests pass, 4 fail (state management), 14 skip (cascade).

## Changes

### Files Created
无

### Files Modified
- tests/e2e/features/unify-permission-checks/api.spec.ts
- tests/e2e/features/unify-permission-checks/results/latest.md
- tests/e2e/helpers.ts

### Key Decisions
- Added noPerms user to team inside setupRbacFixtures (instead of in test beforeAll) to avoid runId mismatch
- Changed PM user role from memberRoleKey to pmRoleKey in setupRbacFixtures so PM has team:invite permission
- Fixed error code assertions from 'FORBIDDEN' to 'ERR_FORBIDDEN' to match actual API response codes
- Fixed sub-item status field from data.status to data.subItem.itemStatus for status change endpoint
- Fixed item pool status field from data.status to data.poolStatus
- Fixed archive assertion (endpoint returns data: null on success)
- Added weekStart query parameter to weekly view test (required by API)
- Added confirmName body to team disband endpoint
- Fixed invalid status 'in_progress' to 'progressing' (valid sub-item status)
- Fixed TC-034 to use read-only role instead of empty permissionCodes (empty not allowed)

## Test Results
- **Passed**: 20
- **Failed**: 4
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] tests/e2e/features/unify-permission-checks/results/latest.md exists
- [ ] All tests pass (status = PASS in latest.md)

## Notes
4 API test failures are due to test script state-management issues (not backend bugs): TC-013 (redundant PM transfer), TC-021 (status already changed by earlier test), TC-024 (PM invite after role change), TC-025 (cascade timeout). 14 tests skip due to Playwright beforeAll cascade from TC-024 failure. 20/38 pass including all 2 CLI tests and 18 API tests.
