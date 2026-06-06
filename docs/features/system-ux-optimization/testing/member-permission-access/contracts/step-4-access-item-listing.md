---
journey: "member-permission-access"
step: 4
step-action: "Access item listing page"
surface_types: ["web", "api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
  - docs/features/system-ux-optimization/design/api-handbook.md
---

# Contract: member-permission-access / Step 4: Access item listing page

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Member user with main_item:list permission can view items | Story 6 | explicit |
| API enforces same permission checks as UI | journey invariant | explicit |
| main_item:list permission required for item listing | prd-spec #6 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- fact: Story 6 -->
- Preconditions: "A member user is logged in with main_item:list permission"
- Input: "Member user navigates to the item listing page"
- Output: "Page loads successfully; items within the user's team are displayed; no permission errors occur"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: api-handbook requires main_item:list permission for item listing endpoint -->
- Preconditions: "A member user is authenticated without the required list permission"
- Input: "An API request to the item listing endpoint without proper authorization"
- Output: "The API returns an authorization error response; no item data is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "not-found"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Derived from API surface not-found common boundary outcome -->
- Preconditions: "A member user is authenticated; the requested resource ID does not exist"
- Input: "The member user sends an API request for a non-existent item ID"
- Output: "The API returns a not-found error response"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "A member user's session has expired"
- Input: "Member user navigates to the item listing page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- API endpoints enforce the same permission checks as the UI
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login
- Menu items and actions visible to a user always match their role's permission set exactly
