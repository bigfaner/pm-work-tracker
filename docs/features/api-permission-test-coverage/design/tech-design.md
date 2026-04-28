---
created: 2026-04-28
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: API Permission Test Coverage

## Overview

This feature adds zero production code. It fills three structural gaps in the RBAC test suite:

1. **U1 — Handler-level permission matrix** (`backend/internal/handler/permission_matrix_test.go`): 12 endpoints × 2 cases (has/no permission), no DB, mock services only.
2. **I-A/I-B/I-C — Integration permission tests** (`backend/tests/integration/rbac_permission_test.go`): preset role matrix, custom role lifecycle, boundary scenarios using real SQLite + full router.
3. **I-D — Coverage gate** (`backend/tests/integration/permission_coverage_test.go`): Go test that asserts every code in `permissions.AllCodes()` appears in the explicit `testedCodes` map; fails CI when a new code is added without a test.

No new CI step is required — all tests run under the existing `go test ./...` step in `ci.yml`.

## Architecture

### Layer Placement

```
Router → TeamScopeMiddleware (sets permCodes) → RequirePermission → Handler
                                                        ↑
                                               U1 tests inject permCodes here
                                               (package handler, no DB)

Integration tests (I-A/I-B/I-C/I-D):
  setupRBACTestDB → setupRBACTestRouter → HTTP → assert status code
  (package integration, real SQLite, full middleware chain)
```

U1 tests live in `package handler` (not `package middleware`) to avoid the import cycle: `middleware` is imported by `handler`, so handler-level tests cannot live in the middleware package.

### Component Diagram

```
+------------------------------------------+
|  backend/internal/handler/               |
|  permission_matrix_test.go               |
|  ┌──────────────────────────────────┐    |
|  │ buildPermTestRouter(code, codes) │    |
|  │  ├─ inject permCodes middleware  │    |
|  │  ├─ RequirePermission(code, ...) │    |
|  │  └─ mock handler stub            │    |
|  └──────────────────────────────────┘    |
|  12 TestPermMatrix_* functions           |
+------------------------------------------+

+------------------------------------------+
|  backend/tests/integration/              |
|  rbac_permission_test.go                 |
|  ┌──────────────────────────────────┐    |
|  │ setupRBACTestDB (existing)       │    |
|  │ setupRBACTestRouter (existing)   │    |
|  │ seedPermMatrixFixtures (new)     │    |
|  └──────────────────────────────────┘    |
|  TestRBACPermMatrix_PresetRoles          |
|  TestCustomRole_PartialPermissions       |
|  TestPermBoundary_EmptyRole              |
|  TestPermBoundary_SuperAdminBypass       |
|  TestPermBoundary_InvalidToken401        |
+------------------------------------------+

+------------------------------------------+
|  backend/tests/integration/              |
|  permission_coverage_test.go             |
|  ┌──────────────────────────────────┐    |
|  │ permissions.AllCodes()           │    |
|  │ testedCodes map[string]bool      │    |
|  │ diff → t.Errorf if missing       │    |
|  └──────────────────────────────────┘    |
|  TestPermissionCodeCoverage              |
+------------------------------------------+
```

### Dependencies

| Dependency | Type | Used by |
|---|---|---|
| `pm-work-tracker/backend/internal/middleware` | internal | U1 (`RequirePermission`) |
| `pm-work-tracker/backend/internal/handler` | internal | U1 (handler structs + mock interfaces) |
| `pm-work-tracker/backend/internal/pkg/permissions` | internal | I-D (`AllCodes()`) |
| `github.com/gin-gonic/gin` | external | U1 (test router) |
| `github.com/stretchr/testify` | external | all test assertions |
| `net/http/httptest` | stdlib | U1 + integration |

No new external dependencies.

## Interfaces

### U1: `buildPermTestRouter` helper

