---
journey: "milestone-lifecycle"
step: 1
step-action: "Create milestone in timeline view"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 1: Create milestone in timeline view

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Milestone map exists and is accessible; parent map is not terminal; user has milestone:create permission; create dialog is open with valid data"
- Input: "PM fills name (1-100 chars), plan completion date (required), and optional description, then confirms"
- Output: "Milestone created with status not_started and completion 0; dialog closes; timeline refreshes showing new node at correct date position"
- State: "New Milestone record persisted with status not_started, completion 0, under parent map"
- Side-effect: "none"

## Outcome "validation-error-missing-name"
- Preconditions: "Create dialog is open; name field is empty"
- Input: "PM submits without filling the name field"
- Output: "Form displays validation error about required name field; form not submitted"
- State: "No new milestone created"
- Side-effect: "none"

## Outcome "validation-error-name-too-long"
- Preconditions: "Create dialog is open; name exceeds 100 characters"
- Input: "PM enters name with 101 characters and submits"
- Output: "Form displays validation error about name length"
- State: "No new milestone created"
- Side-effect: "none"

## Outcome "validation-error-missing-date"
- Preconditions: "Create dialog is open; plan completion date is empty"
- Input: "PM fills name but leaves plan completion date empty and submits"
- Output: "Form displays validation error about required date field; form not submitted"
- State: "No new milestone created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Web surface mandatory session-expired outcome for form submission steps -->
- Preconditions: "User session has expired while create dialog is open"
- Input: "PM submits the create form"
- Output: "User is redirected to login page; no data modified"
- State: "No new milestone created; after re-authentication, milestone retains original values"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps/:mapId/milestones"
- Output: "API returns authentication error with HTTP 401"
- State: "No data returned or modified"
- Side-effect: "none"

## Outcome "terminal-parent-map"
<!-- source: inferred -->
<!-- reasoning: Fact table shows BR-5 blocks milestone creation under terminal maps (milestone_service.go:61-63) -->
- Preconditions: "Parent MilestoneMap is in terminal state (completed or cancelled)"
- Input: "Authenticated POST to create milestone under this map"
- Output: "API returns error indicating parent map is terminal"
- State: "No milestone created"
- Side-effect: "none"

## Outcome "duplicate-name"
<!-- source: inferred -->
<!-- reasoning: Fact table shows duplicate name check within same map (milestone_service.go:66-72) -->
- Preconditions: "A milestone with the same name already exists in the same map"
- Input: "Authenticated POST with duplicate milestone name"
- Output: "API returns 409 conflict error indicating name already exists"
- State: "No milestone created"
- Side-effect: "none"

## Journey Invariants

- Cancellation of a milestone automatically unbinds all associated MainItems within the same transaction.
- A milestone can only be marked completed when all its associated MainItems are in terminal states.
- Cancelled milestones cannot receive new MainItem bindings.
- All mutation operations require their respective RBAC permissions.
