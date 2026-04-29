---
status: "completed"
started: "2026-04-28 19:26"
completed: "2026-04-28 19:56"
time_spent: "~30m"
---

# Task Record: 2.gate Phase 2 Exit Gate: Full Build + Test Pass + Zero Casts

## Summary
Phase 2 exit gate: migrated all repo interfaces (MainItemRepo, SubItemRepo, ProgressRepo, ItemPoolRepo) from uint teamID to int64 teamBizKey, removed all uint(teamBizKey) casts in service layer, updated all test mocks and gorm integration test seeds to use BizKey. All 7 checklist items pass: build clean, GetTeamID has zero callers, zero forbidden casts, progress_service uses teamBizKey directly, all tests pass, coverage 85.5%, isPMRole and UpdateMemberRole accept int64.

## Changes

### Files Created
无

### Files Modified
- backend/internal/repository/main_item_repo.go
- backend/internal/repository/sub_item_repo.go
- backend/internal/repository/progress_repo.go
- backend/internal/repository/item_pool_repo.go
- backend/internal/repository/gorm/main_item_repo.go
- backend/internal/repository/gorm/sub_item_repo.go
- backend/internal/repository/gorm/progress_repo.go
- backend/internal/repository/gorm/item_pool_repo.go
- backend/internal/service/main_item_service.go
- backend/internal/service/sub_item_service.go
- backend/internal/service/progress_service.go
- backend/internal/service/item_pool_service.go
- backend/internal/service/view_service.go
- backend/internal/service/report_service.go
- backend/internal/service/team_service.go
- backend/internal/service/main_item_service_test.go
- backend/internal/service/sub_item_service_test.go
- backend/internal/service/progress_service_test.go
- backend/internal/service/item_pool_service_test.go
- backend/internal/service/view_service_test.go
- backend/internal/handler/router_test_stubs.go
- backend/internal/handler/item_pool_handler_test.go
- backend/internal/handler/main_item_handler_test.go
- backend/internal/repository/gorm/main_item_repo_test.go
- backend/internal/repository/gorm/sub_item_repo_test.go
- backend/internal/repository/gorm/progress_repo_test.go
- backend/internal/repository/gorm/item_pool_repo_test.go

### Key Decisions
- Repo interfaces migrated from uint teamID to int64 teamBizKey to eliminate all uint(teamBizKey) casts at service call sites
- team_service.go CountByTeam call updated to use team.BizKey instead of team.ID
- Gorm integration test seeds updated to set team.BizKey = int64(team.ID) after creation since no BeforeCreate hook auto-sets it
- createMainItem/createSubItem/createItemPool/createProgressRecord helpers updated to accept int64 teamBizKey

## Test Results
- **Passed**: 713
- **Failed**: 0
- **Coverage**: 85.5%

## Acceptance Criteria
- [x] go build ./... passes with zero errors
- [x] grep -rn GetTeamID backend/internal/ returns zero caller results
- [x] grep for uint(.*bizKey) and int64(teamID) in service/handler returns zero results
- [x] progress_service.go TeamKey field assigned from teamBizKey (not int64(teamID))
- [x] go test -race ./internal/... passes with zero failures
- [x] Coverage >= 75% on internal/service and internal/handler
- [x] team_service.go isPMRole accepts int64; UpdateMemberRole last param is int64

## Notes
无
