---
feature: permission-granularity
evaluated_at: 2026-04-28
iteration: 3
score: 92/100
---

# Design Evaluation Report: 细化 user / role 权限粒度

## 1. Summary Table

| Dimension | Score | Max | Grade | Notes |
|-----------|-------|-----|-------|-------|
| Architecture Clarity | 14 | 15 | A- | Execution flow added; layer table column header still "变更内容" |
| Interface & Model Definitions | 17 | 20 | B+ | Cross-Layer Data Map replaced; User/Team types and migration pseudocode still unresolved |
| Error Handling | 14 | 15 | A- | Migration caller error contract still unspecified |
| Testing Strategy | 14 | 15 | A- | user:* tests and canReadRoles guard test added; coverage target still a count |
| Breakdown-Readiness | 19 | 20 | A- | No explicit task count summary |
| Security Considerations | 14 | 15 | A- | Ordering dependency and lockout threats still absent |
| **Total** | **92** | **100** | **A-** | Strong iteration; four low-cost fixes remain before this document is clean |

---

## 2. Issues Fixed Since Iteration 2

| Iteration-2 Issue | Status |
|-------------------|--------|
| Issue 1: user:* route tests missing | **Fixed** — Per-Layer Test Plan now reads "4 个新 role:* 码 + 3 个变更 user:* 绑定"; Key Test Scenarios lists 7 backend tests (cases 5–7 cover GET /admin/users, GET /admin/users/:userId, POST /admin/users) |
| Issue 2: Cross-Layer Data Map useless padding | **Fixed** — Section replaced by "Permission Code Change Classification" table with 7 rows distinguishing new/renamed/semantic-change codes, migration operations, route binding changes, and frontend reference changes |
| Issue 5: TeamManagementPage canReadRoles guard has no test | **Fixed** — Frontend test plan now includes `TeamManagementPage.test.tsx` verifying `listRolesApi` call count === 0 when `canReadRoles === false` |
| Architecture: No request execution flow | **Fixed** — "Request Execution Flow" section added with 5 numbered steps for `GET /admin/roles`, explicitly documenting middleware ordering and the `TeamScopeMiddleware` → `RequirePermission` dependency |

---

## 3. Structure Check

| Section | Required | Present | Status |
|---------|----------|---------|--------|
| Overview | Yes | Yes | Pass |
| Architecture | Yes | Yes (layer table + ASCII diagram + execution flow) | Pass |
| Interfaces | Yes | Yes (9 sections) | Pass |
| Data Models | Yes | Yes (no new tables, confirmed) | Pass |
| Permission Code Change Classification | No | Yes (new in iteration 3) | Bonus — replaces useless padding |
| Error Handling | Yes | Yes (4 error codes) | Pass — caller contract still absent |
| Testing Strategy | Yes | Yes (table + scenarios + overall target) | Pass — coverage target is a count |
| Security Considerations | Yes | Yes (3 threats) | Pass — 2 threats still shallow |
| Open Questions | Optional | Yes (2, both resolved) | Pass |
| Alternatives Considered | Optional | Yes (2 alternatives) | Pass |

---

## 4. Dimension-by-Dimension Findings

### Dimension 1: Architecture Clarity — 14/15 (A-)

**What's there:**
- Request Execution Flow section added with 5 numbered steps for `GET /admin/roles`.
- Explicit statement: "`TeamScopeMiddleware` 必须在 `RequirePermission` 之前注册" with the reason (context dependency).
- Layer Placement table covers all 9 files.
- ASCII component diagram shows static dependency tree.

**Deductions:**

**-1: Layer table descriptions are still change-lists, not responsibilities.** The column header remains "变更内容". For example, `migration/rbac.go` is described as "数据迁移函数（幂等，事务执行）" — this describes what it does in this change, not the layer's standing responsibility. A new developer reading this table cannot understand why the migration layer owns this concern or how it relates to the existing `SyncPresetRoles` call chain. The execution flow section now compensates for the runtime gap, but the layer table remains a diff summary rather than an architecture document.

---

### Dimension 2: Interface & Model Definitions — 17/20 (B+)

**What's there:**
- Permission Code Change Classification table: 7 rows, each distinguishing change type, old code, migration operation, route binding change, and frontend reference change. This is a genuine improvement — a developer can now see at a glance which codes need migration operations and which are purely additive.
- All 9 interface sections with concrete Go and TypeScript snippets.
- API Handbook: 10 endpoints with request/response tables and error codes.

**Deductions:**

**-2: `User` and `Team` response types are still undefined in the API Handbook.** `GET /admin/users` returns `User[]` and `GET /admin/teams` returns `Team[]`. The Data Contracts section defines only `RoleListItem`, `PermissionItem`, and `ResourcePermissions` — unchanged from iteration 1. Neither `User` nor `Team` is defined. A developer implementing the frontend must still infer the shape of the two most-used response types from the response field tables (which list `bizKey`, `username`, `displayName`, `email`, `phone`, `isSuperAdmin`, `status` for User but provide no TypeScript interface). This is a one-paragraph fix.

**-1: Migration function body is still pseudocode.** The migration logic remains a numbered list with pseudocode (`SELECT INTO 临时表或应用层记录`, `INSERT ... 忽略已存在`). The function signature is defined but the implementation contract — what GORM calls, what the idempotency check SQL looks like, what the `schema_migrations` insert looks like — is left to the implementer. The "修正后的步骤 3" note adds clarity on the user:read semantic but does not replace the pseudocode with actual GORM method calls.

