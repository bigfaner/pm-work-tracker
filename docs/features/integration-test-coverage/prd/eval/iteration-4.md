# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 4
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### 三要素 (原因/对象/人员) — 7/7

All three elements are present and specific:

- **原因**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" — concrete numbers with a cited incident (commit `1883499`, timezone and filter bugs, traceable to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3).
- **对象**: "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口。" — precise deliverable.
- **人员**: Two personas with distinct responsibilities: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码").

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% → 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 ≤ 500 行"

All have numeric targets. The PR target now correctly distinguishes test PRs from the helpers infrastructure PR.

### Background and goals logically consistent — 6/6

The goals directly address the stated problem (36 uncovered endpoints out of 54). Both personas have traceable goals: Developer gets coverage/count/speed targets; Code Reviewer gets the PR organization target. The unit test gap target ties back to the background's mention of "permission_handler.go 完全无测试" and other specific gaps.

---

## Dimension 2: Flow Diagrams — 19/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present:
- F1: Item Lifecycle (lines 80-96)
- F2: Item Pool (lines 100-112)
- F3: Team Management (lines 116-130)
- F4: Admin User Management (lines 133-144)
- F5: Views & Reports (lines 149-166) — **new since iteration 3**
- Execution order diagram (lines 170-176)

All use valid `flowchart` syntax.

### Main path complete (start → end) — 7/7

All five domain flow diagrams show complete happy paths:

- **F1**: PM 创建 MainItem → 创建 SubItem → 追加 Progress → Status 变更 → 状态终端? → 归档 → 流程结束
- **F2**: PM 提交池项 → 审查决策 → (assign/convert/reject) → 流程结束
- **F3**: 用户创建团队 → 邀请成员 → 变更角色 → 移除成员 → 解散团队 → 流程结束
- **F4**: SuperAdmin 创建用户 → 编辑用户信息 → 切换状态 → 流程结束
- **F5**: 准备种子数据 → 请求视图端点 → 断言内容/断言空数据 → 验证完成

All start and end with explicit nodes. F5 uses `[准备种子数据]` as start and `([验证完成])` as end.

### Decision points + error branches covered — 5/6

All five diagrams now have decision diamonds and branching logic:

- F1 has `状态是否终端?` diamond with yes/no branches.
- F2 has `审查决策` diamond with three branches (assign/convert/reject).
- F3 has `member{{member 角色}}` hexagon with permission denial branch.
- F4 has validation error terminations at each step.
- F5 has `返回数据是否非空?` diamond and `端点类型?` diamond — a clear improvement from iteration 3.

**Deduction (-1):** F3's permission check at line 130 uses `member{{member 角色}}` — a hexagon node, not a standard diamond decision node. While it conveys the permission check, it is stylistically inconsistent with every other decision node in the document, which uses `{...}` diamond syntax. Additionally, F1 and F2 each have explicit error termination nodes (e.g., `A1[验证错误终止]`, `A2[权限拒绝终止]`), but F5 has no error/exception termination nodes at all. The F5 diagram assumes all requests succeed — no branch for invalid parameters, wrong team, or auth failures. This is a gap for a test flow diagram that should cover error scenarios as well as happy paths.

---

## Dimension 3: Functional Specs — 17/20

### Tables complete — 7/7

Seven spec tables present:
- F1: 17 rows (17 endpoints) with 5 columns
- F2: 6 rows (6 endpoints) with 5 columns
- F3: 9 rows (9 endpoints) with 5 columns
- F4: 6 rows (6 endpoints) with 5 columns
- F5: 6 rows (6 endpoints) with adapted 3-column structure
- F6: 6 rows (unit test gaps)
- F7: 10 rows (shared helpers with signatures and source files)

Every cell is populated with specific expected behavior. F5's adapted column structure (Happy Path / Empty Data / Format Validation) is appropriate for read-only endpoints.

### Field descriptions clear — 5/7

Most endpoint entries are specific about expected behavior. Examples of well-specified entries:

