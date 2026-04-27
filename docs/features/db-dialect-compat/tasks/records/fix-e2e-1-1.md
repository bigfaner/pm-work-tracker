---
status: "completed"
started: "2026-04-27 01:41"
completed: "2026-04-27 01:48"
time_spent: "~7m"
---

# Task Record: fix-e2e-1-1 修复 e2e 测试失败: unknown

## Summary
Fixed 4 integration test failures (TestProgress_AppendToSubItem1_UpdatesMainItemCompletion, TestProgress_AppendToSubItem2_UpdatesMainItemCompletion, TestProgress_RegressionBlocked_Returns422, TestItemPool_Assign_Success) by setting snowflake-generated BizKey on all seeded test models (MainItem, SubItem, ItemPool, ProgressRecord) and using BizKey values in URL paths and request bodies instead of internal auto-increment IDs.

## Changes

### Files Created
无

### Files Modified
- backend/tests/integration/progress_completion_test.go

### Key Decisions
- Seed functions (seedProgressData, seedPoolData, seedReportData) now set BizKey via snowflake.Generate() on all created models, matching production behavior where the API layer resolves path parameters as BizKey via FindByBizKey
- appendProgress helper changed from accepting subID (uint) to subBizKey (int64) since the handler resolves the URL :subId param as a bizKey
- seedPoolData returns poolBizKey and mainItemBizKey in addition to internal IDs, since the assign request body uses mainItemKey as a bizKey resolved via FindByBizKey

## Test Results
- **Passed**: 30
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] Root cause of failures identified (seeded records had BizKey=0, handlers resolve URL params as BizKey via FindByBizKey)
- [x] Code/test scripts fixed (all seed functions now set BizKey, URLs use BizKey values)
- [x] All unit tests pass (30/30 integration tests pass)

## Notes
无
