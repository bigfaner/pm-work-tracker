# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 7
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

## Dimension 2: Flow Diagrams — 19/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle), F2 (Item Pool), F3 (Team Management), F4 (Admin User Management), F5 (Views & Reports), plus an execution order diagram. All use valid `flowchart` syntax.

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths from start to end with explicit start `([...])` and end `([流程结束])` nodes. F5 uses `([验证完成])` as its terminal node. All are traversable from entry to exit.

### Decision points + error branches covered — 5/6

All five diagrams now include decision diamonds and error branches:

- F1: `{状态是否终端?}` diamond with yes/no branches, plus 5 error termination nodes.
- F2: `{审查决策}` diamond with three branches (assign/convert/reject) plus rollback and error terminations.
- F3: Error terminations at each step (duplicate code 422, already-member 409, user-not-found 404, PM role unchangeable 403, PM irremovable 422, non-PM 403), plus a dedicated member-role check node at line 130.
- F4: Error terminations at each step (duplicate username 409, validation errors 422, self-disable 422, not-found 404).
- F5 (previously flagged as having zero error branches): Now includes `ErrAuth{无 token / 非 SuperAdmin?}` (line 158) and `ErrTeam{teamId 不存在?}` (line 160) decision diamonds with corresponding `Assert403` and `Assert404` branches. This resolves the iteration-6 finding.

**Deduction (-1):** F3's permission check at line 130 uses `member{{member 角色}}` which renders as a hexagon node, not a standard diamond decision node. Every other decision point across all diagrams uses `{...}` diamond syntax. This is a stylistic inconsistency and the hexagon convention is not introduced or explained. A correct formulation would be `{member 角色?}`.

---

## Dimension 3: Functional Specs — 18/20

### Tables complete — 7/7

Seven spec tables with full coverage:
- F1: 17 rows x 5 columns — all cells populated.
- F2: 6 rows x 5 columns — all cells populated.
- F3: 9 rows x 5 columns — all cells populated.
- F4: 6 rows x 5 columns — all cells populated.
- F5: 6 rows x 3 columns (adapted for read-only) — all cells populated.
- F6: 6 rows (unit test gaps) — all cells populated.
- F7: 10 rows (shared helpers) — all cells populated with signatures and source files.

No empty cells. No placeholder text. F5's adapted column structure (Happy Path / Empty Data / Format Validation) is appropriate for read-only endpoints. All endpoint paths are now fully qualified (no abbreviated `.../` patterns remain).

### Field descriptions clear — 7/7

Most entries are specific with concrete triggers. Well-specified examples:
- `PUT /admin/users/:userId`: "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 → 422" — field names with character limits and cross-entity validation.
- `PUT /teams/:id`: "缺 name / name 超 100 字符 / description 超 500 字符 → 422" — field names with character limits.
- `POST /teams/:id/item-pool/:poolId/assign`: "无效主项 → 回滚" with cascade effect "已处理 → 409".

Previous iteration-6 issues with vague field descriptions ("错误团队 → 403", "非成员 → 403") have been partially resolved. F3 now specifies "非成员 → 403" (line 226) and "非 PM → 403" (lines 227-228, 232) which are clear about the role being checked.

Minor remaining shorthand: `GET /teams/:id/main-items/:itemId` (line 193) still uses "错误团队 → 403" — "错误团队" could be more precisely stated as "非成员访问其他团队" or similar. However, this is now an isolated case rather than a systemic pattern, and the meaning is inferable from the endpoint context. No deduction warranted at this granularity.

### Validation rules explicit — 4/6

Many validation rules are specific with concrete examples (e.g., "无效转换 如 new→completed → 422"). Permission columns across F2-F4 now specify roles ("member 角色 → 403", "非 SuperAdmin → 403", "非 PM → 403"), resolving the iteration-6 finding.

**Deduction (-1):** Several entries use terse shorthand without enumerating the full set of invalid transitions:
- `PUT /teams/:id/main-items/:itemId/status` (line 195): "无效转换 如 new→completed → 422" — gives one example but does not enumerate the full set of invalid transitions. A test writer must consult the state machine definition elsewhere.
- `PUT /teams/:id/sub-items/:subId/status` (line 202): "无效转换 如 new→completed → 422" — same issue, and the cascade effect description differs in specificity from the MainItem equivalent ("终端级联：主项重算" vs "终端状态：子项自动完成").

**Deduction (-1):** Permission columns still have minor gaps in specificity:
- `GET /teams/:id/main-items/:itemId` (line 193): "错误团队 → 403" — does not specify whether this means non-membership, cross-team access, or a nonexistent team.
- `PUT /teams/:id/main-items/:itemId` (line 194): "→ 403" alone without any role/condition qualifier. The reader must infer from the F1 diagram or Story 1 AC#6 that this is the member role.
- `GET /admin/users/:userId` (line 240): No permission column entry — it uses "—" for the entire permission column. Is this endpoint truly public, or does it require authentication? The F4 diagram at line 138 shows `A2[权限拒绝终止]` on a "非 SuperAdmin → 403" branch for the create endpoint, but the detail endpoint omits this. A test writer would not know whether to write an auth test for this endpoint.

---

## Dimension 4: User Stories — 14/20

### Coverage: one story per target user — 5/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 — covers F1-F6 and F7 helper extraction.
- **Code Reviewer**: Story 7 — addresses PR organization, naming conventions, incremental review.

