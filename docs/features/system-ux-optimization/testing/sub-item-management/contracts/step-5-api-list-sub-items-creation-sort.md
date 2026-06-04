---
journey: "sub-item-management"
step: 5
step-action: "API list sub-items with creation time sort"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#5)
---

# Contract: sub-item-management / Step 5: API list sub-items with creation time sort

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| API returns sub-items ordered by creation time descending | api-handbook Enhanced List Main Items | explicit |
| main_item:read permission required | api-handbook Auth | inferred |

## Outcome "success"
<!-- surface: api -->
<!-- fact: api-handbook Enhanced List Main Items -->
- Preconditions: "An authenticated API request is sent for a main item that has multiple sub-items; user has main_item:read permission"
- Input: "An authenticated API request to the main items endpoint with valid parameters"
- Output: "The response includes sub-items under each main item, ordered by creation time descending (most recently created first)"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook requires main_item:read permission for main items endpoint -->
- Preconditions: "An authenticated user without main_item:read permission sends an API request"
- Input: "The user sends an API request to the main items endpoint without proper authorization"
- Output: "The API returns an authorization error response; no sub-item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthenticated"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E6 -- API request without valid authentication credentials -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E7 -- API request with invalid team ID parameter -->
- Preconditions: "An API request uses a non-numeric value for a team ID parameter"
- Input: "An API request with a non-numeric team ID"
- Output: "The API returns a validation error response listing the invalid parameter; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E8 -- API request for non-existent main item -->
- Preconditions: "The main item ID in the request does not exist in the database"
- Input: "An API request for a non-existent main item"
- Output: "The API returns a not-found error response"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
