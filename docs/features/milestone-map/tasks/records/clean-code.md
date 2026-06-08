---
status: "completed"
started: "2026-06-08 09:48"
completed: "2026-06-08 10:04"
time_spent: "~16m"
---

# Task Record: T-clean-code Simplify and Clean Code

## Summary
Fixed interrupted clean-code task. Restored calcCompletion and countRelatedMIs helper methods in milestone_service.go that were incorrectly removed. Completed dead code cleanup of calcOverallProgress from milestone_map_service.go. Added milestone_key to acceptedFields whitelist.

## Changes

### Files Created
无

### Files Modified
- backend/internal/service/milestone_service.go
- backend/internal/pkg/repo/helpers.go
- backend/internal/service/milestone_map_service.go
- backend/internal/service/milestone_map_service_test.go

### Key Decisions
- Restored calcCompletion and countRelatedMIs as test-only helper methods rather than removing their tests, since they verify computed field logic

## Test Results
- **Tests Executed**: Yes
- **Passed**: 28
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] All changed files pass compile, fmt, lint checks
- [x] No dead code or unnecessary complexity introduced by this feature

## Notes
Pre-existing failures in cmd/server (1 test) and frontend (~160 tests with document-is-not-defined env issue) are unrelated to this task. Backend service tests all pass.
