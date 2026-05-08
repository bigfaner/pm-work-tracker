---
created: 2026-05-08
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 统一权限码校验，移除 bypass 模式

## Overview

Remove all bypass logic and unify frontend + backend to permission-code-only authorization. The `is_super_admin` column is **kept** in the DB as a flag to determine which permission codes to load — not as a bypass mechanism. When a user is SuperAdmin, all 29 permission codes are loaded. Every authorization check (middleware, handler, service, frontend) goes through the same permCodes path.

Changes span: `RequirePermission` tier-1 bypass removal, `TeamScopeMiddleware` injects all 29 codes (not empty), handler-level `isPMOrSuperAdmin` removal, service PM identity check removal, frontend `isSuperAdmin` cleanup.

## Architecture

### Layer Placement

| Layer | Change Type |
|-------|-------------|
| Model | No change (`IsSuperAdmin` field kept) |
| Migration/Seed | `seedPresetRoles` update (superadmin gets 29 codes) |
| Middleware (auth) | No change (keeps `isSuperAdmin` context — used by TeamScopeMiddleware to load all codes) |
| Middleware (permission) | Remove tier-1 SuperAdmin bypass |
| Middleware (team_scope) | SuperAdmin: inject all 29 codes instead of empty; still skip team membership check |
| Handler | Remove `isPMOrSuperAdmin()`, assignee checks, PM substitution |
| Service | Remove PM identity checks, simplify signatures |
| Service (role) | `GetUserPermissions` handles SuperAdmin specially; remove `IsSuperAdmin` from response |
| VO/DTO | Remove `IsSuperAdmin` from response DTOs |
| Frontend store | Remove `isSuperAdmin` state + `hasPermission` bypass |
| Frontend types | Remove `isSuperAdmin` from type definitions |
| Frontend components | Remove all `isSuperAdmin` references |

### Component Diagram

```
┌──────────────────────────────────────────────────────┐
│                 Request Flow (AFTER)                  │
│                                                      │
│  Request                                             │
│    │                                                 │
│    ▼                                                 │
│  AuthMiddleware ─── sets isSuperAdmin in context     │
│    │               (used to load all codes)           │
│    ▼                                                 │
│  TeamScopeMiddleware                                 │
│    ├─ SuperAdmin? ──► inject ALL 29 permCodes       │
│    └─ Member? ──► load role + permCodes              │
│    │                                                 │
│    ▼                                                 │
│  RequirePermission(code)                             │
│    ├─ Tier 1: permCodes contains code → pass/403    │
│    └─ Tier 2: DB query → pass/403                    │
│    │                                                 │
│    ▼                                                 │
│  Handler ─── no isPMOrSuperAdmin, no assignee check  │
│    │        no PM BizKey substitution                 │
│    ▼                                                 │
│  Service ─── no PM identity checks, simplified sigs  │
│    │                                                 │
│    ▼                                                 │
│  Response                                            │
│                                                      │
│  ─── Separate path ───                               │
│                                                      │
│  GET /me/permissions                                 │
│    └─ SuperAdmin? ──► return all 29 codes all teams  │
│    └─ Normal? ──► query team_members → role → codes  │
└──────────────────────────────────────────────────────┘
```

### Dependencies

| Dependency | Usage | New? |
|------------|-------|------|
| `pkg/permissions` | `AllCodeStrings()` for `/me/permissions` SuperAdmin path | Existing |

## Interfaces

### Interface 1: RequirePermission — Tier-1 Bypass Removed

Remove the SuperAdmin short-circuit. All users now pass through permCodes check uniformly.

```go
// BEFORE (3 tiers):
func RequirePermission(code string, roleRepo repository.RoleRepo) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Tier 1: SuperAdmin bypass — REMOVED
        if IsSuperAdmin(c) { c.Next(); return }

        // Tier 2: team-context permCodes check
        // Tier 3: non-team DB query
    }
}

// AFTER (2 tiers):
func RequirePermission(code string, roleRepo repository.RoleRepo) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Tier 1: permCodes contains code → pass/403
        // Tier 2: non-team DB query → pass/403
    }
}
```

