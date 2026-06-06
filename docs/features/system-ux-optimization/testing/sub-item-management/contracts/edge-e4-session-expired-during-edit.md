---
journey: "sub-item-management"
step: "E4"
step-action: "Session expired during sub-item edit"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-management/journey.md
---

# Contract: sub-item-management / Edge E4: Session expired during sub-item edit

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Session expired during edit preserves original values | journey E4 | inferred |
| User redirected to login on expired session | Web surface mandatory outcome | inferred |

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E4 -- session expired while edit dialog is open -->
- Preconditions: "The user was previously authenticated and the session has since expired while the edit dialog is open"
- Input: "PM user submits the edit form"
- Output: "The user is redirected to the login page; no data is modified; after re-authenticating, the sub-item retains its original values"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The sub-item edit dialog always includes a start time field that can be modified and saved
- Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)
