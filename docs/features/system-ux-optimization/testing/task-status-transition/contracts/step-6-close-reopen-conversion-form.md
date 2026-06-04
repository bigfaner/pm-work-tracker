---
journey: "task-status-transition"
step: 6
step-action: "Close and reopen conversion form"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 6: Close and reopen conversion form

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A conversion form has been partially filled with some data entered"
- Input: "PM user closes or cancels the form, then opens any new or conversion form"
- Output: "All fields in the newly opened form are empty with no residual data from the previously closed form"
- State: "No state change to any item"
- Side-effect: "none"

## Outcome "after-successful-submission"
<!-- source: inferred -->
<!-- reasoning: Journey E5 describes reopening form after successful submission; form cleanup invariant applies -->
- Preconditions: "A conversion form was just submitted successfully"
- Input: "PM user opens another new or conversion form"
- Output: "All fields are empty with no residual data from the previously submitted form"
- State: "No state change beyond the prior submission"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before attempting to reopen a form"
- Input: "PM user attempts to open a new form after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- All conversion forms clear all fields on close/cancel or successful submission
- Description field in the todo-to-sub-item conversion form is always disabled and cannot be modified
