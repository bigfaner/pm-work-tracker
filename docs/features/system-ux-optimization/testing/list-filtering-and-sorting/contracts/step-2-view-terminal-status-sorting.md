---
journey: "list-filtering-and-sorting"
step: 2
step-action: "View terminal status sorting"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#11)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 9)
---

# Contract: list-filtering-and-sorting / Step 2: View terminal status sorting

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Terminal items sort to bottom of list | prd-spec #11, Story 9 AC1 | explicit |
| No filters shows all items | Story 8 AC3 | explicit |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "The item list page is displayed with no filters applied; at least one main item is in terminal status and at least one is not"
- Input: "PM user views the item list page"
- Output: "Non-terminal main items appear first in their original order; terminal status main items (closed, completed) are sorted to the bottom of the list"
- State: "No state change; display order only"
- Side-effect: "none"

## Outcome "no-filters-shows-all"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E1 describes no filters selected shows all items -->
- Preconditions: "No status or assignee filters are selected"
- Input: "PM user loads the item list page with no active filters"
- Output: "All items are displayed without any filtering applied; terminal items still sort to the bottom"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible, regardless of active filters or whether all filters are cleared
- When no filters are selected, all items are displayed
