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
| Integration | 2   |
| API  | 34  |
| CLI  | 0   |
| **Total** | **39** |

> **Note**: This feature is a backend permission refactoring with frontend isSuperAdmin removal. API tests cover the core behavioral changes; UI tests cover frontend cleanup verification. Integration tests validate the full middleware-to-handler chain.

---

## UI Test Cases

### TC-001: TypeScript compilation passes after isSuperAdmin removal
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/build
- **Test ID**: ui/build/typescript-compilation-passes-after-issuperadmin-removal
- **Pre-conditions**: All frontend isSuperAdmin references have been removed from types, store, components, and mocks
- **Steps**:
  1. Run `npx tsc --noEmit`
  2. Check the exit code is 0 and stderr contains no type errors referencing `isSuperAdmin`
- **Expected**: Exit code 0; stderr is empty (zero compilation errors)
- **Priority**: P0

### TC-002: Frontend tests pass after isSuperAdmin removal
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/tests
- **Test ID**: ui/tests/frontend-tests-pass-after-issuperadmin-removal
- **Pre-conditions**: All isSuperAdmin references removed; mock data updated
- **Steps**:
  1. Run `npx vitest run`
  2. Check the exit code is 0
- **Expected**: Exit code 0; vitest summary shows 0 failed tests
- **Priority**: P0

### TC-003: SuperAdmin sees all UI elements via hasPermission
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Target**: ui/permissions
- **Test ID**: ui/permissions/superadmin-sees-all-ui-elements-via-haspermission
- **Pre-conditions**: SuperAdmin user logged in; `/api/v1/me/permissions` returns all 29 permission codes
- **Route**: `/items`
- **Element**: `[data-testid="nav-items"]`, `[data-testid="nav-pool"]`, `[data-testid="nav-weekly"]`, `[data-testid="nav-gantt"]`, `[data-testid="nav-table"]`, `[data-testid="nav-report"]`, `[data-testid="nav-teams"]`, `[data-testid="nav-users"]`, `[data-testid="nav-roles"]`
- **Steps**:
  1. Login as SuperAdmin user
  2. Send `GET /api/v1/me/permissions` and assert response body contains an array with all 29 permission code strings
  3. Navigate to `/items` (main layout page)
  4. Assert each navigation element identified by its `data-testid` is visible (not hidden, not `display:none`)
  5. Search the JS bundle source for the string `isSuperAdmin` and assert the result count is 0
- **Expected**: `GET /api/v1/me/permissions` returns 200 with body `{ "codes": [...29 items...] }`. All 9 navigation elements with `data-testid` attributes are visible. `isSuperAdmin` string search returns 0 matches in the served JS.
- **Priority**: P1

---

## API Test Cases

### TC-004: Custom role with sub_item:update edits non-assigned sub-item
- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/custom-role-edits-non-assigned-sub-item
- **Pre-conditions**: User with custom role (e.g., ext-member) granted `sub_item:update` permission code; sub-item exists assigned to a different user
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Create a user with custom role possessing `sub_item:update` (but not `sub_item:assign`)
  2. Create a sub-item assigned to a different user in the same team
  3. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{ "title": "updated-title" }`
- **Expected**: Returns `200` with response body `{ "id": <subId>, "title": "updated-title", ... }`. Subsequent `GET /api/v1/teams/:teamId/sub-items/:subId` confirms the title field equals `"updated-title"`.
- **Priority**: P0

### TC-005: Custom role without sub_item:update gets 403 on edit
- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/custom-role-without-update-gets-403
- **Pre-conditions**: User with custom role having only `sub_item:view` (no `sub_item:update`)
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Create a user with custom role possessing only `sub_item:view`
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{ "title": "hack" }`
- **Expected**: Returns `403` with response body containing `{ "code": "FORBIDDEN", "message": "..." }`
- **Priority**: P0

