---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/design/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 91/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  17      │  20      │ ⚠️          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   6/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  18      │  20      │ ⚠️          │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   5/7    │          │            │
│    Directly implementable    │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  13      │  15      │ ⚠️          │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   3/5    │          │            │
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
│ 6. Security Considerations   │   9      │  10      │ ✅          │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  91      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 19/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Error Handling — propagation | "translated by the existing handler error-mapping helper (`pkg/apperrors` → HTTP status)" — the helper function is never named; a developer cannot locate it without grepping | -2 pts |
| Data Models section | "No model changes. `ProgressRecord.TeamKey` is already `int64`" — the struct is never shown; the `int64` claim is unverifiable from the design alone | -2 pts |
| Architecture — dependencies | "No new dependencies. No package additions." — existing packages in use (`pkg/apperrors`, `pkg.ParseID`, `gorm`, `gin`) are named only in passing throughout the doc, never consolidated in a dependencies list | -2 pts |
| Architecture — component diagram | The ASCII diagram is a linear vertical stack only; no horizontal relationships between services are shown; the diagram conveys layer order but not component topology | -1 pt |
| Breakdown-Readiness — tasks derivable | "7 service test files" and "10 handler test files" are referenced in the test plan but never enumerated by name; a task breakdown cannot derive specific test-update tasks without independent discovery | -1 pt |
| Security — Threat 3 mitigation | "This is a deployment/configuration concern, not a new attack surface" — still not a concrete countermeasure; no deployment guard or startup assertion is described | -1 pt |

---

## Attack Points

### Attack 1: Error Handling — propagation helper unnamed

**Where**: "These errors are translated by the existing handler error-mapping helper (`pkg/apperrors` → HTTP status). This refactor does not change that mapping."

**Why it's weak**: The propagation path from service error → HTTP response is described at a high level, but the mechanism is anonymous. "Existing handler error-mapping helper" could be a middleware, a per-handler switch, a shared utility function, or a Gin error handler — the design does not say. A developer implementing a new handler in this codebase cannot determine where to plug in without grepping. The rubric requires a "stated strategy for how errors flow between layers" — naming the layer boundary is not the same as naming the mechanism. The iteration-2 attack flagged the service→HTTP gap; the table was added but the translation mechanism remains a black box.

**What must improve**: Name the function or middleware responsible for translating `apperrors` errors to HTTP responses (e.g., `apperrors.HandleError(c, err)` or `middleware.ErrorHandler`). One line is sufficient — but it must be explicit enough that a developer knows where to look.

---

### Attack 2: Interface & Model — ProgressRecord struct not shown

**Where**: "No model changes. `ProgressRecord.TeamKey` is already `int64` — the bug was in the assignment, not the field type."

**Why it's weak**: The entire correctness argument for bug fix #1 rests on the claim that `ProgressRecord.TeamKey` is `int64`. The design never shows the struct. A reviewer auditing this document cannot confirm the field type, its GORM tag, or whether any other fields in the struct are affected — they must open `model/progress_record.go`. This was flagged in iteration 2 and remains unfixed. The rubric requires "all model fields named with types and constraints" — a one-line assertion is not a model definition.

**What must improve**: Show the relevant portion of the `ProgressRecord` struct — at minimum the `TeamKey` field with its type and GORM tag. Four lines of Go would close this gap entirely.

---

### Attack 3: Architecture — existing dependencies not consolidated

**Where**: "No new dependencies. No package additions." (Dependencies subsection, Architecture section)

**Why it's weak**: The statement is accurate but incomplete. The refactor touches `pkg/apperrors`, `pkg.ParseID`, `gorm.ErrRecordNotFound`, `github.com/stretchr/testify`, `net/http/httptest`, and `github.com/gin-gonic/gin` — all named in passing across different sections but never listed together. The rubric asks "Are internal modules and external packages named?" — they are named, but scattered across Error Handling, Testing, and Interface sections. A developer onboarding to this change cannot get a complete dependency picture without reading the entire document. "No new dependencies" answers the wrong question; the question is what dependencies are in play.

**What must improve**: Replace the one-liner with a brief table of internal packages (`pkg/apperrors`, `pkg.ParseID`) and external packages (`gorm`, `gin`, `testify`) that this change depends on. Existing, not new — but named in one place.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Error Handling — existing error types unnamed, service→HTTP propagation gap | ✅ (partial) | `ErrCannotAssignPMRole` and `gorm.ErrRecordNotFound` now named in the Service→Handler→HTTP table; propagation mechanism still unnamed (see Attack 1 above) |
| Interface §6 — abbreviated comment signatures missing return types | ✅ | Full typed BEFORE/AFTER signatures now present for `MainItemService`, `SubItemService`, `ItemPoolService` including return types |
| Architecture — change surface not enumerable from diagram | ✅ | "Complete change surface" table added with all 15 files and change types |

---

## Verdict

- **Score**: 91/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 19/20 — can proceed to `/breakdown-tasks`
- **Action**: Iteration 3 complete. All three iteration-2 attacks addressed. Remaining gaps are minor: unnamed error-mapping helper (-2), missing `ProgressRecord` struct (-2), scattered dependencies (-2). None block breakdown. Design is production-ready for task breakdown.
