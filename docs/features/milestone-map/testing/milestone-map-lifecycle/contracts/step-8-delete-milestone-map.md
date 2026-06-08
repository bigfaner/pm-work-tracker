---
journey: "milestone-map-lifecycle"
step: 8
step-action: "Delete milestone map"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 8: Delete milestone map

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone map is in planning status; user has milestone:delete permission"
- Input: "PM triggers delete action and confirms"
- Output: "Milestone map and all its milestones are soft-deleted; page redirects to list view"
- State: "MilestoneMap soft-deleted; all child milestones soft-deleted; all linked MI milestone_keys cleared; all in single transaction"
- Side-effect: "Cascade soft-delete of all milestones and unbinding of all MainItems in single transaction"

## Outcome "non-deletable-status"
- Preconditions: "Milestone map is in executing or completed status"
- Input: "PM views the detail page"
- Output: "Delete action is not displayed"
- State: "No change to milestone map"
- Side-effect: "none"

## Outcome "no-permission"
- Preconditions: "PM does not have milestone:delete permission"
- Input: "PM views the detail page of a planning status milestone map"
- Output: "Delete action is not displayed"
- State: "No change to milestone map"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated DELETE to /api/v1/teams/:teamId/milestone-maps/:mapId"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is modified"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for session-dependent steps -->
- Preconditions: "User session has expired while delete confirmation dialog is open"
- Input: "PM attempts to confirm deletion"
- Output: "User is redirected to the login page; no data is modified"
- State: "MilestoneMap not deleted"
- Side-effect: "none"

## Outcome "api-not-found"
<!-- source: surface-required (API surface) -->
- Preconditions: "Request targets a non-existent milestone map ID"
- Input: "Authenticated DELETE to /api/v1/teams/:teamId/milestone-maps/{non-existent-id}"
- Output: "API returns 404 not-found error"
- State: "No data is modified"
- Side-effect: "none"

## Journey Invariants

- A milestone map can only be deleted when it is in planning, reviewed, or ready status.
- When deleted, all milestones are soft-deleted and all linked MainItems are unbound in the same transaction.
- All mutation operations require their respective RBAC permissions.
