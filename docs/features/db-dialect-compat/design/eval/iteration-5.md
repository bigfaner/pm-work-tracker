---
date: "2026-04-26"
doc_dir: "docs/features/db-dialect-compat/design/"
iteration: "5"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 5

**Score: 91/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  6/7     │          │            │
│    Dependencies listed       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  6/7     │          │            │
│    Directly implementable    │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  14      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  14      │  15      │ ✅         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  17      │  20      │ ✅         │
│    Components enumerable     │  6/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  9       │  10      │ ✅         │
│    Threat model present      │  5/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  91      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 17/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture diagram (lines 17-33) | Component boxes show layer boundaries but still lack directional dependency arrows (e.g., `main.go → dbutil.NewDialect → repo injection`). The boxes are organized top-down but no arrows show which layer depends on which. | -1 pt (Component diagram present) |
| Dependencies section (line 40-49) | Internal import relationships are now listed with `★ (新增)` markers — good improvement from iteration 4. However, `migration/rbac.go` entry says "无新增导入，复用已有 isMySQL 函数" but does not state what `rbac.go` currently imports or whether it needs to import `pkg/dbutil` for any reason. The design says `HasColumn` delegates to existing `columnExists` (line 400), but the dependency list does not confirm this internal delegation chain. | -1 pt (Dependencies listed) |
| Data Models section (line 114-116) | "无需新增数据模型。`Dialect` 是无状态接口，两个实现均为零大小结构体。" The two implementation structs `sqliteDialect{}` and `mysqlDialect{}` are still only described in prose ("两个未导出的结构体" on line 99) — no Go struct definitions are shown. While the interface is fully typed, the concrete implementations are invisible. | -1 pt (Models concrete) |
| Error Handling / HTTP mapping | The document states in the overview (line 11) "不影响 service/handler 层" but the Error Handling section does not explicitly state that there is no HTTP error surface for the `UnsupportedDialectError` panic (it crashes at startup before HTTP listeners are bound). The reader must infer this. | -1 pt (HTTP status codes mapped) |
| Testing Strategy (lines 162-219) | `scripts/lint-staged.sh` modification (Story 3) still lacks an automated test plan. The document includes pass/fail examples in a table (lines 308-317) and a bash function (lines 292-306), but no mechanism to verify the lint blocks bad commits in CI. The per-layer test table (lines 163-168) has no row for `scripts/`. | -1 pt (Per-layer test plan) |
| File Change Summary (line 328) | `backend/internal/repository/main_item_repo.go` listed as "修改" with scope "无改动" — a modification entry with zero actual changes remains misleading in a change summary. Flagged in iteration 3 and 4, still present. | -1 pt (Components enumerable) |
| Tasks derivable | Repo struct definition changes (adding `dialect Dialect` field to `mainItemRepo` / `subItemRepo`) are implied by the constructor signature change (line 390-391) and the DI section, but the actual struct field addition is never shown as concrete Go code. A developer must infer from the Before/After SQL examples that `r.dialect` is a new struct field. | -1 pt (Tasks derivable) |
| PRD AC coverage | Story 3 AC includes a negative case: "Given 开发者在 repo 层使用 `dialect.CastInt()` / `dialect.Substr()` / `dialect.Now()` 生成 SQL ... Then 提交正常通过，不被拦截（无假阳性）". The lint specification addresses false positives in its pass/fail table (line 313: dialect calls → PASS) but does not include an automated CI test or verification procedure for the false-positive-free case. The PRD coverage map (line 264) maps Story 3 to "lint-staged.sh 新增 grep 检查" but no corresponding test verifies the non-interception case. | -1 pt (PRD AC coverage) |
| Security / Mitigations (line 238) | Residual risk for `NewColumnExpr` is acknowledged but the mitigation is "代码审查 + `NewColumnExpr` 构造函数的视觉信号" — this is a process mitigation, not a technical one. The document does not propose a `go vet` rule, linter check, or build-time enforcement. For a security consideration, relying solely on code review for enforcement is weaker than the `ColumnExpr` type constraint itself. | -1 pt (Mitigations concrete) |

---

## Attack Points

### Attack 1: [Breakdown-Readiness — "无改动" file change entry persists as noise across 3 iterations]

