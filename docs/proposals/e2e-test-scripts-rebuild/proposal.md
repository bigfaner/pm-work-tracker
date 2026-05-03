---
created: 2026-04-30
author: fanhuifeng
status: Draft
---

# Proposal: E2E Test Scripts Rebuild & Graduation

## Problem

11 features have `test-cases.md` but their e2e test scripts remain in `testing/scripts/` (feature workspace) instead of being graduated to `tests/e2e/` (regression suite), and the scripts may not conform to current forge `gen-test-scripts` standards.

### Evidence

Exploration of `docs/features/` reveals:

| Feature | test-cases.md | Scripts in testing/scripts/ | Graduated to tests/e2e/ |
|---|---|---|---|
| api-permission-test-coverage | ✅ | ✅ | ❌ |
| bizkey-unification | ✅ | ✅ | ❌ |
| config-yaml | ✅ | ✅ | ❌ |
| db-dialect-compat | ✅ | ✅ | ❌ |
| improve-ui | ✅ | ✅ | ❌ |
| jlc-schema-alignment | ✅ | ✅ | ❌ |
| rbac-permissions | ✅ | ✅ | ❌ |
| schema-alignment-cleanup | ✅ | ✅ | ❌ |
| soft-delete-consistency | ✅ | ✅ | ❌ |
| status-flow-optimization | ✅ | ✅ | ❌ |
| user-management-reset-delete | ✅ | ✅ | ❌ |

Only `integration-test-coverage` and `pm-work-tracker` have been graduated. Scripts for the remaining 11 features are stranded in feature workspaces and excluded from the regression suite.

### Urgency

7 of the 11 stranded scripts reference resources by internal numeric `.id` — a pattern the `bizkey-unification` feature replaced with `.bizKey` across all API responses. Running these scripts against the current backend fails immediately on any assertion that uses a numeric ID as a URL path segment (e.g., `rbac-permissions` has 10 such references, `soft-delete-consistency` has 19). These scripts are not just unrun — they are broken against the current API and will produce false failures if graduated as-is.

## Proposed Solution

For each of the 11 ungraduated features with `test-cases.md`:

1. **Regenerate scripts** using `/gen-test-scripts` against the current `test-cases.md` and `sitemap.json`, ensuring conformance to forge standards (Playwright for UI, `fetch` for API, `child_process` for CLI; `node:test` + `node:assert`; no external test frameworks).
2. **Graduate scripts** using `/graduate-tests` to move them into `tests/e2e/<target>/` with proper imports and graduation markers.

After graduation, `npm test` run from `tests/e2e/` executes the full regression suite. The 11 features are expected to contribute approximately 11–22 additional spec files (each feature has at least `api.spec.ts`; features with CLI coverage add `cli.spec.ts`). The `tests/e2e/package.json` test scripts will be updated to include each graduated spec. A clean run exits with code 0 and prints a per-file summary of passed/skipped/failed test counts.

Features without `test-cases.md` are skipped.

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Do nothing | No effort; zero risk of introducing broken tests into the regression suite | 11 features unprotected by regression; 7 scripts already broken against current API | Rejected: coverage gap grows and stale scripts become harder to fix over time |
| Graduate existing scripts as-is (no regeneration) | Faster (~1 hr total) | 7 scripts use stale `.id` pattern and will fail immediately; others may have stale locators | Rejected: broken scripts in regression suite produce permanent noise |
| Regenerate + graduate (proposed) | Conformant scripts, current API patterns, full regression coverage | ~2 hrs per feature × 11 features ≈ 22 hrs total effort | Accepted: ROI is clear — 22 hrs of rebuild vs. ongoing manual verification of 11 unprotected feature areas |

## Scope

### In Scope

- Regenerate e2e test scripts for all 11 features that have `test-cases.md` but are not yet graduated
- Graduate regenerated scripts to `tests/e2e/` with correct target subdirectory structure
- Create graduation markers in `tests/e2e/.graduated/<slug>`
- Update `sitemap.json` if stale before script generation
- Update `tests/e2e/package.json` test scripts to include each graduated spec file

### Out of Scope

- Features without `test-cases.md` (code-conventions, code-quality-cleanup, decision-log, fix-id-bizKey-mismatch, item-code-redesign, permission-granularity, single-binary-deploy, ui-ux-unification, weekly-stats-optimization)
- Writing new test cases (test-cases.md is treated as source of truth)
- CI pipeline integration
- Already-graduated features (integration-test-coverage, pm-work-tracker)

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| sitemap.json is stale; generated UI locators don't match current DOM | Medium (likely for features >1 month old) | Scripts fail on first run | Run `/gen-sitemap` before script generation |
| Some test cases reference APIs or UI flows that no longer exist | Medium (confirmed: 7/11 scripts already use stale `.id` pattern) | Generated scripts have dead assertions | Before graduating each feature, run `/run-e2e-tests` on its scripts; require ≥80% pass rate. Any failing assertion must either be fixed or documented in a `KNOWN_FAILURES.md` entry with the reason. The author is responsible; no graduation proceeds with unresolved failures and no documented exception. |
| Duplicate coverage with already-graduated tests causes noise | Low | Redundant test failures | `/graduate-tests` agent handles deduplication via split/merge logic |
| Script generation produces non-runnable output for complex features | Low | Graduation blocked | Run `/run-e2e-tests` on each batch before graduating |

## Success Criteria

- [ ] All 11 features have scripts regenerated via `/gen-test-scripts` using `node:test` + `node:assert`, Playwright for UI, `fetch` for API, `child_process` for CLI, and no imports from external test frameworks
- [ ] All 11 features have graduation markers in `tests/e2e/.graduated/<slug>`
- [ ] Graduated scripts are organized under `tests/e2e/<target>/` (api/, ui/, cli/)
- [ ] `tests/e2e/` contains no scripts that import from `testing/scripts/` paths
- [ ] `npm test` in `tests/e2e/` exits with code 0 with ≥80% of graduated scripts passing against the current dev environment; any failures are documented in `KNOWN_FAILURES.md`
- [ ] All 11 features' test-cases.md entries are represented by at least one graduated spec file

## Next Steps

- Proceed to `/gen-test-scripts` + `/graduate-tests` for each feature in batches
