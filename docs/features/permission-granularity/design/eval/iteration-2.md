---
feature: permission-granularity
evaluated_at: 2026-04-28
iteration: 2
score: 83/100
---

# Design Evaluation Report: 细化 user / role 权限粒度

## 1. Summary Table

| Dimension | Score | Max | Grade | Notes |
|-----------|-------|-----|-------|-------|
| Architecture Clarity | 11 | 15 | B | Same as iteration 1 — execution flow still absent |
| Interface & Model Definitions | 14 | 20 | B | All three iteration-1 deductions remain unfixed |
| Error Handling | 14 | 15 | A- | Truncated description fixed; inconsistency fixed; migration caller behavior still unspecified |
| Testing Strategy | 11 | 15 | B | Frontend tooling now named; user:* route tests and TeamManagementPage guard still missing |
| Breakdown-Readiness | 19 | 20 | A- | Ghost component fixed; router_test.go now in layer table; no task count summary |
| Security Considerations | 14 | 15 | A- | Privilege expansion now identified and decided; ordering/lockout threats still absent |
| **Total** | **83** | **100** | **B+** | Significant improvement; two remaining blockers before breakdown |

---

## 2. Issues Fixed Since Iteration 1

| Iteration-1 Issue | Status |
|-------------------|--------|
| Issue 1: Truncated error description | **Fixed** — ERR_FORBIDDEN row now complete; all 4 error codes listed in tech-design |
| Issue 1: Inconsistency between tech-design and API Handbook | **Fixed** — tech-design now lists ERR_ROLE_NAME_EXISTS, ERR_ROLE_IN_USE, ERR_PRESET_ROLE_IMMUTABLE with explanation that they pre-exist |
| Issue 2: Ghost component RoleManagementPage.tsx | **Fixed** — Section 9 added with concrete TSX snippet; component added to Layer Placement table |
| Issue 3: Privilege expansion unaddressed | **Fixed** — Threat 3 added with explicit decision, rationale, and mitigation |
| Issue 4 (partial): Frontend test tooling unnamed | **Fixed** — Per-Layer Test Plan now names "vitest + testing-library" |
| router_test.go absent from Architecture/Interfaces | **Fixed** — Now listed in Layer Placement table |

---

## 3. Structure Check

| Section | Required | Present | Status |
|---------|----------|---------|--------|
| Overview | Yes | Yes | Pass |
| Architecture | Yes | Yes (layer table + ASCII diagram, 9 files) | Pass — execution flow still absent |
| Interfaces | Yes | Yes (9 sections, all components covered) | Pass |
| Data Models | Yes | Yes (no new tables, confirmed) | Pass |
| Error Handling | Yes | Yes (4 error codes, complete descriptions) | Pass |
| Testing Strategy | Yes | Yes (table + scenarios) | Pass — gaps noted |
| Security Considerations | Yes | Yes (3 threats) | Pass — shallow on 2 threats |
| Open Questions | Optional | Yes (2, both resolved) | Pass |
| Alternatives Considered | Optional | Yes (2 alternatives) | Pass |

---

## 4. Dimension-by-Dimension Findings

### Dimension 1: Architecture Clarity — 11/15 (B)

**What's there:**
- Layer Placement table now covers all 9 files including `RoleManagementPage.tsx` and `router_test.go`.
- ASCII component diagram shows static dependency tree.
- Dependencies section confirms no new external dependencies.

**Deductions:**

**-3: No request execution flow.** Unchanged from iteration 1. The diagram shows static dependencies but not the runtime request path. When `GET /admin/roles` arrives, the document still cannot answer: what is the exact middleware chain? How does `RequirePermission` interact with `TeamScopeMiddleware`? The component diagram shows `router.go → middleware/permission.go (RequirePermission)` as a static edge but provides no numbered execution steps. A developer implementing this cannot verify the middleware ordering from this document alone.

**-1: Layer table descriptions are change-lists, not responsibilities.** The column header remains "变更内容". For example, `migration/rbac.go` is listed as "数据迁移函数（幂等，事务执行）" — this describes what it does in this change, not why the migration layer owns this responsibility or how it relates to the existing `SyncPresetRoles` call chain.

---

### Dimension 2: Interface & Model Definitions — 14/20 (B)

