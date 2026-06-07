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

Team members and management with only milestone:read permission access the milestone map list and timeline views in a read-only capacity -- viewing progress, navigating nodes, and clicking through to item details without any mutation controls.

## Setup

- At least one milestone map exists with milestones and associated MainItems
- User has milestone:read permission but no milestone:create, milestone:update, or milestone:delete permissions
- The milestone map list page (/milestones) is accessible

## Happy Path

### Step 1: View list page without create controls

**User Action**: User navigates to /milestones list page.

**Expected Result**: Full card list and filter functionality are visible. The "+ Create Milestone Map" button and dashed create card are NOT displayed.

### Step 2: View empty state without create button

**Precondition**: Team has 0 milestone maps.

**User Action**: User views the list page.

**Expected Result**: Empty state shows "No milestone maps" message but does NOT show the create button.

### Step 3: Navigate to timeline view without edit controls

**User Action**: User clicks a milestone map card to enter the timeline view.

**Expected Result**: Timeline renders fully (breadcrumb, title, info card, filter bar, timeline). The detail title area does NOT show edit or delete buttons. The "+ Create Milestone" button is not displayed or is shown as disabled.

### Step 4: Open milestone detail panel in read-only mode

**User Action**: User clicks a milestone node on the timeline.

**Expected Result**: Detail panel opens showing all read-only information (name, description, status Badge, plan date, progress, associated MI list). The edit button is NOT shown. The delete button is NOT shown. The "+ Add" button is NOT shown. The status Badge is NOT clickable. MI rows do NOT show X unbind buttons.

### Step 5: View tooltip and hover interactions

**User Action**: User hovers over milestone nodes and description text.

**Expected Result**: All hover interactions work normally: node tooltips ("X items, Y completed"), node highlighting, description text tooltips for overflow content.

### Step 6: Navigate to MI detail from timeline

**User Action**: User clicks a MainItem entry on the timeline.

**Expected Result**: Route navigates to /items/:mainItemId main item detail page.

## Edge Cases

### Step 1b: List page API error with retry

**Precondition**: Backend API returns 500 or times out.

**User Action**: User navigates to /milestones.

**Expected Result**: Page shows "Load failed, please retry" message with a retry button. No blank page.

### Step 2b: Access denied without milestone:read

**Precondition**: User does not have milestone:read permission.

**User Action**: User attempts to access /milestones.

**Expected Result**: Page displays a 403 forbidden message.

### Step 3b: Timeline view read-only info displays correctly

**Precondition**: User has milestone:read only.

**User Action**: User views the basic info card and milestone nodes.

**Expected Result**: All read-only information displays correctly: name, status, progress, description, associated MI list. All hover interactions (tooltips, node highlighting) work.

### Step 4b: Panel description tooltip in read-only mode

**Precondition**: Milestone description exceeds 6 lines in the detail panel.

**User Action**: User hovers over the description text.

**Expected Result**: Tooltip displays the full description content, same as for users with edit permission.

### Step 5b: Timeline filters work in read-only mode

**Precondition**: User is in read-only mode on the timeline view.

**User Action**: User uses search, StatusTagFilter, zoom controls, and reset button.

**Expected Result**: All read-only interactive controls (search, filter, zoom, reset, refresh) function identically to users with edit permissions.

### Step 6b: MI navigation from read-only panel

**Precondition**: Detail panel is open in read-only mode.

**User Action**: User clicks an MI's number or title in the panel list.

**Expected Result**: Route navigates to /items/:mainItemId, same as for users with edit permission.

## Journey Invariants

- Users without milestone:create never see create buttons (+ Create Milestone Map, dashed card, + Create Milestone).
- Users without milestone:update never see edit buttons, clickable status Badges, or MI unbind buttons.
- Users without milestone:delete never see delete buttons.
- All read-only interactions (hover, tooltips, navigation, filters, zoom) work identically regardless of mutation permissions.
- Users without milestone:read receive a 403 error and cannot access any milestone pages.
- API errors always display a retry option, never a blank page.