```go
// buildPermTestRouter creates a minimal Gin router for permission boundary testing.
// permCodes is injected into context before RequirePermission runs.
// handlerFn is a stub that returns 200 OK when reached.
func buildPermTestRouter(
    code      string,
    permCodes []string,
    handlerFn gin.HandlerFunc,
) *gin.Engine
```

### U1: Mock service interfaces (per handler)

Each mock implements the minimal interface method needed for the handler stub to return 200. Pattern follows existing mocks in `handler/*_test.go`:

```go
type mockMainItemService struct {
    mock.Mock
}
func (m *mockMainItemService) Archive(ctx context.Context, teamID, itemID uint) error {
    return m.Called(ctx, teamID, itemID).Error(0)
}
// ... other methods return zero values (not called in permission tests)
```

Full list of mock types needed (one per handler under test):

| Mock type | Minimal method | Used by test |
|---|---|---|
| `mockMainItemSvc` | `Create`, `Archive`, `ChangeStatus` | `TestPermMatrix_MainItem*` |
| `mockTeamSvc` | `InviteMember`, `RemoveMember`, `TransferPM` | `TestPermMatrix_Team*` |
| `mockProgressSvc` | `Append`, `CorrectCompletion` | `TestPermMatrix_Progress*` |
| `mockItemPoolSvc` | `Submit`, `Assign` | `TestPermMatrix_ItemPool*` |
| `mockViewSvc` | `Weekly` | `TestPermMatrix_WeeklyView` |
| `mockReportSvc` | `WeeklyExport` | `TestPermMatrix_ReportExport` |

### I-A: `seedPermMatrixFixtures` helper

```go
// seedPermMatrixFixtures creates the minimum DB records needed for the 5 I-A endpoints
// to return 200/403 (not 404/500) for superadmin and pm.
// Returns bizKeys for use in request URLs.
func seedPermMatrixFixtures(t *testing.T, db *gorm.DB, data *seedData) permMatrixFixtures

type permMatrixFixtures struct {
    mainItemBizKey  int64  // for archive endpoint
    progressBizKey  int64  // for progress:update endpoint
    poolItemBizKey  int64  // for item_pool:review endpoint
}
```

### I-D: `TestPermissionCodeCoverage`

```go
// testedCodes is the explicit contract: every code listed here must have
// a corresponding test in permission_matrix_test.go or rbac_permission_test.go.
// When a new code is added to permissions/codes.go, add it here AND write a test.
var testedCodes = map[string]bool{
    // team
    "team:create": true, "team:read": true, "team:update": true,
    "team:delete": true, "team:invite": true, "team:remove": true, "team:transfer": true,
    // main_item
    "main_item:create": true, "main_item:read": true, "main_item:update": true,
    "main_item:archive": true, "main_item:change_status": true,
    // sub_item
    "sub_item:create": true, "sub_item:read": true, "sub_item:update": true,
    "sub_item:assign": true, "sub_item:change_status": true,
    // progress
    "progress:create": true, "progress:read": true, "progress:update": true,
    // item_pool
    "item_pool:submit": true, "item_pool:review": true,
    // view
    "view:weekly": true, "view:gantt": true, "view:table": true,
    // report
    "report:export": true,
    // user
    "user:read": true, "user:update": true, "user:manage_role": true,
}

func TestPermissionCodeCoverage(t *testing.T) {
    allCodes := permissions.AllCodes()
    var missing []string
    for code := range allCodes {
        if !testedCodes[code] {
            missing = append(missing, code)
        }
    }
    sort.Strings(missing)
    if len(missing) > 0 {
        t.Errorf("missing test coverage for: %v", missing)
    }
}
```

## Data Models

This feature introduces no new data models. All test data uses existing models:

