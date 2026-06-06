---
feature: "system-ux-optimization"
journey: "list-filtering-and-sorting"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 8, Story 9)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#10, #11, #12)
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items, Enhanced Gantt View)
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

- A team exists with multiple main items in various statuses (pending, in-progress, closed, completed)
- Assignee A is responsible for some main items directly and some sub-items under other main items
- Assignee B is responsible for different items (no sub-item assignments)
- At least one main item is in terminal status (closed or completed)
- A user without main_item:read permission exists

## Happy Path

### Step 1: Filter by assignee with penetration
<!-- surface: web -->

**Precondition**: Assignee A is the direct assignee of some main items AND the assignee of sub-items under other main items <!-- fact: prd-spec #10, Story 8 AC1 -->

**User Action**: PM user selects assignee A in the assignee filter on the item list page

**Expected Result**: Results include main items where A is the direct assignee AND main items where A is the assignee of a sub-item; main items shown due to sub-item match display a "因子事项匹配" visual indicator, and only the matching sub-items are shown under those main items <!-- source: inferred — "only matching sub-items" derived from api-handbook matchedSubItemIds field -->

### Step 2: View terminal status sorting
<!-- surface: web -->

**Precondition**: The item list page is displayed with no filters applied; at least one main item is in terminal status and at least one is not <!-- fact: prd-spec #11, Story 9 AC1 -->

**User Action**: PM user views the item list page

**Expected Result**: Non-terminal main items appear first in their original order; terminal status main items (closed, completed) are sorted to the bottom of the list

### Step 3: Progress page default filter
<!-- surface: web -->

**Precondition**: The progress page has no stored filter preferences for the current session <!-- fact: prd-spec #12, Story 9 AC2 -->

**User Action**: PM user opens the overall progress page for the first time

**Expected Result**: "进行中" status checkbox is selected by default; only in-progress main items are displayed; terminal status items are filtered out

### Step 4: Clear all status filters
<!-- surface: web -->

**Precondition**: The progress page is loaded with the default "进行中" filter active <!-- fact: Story 9 AC3 -->

**User Action**: PM user deselects all status checkboxes on the progress page

**Expected Result**: All items are displayed (equivalent to "show all"), including terminal status items; terminal status main items still sort to the bottom of the list

### Step 5: Empty state with active filters
<!-- surface: web -->

**Precondition**: Filter criteria are applied that match no items in the system <!-- fact: prd-spec empty state handling -->

**User Action**: PM user applies filters that match no items

**Expected Result**: Empty state message "没有符合条件的事项" is displayed; a "清除过滤条件" action button is available to reset all filters

### Step 6: API list with assignee filter and penetration
<!-- surface: api -->

**Precondition**: A valid team exists with main items assigned to assignee A directly and via sub-items <!-- fact: api-handbook Enhanced List Main Items -->

**User Action**: An authenticated API request is sent to `GET /teams/:teamId/main-items?assigneeKey=<A_bizKey>`

**Expected Result**: The response includes main items with `matchType: "direct"` (where A is the direct assignee) and main items with `matchType: "indirect"` (where A is the assignee of a sub-item); indirect items include `matchedSubItemIds` listing the matching sub-item IDs

### Step 7: API list with multi-status filter
<!-- surface: api -->

**Precondition**: Main items exist in multiple statuses <!-- fact: api-handbook Enhanced Query Parameters, prd-spec #11 -->

**User Action**: An authenticated API request is sent to `GET /teams/:teamId/main-items?status=progressing,closed`

**Expected Result**: Only main items matching the specified statuses are returned; terminal status main items are sorted to the bottom of the response list

## Edge Cases

### Step E1: No filters selected shows all items
<!-- surface: web -->

**Precondition**: No status or assignee filters are selected <!-- fact: Story 8 AC3 -->

**User Action**: PM user loads the item list page

**Expected Result**: All items are displayed without any filtering applied

### Step E2: Filter by assignee with no sub-item matches
<!-- surface: web -->

**Precondition**: Assignee B is only responsible for main items directly, never sub-items

**User Action**: PM user filters by assignee B

**Expected Result**: Only main items directly assigned to B are shown; no "因子事项匹配" indicators appear

### Step E3: Progress page -- select additional status
<!-- surface: web -->

**Precondition**: The progress page is loaded with the default "进行中" filter active; completed main items exist

**User Action**: PM user selects the "已完成" status checkbox in addition to the default "进行中"

**Expected Result**: Both in-progress and completed items are displayed; terminal items still sort to the bottom

### Step E4: Unauthorized list API request
<!-- surface: api -->

**Precondition**: An authenticated user without main_item:read permission attempts to access the list API endpoint <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The user sends `GET /teams/:teamId/main-items` without proper authorization

**Expected Result**: The API returns an authorization error; no item data is returned

### Step E5: Unauthenticated list API request
<!-- surface: api -->

**Precondition**: A list API request is sent without valid credentials <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An API request is sent to `GET /teams/:teamId/main-items` without a valid authentication token

**Expected Result**: The API returns an authentication error; no item data is returned

### Step E6: API validation error for invalid filter parameters
<!-- surface: api -->

**Precondition**: A list API request contains invalid filter values (e.g., non-existent status value, malformed assigneeKey) <!-- source: inferred — derived from API surface `validation-error` outcome -->

**User Action**: An API request is sent with an invalid status value (e.g., `status=invalid_status`)

**Expected Result**: The API returns a validation error response describing the invalid parameter; no filtered results are returned

### Step E7: Session expired during filter interaction (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has since expired while the filter controls are displayed <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM user applies a new filter

**Expected Result**: The user is redirected to the login page; after re-authenticating, the item list is displayed without the filter that was being applied

### Step E8: Validation error on filter parameters (Web)
<!-- surface: web -->

**Precondition**: The assignee filter field accepts text input that is validated against known assignee identifiers <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM user enters special characters or an extremely long string in the assignee filter field and triggers the filter

**Expected Result**: An error message is displayed near the filter field indicating invalid input; no API request with invalid data is sent

### Step E9: API list for non-existent team
<!-- surface: api -->

**Precondition**: The team ID in the request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: An API request is sent to `GET /teams/<non-existent-teamId>/main-items`

**Expected Result**: The API returns a "not found" error

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible, regardless of active filters or whether all filters are cleared
- When no filters are selected, all items are displayed
- Empty filter results always show "没有符合条件的事项" with a "清除过滤条件" action
- Sub-item match indicator ("因子事项匹配") is displayed whenever a main item is shown due to sub-item filter match
