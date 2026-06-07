---
feature: "milestone-map"
journey: "milestone-item-management"
risk_level: "High"
surface_types: ["web", "api"]
surface_keys: ["frontend", "backend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: milestone-item-management

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
-->

## Overview

PM interacts with the milestone detail panel to view, unbind, and quick-add associated MainItems within a milestone's context, including drag-and-drop rebinding on the timeline and handling cancelled milestone states.

## Setup

- A milestone map exists with at least one milestone that has associated MainItems
- User has milestone:update and milestone:create permissions
- At least one unassigned MainItem exists in the same team

## Happy Path

### Step 1: Open milestone detail panel

**User Action**: PM clicks a milestone node on the timeline.

**Expected Result**: Detail panel slides in from the right (translate-x 300ms). Panel shows: name row (name + close button), description area (label + status Badge + edit button, description text max 6 lines + Tooltip), plan completion date, progress (label + progress bar + percentage), associated MI list, danger zone (delete button).

### Step 2: View associated MI list with hover unbind

**User Action**: PM hovers over a MainItem row in the associated items list.

**Expected Result**: An X (unbind) button appears on the right side of the MI row.

### Step 3: Unbind a MainItem

**User Action**: PM clicks the X button on a MainItem row.

**Expected Result**: The MI is removed from the list. An undo toast appears for 5 seconds. The milestone completion percentage is recalculated.

### Step 4: Quick-add a MainItem

**User Action**: PM clicks the "+ Add" button in the panel, fills in title, owner, start date, and expected completion date in the CreateMainItemDialog (milestone field pre-filled and disabled), then clicks Confirm.

**Expected Result**: MainItem is created and auto-bound to the current milestone. Dialog closes. Panel MI list refreshes. Completion recalculates.

### Step 5: Navigate to MI detail from panel

**User Action**: PM clicks a MainItem's number or title in the panel list.

**Expected Result**: Route navigates to /items/:mainItemId main item detail page.

### Step 6: Drag-and-drop MI rebinding on timeline

**User Action**: PM drags a MainItem entry from one milestone node to another milestone node on the timeline.

**Expected Result**: MI is rebound to the target milestone (milestone_key updated via API). During drag, MI shows opacity-50 and target milestone highlights. On completion, an undo toast appears for 5 seconds. Both source and target milestone completion percentages are recalculated.

### Step 7: Close panel via overlay click

**User Action**: PM clicks the overlay area outside the panel.

**Expected Result**: Panel closes with slide-out animation. Focus returns to the milestone node that triggered the panel.

## Edge Cases

### Step 1b: Panel loading skeleton

**Precondition**: Panel opens but data is still loading.

**User Action**: PM opens the panel.

**Expected Result**: Panel shows a skeleton screen during data loading.

### Step 1c: Description tooltip on overflow

**Precondition**: Description text exceeds 6 lines in the panel.

**User Action**: PM hovers over the description text.

**Expected Result**: Tooltip displays the full description content.

### Step 1d: Close panel via Escape key

**Precondition**: Panel is open.

**User Action**: PM presses the Escape key.

**Expected Result**: Panel closes with slide-out animation. Focus returns to the triggering milestone node.

### Step 1e: Close panel via X button

**Precondition**: Panel is open.

**User Action**: PM clicks the X close button.

**Expected Result**: Panel closes.

### Step 1f: No edit permission -- edit button hidden

**Precondition**: PM does not have milestone:update permission.

**User Action**: PM views the panel.

**Expected Result**: Edit button is not displayed. Status Badge is not clickable.

### Step 3b: Undo unbind within 5 seconds

**Precondition**: PM has just unbound a MainItem and the undo toast is showing.

**User Action**: PM clicks "Undo" on the toast within 5 seconds.

**Expected Result**: The MI is re-bound to the milestone. The list and completion percentage are restored.

### Step 4b: Quick-add without title

**Precondition**: Quick-add dialog is open.

**User Action**: PM submits without filling in the title.

**Expected Result**: Form displays title required error and does not submit.

### Step 4c: Quick-add without owner

**Precondition**: Quick-add dialog is open.

**User Action**: PM submits without selecting an owner.

**Expected Result**: Form displays owner required error and does not submit.

### Step 4d: Quick-add without start or end date

**Precondition**: Quick-add dialog is open.

**User Action**: PM submits without filling in start date or expected completion date.

**Expected Result**: Form displays the corresponding required field error and does not submit.

### Step 4e: Quick-add loading state

**Precondition**: Quick-add dialog is open with valid data.

**User Action**: PM submits the form.

**Expected Result**: Confirm button shows loading state, all input fields are disabled.

### Step 4f: Quick-add milestone field is locked

**Precondition**: Quick-add dialog is open.

**User Action**: PM views the milestone field.

**Expected Result**: The milestone field shows the current milestone name and is disabled -- cannot be modified.

### Step 5b: Cancelled milestone panel appearance

**Precondition**: Milestone is in "cancelled" status.

**User Action**: PM opens the detail panel.

**Expected Result**: Panel displays with global grey tone (text-tertiary). Associated MI list is empty (MIs auto-unbound on cancel). "+ Add" button is not shown. Delete button is visible (BR-4).

### Step 6b: Drag-and-drop -- MI belongs to different team

**Precondition**: The target milestone belongs to a different team than the MI.

**User Action**: PM drags an MI to the target milestone.

**Expected Result**: Rebinding is rejected with "Item and milestone do not belong to the same team" error.

### Step 6c: Drag-and-drop -- MI is in terminal state

**Precondition**: The MI is in a terminal state (completed/closed).

**User Action**: PM drags the MI to another milestone.

**Expected Result**: Rebinding is rejected (BR-3: terminal state MI cannot change milestone assignment).

### Step 6d: Drag-and-drop -- target milestone is cancelled

**Precondition**: The target milestone is in "cancelled" status.

**User Action**: PM drags an MI to the cancelled milestone.

**Expected Result**: Rebinding is rejected (BR-3: cancelled milestones cannot receive new MIs).

### Step 7b: Delete button hidden for in_progress/completed milestones

**Precondition**: Milestone is in "in_progress" or "completed" status.

**User Action**: PM views the danger zone in the panel.

**Expected Result**: Delete button is not displayed (BR-4: only not_started and cancelled can be deleted).

## Journey Invariants

- The milestone detail panel always animates in (300ms translate-x) and out (slide-out), returning focus to the triggering node on close.
- Unbind operations show an undo toast valid for 5 seconds; if undone within that window, the binding is restored.
- Quick-add MainItem dialog always pre-fills and locks the milestone field to the current milestone.
- Completion percentage is recalculated whenever MIs are bound, unbound, or rebound.
- Cancelled milestones display with a global grey tone, empty MI list, no add button, but visible delete button.
- All mutation operations (unbind, add, drag-rebind) require milestone:update permission; edit/delete buttons are hidden without permission.
