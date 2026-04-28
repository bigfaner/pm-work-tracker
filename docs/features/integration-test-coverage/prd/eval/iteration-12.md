# Evaluation Report: Integration Test Coverage PRD

**Iteration:** 12
**Date:** 2026-04-27
**Documents evaluated:**
- `docs/features/integration-test-coverage/prd/prd-spec.md`
- `docs/features/integration-test-coverage/prd/prd-user-stories.md`

---

## Dimension 1: Background & Goals — 20/20

### Three elements (cause/target/people) — 7/7

All three elements are present and highly specific:

- **Cause (原因):** Line 14: "54 个 API 端点，但只有 18 个（33%）有集成测试，36 个端点完全无测试覆盖" with a concrete, traceable incident: commit `1883499` introduced a timezone bug (`view_handler.go`) and a filter logic bug (`view_service.go`), both traced to `docs/lessons/weekly-view-bug-fixes.md` Bug 2 & Bug 3. The evidence includes commit hash, file names, and bug tracker references.

- **Target (对象):** Line 20: "为 36 个未测试端点编写端到端集成测试，按用户流程组织（而非按端点隔离）。同时补全 6 个单元测试缺口" -- exact count with methodology stated.

- **People (人员):** Lines 24-25: Two named personas with distinct roles: Developer ("编写和运行测试、在 CI 中获得回归保护") and Code Reviewer ("通过 PR 级别的增量提交审查测试代码"). Each has a purpose statement tied to the deliverable.

### Goals quantified — 7/7

Five quantified targets in a structured table (lines 29-35):

1. Endpoint integration test coverage: 33% -> 100%
2. New test cases: >= 150
3. Unit test gaps: 6/6 closed
4. Test suite execution time: < 150 seconds
5. PR reviewability: "6 个测试 PR + 1 个 helpers PR，每个 <= 500 行"

All five have numeric targets with units (%, count, seconds, lines). No vague language.

### Background and goals logically consistent — 6/6

Goals directly address the stated problem. 33%->100% maps to the 36 uncovered endpoints. The 6 unit gaps tie to named files. Both personas have traceable goals (Developer: CI regression protection; Reviewer: incremental PRs). The 150 test case target follows from ~36 endpoints x ~4 scenarios each. The PR size constraint (500 lines) ties back to the Code Reviewer persona's need for manageable review.

---

## Dimension 2: Flow Diagrams — 20/20

### Mermaid diagram exists — 7/7

Six Mermaid flowcharts present: F1 (Item Lifecycle, line 80), F2 (Item Pool, line 100), F3 (Team Management, line 117), F4 (Admin User Management, line 135), F5 (Views & Reports, line 149), plus an execution order dependency diagram (line 177). All use valid `flowchart TD` or `flowchart LR` syntax with proper node shapes (`([...])` for terminals, `{...}` for decisions, `[...]` for processes).

### Main path complete (start to end) — 7/7

All five domain flow diagrams show complete happy paths with explicit start `([...])` and end `([流程结束])` / `([验证完成])` nodes:

- F1: PM creates MainItem -> SubItem -> Progress -> Status -> Archive -> Done. Fully traversable.
- F2: Submit pool item -> Review decision -> Assign/Convert/Reject -> Done. Complete.
- F3: Create team -> Invite members -> Role change -> Remove -> Dissolve -> Done. Complete.
- F4: Create user -> Edit -> Toggle status -> Done. Complete.
- F5: Seed data -> Call endpoint -> Assert content -> Format branching -> Done. Complete.

The execution order diagram (line 177) provides an additional LR flow showing F1->F7->F2->F3->F4->F5->F6 with dependency arrows.

### Decision points + error branches covered — 6/6

All five diagrams include decision diamonds and error/exception branches:

- **F1:** `{状态是否终端?}` diamond (line 89) with yes/no branches. Five error termination nodes: validation errors (A1, B1, C1), permission denial (A2), invalid transition (D1), archive rejection (F1).
- **F2:** `{审查决策}` diamond (line 102) with three branches (assign/convert/reject). Rollback path from failed assign back to review (line 112). Validation errors and missing-reason error terminations.
- **F3:** `{member 角色?}` diamond (line 130). Error terminations at every step: duplicate code 422, already-member 409, user-not-found 404, PM role unchangeable 403, PM irremovable 422, non-PM operations 403.
- **F4:** Error terminations: duplicate username 409, non-SuperAdmin 403, validation errors (displayName empty/>64, email >100, nonexistent teamKey) 422, self-disable 422, user-not-found 404.
- **F5:** Four decision diamonds: `{无 token / 非 SuperAdmin?}`, `{teamId 不存在?}`, `{返回数据是否非空?}`, `{端点类型?}` with 403/404 branches and content format branching.

