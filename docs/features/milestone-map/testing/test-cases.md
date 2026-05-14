---
feature: "milestone-map"
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
  - docs/features/milestone-map/prd/prd-ui-functions.md
generated: "2026-05-13"
---

# Test Cases: milestone-map

> **Note**: Milestone page (`/milestones`) elements use provisional `data-testid` selectors defined below. Once the milestone page UI is implemented, run `/gen-sitemap` to generate canonical sitemap IDs and update this document.
>
> **Provisional milestone page element map:**
> - `[data-testid='btn-create-map']` — "Create milestone map" button
> - `[data-testid='input-map-name']` — Map name input field
> - `[data-testid='input-map-description']` — Map description textarea
> - `[data-testid='btn-confirm']` — Confirm / Save button in modal/drawer
> - `[data-testid='btn-edit-map']` — Edit button on a milestone map card
> - `[data-testid='btn-delete']` — Delete button
> - `[data-testid='btn-confirm-delete']` — Confirm button in delete confirmation dialog
> - `[data-testid='badge-status']` — Status badge (clickable, opens dropdown)
> - `[data-testid='dropdown-status-options']` — Status transition dropdown menu
> - `[data-testid='filter-status']` — Status filter dropdown
> - `[data-testid='map-card']` — Milestone map card (container)
> - `[data-testid='map-card-title']` — Title text on a map card
> - `[data-testid='timeline-view']` — Timeline view container
> - `[data-testid='milestone-node']` — Milestone node in timeline
> - `[data-testid='zoom-week']`, `[data-testid='zoom-month']`, `[data-testid='zoom-quarter']` — Zoom control buttons
> - `[data-testid='axis-label']` — Time axis label element
> - `[data-testid='btn-create-milestone']` — "+ Create Milestone" button
> - `[data-testid='input-milestone-name']` — Milestone name input
> - `[data-testid='input-planned-date']` — Planned completion date picker
> - `[data-testid='detail-panel']` — Milestone detail panel
> - `[data-testid='btn-edit-milestone']` — Edit button in detail panel
> - `[data-testid='btn-save']` — Save button
> - `[data-testid='btn-unbind-mi']` — Unbind (x) button on MI row
> - `[data-testid='btn-quick-add-mi']` — "+ Add" button in detail panel
> - `[data-testid='form-quick-add-mi']` — Quick add MI form
> - `[data-testid='field-milestone']` — Milestone field in form
> - `[data-testid='empty-state']` — Empty state container
> - `[data-testid='error-state']` — Error state container
> - `[data-testid='btn-retry']` — Retry button
> - `[data-testid='toast-undo']` — Undo toast notification
> - `[data-testid='sidebar-link-milestones']` — Milestones navigation link in sidebar
>
> **Provisional existing-page milestone elements (not yet in sitemap.json):**
> - `[data-testid='btn-edit-item']` — "Edit" button on item detail page (opens edit dialog)
> - `[data-testid='select-milestone']` — Milestone selector dropdown in main item edit dialog
> - `[data-testid='filter-milestone']` — Milestone filter dropdown on /items list page
> - `[data-testid='columnheader-milestone']` — "Milestone" column header in /table view
> - `[data-testid='cell-milestone']` — Milestone cell in a table row (/table view)

## Summary

| Type | Count |
|------|-------|
| UI   | 45   |
| Integration | 7 |
| API  | 18  |
| CLI  | 0  |
| **Total** | **70** |

> **Note**: Integration tests are classified separately from UI. They verify that milestone components are correctly wired into existing pages and test cross-interface behavior (API action reflected in UI), using the same Playwright framework as UI tests.

---

## UI Test Cases

### Milestone Map Page (/milestones)

## TC-001: Create milestone map successfully
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-successfully
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-map']`, `[data-testid='input-map-name']`, `[data-testid='input-map-description']`, `[data-testid='btn-confirm']`, `[data-testid='map-card']`
- **Steps**:
  1. Click `[data-testid='btn-create-map']`
  2. Type "Q3 Release Plan" into `[data-testid='input-map-name']`
  3. (Optional) Type a description into `[data-testid='input-map-description']`
  4. Click `[data-testid='btn-confirm']`

- **Expected**: `[data-testid='map-card']` appears in the list with title "Q3 Release Plan" and status badge showing "planning"

- **Priority**: P0

## TC-002: Create milestone map with name at max length boundary
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-name-max-length
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-map']`, `[data-testid='input-map-name']`, `[data-testid='btn-confirm']`, `[data-testid='map-card']`
- **Steps**:
  1. Click `[data-testid='btn-create-map']`
  2. Type exactly 100 characters into `[data-testid='input-map-name']`
  3. Click `[data-testid='btn-confirm']`

- **Expected**: `[data-testid='map-card']` appears in the list with the 100-character title; form closes

- **Priority**: P1

## TC-003: Create milestone map with name exceeding max length
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-name-exceeds-max
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-map']`, `[data-testid='input-map-name']`, `[data-testid='btn-confirm']`
- **Steps**:
  1. Click `[data-testid='btn-create-map']`
  2. Type 101 characters into `[data-testid='input-map-name']`
  3. Click `[data-testid='btn-confirm']`

- **Expected**: Form displays inline error "Name cannot exceed 100 characters" below `[data-testid='input-map-name']`; form does not submit

- **Priority**: P1

## TC-004: Create milestone map with empty name
- **Source**: Story 1 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-empty-name
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-map']`, `[data-testid='input-map-name']`, `[data-testid='btn-confirm']`
- **Steps**:
  1. Click `[data-testid='btn-create-map']`
  2. Leave `[data-testid='input-map-name']` empty
  3. Click `[data-testid='btn-confirm']`

- **Expected**: Form displays inline error "Name cannot be empty" below `[data-testid='input-map-name']`; form does not submit

- **Priority**: P1

