# PRD Evaluation Report -- db-dialect-compat

**Iteration:** 3
**Date:** 2026-04-26
**Documents evaluated:**
- `docs/features/db-dialect-compat/prd/prd-spec.md`
- `docs/features/db-dialect-compat/prd/prd-user-stories.md`
**Previous report:** `docs/features/db-dialect-compat/prd/eval/iteration-2.md` (88/100)

---

## Dimension 1: Background & Goals -- 19/20

### Background has three elements (7 pts): 7/7

All three elements are present and specific:
- **原因**: Production incident described with specific failure mode ("需求池转主事项"接口返回 500), commit reference `86fd7c7` for emergency fix, P1-P4 priority attribution (production vs code review).
- **对象**: Three concrete deliverables: fix 4 incompatibility points, introduce dialect abstraction layer, establish prevention mechanism.
- **人员**: Two user types with distinct roles: 开发者 (maintains code for both DB environments) and 运维人员 (deploys MySQL production, verifies end-to-end).

No deductions.

### Goals are quantified (7 pts): 6/7

Two of three goals have numeric targets:
- "4/4 不兼容点全部修复" -- count-based.
- "核心业务操作 100% 通过" -- percentage-based.

The third goal "自动化 lint 检出已知 SQLite 关键字" describes a mechanism, not a quantified target. It does not specify: detection rate (e.g., "100% of SUBSTR/CAST/datetime/pragma_ occurrences"), false-positive tolerance (e.g., "0 false positives on dialect module calls"), or keyword coverage count. This was flagged in iterations 1 and 2 and remains unchanged. -1 pt.

### Background and goals are logically consistent (6 pts): 6/6

Causal chain is tight: production 500 errors caused by SQLite-specific SQL in MySQL -> fix all 4 incompatibilities -> verify MySQL end-to-end -> prevent recurrence via lint + abstraction. No gaps.

---

## Dimension 2: Flow Diagrams -- 18/20

### Mermaid diagram exists (7 pts): 7/7

Present at lines 73-95 of prd-spec.md. Proper `flowchart TD` with nodes and labeled edges.

### Main path complete (7 pts): 6/7

The diagram covers the full runtime flow: application startup -> detect driver -> create GORM dialector -> inject into repos/migrations -> business request triggers -> dialect-safe SQL generation -> execution -> result. However, the diagram conflates two distinct phases into one: the startup/initialization flow (first 6 rows) and the request-handling flow (rows 7-12). The "Ready" node serves as both an end-state for startup and a start-state for request handling. There is no trigger annotation connecting them -- what causes the transition from "应用就绪" to "业务请求触发 NextCode / rebuildTable / HasColumn"? Is it a user HTTP request? A cron job? A system event? This ambiguity was flagged in iteration 2 and persists. -1 pt.

### Decision points + error branches covered (6 pts): 5/6

Two decision diamonds exist:
1. `Driver{driver 值?}` -- three branches (sqlite, mysql, other).
2. `Branch{isMySQL = db.Dialector.Name == mysql?}` -- two branches (true, false).

Error branches:
- `Driver --> |其他| UnsupportedErr --> UnsupportedDialectErr` -- startup failure for unsupported driver.
- `Exec --> |SQL 错误| SQLErr` -- runtime SQL error returns 500.

However:
- The spec text says "若 db 为 nil 或 Dialector 未初始化，函数 panic" -- this failure mode (nil dialector causing panic during request handling) is not represented in the diagram. The diagram assumes the dialector is always valid after startup.
- The SQL error branch terminates at "返回 500, 记录错误日志" with no indication of what happens next (does the request retry? is the transaction rolled back? is the error propagated to the caller with a specific error code?). This was flagged in iteration 2 and persists. -1 pt.

---

## Dimension 3: Functional Specs -- 14/20

### Tables complete (7 pts): 5/7

Section 5.4.1 provides a function API table with columns: 函数, 输入, 输出, 行为说明, 输入约束. This covers the Dialect module's 4 public functions with concrete signatures. Section 5.4.2 and 5.4.3 provide pseudocode behavior descriptions.

