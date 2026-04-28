---
feature: permission-granularity
evaluated_at: 2026-04-28
iteration: 1
score: 67/100
---

# Design Evaluation Report: 细化 user / role 权限粒度

## 1. Summary Table

| Dimension | Score | Max | Grade | Notes |
|-----------|-------|-----|-------|-------|
| Architecture Clarity | 11 | 15 | B | Layer table + diagram present; no request execution flow |
| Interface & Model Definitions | 14 | 20 | B | API Handbook solid; truncated error description; missing type defs; useless cross-layer map |
| Error Handling | 7 | 15 | C | Truncated description; inconsistency between tech-design and API Handbook; privilege expansion unaddressed |
| Testing Strategy | 10 | 15 | B- | Structured but misses user:* route tests and TeamManagementPage; no % coverage target |
| Breakdown-Readiness | 15 | 20 | B | PRD map complete; ghost component RoleManagementPage.tsx has no design spec |
| Security Considerations | 10 | 15 | B- | Only 2 threats; privilege expansion from user:read semantic change not identified |
| **Total** | **67** | **100** | **B-** | Passable but not breakdown-ready without fixes |

---

## 2. Structure Check

| Section | Required | Present | Status |
|---------|----------|---------|--------|
| Overview | Yes | Yes | Pass |
| Architecture | Yes | Yes (layer table + ASCII diagram) | Pass — weak on execution flow |
| Interfaces | Yes | Yes (code snippets per component) | Pass — gaps noted below |
| Data Models | Yes | Yes (brief — no new tables) | Pass |
| Error Handling | Yes | Yes (1 row, truncated) | **Fail — truncated + inconsistent** |
| Testing Strategy | Yes | Yes (table + scenarios) | Pass — gaps noted |
| Security Considerations | Yes | Yes (2 threats) | Pass — shallow |
| Open Questions | Optional | Yes (2, both resolved) | Pass |
| Alternatives Considered | Optional | Yes (2 alternatives) | Pass |

---

## 3. Dimension-by-Dimension Findings

### Dimension 1: Architecture Clarity — 11/15 (B)

**What's there:**
- Layer Placement table maps 8 files/packages to their change content.
- ASCII component diagram shows the dependency tree from `permissions/codes.go` down to frontend consumers.
- Dependencies section explicitly states no new external dependencies.

**Deductions:**

**-3: No request execution flow.** The diagram shows static dependencies but not the runtime request path. A developer implementing this cannot answer: when `GET /admin/roles` arrives, what is the exact middleware chain? How does `RequirePermission` interact with `TeamScopeMiddleware`? The existing RBAC design (docs/features/rbac-permissions/design-eval.md) documented the full middleware chain with numbered steps. This design omits it entirely.

**-1: Layer table descriptions are change-lists, not responsibilities.** The table column header is "变更内容" — it lists what changes, not why each layer owns the change. For example, `migration/rbac.go` is listed as "数据迁移函数（幂等，事务执行）" but the table doesn't explain the relationship between `MigratePermissionGranularity` and the existing `SyncPresetRoles` call chain.

---

### Dimension 2: Interface & Model Definitions — 14/20 (B)

**What's there:**
- Permission code constants with descriptions (Go snippets).
- Router binding changes with concrete `deps.perm(...)` calls.
- Migration function signature: `func MigratePermissionGranularity(db *gorm.DB) error`.
- Frontend TypeScript snippets for `PERMISSION_GROUPS` and `useQuery` guard.
- API Handbook: all 10 affected endpoints with request/response tables and error codes.

**Deductions:**

**-3: Cross-Layer Data Map is 7 rows of identical content.** Every row in the Cross-Layer Data Map has the same Storage Layer, Backend Model, API/DTO, Frontend Type, and Validation Rule values — only the `permission_code` value differs. This section adds zero information. It lists `permission_code (user:list)` through `permission_code (role:delete)` as separate rows but they all map to the same column, model field, JSON tag, and frontend type. This is padding, not specification.

**-2: `User` and `Team` response types are undefined in the API Handbook.** `GET /admin/users` returns `User[]` and `GET /admin/teams` returns `Team[]`, but neither `User` nor `Team` is defined in the Data Contracts section. The handbook defines `RoleListItem`, `PermissionItem`, and `ResourcePermissions` but omits the two most-used response types. A developer implementing the frontend must guess the shape.

**-1: Migration function body is pseudocode, not Go.** The migration logic is described as a numbered list with pseudocode (`SELECT INTO 临时表或应用层记录`, `INSERT ... 忽略已存在`). The function signature is defined but the implementation contract (what SQL operations, what GORM calls, what the idempotency check looks like) is left to the implementer. Compare to the RBAC design which showed concrete GORM method calls.

---

