---
journey: "list-filtering-and-sorting"
step: 4
step-action: "Clear all status filters"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 9)
---

# Contract: list-filtering-and-sorting / Step 4: Clear all status filters

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Clearing all filters shows all items | Story 9 AC3 | explicit |
| Terminal items still sort to bottom when all filters cleared | prd-spec #11 | inferred |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "The progress page is loaded with the default in-progress filter active"
- Input: "PM user deselects all status checkboxes on the progress page"
- Output: "All items are displayed including terminal status items; terminal status main items still sort to the bottom of the list"
- State: "No state change; filter is cleared for display"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the progress page is displayed with filters active"
- Input: "PM user deselects filters after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible, regardless of active filters
- When no filters are selected, all items are displayed
