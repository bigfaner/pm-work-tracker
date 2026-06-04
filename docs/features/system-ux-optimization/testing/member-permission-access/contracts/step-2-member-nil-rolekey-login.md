---
journey: "member-permission-access"
step: 2
step-action: "Member user with nil role_key logs in (core bug fix)"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
---

# Contract: member-permission-access / Step 2: Member user with nil role_key logs in (core bug fix)

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A member user exists with an empty or null role_key column in the database"
- Input: "Member user with nil role_key logs in with valid credentials"
- Output: "Authentication succeeds; the middleware falls back to querying the member role's default permission set; user is redirected to the main dashboard; the same menus and permissions are available as the user with a valid role_key"
- State: "User session is established with permissions from the member preset role fallback"
- Side-effect: "none"

## Outcome "unauthorized-api"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface unauthorized mandatory outcome; testing that nil-roleKey user can still authenticate for API -->
- Preconditions: "A member user with nil role_key attempts an authenticated API request"
- Input: "An API request with a valid token obtained by the nil-roleKey user"
- Output: "The API processes the request normally; the user receives the same permissions as a valid-roleKey member"
- State: "User authenticated successfully via API"
- Side-effect: "none"

## Journey Invariants

- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login, regardless of whether their role_key is valid or nil