SuperAdmin passes tier-1 because TeamScopeMiddleware injects all 29 codes.

### Interface 2: TeamScopeMiddleware — SuperAdmin Injects All Codes

```go
// BEFORE:
if IsSuperAdmin(c) {
    c.Set("permCodes", []string{})      // empty slice
    c.Next()
    return
}

// AFTER:
if IsSuperAdmin(c) {
    c.Set("callerTeamRole", "superadmin")
    c.Set("permCodes", permissions.AllCodeStrings())  // all 29 codes
    c.Next()
    return
}
```

SuperAdmin still skips team membership check, but now receives all 29 codes so `RequirePermission` tier-1 (permCodes check) passes for any code.

### Interface 3: RoleService.GetUserPermissions — Modified

Remove `IsSuperAdmin` from response. For SuperAdmin users, return all 29 codes for all teams.

```go
// BEFORE:
type UserPermissions struct {
    IsSuperAdmin    bool
    TeamPermissions map[int64][]string
}

func (s *roleService) GetUserPermissions(ctx, userID) (*UserPermissions, error) {
    user := userRepo.FindByID(ctx, userID)
    teamPerms := roleRepo.GetUserTeamPermissions(ctx, user.BizKey)
    return &UserPermissions{IsSuperAdmin: user.IsSuperAdmin, TeamPermissions: teamPerms}
}

// AFTER:
type UserPermissions struct {
    TeamPermissions map[int64][]string
}

func (s *roleService) GetUserPermissions(ctx, userID) (*UserPermissions, error) {
    user := userRepo.FindByID(ctx, userID)
    if user.IsSuperAdmin {
        // Return all 29 codes for all teams
        allCodes := permissions.AllCodeStrings()
        teams := teamRepo.ListTeamBizKeys(ctx)
        teamPerms := map[int64][]string{}
        for _, bizKey := range teams {
            teamPerms[bizKey] = allCodes
        }
        return &UserPermissions{TeamPermissions: teamPerms}, nil
    }
    teamPerms := roleRepo.GetUserTeamPermissions(ctx, user.BizKey)
    return &UserPermissions{TeamPermissions: teamPerms}, nil
}
```

**New dependency**: `roleService` needs `teamRepo` for the SuperAdmin path. Typed interface addition:

```go
// Add to teamRepo interface:
ListTeamBizKeys(ctx context.Context) ([]int64, error)
```

### Interface 4: TeamService — Simplified Signatures

Remove `pmBizKey` parameters used for PM identity checks. The middleware enforces authorization.

```go
// BEFORE:
UpdateTeam(ctx, pmBizKey, teamBizKey, req) (*Team, error)
RemoveMember(ctx, pmBizKey, teamBizKey, targetUserBizKey) error
TransferPM(ctx, currentPMBizKey, teamBizKey, newPMBizKey) error
DisbandTeam(ctx, callerBizKey, teamBizKey, confirmName) error
UpdateMemberRole(ctx, pmBizKey, targetUserBizKey, teamBizKey, roleBizKey) error
ListTeams(ctx, callerID, isSuperAdmin, search, page, pageSize) ([]*TeamListResp, int64, error)

// AFTER:
UpdateTeam(ctx, teamBizKey, req) (*Team, error)
RemoveMember(ctx, teamBizKey, targetUserBizKey) error
TransferPM(ctx, teamBizKey, newPMBizKey) error
DisbandTeam(ctx, teamBizKey, confirmName) error
UpdateMemberRole(ctx, targetUserBizKey, teamBizKey, roleBizKey) error
ListTeams(ctx, search, page, pageSize) ([]*TeamListResp, int64, error)
```

Internal method changes:
- Remove all `team.PmKey != pmBizKey` checks from UpdateTeam, RemoveMember, TransferPM, DisbandTeam, UpdateMemberRole (5 methods — authorization now enforced by middleware)
- `InviteMember`: keep the PM-role check for the business rule "PM role can only be assigned via TransferPM". Inline the role-name comparison directly (`role.Name == "pm"`) rather than calling the removed `isPMRole()` helper

### Interface 5: ProgressService.Append — Renamed Parameter