```go
// Existing models used in test fixtures (no changes):
model.Role            { ID, BizKey, Name, IsPreset, ... }
model.RolePermission  { RoleID uint, PermissionCode string }
model.TeamMember      { TeamKey, UserKey int64, RoleKey *int64, JoinedAt time.Time }
model.User            { ID, Username, PasswordHash string, IsSuperAdmin bool }
model.MainItem        { ID, BizKey int64, TeamKey int64, Title, Priority, ItemStatus string }
model.ItemPool        { ID, BizKey int64, TeamKey int64, Title, PoolStatus string }
model.ProgressRecord  { BizKey int64, SubItemKey, TeamKey int64, Completion float64 }
```

### Test case shape (U1)

```go
// Table-driven test case used in all TestPermMatrix_* functions
struct {
    name      string    // "has permission" | "no permission"
    permCodes []string  // injected into context
    wantCode  int       // http.StatusOK | http.StatusForbidden
}
```

### Integration test matrix shape (I-A)

```go
// Row in the preset role matrix table
struct {
    role     string  // "superadmin" | "pm" | "member"
    password string
    method   string  // HTTP method
    path     string  // URL with teamId/itemId substituted
    body     string  // JSON body or ""
    wantCode int     // 200 | 403
}
```

## Error Handling

This feature adds test code only. No new error types are introduced.

### Existing error codes exercised by tests

| Error Code | HTTP Status | When triggered |
|---|---|---|
| `ERR_FORBIDDEN` | 403 | `RequirePermission` rejects request (missing permission code) |
| `ERR_UNAUTHORIZED` | 401 | `AuthMiddleware` rejects invalid/missing JWT |
| `NOT_TEAM_MEMBER` | 403 | `TeamScopeMiddleware` rejects non-member (distinct from permission 403) |

### Propagation strategy (existing, verified by tests)

```
AuthMiddleware → 401 if token invalid/missing
TeamScopeMiddleware → 403 (NOT_TEAM_MEMBER) if user not in team
RequirePermission → 403 (ERR_FORBIDDEN) if code not in permCodes
Handler → 200/404/500 based on business logic
```

Tests assert the correct layer fires the correct code. `TestPermBoundary_InvalidToken401` specifically verifies 401 ≠ 403.

## Cross-Layer Data Map

Single-layer feature (test code only). Cross-Layer Data Map not applicable.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|---|---|---|---|---|
| Middleware+Handler boundary | Unit (no DB) | `httptest`, `gin`, `testify/mock` | `RequirePermission` blocks/passes for each of 12 endpoints | 12 endpoints × 2 cases = 24 cases, 100% of target endpoints |
| Full stack (preset roles) | Integration | real SQLite, `httptest` | superadmin/pm → 200, member → 403 for 5 representative endpoints | 3 roles × 5 endpoints = 15 assertions |
| Full stack (custom role) | Integration | real SQLite, `httptest` | partial perms + immediate effect after role update | 5 assertions across 2 phases |
| Full stack (boundary) | Integration | real SQLite, `httptest` | empty role 403, superadmin bypass 200, invalid token 401 | 3 assertions |
| Coverage gate | Unit | `permissions.AllCodes()`, `testify` | all 29 permission codes appear in `testedCodes` | 100% of `codes.go` entries |

### Key Test Scenarios

**U1 — permission boundary (per endpoint):**
- `permCodes = ["<required_code>"]` → handler stub called → 200
- `permCodes = []` → `RequirePermission` aborts → 403 with `ERR_FORBIDDEN`

**I-A — preset role matrix:**
- superadmin: `POST /main-items/:id/archive` → 200 (fixture exists)
- pm: same → 200
- member: same → 403

**I-B — custom role lifecycle:**
- Phase 1: role has `main_item:read` only → GET 200, POST 403
- Phase 2: add `main_item:create` to role → POST 200 (same token, no re-login)

**I-C — boundary:**
- Empty role (no permissions) → any protected endpoint → 403
- superadmin → any protected endpoint → 200 (not 404/500)
- Invalid JWT → any endpoint → 401 (not 403)

