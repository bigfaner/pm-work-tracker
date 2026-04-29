---
feature: "permission-granularity"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
generated: "2026-04-29"
---

# Test Cases: permission-granularity

## Summary

| Type | Count |
|------|-------|
| UI   | 14   |
| API  | 17   |
| CLI  | 1   |
| **Total** | **32** |

---

## UI Test Cases

## TC-001: Role dropdown loads with role:read permission
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-loads-with-role-read-permission
- **Pre-conditions**: User has `role:read` permission but not `role:create/update/delete`
- **Route**: /teams/:teamId
- **Steps**:
  1. Log in as a user with `role:read` permission only
  2. Navigate to team management page
  3. Click "Add Member" button to open the dialog
  4. Observe the role dropdown selector
- **Expected**: Role dropdown loads normally and displays all available roles
- **Priority**: P0

## TC-002: Role dropdown disabled without role:read permission
- **Source**: Story 1 / AC-3; UI Function 1 / States
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-disabled-without-role-read-permission
- **Pre-conditions**: User does not have `role:read` permission
- **Route**: /teams/:teamId
- **Steps**:
  1. Log in as a user without `role:read` permission
  2. Navigate to team management page
  3. Click "Add Member" button to open the dialog
  4. Observe the role dropdown selector
- **Expected**: Role dropdown is disabled and shows "no permission to view role list" message
- **Priority**: P0

## TC-003: Role dropdown shows loading state
- **Source**: UI Function 1 / States
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-shows-loading-state
- **Pre-conditions**: User has `role:read` permission; network is slow or intercepted
- **Route**: /teams/:teamId
- **Steps**:
  1. Log in as a user with `role:read` permission
  2. Throttle network to slow down API responses
  3. Navigate to team management page and open "Add Member" dialog
  4. Observe the role dropdown while the API request is in flight
- **Expected**: Dropdown shows loading indicator during API request
- **Priority**: P2

## TC-004: Role dropdown shows error state on API failure
- **Source**: UI Function 1 / States
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-shows-error-state-on-api-failure
- **Pre-conditions**: User has `role:read` permission; GET /admin/roles returns 5xx
- **Route**: /teams/:teamId
- **Steps**:
  1. Log in as a user with `role:read` permission
  2. Mock GET /admin/roles to return a server error
  3. Open "Add Member" dialog
  4. Observe the role dropdown
- **Expected**: Dropdown shows error state with retry option
- **Priority**: P2

## TC-005: Role dropdown shows empty message when no roles exist
- **Source**: UI Function 1 / Validation Rules
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-shows-empty-message-when-no-roles
- **Pre-conditions**: User has `role:read` permission; GET /admin/roles returns empty array
- **Route**: /teams/:teamId
- **Steps**:
  1. Log in as a user with `role:read` permission
  2. Mock GET /admin/roles to return empty array
  3. Open "Add Member" dialog
  4. Observe the role dropdown and submit button
- **Expected**: Dropdown displays "no available roles" message; submit button is disabled
- **Priority**: P2

## TC-006: User management nav entry hidden without user:list
- **Source**: Story 6 / AC-2; UI Function 2
- **Type**: UI
- **Target**: ui/sidebar
- **Test ID**: ui/sidebar/user-management-nav-hidden-without-user-list
- **Pre-conditions**: User does not have `user:list` permission
- **Route**: / (any page with sidebar)
- **Steps**:
  1. Log in as a user without `user:list` permission
  2. Observe the sidebar navigation menu
- **Expected**: "User Management" menu entry is not displayed
- **Priority**: P0

## TC-007: User management nav entry hidden during permission loading
- **Source**: UI Function 2 / Validation Rules
- **Type**: UI
- **Target**: ui/sidebar
- **Test ID**: ui/sidebar/user-management-nav-hidden-during-permission-loading
- **Pre-conditions**: GET /me/permissions request is pending (not yet returned)
- **Route**: / (any page with sidebar)
- **Steps**:
  1. Log in as a user
  2. Intercept and delay the GET /me/permissions response
  3. Observe the sidebar while permissions are still loading
