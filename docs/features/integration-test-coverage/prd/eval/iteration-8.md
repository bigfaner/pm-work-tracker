# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 8
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### Three elements (cause/target/people) — 7/7

All three elements are present and specific:

- **Cause (原因)**: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" with a concrete incident reference — commit `1883499`, timezone bug in `view_handler.go` and filter logic bug in `view_service.go`, both traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3. This is specific, verifiable evidence.
- **Target (对象)**: "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口" — exact count with methodology distinction.
- **People (人员)**: Two named personas with role-specific needs: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码").

### Goals quantified — 7/7

Five quantified targets in a structured table:

1. Endpoint coverage: 33% to 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 <= 500 行"

All five have numeric targets. No vague language.

### Background and goals logically consistent — 6/6

The goals directly address the stated problem. The 33%->100% coverage goal maps to the 36 uncovered endpoints named in background. The 6 unit test gaps tie back to "permission_handler.go 完全无测试、ConvertToMain/UpdateTeam/GetByBizKey 方法缺少覆盖". Both personas have traceable goals (developer has F1-F6 + F7 stories; reviewer has Story 7). No logical gaps.

---

## Dimension 2: Flow Diagrams — 20/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle), F2 (Item Pool), F3 (Team Management), F4 (Admin User Management), F5 (Views & Reports), plus an execution order dependency diagram. All use valid `flowchart` syntax.

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths with explicit start `([...])` and end `([流程结束])` nodes. F1: PM creates MainItem -> SubItem -> Progress -> Status -> Archive -> Done. F2: Submit -> Review decision (assign/convert/reject) -> Done. F3: Create team -> Invite -> Role change -> Remove -> Dissolve -> Done. F4: Create user -> Edit -> Toggle status -> Done. F5: Seed data -> Call endpoint -> Assert -> Done. All are fully traversable.

### Decision points + error branches covered — 6/6

All five diagrams include decision diamonds and error branches:

- **F1**: `{状态是否终端?}` diamond with yes/no branches, plus 5 error termination nodes (validation, permission, completion rollback, invalid transition).
- **F2**: `{审查决策}` diamond with three branches (assign/convert/reject), rollback path, and error terminations (validation 422, missing reason 422).
- **F3**: Error terminations at every step (duplicate code 422, already-member 409, user-not-found 404, PM role unchangeable 403, PM irremovable 422, non-PM 403), plus a `{member 角色?}` diamond decision node at line 130 checking member permission.
- **F4**: Error terminations at every step (duplicate username 409, non-SuperAdmin 403, validation errors 422, self-disable 422, user-not-found 404).
- **F5**: Now includes `{无 token / 非 SuperAdmin?}` and `{teamId 不存在?}` decision diamonds at lines 158-160 with `Assert403` and `Assert404` branches. This resolves the prior iteration-6 and iteration-7 finding.

The prior iteration-7 deduction for F3's hexagon node is resolved — line 130 now correctly uses `{member 角色?}` diamond syntax.

---

## Dimension 3: Functional Specs — 18/20

### Tables complete — 7/7

Seven spec tables with full coverage:
- F1: 17 rows x 5 columns — all cells populated
- F2: 6 rows x 5 columns — all cells populated
- F3: 9 rows x 5 columns — all cells populated
- F4: 6 rows x 5 columns — all cells populated
- F5: 6 rows x 3 columns (adapted for read-only) — all cells populated
- F6: 6 rows (unit test gaps) — all cells populated
- F7: 10 rows (shared helpers with signatures and source files) — all cells populated

No empty cells, no placeholder text. All endpoint paths are fully qualified. F5's adapted column structure (Happy Path / Empty Data / Format Validation) is appropriate for read-only endpoints.

### Field descriptions clear — 7/7

Most entries are specific with concrete triggers. Examples:
- `PUT /admin/users/:userId` (line 241): "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422" — field names with character limits and cross-entity validation.
- `PUT /teams/:id` (line 227): "缺 name / name 超 100 字符 / description 超 500 字符 → 422" — field names with character limits.
- `POST /teams/:id/item-pool/:poolId/assign` (line 216): "无效主项 → 回滚" with cascade effect "已处理 → 409".

