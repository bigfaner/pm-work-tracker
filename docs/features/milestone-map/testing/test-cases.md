---
feature: "milestone-map"
sources:
  - docs/features/milestone-map/prd/prd-user-stories.md
  - docs/features/milestone-map/prd/prd-spec.md
  - docs/features/milestone-map/prd/prd-ui-functions.md
generated: "2026-05-13"
---

# Test Cases: milestone-map

> **WARNING**: sitemap.json has no page data for `/milestones` route — Element set to `sitemap-missing` for milestone page test cases. Run `/gen-sitemap` to update. For existing pages (`/items`, `/items/:mainItemId`, `/table`), sitemap element IDs are used where available.

## Summary

| Type | Count |
|------|-------|
| UI   | 47   |
| **Integration** | **4** |
| API  | 18  |
| CLI  | 0  |
| **Total** | **65** |

> **Note**: Integration test count is a subset of UI count. Integration tests verify that components are correctly wired into their parent pages, using the same Playwright framework as UI tests.

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
- **Element**: sitemap-missing
- **Steps**:
  1. Click the create milestone map button
  2. Fill in name (valid, 1-100 chars) and optional description
  3. Click confirm button
- **Expected**: Milestone map is created successfully with status "planning"; list view refreshes and shows the new card
- **Priority**: P0

## TC-002: Create milestone map with name at max length boundary
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-name-max-length
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the create milestone map button
  2. Enter a name that is exactly 100 characters
  3. Click confirm button
- **Expected**: Milestone map is created successfully; form closes; list view shows new card
- **Priority**: P1

## TC-003: Create milestone map with name exceeding max length
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-name-exceeds-max
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the create milestone map button
  2. Enter a name that is 101 characters
  3. Observe form validation
- **Expected**: Form displays "Name cannot exceed 100 characters" error; form does not submit
- **Priority**: P1

## TC-004: Create milestone map with empty name
- **Source**: Story 1 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-map-empty-name
- **Pre-conditions**: User has `milestone:create` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the create milestone map button
  2. Leave the name field empty
  3. Click confirm button
- **Expected**: Form displays "Name cannot be empty" error; form does not submit
- **Priority**: P1

## TC-005: Edit milestone map info
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/edit-milestone-map-info
- **Pre-conditions**: A milestone map exists; user has `milestone:update` permission; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the edit button on a milestone map card
  2. Modify name or description
  3. Click save button
- **Expected**: Changes are saved immediately; list view reflects the updated information
- **Priority**: P0

## TC-006: Change milestone map status from planning to reviewed
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-map-status-planning-to-reviewed
- **Pre-conditions**: A milestone map is in "planning" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the status Badge on the milestone map
  2. Observe dropdown menu options
  3. Select "reviewed" option
- **Expected**: Dropdown only shows "reviewed" option; after selection, status changes to "reviewed"
- **Priority**: P0

## TC-007: Change milestone map status from in-progress to completed
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-map-status-inprogress-options
- **Pre-conditions**: A milestone map is in "in-progress" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the status Badge on the milestone map
  2. Observe dropdown menu options
- **Expected**: Dropdown shows "pending-implementation" and "completed" options
- **Priority**: P1

## TC-008: Completed milestone map shows no transitions
- **Source**: Story 3 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/completed-milestone-map-no-transitions
- **Pre-conditions**: A milestone map is in "completed" status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the status Badge on the milestone map
  2. Observe dropdown menu
- **Expected**: Dropdown shows no available options; displays "No transitions available"
- **Priority**: P1

## TC-009: Filter milestone maps by status
- **Source**: Story 3 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/filter-milestone-maps-by-status
- **Pre-conditions**: Team has multiple milestone maps with different statuses; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Use the status filter to select "in-progress"
  2. Observe the filtered list
- **Expected**: List only shows milestone map cards with "in-progress" status
- **Priority**: P0

## TC-010: List view renders all milestone map cards
- **Source**: Story 8 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/list-view-renders-all-cards
- **Pre-conditions**: Team has 3+ milestone maps; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page
  2. Observe the list view
- **Expected**: List view correctly renders all milestone map cards; each card shows name, status, milestone count, item count, overall progress
- **Priority**: P0

