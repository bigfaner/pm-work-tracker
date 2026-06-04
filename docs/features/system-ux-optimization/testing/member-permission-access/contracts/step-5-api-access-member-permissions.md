---
journey: "member-permission-access"
step: 5
step-action: "API access with member permissions"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/member-permission-access/journey.md
---

# Contract: member-permission-access / Step 5: API access with member permissions

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "A member user with nil role_key is authenticated; the middleware correctly falls back to member preset permissions"
- Input: "An authenticated API request to the item listing endpoint"
- Output: "The API returns item data within the user's team scope; no permission error is returned"
- State: "No state change"
- Side-effect: "none"

## Outcome "unauthorized-delete"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface unauthorized mandatory outcome; member role does not include main_item:delete -->
- Preconditions: "A member user is authenticated with default permissions that do not include main_item:delete"
- Input: "The member user sends an API request to a delete endpoint"
- Output: "The API returns HTTP 403 with error code ERR_FORBIDDEN; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from API surface validation-error outcome -->
- Preconditions: "An authenticated API request contains invalid or malformed parameters"
- Input: "An API request with missing required fields or invalid parameter values"
- Output: "The API returns HTTP 400 with error code VALIDATION_ERROR describing the invalid fields; no data is modified"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- API endpoints enforce the same permission checks as the UI
- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
