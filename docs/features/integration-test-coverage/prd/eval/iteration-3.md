# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 3
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### 三要素 (原因/对象/人员) — 7/7

All three elements are present and specific:

- **原因**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" — concrete numbers. Cited incident: commit `1883499` introduced timezone and filter bugs that escaped to manual testing, with a traceable reference to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3.
- **对象**: "为 36 个未测试端点编写端到端集成测试，按用户流程组织...同时补全 6 个单元测试缺口" — clear scope.
- **人员**: Two personas with distinct responsibilities: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码").

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% → 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: 7 PRs, each <= 500 lines

All have numeric targets. The PR target specifically addresses the Code Reviewer persona.

### Background and goals logically consistent — 6/6

The goals directly address the stated problem (36 uncovered endpoints). Both personas have traceable goals: Developer gets coverage/count/speed targets; Code Reviewer gets the PR organization target. The unit test gap target ties back to the background's mention of "permission_handler.go 完全无测试" and other specific gaps.

---

## Dimension 2: Flow Diagrams — 17/20

### Mermaid diagram exists — 7/7

Five Mermaid flowcharts present:
- F1: Item Lifecycle (lines 80-96)
- F2: Item Pool (lines 100-112)
- F3: Team Management (lines 116-130)
- F4: Admin User Management (lines 133-144)
- Execution order diagram (lines 153-160)

All use valid `flowchart` syntax. F3 and F4 are new additions since iteration 2.

### Main path complete (start → end) — 7/7

All four domain flow diagrams show complete happy paths:

- **F1**: PM 创建 MainItem → 创建 SubItem → 追加 Progress → Status 变更 → 状态终端? → 归档 → 流程结束
- **F2**: PM 提交池项 → 审查决策 → (assign/convert/reject) → 流程结束
- **F3**: 用户创建团队 → 邀请成员 → 变更角色 → 移除成员 → 解散团队 → 流程结束
- **F4**: SuperAdmin 创建用户 → 编辑用户信息 → 切换状态 → 流程结束

All start and end with explicit terminal nodes.

### Decision points + error branches covered — 3/6

F1, F2, F3, and F4 all have decision diamonds and error branches. F1 has `状态是否终端?` diamond. F2 has `审查决策` diamond with three branches. F3 has `member{{member 角色}}` shape and error terminations. F4 has validation error terminations.

**Deduction (-3):** F5 (Views & Reports) has no Mermaid diagram. The text at lines 147-149 argues "6 个端点均为只读查询，无状态变更、无决策分支。不适用包含决策菱形的流程图。" This is a weak justification. Even read-only endpoints have test flow logic: seed data → request with parameters → assert response structure → assert empty data case. The F5 table itself has three columns (Happy Path / Empty Data / Format Validation) that describe distinct test scenarios — these are decision points in the test flow. A PRD that provides detailed Mermaid diagrams for 4 of 5 features but omits the 5th with a hand-waving excuse creates an inconsistency. If a flow diagram is required for every in-scope feature, F5 should have one; if read-only flows genuinely don't need diagrams, the rubric should say so rather than leaving it as an exception.

Furthermore, the F3 diagram at line 130 uses `member{{member 角色}}` — this is a hexagon node, not a standard diamond decision node. While it conveys the permission check, it bypasses the standard flowchart convention of using diamond shapes for decision points.

---

## Dimension 3: Functional Specs — 18/20

### Tables complete — 7/7

Six test matrices present:
- F1: 17 rows (17 endpoints)
- F2: 6 rows (6 endpoints)
- F3: 9 rows (9 endpoints)
- F4: 6 rows (6 endpoints)
- F5: 6 rows (6 endpoints, adapted column structure)
- F6: 6 rows (unit gaps)
- F7: 10 rows (shared helpers with signatures)

Every cell is populated with specific expected behavior. F5 appropriately uses a different column structure (Happy Path / Empty Data / Format Validation) suited to read-only endpoints. F7 helper table includes source files, Go function signatures, and usage descriptions.

### Field descriptions clear — 6/7

Most endpoint entries are specific about expected behavior. Examples of well-specified entries:

- `POST /teams/:id/main-items`: "缺标题/无效优先级/无效日期 → 422" — three specific validation triggers.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422"
- `PUT /admin/users/:userId`: "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422"

