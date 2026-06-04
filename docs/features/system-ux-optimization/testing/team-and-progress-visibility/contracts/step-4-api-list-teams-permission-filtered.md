---
journey: "team-and-progress-visibility"
step: 4
step-action: "API list teams with permission filtering"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/team-and-progress-visibility/journey.md
---

# Contract: team-and-progress-visibility / Step 4: API list teams with permission filtering

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "An authenticated user has permission to access Team A and Team B only"
- Input: "An authenticated API request to GET /api/v1/teams"
- Output: "Only Team A and Team B are returned; Team C is not included in the response"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_TEAM_LIST requires authentication but no specific permission code; unauthenticated requests return 401 -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request to GET /api/v1/teams without a valid authentication token"
- Output: "The API returns HTTP 401 with error code UNAUTHORIZED; no team data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; invalid parameters -->
- Preconditions: "An API request contains invalid parameter values"
- Input: "An API request with a non-numeric team ID format"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR describing the invalid parameter; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Team selector always filters to only teams the current user has permission to access
- API endpoints enforce the same permission checks as the UI
