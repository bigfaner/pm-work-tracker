# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 10
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### Three elements (cause/target/people) — 7/7

All three elements are present and specific:

- **Cause**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" with a concrete incident -- commit `1883499`, timezone bug and filter logic bug in `view_handler.go` / `view_service.go`, traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3. Specific, verifiable evidence with commit hash and file references.
- **Target**: "为 36 个未测试端点编写端到端集成测试...同时补全 6 个单元测试缺口" -- exact count with methodology.
- **People**: Two named personas with roles: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码"). Each has a concrete purpose tied to the feature.

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% to 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 <= 500 行"

All five have numeric targets. No vague language.

### Background and goals logically consistent — 6/6

Goals directly address the stated problem. The 33%->100% coverage goal maps to the 36 uncovered endpoints. The 6 unit test gaps tie to the named files in the background. Both personas have traceable goals (Developer gets CI regression protection; Reviewer gets incremental PRs).

---

## Dimension 2: Flow Diagrams — 20/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle), F2 (Item Pool), F3 (Team Management), F4 (Admin User Management), F5 (Views & Reports), plus an execution order dependency diagram. All use valid `flowchart` syntax with proper node shapes and arrow types.

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths with explicit start `([...])` and end `([流程结束])` nodes:
- F1: PM creates MainItem -> SubItem -> Progress -> Status -> Archive -> Done. Fully traversable.
- F2: Submit -> Review decision -> Done. Complete.
- F3: Create team -> Invite -> Role change -> Remove -> Dissolve -> Done. Complete.
- F4: Create user -> Edit -> Toggle status -> Done. Complete.
- F5: Seed data -> Call endpoint -> Assert -> Done. Complete.

### Decision points + error branches covered — 6/6

All five diagrams include decision diamonds and error branches:
- **F1**: `{状态是否终端?}` diamond with yes/no branches, plus 5 error termination nodes (validation, permission, completion rollback, invalid transition).
- **F2**: `{审查决策}` diamond with three branches (assign/convert/reject), rollback path, and error terminations for validation and missing reason.
- **F3**: `{member 角色?}` diamond decision node at line 130, error terminations at every step (duplicate code 422, already-member 409, user-not-found 404, PM role unchangeable 403, PM irremovable 422, non-PM 403).
- **F4**: Error terminations at every step (duplicate username 409, non-SuperAdmin 403, validation errors 422, self-disable 422, user-not-found 404).
- **F5**: `{无 token / 非 SuperAdmin?}`, `{teamId 不存在?}`, `{返回数据是否非空?}`, and `{端点类型?}` decision diamonds with 403/404 branches and content format branching.

---

## Dimension 3: Functional Specs — 18/20

### Tables complete — 7/7

Seven spec tables with full coverage:
- F1: 17 rows x 5 columns -- all cells populated
- F2: 6 rows x 5 columns -- all cells populated
- F3: 9 rows x 5 columns -- all cells populated
- F4: 6 rows x 5 columns -- all cells populated
- F5: 6 rows x 3 columns (adapted for read-only endpoints) -- all cells populated
- F6: 6 rows (unit test gaps) -- all cells populated
- F7: 10 rows (shared helpers with signatures and source files) -- all cells populated

No empty cells, no placeholder text.

### Field descriptions clear — 7/7

Entries are specific with concrete triggers. Examples:
- `PUT /admin/users/:userId` (line 241): "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 -> 422" -- field names with character limits and cross-entity validation.
- `PUT /teams/:id` (line 227): "缺 name / name 超 100 字符 / description 超 500 字符 -> 422" -- field names with character limits.
- `POST /teams/:id/item-pool/:poolId/assign` (line 216): "无效主项 -> 回滚" with cascade effect "已处理 -> 409".

Permission columns specify roles in the vast majority of entries: "member 角色 -> 403", "非 PM -> 403", "非成员 -> 403", "非 SuperAdmin -> 403".

### Validation rules explicit — 4/6

Many validation rules are specific with concrete examples. However:

**Deduction (-1):** Two status-transition entries use "如" (example) phrasing without enumerating the full invalid set:
- `PUT /teams/:id/main-items/:itemId/status` (line 195): "无效转换（如 new→completed、completed→in-progress）→ 422" -- gives two examples but does not list the full set of invalid transitions. A test writer must consult the state machine definition elsewhere to know all cases.
- `PUT /teams/:id/sub-items/:subId/status` (line 202): "无效转换（如 new→completed、completed→in-progress）→ 422" -- same issue. The spec provides representative examples but not an exhaustive enumeration.

**Deduction (-1):** Three F1 entries have permission columns that lack specificity:
- `GET /teams/:id/main-items/:itemId` (line 193): "非成员访问他人团队 → 403" -- the phrasing "非成员访问他人团队" conflates two distinct concepts: non-membership and accessing another team's resources. It is unclear whether a non-member of Team A requesting Team A's item gets 403, or only when accessing Team B's items.
- `GET /teams/:id/sub-items/:subId` (line 200): "非成员访问他人团队 → 403" -- same ambiguity.
- `POST /teams/:id/main-items/:itemId/sub-items` (line 198): the permission column shows "--" (dash), meaning no permission check. But this is a POST endpoint creating a sub-item under a team -- there should be a clear statement of whether member roles can access it. The absence of an explicit "--" rationale leaves ambiguity.

---

## Dimension 4: User Stories — 16/20

### Coverage: one story per target user — 7/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 -- covers F1-F6 and F7 helper extraction. All seven in-scope features are covered.
- **Code Reviewer**: Story 7 -- addresses PR organization, naming conventions, incremental review.

