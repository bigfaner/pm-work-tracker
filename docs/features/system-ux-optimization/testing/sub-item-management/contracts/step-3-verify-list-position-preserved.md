---
journey: "sub-item-management"
step: 3
step-action: "Verify list position preserved after edit"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Step 3: Verify list position preserved after edit

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A sub-item's start time was just edited in the previous step; the main item detail page is displayed"
- Input: "PM user views the sub-item list on the main item detail page"
- Output: "The edited sub-item remains in the same position in the list as before the edit; list order is still determined by creation time, not start time"
- State: "No state change; display order unchanged"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "PM user navigates to the list after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
