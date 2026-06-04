---
journey: "task-status-transition"
step: 4
step-action: "Open conversion form with defaults and validation"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 4: Open conversion form with defaults and validation

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A todo item exists that can be converted to a sub-item"
- Input: "PM user opens the todo-to-sub-item conversion form"
- Output: "Edit dialog renders with all existing fields populated; description field is disabled and greyed out; start date defaults to today; assignee and priority fields show required markers; submit button is disabled until both required fields are filled"
- State: "No state change; form is displayed with default values"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before opening the conversion form"
- Input: "PM user attempts to open the conversion form after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome; form fields have required constraints -->
- Preconditions: "The conversion form is open but required fields are empty"
- Input: "PM user attempts to submit without filling assignee and priority"
- Output: "Submit button is disabled; assignee and priority field labels display required markers"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)
- Description field in the todo-to-sub-item conversion form is always disabled and cannot be modified
