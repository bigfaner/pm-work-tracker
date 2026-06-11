---
journey: "milestone-lifecycle"
step: 6
step-action: "Delete milestone"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 6: Delete milestone

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone is in not_started status; user has milestone:delete permission"
- Input: "PM triggers Delete Milestone action and confirms"
- Output: "Milestone soft-deleted; associated MainItems unbound; panel closes; timeline refreshes"
- State: "Milestone soft-deleted; MI milestone_keys cleared; all in single transaction"
- Side-effect: "Unbind all associated MainItems within delete transaction"

## Outcome "delete-cancelled"
- Preconditions: "Milestone is in cancelled status; PM has milestone:delete permission"
- Input: "PM triggers Delete Milestone and confirms"
- Output: "Milestone soft-deleted successfully"
- State: "Milestone soft-deleted"
- Side-effect: "none"

## Outcome "non-deletable-status"
- Preconditions: "Milestone is in in_progress or completed status"
- Input: "PM views the detail panel"
- Output: "Delete action is not displayed"
- State: "No change to milestone"
- Side-effect: "none"

## Outcome "no-permission"
- Preconditions: "PM does not have milestone:delete permission"
- Input: "PM views the detail panel"
- Output: "Delete action is not displayed"
- State: "No change to milestone"
- Side-effect: "none"

## Outcome "api-not-found"
<!-- source: surface-required (API surface) -->
- Preconditions: "Request targets a non-existent milestone ID"
- Input: "Authenticated request to GET /api/v1/teams/:teamId/milestones/{non-existent-id}"
- Output: "API returns 404 not-found error"
- State: "No data modified"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated DELETE to /api/v1/teams/:teamId/milestones/:milestoneId"
- Output: "API returns authentication error with HTTP 401"
- State: "No data modified"
- Side-effect: "none"

## Journey Invariants

- Delete is only available for milestones in not_started or cancelled status.
- When deleted, associated MainItems are unbound within the same transaction.
- All mutation operations require their respective RBAC permissions.
