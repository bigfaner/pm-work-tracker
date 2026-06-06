---
journey: "list-filtering-and-sorting"
step: "E5"
step-action: "Unauthenticated list API request"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Enhanced List Main Items)
---

# Contract: list-filtering-and-sorting / Edge E5: Unauthenticated list API request

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Unauthenticated requests return authentication error | api-handbook Auth | inferred |
| No data returned for unauthenticated requests | api-handbook Auth | inferred |

## Outcome "unauthenticated"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E5 distinguishes unauthenticated (no valid token) from unauthorized (wrong permissions) -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request to the main items endpoint without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned; no sensitive information is exposed"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items (closed, completed) always sort to the bottom of item lists when they are visible
- Assignee filter always penetrates to sub-item level
