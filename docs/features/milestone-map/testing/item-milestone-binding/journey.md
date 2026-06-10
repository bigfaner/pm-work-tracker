---
feature: "milestone-map"
journey: "item-milestone-binding"
risk_level: "High"
surface_types: ["web"]
surface_keys: ["frontend"]
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

PM uses the MainItem edit dialog to bind a MainItem to a milestone, change its milestone assignment, or unbind it entirely, with validation rules enforcing team consistency, terminal state restrictions, and cancelled milestone constraints. <!-- fact: prd-spec Story 11 -->

## Setup

- A MainItem exists and is not in a terminal state
- At least one milestone exists in the same team, in a non-cancelled state
- User has milestone:update permission
- The MainItem edit dialog is accessible

## Happy Path

### Step 1: Open edit dialog and view milestone selector
<!-- surface: web -->

**User Action**: PM opens the MainItem edit dialog. <!-- fact: prd-spec Story 11 — milestone selector in edit dialog -->

**Expected Result**: Dialog loads with a "Milestone" dropdown selector showing the current assignment or "Unassigned".

### Step 2: Bind unassigned MI to a milestone
<!-- surface: web -->

**Precondition**: MainItem is not assigned to any milestone. <!-- fact: prd-spec — bind sets milestone_key -->

**User Action**: PM selects a milestone from the dropdown and saves.

**Expected Result**: MI is bound to the selected milestone. The milestone's completion percentage is recalculated. <!-- fact: prd-spec — bind sets milestone_key, triggers completion recalc -->

### Step 3: Rebind MI to a different milestone
<!-- surface: web -->

**Precondition**: MainItem is already assigned to milestone A.

**User Action**: PM selects milestone B from the dropdown and saves. <!-- fact: prd-spec — rebind updates milestone_key, auto-unbinds from old -->

**Expected Result**: MI is rebound to milestone B (auto-unbound from A). Both milestones A and B have their completion percentages recalculated. <!-- fact: prd-spec — rebind updates milestone_key, auto-unbinds from old -->

### Step 4: Unbind MI from milestone
<!-- surface: web -->

**Precondition**: MainItem is assigned to a milestone.

**User Action**: PM selects "Unassigned" from the dropdown and saves. <!-- fact: prd-spec — unbind sets milestone_key to null -->

**Expected Result**: MI is unbound from the milestone. The original milestone's completion percentage is recalculated. <!-- fact: prd-spec — unbind sets milestone_key to null -->

### Step 5: Save with no changes
<!-- surface: web -->

**Precondition**: MainItem is assigned to a milestone.

**User Action**: PM opens the edit dialog, does not modify the milestone field, and saves.

**Expected Result**: No change occurs. The original milestone assignment is preserved.

## Edge Cases

### Step 1b: No milestones exist in the team
<!-- surface: web -->

**Precondition**: Team has zero milestones.

**User Action**: PM opens the edit dialog and views the milestone dropdown.

**Expected Result**: Dropdown only shows "Unassigned" option.

### Step 1c: Cancelled milestones excluded from dropdown
<!-- surface: web -->

**Precondition**: Team has milestones including some in "cancelled" status. <!-- fact: prd-spec — cancelled milestones cannot receive new MIs -->

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Cancelled milestones do not appear in the dropdown options.

### Step 2b: Bind MI in terminal state
<!-- surface: web -->

**Precondition**: MainItem is in a terminal state (completed or closed). <!-- fact: prd-spec — terminal state MI cannot change milestone assignment -->

**User Action**: PM selects a milestone and saves.

**Expected Result**: No client-side terminal state guard exists. Server rejection returns error but no user-facing feedback (mutation has no onError handler). <!-- gap: no terminal state guard, no error feedback -->

### Step 2c: Bind MI to milestone in different team
<!-- surface: web -->

**Precondition**: The selected milestone belongs to a different team than the MainItem. <!-- fact: prd-spec — team consistency constraint -->

**User Action**: PM selects the cross-team milestone and saves.

**Expected Result**: Operation is rejected with an error indicating cross-team constraint.

### Step 2d: Bind MI to cancelled milestone
<!-- surface: web -->

**Precondition**: A milestone in "cancelled" status exists (though it should not appear in the dropdown). <!-- fact: prd-spec — cancelled milestones cannot receive new MIs -->

**User Action**: PM attempts to select and bind to a cancelled milestone.

**Expected Result**: Binding is rejected. <!-- fact: prd-spec — cancelled milestones cannot receive new MainItems -->

### Step 3b: Rebind -- original milestone completion recalculated
<!-- surface: web -->

**Precondition**: MainItem is the last non-terminal MI on milestone A. <!-- fact: prd-spec — completion recalculated on rebind -->

**User Action**: PM rebinds MI to milestone B.

**Expected Result**: Milestone A's completion percentage updates to reflect the removal. Milestone B's completion percentage updates to reflect the addition.

### Step 4b: Unbind triggers completion recalculation
<!-- surface: web -->

**Precondition**: MainItem contributes to the milestone's completion calculation.

**User Action**: PM unbinds the MI.

**Expected Result**: The milestone's completion percentage is recalculated excluding the unbound MI.

### Step 2e: Save fails with server validation error
<!-- surface: web -->

**Precondition**: Edit dialog is open with a milestone selected, but the backend rejects the save (e.g., the milestone was deleted by another user while the dialog was open). <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM saves the milestone assignment.

**Expected Result**: Dialog stays open (no explicit error toast shown). The mutation has no onError handler. <!-- gap: no error feedback on mutation failure -->

### Step 5b: Milestone filter with no matching results — NOT_APPLICABLE
<!-- surface: web -->

**Note**: The milestone dropdown is a simple flat list, not a searchable filter. This edge case does not apply to the current implementation.

### Step E1: Session expired during save (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired while the edit dialog is open. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM submits the bind/unbind form.

**Expected Result**: The user is redirected to the login page; no data is modified. After re-authenticating, the MainItem retains its original milestone assignment.

## Journey Invariants

- Terminal state MainItems (completed/closed) cannot have their milestone assignment changed. <!-- fact: prd-spec — terminal state restriction -->
- Cancelled milestones cannot receive new MainItem bindings. <!-- fact: prd-spec — cancelled binding restriction -->
- Cross-team binding is always rejected with an appropriate error message. <!-- fact: prd-spec — team consistency -->
- Every bind/rebind/unbind operation triggers completion recalculation on the affected milestone(s).
- The milestone dropdown never shows cancelled milestones as selectable options.
- When no milestones exist in the team, the dropdown only shows "Unassigned".
