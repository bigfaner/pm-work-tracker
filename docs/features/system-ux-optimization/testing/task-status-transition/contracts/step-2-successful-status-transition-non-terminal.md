---
journey: "task-status-transition"
step: 2
step-action: "Successful status transition (non-terminal)"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 2: Successful status transition (non-terminal)

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "An item exists that is eligible for a non-terminal status transition, and no concurrent modification has occurred"
- Input: "PM user clicks a status transition button targeting a valid non-terminal destination status"
- Output: "The item's status is updated to the new value; the page reflects the new status"
- State: "Item status updated in database; status history record created documenting the transition"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table PERMISSION_MIDDLEWARE shows permission check per route; main_item:change_status required -->
- Preconditions: "A user without main_item:change_status permission attempts a status transition via API"
- Input: "An API request to PUT /api/v1/teams/:teamId/main-items/:itemId/status without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN"
- State: "No state change; no status history record created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user clicks a status transition button after session expiry"
- Output: "The user is redirected to the login page"
- State: "Item status unchanged"
- Side-effect: "none"

## Journey Invariants

- Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips
- Required fields are enforced at both UI level and API level