```go
// BEFORE:
Append(ctx, teamBizKey, authorBizKey, subItemID, completion, achievement, blocker, lesson, isPM bool) (*ProgressRecord, error)

// AFTER:
Append(ctx, teamBizKey, authorBizKey, subItemID, completion, achievement, blocker, lesson, skipRegressionCheck bool) (*ProgressRecord, error)
```

The `skipRegressionCheck` flag is sourced from `sub_item:assign` permCode in the handler. Behavior identical: PM and SuperAdmin bypass regression check; others don't.

### Interface 6: seedPresetRoles — Updated

```go
// BEFORE:
seedRole(tx, "superadmin", "系统超级管理员，绕过所有权限检查", true, nil)

// AFTER:
seedRole(tx, "superadmin", "系统超级管理员", true, permissions.AllCodeStrings())
```

The `seedRole` function is additive — on next startup, it inserts the 29 codes for superadmin role. No migration needed.

## Data Models

Single-layer change (backend internal only). No schema changes.

The `is_super_admin` column remains in `pmw_users`. The superadmin role gains 29 rows in `pmw_role_permissions` (via `seedRole` additive sync).

### Field Quick Reference

| Model | Change | Key Fields | Notes |
|-------|--------|------------|-------|
| pmw_users | No change | is_super_admin kept | Flag to determine permCodes loading |
| pmw_role_permissions | DATA: +29 rows | role_key → superadmin, permission_code | Seeded by updated `seedPresetRoles` |

## Error Handling

### Error Types & Codes

No new error types. Existing errors preserved.

| Error Code | HTTP Status | When |
|------------|-------------|------|
| `ErrForbidden` | 403 | User lacks required permission code |
| `ErrNotTeamMember` | 403 | Non-member accessing team resource |
| `ErrProgressRegression` | 400 | Completion decreasing (when `skipRegressionCheck=false`) |

### Propagation Strategy

Unchanged. Errors flow from repository → service → handler → response via `apperrors`.

### Error Path Migration

The design removes handler-level and service-level authorization checks. This table maps every removed error path to its replacement, confirms the user-facing behavior change, and ties each change to the PRD acceptance criteria that justify it.

#### Removed: Assignee ownership checks (sub_item_handler.go)

| Removed Code Path | Old Behavior | New Behavior | User-Facing Change | PRD AC |
|-------------------|-------------|-------------|-------------------|--------|
| `Update` L157-169: `if !isPMOrSuperAdmin { if item.AssigneeKey != callerBizKey → ErrForbidden }` | Non-PM user without `sub_item:assign` who is not the assignee gets 403 (`ErrForbidden`, code `FORBIDDEN`, message "insufficient permissions") | Check removed. User with `sub_item:update` permCode passes `RequirePermission` middleware and reaches handler → 200 | Users with `sub_item:update` can now update any sub-item regardless of assignee status. Users without `sub_item:update` still get 403 from middleware (same `ErrForbidden`). | AC1: This IS the intended fix — custom roles with `sub_item:update` must not be blocked |
| `ChangeStatus` L211-223: `if !isPMOrSuperAdmin { if item.AssigneeKey != callerBizKey → ErrForbidden }` | Same as above but for status changes | Same as above — check removed | Users with `sub_item:change_status` can now change status of any sub-item. Users without it still get 403 from middleware. | AC2: This IS the intended fix |

#### Removed: PM BizKey substitution (team_handler.go)

