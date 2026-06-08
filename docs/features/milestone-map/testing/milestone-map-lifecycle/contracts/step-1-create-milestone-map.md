---
journey: "milestone-map-lifecycle"
step: 1
step-action: "Create milestone map via list page"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 1: Create milestone map via list page

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "User has milestone:create permission; create dialog is open with valid data"
- Input: "PM fills name (1-100 chars), owner (assigneeBizKey), optional plan dates and description, then submits"
- Output: "Milestone map created with status planning; dialog closes; list refreshes showing new entry with correct name, status badge, and owner info"
- State: "New MilestoneMap record persisted with status planning; list view updated to include new entry"
- Side-effect: "none"

## Outcome "validation-error-missing-name"
- Preconditions: "Create dialog is open; name field is empty"
- Input: "PM submits the form without filling in the name field"
- Output: "Form displays a validation error near the name field; form is not submitted; dialog remains open"
- State: "No new milestone map created; form state preserved"
- Side-effect: "none"

## Outcome "validation-error-name-too-long"
- Preconditions: "Create dialog is open; name exceeds 100 characters"
- Input: "PM enters a name with 101 characters and submits"
- Output: "Form displays a validation error about name length"
- State: "No new milestone map created; form state preserved"
- Side-effect: "none"

## Outcome "validation-error-missing-owner"
- Preconditions: "Create dialog is open; owner field is empty"
- Input: "PM fills the name but leaves the owner field empty and submits"
- Output: "Form displays a validation error about the required owner field; form is not submitted"
- State: "No new milestone map created; form state preserved"
- Side-effect: "none"

## Outcome "validation-error-invalid-date-range"
- Preconditions: "Create dialog is open; plan end date is earlier than plan start date"
- Input: "PM sets the plan end date earlier than the plan start date and submits"
- Output: "Form displays a date range validation error; form is not submitted"
- State: "No new milestone map created; form state preserved"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for form submission steps -->
- Preconditions: "User session has expired while the create form is open"
- Input: "PM submits the create form"
- Output: "User is redirected to the login page; no data is modified"
- State: "No new milestone map created; after re-authentication, form retains no pending changes"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is returned or modified"
- Side-effect: "none"

## Journey Invariants

- A milestone map can only be deleted when it is in planning, reviewed, or ready status.
- Status transitions follow the defined state machine: planning -> reviewed -> ready -> executing -> completed, with rollback and cancel transitions.
- All mutation operations require their respective RBAC permissions.
- Create forms display loading state and prevent further interaction during submission.
