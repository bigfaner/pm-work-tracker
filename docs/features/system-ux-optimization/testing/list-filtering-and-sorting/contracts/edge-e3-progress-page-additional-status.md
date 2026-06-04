---
journey: "list-filtering-and-sorting"
step: "E3"
step-action: "Progress page select additional status"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#12)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 9)
---

# Contract: list-filtering-and-sorting / Edge E3: Progress page select additional status

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Multiple status filters can be combined | prd-spec #12, Story 9 AC2 | explicit |
| Terminal items sort to bottom with any filter combination | prd-spec #11 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E3 -- user adds a second status to the default filter -->
- Preconditions: "The progress page is loaded with the default in-progress filter active; completed main items exist"
- Input: "PM user selects the completed status checkbox in addition to the default in-progress filter"
- Output: "Both in-progress and completed items are displayed; terminal items still sort to the bottom"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the progress page with active filters is displayed"
- Input: "PM user selects additional filters after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible, regardless of active filters
- When no filters are selected, all items are displayed
