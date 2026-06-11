---
journey: "milestone-lifecycle"
step: 3
step-action: "Transition status from not_started to in_progress"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 3: Transition status from not_started to in_progress

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone is in not_started status; parent map is not terminal; user has milestone:update permission"
- Input: "PM selects the In Progress status option in the detail panel"
- Output: "Status changes from not_started to in_progress"
- State: "Milestone status updated to in_progress"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId/status"
- Output: "API returns authentication error with HTTP 401"
- State: "No data modified"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome -->
- Preconditions: "User session has expired"
- Input: "PM attempts to change status"
- Output: "User redirected to login page; no data modified"
- State: "Milestone retains original status"
- Side-effect: "none"

## Journey Invariants

- Status transitions follow defined state machine: not_started -> in_progress -> completed.
- All mutation operations require their respective RBAC permissions.
