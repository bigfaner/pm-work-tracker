---
journey: "task-status-transition"
step: 1
step-action: "Trigger status transition with error"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 1: Trigger status transition with error

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "invalid-transition"
- Preconditions: "An item exists in a status that has no valid transition to the selected target status; the item has not been concurrently modified"
- Input: "PM user clicks a status transition button targeting an invalid destination status for the current item"
- Output: "The item's status remains unchanged; an error message is displayed below the action area with content from the backend explaining why the transition cannot be performed"
- State: "Item status is unchanged in the database"
- Side-effect: "none"

## Outcome "concurrent-edit-conflict"
<!-- source: inferred -->
<!-- reasoning: Fact Table STATUS_TRANSITIONS_MAIN shows self-transition check and transition validation; concurrent edit means the current user's transition is based on stale data -->
- Preconditions: "Another user has changed the item's status after the current user loaded the page; the current user's transition is now based on stale data"
- Input: "PM user clicks a status transition button based on stale state"
- Output: "An inline error message displays: the item status has been updated by another user, please refresh and retry; the item's status remains unchanged"
- State: "Item status is unchanged; no status history record is created"
- Side-effect: "none"

## Outcome "unauthorized-attempt"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_MAIN_ITEM_CHANGE_STATUS requires main_item:change_status permission; API returns ERR_FORBIDDEN for unauthorized users -->
- Preconditions: "The user has member role only, which does not include the main_item:change_status permission"
- Input: "Member-role user sends a status transition API request to PUT /api/v1/teams/:teamId/main-items/:itemId/status"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN indicating insufficient permissions"
- State: "No state change; no status history record created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the status transition page is displayed"
- Input: "PM user clicks a status transition button after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the item retains its original status"
- State: "Item status unchanged; no status history record created"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome; the ChangeStatus endpoint validates the request body -->
- Preconditions: "A direct API request is sent with missing or empty status field"
- Input: "An API request to PUT /api/v1/teams/:teamId/main-items/:itemId/status with an empty or missing status field"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR indicating request validation failed"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips
- Required fields are enforced at both UI level and API level
