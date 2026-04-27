# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 11
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### Three elements (cause/target/people) — 7/7

All three elements are present and specific:

- **Cause (原因):** "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" with a concrete, traceable incident: commit `1883499` introduced a timezone bug (`view_handler.go` rejecting valid week start dates) and a filter logic bug (`view_service.go` returning identical data for all weeks), both traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3. The evidence includes commit hash, file names, and bug tracker references.

- **Target (对象):** "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口" -- exact count (36 + 6) with methodology stated.

- **People (人员):** Two named personas with distinct roles: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码"). Each has a purpose statement tied to the deliverable.

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint integration test coverage: 33% -> 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Test suite execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 <= 500 行"

All five have numeric targets with units (%, count, seconds, lines). No vague language.

### Background and goals logically consistent — 6/6

Goals directly address the stated problem. 33%->100% maps to the 36 uncovered endpoints. The 6 unit gaps tie to named files. Both personas have traceable goals (Developer: CI regression protection via test coverage; Reviewer: incremental PRs for manageable review). The 150 test case target follows from ~36 endpoints x ~4 scenarios each (happy path, validation, permission, not-found).

---

## Dimension 2: Flow Diagrams — 20/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle), F2 (Item Pool), F3 (Team Management), F4 (Admin User Management), F5 (Views & Reports), plus an execution order dependency diagram. All use valid `flowchart TD` or `flowchart LR` syntax with proper node shapes (`([...])` for terminals, `{...}` for decisions, `[...]` for processes).

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths with explicit start `([...])` and end `([流程结束])` / `([验证完成])` nodes:

- F1: PM creates MainItem -> SubItem -> Progress -> Status -> Archive -> Done. Fully traversable.
- F2: Submit pool item -> Review decision -> Assign/Convert/Reject -> Done. Complete.
- F3: Create team -> Invite members -> Role change -> Remove -> Dissolve -> Done. Complete.
- F4: Create user -> Edit -> Toggle status -> Done. Complete.
- F5: Seed data -> Call endpoint -> Assert content -> Format branching -> Done. Complete.

### Decision points + error branches covered — 6/6

All five diagrams include decision diamonds and error/exception branches:

- **F1:** `{状态是否终端?}` diamond with yes/no branches. Five error termination nodes: validation errors (A1, B1, C1), permission denial (A2), invalid transition (D1), archive rejection (F1).
- **F2:** `{审查决策}` diamond with three branches (assign/convert/reject). Rollback path from failed assign back to review. Validation errors and missing-reason error terminations.
- **F3:** `{member 角色?}` diamond at line 130. Error terminations at every step: duplicate code 422, already-member 409, user-not-found 404, PM role unchangeable 403, PM irremovable 422, non-PM operations 403.
- **F4:** Error terminations: duplicate username 409, non-SuperAdmin 403, validation errors (displayName empty/>64, email >100, nonexistent teamKey) 422, self-disable 422, user-not-found 404.
- **F5:** Four decision diamonds: `{无 token / 非 SuperAdmin?}`, `{teamId 不存在?}`, `{返回数据是否非空?}`, `{端点类型?}` with 403/404 branches and content format branching.

---

## Dimension 3: Functional Specs — 18/20

### Tables complete — 7/7

Seven spec tables, all fully populated:

- F1: 17 rows x 5 columns (Happy Path / Validation / Permission / Not Found / Cascade)
- F2: 6 rows x 5 columns
- F3: 9 rows x 5 columns
- F4: 6 rows x 5 columns
- F5: 6 rows x 3 columns (adapted for read-only endpoints: Happy Path / Empty Data / Format)
- F6: 6 rows x 3 columns (File / Gap / Supplement)
- F7: 10 rows x 4 columns (Function / Source File / Signature / Purpose)

No empty cells. No placeholder text ("TBD", "TODO"). The F5 table adapts its columns to the read-only nature of views/reports endpoints rather than forcing the CRUD template.

### Field descriptions clear — 7/7

Entries are specific with concrete triggers, field names, and character limits. Examples:

- `PUT /admin/users/:userId` (line 241): "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 -> 422" -- field names with exact character limits and cross-entity validation.
- `PUT /teams/:id` (line 227): "缺 name / name 超 100 字符 / description 超 500 字符 -> 422" -- field names with character limits.
- `POST /teams/:id/item-pool/:poolId/assign` (line 216): "无效主项 -> 回滚" with cascade effect "已处理 -> 409" -- specifies rollback behavior on invalid reference.
- F5 `GET /teams/:id/views/weekly` (line 249): "3 项（其中 1 项 completed、2 项 in-progress）：`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`" -- exact expected JSON structure with values.

