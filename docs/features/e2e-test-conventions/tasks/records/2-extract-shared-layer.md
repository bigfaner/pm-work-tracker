---
status: "completed"
started: "2026-06-03 20:44"
completed: "2026-06-03 20:57"
time_spent: "~13m"
---

# Task Record: 2 提取 shared 层到 tests/shared/

## Summary
Extracted shared layer to tests/shared/: moved pure Node.js helpers to tests/shared/helpers.ts (no Playwright imports), copied config.yaml to tests/shared/config.yaml, added re-export shim in tests/e2e/helpers.ts to maintain backward compatibility. Also fixed 3 pre-existing Go lint issues (missing package comments) to pass quality gate.

## Changes

### Files Created
- tests/shared/helpers.ts
- tests/shared/config.yaml
- tests/shared/package.json
- tests/shared/package-lock.json

### Files Modified
- tests/e2e/helpers.ts
- tests/e2e/tsconfig.json
- tests/e2e/package.json
- tests/e2e/package-lock.json
- backend/internal/repository/gorm/filter_helpers.go
- backend/internal/repository/item_pool_repo.go
- backend/internal/vo/item_pool_vo.go

### Key Decisions
- config.yaml is a data file read at runtime, not imported via JS module system, so no JS re-export shim needed for it - the file remains in both locations with identical content
- Playwright-specific functions (login, screenshot, browserLogin, navTo, etc.) stayed in tests/e2e/helpers.ts as they depend on Page/Locator types
- shared/helpers.ts already uses findConfigPath() that searches for tests/shared/config.yaml, so no code changes needed for config discovery

## Test Results
- **Tests Executed**: Yes
- **Passed**: 725
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] tests/shared/helpers.ts created with no Playwright imports (grep '@playwright/test' tests/shared/ returns empty)
- [x] tests/shared/config.yaml moved from tests/e2e/config.yaml
- [x] Old paths tests/e2e/helpers.ts and tests/e2e/config.yaml preserved with re-export shim / identical content

## Notes
Work was partially completed by a previous executor. Added package comments to 3 Go files to resolve pre-existing lint errors required by quality gate. TypeScript compilation errors are pre-existing (missing @types/node) and not introduced by this refactoring.
