---
feature: "unify-permission-checks"
sources:
  - docs/features/unify-permission-checks/prd/prd-user-stories.md
  - docs/features/unify-permission-checks/prd/prd-spec.md
generated: "2026-05-09"
---

# Test Cases: unify-permission-checks

## Summary

| Type | Count |
|------|-------|
| UI   | 3   |
| **Integration** | **0** |
| API  | 22  |
| CLI  | 0   |
| **Total** | **25** |

> **Note**: This feature is a backend permission refactoring with frontend isSuperAdmin removal. API tests cover the core behavioral changes; UI tests cover frontend cleanup verification.

---

## UI Test Cases

### TC-001: TypeScript compilation passes after isSuperAdmin removal
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/build
- **Test ID**: ui/build/typescript-compilation-passes-after-issuperadmin-removal
- **Pre-conditions**: All frontend isSuperAdmin references have been removed from types, store, components, and mocks
- **Steps**:
  1. Run TypeScript compiler (`npx tsc --noEmit`)
  2. Verify no type errors related to missing `isSuperAdmin` field
- **Expected**: TypeScript compilation succeeds with zero errors
- **Priority**: P0

### TC-002: Frontend tests pass after isSuperAdmin removal
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/tests
- **Test ID**: ui/tests/frontend-tests-pass-after-issuperadmin-removal
- **Pre-conditions**: All isSuperAdmin references removed; mock data updated
- **Steps**:
  1. Run `npx vitest run`
  2. Verify all tests pass
- **Expected**: All frontend tests pass with zero failures
- **Priority**: P0

### TC-003: SuperAdmin sees all UI elements via hasPermission
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Target**: ui/permissions
- **Test ID**: ui/permissions/superadmin-sees-all-ui-elements-via-haspermission
- **Pre-conditions**: SuperAdmin user logged in; `/api/v1/me/permissions` returns all 29 permission codes
- **Route**: `/items`
- **Element**: L-003, L-004, L-005, L-006, L-007, L-008, L-009, L-010
- **Steps**:
  1. Login as SuperAdmin user
  2. Verify `/api/v1/me/permissions` response contains all 29 permission codes
  3. Navigate to main layout page
  4. Verify all navigation links (items, pool, weekly, gantt, report, teams, users, roles) are visible
  5. Verify no code references `isSuperAdmin` boolean — all visibility driven by `hasPermission()`
- **Expected**: All navigation and action elements visible for SuperAdmin, identical to pre-migration behavior. No isSuperAdmin references in runtime code path.
- **Priority**: P1

---

## API Test Cases

### TC-004: Custom role with sub_item:update edits non-assigned sub-item
- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/custom-role-edits-non-assigned-sub-item
- **Pre-conditions**: User with custom role (e.g., ext-member) granted `sub_item:update` permission code; sub-item exists assigned to a different user
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` — Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Create a user with custom role possessing `sub_item:update` (but not `sub_item:assign`)
  2. Create a sub-item assigned to a different user in the same team
  3. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with updated fields
- **Expected**: Returns 200 OK; sub-item updated successfully. No assignee ownership check blocks the request.
- **Priority**: P0

### TC-005: Custom role without sub_item:update gets 403 on edit
- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/custom-role-without-update-gets-403
- **Pre-conditions**: User with custom role having only `sub_item:view` (no `sub_item:update`)
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` — Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Create a user with custom role possessing only `sub_item:view`
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId`
- **Expected**: Returns 403 FORBIDDEN with permission denied message
- **Priority**: P0

### TC-006: Custom role with sub_item:change_status changes non-assigned sub-item status
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/custom-role-changes-non-assigned-sub-item-status
- **Pre-conditions**: User with custom role granted `sub_item:change_status`; sub-item assigned to different user
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` — Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Create a user with custom role possessing `sub_item:change_status` (but not `sub_item:assign`)
  2. Create a sub-item assigned to a different user in the same team
  3. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with new status
