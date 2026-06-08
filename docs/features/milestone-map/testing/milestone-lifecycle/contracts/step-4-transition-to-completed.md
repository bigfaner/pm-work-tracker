---
journey: "milestone-lifecycle"
step: 4
step-action: "Transition status to completed"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 4: Transition status to completed

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone is in in_progress status; all associated MainItems are in terminal states (completed or closed); parent map is not terminal; user has milestone:update permission"
- Input: "PM selects the Completed status option"
- Output: "Status changes to completed; completion percentage reflects associated MI average"
- State: "Milestone status updated to completed"
- Side-effect: "none"

## Outcome "incomplete-items"
- Preconditions: "Milestone is in in_progress status; at least one associated MainItem is not in a terminal state"
- Input: "PM attempts to transition to Completed"
- Output: "Status change rejected with error indicating not all items are in terminal states"
- State: "Milestone status unchanged"
- Side-effect: "none"

## Outcome "cancel-cascade"
- Preconditions: "Milestone is in not_started or in_progress status and has associated MainItems; parent map is not terminal; user has milestone:update permission"
- Input: "PM selects the Cancelled status option"
- Output: "Status changes to cancelled; all associated MainItems auto-unbound in same transaction; detail panel shows empty MI list"
- State: "Milestone status set to cancelled (terminal); all MI milestone_keys cleared in transaction"
- Side-effect: "Auto-unbind all associated MainItems within same transaction"

## Outcome "cancelled-is-terminal"
- Preconditions: "Milestone is in cancelled status"
- Input: "PM views the status options"
- Output: "No dropdown menu appears; cancelled is terminal"
- State: "No status change possible"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId/status"
- Output: "API returns authentication error with HTTP 401"
- State: "No data modified"
- Side-effect: "none"

## Journey Invariants

- Cancellation of a milestone automatically unbinds all associated MainItems within the same transaction.
- A milestone can only be marked completed when all its associated MainItems are in terminal states.
- Cancelled is a terminal state with no recovery.
- All mutation operations require their respective RBAC permissions.