## TC-011: Empty state when no milestone maps exist
- **Source**: Story 8 / AC-5
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/empty-state-no-maps
- **Pre-conditions**: Team has 0 milestone maps; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page
  2. Observe the empty state
- **Expected**: Page shows empty state message "No milestone maps yet" with a create button
- **Priority**: P1

## TC-012: Error state on API failure
- **Source**: Story 8 / AC-6
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/error-state-api-failure
- **Pre-conditions**: Team has milestone maps but backend API returns 500 error
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page while backend returns 500
  2. Observe the error state
- **Expected**: Page shows "Load failed, please retry" message with retry button; does not show blank page
- **Priority**: P1

### Timeline View

## TC-013: Enter timeline view from card click
- **Source**: Story 8 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/enter-timeline-from-card
- **Pre-conditions**: A milestone map with milestones exists; user is on /milestones list view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click a milestone map card
  2. Observe the timeline view
- **Expected**: Timeline view renders all milestone nodes; each node shows name, planned completion date, status, and completion rate
- **Priority**: P0

## TC-014: Timeline zoom controls
- **Source**: Story 8 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/timeline-zoom-controls
- **Pre-conditions**: User is in timeline view with milestones
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click zoom control to switch to week view
  2. Observe axis labels
  3. Click zoom control to switch to month view
  4. Observe axis labels
  5. Click zoom control to switch to quarter view
  6. Observe axis labels
- **Expected**: Time axis scale labels change accordingly for each zoom level; milestone and item positions rearrange
- **Priority**: P1

### Milestone Creation/Editing

## TC-015: Create milestone successfully
- **Source**: Story 4a / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-successfully
- **Pre-conditions**: User has `milestone:create` permission; user is in a milestone map timeline view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Create Milestone" button
  2. Fill in name and planned completion date
  3. Click confirm button
- **Expected**: Milestone is created with status "not_started" and completion rate 0; timeline refreshes
- **Priority**: P0

## TC-016: Create milestone with name at max length
- **Source**: Story 4a / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-name-max-length
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Create Milestone" button
  2. Enter a name that is exactly 100 characters
  3. Click confirm button
- **Expected**: Milestone is created successfully
- **Priority**: P1

## TC-017: Create milestone with name exceeding max length
- **Source**: Story 4a / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-name-exceeds-max
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Create Milestone" button
  2. Enter a name that is 101 characters
  3. Observe form validation
- **Expected**: Form displays "Name cannot exceed 100 characters" error; form does not submit
- **Priority**: P1

## TC-018: Create milestone with empty name
- **Source**: Story 4a / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-empty-name
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Create Milestone" button
  2. Leave name field empty
  3. Click confirm button
- **Expected**: Form displays "Name cannot be empty" error; form does not submit
- **Priority**: P1

## TC-019: Create milestone shows error on server failure
- **Source**: Story 4a / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/create-milestone-server-error
- **Pre-conditions**: User has `milestone:create` permission; user is in timeline view; backend returns 500
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Create Milestone" button
  2. Fill in name and planned completion date
  3. Click confirm button (backend returns 500)
- **Expected**: Page shows "Create failed, please retry"; form retains entered data without loss
- **Priority**: P1

## TC-020: Edit milestone info
- **Source**: Story 4b / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/edit-milestone-info
- **Pre-conditions**: A milestone exists; user has `milestone:update` permission; user is in timeline view
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel by clicking a milestone node
  2. Click "Edit" button
  3. Modify name or planned completion date
  4. Click save button
- **Expected**: Changes are saved; timeline reflects the updated information
- **Priority**: P0

## TC-021: Concurrent edit conflict on milestone
- **Source**: Story 4b / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/concurrent-edit-conflict-milestone
- **Pre-conditions**: Two PMs are editing the same milestone simultaneously; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. PM-A edits and saves a milestone
  2. PM-B (same test context) attempts to save the same milestone afterward
- **Expected**: PM-B receives conflict message "Data has been modified by another user, please refresh and retry"; no silent overwrite
- **Priority**: P2

