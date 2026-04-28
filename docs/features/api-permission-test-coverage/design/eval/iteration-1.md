---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/design/"
iteration: "1"
target_score: "80"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 93/100** (target: 80)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ✅          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   6/7    │          │            │
│    Dependencies listed       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  17      │  20      │ ⚠️          │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   6/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅          │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   5/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  19      │  20      │ ✅          │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   6/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  10      │  10      │ ✅          │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  93      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 19/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Interfaces § `seedPermMatrixFixtures` | `data *seedData` parameter type never defined in doc | -1 (Interface signatures) |
| Data Models § existing models | `model.Role { ID, BizKey, Name, IsPreset, ... }` — `...` truncation leaves fields unspecified | -1 (Models concrete) |
| Interfaces § `seedPermMatrixFixtures` | `seedData` undefined + 5 I-A endpoints unnamed → developer must guess inputs | -1 (Directly implementable) |
| Component diagram | Diagram shows test file boxes but draws no arrows to production components under test | -1 (Component diagram) |
| Dependencies table | `gorm.io/gorm` used in `seedPermMatrixFixtures(db *gorm.DB)` signature but absent from dependency table | -1 (Dependencies listed) |
| Testing Strategy § I-A | "5 representative endpoints" never enumerated by name or path | -1 (Tasks derivable) |

---

## Attack Points

### Attack 1: Interface & Model — `seedData` type is a phantom

**Where**: `func seedPermMatrixFixtures(t *testing.T, db *gorm.DB, data *seedData) permMatrixFixtures`

**Why it's weak**: `*seedData` appears in the function signature but is never defined anywhere in the document. A developer implementing this helper has no idea what fields `seedData` carries — is it the existing `seedData` struct from `helpers.go`, or a new type? The doc references `helpers.go` only in the Appendix, not as a definition source. This forces the implementer to go spelunking in existing test code before writing a single line.

**What must improve**: Either define `seedData` inline (field list + types) or explicitly state "reuses `seedData` defined in `helpers.go` at line N, fields X/Y/Z are used." The current silence is a concrete implementation blocker.

---

### Attack 2: Breakdown-Readiness — "5 representative endpoints" in I-A are unnamed

**Where**: `| Full stack (preset roles) | Integration | real SQLite, httptest | superadmin/pm → 200, member → 403 for 5 representative endpoints | 3 roles × 5 endpoints = 15 assertions |`

**Why it's weak**: The I-A test matrix is the most complex integration test in this feature, yet the 5 endpoints are never named. The `permMatrixFixtures` struct returns only 3 bizKeys (`mainItemBizKey`, `progressBizKey`, `poolItemBizKey`), which implies 3 of the 5 endpoints need fixture records — but the other 2 are unspecified. A task breakdown cannot assign "implement I-A" without knowing which 5 paths to hit. The example in Key Test Scenarios only shows one (`POST /main-items/:id/archive`).

**What must improve**: List all 5 endpoint paths with HTTP method, required permission code, and whether a fixture bizKey is needed. This is a 5-row table, not a paragraph.

---

### Attack 3: Architecture — component diagram has no edges to production code

**Where**: The three ASCII boxes in the Component Diagram section show only test file internals. No arrows connect them to the production components they exercise (`middleware/permission.go`, `handler/*.go`, `pkg/permissions/codes.go`).

**Why it's weak**: The diagram answers "what test files exist" but not "what production surface each test file covers." For a feature whose entire purpose is test coverage, the diagram should make the coverage relationship explicit. A reader cannot tell from the diagram alone that `permission_matrix_test.go` exercises `RequirePermission` or that `permission_coverage_test.go` reads `codes.go`.

**What must improve**: Add edges (or a second diagram) showing `test file → production component` relationships. Even a simple table mapping each test file to the production file(s) it exercises would close this gap.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 93/100
- **Target**: 80/100
- **Gap**: 0 (target exceeded by 13 points)
- **Breakdown-Readiness**: 19/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. Proceed to task breakdown.
