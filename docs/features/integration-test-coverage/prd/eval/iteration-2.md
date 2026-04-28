# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 2
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### 三要素 (原因/对象/人员) — 7/7

All three elements are present and specific:

- **原因**: 54 endpoints total, only 18 (33%) have integration tests, 36 completely uncovered. Concrete incident cited: commit `1883499` introduced timezone and filter bugs in `view_handler.go` and `view_service.go` that escaped to manual testing (referenced in `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3).
- **对象**: Write integration tests for 36 untested endpoints, organized by user flow. Close 6 unit test gaps.
- **人员**: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码") — each with distinct, specific responsibilities.

### Goals quantified — 7/7

Five quantified targets in a table:

1. Endpoint integration test coverage: 33% → 100%
2. New test cases: ≥ 150
3. Unit test gaps: 6/6 closed
4. Test suite execution time: < 150 seconds
5. PR reviewability: 6 independent PRs, each ≤ 500 lines

All five targets have specific numbers. The fifth target now addresses the Code Reviewer persona identified in the background, resolving the iteration-1 gap.

### Background and goals logically consistent — 6/6

Both user personas now have traceable goals: Developer → coverage/case count/execution time targets; Code Reviewer → PR organization target (6 PRs, ≤ 500 lines each, descriptive function names). The goals directly address the stated problem (36 uncovered endpoints) with measurable outcomes.

---

## Dimension 2: Flow Diagrams — 18/20

### Mermaid diagram exists — 7/7

Three Mermaid flowcharts are present:
- F1: Item Lifecycle (lines 79-95)
- F2: Item Pool (lines 99-111)
- Execution order diagram (lines 116-122)

All use valid `flowchart` syntax with proper node shapes.

### Main path complete (start → end) — 7/7

Both F1 and F2 diagrams show full happy paths from start to finish:

- **F1**: `PM 创建 MainItem` → `创建 SubItem` → `追加 Progress` → `Status 变更` → `状态是否终端?` → `归档` → `流程结束`. This is a complete user flow matching the text description at line 69.
- **F2**: `PM 提交池项` → `审查决策` → (assign/convert/reject) → `流程结束`. Complete happy path with all three review outcomes.

The execution order diagram additionally shows the full F1→F2→F3→F4→F5→F6 chain with helper extraction dependencies.

### Decision points + error branches covered — 4/6

Both F1 and F2 diagrams include decision diamonds and error branches:

- **F1**: Diamond `状态是否终端?` with two branches. Multiple error terminations: validation errors (A1, B1, C1, D1), permission denied (A2), archive rejection (F1).
- **F2**: Diamond `审查决策` with three branches. Rollback branch when MainItem doesn't exist. Validation error terminations (S1, J1).

**Deduction (-2):** Only F1 and F2 have detailed flow diagrams. F3 (Team Management), F4 (Admin User), and F5 (Views & Reports) have no Mermaid diagrams at all. The text at lines 71-73 describes flows for all five features, but only two are visualized. For a PRD that claims "业务流程图" as a section, covering 2 out of 5 flows is incomplete. F3's member lifecycle (invite → role change → remove → dissolve) and F4's admin flow (create → edit → toggle status) have non-trivial decision logic (PM removal protection, self-disable prevention) that would benefit from diagram coverage.

---

## Dimension 3: Functional Specs — 17/20

### Tables complete — 7/7

Five test matrices (F1-F5) and one unit gap table (F6) are present. F1 has 17 rows, F2 has 6, F3 has 9, F4 has 6, F5 has 6, and F6 has 6 rows. Every cell is filled with specific expected behavior — no empty cells, no "TBD" placeholders.

F5 uses a different column structure (Happy Path / Empty Data / Format Validation) instead of the standard 5-column format, which is appropriate for views/reports endpoints that have different test dimensions.

### Field descriptions clear — 6/7

Most endpoint entries are specific about expected behavior. Iteration-1 issues have been partially addressed:

- `PUT /teams/:id/main-items/:itemId`: Now says "终端状态项不可编辑 → 422；`assigneeKey` 非数字 → 422" — concrete validation rules.
- `PUT /teams/:id`: Now says "缺 name / name 超 100 字符 / description 超 500 字符 → 422" — three specific field-level rules.
- `PUT /admin/users/:userId`: Now says "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422" — four specific rules.

**Deduction (-1):** Remaining abbreviated path entries persist. `PUT /.../subId` (line 141) still has "验证 → 422" — what validation? What fields? What invalid values? This is the same vague pattern from iteration 1 that was not addressed. Similarly, `PUT /.../subId/status` says "无效 → 422" — invalid what? The main-item status endpoint at line 135 is more specific ("无效转换 如 new→completed → 422"), but the sub-item equivalent at line 142 just says "无效". Also, `POST /.../item-pool` says "验证/重复 → 422" — "验证" is vague; which fields are validated? Abbreviated paths like `GET /.../search-users`, `POST /.../members`, `DELETE /.../members/:userId`, `PUT /.../members/:userId/role` force the test writer to infer the full URL.

### Validation rules explicit — 4/6

Many validation rules are now concrete (the iteration-1 gaps in F3 and F4 are fixed). However:

**Deduction (-2):** Two entries remain vague:
- `PUT /.../subId` (line 141): "验证 → 422" — no fields, no invalid values specified.
- `POST /.../item-pool` (line 153): "验证/重复 → 422" — "验证" without specifying which fields are validated and what constitutes invalid input. Only "重复" is clear.

These are test specifications. A test writer seeing "验证 → 422" cannot write a validation test case without consulting the source code, which defeats the purpose of a PRD.

---

## Dimension 4: User Stories — 18/20

### Coverage: one story per target user — 6/7

The background defines two user types: **Developer** and **Code Reviewer**.

- **Developer**: Stories 1-5 and Story 7 — comprehensive coverage across all five flow domains and unit test gaps.
- **Code Reviewer**: Story 6 — "As a 代码审查者" addressing PR organization, test naming conventions, and incremental review workflow.

Iteration-1 gap fully addressed. Both personas now have story coverage.

**Deduction (-1):** Story 6 is more about process compliance (naming conventions, PR structure) than about user need. The "So that" clause is strong ("我能在每次 PR 审查中聚焦一个业务域的测试逻辑"), but the story itself describes the output format rather than a testable user scenario. The acceptance criteria are about naming patterns and commit message prefixes — these are quality standards, not user-facing behaviors. This is acceptable but borderline for a user story format.

### Format correct (As a / I want / So that) — 7/7

All seven stories follow the format precisely. Each "I want" clause is concrete and specific:

- Story 1: "验证 MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive 的完整生命周期" — exact flow steps.
- Story 6: "每个 Feature（F1-F6）作为独立 PR 提交，PR 内测试按用户流程组织且命名清晰（如 `TestItemLifecycle_CreateSubItem_TracksCompletionCascade`）" — includes example naming convention.
- Story 7: "为 permission_handler、ConvertToMain、UpdateTeam、GetByBizKey 等未测试方法补充单元测试" — names exact methods.

No vague actions like "manage", "handle", or "improve". Each is a concrete, testable action.

### AC per story (Given/When/Then) — 5/6

Every story has multiple ACs in Given/When/Then format:

- Story 1: 7 ACs
- Story 2: 6 ACs
- Story 3: 6 ACs
- Story 4: 5 ACs
- Story 5: 5 ACs
- Story 6: 4 ACs
- Story 7: 4 ACs

Total: 37 acceptance criteria across 7 stories. All use Given/When/Then. All include specific values (HTTP status codes, data values, naming patterns).

**Deduction (-1):** Story 6 AC #2 and AC #3 are unusually process-oriented for acceptance criteria: "每个测试函数包含 3-10 步有序操作" and "helpers 提供带类型签名的工厂函数（如 `createTeamWithMembers(t, pmID, memberCount) uint`）". These prescribe implementation details (exact step count range, function signature format) rather than observable outcomes. An AC should verify the *result* (e.g., "tests are organized as sequential user flows, not isolated endpoint calls"), not mandate the implementation approach. AC #4 ("git log 显示 6 个独立 commit") is a good example of a verifiable outcome.

---

## Dimension 5: Scope Clarity — 17/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:

- F1-F5: each names exact endpoint counts and test domains
- F6: "6 个缺口（permission_handler、ConvertToMain、UpdateTeam、3x GetByBizKey）"
- Shared test helpers (extracted from existing tests)

No vague items. Each is a specific feature with measurable scope.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with reasons:

1. "前端测试变更" — reason: "前端测试套件已覆盖组件和 E2E 流程"
2. "性能/负载测试" — named, no ambiguity
3. "E2E 浏览器测试" — reason: "独立工作流"
4. "新功能或 bug 修复" — reason: "本需求纯粹是测试覆盖"

All named, not implied by absence. Each has a brief rationale.

### Scope consistent with functional specs and user stories — 3/6

The six scope items (F1-F6) map to:
- Six functional spec tables (F1-F6) in prd-spec.md
- Seven user stories (Stories 1-6, Story 7) in prd-user-stories.md

**Deduction (-3):** Two consistency issues:

1. **Shared helpers have no spec or story.** The in-scope list includes "共享测试辅助函数（从现有集成测试中提取复用）" as a checked deliverable. No functional spec table describes what these helpers are, what functions they expose, what existing tests they are extracted from, or what their API surface looks like. No user story covers the helper extraction task. The Mermaid execution order diagram shows helpers as a dependency node, but they are never specified. This is the same traceability gap flagged in iteration 1, unaddressed.

2. **F5 (Views & Reports) has no flow diagram.** The scope section states F5 as an in-scope deliverable, the functional spec table exists, and Story 5 covers it — but the "业务流程图" section only diagrams F1 and F2. F5 is described in text at line 73 ("基于种子数据验证聚合统计、导出格式、空数据处理"), but this is a one-liner compared to the detailed Mermaid diagrams given to F1 and F2. The scope promises 6 endpoints for F5, yet no diagram shows the flow.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 18 | 20 |
| Functional Specs | 17 | 20 |
| User Stories | 18 | 20 |
| Scope Clarity | 17 | 20 |
| **Total** | **90** | **100** |

---

SCORE: 90/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 18/20
- Functional Specs: 17/20
- User Stories: 18/20
- Scope Clarity: 17/20

ATTACKS:
1. [Scope Clarity]: Shared test helpers are a checked in-scope deliverable with zero specification — the in-scope list includes "共享测试辅助函数（从现有集成测试中提取复用）" but no functional spec table, user story, or Mermaid diagram describes what these helpers are, what functions they expose, or what existing tests they are extracted from. This was flagged in iteration 1 and remains unaddressed. Either remove it from scope (if it is a byproduct, not a deliverable) or give it a spec section.
2. [Functional Specs]: Two endpoints still have vague validation rules — `PUT /.../subId` says "验证 → 422" without specifying which fields or invalid values, and `POST /.../item-pool` says "验证/重复 → 422" where "验证" is unspecified. For a test-coverage PRD, a test writer cannot write a validation test from "验证 → 422" without reading the source code, which defeats the PRD's purpose as an executable specification.
3. [Flow Diagrams]: Only F1 and F2 have Mermaid diagrams; F3, F4, and F5 are undocumented visually — the "业务流程图" section provides detailed flowcharts for 2 of 5 flow features. F3 (team management with member lifecycle, role protection) and F4 (admin with self-disable prevention) have non-trivial decision logic described in text but not diagrammed. The text at lines 71-73 gives one-liner descriptions for F3-F5, but these are not equivalent to the detailed decision-point diagrams provided for F1 and F2.