## TC-022: Delete milestone
- **Source**: Story 4c / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/delete-milestone
- **Pre-conditions**: A milestone exists with associated MIs; user has `milestone:delete` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel
  2. Click "Delete" button
  3. Confirm deletion in the confirmation dialog
- **Expected**: Milestone is soft-deleted; all associated MIs have their milestone_key cleared in the same transaction
- **Priority**: P0

### Milestone Status Changes

## TC-023: Change milestone status from not_started to in_progress
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/change-milestone-status-not-started-to-in-progress
- **Pre-conditions**: A milestone is in `not_started` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel
  2. Click status Badge
  3. Select `in_progress` from dropdown
- **Expected**: Status changes to `in_progress`; timeline node style updates accordingly
- **Priority**: P0

## TC-024: Completed milestone shows only cancelled option
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/completed-milestone-only-cancelled
- **Pre-conditions**: A milestone is in `completed` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel
  2. Click status Badge
  3. Observe dropdown menu options
- **Expected**: Dropdown only shows "cancelled" option; no `in_progress` rollback option
- **Priority**: P1

## TC-025: Cancelled milestone shows no transitions
- **Source**: Story 5 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/cancelled-milestone-no-transitions
- **Pre-conditions**: A milestone is in `cancelled` status; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel
  2. Click status Badge
  3. Observe dropdown menu
- **Expected**: Dropdown shows no available options; displays "No transitions available"
- **Priority**: P1

## TC-026: Cancel milestone auto-unbinds associated MIs
- **Source**: Story 5 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/cancel-milestone-auto-unbind
- **Pre-conditions**: A milestone is in `completed` or `in_progress` status with associated MIs; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Open milestone detail panel
  2. Click status Badge
  3. Select "cancelled"
- **Expected**: Milestone status changes to `cancelled`; all associated MIs are automatically unbound
- **Priority**: P0

### Milestone Detail Panel

## TC-027: Unbind MI from milestone detail panel
- **Source**: Story 7 / AC-1
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/unbind-mi-from-detail-panel
- **Pre-conditions**: Milestone detail panel is open with associated MIs; user has `milestone:update` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click the x button on the right side of an MI row
- **Expected**: MI is unbound from the milestone; MI list refreshes; undo toast is displayed
- **Priority**: P0

## TC-028: Quick add MI from detail panel
- **Source**: Story 7 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-from-detail-panel
- **Pre-conditions**: Milestone detail panel is open; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Click "+ Add" button in the detail panel
  2. Observe the create MI form opens
- **Expected**: Create MI form opens with milestone field auto-populated to current milestone and disabled (not editable)
- **Priority**: P0

## TC-029: Quick add MI form creates and binds
- **Source**: Story 7 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-form-create-bind
- **Pre-conditions**: Quick add MI form is open; user has `milestone:create` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Fill in title, assignee, start date, expected end date
  2. Click confirm button
- **Expected**: MI is created and automatically bound to current milestone; detail panel MI list refreshes
- **Priority**: P0

## TC-030: Quick add MI form milestone field is disabled
- **Source**: Story 7 / AC-4
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/quick-add-mi-milestone-field-disabled
- **Pre-conditions**: Quick add MI form is open
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Observe the milestone field in the form
- **Expected**: Milestone field displays current milestone name in disabled state; cannot be modified
- **Priority**: P1

### MI-Milestone Binding (via Items Page)

## TC-031: Bind MI to milestone from item edit page
- **Source**: Story 6 / AC-1
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/bind-mi-to-milestone
- **Pre-conditions**: A MainItem is not assigned to any milestone; milestones exist in the team
- **Route**: /items/:mainItemId
- **Element**: E-035, E-036, E-037
- **Steps**:
  1. Open main item edit dialog
  2. Select a milestone from the milestone dropdown
  3. Click save
- **Expected**: MI's milestone_key is set to the selected milestone's bizKey; milestone completion rate updates
- **Priority**: P0

## TC-032: Unbind MI from milestone via item edit page
- **Source**: Story 6 / AC-2
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/unbind-mi-from-milestone
- **Pre-conditions**: A MainItem belongs to a milestone; user is on item edit page
- **Route**: /items/:mainItemId
- **Element**: E-035, E-036, E-037
- **Steps**:
  1. Open main item edit dialog
  2. Clear the milestone field (select "unassigned")
  3. Click save