Permission columns now specify roles in the vast majority of entries: "member 角色 → 403" (F1 line 191, F2 lines 213/216-218, F3 lines 229-232), "非 PM → 403" (F3 lines 227-228), "非成员 → 403" (F3 line 226, F1 line 204), "非 SuperAdmin → 403" (F4 lines 238-239).

### Validation rules explicit — 4/6

Many validation rules are specific with concrete examples. However:

**Deduction (-1):** Two status-transition entries use "如" (example) phrasing without enumerating the full invalid set:
- `PUT /teams/:id/main-items/:itemId/status` (line 195): "无效转换 如 new→completed → 422" — gives one example but does not list the full set of invalid transitions. A test writer must consult the state machine definition elsewhere to know all cases.
- `PUT /teams/:id/sub-items/:subId/status` (line 202): "无效转换 如 new→completed → 422" — same issue, and the cascade description differs in specificity from the MainItem equivalent ("终端级联：主项重算" vs "终端状态：子项自动完成").

**Deduction (-1):** Three F1 entries still have permission columns that do not specify the role:
- `PUT /teams/:id/main-items/:itemId` (line 194): "→ 403" — no role or condition specified. The reader must infer this is the member role from the F1 diagram (line 84) or Story 1 AC#6.
- `PUT /teams/:id/main-items/:itemId/status` (line 195): "→ 403" — same bare arrow notation without role.
- `GET /teams/:id/main-items/:itemId` (line 193): "错误团队 → 403" — "错误团队" is ambiguous. Does this mean accessing a team the user is not a member of? Accessing a nonexistent team? The term is inferable but imprecise compared to the clearer "非成员 → 403" used elsewhere in F3.

---

## Dimension 4: User Stories — 16/20

### Coverage: one story per target user — 7/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 — covers F1-F6 and F7 helper extraction. All seven in-scope features are covered.
- **Code Reviewer**: Story 7 — addresses PR organization, naming conventions, incremental review.

Story 6's AC#1 (line 93) now correctly lists the 8 existing helper function names matching the F7 spec table: `setupTestDB`/`setupTestRouter`/`loginAs`/`signTokenWithClaims`/`seedProgressData`/`appendProgress`/`seedPoolData`/`seedReportData`. The prior naming conflict (5 mismatched names) is resolved.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the format precisely. Each "I want" is concrete with enumerated steps or named methods:

- Story 1: Exact lifecycle steps enumerated ("MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive")
- Story 6: Specifies file name (`helpers.go`) and enumerates all 10 function names explicitly
- Story 7: Includes naming convention example (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`)
- Story 8: Names exact methods to test

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 2/6

Every story has ACs in Given/When/Then format. Total: approximately 30 acceptance criteria across 8 stories. Story 5 AC#1 data (`{NEW:0, completed:1, inProgress:2, overdue:0}`) matches the F5 spec table at line 249 exactly.

**Deduction (-2):** Story 7's ACs contain implementation prescriptions rather than observable outcomes:
- AC#3 (line 108): "Given F7 helpers PR 提交, When 审查 `backend/tests/integration/helpers.go`, Then 文件包含 F7 规格表所列的全部辅助函数，且现有测试文件中的重复定义已删除" — the AC mandates checking a specific file path (`backend/tests/integration/helpers.go`) and verifying deletion of duplicate definitions. While this has improved from prior iterations (removing the GoDoc mandate), the AC still prescribes implementation details (which file to check, that duplicates must be deleted) rather than stating the outcome (helpers are reusable, no test breakage).
- AC#4 (line 109): "Given 6 个测试 PR + 1 个 helpers PR 全部合并, When 查看 git log, Then 每个 PR 对应一个 Feature，commit message 清晰标识测试域" — mandates commit message format ("清晰标识测试域"). This is a contribution guideline/process rule, not a product acceptance criterion. This AC has also improved from prior iterations (removing the `test(<domain>):` prefix mandate) but still prescribes how git history should look rather than what the reviewer needs to verify.

**Deduction (-2):** Story 7's AC#1 and AC#2 (lines 106-107) describe the reviewer's reading experience but frame ACs around the developer's submission structure rather than the reviewer's verification task:
- AC#1: "Given F1 的 17 个端点测试完成, When 提交 PR, Then 所有测试函数以 `Test<ItemLifecycle|SubItem|Progress|Status|Archive>_<Scenario>` 命名" — the Story is from the Code Reviewer perspective ("As a 代码审查者"), but this AC describes what the developer submits, not what the reviewer verifies. The reviewer cannot control test function naming; they can only accept or reject. A proper reviewer AC would say "When reviewing a test PR, Then I can identify the tested flow and scenario from the function name."
- AC#2: "Given 审查任一测试 PR 的 diff, When 检查测试结构, Then 测试函数按用户流程分组" — this is closer to a reviewer AC but still prescribes the structural organization ("按用户流程分组") rather than the reviewer outcome ("can trace a complete user operation sequence through the diff").

---

## Dimension 5: Scope Clarity — 20/20

### In-scope items are concrete deliverables — 7/7

All seven in-scope items are specific, countable deliverables:
- F1: 17 endpoints (enumerated flow: MainItem CRUD -> SubItem -> Progress -> Status -> Archive)
- F2: 6 endpoints (enumerated flow: Submit -> Assign/Convert/Reject)
- F3: 9 endpoints (enumerated flow: CRUD + member management + role change)
- F4: 6 endpoints (enumerated flow: User CRUD + status toggle + team list)
- F5: 6 endpoints (enumerated: Weekly/Gantt/Table/CSV/Report Preview/Export)
- F6: 6 gaps (named: permission_handler, ConvertToMain, UpdateTeam, 3x GetByBizKey)
- F7: shared helpers (10 functions enumerated with signatures in spec table)

Each has exact counts and named contents.

### Out-of-scope explicitly lists deferred items — 7/7

Four items explicitly deferred with brief rationales:
1. Front-end test changes — "前端测试套件已覆盖组件和 E2E 流程"
2. Performance/load testing — named without further explanation
3. E2E browser testing — "独立工作流"
4. New features or bug fixes — "本需求纯粹是测试覆盖"

All named and reasoned.

### Scope consistent with functional specs and user stories — 6/6

The seven scope items (F1-F7) map to seven functional spec tables and eight user stories. Story 6 now correctly enumerates all 10 helper function names matching the F7 spec table. The prior three-way consistency gap between scope/functional-specs/user-stories for F7 is resolved. Each in-scope feature has a corresponding spec table and at least one user story.

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
1. [User Stories]: Story 7 ACs #1 and #2 are written from the developer-submission perspective but the story persona is Code Reviewer ("As a 代码审查者"). AC#1 mandates test function naming conventions (`Test<ItemLifecycle|SubItem|Progress|Status|Archive>_<Scenario>`) — this is something the developer controls, not what the reviewer verifies. AC#3 prescribes checking a specific file path (`backend/tests/integration/helpers.go`) and AC#4 mandates commit message format. These are implementation instructions masquerading as acceptance criteria, not verifiable outcomes from the reviewer's perspective.
2. [Functional Specs]: Two F1 entries (lines 194-195) use bare "→ 403" without specifying which role or condition triggers the rejection. `PUT /teams/:id/main-items/:itemId` and `PUT /teams/:id/main-items/:itemId/status` both show only "→ 403" in the permission column. Every other permission column across F2-F4 specifies the role (e.g., "member 角色 → 403", "非 PM → 403", "非 SuperAdmin → 403"). A test writer cannot determine from the spec table alone what role to use for these two endpoints.
3. [Functional Specs]: Two status-transition endpoints (lines 195, 202) use "无效转换 如 new→completed → 422" — giving one example but not enumerating the full set of invalid transitions. A test writer must look up the state machine elsewhere to know all invalid cases to test. Additionally, the cascade descriptions are inconsistent: "终端状态：子项自动完成" (line 195) vs "终端级联：主项重算" (line 202) use different terminology for the same pattern type.
