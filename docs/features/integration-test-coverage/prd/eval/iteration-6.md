# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 6
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### Three elements (cause/target/people) — 7/7

All three elements are present and specific:

- **Cause**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" with a concrete incident reference (commit `1883499`, timezone and filter bugs traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3).
- **Target**: "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口" — exact deliverable count with methodology.
- **People**: Two named personas: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码").

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% to 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 <= 500 行"

All five have numeric or countable targets. No vague language.

### Background and goals logically consistent — 6/6

The goals directly address the stated problem. Both personas have traceable goals. The unit test gap goal ties back to background's mention of "permission_handler.go 完全无测试" and other specific gaps. No logical gaps.

---

## Dimension 2: Flow Diagrams — 18/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle), F2 (Item Pool), F3 (Team Management), F4 (Admin User Management), F5 (Views & Reports), plus an execution order diagram. All use valid `flowchart` syntax.

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths from start to end with explicit start `([...])` and end `([流程结束])` nodes. F5 uses `([验证完成])` as its terminal node. All are traversable from entry to exit.

### Decision points + error branches covered — 4/6

Four of five diagrams have decision diamonds with error branches:

- F1: `{状态是否终端?}` diamond with yes/no branches, plus 5 error termination nodes.
- F2: `{审查决策}` diamond with three branches (assign/convert/reject) plus rollback and error terminations.
- F4: Error terminations at each step (duplicate username 409, validation errors 422, self-disable 422, not-found 404).

**Deduction (-1):** F3's permission check at line 130 uses `member{{member 角色}}` which renders as a hexagon node, not a standard diamond decision node. Every other decision point across all diagrams uses `{...}` diamond syntax. This is a stylistic inconsistency and the hexagon convention is not introduced or explained. A correct formulation would be `{member 角色?}`.

**Deduction (-1):** F5's diagram (lines 149-166) has zero error/exception branches. The entire diagram assumes every request succeeds. No branch for invalid parameters, wrong team ID, unauthorized access, or malformed queries. This is a test coverage PRD that plans to test error scenarios (the F5 spec table lists validation errors for F5 endpoints), yet the flow diagram only shows the happy path. A test writer using this diagram as guidance would miss error-case test planning for F5.

---

## Dimension 3: Functional Specs — 17/20

### Tables complete — 7/7

Seven spec tables with full coverage:
- F1: 17 rows x 5 columns — all cells populated.
- F2: 6 rows x 5 columns — all cells populated.
- F3: 9 rows x 5 columns — all cells populated.
- F4: 6 rows x 5 columns — all cells populated.
- F5: 6 rows x 3 columns (adapted for read-only) — all cells populated.
- F6: 6 rows (unit test gaps) — all cells populated.
- F7: 10 rows (shared helpers) — all cells populated with signatures and source files.

No empty cells. F5's adapted column structure (Happy Path / Empty Data / Format Validation) is appropriate for read-only endpoints.

### Field descriptions clear — 6/7

Most entries are specific with concrete triggers. Well-specified examples:
- `PUT /admin/users/:userId`: "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422" — field names with character limits and cross-entity validation.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422" — field names with character limits.
- `POST /teams/:id/item-pool/:poolId/assign`: "无效主项 → 回滚" with cascade effect "已处理 → 409".

**Deduction (-1):** Several entries use terse shorthand that requires inference:
- `GET /teams/:id/main-items/:itemId` (line 187): "错误团队 → 403" — "错误团队" is vague. What constitutes a "wrong team"? A team the user is not a member of? A nonexistent team? A team from another context?
- `GET /teams/:id/sub-items/:subId` (line 194): "错误团队 → 403" — same vague phrasing.
- `PUT /teams/:id/sub-items/:subId/assignee` (line 198): "非成员 → 403" — "非成员" is ambiguous. Non-member of the team? Non-member as the assignee? The assignee must be a team member, but the table does not specify whose membership is being checked.
- `PATCH /teams/:id/progress/:recordId/completion` (line 201): "修正最新 → 同步子项" and "修正非最新 → 不级联" — the word "修正" is used without defining what distinguishes a "修正" operation from a regular progress update.

### Validation rules explicit — 4/6

Many validation rules are specific. However:

**Deduction (-1):** Several F1 entries use terse shorthand without examples:
- `PUT /teams/:id/main-items/:itemId/status` (line 189): "无效转换 如 new→completed → 422" — gives one example but does not enumerate the full set of invalid transitions. A test writer must consult the state machine definition elsewhere.
- `PUT /teams/:id/sub-items/:subId/status` (line 196): "无效转换 如 new→completed → 422" — same issue, and unlike the MainItem status entry at line 189 which mentions cascade effects ("终端状态：子项自动完成"), this entry only mentions "终端级联：主项重算" without specifying what the cascade does concretely.

**Deduction (-1):** Permission columns across F2-F4 frequently use bare arrow notation "→ 403" without specifying what role or condition triggers the rejection:
- F2 rows for assign/convert/reject (lines 210-212): "→ 403" with no role specified. A test writer must infer from context (F3 diagram shows member role check) what role to use.
- F3 `GET /teams/:id/search-users` (line 223): "→ 403" — which role is denied? From context it appears to be non-members, but this is not stated.
- F4 `GET /admin/users` (line 232): "→ 403" — Story 4 AC #5 clarifies this is "普通用户", but the spec table itself does not specify.

---

## Dimension 4: User Stories — 15/20

### Coverage: one story per target user — 5/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 — covers F1-F6 and F7 helper extraction.
- **Code Reviewer**: Story 7 — addresses PR organization, naming conventions, incremental review.

