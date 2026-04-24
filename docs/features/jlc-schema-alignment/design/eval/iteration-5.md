---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "5"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 5

**Score: 90/100** (target: N/A)

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
│ 2. Interface & Model Defs    │  16      │  20      │ ⚠️          │
│    Interface signatures typed│   5/7    │          │            │
│    Models concrete           │   6/7    │          │            │
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
│ TOTAL (before deductions)    │  92      │  100     │            │
│ Deductions                   │  -2      │          │            │
│ TOTAL                        │  90      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 20/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| `## Open Questions` line 4 | `[ ] progress_records 和 status_histories 是否需要 biz_key？` — checkbox unchecked despite inline answer "(当前设计：不需要，append-only 表无业务关联需求)". Inconsistency signals the doc is not finalized. | -2 pts |

---

## Attack Points

### Attack 1: Interface & Model — Repo interface type definitions still absent

**Where**: `### 3. SoftDelete 方法（每个 repo 接口）` — the table lists `TeamRepo`, `SubItemRepo`, `TeamMemberRepo` with "Delete → SoftDelete(ctx, id uint) error", and the implementation is shown for `teamRepo`. No Go interface type block appears anywhere (e.g., `type TeamRepo interface { ... SoftDelete(ctx context.Context, id uint) error }`).

**Why it's weak**: The rubric criterion "Interface signatures typed" requires typed params and return values in interface form, not just in implementation form. This was Attack 1 in iteration 4 and remains unaddressed. For `TeamMemberRepo` specifically — described as "新增 SoftDelete" — there is no existing interface to reference, so the developer must infer the full interface from the implementation example alone. The design gives the implementation but not the contract.

**What must improve**: Add the updated interface type block for at least `TeamRepo` showing the full method set with `SoftDelete` added. For `TeamMemberRepo` (new method on a new interface), show the complete interface definition.

---

### Attack 2: Architecture — Component diagram arrow misrepresents snowflake call site

**Where**: `### Component Diagram` — the ASCII diagram shows `pkg/snowflake/generator.go` connected via arrow into `repo/gorm/*.go`, which then connects to `service/*.go`. But `### 5. Service 层 biz_key 赋值` explicitly states that `Create()` in service files calls `snowflake.Generate()` — the call originates in the service layer, not the repo layer.

**Why it's weak**: The diagram is the first thing a developer reads to understand data flow. Showing snowflake feeding into repo implies the repo is responsible for generating biz_key, which contradicts the implementation spec. A developer following the diagram would put `snowflake.Generate()` in the wrong layer, creating a design inconsistency that tests would not catch (both layers compile fine).

**What must improve**: Redraw the arrow so `pkg/snowflake/generator.go` connects directly to `service/*.go`, not to `repo/gorm/*.go`. This is a one-line diagram fix with zero ambiguity.

---

### Attack 3: Error Handling — Error types described in table, not defined as Go code

**Where**: `### Error Types & Codes` — the table shows `ERR_VALIDATION → ErrValidation` and `ERR_NOT_FOUND → ErrNotFound` with descriptions and HTTP codes. The propagation section references `apperrors.ErrNotFound` and `apperrors.RespondError` by name.

**Why it's weak**: This was Attack 3 in iteration 4 and remains unaddressed. The table describes the errors but does not show how they are defined in the `apperrors` package — no `var ErrNotFound = ...`, no sentinel value, no struct type. A developer implementing `SoftDelete` must open `backend/pkg/errors/` to find the correct error value. If `ErrNotFound` does not yet exist in that package, the developer has no spec for what to create.

**What must improve**: Add a one-line Go snippet showing the error definition (e.g., `var ErrNotFound = errors.New("record not found")`), or explicitly state "these errors already exist in `backend/pkg/errors/errors.go`". Either confirms the contract; the current table does neither.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 4): Repo interface type definitions absent | ❌ | No `type TeamRepo interface{...}` block added. Table still lists method changes in prose form only. |
| Attack 2 (iter 4): Testing packages missing from dependencies table | ❌ | Dependencies table still lists only 3 entries: `gorm.io/gorm`, `gorm.io/driver/mysql`, `bwmarrin/snowflake`. `DATA-DOG/go-sqlmock` and `@playwright/test` absent. |
| Attack 3 (iter 4): Error types not shown as Go code | ❌ | `### Error Types & Codes` still a prose table. No Go sentinel definition added. |
| Rule change (iter 5 trigger): biz_key exposed as `json:"bizKey"`, id hidden as `json:"-"` | ✅ | `BaseModel` correctly shows `BizKey int64 ... json:"bizKey"` and `ID uint ... json:"-"`. Security section updated: "biz_key 通过 `json:"bizKey"` 对外暴露". PRD Coverage Map row added: "biz_key 对外暴露（json:"bizKey"），id 不对外暴露（json:"-"）". Frontend type diff table correctly shows `id: number → removed (json:"-"，不出现在响应体)` and `bizKey: string (新增)`. |

---

## Verdict

- **Score**: 90/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 20/20 — can proceed to `/breakdown-tasks`
- **Action**: Score holds at 90. The iter-5 rule change (biz_key/id json tags) is correctly reflected throughout the document. All three iter-4 attacks remain unaddressed and continue to cap the score at 90. None block breakdown.
