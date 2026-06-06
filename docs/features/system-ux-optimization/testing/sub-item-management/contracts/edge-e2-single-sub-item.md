---
journey: "sub-item-management"
step: "E2"
step-action: "Sub-item list with single sub-item"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 5)
---

# Contract: sub-item-management / Edge E2: Sub-item list with single sub-item

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Sorting applies regardless of item count | Story 5 | inferred |
| Single sub-item displayed correctly | Story 5 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E2 -- sorting applies even with single item; no edge case failure -->
- Preconditions: "A main item has exactly one sub-item"
- Input: "PM user views the main item detail page"
- Output: "The single sub-item is displayed; no sorting issues occur; empty state is not triggered"
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
