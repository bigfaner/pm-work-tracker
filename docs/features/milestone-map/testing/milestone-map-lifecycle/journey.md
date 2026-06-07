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
<!-- surface: web -->

**User Action**: PM opens the create dialog from the list page, fills in name (1-100 chars), owner, optional plan dates and description, then submits. <!-- fact: prd-spec Story 1 — initial status is planning -->

**Expected Result**: Milestone map is created with status "planning", the dialog closes, and the list refreshes showing the new entry with correct name, status badge, and owner info.

### Step 2: Edit milestone map information
<!-- surface: web -->

**User Action**: PM navigates to the milestone map detail page, opens the edit function, modifies the name and description, then saves. <!-- fact: prd-spec Story 2 -->

**Expected Result**: Changes are saved immediately, the edit dialog closes, and the detail page refreshes to show updated information.

### Step 3: Transition status from planning to reviewed
<!-- surface: web -->

**User Action**: PM selects the "Reviewed" status option on the detail page. <!-- fact: prd-spec MilestoneMap state machine — planning → reviewed -->

**Expected Result**: Status changes from "planning" to "reviewed", the badge updates visually, and no error occurs.

### Step 4: Transition status from reviewed to ready
<!-- surface: web -->

**User Action**: PM selects the "Ready for Implementation" status option. <!-- fact: prd-spec MilestoneMap state machine — reviewed → ready -->

**Expected Result**: Status changes from "reviewed" to "ready".

### Step 5: Transition status from ready to executing
<!-- surface: web -->

**User Action**: PM selects the "In Progress" status option. <!-- fact: prd-spec MilestoneMap state machine — ready → executing -->

**Expected Result**: Status changes from "ready" to "executing".

### Step 6: Transition status from executing to completed
<!-- surface: web -->

**User Action**: PM selects the "Completed" status option. All milestones under this map are in terminal states (completed or cancelled). <!-- fact: prd-spec — executing → completed requires all milestones in terminal states -->

**Expected Result**: Status changes from "executing" to "completed". This is a terminal state -- no further transitions are available.

### Step 7: Rollback status from reviewed back to planning
<!-- surface: web -->

**User Action**: PM selects the "Planning" rollback option when the map is in "reviewed" state. <!-- fact: prd-spec MilestoneMap state machine — reviewed → planning rollback -->

**Expected Result**: Status reverts from "reviewed" to "planning".

### Step 8: Delete milestone map in planning status
<!-- surface: web -->

**User Action**: PM triggers the delete action on a milestone map that is in "planning" status and confirms. <!-- fact: prd-spec Story 3 — soft delete with cascade -->

**Expected Result**: The milestone map and all its milestones are soft-deleted in a single transaction. All associated MainItems are unbound. The page redirects back to the list view.

## Edge Cases

### Step 1b: Create with missing required fields
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM submits the form without filling in the name field.

**Expected Result**: Form displays a validation error near the name field, the form is not submitted, and the dialog remains open.

### Step 1c: Create with name exceeding 100 characters
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM enters a name with 101 characters and submits. <!-- fact: prd-spec — name 1-100 chars -->

**Expected Result**: Form displays a validation error about name length. When the name is exactly 100 characters, creation succeeds.

### Step 1d: Create without selecting owner
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM fills the name but leaves the owner field empty and submits.

**Expected Result**: Form displays a validation error about the required owner field, the form is not submitted.

### Step 1e: Create with invalid date range
<!-- surface: web -->

**Precondition**: Create dialog is open with both plan start and plan end dates available. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM sets the plan end date earlier than the plan start date and submits.

**Expected Result**: Form displays a date range validation error, the form is not submitted.

### Step 1f: Create during server error
<!-- surface: web -->

**Precondition**: Create dialog is open with valid data entered; the backend is unavailable. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM submits the form, but the backend returns an error.

**Expected Result**: Page displays a retryable error message. The dialog stays open with all entered data preserved. No duplicate submission occurs.

### Step 1g: Create with loading state
<!-- surface: web -->

**Precondition**: Create dialog is open with valid data entered.

**User Action**: PM submits the form and the request is in flight.

**Expected Result**: The submit button shows a loading state and the form prevents further interaction until the request completes.

### Step 2b: Edit with no changes
<!-- surface: web -->

**Precondition**: Edit dialog is open with pre-filled current values.

**User Action**: PM saves without making any modifications.

**Expected Result**: The dialog closes as a no-op, equivalent to Cancel.

### Step 2c: Edit with concurrent modification conflict
<!-- surface: web -->

**Precondition**: Another PM has edited the same milestone map while the edit dialog is open. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM submits their changes.

**Expected Result**: A conflict notification appears. No silent overwrite occurs.