**What's there:**
- Permission code constants with descriptions (Go snippets).
- Router binding changes with concrete `deps.perm(...)` calls.
- Migration function signature and detailed pseudocode steps.
- Frontend TypeScript snippets for all 4 affected components.
- API Handbook: all 10 affected endpoints with request/response tables and error codes.
- Section 9 now fully specifies the `RoleManagementPage.tsx` button guard.

**Deductions:**

**-3: Cross-Layer Data Map is still 7 rows of identical content.** Unchanged from iteration 1. Every row has the same Storage Layer, Backend Model, API/DTO, Frontend Type, and Validation Rule values — only the `permission_code` value differs. The section still conveys no information beyond "these 7 permission codes exist." All rows read `同上` for every column after the first. This is padding.

**-2: `User` and `Team` response types are still undefined in the API Handbook.** `GET /admin/users` returns `User[]` and `GET /admin/teams` returns `Team[]`. The Data Contracts section defines `RoleListItem`, `PermissionItem`, and `ResourcePermissions` — the same three types as iteration 1. Neither `User` nor `Team` is defined. A developer implementing the frontend must still guess the shape of the two most-used response types.

**-1: Migration function body is still pseudocode.** The migration logic remains a numbered list with pseudocode (`SELECT INTO 临时表或应用层记录`, `INSERT ... 忽略已存在`). The function signature is defined but the implementation contract — what GORM calls, what the idempotency check SQL looks like, what `schema_migrations` insert looks like — is left to the implementer.

---

### Dimension 3: Error Handling — 14/15 (A-)

**What's there:**
- Error Types & Codes table now has 4 complete rows: ERR_FORBIDDEN (complete description), ERR_ROLE_NAME_EXISTS, ERR_ROLE_IN_USE, ERR_PRESET_ROLE_IMMUTABLE.
- Explanation that the last three pre-exist and are not new: "后三个错误码在现有角色管理逻辑中已存在".
- Propagation strategy unchanged and correct.
- API Handbook error codes table consistent with tech-design.

**Deductions:**

**-1: No error handling for migration failure at the caller level.** The Mitigations section states "迁移在单一数据库事务中执行，失败自动回滚" but the Error Handling section still does not specify what the caller of `MigratePermissionGranularity` should do when it returns an error. Does the application fail to start? Log and continue? Alert ops? This is a deployment-critical question. The function signature returns `error` but the contract for that error path is unspecified.

---

### Dimension 4: Testing Strategy — 11/15 (B)

**What's there:**
- Per-Layer Test Plan table with 4 rows; frontend row now names "vitest + testing-library".
- Key Test Scenarios with concrete Go and TypeScript code snippets.
- `router_test.go` seed data update specified.

**Deductions:**

**-2: Router middleware tests still cover only role:* codes, not the changed user:* codes.** The test plan still reads "4 个新权限码（role:read/create/update/delete）各自在无权限时返回 403". The Key Test Scenarios still list only 4 backend tests, all for `role:*` routes. `GET /admin/users` (now requires `user:list`), `GET /admin/users/:userId` (now requires `user:read` with new semantics), and `POST /admin/users` (now requires `user:assign_role`) have no corresponding test cases. These are regression risks — the changed bindings are untested.

**-1: No test for `TeamManagementPage.tsx` `canReadRoles` guard.** Section 8 adds `enabled: canReadRoles` to the `useQuery` call. This is a new conditional behavior. The frontend test plan still only covers `App.tsx` route guards (`user:list` controls user management page, `role:read` controls role management page). The `canReadRoles` guard in `TeamManagementPage.tsx` has no test case.

**-1: Coverage target is a count, not a percentage.** `"新增 7 个测试用例（4 路由 + 3 迁移）"` is a task count. The project's established standard specifies percentage targets. A count of 7 test cases says nothing about whether the changed code paths are adequately covered.

---

### Dimension 5: Breakdown-Readiness — 19/20 (A-)

**What's there:**
- PRD Coverage Map with 13 rows, all components now specified.
- `RoleManagementPage.tsx` fully specified in Section 9 with TSX snippet.
- `router_test.go` now listed in Layer Placement table.
- Both Open Questions resolved with decisions recorded.
- File-level changes enumerated across 9 sections.

