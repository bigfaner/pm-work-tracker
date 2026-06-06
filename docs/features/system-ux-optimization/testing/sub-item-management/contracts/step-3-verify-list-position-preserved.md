---
journey: "sub-item-management"
step: 3
step-action: "Verify list position preserved after edit"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#5)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 5)
---

# Contract: sub-item-management / Step 3: Verify list position preserved after edit

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Sort order is by creation time, not start time | prd-spec #5 | explicit |
| Editing start time does not affect list position | journey invariant | explicit |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #5 -->
- Preconditions: "A sub-item exists whose start time differs from its original value; the main item detail page is accessible"
- Input: "PM user views the sub-item list on the main item detail page"
- Output: "The edited sub-item remains in the same position in the list as before the edit; list order is still determined by creation time, not start time"
- State: "No state change; display order unchanged"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user navigates to the list after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
