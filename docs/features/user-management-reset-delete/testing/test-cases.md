---
feature: "user-management-reset-delete"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
generated: "2026-04-27"
---

# Test Cases: user-management-reset-delete

## Summary

| Type | Count |
|------|-------|
| UI   | 16   |
| API  | 10  |
| CLI  | 0  |
| **Total** | **26** |

---

## UI Test Cases

## TC-001: Super admin can see reset password and delete buttons
- **Source**: Story 5 / AC-1, UI Function 1
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/super-admin-can-see-reset-password-and-delete-buttons
- **Pre-conditions**: Logged in as super admin; at least one other user exists in the list
- **Route**: /users
- **Element**: E-160
- **Steps**:
  1. Log in as super admin
  2. Navigate to /users page
  3. Observe the action column for each user row
- **Expected**: Each user row displays "reset password" and "delete" buttons in the action column
- **Priority**: P0

## TC-002: Non-super-admin cannot see reset password and delete buttons
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/non-super-admin-cannot-see-reset-password-and-delete-buttons
- **Pre-conditions**: Logged in as non-super-admin user; user management page is accessible
- **Route**: /users
- **Element**: E-160
- **Steps**:
  1. Log in as a non-super-admin user
  2. Navigate to /users page (if permitted)
  3. Observe the action column for each user row
- **Expected**: "Reset password" and "Delete" buttons are NOT rendered in any user row
- **Priority**: P0

## TC-003: Delete button disabled on own row with tooltip
- **Source**: Story 4 / AC-1, UI Function 1 States
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/delete-button-disabled-on-own-row-with-tooltip
- **Pre-conditions**: Logged in as super admin; current user's own row is visible in the list
- **Route**: /users
- **Element**: E-160
- **Steps**:
  1. Log in as super admin
  2. Navigate to /users page
  3. Find the row matching the logged-in user
  4. Observe the delete button state on that row
  5. Hover over the disabled delete button
- **Expected**: The delete button is disabled (grey, not clickable). Tooltip displays "Cannot delete your own account" (or equivalent message)
- **Priority**: P0

## TC-004: Clicking reset password opens dialog with user display name
- **Source**: Story 1 / AC-1, UI Function 2
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/clicking-reset-password-opens-dialog-with-display-name
- **Pre-conditions**: Logged in as super admin; at least one other user exists
- **Route**: /users
- **Element**: E-160
- **Steps**:
  1. Log in as super admin
  2. Navigate to /users page
  3. Click "Reset Password" button on a user row (e.g. user with displayName "Alice Wang")
  4. Observe the dialog title
- **Expected**: A modal dialog opens with title containing "Reset Password" and the user's display name (e.g. "Reset Password -- Alice Wang")
- **Priority**: P0

## TC-005: Reset password empty validation on submit
- **Source**: Story 2 / AC-1, UI Function 2 Validation Rules
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/reset-password-empty-validation-on-submit
- **Pre-conditions**: Reset password dialog is open for a user
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog for a user
  2. Leave both password fields empty
  3. Click the confirm/submit button
- **Expected**: Dialog stays open. Red error text "Please enter a new password" (or equivalent) appears below the new password field. No network request is sent.
- **Priority**: P0

## TC-006: Reset password strength validation on blur and submit
- **Source**: Story 2 / AC-1, PRD Spec Section 5.3 Validation Rules
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/reset-password-strength-validation-on-blur-and-submit
- **Pre-conditions**: Reset password dialog is open for a user
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog for a user
  2. Enter a password shorter than 8 characters (e.g. "abc12") into the new password field
  3. Blur the field (click elsewhere)
  4. Observe error message
  5. Enter a password with 8+ chars but only letters (e.g. "abcdefgh")
  6. Blur the field
  7. Observe error message
  8. Enter a valid password (e.g. "NewPass123") and blur
  9. Observe no error message
- **Expected**: Invalid passwords (too short, or missing letters/digits) show error "Password must be at least 8 characters and contain letters and digits" (or equivalent) on blur. Valid passwords show no error.
- **Priority**: P0

## TC-007: Reset password confirm mismatch validation
- **Source**: Story 2 / AC-1, UI Function 2 Validation Rules
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/reset-password-confirm-mismatch-validation
- **Pre-conditions**: Reset password dialog is open for a user
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog
  2. Enter a valid new password (e.g. "NewPass123")
  3. Enter a different password in the confirm field (e.g. "Different1")
  4. Blur the confirm field or click submit