- **Expected**: Returns 200 OK; status changed successfully. No assignee ownership check.
- **Priority**: P0

### TC-007: Custom role without sub_item:change_status gets 403 on status change
- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/custom-role-without-change-status-gets-403
- **Pre-conditions**: User with custom role lacking `sub_item:change_status`
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` — Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Create a user with custom role without `sub_item:change_status`
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status`
- **Expected**: Returns 403 FORBIDDEN
- **Priority**: P0

### TC-008: SuperAdmin creates team (201)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-creates-team
- **Pre-conditions**: SuperAdmin user (`is_super_admin=true`); seed data loaded with all 29 permission codes
- **Route**: `POST /api/v1/teams` — Matched (backend/internal/handler/router.go:167)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams` with team name and code
- **Expected**: Returns 201 Created
- **Priority**: P0

### TC-009: SuperAdmin updates team (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-updates-team
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `PUT /api/v1/teams/:teamId` — Matched (backend/internal/handler/router.go:103)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId` with updated fields
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-010: SuperAdmin invites member (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members
- **Test ID**: api/teams-members/superadmin-invites-member
- **Pre-conditions**: SuperAdmin user; team exists; target user exists
- **Route**: `POST /api/v1/teams/:teamId/members` — Matched (backend/internal/handler/router.go:109)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/members` with user ID and role
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-011: SuperAdmin modifies member role (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members-role
- **Test ID**: api/teams-members-role/superadmin-modifies-member-role
- **Pre-conditions**: SuperAdmin user; team has existing members
- **Route**: `PUT /api/v1/teams/:teamId/members/:userId/role` — Matched (backend/internal/handler/router.go:111)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/members/:userId/role` with new role
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-012: SuperAdmin removes member (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members
- **Test ID**: api/teams-members/superadmin-removes-member
- **Pre-conditions**: SuperAdmin user; team has a removable member
- **Route**: `DELETE /api/v1/teams/:teamId/members/:userId` — Matched (backend/internal/handler/router.go:110)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `DELETE /api/v1/teams/:teamId/members/:userId`
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-013: SuperAdmin transfers PM (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-pm
- **Test ID**: api/teams-pm/superadmin-transfers-pm
- **Pre-conditions**: SuperAdmin user; team has at least 2 members
- **Route**: `PUT /api/v1/teams/:teamId/pm` — Matched (backend/internal/handler/router.go:112)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/pm` with new PM user ID
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-014: SuperAdmin disbands team (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-disbands-team
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `DELETE /api/v1/teams/:teamId` — Matched (backend/internal/handler/router.go:104)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `DELETE /api/v1/teams/:teamId`
- **Expected**: Returns 200 OK
- **Priority**: P0

### TC-015: SuperAdmin creates main item (201)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/superadmin-creates-main-item
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `POST /api/v1/teams/:teamId/main-items` — Matched (backend/internal/handler/router.go:115)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items` with title and required fields
- **Expected**: Returns 201 Created
- **Priority**: P0

### TC-016: SuperAdmin edits main item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/superadmin-edits-main-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `PUT /api/v1/teams/:teamId/main-items/:itemId` — Matched (backend/internal/handler/router.go:118)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/main-items/:itemId` with updated fields
- **Expected**: Returns 200 OK
- **Priority**: P1

### TC-017: SuperAdmin archives main item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items-archive
- **Test ID**: api/main-items-archive/superadmin-archives-main-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `POST /api/v1/teams/:teamId/main-items/:itemId/archive` — Matched (backend/internal/handler/router.go:121)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items/:itemId/archive`
- **Expected**: Returns 200 OK
- **Priority**: P1

### TC-018: SuperAdmin creates sub-item (201)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/superadmin-creates-sub-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` — Matched (backend/internal/handler/router.go:124)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` with title
- **Expected**: Returns 201 Created
- **Priority**: P1

### TC-019: SuperAdmin edits sub-item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/superadmin-edits-sub-item
- **Pre-conditions**: SuperAdmin user; sub-item exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` — Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with updated fields
- **Expected**: Returns 200 OK
- **Priority**: P1

### TC-020: SuperAdmin assigns sub-item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items-assignee
- **Test ID**: api/sub-items-assignee/superadmin-assigns-sub-item
- **Pre-conditions**: SuperAdmin user; sub-item exists; team member exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` — Matched (backend/internal/handler/router.go:132)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` with assignee user ID
- **Expected**: Returns 200 OK
- **Priority**: P1

### TC-021: SuperAdmin changes sub-item status (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/superadmin-changes-sub-item-status
- **Pre-conditions**: SuperAdmin user; sub-item exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` — Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with new status
- **Expected**: Returns 200 OK
- **Priority**: P1

### TC-022: SuperAdmin accesses non-member team resources (200)
- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/teams-resources
- **Test ID**: api/teams-resources/superadmin-accesses-non-member-team-resources
- **Pre-conditions**: SuperAdmin user (`is_super_admin=true`) NOT in `team_members` table for target team
- **Route**: `GET /api/v1/teams/:teamId/main-items` — Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as SuperAdmin who is not a member of the target team
  2. Send `GET /api/v1/teams/:teamId/main-items`
- **Expected**: Returns 200 OK. TeamScopeMiddleware injects all 29 permission codes and skips team membership check for SuperAdmin.
- **Priority**: P0

### TC-023: Non-member user without SuperAdmin gets 403 on cross-team access
- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/teams-resources
- **Test ID**: api/teams-resources/non-member-user-gets-403-on-cross-team-access
- **Pre-conditions**: Regular user (non-SuperAdmin) who is not a member of the target team
- **Route**: `GET /api/v1/teams/:teamId/main-items` — Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as regular user who is not a member of the target team
  2. Send `GET /api/v1/teams/:teamId/main-items`
- **Expected**: Returns 403 FORBIDDEN. TeamScopeMiddleware rejects non-member access.
- **Priority**: P0

### TC-024: PM team management operations succeed via permission codes
- **Source**: Story 6 / AC-1
- **Type**: API
- **Target**: api/teams-management
- **Test ID**: api/teams-management/pm-team-management-operations-succeed
- **Pre-conditions**: User with PM role (holding `team:invite`, `team:remove`, `team:transfer` permission codes); team with at least 2 members
- **Route**: `POST /api/v1/teams/:teamId/members` — Matched (backend/internal/handler/router.go:109)
- **Steps**:
  1. Login as PM user
  2. Send `POST /api/v1/teams/:teamId/members` (invite) — expect 200
  3. Send `DELETE /api/v1/teams/:teamId/members/:userId` (remove) — expect 200
  4. Send `PUT /api/v1/teams/:teamId/pm` (transfer PM) — expect 200
- **Expected**: All three operations return 200 OK. Service layer no longer performs PM identity checks; authorization handled by middleware permission code check only.
- **Priority**: P0

### TC-025: Custom role with progress:create adds progress to non-assigned sub-item
- **Source**: Story 7 / AC-1
- **Type**: API
- **Target**: api/progress
- **Test ID**: api/progress/custom-role-adds-progress-to-non-assigned-sub-item
- **Pre-conditions**: User with custom role granted `progress:create`; sub-item assigned to a different user
- **Route**: `POST /api/v1/teams/:teamId/sub-items/:subId/progress` — Matched (backend/internal/handler/router.go:135)
- **Steps**:
  1. Create a user with custom role possessing `progress:create`
  2. Create a sub-item assigned to a different user
  3. Send `POST /api/v1/teams/:teamId/sub-items/:subId/progress` with progress data
- **Expected**: Returns 201 Created (or 200 OK). Progress created successfully. No `isPMOrSuperAdmin()` check blocks the request.
- **Priority**: P0

---

## CLI Test Cases

_No CLI test cases — this project does not expose a CLI interface._

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 5 / AC-1 | UI | ui/build | P0 |
| TC-002 | Story 5 / AC-1 | UI | ui/tests | P0 |
| TC-003 | Story 5 / AC-2 | UI | ui/permissions | P1 |
| TC-004 | Story 1 / AC-1 | API | api/sub-items | P0 |
| TC-005 | Story 1 / AC-2 | API | api/sub-items | P0 |
| TC-006 | Story 2 / AC-1 | API | api/sub-items-status | P0 |
| TC-007 | Story 2 / AC-2 | API | api/sub-items-status | P0 |
| TC-008 | Story 3 / AC 3a | API | api/teams | P0 |
| TC-009 | Story 3 / AC 3a | API | api/teams | P0 |
| TC-010 | Story 3 / AC 3a | API | api/teams-members | P0 |
| TC-011 | Story 3 / AC 3a | API | api/teams-members-role | P0 |
| TC-012 | Story 3 / AC 3a | API | api/teams-members | P0 |
| TC-013 | Story 3 / AC 3a | API | api/teams-pm | P0 |
| TC-014 | Story 3 / AC 3a | API | api/teams | P0 |
| TC-015 | Story 3 / AC 3b | API | api/main-items | P0 |
| TC-016 | Story 3 / AC 3b | API | api/main-items | P1 |
| TC-017 | Story 3 / AC 3b | API | api/main-items-archive | P1 |
| TC-018 | Story 3 / AC 3b | API | api/sub-items | P1 |
| TC-019 | Story 3 / AC 3b | API | api/sub-items | P1 |
| TC-020 | Story 3 / AC 3b | API | api/sub-items-assignee | P1 |
| TC-021 | Story 3 / AC 3b | API | api/sub-items-status | P1 |
| TC-022 | Story 4 / AC-1 | API | api/teams-resources | P0 |
| TC-023 | Story 4 / AC-2 | API | api/teams-resources | P0 |
| TC-024 | Story 6 / AC-1 | API | api/teams-management | P0 |
| TC-025 | Story 7 / AC-1 | API | api/progress | P0 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| `PUT /api/v1/teams/:teamId/sub-items/:subId` | Matched | TC-004, TC-005, TC-019 | router.go:129 |
| `PUT /api/v1/teams/:teamId/sub-items/:subId/status` | Matched | TC-006, TC-007, TC-021 | router.go:130 |
| `POST /api/v1/teams` | Matched | TC-008 | router.go:167 |
| `PUT /api/v1/teams/:teamId` | Matched | TC-009 | router.go:103 |
| `POST /api/v1/teams/:teamId/members` | Matched | TC-010, TC-024 | router.go:109 |
| `PUT /api/v1/teams/:teamId/members/:userId/role` | Matched | TC-011 | router.go:111 |
| `DELETE /api/v1/teams/:teamId/members/:userId` | Matched | TC-012, TC-024 | router.go:110 |
| `PUT /api/v1/teams/:teamId/pm` | Matched | TC-013, TC-024 | router.go:112 |
| `DELETE /api/v1/teams/:teamId` | Matched | TC-014 | router.go:104 |
| `POST /api/v1/teams/:teamId/main-items` | Matched | TC-015 | router.go:115 |
| `PUT /api/v1/teams/:teamId/main-items/:itemId` | Matched | TC-016 | router.go:118 |
| `POST /api/v1/teams/:teamId/main-items/:itemId/archive` | Matched | TC-017 | router.go:121 |
| `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` | Matched | TC-018 | router.go:124 |
| `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` | Matched | TC-020 | router.go:132 |
| `GET /api/v1/teams/:teamId/main-items` | Matched | TC-022, TC-023 | router.go:116 |
| `POST /api/v1/teams/:teamId/sub-items/:subId/progress` | Matched | TC-025 | router.go:135 |