## TC-005: Edit milestone map info
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/edit-milestone-map-info
- **Pre-conditions**: A milestone map exists; user has `milestone:update` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='map-card']`, `[data-testid='btn-edit-map']`, `[data-testid='input-map-name']`, `[data-testid='btn-save']`
- **Steps**:
  1. Hover over the first `[data-testid='map-card']` and click `[data-testid='btn-edit-map']`
  2. Clear `[data-testid='input-map-name']` and type "Updated Map Name"
  3. Click `[data-testid='btn-save']`

- **Expected**: The first `[data-testid='map-card']` title updates to "Updated Map Name" within 1 second; no page reload required

- **Priority**: P0

## TC-006: Change milestone map status from planning to reviewed
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-map-status-planning-to-reviewed
- **Pre-conditions**: A milestone map is in "planning" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Click `[data-testid='badge-status']` on the "planning" map card
  2. Verify `[data-testid='dropdown-status-options']` contains exactly one option: "reviewed"
  3. Click the "reviewed" option

- **Expected**: `[data-testid='badge-status']` text changes from "planning" to "reviewed"; dropdown closes

- **Priority**: P0

## TC-007: Change milestone map status from in-progress to completed
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-map-status-inprogress-options
- **Pre-conditions**: A milestone map is in "in-progress" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Click `[data-testid='badge-status']` on the "in-progress" map card
  2. Verify `[data-testid='dropdown-status-options']` lists exactly two options: "pending-implementation" and "completed"

- **Expected**: Dropdown displays "pending-implementation" and "completed" as the only selectable options

- **Priority**: P1

## TC-008: Completed milestone map shows no transitions
- **Source**: Story 3 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/completed-milestone-map-no-transitions
- **Pre-conditions**: A milestone map is in "completed" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Click `[data-testid='badge-status']` on the "completed" map card
  2. Verify `[data-testid='dropdown-status-options']` is empty or shows text "No transitions available"

- **Expected**: Dropdown shows no selectable transition options; displays "No transitions available" message

- **Priority**: P1

## TC-009: Filter milestone maps by status
- **Source**: Story 3 / AC-4, Story 8 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/filter-milestone-maps-by-status
- **Pre-conditions**: Team has multiple milestone maps with different statuses; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='filter-status']`, `[data-testid='map-card']`, `[data-testid='badge-status']`
- **Steps**:
  1. Click `[data-testid='filter-status']` and select "in-progress"
  2. Verify each visible `[data-testid='map-card']` has `[data-testid='badge-status']` text equal to "in-progress"

- **Expected**: Only map cards with "in-progress" status are visible; all other cards are hidden

- **Priority**: P0

## TC-010: List view renders all milestone map cards
- **Source**: Story 8 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/list-view-renders-all-cards
- **Pre-conditions**: Team has 3+ milestone maps; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='map-card']`, `[data-testid='map-card-title']`, `[data-testid='badge-status']`
- **Steps**:
  1. Navigate to /milestones page
  2. Verify the count of `[data-testid='map-card']` elements matches the number of maps returned by the API
  3. Verify each `[data-testid='map-card']` contains `[data-testid='map-card-title']`, `[data-testid='badge-status']`, milestone count, item count, and progress percentage

- **Expected**: All milestone map cards render with name, status, milestone count, item count, and overall progress visible

- **Priority**: P0

## TC-011: Empty state when no milestone maps exist
- **Source**: Story 8 / AC-5
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/empty-state-no-maps
- **Pre-conditions**: Team has 0 milestone maps; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: `[data-testid='empty-state']`, `[data-testid='btn-create-map']`
- **Steps**:
  1. Navigate to /milestones page
  2. Verify `[data-testid='empty-state']` is visible with text "No milestone maps yet"
  3. Verify `[data-testid='btn-create-map']` is visible inside the empty state

- **Expected**: `[data-testid='empty-state']` displays "No milestone maps yet"; create button is present

- **Priority**: P1

## TC-012: Error state on API failure
- **Source**: Story 8 / AC-6
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/error-state-api-failure
- **Pre-conditions**: Team has milestone maps but backend API returns 500 error
- **Route**: /milestones
- **Element**: `[data-testid='error-state']`, `[data-testid='btn-retry']`
- **Steps**:
  1. Navigate to /milestones page while backend returns 500
  2. Verify `[data-testid='error-state']` is visible with text "Load failed, please retry"
  3. Verify `[data-testid='btn-retry']` is visible

- **Expected**: `[data-testid='error-state']` displays "Load failed, please retry"; `[data-testid='btn-retry']` is present; page does not render blank

- **Priority**: P1

### Timeline View

## TC-013: Enter timeline view from card click
- **Source**: Story 8 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/enter-timeline-from-card
- **Pre-conditions**: A milestone map with milestones exists; user is on /milestones list view
- **Route**: /milestones
- **Element**: `[data-testid='map-card']`, `[data-testid='timeline-view']`, `[data-testid='milestone-node']`
- **Steps**:
  1. Click the first `[data-testid='map-card']`
  2. Verify `[data-testid='timeline-view']` is visible
  3. Verify each `[data-testid='milestone-node']` displays name, planned completion date, status, and completion rate

- **Expected**: `[data-testid='timeline-view']` renders all milestone nodes; each node shows name, date, status, and completion rate

- **Priority**: P0

## TC-014: Timeline zoom controls
- **Source**: Story 8 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/timeline-zoom-controls
- **Pre-conditions**: User is in timeline view with milestones
- **Route**: /milestones
- **Element**: `[data-testid='zoom-week']`, `[data-testid='zoom-month']`, `[data-testid='zoom-quarter']`, `[data-testid='axis-label']`
- **Steps**:
  1. Click `[data-testid='zoom-week']`; verify `[data-testid='axis-label']` elements show week-level granularity (e.g., "W23", "W24")
  2. Click `[data-testid='zoom-month']`; verify `[data-testid='axis-label']` elements show month names (e.g., "Jun", "Jul")
  3. Click `[data-testid='zoom-quarter']`; verify `[data-testid='axis-label']` elements show quarter labels (e.g., "Q2", "Q3")

- **Expected**: Axis labels update to match the selected zoom level at each step

- **Priority**: P1

### Milestone Creation/Editing

## TC-015: Create milestone successfully
- **Source**: Story 4a / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-successfully
- **Pre-conditions**: User has `milestone:create` permission; user is in a milestone map timeline view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='input-planned-date']`, `[data-testid='btn-confirm']`, `[data-testid='milestone-node']`
- **Steps**:
  1. Click `[data-testid='btn-create-milestone']`
  2. Type "Phase 1" into `[data-testid='input-milestone-name']`
  3. Select a date in `[data-testid='input-planned-date']`
  4. Click `[data-testid='btn-confirm']`

- **Expected**: A new `[data-testid='milestone-node']` appears in the timeline with status "not_started" and completion rate 0

- **Priority**: P0

## TC-016: Create milestone with name at max length
- **Source**: Story 4a / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-name-max-length
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='btn-confirm']`, `[data-testid='milestone-node']`
- **Steps**:
  1. Click `[data-testid='btn-create-milestone']`
  2. Type exactly 100 characters into `[data-testid='input-milestone-name']`
  3. Click `[data-testid='btn-confirm']`

- **Expected**: A new `[data-testid='milestone-node']` appears with the 100-character name

- **Priority**: P1

## TC-017: Create milestone with name exceeding max length
- **Source**: Story 4a / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-name-exceeds-max
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='btn-confirm']`
- **Steps**:
  1. Click `[data-testid='btn-create-milestone']`
  2. Type 101 characters into `[data-testid='input-milestone-name']`
  3. Click `[data-testid='btn-confirm']`

