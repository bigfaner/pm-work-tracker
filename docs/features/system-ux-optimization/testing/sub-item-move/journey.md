---
feature: "system-ux-optimization"
journey: "sub-item-move"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 7)
  - docs/features/system-ux-optimization/prd/prd-spec.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Move Sub Item)
generated: "2026-06-04"
---

# Journey: sub-item-move

**Risk Level**: High

<!-- Risk Classification Criteria:
  High = Workflow involves state mutation, data loss risk, or irreversible operations
  Moving a sub-item changes its parent relationship and reassigns its code number.
-->

## Overview

PM user moves a sub-item from one main item to another, with automatic re-numbering under the target main item while preserving status and assignee.

## Setup

- PM user is logged in with appropriate permissions
- A sub-item exists under main item A
- Another main item B exists in a non-terminal (non-closed, non-completed) status
- A main item C exists in terminal status (closed or completed)
- At least two PM users have access to the team (for concurrent operations)

## Happy Path

### Step 1: Initiate sub-item move
<!-- surface: web -->

**Precondition**: PM user is viewing a sub-item detail page that belongs to main item A

**User Action**: PM user selects the "Move to another main item" action <!-- fact: prd-spec Story 7 AC1 -->

**Expected Result**: A target main item selector is displayed, listing available (non-closed, non-completed) main items excluding the current parent

### Step 2: Select valid target and confirm
<!-- surface: web -->

**Precondition**: The move target selector is open, showing at least one valid target

**User Action**: PM user selects main item B as the target and confirms the move

**Expected Result**: Sub-item is moved to main item B; the code number is automatically regenerated under the target; status and assignee remain unchanged; the page updates to reflect the new parent

### Step 3: Verify sub-item in new location
<!-- surface: web -->

**Precondition**: The sub-item has been moved to main item B

**User Action**: PM user navigates to main item B's detail page

**Expected Result**: The moved sub-item appears in the sub-item list under main item B with its new code number; the sub-item no longer appears under main item A

## Edge Cases

### Step E1: Move to a closed main item
<!-- surface: web -->

**Precondition**: A main item C is in terminal status (closed or completed)

**User Action**: PM user attempts to select main item C as the move target

**Expected Result**: The closed main item is shown as disabled or non-selectable in the selector; if somehow selected, the operation is rejected with a message indicating the target is closed

### Step E2: Move to the same main item
<!-- surface: web -->

**Precondition**: Sub-item currently belongs to main item A; the move target selector shows main item A as excluded

**User Action**: PM user attempts to move the sub-item to main item A (same parent)

**Expected Result**: The operation is rejected; a message indicates the sub-item is already under this main item; no change occurs

### Step E3: Concurrent move of same sub-item
<!-- surface: web, api -->

**Precondition**: Two PM users have the same sub-item's move dialog open simultaneously

**User Action**: Both users confirm the move to different targets at nearly the same time

**Expected Result A** (first request arrives): The first transaction succeeds; the moved sub-item appears under the first user's target with a new code number.

**Expected Result B** (second request arrives after first transaction committed): The second transaction detects the sub-item's parent has changed and returns a conflict error; no data corruption occurs. <!-- source: inferred — derived from concurrent edit pattern in prd-spec -->

### Step E4: Move sub-item from a deleted main item
<!-- surface: web -->

**Precondition**: Main item A has been soft-deleted by another user while the current user has the move dialog open

**User Action**: PM user confirms the move

**Expected Result**: An error message is displayed indicating the source main item no longer exists; no move is performed

### Step E5: Unauthorized move attempt (API)
<!-- surface: api -->

**Precondition**: A user without the sub_item:update permission sends a move API request <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The user sends a move API request without proper authorization

**Expected Result**: The API returns an authorization error; no data is modified

### Step E6: Move without selecting a target (Web validation)
<!-- surface: web -->

**Precondition**: The move target selector is open but no target has been selected <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM user attempts to confirm the move without selecting a target

**Expected Result**: The confirm button is disabled or a validation error message is displayed indicating a target must be selected; no API request is sent

### Step E7: Session expired during move (Web)
<!-- surface: web -->

**Precondition**: The user's session has expired while the move target selector is open <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM user selects a target and confirms the move

**Expected Result**: The user is redirected to the login page; after re-authenticating, the move dialog is no longer open

### Step E8: Move non-existent sub-item (API)
<!-- surface: api -->

**Precondition**: The sub-item ID in the request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: A move API request is sent with a non-existent sub-item ID

**Expected Result**: The API returns a "not found" error; no data is modified

### Step E9: Move to non-existent target main item (API)
<!-- surface: api -->

**Precondition**: The target main item ID in the request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: A move API request is sent with a non-existent target main item ID

**Expected Result**: The API returns a "not found" error; no data is modified

### Step E10: API validation error for missing fields
<!-- surface: api -->

**Precondition**: A move API request is sent with missing or invalid fields (e.g., empty target main item ID) <!-- source: inferred — derived from API surface `validation-error` outcome -->

**User Action**: An API request is sent to the move endpoint with an invalid or missing target ID

**Expected Result**: The API returns a validation error response listing the missing or invalid fields; no data is modified

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Status and assignee of the moved sub-item are never changed by the move operation
- Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)
- Closed/completed main items are never valid targets for sub-item moves
- Moving to the same parent main item is always rejected