### TC-006: Custom role with sub_item:change_status changes non-assigned sub-item status
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/custom-role-changes-non-assigned-sub-item-status
- **Pre-conditions**: User with custom role granted `sub_item:change_status`; sub-item assigned to different user
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` -- Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Create a user with custom role possessing `sub_item:change_status` (but not `sub_item:assign`)
  2. Create a sub-item assigned to a different user in the same team
  3. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with body `{ "status": "in_progress" }`
- **Expected**: Returns `200` with response body containing `{ "id": <subId>, "status": "in_progress" }`
- **Priority**: P0

### TC-007: Custom role without sub_item:change_status gets 403 on status change
- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/custom-role-without-change-status-gets-403
- **Pre-conditions**: User with custom role lacking `sub_item:change_status`
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` -- Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Create a user with custom role without `sub_item:change_status`
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with body `{ "status": "done" }`
- **Expected**: Returns `403` with response body containing `{ "code": "FORBIDDEN", "message": "..." }`
- **Priority**: P0

### TC-008: SuperAdmin creates team (201)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-creates-team
- **Pre-conditions**: SuperAdmin user (`is_super_admin=true`); seed data loaded with all 29 permission codes
- **Route**: `POST /api/v1/teams` -- Matched (backend/internal/handler/router.go:167)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams` with body `{ "name": "test-team", "code": "TT" }`
- **Expected**: Returns `201` with response body containing `{ "id": <number>, "name": "test-team", "code": "TT" }`
- **Priority**: P0

### TC-009: SuperAdmin updates team (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-updates-team
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `PUT /api/v1/teams/:teamId` -- Matched (backend/internal/handler/router.go:103)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId` with body `{ "name": "updated-team" }`
- **Expected**: Returns `200` with response body containing `{ "id": <teamId>, "name": "updated-team" }`
- **Priority**: P0

### TC-010: SuperAdmin invites member (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members
- **Test ID**: api/teams-members/superadmin-invites-member
- **Pre-conditions**: SuperAdmin user; team exists; target user exists
- **Route**: `POST /api/v1/teams/:teamId/members` -- Matched (backend/internal/handler/router.go:109)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/members` with body `{ "userId": <targetUserId>, "role": "member" }`
- **Expected**: Returns `200` with response body containing `{ "teamId": <teamId>, "userId": <targetUserId> }`
- **Priority**: P0

### TC-011: SuperAdmin modifies member role (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members-role
- **Test ID**: api/teams-members-role/superadmin-modifies-member-role
- **Pre-conditions**: SuperAdmin user; team has existing members
- **Route**: `PUT /api/v1/teams/:teamId/members/:userId/role` -- Matched (backend/internal/handler/router.go:111)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/members/:userId/role` with body `{ "role": "pm" }`
- **Expected**: Returns `200` with response body containing `{ "userId": <userId>, "role": "pm" }`
- **Priority**: P0

### TC-012: SuperAdmin removes member (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-members
- **Test ID**: api/teams-members/superadmin-removes-member
- **Pre-conditions**: SuperAdmin user; team has a removable member
- **Route**: `DELETE /api/v1/teams/:teamId/members/:userId` -- Matched (backend/internal/handler/router.go:110)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `DELETE /api/v1/teams/:teamId/members/:userId`
- **Expected**: Returns `200` with response body `{ "message": "ok" }`. Subsequent `GET /api/v1/teams/:teamId/members` returns a list that does not contain the removed `userId`.
- **Priority**: P0

### TC-013: SuperAdmin transfers PM (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams-pm
- **Test ID**: api/teams-pm/superadmin-transfers-pm
- **Pre-conditions**: SuperAdmin user; team has at least 2 members
- **Route**: `PUT /api/v1/teams/:teamId/pm` -- Matched (backend/internal/handler/router.go:112)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/pm` with body `{ "newPmUserId": <userId> }`
- **Expected**: Returns `200` with response body containing `{ "teamId": <teamId>, "pmUserId": <userId> }`
- **Priority**: P0

### TC-014: SuperAdmin disbands team (200)
- **Source**: Story 3 / AC 3a
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/superadmin-disbands-team
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `DELETE /api/v1/teams/:teamId` -- Matched (backend/internal/handler/router.go:104)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `DELETE /api/v1/teams/:teamId`
- **Expected**: Returns `200` with response body `{ "message": "ok" }`. Subsequent `GET /api/v1/teams/:teamId` returns `404` or response body no longer contains the disbanded teamId, confirming the team is deleted.
- **Priority**: P0

### TC-015: SuperAdmin creates main item (201)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/superadmin-creates-main-item
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `POST /api/v1/teams/:teamId/main-items` -- Matched (backend/internal/handler/router.go:115)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items` with body `{ "title": "test-item" }`
- **Expected**: Returns `201` with response body containing `{ "id": <number>, "title": "test-item" }`
- **Priority**: P0

