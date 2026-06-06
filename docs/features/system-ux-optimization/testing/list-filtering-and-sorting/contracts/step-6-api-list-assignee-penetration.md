---
journey: "list-filtering-and-sorting"
step: 6
step-action: "API list with assignee filter and penetration"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#10)
---

# Contract: list-filtering-and-sorting / Step 6: API list with assignee filter and penetration

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Assignee filter penetrates to sub-items via API | api-handbook Enhanced List Main Items | explicit |
| Response includes matchType direct and indirect | api-handbook Enhanced List Main Items | explicit |
| Indirect items include matchedSubItemIds | api-handbook matchedSubItemIds | explicit |
| main_item:read permission required | api-handbook Auth | inferred |

## Outcome "success"
<!-- surface: api -->
- Preconditions: "A valid team exists with main items assigned to assignee A directly and via sub-items; user has main_item:read permission"
- Input: "An authenticated API request to the main items endpoint with assignee filter parameter set to assignee A"
- Output: "The response includes main items where A is the direct assignee and main items where A is the assignee of a sub-item; indirect items include the identifiers of the matching sub-items"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Enhanced List Main Items requires main_item:read permission -->
- Preconditions: "An authenticated user without main_item:read permission attempts to access the list API"
- Input: "The user sends a list API request without proper authorization"
- Output: "The API returns an authorization error response; no item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthenticated"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E5 distinguishes unauthenticated (no token) from unauthorized (wrong permissions) -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; invalid filter parameter values -->
- Preconditions: "A list API request contains invalid filter values such as non-existent status or malformed assignee identifier"
- Input: "An API request with an invalid status value"
- Output: "The API returns a validation error response describing the invalid parameter; no filtered results are returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Terminal status main items always sort to the bottom of response lists when visible
