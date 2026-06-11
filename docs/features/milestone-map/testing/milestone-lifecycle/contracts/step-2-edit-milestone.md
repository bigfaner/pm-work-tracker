---
journey: "milestone-lifecycle"
step: 2
step-action: "Edit milestone information"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 2: Edit milestone information

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone exists; edit dialog is open with pre-filled values; user has milestone:update permission; parent map is not terminal"
- Input: "PM modifies name and plan completion date in the edit dialog, then saves"
- Output: "Changes saved; dialog closes; panel and timeline refresh; node position recalculated based on new date"
- State: "Milestone record updated with new name and date"
- Side-effect: "none"

## Outcome "no-changes"
- Preconditions: "Edit dialog is open with pre-filled current values"
- Input: "PM saves without modifying anything"
- Output: "Dialog closes as a no-op, equivalent to Cancel"
- State: "Milestone record unchanged"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for form submission steps -->
- Preconditions: "User session has expired while edit dialog is open"
- Input: "PM submits the edit form"
- Output: "User is redirected to login page; no data modified"
- State: "Milestone retains original values"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId"
- Output: "API returns authentication error with HTTP 401"
- State: "No data modified"
- Side-effect: "none"

## Journey Invariants

- All mutation operations require their respective RBAC permissions.
- Parent map must not be terminal for status-relevant updates (name, date).