### TC-016: SuperAdmin edits main item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/superadmin-edits-main-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `PUT /api/v1/teams/:teamId/main-items/:itemId` -- Matched (backend/internal/handler/router.go:118)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/main-items/:itemId` with body `{ "title": "edited-item" }`
- **Expected**: Returns `200` with response body containing `{ "id": <itemId>, "title": "edited-item" }`
- **Priority**: P1

### TC-017: SuperAdmin archives main item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/main-items-archive
- **Test ID**: api/main-items-archive/superadmin-archives-main-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `POST /api/v1/teams/:teamId/main-items/:itemId/archive` -- Matched (backend/internal/handler/router.go:121)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items/:itemId/archive` (empty body)
- **Expected**: Returns `200` with response body containing `{ "id": <itemId>, "status": "archived" }`
- **Priority**: P1

### TC-018: SuperAdmin creates sub-item (201)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/superadmin-creates-sub-item
- **Pre-conditions**: SuperAdmin user; main item exists
- **Route**: `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` -- Matched (backend/internal/handler/router.go:124)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` with body `{ "title": "sub-1" }`
- **Expected**: Returns `201` with response body containing `{ "id": <number>, "title": "sub-1" }`
- **Priority**: P1

### TC-019: SuperAdmin edits sub-item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/superadmin-edits-sub-item
- **Pre-conditions**: SuperAdmin user; sub-item exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{ "title": "edited-sub" }`
- **Expected**: Returns `200` with response body containing `{ "id": <subId>, "title": "edited-sub" }`
- **Priority**: P1

### TC-020: SuperAdmin assigns sub-item (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items-assignee
- **Test ID**: api/sub-items-assignee/superadmin-assigns-sub-item
- **Pre-conditions**: SuperAdmin user; sub-item exists; team member exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` -- Matched (backend/internal/handler/router.go:132)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` with body `{ "assigneeId": <userId> }`
- **Expected**: Returns `200` with response body containing `{ "id": <subId>, "assigneeId": <userId> }`
- **Priority**: P1

### TC-021: SuperAdmin changes sub-item status (200)
- **Source**: Story 3 / AC 3b
- **Type**: API
- **Target**: api/sub-items-status
- **Test ID**: api/sub-items-status/superadmin-changes-sub-item-status
- **Pre-conditions**: SuperAdmin user; sub-item exists
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId/status` -- Matched (backend/internal/handler/router.go:130)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with body `{ "status": "in_progress" }`
- **Expected**: Returns `200` with response body containing `{ "id": <subId>, "status": "in_progress" }`
- **Priority**: P1

### TC-022: SuperAdmin accesses non-member team resources (200)
- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/teams-resources
- **Test ID**: api/teams-resources/superadmin-accesses-non-member-team-resources
- **Pre-conditions**: SuperAdmin user (`is_super_admin=true`) NOT in `team_members` table for target team
- **Route**: `GET /api/v1/teams/:teamId/main-items` -- Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as SuperAdmin who is not a member of the target team
  2. Send `GET /api/v1/teams/:teamId/main-items`
- **Expected**: Returns `200` with response body `{ "items": [...], "total": <number> }` (a non-empty or empty array; not a 403 error)
- **Priority**: P0

### TC-023: Non-member user without SuperAdmin gets 403 on cross-team access
- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/teams-resources
- **Test ID**: api/teams-resources/non-member-user-gets-403-on-cross-team-access
- **Pre-conditions**: Regular user (non-SuperAdmin) who is not a member of the target team
- **Route**: `GET /api/v1/teams/:teamId/main-items` -- Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as regular user who is not a member of the target team
  2. Send `GET /api/v1/teams/:teamId/main-items`