| Removed Code Path | Old Behavior | New Behavior | User-Facing Change | PRD AC |
|-------------------|-------------|-------------|-------------------|--------|
| `Update` L99-106: `if IsSuperAdmin { pmBizKey = team.PmKey }` | SuperAdmin's BizKey replaced with actual PM's BizKey so service-layer `team.PmKey != pmBizKey` check passes | Substitution removed. Service no longer checks PM identity — `UpdateTeam(ctx, teamBizKey, req)` has no `pmBizKey` param | SuperAdmin with `team:update` permCode passes middleware and reaches handler → 200 (was already 200). PM with `team:update` → 200 (unchanged). Non-PM without `team:update` → 403 from middleware (was 403 from service, now from middleware — same HTTP status, same error code `FORBIDDEN`) | AC3a, AC6 |
| `Disband` L128-134: same pattern | SuperAdmin BizKey substituted to PM | Same as above for `DisbandTeam` | Same pattern: authorization shifts from service `team.PmKey != callerBizKey` → ErrForbidden to middleware `RequirePermission("team:delete")` → ErrForbidden | AC3a, AC6 |
| `TransferPM` L276-283: same pattern | SuperAdmin BizKey substituted to PM | Same as above for `TransferPM` | Same shift from service 403 to middleware 403 | AC3a, AC6 |
| `RemoveMember` L194 (implied): same pattern | SuperAdmin BizKey substituted to PM | Same as above for `RemoveMember` | Same shift | AC3a, AC6 |
| `UpdateMemberRole` L233 (implied): same pattern | SuperAdmin BizKey substituted to PM | Same as above | Same shift | AC3a, AC6 |

#### Removed: Service-layer PM identity checks (team_service.go)

| Removed Code Path | Old Behavior | New Behavior | User-Facing Change |
|-------------------|-------------|-------------|-------------------|
| `UpdateTeam` L159: `if team.PmKey != pmBizKey → ErrForbidden` | Non-PM caller gets 403 | Check removed. Method signature no longer takes `pmBizKey` | Users without `team:update` get 403 from middleware instead of service. Same status code, different layer. |
| `RemoveMember` L208: `if team.PmKey != pmBizKey → ErrForbidden` | Non-PM caller gets 403 | Check removed | Same as above for `team:remove` |
| `TransferPM` L223: `if team.PmKey != currentPMBizKey → ErrForbidden` | Non-PM caller gets 403 | Check removed | Same as above for `team:transfer` |
| `DisbandTeam` L272: `if team.PmKey != callerBizKey → ErrForbidden` | Non-PM caller gets 403 | Check removed | Same as above for `team:delete` |
| `UpdateMemberRole` L287: `if team.PmKey != pmBizKey → ErrForbidden` | Non-PM caller gets 403 | Check removed | Same as above for `team:update_role` |

#### Removed: pmFlag / isPMOrSuperAdmin (progress_handler.go)

| Removed Code Path | Old Behavior | New Behavior | User-Facing Change | PRD AC |
|-------------------|-------------|-------------|-------------------|--------|
| `Append` L86: `pmFlag := isPMOrSuperAdmin(c)` passed to service | `pmFlag=true` bypasses regression check; `pmFlag=false` enforces it | `pmFlag` replaced with `skipRegressionCheck bool`, sourced from `sub_item:assign` permCode in handler | Users with `sub_item:assign` skip regression check (unchanged). Users with `progress:create` but not `sub_item:assign` get regression check enforced. Previously these users were blocked entirely by `isPMOrSuperAdmin` returning false → the handler never reached the service. Now they reach the service and regression check applies. | AC7: Custom role with `progress:create` can now create progress. |

#### Summary of user-facing behavior changes

1. **Custom roles with `sub_item:update`**: previously 403 (assignee check), now 200. **This is the core fix (AC1).**
2. **Custom roles with `sub_item:change_status`**: previously 403 (assignee check), now 200. **This is the core fix (AC2).**
3. **Custom roles with `progress:create`**: previously blocked by `isPMOrSuperAdmin` (pmFlag=false caused no functional regression check issue but the handler behavior differed), now proceed normally. **This is the core fix (AC7).**
4. **Non-PM users calling team management endpoints**: previously 403 from service (`ErrForbidden`), now 403 from middleware (`ErrForbidden`). Same HTTP response to the client.
5. **SuperAdmin users**: no behavior change. Middleware loads all 29 codes; RequirePermission passes via tier-1 (permCodes check); all operations return same status codes as before.

#### Edge case: `GetUserPermissions` with no teams

When `ListTeamBizKeys(ctx)` returns an empty slice (no teams exist in the system), `GetUserPermissions` returns an empty `TeamPermissions` map with `nil` error. This is correct: a SuperAdmin in a fresh system has no teams, so `hasPermission()` in the frontend evaluates against zero permission entries and returns false for everything. No special error needed — the empty map is the correct response.