- **Expected**: MI's milestone_key is cleared; original milestone completion rate recalculates
- **Priority**: P0

## TC-033: Filter items by milestone on items list page
- **Source**: Story 6 / AC-3
- **Type**: UI
- **Target**: ui/items
- **Test ID**: ui/items/filter-items-by-milestone
- **Pre-conditions**: MIs are assigned to various milestones; user is on /items page
- **Route**: /items
- **Element**: E-010, E-011
- **Steps**:
  1. Select a specific milestone from the milestone filter dropdown
  2. Observe the filtered item list
- **Expected**: List only shows MIs belonging to the selected milestone
- **Priority**: P0

## TC-034: No milestones available shows only unassigned option
- **Source**: Story 6 / AC-4
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/no-milestones-only-unassigned
- **Pre-conditions**: Team has no milestones created; user opens item edit dialog
- **Route**: /items/:mainItemId
- **Element**: E-035, E-036, E-037
- **Steps**:
  1. Open main item edit dialog
  2. Observe the milestone dropdown options
- **Expected**: Dropdown only shows "unassigned" option; no other selectable options
- **Priority**: P1

### Permission-Based Views

## TC-035: Read-only user sees milestones page without action buttons
- **Source**: Story 9 / AC-1, Story 9 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/read-only-user-view
- **Pre-conditions**: User has `milestone:read` permission only (no create/update/delete)
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page
  2. Observe list view and timeline view
- **Expected**: Full list view and timeline view are visible in read-only mode; no create/edit/delete buttons shown; create button is disabled with tooltip "No permission"
- **Priority**: P0

## TC-036: Read-only user with no maps sees empty state without create button
- **Source**: Story 9 / AC-3
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/read-only-empty-no-create
- **Pre-conditions**: User has `milestone:read` permission only; team has 0 milestone maps
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page
  2. Observe empty state
- **Expected**: Empty state message "No milestone maps yet" is shown; no create button displayed
- **Priority**: P1

## TC-037: No milestone read permission shows 403
- **Source**: Story 11 / AC-2
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/no-read-permission-403
- **Pre-conditions**: User does NOT have `milestone:read` permission
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Navigate to /milestones page
- **Expected**: Page returns 403 permission denied message
- **Priority**: P0

### Table View Milestone Column

## TC-038: Table view shows milestone column
- **Source**: Story 10 / AC-1
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/shows-milestone-column
- **Pre-conditions**: Table view is loaded; milestone feature is available
- **Route**: /table
- **Element**: E-089, E-090
- **Steps**:
  1. Navigate to /table page
  2. Observe the table columns
- **Expected**: Table has a "milestone" column between "title" and "priority" columns; shows milestone name or "-" for unassigned
- **Priority**: P0

## TC-039: Table view filter by milestone column
- **Source**: Story 10 / AC-2
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/filter-by-milestone-column
- **Pre-conditions**: Table view has milestone column with various milestone assignments
- **Route**: /table
- **Element**: E-089
- **Steps**:
  1. Use milestone column filter to select a specific milestone
  2. Observe filtered results
- **Expected**: Only MIs belonging to the selected milestone are shown
- **Priority**: P0

## TC-040: Table view milestone column error fallback
- **Source**: Story 10 / AC-3
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/milestone-column-error-fallback
- **Pre-conditions**: Table view is loaded; milestone data loading fails
- **Route**: /table
- **Element**: E-089
- **Steps**:
  1. Navigate to /table page with milestone data loading failure
  2. Observe the milestone column
- **Expected**: Milestone column displays "--"; other columns render normally; table loading is not blocked
- **Priority**: P2

## TC-041: Table view shows "--" for soft-deleted milestone MI
- **Source**: Story 10 / AC-4
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/soft-deleted-milestone-shows-dash
- **Pre-conditions**: MI's milestone_key points to a soft-deleted milestone
- **Route**: /table
- **Element**: E-089
- **Steps**:
  1. Navigate to /table page
  2. Find the MI with a soft-deleted milestone
  3. Observe its milestone column
- **Expected**: Milestone column shows "--"
- **Priority**: P1

