---
feature: "system-ux-optimization"
journey: "item-deletion"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 3)
  - docs/features/system-ux-optimization/prd/prd-spec.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Delete Main Item, Delete Sub Item)
generated: "2026-06-04"
---

# Journey: item-deletion

**Risk Level**: High

<!-- Risk Classification Criteria:
  High = Workflow involves state mutation, data loss risk, or irreversible operations
  Deletion is a destructive operation (soft delete) that cascades to child items.
-->

## Overview

PM user deletes a main item (cascading to its sub-items) or an individual sub-item, with confirmation dialog and status_history audit trail.

## Setup

- PM user is logged in with main_item:delete and sub_item:delete permissions
- A main item exists with at least 2 sub-items
- A standalone sub-item exists under a different main item
- A member-role user exists without delete permissions

## Happy Path

### Step 1: Delete main item with cascading sub-items
<!-- surface: web -->

**Precondition**: A main item exists with 3 sub-items <!-- fact: prd-spec Story 3 AC1 -->

**User Action**: PM user clicks the delete button on the main item

**Expected Result**: A confirmation dialog appears showing the sub-item count; after confirming, the main item and all 3 sub-items are soft-deleted; a status_history audit record is created; the page updates to remove the deleted item

### Step 2: Delete individual sub-item
<!-- surface: web -->

**Precondition**: A sub-item exists under a main item <!-- fact: prd-spec Story 3 AC2 -->

**User Action**: PM user clicks the delete button on the sub-item detail view

**Expected Result**: A confirmation dialog appears; after confirming, the sub-item is soft-deleted; a status_history audit record is created; the parent item's sub-item list updates; the parent's completion percentage is recalculated

### Step 3: Non-PM user sees no delete button
<!-- surface: web -->

**Precondition**: A member-role user is viewing the same item <!-- fact: prd-spec Story 3 AC3 -->

**User Action**: Member-role user views a main item or sub-item detail page

**Expected Result**: No delete button is visible on the page

## Edge Cases

### Step E1: Cancel deletion of main item
<!-- surface: web -->

**Precondition**: A main item with sub-items is displayed; the delete confirmation dialog is open

**User Action**: PM user clicks cancel on the confirmation dialog

**Expected Result**: No data is deleted; the main item and all sub-items remain unchanged; the dialog closes

### Step E2: Concurrent deletion during sub-item move
<!-- surface: web, api -->

**Precondition**: Another PM user is moving a sub-item out of the main item while the current user has the delete confirmation dialog open

**User Action**: PM user confirms deletion of the main item

**Expected Result**: The delete transaction completes; if the move operation was in progress, it fails with an error indicating the source main item no longer exists; no data corruption occurs

### Step E3: Delete last sub-item of a main item
<!-- surface: web -->

**Precondition**: A main item has exactly 1 remaining sub-item

**User Action**: PM user deletes the last sub-item

**Expected Result**: The sub-item is soft-deleted; the main item still exists with zero sub-items; a status_history audit record is created

### Step E4: Unauthorized deletion attempt (API)
<!-- surface: api -->

**Precondition**: A user without main_item:delete permission sends a delete request <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The user sends a delete API request for a main item

**Expected Result**: The API returns an authorization error; no data is modified; no audit record is created

### Step E5: Unauthenticated deletion attempt (API)
<!-- surface: api -->

**Precondition**: A delete API request is sent without valid credentials <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: A delete API request is sent without a valid authentication token

**Expected Result**: The API returns an authentication error; no data is modified

### Step E6: Delete non-existent item (API)
<!-- surface: api -->

**Precondition**: The item ID in the delete request does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: A delete API request is sent with a non-existent item ID

**Expected Result**: The API returns a "not found" error

### Step E7: Delete transaction failure
<!-- surface: web -->

**Precondition**: An unexpected database error occurs during the delete transaction

**User Action**: PM user confirms deletion

**Expected Result**: The transaction is rolled back; no data is deleted; an error message is displayed to the user; data remains unchanged

### Step E8: Session expired during deletion (Web)
<!-- surface: web -->

**Precondition**: The user's session has expired while the deletion confirmation dialog is open <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: PM user confirms the deletion

**Expected Result**: The user is redirected to the login page; after re-authenticating, the item still exists (no deletion was performed)

### Step E9: Validation error on invalid item ID (Web)
<!-- surface: web -->

**Precondition**: The item URL contains a malformed or invalid item identifier <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: PM user navigates to a delete action with an invalid item identifier

**Expected Result**: An error message is displayed indicating the item identifier is invalid; no deletion is attempted

## Journey Invariants

- Deletion from the web UI always requires a confirmation dialog before executing
- Main item deletion always cascades to all sub-items atomically (the operation either fully completes or fully rolls back)
- Every successful deletion creates a status_history audit record documenting the deletion event
- The delete button is only visible to users with appropriate permissions (main_item:delete, sub_item:delete)
- Deletion is always soft-delete — data is marked as deleted but never permanently removed
