---
journey: "sub-item-move"
step: 2
step-action: "Select valid target and confirm"
surface_types: ["web", "api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/sub-item-move/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 7)
  - docs/features/system-ux-optimization/prd/prd-spec.md
  - docs/features/system-ux-optimization/design/api-handbook.md (Move Sub Item)
---

# Contract: sub-item-move / Step 2: Select valid target and confirm

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Move reassigns sub-item to target with new code number | prd-spec Story 7 AC1 | explicit |
| Terminal target rejected with 400 error | api-handbook.md Move Sub Item Error Responses | explicit |
| Same parent rejected with 400 error | api-handbook.md Move Sub Item Error Responses | explicit |
| sub_item:update auth required | api-handbook.md Move Sub Item Auth | explicit |
| 404 for non-existent sub-item | api-handbook.md Move Sub Item Error Responses | inferred |
| 404 for non-existent target | api-handbook.md Move Sub Item Error Responses | inferred |
| 400 for validation errors | api-handbook.md Move Sub Item Error Responses | inferred |
| 409 for concurrent move conflict | journey E3 — concurrent edit pattern | inferred |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "The move target selector is open showing at least one valid non-terminal target"
- Input: "PM user selects main item B as the target and confirms the move"
- Output: "Sub-item is moved to main item B; the code number is automatically regenerated under the target; status and assignee remain unchanged; the page updates to reflect the new parent"
- State: "Sub-item is reassigned to the target main item with a new code number; completion percentages updated for both source and target main items"
- Side-effect: "Source main item completion percentage recalculated; target main item completion percentage recalculated"

## Outcome "move-to-closed-target"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item checks IsMainTerminal and returns error for closed target -->
- Preconditions: "A main item is in terminal status (closed or completed)"
- Input: "PM user attempts to select the terminal main item as the move target"
- Output: "The closed main item is shown as disabled or non-selectable in the selector; if selected, the operation is rejected with a message indicating the target is closed"
- State: "No state change"
- Side-effect: "none"

## Outcome "move-to-same-parent"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item rejects same parent with error -->
- Preconditions: "Sub-item currently belongs to main item A; the move target is also main item A"
- Input: "PM user attempts to move the sub-item to the same parent"
- Output: "The operation is rejected; a message indicates the sub-item is already under this main item; no change occurs"
- State: "No state change"
- Side-effect: "none"

## Outcome "no-target-selected"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome -->
- Preconditions: "The move target selector is open but no target has been selected"
- Input: "PM user attempts to confirm the move without selecting a target"
- Output: "The confirm button is disabled or a validation error message is displayed indicating a target must be selected; no API request is sent"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired while the move target selector is open"
- Input: "PM user selects a target and confirms the move after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the move dialog is no longer open"
- State: "No state change; sub-item remains under original parent"
- Side-effect: "none"

## Outcome "unauthorized"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item requires sub_item:update permission -->
- Preconditions: "A user without sub_item:update permission sends a move API request"
- Input: "The user sends a move API request without proper authorization"
- Output: "The API returns an authorization error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "concurrent-move-conflict"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E3 -- second transaction detects parent change; concurrent edit pattern -->
- Preconditions: "Another transaction has already moved this sub-item to a different parent"
- Input: "A second move request arrives for the same sub-item after the first transaction committed"
- Output: "The API returns a conflict error response indicating the sub-item's parent has changed; no data corruption occurs"
- State: "No state change; sub-item remains at first transaction's target"
- Side-effect: "none"

## Outcome "source-deleted-during-move"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E4 describes source main item deleted while move dialog is open -->
- Preconditions: "Another user has soft-deleted main item A while the current user was performing the move"
- Input: "PM user confirms the move after the source was deleted"
- Output: "An error message is displayed indicating the source main item no longer exists; no move is performed"
- State: "No state change"
- Side-effect: "none"

## Outcome "non-existent-sub-item-api"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item returns not-found for non-existent sub-item -->
- Preconditions: "The sub-item ID in the request does not exist in the database"
- Input: "A move API request is sent with a non-existent sub-item ID"
- Output: "The API returns a not-found error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "non-existent-target-api"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item returns not-found for non-existent target -->
- Preconditions: "The target main item ID in the request does not exist in the database"
- Input: "A move API request is sent with a non-existent target main item ID"
- Output: "The API returns a not-found error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook Move Sub Item returns validation error for missing/invalid fields -->
- Preconditions: "A move API request is sent with missing or invalid fields"
- Input: "An API request to the move endpoint with an invalid or missing target ID"
- Output: "The API returns a validation error response listing the missing or invalid fields; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Sub-item code number is always automatically regenerated under the target main item
- Status and assignee of the moved sub-item are never changed by the move operation
- Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)
- Closed/completed main items are never valid targets for sub-item moves
- Moving to the same parent main item is always rejected
