---
journey: "list-filtering-and-sorting"
step: 5
step-action: "Empty state with active filters"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
---

# Contract: list-filtering-and-sorting / Step 5: Empty state with active filters

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Filter criteria are applied that match no items in the system"
- Input: "PM user applies filters that match no items"
- Output: "Empty state message indicating no matching items is displayed; a clear filter action button is available to reset all filters"
- State: "No state change; empty result set displayed"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user applies filters after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Empty filter results always show a message with a clear filter action
- Assignee filter always penetrates to sub-item level
