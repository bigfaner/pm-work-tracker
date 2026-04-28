---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/design/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 85/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  16      │  20      │ ⚠️          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   5/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  17      │  20      │ ⚠️          │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   5/7    │          │            │
│    Directly implementable    │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  10      │  15      │ ⚠️          │
│    Error types defined       │   2/5    │          │            │
│    Propagation strategy clear│   3/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅          │
│    Components enumerable     │   6/7    │          │            │
│    Tasks derivable           │   6/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   9      │  10      │ ✅          │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  85      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Error Handling section | "No new error types. Existing error propagation is unchanged." — the section never names which existing error types are in play; `ErrCannotAssignPMRole` appears only in the Testing section, not here | -3 pts |
| Error Handling section | Propagation from service → handler → HTTP response is undocumented; only the middleware → HTTP path is mapped; a reader cannot trace what happens when `ProgressService.Append` returns an error | -2 pts |
| Interface §6 | `// BEFORE: Create(ctx, teamID, pmID uint, req) / List(ctx, teamID uint, ...)` — abbreviated comment-style, return types entirely absent; not actual typed Go signatures | -1 pt |
| Data Models section | "No model changes. `ProgressRecord.TeamKey` is already `int64`" — the `ProgressRecord` struct is never shown; a reviewer cannot confirm the field type without reading the source | -2 pts |
| Architecture section | "All 7 handler files" and "8 service files" referenced in prose (§7) but never enumerated by name; the component diagram shows layer flow only, not the full change surface | -2 pts |
| Security section | Threat 3 mitigation is "this is not a new problem" — not a concrete countermeasure; the deployment guard (middleware registration check) is not described | -1 pt |

---

## Attack Points

### Attack 1: Error Handling — existing error types unnamed, service→HTTP propagation gap

**Where**: "No new error types. Existing error propagation is unchanged." and the entire Error Handling section, which maps only middleware failures to HTTP status codes.

**Why it's weak**: The middleware → HTTP table is now present and correct, but the section stops there. The service layer returns errors (`ErrCannotAssignPMRole` is mentioned in Testing §2, not here) and those errors must propagate through the handler to an HTTP response. The design never states: which existing error types are relevant to this refactor, how the handler translates a service error to a status code, or whether the `FindByBizKey(0)` not-found error from a zero-bizKey call produces a 404 or 500. "Existing error propagation is unchanged" is an assertion that cannot be verified without reading the source — it is not a design decision.

**What must improve**: Add a table mapping the relevant existing error types (at minimum `ErrCannotAssignPMRole`, `ErrRecordNotFound` from `FindByBizKey`) to their HTTP status codes. One sentence confirming the handler error-mapping middleware (if any) handles these is sufficient — but it must be explicit.

---

### Attack 2: Interface §6 — abbreviated comment signatures missing return types

**Where**: "// BEFORE: Create(ctx, teamID, pmID uint, req) / List(ctx, teamID uint, ...)" and "// AFTER: Create(ctx, teamBizKey int64, pmID uint, req) / List(ctx, teamBizKey int64, ...)"

**Why it's weak**: These are comment-style abbreviations, not typed Go signatures. The return types are entirely absent — a developer cannot tell whether `Create` returns `(*model.MainItem, error)` or `(uint, error)` or something else. The same applies to `SubItemService` and `ItemPoolService`. Sections 1–5 all provide full typed BEFORE/AFTER signatures; §6 breaks that pattern for the three highest-traffic services in the codebase. The rubric requires "typed params and return values (not prose)" — these fail that bar.

**What must improve**: Replace the comment abbreviations with full typed Go signatures matching the pattern used in §2–§5. At minimum show one complete method per service (e.g., `Create` and `List`) with return types.

---

### Attack 3: Architecture — change surface not enumerable from diagram

**Where**: "All 7 handler files replace: `teamID := middleware.GetTeamID(c)` → `teamBizKey := middleware.GetTeamBizKey(c)`" and the ASCII layer diagram.

**Why it's weak**: The diagram shows the layer flow correctly, but "all 7 handler files" and "8 service files" are never listed by name. A developer picking up this design cannot derive the complete file list — they must grep the codebase. This matters for breakdown: if a task is "update all handler call sites," the assignee has no authoritative list to work from and no way to verify completeness without independent discovery. The iteration 1 report flagged this; it remains unfixed.

**What must improve**: Add a flat enumeration of the affected files — either inline in §7 or as an appendix table. Format: filename → change type (e.g., `progress_handler.go` → replace `GetTeamID` with `GetTeamBizKey`). This is the minimum a task breakdown needs.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Error Handling — complete absence of HTTP error contract | ✅ | Middleware failure → HTTP status map table added with 4 conditions; `GetTeamBizKey` failure behavior documented with code |
| Testing Strategy — no numeric coverage target, no tooling named | ✅ | "Maintain ≥ 75% statement coverage" baseline stated; tooling table names `go test -race`, `testify/assert`, `testify/require`, `httptest`, `gin.CreateTestContext` |
| Security — one-sentence dismissal in RBAC system | ✅ | Three named threats with analysis: cross-team bizKey, isPMRole privilege escalation, zero bizKey routing misconfiguration |

---

## Verdict

- **Score**: 85/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Iteration 2 complete. All three iteration-1 attacks addressed. Remaining gaps are in Error Handling (10/15) and Interface §6 abbreviated signatures — neither blocks breakdown. Design is production-ready for task breakdown.
