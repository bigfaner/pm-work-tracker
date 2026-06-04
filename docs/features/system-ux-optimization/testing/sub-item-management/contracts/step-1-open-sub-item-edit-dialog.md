---
journey: "sub-item-management"
step: 1
step-action: "Open sub-item edit dialog"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Step 1: Open sub-item edit dialog

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A sub-item exists with a start time value set; user has sub_item:update permission"
- Input: "PM user opens the edit dialog for a sub-item"
- Output: "Edit dialog renders with all existing fields populated including the start time field showing the current start time value"
- State: "No state change; dialog is displayed"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before opening the edit dialog"
- Input: "PM user attempts to open the edit dialog after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order
