---
journey: "list-filtering-and-sorting"
step: "E2"
step-action: "Filter by assignee with no sub-item matches"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#10)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 8)
---

# Contract: list-filtering-and-sorting / Edge E2: Filter by assignee with no sub-item matches

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Assignee with only direct matches shows no sub-item indicators | prd-spec #10, Story 8 | inferred |
| Filter still penetrates but finds no sub-item matches for this assignee | prd-spec #10 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E2 -- assignee B has only direct main item assignments, no sub-item assignments -->
- Preconditions: "Assignee B is only responsible for main items directly, never as a sub-item assignee"
- Input: "PM user filters by assignee B"
- Output: "Only main items directly assigned to B are shown; no sub-item match indicators appear"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while filter controls are displayed"
- Input: "PM user applies the assignee filter after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Assignee filter always penetrates to sub-item level, surfacing parent main items of matching sub-items
- Sub-item match indicator is displayed whenever a main item is shown due to sub-item filter match
