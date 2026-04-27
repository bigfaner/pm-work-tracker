# Tech Design Evaluation: db-dialect-compat — Iteration 1

**Date:** 2026-04-26
**Document:** `docs/features/db-dialect-compat/design/tech-design.md`
**Evaluator:** Adversarial doc-scorer

---

## Dimension 1: Architecture Clarity (20 pts)

### Layer placement explicit (0-7): 6/7

The doc explicitly states "改动集中在 repository 层（2 个 repo）和 migration 层" in the overview, and the ASCII box diagram clearly shows four layers: `pkg/dbutil` (new), `repository/gorm` (modify), `migration` (modify), `scripts` (modify). Layer placement is unambiguous.

**Deduction (-1):** The ASCII diagram lists `pkg/dbutil` but the File Change Summary table later refers to `backend/internal/pkg/dbutil/dialect.go`. The diagram says "pkg/dbutil" while the actual Go import path is `backend/internal/pkg/dbutil` — the `internal` segment matters for Go package visibility. A developer reading the diagram might assume this is a public `pkg/` package.

### Component diagram present (0-7): 6/7

An ASCII box diagram is present showing the four affected areas with file names and brief descriptions of changes. This is adequate.

**Deduction (-1):** The diagram shows component boxes but does not show data flow or dependencies between the boxes. There are no arrows indicating that `main.go` creates a `Dialect` and injects it into repos, or that `migration/rbac.go` depends on the `isMySQL()` helper. The DI change details are buried in prose below. A component relationship diagram with directional arrows would have been more informative.

### Dependencies listed (0-6): 5/6

The doc lists: "无新增外部依赖。使用已有的 `gorm.io/gorm` 和 `github.com/glebarez/sqlite`、`gorm.io/driver/mysql`。" Internal dependencies are implied by the layer diagram.

**Deduction (-1):** Internal module dependencies are only implied, not explicit. The doc does not state which existing internal packages the new `pkg/dbutil` depends on (it depends on `gorm.io/gorm`), nor does it list the reverse dependency (which packages will import `pkg/dbutil`). The DI section mentions `main.go` importing `dbutil`, but this is not listed as a formal dependency relationship.

**Subtotal: 17/20**

---

## Dimension 2: Interface & Model Definitions (20 pts)

### Interface signatures typed (0-7): 7/7

The `Dialect` interface has full typed signatures with parameters and return types:
- `CastInt(expr string) string`
- `Substr(str string, start int) string`
- `Now() string`

The factory function `NewDialect(db *gorm.DB) Dialect` is also fully typed. The `HasColumn(db *gorm.DB, table, column string) bool` signature is provided. All signatures are clear and implementable.

### Models concrete (0-7): 5/7

The doc explicitly states "无需新增数据模型。`Dialect` 是无状态接口，两个实现均为零大小结构体。" and the two unexported structs (`sqliteDialect{}`, `mysqlDialect{}`) are named. The File Change Summary gives line-count estimates.

**Deduction (-2):** The concrete struct implementations are only described in prose ("两个未导出的结构体 `sqliteDialect{}` 和 `mysqlDialect{}`") — no Go struct literal is shown. A developer implementing this must infer the struct definition. While simple, showing the struct declarations (even just `type sqliteDialect struct{}`) would remove ambiguity. Additionally, the `main_item_repo` struct change (adding a `dialect Dialect` field) is described in prose but never shown as a concrete struct definition.

### Directly implementable (0-6): 4/6

The interface is implementable. However, several implementation details require inference:

**Deduction (-1):** The DI change shows the constructor call (`gormrepo.NewGormMainItemRepo(db, dialect)`) but does not show the updated repo struct definition. The developer must infer that `mainItemRepo` gains a `dialect Dialect` field and that the constructor assigns it.

**Deduction (-1):** The `NextCode` method change is described conceptually ("NextCode 使用 dialect") but the actual SQL query change — the core of this entire feature — is not shown. The doc says `CAST(SUBSTR(code, ?) AS INTEGER)` must become dialect-aware, but never shows the before/after SQL string in the repo. This is the most critical line of code in the entire feature and it is left as an exercise for the reader.

