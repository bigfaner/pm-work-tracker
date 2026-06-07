---
feature: "milestone-map"
journey: "item-milestone-binding"
risk_level: "High"
surface_types: ["web", "api"]
surface_keys: ["frontend", "backend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: item-milestone-binding

**Risk Level**: High

<!-- Risk Classification Criteria:
  High   = Workflow involves state mutation, data loss risk, or irreversible operations
-->

## Overview

PM uses the MainItem edit dialog to bind a MainItem to a milestone, change its milestone assignment, or unbind it entirely, with validation rules enforcing team consistency, terminal state restrictions, and cancelled milestone constraints.

## Setup

- A MainItem exists and is not in a terminal state
- At least one milestone exists in the same team, in a non-cancelled state
- User has milestone:update permission
- The MainItem edit dialog is accessible

## Happy Path

### Step 1: Open edit dialog and view milestone selector

**User Action**: PM opens the MainItem edit dialog.

**Expected Result**: Dialog loads with a "Milestone" dropdown selector (positioned below the "Owner" field) showing the current assignment or "Unassigned".

### Step 2: Bind unassigned MI to a milestone

**Precondition**: MainItem is not assigned to any milestone.

**User Action**: PM selects a milestone from the dropdown and clicks Save.

**Expected Result**: MI's milestone_key is set to the selected milestone's bizKey. The milestone's completion percentage is recalculated.

### Step 3: Rebind MI to a different milestone

**Precondition**: MainItem is already assigned to milestone A.

**User Action**: PM selects milestone B from the dropdown and clicks Save.

**Expected Result**: MI is rebound to milestone B (auto-unbound from A). Both milestones A and B have their completion percentages recalculated.

### Step 4: Unbind MI from milestone

**Precondition**: MainItem is assigned to a milestone.

**User Action**: PM selects "Unassigned" from the dropdown and clicks Save.

**Expected Result**: MI's milestone_key is set to null. The original milestone's completion percentage is recalculated.

### Step 5: Save with no changes

**Precondition**: MainItem is assigned to a milestone.

**User Action**: PM opens the edit dialog, does not modify the milestone field, and clicks Save.

**Expected Result**: No change occurs. The original milestone assignment is preserved.

## Edge Cases

### Step 1b: No milestones exist in the team

**Precondition**: Team has zero milestones.

**User Action**: PM opens the edit dialog and views the milestone dropdown.

**Expected Result**: Dropdown only shows "Unassigned" option.

### Step 1c: Cancelled milestones excluded from dropdown

**Precondition**: Team has milestones including some in "cancelled" status.

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Cancelled milestones do not appear in the dropdown options.

### Step 2b: Bind MI in terminal state

**Precondition**: MainItem is in a terminal state (completed or closed).

**User Action**: PM selects a milestone and clicks Save.

**Expected Result**: Operation is rejected (BR-3: terminal state MI cannot change milestone assignment).

### Step 2c: Bind MI to milestone in different team

**Precondition**: The selected milestone belongs to a different team than the MainItem.

**User Action**: PM selects the cross-team milestone and clicks Save.

**Expected Result**: Operation is rejected with "Item and milestone do not belong to the same team" error.

### Step 2d: Bind MI to cancelled milestone

**Precondition**: A milestone in "cancelled" status exists (though it should not appear in the dropdown).

**User Action**: PM attempts to bind to a cancelled milestone (e.g., via direct API call).

**Expected Result**: Binding is rejected (BR-3: cancelled milestones cannot receive new MIs).

### Step 3b: Rebind -- original milestone completion recalculated

**Precondition**: MainItem is the last non-terminal MI on milestone A.

**User Action**: PM rebinds MI to milestone B.

**Expected Result**: Milestone A's completion percentage updates to reflect the removal. Milestone B's completion percentage updates to reflect the addition.

### Step 4b: Unbind triggers completion recalculation

**Precondition**: MainItem contributes to the milestone's completion calculation.

**User Action**: PM unbinds the MI.

**Expected Result**: The milestone's completion percentage is recalculated excluding the unbound MI.

### Step 5b: Invalid bizKey in filter

**Precondition**: An invalid bizKey is passed to the milestone filter.

**User Action**: Filter is triggered with an invalid value.

**Expected Result**: Filter falls back to "All" without producing an error.

## Journey Invariants

- Terminal state MainItems (completed/closed) cannot have their milestone_key changed -- BR-3.
- Cancelled milestones cannot receive new MainItem bindings -- BR-3.
- Cross-team binding is always rejected with an appropriate error message.
- Every bind/rebind/unbind operation triggers completion recalculation on the affected milestone(s).
- The milestone dropdown never shows cancelled milestones as selectable options.
- When no milestones exist in the team, the dropdown only shows "Unassigned".
