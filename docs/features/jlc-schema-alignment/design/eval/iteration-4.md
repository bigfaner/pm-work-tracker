---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "4"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 4

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
| `## Open Questions` line 4 | `[ ] progress_records 和 status_histories 是否需要 biz_key？` — unchecked checkbox despite providing an inline answer "(当前设计：不需要，append-only 表无业务关联需求)". Inconsistency signals the doc is not finalized. | -2 pts |

---

## Attack Points

### Attack 1: Interface & Model — Repo interface signatures absent as typed Go code

**Where**: `### 3. SoftDelete 方法（每个 repo 接口）` — the table lists `TeamRepo`, `SubItemRepo`, `TeamMemberRepo` with "Delete → SoftDelete(ctx, id uint) error", and the implementation is shown for `teamRepo`. But the actual Go interface type definitions are never written out (e.g., `type TeamRepo interface { ... SoftDelete(ctx context.Context, id uint) error }`).

**Why it's weak**: The rubric criterion "Interface signatures typed" requires typed params and return values in interface form, not just in implementation form. A developer adding `SoftDelete` to an existing interface must open the current interface file to find the right method set, then add the new signature — the design gives them the implementation but not the contract. For `TeamMemberRepo`, which is described as "新增 SoftDelete", there is no existing interface to reference, making the gap worse: the developer must infer the full interface from the implementation example alone.

**What must improve**: Add the updated interface type block for at least one repo (e.g., `TeamRepo`) showing the full method set with `SoftDelete` added. For `TeamMemberRepo` (new method), show the complete interface definition since there is no existing contract to extend from.

---

### Attack 2: Architecture — Dependencies table omits all testing packages

**Where**: `### Dependencies` — the table lists exactly 3 entries: `gorm.io/gorm`, `gorm.io/driver/mysql`, `bwmarrin/snowflake`. The Testing Strategy section names `sqlmock` (for repo unit tests) and `Playwright` (for E2E), but neither appears in the dependencies table.

**Why it's weak**: The rubric asks "Are internal modules and external packages named?" Testing dependencies are external packages that must be present in `go.mod` (sqlmock) and `package.json` / installed separately (Playwright). A developer setting up the test environment from this doc would not know to add `DATA-DOG/go-sqlmock` to `go.mod` or run `npx playwright install`. The dependencies section is the canonical place to list what must be installed; scattering package names across sections without a single authoritative list creates setup friction.

**What must improve**: Add `DATA-DOG/go-sqlmock` (new, for repo tests) and `@playwright/test` (existing, confirm version) to the dependencies table, or add a "Test Dependencies" sub-table. At minimum, note whether sqlmock is already in `go.mod` or needs to be added.

---

### Attack 3: Error Handling — Error types described in prose, not defined as Go code

**Where**: `### Error Types & Codes` — the table shows `ERR_VALIDATION → ErrValidation` and `ERR_NOT_FOUND → ErrNotFound` with descriptions and HTTP codes. The propagation section references `apperrors.ErrNotFound` and `apperrors.RespondError` by name.

**Why it's weak**: The rubric asks "Are custom error types or error codes explicitly defined?" The table describes the errors but does not show how they are defined in the `apperrors` package — no `var ErrNotFound = ...`, no sentinel value, no struct type. A developer implementing `SoftDelete` must open `backend/pkg/errors/` (or wherever `apperrors` lives) to find the correct error value to return. If `ErrNotFound` does not yet exist in that package (this is a new feature), the developer has no spec for what to create. The doc assumes the package already has these errors without confirming it.

**What must improve**: Add a one-line Go snippet showing the error definition (e.g., `var ErrNotFound = errors.New("record not found")` or the actual sentinel pattern used in this codebase), or explicitly state "these errors already exist in `backend/pkg/errors/errors.go` at line N". Either confirms the contract; the current table does neither.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Go model structs absent for all domain models | ✅ Partial | `ProgressRecord`, `StatusHistory`, and `TeamMember` now have full Go struct definitions in `### 1b. Deviation Model Structs`. Type mapping note added: "BIGINT UNSIGNED → uint, BIGINT → int64, DECIMAL(5,2) → float64". `MainItem`, `SubItem`, `Team`, `User`, `ItemPool` remain DDL-only, but the deviation models (the highest-risk ones) are now covered. |
| Attack 2: Frontend coverage target non-numeric ("不降低") | ✅ | `### Overall Coverage Target` now reads "前端：≥70%（当前基线，维持不降低）" — numeric floor established. |
| Attack 3: id exposure unmitigated; snowflake timing leak omitted | ✅ | Threat 3 now states "auth middleware 在所有资源端点校验 team membership，枚举越权 id 返回 403，不泄露数据". Threat 4 explicitly covers biz_key indirect leak via logs with concrete mitigation "logging middleware 不得打印完整 model struct（应只记录 id）". |

---

## Verdict

- **Score**: 90/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 20/20 — can proceed to `/breakdown-tasks`
- **Action**: Score improves from 78 → 90. All three iter-3 attacks addressed. Remaining gaps are minor: repo interface type definitions absent (Attack 1), testing packages missing from dependencies table (Attack 2), error types not shown as Go code (Attack 3). None block breakdown.
