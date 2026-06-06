---
journey: "sub-item-management"
step: "E7"
step-action: "API validation error for malformed request"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md
---

# Contract: sub-item-management / Edge E7: API validation error for malformed request

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Non-numeric team ID returns validation error | api-handbook validation | inferred |
| Validation error describes invalid parameter | api-handbook validation | inferred |

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E7 -- non-numeric team ID parameter triggers validation error -->
- Preconditions: "An API request uses a non-numeric value for a team ID parameter"
- Input: "An API request is sent with a non-numeric team ID"
- Output: "The API returns a validation error response listing the invalid parameter; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
