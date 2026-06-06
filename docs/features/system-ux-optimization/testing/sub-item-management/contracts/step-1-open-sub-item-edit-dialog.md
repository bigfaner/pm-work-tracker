---
journey: "sub-item-management"
step: 1
step-action: "Open sub-item edit dialog"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#2)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 2)
---

# Contract: sub-item-management / Step 1: Open sub-item edit dialog

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Edit dialog shows start time field with current value | prd-spec #2, Story 2 AC1 | explicit |
| Dialog includes all existing fields populated | prd-spec #2 | explicit |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #2, Story 2 AC1 -->
- Preconditions: "A sub-item exists with a start time value set; user has sub_item:update permission"
- Input: "PM user opens the edit dialog for a sub-item"
- Output: "Edit dialog renders with all existing fields populated including the start time field showing the current start time value"
- State: "No state change; dialog is displayed"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before opening the edit dialog"
- Input: "PM user attempts to open the edit dialog after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
