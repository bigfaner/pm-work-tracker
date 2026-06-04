---
feature: "system-ux-optimization"
journey: "list-filtering-and-sorting"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 8, Story 9)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: list-filtering-and-sorting

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Filtering and sorting change displayed data but do not mutate underlying state.
-->

## Overview

PM user applies assignee filters with sub-item penetration, multi-select status filters, and observes terminal status sorting on the item list page and progress page.

## Setup

- Multiple main items exist with various statuses (in-progress, closed, completed)
- Assignee A is responsible for some main items directly and some sub-items under other main items
- Assignee B is responsible for different items
- At least one main item is in terminal status (closed or completed)

## Happy Path

### Step 1: Filter by assignee with penetration

**User Action**: PM user selects assignee A in the assignee filter on the item list page

**Expected Result**: Results include main items where A is the direct assignee AND main items where A is the assignee of a sub-item; main items shown due to sub-item match display a "matched via sub-item" visual indicator, and only the matching sub-items are shown under those main items

### Step 2: View terminal status sorting

**User Action**: PM user views the item list page with no filters applied

**Expected Result**: Non-terminal main items appear first in their original order; terminal status main items (closed, completed) are sorted to the bottom of the list

### Step 3: Progress page default filter

**User Action**: PM user opens the overall progress page for the first time

**Expected Result**: "In progress" status checkbox is selected by default; only in-progress main items are displayed; terminal status items are filtered out

### Step 4: Clear all filters

**User Action**: PM user deselects all status checkboxes on the progress page

**Expected Result**: All items are displayed (equivalent to "show all"), including terminal status items

### Step 5: Empty state with active filters

**User Action**: PM user applies filters that match no items

**Expected Result**: Empty state message "No items match the criteria" is displayed; a "Clear filters" action button is available to reset all filters

## Edge Cases

### Step 1b: No filters selected shows all items

**Precondition**: No status or assignee filters are selected

**User Action**: PM user loads the item list page

**Expected Result**: All items are displayed without any filtering applied

### Step 2b: Filter by assignee with no sub-item matches

**Precondition**: Assignee C is only responsible for main items, never sub-items

**User Action**: PM user filters by assignee C

**Expected Result**: Only main items directly assigned to C are shown; no "matched via sub-item" indicators appear

### Step 3b: Progress page -- select terminal status

**Precondition**: Progress page is loaded

**User Action**: PM user selects the "Completed" status checkbox in addition to the default "In progress"

**Expected Result**: Both in-progress and completed items are displayed; terminal items still sort to the bottom

### Step 4b: Performance under large dataset

**Precondition**: 1000 main items and 5000 sub-items exist in the system

**User Action**: PM user applies an assignee filter with penetration

**Expected Result**: Filter response returns within 500ms; page renders without noticeable lag

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Terminal status main items (closed, completed) always sort to the bottom of item lists, regardless of active filters
- When no filters are selected, all items are displayed
- Empty filter results always show a clear empty state with a "Clear filters" action
- Sub-item match indicator is displayed whenever a main item is shown due to sub-item filter match