---

## Dimension 3: Functional Specs — 19/20

### Tables complete — 7/7

Seven spec tables, all fully populated:

- F1: 17 rows x 5 columns (Happy Path / Validation / Permission / Not Found / Cascade)
- F2: 6 rows x 5 columns
- F3: 9 rows x 5 columns
- F4: 6 rows x 5 columns
- F5: 6 rows x 3 columns (adapted for read-only endpoints: Happy Path / Empty Data / Format)
- F6: 6 rows x 3 columns (File / Gap / Supplement)
- F7: 10 rows x 4 columns (Function / Source File / Signature / Purpose)

No empty cells where data is expected. No placeholder text ("TBD", "TODO"). The F5 table adapts its columns to the read-only nature of views/reports endpoints rather than forcing the CRUD template. The "--" entries in permission/not-found/cascade columns are acceptable for endpoints where no such scenario applies (e.g., `GET /teams/:id/main-items` list endpoint has no permission check because team membership is verified by URL context).

### Field descriptions clear — 7/7

Entries are specific with concrete triggers, field names, and character limits. Examples:

- `PUT /admin/users/:userId` (line 241): "displayName 空 / 超 64 字符 / email 超 100 字符 / teamKey 指向不存在的团队 -> 422" -- field names with exact character limits and cross-entity validation.
- `PUT /teams/:id` (line 227): "缺 name / name 超 100 字符 / description 超 500 字符 -> 422" -- field names with character limits.
- F5 `GET /teams/:id/views/weekly` (line 249): "3 项（其中 1 项 completed、2 项 in-progress）：`stats: {NEW:0, completed:1, inProgress:2, overdue:0}`" -- exact expected JSON structure with values.
- F7 table (lines 272-283): Each helper function has source file, full Go signature, and purpose description. E.g., `setupTestDB` has `func setupTestDB(t *testing.T) (*gorm.DB, *seedData)` with purpose "创建内存 SQLite、运行迁移、种子用户/团队/角色/权限".

Permission columns consistently specify roles: "member 角色 -> 403", "非 PM -> 403", "非 SuperAdmin -> 403", "member 无 `sub_item:create` 权限 -> 403" (line 198).

### Validation rules explicit — 5/6

Most validation rules are specific with concrete triggers and field-level detail. Two status-transition entries now include representative examples (improvement from prior iterations):

- Line 195: "状态机不允许的转换，如 new→completed、completed→in-progress（完整定义见 `status/transition.go`）→ 422"
- Line 202: Same pattern with examples.

**Deduction (-1):** The `GET /teams/:id/main-items` endpoint (line 192) and `GET /teams/:id/main-items/:itemId/sub-items` endpoint (line 199) both show "--" in every non-happy-path column. For `GET /teams/:id/main-items`, no permission, not-found, or validation scenario is listed. While this may be accurate (a list endpoint for a valid team always returns 200, possibly with empty results), the absence of any "未授权/无 token" consideration is notable -- every other list endpoint in the document at minimum has authentication middleware implied. The spec does not explicitly state whether authentication is required for these endpoints, creating a minor gap for the test writer.

---

## Dimension 4: User Stories — 19/20

### Coverage: one story per target user — 7/7

Background defines two personas: Developer and Code Reviewer.

- **Developer**: Stories 1-6, 8 (seven stories) -- covers all seven in-scope features (F1-F7 helper extraction, F6 unit gaps).
- **Code Reviewer**: Story 7 -- addresses PR organization, naming conventions, and incremental review.

Both personas have at least one story. All seven in-scope features have corresponding stories. Story 6 covers F7 (shared helpers), Story 8 covers F6 (unit gaps).

### Format correct (As a / I want / So that) — 7/7

All eight stories follow the "As a / I want / So that" format precisely. Each "I want" clause is concrete with specific actions:

