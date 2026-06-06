---
journey: "member-permission-access"
step: 2
step-action: "Member user with nil role_key logs in (core bug fix)"
surface_types: ["web", "api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#6)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
---

# Contract: member-permission-access / Step 2: Member user with nil role_key logs in (core bug fix)

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Nil role_key triggers fallback to member preset permissions | prd-spec #6 | explicit |
| Member permissions include item_pool:submit and main_item:list | prd-spec #6 | explicit |
| Nil-roleKey user gets same experience as valid-roleKey user | journey invariant | inferred |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #6 -->
- Preconditions: "A member user exists with an empty or null role_key column in the database; valid credentials are available"
- Input: "Member user with nil role_key logs in with valid credentials"
- Output: "Authentication succeeds; the system provides the member role's default permissions; user is redirected to the main dashboard; the same menus and permissions are available as the user with a valid role_key"
- State: "User session is established with permissions from the member role's default permission set"
- Side-effect: "none"

## Outcome "forbidden-endpoint"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Nil-roleKey member attempting a restricted API endpoint (e.g., delete) should still be blocked; covers the actual unauthorized scenario for nil-roleKey users -->
- Preconditions: "A member user with nil role_key is authenticated; the user does not have permission for a restricted endpoint"
- Input: "The member user sends an API request to a restricted endpoint (e.g., delete)"
- Output: "The API returns an authorization error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome; login form validation -->
- Preconditions: "The login form is displayed"
- Input: "User submits login with an empty password or malformed email format"
- Output: "A validation error message is displayed near the relevant field; no API request is sent; the form remains on screen for correction"
- State: "No session created"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "A previous session has expired"
- Input: "User attempts to navigate to a protected page with expired session"
- Output: "The user is redirected to the login page; no partial data is displayed from the previous session"
- State: "No active session"
- Side-effect: "none"

## Journey Invariants

- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login, regardless of whether their role_key is valid or nil
- Menu items and actions visible to a user always match their role's permission set exactly
- API endpoints enforce the same permission checks as the UI
