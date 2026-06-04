---
journey: "item-deletion"
step: 2
step-action: "Delete individual sub-item"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/item-deletion/journey.md
---

# Contract: item-deletion / Step 2: Delete individual sub-item

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A sub-item exists under a main item; PM user has sub_item:delete permission"
- Input: "PM user clicks the delete button on the sub-item detail view and confirms in the confirmation dialog"
- Output: "A confirmation dialog appears; after confirming, the sub-item is soft-deleted; the parent item's sub-item list updates; the parent's completion percentage is recalculated"
- State: "Sub-item is soft-deleted (deleted_flag set); status_history audit record created; parent main item completion recalculated"
- Side-effect: "Parent main item completion percentage recalculated; linkage evaluation triggered on parent"

## Outcome "delete-last-sub-item"
<!-- source: inferred -->
<!-- reasoning: Journey E3 describes deleting the last sub-item of a main item; completion recalc with zero sub-items returns 0 -->
- Preconditions: "A main item has exactly 1 remaining sub-item"
- Input: "PM user deletes the last sub-item"
- Output: "The sub-item is soft-deleted; the main item still exists with zero sub-items; a status_history audit record is created"
- State: "Sub-item soft-deleted; main item completion recalculated to 0; status_history audit record created"
- Side-effect: "Parent main item completion recalculated"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_SUB_ITEM_DELETE requires sub_item:delete permission -->
- Preconditions: "A user without sub_item:delete permission sends a delete request"
- Input: "The user sends a DELETE API request for a sub-item"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user confirms sub-item deletion after session expiry"
- Output: "The user is redirected to the login page; the sub-item still exists"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The sub-item ID does not exist in the database"
- Input: "A DELETE API request with a non-existent sub-item ID"
- Output: "The API returns HTTP 404 with error code ITEM_NOT_FOUND"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Deletion from the web UI always requires a confirmation dialog before executing
- Every successful deletion creates a status_history audit record
- The delete button is only visible to users with appropriate permissions
- Deletion is always soft-delete
