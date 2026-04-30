---
status: "completed"
started: "2026-04-30 18:01"
completed: "2026-04-30 18:03"
time_spent: "~2m"
---

# Task Record: 2.summary Phase 2 Summary

## Summary
## Tasks Completed
- 2.1: API-only features (api-permission-test-coverage, soft-delete-consistency) graduated to tests/e2e/api/<slug>/api.spec.ts; KNOWN_FAILURES.md created
- 2.2: API+CLI features (bizkey-unification, config-yaml, db-dialect-compat) graduated to tests/e2e/api/ and tests/e2e/cli/; fixed false-positive MISSING_TRACEABILITY in db-dialect-compat cli spec
- 2.3: API+UI features (improve-ui, schema-alignment-cleanup, status-flow-optimization, user-management-reset-delete) graduated to tests/e2e/api/ and tests/e2e/ui/; package.json test scripts updated
- 2.4: API+UI+CLI features (jlc-schema-alignment, rbac-permissions) graduated to tests/e2e/api/, tests/e2e/ui/, tests/e2e/cli/; fixed TEST_CALL_RE regex in validate-spec.ts to exclude method calls

## Key Decisions
- 2.1: Fixed MISSING_TRACEABILITY by moving traceability comment to within 3 lines of test() call in soft-delete spec
- 2.1: Import paths rewritten from ./helpers.js to ../../helpers.js for tests/e2e/api/<slug>/ depth
- 2.1: Backend server not running at graduation time — failures documented in KNOWN_FAILURES.md per task AC
- 2.2: Fixed false-positive in validateSpec: comment containing test( triggered MISSING_TRACEABILITY; renamed comment to avoid regex match
- 2.2: config-yaml cli helpers copied verbatim (feature-specific exports) rather than re-exporting from main helpers
- 2.2: db-dialect-compat graduation marker overwritten with new 2026-04-30 timestamp as required by task spec
- 2.3: Added traceability comments before all test() calls in weekly-view.spec.ts including regex .test() method calls that triggered false-positive MISSING_TRACEABILITY errors
- 2.3: Copied feature helpers.ts to both api/ and ui/ graduated directories since each spec imports from ./helpers.js
- 2.4: Fixed TEST_CALL_RE in validate-spec.ts to exclude '.' as preceding char, preventing false positives on regex .test() method calls
- 2.4: Added test:cli:features script to package.json for new CLI specs; existing test:cli unchanged to avoid breaking current CI

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|
| (none — Phase 2 is execution-only) | — | — |

## Conventions Established
- All graduated specs live under tests/e2e/<type>/<slug>/<type>.spec.ts with a co-located helpers.ts
- Import paths in graduated specs use ../../helpers.js (relative to tests/e2e/<type>/<slug>/)
- Graduation markers written to tests/e2e/.graduated/<slug>
- All test failures at graduation time (server unavailable) documented in tests/e2e/KNOWN_FAILURES.md with feature slug, test ID, reason, and owner
- validateSpec() must pass before graduation; traceability comments must appear within 3 lines of test() call
- TEST_CALL_RE in validate-spec.ts excludes method calls (preceded by '.') to avoid false positives on regex .test() usage

## Deviations from Design
- All 11 graduated features have 100% of tests in KNOWN_FAILURES.md due to backend/frontend servers being offline at graduation time; no tests were actually executed against a live stack
- TC-003/004/017/019/021 in soft-delete-consistency have partial coverage only — no public API for direct sub-item soft-delete
- api-permission-test-coverage TC-011/TC-012 have a hardcoded backend path that may need updating in production use
- jlc-schema-alignment TC-008 to TC-010 require MySQL 8.0 which was unavailable at graduation time

## Changes

### Files Created
无

### Files Modified
无

### Key Decisions
- All 11 features graduated: api-permission-test-coverage, soft-delete-consistency, bizkey-unification, config-yaml, db-dialect-compat, improve-ui, schema-alignment-cleanup, status-flow-optimization, user-management-reset-delete, jlc-schema-alignment, rbac-permissions
- All failures documented in KNOWN_FAILURES.md — backend/frontend offline at graduation time
- validate-spec.ts TEST_CALL_RE fixed to exclude .test() method calls (false-positive fix)

## Test Results
- **Passed**: 0
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] All phase task records read and analyzed
- [x] Summary follows the exact 5-section template
- [x] Graduated feature count and any blocked features documented

## Notes
无
