---
created: 2026-04-30
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: E2E Test Scripts Rebuild & Graduation

## Overview

This is a tooling workflow feature. There is no new application code, backend model, or API surface. The work consists of running forge skills (`/gen-test-scripts`, `/run-e2e-tests`, `/graduate-tests`) for each of 11 features in sequence, then updating `tests/e2e/package.json` to include all graduated specs.

The execution unit is one feature at a time. Each feature goes through: generate → validate → graduate. Features are independent; failure in one does not block others.

## Architecture

### Layer Placement

This feature operates entirely in the test infrastructure layer:

```
docs/features/<slug>/testing/scripts/   ← source (generated, per-feature)
tests/e2e/                              ← target (regression suite, shared)
tests/e2e/.graduated/                   ← graduation markers
tests/e2e/KNOWN_FAILURES.md             ← failure registry
```

No application layers (handler/service/repository/model) are touched.

### Component Diagram

Shows components, data stores, and directed read/write edges.

```
                    ┌──────────────────────────────────┐
                    │           Orchestrator            │
                    │     (executeFeature per slug)     │
                    └──┬──────────────┬──────────┬─────┘
                       │ invoke       │ invoke   │ invoke
          ┌────────────┘              │          └──────────────────┐
          ▼                           ▼                             ▼
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ /gen-test-       │       │ /run-e2e-tests   │       │ /graduate-tests  │
│ scripts          │       │                  │       │                  │
│ [forge skill]    │       │ [forge skill]    │       │ [forge skill]    │
└──────────────────┘       └──────────────────┘       └──────────────────┘
  reads:                     reads:                     reads:
  · test-cases.md            · testing/scripts/         · testing/scripts/
  · sitemap.json               *.spec.ts                  *.spec.ts
  writes:                    returns:                   writes:
  · testing/scripts/         · pass rate → Orchestrator · tests/e2e/<type>/
    *.spec.ts                                              <slug>/*.spec.ts
                                                         · .graduated/<slug>

Orchestrator writes:
  · KNOWN_FAILURES.md        (on ERR_PASS_RATE_LOW, after author confirmation)
  · tests/e2e/package.json   (after all features complete, via updatePackageJson)
```

The pipeline flow (what comes after what) is shown in the Appendix.

### Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `node:test` | Node.js built-in | Test runner (no install needed) |
| `node:assert` | Node.js built-in | Assertions |
| `playwright` | ^1.44.0 (already in tests/e2e/package.json) | UI test execution |
| `tsx` | ^4.19.0 (already in tests/e2e/package.json) | TypeScript ESM loader |
| `yaml` | ^2.7.0 (already in tests/e2e/package.json) | Config parsing |

No new dependencies required.

## Interfaces

### Interface 1: Per-Feature Execution Contract

The orchestrator calls `executeFeature` once per slug. This is the only typed callable in the workflow.

```typescript
executeFeature(slug: string): FeatureResult

interface FeatureResult {
  slug: string
  status: "completed" | "blocked" | "skipped"
  specFiles: string[]          // paths relative to tests/e2e/
  passRate: number             // 0.0–1.0; -1 if not run
  knownFailures: KnownFailure[]
  graduationMarker: string     // ISO 8601 timestamp; empty if not graduated
}

interface KnownFailure {
  feature: string
  testId: string               // e.g. "TC-003"
  description: string
  reason: string
  owner: string                // GitHub handle
}
```

Implementor: the per-feature loop script at `docs/features/e2e-test-scripts-rebuild/testing/scripts/run-all.ts`. It shells out to the three forge skills in sequence.

### Interface 2: Spec File Validator

Called by the orchestrator before invoking `/graduate-tests`. Rejects specs that import external frameworks or lack traceability comments.

```typescript
validateSpec(filePath: string): ValidationResult

interface ValidationResult {
  valid: boolean
  errors: ValidationError[]   // empty when valid === true
}

interface ValidationError {
  filePath: string
  line: number
  code: "EXTERNAL_IMPORT" | "MISSING_TRACEABILITY" | "STALE_IMPORT_PATH"
  message: string
}
```

Implementor: `testing/scripts/validate-spec.ts`. Uses Node.js `fs.readFileSync` + regex scan (no AST library required for these three checks).

### Interface 3: package.json Updater