However, the rubric asks for "list page 7 elements, button 4 elements, form 2 elements." While this is a backend feature with no UI, the equivalent backend elements are incomplete:
- **Affected HTTP endpoint contracts**: The spec's "受影响的 HTTP 端点映射" table (lines 118-125) lists endpoints and repo methods but does not provide request/response schemas, status codes, or error response formats for any of the listed endpoints. A developer reading this table cannot determine what the API response looks like after the fix.
- **Error response specifications**: The spec says "事务失败时返回 error，调用方返回 500" (line 148) but does not specify the error response body format, error codes, or how to distinguish different failure modes (e.g., dialect detection failure vs SQL execution error vs transaction conflict).
- **Section numbering gap**: Sections 5.1-5.3 are declared "不适用" with a brief explanation. While the explanation is reasonable, the explicit "不适用说明" section (lines 99-101) is new since iteration 2 and is a welcome clarification. But the numbering still implies a gap in the document structure. -2 pts.

### Field descriptions clear (7 pts): 5/7

The function table in 5.4.1 describes purpose, I/O, and behavior for each function. The pseudocode blocks in 5.4.2 and 5.4.3 describe step-by-step behavior.

Weaknesses:
- `Substr` (line 115) now states "起始位置从 1 开始（两种数据库一致，1-indexed）" -- this addresses the iteration 2 finding about 0-vs-1 indexing. Good.
- `CastInt` (line 114) says "SQLite 返回 `CAST(expr AS INTEGER)`；MySQL 返回 `CAST(expr AS SIGNED)`" but does not specify what `expr` is expected to be. Is it a column reference? A computed expression? A literal? The "输入约束" column says "expr 为非空字符串" but does not clarify the format (e.g., "SQL column reference without table prefix" or "any valid SQL expression").
- `Now` (line 116) now states "返回服务器本地时区时间，精度为秒（无毫秒/微秒）" -- this partially addresses the iteration 2 finding about precision. However, it does not address timezone behavior: SQLite's `datetime('now')` returns UTC, while MySQL's `CURRENT_TIMESTAMP` returns the server's configured timezone. The spec claims "返回服务器本地时区时间" but `datetime('now')` returns UTC, not local time. This is a factual inaccuracy in the spec.
- `rebuildTeamMembersTable` (lines 159-167) describes branching logic but does not show the actual DDL differences beyond "SQLite: 使用 AUTOINCREMENT 语法，datetime('now') 默认值" and "MySQL: 使用 AUTO_INCREMENT 语法，CURRENT_TIMESTAMP 默认值". Are there column type differences? Constraint differences? Index differences? A developer implementing P3 needs the full DDL for both dialects. This was flagged in iteration 2 and is only partially addressed. -2 pts.

### Validation rules explicit (6 pts): 4/6

The spec includes "可验证条件" for NextCode (line 151): "对同一 teamID 连续调用 NextCode，在 SQLite 和 MySQL 下均产生严格递增且无间隔的编号序列." This is concrete.

The "输入约束" column in the function table (lines 111-116) specifies: `db` non-nil and Dialector initialized (else panic), `expr` non-empty string, `str` non-empty, `start >= 1`. These are input preconditions.

Weaknesses:
- No validation rules for what happens when preconditions are violated beyond "panic." Is panic acceptable in production? Should there be a graceful degradation? The spec does not address recovery from panic.
- No per-function expected output examples. For `CastInt`, what does `CastInt("code", db)` produce? The spec shows the SQL fragment pattern but not a concrete input-output example (e.g., given input "SUBSTR(code, 8)", SQLite output is `CAST(SUBSTR(code, 8) AS INTEGER)`).
- `rebuildTeamMembersTable` and `HasColumn` have no explicit validation rules or testable conditions in the spec. The "可验证条件" in 5.4.2 only applies to NextCode. This was flagged in iterations 1 and 2. -2 pts.

