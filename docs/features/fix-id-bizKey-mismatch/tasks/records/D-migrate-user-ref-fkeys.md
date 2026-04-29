---
status: "completed"
started: "2026-04-29 01:51"
completed: "2026-04-29 01:51"
time_spent: ""
---

# Task Record: D Migrate all user-reference foreign keys to store BizKey

## Summary
Migrated all user-reference foreign keys (MainItem.ProposerKey, ItemPool.SubmitterKey, ItemPool.ReviewerKey, ProgressRecord.AuthorKey, StatusHistory.ChangedBy) from storing auto-increment uint ID to storing BizKey int64. Updated service interfaces to accept callerBizKey int64, handlers to use GetUserBizKey instead of GetUserID for FK storage, and fixed all test mocks/stubs to match new signatures. Fixed router_test.go seed data to ensure user BizKey matches auto-inc ID so GetUserBizKey returns correct values in handler tests.

## Changes

### Files Created
无

### Files Modified
- backend/internal/service/status_history_helper.go
- backend/internal/service/main_item_service.go
- backend/internal/service/sub_item_service.go
- backend/internal/service/progress_service.go
- backend/internal/service/item_pool_service.go
- backend/internal/handler/main_item_handler.go
- backend/internal/handler/sub_item_handler.go
- backend/internal/handler/progress_handler.go
- backend/internal/handler/item_pool_handler.go
- backend/internal/handler/router_test_stubs.go
- backend/internal/handler/permission_matrix_test.go
- backend/internal/handler/main_item_handler_test.go
- backend/internal/handler/sub_item_handler_test.go
- backend/internal/handler/progress_handler_test.go
- backend/internal/handler/item_pool_handler_test.go
- backend/internal/handler/router_test.go
- backend/internal/service/main_item_service_test.go
- backend/internal/service/sub_item_service_test.go
- backend/internal/service/progress_service_test.go
- backend/internal/service/item_pool_service_test.go

### Key Decisions
- Service methods receive callerBizKey int64 (snowflake) instead of callerID uint (auto-inc) for user-reference FK fields
- Handlers use GetUserBizKey(c) for FK storage, GetUserID(c) retained only for lookup-by-ID purposes
- StatusHistory.ChangedBy parameter renamed from changedBy uint to changedByBizKey int64, removing int64() cast
- Test seed data in router_test.go updated to set BizKey=int64(i) for loop-created users, ensuring BizKey matches auto-inc ID
- Removed redundant explicit testmember5 user creation; loop now creates all users 3-5 with matching BizKeys

## Test Results
- **Passed**: 909
- **Failed**: 0
- **Coverage**: 87.7%

## Acceptance Criteria
- [x] MainItem.ProposerKey stores BizKey from pmBizKey parameter
- [x] ItemPool.SubmitterKey stores BizKey from submitterBizKey parameter
- [x] ItemPool.ReviewerKey stores BizKey from pmBizKey parameter
- [x] ProgressRecord.AuthorKey stores BizKey from authorBizKey parameter
- [x] StatusHistory.ChangedBy stores BizKey from changedByBizKey parameter
- [x] All service tests pass
- [x] All handler tests pass
- [x] All integration tests pass
- [x] go build ./... succeeds
- [x] go vet ./... clean

## Notes
无