- **Expected**: Returns `403` with response body containing `{ "code": "FORBIDDEN", "message": "..." }`
- **Priority**: P0

### TC-024: PM team management operations succeed via permission codes
- **Source**: Story 6 / AC-1
- **Type**: API
- **Target**: api/teams-management
- **Test ID**: api/teams-management/pm-team-management-operations-succeed
- **Pre-conditions**: User with PM role (holding `team:invite`, `team:remove`, `team:transfer` permission codes); team with at least 2 members
- **Route**: `POST /api/v1/teams/:teamId/members` -- Matched (backend/internal/handler/router.go:109)
- **Steps**:
  1. Login as PM user
  2. Send `POST /api/v1/teams/:teamId/members` with body `{ "userId": <newUserId>, "role": "member" }` -- assert status 200
  3. Send `DELETE /api/v1/teams/:teamId/members/:userId` -- assert status 200
  4. Send `PUT /api/v1/teams/:teamId/pm` with body `{ "newPmUserId": <userId> }` -- assert status 200
- **Expected**: Step 2 returns `200` with response body `{ "teamId": <teamId>, "userId": <newUserId>, "role": "member" }`. Step 3 returns `200` with response body `{ "message": "ok" }`. Step 4 returns `200` with response body `{ "teamId": <teamId>, "pmUserId": <userId> }`.
- **Priority**: P0

### TC-025: Custom role with progress:create adds progress to non-assigned sub-item
- **Source**: Story 7 / AC-1
- **Type**: API
- **Target**: api/progress
- **Test ID**: api/progress/custom-role-adds-progress-to-non-assigned-sub-item
- **Pre-conditions**: User with custom role granted `progress:create`; sub-item exists in the same team, assigned to a different user
- **Route**: `POST /api/v1/teams/:teamId/sub-items/:subId/progress` -- Matched (backend/internal/handler/router.go:135)
- **Steps**:
  1. Create a user with custom role possessing `progress:create` in the same team as the sub-item
  2. Create a sub-item assigned to a different user in that team
  3. Send `POST /api/v1/teams/:teamId/sub-items/:subId/progress` with body `{ "content": "test progress", "completion": 50 }`
- **Expected**: Returns `200` with response body containing `{ "id": <recordId>, "content": "test progress", "completion": 50 }`
- **Priority**: P0

### TC-026: SuperAdmin submits item pool entry (201)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/item-pool
- **Test ID**: api/item-pool/superadmin-submits-item-pool-entry
- **Pre-conditions**: SuperAdmin user; team exists
- **Route**: `POST /api/v1/teams/:teamId/item-pool` -- Matched (backend/internal/handler/router.go:146)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/item-pool` with body `{ "title": "todo-entry", "description": "test" }`
- **Expected**: Returns `201` with response body containing `{ "id": <number>, "title": "todo-entry" }`
- **Priority**: P0

### TC-027: SuperAdmin reviews (rejects) item pool entry (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/item-pool-review
- **Test ID**: api/item-pool-review/superadmin-rejects-item-pool-entry
- **Pre-conditions**: SuperAdmin user; item pool entry exists in pending status
- **Route**: `POST /api/v1/teams/:teamId/item-pool/:poolId/reject` -- Matched (backend/internal/handler/router.go:152)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `POST /api/v1/teams/:teamId/item-pool/:poolId/reject` with body `{ "reason": "not needed" }`
- **Expected**: Returns `200` with response body containing `{ "id": <poolId>, "status": "rejected" }`
- **Priority**: P0

### TC-028: SuperAdmin views weekly report (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/views-weekly
- **Test ID**: api/views-weekly/superadmin-views-weekly
- **Pre-conditions**: SuperAdmin user; team has items with sub-items in the current week
- **Route**: `GET /api/v1/teams/:teamId/views/weekly` -- Matched (backend/internal/handler/router.go:155)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/teams/:teamId/views/weekly`
- **Expected**: Returns `200` with response body containing a JSON array or object (non-null weekly view data structure)
- **Priority**: P0

