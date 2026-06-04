---
journey: "sub-item-management"
step: "E3"
step-action: "Sub-items with identical creation times"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Edge E3: Sub-items with identical creation times

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Identical creation times produce stable deterministic order | journey E3 | inferred |
| No random reordering on repeated page loads | journey E3 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E3 -- tiebreaker behavior when creation times match -->
- Preconditions: "Multiple sub-items were created simultaneously (same timestamp)"
- Input: "PM user views the main item detail page"
- Output: "Sub-items with identical creation times are displayed in a stable, deterministic order"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user navigates to the page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
