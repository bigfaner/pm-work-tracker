---
journey: "milestone-map-lifecycle"
step: 5
step-action: "Transition status from ready to executing"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 5: Transition status from ready to executing

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone map is in ready status; user has milestone:update permission"
- Input: "PM selects the In Progress status option"
- Output: "Status changes from ready to executing"
- State: "MilestoneMap status updated from ready to executing"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestone-maps/:mapId/status"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is modified"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome -->
- Preconditions: "User session has expired"
- Input: "PM attempts to change the status"
- Output: "User is redirected to the login page; no data is modified"
- State: "MilestoneMap retains original status"
- Side-effect: "none"

## Journey Invariants

- Status transitions follow the defined state machine.
- All mutation operations require their respective RBAC permissions.
