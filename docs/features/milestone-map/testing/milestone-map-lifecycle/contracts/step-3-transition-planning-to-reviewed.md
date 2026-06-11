---
journey: "milestone-map-lifecycle"
step: 3
step-action: "Transition status from planning to reviewed"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 3: Transition status from planning to reviewed

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone map is in planning status; user has milestone:update permission"
- Input: "PM selects the Reviewed status option on the detail page"
- Output: "Status changes from planning to reviewed; badge updates visually; no error occurs"
- State: "MilestoneMap status updated from planning to reviewed"
- Side-effect: "none"

## Outcome "server-error"
<!-- source: inferred -->
<!-- reasoning: Status transition involves API call; backend may be unavailable -->
- Preconditions: "Backend is unavailable when status transition is attempted"
- Input: "PM selects a new status and the backend returns an error"
- Output: "Error message is displayed; status reverts to original value"
- State: "MilestoneMap status unchanged; detail page does not refresh"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for session-dependent steps -->
- Preconditions: "User session has expired"
- Input: "PM attempts to change the status"
- Output: "User is redirected to the login page; no data is modified"
- State: "MilestoneMap retains original status"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestone-maps/:mapId/status"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is modified"
- Side-effect: "none"

## Journey Invariants

- Status transitions follow the defined state machine: planning -> reviewed -> ready -> executing -> completed.
- All mutation operations require their respective RBAC permissions.