## TC-042: Table view sort by milestone column descending
- **Source**: Story 10 / AC-5
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/sort-milestone-column-desc
- **Pre-conditions**: Table view has milestone column with various assignments
- **Route**: /table
- **Element**: E-089
- **Steps**:
  1. Sort milestone column in descending order
  2. Observe the ordering
- **Expected**: MIs with assigned milestones are sorted by milestone name descending; unassigned MIs appear at the end
- **Priority**: P1

## TC-043: Table view default milestone sort ascending
- **Source**: Story 10 / AC-6
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/default-milestone-sort-ascending
- **Pre-conditions**: Table view has milestone column; no sort setting has been changed
- **Route**: /table
- **Element**: E-089
- **Steps**:
  1. Navigate to /table page (fresh load, default sort)
  2. Observe milestone column ordering
- **Expected**: MIs with assigned milestones are sorted by milestone name ascending (default); unassigned MIs appear at the end
- **Priority**: P1

### Integration Tests (Existing Page Modifications)

## TC-044: Integration -- Milestone filter visible on Items page
- **Source**: PRD UI Function "UF-4" Placement + Integration Spec
- **Type**: UI
- **Target**: ui/items
- **Test ID**: ui/items/integration-milestone-filter
- **Pre-conditions**: Milestone filter component build complete, integration task complete
- **Route**: /items
- **Element**: E-010, E-011 (filter area)
- **Steps**:
  1. Navigate to /items
  2. Verify milestone filter dropdown is visible in the filter bar area, to the right of "assignee" filter
  3. Verify milestone filter renders with "all" as default value
- **Expected**: Milestone filter dropdown appears at the specified position and shows default "all" state
- **Priority**: P0

## TC-045: Integration -- Milestone selector visible in Item Edit dialog
- **Source**: PRD UI Function "UF-5" Placement + Integration Spec
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/integration-milestone-selector
- **Pre-conditions**: Milestone selector component build complete, integration task complete
- **Route**: /items/:mainItemId
- **Element**: E-035, E-037 (edit modal, below assignee field)
- **Steps**:
  1. Navigate to /items/:mainItemId
  2. Click "Edit" button to open edit dialog
  3. Verify "Milestone" selector is visible below the "Assignee" field
  4. Verify selector renders with current milestone or "unassigned" value
- **Expected**: Milestone selector appears below assignee field and displays current assignment correctly
- **Priority**: P0

## TC-046: Integration -- Milestone column visible in Table view
- **Source**: PRD UI Function "UF-6" Placement + Integration Spec
- **Type**: UI
- **Target**: ui/table
- **Test ID**: ui/table/integration-milestone-column
- **Pre-conditions**: Milestone column component build complete, integration task complete
- **Route**: /table
- **Element**: E-089, E-090 (between title and priority columns)
- **Steps**:
  1. Navigate to /table
  2. Verify "Milestone" column header is visible between "Title" and "Priority" columns
  3. Verify column cells display milestone names or "-" for unassigned
- **Expected**: Milestone column appears between title and priority columns and displays data correctly
- **Priority**: P0

## TC-047: Integration -- Milestones page navigation link exists
- **Source**: PRD UI Function "UF-1" Navigation Architecture
- **Type**: UI
- **Target**: ui/milestones
- **Test ID**: ui/milestones/integration-navigation-link
- **Pre-conditions**: Milestones page route is registered in App.tsx; navigation is updated
- **Route**: /milestones
- **Element**: sitemap-missing
- **Steps**:
  1. Verify a "Milestones" navigation link exists in the main sidebar, between "Items" and "Gantt" links
  2. Click the link
  3. Verify navigation to /milestones route
- **Expected**: Navigation link exists at correct position and navigates to /milestones page
- **Priority**: P0

---

## API Test Cases

### MilestoneMap CRUD

## TC-048: API Create milestone map
- **Source**: Story 1 / AC-1, PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/create-milestone-map
- **Pre-conditions**: User authenticated with `milestone:create` permission; valid team context
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps with body `{name: "Test Map", description: "desc"}`
- **Expected**: Returns 200 with created milestone map; status is "planning"; name and description match input
- **Priority**: P0

