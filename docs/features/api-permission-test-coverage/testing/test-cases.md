---
feature: "api-permission-test-coverage"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-28"
---

# Test Cases: api-permission-test-coverage

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 12    |
| CLI  | 0     |
| **Total** | **12** |

---

## API Test Cases

## TC-001: Permission injection grants access
- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/permission-middleware
- **Test ID**: api/permission-middleware/permission-injection-grants-access
- **Pre-conditions**: Handler test context with `c.Set("permCodes", []string{"main_item:archive"})` injected
- **Steps**:
  1. Inject `permCodes` containing `"main_item:archive"` into handler context
  2. Call `POST /teams/:teamId/main-items/:id/archive`
- **Expected**: Response status 200
- **Priority**: P0

---

## TC-002: Empty permission injection denies access
- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/permission-middleware
- **Test ID**: api/permission-middleware/empty-permission-injection-denies-access
- **Pre-conditions**: Handler test context with `c.Set("permCodes", []string{})` injected
- **Steps**:
  1. Inject empty `permCodes` (`[]string{}`) into handler context
  2. Call `POST /teams/:teamId/main-items/:id/archive`
- **Expected**: Response status 403
- **Priority**: P0

---

## TC-003: Preset roles matrix — archive endpoint
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/main-item
- **Test ID**: api/main-item/preset-roles-matrix-archive-endpoint
- **Pre-conditions**: Three users bound to superadmin / pm / member roles respectively; team created; target main item exists with completed status
- **Steps**:
  1. Authenticate as superadmin user, call `POST /teams/:teamId/main-items/:id/archive`
  2. Authenticate as pm user, call same endpoint
  3. Authenticate as member user, call same endpoint
- **Expected**: superadmin → 200, pm → 200, member → 403
- **Priority**: P0

---

## TC-004: Preset roles matrix — team invite endpoint
- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/team
- **Test ID**: api/team/preset-roles-matrix-team-invite-endpoint
- **Pre-conditions**: Three users bound to superadmin / pm / member roles respectively; team created
- **Steps**:
  1. Authenticate as superadmin user, call `POST /teams/:teamId/members`
  2. Authenticate as pm user, call same endpoint
  3. Authenticate as member user, call same endpoint
- **Expected**: superadmin → 200, pm → 200, member → 403
- **Priority**: P0

---

## TC-005: Custom role with partial permissions allows read
- **Source**: Story 3 / AC-1
- **Type**: API
- **Target**: api/custom-role
- **Test ID**: api/custom-role/custom-role-with-partial-permissions-allows-read
- **Pre-conditions**: Custom role created with only `main_item:read` + `progress:read`; user assigned to this role
- **Steps**:
  1. Authenticate as the custom-role user
  2. Call `GET /teams/:teamId/main-items`
- **Expected**: Response status 200
- **Priority**: P0

---

## TC-006: Custom role without create permission denies write
- **Source**: Story 3 / AC-2
- **Type**: API
- **Target**: api/custom-role
- **Test ID**: api/custom-role/custom-role-without-create-permission-denies-write
- **Pre-conditions**: Custom role created with only `main_item:read` + `progress:read`; user assigned to this role
- **Steps**:
  1. Authenticate as the custom-role user
  2. Call `POST /teams/:teamId/main-items` (requires `main_item:create`)
- **Expected**: Response status 403
- **Priority**: P0

---

## TC-007: Permission change takes effect immediately without re-login
- **Source**: Story 3 / AC-3
- **Type**: API
- **Target**: api/custom-role
- **Test ID**: api/custom-role/permission-change-takes-effect-immediately-without-re-login
- **Pre-conditions**: Custom role has only `main_item:read` + `progress:read`; user authenticated with existing token; admin adds `main_item:create` to the role via DB
- **Steps**:
  1. Using the same token (no re-login), call `POST /teams/:teamId/main-items`
- **Expected**: Response status 200 (permission change reflected immediately, no cache)
- **Priority**: P0

---