**Deduction (-1):** Abbreviated paths persist throughout. Entries like `PUT /.../subId` (line 178), `GET /.../search-users` (line 206), `POST /.../members` (line 207), `DELETE /.../members/:userId` (line 208), `PUT /.../members/:userId/role` (line 209), `POST /.../poolId/assign` (line 193), `POST /.../poolId/convert-to-main` (line 194), `POST /.../poolId/reject` (line 195) all use `...` abbreviation. While context makes these inferable, a test writer should not need to reconstruct full URL paths from abbreviated notation. The first occurrence of each endpoint group should use the full path, with abbreviations only for subsequent references. F1's first two rows use full paths (`POST /teams/:id/main-items`, `GET /teams/:id/main-items`), then switches to abbreviated `PUT /.../status` — inconsistent.

### Validation rules explicit — 5/6

The iteration-2 vague entries have been fixed:
- `PUT /.../subId` now says "`assigneeKey` 无法解析为有效 bizKey → 422" — specific.
- `POST /.../item-pool` now says "缺 title / title 超 100 字符 → 422" — specific.

**Deduction (-1):** `POST /.../subId/progress` at line 182 says "回退 → 422" for validation error. "回退" (regression/decrease) is somewhat specific — it means completion value lower than the previous record — but the entry does not specify what "回退" means quantitatively. Compare with Story 1 AC: "Given 一个完成度为 0% 的 SubItem, When PM 追加 Progress (completion=60), Then SubItem 完成度更新为 60" — that AC gives concrete numbers. The test matrix entry should similarly specify: e.g., "completion < 上次记录值 → 422". The Mermaid diagram at line 88 is more specific ("回退 completion < 上次 → 422") than the table entry.

Also, `PUT /.../subId/status` at line 179 says "无效 → 422" — "无效" alone is vague. The F1 equivalent at line 172 says "无效转换 → 422" which is slightly better but still doesn't specify which transitions are invalid. Compare with the Mermaid diagram at line 90: "无效转换 如 new→completed → 422" — the diagram gives an example but the table doesn't.

---

## Dimension 4: User Stories — 17/20

### Coverage: one story per target user — 6/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-5 (integration tests) and Story 7 (unit test gaps) — complete coverage.
- **Code Reviewer**: Story 6 — addresses PR organization, naming conventions, incremental review.

**Deduction (-1):** Story 7 (unit test gaps) covers F6 in the scope, but the scope now includes F7 (Shared Test Helpers) as a separate deliverable. F7 has a spec table in prd-spec.md, is listed as an in-scope item, and appears in the execution order diagram — but no user story explicitly covers the helper extraction work. Story 6's AC #3 mentions helpers ("helpers 提供带类型签名的工厂函数"), but Story 6 is about the Code Reviewer reviewing PRs, not about a Developer creating the helpers. The helper extraction task (refactoring existing test code into shared utilities) is a distinct work item that deserves its own story or at minimum explicit coverage within an existing story.

### Format correct (As a / I want / So that) — 7/7

All seven stories follow the format precisely. Each "I want" is concrete:

