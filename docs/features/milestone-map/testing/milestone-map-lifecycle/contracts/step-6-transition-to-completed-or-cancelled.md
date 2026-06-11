---
journey: "milestone-map-lifecycle"
step: 6
step-action: "Transition status to completed or cancelled"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 6: Transition status to completed or cancelled

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success-completed"
- Preconditions: "Milestone map is in executing status; all milestones under this map are in terminal states (completed or cancelled); user has milestone:update permission"
- Input: "PM selects the Completed status option"
- Output: "Status changes from executing to completed; this is a terminal state with no further transitions"
- State: "MilestoneMap status updated to completed; no further status transitions available"
- Side-effect: "none"

## Outcome "success-cancelled"
- Preconditions: "Milestone map is in executing status and has milestones with associated MainItems; user has milestone:update permission"
- Input: "PM selects the Cancelled status option"
- Output: "Status changes to cancelled; all milestones cancelled in cascade; all associated MainItems unbound"
- State: "MilestoneMap status updated to cancelled (terminal); all non-terminal milestones set to cancelled; all MI milestone_keys cleared"
- Side-effect: "Cascade: cancel non-terminal milestones + unbind all MIs in single transaction"

## Outcome "completed-with-incomplete-milestones"
- Preconditions: "Milestone map is in executing status; at least one milestone is not in a terminal state"
- Input: "PM attempts to transition to Completed"
- Output: "Status change is rejected with error message indicating not all milestones are in terminal states"
- State: "MilestoneMap status unchanged"
- Side-effect: "none"

## Outcome "completed-is-terminal"
- Preconditions: "Milestone map is in completed status"
- Input: "PM views the status options"
- Output: "No transition options are available; completed is terminal"
- State: "MilestoneMap status remains completed"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestone-maps/:mapId/status"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is modified"
- Side-effect: "none"

## Journey Invariants

- Transitioning to completed requires all milestones under the map to be in terminal states.
- Any non-terminal state can transition to cancelled, which is a terminal state.
- Cancelled cascade: all non-terminal milestones cancelled + all MIs unbound in single transaction.
- completed and cancelled are terminal states with no further transitions.
