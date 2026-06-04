---
status: "completed"
started: "2026-06-03 22:24"
completed: "2026-06-03 22:29"
time_spent: "~5m"
---

# Task Record: 6 清理迁移残留 + CI 适配

## Summary
Cleaned up migration residuals: deleted re-export shim files (helpers.ts, config.yaml) and entire tests/e2e/ directory, moved .graduated/ and KNOWN_FAILURES.md to tests/ root, updated tests/infra/e2e-rebuild.spec.ts to reference new surface-based directories, updated convention docs to reflect new paths, and added surface-based CI jobs (e2e-api, e2e-web, e2e-infra) to .github/workflows/ci.yml.

## Changes

### Files Created
- tests/.graduated/
- tests/KNOWN_FAILURES.md

### Files Modified
- tests/infra/e2e-rebuild.spec.ts
- docs/conventions/testing/index.md
- docs/conventions/testing/web/core.md
- .github/workflows/ci.yml

### Key Decisions
- Moved .graduated/ and KNOWN_FAILURES.md to tests/ root instead of deleting them, as they are project artifacts spanning all surfaces
- Rewrote e2e-rebuild.spec.ts to validate new surface-based structure (vitest for api/infra, playwright for web) instead of old monolithic tests/e2e/ directory
- CI jobs: e2e-api starts backend + runs vitest, e2e-web starts full stack + runs playwright, e2e-infra runs vitest with no services needed

## Test Results
- **Tests Executed**: Yes
- **Passed**: 6
- **Failed**: 0
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] Re-export shim files (tests/e2e/helpers.ts, tests/e2e/config.yaml) deleted
- [x] tests/e2e/ directory deleted (confirmed empty then removed)
- [x] CI has independent api/web/infra jobs, old unified e2e job removed

## Notes
19 pre-existing infra test failures due to missing server binaries and backend services — not caused by this cleanup. All 6 e2e-rebuild tests pass after rewrite.