---

## Dimension 4: User Stories -- 18/20

### Coverage: one story per target user (7 pts): 7/7

Both user types from background have stories:
- 运维人员: Stories 1, 2, and 4 (MySQL deployment scenarios).
- 开发者: Stories 3 and 5 (lint prevention, SQLite regression).

All 4 in-scope fixes (P1-P4) are covered:
- P1/P2: Story 1.
- P3: Story 2.
- P4: Story 4.
- Prevention: Story 3.
- Non-regression: Story 5.

### Format correct (As a / I want / So that) (7 pts): 7/7

All 5 stories follow the As a / I want / So that format. Actions are concrete and positive:
- Story 1: "在 MySQL 生产环境下将需求池条目转换为主事项"
- Story 2: "在空 MySQL 数据库上启动应用并自动完成 RBAC 数据迁移"
- Story 3: "在提交包含硬编码 SQLite 专属 SQL 的代码时收到明确提示"
- Story 4: "在 MySQL 生产环境下执行 RBAC 迁移时 HasColumn 函数正确检测列是否存在"
- Story 5: "在 SQLite 环境下运行所有现有测试并全部通过"

No negation-form actions. No vague verbs like "manage" or "handle."

### AC per story (Given/When/Then) (6 pts): 4/6

All 5 stories have at least one AC in Given/When/Then format.

Improvements since iteration 2:
- Story 1 AC now specifies both number formats explicitly: "主事项编号格式为 `{teamCode}-{seq:05d}`（如 `TEAM-00042`）" and "子事项编号格式为 `{mainCode}-{seq:02d}`（如 `TEAM-00042-03`）". This resolves the iteration 2 finding about conflated formats.
- Story 3 now includes a second AC block for the false-positive case: "Given 开发者在 repo 层使用 `dialect.CastInt()` / `dialect.Substr()` / `dialect.Now()` 生成 SQL（而非硬编码 SQLite 关键字）... Then 提交正常通过，不被拦截（无假阳性）". This resolves the iteration 2 finding about missing false-positive coverage.
- Story 4 AC now specifies concrete inputs and expected outputs: "`HasColumn(db, 'pmw_team_members', 'role_key')` 返回 `true`，`HasColumn(db, 'pmw_team_members', 'nonexistent')` 返回 `false`". This resolves the iteration 2 finding about circular AC.

Remaining weaknesses:
- Story 1 AC verifies format correctness but not sequence correctness across multiple calls. The spec's "可验证条件" says "连续调用 NextCode 验证编号递增" but the AC only checks a single call's output format. The AC should include: "Given 需求池条目已存在且 team 已有主事项编号 TEAM-00041, When 连续转换 2 个需求池条目, Then 第一个生成 TEAM-00042，第二个生成 TEAM-00043".
- Story 5 AC: "运行 `go test ./internal/... ./config/... ./cmd/...`" specifies the command but no test count baseline. "所有现有测试通过" is vague -- how many tests? What counts as "all"? The AC should specify the expected test count or at minimum "N tests pass, 0 failures, 0 skipped" where N is determinable from the codebase. -2 pts.

---

## Dimension 5: Scope Clarity -- 19/20

### In-scope items are concrete deliverables (7 pts): 7/7

All 7 in-scope items name specific technical artifacts: P1-P4 with exact SQL patterns and affected components, dialect module, code standard update, lint check. Each is a specific deliverable, not a vague area.

### Out-of-scope explicitly lists deferred items (7 pts): 7/7

Three items explicitly excluded with justifications:
- Test files using SQLite syntax (in-memory SQLite, no change needed).
- Schema SQL files (already correct per dialect).
- GORM ORM calls (auto-adapts, no issue).

Each has a clear reason for exclusion.

### Scope consistent with functional specs and user stories (6 pts): 5/6

The iteration 1 inconsistency (P4 in-scope with no matching story) is resolved. The "范围项与验证方式映射" table (lines 52-61) maps each in-scope item to a user story and verification method. This is a significant improvement from iteration 2.

