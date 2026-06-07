---
feature: "milestone-map"
journey: "item-list-milestone-integration"
risk_level: "Low"
surface_types: ["web"]
surface_keys: ["frontend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: item-list-milestone-integration

**Risk Level**: Low

<!-- Risk Classification Criteria:
  Low    = Workflow is read-only or purely observational
-->

## Overview

PM uses the milestone filter on the items list page and views the milestone column in the table view to quickly find and sort items by their milestone assignment.

## Setup

- Multiple MainItems exist, some assigned to milestones and some unassigned
- At least one milestone exists in the team in a non-cancelled state
- User has milestone:read permission
- Items list page and table view are accessible

## Happy Path

### Step 1: View milestone filter on items list page

**User Action**: PM loads the items list page.

**Expected Result**: A "Milestone" dropdown filter appears in the filter bar, to the right of the "Owner" filter, with default value "All".

### Step 2: Filter by specific milestone

**User Action**: PM opens the milestone dropdown and selects a specific milestone.

**Expected Result**: The list only shows MainItems belonging to that milestone.

### Step 3: Filter by "Unassigned"

**User Action**: PM selects "Unassigned" from the milestone dropdown.

**Expected Result**: The list only shows MainItems with an empty milestone_key.

### Step 4: Filter by "All" (default)

**User Action**: PM selects "All" from the milestone dropdown.

**Expected Result**: All MainItems are shown regardless of milestone assignment.

### Step 5: View milestone badge on list items

**User Action**: PM views MainItem entries in the list.

**Expected Result**: Each MI assigned to a milestone shows a milestone name badge (rounded-full, text-xs) next to it. Unassigned MIs do not show a badge.

### Step 6: View milestone column in table view

**User Action**: PM switches to the table view.

**Expected Result**: A "Milestone" column appears between the "Title" and "Priority" columns (width w-32). Assigned MIs show the milestone name (text-secondary). Unassigned MIs show "-" (text-tertiary).

### Step 7: Sort milestone column ascending

**User Action**: PM clicks the milestone column header sort icon and selects ascending order.

**Expected Result**: Assigned MIs are sorted by milestone name alphabetically ascending. Unassigned MIs appear at the bottom.

### Step 8: Sort milestone column descending

**User Action**: PM clicks the milestone column header sort icon and selects descending order.

**Expected Result**: Assigned MIs are sorted by milestone name alphabetically descending. Unassigned MIs still appear at the bottom.

## Edge Cases

### Step 2b: Filter by cancelled milestone -- not in dropdown

**Precondition**: Team has milestones in "cancelled" status.

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Cancelled milestones do not appear as options in the dropdown.

### Step 5b: Filter with invalid bizKey

**Precondition**: An invalid bizKey is passed to the milestone filter.

**User Action**: Filter is triggered with an invalid value.

**Expected Result**: Filter falls back to "All" without producing an error.

### Step 6b: MI references soft-deleted milestone

**Precondition**: A MainItem's milestone_key points to a milestone that has been soft-deleted.

**User Action**: PM views the table.

**Expected Result**: The milestone column for that MI displays "--".

### Step 6c: Milestone data load failure in table

**Precondition**: Milestone data API call fails.

**User Action**: PM views the table.

**Expected Result**: Milestone column shows "--" for all rows. The rest of the table renders normally -- milestone load failure does not block table display.

### Step 6d: Default sort on first load

**Precondition**: Table view has not been sorted by the user.

**User Action**: PM loads the table view for the first time.

**Expected Result**: Milestone column defaults to ascending order. Unassigned MIs appear at the bottom.

### Step 6e: Column header filter by milestone

**Precondition**: Table view is loaded with milestone column visible.

**User Action**: PM uses the column header filter to select a specific milestone.

**Expected Result**: Only MIs belonging to that milestone are shown.

### Step 7b: Switch team resets milestone filter

**Precondition**: PM switches to a different team.

**User Action**: Team switch completes and new data loads.

**Expected Result**: Milestone filter resets to "All". Dropdown options refresh to show the new team's milestones.

### Step 7c: Milestone dropdown load failure

**Precondition**: Milestone options API call fails.

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Dropdown shows "Load failed" and is disabled. Other filters on the page remain functional.

## Journey Invariants

- Cancelled milestones never appear in dropdown options or filter selections.
- Unassigned MIs consistently display as "-" or without a badge, and always sort to the bottom of milestone column sorts.
- Milestone-related failures (load errors, missing data) never block the rendering of the rest of the items list or table view.
- Switching teams always resets the milestone filter to "All" and refreshes the dropdown options.
- Invalid filter values gracefully fall back to "All" without errors.
