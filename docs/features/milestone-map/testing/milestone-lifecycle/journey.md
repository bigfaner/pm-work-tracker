---
feature: "milestone-map"
journey: "milestone-lifecycle"
risk_level: "High"
surface_types: ["web", "api"]
surface_keys: ["frontend", "backend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: milestone-lifecycle

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
-->

## Overview

PM creates a milestone within a milestone map's timeline, edits its information, transitions it through the four-state lifecycle (not_started -> in_progress -> completed -> cancelled), and optionally deletes it -- covering the full Milestone lifecycle including cancellation cascades that auto-unbind associated MainItems. <!-- fact: prd-spec Milestone lifecycle -->

## Setup

- A milestone map exists and is accessible via its timeline view
- User has milestone:create, milestone:update, and milestone:delete permissions
- At least one MainItem exists in the same team

## Happy Path

### Step 1: Create milestone in timeline view
<!-- surface: web -->

**User Action**: PM opens the create milestone function in the timeline view, fills in name (1-100 chars), plan completion date (required), and optional description, then confirms. <!-- fact: prd-spec Story 5 — initial status is not_started -->

**Expected Result**: Milestone is created with status "not_started" and completion 0. The dialog closes and the timeline refreshes showing the new node at the correct date position.

### Step 2: Edit milestone information
<!-- surface: web -->

**User Action**: PM opens the milestone detail panel, triggers the edit function, modifies the name and plan completion date, then saves. <!-- fact: prd-spec Story 6 -->

**Expected Result**: Changes are saved, the dialog closes, the panel and timeline refresh. The node position on the timeline is recalculated based on the new date.

### Step 3: Transition status from not_started to in_progress
<!-- surface: web -->

**User Action**: PM selects the "In Progress" status option in the detail panel. <!-- fact: prd-spec Milestone state machine — not_started → in_progress -->

**Expected Result**: Status changes from "not_started" to "in_progress".

### Step 4: Transition status from in_progress to completed
<!-- surface: web -->

**User Action**: PM selects the "Completed" status option. All associated MainItems are in terminal states (completed/closed). <!-- fact: prd-spec — in_progress → completed requires all MIs in terminal states -->

**Expected Result**: Status changes to "completed". Completion percentage reflects the associated MI average.

### Step 5: Transition status from completed back to cancelled
<!-- surface: web -->

**User Action**: PM selects the "Cancelled" status option on a completed milestone. <!-- fact: prd-spec Milestone state machine — completed → cancelled -->

**Expected Result**: Status changes to "cancelled". All associated MainItems are auto-unbound in the same transaction. This is a terminal state with no further transitions.

### Step 6: Delete milestone in not_started status
<!-- surface: web -->

**User Action**: PM triggers the "Delete Milestone" action for a not_started milestone and confirms. <!-- fact: prd-spec Story 7 — soft delete -->

**Expected Result**: The milestone is soft-deleted. Associated MainItems are unbound within the same transaction. The panel closes and timeline refreshes.

## Edge Cases

### Step 1b: Create with missing name
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM submits without filling in the name field.

**Expected Result**: Form displays a validation error about the required name field, the form is not submitted.

### Step 1c: Create with name exceeding 100 characters
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM enters a name with 101 characters and submits. <!-- fact: prd-spec — name 1-100 chars -->

**Expected Result**: Form displays a validation error about name length. When exactly 100 characters, creation succeeds.

### Step 1d: Create without plan completion date
<!-- surface: web -->

**Precondition**: Create dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM fills the name but leaves the plan completion date empty and submits.

**Expected Result**: Form displays a validation error about the required date field, the form is not submitted.

### Step 1e: Create during server error
<!-- surface: web -->

**Precondition**: Create dialog is open with valid data; the backend is unavailable. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM submits and the backend returns an error.

**Expected Result**: Page displays a retryable error message. Dialog stays open with data preserved.

### Step 1f: Create with loading state
<!-- surface: web -->

**Precondition**: Create dialog is open with valid data.

**User Action**: PM submits the form and request is in flight.

**Expected Result**: Confirm button shows loading state and the form prevents further interaction.

### Step 1g: Cancel create dialog
<!-- surface: web -->

**Precondition**: Create dialog is open.

**User Action**: PM cancels or closes the dialog.

**Expected Result**: Dialog closes, no operation is performed.

### Step 2b: Edit with no changes
<!-- surface: web -->

**Precondition**: Edit dialog is open with pre-filled current values.

**User Action**: PM saves without modifying anything.

**Expected Result**: Dialog closes as a no-op, equivalent to Cancel.

### Step 2c: Edit with concurrent modification conflict
<!-- surface: web -->

**Precondition**: Another PM has edited the same milestone while the edit dialog is open. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM submits their changes.

**Expected Result**: A conflict notification appears. No silent overwrite.

### Step 2d: Cancel edit dialog
<!-- surface: web -->

**Precondition**: Edit dialog is open.

**User Action**: PM cancels or closes the dialog.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3b: Transition to completed with incomplete MIs
<!-- surface: web -->

**Precondition**: The milestone is in "in_progress" status and has at least one associated MainItem not in a terminal state. <!-- fact: prd-spec — completed requires all MIs in terminal states -->

**User Action**: PM attempts to transition to "Completed".

**Expected Result**: Status change is rejected with an error indicating not all items are in terminal states.

### Step 4b: Transition not_started to cancelled with cascade
<!-- surface: web -->

**Precondition**: The milestone is in "not_started" status and has associated MainItems. <!-- fact: prd-spec Milestone state machine — not_started → cancelled -->

**User Action**: PM selects the "Cancelled" status option.

**Expected Result**: Status changes to "cancelled". All associated MainItems are auto-unbound within the same transaction. The detail panel shows an empty associated items list.

### Step 4c: Transition in_progress to cancelled with cascade
<!-- surface: web -->

**Precondition**: The milestone is in "in_progress" status and has associated MainItems. <!-- fact: prd-spec Milestone state machine — in_progress → cancelled -->

**User Action**: PM selects the "Cancelled" status option.

**Expected Result**: Same cascade behavior as Step 4b -- all MIs auto-unbound, panel shows empty list.

### Step 5b: Cancelled is terminal -- no status options
<!-- surface: web -->

**Precondition**: Milestone is in "cancelled" status. <!-- fact: prd-spec — cancelled is a terminal state -->

**User Action**: PM views the status options.

**Expected Result**: No dropdown menu appears. Cancelled is a terminal state with no recovery.

### Step 5c: Cancelled milestone cannot receive new MI bindings
<!-- surface: web -->

**Precondition**: Milestone is in "cancelled" status. <!-- fact: prd-spec — cancelled milestones cannot receive new MainItems -->

**User Action**: PM tries to bind a MainItem to this cancelled milestone.

**Expected Result**: Binding is rejected with an appropriate error message.

### Step 5d: Cancelled milestone panel appearance
<!-- surface: web -->

**Precondition**: Milestone is in "cancelled" status.

**User Action**: PM views the detail panel.

**Expected Result**: Panel displays with a muted visual tone. Associated MI list is empty. Add button is not shown. Delete button is visible. <!-- fact: prd-spec Story 7 — cancelled milestones can be deleted -->

### Step 6b: Delete cancelled milestone
<!-- surface: web -->

**Precondition**: Milestone is in "cancelled" status and PM has milestone:delete permission. <!-- fact: prd-spec — not_started and cancelled can be deleted -->

**User Action**: PM triggers "Delete Milestone" and confirms.

**Expected Result**: Milestone is soft-deleted successfully.

### Step 6c: Delete in_progress or completed milestone -- action hidden
<!-- surface: web -->

**Precondition**: Milestone is in "in_progress" or "completed" status. <!-- fact: prd-spec — only not_started and cancelled can be deleted -->

**User Action**: PM views the detail panel.

**Expected Result**: Delete action is not displayed.

### Step 6d: Delete without permission
<!-- surface: web -->

**Precondition**: PM does not have milestone:delete permission.

**User Action**: PM views the detail panel.

**Expected Result**: Delete action is not displayed.

### Step 6e: Cancel delete confirmation
<!-- surface: web -->

**Precondition**: Delete confirmation dialog is open.

**User Action**: PM cancels the operation.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3c: Status transition server error
<!-- surface: web -->

**Precondition**: The backend is unavailable when a status transition is attempted. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM selects a new status and the backend returns an error.

**Expected Result**: Error message is displayed, status reverts to original, panel and timeline do not refresh.

### Step E1: Session expired during form submission (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired while a form is open. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM submits the create or edit form.

**Expected Result**: The user is redirected to the login page; no data is modified. After re-authenticating, the milestone retains its original values.

### Step E2: Unauthenticated API access (API)
<!-- surface: api -->

**Precondition**: An API request is sent without valid credentials. <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An unauthenticated request is sent to any Milestone endpoint.

**Expected Result**: The API returns an authentication error; no data is returned or modified.

## Journey Invariants

- Cancellation of a milestone (from any non-terminal state) automatically unbinds all associated MainItems within the same transaction. <!-- fact: prd-spec — cancel cascade -->
- A milestone can only be marked as "completed" when all its associated MainItems are in terminal states (completed or closed). <!-- fact: prd-spec — completed transition guard -->
- Cancelled milestones cannot receive new MainItem bindings. <!-- fact: prd-spec — cancelled binding restriction -->
- Delete is only available for milestones in "not_started" or "cancelled" status; the action is hidden for "in_progress" and "completed". <!-- fact: prd-spec Story 7 -->
- "cancelled" is a terminal state: no status transitions are available. <!-- fact: prd-spec Milestone state machine -->
- Status machine: not_started -> in_progress -> completed (with rollback to cancelled from any non-terminal state); completed -> cancelled (manual cancel); completed -> in_progress (reopen); cancelled is terminal. <!-- fact: prd-spec Milestone state machine -->
- All mutation operations require their respective RBAC permissions.
- Create/update forms display loading state and prevent further interaction during submission.