- `POST /teams/:id/main-items`: "缺标题/无效优先级/无效日期 → 422" — three specific validation triggers.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422"
- `PUT /admin/users/:userId`: "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422"

**Deduction (-2):** Abbreviated paths persist throughout F1-F4. Entries like `PUT /.../subId` (line 195), `GET /.../search-users` (line 223), `POST /.../members` (line 224), `DELETE /.../members/:userId` (line 225), `PUT /.../members/:userId/role` (line 226), `POST /.../poolId/assign` (line 210), `POST /.../poolId/convert-to-main` (line 211), `POST /.../poolId/reject` (line 212), `PUT /.../userId/status` (line 236) all use `...` abbreviation. This issue was flagged in iteration 2 and iteration 3 and remains unfixed. A test writer should not need to reconstruct full URL paths from abbreviated notation. The first occurrence of each endpoint group should use the full path; abbreviations should only appear for subsequent references within the same logical group. The inconsistency is especially visible in F1 where the first two rows use full paths (`POST /teams/:id/main-items`, `GET /teams/:id/main-items`) and then switches to `PUT /.../status`.

### Validation rules explicit — 5/6

Most validation rules are specific with concrete triggers:

- `POST /teams/:id/main-items`: "缺标题/无效优先级/无效日期 → 422" — three specific triggers.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422"
- `POST /.../item-pool`: "缺 title / title 超 100 字符 → 422"

**Deduction (-1):** `PUT /.../subId/status` at line 196 says "无效 → 422" — "无效" alone is vague. The F1 equivalent `PUT /.../status` at line 189 says "无效转换 → 422" which is slightly better but still does not specify which transitions are invalid. The Mermaid diagram at line 90 provides an example ("无效转换 如 new→completed → 422") but the table does not carry this specificity. A test writer would need to cross-reference the diagram to understand what "无效" means.

Similarly, `POST /.../subId/progress` at line 199 says "回退 → 422". While "回退" is a specific business concept (completion value lower than previous), the table does not define it quantitatively. The Mermaid diagram at line 88 says "回退 completion < 上次 → 422" which is more explicit — the table should match that level of specificity.

---

## Dimension 4: User Stories — 16/20

### Coverage: one story per target user — 5/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-5 (integration tests) and Story 7 (unit test gaps) — complete coverage.
- **Code Reviewer**: Story 6 — addresses PR organization, naming conventions, incremental review.

**Deduction (-2):** F7 (Shared Test Helpers) is a distinct in-scope deliverable with its own spec table (10 helper functions), its own execution order slot (F1 → F7 → F2...), and its own PR — but no user story explicitly covers the helper extraction work from a Developer perspective. Story 6's AC #3 mentions verifying the helpers file during code review, but Story 6 is from the Code Reviewer persona. The act of extracting, refactoring, and creating the `helpers.go` file is Developer work that lacks a story. This was flagged in iteration 3 and remains unaddressed.

### Format correct (As a / I want / So that) — 7/7

All seven stories follow the format precisely. Each "I want" is concrete:

