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

PM creates a milestone within a milestone map's timeline, edits its information, transitions it through the four-state lifecycle (not_started -> in_progress -> completed -> cancelled), and optionally deletes it -- covering the full Milestone lifecycle including cancellation cascades that auto-unbind associated MainItems.

## Setup

- A milestone map exists and is accessible via its timeline view
- User has milestone:create, milestone:update, and milestone:delete permissions
- At least one MainItem exists in the same team

## Happy Path

### Step 1: Create milestone in timeline view

**User Action**: PM clicks the "+ Create Milestone" button in the timeline view, fills in name (1-100 chars), plan completion date (required), and optional description, then clicks Confirm.

**Expected Result**: Milestone is created with status "not_started" and completion 0. The dialog closes and the timeline refreshes showing the new node at the correct date position.

### Step 2: Edit milestone information

**User Action**: PM opens the milestone detail panel, clicks the edit button, modifies the name and plan completion date, then clicks Save.

**Expected Result**: Changes are saved, the dialog closes, the panel and timeline refresh. The node position on the timeline is recalculated based on the new date.

### Step 3: Transition status from not_started to in_progress

**User Action**: PM clicks the status Badge in the detail panel and selects "In Progress".

**Expected Result**: Status changes from "not_started" to "in_progress".

### Step 4: Transition status from in_progress to completed

**User Action**: PM clicks the status Badge and selects "Completed". All associated MainItems are in terminal states (completed/closed).

**Expected Result**: Status changes to "completed". Completion percentage reflects the associated MI average.

### Step 5: Transition status from completed back to cancelled

**User Action**: PM clicks the status Badge on a completed milestone and selects "Cancelled".

**Expected Result**: Status changes to "cancelled". This is a terminal state with no further transitions.

### Step 6: Delete milestone in not_started status

**User Action**: PM clicks "Delete Milestone" in the detail panel for a not_started milestone, confirms in the confirmation dialog.

**Expected Result**: The milestone is soft-deleted. Associated MainItems have their milestone_key set to null within the same transaction. The panel closes and timeline refreshes.

## Edge Cases

### Step 1b: Create with missing name

**Precondition**: Create dialog is open.

**User Action**: PM submits without filling in the name field.

**Expected Result**: Form displays "Name cannot be empty" error, the form is not submitted.

### Step 1c: Create with name exceeding 100 characters

**Precondition**: Create dialog is open.

**User Action**: PM enters a name with 101 characters and submits.

**Expected Result**: Form displays "Name cannot exceed 100 characters" error. When exactly 100 characters, creation succeeds.

### Step 1d: Create without plan completion date

**Precondition**: Create dialog is open.

**User Action**: PM fills the name but leaves the plan completion date empty and submits.

**Expected Result**: Form displays "Plan completion date cannot be empty" error, the form is not submitted.

### Step 1e: Create during server error

**Precondition**: Create dialog is open with valid data.

**User Action**: PM submits and the backend returns 500.

**Expected Result**: Page displays "Creation failed, please retry". Dialog stays open with data preserved.

### Step 1f: Create with loading state

**Precondition**: Create dialog is open with valid data.

**User Action**: PM submits the form and request is in flight.

**Expected Result**: Confirm button shows loading state, all input fields are disabled, preventing duplicate submission.

### Step 1g: Cancel create dialog

**Precondition**: Create dialog is open.

**User Action**: PM clicks Cancel or the X button.

**Expected Result**: Dialog closes, no operation is performed.

### Step 2b: Edit with no changes

**Precondition**: Edit dialog is open with pre-filled current values.

**User Action**: PM clicks Save without modifying anything.

**Expected Result**: Dialog closes as a no-op, equivalent to Cancel.

### Step 2c: Edit with concurrent modification conflict

**Precondition**: Another PM has edited the same milestone while the edit dialog is open.

**User Action**: PM submits their changes.

**Expected Result**: Conflict notification: "Data has been modified by someone else, please refresh and retry". No silent overwrite.

### Step 2d: Cancel edit dialog

