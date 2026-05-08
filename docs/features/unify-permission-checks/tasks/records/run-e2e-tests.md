---
status: "blocked"
started: "2026-05-09 02:58"
completed: "N/A"
time_spent: ""
---

# Task Record: T-test-3 Run e2e Tests

## Summary
Executed e2e test scripts for unify-permission-checks. 0/38 passed, 3 failed, 35 skipped (cascade from beforeAll failure). Two root causes found: (1) ui.spec.ts uses wrong cwd 'ui' instead of 'frontend' in runCli calls (TC-001, TC-002); (2) RequirePermission middleware lacks SuperAdmin bypass for non-team-scoped routes, causing setupRbacFixtures to fail on GET /admin/roles (TC-004 cascade). Created two P0 fix tasks: disc-2 (test script cwd) and disc-3 (middleware bypass). Results report written to tests/e2e/features/unify-permission-checks/results/latest.md.

## Changes

### Files Created
- tests/e2e/features/unify-permission-checks/results/latest.md

### Files Modified
无

### Key Decisions
- Marked T-test-3 as blocked pending fix tasks disc-2 and disc-3
- Created disc-2: fix test script cwd 'ui' -> 'frontend' in ui.spec.ts
- Created disc-3: add SuperAdmin bypass to RequirePermission middleware for non-team-scoped routes

## Test Results
- **Passed**: 0
- **Failed**: 3
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] tests/e2e/features/unify-permission-checks/results/latest.md exists
- [ ] All tests pass (status = PASS in latest.md)

## Notes
TC-001 and TC-002: runCli cwd bug ('ui' should be 'frontend'). TC-004: RequirePermission middleware only checks roleRepo.HasPermission() for non-team routes, missing isSuperAdmin bypass that TeamScopeMiddleware provides for team-scoped routes. All 35 API tests skipped due to cascade from TC-004 beforeAll failure.