## TC-049: API Create milestone map validation errors
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

## TC-050: API List milestone maps
- **Source**: Story 8 / AC-1, PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/list-milestone-maps
- **Pre-conditions**: Team has 3+ milestone maps; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps
- **Expected**: Returns 200 with list of milestone maps; each item includes name, status, milestone count, item count, overall progress
- **Priority**: P0

## TC-051: API Get milestone map by ID
- **Source**: PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/get-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps/:mapId
- **Expected**: Returns 200 with milestone map details including name, description, status, computed fields
- **Priority**: P0

## TC-052: API Update milestone map
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/update-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestone-maps/:mapId with body `{name: "Updated Name", description: "Updated desc"}`
- **Expected**: Returns 200 with updated milestone map; name and description reflect changes
- **Priority**: P0

## TC-053: API Delete milestone map
- **Source**: PRD Spec Related Changes #1
- **Type**: API
- **Target**: api/milestone-maps
- **Test ID**: api/milestone-maps/delete-milestone-map
- **Pre-conditions**: A milestone map exists; user has `milestone:delete` permission
- **Steps**:
  1. DELETE /api/v1/teams/:teamId/milestone-maps/:mapId
- **Expected**: Returns 200; milestone map is soft-deleted; subsequent GET returns 404
- **Priority**: P0

## TC-054: API Change milestone map status
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

## TC-055: API Get available transitions for milestone map
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

## TC-056: API Create milestone
- **Source**: Story 4a / AC-1
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/create-milestone
- **Pre-conditions**: User has `milestone:create` permission; a milestone map exists
- **Steps**:
  1. POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones with body `{name: "Phase 1", planned_completion_date: "2026-06-30"}`
- **Expected**: Returns 200 with created milestone; status is "not_started"; completion rate is 0
- **Priority**: P0

## TC-057: API Create milestone validation errors
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

## TC-058: API List milestones by map
- **Source**: PRD Spec Related Changes #2
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/list-by-map
- **Pre-conditions**: A milestone map has milestones; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
- **Expected**: Returns 200 with list of milestones under the specified map; each includes computed completion rate
- **Priority**: P0

## TC-059: API List milestones by team
- **Source**: PRD Spec Related Changes #2, Story 6 / AC-3
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/list-by-team
- **Pre-conditions**: Team has milestones across multiple maps; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestones
- **Expected**: Returns 200 with all non-cancelled milestones in the team
- **Priority**: P0

## TC-060: API Get milestone by ID
- **Source**: PRD Spec Related Changes #2
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/get-milestone
- **Pre-conditions**: A milestone exists; user has `milestone:read` permission
- **Steps**:
  1. GET /api/v1/teams/:teamId/milestones/:milestoneId
- **Expected**: Returns 200 with milestone details including computed completion rate and associated MI count
- **Priority**: P0

## TC-061: API Update milestone
- **Source**: Story 4b / AC-1
- **Type**: API
- **Target**: api/milestones
- **Test ID**: api/milestones/update-milestone
- **Pre-conditions**: A milestone exists; user has `milestone:update` permission
- **Steps**:
  1. PUT /api/v1/teams/:teamId/milestones/:milestoneId with body `{name: "Updated Phase", planned_completion_date: "2026-07-31"}`
- **Expected**: Returns 200 with updated milestone
- **Priority**: P0

## TC-062: API Delete milestone unbinds associated MIs
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

## TC-063: API Change milestone status
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

## TC-064: API Get available transitions for milestone
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

