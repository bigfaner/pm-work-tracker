---
status: "completed"
started: "2026-06-03 21:50"
completed: "2026-06-03 21:56"
time_spent: "~6m"
---

# Task Record: 3 迁移 API 测试到 tests/api/ + Vitest

## Summary
Migrated API tests from tests/e2e/ to tests/api/ with Vitest runner. Fixed vitest.config.ts sequence setting to match proposal spec (sequential: true) and fixed setupRbacFixtures username collision by adding random suffix to runId. All 26 test files pass (207 tests).

## Changes

### Files Created
无

### Files Modified
- tests/api/vitest.config.ts
- tests/shared/helpers.ts

### Key Decisions
- Changed vitest.config.ts sequence from { concurrent: false } to { sequential: true } to match proposal spec exactly
- Fixed setupRbacFixtures runId from Date.now() to Date.now() + random suffix to prevent UNIQUE constraint failures across sequential test files running in the same millisecond

## Test Results
- **Tests Executed**: Yes
- **Passed**: 207
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] tests/api/ directory structure with vitest.config.ts (testTimeout 30s, hookTimeout 60s, sequential) and package.json
- [x] API test files migrated to tests/api/<journey>/, all imports from vitest
- [x] request.newContext files rewritten to curl(), shared helpers via ../shared/helpers
- [x] .serial tests migrated to describe.sequential(), failure propagation semantics checked
- [x] npx vitest run from tests/api/ passes all migrated tests

## Notes
The previous session had already performed the bulk migration (directory creation, file moves, import changes). This session verified all AC items and fixed 2 issues: (1) vitest.config.ts sequence config mismatch with proposal spec, (2) setupRbacFixtures username collision causing intermittent test failures. No API test files used .serial (those were web-only files per audit manifest).
