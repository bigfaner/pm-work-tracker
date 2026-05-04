---
status: "completed"
started: "2026-05-04 15:41"
completed: "2026-05-04 15:41"
time_spent: ""
---

# Task Record: disc-2 Fix: DecisionLog handler double-resolution bug

## Summary
Fix double bizKey resolution in DecisionLog handler. Handler was resolving URL bizKey param to internal ID via ResolveBizKey, then passing that ID to service methods that treated it as a bizKey again. Changed handler to use ParseBizKeyParam and pass raw bizKey directly. Updated service interface (Create, List) and repo interface (ListByItem) to accept int64 bizKey instead of uint internal ID.

## Changes

### Files Created
无

### Files Modified
- backend/internal/handler/decision_log_handler.go
- backend/internal/handler/decision_log_handler_test.go
- backend/internal/handler/router_test_stubs.go
- backend/internal/repository/decision_log_repo.go
- backend/internal/repository/gorm/decision_log_repo.go
- backend/internal/repository/gorm/decision_log_repo_test.go
- backend/internal/service/decision_log_service.go
- backend/internal/service/decision_log_service_test.go

### Key Decisions
- Pass raw bizKey (int64) through handler->service->repo chain instead of resolving to internal uint ID, matching the data model where MainItemKey stores bizKey values

## Test Results
- **Passed**: 31
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] Handler no longer double-resolves mainItemID bizKey
- [x] Service receives and uses bizKey directly (int64)
- [x] Repo ListByItem uses int64 mainItemKey param
- [x] All 31 decision-log tests pass (handler 22 + service 4 + repo 5)

## Notes
Pre-existing unrelated CORS test failure (TestCORS_WildcardWhenNoOriginsConfigured) not related to this fix.
