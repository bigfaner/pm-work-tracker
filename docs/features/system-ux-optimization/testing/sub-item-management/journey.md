---
feature: "system-ux-optimization"
journey: "sub-item-management"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 2, Story 5)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#2, #5)
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
generated: "2026-06-04"
---

# Journey: sub-item-management

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Workflow involves data mutation (start time edit) without destructive operations.
  Sorting is read-only. No security-sensitive operations.
-->

## Overview

PM user edits sub-item start time in the edit dialog and views sub-items sorted by creation time in descending order on the main item detail page. Note: sub-item start time update has no standalone API endpoint; the mutation is handled via the web form only. The API surface covers read-path operations (listing sub-items via main items endpoint).

## Setup

- A team exists with a main item that has multiple sub-items created at different times
- Sub-items have various start times assigned
- PM user is logged in with sub_item:update and main_item:read permissions
- At least one sub-item has both a start time and an end time set

## Happy Path

### Step 1: Open sub-item edit dialog
<!-- surface: web -->

**Precondition**: A sub-item exists with a start time value set <!-- fact: prd-spec #2, Story 2 AC1 -->

**User Action**: PM user opens the edit dialog for a sub-item

**Expected Result**: Edit dialog renders with all existing fields populated, including the "开始时间" field showing the current start time value

### Step 2: Edit start time and save
<!-- surface: web -->

**Precondition**: The sub-item edit dialog is open with a start time field <!-- fact: Story 2 AC1 -->

**User Action**: PM user changes the start time to a new date and clicks save

**Expected Result**: Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time

### Step 3: Verify list position preserved after edit
<!-- surface: web -->

**Precondition**: A sub-item's start time was just edited in Step 2; the main item detail page is displayed <!-- fact: prd-spec #5 — sort order is by creation time, not start time -->

**User Action**: PM user views the sub-item list on the main item detail page

**Expected Result**: The edited sub-item remains in the same position in the list as before the edit; list order is still determined by creation time, not start time

### Step 4: View sub-item list sorted by creation time
<!-- surface: web -->

**Precondition**: A main item has multiple sub-items created on different dates <!-- fact: Story 5 AC1 -->

**User Action**: PM user navigates to the main item detail page

**Expected Result**: Sub-items are displayed in descending order by creation time -- the most recently created sub-item appears first in the list

### Step 5: API list sub-items with creation time sort
<!-- surface: api -->

**Precondition**: An authenticated API request is sent for a main item that has multiple sub-items <!-- fact: api-handbook Enhanced List Main Items -->

**User Action**: An authenticated API request is sent to the main items endpoint for a team

**Expected Result**: The response includes sub-items under each main item, ordered by creation time descending (most recently created first)

## Edge Cases

### Step E1: Edit start time to an invalid date format
<!-- surface: web -->

**Precondition**: The sub-item edit dialog is open <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM user manually enters an invalid date format in the start time field and attempts to save

**Expected Result**: Date input validation prevents saving with an invalid date; an error message is shown near the date field indicating the invalid format

### Step E1b: Edit start time to a date after end time
<!-- surface: web -->

**Precondition**: A sub-item has an end time set to 2026-05-01 <!-- fact: prd-ui-functions.md line 115 — start time must not be later than end time -->

**User Action**: PM user sets the sub-item start time to 2026-06-01 (after the end time) and attempts to save

**Expected Result**: A validation error message is displayed indicating that start time must not be later than end time; the change is not saved

### Step E2: Sub-item list with single sub-item
<!-- surface: web -->

**Precondition**: A main item has exactly one sub-item <!-- fact: Story 5 — sorting applies regardless of item count -->

**User Action**: PM user views the main item detail page

**Expected Result**: The single sub-item is displayed; no sorting issues occur; empty state is not triggered

### Step E3: Sub-items with identical creation times
<!-- surface: web -->

**Precondition**: Multiple sub-items were created simultaneously (same timestamp) <!-- source: inferred — tiebreaker behavior when creation times match -->

**User Action**: PM user views the main item detail page

**Expected Result**: Sub-items with identical creation times are displayed in a stable, deterministic order

### Step E4: Session expired during sub-item edit (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has since expired while the edit dialog is open <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM user submits the edit form

**Expected Result**: The user is redirected to the login page; no data is modified; after re-authenticating, the sub-item retains its original values

### Step E5: Unauthorized API access to sub-item data
<!-- surface: api -->

**Precondition**: An authenticated user without main_item:read permission sends an API request <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The user sends an API request to the main items endpoint

**Expected Result**: The API returns an authorization error; no sub-item data is returned

### Step E6: Unauthenticated API request for sub-items
<!-- surface: api -->

**Precondition**: An API request is sent without valid credentials <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An API request is sent without a valid authentication token

**Expected Result**: The API returns an authentication error; no data is returned

### Step E7: API validation error for malformed request
<!-- surface: api -->

**Precondition**: An API request uses a non-numeric value for a team ID parameter <!-- source: inferred — derived from API surface `validation-error` outcome -->

**User Action**: An API request is sent with a non-numeric team ID (e.g., `abc` instead of a numeric identifier)

**Expected Result**: The API returns a validation error response listing the invalid parameter; no data is returned

### Step E8: API request for non-existent main item
<!-- surface: api -->

**Precondition**: The main item ID in the request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: An API request is sent for a non-existent main item

**Expected Result**: The API returns a "not found" error

### Step E9: Edit a concurrently deleted sub-item (Web)
<!-- surface: web -->

**Precondition**: Another user has deleted the sub-item while the current user has the edit dialog open <!-- source: inferred — derived from Web surface `not-found` boundary outcome -->

**User Action**: PM user submits the edit form

**Expected Result**: An error message is displayed indicating the sub-item no longer exists; the edit dialog closes; no data is modified

## Journey Invariants

- The sub-item edit dialog always includes a "开始时间" field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
