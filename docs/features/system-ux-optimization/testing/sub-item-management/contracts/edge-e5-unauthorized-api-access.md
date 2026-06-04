---
journey: "sub-item-management"
step: "E5"
step-action: "Unauthorized API access to sub-item data"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
---

# Contract: sub-item-management / Edge E5: Unauthorized API access to sub-item data

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| main_item:read permission required for sub-item listing | api-handbook Auth | inferred |
| Unauthorized request returns authorization error | api-handbook Auth | inferred |

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E5 -- authenticated user without main_item:read permission attempts API access -->
- Preconditions: "An authenticated user without main_item:read permission sends an API request"
- Input: "The user sends an API request to the main items endpoint"
- Output: "The API returns an authorization error response; no sub-item data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- API endpoints enforce the same permission checks as the UI