## TC-065: API milestone operations without permission return 403
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
| TC-009 | Story 3 / AC-4 | UI | ui/milestones | P0 |
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
| TC-035 | Story 9 / AC-1, AC-2 | UI | ui/milestones | P0 |
| TC-036 | Story 9 / AC-3 | UI | ui/milestones | P1 |
| TC-037 | Story 11 / AC-2 | UI | ui/milestones | P0 |
| TC-038 | Story 10 / AC-1 | UI | ui/table | P0 |
| TC-039 | Story 10 / AC-2 | UI | ui/table | P0 |
| TC-040 | Story 10 / AC-3 | UI | ui/table | P2 |
| TC-041 | Story 10 / AC-4 | UI | ui/table | P1 |
| TC-042 | Story 10 / AC-5 | UI | ui/table | P1 |
| TC-043 | Story 10 / AC-6 | UI | ui/table | P1 |
| TC-044 | UF-4 Placement + Integration | UI | ui/items | P0 |
| TC-045 | UF-5 Placement + Integration | UI | ui/items-detail | P0 |
| TC-046 | UF-6 Placement + Integration | UI | ui/table | P0 |
| TC-047 | UF-1 Navigation Architecture | UI | ui/milestones | P0 |
| TC-048 | Story 1 / AC-1, Related Changes #1 | API | api/milestone-maps | P0 |
| TC-049 | Story 1 / AC-2, AC-3 | API | api/milestone-maps | P0 |
| TC-050 | Story 8 / AC-1, Related Changes #1 | API | api/milestone-maps | P0 |
| TC-051 | Related Changes #1 | API | api/milestone-maps | P0 |
| TC-052 | Story 2 / AC-1 | API | api/milestone-maps | P0 |
| TC-053 | Related Changes #1 | API | api/milestone-maps | P0 |
| TC-054 | Story 3 / AC-1, AC-2, AC-3 | API | api/milestone-maps | P0 |
| TC-055 | Story 3 / AC-1, AC-2, AC-3 | API | api/milestone-maps | P1 |
| TC-056 | Story 4a / AC-1 | API | api/milestones | P0 |
| TC-057 | Story 4a / AC-2, AC-3 | API | api/milestones | P0 |
| TC-058 | Related Changes #2 | API | api/milestones | P0 |
| TC-059 | Related Changes #2, Story 6 / AC-3 | API | api/milestones | P0 |
| TC-060 | Related Changes #2 | API | api/milestones | P0 |
| TC-061 | Story 4b / AC-1 | API | api/milestones | P0 |
| TC-062 | Story 4c / AC-1 | API | api/milestones | P0 |
| TC-063 | Story 5 / AC-1, AC-2, AC-3, AC-4 | API | api/milestones | P0 |
| TC-064 | Story 5 / AC-2, AC-3 | API | api/milestones | P1 |
| TC-065 | PRD Security Requirements | API | api/milestones | P0 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| /milestones | ⚠️ No sitemap page data | TC-001..TC-026, TC-035..TC-037, TC-047 | Frontend: `App.tsx:37 <Route path="/milestones">` -- sitemap has no page entry for this route |
| /items | ✅ Matched | TC-033, TC-044 | Frontend route exists; sitemap page exists with elements |
| /items/:mainItemId | ✅ Matched | TC-031, TC-032, TC-034, TC-045 | Frontend route exists; sitemap page exists with elements |
| /table | ✅ Matched | TC-038..TC-043, TC-046 | Frontend route exists; sitemap page exists with elements |
| POST /api/v1/teams/:teamId/milestone-maps | ✅ Matched | TC-048, TC-049 | `router.go:158` |
| GET /api/v1/teams/:teamId/milestone-maps | ✅ Matched | TC-050 | `router.go:159` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-051 | `router.go:160` |
| PUT /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-052 | `router.go:161` |
| DELETE /api/v1/teams/:teamId/milestone-maps/:mapId | ✅ Matched | TC-053 | `router.go:162` |
| PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status | ✅ Matched | TC-054 | `router.go:163` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions | ✅ Matched | TC-055 | `router.go:164` |
| POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones | ✅ Matched | TC-056, TC-057 | `router.go:167` |
| GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones | ✅ Matched | TC-058 | `router.go:168` |
| GET /api/v1/teams/:teamId/milestones | ✅ Matched | TC-059 | `router.go:171` |
| GET /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-060 | `router.go:172` |
| PUT /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-061 | `router.go:173` |
| DELETE /api/v1/teams/:teamId/milestones/:milestoneId | ✅ Matched | TC-062 | `router.go:174` |
| PUT /api/v1/teams/:teamId/milestones/:milestoneId/status | ✅ Matched | TC-063 | `router.go:175` |
| GET /api/v1/teams/:teamId/milestones/:milestoneId/available-transitions | ✅ Matched | TC-064 | `router.go:176` |