- Story 1: Exact lifecycle steps enumerated.
- Story 2: Exact pool flow steps enumerated.
- Story 6: Includes example naming convention (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`).
- Story 7: Names exact methods to test.

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 4/6

Every story has multiple ACs in Given/When/Then format. Total: 37 acceptance criteria across 7 stories.

**Deduction (-2):** Story 6 has four ACs, but three of them prescribe implementation details rather than observable outcomes:

- AC #2: "每个测试函数包含 3-10 步有序操作（如 Create→Assert→SubCreate→Assert→Progress→Assert）" — mandates exact step count range and implementation pattern. An AC should verify the result ("tests simulate real user flows as multi-step sequences"), not dictate implementation.
- AC #3: "helpers 提供带类型签名的工厂函数（如 `createTeamWithMembers(t, pmID, memberCount) uint`）" — specifies exact function signatures. This is a design specification, not acceptance criteria.
- AC #4: "git log 显示 7 个独立 commit，每个 commit 对应一个 Feature，commit message 以 `test(<domain>):` 前缀" — mandates commit message format and count. While verifiable, this is a process rule, not a behavior.

Only AC #1 ("所有测试函数以 `Test<ItemLifecycle|SubItem|Progress|Status|Archive>_<Scenario>` 命名，审查者可从函数名识别被测流程和场景") is a legitimate acceptance criterion — it describes a verifiable property of the output (test naming convention) without over-constraining implementation.

Additionally, Story 5 AC #1 has a precision issue: "Given 3 项创建+1 项完成于本周, When 请求周视图, Then stats 为 `{NEW:0, completed:1, inProgress:2, overdue:0}`". The stats object shows NEW:0, completed:1, inProgress:2 — that's 3 items total (1+2). But the Given says "3 项创建+1 项完成于本周" which could mean 4 items (3 created + 1 completed) or 3 items (of which 1 completed). The F5 test matrix at line 226 is clearer: "3 项创建+1 完成：`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`" — this reads as 3 items total, 1 completed, 2 in-progress. But the AC "3 项创建+1 项完成" is ambiguous. Is it 3+1=4 items, or 3 items of which 1 is completed?

---

## Dimension 5: Scope Clarity — 17/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables with exact endpoint/method counts:

- F1: 17 endpoints
- F2: 6 endpoints
- F3: 9 endpoints
- F4: 6 endpoints
- F5: 6 endpoints
- F6: 6 gaps (named)
- F7: shared helpers (with reference to F7 spec table)

Each is a specific feature with measurable scope.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:

1. 前端测试变更 — "前端测试套件已覆盖组件和 E2E 流程"
2. 性能/负载测试 — named
3. E2E 浏览器测试 — "独立工作流"
4. 新功能或 bug 修复 — "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 3/6

The seven scope items (F1-F7) map to:
- Seven functional spec tables (F1-F7) in prd-spec.md
- Seven user stories in prd-user-stories.md

**Deduction (-3):** Two consistency issues:

1. **PR count inconsistency.** The goals table at line 35 says "7 个独立 PR，每个 <= 500 行测试代码" — explicitly listing 7 PRs. Story 6 AC #4 says "7 个独立 commit" and "F1-F7 共 7 个 PR". The execution order at line 56 says "F1 → F7 → F2 → F3 → F4 → F5 → F6". This is 7 items. But the iteration-2 report mentioned 6 PRs. The document now consistently says 7 — this is resolved. However, the "推荐执行顺序" at line 56 lists F7 as the second item to implement, but F7's spec table at line 260 says `createTeamWithMembers` is "待提取" and `createMainItem` is "待提取" — these are extracted from F1. If F7 is implemented as a standalone PR before F2-F4, the PR would contain only functions already extracted from F1, with no new test code of its own. This means the "7 个独立 PR，每个 <= 500 行测试代码" target is misleading — F7's PR would contain helper code, not test code. The 500-line limit says "500 行测试代码", but helpers are not test code; they are test infrastructure. This is a minor but real inconsistency in the scope/goals alignment.

2. **F5 has no flow diagram despite being in scope.** F5 is a checked in-scope item, has a functional spec table, and has a user story — but the flow diagram section explicitly refuses to provide one ("不适用包含决策菱形的流程图"). If every in-scope feature should have consistent documentation across all sections, F5's missing diagram is a gap. If the document's position is that F5 genuinely doesn't need a diagram, this should be stated as a deliberate scoping decision in the Scope section, not buried as a footnote in the flow diagram section. As-is, the inconsistency between "all 5 flows described in 业务流程说明" (lines 68-74) and "only 4 flows get diagrams" creates confusion about whether F5 is fully specified.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 17 | 20 |
| Functional Specs | 18 | 20 |
| User Stories | 17 | 20 |
| Scope Clarity | 17 | 20 |
| **Total** | **89** | **100** |

---

SCORE: 89/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 17/20
- Functional Specs: 18/20
- User Stories: 17/20
- Scope Clarity: 17/20

ATTACKS:
1. [Flow Diagrams]: F5 (Views & Reports) lacks a Mermaid diagram despite being a checked in-scope feature with its own spec table and user story — the text at lines 147-149 dismisses this with "不适用包含决策菱形的流程图" but the F5 table itself defines three distinct test scenario types (Happy Path / Empty Data / Format Validation) that represent decision branches in the test flow. Every other in-scope flow feature gets a diagram; F5's exclusion is inconsistent, not justified.
2. [User Stories]: Story 6's ACs #2-#4 prescribe implementation details (step count ranges, function signatures, commit message prefixes) rather than observable outcomes — AC #2 mandates "3-10 步有序操作" which is an implementation constraint, not a testable behavior. AC #3 specifies exact Go function signatures like `createTeamWithMembers(t, pmID, memberCount) uint`. These are design specifications disguised as acceptance criteria.
3. [Scope Clarity]: The goal "7 个独立 PR，每个 <= 500 行测试代码" is inconsistent with F7's nature — the F7 spec table shows that all helpers are either extracted from existing code or "待提取" from F1; F7 contains zero test cases, only infrastructure helpers. Counting F7 as a PR of "测试代码" inflates the PR count and misrepresents the deliverable type.