### Step 2d: Edit with invalid date range
<!-- surface: web -->

**Precondition**: Edit dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM modifies the plan end date to be earlier than the plan start date and submits.

**Expected Result**: Form displays date validation error and does not submit.

### Step 6b: Transition to completed with incomplete milestones
<!-- surface: web -->

**Precondition**: The milestone map is in "executing" status, but has at least one milestone that is not in a terminal state (not completed or cancelled). <!-- fact: prd-spec — completed transition requires all milestones in terminal states -->

**User Action**: PM attempts to transition the status to "Completed".

**Expected Result**: Status change is rejected with an error message indicating not all milestones are in terminal states.

### Step 8b: Delete non-deletable milestone map
<!-- surface: web -->

**Precondition**: The milestone map is in a non-deletable status ("executing" or "completed"). <!-- fact: prd-spec Story 3 — only planning, reviewed, ready statuses are deletable -->

**User Action**: PM views the detail page.

**Expected Result**: The delete action is not displayed. Only milestone maps in planning, reviewed, or ready status can be deleted.

### Step 8c: Delete without permission
<!-- surface: web -->

**Precondition**: PM does not have milestone:delete permission.

**User Action**: PM views the detail page of a "planning" status milestone map.

**Expected Result**: The delete action is not displayed.

### Step 8d: Cancel delete confirmation
<!-- surface: web -->

**Precondition**: Delete confirmation dialog is open.

**User Action**: PM cancels the operation.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3b: Status transition server error
<!-- surface: web -->

**Precondition**: The backend is unavailable when a status transition is attempted. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM selects a new status and the backend returns an error.

**Expected Result**: Error message is displayed, and the status reverts to the original value.

### Step 4b: Filter milestone maps by status
<!-- surface: web -->

**Precondition**: List page has multiple milestone maps with various statuses.

**User Action**: PM uses the status filter to select "In Progress".

**Expected Result**: Only milestone maps with "executing" status are shown in the list.

### Step 4c: Filter milestone maps by owner
<!-- surface: web -->

**Precondition**: List page has milestone maps owned by different team members.

**User Action**: PM uses the owner filter to select a specific team member.

**Expected Result**: Only milestone maps owned by that member are shown.

### Step 4d: Search milestone maps by name
<!-- surface: web -->

**Precondition**: List page has multiple milestone maps.

**User Action**: PM types a keyword in the name search box.

**Expected Result**: Only milestone maps whose names contain the keyword are shown.

### Step 6c: Completed status is terminal
<!-- surface: web -->

**Precondition**: Milestone map is in "completed" status. <!-- fact: prd-spec — completed is a terminal state -->

**User Action**: PM views the status options.

**Expected Result**: No transition options are available. Completed is a terminal state with no rollback.

### Step 1h: Create from empty state page
<!-- surface: web -->

**Precondition**: The team has 0 milestone maps and the empty state page is shown.

**User Action**: PM clicks the create button on the empty state page.

**Expected Result**: The create dialog opens.

### Step 1i: Create from grid dashed card
<!-- surface: web -->

**Precondition**: The milestone map list page has a dashed create card in the grid.

**User Action**: PM clicks the dashed create card.

**Expected Result**: The create dialog opens.

### Step E1: Session expired during form submission (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired while a form is open. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM submits a create or edit form.

**Expected Result**: The user is redirected to the login page; no data is modified. After re-authenticating, the form retains no pending changes.

### Step E2: Unauthenticated API access (API)
<!-- surface: api -->

**Precondition**: An API request is sent without valid credentials. <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An unauthenticated request is sent to any MilestoneMap endpoint.

**Expected Result**: The API returns an authentication error; no data is returned or modified.

## Journey Invariants

- A milestone map can only be deleted when it is in "planning", "reviewed", or "ready" status; the delete action is hidden for "executing", "completed", and "cancelled" statuses. <!-- fact: prd-spec Story 3 — deletable statuses -->
- Status transitions follow the defined state machine: planning -> reviewed -> ready -> executing -> completed, with rollback allowed from non-terminal states to their immediate predecessor only. "completed" and "cancelled" are terminal states with no transitions available. <!-- fact: prd-spec MilestoneMap state machine -->
- When a milestone map is deleted, all associated milestones are soft-deleted and all linked MainItems are unbound within the same transaction. <!-- fact: prd-spec Story 3 -->
- Transitioning to "completed" requires all milestones under the map to be in terminal states (completed or cancelled). <!-- fact: prd-spec — completed transition guard -->
- Any non-terminal state can transition to "cancelled", which is a terminal state. <!-- fact: prd-spec — cancelled transition -->
- All mutation operations (create, update, delete, status transition) require their respective RBAC permissions (milestone:create, milestone:update, milestone:delete).
