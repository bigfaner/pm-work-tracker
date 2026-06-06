---
journey: "task-status-transition"
step: 3
step-action: "Successful status transition (terminal with confirmation)"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 3: Successful status transition (terminal with confirmation)

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "An item exists that is eligible for transition to a terminal status (completed or closed); all sub-items are in terminal status"
- Input: "PM user clicks the terminal status transition button, a confirmation dialog appears, user confirms"
- Output: "The transition completes; the page reflects the new terminal status; completion is set to 100 percent"
- State: "Item status updated to terminal status; completion set to 100; actual_end_date set to current timestamp; status history record created"
- Side-effect: "Parent main item completion percentage recalculated; linkage evaluation triggered"

## Outcome "cancelled"
- Preconditions: "An item is eligible for transition to a terminal status; the confirmation dialog is open"
- Input: "PM user clicks the terminal transition button, confirmation dialog appears, user clicks cancel"
- Output: "No status change occurs; the dialog closes; the item remains in its current status"
- State: "No state change; no status history record created"
- Side-effect: "none"

## Outcome "sub-items-not-terminal"
<!-- source: inferred -->
<!-- reasoning: Fact Table STATUS_TRANSITIONS_MAIN shows terminal guard checks all sub-items are terminal; returns ErrSubItemsNotTerminal -->
- Preconditions: "An item is eligible for terminal transition but has at least one sub-item in non-terminal status"
- Input: "PM user attempts to transition the main item to terminal status"
- Output: "The transition is rejected; an error message indicates all sub-items must be completed or closed before closing the main item"
- State: "No state change; item status remains unchanged"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table PERMISSION_MIDDLEWARE shows permission check; main_item:change_status required -->
- Preconditions: "A user without main_item:change_status permission attempts a terminal transition"
- Input: "Unauthorized API request to change status to terminal"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the confirmation dialog is displayed"
- Input: "PM user confirms the terminal transition after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the item retains its original status"
- State: "Item status unchanged"
- Side-effect: "none"

## Journey Invariants

- Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips
- All conversion forms clear all fields on close/cancel or successful submission
