---
journey: "sub-item-management"
step: 2
step-action: "Edit start time and save"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#2)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 2)
---

# Contract: sub-item-management / Step 2: Edit start time and save

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Start time can be edited and saved | prd-spec #2, Story 2 AC1 | explicit |
| Invalid date format rejected by validation | journey E1 | inferred |
| Start time must not be later than end time | prd-ui-functions.md line 115 | explicit |
| Session expired during edit preserves original values | journey E4 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #2, Story 2 AC1 -->
- Preconditions: "The sub-item edit dialog is open with a start time field"
- Input: "PM user changes the start time to a new date and clicks save"
- Output: "Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time"
- State: "Sub-item start time is persisted"
- Side-effect: "none"

## Outcome "invalid-date-format"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E1 -- invalid date format entered manually -->
- Preconditions: "The sub-item edit dialog is open"
- Input: "PM user manually enters an invalid date format in the start time field and attempts to save"
- Output: "Date input validation prevents saving with an invalid date; an error message is shown near the date field indicating the invalid format"
- State: "No state change; start time unchanged"
- Side-effect: "none"

## Outcome "start-after-end"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E1b -- start time after end time violates business rule -->
- Preconditions: "A sub-item has an end time set"
- Input: "PM user sets the start time to a date after the end time and attempts to save"
- Output: "A validation error message is displayed indicating that start time must not be later than end time; the change is not saved"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E4 -- session expired while edit dialog is open -->
- Preconditions: "The user was previously authenticated and the session has since expired while the edit dialog is open"
- Input: "PM user submits the edit form after session expiry"
- Output: "The user is redirected to the login page; no data is modified; after re-authenticating, the sub-item retains its original values"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
- Sub-item creation time determines sort order, not start time or update time
- Editing start time does not affect the sub-item's position in the sorted list
