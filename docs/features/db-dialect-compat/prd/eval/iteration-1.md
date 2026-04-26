# PRD Evaluation Report — db-dialect-compat

**Iteration:** 1
**Date:** 2026-04-26
**Documents evaluated:**
- `docs/features/db-dialect-compat/prd/prd-spec.md`
- `docs/features/db-dialect-compat/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 19/20

### Background has three elements (原因/对象/人员): 7/7

All three elements are present and specific:
- **原因**: Production incident identified ("需求池转主事项"接口返回 500), commit reference `86fd7c7` for the emergency fix, 4 known incompatibilities with priority levels P1-P4 clearly attributed to their discovery source.
- **对象**: Three concrete deliverables -- fix all 4 points, introduce dialect abstraction layer, establish prevention mechanism.
- **人员**: Two user types named with roles -- 开发者 (maintains code for both DB environments) and 运维人员 (deploys MySQL production, verifies end-to-end).

No deductions.

### Goals are quantified: 6/7

Two of three goals have numeric targets:
- "4/4 不兼容点全部修复" -- count-based target.
- "核心业务操作 100% 通过" -- percentage-based target.

The third goal "新增自动化 lint 检查" is a deliverable, not a quantified target. It should specify detection coverage (e.g., "detect 100% of known SQLite-specific keywords") or false-positive tolerance. -1 pt.

### Background and goals are logically consistent: 6/6

Strong causal chain: production 500 errors caused by SQLite-specific SQL in MySQL environment -> fix all 4 incompatibilities -> verify MySQL end-to-end -> prevent recurrence via lint + abstraction. No gaps.

---

## Dimension 2: Flow Diagrams — 16/20

### Mermaid diagram exists: 7/7

A Mermaid flowchart is present (lines 61-74 of prd-spec.md).

### Main path complete (start to end): 5/7

The diagram covers the *prevention workflow* (developer writes code -> uses dialect module? -> generates correct SQL for both DBs -> success). However, this is a future-state operational flow, not the feature delivery flow. The PRD's primary purpose is fixing 4 existing bugs (P1-P4). There is no diagram showing: identify incompatibility point -> apply dialect fix -> verify in both SQLite and MySQL -> mark as resolved. The main path of the actual feature work is missing. -2 pts.

### Decision points + error branches covered: 4/6

One decision point exists ("是否使用方言模块?") with yes/no branches. One error branch: lint detects SQLite keywords -> commit blocked. However:
- No runtime error branch: what happens if dialect auto-detection fails at startup?
- No error branch for: dialect module produces incorrect SQL for one dialect.
- The compatibility requirements state "方言判断必须在应用启动时自动完成" but this startup flow and its failure modes are not diagrammed. -2 pts.

---

## Dimension 3: Functional Specs — 11/20

### Tables complete (list page 7 elements, button 4 elements, form 2 elements): 3/7

The document contains a "关联性需求改动" table with columns: 序号, 涉及项目, 功能模块, 关联改动点, 更改后逻辑说明. This is a change-tracking table, not a functional spec table in the standard format. While this is a backend feature (no UI), the equivalent spec elements are incomplete:
- No API contract details (request/response schemas for affected endpoints).
- No operation triggers (which user actions or system events trigger each change).
- No state transition specifications.
- Rows 5 and 6 (代码规范, 自动化检查) are process/tooling items, not functional specifications of system behavior. -4 pts.

### Field descriptions clear: 4/7

Descriptions are directional but lack specificity:
- Row 1 says "SQLite 用 `CAST(SUBSTR(...) AS INTEGER)`，MySQL 用 `CAST(SUBSTRING(...) AS SIGNED)`" but does not specify: function signature, input parameters, return type, or the complete mapping table for all dialect variants.
- Row 3 says "按方言分支生成建表语句" but does not include the actual DDL differences or the branching logic.
- Row 4 says "MySQL 使用 information_schema.columns" but does not specify the query or how the function signature changes. -3 pts.

### Validation rules explicit: 4/6

The compatibility requirements section states constraints (SQLite 3.x + MySQL 8.0, all existing tests pass, automatic dialect detection). Row 6 describes lint-detected keywords. However:
- No per-field validation rules (e.g., "P1 fix must produce identical results to the original SQLite behavior for codes T001-T999").
- No test case specifications in the functional spec section (the user stories partially compensate, but the spec itself should define validation).
- The "成功标准" section lists environment prerequisites but not pass/fail criteria per fix. -2 pts.

---

## Dimension 4: User Stories — 17/20

### Coverage: one story per target user: 6/7

Two user types from background (开发者, 运维人员) both have stories:
- 运维人员: Stories 1 and 2.
- 开发者: Stories 3 and 4.

However, P4 (HasColumn / pragma_table_info fix) has no dedicated story. It is listed as an in-scope deliverable and described in the functional spec, but no story captures the scenario of deploying with the fixed HasColumn function on MySQL. Story 2's AC ("无 SQL 语法错误") is too broad to specifically cover P4. -1 pt.

### Format correct (As a / I want / So that): 6/7

All 4 stories follow the As a / I want / So that format. Actions are mostly concrete:
- Story 1: "将需求池条目转换为主事项" -- concrete.
- Story 2: "在空 MySQL 数据库上启动应用并自动完成 RBAC 数据迁移" -- concrete.
- Story 3: "在提交包含硬编码 SQLite 专属 SQL 的代码时收到明确提示" -- concrete.
- Story 4: "在 SQLite 环境下开发和测试不受方言改造影响" -- phrased as a negative condition ("不受...影响") rather than a positive action. A stronger formulation: "I want to run all existing tests in SQLite mode and see them pass after the dialect refactoring." -1 pt.

### AC per story (Given/When/Then): 5/6

All stories have at least one AC in Given/When/Then format. Observations:
- Story 2's Then clause says "roles 表包含 3 条预设角色" but does not name the 3 roles (superadmin, pm, member are only mentioned in Story 2's "So that" clause, not in the AC's Then).
- Story 3 has no AC for the false-positive case (developer uses dialect module correctly but lint still triggers).
- Story 4's When clause specifies `go test ./internal/... ./config/... ./cmd/...` but does not define what "all existing tests" means -- is there a specific test count or test suite name? -1 pt.

---

## Dimension 5: Scope Clarity — 19/20

### In-scope items are concrete deliverables: 7/7

All 7 in-scope items name specific technical artifacts: P1-P4 each name the exact SQL pattern and affected component. The dialect module, code standard update, and lint check are concrete deliverables. No vague items like "improve database support."

### Out-of-scope explicitly lists deferred items: 7/7

Three items explicitly excluded with justifications:
- Test files using SQLite syntax (uses in-memory SQLite, no change needed).
- Schema SQL files (already correct per dialect).
- GORM ORM calls (auto-adapts, no issue).

Each has a reason for exclusion. Well done.

### Scope consistent with functional specs and user stories: 5/6

Minor inconsistency: P4 (HasColumn pragma_table_info) is listed as in-scope and described in the functional spec (row 4) but has no dedicated user story. Story 2's AC ("无 SQL 语法错误") broadly covers it but does not specifically validate the HasColumn function's behavior on MySQL. -1 pt.

---

## Summary of Deductions

| Dimension | Sub-criterion | Deduction | Reason |
|-----------|--------------|-----------|--------|
| Background & Goals | Goals quantified | -1 | "新增自动化 lint 检查" has no numeric target |
| Flow Diagrams | Main path complete | -2 | Diagram shows prevention workflow, not the fix/delivery flow for P1-P4 |
| Flow Diagrams | Error branches | -2 | No runtime failure branches (dialect detection failure, incorrect SQL generation) |
| Functional Specs | Tables complete | -4 | Change-tracking table, not functional spec format; missing API contracts, operation triggers, state transitions |
| Functional Specs | Field descriptions | -3 | Descriptions directional but lack function signatures, input/output types, complete dialect mappings |
| Functional Specs | Validation rules | -2 | No per-fix pass/fail criteria in spec section |
| User Stories | Coverage | -1 | P4 (HasColumn) has no dedicated story |
| User Stories | Format | -1 | Story 4 "I want" is a negation, not a positive action |
| User Stories | AC precision | -1 | Story 2 AC does not name the 3 roles; Story 3 no false-positive AC |
| Scope Clarity | Consistency | -1 | P4 in-scope but no matching user story |

---

SCORE: 82/100

DIMENSIONS:
- Background & Goals: 19/20
- Flow Diagrams: 16/20
- Functional Specs: 11/20
- User Stories: 17/20
- Scope Clarity: 19/20

ATTACKS:
1. Functional Specs: The "关联性需求改动" table is a change log, not a functional specification -- it lacks API contracts, operation triggers, input/output schemas, and state transitions. Rows 1-4 describe WHERE changes happen but not WHAT the system behavior looks like after the change (function signatures, return types, error responses). This is the weakest dimension and needs the most work.
2. Flow Diagrams: The Mermaid diagram shows a future-state prevention workflow, not the actual feature delivery flow. A reader looking at the diagram alone would not understand that the feature is about fixing 4 specific existing bugs. The diagram should show: identify each incompatibility -> apply dialect abstraction -> verify on both DBs -> close fix. Additionally, no runtime error branches are shown (dialect detection failure, SQL generation mismatch).
3. Functional Specs field descriptions lack engineering specificity -- Row 1 says "SQLite 用 CAST(SUBSTR(...) AS INTEGER)，MySQL 用 CAST(SUBSTRING(...) AS SIGNED)" but the full dialect mapping is not tabulated, function signatures are absent, and the branching mechanism is unspecified. A developer cannot implement from this spec alone without re-discovering the technical details.
