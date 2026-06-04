---
journey: "sub-item-move"
step: 1
step-action: "Initiate sub-item move"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
---

# Contract: sub-item-move / Step 1: Initiate sub-item move

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "PM user is viewing a sub-item detail page that belongs to main item A; at least one other non-terminal main item exists"
- Input: "PM user selects the move to another main item action"
- Output: "A target main item selector is displayed listing available non-closed non-completed main items excluding the current parent"
- State: "No state change; move dialog is displayed"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before initiating the move"
- Input: "PM user selects the move action after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_SUB_ITEM_MOVE requires sub_item:update permission -->
- Preconditions: "A user without sub_item:update permission attempts the move action via API"
- Input: "An API request to PUT /api/v1/teams/:teamId/sub-items/:subId/move without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Status and assignee of the moved sub-item are never changed by the move operation
