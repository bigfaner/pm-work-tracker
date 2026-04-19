---
feature: rbac-permissions
evaluated_at: 2026-04-19
---

# Design Evaluation Report: RBAC 权限体系

## 1. Summary Table

| Dimension | Grade | Notes |
|-----------|-------|-------|
| Architecture Clarity | A | Layer placement explicit, diagram present, dependencies listed, consistent with codebase |
| Interface & Model Definitions | A | All interfaces typed with concrete signatures, all models concrete with SQL DDL |
| Error Handling | A | Custom error codes defined, HTTP status mapped, propagation strategy clear |
| Testing Strategy | B | Per-layer plan with test types and tooling; coverage target present but integration test tooling not named |
| Breakdown-Readiness | A | All components enumerable, tasks clearly derivable, PRD fully covered |
| Security Considerations | A | Threat model identifies 5 threats, each with concrete mitigations |
| **Overall** | **A** | All required dimensions A/B, 4 A's, Breakdown-Readiness A |

## 2. Structure Check

| Section | Required | Present | Notes |
|---------|----------|---------|-------|
| Overview | Yes | Yes | High-level approach, tech stack, and change scope table |
| Architecture | Yes | Yes | Layer placement, component diagram (ASCII), middleware chain, dependencies table |
| Interfaces | Yes | Yes | PermissionCodeRegistry, RoleService, RoleRepo, UserPermissions -- all with typed signatures |
| Data Models | Yes | Yes | Role, RolePermission DDL; TeamMember/User modifications; ER diagram; seed data |
| Error Handling | Yes | Yes | 5 new error codes with HTTP status mapping; permission denied response format |
| Testing Strategy | Yes | Yes | Unit tests per component, integration test scenarios, coverage targets (85%/90%) |
| Security Considerations | Yes (PRD has auth) | Yes | 5-row threat model table, defense-in-depth layers |
| Open Questions | Optional | Yes | 2 open items (SQLite migration, polling interval) |
| Alternatives Considered | Optional | Yes | 4 alternatives with pros/cons/why-not |

**All required sections present.**

## 3. Detailed Findings

### Dimension 1: Architecture Clarity -- Grade A

**Evidence:**
- Layer placement is explicit with a file-tree mapping (tech-design.md lines 28-43): router, middleware, handler, service, repository, and pkg/permissions are each named with their responsibility.
- ASCII component diagram (lines 47-69) shows the data flow: Router -> RequirePermission -> Handler, with TeamScopeMiddleware loading perm codes, all flowing down to Repository Layer and Database.
- Middleware chain (lines 72-93) documents the exact request flow through AuthMiddleware -> TeamScopeMiddleware -> RequirePermission(code) with numbered steps for each middleware's behavior.
- Dependencies table (lines 96-103) lists 4 existing packages; explicitly states "No new external dependencies needed."
- Consistent with codebase: router.go confirms the gin-based routing pattern, middleware/auth.go confirms JWT-based AuthMiddleware, middleware/rbac.go confirms RequireRole/RequireTeamRole pattern that will be replaced, middleware/team_scope.go confirms TeamScopeMiddleware pattern that will be extended.

### Dimension 2: Interface & Model Definitions -- Grade A

**Evidence:**
- `PermissionCodeRegistry` (lines 109-138): typed as `map[string][]ActionDef` with `AllCodes()` and `ValidateCode()` function signatures. Example codes provided inline.
- `RoleService` interface (lines 143-153): 7 methods with typed params and return values (`ctx context.Context`, `uint`, `*Role`, `CreateRoleReq`, `UpdateRoleReq`, `*UserPermissions`).
- `RoleRepo` interface (lines 157-173): 10 methods covering CRUD, permission bindings (`ListPermissions`, `SetPermissions`), and usage counting (`CountMembersByRoleID`).
- `UserPermissions` response struct (lines 177-182): concrete Go struct with JSON tags.
- Data Models (lines 185-267): Full SQL DDL for `roles` and `role_permissions` tables, ALTER statements for `team_members` and `users`, ER diagram, and seed data table.
- API Handbook complements with Go response structs: `RoleDetailResp`, `PermissionItem`, `PermissionGroup`, `UserPermissionsResp` (api-handbook.md lines 352-393).
- Frontend types defined: `PermissionMap`, `Role`, `PermissionGroup` (tech-design.md lines 403-532).
- A developer can implement directly from these definitions without guessing.

### Dimension 3: Error Handling -- Grade A

**Evidence:**
- 5 new error codes defined with HTTP status mapping (tech-design.md lines 537-545): `ERR_ROLE_NOT_FOUND` (404), `ERR_ROLE_NAME_EXISTS` (409), `ERR_ROLE_IN_USE` (422), `ERR_PRESET_ROLE_IMMUTABLE` (403), `ERR_INVALID_PERMISSION_CODE` (400).
- API Handbook adds 3 more codes (lines 395-405): `ERR_CANNOT_MODIFY_SELF` (403), `ERR_ALREADY_MEMBER` (422), `ERR_FORBIDDEN` (403).
- Permission denied response format is explicit with JSON example (line 549-554).
- Propagation strategy is clear: middleware returns 403, handler does business-rule checks (e.g., assignee), service does ownership checks.
- Consistent with existing error pattern: codebase uses `apperrors.RespondError(c, apperrors.ErrForbidden)` which matches the design's approach.