## Cross-Layer Data Map

The `isSuperAdmin` field is **kept in backend model** but **removed from API responses and frontend**.

| Field Name | Storage Layer | Backend Model | API/DTO | Frontend Type | Change |
|------------|---------------|---------------|---------|---------------|--------|
| `is_super_admin` / `isSuperAdmin` | `pmw_users.is_super_admin` KEPT | `User.IsSuperAdmin` KEPT | `json:"isSuperAdmin"` **REMOVED** | `isSuperAdmin: boolean` **REMOVED** | Removed from API & frontend only |

Post-removal, the frontend derives SuperAdmin capabilities from permission codes:
- `/me/permissions` returns all 29 codes for all teams → `hasPermission()` returns true for everything
- No special SuperAdmin UI path needed

## Frontend Change File Enumeration

Every file containing `isSuperAdmin` references in `frontend/src/`, grouped by module. Files marked **(source)** contain production logic that must be rewritten; files marked **(test)** contain test fixtures that must be updated to remove `isSuperAdmin` fields and assertions.

### Types & Store (core — changes affect all downstream consumers)

| File | Type | `isSuperAdmin` refs | Change Required |
|------|------|---------------------|-----------------|
| `src/types/index.ts` | source | 4 (lines 4, 70, 433, 494) | Remove `isSuperAdmin: boolean` from `UserInfo`, `PermissionsResponse`, `LoginResponse`, and the 4th interface |
| `src/store/auth.ts` | source | 4 (lines 10, 28, 37, 44, 62) | Remove state field, setter, login handler, and `hasPermission` bypass (`if permissions.isSuperAdmin return true`) |
| `src/mocks/handlers.ts` | source | 1 (line 18) | Remove `isSuperAdmin` from mock response fixture |
| `src/types/index.test.ts` | test | 3 | Remove `isSuperAdmin` from test fixtures and assertions |
| `src/store/auth.test.ts` | test | 18 | Remove all `isSuperAdmin`-related test cases and fixtures |

### Page Components (production logic)

| File | Type | `isSuperAdmin` refs | Change Required |
|------|------|---------------------|-----------------|
| `src/pages/UserManagementPage.tsx` | source | 3 (lines 287, 546, 557) | Replace `const isSuperAdmin = useAuthStore(...)` with `hasPermission("user:manage_role")` or equivalent; update conditional rendering |

### Test Files (mock state setup)

| File | `isSuperAdmin` refs |
|------|---------------------|
| `src/App.test.tsx` | 3 |
| `src/api/client.test.ts` | 2 |
| `src/__tests__/permission-driven-ui.test.tsx` | 9 |
| `src/components/PermissionGuard.test.tsx` | 4 |
| `src/components/layout/Sidebar.test.tsx` | 5 |
| `src/components/layout/AppLayout.test.tsx` | 1 |
| `src/pages/UserManagementPage.test.tsx` | 14 |
| `src/pages/TeamManagementPage.test.tsx` | 1 |
| `src/pages/TeamDetailPage.test.tsx` | 3 |
| `src/pages/MainItemDetailPage.test.tsx` | 2 |
| `src/pages/SubItemDetailPage.test.tsx` | 1 |
| `src/pages/ItemViewPage.test.tsx` | 1 |
| `src/pages/ItemPoolPage.test.tsx` | 4 |
| `src/pages/ReportPage.test.tsx` | 1 |
| `src/pages/LoginPage.test.tsx` | 1 |
| `src/pages/main-item-detail/DecisionTimeline.test.tsx` | 2 |

### Summary for Task Breakdown

- **Task Group A (1 task)**: `types/index.ts` — remove 4 field declarations; triggers TypeScript compile errors across codebase
- **Task Group B (1 task)**: `store/auth.ts` — remove state field + `hasPermission` bypass; highest risk (auth logic)
- **Task Group C (1 task)**: `UserManagementPage.tsx` — replace `isSuperAdmin` usage with `hasPermission()`
- **Task Group D (1 task)**: `mocks/handlers.ts` — remove from mock fixture
- **Task Group E (batch, ~4 tasks of 4 files each)**: Update all 16 test files — remove `isSuperAdmin` from mock state setup objects; mechanical change