**Deduction (-2):** Story 6's function list in AC#1 does not match the F7 spec table. Story 6 AC#1 lists 8 "existing" helper functions: `setupTestDB`/`setupTestRouter`/`loginAs`/`createTestUser`/`getAuthToken`/`makeRequest`/`assertJSON`/`parseResponseBody`. The F7 spec table lists 8 "existing" helper functions with different names: `setupTestDB`/`setupTestRouter`/`loginAs`/`signTokenWithClaims`/`seedProgressData`/`appendProgress`/`seedPoolData`/`seedReportData`. Only 3 of 8 names overlap (setupTestDB, setupTestRouter, loginAs). The other 5 names in Story 6 (createTestUser, getAuthToken, makeRequest, assertJSON, parseResponseBody) do not appear anywhere in the F7 spec table. Conversely, 5 names in the F7 spec table (signTokenWithClaims, seedProgressData, appendProgress, seedPoolData, seedReportData) do not appear in Story 6's AC. This is a factual inconsistency between two documents in the same PRD. A developer following Story 6's AC#1 would extract the wrong set of functions.

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the format precisely. Each "I want" is concrete with enumerated steps or named methods:

- Story 1: Exact lifecycle steps enumerated ("MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive").
- Story 6: Specifies file name (`helpers.go`) and function count (10).
- Story 7: Includes naming convention example (`TestItemLifecycle_CreateSubItem_TracksCompletionCascade`).
- Story 8: Names exact methods to test.

No vague actions like "manage", "handle", "improve".

### AC per story (Given/When/Then) — 2/6

Every story has ACs in Given/When/Then format. Total: ~30 acceptance criteria across 8 stories. The data in Story 5 AC#1 (`{NEW:0, completed:1, inProgress:2, overdue:0}`) matches the F5 spec table at line 249 exactly — this prior issue is resolved.

**Deduction (-2):** Story 6's AC#1 contains a factually incorrect function list. The AC says: "从中提取 `setupTestDB`/`setupTestRouter`/`loginAs`/`createTestUser`/`getAuthToken`/`makeRequest`/`assertJSON`/`parseResponseBody` 8 个现有辅助函数". Five of these 8 names (`createTestUser`, `getAuthToken`, `makeRequest`, `assertJSON`, `parseResponseBody`) do not exist in the F7 spec table. The AC is not verifiable as written — a developer cannot confirm they have extracted the correct functions because the AC names functions that the spec does not define. This is an acceptance criterion that is internally contradictory with the spec it references ("Then 函数签名与 F7 规格表一致" is impossible when the listed function names do not match the table).

**Deduction (-2):** Story 7's ACs continue to contain implementation prescriptions rather than observable behavior:
- AC #3: "Then 文件包含 F7 规格表所列的全部辅助函数，每个函数带 GoDoc 注释说明用途，且现有测试文件中的重复定义已删除" — mandates specific file path, GoDoc annotation format, and deletion of existing duplicates. These are implementation instructions ("how"), not acceptance criteria ("what outcome").
- AC #4: "Then 每个 PR 对应一个 Feature，commit message 以 `test(<domain>):` 前缀（helpers PR 以 `refactor(test):` 前缀）" — mandates commit message format. This is a contribution guideline/process rule, not a product acceptance criterion.

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
- F7: shared helpers (with reference to F7 spec table, 10 functions enumerated with signatures)

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

**Deduction (-2):** F7 (Shared Test Helpers) is a named in-scope item with a 10-row spec table, its own execution-order slot, and its own PR. Story 6 is the only story covering F7. However, Story 6's AC#1 lists 5 function names that do not appear in the F7 spec table and omits 5 function names that do. This is a direct factual inconsistency between the user stories document and the spec document — not a gap in coverage or a vague reference, but a naming conflict where the two documents describe different deliverable sets. A developer following Story 6's AC#1 literally would extract different functions than the F7 spec table requires. This undermines the three-way consistency between scope/functional-specs/user-stories.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 20 | 20 |
| Flow Diagrams | 19 | 20 |
| Functional Specs | 18 | 20 |
| User Stories | 14 | 20 |
| Scope Clarity | 18 | 20 |
| **Total** | **89** | **100** |

---

SCORE: 89/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 19/20
- Functional Specs: 18/20
- User Stories: 14/20
- Scope Clarity: 18/20

ATTACKS:
1. [User Stories]: Story 6 AC#1 lists 8 "existing" helper function names, but 5 of them (`createTestUser`, `getAuthToken`, `makeRequest`, `assertJSON`, `parseResponseBody`) do not exist anywhere in the F7 spec table, which instead lists `signTokenWithClaims`, `seedProgressData`, `appendProgress`, `seedPoolData`, `seedReportData`. The AC says "Then 函数签名与 F7 规格表一致" but this is impossible when the listed function names contradict the table. This is a factual naming conflict between two documents in the same PRD — a developer cannot satisfy both the AC and the spec simultaneously.
2. [User Stories]: Story 7 contains implementation prescriptions instead of acceptance criteria. AC#3 mandates GoDoc annotation format and deletion of duplicate definitions. AC#4 mandates commit message prefixes (`test(<domain>):`, `refactor(test):`). These are contribution guidelines, not product acceptance criteria — they prescribe "how" rather than "what outcome".
3. [Flow Diagrams]: F3's permission check node at line 130 uses hexagon syntax `member{{member 角色}}` instead of the standard diamond `{...}` decision syntax used by every other decision node across all diagrams. This is an unexplained stylistic inconsistency that could confuse a reader unfamiliar with Mermaid shape syntax.