### Dimension 3: Error Handling — 7/15 (C)

**What's there:**
- One-row error table in tech-design.md.
- Propagation strategy: `RequirePermission → apperrors.RespondError → JSON`.
- API Handbook error codes table with 6 entries.

**Deductions:**

**-4: Truncated error description — literal document defect.** The Error Types & Codes table in tech-design.md reads: `"缺少对应权限码时，`RequirePermission` 中间件返"` — the sentence is cut off mid-word. This is not a minor typo; it is an incomplete specification in a required section.

**-3: Inconsistency between tech-design and API Handbook.** tech-design.md states `"无新增错误类型"` (no new error types). The API Handbook defines `ERR_ROLE_NAME_EXISTS` (409), `ERR_ROLE_IN_USE` (422), and `ERR_PRESET_ROLE_IMMUTABLE` (403). These are either new error codes or newly-relevant ones — either way, the tech-design's claim is false. A developer reading only the tech-design would not know these codes exist.

**-1: No error handling for migration failure beyond "transaction rollback".** The Security Considerations section mentions transaction rollback, but the Error Handling section says nothing about what the caller of `MigratePermissionGranularity` should do if it returns an error. Does the application fail to start? Log and continue? This is a deployment-critical question left unanswered.

---

### Dimension 4: Testing Strategy — 10/15 (B-)

**What's there:**
- Per-Layer Test Plan table with 4 rows (router middleware, migration, preset roles, frontend guards).
- Key Test Scenarios with concrete Go and TypeScript code snippets.
- Overall Coverage Target stated.

**Deductions:**

**-2: Router middleware tests cover only role:* codes, not the changed user:* codes.** The test plan says "4 个新权限码（role:read/create/update/delete）各自在无权限时返回 403". But `user:list`, `user:read` (semantic change), and `user:assign_role` also changed their route bindings. There are no tests verifying that `GET /admin/users` now requires `user:list` instead of the old `user:read`, or that `GET /admin/users/:userId` now requires the new `user:read`. These are regression risks.

**-1: No test for `TeamManagementPage.tsx` `canReadRoles` guard.** Section 8 of the Interfaces adds a `canReadRoles` guard to `useQuery` in `TeamManagementPage.tsx`. This is a new conditional behavior but the Testing Strategy has no corresponding test case. The frontend test plan only covers `App.tsx` route guards.

**-1: Coverage target is a count, not a percentage.** `"新增 7 个测试用例（4 路由 + 3 迁移）"` is a task count, not a coverage target. The project's existing design evals specify percentage targets (85%/90%). A count of 7 test cases says nothing about whether the changed code paths are adequately covered.

**-1: No mention of test tooling for frontend.** The backend test tooling is implicit (Go `httptest` + Gin). The frontend test tooling is not named at all. The project uses vitest + testing-library per `.claude/rules/testing.md` — naming it would be consistent with the project standard.

---

### Dimension 5: Breakdown-Readiness — 15/20 (B)

**What's there:**
- PRD Coverage Map with 13 rows mapping every story/AC to a design component and interface.
- File-level changes enumerated across 8 sections of the Interfaces chapter.
- Both Open Questions resolved with decisions recorded.

**Deductions:**

**-3: `RoleManagementPage.tsx` appears in the PRD Coverage Map without any design specification.** Story 6 maps `"无 role:create → 创建角色按钮不显示"` to `RoleManagementPage.tsx` with `usePermission("role:create")`. This component is not mentioned anywhere else in the design — not in the Architecture layer table, not in the Interfaces sections, not in the component diagram. A developer assigned this task has no specification to work from. This is a ghost component.

**-1: `router_test.go` appears only in the Testing Strategy, not in the Architecture or Interfaces sections.** The design says seed data in `router_test.go` must be updated, but this file is not listed in the Layer Placement table or the component diagram. It will be missed in task breakdown unless the developer reads the Testing Strategy section carefully.

**-1: No explicit task count or enumeration.** The RBAC design eval noted 20+ discrete components. This design's components are scattered across 8 interface sections without a summary count. A task breakdown author must manually enumerate them.

---

### Dimension 6: Security Considerations — 10/15 (B-)

**What's there:**
- 2 threats identified: permission code semantic collision, migration partial execution.
- 4 mitigations: transaction rollback, snapshot, no-delete of user:read row, CI grep assertion.

**Deductions:**

**-3: Privilege expansion from `user:read` semantic change is not identified as a threat.** The design's mitigation for the `user:read` semantic collision is: "旧 `user:read` 行不删除（语义随路由绑定更新而更新）". The consequence: every role that previously had `user:read` (meaning "list users") now silently gains `user:read` (meaning "view user details including sensitive fields: email, phone"). This is a privilege expansion — roles that were only authorized to list users now have access to sensitive PII fields. The design does not identify this as a threat, does not evaluate whether it is acceptable, and does not propose any mitigation (e.g., auditing which roles are affected, notifying admins).