---

### Dimension 3: Error Handling — 14/15 (A-)

**What's there:**
- Error Types & Codes table: 4 complete rows with HTTP status codes and descriptions.
- Propagation strategy: `RequirePermission` → `apperrors.RespondError` → JSON response.
- API Handbook error codes table consistent with tech-design.

**Deductions:**

**-1: No error handling for migration failure at the caller level.** The Mitigations section states "迁移在单一数据库事务中执行，失败自动回滚" but the Error Handling section still does not specify what the caller of `MigratePermissionGranularity` should do when it returns a non-nil error. The function signature returns `error`. Does the application fail to start? Log and continue? Alert ops? This is a deployment-critical question with a one-sentence answer. Three iterations have passed without this being addressed.

---

### Dimension 4: Testing Strategy — 14/15 (A-)

**What's there:**
- Per-Layer Test Plan: 4 rows; route middleware row now covers "4 个新 role:* 码 + 3 个变更 user:* 绑定".
- Key Test Scenarios: 7 backend tests (4 role:* + 3 user:*), 3 migration tests, frontend tests including `TeamManagementPage.test.tsx`.
- Overall Coverage Target section added.

**Deductions:**

**-1: Coverage target is still a count, not a percentage.** The Overall Coverage Target reads "新增 11 个测试用例（7 路由中间件 + 3 迁移 + 1 前端守卫）". This is a task count. It says nothing about whether the changed code paths are adequately covered. The project's established standard specifies percentage targets. 11 test cases for a refactor touching 9 files and 14 route bindings may or may not be sufficient — the document provides no way to evaluate this.

---

### Dimension 5: Breakdown-Readiness — 19/20 (A-)

**What's there:**
- PRD Coverage Map: 13 rows, all stories covered.
- 9 interface sections enumerate file-level changes.
- Both Open Questions resolved.
- Permission Code Change Classification table provides a clear migration action per code.

**Deductions:**

**-1: No explicit task count or enumeration.** The document enumerates changes across 9 sections but provides no summary count of discrete implementation tasks. A task breakdown author must manually count: 1 constants file, 1 router file, 1 migration function, 1 seedPresetRoles update, 4 frontend files, 1 router_test.go seed update = 8 discrete tasks minimum, plus 11 test cases. This is not stated. The reader must derive it.

---

### Dimension 6: Security Considerations — 14/15 (A-)

**What's there:**
- 3 threats: permission code semantic collision, migration partial execution, user:read privilege expansion.
- Threat 3 includes an explicit decision with rationale and mitigation.
- 4 mitigations: transaction rollback, snapshot, no-delete of user:read row, CI grep assertion.

**Deductions:**

**-1: Ordering dependency and lockout threats still absent.** Two threats identified in iteration 1 remain unaddressed after three iterations: (1) what happens if `MigratePermissionGranularity` is called before `SyncPresetRoles` — are the new codes registered in `permissions.Registry` in time for the migration to reference them? The execution flow section now documents the runtime middleware order but says nothing about startup initialization order. (2) Lockout: if a migration bug incorrectly removes all `role:*` permissions, no one can manage roles. Neither scenario is discussed or mitigated.

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

**13/13 stories covered.**

---

## 6. Top Issues

### Issue 1: User and Team Response Types Undefined (Dimension 2 — B+)

`GET /admin/users` returns `User[]` and `GET /admin/teams` returns `Team[]`. The Data Contracts section defines only `RoleListItem`, `PermissionItem`, `ResourcePermissions`. Fix: add `User` and `Team` TypeScript interface definitions to the Data Contracts section of the API Handbook. The fields are already listed in the response tables — this is a mechanical extraction.

### Issue 2: Migration Caller Error Contract Unspecified (Dimension 3 — A-)

`MigratePermissionGranularity` returns `error` but the document does not specify what the caller should do on failure. Fix: add one sentence to the Error Handling section, e.g., "调用方（`main.go` 启动流程）若收到非 nil error，应终止启动并记录错误日志，不允许静默忽略。" This has been flagged in all three iterations.

### Issue 3: Coverage Target Is a Count, Not a Percentage (Dimension 4 — A-)

"新增 11 个测试用例" is a task count, not a coverage target. Fix: replace with a statement like "变更代码行覆盖率目标 ≥ 80%，或明确说明为何计数目标等价于覆盖率目标。"

### Issue 4: Ordering Dependency and Lockout Threats Absent (Dimension 6 — A-)

The threat model does not address startup initialization order (`SyncPresetRoles` vs `MigratePermissionGranularity`) or the lockout scenario if migration removes all `role:*` permissions. Fix: add two sentences to the Threat Model — one stating the required call order and where it is enforced, one acknowledging the lockout risk and pointing to the snapshot mitigation.

---

## 7. Verdict

**Overall Score: 92/100 — A-**

Iteration 3 addressed four of the five iteration-2 issues. The execution flow section is now the strongest part of the architecture section. The Permission Code Change Classification table is a genuine improvement over the useless Cross-Layer Data Map. The test plan now covers all 7 changed route bindings and the `canReadRoles` guard.

**Breakdown-Readiness: Yes.** All 13 PRD stories are covered. The four remaining issues are quality gaps, not blockers. A developer can proceed to `/breakdown-tasks` immediately. The two highest-priority fixes (User/Team types, migration caller contract) are each one paragraph and should be resolved before implementation begins.
