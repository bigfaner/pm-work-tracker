---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "6"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 6

**Score: 92/100** (target: N/A)

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
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   7/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  14      │  15      │ ⚠️          │
│    Error types defined       │   4/5    │          │            │
│    Propagation strategy clear│   5/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  20      │  20      │ ✅          │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  10      │  10      │ ✅          │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  94      │  100     │            │
│ Deductions                   │  -2      │          │            │
│ TOTAL                        │  92      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 20/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| `## Open Questions` last item | `[ ] progress_records 和 status_histories 是否需要 biz_key？` — checkbox unchecked despite inline answer "(当前设计：不需要，append-only 表无业务关联需求)". Inconsistency signals the doc is not finalized. Carry-over from iter 5. | -2 pts |

---

## Attack Points

### Attack 1: Interface & Model — Formal Go interface type blocks still absent

**Where**: `### 3b. FindByBizKey 方法（每个 repo 接口）` — the table lists five repos with typed signatures (`FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error)`) and an implementation example is shown. But no `type MainItemRepo interface { ... }` block appears anywhere in the document. Same gap exists for `SoftDelete` in `### 3. SoftDelete 方法`.

**Why it's weak**: The rubric criterion "Interface signatures typed" requires typed params and return values in interface form. The table format is an improvement over prose, but it is not a Go interface contract. A developer implementing `UserRepo` or `ItemPoolRepo` must open the existing source files to discover the full method set before they can add `FindByBizKey` — the design does not give them the complete interface. For `TeamMemberRepo.SoftDelete` (a new method on an existing interface), there is no interface block at all, so the developer cannot verify the full contract without reading the source. This has been Attack 1 for iterations 4 and 5 and remains unaddressed.

**What must improve**: Add a Go interface type block for at least one representative repo (e.g., `MainItemRepo`) showing the complete method set with `FindByBizKey` and `SoftDelete` added. For `TeamMemberRepo`, show the full interface definition since `SoftDelete` is a new addition to it.

---

### Attack 2: Architecture — Component diagram snowflake arrow misrepresents call site

**Where**: `### Component Diagram` — `pkg/snowflake/generator.go` is connected via arrow into `repo/gorm/*.go`, which then connects to `service/*.go`. But `### 5. Service 层 biz_key 赋值` explicitly states that `Create()` in service files calls `snowflake.Generate()` — the call originates in the service layer.

**Why it's weak**: The diagram is the first artifact a developer reads to understand data flow. Showing snowflake feeding into repo implies the repo is responsible for generating `biz_key`, which directly contradicts the implementation spec in Section 5. A developer following the diagram would place `snowflake.Generate()` in the wrong layer. This has been Attack 2 since iteration 5 and remains unaddressed despite being a one-line diagram fix.

**What must improve**: Redraw the arrow so `pkg/snowflake/generator.go` connects directly to `service/*.go`, not to `repo/gorm/*.go`.

---

### Attack 3: Error Handling — Error types described in table, not defined as Go code

**Where**: `### Error Types & Codes` — the table shows `ERR_NOT_FOUND → ErrNotFound` and `ERR_VALIDATION → ErrValidation` with HTTP codes. The propagation section references `apperrors.ErrNotFound` and `apperrors.RespondError` by name.

**Why it's weak**: The table describes the errors but does not show how they are defined in the `apperrors` package — no `var ErrNotFound = ...`, no sentinel value, no struct type. A developer implementing `SoftDelete` or `FindByBizKey` must open `backend/pkg/errors/` to find the correct error value. If `ErrNotFound` does not yet exist in that package, the developer has no spec for what to create. This has been Attack 3 for iterations 4 and 5 and remains unaddressed.

**What must improve**: Add a one-line Go snippet showing the error definition (e.g., `var ErrNotFound = errors.New("record not found")`), or explicitly state "these errors already exist in `backend/pkg/errors/errors.go`". Either confirms the contract; the current table does neither.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 5): Repo interface type definitions absent | ❌ | No `type TeamRepo interface{...}` block added. FindByBizKey section adds typed signatures in table form — an improvement — but formal Go interface type blocks remain absent. |
| Attack 2 (iter 5): Component diagram snowflake arrow points to repo instead of service | ❌ | Diagram unchanged. `pkg/snowflake/generator.go` still arrows into `repo/gorm/*.go`. |
| Attack 3 (iter 5): Error types not shown as Go code | ❌ | `### Error Types & Codes` still a prose table. No Go sentinel definition added. |
| Iter-6 trigger: FindByBizKey interface + handler layer changes | ✅ | Section 3b added with typed signatures for all 5 repos and implementation pattern. Handler before/after code snippets added. PRD Coverage Map row added: "API 路径参数改用 bizKey，后端通过 FindByBizKey 定位记录". Service table updated with `GetByBizKey()` entries. |

---

## Verdict

- **Score**: 92/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 20/20 — can proceed to `/breakdown-tasks`
- **Action**: Score improves from 90 to 92. The iter-6 additions (FindByBizKey typed signatures in table form, handler before/after snippets) partially address the interface signatures gap (+1 pt). All three iter-5 attacks remain unaddressed and continue to cap the score at 92.