Permission columns consistently specify roles: "member 角色 -> 403", "非 PM -> 403", "非成员 -> 403", "非 SuperAdmin -> 403".

### Validation rules explicit — 4/6

Many validation rules are specific with concrete triggers and field-level detail. However:

**Deduction (-1):** Two status-transition entries use a reference-by-proxy pattern instead of enumerating the full invalid set:
- `PUT /teams/:id/main-items/:itemId/status` (line 195): "状态机不允许的转换（参考 `status/transition.go` 定义）-> 422"
- `PUT /teams/:id/sub-items/:subId/status` (line 202): "状态机不允许的转换（参考 `status/transition.go` 定义）-> 422"

While the file reference (`status/transition.go`) is a reasonable fallback, a PRD should be self-contained for the test writer. The spec does not enumerate which transitions are invalid, nor does it provide examples. A test writer must open the source file to determine test cases. The previous iteration had examples ("如 new->completed") which were removed; this is a regression -- the current version provides less information.

**Deduction (-1):** Three F1 endpoints have ambiguous or missing permission specifications:
- `POST /teams/:id/main-items/:itemId/sub-items` (line 198): Permission column shows "--" (no check). This is a POST endpoint creating a sub-item under a team's main item. The absence of any permission note is ambiguous -- can a member create sub-items? Can a non-team member? The "--" implies no check exists, but this is never confirmed.
- `GET /teams/:id/main-items/:itemId` (line 193): "非本团队成员 -> 403". The phrase "非本团队成员" is clearer than the previous "非成员访问他人团队" but still leaves ambiguity: is a non-member of the team requesting a valid team's item denied with 403, or only cross-team access?
- `GET /teams/:id/sub-items/:subId` (line 200): "非本团队成员 -> 403" -- same ambiguity.

---

## Dimension 4: User Stories — 16/20

### Coverage: one story per target user — 7/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 (seven stories) -- covers all seven in-scope features (F1-F7 helper extraction, F6 unit gaps).
- **Code Reviewer**: Story 7 -- addresses PR organization, naming conventions, and incremental review.

Both personas have at least one story. All seven in-scope features have corresponding stories. Story 6 covers F7 (shared helpers) which is the infrastructure PR, distinct from the test-flow stories.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the "As a / I want / So that" format precisely. Each "I want" clause is concrete with specific actions:

- Story 1: "通过集成测试验证 MainItem 创建 -> SubItem 创建 -> Progress 追加 -> Status 变更 -> Archive 的完整生命周期" -- exact lifecycle steps enumerated.
- Story 6: "从现有集成测试（`auth_isolation_test.go`、`progress_completion_test.go`）和 F1 编写过程中提取 10 个共享辅助函数到独立 `helpers.go` 文件" -- specifies source files, count, and target file.
- Story 7: "每个 Feature（F1-F6）作为独立测试 PR 提交，F7 helpers 作为独立基础设施 PR 提交" -- specifies PR structure.
- Story 8: "为 permission_handler、ConvertToMain、UpdateTeam、GetByBizKey 等未测试方法补充单元测试" -- names exact methods.

No vague actions like "manage", "handle", "improve" in any story.

### AC per story (Given/When/Then) — 2/6

Every story has ACs in Given/When/Then format. Approximately 30 acceptance criteria across 8 stories.

**Deduction (-2):** Story 7 (Code Reviewer persona) has acceptance criteria that conflate multiple verifiable outcomes or prescribe implementation details inconsistent with the persona:

- AC#1 (line 106): "Given 收到一个测试 PR, When 审查 diff, Then 测试函数名包含业务语义可识别被测场景，且同一流程的端点测试在相邻函数中可追踪用户操作序列" -- the Then clause contains two verifiable outcomes joined by "且" (AND): (a) function names are semantically identifiable, and (b) endpoint tests within a flow are in adjacent functions. These should be separate ACs because partial fulfillment (identifiable but not adjacent) is ambiguous as pass or fail.

- AC#2 (line 107): "Given 收到 F7 helpers PR, When 审查文件, Then 包含 F7 规格表所列全部 10 个辅助函数，且现有测试文件中的重复定义已删除" -- again two outcomes conjoined: (a) all 10 functions present, and (b) duplicate definitions removed from source files. The second condition ("重复定义已删除") is an implementation verification (checking that source files were cleaned up), not a reviewer acceptance criterion. A code reviewer verifies the helpers PR itself, not the state of other files after merge.