### TC-029: SuperAdmin views gantt chart (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/views-gantt
- **Test ID**: api/views-gantt/superadmin-views-gantt
- **Pre-conditions**: SuperAdmin user; team has items with date ranges
- **Route**: `GET /api/v1/teams/:teamId/views/gantt` -- Matched (backend/internal/handler/router.go:156)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/teams/:teamId/views/gantt`
- **Expected**: Returns `200` with response body containing a JSON array or object (non-null gantt view data)
- **Priority**: P0

### TC-030: SuperAdmin views table view (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/views-table
- **Test ID**: api/views-table/superadmin-views-table
- **Pre-conditions**: SuperAdmin user; team has items
- **Route**: `GET /api/v1/teams/:teamId/views/table` -- Matched (backend/internal/handler/router.go:157)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/teams/:teamId/views/table`
- **Expected**: Returns `200` with response body containing a JSON array or object (non-null table view data)
- **Priority**: P0

### TC-031: SuperAdmin exports table report (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/views-export
- **Test ID**: api/views-export/superadmin-exports-table-report
- **Pre-conditions**: SuperAdmin user; team has items with sub-items
- **Route**: `GET /api/v1/teams/:teamId/views/table/export` -- Matched (backend/internal/handler/router.go:158)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/teams/:teamId/views/table/export`
- **Expected**: Returns `200` with `Content-Type: application/octet-stream` (or `text/csv`) and a non-empty response body containing the exported file data
- **Priority**: P0

### TC-032: SuperAdmin views user list (200)
- **Source**: Story 3 / AC 3c
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/superadmin-views-user-list
- **Pre-conditions**: SuperAdmin user; at least one user exists in the system
- **Route**: `GET /api/v1/admin/users` -- Matched (backend/internal/handler/router.go:174)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/admin/users`
- **Expected**: Returns `200` with response body containing `{ "users": [...], "total": <number> }` where the array length is >= 1
- **Priority**: P0

### TC-033: Custom role without progress:create gets 403 on progress add
- **Source**: Story 7 / AC-2
- **Type**: API
- **Target**: api/progress
- **Test ID**: api/progress/custom-role-without-progress-create-gets-403
- **Pre-conditions**: User with custom role lacking `progress:create` (e.g., has only `sub_item:view`); sub-item exists in the same team assigned to a different user
- **Route**: `POST /api/v1/teams/:teamId/sub-items/:subId/progress` -- Matched (backend/internal/handler/router.go:135)
- **Steps**:
  1. Create a user with custom role possessing only `sub_item:view` (no `progress:create`)
  2. Send `POST /api/v1/teams/:teamId/sub-items/:subId/progress` with body `{ "content": "unauthorized", "completion": 10 }`
- **Expected**: Returns `403` with response body containing `{ "code": "FORBIDDEN", "message": "..." }`
- **Priority**: P0

### TC-034: SuperAdmin with no seed data gets 403 on team-scoped endpoint
- **Source**: Story 3 / (edge case)
- **Type**: API
- **Target**: api/seed-edge
- **Test ID**: api/seed-edge/superadmin-no-seed-data-gets-403
- **Pre-conditions**: SuperAdmin user exists but all entries in `role_permissions` for that user's role have been deleted (manually cleared before test); team exists
- **Route**: `POST /api/v1/teams/:teamId/main-items` -- Matched (backend/internal/handler/router.go:115)
- **Steps**:
  1. Login as the SuperAdmin user whose role has no permission codes
  2. Send `POST /api/v1/teams/:teamId/main-items` with body `{ "title": "edge-case" }`
- **Expected**: Returns `403` with response body containing `{ "code": "FORBIDDEN", "message": "..." }`. Permission middleware blocks access when permCodes list is empty, even for SuperAdmin identity.
- **Priority**: P1

### TC-035: SuperAdmin requests non-existent team resource returns 404 (not 403)
- **Source**: Story 3 / (edge case)
- **Type**: API
- **Target**: api/teams-edge
- **Test ID**: api/teams-edge/superadmin-nonexistent-team-returns-404
- **Pre-conditions**: SuperAdmin user; teamId `999999` does not exist in the database
- **Route**: `GET /api/v1/teams/999999/main-items` -- Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as SuperAdmin
  2. Send `GET /api/v1/teams/999999/main-items`