## Integration Specs

No existing-page integrations — not applicable. No new UI components; only removing `isSuperAdmin` references.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Handler | Unit | `testify/mock` for repository mocks, `testify/assert` for assertions; runner: `cd backend && go test ./internal/handler/...` | sub_item/progress handlers without assignee checks; team handler without PM substitution | 85% |
| Service | Unit | `testify/mock` for repository mocks, `testify/assert`; runner: `cd backend && go test ./internal/service/...` | Simplified signatures; no PM identity checks | 80% |
| RoleService | Unit | `testify/mock` for `teamRepo`/`roleRepo`; runner: `cd backend && go test ./internal/service/... -run TestRoleService` | `GetUserPermissions` SuperAdmin path returns all 29 codes | 90% |
| RBAC | Integration | `net/http/httptest` server backed by in-memory SQLite (via GORM `:memory:` DSN) to exercise real DB queries; runner: `cd backend && go test ./internal/handler/... -run TestRBAC` | Full permission matrix for superadmin/PM/member/custom roles | 90% |
| Frontend store | Unit | `vitest`; runner: `npx vitest run src/store/auth.test.ts` | `hasPermission()` without SuperAdmin bypass | 90% |
| Frontend UI | Integration | `vitest` + `@testing-library/react`; runner: `npx vitest run` | Permission-driven UI visibility | 80% |

### Key Test Scenarios

1. **SuperAdmin full flow**: SuperAdmin performs all 22 operations (AC 3a/3b/3c) → all return expected status codes (all 29 codes loaded, permCodes check passes)
2. **Custom role sub_item:update**: ext-member with `sub_item:update` edits unassigned sub-item → 200 (core fix)
3. **Custom role sub_item:update missing**: ext-member without `sub_item:update` → 403
4. **SuperAdmin /me/permissions**: Returns all 29 codes for all teams
5. **PM team management**: PM performs invite/remove/transfer → all succeed (middleware passes via `team:invite/remove/transfer` codes)
6. **Custom role progress:create**: ext-member with `progress:create` adds progress to unassigned sub-item → 200
7. **PM team ops without PM substitution**: PM calls UpdateTeam without PM BizKey → 200 (middleware enforces `team:update`)
8. **Frontend cleanup**: No `isSuperAdmin` references; TypeScript compiles; all tests pass
9. **seedPresetRoles**: Superadmin role has 29 codes after startup
10. **GetUserPermissions normal user**: Non-SuperAdmin user gets team-specific codes (unchanged behavior)

### Overall Coverage Target

85%

## Security Considerations

### Threat Model

| Threat | Risk | Mitigation (verification mechanism) |
|--------|------|--------------------------------------|
| Custom role gains unintended access via assignee check removal | Low | RBAC integration test scenario 2 asserts: user with only `sub_item:view` receives 403 on update — confirming no permission escalation beyond granted codes. CI runs this matrix on every push. |
| Frontend UI shows too many elements to non-SuperAdmin | Low | Frontend integration test suite asserts each UI element appears only when the corresponding permission code is present in the store. `permission-driven-ui.test.tsx` (9 existing assertions) is extended to cover every removed `isSuperAdmin` guard. |
| SuperAdmin detection still relies on DB column | Info | `IsSuperAdmin` removed from all response DTOs (verified by RoleService unit test asserting `UserPermissions` struct has no `IsSuperAdmin` field). API contract test confirms `/me/permissions` response JSON contains no `isSuperAdmin` key. Column is unreachable from outside the backend. |

### Mitigations

