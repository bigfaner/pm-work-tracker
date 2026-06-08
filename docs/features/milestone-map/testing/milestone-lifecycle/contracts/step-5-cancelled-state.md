---
journey: "milestone-lifecycle"
step: 5
step-action: "Cancelled state interactions"
generated: "2026-06-08"
sources:
  - docs/features/milestone-map/testing/milestone-lifecycle/journey.md
---

# Contract: milestone-lifecycle / Step 5: Cancelled state interactions

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "cannot-receive-bindings"
- Preconditions: "Milestone is in cancelled status"
- Input: "PM tries to bind a MainItem to this cancelled milestone"
- Output: "Binding is rejected with appropriate error message"
- State: "No new MI bindings created"
- Side-effect: "none"

## Outcome "panel-muted-appearance"
- Preconditions: "Milestone is in cancelled status"
- Input: "PM views the detail panel"
- Output: "Panel displays with muted visual tone; associated MI list is empty; add button not shown; delete button visible"
- State: "Milestone remains in cancelled status"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: surface-required (API surface) -->
- Preconditions: "API request sent without valid credentials"
- Input: "Unauthenticated request to milestone endpoint"
- Output: "API returns authentication error with HTTP 401"
- State: "No data modified"
- Side-effect: "none"

## Journey Invariants

- Cancelled milestones cannot receive new MainItem bindings.
- cancelled is a terminal state with no status transitions available.
- Panel for cancelled milestones shows muted tone, empty MI list, no add control, but visible delete.
