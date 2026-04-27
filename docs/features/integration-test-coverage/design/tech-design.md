---
created: 2026-04-27
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Integration Test Coverage

## Overview

Add 5 integration test files and 1 shared helpers file to `backend/tests/integration/`, plus 6 unit test gap fixes in existing `*_test.go` files. All tests use the existing in-memory SQLite + real DI router pattern. No new production code.

## Architecture

### Layer Placement

Single-layer feature: test code only. Changes confined to `backend/tests/integration/` (integration tests) and `backend/internal/*/` (unit test gap fixes). Zero production code changes.

### Component Diagram

```
backend/tests/integration/
├── helpers.go                    ← F7: extracted shared helpers (NEW)
├── auth_isolation_test.go        ← existing (helper definitions REMOVED)
├── progress_completion_test.go   ← existing (helper definitions REMOVED)
├── rbac_test.go                  ← existing (helper definitions REMOVED)
├── team_transfer_pm_test.go      ← existing (minimal, no helpers to move)
├── item_lifecycle_test.go        ← F1: NEW
├── item_pool_test.go             ← F2: NEW
├── team_management_test.go       ← F3: NEW
├── admin_user_test.go            ← F4: NEW
└── views_reports_test.go         ← F5: NEW

backend/internal/handler/permission_handler_test.go  ← F6: NEW
backend/internal/service/  (*_test.go edits)          ← F6: edits to existing
```

### Dependencies

- **Existing**: `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `github.com/gin-gonic/gin`, `net/http/httptest`, `gorm.io/gorm`, `encoding/json`
- **Internal packages referenced**: `backend/internal/handler` (`handler.Dependencies`, `handler.SetupRouter`), `backend/internal/pkg/jwt` (`appjwt.Claims`), `backend/internal/dto` (`dto.MainItemCreateReq`), `backend/internal/model`, `backend/internal/config`
- **New**: None. All dependencies already in `go.mod`.

## Interfaces

### Helper Function Signatures (helpers.go)

Extracted from existing test files, preserving exact signatures. All functions are in `package integration`.

```go
// From auth_isolation_test.go
func setupTestDB(t *testing.T) (*gorm.DB, *seedData)
func setupTestRouter(t *testing.T) (*gin.Engine, *seedData)
func loginAs(t *testing.T, r *gin.Engine, username, password string) string
func signTokenWithClaims(t *testing.T, claims *appjwt.Claims) string

// From progress_completion_test.go
func seedProgressData(t *testing.T, db *gorm.DB, teamID, userID uint) (mainItemID, subItem1ID, subItem2ID uint, subItem1BizKey, subItem2BizKey int64)
func appendProgress(t *testing.T, r *gin.Engine, token string, teamBizKey, subBizKey int64, completion float64) *httptest.ResponseRecorder
func seedPoolData(t *testing.T, db *gorm.DB, teamID, userID uint) (poolID, mainItemID uint, poolBizKey, mainItemBizKey int64)
func seedReportData(t *testing.T, db *gorm.DB, teamID, userID uint, weekStart time.Time) string

