---
journey: "task-status-transition"
step: 7
step-action: "Submit todo-to-main-item conversion successfully"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 7: Submit todo-to-main-item conversion successfully

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A todo item exists that can be converted to a main item; the todo-to-main-item conversion form is open with all required fields filled"
- Input: "PM user submits the todo-to-main-item conversion form"
- Output: "Conversion succeeds; page updates to reflect the new main item; all form fields are cleared"
- State: "New main item is created in the database with auto-generated code and snowflake bizKey"
- Side-effect: "none"

## Outcome "required-fields-missing"
<!-- source: inferred -->
<!-- reasoning: Journey E3 describes missing required fields scenario; form validation prevents submission -->
- Preconditions: "A todo item exists; the conversion form is open without assignee or priority selected"
- Input: "PM user opens the todo-to-main-item conversion form without selecting assignee or priority"
- Output: "Submit button is disabled; assignee and priority field labels display required markers"
- State: "No new item created"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table PERMISSION_MIDDLEWARE shows permission check; item_pool:review required for conversion -->
- Preconditions: "A user without conversion permission submits the form"
- Input: "Member-role user submits a conversion form API request"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN"
- State: "No item created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while filling the conversion form"
- Input: "PM user submits the form after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the form data is not preserved"
- State: "No item created"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome -->
- Preconditions: "A direct API request is sent for conversion with missing required fields"
- Input: "An API request submits a conversion form with empty assignee or priority fields"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR listing the missing fields; no item is created"
- State: "No item created"
- Side-effect: "none"

## Journey Invariants

- All conversion forms clear all fields on close/cancel or successful submission
- Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)
