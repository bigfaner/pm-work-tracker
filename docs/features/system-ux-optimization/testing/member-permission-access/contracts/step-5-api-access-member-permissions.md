---
journey: "member-permission-access"
step: 5
step-action: "API access with member permissions"
surface_types: ["api"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#6)
  - docs/features/system-ux-optimization/design/api-handbook.md
---

# Contract: member-permission-access / Step 5: API access with member permissions

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Nil-roleKey member can access API with default permissions | prd-spec #6 | explicit |
| Member role does not include main_item:delete | journey E3 | inferred |
| API returns data within team scope | prd-spec #6 | inferred |

## Outcome "success"
<!-- surface: api -->
<!-- fact: prd-spec #6 -->
- Preconditions: "A member user with nil role_key is authenticated"
- Input: "An authenticated API request to the item listing endpoint"
- Output: "The API returns item data within the user's team scope; no permission error is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized-delete"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E3 -- member role does not include main_item:delete; API enforces permission checks -->
- Preconditions: "A member user is authenticated with default permissions that do not include main_item:delete"
- Input: "The member user sends an API request to a delete endpoint"
- Output: "The API returns an authorization error response; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome -->
- Preconditions: "An authenticated API request contains invalid or malformed parameters"
- Input: "An API request with missing required fields or invalid parameter values"
- Output: "The API returns a validation error response describing the invalid fields; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthenticated"
<!-- surface: api -->
<!-- source: inferred -->
<!-- reasoning: Journey E2 -- API request without valid authentication credentials -->
- Preconditions: "An API request is sent without valid credentials"
- Input: "An API request without a valid authentication token"
- Output: "The API returns an authentication error response; no data is returned"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- API endpoints enforce the same permission checks as the UI
- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- Member-role users always receive at least item_pool:submit and main_item:list permissions after login