- Story 1: "通过集成测试验证 MainItem 创建 -> SubItem 创建 -> Progress 追加 -> Status 变更 -> Archive 的完整生命周期" -- exact lifecycle steps enumerated.
- Story 6: "从现有集成测试（`auth_isolation_test.go`、`progress_completion_test.go`）和 F1 编写过程中提取 10 个共享辅助函数到独立 `helpers.go` 文件" -- specifies source files, count, and target file.
- Story 8: "为 permission_handler、ConvertToMain、UpdateTeam、GetByBizKey 等未测试方法补充单元测试" -- names exact methods.

No vague actions like "manage", "handle", "improve" in any story.

### AC per story (Given/When/Then) — 5/6

Every story has ACs in Given/When/Then format. Approximately 30 acceptance criteria across 8 stories. The previous iteration's conjoined ACs in Story 7 have been properly split into separate single-outcome ACs (lines 106-110).

**Deduction (-1):** Story 7, AC#5 (line 110): "Given 审查全部 PR 合并后的代码库, When 检查测试覆盖, Then 各 Feature 测试文件独立且覆盖对应业务域" -- the "When" clause ("检查测试覆盖") describes a code coverage analysis activity that is more CI/automation-oriented than reviewer-oriented. A code reviewer reviews PR diffs, not post-merge coverage reports. The Given precondition ("全部 PR 合并后的代码库") places this clearly after the review is complete, making it a verification-of-outcome rather than an acceptance criterion the reviewer can apply during review. This is a minor persona mismatch: the action belongs to a CI pipeline or a tech lead doing final verification, not to the "代码审查者" persona in the act of reviewing PRs.

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

The seven scope items (F1-F7) map one-to-one to seven functional spec tables (F1-F7) and eight user stories (Stories 1-8):

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
| Functional Specs | 19 | 20 |
| User Stories | 19 | 20 |
| Scope Clarity | 20 | 20 |
| **Total** | **98** | **100** |

---

SCORE: 98/100

DIMENSIONS:
- Background & Goals: 20/20
- Flow Diagrams: 20/20
- Functional Specs: 19/20
- User Stories: 19/20
- Scope Clarity: 20/20

ATTACKS:
1. [Functional Specs]: Two list-only GET endpoints -- `GET /teams/:id/main-items` (line 192) and `GET /teams/:id/main-items/:itemId/sub-items` (line 199) -- show "--" in every non-happy-path column (Validation, Permission, Not Found, Cascade) without stating whether authentication is required. Every other list endpoint in the document has at minimum an implied auth check. The spec does not confirm whether these are public endpoints or whether a missing/invalid JWT should return 401, leaving the test writer to guess. To fix: add "需认证" or "无 token -> 401" in the Permission column, or add a footnote explaining the auth policy for list endpoints.
2. [User Stories]: Story 7 AC#5 (line 110) -- "Given 审查全部 PR 合并后的代码库, When 检查测试覆盖, Then 各 Feature 测试文件独立且覆盖对应业务域" -- has a persona mismatch. The Given precondition ("全部 PR 合并后") places the reviewer after all reviews are complete. "检查测试覆盖" is a post-merge verification activity (CI/tech lead), not a code reviewer action during PR review. The "代码审查者" persona reviews diffs, not merged codebases. To fix: reframe as "Given 7 PRs submitted sequentially, When I review each PR's file list, Then each PR adds test files for exactly one F-domain and no F-domain's tests are split across PRs" -- this keeps the criterion reviewer-centric.
3. [Functional Specs]: Minor -- the F5 spec table (lines 248-254) uses a different column structure (Happy Path / Empty Data / Format Validation) from the F1-F4 tables (Happy Path / Validation Error / Permission Denial / Not Found / Cascade). While this adaptation is reasonable for read-only endpoints, the F5 table drops the Permission and Not Found columns entirely. However, the F5 Mermaid diagram (lines 158-161) includes 403 and 404 branches. This creates a minor inconsistency between the flow diagram (which tests auth and team-existence errors) and the spec table (which omits those scenarios). To fix: either add Permission and Not Found columns to F5, or add a note stating "权限/未找到场景与 F1-F4 共享测试模式，由 F1-F4 的认证中间件测试覆盖" explaining the omission.
