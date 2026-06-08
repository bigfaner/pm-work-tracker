---
journey: "milestone-map-lifecycle"
step: 2
step-action: "Edit milestone map information"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
---

# Contract: milestone-map-lifecycle / Step 2: Edit milestone map information

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone map exists; edit dialog is open with pre-filled current values; user has milestone:update permission"
- Input: "PM modifies the name and description in the edit dialog, then saves"
- Output: "Changes are saved immediately; edit dialog closes; detail page refreshes to show updated information"
- State: "MilestoneMap record updated with new name and description values"
- Side-effect: "none"

## Outcome "no-changes"
- Preconditions: "Edit dialog is open with pre-filled current values"
- Input: "PM saves without making any modifications"
- Output: "Dialog closes as a no-op, equivalent to Cancel"
- State: "MilestoneMap record unchanged"
- Side-effect: "none"

## Outcome "validation-error-invalid-date-range"
- Preconditions: "Edit dialog is open; PM modifies plan end date to be earlier than plan start date"
- Input: "PM sets invalid date range and submits"
- Output: "Form displays date validation error and does not submit"
- State: "MilestoneMap record unchanged; dialog remains open"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for form submission steps -->
- Preconditions: "User session has expired while the edit form is open"
- Input: "PM submits the edit form"
- Output: "User is redirected to the login page; no data is modified"
- State: "MilestoneMap retains original values; after re-authentication, form retains no pending changes"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestone-maps/:mapId"
- Output: "API returns authentication error response with HTTP 401"
- State: "No data is returned or modified"
- Side-effect: "none"

## Journey Invariants

- A milestone map can only be deleted when it is in planning, reviewed, or ready status.
- All mutation operations require their respective RBAC permissions.
- Edit forms display loading state and prevent further interaction during submission.
