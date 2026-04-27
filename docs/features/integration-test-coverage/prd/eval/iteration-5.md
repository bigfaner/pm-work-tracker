# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 5
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### 三要素 (原因/对象/人员) — 7/7

All three elements are present and specific:

- **原因**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" — precise numbers, plus a concrete incident reference (commit `1883499`, timezone and filter bugs traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3).
- **对象**: "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口。" — exact deliverable with methodology.
- **人员**: Two named personas with distinct responsibilities: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码").

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% → 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 ≤ 500 行"

All five have numeric or countable targets. No vague language.

### Background and goals logically consistent — 6/6

The goals directly address the stated problem: 36 uncovered endpoints out of 54 total. Both personas have traceable goals — Developer gets coverage/count/speed targets; Code Reviewer gets the PR organization target. The unit test gap goal ties back to the background's mention of "permission_handler.go 完全无测试" and other specific gaps. No logical gaps detected.

---

## Dimension 2: Flow Diagrams — 18/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present:
- F1: Item Lifecycle (lines 80-96)
- F2: Item Pool (lines 100-112)
- F3: Team Management (lines 116-130)
- F4: Admin User Management (lines 133-144)
- F5: Views & Reports (lines 149-166)
- Execution order diagram (lines 170-176)

All use valid `flowchart` syntax.

### Main path complete (start → end) — 7/7

All five domain flow diagrams show complete happy paths from start to end:

- **F1**: PM 创建 MainItem → 创建 SubItem → 追加 Progress → Status 变更 → 状态终端? → 归档 → 流程结束
- **F2**: PM 提交池项 → 审查决策 → (assign/convert/reject) → 流程结束
- **F3**: 用户创建团队 → 邀请成员 → 变更角色 → 移除成员 → 解散团队 → 流程结束
- **F4**: SuperAdmin 创建用户 → 编辑用户信息 → 切换状态 → 流程结束
- **F5**: 准备种子数据 → 请求视图端点 → 断言内容/断言空数据 → 验证完成

All have explicit start and end nodes.

### Decision points + error branches covered — 4/6

All five diagrams have decision logic, but with quality variance:

- F1 has `{状态是否终端?}` diamond with yes/no branches plus multiple error termination nodes.
- F2 has `{审查决策}` diamond with three branches (assign/convert/reject) plus rollback branch.
- F3 has `{member 角色}}` — uses a hexagon `{{...}}` node, not a standard diamond `{...}`. While it conveys a permission check, it is stylistically inconsistent with every other decision node in the document.
- F4 has validation error terminations at each step.
- F5 has two diamonds (`{返回数据是否非空?}`, `{端点类型?}`).

**Deduction (-1):** F3's permission check at line 130 uses `member{{member 角色}}` (hexagon syntax) instead of a standard diamond decision node `{member 角色?}`. This is a stylistic inconsistency — every other decision point across all diagrams uses `{...}` diamond syntax. The hexagon convention is not introduced or explained.

**Deduction (-1):** F5's diagram (lines 149-166) has zero error/exception branches. The entire diagram assumes every request succeeds — no branch for invalid parameters, wrong team ID, unauthorized access, or malformed queries. This is a test coverage PRD that explicitly plans to test error scenarios (the spec tables list validation errors for F5 endpoints), yet the flow diagram only shows the happy path. A test writer using this diagram as guidance would miss error-case test planning for F5.

---

## Dimension 3: Functional Specs — 16/20

### Tables complete — 7/7

Seven spec tables with full coverage:
- F1: 17 rows (17 endpoints) x 5 columns — all cells populated.
- F2: 6 rows (6 endpoints) x 5 columns — all cells populated.
- F3: 9 rows (9 endpoints) x 5 columns — all cells populated.
- F4: 6 rows (6 endpoints) x 5 columns — all cells populated.
- F5: 6 rows (6 endpoints) x 3 columns (adapted for read-only) — all cells populated.
- F6: 6 rows (unit test gaps) — all cells populated.
- F7: 10 rows (shared helpers) — all cells populated with signatures and source files.

No empty cells. F5's adapted column structure (Happy Path / Empty Data / Format Validation) is appropriate for read-only endpoints.

### Field descriptions clear — 5/7

Most entries are specific. Well-specified examples:
- `POST /teams/:id/main-items`: "缺标题/无效优先级/无效日期 → 422" — three specific validation triggers.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422" — field names with character limits.
- `PUT /admin/users/:userId`: "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422".

