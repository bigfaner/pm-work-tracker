---
feature: "e2e-test-scripts-rebuild"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-30"
---

# Test Cases: e2e-test-scripts-rebuild

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 0     |
| CLI  | 6     |
| **Total** | **6** |

---

## UI Test Cases

_None — this feature has no UI surface._

---

## API Test Cases

_None — this feature operates entirely in the test infrastructure layer._

---

## CLI Test Cases

## TC-001: npm test exits 0 after all features graduated

- **Source**: Story 1 / AC-1
- **Type**: CLI
- **Target**: cli/npm-test
- **Test ID**: cli/npm-test/npm-test-exits-0-after-all-features-graduated
- **Pre-conditions**: All 11 features have graduation markers in `tests/e2e/.graduated/`; dev server running on localhost:8080 and localhost:5174
- **Steps**:
  1. Change directory to `tests/e2e/`
  2. Run `npm test`
  3. Capture exit code and stdout
- **Expected**: Exit code is 0; stdout contains per-file pass/skip/fail counts for each spec file; overall pass rate ≥ 80%
- **Priority**: P0

---

## TC-002: Graduate single feature via full pipeline

- **Source**: Story 2 / AC-1
- **Type**: CLI
- **Target**: cli/graduate-tests
- **Test ID**: cli/graduate-tests/graduate-single-feature-via-full-pipeline
- **Pre-conditions**: Feature `api-permission-test-coverage` has `test-cases.md`; not yet graduated (no `.graduated/api-permission-test-coverage` marker)
- **Steps**:
  1. Run `/gen-test-scripts` for `api-permission-test-coverage`
  2. Run `/run-e2e-tests` for the generated scripts
  3. Verify pass rate ≥ 80%
  4. Run `/graduate-tests` for `api-permission-test-coverage`
- **Expected**: `tests/e2e/.graduated/api-permission-test-coverage` marker file exists with ISO 8601 timestamp; graduated spec appears under `tests/e2e/api/api-permission-test-coverage/api.spec.ts`
- **Priority**: P0

---

## TC-003: Graduated spec contains no stale import paths

- **Source**: Story 2 / AC-1 (no stale imports)
- **Type**: CLI
- **Target**: cli/validate-spec
- **Test ID**: cli/validate-spec/graduated-spec-contains-no-stale-import-paths
- **Pre-conditions**: At least one feature has been graduated to `tests/e2e/`
- **Steps**:
  1. Run `validateSpec()` on each graduated spec file in `tests/e2e/`
  2. Check for `STALE_IMPORT_PATH` errors (imports containing `testing/scripts/`)
- **Expected**: `validateSpec()` returns `{ valid: true, errors: [] }` for all graduated specs; no spec imports from `testing/scripts/` paths
- **Priority**: P0

---

## TC-004: Low pass rate triggers KNOWN_FAILURES.md entry

- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/known-failures
- **Test ID**: cli/known-failures/low-pass-rate-triggers-known-failures-entry
- **Pre-conditions**: A feature's generated scripts have pass rate < 80% when run via `/run-e2e-tests`
- **Steps**:
  1. Run `/run-e2e-tests` for a feature with known failing assertions
  2. Observe pass rate < 80%
  3. Record failing assertions in `tests/e2e/KNOWN_FAILURES.md` with feature slug, test ID, description, reason, and owner
  4. Confirm with author
  5. Run `/graduate-tests` for the feature
- **Expected**: `tests/e2e/KNOWN_FAILURES.md` contains an entry for the feature with all required fields (slug, test ID, description, reason, owner, date); graduation proceeds; no undocumented failures enter the regression suite
- **Priority**: P0

---

## TC-005: Spec file conformance — node:test only, no external frameworks

- **Source**: Story 2 / AC-1 + Spec §5.1
- **Type**: CLI
- **Target**: cli/validate-spec
- **Test ID**: cli/validate-spec/spec-file-conformance-node-test-only-no-external-frameworks
- **Pre-conditions**: Scripts have been generated via `/gen-test-scripts` for at least one feature
- **Steps**:
  1. Run `validateSpec()` on each generated spec file in `testing/scripts/`
  2. Check for `EXTERNAL_IMPORT` errors (mocha, jest, chai, jasmine, vitest, describe, it)
  3. Check for `MISSING_TRACEABILITY` errors (test() blocks without `// Traceability: TC-` comment)
- **Expected**: All generated specs return `{ valid: true, errors: [] }`; no external framework imports; every `test()` block has a traceability comment
- **Priority**: P0

---

## TC-006: package.json updated with all graduated spec paths

- **Source**: Story 1 / AC-1 + Spec §5.4
- **Type**: CLI
- **Target**: cli/update-package-json
- **Test ID**: cli/update-package-json/package-json-updated-with-all-graduated-spec-paths
- **Pre-conditions**: All 11 features have been graduated; `updatePackageJson()` has been called
- **Steps**:
  1. Read `tests/e2e/package.json`
  2. Parse `test:api`, `test:ui`, and `test:cli` script strings
  3. Verify each graduated spec path appears in the appropriate script
  4. Verify no path points to a non-existent file
- **Expected**: `test:api` includes all `tests/e2e/api/*/api.spec.ts` paths; `test:ui` exists and includes all `tests/e2e/ui/*/` spec paths; `test:cli` includes all `tests/e2e/cli/*/cli.spec.ts` paths; `npm test` script chains all three; no duplicate paths
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | CLI | cli/npm-test | P0 |
| TC-002 | Story 2 / AC-1 | CLI | cli/graduate-tests | P0 |
| TC-003 | Story 2 / AC-1 | CLI | cli/validate-spec | P0 |
| TC-004 | Story 3 / AC-1 | CLI | cli/known-failures | P0 |
| TC-005 | Story 2 / AC-1 + Spec §5.1 | CLI | cli/validate-spec | P0 |
| TC-006 | Story 1 / AC-1 + Spec §5.4 | CLI | cli/update-package-json | P0 |