- **Expected**: Returns `404` with response body containing `{ "code": "NOT_FOUND", "message": "..." }` (team not found takes precedence over permission check; distinguishes "no access" from "resource does not exist")
- **Priority**: P1

### TC-038: PUT sub-item with empty body returns 400
- **Source**: Story 1 / (boundary case)
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/put-empty-body-returns-400
- **Pre-conditions**: User with custom role possessing `sub_item:update`; sub-item exists in the same team
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Login as the custom role user
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{}`
- **Expected**: Returns `400` with response body containing `{ "code": "BAD_REQUEST", "message": "..." }`. The permission check passes (user has `sub_item:update`) but the handler rejects the empty body due to missing required fields.
- **Priority**: P1

### TC-039: PUT sub-item with leftover isSuperAdmin field in body is ignored
- **Source**: Story 5 / (boundary case -- prd-spec.md Failure Scenarios table)
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/put-issuperadmin-field-ignored
- **Pre-conditions**: User with custom role possessing `sub_item:update`; sub-item exists in the same team assigned to a different user
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Login as the custom role user
  2. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{ "title": "updated-title", "isSuperAdmin": true }`
- **Expected**: Returns `200` with response body containing `{ "id": <subId>, "title": "updated-title" }`. The backend ignores the unknown `isSuperAdmin` JSON field; the update succeeds normally based on permission codes alone.
- **Priority**: P1

---

## Integration Test Cases

### TC-036: Full chain -- SuperAdmin cross-team access flows through TeamScopeMiddleware to handler
- **Source**: Story 4 / AC-1 + Story 3
- **Type**: Integration
- **Target**: integration/middleware-handler-chain
- **Test ID**: integration/middleware-handler-chain/superadmin-cross-team-full-chain
- **Pre-conditions**: SuperAdmin user NOT in `team_members` for target team; target team has 3 main items; SuperAdmin has all 29 permCodes loaded
- **Route**: `GET /api/v1/teams/:teamId/main-items` -- Matched (backend/internal/handler/router.go:116)
- **Steps**:
  1. Login as SuperAdmin who is not a member of the target team
  2. Send `GET /api/v1/teams/:teamId/main-items`
  3. Assert response status is 200 and body contains 3 items
  4. Send `PUT /api/v1/teams/:teamId/main-items/:itemId` with body `{ "title": "cross-team-edit" }`
  5. Assert response status is 200 and body contains `{ "title": "cross-team-edit" }`
  6. Send `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` with body `{ "title": "cross-team-sub" }`
  7. Assert response status is 201
- **Expected**: All three requests succeed (200, 200, 201). The cross-team read-write flow passes through AuthMiddleware, TeamScopeMiddleware (team membership bypass + permCodes injection), RequirePermission (code validation), and the handler without any layer rejecting the request. Response bodies contain the expected data structures.
- **Priority**: P0

### TC-037: Full chain -- Custom role permission denial propagates through middleware to 403 response
- **Source**: Story 1 / AC-2 + Story 7 / AC-2
- **Type**: Integration
- **Target**: integration/permission-denial-chain
- **Test ID**: integration/permission-denial-chain/custom-role-denial-propagates
- **Pre-conditions**: User with custom role possessing only `sub_item:view` and `sub_item:read` (no `sub_item:update`, no `progress:create`); sub-item exists in the same team
- **Route**: `PUT /api/v1/teams/:teamId/sub-items/:subId` -- Matched (backend/internal/handler/router.go:129)
- **Steps**:
  1. Login as the custom role user
  2. Send `GET /api/v1/teams/:teamId/main-items` -- assert 200 (has read access)
  3. Send `PUT /api/v1/teams/:teamId/sub-items/:subId` with body `{ "title": "blocked" }` -- assert 403
  4. Send `POST /api/v1/teams/:teamId/sub-items/:subId/progress` with body `{ "content": "blocked" }` -- assert 403
  5. Verify both 403 response bodies contain `{ "code": "FORBIDDEN" }`
