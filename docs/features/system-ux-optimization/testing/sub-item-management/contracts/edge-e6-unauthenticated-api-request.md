---
journey: "sub-item-management"
step: "E6"
step-action: "Unauthenticated API request for sub-items"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md
---

# Contract: sub-item-management / Edge E6: Unauthenticated API request for sub-items

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Unauthenticated request returns authentication error | api-handbook Auth | inferred |
| No data returned without authentication | api-handbook Auth | inferred |

## Outcome "unauthenticated"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E6 -- API request without valid authentication credentials -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- API endpoints enforce the same permission checks as the UI
