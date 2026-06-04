---
journey: "sub-item-move"
step: 2
step-action: "Select valid target and confirm"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
---

# Contract: sub-item-move / Step 2: Select valid target and confirm

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "The move target selector is open showing at least one valid non-terminal target"
- Input: "PM user selects main item B as the target and confirms the move"
- Output: "Sub-item is moved to main item B; the code number is automatically regenerated under the target; status and assignee remain unchanged; the page updates to reflect the new parent"
- State: "Sub-item main_item_key updated to target BizKey; item_code regenerated via atomic increment; completion recalculated for both source and target main items"
- Side-effect: "Source main item completion percentage recalculated; target main item completion percentage recalculated"

## Outcome "move-to-closed-target"
<!-- source: inferred -->
<!-- reasoning: Fact Table MOVE_TERMINAL_CHECK shows Move checks IsMainTerminal and returns ErrTargetClosed -->
- Preconditions: "A main item is in terminal status (closed or completed)"
- Input: "PM user attempts to select the terminal main item as the move target"
- Output: "The closed main item is shown as disabled or non-selectable in the selector; if selected, the operation is rejected with a message indicating the target is closed"
- State: "No state change"
- Side-effect: "none"

## Outcome "move-to-same-parent"
<!-- source: inferred -->
<!-- reasoning: Fact Table MOVE_SAME_PARENT_CHECK shows Move rejects same parent with ErrSameMainItem -->
- Preconditions: "Sub-item currently belongs to main item A; the move target is also main item A"
- Input: "PM user attempts to move the sub-item to the same parent"
- Output: "The operation is rejected; a message indicates the sub-item is already under this main item; no change occurs"
- State: "No state change"
- Side-effect: "none"

## Outcome "no-target-selected"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome -->
- Preconditions: "The move target selector is open but no target has been selected"
- Input: "PM user attempts to confirm the move without selecting a target"
- Output: "The confirm button is disabled or a validation error message is displayed indicating a target must be selected; no API request is sent"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the move target selector is open"
- Input: "PM user selects a target and confirms the move after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the move dialog is no longer open"
- State: "No state change; sub-item remains under original parent"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- source: inferred -->
<!-- reasoning: Fact Table ROUTE_SUB_ITEM_MOVE requires sub_item:update permission -->
- Preconditions: "A user without sub_item:update permission sends a move API request"
- Input: "The user sends a move API request without proper authorization"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)
- Closed or completed main items are never valid targets for sub-item moves
