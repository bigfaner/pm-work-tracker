---
status: "completed"
started: "2026-05-04 15:17"
completed: "2026-05-04 15:29"
time_spent: "~12m"
---

# Task Record: T-test-3 Run e2e Tests

## Summary
Executed e2e tests for decision-log feature. Results: 3 passed, 19 failed, 6 skipped out of 28 total. Root cause: handler double-resolves mainItemID (bizKey vs internal ID) causing 17 failures. Secondary issue: role creation with empty permissionCodes in test scripts. Report written to tests/e2e/decision-log/results/latest.md. Fix tasks needed for handler bug and test script issues.

## Changes

### Files Created
- tests/e2e/results/test-results.json

### Files Modified
- tests/e2e/decision-log/results/latest.md
- tests/e2e/playwright.config.ts

### Key Decisions
- Added JSON reporter to playwright.config.ts for structured test output collection
- Rebuilt backend binary to include handler wiring fix from disc-1
- Identified double-resolution bug in decision_log_handler.go as root cause of 17/19 failures

## Test Results
- **Passed**: 3
- **Failed**: 19
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] tests/e2e/decision-log/results/latest.md exists
- [ ] All tests pass (status = PASS in latest.md)

## Notes
Tests failed due to a code bug (handler double-resolution of mainItemID) not a test script issue. The handler resolves URL bizKey to internal ID, then passes that internal ID to the service which treats it as a bizKey again. This is a P0 backend fix. 6 UI tests skipped because they depend on createDecisionViaApi helper which hits the same bug. Fix tasks should be created for: (1) handler double-resolution bug, (2) test script role creation with empty permissionCodes.