- **Expected**: "User Management" menu entry is hidden by default during loading; no flash when permissions load
- **Priority**: P1

## TC-008: User management nav entry hidden on permission API failure
- **Source**: UI Function 2 / Validation Rules
- **Type**: UI
- **Target**: ui/sidebar
- **Test ID**: ui/sidebar/user-management-nav-hidden-on-permission-api-failure
- **Pre-conditions**: GET /me/permissions returns an error
- **Route**: / (any page with sidebar)
- **Steps**:
  1. Log in as a user
  2. Mock GET /me/permissions to return an error
  3. Observe the sidebar navigation
- **Expected**: "User Management" menu entry remains hidden; not exposed due to API failure
- **Priority**: P1

## TC-009: Create role button hidden without role:create
- **Source**: Story 6 / AC-1; UI Function 3
- **Type**: UI
- **Target**: ui/role-management
- **Test ID**: ui/role-management/create-role-button-hidden-without-role-create
- **Pre-conditions**: User does not have `role:create` permission
- **Route**: /admin/roles (role management page)
- **Steps**:
  1. Log in as a user without `role:create` permission
  2. Navigate to role management page
  3. Observe the page header area
- **Expected**: "Create Role" button is not displayed
- **Priority**: P0

## TC-010: Edit button hidden without role:update
- **Source**: UI Function 3 / States
- **Type**: UI
- **Target**: ui/role-management
- **Test ID**: ui/role-management/edit-button-hidden-without-role-update
- **Pre-conditions**: User does not have `role:update` permission
- **Route**: /admin/roles
- **Steps**:
  1. Log in as a user without `role:update` permission
  2. Navigate to role management page
  3. Observe each role row in the table
- **Expected**: "Edit" button is not displayed on any role row
- **Priority**: P0

## TC-011: Delete button hidden without role:delete
- **Source**: UI Function 3 / States
- **Type**: UI
- **Target**: ui/role-management
- **Test ID**: ui/role-management/delete-button-hidden-without-role-delete
- **Pre-conditions**: User does not have `role:delete` permission
- **Route**: /admin/roles
- **Steps**:
  1. Log in as a user without `role:delete` permission
  2. Navigate to role management page
  3. Observe each role row in the table
- **Expected**: "Delete" button is not displayed on any role row
- **Priority**: P0

## TC-012: Preset role delete button always disabled
- **Source**: UI Function 3 / Validation Rules
- **Type**: UI
- **Target**: ui/role-management
- **Test ID**: ui/role-management/preset-role-delete-button-always-disabled
- **Pre-conditions**: User has `role:delete` permission; preset roles (superadmin/pm/member) exist
- **Route**: /admin/roles
- **Steps**:
  1. Log in as a user with `role:delete` permission
  2. Navigate to role management page
  3. Observe the "Delete" button for preset roles (superadmin, pm, member)
- **Expected**: Delete buttons for preset roles are always disabled regardless of `role:delete` permission
- **Priority**: P1

## TC-013: Delete button disabled for role with bound members
- **Source**: UI Function 3 / Validation Rules
- **Type**: UI
- **Target**: ui/role-management
- **Test ID**: ui/role-management/delete-button-disabled-for-role-with-bound-members
- **Pre-conditions**: User has `role:delete` permission; a custom role has members bound to it
- **Route**: /admin/roles
- **Steps**:
  1. Log in as a user with `role:delete` permission
  2. Navigate to role management page
  3. Observe the "Delete" button for a custom role that has members bound to it
  4. Hover over the disabled delete button
- **Expected**: Delete button is disabled; tooltip shows "this role has members and cannot be deleted"
- **Priority**: P1

