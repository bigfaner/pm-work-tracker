---
journey: "task-status-transition"
step: 5
step-action: "Submit conversion form successfully"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/task-status-transition/journey.md
---

# Contract: task-status-transition / Step 5: Submit conversion form successfully

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "The todo-to-sub-item conversion form is open with assignee and priority filled in"
- Input: "PM user submits the conversion form with all required fields populated"
- Output: "Conversion succeeds; page updates to reflect the new sub-item; all form fields are cleared"
- State: "New sub-item is created in the database with a snowflake bizKey and auto-generated code"
- Side-effect: "Linkage evaluation triggered on the parent main item"

## Outcome "backend-validation-failure"
<!-- source: inferred -->
<!-- reasoning: Fact Table ERROR_VALIDATION shows backend returns validation errors for invalid data; journey E4 describes backend rejection scenario -->
- Preconditions: "All required fields are filled but the backend rejects the submission due to business rule violation"
- Input: "PM user submits the conversion form with data that fails backend validation"
- Output: "Form fields retain their values; an error message is displayed indicating the specific validation failure; the user can correct and retry"
- State: "No new item created; form data preserved"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table PERMISSION_MIDDLEWARE shows permission check; sub_item:create required for creating sub-items -->
- Preconditions: "A user without the required permission sends a conversion form submission"
- Input: "Member-role user submits a conversion form API request"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN"
- State: "No item created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while they are filling the conversion form"
- Input: "PM user submits the conversion form after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the previous unsaved data is not preserved"
- State: "No item created"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome -->
- Preconditions: "The conversion form is open with empty assignee or priority field"
- Input: "PM user submits the form with an empty required field"
- Output: "An error message is displayed near the empty field indicating it is required; the form is not submitted"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- All conversion forms clear all fields on close/cancel or successful submission
- Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)
