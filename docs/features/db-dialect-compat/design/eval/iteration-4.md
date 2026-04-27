---
date: "2026-04-26"
doc_dir: "docs/features/db-dialect-compat/design/"
iteration: "4"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 4

**Score: 86/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  17      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  6/7     │          │            │
│    Dependencies listed       │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  18      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  5/7     │          │            │
│    Directly implementable    │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  12      │  15      │ ⚠️         │
│    Error types defined       │  3/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  14      │  15      │ ✅         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  6/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  7       │  10      │ ⚠️         │
│    Threat model present      │  3/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  86      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture diagram (lines 17-33) | Component boxes still lack directional arrows showing dependency flow (e.g., `main.go` → `dbutil.NewDialect` → repo injection). Unchanged from iteration 3. | -1 pt (Component diagram present) |
| Dependencies section (line 38) | Internal module import relationships not stated. No list of which packages import `pkg/dbutil`, what `pkg/dbutil` imports, and the `cmd/server/main.go` import graph. Flagged in iteration 2 and 3, still not addressed. | -2 pts (Dependencies listed) |
| Data Models section (line 105) | `type sqliteDialect struct{}` and `type mysqlDialect struct{}` definitions described in prose only ("两个未导出的结构体"), not shown as Go code. Repo struct field addition (`dialect Dialect`) implied by constructor signature but never shown as concrete struct definition. | -1 pt (Models concrete) |
| Data Models / DI section (lines 329-342) | `mainItemRepo` and `subItemRepo` struct definitions are never shown with the added `dialect Dialect` field. The DI section shows constructor call changes but not the struct definition change. | -1 pt (Models concrete) |
| Error Handling / Unrecognized Dialect (line 115) | No custom error type or sentinel variable for unrecognized dialect panic. Uses raw string: `panic("unsupported dialect: postgres, only 'sqlite' and 'mysql' are supported")`. Tests can only assert on substring, not type. Flagged in iteration 2 and 3, still not addressed. | -2 pts (Error types defined) |
| Error Handling / HTTP mapping | No explicit error-to-HTTP-status mapping table. While this is a repo-layer change, the panic from `NewDialect` crashes the server at startup — the document does not explicitly state that there is no HTTP surface (it merely says "不影响 service/handler 层" in the overview, but the error handling section itself lacks this clarification). | -1 pt (HTTP status codes mapped) |
| Testing Strategy (lines 218-266) | The `scripts/lint-staged.sh` change (Story 3) has no automated test plan. The lint spec includes pass/fail examples in a table, but no mechanism to verify the lint blocks bad commits and passes good ones in CI. Flagged in iteration 3, still not addressed. | -1 pt (Per-layer test plan) |
| File Change Summary (line 276) | `backend/internal/repository/main_item_repo.go` listed as "修改" (modify) with scope "无改动" (no changes). A modification entry with no actual changes is misleading in a change summary. Flagged in iteration 3, still not addressed. | -1 pt (Components enumerable) |
| Tasks derivable | Repo struct definition change (adding `dialect Dialect` field to `mainItemRepo` / `subItemRepo`) is only implied by constructor signature and DI section, never shown as concrete struct code. A developer must guess the struct field addition. | -1 pt (Tasks derivable) |
| Security section (lines 189-204) | `HasColumn(db *gorm.DB, table, column string)` accepts arbitrary `string` parameters passed to SQL queries. This function is explicitly in scope (Story 4, line 98, PRD Coverage Map line 213), but its injection surface is completely unanalyzed. Security section only covers `CastInt`/`Substr`. Flagged in iteration 3, still not addressed. | -2 pts (Threat model present) |
| Security section (lines 189-204) | `HasColumn` has no mitigation analysis — no confirmation that `table` and `column` parameters flow into parameterized queries, no identification of call sites, no analysis of whether any receive user-controlled input. Flagged in iteration 3, still not addressed. | -1 pt (Mitigations concrete) |

---

## Attack Points

### Attack 1: [Security — HasColumn injection surface unanalyzed, persists across 2 iterations]

