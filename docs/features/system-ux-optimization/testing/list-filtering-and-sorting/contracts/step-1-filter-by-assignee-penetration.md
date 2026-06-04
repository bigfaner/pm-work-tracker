---
journey: "list-filtering-and-sorting"
step: 1
step-action: "Filter by assignee with penetration"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
---

# Contract: list-filtering-and-sorting / Step 1: Filter by assignee with penetration

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Assignee A is the direct assignee of some main items AND the assignee of sub-items under other main items"
- Input: "PM user selects assignee A in the assignee filter on the item list page"
- Output: "Results include main items where A is the direct assignee AND main items where A is the assignee of a sub-item; main items shown due to sub-item match display a sub-item match visual indicator; only the matching sub-items are shown under those main items"
- State: "No state change; filter is applied for display only"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while filter controls are displayed"
- Input: "PM user applies a new filter after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the item list is displayed without the filter"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome -->
- Preconditions: "The assignee filter field accepts text input that is validated against known assignee identifiers"
- Input: "PM user enters special characters or an extremely long string in the assignee filter field and triggers the filter"
- Output: "An error message is displayed near the filter field indicating invalid input; no API request with invalid data is sent"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Terminal status main items always sort to the bottom of item lists when visible
- Sub-item match indicator is displayed whenever a main item is shown due to sub-item filter match
