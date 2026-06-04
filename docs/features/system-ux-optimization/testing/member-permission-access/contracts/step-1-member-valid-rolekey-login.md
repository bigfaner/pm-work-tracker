---
journey: "member-permission-access"
step: 1
step-action: "Member user with valid role_key logs in"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#6)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
---

# Contract: member-permission-access / Step 1: Member user with valid role_key logs in

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Member user with valid role_key receives correct permissions | prd-spec #6 | explicit |
| Member permissions include item_pool:submit and main_item:list | prd-spec #6 | explicit |

## Outcome "success"
<!-- surface: web -->
<!-- fact: prd-spec #6 -->
- Preconditions: "A member user with a valid role_key exists in the system"
- Input: "Member user logs in with valid credentials"
- Output: "Authentication succeeds; user is redirected to the main dashboard; no permission errors occur during login"
- State: "User session is established with the correct permission set for the member role"
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

- Member-role users always receive at least item_pool:submit and main_item:list permissions after login, regardless of whether their role_key is valid or nil/empty
- Menu items and actions visible to a user always match their role's permission set exactly
- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- API endpoints enforce the same permission checks as the UI