Both personas have at least one story. All seven in-scope features have corresponding stories.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the format precisely. Each "I want" is concrete:
- Story 1: Exact lifecycle steps enumerated ("创建 -> SubItem 创建 -> Progress 追加 -> Status 变更 -> Archive")
- Story 6: Specifies file names and function names ("从现有集成测试...提取 10 个共享辅助函数到独立 helpers.go 文件")
- Story 7: Names specific PR structure ("每个 Feature（F1-F6）作为独立测试 PR 提交，F7 helpers 作为独立基础设施 PR")
- Story 8: Names exact methods ("permission_handler、ConvertToMain、UpdateTeam、GetByBizKey")

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 2/6

Every story has ACs in Given/When/Then format. Approximately 30 acceptance criteria across 8 stories.

**Deduction (-2):** Story 7's ACs #1 and #2 (lines 106-107) are written from the developer-submission perspective but the story persona is Code Reviewer ("As a 代码审查者"):
- AC#1 (line 106): "Given 收到一个测试 PR, When 审查 diff, Then 可识别被测流程和场景（函数名包含业务语义），可追踪完整的用户操作序列（同一流程的端点测试在相邻函数中）" -- this is written from the reviewer perspective, which is an improvement over prior iterations. However, the AC still conflates two verifiable outcomes into a single Then clause ("可识别被测流程和场景" AND "可追踪完整的用户操作序列"). Per AC best practices, each Then should verify a single observable outcome. This makes it ambiguous whether partial fulfillment (e.g., identifiable but not traceable) would pass or fail the AC.
- AC#2 (line 107): "Given 收到 F7 helpers PR, When 审查文件, Then 包含 F7 规格表所列全部 10 个辅助函数，且现有测试文件中的重复定义已删除" -- the AC mandates checking a specific file path and verifying deletion of duplicate definitions. This prescribes implementation details rather than stating the outcome (helpers are reusable, no test breakage).

**Deduction (-2):** Story 7's AC#3 (line 108) contains an implementation prescription rather than an observable outcome:
- AC#3: "Given 6 个测试 PR + 1 个 helpers PR 全部合并, When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 清晰标识测试域" -- mandates commit message format. This is a process rule, not a product acceptance criterion. A reviewer does not verify git log messages as an acceptance condition; this belongs in a definition-of-done checklist, not a user story AC.

---

## Dimension 5: Scope Clarity — 20/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:
- F1: 17 endpoints enumerated with exact names
- F2: 6 endpoints enumerated
- F3: 9 endpoints enumerated
- F4: 6 endpoints enumerated
- F5: 6 endpoints enumerated
- F6: 6 gaps named with file paths
- F7: shared helpers with 10 functions enumerated with full signatures

Each has exact counts and named contents.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:
1. Front-end test changes -- "前端测试套件已覆盖组件和 E2E 流程"
2. Performance/load testing -- named without further explanation
3. E2E browser testing -- "独立工作流"
4. New features or bug fixes -- "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 6/6

The seven scope items (F1-F7) map to seven functional spec tables and eight user stories. Story 6 correctly enumerates the helper function names matching the F7 spec table. Each in-scope feature has a corresponding spec table and at least one user story. No inconsistencies detected between scope, spec tables, and stories.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 20 | 20 |
| Functional Specs | 18 | 20 |
| User Stories | 16 | 20 |
| Scope Clarity | 20 | 20 |
| **Total** | **94** | **100** |

---

SCORE: 94/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 20/20
- Functional Specs: 18/20
- User Stories: 16/20
- Scope Clarity: 20/20

ATTACKS:
1. [Functional Specs]: Two status-transition endpoints -- `PUT /teams/:id/main-items/:itemId/status` (line 195) and `PUT /teams/:id/sub-items/:subId/status` (line 202) -- use "无效转换（如 new→completed、completed→in-progress）→ 422" with representative examples but never enumerate the full set of invalid transitions. A test writer must look up the state machine definition elsewhere to write complete test cases. Quote from line 195: "无效转换（如 new→completed、completed→in-progress）→ 422". To fix: enumerate all invalid transition pairs explicitly, or link to the canonical state machine definition with a reference like "see state machine table in [design doc]".
2. [Functional Specs]: The permission column for `POST /teams/:id/main-items/:itemId/sub-items` (line 198) shows "--" (no permission check) without explanation. Since this is a POST endpoint creating a sub-item under a team, the absence of an explicit statement about whether member roles can access it is ambiguous. Additionally, `GET /teams/:id/main-items/:itemId` (line 193) and `GET /teams/:id/sub-items/:subId` (line 200) use "非成员访问他人团队 → 403" which conflates non-membership with cross-team access. Quote from line 193: "非成员访问他人团队 → 403". To fix: split the permission condition into distinct cases (non-member of team -> 403) and add an explicit note for endpoints where member access is allowed.
3. [User Stories]: Story 7 (Code Reviewer persona) has 3 acceptance criteria that do not match the persona or prescribe implementation details. AC#2 (line 107) mandates checking specific file contents and duplicate deletion -- this is an implementation instruction, not a reviewer outcome. AC#3 (line 108) mandates checking git log commit message format -- this is a process rule, not a product acceptance criterion. AC#1 (line 106) conflates two verifiable outcomes in a single Then clause, making partial fulfillment ambiguous. Quote from line 108: "When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 清晰标识测试域". To fix: rewrite ACs around reviewer verification outcomes (e.g., "When I review a test PR, Then I can identify which business flow and scenario each test covers from its function name alone") and move process rules to a definition-of-done checklist.
