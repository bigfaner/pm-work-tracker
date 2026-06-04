---
journey: "list-filtering-and-sorting"
step: 7
step-action: "API list with multi-status filter"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#11)
---

# Contract: list-filtering-and-sorting / Step 7: API list with multi-status filter

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Multi-status filter returns matching items | api-handbook Enhanced Query Parameters | explicit |
| Terminal items sort to bottom in API response | prd-spec #11 | explicit |
| main_item:read permission required | api-handbook Auth | inferred |

## Outcome "success"
<!-- surface: api -->
- Preconditions: "Main items exist in multiple statuses; user has main_item:read permission"
- Input: "An authenticated API request to the main items endpoint with status query parameter set to multiple statuses"
- Output: "Only main items matching the specified statuses are returned; terminal status main items are sorted to the bottom of the response list"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Derived from API surface unauthorized mandatory outcome -->
- Preconditions: "An unauthenticated user sends a list API request"
- Input: "An API request without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found-team"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The team ID in the request does not exist in the database"
- Input: "An API request with a non-existent team identifier"
- Output: "The API returns a not-found error response"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E6 describes invalid filter parameter values for API -->
- Preconditions: "A list API request contains invalid status filter values"
- Input: "An API request with an invalid status value"
- Output: "The API returns a validation error response describing the invalid parameter; no filtered results are returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible
- Assignee filter always penetrates to sub-item level
