---
feature: "milestone-map"
journey: "milestone-map-visualization"
risk_level: "Medium"
surface_types: ["web"]
surface_keys: ["frontend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: milestone-map-visualization

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Workflow involves multi-step interaction without irreversible side effects
-->

## Overview

PM navigates the milestone map list view and timeline view to observe progress, using filters, search, zoom controls, and tooltips to understand the state of milestones and associated items across stages.

## Setup

- At least 3 milestone maps exist with various statuses and owners
- At least one milestone map has milestones with associated MainItems
- User has milestone:read permission
- The milestone map list page (/milestones) is accessible

## Happy Path

### Step 1: Load list view with cards

**User Action**: PM navigates to /milestones.

**Expected Result**: Page loads showing all milestone map cards. Each card displays: name + status badge (row 1), milestone count + item count + owner (row 2), plan date range + overall progress bar + percentage (row 3), milestone node thumbnails (bottom). Loading skeleton appears during data fetch.

### Step 2: Hover card for highlight

**User Action**: PM hovers over a milestone map card.

**Expected Result**: Card border highlights with shadow effect.

### Step 3: Navigate to timeline view

**User Action**: PM clicks a milestone map card.

**Expected Result**: Route navigates to /milestones/:mapId. Timeline view loads with breadcrumb ("Milestone Maps" clickable to return), detail title area (name + status Badge + edit/delete buttons), basic info card (owner + plan dates + progress + description), filter bar (search + StatusTagFilter + reset + refresh + create button + zoom controls), and horizontal timeline.

### Step 4: Interact with timeline nodes

**User Action**: PM hovers over a milestone node on the timeline.

**Expected Result**: Tooltip appears showing "X items, Y completed". Node background highlights.

### Step 5: Filter timeline by name search

**User Action**: PM types a keyword in the search box (debounce 300ms).

**Expected Result**: Only milestone nodes whose names match the keyword are shown. Non-matching nodes are hidden.

### Step 6: Filter timeline by status tags

**User Action**: PM uses StatusTagFilter to toggle one or more status tags.

**Expected Result**: Only milestone nodes matching the selected statuses are shown.

### Step 7: Reset filters

**User Action**: PM clicks the reset button in the filter bar.

**Expected Result**: All filter conditions (search and status tags) return to default. All nodes are shown.

### Step 8: Zoom timeline scale

**User Action**: PM clicks zoom controls to switch between week/month/quarter view.

**Expected Result**: Timeline scale labels change accordingly (week = every 7 days, month = every 30 days, quarter = every 90 days). Milestone node positions remain the same. Transition animates in 200ms.

### Step 9: Navigate back to list view

**User Action**: PM clicks the breadcrumb "Milestone Maps" link.

**Expected Result**: Route navigates back to /milestones list view.

### Step 10: Click MI item in timeline

**User Action**: PM clicks a MainItem entry displayed below a milestone node on the timeline.

**Expected Result**: Route navigates to /items/:mainItemId main item detail page.

## Edge Cases

### Step 1b: List view empty state

**Precondition**: The team has 0 milestone maps.

**User Action**: PM navigates to /milestones.

**Expected Result**: Empty state page shows "No milestone maps" message and a create button (if PM has milestone:create permission).

### Step 1c: List view API error

**Precondition**: Backend API returns 500 error.

**User Action**: PM navigates to /milestones.

**Expected Result**: Page shows "Load failed, please retry" message with a retry button. No blank page is shown.

### Step 1d: Filter list by status

**Precondition**: List has milestone maps with various statuses.

**User Action**: PM uses status filter to select a specific status.

**Expected Result**: Only cards matching the selected status are shown.

### Step 1e: Filter list by owner

**Precondition**: List has milestone maps with different owners.

**User Action**: PM uses owner filter to select a team member.

**Expected Result**: Only cards owned by that member are shown.

### Step 1f: Search list by name

**Precondition**: List has multiple milestone maps.

**User Action**: PM types a keyword in the search box (debounce 300ms).

**Expected Result**: Only cards whose names contain the keyword are shown.

### Step 1g: Refresh list

**Precondition**: List view is loaded.

**User User Action**: PM clicks the refresh button.

**Expected Result**: List reloads, refresh button shows loading state during fetch.

### Step 3b: Timeline loading skeleton

**Precondition**: PM enters a milestone map timeline view.

**User Action**: Data is loading.

**Expected Result**: Skeleton screen is shown during loading.

### Step 3c: Timeline empty state

**Precondition**: The milestone map has 0 milestones.

**User Action**: Timeline view loads.

**Expected Result**: Shows "No milestones" empty state with a create button.

### Step 3d: Timeline data load failure

**Precondition**: Timeline data fetch fails.

**User Action**: PM enters timeline view.

**Expected Result**: Error message is shown with a retry button.

### Step 3e: Description tooltip on overflow

**Precondition**: Detail title area description text exceeds 3 lines.

**User Action**: PM hovers over the description text.

**Expected Result**: Tooltip displays the full description content.

### Step 8b: Horizontal scroll for dense timeline

**Precondition**: Milestone nodes are densely packed (adjacent nodes x-distance < 184px).

**User Action**: PM views the timeline.

**Expected Result**: Timeline area automatically shows a horizontal scrollbar. PM can scroll horizontally to view all nodes.

### Step 3f: Delete button hidden for non-planning maps

**Precondition**: Milestone map status is not "planning".

**User Action**: PM views the detail title area.

**Expected Result**: Delete button is not displayed.

## Journey Invariants

- List view always shows a loading skeleton during data fetch, never a blank page.
- Timeline node positions are determined by plan completion date and do not change when zoom scale changes.
- All filter operations (search, status, owner) are client-side or debounced (search at 300ms) to avoid excessive API calls.
- Breadcrumb navigation provides a consistent way to return from timeline to list view.
- Error states always provide a retry option, never a blank page.