- **Expected**: Red error text "The two passwords entered do not match" (or equivalent) appears below the confirm password field. Dialog stays open. No network request is sent.
- **Priority**: P0

## TC-008: Reset password success flow
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/reset-password-success-flow
- **Pre-conditions**: Logged in as super admin; reset password dialog is open with valid input
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog for a user
  2. Enter a valid new password (e.g. "NewPass123") that meets strength rules
  3. Enter matching confirm password
  4. Click confirm/submit button
  5. Wait for API response (200)
- **Expected**: Dialog closes. A toast notification displays "Password has been reset" (or equivalent success message). The user can log in with the new password.
- **Priority**: P0

## TC-009: Reset password API error keeps dialog open
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/reset-password-api-error-keeps-dialog-open
- **Pre-conditions**: Reset password dialog is open with valid input; backend will return error
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog for a user
  2. Enter valid password and confirm password
  3. Click submit
  4. Simulate backend returning non-200 response (e.g. 500 internal error or network timeout)
- **Expected**: Dialog remains open. Error message from the backend (or "Network error, please retry") is displayed inside the dialog. User can modify and resubmit.
- **Priority**: P1

## TC-010: Clicking delete opens confirmation dialog with username
- **Source**: Story 3 / AC-1, UI Function 3
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/clicking-delete-opens-confirmation-dialog-with-username
- **Pre-conditions**: Logged in as super admin; at least one other user exists
- **Route**: /users
- **Element**: E-160
- **Steps**:
  1. Log in as super admin
  2. Navigate to /users page
  3. Click "Delete" button on a user row (e.g. user with username "bob")
  4. Observe the confirmation dialog content
- **Expected**: A confirmation dialog opens. The dialog body includes the target username (e.g. "Confirm deleting user bob? This action cannot be undone via the UI."). Both "Confirm Delete" and "Cancel" buttons are visible and clickable.
- **Priority**: P0

## TC-011: Delete user success removes row and shows toast
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/delete-user-success-removes-row-and-shows-toast
- **Pre-conditions**: Logged in as super admin; delete confirmation dialog is open for a user
- **Route**: /users
- **Steps**:
  1. Open the delete confirmation dialog for a user
  2. Click "Confirm Delete"
  3. Wait for API response (200)
- **Expected**: Dialog closes. The deleted user's row disappears from the list. A toast notification displays "User deleted" (or equivalent).
- **Priority**: P0

## TC-012: Delete user 404 shows message and removes row
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/delete-user-404-shows-message-and-removes-row
- **Pre-conditions**: Logged in as super admin; target user was already deleted by another admin (stale list)
- **Route**: /users
- **Steps**:
  1. Open the delete confirmation dialog for a user whose row is stale (already deleted by another admin)
  2. Click "Confirm Delete"
  3. Simulate backend returning 404 / USER_NOT_FOUND
- **Expected**: The row is removed from the list. A toast or inline message displays "This user has been deleted or does not exist" (or equivalent).
- **Priority**: P1

## TC-013: Delete user API error shows error in dialog
- **Source**: UI Function 3 States
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/delete-user-api-error-shows-error-in-dialog
- **Pre-conditions**: Logged in as super admin; delete confirmation dialog is open
- **Route**: /users
- **Steps**:
  1. Open the delete confirmation dialog for a user
  2. Click "Confirm Delete"
  3. Simulate backend returning a non-404 error (e.g. 500)
- **Expected**: Dialog remains open. Error message is displayed inside the dialog.
- **Priority**: P1

## TC-014: Copy credentials button copies to clipboard
- **Source**: Story 6 / AC-1, UI Function 4
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/copy-credentials-button-copies-to-clipboard
- **Pre-conditions**: Super admin has just created a new user; result dialog is showing with username and initialPassword
- **Route**: /users
- **Steps**:
  1. Create a new user successfully (result dialog appears with username and initial password)
  2. Click the "Copy Account & Password" button
  3. Check clipboard contents
- **Expected**: Clipboard contains "Account: {username}\nPassword: {initialPassword}" (or equivalent format). Button text briefly changes to "Copied" and reverts after about 2 seconds.
- **Priority**: P1