**I-D — coverage gate:**
- All 29 codes in `permissions.AllCodes()` present in `testedCodes` → pass
- Any code missing → `t.Errorf("missing test coverage for: [code]")`

### Overall Coverage Target

100% of the 12 target endpoints (U1) and 100% of permission codes in `codes.go` (I-D).

## Security Considerations

### Threat Model

This feature adds test code only. The threats it validates (not introduces) are:

| Threat | Validated by |
|---|---|
| Wrong permission code bound to route (typo/copy-paste) | U1: each endpoint tested with exact required code |
| Preset role missing a permission code | I-A: matrix asserts expected 200/403 per role |
| Permission change not taking effect immediately (caching bug) | I-B: phase 2 asserts immediate effect |
| 401 vs 403 conflation (auth failure masked as authz failure) | I-C: `TestPermBoundary_InvalidToken401` |
| New permission code added without test coverage | I-D: `TestPermissionCodeCoverage` fails CI |

### Mitigations

No new production security surface. Tests themselves are read-only against the DB (SQLite in-memory, per-test isolation via `setupRBACTestDB`).

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|---|---|---|
| Story 1: `permCodes=["x"]` → 200 | `TestPermMatrix_*` (has permission case) | `buildPermTestRouter` + mock handler |
| Story 1: `permCodes=[]` → 403 | `TestPermMatrix_*` (no permission case) | `RequirePermission` middleware |
| Story 2: superadmin/pm → 200, member → 403 | `TestRBACPermMatrix_PresetRoles` | `seedPermMatrixFixtures` + `setupRBACTestRouter` |
| Story 3: custom role partial perms | `TestCustomRole_PartialPermissions` (phase 1) | `model.RolePermission` insert |
| Story 3: permission change immediate effect | `TestCustomRole_PartialPermissions` (phase 2) | `model.RolePermission` update + same token |
| Story 4: empty role → 403 | `TestPermBoundary_EmptyRole` | `model.Role` with no `RolePermission` rows |
| Story 4: superadmin → 200 | `TestPermBoundary_SuperAdminBypass` | `model.User.IsSuperAdmin = true` |
| Story 4: invalid token → 401 | `TestPermBoundary_InvalidToken401` | malformed JWT + `AuthMiddleware` |
| Story 5: missing code → CI fail | `TestPermissionCodeCoverage` | `permissions.AllCodes()` diff vs `testedCodes` |
| Story 5: all codes covered → pass | `TestPermissionCodeCoverage` (green path) | `testedCodes` map completeness |

## Open Questions

- [x] ~~Where does I-D live — Go test or CI shell script?~~ → Go test in `tests/integration/`, runs under existing `go test ./...`
- [x] ~~Where is `codes.go`?~~ → `backend/internal/pkg/permissions/codes.go`, exposes `AllCodes()`
- [x] ~~Import cycle for U1?~~ → U1 lives in `package handler`, not `package middleware`

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|---|---|---|---|
| I-D as CI shell script (grep codes.go vs test files) | No Go code needed | Fragile string matching, counts comments/logs, hard to maintain | Explicit `testedCodes` map is the developer contract; Go test is more reliable |
| U1 in `package middleware` | Co-located with `permission.go` | Import cycle: `middleware` ← `handler`; can't wire real handlers | `package handler` avoids cycle |
| Separate `TestMain` with shared DB for I-A/I-B/I-C | Faster (one DB init) | Shared state between tests, harder to isolate failures | Per-test `setupRBACTestDB` matches existing pattern, isolation is more important |

### References

- `backend/internal/pkg/permissions/codes.go` — permission code registry (source of truth for I-D)
- `backend/internal/middleware/permission.go` — `RequirePermission` implementation
- `backend/internal/handler/router.go` — all 53 `perm()` bindings
- `backend/tests/integration/helpers.go` — `setupRBACTestDB`, `setupRBACTestRouter`, `makeRequest`
- `backend/tests/integration/rbac_test.go` — existing RBAC integration tests (pattern reference)