**Deductions:**

**-1: No explicit task count or enumeration.** The 9 interface sections enumerate changes but provide no summary count of discrete implementation tasks. A task breakdown author must manually count: 1 Go constants file, 1 router file, 1 migration function, 1 seedPresetRoles update, 4 frontend files = 8 discrete tasks minimum, but this is not stated. The RBAC design eval noted 20+ discrete components with an explicit count. This design leaves the enumeration to the reader.

---

### Dimension 6: Security Considerations — 14/15 (A-)

**What's there:**
- 3 threats identified: permission code semantic collision, migration partial execution, `user:read` privilege expansion.
- Threat 3 includes an explicit decision with rationale: "接受此扩张，理由如下" with 3 bullet points.
- Mitigation: "上线前由管理员审查持有 user:read 的角色列表，确认扩张范围可接受".
- 4 mitigations covering transaction rollback, snapshot, no-delete of user:read row, CI grep assertion.

**Deductions:**

**-1: Ordering dependency and lockout threats still absent.** The threat model grew from 2 to 3 threats but two threats identified in iteration 1 remain unaddressed: (1) what happens if `MigratePermissionGranularity` is called before `SyncPresetRoles` — are the new codes registered in `permissions.Registry` in time for the migration to reference them? (2) lockout: if a migration bug incorrectly removes all `role:*` permissions, no one can manage roles. Neither scenario is discussed or mitigated.

---

## 5. PRD Traceability Matrix

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
| Story 6: 无 role:create → 创建角色按钮不显示 | RoleManagementPage.tsx Section 9 | Covered |
| Story 6: 后端是安全防线 | middleware/permission.go | Covered |

**13/13 stories covered. Ghost component resolved.**

---

## 6. Top Issues

### Issue 1: user:* Route Tests Missing (Dimension 4 — B, Blocking for regression safety)

The test plan covers only the 4 new `role:*` codes. `GET /admin/users` (changed from `user:read` to `user:list`), `GET /admin/users/:userId` (semantic change on `user:read`), and `POST /admin/users` (changed from `user:manage_role` to `user:assign_role`) have no test cases. Fix: add 3 test cases verifying the changed `user:*` bindings return 403 without the new codes.

### Issue 2: Cross-Layer Data Map Is Still Useless Padding (Dimension 2 — B)

All 7 rows in the Cross-Layer Data Map have identical Storage Layer, Backend Model, API/DTO, Frontend Type, and Validation Rule values. The section conveys no information. Fix: remove the section entirely, or replace it with a table that distinguishes new codes from renamed codes from semantic-change-only codes.

### Issue 3: User and Team Response Types Undefined (Dimension 2 — B)

`GET /admin/users` returns `User[]` and `GET /admin/teams` returns `Team[]`. Neither type is defined in the Data Contracts section. Fix: add `User` and `Team` interface definitions to the Data Contracts section of the API Handbook.

### Issue 4: Migration Caller Error Contract Unspecified (Dimension 3 — A-)

`MigratePermissionGranularity` returns `error` but the document does not specify what the caller should do on failure. Fix: add one sentence to the Error Handling section: e.g., "调用方（`main.go` 启动流程）若收到非 nil error，应终止启动并记录错误日志，不允许静默忽略。"

### Issue 5: TeamManagementPage canReadRoles Guard Has No Test (Dimension 4 — B)

Section 8 adds `enabled: canReadRoles` to `useQuery` in `TeamManagementPage.tsx`. No test case covers this guard. Fix: add a frontend test case verifying that without `role:read`, the roles query is not issued.

---

## 7. Verdict

**Overall Score: 83/100 — B+**

Iteration 2 addressed all five blocking issues from iteration 1. The ghost component is fully specified, the error section is complete and consistent, and the privilege expansion is now an explicit, decided threat. The API Handbook remains the strongest artifact.

**Breakdown-Readiness: Conditional.** The design is implementable for 12 of 13 stories. The remaining gaps (missing user:* route tests, undefined User/Team types, useless Cross-Layer Data Map) are quality issues, not blockers. A developer can proceed to `/breakdown-tasks` but should note that the test plan is incomplete and the Cross-Layer Data Map should be removed to avoid confusion.