## TC-014: Role dropdown visible but user management nav hidden
- **Source**: Story 6 / AC-3
- **Type**: UI
- **Target**: ui/team-management
- **Test ID**: ui/team-management/role-dropdown-visible-but-user-mgmt-nav-hidden
- **Pre-conditions**: User has `role:read` but not `user:list`
- **Route**: /teams/:teamId; / (sidebar)
- **Steps**:
  1. Log in as a user with `role:read` but not `user:list`
  2. Verify sidebar does not show "User Management" entry
  3. Navigate to team management page and open "Add Member" dialog
  4. Observe the role dropdown
- **Expected**: Role dropdown loads normally and displays roles; user management page entry is not visible in navigation
- **Priority**: P0

---

## API Test Cases

## TC-015: GET /admin/roles returns 200 with role:read
- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/get-roles-returns-200-with-role-read
- **Pre-conditions**: User has `role:read` permission
- **Steps**:
  1. Authenticate as a user with `role:read` permission
  2. Send `GET /admin/roles`
- **Expected**: Returns HTTP 200 with role list
- **Priority**: P0

## TC-016: GET /admin/roles returns 403 without role:read
- **Source**: Story 1 / AC-3
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/get-roles-returns-403-without-role-read
- **Pre-conditions**: User does not have `role:read` permission
- **Steps**:
  1. Authenticate as a user without `role:read` permission
  2. Send `GET /admin/roles`
- **Expected**: Returns HTTP 403
- **Priority**: P0

## TC-017: GET /admin/users returns 200 with user:list (without sensitive fields)
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/get-users-returns-200-with-user-list
- **Pre-conditions**: User has `user:list` permission but not `user:read`
- **Steps**:
  1. Authenticate as a user with `user:list` but not `user:read`
  2. Send `GET /admin/users`
- **Expected**: Returns HTTP 200 with user list; response does not include sensitive fields (email, phone)
- **Priority**: P0

## TC-018: GET /admin/users/:userId returns 403 with user:list but not user:read
- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/get-user-detail-returns-403-without-user-read
- **Pre-conditions**: User has `user:list` but not `user:read`
- **Steps**:
  1. Authenticate as a user with `user:list` but not `user:read`
  2. Send `GET /admin/users/:userId`
- **Expected**: Returns HTTP 403; cannot access user detail
- **Priority**: P0

## TC-019: GET /admin/users/:userId returns 200 with user:list and user:read
- **Source**: Story 2 / AC-3
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/get-user-detail-returns-200-with-user-read
- **Pre-conditions**: User has both `user:list` and `user:read` permissions
- **Steps**:
  1. Authenticate as a user with both `user:list` and `user:read`
  2. Send `GET /admin/users/:userId`
- **Expected**: Returns HTTP 200 with full user detail including sensitive fields (email, phone)
- **Priority**: P0

## TC-020: POST /admin/roles returns 201 with role:create
- **Source**: Story 3 / AC-1
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/create-role-returns-201-with-role-create
- **Pre-conditions**: User has `role:create` permission; role name is valid and permission codes are valid
- **Steps**:
  1. Authenticate as a user with `role:create` permission
  2. Send `POST /admin/roles` with valid role name and permission codes
- **Expected**: Returns HTTP 201; new role is created successfully
- **Priority**: P0

## TC-021: PUT /admin/roles/:id returns 200 with role:update
- **Source**: Story 3 / AC-2
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/update-role-returns-200-with-role-update
- **Pre-conditions**: User has `role:update` permission
- **Steps**:
  1. Authenticate as a user with `role:update` permission
  2. Send `PUT /admin/roles/:id` with updated role data
- **Expected**: Returns HTTP 200; role is updated successfully
- **Priority**: P0

## TC-022: DELETE /admin/roles/:id returns 200 with role:delete and no bound members
- **Source**: Story 3 / AC-3
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/delete-role-returns-200-with-role-delete
- **Pre-conditions**: User has `role:delete` permission; target role has no bound members
- **Steps**:
  1. Authenticate as a user with `role:delete` permission
  2. Send `DELETE /admin/roles/:id` for a custom role with no members