- **Threat 1 (assignee check removal)**: The RBAC integration test suite (`go test ./internal/handler/... -run TestRBAC`) runs a permission matrix covering superadmin/PM/member/custom-role against all 22 operations (AC 3a/3b/3c). Scenario 2 (`sub_item:view`-only user attempts update) and scenario 3 (user without `sub_item:update` gets 403) serve as the security regression gate — any permission escalation beyond the granted codes causes a test failure.
- **Threat 2 (frontend UI overexposure)**: Frontend integration tests (`npx vitest run src/__tests__/permission-driven-ui.test.tsx`) verify each UI element's visibility is determined solely by `hasPermission(code)`. The removal of all `isSuperAdmin` references is validated by TypeScript compilation (no undefined-property access) and by the grep-based audit in test scenario 8.
- **Threat 3 (SuperAdmin DB column exposure)**: `IsSuperAdmin` is absent from the `UserPermissions` response struct and all JSON DTOs. Unit test for `GetUserPermissions` asserts the returned struct contains only `TeamPermissions`. `seedPresetRoles` is additive (only inserts, never deletes) and the superadmin role is a preset protected from API mutation.

## Scope Alignment with PRD

No divergence. The design fully aligns with the PRD scope:

- **In Scope items**: All addressed as specified (RequirePermission bypass removed, handler-level checks removed, frontend isSuperAdmin removed, VO/DTO fields removed, seed updated)
- **Intentionally Kept items**: All respected (`User.IsSuperAdmin` model field kept, `AuthMiddleware` context kept, `config/seed.go` kept, `TeamScopeMiddleware` SuperAdmin team-membership skip kept)

**What this means for task breakdown**: No migration task to drop the column. The `is_super_admin` column is used by TeamScopeMiddleware (load all 29 codes) and `GetUserPermissions` (return all 29 codes). It is NOT used as a bypass mechanism.

## PRD Coverage Map

| PRD AC | Design Component | Status |
|--------|-----------------|--------|
| AC1: Custom role `sub_item:update` → 200 | Remove assignee check from sub_item_handler.Update | Addressed |
| AC1: Custom role without `sub_item:update` → 403 | RequirePermission tier-1 (permCodes check) | Addressed |
| AC2: Custom role `sub_item:change_status` → 200 | Remove assignee check from sub_item_handler.ChangeStatus | Addressed |
| AC2: Custom role without → 403 | RequirePermission tier-1 (permCodes check) | Addressed |
| AC3a-c: SuperAdmin all ops | TeamScopeMiddleware injects all 29 codes + RequirePermission permCodes check | Addressed |
| AC4: SuperAdmin cross-team → 200 | TeamScopeMiddleware SuperAdmin path (inject all codes, skip membership) | Addressed |
| AC4: Non-member non-superadmin → 403 | TeamScopeMiddleware normal flow | Addressed |
| AC5: No `isSuperAdmin` in frontend | Remove from types, store, components | Addressed |
| AC5: `/me/permissions` returns 29 codes | GetUserPermissions SuperAdmin path | Addressed |
| AC6: PM ops unaffected | PM role has `team:invite/remove/transfer`; service no longer checks PM identity | Addressed |
| AC7: Custom role `progress:create` → 200 | Remove pmFlag gate; RequirePermission tier-1 | Addressed |
| AC7: Custom role without → 403 | RequirePermission tier-1 (permCodes check) | Addressed |
| PRD: Keep `User.IsSuperAdmin` model field | Column kept in DB and backend model; removed from API DTOs and frontend | **Aligned** — see Scope Alignment section |
| PRD: Remove `RequirePermission` short-circuit | Tier-1 bypass removed; SuperAdmin passes via all 29 codes in permCodes | Addressed |
| PRD: Remove `AuthMiddleware` isSuperAdmin context | Context value kept (used by TeamScopeMiddleware to load all codes) | Addressed |

## Open Questions

None — all decisions resolved during design.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Remove `is_super_admin` column entirely (early draft) | Pure RBAC; no special cases | Requires migration; team_members insertion for every team; bootstrapping problem for seed admin | User decision: keep column as permCodes-loading flag, use permCodes for all checks |
| Keep `isPMOrSuperAdmin()` but fix it | Minimal change | Still blocks custom roles without `sub_item:assign` | Core fix requires removing the handler-level check |
| Add global role assignment table | Clean separation of global vs team roles | Significant schema change for one special case | Over-engineering; middleware bypass is sufficient |