**Deduction (-2):** Story 7 AC#3 (line 108): "Given 全部 PR 合并, When 运行 `go test ./tests/integration/...`, Then 所有测试通过" -- this is a CI/build criterion, not a code reviewer's acceptance criterion. The persona is "代码审查者" (Code Reviewer), but verifying that all tests pass after merge is a CI outcome, not something a reviewer does during PR review. This belongs in a Definition of Done checklist, not in a user story AC for the Reviewer persona. It conflates the reviewer role with a CI/automation concern.

---

## Dimension 5: Scope Clarity — 20/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:

- F1: "17 个端点（MainItem CRUD -> SubItem -> Progress -> Status -> Archive）"
- F2: "6 个端点（Submit -> Assign/Convert/Reject）"
- F3: "9 个端点（CRUD + 成员管理 + 角色变更）"
- F4: "6 个端点（用户 CRUD + 状态切换 + 团队列表）"
- F5: "6 个端点（Weekly/Gantt/Table/CSV/Report Preview/Export）"
- F6: "6 个缺口（permission_handler、ConvertToMain、UpdateTeam、3x GetByBizKey）"
- F7: "从现有集成测试中提取复用（见 F7 规格表）"

Each has exact counts and named contents. No vague areas.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:

1. "前端测试变更（前端测试套件已覆盖组件和 E2E 流程）" -- named with justification.
2. "性能/负载测试" -- named explicitly.
3. "E2E 浏览器测试（独立工作流）" -- named with rationale.
4. "新功能或 bug 修复 -- 本需求纯粹是测试覆盖" -- named with scope boundary.

All four are named, not implied by absence. Each has a reason for exclusion.

### Scope consistent with functional specs and user stories — 6/6

The seven scope items (F1-F7) map one-to-one to seven functional spec tables (F1-F7) and eight user stories (Stories 1-8). Verification:

- F1 -> F1 spec table (17 endpoints) -> Story 1 (item lifecycle). Consistent.
- F2 -> F2 spec table (6 endpoints) -> Story 2 (item pool). Consistent.
- F3 -> F3 spec table (9 endpoints) -> Story 3 (team management). Consistent.
- F4 -> F4 spec table (6 endpoints) -> Story 4 (admin user). Consistent.
- F5 -> F5 spec table (6 endpoints) -> Story 5 (views & reports). Consistent.
- F6 -> F6 spec table (6 gaps) -> Story 8 (unit test gaps). Consistent.
- F7 -> F7 spec table (10 helpers) -> Story 6 (shared helpers). Consistent.
- Story 7 (code reviewer) spans F1-F7 as the review process story. No inconsistency.

No gaps or mismatches detected between scope, specs, and stories.

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
1. [Functional Specs]: Two status-transition endpoints -- `PUT /teams/:id/main-items/:itemId/status` (line 195) and `PUT /teams/:id/sub-items/:subId/status` (line 202) -- delegate validation rule enumeration to an external source file with "状态机不允许的转换（参考 `status/transition.go` 定义）→ 422" without providing even representative examples of invalid transitions. The previous iteration had example transitions ("如 new→completed") which were removed, making this a regression in specificity. A test writer must open `status/transition.go` to determine which transitions to test. To fix: add representative invalid transition examples back (e.g., "new→completed, completed→in-progress are invalid per transition.go") or inline the full transition table.
2. [Functional Specs]: Three F1 endpoints have ambiguous permission specifications. `POST /teams/:id/main-items/:itemId/sub-items` (line 198) shows "--" in the permission column without explaining whether member/non-member access is allowed. `GET /teams/:id/main-items/:itemId` (line 193) and `GET /teams/:id/sub-items/:subId` (line 200) use "非本团队成员 → 403" which is clearer than before but still does not specify the exact access policy: is a team member guaranteed access, and is any authenticated non-member denied? To fix: add a note like "member/PM 均可访问" or "仅团队内角色可访问" for the "--" entries, and for the "非本团队成员" entries clarify whether team membership is the sole access criterion.
3. [User Stories]: Story 7 (Code Reviewer persona) has structural AC problems. AC#1 (line 106) and AC#2 (line 107) each conflate two verifiable outcomes in a single Then clause connected by "且" (AND), making partial fulfillment ambiguous. AC#3 (line 108) prescribes running `go test ./tests/integration/...` and verifying all tests pass -- this is a CI/automation outcome, not something a code reviewer does during PR review. The persona mismatch means this AC tests the build system, not the reviewer's ability to review incrementally. To fix: split AC#1 and AC#2 into separate single-outcome ACs, and replace AC#3 with a reviewer-centric criterion (e.g., "Given all 7 PRs, When I review each one in sequence, Then each PR diff is <= 500 lines and covers exactly one F-domain").