Called once after all features complete. Reads current `tests/e2e/package.json`, merges new spec paths into the `test:api`, `test:ui`, and `test:cli` script strings, and writes the file back.

```typescript
updatePackageJson(specPaths: SpecPaths): void

interface SpecPaths {
  api: string[]   // e.g. ["api/bizkey-unification/api.spec.ts"]
  ui:  string[]
  cli: string[]
}
```

Implementor: `testing/scripts/update-package-json.ts`. Throws `Error` with code `ERR_PACKAGE_JSON_WRITE` on write failure.

## File Format Conventions

### Convention 1: Graduated Spec File

Every spec file placed in `tests/e2e/` must use only these imports:

```typescript
import { test, before, after } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { parse } from 'yaml'
// Playwright (UI specs only):
import { chromium } from 'playwright'
// Shared helpers (path relative to spec location):
import { ... } from '../../helpers.js'
```

Each `test()` block must include a traceability comment:
```typescript
// Traceability: TC-NNN → Story X / AC-Y
```

`validateSpec()` enforces both rules before graduation.

### Convention 2: KNOWN_FAILURES.md Entry Format

```markdown
## <feature-slug>

| Test ID | Description | Reason | Owner | Date |
|---------|-------------|--------|-------|------|
| TC-NNN  | <what fails> | <why it fails> | <github handle> | YYYY-MM-DD |
```

## Data Models

### Model 1: Graduation Marker File

```
File path: tests/e2e/.graduated/<slug>
Content:   ISO 8601 timestamp string, e.g. "2026-04-30T10:00:00+08:00"
Encoding:  UTF-8, no trailing newline required
```

### Model 2: Feature Classification

Determines which subdirectory a spec lands in after graduation:

```
FeatureClassification = {
  slug: string
  hasApi: boolean    // api.spec.ts exists in testing/scripts/
  hasUi: boolean     // ui.spec.ts or weekly-view.spec.ts exists
  hasCli: boolean    // cli.spec.ts exists
}
```

Target directories:
- `hasApi` → `tests/e2e/api/<slug>/api.spec.ts`
- `hasUi` → `tests/e2e/ui/<slug>/ui.spec.ts`
- `hasCli` → `tests/e2e/cli/<slug>/cli.spec.ts`

### Model 3: Feature Inventory (11 features)

| Slug | API | UI | CLI | Notes |
|------|-----|----|-----|-------|
| api-permission-test-coverage | ✅ | ❌ | ❌ | |
| bizkey-unification | ✅ | ❌ | ✅ | |
| config-yaml | ✅ | ❌ | ✅ | |
| db-dialect-compat | ✅ | ❌ | ✅ | already has graduation marker — re-graduate |
| improve-ui | ✅ | ✅ | ❌ | has weekly-view.spec.ts (treat as UI) |
| jlc-schema-alignment | ✅ | ✅ | ✅ | |
| rbac-permissions | ✅ | ✅ | ✅ | |
| schema-alignment-cleanup | ✅ | ✅ | ❌ | |
| soft-delete-consistency | ✅ | ❌ | ❌ | no package.json in scripts dir |
| status-flow-optimization | ✅ | ✅ | ❌ | |
| user-management-reset-delete | ✅ | ✅ | ❌ | |

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | Action |
|------------|------|-------------|--------|
| ERR_NO_TEST_CASES | MissingTestCases | test-cases.md not found for feature | Skip feature, log |
| ERR_GEN_FAILED | ScriptGenFailed | /gen-test-scripts exits non-zero | Block graduation, report to maintainer |
| ERR_PASS_RATE_LOW | PassRateBelowThreshold | pass rate < 80% after /run-e2e-tests | Record to KNOWN_FAILURES.md, await author confirmation |
| ERR_NO_MARKER | GraduationMarkerMissing | .graduated/<slug> not created by /graduate-tests | Retry once, then report |
| ERR_STALE_IMPORT | StaleImportPath | graduated spec imports from testing/scripts/ | Block graduation, fix import path |

### Propagation Strategy

Errors are non-fatal per feature. A blocked feature is logged and skipped; the loop continues to the next feature. Final summary reports all blocked features for manual follow-up.

## Cross-Layer Data Map