// From rbac_test.go
func makeRequest(t *testing.T, r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder

// New helpers (extracted from F1 patterns)
func createTeamWithMembers(t *testing.T, db *gorm.DB, pmID uint, memberCount int) uint
func createMainItem(t *testing.T, r *gin.Engine, token string, teamBizKey int64, req dto.MainItemCreateReq) int64
```

### Unified Setup Function

The existing three setup variants (`setupTestRouter`, `setupTestRouterWithDB`, `setupRBACTestRouter`) share ~40 lines of identical DI wiring. Consolidate into a single internal helper:

```go
// wireHandlers creates all repos+services+handlers and returns *handler.Dependencies.
// Called by all setup* functions.
func wireHandlers(t *testing.T, db *gorm.DB, data *seedData) *handler.Dependencies
```

Existing public signatures (`setupTestRouter`, etc.) remain unchanged — they call `wireHandlers` internally.

### Test Request Pattern

All new test files use the same request pattern:

```go
// makeRequest sends an HTTP request and returns the response recorder.
// Body is a JSON string (empty string for no body).
// Token is the JWT (empty string for unauthenticated requests).
func makeRequest(t *testing.T, r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder
```

Already exists in `rbac_test.go`, will be moved to `helpers.go`.

## Data Models

### seedData Struct

Existing struct, shared across all test files:

```go
type seedData struct {
    userAID, userBID, memberAID, superAdminID uint
    teamAID, teamBID                          uint
    teamABizKey, teamBBizKey                  int64
    pmRoleID, memberRoleID, superAdminRoleID  uint  // added for RBAC tests
}
```

### Status Machine Reference

From `status/transition.go`. Test writers use this to determine valid/invalid transitions:

**MainItem**: pending → progressing/closed, progressing → blocking/pausing/reviewing/closed, blocking → progressing, pausing → progressing/closed, reviewing → completed/progressing. Terminal: completed, closed.

**SubItem**: pending → progressing/closed, progressing → blocking/pausing/completed/closed, blocking → progressing, pausing → progressing/closed. Terminal: completed, closed.

Key difference: SubItem can go progressing → completed directly; MainItem must go reviewing → completed.

## Error Handling

### Error Assertion Pattern

Tests assert on HTTP status codes and the structured error response defined in `backend/internal/pkg/errors/`. The canonical error response shape is:

```json
{
  "error": "ERR_CODE",
  "message": "Human-readable description"
}
```

Specific error codes (e.g., `PROGRESS_REGRESSION`, `ERR_ITEM_NOT_FOUND`, `ERR_ROLE_IN_USE`) are defined in `backend/internal/pkg/errors/codes.go`. Tests assert on the `error` field value when the code is business-meaningful, and on the HTTP status code otherwise:

```go
assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
var resp map[string]interface{}
json.Unmarshal(w.Body.Bytes(), &resp)
assert.Equal(t, "PROGRESS_REGRESSION", resp["error"])
```

### Bug Severity Threshold

| Severity | Criteria | Action |
|----------|----------|--------|
| Critical | Data loss, wrong status transition accepted, auth bypass | Fix immediately in same PR |
| Non-critical | Wrong error message format, minor response field mismatch | File bug, continue |

## Cross-Layer Data Map

Single-layer feature — not applicable. All work is in test files.

## Testing Strategy

### DB Isolation Strategy (PRD Deviation)

**PRD specifies**: "使用事务回滚（tx.Begin() + t.Cleanup(tx.Rollback)）"
**Design uses**: Per-test in-memory SQLite databases via `fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())`

**Rationale for deviation**: Transaction rollback with GORM on SQLite shared-cache has proven fragile in this codebase. The existing 39 integration tests already use the per-test DB pattern successfully, and it provides stronger isolation (each test gets a completely fresh schema). See Alternatives Considered below.

Each test creates a unique in-memory SQLite DB via `fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())`. No cleanup needed — DB is scoped to the test name and dies when the process exits.

Tests do NOT share databases. Each test function gets its own `setupTestDB` call.

### Test Naming Convention

```
Test<FlowName>_<Action>_<ExpectedOutcome>

Examples:
TestItemLifecycle_CreateMainItem_Returns201
TestItemLifecycle_CreateMainItem_MissingTitle_Returns422
TestItemLifecycle_ChangeStatus_InvalidTransition_Returns422
TestItemLifecycle_AppendProgress_CompletionRollsUpToMainItem
TestItemPool_Assign_CreatesSubItemAndUpdatesPoolStatus
TestItemPool_Assign_InvalidMainItem_RollsBack
TestTeamManagement_RemovePM_Returns422
TestAdminUser_DisableSelf_Returns422
```

### Test File Structure

Each flow file follows this pattern:

```go
package integration

