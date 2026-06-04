---
journey: "list-filtering-and-sorting"
step: 6
step-action: "API list with assignee filter and penetration"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
---

# Contract: list-filtering-and-sorting / Step 6: API list with assignee filter and penetration

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A valid team exists with main items assigned to assignee A directly and via sub-items; user has main_item:read permission"
- Input: "An authenticated API request to GET /api/v1/teams/:teamId/main-items with assigneeKey query parameter set to assignee A's bizKey"
- Output: "The response includes main items with matchType direct (where A is the direct assignee) and main items with matchType indirect (where A is the assignee of a sub-item); indirect items include matchedSubItemIds listing the matching sub-item IDs"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_MAIN_ITEM_LIST requires main_item:read permission; API surface unauthorized mandatory outcome -->
- Preconditions: "An authenticated user without main_item:read permission attempts to access the list API"
- Input: "The user sends GET /api/v1/teams/:teamId/main-items without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; invalid filter parameter values -->
- Preconditions: "A list API request contains invalid filter values such as non-existent status or malformed assigneeKey"
- Input: "An API request with an invalid status value"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR describing the invalid parameter; no filtered results are returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Terminal status main items always sort to the bottom of response lists when visible