**-2: Only 2 threats in the threat model.** The RBAC design identified 5 threats. This design has a narrower scope but still has unaddressed threats: (1) the privilege expansion above, (2) what happens if `MigratePermissionGranularity` is called before `SyncPresetRoles` — are the new codes registered in time? (3) lockout: if a migration bug removes all `role:*` permissions, no one can manage roles. None of these are discussed.

---

## 4. PRD Traceability Matrix

| PRD Story / AC | Design Coverage | Status |
|----------------|-----------------|--------|
| Story 1: role:read → GET /admin/roles 200 | router.go binding | Covered |
| Story 1: 无 role:read → 403 | middleware/permission.go | Covered |
| Story 2: user:list → GET /admin/users 200 | router.go binding | Covered |
| Story 2: 无 user:read → GET /admin/users/:userId 403 | router.go binding | Covered |
| Story 3: role:create → POST /admin/roles 201 | router.go binding | Covered |
| Story 3: 只有 role:read → POST /admin/roles 403 | middleware/permission.go | Covered |
| Story 4: pm 有完整 user:* + role:* | seedPresetRoles pmCodes | Covered |
| Story 5: user:manage_role → role:create+update+delete | MigratePermissionGranularity | Covered |
| Story 5: 旧 user:read → user:list | MigratePermissionGranularity | Covered |
| Story 5: 事务回滚 | MigratePermissionGranularity | Covered |
| Story 6: 无 user:list → 用户管理页入口不显示 | App.tsx + Sidebar.tsx | Covered |
| Story 6: 无 role:create → 创建角色按钮不显示 | RoleManagementPage.tsx | **Ghost — no design spec** |
| Story 6: 后端是安全防线 | middleware/permission.go | Covered |

**12/13 stories covered. 1 ghost component.**

---

## 5. Top Issues

### Issue 1: Truncated Error Description (Dimension 3 — C, Blocking)

tech-design.md Error Types & Codes table: `"缺少对应权限码时，`RequirePermission` 中间件返"` — sentence cut off. This is a defect in a required section. Fix: complete the description and reconcile with the API Handbook's 6 error codes. The tech-design must not claim "无新增错误类型" while the API Handbook defines ERR_ROLE_NAME_EXISTS, ERR_ROLE_IN_USE, and ERR_PRESET_ROLE_IMMUTABLE.

### Issue 2: Ghost Component RoleManagementPage.tsx (Dimension 5 — B, Blocking for breakdown)

PRD Coverage Map Story 6 references `RoleManagementPage.tsx` with `usePermission("role:create")` but this component has no design specification anywhere in the document. A developer cannot implement it from this design. Fix: add a frontend section specifying the `role:create` button guard in `RoleManagementPage.tsx`, or clarify that this component already exists and only needs a `usePermission` call added.

### Issue 3: Privilege Expansion Risk Unaddressed (Dimension 6 — B-)

The mitigation for `user:read` semantic collision is to not delete the old row. This silently grants every role that had `user:read` (list users) access to `user:read` (view user details + sensitive fields: email, phone). This is a privilege expansion affecting production data. Fix: add this as a named threat, evaluate whether it is acceptable (it may be — if all roles with user:read should also have user:read-detail), and document the decision explicitly.

### Issue 4: Router Middleware Tests Miss Changed user:* Bindings (Dimension 4 — B-)

The test plan covers only the 4 new `role:*` codes. The changed `user:list`, `user:read` (semantic), and `user:assign_role` bindings have no corresponding test cases. Fix: add test cases verifying `GET /admin/users` requires `user:list`, `GET /admin/users/:userId` requires `user:read`, and `POST /admin/users` requires `user:assign_role`.

### Issue 5: Cross-Layer Data Map Is Useless Padding (Dimension 2 — B)

All 7 rows in the Cross-Layer Data Map have identical Storage Layer, Backend Model, API/DTO, Frontend Type, and Validation Rule values. The section conveys no information beyond "these 7 permission codes exist." Fix: either remove the section or replace it with a table that shows something non-trivial (e.g., which codes are new vs. renamed vs. semantic-change-only).

---

## 6. Verdict

**Overall Score: 67/100 — B-**

The design is functional for a narrow refactor but has a blocking defect (truncated error description), a breakdown blocker (ghost component), and an unaddressed security risk (privilege expansion). The API Handbook is the strongest artifact — comprehensive and directly implementable. The tech-design itself is thinner than the project's established standard.

**Breakdown-Readiness: Not yet.** Fix Issues 1 and 2 before proceeding to `/breakdown-tasks`.
