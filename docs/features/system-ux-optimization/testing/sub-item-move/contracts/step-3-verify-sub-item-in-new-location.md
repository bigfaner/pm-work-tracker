---
journey: "sub-item-move"
step: 3
step-action: "Verify sub-item in new location"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
---

# Contract: sub-item-move / Step 3: Verify sub-item in new location

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "The sub-item has been moved to main item B in a previous step"
- Input: "PM user navigates to main item B's detail page"
- Output: "The moved sub-item appears in the sub-item list under main item B with its new code number; the sub-item no longer appears under main item A"
- State: "Sub-item is correctly associated with main item B with updated code"
- Side-effect: "none"

## Outcome "source-deleted-during-move"
<!-- source: inferred -->
<!-- reasoning: Journey E4 describes source main item deleted while move dialog is open; concurrent operation boundary -->
- Preconditions: "Another user has soft-deleted main item A while the current user was performing the move"
- Input: "PM user confirms the move after the source was deleted"
- Output: "An error message is displayed indicating the source main item no longer exists; no move is performed"
- State: "No state change"
- Side-effect: "none"

## Outcome "non-existent-sub-item-api"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The sub-item ID in the request does not exist in the database"
- Input: "A move API request is sent with a non-existent sub-item ID"
- Output: "The API returns HTTP 404 with error code ITEM_NOT_FOUND; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "non-existent-target-api"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The target main item ID in the request does not exist in the database"
- Input: "A move API request is sent with a non-existent target main item ID"
- Output: "The API returns HTTP 404 with error code ITEM_NOT_FOUND; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; missing or invalid target ID -->
- Preconditions: "A move API request is sent with missing or invalid fields"
- Input: "An API request to the move endpoint with an invalid or missing target ID"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR listing the missing or invalid fields; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Status and assignee of the moved sub-item are never changed by the move operation
- Moving to the same parent main item is always rejected
