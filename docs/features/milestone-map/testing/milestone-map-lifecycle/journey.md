---
feature: "milestone-map"
journey: "milestone-map-lifecycle"
risk_level: "High"
surface_types: ["web", "api"]
surface_keys: ["frontend", "backend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: milestone-map-lifecycle

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
-->

## Overview

PM creates a milestone map, edits its information, transitions it through status stages, and optionally deletes it -- covering the full MilestoneMap CRUD lifecycle with status machine transitions.

## Setup

- User has milestone:create, milestone:update, and milestone:delete permissions
- At least one team exists in the system
- The milestone map list page (/milestones) is accessible

## Happy Path

### Step 1: Create milestone map via list page

**User Action**: PM clicks the "+ Create Milestone Map" button on the list page, fills in name (1-100 chars), owner, optional plan dates and description, then clicks Create.

**Expected Result**: Milestone map is created with status "planning", the dialog closes, and the list refreshes showing the new card with correct name, status badge, and owner info.

### Step 2: Edit milestone map information

**User Action**: PM navigates to the milestone map detail page, clicks the edit button in the title area, modifies the name and description, then clicks Save.

**Expected Result**: Changes are saved immediately, the edit dialog closes, and the detail page refreshes to show updated information.

### Step 3: Transition status from planning to reviewed

**User Action**: PM clicks the status Badge on the detail page, selects "Reviewed" from the dropdown menu.

**Expected Result**: Status changes from "planning" to "reviewed", the Badge updates visually, and no error occurs.

### Step 4: Transition status from reviewed to ready

**User Action**: PM clicks the status Badge and selects "Ready for Implementation".

**Expected Result**: Status changes from "reviewed" to "ready".

### Step 5: Transition status from ready to executing

**User Action**: PM clicks the status Badge and selects "In Progress".

**Expected Result**: Status changes from "ready" to "executing".

### Step 6: Transition status from executing to completed

**User Action**: PM clicks the status Badge and selects "Completed". All milestones under this map are in terminal states (completed or cancelled).

**Expected Result**: Status changes from "executing" to "completed". This is a terminal state -- no further transitions are available.

### Step 7: Rollback status from reviewed back to planning

**User Action**: PM clicks the status Badge when the map is in "reviewed" state and selects "Planning" (rollback option).

**Expected Result**: Status reverts from "reviewed" to "planning".

### Step 8: Delete milestone map in planning status

**User Action**: PM clicks the delete button on a milestone map that is in "planning" status, confirms in the confirmation dialog.

**Expected Result**: The milestone map and all its milestones are soft-deleted in a single transaction. All associated MainItems have their milestone_key set to null. The page redirects back to the list view.

## Edge Cases

### Step 1b: Create with missing required fields

**Precondition**: Create dialog is open.

**User Action**: PM submits the form without filling in the name field.

**Expected Result**: Form displays "Name cannot be empty" error near the name field, the form is not submitted, and the dialog remains open.

### Step 1c: Create with name exceeding 100 characters

**Precondition**: Create dialog is open.

**User Action**: PM enters a name with 101 characters and submits.

**Expected Result**: Form displays "Name cannot exceed 100 characters" error. When the name is exactly 100 characters, creation succeeds.

### Step 1d: Create without selecting owner

**Precondition**: Create dialog is open.

**User Action**: PM fills the name but leaves the owner field empty and submits.

**Expected Result**: Form displays "Please select an owner" error, the form is not submitted.

### Step 1e: Create with invalid date range

**Precondition**: Create dialog is open with both plan start and plan end dates available.

**User Action**: PM sets the plan end date earlier than the plan start date and submits.

**Expected Result**: Form displays "Plan end date must not be earlier than plan start date" error, the form is not submitted.

### Step 1f: Create during server error

**Precondition**: Create dialog is open with valid data entered.

**User Action**: PM submits the form, but the backend returns a 500 error.

**Expected Result**: Page displays "Creation failed, please retry" error message. The dialog stays open with all entered data preserved. No duplicate submission occurs.

### Step 1g: Create with loading state

**Precondition**: Create dialog is open with valid data entered.

**User Action**: PM submits the form and the request is in flight.

**Expected Result**: The create button shows a loading state, all input fields are disabled, preventing duplicate submission.

### Step 2b: Edit with no changes

**Precondition**: Edit dialog is open with pre-filled current values.

**User Action**: PM clicks Save without making any modifications.

**Expected Result**: The dialog closes as a no-op, equivalent to Cancel.

### Step 2c: Edit with concurrent modification conflict

**Precondition**: Another PM has edited the same milestone map while the edit dialog is open.

**User Action**: PM submits their changes.

**Expected Result**: Conflict notification appears: "Data has been modified by someone else, please refresh and retry". No silent overwrite occurs.

### Step 2d: Edit with invalid date range

**Precondition**: Edit dialog is open.

**User Action**: PM modifies the plan end date to be earlier than the plan start date and submits.

**Expected Result**: Form displays date validation error and does not submit.

### Step 6b: Transition to completed with incomplete milestones

**Precondition**: The milestone map is in "executing" status, but has at least one milestone that is not in a terminal state (not completed or cancelled).

**User Action**: PM attempts to transition the status to "Completed".

**Expected Result**: Status change is rejected with the message "All milestones must be completed before marking as complete" (BR-2).

### Step 8b: Delete non-planning milestone map

**Precondition**: The milestone map is in a status other than "planning" (e.g., "reviewed", "executing", "completed").

**User Action**: PM views the detail page.

**Expected Result**: The delete button is not displayed. Only milestone maps in "planning" status can be deleted (BR-4).

### Step 8c: Delete without permission

**Precondition**: PM does not have milestone:delete permission.

**User Action**: PM views the detail page of a "planning" status milestone map.

**Expected Result**: The delete button is not displayed.

### Step 8d: Cancel delete confirmation

**Precondition**: Delete confirmation dialog is open.

**User Action**: PM clicks Cancel.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3b: Status transition server error

**Precondition**: PM clicks the status Badge and selects a new status.

**User Action**: The backend returns an error for the status transition request.

**Expected Result**: Error message is displayed, and the Badge reverts to the original status.

### Step 4b: Filter milestone maps by status

**Precondition**: List page has multiple milestone maps with various statuses.

**User Action**: PM uses the status filter to select "In Progress".

**Expected Result**: Only milestone maps with "executing" status are shown in the list.

### Step 4c: Filter milestone maps by owner

**Precondition**: List page has milestone maps owned by different team members.

**User Action**: PM uses the owner filter to select a specific team member.

**Expected Result**: Only milestone maps owned by that member are shown.

### Step 4d: Search milestone maps by name

**Precondition**: List page has multiple milestone maps.

**User Action**: PM types a keyword in the name search box (debounce 300ms).

**Expected Result**: Only milestone maps whose names contain the keyword are shown.

### Step 6c: Completed status is terminal

**Precondition**: Milestone map is in "completed" status.

**User Action**: PM clicks the status Badge.

**Expected Result**: No transition options are available in the dropdown. Completed is a terminal state with no rollback.

### Step 1h: Create from empty state page

**Precondition**: The team has 0 milestone maps and the empty state page is shown.

**User Action**: PM clicks the create button on the empty state page.

**Expected Result**: The create dialog opens.

### Step 1i: Create from grid dashed card

**Precondition**: The milestone map list page has a dashed create card in the grid.

**User Action**: PM clicks the dashed create card.

**Expected Result**: The create dialog opens.

## Journey Invariants

- A milestone map can only be deleted when it is in "planning" status; the delete button is hidden in all other statuses.
- Status transitions follow the defined state machine: planning -> reviewed -> ready -> executing -> completed, with rollback allowed from any non-terminal state to its predecessor, and no transitions from terminal states (completed).
- When a milestone map is deleted, all associated milestones are soft-deleted and all linked MainItems have their milestone_key set to null within the same transaction.
- Transitioning to "completed" requires all milestones under the map to be in terminal states (completed or cancelled) -- BR-2.
- All mutation operations (create, update, delete, status transition) require their respective RBAC permissions (milestone:create, milestone:update, milestone:delete).
- The create/update forms must display loading state during submission and disable all inputs to prevent duplicate requests.