Remaining inconsistency:
- The "测试环境前置条件" section (lines 192-200) references "成功标准 1-3（MySQL 集成测试）" and "成功标准 4-6（单元测试）" but these 6 success criteria are never explicitly listed or numbered anywhere in the document. The mapping table (lines 52-61) references "成功标准 1" through "成功标准 6" but the reader must infer what each success standard is from the "验证方式" column. The success criteria should be listed explicitly in a numbered list so that "成功标准 1" through "成功标准 6" are unambiguous references. -1 pt.

---

## Summary of Deductions

| Dimension | Sub-criterion | Deduction | Reason |
|-----------|--------------|-----------|--------|
| Background & Goals | Goals quantified | -1 | Third goal "自动化 lint 检出已知 SQLite 关键字" has no numeric target; flagged in iterations 1-2, unchanged |
| Flow Diagrams | Main path complete | -1 | Startup flow and request flow conflated with ambiguous transition at "Ready" node; flagged in iteration 2, unchanged |
| Flow Diagrams | Error branches | -1 | Nil-dialector panic branch not in diagram; SQL error recovery unspecified; flagged in iteration 2, unchanged |
| Functional Specs | Tables complete | -2 | No HTTP endpoint request/response schemas; no error response body format; HTTP endpoint mapping lacks status codes |
| Functional Specs | Field descriptions | -2 | `Now` claims "服务器本地时区时间" but SQLite `datetime('now')` returns UTC (factual inaccuracy); `CastInt` expr format unspecified; rebuildTeamMembersTable DDL not fully shown |
| Functional Specs | Validation rules | -2 | No expected output examples for dialect functions; no validation rules for rebuildTeamMembersTable or HasColumn; panic recovery unspecified |
| User Stories | AC precision | -2 | Story 1 lacks multi-call sequence verification; Story 5 has no test count baseline |
| Scope Clarity | Consistency | -1 | "成功标准 1-6" referenced but never explicitly listed as a numbered set |

---

SCORE: 88/100

DIMENSIONS:
- Background & Goals: 19/20
- Flow Diagrams: 18/20
- Functional Specs: 14/20
- User Stories: 18/20
- Scope Clarity: 19/20

ATTACKS:
1. Functional Specs: The spec contains a factual inaccuracy -- `Now` claims "返回服务器本地时区时间，精度为秒" but SQLite's `datetime('now')` returns UTC, not local time. MySQL's `CURRENT_TIMESTAMP` returns the server's timezone-configured time. The two functions produce different timezone results, which contradicts the spec's claim of identical behavior. Additionally, `CastInt` does not specify the expected format of `expr` input, and `rebuildTeamMembersTable` only shows 2 DDL differences (AUTOINCREMENT vs AUTO_INCREMENT, datetime vs CURRENT_TIMESTAMP) without showing the full DDL for either dialect, leaving a developer to guess about column types and constraints. The HTTP endpoint mapping table lists affected endpoints but provides no request/response schemas or error response formats.
2. User Stories: Story 1's AC verifies output format for a single call (`TEAM-00042`, `TEAM-00042-03`) but does not verify sequence behavior across consecutive calls. The spec's own "可验证条件" says "连续调用 NextCode...产生严格递增且无间隔的编号序列" -- this condition should be an AC. Story 5 says "所有现有测试通过" without specifying how many tests exist or providing a baseline count, making the AC unverifiable by a reviewer who does not have the codebase open.
3. Functional Specs validation: The 4 dialect functions (`IsMySQL`, `CastInt`, `Substr`, `Now`) have input constraints (non-nil db, non-empty expr) but no concrete input-to-output examples. A developer cannot verify their implementation against expected outputs because none are given. `rebuildTeamMembersTable` and `HasColumn` have no "可验证条件" equivalent -- only `NextCode` has one. The spec references "成功标准 1" through "成功标准 6" but these are never listed as a numbered set; they exist only as references in the mapping table, making traceability incomplete.
