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

PM uses the milestone filter on the items list page and views the milestone column in the table view to quickly find and sort items by their milestone assignment. <!-- fact: prd-spec Story 12, Story 13 -->

## Setup

- Multiple MainItems exist, some assigned to milestones and some unassigned
- At least one milestone exists in the team in a non-cancelled state
- User has milestone:read permission
- Items list page and table view are accessible

## Happy Path

### Step 1: View milestone filter on items list page
<!-- surface: web -->

**User Action**: PM loads the items list page. <!-- fact: prd-spec Story 12 — milestone filter on items list -->

**Expected Result**: A "Milestone" dropdown filter appears in the filter bar with default value "All".

### Step 2: Filter by specific milestone
<!-- surface: web -->

**User Action**: PM opens the milestone dropdown and selects a specific milestone.

**Expected Result**: The list only shows MainItems belonging to that milestone.

### Step 3: Filter by "Unassigned"
<!-- surface: web -->

**User Action**: PM selects "Unassigned" from the milestone dropdown. <!-- fact: prd-spec — filter by unassigned MIs -->

**Expected Result**: The list only shows MainItems with no milestone assignment.

### Step 4: Filter by "All" (default)
<!-- surface: web -->

**User Action**: PM selects "All" from the milestone dropdown.

**Expected Result**: All MainItems are shown regardless of milestone assignment.

### Step 5: View milestone badge on list items
<!-- surface: web -->

**User Action**: PM views MainItem entries in the list. <!-- fact: prd-spec — milestone badge on items -->

**Expected Result**: Each MI assigned to a milestone shows a milestone name badge. Unassigned MIs do not show a badge.

### Step 6: View milestone column in table view
<!-- surface: web -->

**User Action**: PM switches to the table view. <!-- fact: prd-spec Story 13 — milestone column in table -->

**Expected Result**: A "Milestone" column appears showing the milestone name for assigned MIs and a placeholder for unassigned MIs.

### Step 7: Sort milestone column ascending
<!-- surface: web -->

**User Action**: PM triggers ascending sort on the milestone column header.

**Expected Result**: Assigned MIs are sorted by milestone name alphabetically ascending. Unassigned MIs appear at the bottom.

### Step 8: Sort milestone column descending
<!-- surface: web -->

**User Action**: PM triggers descending sort on the milestone column header.

**Expected Result**: Assigned MIs are sorted by milestone name alphabetically descending. Unassigned MIs still appear at the bottom.

## Edge Cases

### Step 2b: Filter by cancelled milestone -- not in dropdown
<!-- surface: web -->

**Precondition**: Team has milestones in "cancelled" status. <!-- fact: prd-spec — cancelled milestones excluded -->

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Cancelled milestones do not appear as options in the dropdown.

### Step 5b: Filter with invalid bizKey
<!-- surface: web -->

**Precondition**: An invalid bizKey is passed to the milestone filter.

**User Action**: Filter is triggered with an invalid value.

**Expected Result**: Filter falls back to "All" without producing an error.

### Step 6b: MI references soft-deleted milestone
<!-- surface: web -->

**Precondition**: A MainItem's milestone assignment points to a milestone that has been soft-deleted. <!-- fact: prd-spec — soft-deleted milestone data -->

**User Action**: PM views the table.

**Expected Result**: The milestone column for that MI displays a placeholder indicating the milestone is no longer available.

### Step 6c: Milestone data load failure in table
<!-- surface: web -->

**Precondition**: Milestone data API call fails. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM views the table.

**Expected Result**: Milestone column shows a fallback value for all rows. The rest of the table renders normally -- milestone load failure does not block table display.

### Step 6d: Default sort on first load
<!-- surface: web -->

**Precondition**: Table view has not been sorted by the user.

**User Action**: PM loads the table view for the first time.

**Expected Result**: Table loads with no default sort. Milestone column is unsorted initially. <!-- gap: no default sort by milestone on first load -->

### Step 6e: Filter bar dropdown by milestone
<!-- surface: web -->

**Precondition**: Table view is loaded with milestone filter available.

**User Action**: PM uses the milestone dropdown in the filter bar (not on the column header) to select a specific milestone.

**Expected Result**: Only MIs belonging to that milestone are shown.

### Step 7b: Switch team resets milestone filter
<!-- surface: web -->

**Precondition**: PM switches to a different team.

**User Action**: Team switch completes and new data loads.

**Expected Result**: React Query refetches data because teamId changes, but the milestone filter dropdown selection persists (does not reset to "All"). Dropdown options refresh to show the new team's milestones. <!-- gap: filter state does not reset on team switch -->

### Step 7c: Milestone dropdown load failure
<!-- surface: web -->

**Precondition**: Milestone options API call fails. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM opens the milestone dropdown.

**Expected Result**: Dropdown shows a load failure state and is disabled. Other filters on the page remain functional.

### Step E1: Session expired during list interaction (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM attempts to interact with the milestone filter or navigate to a new page.

**Expected Result**: The user is redirected to the login page. After re-authenticating, the items list reloads with default filters.

## Journey Invariants

- Cancelled milestones never appear in dropdown options or filter selections.
- Unassigned MIs consistently display without a badge or with a placeholder, and always sort to the bottom of milestone column sorts.
- Milestone-related failures (load errors, missing data) never block the rendering of the rest of the items list or table view.
- Switching teams refreshes the dropdown options but does NOT reset the milestone filter selection. <!-- gap: filter persists on team switch -->
- Invalid filter values gracefully fall back to "All" without errors.