- **Expected**: Form displays inline error "Name cannot exceed 100 characters" below `[data-testid='input-milestone-name']`; form does not submit

- **Priority**: P1

## TC-018: Create milestone with empty name
- **Source**: Story 4a / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-empty-name
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='btn-confirm']`
- **Steps**:
  1. Click `[data-testid='btn-create-milestone']`
  2. Leave `[data-testid='input-milestone-name']` empty
  3. Click `[data-testid='btn-confirm']`

- **Expected**: Form displays inline error "Name cannot be empty" below `[data-testid='input-milestone-name']`; form does not submit

- **Priority**: P1

## TC-019: Create milestone shows error on server failure
- **Source**: Story 4a / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-server-error
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view; backend returns 500
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='input-planned-date']`, `[data-testid='btn-confirm']`
- **Steps**:
  1. Click `[data-testid='btn-create-milestone']`
  2. Type "Phase 1" into `[data-testid='input-milestone-name']`; select a date in `[data-testid='input-planned-date']`
  3. Click `[data-testid='btn-confirm']` (backend returns 500)

- **Expected**: Page displays "Create failed, please retry"; `[data-testid='input-milestone-name']` retains "Phase 1"; date field retains selected value

- **Priority**: P1

## TC-020: Edit milestone info
- **Source**: Story 4b / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/edit-milestone-info
- **Pre-conditions**: A milestone exists; user has `milestone:update` permission; user is in timeline view
- **Route**: /milestones
- **Element**: `[data-testid='milestone-node']`, `[data-testid='detail-panel']`, `[data-testid='btn-edit-milestone']`, `[data-testid='input-milestone-name']`, `[data-testid='input-planned-date']`, `[data-testid='btn-save']`
- **Steps**:
  1. Click the first `[data-testid='milestone-node']` to open `[data-testid='detail-panel']`
  2. Click `[data-testid='btn-edit-milestone']`
  3. Clear `[data-testid='input-milestone-name']` and type "Updated Phase"
  4. Click `[data-testid='btn-save']`

- **Expected**: `[data-testid='detail-panel']` reflects the updated name; the corresponding `[data-testid='milestone-node']` label updates to "Updated Phase"

- **Priority**: P0

## TC-021: Concurrent edit conflict on milestone
- **Source**: Story 4b / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/concurrent-edit-conflict-milestone
- **Pre-conditions**: Two PMs are editing the same milestone simultaneously; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='detail-panel']`, `[data-testid='btn-save']`
- **Steps**:
  1. In browser tab A, open `[data-testid='detail-panel']` and edit the milestone name
  2. In browser tab B, open `[data-testid='detail-panel']` for the same milestone and save first
  3. In tab A, click `[data-testid='btn-save']`

- **Expected**: Tab A displays conflict message "Data has been modified by another user, please refresh and retry"; no silent overwrite occurs

- **Priority**: P2

## TC-022: Delete milestone
- **Source**: Story 4c / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/delete-milestone
- **Pre-conditions**: A milestone exists with associated MIs; user has `milestone:delete` permission
- **Route**: /milestones
- **Element**: `[data-testid='milestone-node']`, `[data-testid='detail-panel']`, `[data-testid='btn-delete']`, `[data-testid='btn-confirm-delete']`
- **Steps**:
  1. Click a `[data-testid='milestone-node']` to open `[data-testid='detail-panel']`
  2. Click `[data-testid='btn-delete']`
  3. Click `[data-testid='btn-confirm-delete']` in the confirmation dialog

- **Expected**: The `[data-testid='milestone-node']` is removed from the timeline; associated MIs have milestone_key cleared (verified via API GET)

- **Priority**: P0

### Milestone Status Changes

## TC-023: Change milestone status from not_started to in_progress
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-status-not-started-to-in-progress
- **Pre-conditions**: A milestone is in `not_started` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='milestone-node']`, `[data-testid='detail-panel']`, `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Click a `[data-testid='milestone-node']` to open `[data-testid='detail-panel']`
  2. Click `[data-testid='badge-status']`
  3. Click "in_progress" in `[data-testid='dropdown-status-options']`

- **Expected**: `[data-testid='badge-status']` text changes to "in_progress"; the `[data-testid='milestone-node']` visual style updates to in-progress appearance

- **Priority**: P0

## TC-024: Completed milestone shows only cancelled option
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/completed-milestone-only-cancelled
- **Pre-conditions**: A milestone is in `completed` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='detail-panel']`, `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Open `[data-testid='detail-panel']` for the completed milestone
  2. Click `[data-testid='badge-status']`
  3. Verify `[data-testid='dropdown-status-options']` contains exactly one option: "cancelled"

- **Expected**: Dropdown lists only "cancelled"; no "in_progress" rollback option is present

- **Priority**: P1

## TC-025: Cancelled milestone shows no transitions
- **Source**: Story 5 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/cancelled-milestone-no-transitions
- **Pre-conditions**: A milestone is in `cancelled` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='detail-panel']`, `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Open `[data-testid='detail-panel']` for the cancelled milestone
  2. Click `[data-testid='badge-status']`
  3. Verify `[data-testid='dropdown-status-options']` is empty or shows "No transitions available"

- **Expected**: Dropdown shows no selectable transition options; displays "No transitions available"

- **Priority**: P1

## TC-026: Cancel milestone auto-unbinds associated MIs
- **Source**: Story 5 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/cancel-milestone-auto-unbind
- **Pre-conditions**: A milestone is in `completed` or `in_progress` status with associated MIs; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='detail-panel']`, `[data-testid='badge-status']`, `[data-testid='dropdown-status-options']`
- **Steps**:
  1. Open `[data-testid='detail-panel']` for the milestone with associated MIs
  2. Click `[data-testid='badge-status']`
  3. Click "cancelled" in `[data-testid='dropdown-status-options']`

