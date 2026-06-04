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

PM user performs status transitions on items and uses conversion forms (todo-to-sub-item, todo-to-main-item, sub-item conversion) with proper error display, validation, and form cleanup.

## Setup

- PM user is logged in with appropriate permissions
- At least one item exists in a non-terminal status that can be transitioned
- At least one todo item exists that can be converted to a sub-item or main-item

## Happy Path

### Step 1: Trigger status transition with error

**User Action**: PM user clicks a status transition button for an item that cannot be transitioned due to backend business rules

**Expected Result**: Backend returns a specific error reason; frontend displays an inline Alert component below the action area with the backend-provided error message (not a tooltip that disappears after 2 seconds)

### Step 2: Successful status transition

**User Action**: PM user clicks a status transition button for an item that is eligible for transition

**Expected Result**: Status is updated; if transitioning to a terminal status, a confirmation dialog appears first. On confirm, the transition completes and page reflects the new status

### Step 3: Open conversion form with defaults and validation

**User Action**: PM user opens the todo-to-sub-item conversion form

**Expected Result**: Description field is disabled (greyed out, not editable); start time defaults to today; assignee and priority fields show required markers; submit button is disabled until both required fields are filled

### Step 4: Submit conversion form successfully

**User Action**: PM user fills in assignee and priority, then submits the conversion form

**Expected Result**: Conversion succeeds; page updates to reflect the new item; all form fields are cleared

### Step 5: Close and reopen conversion form

**User Action**: PM user closes (cancels) a partially filled conversion form, then opens any new/conversion form

**Expected Result**: All fields in the newly opened form are empty -- no residual data from the previously closed form

## Edge Cases

### Step 1b: Status transition conflict (concurrent edit)

**Precondition**: Another user has changed the item's status between the time the current user loaded the page and attempted the transition

**User Action**: PM user clicks a status transition button

**Expected Result**: Backend returns a conflict error; frontend displays "This item status has been updated by another user, please refresh and retry" as an inline error message

### Step 2b: Terminal status transition cancelled

**Precondition**: Item is eligible for transition to a terminal status

**User Action**: PM user clicks transition, confirmation dialog appears, user clicks cancel

**Expected Result**: No status change occurs; dialog closes; item remains in current status

### Step 3b: Open todo-to-main-item conversion form without required fields

**Precondition**: A todo item exists that can be converted to a main item

**User Action**: PM user opens the todo-to-main-item conversion form without selecting assignee or priority

**Expected Result**: Submit button is disabled; assignee and priority field labels display required markers

### Step 4b: Conversion form submission fails (backend validation)

**Precondition**: PM user has filled all required fields but backend rejects the submission (e.g., duplicate name, business rule violation)

**User Action**: PM user submits the conversion form

**Expected Result**: Form fields retain their values (not cleared); an error message is displayed for the user to correct and retry

### Step 5b: Reopen form after successful submission

**Precondition**: A conversion form was just submitted successfully

**User Action**: PM user opens another new or conversion form

**Expected Result**: All fields are empty -- no residual data from the previously submitted form

### Step 6b: Unauthorized status transition attempt

**Precondition**: A member-role user tries to access the status transition endpoint

**User Action**: Member user sends a status transition request via API

**Expected Result**: API returns 403 Forbidden; frontend does not show the transition button for unauthorized users

## Journey Invariants

- Inline error messages (Alert component) are always used for status transition errors, never auto-disappearing tooltips
- All conversion forms clear all fields on close/cancel or successful submission
- Required fields (assignee, priority) are enforced at both UI level (disabled submit) and API level (validation error response)
- Description field in todo-to-sub-item conversion form is always disabled and cannot be modified