- **Expected**: Returns HTTP 200; role is deleted successfully
- **Priority**: P0

## TC-023: POST /admin/roles returns 403 with role:read only
- **Source**: Story 3 / AC-4
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/create-role-returns-403-with-role-read-only
- **Pre-conditions**: User has `role:read` permission only (not `role:create`)
- **Steps**:
  1. Authenticate as a user with only `role:read` permission
  2. Send `POST /admin/roles` with valid data
- **Expected**: Returns HTTP 403
- **Priority**: P0

## TC-024: GET /admin/users returns 200 with user:list
- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/get-users-returns-200-with-user-list-story4
- **Pre-conditions**: User has `user:list` permission
- **Steps**:
  1. Authenticate as a user with `user:list` permission
  2. Send `GET /admin/users`
- **Expected**: Returns HTTP 200 with user list
- **Priority**: P1

## TC-025: GET /admin/users/:userId returns 200 with user:read
- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/get-user-detail-returns-200-with-user-read-story4
- **Pre-conditions**: User has `user:read` permission
- **Steps**:
  1. Authenticate as a user with `user:read` permission
  2. Send `GET /admin/users/:userId`
- **Expected**: Returns HTTP 200 with complete user detail
- **Priority**: P1

## TC-026: POST /admin/users returns 200 with user:assign_role
- **Source**: Story 4 / AC-3
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/assign-role-returns-200-with-user-assign-role
- **Pre-conditions**: User has `user:assign_role` permission
- **Steps**:
  1. Authenticate as a user with `user:assign_role` permission
  2. Send `POST /admin/users` with user ID and role assignment data
- **Expected**: Returns HTTP 200; role assignment is successful
- **Priority**: P0

## TC-027: Backend returns 403 when frontend-bypassed API is called without permission
- **Source**: Story 6 / AC-4
- **Type**: API
- **Target**: api/permissions
- **Test ID**: api/permissions/backend-returns-403-on-frontend-bypass
- **Pre-conditions**: User does not have the required permission for a specific API endpoint
- **Steps**:
  1. Authenticate as a user lacking a specific permission (e.g., no `role:create`)
  2. Directly call the protected API endpoint (e.g., `POST /admin/roles`) bypassing frontend UI
- **Expected**: Backend middleware returns HTTP 403; backend enforces permission regardless of frontend visibility
- **Priority**: P0

## TC-028: Migration replaces user:manage_role with role:create+update+delete
- **Source**: Story 5 / AC-1
- **Type**: API
- **Target**: api/migration
- **Test ID**: api/migration/replace-manage-role-with-role-crud
- **Pre-conditions**: Database has a custom role holding `user:manage_role` permission
- **Steps**:
  1. Run step-one migration script
  2. Query the role's permission codes from `pmw_role_permissions`
- **Expected**: `user:manage_role` is replaced by `role:create` + `role:update` + `role:delete`; original code no longer exists
- **Priority**: P0

## TC-029: Migration replaces old user:read with user:list
- **Source**: Story 5 / AC-2
- **Type**: API
- **Target**: api/migration
- **Test ID**: api/migration/replace-old-user-read-with-user-list
- **Pre-conditions**: Database has a role holding old semantic `user:read` permission
- **Steps**:
  1. Run step-one migration script
  2. Query the role's permission codes from `pmw_role_permissions`
- **Expected**: Old `user:read` is replaced by `user:list`; new semantic `user:read` is atomically written
- **Priority**: P0

## TC-030: Migration rollback on error
- **Source**: Story 5 / AC-3
- **Type**: API
- **Target**: api/migration
- **Test ID**: api/migration/rollback-on-error
- **Pre-conditions**: Database is in pre-migration state; migration script encounters an error during execution
- **Steps**:
  1. Force a failure during migration execution
  2. Verify transaction rollback occurs
  3. Query the database state