**Precondition**: Edit dialog is open.

**User Action**: PM clicks Cancel or the X button.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3b: Transition to completed with incomplete MIs

**Precondition**: The milestone is in "in_progress" status and has at least one associated MainItem not in a terminal state.

**User Action**: PM attempts to transition to "Completed".

**Expected Result**: Status change is rejected with "All associated items must be completed before marking as complete" (BR-1).

### Step 4b: Transition not_started to cancelled with cascade

**Precondition**: The milestone is in "not_started" status and has associated MainItems.

**User Action**: PM clicks the status Badge and selects "Cancelled".

**Expected Result**: Status changes to "cancelled". All associated MainItems are auto-unbound (milestone_key set to null in a transaction). The detail panel shows an empty associated items list.

### Step 4c: Transition in_progress to cancelled with cascade

**Precondition**: The milestone is in "in_progress" status and has associated MainItems.

**User Action**: PM clicks the status Badge and selects "Cancelled".

**Expected Result**: Same cascade behavior as Step 4b -- all MIs auto-unbound, panel shows empty list.

### Step 5b: Cancelled is terminal -- no Badge dropdown

**Precondition**: Milestone is in "cancelled" status.

**User Action**: PM clicks the status Badge.

**Expected Result**: No dropdown menu appears. Cancelled is a terminal state with no recovery.

### Step 5c: Cancelled milestone cannot receive new MI bindings

**Precondition**: Milestone is in "cancelled" status.

**User Action**: PM tries to bind a MainItem to this cancelled milestone via the item edit dialog.

**Expected Result**: Binding is rejected (BR-3: cancelled milestones cannot receive new MainItems).

### Step 5d: Cancelled milestone panel appearance

**Precondition**: Milestone is in "cancelled" status.

**User Action**: PM views the detail panel.

**Expected Result**: Panel displays with a global grey tone (text-tertiary). Associated MI list is empty. "+ Add" button is not shown. Delete button is visible (BR-4 allows deleting cancelled milestones).

### Step 6b: Delete cancelled milestone

**Precondition**: Milestone is in "cancelled" status and PM has milestone:delete permission.

**User Action**: PM clicks "Delete Milestone" and confirms.

**Expected Result**: Milestone is soft-deleted successfully.

### Step 6c: Delete in_progress or completed milestone -- button hidden

**Precondition**: Milestone is in "in_progress" or "completed" status.

**User Action**: PM views the detail panel.

**Expected Result**: Delete button is not displayed. Only not_started and cancelled milestones can be deleted (BR-4).

### Step 6d: Delete without permission

**Precondition**: PM does not have milestone:delete permission.

**User Action**: PM views the detail panel.

**Expected Result**: Delete button is not displayed.

### Step 6e: Cancel delete confirmation

**Precondition**: Delete confirmation dialog is open.

**User Action**: PM clicks Cancel.

**Expected Result**: Dialog closes, no operation is performed.

### Step 3c: Status transition server error

**Precondition**: PM selects a new status from the Badge dropdown.

**User Action**: Backend returns an error for the transition request.

**Expected Result**: Error message is displayed, Badge reverts to original status, panel and timeline do not refresh.

## Journey Invariants

- Cancellation of a milestone (from any non-terminal state) automatically unbinds all associated MainItems by setting their milestone_key to null within the same transaction.
- A milestone can only be marked as "completed" when all its associated MainItems are in terminal states (completed or closed) -- BR-1.
- Cancelled milestones cannot receive new MainItem bindings -- BR-3.
- Delete is only available for milestones in "not_started" or "cancelled" status; the button is hidden for "in_progress" and "completed" -- BR-4.
- "cancelled" is a terminal state: no status transitions are available, no Badge dropdown appears.
- Status machine: not_started -> in_progress -> completed (with rollback to cancelled from any non-terminal state); completed -> cancelled (manual cancel); cancelled is terminal.
- All mutation operations require their respective RBAC permissions.
- Create/update forms display loading state and disable inputs during submission to prevent duplicate requests.