**Where**: PRD Coverage Map line 213: "Story 4: HasColumn MySQL 兼容 | HasColumn 内部增加 isMySQL 分支" / Interface 2 line 98: `func HasColumn(db *gorm.DB, table, column string) bool`
**Why it's weak**: `HasColumn` accepts raw `string` parameters for `table` and `column` and passes them into SQL queries (`information_schema.columns WHERE table_name = ? AND column_name = ?` for MySQL, `pragma_table_info(?)` for SQLite). These are not wrapped in `ColumnExpr` type constraints like `CastInt`/`Substr` are. The security section (lines 189-204) provides a detailed threat model for `CastInt`/`Substr` with a full mitigation table, but `HasColumn` — which has the same class of vulnerability (string parameters embedded in SQL) — receives zero security analysis. This is especially problematic because `HasColumn` uses plain `string` parameters while `CastInt`/`Substr` got the `ColumnExpr` type treatment, creating an inconsistency in the security posture. This was flagged in iteration 3 and remains completely unaddressed.
**What must improve**: Add a paragraph in Security Considerations analyzing `HasColumn`'s injection surface: (1) confirm that `table` and `column` parameters flow into GORM parameterized queries (not string interpolation), (2) enumerate all call sites where `HasColumn` is invoked, (3) confirm none receive user-controlled input, and (4) explain why `string` parameters (vs. `ColumnExpr`) are acceptable here.

### Attack 2: [Error Handling — No typed error for unrecognized dialect panic, persists across 2 iterations]

**Where**: Error Handling / Unrecognized Dialect Fail-Fast (line 115): `panic("unsupported dialect: postgres, only 'sqlite' and 'mysql' are supported")`
**Why it's weak**: The fail-fast panic for unrecognized dialects uses a raw string message. No typed error value exists (`var ErrUnsupportedDialect` or `type UnsupportedDialectError struct{}`). This means: (1) test assertions can only match on substring content, not on type identity, (2) if this code is ever refactored to return errors instead of panicking, there is no error type to migrate to, (3) the panic message format is not guaranteed stable, making test assertions fragile. The iteration-3 report explicitly called this out as Attack 1 and recommended `type UnsupportedDialectError struct { Name string }` with an `Error()` method. This remains unaddressed.
**What must improve**: Define a typed panic value, e.g., `type UnsupportedDialectError struct { Name string } func (e UnsupportedDialectError) Error() string { return fmt.Sprintf("unsupported dialect: %s, only 'sqlite' and 'mysql' are supported", e.Name) }`, and use `panic(UnsupportedDialectError{Name: name})` instead of `panic(fmt.Sprintf(...))`. Tests can then assert on the type using `recover()`.

### Attack 3: [Architecture — Internal import graph remains implicit, persists across 3 iterations]

**Where**: Dependencies section (line 38): "无新增外部依赖。使用已有的 `gorm.io/gorm` 和 `github.com/glebarez/sqlite`、`gorm.io/driver/mysql`。"
**Why it's weak**: External dependencies are listed but internal import relationships are completely absent. The reader must infer from scattered prose across three sections that: `pkg/dbutil` imports `gorm.io/gorm`; `repository/gorm/main_item_repo.go` and `sub_item_repo.go` import `pkg/dbutil`; `cmd/server/main.go` imports both `pkg/dbutil` and `repository/gorm`. This was flagged as Attack 2 in iteration 2, flagged again in iteration 3, and remains unaddressed in iteration 4. For a design document whose purpose is to enable unambiguous implementation, the import graph should be stated explicitly, not left for the reader to deduce from code examples in different sections.
**What must improve**: Add an explicit internal import dependency list, e.g.:
```
pkg/dbutil          → gorm.io/gorm
repository/gorm/*   → pkg/dbutil (new import)
cmd/server/main.go  → pkg/dbutil (new import)
migration/rbac.go   → (no new imports, uses existing isMySQL)
```

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 3): No typed error for unrecognized dialect panic | ❌ No | Line 115 still shows prose description of string-based panic with no typed error. No `UnsupportedDialectError` type or equivalent defined anywhere in the document. |
| Attack 2 (iter 3): Internal dependency relationships remain implicit | ❌ No | Dependencies section (line 38) is unchanged — still lists only external dependencies, no internal import graph. |
| Attack 3 (iter 3): HasColumn injection surface unanalyzed | ❌ No | Security section (lines 189-204) is unchanged — still analyzes only `CastInt`/`Substr`, `HasColumn` receives zero security analysis. |

---

## Verdict

- **Score**: 86/100
- **Target**: 85/100
- **Gap**: 0 points (target reached)
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. Score unchanged from iteration 3 (86/100). All three iteration-3 attack points remain unaddressed, but the document still exceeds the 85-point target. The recurring gaps (typed error for unrecognized dialect, explicit internal import list, HasColumn security analysis) are stable and non-blocking for task breakdown. These can be refined during implementation if desired.
