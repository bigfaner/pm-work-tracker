---
feature: "system-ux-optimization"
journey: "item-deletion"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 3)
  - docs/features/system-ux-optimization/prd/prd-spec.md
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

## Happy Path

### Step 1: Delete main item with cascading sub-items

**User Action**: PM user clicks the delete button on a main item that has 3 sub-items

**Expected Result**: Confirmation dialog appears with the message "Will also delete 3 sub-items"; after confirming, the main item and all 3 sub-items are soft-deleted in a single transaction; status_history records the deletion event; page updates to remove the deleted item

### Step 2: Delete individual sub-item

**User Action**: PM user clicks the delete button on a sub-item in its detail view

**Expected Result**: Confirmation dialog appears; after confirming, the sub-item is soft-deleted; status_history records the deletion event; sub-item is removed from the parent item's sub-item list

### Step 3: Non-PM user sees no delete button

**User Action**: A member-role user views the same main item or sub-item detail page

**Expected Result**: No delete button is visible on the page

## Edge Cases

### Step 1b: Cancel deletion of main item

**Precondition**: Main item with sub-items is displayed

**User Action**: PM user clicks delete, then clicks cancel on the confirmation dialog

**Expected Result**: No data is deleted; the main item and all sub-items remain unchanged; dialog closes

### Step 2b: Concurrent deletion during sub-item move

**Precondition**: Another PM user is moving a sub-item out of the main item while the current user initiates deletion of the same main item

**User Action**: PM user confirms deletion of the main item

**Expected Result**: The move operation and delete operation are handled within transactions; if the main item is already soft-deleted by the time the move completes, the move fails with "source main item does not exist" error

### Step 3b: Delete last sub-item of a main item

**Precondition**: A main item has exactly 1 remaining sub-item

**User Action**: PM user deletes the last sub-item

**Expected Result**: Sub-item is soft-deleted; main item still exists with zero sub-items; status_history records the deletion

### Step 4b: Unauthorized deletion attempt via API

**Precondition**: A member-role user sends a DELETE request to the item endpoint

**User Action**: Member user sends DELETE /api/main-items/:id

**Expected Result**: API returns 403 Forbidden; no data is modified; deletion audit trail is not created

### Step 5b: Delete transaction failure

**Precondition**: Database constraint or unexpected error occurs during the delete transaction

**User Action**: PM user confirms deletion

**Expected Result**: Transaction is rolled back; no data is deleted; frontend displays an error message; data remains unchanged

## Journey Invariants

- Deletion always requires a confirmation dialog before executing
- Main item deletion always cascades to all sub-items in a single transaction
- Every successful deletion creates a status_history record documenting the deletion event
- Delete button is only visible to PM users with appropriate permissions (main_item:delete, sub_item:delete)
- Deletion is always soft-delete -- data is marked as deleted but never permanently removed
