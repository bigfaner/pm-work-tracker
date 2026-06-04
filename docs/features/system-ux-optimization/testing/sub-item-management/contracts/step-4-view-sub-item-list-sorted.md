---
journey: "sub-item-management"
step: 4
step-action: "View sub-item list sorted by creation time"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Step 4: View sub-item list sorted by creation time

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A main item has multiple sub-items created on different dates"
- Input: "PM user navigates to the main item detail page"
- Output: "Sub-items are displayed in descending order by creation time with the most recently created sub-item appearing first"
- State: "No state change; display order only"
- Side-effect: "none"

## Outcome "session-expired"
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
