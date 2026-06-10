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

PM navigates the milestone map list view and timeline view to observe progress, using filters, search, zoom controls, and tooltips to understand the state of milestones and associated items across stages. <!-- fact: prd-spec Story 9 -->

## Setup

- At least 3 milestone maps exist with various statuses and owners
- At least one milestone map has milestones with associated MainItems
- User has milestone:read permission
- The milestone map list page (/milestones) is accessible

## Happy Path

### Step 1: Load list view with cards
<!-- surface: web -->

**User Action**: PM navigates to /milestones. <!-- fact: prd-spec Story 9 — list page -->

**Expected Result**: Page loads showing all milestone map cards. Each card displays: name and status badge, milestone count, item count, owner, plan date range, overall progress, and milestone node thumbnails. A loading placeholder appears during data fetch. <!-- fact: prd-spec Story 9 — list page card content -->

### Step 2: Hover card for highlight
<!-- surface: web -->

**User Action**: PM hovers over a milestone map card.

**Expected Result**: Card provides visual highlight feedback.

### Step 3: Navigate to timeline view
<!-- surface: web -->

**User Action**: PM clicks a milestone map card. <!-- fact: prd-spec — timeline view -->

**Expected Result**: Route navigates to the map detail. Timeline view loads with breadcrumb, detail title area (name, status badge, edit/delete controls), basic info card (owner, plan dates, progress, description), filter bar (search, status filter, reset, refresh, create button, zoom controls), and horizontal timeline. <!-- fact: prd-spec — timeline view components -->

### Step 4: Interact with timeline nodes
<!-- surface: web -->

**User Action**: PM hovers over a milestone node on the timeline.

**Expected Result**: Tooltip shows milestone description text. Node provides visual highlight feedback.

### Step 5: Filter timeline by name search
<!-- surface: web -->

**User Action**: PM types a keyword in the search box.

**Expected Result**: Only milestone nodes whose names match the keyword are shown. Non-matching nodes are hidden.

### Step 6: Filter timeline by status tags
<!-- surface: web -->

**User Action**: PM uses the status filter to toggle one or more status tags.

**Expected Result**: Only milestone nodes matching the selected statuses are shown.

### Step 7: Reset filters
<!-- surface: web -->

**User Action**: PM clicks the reset button in the filter bar.

**Expected Result**: All filter conditions return to default. All nodes are shown.

### Step 8: Zoom timeline scale
<!-- surface: web -->

**User Action**: PM uses zoom controls to switch between compact/standard/relaxed spacing modes (紧凑/标准/宽松). <!-- note: zoom controls spacing modes, not time scales -->

**Expected Result**: Timeline spacing between nodes changes accordingly. Milestone node positions remain consistent across zoom levels.

### Step 9: Navigate back to list view
<!-- surface: web -->

**User Action**: PM clicks the breadcrumb "Milestone Maps" link.

**Expected Result**: Route navigates back to the list view.

### Step 10: Click MI item in timeline
<!-- surface: web -->

**User Action**: PM clicks a MainItem entry displayed below a milestone node on the timeline.

**Expected Result**: Route navigates to the main item detail page.

## Edge Cases

### Step 1b: List view empty state
<!-- surface: web -->

**Precondition**: The team has 0 milestone maps.

**User Action**: PM navigates to /milestones.

**Expected Result**: Empty state page shows a "No milestone maps" message and a create button (if PM has milestone:create permission).

### Step 1c: List view server error
<!-- surface: web -->

**Precondition**: Backend returns an error.

**User Action**: PM navigates to /milestones. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**Expected Result**: Page shows a retryable error message with a retry control. No blank page is shown.

### Step 1d: Filter list by status
<!-- surface: web -->

**Precondition**: List has milestone maps with various statuses.

**User Action**: PM uses status filter to select a specific status.

**Expected Result**: Only cards matching the selected status are shown.

### Step 1e: Filter list by owner
<!-- surface: web -->

**Precondition**: List has milestone maps with different owners.

**User Action**: PM uses owner filter to select a team member.

**Expected Result**: Only cards owned by that member are shown.

### Step 1f: Search list by name
<!-- surface: web -->

**Precondition**: List has multiple milestone maps.

**User Action**: PM types a keyword in the search box.

**Expected Result**: Only cards whose names contain the keyword are shown.

### Step 1g: Refresh list
<!-- surface: web -->

**Precondition**: List view is loaded.

**User Action**: PM clicks the refresh button.

**Expected Result**: List reloads, refresh button shows loading state during fetch.

### Step 3b: Timeline loading state
<!-- surface: web -->

**Precondition**: PM navigates to a milestone map timeline view and the data has not yet loaded.

**User Action**: PM views the timeline area while data is being fetched.

**Expected Result**: A loading placeholder is shown during data fetch. The placeholder is replaced with actual content once data arrives.

### Step 3c: Timeline empty state
<!-- surface: web -->

**Precondition**: The milestone map has 0 milestones.

**User Action**: Timeline view loads.

**Expected Result**: Shows empty state with a create button.

### Step 3d: Timeline data load failure
<!-- surface: web -->

**Precondition**: Timeline data fetch fails. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: PM enters timeline view.

**Expected Result**: Error message is shown with a retry control.

### Step 3e: Description tooltip on overflow
<!-- surface: web -->

**Precondition**: Detail title area description text overflows the display area.

**User Action**: PM hovers over the description text.

**Expected Result**: Description text is truncated at 3 lines via CSS line-clamp. No tooltip for overflow. <!-- gap: no tooltip wrapping for truncated descriptions -->

### Step 8b: Horizontal scroll for dense timeline
<!-- surface: web -->

**Precondition**: Milestone nodes are densely packed.

**User Action**: PM views the timeline.

**Expected Result**: Timeline area provides horizontal scrolling to view all nodes.

### Step 3f: Delete action hidden for non-deletable maps
<!-- surface: web -->

**Precondition**: Milestone map status is not in a deletable state ("executing", "completed", "cancelled"). <!-- fact: prd-spec — only planning, reviewed, ready are deletable -->

**User Action**: PM views the detail title area.

**Expected Result**: Delete action is not displayed.

### Step 3g: Access denied without milestone:read permission
<!-- surface: web -->

**Precondition**: User does not have milestone:read permission. <!-- fact: prd-spec — milestone:read required -->

**User Action**: User attempts to navigate to /milestones.

**Expected Result**: Page displays a forbidden access message.

### Step E1: Session expired during navigation (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM attempts to load a new page or interact with the timeline.

**Expected Result**: The user is redirected to the login page. After re-authenticating, the user must re-navigate to the desired page.

## Journey Invariants

<!-- Note on validation-error: This journey covers read-only visualization (search, filters, zoom). The only user inputs are search keywords and filter selections from predefined options, none of which have free-text validation constraints. The validation-error mandatory outcome is not applicable to this journey's scope. -->

- List view always shows a loading placeholder during data fetch, never a blank page.
- Timeline node positions are determined by plan completion date and do not change when zoom scale changes.
- All filter operations (search, status, owner) are client-side or debounced to avoid excessive API calls.
- Breadcrumb navigation provides a consistent way to return from timeline to list view.
- Error states always provide a retry option, never a blank page.
