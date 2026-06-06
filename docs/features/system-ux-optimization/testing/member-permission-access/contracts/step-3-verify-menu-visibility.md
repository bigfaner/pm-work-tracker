---
journey: "member-permission-access"
step: 3
step-action: "Verify menu visibility matches permission set"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#6)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
---

# Contract: member-permission-access / Step 3: Verify menu visibility matches permission set

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Menu items match permission set exactly | prd-spec #6, Story 6 | explicit |
| PM-only actions hidden from member users | journey E4 | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #6, Story 6 -->
- Preconditions: "A member user is logged in with either valid or nil role_key"
- Input: "Member user views the navigation menu"
- Output: "Menu displays items corresponding to item_pool:submit and main_item:list permissions; menus outside the member role's permissions (delete, admin functions) are not visible"
- State: "No state change; menu reflects permission set"
- Side-effect: "none"

## Outcome "pm-only-actions-hidden"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Journey E4 describes member user viewing item detail page without PM controls -->
- Preconditions: "A member user is logged in; an item detail page is accessible"
- Input: "Member user views the item detail page"
- Output: "The delete button is not visible; no admin or PM-only controls are displayed"
- State: "No state change"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "A member user was previously authenticated and the session has since expired"
- Input: "Member user attempts to navigate to a page requiring authentication"
- Output: "The user is redirected to the login page; no partial data is displayed from the previous session"
- State: "No active session"
- Side-effect: "none"

## Journey Invariants

- Menu items and actions visible to a user always match their role's permission set exactly
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login, regardless of whether their role_key is valid or nil/empty
- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- API endpoints enforce the same permission checks as the UI
