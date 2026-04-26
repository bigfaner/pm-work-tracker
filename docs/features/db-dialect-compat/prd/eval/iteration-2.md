# PRD Evaluation Report — db-dialect-compat

**Iteration:** 2
**Date:** 2026-04-26
**Documents evaluated:**
- `docs/features/db-dialect-compat/prd/prd-spec.md`
- `docs/features/db-dialect-compat/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 19/20

### Background has three elements (原因/对象/人员): 7/7

All three elements are present and specific:
- **原因**: Production incident identified ("需求池转主事项"接口返回 500), commit reference `86fd7c7`, P1-P4 with attribution (production incident vs code review).
- **对象**: Three concrete deliverables: fix all 4 points, introduce dialect abstraction layer, establish prevention mechanism.
- **人员**: Two user types with clear roles: 开发者 (maintains code for both DB environments) and 运维人员 (deploys MySQL production, verifies end-to-end).

No deductions.

### Goals are quantified: 6/7

Two of three goals have numeric targets:
- "4/4 不兼容点全部修复" — count-based.
- "核心业务操作 100% 通过" — percentage-based.

The third goal "自动化 lint 检出已知 SQLite 关键字" describes a mechanism, not a quantified target. It should specify: how many keywords, detection rate, or false-positive tolerance (e.g., "detect SUBSTR/CAST/datetime/pragma_ with 0 false positives on dialect module usage"). -1 pt.

### Background and goals are logically consistent: 6/6

Strong causal chain: production 500 errors → fix 4 incompatibilities → verify MySQL end-to-end → prevent recurrence via lint + abstraction. No gaps.

---

## Dimension 2: Flow Diagrams — 18/20

### Mermaid diagram exists: 7/7

Present at lines 61-83 of prd-spec.md. A proper flowchart with `flowchart TD`.

### Main path complete (start to end): 6/7

The diagram now correctly shows the runtime flow: application startup → detect driver → create GORM dialector → inject into repos/migrations → handle business requests → generate dialect-safe SQL → execute → return result. This is a significant improvement from iteration 1. However, the diagram conflates two separate flows into one: the startup/initialization flow (rows 1-6) and the request-handling flow (rows 7-12). The transition from "应用就绪" to "业务请求触发 NextCode / rebuildTable / HasColumn" is a jump — there is no indication of what connects these two phases (e.g., user action, scheduled task). The "Ready" node is both an end-state for startup and a start-state for request handling, which makes the flow ambiguous about scope. -1 pt.

### Decision points + error branches covered: 5/6

Two decision points exist:
1. `Driver{driver 值?}` — three branches: sqlite, mysql, other.
2. `Branch{isMySQL = db.Dialector.Name == mysql?}` — two branches.

One explicit error branch: `Driver --> |其他| UnsupportedErr` leading to `启动失败`. One runtime error: `Exec -->|SQL 错误| SQLErr`.

Improvements from iteration 1: startup failure branch is now present. However:
- The SQL error branch (`Exec -->|SQL 错误| SQLErr([返回 500, 记录错误日志])`) is present but the retry/recovery behavior is unspecified. If a SQL error occurs in NextCode during a transaction, does it retry? Does it return a specific error to the caller? The branch just says "返回 500, 记录错误日志" which is generic.
- No branch for the case where `db.Dialector` is nil (the spec itself says "若 db 为 nil 或 Dialector 未初始化，函数 panic" — this failure mode is described in text but not in the diagram). -1 pt.

---

## Dimension 3: Functional Specs — 14/20

### Tables complete (list page 7 elements, button 4 elements, form 2 elements): 5/7

Major improvement from iteration 1. Section 5.4.1 provides a function API table with columns: 函数, 输入, 输出, 行为说明. This covers the Dialect module's 4 public functions with concrete signatures.

However, the rubric asks for "list page 7 elements, button 4 elements, form 2 elements" — this is a backend-only feature with no UI, so those specific elements don't apply. The equivalent backend elements that should be present:
- **API endpoint contracts**: The spec references affected endpoints (e.g., `POST /v1/teams/:id/item-pool/:id/convert-to-main` appears only in user stories, not in the spec itself). The spec does not list the affected HTTP endpoints, their request/response schemas, or status codes.
- **Error response specifications**: The spec says "事务失败时返回 error，调用方返回 500" but does not specify the error response body format, error codes, or distinguish between different failure modes.
- Section numbering "5.4 系统行为规格" and "5.4.1 / 5.4.2 / 5.4.3" suggests these are subsections of a larger template, but sections 5.1-5.3 are absent, which is confusing. -2 pts.

### Field descriptions clear: 5/7

The function table in 5.4.1 is clear on purpose and I/O for each function. The pseudocode blocks in 5.4.2 and 5.4.3 describe behavior step-by-step.

Weaknesses:
- `CastInt` says "SQLite 返回 `CAST(expr AS INTEGER)`；MySQL 返回 `CAST(expr AS SIGNED)`" — but does not specify the complete SQL expression that wraps these fragments. The reader must infer that the result is interpolated into a larger query. What happens if `expr` contains SQL injection? What is the expected input format (column reference? literal?)?
- `Substr` takes a `start int` parameter but does not specify: is it 0-indexed or 1-indexed? SQLite's SUBSTR and MySQL's SUBSTRING both use 1-based indexing, but this is not stated.
- `Now` function says "SQLite 返回 `datetime('now')`；MySQL 返回 `CURRENT_TIMESTAMP`" — but does not specify precision (millisecond? microsecond?) or timezone behavior. These differ between SQLite and MySQL.
- Section 5.4.3 for `rebuildTeamMembersTable` describes the branching logic but does not show the actual DDL differences. A developer needs to know: what columns, types, constraints differ between the SQLite and MySQL DDL? The spec says "SQLite: 使用 AUTOINCREMENT 语法，datetime('now') 默认值" and "MySQL: 使用 AUTO_INCREMENT 语法，CURRENT_TIMESTAMP 默认值" but this is only two differences — are there others? -2 pts.

### Validation rules explicit: 4/6

Improvements: The spec now includes "可验证条件" for NextCode: "对同一 teamID 连续调用 NextCode，在 SQLite 和 MySQL 下均产生严格递增且无间隔的编号序列." This is a concrete acceptance test.

Weaknesses:
- No validation rules for `CastInt`, `Substr`, or `Now` functions. What are the valid input ranges? What happens with empty strings, null values, or negative start positions?
- The "测试环境前置条件" section describes environment setup but not per-function test cases or expected outputs.
- The spec says "若 db 为 nil 或 Dialector 未初始化，函数 panic" — this is a behavior specification, but there is no validation rule for how this panic should be tested or what the expected recovery behavior is. -2 pts.

---

## Dimension 4: User Stories — 18/20

### Coverage: one story per target user: 7/7

Both user types from background have stories:
- 运维人员: Stories 1, 2, and 4 (MySQL deployment scenarios).
- 开发者: Stories 3 and 5 (lint prevention, SQLite regression).

All 4 in-scope fixes (P1-P4) are now covered:
- P1/P2 (NextCode SUBSTR): Story 1.
- P3 (RBAC DDL AUTOINCREMENT): Story 2.
- P4 (HasColumn pragma_table_info): Story 4.
- Prevention mechanism: Story 3.
- SQLite non-regression: Story 5.

The P4 gap from iteration 1 is resolved. No deductions.

### Format correct (As a / I want / So that): 7/7

All 5 stories follow the As a / I want / So that format. Actions are concrete and positive:
- Story 1: "在 MySQL 生产环境下将需求池条目转换为主事项" — concrete action.
- Story 2: "在空 MySQL 数据库上启动应用并自动完成 RBAC 数据迁移" — concrete action.
- Story 3: "在提交包含硬编码 SQLite 专属 SQL 的代码时收到明确提示" — concrete action.
- Story 4: "在 MySQL 生产环境下执行 RBAC 迁移时 HasColumn 函数正确检测列是否存在" — concrete action.
- Story 5: "在 SQLite 环境下运行所有现有测试并全部通过" — concrete action.

The iteration 1 weakness (Story 4 was a negation "不受...影响") is now fixed. No deductions.

### AC per story (Given/When/Then): 4/6

All 5 stories have at least one AC in Given/When/Then format. Improvements from iteration 1: Story 2's Then clause now names the 3 roles explicitly ("superadmin、pm、member").

Weaknesses:
- Story 1 AC: "接口返回 200，主事项和子事项编号按规则正确生成（如 `T001-S001`）" — the example `T001-S001` is a single format. The spec describes two different formats: `"{teamCode}-{seq:05d}"` for main items and `"{mainCode}-{seq:02d}"` for sub-items. The AC example conflates these. Also, the "按规则正确生成" is vague — what rule? The AC should specify expected output format for at least 2 consecutive calls to verify the sequence.
- Story 3 AC: Only covers the happy path (lint detects SQLite keywords). No AC for: (a) developer uses dialect module correctly and lint does NOT block the commit (false-positive case), (b) lint detects a keyword inside a string literal or comment (should it trigger?). The iteration 1 report flagged this exact gap and it persists.
- Story 4 AC: "HasColumn 通过 information_schema.columns 查询返回正确的布尔值" — "正确的布尔值" is circular. The AC should specify: given a column that exists, returns true; given a column that does not exist, returns false.
- Story 5 AC: "运行 `go test ./internal/... ./config/... ./cmd/...`" — this is the When clause. But no test count baseline is given. How does the reviewer know "all" tests ran? Should specify: "N tests pass, 0 failures" where N is the known test count. -2 pts.

---

## Dimension 5: Scope Clarity — 19/20

### In-scope items are concrete deliverables: 7/7

All 7 in-scope items name specific technical artifacts: P1-P4 with exact SQL patterns and affected components, dialect module, code standard update, lint check. No vague items.

### Out-of-scope explicitly lists deferred items: 7/7

Three items explicitly excluded with justifications:
- Test files using SQLite syntax (in-memory SQLite, no change needed).
- Schema SQL files (already correct per dialect).
- GORM ORM calls (auto-adapts, no issue).

### Scope consistent with functional specs and user stories: 5/6

The iteration 1 inconsistency (P4 in-scope but no matching user story) is now resolved — Story 4 covers HasColumn.

Minor inconsistency: The "测试环境前置条件" section mentions "成功标准 1-3（MySQL 集成测试）" and "成功标准 4-6（单元测试）" but there is no explicit mapping of these 6 success criteria to the in-scope items or user stories. A reader cannot trace which success criterion validates which deliverable. The functional spec's "可验证条件" for NextCode maps to P1/P2 but no equivalent verification criteria are stated for P3 or P4 in the spec. -1 pt.

---

## Summary of Deductions

| Dimension | Sub-criterion | Deduction | Reason |
|-----------|--------------|-----------|--------|
| Background & Goals | Goals quantified | -1 | Lint goal has no numeric target (detection rate, keyword count) |
| Flow Diagrams | Main path complete | -1 | Startup flow and request flow are conflated in one diagram with ambiguous transition |
| Flow Diagrams | Error branches | -1 | Panics and SQL error recovery paths not shown in diagram |
| Functional Specs | Tables complete | -2 | No affected HTTP endpoint contracts; missing error response body format; section numbering gap (5.1-5.3 absent) |
| Functional Specs | Field descriptions | -2 | Substr indexing not specified; Now timezone/precision unspecified; rebuildTeamMembersTable DDL not fully shown |
| Functional Specs | Validation rules | -2 | No per-function input validation rules; panic-on-nil behavior untested |
| User Stories | AC precision | -2 | Story 1 conflates two number formats; Story 3 still lacks false-positive AC; Story 4 AC is circular; Story 5 has no test count baseline |
| Scope Clarity | Consistency | -1 | 6 success criteria mentioned but not mapped to in-scope items or user stories |

---

SCORE: 88/100

DIMENSIONS:
- Background & Goals: 19/20
- Flow Diagrams: 18/20
- Functional Specs: 14/20
- User Stories: 18/20
- Scope Clarity: 19/20

ATTACKS:
1. Functional Specs: The API table in 5.4.1 is a good start but incomplete as a specification — no affected HTTP endpoint contracts are listed in the spec itself (they only appear in user stories), error response body format is absent, Substr does not specify 0-vs-1 indexing, Now does not specify timezone/precision behavior, and rebuildTeamMembersTable's DDL differences are only partially shown. A developer reading only the spec (not the codebase) could not implement P3 or P4 correctly without guessing. The section numbering (5.4) implies missing sections 5.1-5.3 which creates confusion.
2. User Stories: Story 1's AC example `T001-S001` conflates the main item format (`{teamCode}-{seq:05d}`) with the sub-item format (`{mainCode}-{seq:02d}`), making it ambiguous which format the AC validates. Story 3's AC still lacks a false-positive case (developer uses dialect module correctly — lint should NOT trigger), which was flagged in iteration 1 and remains unaddressed. Story 4's AC "返回正确的布尔值" is tautological — it should specify the actual expected values for specific inputs.
3. Functional Specs validation: No per-function input validation rules are stated for `CastInt`, `Substr`, or `Now`. The spec says these functions panic on nil input but provides no validation criteria for boundary conditions (empty strings, negative positions, null expressions). The "测试环境前置条件" references 6 success criteria but they are not mapped to the 7 in-scope items, leaving P3 and P4 without explicit verification criteria in the spec section.
