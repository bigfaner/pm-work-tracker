---
date: "2026-04-26"
doc_dir: "docs/features/db-dialect-compat/design/"
iteration: "3"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 86/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┬────────────┤
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
│ 3. Error Handling            │  11      │  15      │ ⚠️         │
│    Error types defined       │  3/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  3/5     │          │            │
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
│ 6. Security Considerations   │  8       │  10      │ ✅         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  86      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture diagram | Component boxes lack directional arrows showing dependency flow (e.g., `main.go` → `dbutil.NewDialect` → repo injection) | -1 pt (Component diagram present) |
| Dependencies section | Internal module import relationships not stated explicitly — no list of which packages import `pkg/dbutil` and what `pkg/dbutil` imports | -2 pts (Dependencies listed) |
| Data Models section | Concrete struct definitions (`type sqliteDialect struct{}`) and repo struct field additions (`dialect Dialect`) are still described in prose, not shown as Go code | -1 pt (Models concrete) |
| Data Models / DI section | Repo struct field addition (adding `dialect Dialect` field to `mainItemRepo` / `subItemRepo`) is implied by constructor signature change but never shown as a concrete struct definition | -1 pt (Models concrete) |
| Error Handling | No custom error type or sentinel variable defined for unrecognized dialect — only string-based panic messages; programmatic error handling is impossible | -2 pts (Error types defined) |
| Error Handling | No explicit error-to-HTTP-status mapping table; relies on "existing behavior" prose and the reader's knowledge of the `apperrors` package | -2 pts (HTTP status codes mapped) |
| Testing Strategy / scripts layer | The `scripts/lint-staged.sh` change (Story 3) has no automated test plan — how to verify the lint blocks bad commits and passes good ones | -1 pt (Per-layer test plan) |
| File Change Summary | `backend/internal/repository/main_item_repo.go` listed with "修改" change type but "无改动" scope — misleading in a change summary | -1 pt (Components enumerable) |
| Tasks derivable | Repo struct definition change (adding `dialect Dialect` field) is only implied by constructor signature, never shown as concrete struct code | -1 pt (Tasks derivable) |
| Security section | `HasColumn(db *gorm.DB, table, column string)` accepts arbitrary string parameters passed to SQL queries — this is in scope (Story 4) but its injection surface is not analyzed | -1 pt (Threat model present) |
| Security section | `HasColumn` mitigation gap — no analysis of whether `table`/`column` parameters could carry user-controlled input in the current call chain | -1 pt (Mitigations concrete) |

---

## Attack Points

### Attack 1: [Error Handling — No typed error for unrecognized dialect panic]

**Where**: Error Handling / Unrecognized Dialect Fail-Fast: `panic("unsupported dialect: postgres, only 'sqlite' and 'mysql' are supported")`
**Why it's weak**: The fail-fast panic for unrecognized dialects uses a raw string message, not a typed error. This means no caller can programmatically distinguish this panic from other panics (e.g., nil pointer). While this is an init-time failure that crashes the app, the lack of a sentinel like `var ErrUnsupportedDialect` or a typed panic value means: (1) test assertions can only match on substring, not type, and (2) if this code is ever refactored to return errors instead of panicking, there is no error type to migrate to. The iteration-2 report explicitly called out "no `var ErrUnsupportedDialect` or similar sentinel" and this was not addressed.
**What must improve**: Define a typed panic value or sentinel error variable (e.g., `type UnsupportedDialectError struct { Name string }` with an `Error()` method), and use `panic(UnsupportedDialectError{Name: name})` instead of `panic(fmt.Sprintf(...))`. Tests can then assert on the type.

### Attack 2: [Architecture — Internal dependency relationships remain implicit]

**Where**: Dependencies section: "无新增外部依赖。使用已有的 `gorm.io/gorm` 和 `github.com/glebarez/sqlite`、`gorm.io/driver/mysql`。"
**Why it's weak**: External dependencies are listed but internal import relationships are not. The reader must infer from scattered prose that: `pkg/dbutil` imports `gorm.io/gorm`; `repository/gorm/main_item_repo.go` and `sub_item_repo.go` import `pkg/dbutil`; `cmd/server/main.go` imports both `pkg/dbutil` and `repository/gorm`. This is a two-step deduction across three sections (Dependencies, DI section, File Change Summary). The iteration-2 report explicitly flagged "no explicit list of which packages import `pkg/dbutil`" and this was not addressed. For a design document that should enable unambiguous implementation, the import graph should be stated, not inferred.
**What must improve**: Add an explicit import dependency list, e.g.:
```
pkg/dbutil → gorm.io/gorm
repository/gorm/* → pkg/dbutil
cmd/server/main.go → pkg/dbutil, repository/gorm
migration/rbac.go → (no new imports, uses existing isMySQL)
```

### Attack 3: [Security — HasColumn injection surface unanalyzed]

**Where**: PRD Coverage Map, Story 4: "HasColumn MySQL 兼容" / Interface 2: `func HasColumn(db *gorm.DB, table, column string) bool`
**Why it's weak**: The security section provides a detailed threat model for `CastInt`/`Substr` (covering `ColumnExpr` type constraints), but completely ignores `HasColumn` which accepts raw `string` parameters for `table` and `column` and passes them into SQL queries (`information_schema.columns WHERE table_name = ? AND column_name = ?` for MySQL, `pragma_table_info(?)` for SQLite). While these appear to use parameterized queries via GORM, the doc does not verify this — it only analyzes `CastInt`/`Substr`. Story 4 brings `HasColumn` into scope, making it part of this feature's security surface. The inconsistency is notable: `CastInt`/`Substr` get a full injection analysis with typed mitigations, while `HasColumn` — which also embeds string parameters in SQL — receives zero security analysis.
**What must improve**: Add a sentence in Security Considerations analyzing `HasColumn`'s injection surface: confirm that `table` and `column` parameters flow into GORM parameterized queries (not string interpolation), identify all call sites, and confirm none receive user-controlled input.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: MySQL repo path testing gap | ✅ Yes | New "MySQL Repo Path Automated Test Plan" section (lines 157-181) provides concrete test code for `TestNextCode_SelectClause_MySQL` and `TestNextCode_SelectClause_SQLite`, plus a three-layer testing strategy (unit → string assertion → manual integration). The gap is substantially closed. |
| Attack 2: Security enforcement mechanism (comment-only) | ✅ Yes | The `ColumnExpr` type with pre-defined constants (`ColCode`) and compile-time type checking is now the enforcement mechanism. `string` cannot be implicitly passed to `ColumnExpr` parameters. The `NewColumnExpr` export trade-off is explicitly documented. This is a concrete type-level constraint, not just a comment. |
| Attack 3: Story 3 lint check implementation undefined | ✅ Yes | New "Lint Check Specification (Story 3)" section (lines 218-266) defines: exact keyword patterns in a table, scan scope (`backend/internal/repository/gorm/*.go`), exclusion logic (test files, pkg/dbutil, not migration/), concrete bash implementation, and pass/fail examples. Fully specified. |

---

## Verdict

- **Score**: 86/100
- **Target**: 85/100
- **Gap**: 0 points (target reached)
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. All three iteration-2 attacks have been substantively addressed. The remaining gaps (typed error for unrecognized dialect, explicit internal import list, HasColumn security analysis) are minor and can be refined during implementation.
