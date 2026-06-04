---
journey: "list-filtering-and-sorting"
step: 7
step-action: "API list with multi-status filter"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
---

# Contract: list-filtering-and-sorting / Step 7: API list with multi-status filter

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Main items exist in multiple statuses; user has main_item:read permission"
- Input: "An authenticated API request to GET /api/v1/teams/:teamId/main-items with status query parameter set to multiple statuses"
- Output: "Only main items matching the specified statuses are returned; terminal status main items are sorted to the bottom of the response list"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface unauthorized mandatory outcome -->
- Preconditions: "An unauthenticated user sends a list API request"
- Input: "An API request without a valid authentication token"
- Output: "The API returns HTTP 401 with error code UNAUTHORIZED; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found-team"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "The team ID in the request does not exist in the database"
- Input: "An API request to GET /api/v1/teams/<non-existent-teamId>/main-items"
- Output: "The API returns HTTP 404 with error code TEAM_NOT_FOUND"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible
- Assignee filter always penetrates to sub-item level
