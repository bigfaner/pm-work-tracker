---
feature: "system-ux-optimization"
journey: "task-status-transition"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 1, Story 4)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: task-status-transition

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
  Status transition changes item state; conversion form creates new items (state mutation).
-->

## Overview

PM user performs status transitions on items and uses conversion forms (todo-to-sub-item, todo-to-main-item) with proper error display, validation, and form cleanup.

## Setup

- PM user is logged in with appropriate permissions
- At least one item exists in a non-terminal status that can be transitioned
- At least one todo item exists that can be converted to a sub-item or main-item

## Happy Path

### Step 1: Trigger status transition with error
<!-- surface: web -->

**Precondition**: An item exists in a status that has no valid transition to the selected target status, independent of concurrent modifications

**User Action**: PM user clicks a status transition button for that item

**Expected Result**: The item's status remains unchanged; an error message is displayed below the action area with content provided by the backend explaining why the transition cannot be performed

### Step 2: Successful status transition (non-terminal)
<!-- surface: web -->

**Precondition**: An item exists that is eligible for a non-terminal status transition, and no concurrent modification has occurred

**User Action**: PM user clicks a status transition button

**Expected Result**: The item's status is updated to the new value; the page reflects the new status

### Step 3: Successful status transition (terminal with confirmation)
<!-- surface: web -->

**Precondition**: An item exists that is eligible for transition to a terminal status

**User Action**: PM user clicks the terminal status transition button, a confirmation dialog appears, user confirms

**Expected Result**: The transition completes; the page reflects the new terminal status

### Step 4: Open conversion form with defaults and validation
<!-- surface: web -->

**Precondition**: A todo item exists that can be converted to a sub-item

**User Action**: PM user opens the todo-to-sub-item conversion form

**Expected Result**: Description field is disabled (greyed out, not editable); start date defaults to today; assignee and priority fields show required markers (*); submit button is disabled until both required fields are filled

### Step 5: Submit conversion form successfully
<!-- surface: web -->

**Precondition**: The todo-to-sub-item conversion form is open with assignee and priority filled in

**User Action**: PM user submits the conversion form

**Expected Result**: Conversion succeeds; page updates to reflect the new sub-item; all form fields are cleared

### Step 6: Close and reopen conversion form
<!-- surface: web -->

**Precondition**: A conversion form has been partially filled

**User Action**: PM user closes (cancels) the form, then opens any new or conversion form

**Expected Result**: All fields in the newly opened form are empty — no residual data from the previously closed form

### Step 7: Submit todo-to-main-item conversion successfully
<!-- surface: web -->

**Precondition**: A todo item exists that can be converted to a main item; the todo-to-main-item conversion form is open with all required fields filled

**User Action**: PM user submits the todo-to-main-item conversion form

**Expected Result**: Conversion succeeds; page updates to reflect the new main item; all form fields are cleared

## Edge Cases

### Step E1: Status transition conflict (concurrent edit)
<!-- surface: web -->

**Precondition**: Another user has changed the item's status after the current user loaded the page; the current user's transition is now based on stale data

**User Action**: PM user clicks a status transition button

**Expected Result**: An inline error message displays: "该事项状态已被他人更新，请刷新后重试"; the item's status remains unchanged

### Step E2: Terminal status transition cancelled
<!-- surface: web -->

**Precondition**: An item is eligible for transition to a terminal status

**User Action**: PM user clicks the terminal transition, confirmation dialog appears, user clicks cancel

**Expected Result**: No status change occurs; the dialog closes; the item remains in its current status

### Step E3: Open todo-to-main-item conversion form without required fields
<!-- surface: web -->

**Precondition**: A todo item exists that can be converted to a main item

**User Action**: PM user opens the todo-to-main-item conversion form without selecting assignee or priority

**Expected Result**: Submit button is disabled; assignee and priority field labels display required markers (*)

### Step E4: Conversion form submission fails (backend validation)
<!-- surface: web -->

**Precondition**: All required fields are filled but the backend rejects the submission — a sub-item with the same name already exists under the target parent item <!-- source: inferred — derived from conversion form backend validation failure described in prd-spec -->

**User Action**: PM user submits the conversion form

**Expected Result**: Form fields retain their values (not cleared); an error message is displayed indicating the specific validation failure; the user can correct and retry

### Step E5: Reopen form after successful submission
<!-- surface: web -->

**Precondition**: A conversion form was just submitted successfully

**User Action**: PM user opens another new or conversion form

**Expected Result**: All fields are empty — no residual data from the previously submitted form

### Step E6: Unauthorized status transition attempt (API)
<!-- surface: api -->

**Precondition**: The user has member role only, which does not include the status transition permission <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: Member-role user sends a status transition API request

**Expected Result**: The API returns an authorization error; the frontend does not display status transition controls for users without the required permission

### Step E7: Unauthorized conversion form API request
<!-- surface: api -->

**Precondition**: The user has member role only, which does not include the conversion form submission permission <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: Member-role user sends a conversion form submission API request

**Expected Result**: The API returns an authorization error; no item is created

### Step E8: API-level validation error for conversion form
<!-- surface: api -->

**Precondition**: A direct API request is sent for conversion form submission with missing assignee or priority fields <!-- source: inferred — derived from API surface `validation-error` outcome and Web surface `validation-error` mandatory outcome -->

**User Action**: An API request submits a conversion form with empty required fields

**Expected Result**: The API returns a validation error response listing the missing fields; no item is created

### Step E9: Session expired during form interaction (Web)
<!-- surface: web -->

**Precondition**: The user's session has expired while they are filling a conversion form <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM user submits the conversion form

**Expected Result**: The user is redirected to the login page; after re-authenticating, the user can access the form again (previous unsaved data is not preserved)

### Step E10: Validation error on conversion form submission (Web)
<!-- surface: web -->

**Precondition**: The conversion form is open and a direct API request or browser developer tools bypass has cleared the assignee field value <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM user submits the form with an empty assignee field

**Expected Result**: An error message is displayed near the assignee field indicating that it is required; the form is not submitted; the user can fill the field and retry

## Journey Invariants

- Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips
- All conversion forms clear all fields on close/cancel or successful submission
- Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)
- Description field in the todo-to-sub-item conversion form is always disabled and cannot be modified