## TC-015: Copy credentials failure shows error toast
- **Source**: Story 6 / AC-2, UI Function 4 States
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/copy-credentials-failure-shows-error-toast
- **Pre-conditions**: Create-user result dialog is open; browser Clipboard API is unavailable or permission denied
- **Route**: /users
- **Steps**:
  1. Create a new user successfully (result dialog appears)
  2. Simulate Clipboard API failure (browser does not support or permission denied)
  3. Click "Copy Account & Password" button
- **Expected**: A toast error message "Copy failed, please select and copy manually" (or equivalent) is displayed. Button text does NOT change to "Copied".
- **Priority**: P2

## TC-016: Password visibility toggle in reset password dialog
- **Source**: UI Function 2 (password field with show/hide toggle)
- **Type**: UI
- **Target**: ui/users
- **Test ID**: ui/users/password-visibility-toggle-in-reset-password-dialog
- **Pre-conditions**: Reset password dialog is open
- **Route**: /users
- **Steps**:
  1. Open the reset password dialog for a user
  2. Enter text in the new password field
  3. Click the eye/toggle icon next to the password field
  4. Observe the field changes from masked to visible
  5. Click the toggle icon again
  6. Observe the field changes back to masked
- **Expected**: Password field toggles between masked (dots) and visible (plain text) when the toggle icon is clicked.
- **Priority**: P2

---

## API Test Cases

## TC-017: Reset password with valid request returns 200
- **Source**: Story 1 / AC-1, API Handbook - Reset Password
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/reset-password-with-valid-request-returns-200
- **Pre-conditions**: Authenticated as super admin; target user exists and is not deleted
- **Steps**:
  1. Send `PUT /api/v1/admin/users/:userId/password` with valid JWT and body `{"newPassword": "NewPass123"}`
- **Expected**: Response status 200. Response body contains `{ "code": 0, "data": { "bizKey": "...", "username": "...", "displayName": "..." } }`
- **Priority**: P0

## TC-018: Reset password without auth returns 401
- **Source**: API Handbook - Reset Password Error Responses, PRD Spec Security Requirements
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/reset-password-without-auth-returns-401
- **Pre-conditions**: No JWT token provided
- **Steps**:
  1. Send `PUT /api/v1/admin/users/:userId/password` without Authorization header, body `{"newPassword": "NewPass123"}`
- **Expected**: Response status 401 with error code UNAUTHORIZED
- **Priority**: P0

## TC-019: Reset password by non-super-admin returns 403
- **Source**: Story 5 / AC-1, API Handbook - Reset Password Error Responses, PRD Spec Security Requirements
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/reset-password-by-non-super-admin-returns-403
- **Pre-conditions**: Authenticated as a non-super-admin user
- **Steps**:
  1. Send `PUT /api/v1/admin/users/:userId/password` with a non-super-admin JWT, body `{"newPassword": "NewPass123"}`
- **Expected**: Response status 403 with error code FORBIDDEN
- **Priority**: P0

## TC-020: Reset password with weak password returns 400
- **Source**: PRD Spec Section 5.3 Validation Rules, API Handbook - Reset Password Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/reset-password-with-weak-password-returns-400
- **Pre-conditions**: Authenticated as super admin
- **Steps**:
  1. Send `PUT /api/v1/admin/users/:userId/password` with body `{"newPassword": "abc"}`
- **Expected**: Response status 400 with error code VALIDATION_ERROR
- **Priority**: P0

## TC-021: Reset password for non-existent user returns 404
- **Source**: API Handbook - Reset Password Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/reset-password-for-non-existent-user-returns-404
- **Pre-conditions**: Authenticated as super admin; target userId does not exist
- **Steps**:
  1. Send `PUT /api/v1/admin/users/:nonExistentId/password` with body `{"newPassword": "NewPass123"}`
- **Expected**: Response status 404 with error code USER_NOT_FOUND
- **Priority**: P1

## TC-022: Delete user with valid request returns 200
- **Source**: Story 3 / AC-1, API Handbook - Delete User
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/delete-user-with-valid-request-returns-200
- **Pre-conditions**: Authenticated as super admin; target user exists and is not the caller
- **Steps**:
  1. Send `DELETE /api/v1/admin/users/:userId` with valid super-admin JWT