**Subtotal: 16/20**

---

## Dimension 3: Error Handling (15 pts)

### Error types defined (0-5): 2/5

The doc defines one error behavior: panic-on-nil for `NewDialect`. No custom error types or error codes are defined.

**Deduction (-3):** The doc does not address what happens when `NewDialect` is called with a dialector whose name is neither "mysql" nor "sqlite". The prose says "dialect 名称非 'mysql' 时回退到 SQLite（向后兼容）" — this is a design decision, but it silently swallows an unexpected state rather than signaling it. If a new database driver is added (e.g., PostgreSQL), the system would silently use SQLite SQL syntax, which would produce confusing runtime errors rather than a clear startup failure.

### Propagation strategy clear (0-5): 3/5

The doc states that `rebuildTeamMembersTable` returns errors and rolls back transactions (existing behavior). The panic-on-nil pattern is noted.

**Deduction (-2):** The error propagation for the Dialect methods themselves is unclear. If `CastInt`, `Substr`, or `Now` somehow produce an invalid SQL fragment (e.g., empty string input), is that an error? A panic? A silent empty result? The doc does not say. The PRD specifies input constraints ("expr 为非空字符串") but the tech design does not define what happens when constraints are violated.

### HTTP status codes mapped (0-5): 4/5

The doc notes "事务失败时返回 error，调用方返回 500" and the PRD maps affected endpoints. The error mapping is consistent with existing patterns (`apperrors.RespondError`).

**Deduction (-1):** No explicit mapping table of error scenario to HTTP status code is provided. The doc relies on "existing behavior" and the reader's knowledge of the codebase's `apperrors` package rather than stating the mapping explicitly.

**Subtotal: 9/15**

---

## Dimension 4: Testing Strategy (15 pts)

### Per-layer test plan (0-5): 4/5

A test table is provided with four layers: `pkg/dbutil` (unit), `repository/gorm` (existing unit), `migration` (existing unit), and key test scenarios with specific inputs and expected outputs.

**Deduction (-1):** The `scripts` layer (lint-staged.sh) has no test plan. Story 3's AC requires that the lint both blocks bad commits and passes good commits — this is a testable behavior but no test approach is described. Manual verification is implied but not stated.

### Coverage target numeric (0-5): 4/5

The doc states "`dialect.go` 100%" and "repo 和 migration 层现有测试全部通过". The test table column has "100%" for `pkg/dbutil`.

**Deduction (-1):** The "100%" target applies only to `dialect.go`. For repo and migration layers, the target is "现有测试通过" — this is a regression gate, not a coverage target. There is no numeric coverage target for the modified repo code paths (the new dialect-using SQL in NextCode/NextSubCode). A developer could make existing tests pass without ever testing the MySQL dialect path in repos.

### Test tooling named (0-5): 5/5

The doc names `go test + testify` for all Go test layers. This is consistent with the project's testing conventions documented in `.claude/rules/testing.md`.

**Subtotal: 13/15**

---

## Dimension 5: Breakdown-Readiness (20 pts) — critical gate

### Components enumerable (0-7): 6/7

The File Change Summary table lists 8 files with change type and scope. All components can be counted and listed.

**Deduction (-1):** The doc lists `backend/internal/repository/main_item_repo.go` with "修改" change type but "无改动" scope. Including a file with zero changes in a change summary is misleading and creates confusion during task breakdown — is this a task or not?

### Tasks derivable (0-7): 5/7

Each interface maps to an implementation task. The Dialect interface produces `dialect.go` + `dialect_test.go`. The repo changes produce modification tasks. Migration changes produce modification tasks.

**Deduction (-2):** The critical NextCode/NextSubCode SQL refactoring — the core value of this feature — is not decomposed into a clear task with acceptance criteria. The doc says "NextCode 使用 dialect" (~5 lines) but a developer needs to know: (1) which exact SQL string to modify, (2) how to call `dialect.CastInt(dialect.Substr(...))` in the query builder, (3) how to test the MySQL path without a MySQL instance. These are not trivial decisions and the doc does not address them. Also, the `main.go` DI wiring task is mentioned but not explicitly listed in the File Change Summary as a separate task item.

