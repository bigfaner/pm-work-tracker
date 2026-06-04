---
feature: "system-ux-optimization"
journey: "sub-item-management"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 2, Story 5)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: sub-item-management

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Editing start time and viewing sorted lists involve data mutation but no destructive operations.
-->

## Overview

PM user edits sub-item start time in the edit dialog and views sub-items sorted by creation time in descending order on the main item detail page.

## Setup

- PM user is logged in
- A main item exists with multiple sub-items created at different times
- Sub-items have various start times assigned

## Happy Path

### Step 1: Open sub-item edit dialog

**User Action**: PM user opens the edit dialog for a sub-item

**Expected Result**: Edit dialog renders with all existing fields populated, including the "Start Time" field showing the current start time value

### Step 2: Edit start time and save

**User Action**: PM user changes the start time to a new date and clicks save

**Expected Result**: Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time

### Step 3: View sub-item list sorted by creation time

**User Action**: PM user navigates to a main item detail page that has 5 sub-items created on different dates

**Expected Result**: Sub-items are displayed in descending order by creation time -- the most recently created sub-item appears first in the list

## Edge Cases

### Step 1b: Edit start time to a date before parent main item's start time

**Precondition**: The main item has a start time of 2026-06-01

**User Action**: PM user sets the sub-item start time to 2026-05-15 (before parent) and saves

**Expected Result**: System behavior depends on backend validation rules; if disallowed, an inline error message is shown indicating the constraint; if allowed, the date is saved

### Step 2b: Edit start time to an invalid date

**Precondition**: Sub-item edit dialog is open

**User Action**: PM user manually enters an invalid date format in the start time field

**Expected Result**: Date picker or input validation prevents saving with an invalid date; an error message is shown if the user attempts to save

### Step 3b: Sub-item list with single sub-item

**Precondition**: A main item has exactly one sub-item

**User Action**: PM user views the main item detail page

**Expected Result**: The single sub-item is displayed; no sorting issues occur; empty state is not triggered

### Step 4b: Sub-item list with items created at the same time

**Precondition**: Multiple sub-items were created simultaneously (same timestamp)

**User Action**: PM user views the main item detail page

**Expected Result**: Sub-items with identical creation times are displayed in a stable, deterministic order (e.g., by sub-item code number as tiebreaker)

## Journey Invariants

- The sub-item edit dialog always includes a "Start Time" field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list (list is sorted by creation time, not start time)