**Where**: File Change Summary line 328: `backend/internal/repository/main_item_repo.go | 修改 | 无改动`
**Why it's weak**: A file listed as "modified" with scope "no changes" is self-contradictory. It was flagged in iteration 3 and iteration 4, and remains unchanged. This creates ambiguity for task breakdown: does a developer create a task for this file or not? Either the file needs changes (interface file must be updated to match repo constructor signatures) and the scope should reflect that, or it does not and the row should be removed from the change summary entirely. Leaving it as-is forces every reader to stop and reason about why a "modification" has no modifications.
**What must improve**: Either remove the row from the File Change Summary (if truly no changes needed) or update the scope to describe the actual change (if the interface definition needs updating to match new constructor signatures).

### Attack 2: [Testing — lint-staged.sh has no automated verification plan across 3 iterations]

**Where**: Testing Strategy per-layer table (lines 163-168) — no row for `scripts/` layer; Lint Check Specification (lines 270-318) has pass/fail examples but no CI test.
**Why it's weak**: Story 3 is one of the PRD's 5 user stories and its AC is specific: lint must block bad commits and pass good ones. The design provides the shell function and example table but no automated way to verify either behavior. If the grep pattern breaks (e.g., regex escaping issue on a different platform), there is no test to catch it. This is the only component in the design with zero automated test coverage, and it was flagged in iteration 3 and 4.
**What must improve**: Add a test row for `scripts/` in the per-layer test table with a concrete verification approach: either a dedicated shellcheck/lint test script (e.g., `tests/test_lint_staged.sh` that creates a temp file with hardcoded SQLite keywords and asserts the function returns 1, then creates a dialect-using file and asserts it returns 0), or a CI step that exercises the lint function against known-good and known-bad files.

### Attack 3: [Interface & Models — Concrete dialect struct definitions remain invisible]

**Where**: Data Models section line 114-116: "无需新增数据模型。`Dialect` 是无状态接口，两个实现均为零大小结构体。" and Interface 1 line 99: "两个未导出的结构体 `sqliteDialect{}` 和 `mysqlDialect{}`"
**Why it's weak**: The interface `Dialect` is fully typed with method signatures. But its two implementations — the core of the feature — are never shown as Go code. The reader knows they exist and are "zero-size structs" but not what they look like. For a developer implementing from this design, the question arises: do they need methods on the struct? Are there any fields? Does `mysqlDialect` embed anything? The document describes behavior in prose but does not provide the struct definitions. This is a minor gap since the implementations are trivial, but for "directly implementable" standard, showing `type sqliteDialect struct{}` and `func (sqliteDialect) CastInt(expr ColumnExpr) string { return fmt.Sprintf("CAST(%s AS INTEGER)", string(expr)) }` would eliminate all guesswork.
**What must improve**: Add concrete Go code for the two struct definitions and at least one method implementation (e.g., `sqliteDialect.CastInt`) as a reference pattern. The others follow trivially.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 4): HasColumn injection surface unanalyzed | ✅ Yes | Lines 240-256 now contain a full "HasColumn 注入面分析" section with parameterized query confirmation, call site enumeration (`migration/rbac.go` and `migration/rbac_test.go`), user-input reachability analysis, and rationale for `string` vs `ColumnExpr`. |
| Attack 2 (iter 4): No typed error for unrecognized dialect panic | ✅ Yes | Lines 126-147 now define `UnsupportedDialectError` struct with `Error()` method and show the panic call `panic(UnsupportedDialectError{Name: name})` with test assertion via `recover()`. |
| Attack 3 (iter 4): Internal import graph remains implicit | ✅ Yes | Lines 40-49 now list explicit internal import relationships with `★ (新增)` markers for all new imports. |

---

## Verdict

- **Score**: 91/100
- **Target**: 85/100
- **Gap**: 0 points (target exceeded by 6 points)
- **Breakdown-Readiness**: 17/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. All three iteration-4 attack points have been addressed. The document improved by +5 points (86 → 91). Remaining gaps are minor: the "无改动" file entry (cosmetic), the untested lint script (process gap), and the hidden concrete struct definitions (minor implementability gap). These are non-blocking for task breakdown.
