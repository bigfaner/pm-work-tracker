---
status: "completed"
started: "2026-04-27 02:01"
completed: "2026-04-27 02:13"
time_spent: "~12m"
---

# Task Record: fix-e2e-3-1 修复 e2e 测试失败: unknown

## Summary
Fixed all 4 API e2e test failures (TC-001 through TC-004). Root causes: (1) API tests used /api/v1/... URL prefix but backend routes are at /v1/... (no /api prefix), (2) convert-to-main and sub-item-create endpoints require JSON body fields (priority, assigneeKey, startDate, expectedEndDate) that were entirely missing from test requests, (3) convert-to-main returns mainItemBizKey as a raw JSON number exceeding JavaScript safe integer range (int64 snowflake), causing precision loss in JSON.parse — fixed by extracting via regex as a string, (4) tests assumed pre-existing data with sequential integer IDs (1,2,3) but the API uses snowflake BizKeys — tests now create their own test data programmatically via the API, (5) submit-pool returns 201 (not 200) and sub-item-create returns 201 (not 200) — assertions updated to accept both.

## Changes

### Files Created
无

### Files Modified
- docs/features/db-dialect-compat/testing/scripts/helpers.ts
- docs/features/db-dialect-compat/testing/scripts/api.spec.ts

### Key Decisions
- Changed auth URL from /api/v1/auth/login to /v1/auth/login to match actual backend router prefix
- Tests create their own test data via API calls instead of relying on pre-existing data with hardcoded IDs
- Extract mainItemBizKey from raw JSON body via regex to avoid JavaScript int64 precision loss from JSON.parse
- Use assigneeKey '0' instead of empty string to satisfy Gin binding:required validation
- Accept both 200 and 201 status codes for resource-creating endpoints

## Test Results
- **Passed**: 9
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] Root cause of failures identified
- [x] Code or test scripts fixed
- [x] All unit tests pass

## Notes
无