- Story 1: Exact lifecycle steps enumerated.
- Story 2: Exact pool flow steps enumerated.
- Story 6: Includes example naming convention (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`).
- Story 7: Names exact methods to test.

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 4/6

Every story has multiple ACs in Given/When/Then format. Total: ~25 acceptance criteria across 7 stories.

**Deduction (-2):** Story 6's ACs have been revised since iteration 3, but still contain implementation prescriptions:

- AC #3: "Given F7 helpers PR 提交, When 审查 `backend/tests/integration/helpers.go`, Then 文件包含 F7 规格表所列的全部辅助函数，每个函数带 GoDoc 注释说明用途，且现有测试文件中的重复定义已删除" — mandates specific file path, GoDoc annotation requirements, and deletion of existing duplicates. These are implementation instructions, not observable behavior. An AC should verify the outcome ("helpers are available and documented"), not prescribe file paths and annotation formats.

- AC #4: "Given 6 个测试 PR + 1 个 helpers PR 全部合并, When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 以 `test(<domain>):` 前缀（helpers PR 以 `refactor(test):` 前缀）" — mandates commit message format. While verifiable, this is a process rule (contribution guideline), not a product acceptance criterion. Acceptance criteria should describe what the delivered system does, not how the git history looks.

Additionally, Story 5 AC #1 has an ambiguity: "Given 3 项创建+1 项完成于本周" — this reads as 3+1=4 items, but the F5 spec table at line 243 says "3 项创建+1 完成：`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`" which sums to 3 items. The AC's phrasing is inconsistent with the spec table's intended meaning.

---

## Dimension 5: Scope Clarity — 18/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables with exact endpoint/method counts:

- F1: 17 endpoints
- F2: 6 endpoints
- F3: 9 endpoints
- F4: 6 endpoints
- F5: 6 endpoints
- F6: 6 gaps (named methods)
- F7: shared helpers (with reference to F7 spec table)

Each is a specific feature with measurable scope.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:

1. 前端测试变更 — "前端测试套件已覆盖组件和 E2E 流程"
2. 性能/负载测试 — named
3. E2E 浏览器测试 — "独立工作流"
4. 新功能或 bug 修复 — "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 4/6

The seven scope items (F1-F7) map to:
- Seven functional spec tables (F1-F7) in prd-spec.md
- Seven user stories in prd-user-stories.md

**Deduction (-2):** The goals table at line 35 now correctly says "6 个测试 PR + 1 个 helpers PR，每个 ≤ 500 行" and clarifies F7 is "helpers 基础设施 PR（非测试代码）". This resolves iteration-3's PR count inconsistency. However, two consistency gaps remain:

1. **F7 has no dedicated user story.** F7 is an in-scope item, has a spec table, has an execution-order slot, and has its own PR — but no story describes the Developer's work of extracting and creating the helpers. Story 6 (Code Reviewer persona) references F7 in AC #3, but only from the review perspective. The scope says F7 is in scope, the functional specs describe it in detail, but the user stories do not cover the creation of F7. This is a three-way consistency failure between scope/functional-specs/user-stories.

2. **Story 5 AC #1 ambiguity vs. F5 spec table.** The story says "3 项创建+1 项完成于本周" (which reads as 4 items), while the spec table at line 243 says "3 项创建+1 完成" (which reads as 3 items with 1 completed). The scope section counts F5 as "6 个端点" which matches the spec table. The story's ambiguous phrasing creates a mismatch between the story AC and the functional spec.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 19 | 20 |
| Functional Specs | 17 | 20 |
| User Stories | 16 | 20 |
| Scope Clarity | 18 | 20 |
| **Total** | **90** | **100** |

---

SCORE: 90/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 19/20
- Functional Specs: 17/20
- User Stories: 16/20
- Scope Clarity: 18/20

ATTACKS:
1. [User Stories]: F7 (Shared Test Helpers) is a named in-scope deliverable with a 10-row spec table and its own execution-order slot, yet has no user story covering the Developer's extraction/refactoring work — Story 6 AC #3 addresses F7 only from the Code Reviewer perspective, leaving the creation work without a story. This is a coverage gap that has persisted since iteration 3.
2. [Functional Specs]: Abbreviated endpoint paths (`PUT /.../subId`, `GET /.../search-users`, `POST /.../poolId/assign`, etc.) force test writers to reconstruct full URLs from context — this issue was flagged in both iteration 2 and iteration 3 and remains unfixed. F1 inconsistently uses full paths for the first two rows then switches to abbreviation.
3. [Functional Specs]: `PUT /.../subId/status` at line 196 says "无效 → 422" without specifying which status transitions are invalid, while the Mermaid diagram at line 90 provides a concrete example ("无效转换 如 new→completed → 422") — the table should carry the same specificity so test writers don't need to cross-reference diagrams to understand validation rules.
