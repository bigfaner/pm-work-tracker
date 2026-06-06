---
journey: "sub-item-management"
step: 4
step-action: "View sub-item list sorted by creation time"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#5)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 5)
---

# Contract: sub-item-management / Step 4: View sub-item list sorted by creation time

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Sub-items sorted by creation time descending | Story 5 AC1 | explicit |
| Most recently created sub-item appears first | Story 5 AC1 | explicit |

## Outcome "success"
<!-- surface: web -->
<!-- fact: Story 5 AC1 -->
- Preconditions: "A main item has multiple sub-items created on different dates"
- Input: "PM user navigates to the main item detail page"
- Output: "Sub-items are displayed in descending order by creation time with the most recently created sub-item appearing first"
- State: "No state change; display order only"
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

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