- **Expected**: `[data-testid='badge-status']` shows "cancelled"; the MI list in `[data-testid='detail-panel']` is empty (verified via API: all MIs' milestone_key is null)

- **Priority**: P0

### Milestone Detail Panel

## TC-027: Unbind MI from milestone detail panel
- **Source**: Story 7 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/unbind-mi-from-detail-panel
- **Pre-conditions**: Milestone detail panel is open with associated MIs; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: `[data-testid='detail-panel']`, `[data-testid='btn-unbind-mi']`, `[data-testid='toast-undo']`
- **Steps**:
  1. Click `[data-testid='btn-unbind-mi']` on the first MI row in `[data-testid='detail-panel']`

- **Expected**: The MI row is removed from the list; `[data-testid='toast-undo']` appears within 1 second

- **Priority**: P0

## TC-028: Quick add MI from detail panel
- **Source**: Story 7 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-from-detail-panel
- **Pre-conditions**: Milestone detail panel is open; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: `[data-testid='btn-quick-add-mi']`, `[data-testid='form-quick-add-mi']`, `[data-testid='field-milestone']`
- **Steps**:
  1. Click `[data-testid='btn-quick-add-mi']`
  2. Verify `[data-testid='form-quick-add-mi']` is visible
  3. Verify `[data-testid='field-milestone']` displays the current milestone name and has attribute `disabled`

- **Expected**: Quick add form opens; milestone field is pre-populated with the current milestone name and is non-editable

- **Priority**: P0

## TC-029: Quick add MI form creates and binds
- **Source**: Story 7 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-form-create-bind
- **Pre-conditions**: Quick add MI form is open; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: `[data-testid='form-quick-add-mi']`, `[data-testid='btn-confirm']`, `[data-testid='detail-panel']`
- **Steps**:
  1. Fill in title, assignee, start date, and expected end date in `[data-testid='form-quick-add-mi']`
  2. Click `[data-testid='btn-confirm']`

- **Expected**: New MI row appears in `[data-testid='detail-panel']` MI list; the MI's milestone_key equals the current milestone (verified via API)

- **Priority**: P0

## TC-030: Quick add MI form milestone field is disabled
- **Source**: Story 7 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-milestone-field-disabled
- **Pre-conditions**: Quick add MI form is open
- **Route**: /milestones
- **Element**: `[data-testid='field-milestone']`
- **Steps**:
  1. Verify `[data-testid='field-milestone']` has attribute `disabled` (or `aria-disabled="true"`)
  2. Verify `[data-testid='field-milestone']` value matches the current milestone name

- **Expected**: Field is disabled and displays the current milestone name; user cannot type or select a different value

- **Priority**: P1

### MI-Milestone Binding (via Items Page)

## TC-031: Bind MI to milestone from item edit page
- **Source**: Story 6 / AC-1
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/bind-mi-to-milestone
- **Pre-conditions**: A MainItem exists with milestone_key=null (created via POST /api/v1/teams/:teamId/main-items); at least one milestone exists in the team (created via POST /api/v1/teams/:teamId/milestones)
- **Route**: /items/:mainItemId
- **Element**: `[data-testid='btn-edit-item']`, `[data-testid='select-milestone']`, `[data-testid='btn-save']`
- **Steps**:
  1. Navigate to /items/:mainItemId (using the unassigned MI's bizKey)
  2. Click `[data-testid='btn-edit-item']` to open the edit dialog
  3. Click `[data-testid='select-milestone']` and select a milestone option
  4. Click `[data-testid='btn-save']`
- **Expected**: Dialog closes; MI detail page shows the selected milestone name; GET /api/v1/teams/:teamId/main-items/:mainItemId returns milestone_key equal to the selected milestone's bizKey
- **Priority**: P0

## TC-032: Unbind MI from milestone via item edit page
- **Source**: Story 6 / AC-2
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/unbind-mi-from-milestone
- **Pre-conditions**: A MainItem exists with milestone_key set to an active milestone (created via API); user has `milestone:update` permission
- **Route**: /items/:mainItemId
- **Element**: `[data-testid='btn-edit-item']`, `[data-testid='select-milestone']`, `[data-testid='btn-save']`
- **Steps**:
  1. Navigate to /items/:mainItemId (using the bound MI's bizKey)
  2. Click `[data-testid='btn-edit-item']` to open the edit dialog
  3. Click `[data-testid='select-milestone']` and select the "unassigned" option
  4. Click `[data-testid='btn-save']`
- **Expected**: Dialog closes; MI detail page shows no milestone; GET /api/v1/teams/:teamId/main-items/:mainItemId returns milestone_key=null; the original milestone's completion rate recalculates
- **Priority**: P0

## TC-033: Filter items by milestone on items list page
- **Source**: Story 6 / AC-3
- **Type**: UI
- **Target**: ui/items
- **Test ID**: ui/items/filter-items-by-milestone
- **Pre-conditions**: MIs are assigned to at least 2 different milestones (created via API); user is on /items page
- **Route**: /items
- **Element**: `[data-testid='filter-milestone']`
- **Steps**:
  1. Click `[data-testid='filter-milestone']` and select a specific milestone option
  2. Verify each visible item card shows the selected milestone name in its milestone field
- **Expected**: List only shows MIs belonging to the selected milestone; count of visible items matches the API count for that milestone
- **Priority**: P0

## TC-034: No milestones available shows only unassigned option
- **Source**: Story 6 / AC-4
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/no-milestones-only-unassigned
- **Pre-conditions**: Team has no milestones created (GET /api/v1/teams/:teamId/milestones returns empty list); user has permission to edit a MainItem
- **Route**: /items/:mainItemId
- **Element**: `[data-testid='btn-edit-item']`, `[data-testid='select-milestone']`
- **Steps**:
  1. Navigate to /items/:mainItemId
  2. Click `[data-testid='btn-edit-item']` to open the edit dialog
  3. Click `[data-testid='select-milestone']`
  4. Verify the dropdown contains exactly one option: "unassigned"
- **Expected**: `[data-testid='select-milestone']` dropdown only shows "unassigned" option; no other selectable options
- **Priority**: P1

### Permission-Based Views

## TC-035: Read-only user sees milestones page without action buttons
- **Source**: Story 9 / AC-1, Story 9 / AC-2, Story 11 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/read-only-user-view
- **Pre-conditions**: User has `milestone:read` permission only (no create/update/delete)
- **Route**: /milestones
- **Element**: `[data-testid='btn-create-map']`, `[data-testid='btn-edit-map']`, `[data-testid='btn-delete']`, `[data-testid='map-card']`, `[data-testid='timeline-view']`
- **Steps**:
  1. Navigate to /milestones page
  2. Verify `[data-testid='map-card']` elements are visible (list view renders)
  3. Verify `[data-testid='btn-create-map']` is disabled with tooltip "No permission"
  4. Verify `[data-testid='btn-edit-map']` and `[data-testid='btn-delete']` are not present on any card
- **Expected**: List and timeline views render in read-only mode; no create/edit/delete buttons are functional or visible
- **Priority**: P0

## TC-036: Read-only user with no maps sees empty state without create button
- **Source**: Story 9 / AC-3, Story 11 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/read-only-empty-no-create
- **Pre-conditions**: User has `milestone:read` permission only; team has 0 milestone maps
- **Route**: /milestones
- **Element**: `[data-testid='empty-state']`, `[data-testid='btn-create-map']`
- **Steps**:
  1. Navigate to /milestones page
  2. Verify `[data-testid='empty-state']` is visible with text "No milestone maps yet"
  3. Verify `[data-testid='btn-create-map']` is not present in the DOM
- **Expected**: Empty state displays "No milestone maps yet"; no create button is rendered
- **Priority**: P1

## TC-037: No milestone read permission shows 403
- **Source**: Story 11 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/no-read-permission-403
- **Pre-conditions**: User does NOT have `milestone:read` permission
- **Route**: /milestones
- **Element**: `[role='alert']` (403 permission denied message)
- **Steps**:
  1. Navigate to /milestones page
- **Expected**: Page displays a 403 permission denied message within `[role='alert']`
- **Priority**: P0

## TC-048: Management user sees error state with retry on API failure
- **Source**: Story 11 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/management-user-api-error-retry
- **Pre-conditions**: User has `milestone:read` permission only (management persona); backend API returns 500 or times out
- **Route**: /milestones
- **Element**: `[data-testid='error-state']`, `[data-testid='btn-retry']`
- **Steps**:
  1. Navigate to /milestones page while backend returns 500
  2. Verify `[data-testid='error-state']` is visible with text "Load failed, please retry"
  3. Verify `[data-testid='btn-retry']` is visible
- **Expected**: `[data-testid='error-state']` displays "Load failed, please retry"; `[data-testid='btn-retry']` is present; page does not render blank
- **Priority**: P1

### Table View Milestone Column

## TC-038: Table view shows milestone column
- **Source**: Story 10 / AC-1
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/shows-milestone-column
- **Pre-conditions**: Table view is loaded with at least one MI; at least one MI is bound to a milestone (via API)
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Navigate to /table page
  2. Verify `[data-testid='columnheader-milestone']` is visible in the header row
  3. Verify `[data-testid='columnheader-milestone']` is present within the `<thead>` row and its DOM index equals the expected column position (e.g., `headerRow.locator('th').nth(2)`)
  4. Verify each `[data-testid='cell-milestone']` displays the milestone name or "-" for unassigned MIs
- **Expected**: `[data-testid='columnheader-milestone']` is present at the expected DOM index within `<thead>`; cells show milestone names or "-" for unassigned
- **Priority**: P0

## TC-039: Table view filter by milestone column
- **Source**: Story 10 / AC-2
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/filter-by-milestone-column
- **Pre-conditions**: Table view has MIs assigned to at least 2 different milestones
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Click the filter icon on `[data-testid='columnheader-milestone']`
  2. Select a specific milestone name from the filter dropdown
  3. Verify each visible `[data-testid='cell-milestone']` displays the selected milestone name
- **Expected**: Only rows belonging to the selected milestone are visible; unselected milestone rows are hidden
- **Priority**: P0

## TC-040: Table view milestone column error fallback
- **Source**: Story 10 / AC-3
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/milestone-column-error-fallback
- **Pre-conditions**: Table view is loaded; milestone API endpoint (GET /api/v1/teams/:teamId/milestones) returns 500
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Navigate to /table page with milestone API returning 500
  2. Verify each `[data-testid='cell-milestone']` displays "--"
  3. Verify other columns (E-089 title, E-090 priority, etc.) render normally
- **Expected**: `[data-testid='cell-milestone']` cells display "--"; other columns render normally; table loading is not blocked
- **Priority**: P2

## TC-041: Table view shows "--" for soft-deleted milestone MI
- **Source**: Story 10 / AC-4
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/soft-deleted-milestone-shows-dash
- **Pre-conditions**: An MI exists whose milestone_key references a soft-deleted milestone (milestone deleted via DELETE /api/v1/teams/:teamId/milestones/:id, but MI still has the stale milestone_key)
- **Route**: /table
- **Element**: `[data-testid='cell-milestone']`
- **Steps**:
  1. Navigate to /table page
  2. Locate the row for the MI with the soft-deleted milestone
  3. Verify its `[data-testid='cell-milestone']` displays "--"
- **Expected**: `[data-testid='cell-milestone']` shows "--" for the affected MI row
- **Priority**: P1

## TC-042: Table view sort by milestone column descending
- **Source**: Story 10 / AC-5
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/sort-milestone-column-desc
- **Pre-conditions**: Table view has MIs assigned to at least 2 milestones with different names, plus at least 1 unassigned MI
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Click `[data-testid='columnheader-milestone']` once to sort ascending, then again to sort descending
  2. Verify assigned MIs' `[data-testid='cell-milestone']` values are ordered by milestone name in descending alphabetical order
  3. Verify unassigned MIs (showing "-") appear after all assigned ones
- **Expected**: Rows with milestone names are sorted Z-to-A; unassigned rows appear at the bottom
- **Priority**: P1

## TC-043: Table view default milestone sort ascending
- **Source**: Story 10 / AC-6
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/default-milestone-sort-ascending
- **Pre-conditions**: Table view has MIs assigned to at least 2 milestones with different names, plus at least 1 unassigned MI; no sort setting has been changed
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Navigate to /table page (fresh load, default sort)
  2. Verify assigned MIs' `[data-testid='cell-milestone']` values are ordered by milestone name in ascending alphabetical order
  3. Verify unassigned MIs (showing "-") appear after all assigned ones
- **Expected**: Rows with milestone names are sorted A-to-Z; unassigned rows appear at the bottom
- **Priority**: P1

### Navigation

## TC-047: Milestones page navigation link exists
- **Source**: PRD UI Function "UF-1" Navigation Architecture
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/navigation-link
- **Pre-conditions**: Milestones page route is registered in App.tsx; navigation is updated
- **Route**: /milestones
- **Element**: `[data-testid='sidebar-link-milestones']`
- **Steps**:
  1. Verify `[data-testid='sidebar-link-milestones']` is visible in the sidebar
  2. Click `[data-testid='sidebar-link-milestones']`
  3. Verify browser URL is `/milestones`
- **Expected**: Navigation link exists in sidebar and navigates to /milestones page
- **Priority**: P0

### Integration Tests (Existing Page Modifications)

## TC-044: Integration -- Milestone filter visible on Items page
- **Source**: Story 6 / AC-3, PRD UI Function "UF-4"
- **Type**: Integration
- **Target**: ui/items
- **Test ID**: ui/items/integration-milestone-filter
- **Pre-conditions**: Milestone filter component build complete, integration task complete
- **Route**: /items
- **Element**: `[data-testid='filter-milestone']`
- **Steps**:
  1. Navigate to /items
  2. Verify `[data-testid='filter-milestone']` is visible within the filter bar container (`[data-testid='filter-bar']` or equivalent parent)
  3. Verify `[data-testid='filter-milestone']` renders with "all" as default value
- **Expected**: `[data-testid='filter-milestone']` is a child of the filter bar container and shows default "all" state
- **Priority**: P0

## TC-045: Integration -- Milestone selector visible in Item Edit dialog
- **Source**: Story 6 / AC-1, PRD UI Function "UF-5"
- **Type**: Integration
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/integration-milestone-selector
- **Pre-conditions**: Milestone selector component build complete, integration task complete
- **Route**: /items/:mainItemId
- **Element**: `[data-testid='btn-edit-item']`, `[data-testid='select-milestone']`
- **Steps**:
  1. Navigate to /items/:mainItemId
  2. Click `[data-testid='btn-edit-item']` to open the edit dialog
  3. Verify `[data-testid='select-milestone']` is visible within the edit dialog form container
  4. Verify `[data-testid='select-milestone']` displays the current milestone name or "unassigned" value
- **Expected**: `[data-testid='select-milestone']` is present inside the edit dialog and displays current assignment correctly
- **Priority**: P0

## TC-046: Integration -- Milestone column visible in Table view
- **Source**: Story 10 / AC-1, PRD UI Function "UF-6"
- **Type**: Integration
- **Target**: ui/table
- **Test ID**: ui/table/integration-milestone-column
- **Pre-conditions**: Milestone column component build complete, integration task complete
- **Route**: /table
- **Element**: `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`
- **Steps**:
  1. Navigate to /table
  2. Verify `[data-testid='columnheader-milestone']` is visible within the `<thead>` row at the expected DOM index
  3. Verify each `[data-testid='cell-milestone']` displays milestone names or "-" for unassigned
- **Expected**: `[data-testid='columnheader-milestone']` is present at the expected DOM index within `<thead>`; cells display data correctly
- **Priority**: P0


### Cross-Interface Integration Tests

## TC-049: API-created milestone appears in items page filter
- **Source**: Story 6 / AC-3, Story 6 / AC-1
- **Type**: Integration
- **Target**: ui/items
- **Test ID**: integration/api-create-milestone-appears-in-filter
- **Pre-conditions**: User has `milestone:create` permission; team has an existing milestone map
- **Route**: /items
- **Element**: `[data-testid='filter-milestone']`
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones with `{name: "Integration Test MS", planned_completion_date: "2026-08-01"}`
  2. Navigate to /items page
  3. Click `[data-testid='filter-milestone']`
  4. Verify "Integration Test MS" appears as an option in the dropdown
- **Expected**: The milestone created via API is immediately selectable in the UI filter without page refresh
- **Priority**: P0

## TC-050: API-bound MI shows milestone name in table view
- **Source**: Story 10 / AC-1, Story 6 / AC-1
- **Type**: Integration
- **Target**: ui/table
- **Test ID**: integration/api-bind-mi-reflects-in-table
- **Pre-conditions**: A milestone and an unassigned MI exist (created via API)
- **Route**: /table
- **Element**: `[data-testid='cell-milestone']`
- **Steps**:
  1. PUT /api/v1/teams/:teamId/main-items/:mainItemId with `{milestone_key: "<milestoneBizKey>"}` to bind the MI
  2. Navigate to /table page
  3. Locate the row for the bound MI
  4. Verify its `[data-testid='cell-milestone']` displays the milestone name (not "-")
- **Expected**: Table view reflects the API-performed binding without manual refresh; the cell shows the correct milestone name
- **Priority**: P0

## TC-051: API-deleted milestone shows "--" for affected MIs in table view
- **Source**: Story 10 / AC-4, Story 4c / AC-1
- **Type**: Integration
- **Target**: ui/table
- **Test ID**: integration/api-delete-milestone-reflects-in-table
- **Pre-conditions**: A milestone with bound MIs exists
- **Route**: /table
- **Element**: `[data-testid='cell-milestone']`
- **Steps**:
  1. DELETE /api/v1/teams/:teamId/milestones/:milestoneId (soft-delete)
  2. Navigate to /table page
  3. Locate the rows for previously bound MIs
  4. Verify their `[data-testid='cell-milestone']` displays "--"
- **Expected**: All MIs that were bound to the deleted milestone show "--" in the table view
- **Priority**: P0

## TC-052: API-cancelled milestone unbinds MIs reflected in table view
- **Source**: Story 5 / AC-4, Story 10 / AC-1
- **Type**: Integration
- **Target**: ui/table
- **Test ID**: integration/api-cancel-milestone-unbind-reflects-in-table
- **Pre-conditions**: A milestone in `in_progress` status with bound MIs exists
- **Route**: /table
- **Element**: `[data-testid='cell-milestone']`
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestones/:milestoneId/status with `{status: "cancelled"}`
  2. Navigate to /table page
  3. Locate the rows for previously bound MIs
  4. Verify their `[data-testid='cell-milestone']` displays "--"
- **Expected**: Cancelling the milestone via API causes all bound MIs to show "--" in the table view
- **Priority**: P0

---

## API Test Cases

### MilestoneMap CRUD

## TC-053: API Create milestone map
- **Source**: Story 1 / AC-1, PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/create-milestone-map
- **Pre-conditions**: User authenticated with `milestone:create` permission; valid team context
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps with body `{name: "Test Map", description: "desc"}`
- **Expected**: Returns 200 with created milestone map; status is "planning"; name and description match input
- **Priority**: P0

## TC-054: API Create milestone map validation errors
- **Source**: Story 1 / AC-2, Story 1 / AC-3
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/create-validation-errors
- **Pre-conditions**: User authenticated with `milestone:create` permission
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps with body `{name: ""}` (empty name)
  2. POST /api/v1/teams/:teamId/milestone-maps with body `{name: "x".repeat(101)}` (name too long)
- **Expected**: Empty name returns 400 with error message; name exceeding 100 chars returns 400 with error message
- **Priority**: P0

## TC-055: API List milestone maps
- **Source**: Story 8 / AC-1, PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/list-milestone-maps
- **Pre-conditions**: Team has 3+ milestone maps; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps
- **Expected**: Returns 200 with list of milestone maps; each item includes name, status, milestone count, item count, overall progress
- **Priority**: P0

## TC-056: API Get milestone map by ID
- **Source**: PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/get-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps/:mapId
- **Expected**: Returns 200 with milestone map details including name, description, status, computed fields
- **Priority**: P0

## TC-057: API Update milestone map
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/update-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestone-maps/:mapId with body `{name: "Updated Name", description: "Updated desc"}`
- **Expected**: Returns 200 with updated milestone map; name and description reflect changes
- **Priority**: P0

## TC-058: API Delete milestone map
- **Source**: PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/delete-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:delete` permission
- **Steps**:
  1. DELETE /api/v1/teams/:teamId/milestone-maps/:mapId
- **Expected**: Returns 200; milestone map is soft-deleted; subsequent GET returns 404
- **Priority**: P0

## TC-059: API Change milestone map status
- **Source**: Story 3 / AC-1, Story 3 / AC-2, Story 3 / AC-3
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/change-status
- **Pre-conditions**: Milestone maps exist in various states; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status with `{status: "reviewed"}` on a "planning" map
  2. PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status with invalid transition
  3. PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status on a "completed" map
- **Expected**: Valid transition returns 200; invalid transition returns 400 with error; completed map returns no available transitions
- **Priority**: P0

## TC-060: API Get available transitions for milestone map
- **Source**: Story 3 / AC-1, Story 3 / AC-2, Story 3 / AC-3
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/available-transitions
- **Pre-conditions**: Milestone maps exist in various states; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions for "planning" status
  2. GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions for "completed" status
- **Expected**: Planning map returns ["reviewed"]; completed map returns empty array
- **Priority**: P1

### Milestone CRUD

## TC-061: API Create milestone
- **Source**: Story 4a / AC-1
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/create-milestone
- **Pre-conditions**: User has `milestone:create` permission; a milestone map exists
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones with body `{name: "Phase 1", planned_completion_date: "2026-06-30"}`
- **Expected**: Returns 200 with created milestone; status is "not_started"; completion rate is 0
- **Priority**: P0

## TC-062: API Create milestone validation errors
- **Source**: Story 4a / AC-2, Story 4a / AC-3
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/create-validation-errors
- **Pre-conditions**: User has `milestone:create` permission
- **Steps**:
  1. POST with empty name
  2. POST with name exceeding 100 chars
- **Expected**: Returns 400 for both cases with appropriate validation error messages
- **Priority**: P0

## TC-063: API List milestones by map
- **Source**: PRD Spec Related Changes #2
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/list-by-map
- **Pre-conditions**: A milestone map has milestones; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
- **Expected**: Returns 200 with list of milestones under the specified map; each includes computed completion rate
- **Priority**: P0

## TC-064: API List milestones by team
- **Source**: PRD Spec Related Changes #2, Story 6 / AC-3
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/list-by-team
- **Pre-conditions**: Team has milestones across multiple maps; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestones
- **Expected**: Returns 200 with all non-cancelled milestones in the team
- **Priority**: P0

## TC-065: API Get milestone by ID
- **Source**: PRD Spec Related Changes #2
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/get-milestone
- **Pre-conditions**: A milestone exists; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestones/:milestoneId
- **Expected**: Returns 200 with milestone details including computed completion rate and associated MI count
- **Priority**: P0

## TC-066: API Update milestone
- **Source**: Story 4b / AC-1
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/update-milestone
- **Pre-conditions**: A milestone exists; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestones/:milestoneId with body `{name: "Updated Phase", planned_completion_date: "2026-07-31"}`
- **Expected**: Returns 200 with updated milestone
- **Priority**: P0

## TC-067: API Delete milestone unbinds associated MIs
- **Source**: Story 4c / AC-1
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/delete-milestone-unbind
- **Pre-conditions**: A milestone with associated MIs exists; user has `milestone:delete` permission
- **Steps**:
  1. DELETE /api/v1/teams/:teamId/milestones/:milestoneId
  2. Verify associated MIs' milestone_key is null
- **Expected**: Returns 200; milestone is soft-deleted; all associated MIs have milestone_key cleared in same transaction
- **Priority**: P0

## TC-068: API Change milestone status
- **Source**: Story 5 / AC-1, Story 5 / AC-2, Story 5 / AC-3, Story 5 / AC-4
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/change-status
- **Pre-conditions**: Milestones exist in various states; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestones/:milestoneId/status with `{status: "in_progress"}` on not_started
  2. PUT with invalid transition (cancelled -> any)
  3. PUT with `{status: "cancelled"}` on in_progress milestone with associated MIs
- **Expected**: Valid transition returns 200; invalid transition returns 400; cancellation auto-unbinds MIs
- **Priority**: P0

## TC-069: API Get available transitions for milestone
- **Source**: Story 5 / AC-2, Story 5 / AC-3
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/available-transitions
- **Pre-conditions**: Milestones in various states; user has `milestone:read` permission
- **Steps**:
  1. GET available transitions for not_started milestone
  2. GET available transitions for completed milestone
  3. GET available transitions for cancelled milestone
- **Expected**: not_started returns ["in_progress", "cancelled"]; completed returns ["cancelled"]; cancelled returns empty array
- **Priority**: P1

### Permission Tests

## TC-070: API milestone operations without permission return 403
- **Source**: PRD Spec Security Requirements
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/permission-denied
- **Pre-conditions**: User lacks specific milestone permission
- **Steps**:
  1. POST create without `milestone:create` permission
  2. PUT update without `milestone:update` permission
  3. DELETE without `milestone:delete` permission
  4. GET without `milestone:read` permission
- **Expected**: Each operation returns 403 Forbidden
- **Priority**: P0

---

## CLI Test Cases

_No CLI test cases — project exposes UI and API interfaces only (no user-facing CLI binary)._

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | UI | ui/milestones | P0 |
| TC-002 | Story 1 / AC-2 | UI | ui/milestones | P1 |
| TC-003 | Story 1 / AC-2 | UI | ui/milestones | P1 |
| TC-004 | Story 1 / AC-3 | UI | ui/milestones | P1 |
| TC-005 | Story 2 / AC-1 | UI | ui/milestones | P0 |
| TC-006 | Story 3 / AC-1 | UI | ui/milestones | P0 |
| TC-007 | Story 3 / AC-2 | UI | ui/milestones | P1 |
| TC-008 | Story 3 / AC-3 | UI | ui/milestones | P1 |
| TC-009 | Story 3 / AC-4, Story 8 / AC-2 | UI | ui/milestones | P0 |
| TC-010 | Story 8 / AC-1 | UI | ui/milestones | P0 |
| TC-011 | Story 8 / AC-5 | UI | ui/milestones | P1 |
| TC-012 | Story 8 / AC-6 | UI | ui/milestones | P1 |
| TC-013 | Story 8 / AC-3 | UI | ui/milestones | P0 |
| TC-014 | Story 8 / AC-4 | UI | ui/milestones | P1 |
| TC-015 | Story 4a / AC-1 | UI | ui/milestones | P0 |
| TC-016 | Story 4a / AC-2 | UI | ui/milestones | P1 |
| TC-017 | Story 4a / AC-2 | UI | ui/milestones | P1 |
| TC-018 | Story 4a / AC-3 | UI | ui/milestones | P1 |
| TC-019 | Story 4a / AC-4 | UI | ui/milestones | P1 |
| TC-020 | Story 4b / AC-1 | UI | ui/milestones | P0 |
| TC-021 | Story 4b / AC-2 | UI | ui/milestones | P2 |
| TC-022 | Story 4c / AC-1 | UI | ui/milestones | P0 |
| TC-023 | Story 5 / AC-1 | UI | ui/milestones | P0 |
| TC-024 | Story 5 / AC-2 | UI | ui/milestones | P1 |
| TC-025 | Story 5 / AC-3 | UI | ui/milestones | P1 |
| TC-026 | Story 5 / AC-4 | UI | ui/milestones | P0 |
| TC-027 | Story 7 / AC-1 | UI | ui/milestones | P0 |
| TC-028 | Story 7 / AC-2 | UI | ui/milestones | P0 |
| TC-029 | Story 7 / AC-3 | UI | ui/milestones | P0 |
| TC-030 | Story 7 / AC-4 | UI | ui/milestones | P1 |
| TC-031 | Story 6 / AC-1 | UI | ui/items-detail | P0 |
| TC-032 | Story 6 / AC-2 | UI | ui/items-detail | P0 |
| TC-033 | Story 6 / AC-3 | UI | ui/items | P0 |
| TC-034 | Story 6 / AC-4 | UI | ui/items-detail | P1 |
| TC-035 | Story 9 / AC-1, AC-2, Story 11 / AC-1 | UI | ui/milestones | P0 |
| TC-036 | Story 9 / AC-3, Story 11 / AC-3 | UI | ui/milestones | P1 |
| TC-037 | Story 11 / AC-2 | UI | ui/milestones | P0 |
| TC-038 | Story 10 / AC-1 | UI | ui/table | P0 |
| TC-039 | Story 10 / AC-2 | UI | ui/table | P0 |
| TC-040 | Story 10 / AC-3 | UI | ui/table | P2 |
| TC-041 | Story 10 / AC-4 | UI | ui/table | P1 |
| TC-042 | Story 10 / AC-5 | UI | ui/table | P1 |
| TC-043 | Story 10 / AC-6 | UI | ui/table | P1 |
| TC-044 | Story 6 / AC-3, UF-4 | Integration | ui/items | P0 |
| TC-045 | Story 6 / AC-1, UF-5 | Integration | ui/items-detail | P0 |
| TC-046 | Story 10 / AC-1, UF-6 | Integration | ui/table | P0 |
| TC-047 | UF-1 Navigation Architecture | UI | ui/milestones | P0 |
| TC-048 | Story 11 / AC-4 | UI | ui/milestones | P1 |
| TC-049 | Story 6 / AC-3, Story 6 / AC-1 | Integration | ui/items | P0 |
| TC-050 | Story 10 / AC-1, Story 6 / AC-1 | Integration | ui/table | P0 |
| TC-051 | Story 10 / AC-4, Story 4c / AC-1 | Integration | ui/table | P0 |
| TC-052 | Story 5 / AC-4, Story 10 / AC-1 | Integration | ui/table | P0 |
| TC-053 | Story 1 / AC-1, Related Changes #1 | API | api/milestone-maps | P0 |
| TC-054 | Story 1 / AC-2, AC-3 | API | api/milestone-maps | P0 |
| TC-055 | Story 8 / AC-1, Related Changes #1 | API | api/milestone-maps | P0 |
| TC-056 | Related Changes #1 | API | api/milestone-maps | P0 |
| TC-057 | Story 2 / AC-1 | API | api/milestone-maps | P0 |
| TC-058 | Related Changes #1 | API | api/milestone-maps | P0 |
| TC-059 | Story 3 / AC-1, AC-2, AC-3 | API | api/milestone-maps | P0 |
| TC-060 | Story 3 / AC-1, AC-2, AC-3 | API | api/milestone-maps | P1 |
| TC-061 | Story 4a / AC-1 | API | api/milestones | P0 |
| TC-062 | Story 4a / AC-2, AC-3 | API | api/milestones | P0 |
| TC-063 | Related Changes #2 | API | api/milestones | P0 |
| TC-064 | Related Changes #2, Story 6 / AC-3 | API | api/milestones | P0 |
| TC-065 | Related Changes #2 | API | api/milestones | P0 |
| TC-066 | Story 4b / AC-1 | API | api/milestones | P0 |
| TC-067 | Story 4c / AC-1 | API | api/milestones | P0 |
| TC-068 | Story 5 / AC-1, AC-2, AC-3, AC-4 | API | api/milestones | P0 |
| TC-069 | Story 5 / AC-2, AC-3 | API | api/milestones | P1 |
| TC-070 | PRD Security Requirements | API | api/milestones | P0 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| /milestones | ⚠️ Provisional selectors | TC-001..TC-026, TC-035..TC-037, TC-047, TC-048 | Frontend: `App.tsx:37 <Route path="/milestones">` -- provisional data-testid selectors defined; sitemap not yet generated |
| /items | ✅ Matched | TC-033, TC-044, TC-049 | Frontend route exists; sitemap page exists with elements |
| /items/:mainItemId | ✅ Matched | TC-031, TC-032, TC-034, TC-045 | Frontend route exists; sitemap page exists with elements |
| /table | ✅ Matched | TC-038..TC-043, TC-046, TC-050..TC-052 | Frontend route exists; sitemap page exists with elements |
| POST /api/v1/teams/:teamId/milestone-maps | ✅ Matched | TC-053, TC-054 | `router.go:158` |
| GET /api/v1/teams/:teamId/milestone-maps | ✅ Matched | TC-055 | `router.go:159` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-056 | `router.go:160` |
| PUT /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-057 | `router.go:161` |
| DELETE /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-058 | `router.go:162` |
| PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status | ✅ Matched | TC-059 | `router.go:163` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions | ✅ Matched | TC-060 | `router.go:164` |
| POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones | ✅ Matched | TC-061, TC-062 | `router.go:167` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones | ✅ Matched | TC-063 | `router.go:168` |
| GET /api/v1/teams/:teamId/milestones | ✅ Matched | TC-064 | `router.go:171` |
| GET /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-065 | `router.go:172` |
| PUT /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-066 | `router.go:173` |
| DELETE /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-067 | `router.go:174` |
| PUT /api/v1/teams/:teamId/milestones/:milestoneId/status | ✅ Matched | TC-068 | `router.go:175` |
| GET /api/v1/teams/:teamId/milestones/:milestoneId/available-transitions | ✅ Matched | TC-069 | `router.go:176` |