### Dimension 4: Testing Strategy -- Grade B

**Evidence:**
- Unit tests specified per component with file paths (lines 561-568): 6 test files covering permissions registry, middleware, service, repo, and JWT.
- Integration tests specified with 4 scenarios (lines 571-577): role CRUD, invite with role selection, permission-driven access, data migration.
- Coverage target: >=85% for new code, >=90% for modified middleware (line 580).
- Tooling implicitly named via file paths (Go standard testing), but integration test tooling (e.g., test HTTP server setup, test database) is not explicitly named.

**Gap:** Integration test infrastructure (how to set up test DB, test HTTP server for API integration tests) is not specified. This is minor since Go's `httptest` package is standard practice, but naming it would strengthen the plan.

### Dimension 5: Breakdown-Readiness -- Grade A

**Evidence:**
- Enumerable components:
  - 2 new DB tables: `roles`, `role_permissions`
  - 2 DB schema modifications: `team_members` (add `role_id`), `users` (drop `can_create_team`)
  - 1 new Go package: `pkg/permissions` (codes.go)
  - 1 new middleware: `RequirePermission`
  - 3 modified middlewares: AuthMiddleware, TeamScopeMiddleware, remove rbac.go
  - 2 new handlers: `role_handler.go`, `permission_handler.go`
  - 1 new service: `role_service.go`
  - 1 new repository: `role_repo.go`
  - 5 modified handlers (invite member, JWT claims, assignee pattern)
  - 3 frontend additions: `PermissionGuard`, `useHasPermission`, permission types
  - 1 frontend store modification: auth.ts
  - 6 new API endpoints + 3 modified endpoints
  - 1 data migration script
  - 1 seed data initialization

- PRD traceability: All 10 user stories and their acceptance criteria are addressed. See Section 4 for the full matrix.

- No ambiguous ownership: every file, interface, and model has a clear owner (middleware, handler, service, repository, or frontend).

### Dimension 6: Security Considerations -- Grade A

**Evidence:**
- Threat model (lines 587-594) identifies 5 threats with concrete mitigations:
  1. Frontend bypass -> RequirePermission middleware enforces server-side
  2. Permission code injection -> Go code validates code legitimacy
  3. Privilege escalation -> Role CRUD superadmin-only; DB-loaded permissions
  4. Horizontal privilege escalation -> TeamScopeMiddleware isolates by teamID
  5. Role deletion risk -> Usage count check before deletion
- Defense-in-depth pattern (lines 597-602): 3 layers documented (Router Middleware, Handler assignee check, Service ownership check).
- PRD security requirements (PRD section 8) are all addressed: backend as final authority, code-defined permission codes, JWT without permission fields, real-time DB queries.

### Codebase Consistency Findings

Comparing the design against actual source files:

| Design Assertion | Codebase Reality | Consistent? |
|-----------------|-------------------|-------------|
| Replace `RequireRole`/`RequireTeamRole` in middleware/rbac.go | File exists with both functions | Yes |
| AuthMiddleware sets `userRole` from JWT Claims.Role | auth.go line 38: `c.Set("userRole", claims.Role)` | Yes |
| JWT Claims has `Role` field to remove | jwt.go line 15: `Role string` field exists | Yes |
| TeamMember has `Role string` field to replace with `RoleID` | team.go line 22: `Role string` field exists | Yes |
| User has `CanCreateTeam bool` to remove | user.go line 12: `CanCreateTeam bool` exists | Yes |
| Admin routes use `RequireRole("superadmin")` | router.go line 134: confirmed | Yes |
| Frontend auth store has `isSuperAdmin` boolean | auth.ts line 9: `isSuperAdmin: boolean` | Yes |
| Frontend User type has `canCreateTeam` | types/index.ts line 15: `canCreateTeam: boolean` | Yes |
| Design says new JWT Claims: `{UserID, Username}` | Current: `{UserID, Role}` -- Role replaced by Username | Yes |
| Design says AuthMiddleware sets `username` | Current sets `userRole` -- will change | Yes |

**All design assertions are consistent with the current codebase.** No contradictions found.

## 4. PRD Traceability Matrix

### PRD Sections Coverage

