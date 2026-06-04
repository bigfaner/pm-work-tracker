---
journey: "sub-item-management"
step: 5
step-action: "API list sub-items with creation time sort"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Step 5: API list sub-items with creation time sort

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "An authenticated API request is sent for a main item that has multiple sub-items; user has main_item:read permission"
- Input: "An authenticated API request to GET /api/v1/teams/:teamId/main-items with valid parameters"
- Output: "The response includes sub-items under each main item, ordered by creation time descending (most recently created first)"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_MAIN_ITEM_LIST requires main_item:read permission -->
- Preconditions: "An authenticated user without main_item:read permission sends an API request"
- Input: "The user sends an API request to the main items endpoint without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no sub-item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; invalid team ID parameter -->
- Preconditions: "An API request uses a non-numeric value for a team ID parameter"
- Input: "An API request with a non-numeric team ID"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR listing the invalid parameter; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
