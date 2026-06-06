---
journey: "item-deletion"
step: 3
step-action: "Non-PM user sees no delete button"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/item-deletion/journey.md
---

# Contract: item-deletion / Step 3: Non-PM user sees no delete button

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "no-delete-button"
- Preconditions: "A member-role user is viewing a main item or sub-item detail page; member role does not include main_item:delete or sub_item:delete permissions"
- Input: "Member-role user views a main item or sub-item detail page"
- Output: "No delete button is visible on the page"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "A member user was previously authenticated and the session has since expired"
- Input: "Member user attempts to view an item detail page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome; invalid item URL -->
- Preconditions: "The item URL contains a malformed or invalid item identifier"
- Input: "PM user navigates to a page with an invalid item identifier"
- Output: "An error message is displayed indicating the item identifier is invalid; no deletion is attempted"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- The delete button is only visible to users with appropriate permissions (main_item:delete, sub_item:delete)
- Deletion is always soft-delete
