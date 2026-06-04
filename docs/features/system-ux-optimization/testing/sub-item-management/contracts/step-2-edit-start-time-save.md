---
journey: "sub-item-management"
step: 2
step-action: "Edit start time and save"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Step 2: Edit start time and save

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "The sub-item edit dialog is open with a start time field"
- Input: "PM user changes the start time to a new date and clicks save"
- Output: "Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time"
- State: "Sub-item plan_start_date updated in database"
- Side-effect: "none"

## Outcome "invalid-date-format"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome; date field validation -->
- Preconditions: "The sub-item edit dialog is open"
- Input: "PM user manually enters an invalid date format in the start time field and attempts to save"
- Output: "Date input validation prevents saving with an invalid date; an error message is shown near the date field indicating the invalid format"
- State: "No state change; start time unchanged"
- Side-effect: "none"

## Outcome "start-after-end"
<!-- source: inferred -->
<!-- reasoning: Journey E1b describes start time after end time scenario; business rule validation -->
- Preconditions: "A sub-item has an end time set to a specific date"
- Input: "PM user sets the start time to a date after the end time and attempts to save"
- Output: "A validation error message is displayed indicating that start time must not be later than end time; the change is not saved"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user was previously authenticated and the session has since expired while the edit dialog is open"
- Input: "PM user submits the edit form after session expiry"
- Output: "The user is redirected to the login page; no data is modified; after re-authenticating, the sub-item retains its original values"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Editing start time does not affect the sub-item's position in the sorted list
