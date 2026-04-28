---
status: "completed"
started: "2026-04-29 02:08"
completed: "2026-04-29 02:08"
time_spent: ""
---

# Task Record: E-F AssigneeKey read-side BizKey unification + StatusHistory.ItemKey BizKey migration

## Summary
Unified AssigneeKey read-side to use BizKey lookups (FindByBizKeys) instead of auto-inc ID lookups (FindByIDs). Fixed StatusHistory.ItemKey to store BizKey instead of auto-inc ID. Updated sub_item_handler assignee checks from GetUserID to GetUserBizKey. Updated progress_handler, item_pool_handler, and view_service batch lookups to use FindByBizKeys. Fixed all test mocks to implement the new UserRepo.FindByBizKeys method.

## Changes

### Files Created
无

### Files Modified
- backend/internal/repository/user_repo.go
- backend/internal/repository/gorm/user_repo.go
- backend/internal/handler/sub_item_handler.go
- backend/internal/handler/progress_handler.go
- backend/internal/handler/item_pool_handler.go
- backend/internal/service/view_service.go
- backend/internal/service/sub_item_service.go
- backend/internal/service/main_item_service.go
- backend/internal/service/progress_service.go
- backend/internal/handler/router_test_stubs.go
- backend/internal/handler/team_handler_test.go
- backend/internal/handler/progress_handler_test.go
- backend/internal/handler/item_pool_handler_test.go
- backend/internal/middleware/auth_test.go
- backend/internal/service/auth_service_test.go
- backend/internal/service/role_service_test.go
- backend/internal/service/admin_service_test.go
- backend/internal/service/team_service_test.go
- backend/internal/service/view_service_test.go
- backend/internal/service/main_item_service_test.go
- backend/internal/service/sub_item_service_test.go

### Key Decisions
- Added FindByBizKeys to UserRepo interface for batch BizKey lookups, mirroring FindByIDs pattern
- Changed resolveSubItemAssigneeNames return type from map[uint]string to map[int64]string since AssigneeKey stores BizKey
- Changed resolveAssigneeNames (table view) to use FindByBizKeys with int64 keys parsed from AssigneeID strings
- Changed buildProgressRecordVOs and buildItemPoolVOs to use FindByBizKeys for user name resolution
- Changed sub_item_handler Update/ChangeStatus assignee checks from GetUserID/uint comparison to GetUserBizKey/int64 comparison
- Changed RecordStatusChange callers to pass item.BizKey instead of int64(itemID) for StatusHistory.ItemKey
- Removed unused mapKeysToSlice helper from item_pool_handler.go

## Test Results
- **Passed**: 320
- **Failed**: 0
- **Coverage**: 83.0%

## Acceptance Criteria
- [x] UserRepo.FindByBizKeys added to interface and GORM implementation
- [x] sub_item_handler Update/ChangeStatus use GetUserBizKey instead of GetUserID for assignee check
- [x] resolveSubItemAssigneeNames uses FindByBizKeys and returns map[int64]string
- [x] resolveAssigneeNames (table view) uses FindByBizKeys
- [x] buildProgressRecordVOs uses FindByBizKeys for AuthorKey lookups
- [x] buildItemPoolVOs uses FindByBizKeys for SubmitterKey lookups
- [x] StatusHistory.ItemKey now receives BizKey from all callers (sub_item, main_item, progress)
- [x] All test mocks implement FindByBizKeys
- [x] All tests pass with zero failures

## Notes
无