**Deduction (-2):** Abbreviated endpoint paths persist throughout F1-F4. Entries like `PUT /.../subId` (line 195), `GET /.../search-users` (line 223), `POST /.../members` (line 224), `DELETE /.../members/:userId` (line 225), `PUT /.../members/:userId/role` (line 226), `POST /.../poolId/assign` (line 210), `POST /.../poolId/convert-to-main` (line 211), `POST /.../poolId/reject` (line 212), `PUT /.../userId/status` (line 236) all use `...` abbreviation. This forces a test writer to mentally reconstruct full URL paths. F1 inconsistently uses full paths for the first two rows (`POST /teams/:id/main-items`, `GET /teams/:id/main-items`) then switches to `PUT /.../status` mid-table. This issue was flagged in iterations 2, 3, and 4 and remains unfixed.

### Validation rules explicit — 4/6

Many validation rules are specific with concrete triggers. However:

**Deduction (-1):** `PUT /.../subId/status` at line 196 says "无效转换 如 new→completed → 422" — while an example is given, this is less specific than the F1 equivalent at line 189 which says "无效转换 如 new→completed → 422" and includes cascade effects. More critically, several entries use terse shorthand without examples:
- `GET /teams/:id/main-items/:itemId` (line 187): "错误团队 → 403" — "错误团队" is vague. What constitutes a "wrong team"? A team the user is not a member of? A nonexistent team? A team ID from another organization?
- `GET /teams/:id/sub-items/:subId` (line 194): "错误团队 → 403" — same vague phrasing.
- `PUT /teams/:id/sub-items/:subId/assignee` (line 198): "非成员 → 403" — "非成员" is ambiguous. Non-member of the team? Non-member as the assignee? The assignee must be a team member, but the table does not specify whose membership is being checked.

**Deduction (-1):** `POST /.../subId/progress` at line 199 says "回退 completion < 上次 → 422" — while the concept is defined in the Mermaid diagram (line 88: "回退 completion < 上次 → 422"), the table's terse "回退" without the diagram's `completion < 上次` clarification forces cross-referencing. Similarly, `PATCH /teams/:id/progress/:recordId/completion` at line 201 says "修正最新 → 同步子项" and "修正非最新 → 不级联" — the word "修正" is used without defining what constitutes a "修正" operation vs a regular progress update. A test writer needs to understand the semantic difference between append (POST) and amend (PATCH), which is not explained in the table or in the flow description.

---

## Dimension 4: User Stories — 14/20

### Coverage: one story per target user — 5/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-5 (integration tests) and Story 8 (unit test gaps) — complete coverage for F1-F6.
- **Code Reviewer**: Story 7 — addresses PR organization, naming conventions, incremental review.

**Deduction (-2):** F7 (Shared Test Helpers) is a distinct in-scope deliverable with its own spec table (10 helper functions with signatures), its own execution-order slot (F1 → F7 → F2...), and its own PR — yet no user story explicitly covers the Developer's work of extracting and creating the `helpers.go` file. Story 6 covers helper extraction from the Developer perspective but is framed narrowly: "I want to 从现有集成测试（...）和 F1 编写过程中提取 10 个共享辅助函数到独立 `helpers.go` 文件". However, Story 6's "So that" clause says "F2-F5 的测试编写可直接复用" — this describes the *purpose* but the story's AC only covers the extraction process and backward compatibility, not the full scope of F7's spec table (which includes 8 existing helpers + 2 new ones with specific signatures). The story does not reference the F7 spec table by name or enumerate which 10 functions must be extracted. A developer reading only the stories (not the spec table) would not know the full scope.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the format precisely. Each "I want" is concrete with enumerated steps or named methods:

- Story 1: Exact lifecycle steps enumerated ("MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive").
- Story 2: Exact pool flow steps enumerated.
- Story 6: Specifies file name (`helpers.go`) and function count (10).
- Story 7: Includes naming convention example (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`).
- Story 8: Names exact methods to test.

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 2/6

Every story has ACs in Given/When/Then format. Total: ~30 acceptance criteria across 8 stories.

**Deduction (-2):** Story 7's ACs contain implementation prescriptions rather than observable behavior:

- AC #3: "Given F7 helpers PR 提交, When 审查 `backend/tests/integration/helpers.go`, Then 文件包含 F7 规格表所列的全部辅助函数，每个函数带 GoDoc 注释说明用途，且现有测试文件中的重复定义已删除" — mandates specific file path, GoDoc annotation format, and deletion of existing duplicates. These are implementation instructions ("how"), not acceptance criteria ("what outcome"). An AC should verify the helpers are available and documented, not prescribe the annotation format.
- AC #4: "Given 6 个测试 PR + 1 个 helpers PR 全部合并, When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 以 `test(<domain>):` 前缀（helpers PR 以 `refactor(test):` 前缀）" — mandates commit message format. This is a contribution guideline / process rule, not a product acceptance criterion. Acceptance criteria describe what the delivered system does, not how the git history looks.

**Deduction (-2):** Story 5 AC #1 has a numeric ambiguity: "Given 3 项创建+1 项完成于本周" — this reads as 3 items created + 1 item completed = 4 total items. But the F5 spec table at line 243 says "3 项创建+1 完成：`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`" which sums to 3 items total (0 new + 1 completed + 2 in-progress = 3). The AC's phrasing "3 项创建+1 项完成于本周" is inconsistent with the spec table's "3 项创建+1 完成". A test writer following the story AC literally would create 4 items and get wrong stats.

Story 6's AC #1 says "从中提取 `setupTestDB`/`setupTestRouter`/`loginAs` 等 8 个现有辅助函数" — the "等 8 个" phrasing implies the AC lists all 8 functions but only names 3 explicitly. A test writer cannot verify completeness from this AC alone without cross-referencing the F7 spec table.

---

## Dimension 5: Scope Clarity — 17/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:
- F1: 17 endpoints (enumerated flow: MainItem CRUD → SubItem → Progress → Status → Archive)
- F2: 6 endpoints (enumerated flow: Submit → Assign/Convert/Reject)
- F3: 9 endpoints (enumerated flow: CRUD + 成员管理 + 角色变更)
- F4: 6 endpoints (enumerated flow: 用户 CRUD + 状态切换 + 团队列表)
- F5: 6 endpoints (enumerated: Weekly/Gantt/Table/CSV/Report Preview/Export)
- F6: 6 gaps (named: permission_handler, ConvertToMain, UpdateTeam, 3x GetByBizKey)
- F7: shared helpers (with reference to F7 spec table)

Each has exact counts and named contents.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:
1. 前端测试变更 — "前端测试套件已覆盖组件和 E2E 流程"
2. 性能/负载测试 — named without further explanation
3. E2E 浏览器测试 — "独立工作流"
4. 新功能或 bug 修复 — "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 3/6

The seven scope items (F1-F7) map to seven functional spec tables in prd-spec.md and eight user stories in prd-user-stories.md.

**Deduction (-2):** F7 (Shared Test Helpers) is a named in-scope item with a 10-row spec table and its own PR, but has no dedicated user story from the Developer persona. Story 6 partially covers helper extraction, but its AC does not reference the F7 spec table by name or enumerate all 10 functions. This is a three-way consistency failure: scope lists F7, functional specs detail F7 (10 rows), but user stories do not fully cover F7's creation work. This was flagged in iterations 3 and 4 and remains unaddressed.

**Deduction (-1):** Story 5 AC #1's "3 项创建+1 项完成于本周" (reading as 4 items) is inconsistent with the F5 spec table's "3 项创建+1 完成" (3 items). The scope section counts F5 as "6 个端点" which matches the spec table, but the story's ambiguous phrasing creates a discrepancy between the user story and the functional spec for the same deliverable.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 18 | 20 |
| Functional Specs | 16 | 20 |
| User Stories | 14 | 20 |
| Scope Clarity | 17 | 20 |
| **Total** | **85** | **100** |

---

SCORE: 85/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 18/20
- Functional Specs: 16/20
- User Stories: 14/20
- Scope Clarity: 17/20

ATTACKS:
1. [User Stories]: F7 (Shared Test Helpers) is a named in-scope deliverable with a 10-row spec table, its own execution-order slot, and its own PR — yet no user story adequately covers the Developer's extraction and creation work. Story 6 mentions extraction but does not reference the F7 spec table by name or enumerate all 10 functions, and its AC only names 3 of 8 existing helpers explicitly. This three-way consistency failure between scope/functional-specs/user-stories has persisted since iteration 3.
2. [User Stories]: Story 5 AC #1 says "Given 3 项创建+1 项完成于本周" which reads as 4 items total, while the F5 spec table at line 243 says "3 项创建+1 完成：stats: {NEW:0, completed:1, inProgress:2, overdue:0}" which sums to 3 items. A test writer following the story literally would create wrong seed data and get failing tests. This is a concrete inconsistency that produces incorrect test implementation.
3. [Functional Specs]: Abbreviated endpoint paths (`PUT /.../subId`, `GET /.../search-users`, `POST /.../poolId/assign`, etc.) force test writers to mentally reconstruct full URL paths from context — F1 even inconsistently uses full paths for the first two rows then switches to abbreviation mid-table. This issue was flagged in iterations 2, 3, and 4 and remains unfixed across four evaluation cycles, indicating a systematic neglect of this feedback.