## TC-008: Empty permission role is denied on protected endpoint
- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/permission-boundary
- **Test ID**: api/permission-boundary/empty-permission-role-is-denied-on-protected-endpoint
- **Pre-conditions**: User bound to a role with no permission codes
- **Steps**:
  1. Authenticate as the empty-role user
  2. Call any protected endpoint (e.g. `POST /teams/:teamId/main-items/:id/archive`)
- **Expected**: Response status 403
- **Priority**: P0

---

## TC-009: Superadmin bypasses permission check
- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/permission-boundary
- **Test ID**: api/permission-boundary/superadmin-bypasses-permission-check
- **Pre-conditions**: superadmin user exists; target resource exists (fixture data present)
- **Steps**:
  1. Authenticate as superadmin user
  2. Call any protected endpoint (e.g. `POST /teams/:teamId/main-items/:id/archive`)
- **Expected**: Response status 200 (not 403; 404/500 indicates fixture missing, not a pass)
- **Priority**: P0

---

## TC-010: Invalid token returns 401 not 403
- **Source**: Story 4 / AC-3
- **Type**: API
- **Target**: api/permission-boundary
- **Test ID**: api/permission-boundary/invalid-token-returns-401-not-403
- **Pre-conditions**: None
- **Steps**:
  1. Send request with an invalid/malformed JWT token
  2. Call any endpoint
- **Expected**: Response status 401 (authentication failure, distinct from 403 authorization failure)
- **Priority**: P0

---

## TC-011: CI fails when permission code lacks test coverage
- **Source**: Story 5 / AC-1
- **Type**: API
- **Target**: api/permission-coverage
- **Test ID**: api/permission-coverage/ci-fails-when-permission-code-lacks-test-coverage
- **Pre-conditions**: `codes.go` defines a permission code `foo:bar`; `foo:bar` does not appear as a `permCodes` parameter or test matrix value in any test file (comments/logs excluded)
- **Steps**:
  1. Run the permission coverage CI assertion step (e.g. `go test ./... -run TestPermissionCodeCoverage`)
- **Expected**: Build fails; output contains `missing test coverage for: foo:bar`
- **Priority**: P1

---

## TC-012: CI passes when all permission codes have test coverage
- **Source**: Story 5 / AC-2
- **Type**: API
- **Target**: api/permission-coverage
- **Test ID**: api/permission-coverage/ci-passes-when-all-permission-codes-have-test-coverage
- **Pre-conditions**: Every permission code defined in `codes.go` appears in `middleware/permission_test.go` or `tests/integration/rbac_test.go` as a `permCodes` value or test matrix entry
- **Steps**:
  1. Run the permission coverage CI assertion step
- **Expected**: Assertion passes; build continues
- **Priority**: P1

---

## Traceability

| TC ID  | Source        | Type | Target                      | Priority |
|--------|---------------|------|-----------------------------|----------|
| TC-001 | Story 1 / AC-1 | API  | api/permission-middleware   | P0       |
| TC-002 | Story 1 / AC-2 | API  | api/permission-middleware   | P0       |
| TC-003 | Story 2 / AC-1 | API  | api/main-item               | P0       |
| TC-004 | Story 2 / AC-2 | API  | api/team                    | P0       |
| TC-005 | Story 3 / AC-1 | API  | api/custom-role             | P0       |
| TC-006 | Story 3 / AC-2 | API  | api/custom-role             | P0       |
| TC-007 | Story 3 / AC-3 | API  | api/custom-role             | P0       |
| TC-008 | Story 4 / AC-1 | API  | api/permission-boundary     | P0       |
| TC-009 | Story 4 / AC-2 | API  | api/permission-boundary     | P0       |
| TC-010 | Story 4 / AC-3 | API  | api/permission-boundary     | P0       |
| TC-011 | Story 5 / AC-1 | API  | api/permission-coverage     | P1       |
| TC-012 | Story 5 / AC-2 | API  | api/permission-coverage     | P1       |