- **Expected**: Database is restored to pre-migration state; no partial migration residue exists
- **Priority**: P0

## TC-031: GET /admin/permissions returns 200 with role:read
- **Source**: Spec Section 5.2 (route binding table)
- **Type**: API
- **Target**: api/permissions
- **Test ID**: api/permissions/get-permissions-returns-200-with-role-read
- **Pre-conditions**: User has `role:read` permission
- **Steps**:
  1. Authenticate as a user with `role:read` permission
  2. Send `GET /admin/permissions`
- **Expected**: Returns HTTP 200 with permission code list
- **Priority**: P1

---

## CLI Test Cases

## TC-032: CI grep assertion finds zero old permission code references
- **Source**: Story 5 / AC-4
- **Type**: CLI
- **Target**: cli/ci-grep
- **Test ID**: cli/ci-grep/assert-zero-old-permission-code-references
- **Pre-conditions**: Step-one migration is complete
- **Steps**:
  1. Run CI grep assertion step in the pipeline
  2. Search codebase for `user:manage_role` and old semantic `user:read` references
- **Expected**: Reference count for `user:manage_role` and old `user:read` is zero; assertion passes
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | UI | ui/team-management | P0 |
| TC-002 | Story 1 / AC-3; UI Function 1 | UI | ui/team-management | P0 |
| TC-003 | UI Function 1 / States | UI | ui/team-management | P2 |
| TC-004 | UI Function 1 / States | UI | ui/team-management | P2 |
| TC-005 | UI Function 1 / Validation Rules | UI | ui/team-management | P2 |
| TC-006 | Story 6 / AC-2; UI Function 2 | UI | ui/sidebar | P0 |
| TC-007 | UI Function 2 / Validation Rules | UI | ui/sidebar | P1 |
| TC-008 | UI Function 2 / Validation Rules | UI | ui/sidebar | P1 |
| TC-009 | Story 6 / AC-1; UI Function 3 | UI | ui/role-management | P0 |
| TC-010 | UI Function 3 / States | UI | ui/role-management | P0 |
| TC-011 | UI Function 3 / States | UI | ui/role-management | P0 |
| TC-012 | UI Function 3 / Validation Rules | UI | ui/role-management | P1 |
| TC-013 | UI Function 3 / Validation Rules | UI | ui/role-management | P1 |
| TC-014 | Story 6 / AC-3 | UI | ui/team-management | P0 |
| TC-015 | Story 1 / AC-2 | API | api/roles | P0 |
| TC-016 | Story 1 / AC-3 | API | api/roles | P0 |
| TC-017 | Story 2 / AC-1 | API | api/users | P0 |
| TC-018 | Story 2 / AC-2 | API | api/users | P0 |
| TC-019 | Story 2 / AC-3 | API | api/users | P0 |
| TC-020 | Story 3 / AC-1 | API | api/roles | P0 |
| TC-021 | Story 3 / AC-2 | API | api/roles | P0 |
| TC-022 | Story 3 / AC-3 | API | api/roles | P0 |
| TC-023 | Story 3 / AC-4 | API | api/roles | P0 |
| TC-024 | Story 4 / AC-1 | API | api/users | P1 |
| TC-025 | Story 4 / AC-2 | API | api/users | P1 |
| TC-026 | Story 4 / AC-3 | API | api/users | P0 |
| TC-027 | Story 6 / AC-4 | API | api/permissions | P0 |
| TC-028 | Story 5 / AC-1 | API | api/migration | P0 |
| TC-029 | Story 5 / AC-2 | API | api/migration | P0 |
| TC-030 | Story 5 / AC-3 | API | api/migration | P0 |
| TC-031 | Spec Section 5.2 | API | api/permissions | P1 |
| TC-032 | Story 5 / AC-4 | CLI | cli/ci-grep | P0 |
