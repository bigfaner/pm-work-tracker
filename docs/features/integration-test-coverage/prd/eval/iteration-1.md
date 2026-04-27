# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 1
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 19/20

### 三要素 (原因/对象/人员) — 7/7

All three elements are present and specific:

- **原因**: 54 endpoints total, only 18 (33%) have integration tests, 36 completely uncovered. Concrete incident cited: commit `1883499` introduced timezone and filter bugs in `view_handler.go` and `view_service.go` that escaped to manual testing (referenced in `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3).
- **对象**: Write integration tests for 36 untested endpoints, organized by user flow. Close 6 unit test gaps.
- **人员**: Developer and Code Reviewer, each with distinct responsibilities.

All three elements are concrete, not vague. The incident citation with commit hash and file names is commendable.

### Goals quantified — 7/7

Four quantified targets in a table:

1. Endpoint integration test coverage: 33% → 100%
2. New test cases: ≥ 150
3. Unit test gaps: 6/6 closed
4. Test suite execution time: < 150 seconds

All four targets have specific numbers with units. No vague targets like "improve coverage" or "add more tests."

### Background and goals logically consistent — 5/6

The goals directly address the stated problem: 36 uncovered endpoints → 100% coverage target. The 150+ test case target is reasonable for 36 endpoints across multiple test scenarios (happy path, validation, permission, not found, cascade).

**Deduction (-1):** The background names "Code Reviewer" as a target user type ("通过 PR 级别的增量提交审查测试代码"), but none of the four goals address the reviewer's workflow. No goal mentions PR organization, incremental commit structure, review readability, or anything traceable to the reviewer persona. The goals are entirely developer-centric. A user persona stated in the background should have at least one goal that traces back to it.

---

## Dimension 2: Flow Diagrams — 16/20

### Mermaid diagram exists — 7/7

A Mermaid flowchart is present at lines 76-94 of prd-spec.md. It uses `flowchart TD` syntax with standard node shapes (rounded rectangles, diamonds, circles).

### Main path complete (start → end) — 5/7

The diagram shows a linear chain: Start → F1 → F2 → F3 → F4 → F5 → F6 → Check → Done, with a feedback loop through Gap → Check. All six features are traversed.

**Deduction (-2):** The flowchart depicts the **project execution sequence** (which feature to implement first), not the **business logic flow** being tested. The text at lines 66-73 describes actual user flows (e.g., "创建主项 → 创建子项 → 追加进度 → 状态变更 → 归档"), but the Mermaid diagram reduces each rich flow to a single box labeled "F1: Item Lifecycle / 17 endpoints." For a test coverage PRD, the diagram should ideally show the flow within each test scenario (create → assert → next step), not just the implementation order. The "Shared Test Helpers" dotted lines are a nice touch showing dependency, but the main flow is overly abstracted.

### Decision points + error branches — 4/6

There is one decision diamond (`Check{150+ 测试用例? / 54 端点全覆盖?}`) with two branches: "是" → Done, "否" → Gap → loop back. This is one decision point with one error/recovery branch.

**Deduction (-2):** Only a single decision point in the entire diagram. Missing branches for realistic project risks: what happens if integration tests reveal actual bugs in the code under test? What happens if a flow file exceeds the 30-second execution budget? What happens if shared helpers need refactoring mid-stream? The diagram presents a frictionless linear path with only one checkpoint, which is optimistic for a 40-hour effort spanning 54 endpoints.

---

## Dimension 3: Functional Specs — 16/20

### Tables complete — 7/7

Five test matrices (F1-F5) and one unit gap table (F6) are present. F1 has 17 rows, F2 has 6, F3 has 9, F4 has 6, F5 has 6, and F6 has 6 rows. Every cell is filled with specific expected behavior — no empty cells, no "TBD" placeholders. The total matches the stated scope: 17+6+9+6+6 = 44 endpoint test rows (36 integration + 6 unit + 2 extra for sub-item detail views) plus the 6 unit gap rows.

### Field descriptions clear — 5/7

Most endpoint entries are specific about expected behavior. For example:
- `POST /.../subId/progress`: "追加 200" / "回退 → 422" / "100% 自动状态转换；完成度上卷主项"
- `POST /.../poolId/assign`: "分配：创建子项+更新池状态 200" / "无效主项 → 回滚" / "已处理 → 409"

**Deduction (-2):** Several endpoints use abbreviated paths that omit the full URL:
- `PUT /.../status` — which exact path?
- `GET /.../available-transitions` — for main-item or sub-item?
- `POST /.../archive` — full path not shown
- `PUT /.../subId` — abbreviated
- `PUT /.../subId/status` — abbreviated
- `POST /.../subId/progress` — abbreviated
- `PATCH /.../recordId/completion` — abbreviated
- All F2 endpoints use `.../poolId/` abbreviation
- All F3 member endpoints use `.../` abbreviation

While the context makes these inferable, a test writer should not need to guess. The first two tables (F1 main-item section) use full paths, then the document switches to abbreviations inconsistently. Additionally, F4's permission column entries like "→ 403" don't specify who gets rejected — the reader must cross-reference to user stories to learn it's non-SuperAdmin users.

### Validation rules explicit — 4/6

Many validation rules are concrete:
- "缺标题/无效优先级/无效日期 → 422"
- "权重/重复 code → 422"
- "无效转换 → 422"
- "缺原因 → 422"
- "PM 不可移除 → 422"
- "禁用自身 → 422"

**Deduction (-2):** Three endpoints have vague validation entries:
- `PUT /teams/:id/main-items/:itemId`: "验证错误 → 422" — what errors? Which fields? What invalid values?
- `PUT /teams/:id`: "验证 → 422" — same problem, even vaguer
- `PUT /admin/users/:userId`: "验证 → 422" — same problem

For a PRD whose entire purpose is to define exactly what tests to write, saying an endpoint has "validation errors" without specifying what is being validated is a significant gap. A test writer cannot write a validation test case from "验证 → 422" — they need to know what input triggers the 422.

---

## Dimension 4: User Stories — 16/20

### Coverage: one story per target user — 3/7

The background defines two user types: **Developer** and **Code Reviewer**.

All six stories are "As a 开发者". Zero stories address the Code Reviewer persona.

**Deduction (-4):** The Code Reviewer is explicitly named in the background as a target user ("代码审查者：通过 PR 级别的增量提交审查测试代码"). The document even describes a specific workflow for reviewers (incremental PR-based review). Yet no story captures what a reviewer needs: organized commits, readable test structure, meaningful test names, or review documentation. Half of the stated user personas have zero story coverage.

### Format correct (As a / I want / So that) — 7/7

All six stories follow the format precisely:
- **As a** 开发者 (consistent persona)
- **I want to** [specific action with concrete details]
- **So that** [business justification]

The "I want" clauses list specific flow steps (not vague verbs like "manage" or "handle"):
- Story 1: "验证 MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive 的完整生命周期"
- Story 6: "为 permission_handler、ConvertToMain、UpdateTeam、GetByBizKey 等未测试方法补充单元测试" (names exact methods)

No vague actions. Each "I want" is a concrete, testable action.

### AC per story (Given/When/Then) — 6/6

Every story has multiple ACs in Given/When/Then format:
- Story 1: 7 ACs
- Story 2: 6 ACs
- Story 3: 6 ACs
- Story 4: 5 ACs
- Story 5: 5 ACs
- Story 6: 4 ACs

Total: 33 acceptance criteria across 6 stories. All follow Given/When/Then. All are specific and testable — they include exact HTTP status codes (201, 200, 403, 404, 409, 422), specific data values (`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`), and concrete preconditions.

---

## Dimension 5: Scope Clarity — 18/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:
- F1: "17 个端点（MainItem CRUD → SubItem → Progress → Status → Archive）"
- F2: "6 个端点（Submit → Assign/Convert/Reject）"
- F3: "9 个端点（CRUD + 成员管理 + 角色变更）"
- F4: "6 个端点（用户 CRUD + 状态切换 + 团队列表）"
- F5: "6 个端点（Weekly/Gantt/Table/CSV/Report Preview/Export）"
- F6: "6 个缺口（permission_handler、ConvertToMain、UpdateTeam、3x GetByBizKey）"
- Shared test helpers (extracted from existing tests)

No vague items like "improve testing" or "better coverage." Each item names the specific endpoints or methods.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with reasons:
1. "前端测试变更" — with reason: "前端测试套件已覆盖组件和 E2E 流程"
2. "性能/负载测试" — named, no ambiguity
3. "E2E 浏览器测试" — with reason: "独立工作流"
4. "新功能或 bug 修复" — with reason: "本需求纯粹是测试覆盖"

All are named, not implied by absence. Each has a brief rationale.

### Scope consistent with functional specs and user stories — 4/6

The six scope items (F1-F6) map directly to:
- Six functional spec tables (F1-F6) in prd-spec.md
- Six user stories (Stories 1-6) in prd-user-stories.md

This is clean traceability for the core scope.

**Deduction (-2):** The in-scope list includes "共享测试辅助函数（从现有集成测试中提取复用）" as a deliverable. However:
- No functional spec table describes what these helpers are, what functions they expose, or what existing tests they are extracted from.
- No user story covers the helper extraction task.
- The Mermaid diagram shows helpers as a dotted dependency, but they are never specified.

A deliverable that appears in scope but has no corresponding spec or story is a traceability gap. Either it should have its own spec section, or it should be noted as a supporting artifact rather than a separate in-scope feature.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 19 | 20 |
| Flow Diagrams | 16 | 20 |
| Functional Specs | 16 | 20 |
| User Stories | 16 | 20 |
| Scope Clarity | 18 | 20 |
| **Total** | **85** | **100** |

---

SCORE: 85/100

DIMENSIONS:
- Background & Goals: 19/20
- Flow Diagrams: 16/20
- Functional Specs: 16/20
- User Stories: 16/20
- Scope Clarity: 18/20

ATTACKS:
1. [User Stories]: Code Reviewer persona has zero story coverage — the background names "代码审查者" as a target user with a specific workflow ("通过 PR 级别的增量提交审查测试代码"), but all 6 stories are "As a 开发者". Either add stories for the reviewer, or remove the persona from the background if it was aspirational rather than real.
2. [Functional Specs]: Three endpoints have unspecified validation rules — `PUT /teams/:id/main-items/:itemId` says "验证错误 → 422", `PUT /teams/:id` says "验证 → 422", and `PUT /admin/users/:userId` says "验证 → 422". For a test-coverage PRD, "validation error" without stating what is validated is unactionable. A test writer must know which fields and which invalid values trigger the 422.
3. [Flow Diagrams]: Mermaid diagram shows project execution order, not business logic flows — the text section describes rich multi-step flows (e.g., "创建主项 → 创建子项 → 追加进度 → 状态变更 → 归档"), but the diagram collapses each to a single box. The only decision point is a coverage check at the end. The diagram adds little information beyond what a task list would convey.