| PRD Section | Design Coverage | Status |
|-------------|----------------|--------|
| 5.1 角色管理（超级管理员） | Router permission mapping, RoleService CRUD, RoleHandler, UI components 1-3 | Covered |
| 5.2 团队成员角色管理 | InviteMemberReq change (roleId), ChangeMemberRole endpoint, UI component 4 & 6 | Covered |
| 5.3 前端权限渲染 | PermissionGuard, useHasPermission, refresh strategy, UI component 5 | Covered |
| 5.4 预置角色定义 | Seed data table (3 preset roles), preset immutability rules in RoleService | Covered |
| 5.5 数据迁移 | Migration strategy described (SQLite rebuild), not a separate script spec but approach documented | Covered |
| 5.6 JWT Claims | Claims struct change documented (remove Role, add Username) | Covered |
| 5.7 关联改动 | All 6 items addressed: invite flow, middleware replacement, nav, admin, team create, store | Covered |

### User Story Acceptance Criteria Coverage

| Story | AC Count | Covered | Notes |
|-------|----------|---------|-------|
| Story 1: 超级管理员在线管理角色 | 8 | 8 | RoleService CRUD + preset protection rules + permission list endpoint |
| Story 2: PM 在邀请成员时指定角色 | 5 | 5 | InviteMemberReq with roleId + role list endpoint excluding superadmin |
| Story 3: 前端根据权限动态渲染 UI | 7 | 7 | PermissionGuard + useHasPermission + team-scoped permission checking |
| Story 4: 现有数据无缝迁移到 RBAC | 7 | 7 | Migration strategy with transaction + idempotency + field removal |
| Story 5: 团队创建权限控制 | 3 | 3 | team:create permission code + PermissionGuard on button |
| Story 6: PM 的权限驱动操作 | 5 | 5 | All permission codes mapped to routes in router changes table |
| Story 7: Member 的受限操作 | 6 | 6 | Permission codes + assignee pattern preserved |
| Story 8: 跨团队权限隔离 | 3 | 3 | TeamScopeMiddleware per-team isolation + team-scoped PermissionMap |
| Story 9: 后端权限强制执行 | 4 | 4 | RequirePermission middleware as Layer 1 defense |
| Story 10: 角色编辑即时生效 | 3 | 3 | DB-based permission loading (not JWT) + frontend refresh strategy |

**Total: 51/51 acceptance criteria covered.**

### UI Functions Traceability

| UI Function | Design Coverage | Status |
|-------------|----------------|--------|
| UI Function 1: 角色列表页 | UI Design Component 1 (RoleManagementPage) | Covered |
| UI Function 2: 角色编辑表单 | UI Design Component 2 (RoleEditDialog) | Covered |
| UI Function 3: 权限码浏览视图 | UI Design Component 3 (PermissionBrowseDialog) | Covered |
| UI Function 4: 邀请成员角色选择 | UI Design Component 4 (invite member Dialog extension) | Covered |
| UI Function 5: 权限驱动的 UI 渲染 | UI Design Component 5 (usePermission + PermissionGuard) | Covered |
| UI Function 6: 变更已有成员角色 | UI Design Component 6 (inline role change in member list) | Covered |

## 5. Action Items

### Mandatory Fixes (none required for proceeding)

No blocking issues found. All required dimensions pass at A or B.

### Optional Improvements

| Priority | Item | Rationale |
|----------|------|-----------|
| P2 | Name integration test tooling explicitly | Add `net/http/httptest` and test DB setup approach to testing strategy section. Minor gap, standard Go practice. |
| P2 | Specify data migration script as a separate component | The migration approach (SQLite rebuild) is described but not broken out as a named task with its own interface. Adding a `Migration` function signature would make task breakdown even more straightforward. |
| P3 | Add `ActionDef` struct definition | The `PermissionCodeRegistry` uses `[]ActionDef` but `ActionDef` struct is not explicitly defined in the design. It is inferable from context (`Code` and `Description` fields shown in usage) but would be more complete if shown. |
| P3 | Clarify `RequirePermission` for non-team routes | The middleware chain doc (line 92) says "If non-team context -> query DB: any role of this user has the code". The specific repo method for this query is not named in `RoleRepo`. Adding a `HasPermission(ctx, userID, code string) (bool, error)` method would close this gap. |

## 6. Verdict

**Overall Grade: A**

- All 6 required dimensions pass at A or B.
- 4 out of 6 dimensions are A (Architecture, Interfaces/Models, Error Handling, Security, Breakdown-Readiness).
- Breakdown-Readiness is A: all components are enumerable, tasks are clearly derivable, PRD is fully covered (51/51 AC).
- Design is consistent with the actual codebase across all checked files.
- API Handbook provides complete endpoint specifications with request/response examples and error code tables.
- UI Design document provides 6 detailed component specifications with layout, states, interactions, and data binding.

**Breakdown-Readiness: A -- can proceed to /breakdown-tasks.**

Top 3 strengths:
1. Complete PRD traceability with 51/51 acceptance criteria covered.
2. All interfaces are typed and directly implementable -- a developer can code from the design without guessing.
3. Full codebase consistency: every design assertion about current code matches the actual source.

Top 3 minor gaps (non-blocking):
1. Integration test tooling not explicitly named.
2. `ActionDef` struct not explicitly defined (inferable).
3. Non-team-context permission query method not named in RoleRepo interface.
