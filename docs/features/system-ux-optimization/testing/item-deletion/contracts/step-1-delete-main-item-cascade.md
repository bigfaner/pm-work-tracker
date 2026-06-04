---
journey: "item-deletion"
step: 1
step-action: "Delete main item with cascading sub-items"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/item-deletion/journey.md
---

# Contract: item-deletion / Step 1: Delete main item with cascading sub-items

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A main item exists with 3 sub-items; PM user has main_item:delete permission"
- Input: "PM user clicks the delete button on the main item and confirms the deletion in the confirmation dialog"
- Output: "A confirmation dialog appears showing the sub-item count; after confirming, the main item and all 3 sub-items are soft-deleted; the page updates to remove the deleted item"
- State: "Main item and all sub-items are soft-deleted (deleted_flag set); status_history audit records created for each deleted item"
- Side-effect: "Cascade soft-delete executes atomically in a single transaction"

## Outcome "cancelled"
- Preconditions: "A main item with sub-items is displayed; the delete confirmation dialog is open"
- Input: "PM user clicks cancel on the confirmation dialog"
- Output: "No data is deleted; the main item and all sub-items remain unchanged; the dialog closes"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_MAIN_ITEM_DELETE requires main_item:delete permission -->
- Preconditions: "A user without main_item:delete permission sends a delete request"
- Input: "The user sends a DELETE API request for a main item"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no data is modified; no audit record is created"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the deletion confirmation dialog is open"
- Input: "PM user confirms the deletion after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the item still exists"
- State: "No state change; item not deleted"
- Side-effect: "none"

## Outcome "not-found"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The item ID in the delete request does not exist in the database"
- Input: "A DELETE API request is sent with a non-existent item ID"
- Output: "The API returns HTTP 404 with error code ITEM_NOT_FOUND"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Deletion from the web UI always requires a confirmation dialog before executing
- Main item deletion always cascades to all sub-items atomically
- Every successful deletion creates a status_history audit record
- Deletion is always soft-delete