**Deduction (-2):** F7 (Shared Test Helpers) is a named in-scope deliverable with its own 10-row spec table, its own execution-order slot (F1 -> F7 -> F2...), and its own PR. Story 6 covers helper extraction but is framed as a byproduct of F1 writing: "从现有集成测试和 F1 编写过程中提取 10 个共享辅助函数". The story's scope conflates two distinct activities: (a) extracting existing helpers from `auth_isolation_test.go` and `progress_completion_test.go`, and (b) creating two new helpers (`createTeamWithMembers`, `createMainItem`) that do not yet exist. The AC does not reference the F7 spec table by name and only names 3 of 8 existing helpers explicitly ("`setupTestDB`/`setupTestRouter`/`loginAs` 等 8 个现有辅助函数"). A developer reading only the stories would not know the full enumeration of all 10 functions without cross-referencing the spec table.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the format precisely. Each "I want" is concrete with enumerated steps or named methods:

- Story 1: Exact lifecycle steps enumerated ("MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive").
- Story 6: Specifies file name (`helpers.go`) and function count (10).
- Story 7: Includes naming convention example (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`).
- Story 8: Names exact methods to test.

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 3/6

Every story has ACs in Given/When/Then format. Total: ~30 acceptance criteria across 8 stories.

**Deduction (-2):** Story 7's ACs contain implementation prescriptions rather than observable behavior:
- AC #3: "Given F7 helpers PR 提交, When 审查 `backend/tests/integration/helpers.go`, Then 文件包含 F7 规格表所列的全部辅助函数，每个函数带 GoDoc 注释说明用途，且现有测试文件中的重复定义已删除" — mandates specific file path, GoDoc annotation format, and deletion of existing duplicates. These are implementation instructions ("how"), not acceptance criteria ("what outcome").
- AC #4: "Given 6 个测试 PR + 1 个 helpers PR 全部合并, When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 以 `test(<domain>):` 前缀" — mandates commit message format. This is a contribution guideline/process rule, not a product acceptance criterion.

**Deduction (-1):** Story 6 AC #1 says "从中提取 `setupTestDB`/`setupTestRouter`/`loginAs` 等 8 个现有辅助函数" — the "等 8 个" phrasing implies the AC lists all 8 functions but only names 3 explicitly. A test writer cannot verify completeness from this AC alone without cross-referencing the F7 spec table. AC should enumerate all function names or explicitly reference "all functions listed in the F7 spec table".

---

## Dimension 5: Scope Clarity — 18/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:
- F1: 17 endpoints (enumerated flow: MainItem CRUD -> SubItem -> Progress -> Status -> Archive)
- F2: 6 endpoints (enumerated flow: Submit -> Assign/Convert/Reject)
- F3: 9 endpoints (enumerated flow: CRUD + member management + role change)
- F4: 6 endpoints (enumerated flow: User CRUD + status toggle + team list)
- F5: 6 endpoints (enumerated: Weekly/Gantt/Table/CSV/Report Preview/Export)
- F6: 6 gaps (named: permission_handler, ConvertToMain, UpdateTeam, 3x GetByBizKey)
- F7: shared helpers (with reference to F7 spec table, 10 functions enumerated)

Each has exact counts and named contents.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:
1. Front-end test changes — "前端测试套件已覆盖组件和 E2E 流程"
2. Performance/load testing — named without further explanation
3. E2E browser testing — "独立工作流"
4. New features or bug fixes — "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 4/6

The seven scope items (F1-F7) map to seven functional spec tables and eight user stories.

**Deduction (-2):** F7 (Shared Test Helpers) is a named in-scope item with a 10-row spec table and its own PR, but no user story adequately covers the full scope of F7's creation work from the Developer persona. Story 6 partially covers helper extraction but (a) conflates extraction of existing helpers with creation of new ones, (b) does not reference the F7 spec table by name, and (c) only enumerates 3 of 10 functions in its AC. This is a three-way consistency gap: scope lists F7, functional specs detail F7 (10 rows with signatures), but user stories do not fully cover F7's deliverable scope.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 18 | 20 |
| Functional Specs | 17 | 20 |
| User Stories | 15 | 20 |
| Scope Clarity | 18 | 20 |
| **Total** | **88** | **100** |

---

SCORE: 88/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 18/20
- Functional Specs: 17/20
- User Stories: 15/20
- Scope Clarity: 18/20

ATTACKS:
1. [User Stories + Scope Clarity]: F7 (Shared Test Helpers) is a named in-scope deliverable with a 10-row spec table listing every function signature, its own execution-order slot, and its own PR — yet Story 6 only names 3 of 10 functions explicitly and does not reference the F7 spec table by name. This is a three-way consistency gap between scope/functional-specs/user-stories that has persisted since iteration 3. Story 6's AC #1 should enumerate all 10 function names or explicitly state "all functions listed in the F7 spec table".
2. [Flow Diagrams]: F5's diagram has zero error/exception branches. The entire Views & Reports flow assumes every request succeeds — no branch for invalid parameters, wrong team ID, unauthorized access, or malformed queries. This is a test coverage PRD that explicitly plans to test error scenarios (the F5 spec table lists validation error columns), yet the flow diagram provides no guidance for error-case test planning.
3. [Functional Specs]: Permission columns across F2-F4 use bare arrow notation "→ 403" without specifying which role or condition triggers the rejection. F2's assign/convert/reject rows (lines 210-212) all say just "→ 403" with no role specified. F3's `GET /teams/:id/search-users` (line 223) says "→ 403" without clarifying which role is denied. A test writer must infer from Story 3's AC #6 or the F3 diagram what role to use, breaking the spec table's self-contained usability.
