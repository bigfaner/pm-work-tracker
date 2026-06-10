---
feature: "milestone-map"
journey: "read-only-milestone-access"
risk_level: "Low"
surface_types: ["web"]
surface_keys: ["frontend"]
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
generated: "2026-06-08"
---

# Journey: read-only-milestone-access

**Risk Level**: Low

<!-- Risk Classification Criteria:
  Low    = Workflow is read-only or purely observational
-->

## Overview

Team members and management with only milestone:read permission access the milestone map list and timeline views in a read-only capacity -- viewing progress, navigating nodes, and clicking through to item details without any mutation controls. <!-- fact: prd-spec Story 14 -->

## Setup

- At least one milestone map exists with milestones and associated MainItems
- User has milestone:read permission but no milestone:create, milestone:update, or milestone:delete permissions
- The milestone map list page (/milestones) is accessible

## Happy Path

### Step 1: View list page without create controls
<!-- surface: web -->

**User Action**: User navigates to /milestones list page. <!-- fact: prd-spec Story 14 — read-only access -->

**Expected Result**: Full card list and filter functionality are visible. The create button and dashed create card are NOT displayed. <!-- fact: prd-spec — RBAC controls visibility, milestone:create required for create buttons -->

### Step 2: View empty state without create button
<!-- surface: web -->

**Precondition**: Team has 0 milestone maps.

**User Action**: User views the list page.

**Expected Result**: Empty state shows "No milestone maps" message but does NOT show the create button. <!-- fact: prd-spec — create button requires milestone:create permission -->

### Step 3: Navigate to timeline view without edit controls
<!-- surface: web -->

**User Action**: User clicks a milestone map card to enter the timeline view. <!-- fact: prd-spec — timeline view for read-only -->

**Expected Result**: Timeline renders fully (breadcrumb, title, info card, filter bar, timeline). Edit button is hidden without milestone:update permission. Delete button is hidden without milestone:delete permission. StatusTransitionDropdown is hidden without milestone:update permission. The create milestone button is not displayed. <!-- fixed: permission-based visibility now enforced -->

### Step 4: Open milestone detail panel in read-only mode
<!-- surface: web -->

**User Action**: User clicks a milestone node on the timeline. <!-- fact: prd-spec — read-only panel -->

**Expected Result**: Detail panel opens showing all read-only information (name, description, status badge, plan date, progress, associated MI list). Edit button is hidden without milestone:update permission. Delete button is hidden without milestone:delete permission. StatusTransitionDropdown is hidden without milestone:update permission. Quick-add button and bind-existing button are hidden without milestone:update. MI rows do NOT show unbind controls without milestone:update. <!-- fixed: permission-based visibility now enforced -->

### Step 5: View tooltip and hover interactions
<!-- surface: web -->

**User Action**: User hovers over milestone nodes and description text.

**Expected Result**: All hover interactions work normally: node tooltips with item summaries, node highlighting, description text tooltips for overflow content.

### Step 6: Navigate to MI detail from timeline
<!-- surface: web -->

**User Action**: User clicks a MainItem entry on the timeline.

**Expected Result**: Route navigates to the main item detail page.

## Edge Cases

### Step 1b: List page API error with retry
<!-- surface: web -->

**Precondition**: Backend returns an error or times out. <!-- source: inferred — derived from Web surface `server-error` boundary outcome -->

**User Action**: User navigates to /milestones.

**Expected Result**: Page shows a retryable error message with a retry control. No blank page.

### Step 2b: Access denied without milestone:read
<!-- surface: web -->

**Precondition**: User does not have milestone:read permission. <!-- fact: prd-spec — milestone:read required -->

**User Action**: User attempts to access /milestones.

**Expected Result**: Page displays a forbidden access message. <!-- known gap: no page-level permission gate on milestone pages; this behavior is not yet implemented -->

### Step 3b: Timeline view read-only info displays correctly
<!-- surface: web -->

**Precondition**: User has milestone:read only.

**User Action**: User views the basic info card and milestone nodes.

**Expected Result**: All read-only information displays correctly: name, status, progress, description, associated MI list. All hover interactions work.

### Step 4b: Panel description tooltip in read-only mode
<!-- surface: web -->

**Precondition**: Milestone description overflows the panel area.

**User Action**: User hovers over the description text.

**Expected Result**: Tooltip displays the full description content, same as for users with edit permission.

### Step 5b: Timeline filters work in read-only mode
<!-- surface: web -->

**Precondition**: User is in read-only mode on the timeline view.

**User Action**: User uses search, status filter, zoom controls, and reset button.

**Expected Result**: All read-only interactive controls (search, filter, zoom, reset, refresh) function identically to users with edit permissions.

### Step 6b: MI navigation from read-only panel
<!-- surface: web -->

**Precondition**: Detail panel is open in read-only mode.

**User Action**: User clicks an MI's number or title in the panel list.

**Expected Result**: Route navigates to the main item detail page, same as for users with edit permission.

### Step E1: Session expired during read-only browsing (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has expired. <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: User attempts to navigate to a new page or interact with the timeline.

**Expected Result**: The user is redirected to the login page. After re-authenticating, the user must re-navigate to the desired page.

## Journey Invariants

<!-- Note on validation-error: This journey covers read-only access only. No forms or mutation controls are available to the user, so validation-error outcomes are not applicable. -->

- Users without milestone:create never see create buttons. <!-- fact: prd-spec — RBAC controls visibility -->
- Users without milestone:update never see edit controls, interactive status badges, or MI unbind controls.
- Users without milestone:delete never see delete controls.
- All read-only interactions (hover, tooltips, navigation, filters, zoom) work identically regardless of mutation permissions.
- Users without milestone:read receive a forbidden error and cannot access any milestone pages.
- API errors always display a retry option, never a blank page.
