---
journey: "sub-item-management"
step: "E9"
step-action: "Edit a concurrently deleted sub-item"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Edge E9: Edit a concurrently deleted sub-item

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Concurrently deleted sub-item shows error on save | journey E9 | inferred |
| No data modified for non-existent sub-item | journey E9 | inferred |

## Outcome "concurrently-deleted"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E9 -- another user deleted the sub-item while edit dialog is open -->
- Preconditions: "Another user has deleted the sub-item while the current user has the edit dialog open"
- Input: "PM user submits the edit form"
- Output: "An error message is displayed indicating the sub-item no longer exists; the edit dialog closes; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Editing start time does not affect the sub-item's position in the sorted list
