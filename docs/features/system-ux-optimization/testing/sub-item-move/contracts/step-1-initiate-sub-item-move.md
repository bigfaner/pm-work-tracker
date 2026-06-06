---
journey: "sub-item-move"
step: 1
step-action: "Initiate sub-item move"
surface_types: ["web", "api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 7)
  - docs/features/system-ux-optimization/prd/prd-spec.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Move Sub Item)
---

# Contract: sub-item-move / Step 1: Initiate sub-item move

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Target selector shows non-closed, non-completed items | prd-spec Story 7 AC1 | explicit |
| sub_item:update auth required for move | api-handbook.md Move Sub Item Auth | explicit |
| Closed/completed items excluded from selector | journey invariant | explicit |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "PM user is viewing a sub-item detail page that belongs to main item A; at least one other non-terminal main item exists; user has sub_item:update permission"
- Input: "PM user selects the move to another main item action"
- Output: "A target main item selector is displayed listing available non-closed non-completed main items excluding the current parent"
- State: "No state change; move dialog is displayed"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before initiating the move"
- Input: "PM user selects the move action after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item requires sub_item:update permission -->
- Preconditions: "A user without sub_item:update permission attempts the move action via API"
- Input: "An unauthorized move API request referencing the sub-item"
- Output: "The API returns an authorization error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Status and assignee of the moved sub-item are never changed by the move operation
- Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)
- Closed/completed main items are never valid targets for sub-item moves
- Moving to the same parent main item is always rejected
