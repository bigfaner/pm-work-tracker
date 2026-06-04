---
journey: "sub-item-management"
step: "E8"
step-action: "API request for non-existent main item"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md
---

# Contract: sub-item-management / Edge E8: API request for non-existent main item

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Non-existent main item returns not-found error | api-handbook | inferred |

## Outcome "not-found"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E8 -- API request for non-existent main item ID -->
- Preconditions: "The main item ID in the request does not exist in the database"
- Input: "An API request for a non-existent main item"
- Output: "The API returns a not-found error response"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