func TestItemLifecycle_CreateMainItem_Returns201(t *testing.T) {
    r, data := setupTestRouter(t)
    token := loginAs(t, r, "userA", "passwordA")

    body := `{"title":"Item 1","priority":"P0",...}`
    w := makeRequest(t, r, http.MethodPost, "/api/v1/teams/"+strconv.FormatInt(data.teamABizKey, 10)+"/main-items", body, token)

    assert.Equal(t, http.StatusCreated, w.Code)
    // assert response body...
}
```

### Per-Flow Test Count Estimates

| Flow | File | Estimated Tests | Key Scenarios |
|------|------|----------------|---------------|
| F1 | `item_lifecycle_test.go` | ~50 | 17 endpoints x ~3 cases each, plus cascade scenarios |
| F2 | `item_pool_test.go` | ~20 | 6 endpoints x ~3 cases, plus rollback and 409 |
| F3 | `team_management_test.go` | ~30 | 9 endpoints x ~3 cases, plus PM protection |
| F4 | `admin_user_test.go` | ~20 | 6 endpoints x ~3 cases, plus self-disable |
| F5 | `views_reports_test.go` | ~25 | 6 endpoints x ~4 cases (happy/empty/format/auth) |
| F6 | Unit gaps | ~15 | 6 gaps x ~2 cases each |
| **Total** | | **~160** | |

### F6 Unit Test Gap Strategy

Unit test gaps use the existing handler/service test patterns:

| Gap | Test File | Pattern |
|-----|-----------|---------|
| `permission_handler.go` | New `permission_handler_test.go` in `handler/` package | `httptest` + mock `RoleService` interface |
| `ConvertToMain` | Edit `item_pool_service_test.go` in `service/` package | Mock `mainItemRepo` + `itemPoolRepo`, verify transaction |
| `UpdateTeam` | Edit `team_service_test.go` in `service/` package | Mock `teamRepo`, verify PM check + field update |
| `GetByBizKey` (x3) | Edit respective `*_service_test.go` files | Mock repo, test found/not-found |

### Overall Coverage Target

- 54/54 API endpoints with integration tests (100%)
- 150+ new test cases
- 6 unit test gaps closed
- Suite execution < 150s

## Security Considerations

### Threat Model

No new production attack surface. Test-only changes.

### Mitigations

- Tests use in-memory SQLite only — never connect to production databases
- JWT test secret is a hardcoded constant (`test-secret-that-is-at-least-32-bytes!!`), only used in test context
- Test users are ephemeral — no credentials persist

## PRD Coverage Map

| PRD AC | Design Component | File |
|--------|------------------|------|
| S1: PM creates MainItem → 201 | `TestItemLifecycle_CreateMainItem_Returns201` | `item_lifecycle_test.go` |
| S1: SubItem with weight → linked | `TestItemLifecycle_CreateSubItem_TracksWeight` | `item_lifecycle_test.go` |
| S1: Progress → completion rollup | `TestItemLifecycle_AppendProgress_CompletionRollsUp` | `item_lifecycle_test.go` |
| S1: Valid status transition → cascade | `TestItemLifecycle_ChangeStatus_TerminalCascade` | `item_lifecycle_test.go` |
| S1: Invalid transition → 422 | `TestItemLifecycle_ChangeStatus_InvalidTransition_Returns422` | `item_lifecycle_test.go` |
| S1: Archive completed → 200 | `TestItemLifecycle_Archive_CompletedItem_Returns200` | `item_lifecycle_test.go` |
| S1: Member denied → 403 | `TestItemLifecycle_CreateMainItem_MemberDenied_Returns403` | `item_lifecycle_test.go` |
| S2: Submit pool item → 201 | `TestItemPool_Submit_Returns201` | `item_pool_test.go` |
| S2: Assign to main → SubItem created | `TestItemPool_Assign_CreatesSubItem` | `item_pool_test.go` |
| S2: Invalid main → rollback | `TestItemPool_Assign_InvalidMainItem_RollsBack` | `item_pool_test.go` |
| S2: Already processed → 409 | `TestItemPool_Assign_AlreadyProcessed_Returns409` | `item_pool_test.go` |
| S2: Reject with reason → 200 | `TestItemPool_Reject_WithReason_Returns200` | `item_pool_test.go` |
| S2: Reject no reason → 422 | `TestItemPool_Reject_NoReason_Returns422` | `item_pool_test.go` |
| S3: Create team → auto PM | `TestTeamManagement_CreateTeam_AutoJoinAsPM` | `team_management_test.go` |
| S3: Invite with role → correct perms | `TestTeamManagement_InviteMember_HasCorrectRole` | `team_management_test.go` |
| S3: Change role → immediate effect | `TestTeamManagement_ChangeRole_ImmediateEffect` | `team_management_test.go` |
| S3: Disband → cascade delete | `TestTeamManagement_Disband_CascadeDeletes` | `team_management_test.go` |
| S3: Remove PM → 422 | `TestTeamManagement_RemovePM_Returns422` | `team_management_test.go` |
| S3: Member denied → 403 | `TestTeamManagement_Invite_MemberDenied_Returns403` | `team_management_test.go` |
| S4: Create user → 201 | `TestAdminUser_Create_Returns201` | `admin_user_test.go` |
| S4: Duplicate username → 409 | `TestAdminUser_Create_DuplicateUsername_Returns409` | `admin_user_test.go` |
| S4: Disable user → 200 | `TestAdminUser_ToggleStatus_Disable_Returns200` | `admin_user_test.go` |
| S4: Disable self → 422 | `TestAdminUser_ToggleStatus_DisableSelf_Returns422` | `admin_user_test.go` |
| S4: Non-admin → 403 | `TestAdminUser_Create_NonAdmin_Returns403` | `admin_user_test.go` |
| S5: Weekly stats correct | `TestViews_Weekly_CorrectStats` | `views_reports_test.go` |
| S5: Table filter works | `TestViews_Table_FilterByStatus` | `views_reports_test.go` |
| S5: CSV export has BOM | `TestViews_TableExport_HasBOM` | `views_reports_test.go` |
| S5: Empty team → zeros | `TestViews_Weekly_EmptyTeam_ReturnsZeroStats` | `views_reports_test.go` |
| S5: Report preview has sections | `TestReports_WeeklyPreview_ContainsSections` | `views_reports_test.go` |
| S6: Extract 8 existing helpers | `helpers.go` with all moved functions | `helpers.go` |
| S6: Extract 2 new helpers | `createTeamWithMembers`, `createMainItem` | `helpers.go` |
| S6: No behavior change | All existing tests pass after extraction | — |
| S7: Function names have business semantics | Naming convention enforced | all `*_test.go` |
| S7: Adjacent functions trace flow | Grouped by user flow in each file | all `*_test.go` |
| S7: 10 helpers in file | `helpers.go` contains all listed functions | `helpers.go` |
| S7: Duplicate definitions removed | Original files cleaned | `auth_isolation_test.go` etc. |
| S8: permission_handler tests | `permission_handler_test.go` | new file |
| S8: ConvertToMain test | `item_pool_service_test.go` edit | existing file |
| S8: UpdateTeam test | `team_service_test.go` edit | existing file |
| S8: GetByBizKey tests (x3) | respective `*_service_test.go` edits | existing files |

## Open Questions

None. All decisions resolved in the PRD evaluation.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Transaction rollback per test | Automatic cleanup | SQLite shared-cache + GORM tx support is fragile; existing tests already use per-test DB isolation | Existing pattern (unique in-memory DB per test) is simpler and proven |
| Testcontainers (real MySQL) | Tests against real DB dialect | CI complexity, slow startup (~5s per test), Docker dependency | In-memory SQLite is sufficient for integration testing; dialect-specific tests already exist in `dialect_sql_test.go` |
| Table-driven subtests | Fewer test functions | Harder to trace which endpoint failed in CI; flow-based tests need sequential state | Individual test functions with descriptive names are more reviewable |

### References

- `backend/tests/integration/auth_isolation_test.go` — existing helper pattern
- `backend/tests/integration/progress_completion_test.go` — seed data pattern
- `backend/tests/integration/rbac_test.go` — makeRequest pattern
- `backend/internal/pkg/status/transition.go` — status machine definition
- `backend/internal/handler/router.go` — Dependencies struct and route wiring
