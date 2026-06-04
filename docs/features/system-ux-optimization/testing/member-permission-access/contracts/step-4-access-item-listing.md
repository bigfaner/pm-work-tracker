---
journey: "member-permission-access"
step: 4
step-action: "Access item listing page"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
---

# Contract: member-permission-access / Step 4: Access item listing page

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A member user is logged in with main_item:list permission"
- Input: "Member user navigates to the item listing page"
- Output: "Page loads successfully; items within the user's team are displayed; no permission errors occur"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_MAIN_ITEM_LIST requires main_item:read permission -->
- Preconditions: "A member user without main_item:read permission attempts to access the list API"
- Input: "An API request to GET /api/v1/teams/:teamId/main-items without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "A member user is authenticated; the requested resource ID does not exist"
- Input: "The member user sends an API request for a non-existent item ID"
- Output: "The API returns HTTP 404 with error code ITEM_NOT_FOUND"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- API endpoints enforce the same permission checks as the UI
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login