- **Expected**: Response status 200 with `{ "code": 0, "data": null }`
- **Priority**: P0

## TC-023: Delete user without auth returns 401
- **Source**: API Handbook - Delete User Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/delete-user-without-auth-returns-401
- **Pre-conditions**: No JWT token provided
- **Steps**:
  1. Send `DELETE /api/v1/admin/users/:userId` without Authorization header
- **Expected**: Response status 401 with error code UNAUTHORIZED
- **Priority**: P0

## TC-024: Delete user by non-super-admin returns 403
- **Source**: Story 5 / AC-1, API Handbook - Delete User Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/delete-user-by-non-super-admin-returns-403
- **Pre-conditions**: Authenticated as a non-super-admin user
- **Steps**:
  1. Send `DELETE /api/v1/admin/users/:userId` with a non-super-admin JWT
- **Expected**: Response status 403 with error code FORBIDDEN
- **Priority**: P0

## TC-025: Delete self returns 422
- **Source**: Story 4, API Handbook - Delete User Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/delete-self-returns-422
- **Pre-conditions**: Authenticated as super admin
- **Steps**:
  1. Send `DELETE /api/v1/admin/users/:ownUserId` with the caller's own JWT
- **Expected**: Response status 422 with error code CANNOT_DELETE_SELF
- **Priority**: P0

## TC-026: Delete non-existent user returns 404
- **Source**: Story 3 / AC-2, API Handbook - Delete User Error Responses
- **Type**: API
- **Target**: api/admin-users
- **Test ID**: api/admin-users/delete-non-existent-user-returns-404
- **Pre-conditions**: Authenticated as super admin; target userId does not exist or already deleted
- **Steps**:
  1. Send `DELETE /api/v1/admin/users/:nonExistentId` with valid super-admin JWT
- **Expected**: Response status 404 with error code USER_NOT_FOUND
- **Priority**: P1

---

## CLI Test Cases

_No CLI test cases — this feature has no CLI components._

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 5 / AC-1, UI Function 1 | UI | ui/users | P0 |
| TC-002 | Story 5 / AC-1 | UI | ui/users | P0 |
| TC-003 | Story 4 / AC-1, UI Function 1 States | UI | ui/users | P0 |
| TC-004 | Story 1 / AC-1, UI Function 2 | UI | ui/users | P0 |
| TC-005 | Story 2 / AC-1, UI Function 2 Validation Rules | UI | ui/users | P0 |
| TC-006 | Story 2 / AC-1, PRD Spec 5.3 | UI | ui/users | P0 |
| TC-007 | Story 2 / AC-1, UI Function 2 Validation Rules | UI | ui/users | P0 |
| TC-008 | Story 1 / AC-1 | UI | ui/users | P0 |
| TC-009 | Story 1 / AC-2 | UI | ui/users | P1 |
| TC-010 | Story 3 / AC-1, UI Function 3 | UI | ui/users | P0 |
| TC-011 | Story 3 / AC-1 | UI | ui/users | P0 |
| TC-012 | Story 3 / AC-2 | UI | ui/users | P1 |
| TC-013 | UI Function 3 States | UI | ui/users | P1 |
| TC-014 | Story 6 / AC-1, UI Function 4 | UI | ui/users | P1 |
| TC-015 | Story 6 / AC-2, UI Function 4 States | UI | ui/users | P2 |
| TC-016 | UI Function 2 | UI | ui/users | P2 |
| TC-017 | Story 1 / AC-1, API Handbook | API | api/admin-users | P0 |
| TC-018 | API Handbook, PRD Security | API | api/admin-users | P0 |
| TC-019 | Story 5 / AC-1, API Handbook | API | api/admin-users | P0 |
| TC-020 | PRD Spec 5.3, API Handbook | API | api/admin-users | P0 |
| TC-021 | API Handbook | API | api/admin-users | P1 |
| TC-022 | Story 3 / AC-1, API Handbook | API | api/admin-users | P0 |
| TC-023 | API Handbook | API | api/admin-users | P0 |
| TC-024 | Story 5 / AC-1, API Handbook | API | api/admin-users | P0 |
| TC-025 | Story 4, API Handbook | API | api/admin-users | P0 |
| TC-026 | Story 3 / AC-2, API Handbook | API | api/admin-users | P1 |
