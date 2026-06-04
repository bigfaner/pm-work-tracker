---
journey: "sub-item-move"
step: 3
step-action: "Verify sub-item in new location"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 7)
  - docs/features/system-ux-optimization/prd/prd-spec.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Move Sub Item)
---

# Contract: sub-item-move / Step 3: Verify sub-item in new location

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Moved sub-item appears under new parent with new code | prd-spec Story 7 AC1 | explicit |
| Sub-item no longer appears under old parent | prd-spec Story 7 AC1 | inferred |
| Code number regenerated automatically | journey invariant | explicit |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "The sub-item has been moved to main item B in a previous step; main item B is accessible"
- Input: "PM user navigates to main item B's detail page"
- Output: "The moved sub-item appears in the sub-item list under main item B with its new code number; the sub-item no longer appears under main item A"
- State: "Sub-item is correctly associated with main item B with updated code"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before navigating to verify the move"
- Input: "PM user navigates to main item B's detail page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "moved-sub-item-no-longer-visible"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Another user moved the sub-item again during navigation; read-only verification edge case -->
- Preconditions: "The sub-item was moved to main item B but another user moved it elsewhere before the current user navigated to verify"
- Input: "PM user navigates to main item B's detail page"
- Output: "The sub-item is not visible under main item B; it is no longer at this location"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Status and assignee of the moved sub-item are never changed by the move operation
- Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)
- Closed/completed main items are never valid targets for sub-item moves
- Moving to the same parent main item is always rejected
