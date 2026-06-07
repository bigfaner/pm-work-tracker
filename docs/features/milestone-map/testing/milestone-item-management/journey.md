---
feature: "milestone-map"
journey: "milestone-item-management"
risk_level: "High"
surface_types: ["web"]
surface_keys: ["frontend"]
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

PM interacts with the milestone detail panel to view, unbind, and quick-add associated MainItems within a milestone's context, including drag-and-drop rebinding on the timeline and handling cancelled milestone states. <!-- fact: prd-spec Story 10 -->

## Setup

- A milestone map exists with at least one milestone that has associated MainItems
- User has milestone:update and milestone:create permissions
- At least one unassigned MainItem exists in the same team

## Happy Path

### Step 1: Open milestone detail panel
<!-- surface: web -->

**User Action**: PM clicks a milestone node on the timeline. <!-- fact: prd-spec Story 9 — timeline node interaction -->

**Expected Result**: Detail panel opens showing: name row with close control, description area with status badge and edit control (description text with overflow handling), plan completion date, progress bar with percentage, associated MI list, and danger zone with delete action.

### Step 2: View associated MI list with hover unbind
<!-- surface: web -->

**User Action**: PM hovers over a MainItem row in the associated items list.

**Expected Result**: An unbind control appears on the MI row.

### Step 3: Unbind a MainItem
<!-- surface: web -->

**User Action**: PM triggers the unbind action on a MainItem row. <!-- fact: prd-spec — unbind sets milestone_key to null -->

**Expected Result**: The MI is removed from the list. An undo option appears briefly. The milestone completion percentage is recalculated.

### Step 4: Quick-add a MainItem
<!-- surface: web -->

**User Action**: PM opens the add function in the panel, fills in title, owner, start date, and expected completion date (milestone field pre-filled and locked), then confirms. <!-- fact: prd-spec Story 10 — quick-add creates and auto-binds MI -->

**Expected Result**: MainItem is created and auto-bound to the current milestone. Dialog closes. Panel MI list refreshes. Completion recalculates.

### Step 5: Navigate to MI detail from panel
<!-- surface: web -->

**User Action**: PM clicks a MainItem's number or title in the panel list.

**Expected Result**: Route navigates to the main item detail page.

### Step 6: Drag-and-drop MI rebinding on timeline
<!-- surface: web -->

**User Action**: PM drags a MainItem entry from one milestone node to another milestone node on the timeline. <!-- fact: prd-spec — drag-drop updates milestone_key -->

**Expected Result**: MI is rebound to the target milestone. Visual feedback is shown during drag. On completion, an undo option appears briefly. Both source and target milestone completion percentages are recalculated.

### Step 7: Close panel via overlay click
<!-- surface: web -->

**Precondition**: Panel is open.

**User Action**: PM clicks the overlay area outside the panel.

**Expected Result**: Panel closes. Focus returns to the milestone node that triggered the panel.

## Edge Cases

### Step 1b: Panel loading skeleton
<!-- surface: web -->

**Precondition**: Panel opens but data is still loading.

**User Action**: PM opens the panel.

**Expected Result**: Panel shows a loading placeholder during data loading.

### Step 1c: Description tooltip on overflow
<!-- surface: web -->

**Precondition**: Description text overflows the panel area.

**User Action**: PM hovers over the description text.

**Expected Result**: Tooltip displays the full description content.

### Step 1d: Close panel via Escape key
<!-- surface: web -->

**Precondition**: Panel is open.

**User Action**: PM presses the Escape key.

**Expected Result**: Panel closes. Focus returns to the triggering milestone node.

### Step 1e: Close panel via close control
<!-- surface: web -->

**Precondition**: Panel is open.

**User Action**: PM clicks the close control.

**Expected Result**: Panel closes.

### Step 1f: No edit permission -- edit control hidden
<!-- surface: web -->

**Precondition**: PM does not have milestone:update permission.

**User Action**: PM views the panel.