### PRD AC coverage (0-6): 5/6

The PRD Coverage Map table maps all 5 stories to design components. All PRD acceptance criteria have corresponding design elements.

**Deduction (-1):** Story 3's AC includes a negative test case: "repo 层写入硬编码 SQLite 关键字时提交被拦截" AND a positive case "使用 dialect 包时提交通过（无假阳性）". The design addresses the lint check in the File Change Summary ("lint-staged.sh 新增 SQLite 关键字 grep 检查") but does not define the exact grep pattern, the exclusion logic (how to allow dialect.go itself to contain these keywords without triggering), or how to avoid false positives from comments, string literals in test files, or the dialect implementation itself. This is a subtle but important implementation detail that a developer would need to guess.

**Subtotal: 16/20**

---

## Dimension 6: Security Considerations (10 pts)

### Threat model present (0-5): 3/5

The doc states: "无新增安全风险。`Dialect` 只生成 SQL 片段，不拼接用户输入（参数仍通过 GORM 占位符传递）。"

**Deduction (-2):** This is a claim, not a threat model. A proper threat model would identify specific attack vectors: SQL injection via the `expr`/`str` parameters of `CastInt`/`Substr`, the risk of the SQLite fallback masking a misconfigured driver, or the risk that `HasColumn` (which takes table/column names as strings) could be used for information disclosure via `information_schema`. The claim that "parameters still go through GORM placeholders" needs verification — `CastInt` takes a raw string `expr` and embeds it in a CAST expression. If that `expr` ever contains user input (now or in the future), it would be a SQL injection vector.

### Mitigations concrete (0-5): 3/5

The mitigation is stated: "不拼接用户输入（参数仍通过 GORM 占位符传递）".

**Deduction (-2):** This mitigation is not enforced by the design. The `CastInt(expr string)` and `Substr(str string, start int)` methods accept arbitrary strings. There is no type-level constraint (e.g., accepting only an enum of known column expressions) or runtime validation. The security relies entirely on the caller's discipline. A concrete mitigation would be to document which values are expected or add validation.

**Subtotal: 6/10**

---

## Placeholder / TBD Check

No instances of "TBD", "TODO", or placeholder text found. No deduction.

---

## PRD AC Gap Analysis

All 5 user stories are mapped in the PRD Coverage Map. No unaddressed acceptance criteria were found that are completely missing from the design. The partial gaps are captured in the dimension scores above.

---

## Score Summary

SCORE: 77/100

DIMENSIONS:
- Architecture Clarity: 17/20
- Interface & Model Definitions: 16/20
- Error Handling: 9/15
- Testing Strategy: 13/15
- Breakdown-Readiness: 16/20
- Security Considerations: 6/10

ATTACKS:
1. [Interface & Model Definitions]: The core SQL query change — the entire reason this feature exists — is not shown. The doc says "NextCode 使用 dialect" but never provides the before/after SQL string for NextCode or NextSubCode. A developer must open the repo file, find the raw SQL, understand how to decompose `CAST(SUBSTR(code, ?) AS INTEGER)` into `dialect.CastInt(dialect.Substr("code", prefixLen+1))` calls, and figure out how to integrate this with GORM's query builder. This is the hardest part of the implementation and it is left entirely unspecified.
2. [Error Handling]: The silent SQLite fallback for unknown dialect names is a latent defect. The doc says "dialect 名称非 'mysql' 时回退到 SQLite（向后兼容）" — if a future developer configures a PostgreSQL or cockroachdb driver, the system will silently generate SQLite SQL and fail at runtime with confusing syntax errors instead of failing fast at startup. This should be an explicit error or at minimum a logged warning.
3. [Security Considerations]: The claim "不拼接用户输入" is not enforced by the interface design. `CastInt(expr string)` accepts arbitrary strings and directly embeds them into SQL CAST expressions. While current callers pass column names, nothing prevents a future caller from passing user-controlled input. The interface design should either validate input (allowlist of known expressions) or document the security contract explicitly as a code comment and in the design doc.
