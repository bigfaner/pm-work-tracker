---
feature: "system-ux-optimization"
journey: "team-and-progress-visibility"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 10, Story 11)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#16)
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced Gantt View)
generated: "2026-06-04"
---

# Journey: team-and-progress-visibility

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Permission boundary enforcement with data leakage risk
  Team selector filtering prevents users from seeing unauthorized teams; incorrect filtering leaks team names.
-->

## Overview

User sees only teams they have permission to in the team selector, and the weekly progress page hides terminal main items with no activity in the current or previous week.

## Setup

- A team exists with multiple main items in various statuses and activity levels
- User belongs to Team A and Team B, but not Team C
- Team C exists in the system with other users as members
- At least one main item is in terminal status (closed or completed) with no activity in the current or previous week

## Happy Path

### Step 1: View team selector with permission filtering
<!-- surface: web -->

**Precondition**: User has permission to access Team A and Team B only <!-- fact: Story 10 AC1 -->

**User Action**: User opens the team selector after login

**Expected Result**: Only teams the user has permission to access are displayed (Team A and Team B); Team C is not shown

### Step 2: View weekly progress page with mixed activity
<!-- surface: web -->

**Precondition**: Non-terminal main items exist; a terminal main item has activity (status change or sub-item update) this week <!-- fact: Story 11, prd-spec #16 -->

**User Action**: User opens the weekly progress page

**Expected Result**: All non-terminal main items are displayed; terminal main items with activity this week or last week are displayed; terminal main items with no activity in either week are hidden

### Step 3: View weekly progress page with no active terminal items
<!-- surface: web -->

**Precondition**: All terminal main items have had no activity (status changes, sub-item updates, or progress updates) in the current or previous week <!-- fact: prd-spec #16 -->

**User Action**: User opens the weekly progress page

**Expected Result**: Only non-terminal main items are displayed; terminal items are hidden

### Step 4: API list teams with permission filtering
<!-- surface: api -->

**Precondition**: An authenticated user has permission to access Team A and Team B only <!-- fact: api-handbook team endpoints -->

**User Action**: An authenticated API request is sent to the team listing endpoint

**Expected Result**: Only Team A and Team B are returned; Team C is not included in the response

## Edge Cases

### Step E1: User with access to only one team
<!-- surface: web -->

**Precondition**: User has permission to access exactly one team

**User Action**: User opens the team selector

**Expected Result**: Only that one team is shown in the selector

### Step E2: Terminal item with activity in current week only
<!-- surface: web -->

**Precondition**: A completed main item had a sub-item update in the current week but no activity in the previous week <!-- fact: prd-spec #16 activity definition -->

**User Action**: User views the weekly progress page

**Expected Result**: The completed main item is displayed because it has activity in the current week

### Step E3: Terminal item with activity in previous week only
<!-- surface: web -->

**Precondition**: A closed main item had a status change in the previous week but no activity in the current week <!-- fact: prd-spec #16 activity definition -->

**User Action**: User views the weekly progress page

**Expected Result**: The closed main item is displayed because it had activity in the previous week

### Step E4: Session expired while viewing weekly progress (Web)
<!-- surface: web -->

**Precondition**: The user was previously authenticated and the session has since expired while viewing the weekly progress page <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: User attempts to apply a filter or navigate within the weekly progress page

**Expected Result**: The user is redirected to the login page; after re-authenticating, the weekly progress page loads with default filters

### Step E5: Validation error on team selection (Web)
<!-- surface: web -->

**Precondition**: The team selector URL is manipulated to contain an invalid team identifier <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: User navigates to a URL containing an invalid team identifier

**Expected Result**: An error message is displayed indicating the team identifier is invalid; the user is not shown any data for the invalid team

### Step E6: Unauthorized API access to team list
<!-- surface: api -->

**Precondition**: An authenticated user without team listing permission sends an API request <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The user sends an API request to the team listing endpoint without proper authorization

**Expected Result**: The API returns an authorization error; no team data is returned

### Step E7: Unauthenticated API request for weekly progress
<!-- surface: api -->

**Precondition**: An API request is sent without valid credentials <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An API request is sent to the weekly progress endpoint without a valid authentication token

**Expected Result**: The API returns an authentication error; no progress data is returned

### Step E8: API request for non-existent team's weekly progress
<!-- surface: api -->

**Precondition**: The team ID in the API request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: An API request is sent for a non-existent team's weekly progress data

**Expected Result**: The API returns a "not found" error

### Step E9: API validation error for invalid parameters
<!-- surface: api -->

**Precondition**: An API request contains a non-numeric value for a team ID parameter <!-- source: inferred — derived from API surface `validation-error` outcome -->

**User Action**: An API request is sent with an invalid team ID format

**Expected Result**: The API returns a validation error response describing the invalid parameter; no data is returned

## Journey Invariants

- Team selector always filters to only teams the current user has permission to access
- Non-terminal main items are always displayed on the weekly progress page regardless of activity
- Terminal main items are hidden only when they have zero activity in both the current week (Monday through Sunday) and the previous week
- Activity for terminal items includes: status changes, sub-item creation or updates, and main item progress updates within the time range
- The weekly progress page uses natural week boundaries (Monday through Sunday) for activity calculation