**Expected Result**: Edit control is not displayed. Status badge is not interactive.

### Step 3b: Undo unbind within brief window
<!-- surface: web -->

**Precondition**: PM has just unbound a MainItem and the undo option is visible.

**User Action**: PM triggers undo within the allowed time window.

**Expected Result**: The MI is re-bound to the milestone. The list and completion percentage are restored.

### Step 4b: Quick-add without title
<!-- surface: web -->

**Precondition**: Quick-add dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM submits without filling in the title.

**Expected Result**: Form displays a validation error and does not submit.

### Step 4c: Quick-add without owner
<!-- surface: web -->

**Precondition**: Quick-add dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM submits without selecting an owner.

**Expected Result**: Form displays a validation error and does not submit.

### Step 4d: Quick-add without start or end date
<!-- surface: web -->

**Precondition**: Quick-add dialog is open. <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM submits without filling in start date or expected completion date.

**Expected Result**: Form displays the corresponding required field error and does not submit.

### Step 4e: Quick-add loading state
<!-- surface: web -->

**Precondition**: Quick-add dialog is open with valid data.

**User Action**: PM submits the form.

**Expected Result**: Confirm control shows loading state and the form prevents further interaction.

### Step 4f: Quick-add milestone field is locked
<!-- surface: web -->

**Precondition**: Quick-add dialog is open.

**User Action**: PM views the milestone field.

**Expected Result**: The milestone field shows the current milestone name and cannot be modified.

### Step 5b: Cancelled milestone panel appearance
<!-- surface: web -->

**Precondition**: Milestone is in "cancelled" status. <!-- fact: prd-spec — cancelled milestones -->

**User Action**: PM opens the detail panel.

**Expected Result**: Panel displays with a muted visual tone. Associated MI list is empty (MIs auto-unbound on cancel). Add control is not shown. Delete action is visible.

### Step 6b: Drag-and-drop -- MI belongs to different team
<!-- surface: web -->

**Precondition**: The target milestone belongs to a different team than the MI. <!-- fact: prd-spec — team consistency constraint -->

**User Action**: PM drags an MI to the target milestone.

**Expected Result**: Rebinding is rejected with an error indicating cross-team constraint.

### Step 6c: Drag-and-drop -- MI is in terminal state
<!-- surface: web -->

**Precondition**: The MI is in a terminal state (completed/closed). <!-- fact: prd-spec — terminal state MI cannot change milestone -->

**User Action**: PM drags the MI to another milestone.

**Expected Result**: Rebinding is rejected.

### Step 6d: Drag-and-drop -- target milestone is cancelled
<!-- surface: web -->

**Precondition**: The target milestone is in "cancelled" status. <!-- fact: prd-spec — cancelled milestones cannot receive new MIs -->

**User Action**: PM drags an MI to the cancelled milestone.

**Expected Result**: Rebinding is rejected.

### Step 7b: Delete action hidden for in_progress/completed milestones
<!-- surface: web -->

**Precondition**: Milestone is in "in_progress" or "completed" status. <!-- fact: prd-spec — only not_started and cancelled can be deleted -->

**User Action**: PM views the danger zone in the panel.

**Expected Result**: Delete action is not displayed.

### Step E1: Session expired during panel interaction (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired while the panel is open. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM attempts an action (unbind, add, drag-rebind) in the panel.

**Expected Result**: The user is redirected to the login page; no data is modified. After re-authenticating, the panel retains its original state.

## Journey Invariants

- Unbind operations show an undo option valid for a brief window; if undone within that window, the binding is restored.
- Quick-add MainItem dialog always pre-fills and locks the milestone field to the current milestone.
- Completion percentage is recalculated whenever MIs are bound, unbound, or rebound.
- Cancelled milestones display with a muted tone, empty MI list, no add control, but visible delete action.
- All mutation operations (unbind, add, drag-rebind) require milestone:update permission; edit/delete controls are hidden without permission.