Single-layer feature (test infrastructure only) — Cross-Layer Data Map not applicable.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Script conformance | Static check | grep / AST scan | No external framework imports; traceability comments present | 100% of graduated files |
| Script execution | Integration | node:test via /run-e2e-tests | Each spec passes ≥80% of its test cases | ≥80% per feature |
| Graduation marker | File check | ls / read | .graduated/<slug> exists with valid ISO 8601 timestamp | 100% of graduated features |
| package.json | Smoke test | npm test (dry run) | All spec paths in scripts resolve to existing files | 100% |

### Key Test Scenarios

1. **Happy path**: feature has test-cases.md, scripts generate cleanly, all tests pass, graduation marker created
2. **Stale sitemap**: UI feature with outdated sitemap.json — sitemap regenerated before script generation
3. **Low pass rate**: feature scripts fail >20% — failures recorded in KNOWN_FAILURES.md, author confirms, graduation proceeds
4. **Stale import path**: generated script imports from testing/scripts/ — caught before graduation, path rewritten to ../../helpers.js
5. **db-dialect-compat re-graduation**: already has .graduated marker — overwrite marker with new timestamp

### Overall Coverage Target

≥80% of graduated spec test cases pass against current dev environment.

## Security Considerations

### Threat Model

| Threat | Concrete Risk | Countermeasure |
|--------|--------------|----------------|
| `config.yaml` committed to git | Test credentials (username/password) leak into repository history | `config.yaml` is listed in `.gitignore`; verified with `git check-ignore -v config.yaml` before any commit. CI reads credentials from environment variables injected by the secrets manager, never from a committed file. |
| Spec injection | A generated spec file contains a shell command (e.g. via a crafted test-cases.md title) that executes when `node:test` loads the file | `validateSpec()` rejects any spec whose source contains backtick template literals with `$()` or `process.exec` patterns before graduation. Generated specs are reviewed by the maintainer before the `/graduate-tests` call. |
| CI secret exposure | Test account credentials passed as env vars appear in CI build logs | CI pipeline uses masked secrets (`TEST_USER`, `TEST_PASSWORD`). `config.yaml` is written from env vars at job start and deleted at job end via a `post` step. |
| Overly broad test account permissions | A test account with admin rights leaks data or mutates production-adjacent state | Test accounts are provisioned with the minimum role required per feature (reader for read-only specs, member for write specs). Credentials are rotated after each release cycle by the team lead. |

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| 11 features graduate to tests/e2e/ | Per-Feature Execution Contract | executeFeature() → FeatureResult |
| Scripts use node:test + node:assert only | Graduated Spec File Convention | Convention 1 import rules + validateSpec() |
| Graduation markers created | Graduation Marker File model | Model 1 |
| ≥80% pass rate before graduation | Per-Feature Execution Contract | ERR_PASS_RATE_LOW error handling |
| KNOWN_FAILURES.md for unresolved failures | KNOWN_FAILURES.md Entry Format | Convention 2 |
| package.json updated with new specs | package.json Updater | updatePackageJson() → Interface 3 |
| sitemap.json updated for UI features | Per-Feature Execution Contract | executeFeature() pre-step |
| Story 1: developer runs npm test | package.json Updater | Interface 3 |
| Story 2: maintainer graduates single feature | Per-Feature Execution Contract | executeFeature() |
| Story 3: maintainer handles low pass rate | KNOWN_FAILURES.md + ERR_PASS_RATE_LOW | Convention 2 + Error Handling |

## Open Questions

- [x] Does `db-dialect-compat` need re-graduation? It has a marker but scripts were not regenerated. → Yes, regenerate and overwrite marker.
- [x] Does `improve-ui`'s `weekly-view.spec.ts` graduate as a separate file or merged into `ui.spec.ts`? → Graduate as separate file under `tests/e2e/ui/improve-ui/`.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Graduate existing scripts as-is | Faster | 7 scripts use stale .id pattern; fail immediately | Broken scripts pollute regression suite |
| Batch all 11 features in one /graduate-tests call | Single operation | No per-feature pass rate gate; failures hard to isolate | Quality gate requires per-feature validation |

### References

- `tests/e2e/helpers.ts` — shared test utilities
- `tests/e2e/package.json` — current test scripts
- `docs/sitemap/sitemap.json` — element IDs for Playwright locators
- `tests/e2e/api/main-items/api.spec.ts` — reference graduated spec pattern