- **Expected**: Step 2 returns `200` (read succeeds). Steps 3 and 4 each return `403` with `{ "code": "FORBIDDEN" }`. This confirms the RequirePermission middleware denies operations the role lacks codes for, while allowing operations the role has codes for, in a single session with consistent middleware state.
- **Priority**: P0

---

## CLI Test Cases

_No CLI test cases -- this project does not expose a CLI interface._

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
| TC-026 | Story 3 / AC 3c | API | api/item-pool | P0 |
| TC-027 | Story 3 / AC 3c | API | api/item-pool-review | P0 |
| TC-028 | Story 3 / AC 3c | API | api/views-weekly | P0 |
| TC-029 | Story 3 / AC 3c | API | api/views-gantt | P0 |
| TC-030 | Story 3 / AC 3c | API | api/views-table | P0 |
| TC-031 | Story 3 / AC 3c | API | api/views-export | P0 |
| TC-032 | Story 3 / AC 3c | API | api/admin-users | P0 |
| TC-033 | Story 7 / AC-2 | API | api/progress | P0 |
| TC-034 | Story 3 / edge | API | api/seed-edge | P1 |
| TC-035 | Story 3 / edge | API | api/teams-edge | P1 |
| TC-038 | Story 1 / boundary | API | api/sub-items | P1 |
| TC-039 | Story 5 / boundary | API | api/sub-items | P1 |
| TC-036 | Story 4 / AC-1 + Story 3 | Integration | integration/middleware-handler-chain | P0 |
| TC-037 | Story 1 / AC-2 + Story 7 / AC-2 | Integration | integration/permission-denial-chain | P0 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| `PUT /api/v1/teams/:teamId/sub-items/:subId` | Matched | TC-004, TC-005, TC-019, TC-038, TC-039 | router.go:129 |
| `PUT /api/v1/teams/:teamId/sub-items/:subId/status` | Matched | TC-006, TC-007, TC-021 | router.go:130 |
| `POST /api/v1/teams` | Matched | TC-008 | router.go:167 |
| `PUT /api/v1/teams/:teamId` | Matched | TC-009 | router.go:103 |
| `POST /api/v1/teams/:teamId/members` | Matched | TC-010, TC-024 | router.go:109 |
| `PUT /api/v1/teams/:teamId/members/:userId/role` | Matched | TC-011 | router.go:111 |
| `DELETE /api/v1/teams/:teamId/members/:userId` | Matched | TC-012, TC-024 | router.go:110 |
| `PUT /api/v1/teams/:teamId/pm` | Matched | TC-013, TC-024 | router.go:112 |
| `DELETE /api/v1/teams/:teamId` | Matched | TC-014 | router.go:104 |
| `POST /api/v1/teams/:teamId/main-items` | Matched | TC-015, TC-034 | router.go:115 |
| `PUT /api/v1/teams/:teamId/main-items/:itemId` | Matched | TC-016, TC-036 | router.go:118 |
| `POST /api/v1/teams/:teamId/main-items/:itemId/archive` | Matched | TC-017 | router.go:121 |
| `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` | Matched | TC-018, TC-036 | router.go:124 |
| `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` | Matched | TC-020 | router.go:132 |
| `GET /api/v1/teams/:teamId/main-items` | Matched | TC-022, TC-023, TC-035, TC-036 | router.go:116 |
| `POST /api/v1/teams/:teamId/sub-items/:subId/progress` | Matched | TC-025, TC-033, TC-037 | router.go:135 |
| `POST /api/v1/teams/:teamId/item-pool` | Matched | TC-026 | router.go:146 |
| `POST /api/v1/teams/:teamId/item-pool/:poolId/reject` | Matched | TC-027 | router.go:152 |
| `GET /api/v1/teams/:teamId/views/weekly` | Matched | TC-028 | router.go:155 |
| `GET /api/v1/teams/:teamId/views/gantt` | Matched | TC-029 | router.go:156 |
| `GET /api/v1/teams/:teamId/views/table` | Matched | TC-030 | router.go:157 |
| `GET /api/v1/teams/:teamId/views/table/export` | Matched | TC-031 | router.go:158 |
| `GET /api/v1/admin/users` | Matched | TC-032 | router.go:174 |
