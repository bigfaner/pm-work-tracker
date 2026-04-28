---
created: 2026-04-28
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: BizKey Unification

## Overview

Pure internal refactor — no API surface changes, no schema changes, no frontend changes. The goal is to make `int64` bizKey the only identifier that crosses the Middleware → Handler → Service boundary for team-scoped requests. Repository internals (`uint` FK lookups) are untouched.

Three concrete bugs are fixed as part of this change:
1. `progress_service.go:57` — `TeamKey: int64(teamID)` stores the wrong value in every progress record
2. `team_service.go:175` — `uint(roleID)` silently truncates a snowflake int64 in `InviteMember`
3. `team_service.go:305` — `isPMRole` is typed `uint` but semantically expects a bizKey

## Architecture

### Layer Placement

This change touches three layers:

```
Router (unchanged)
  ↓
Middleware  ← injects teamBizKey int64 instead of teamID uint
  ↓
Handler     ← reads teamBizKey int64 from context; passes int64 to service
  ↓
Service     ← interface params for teamID/roleID change uint → int64
  ↓
Repository  ← UNCHANGED: FindByID(uint) stays internal
```

### Boundary Rule

> If a value enters the system via an HTTP request (URL param, request body, query string), it must be `int64` at the Service interface boundary. If a value is an internal FK resolved inside a Repository, it stays `uint`.

| Value | Origin | Type at Service boundary |
|-------|--------|--------------------------|
| `teamBizKey` | URL param `:teamId` | `int64` |
| `roleBizKey` | Request body `roleKey` string | `int64` |
| `pmID` / `callerID` / `authorID` | JWT auth context (internal user ID) | `uint` (unchanged) |
| `subItemID` / `poolItemID` / `itemID` | URL param — **out of scope** | `uint` (unchanged, addressed separately) |

### Dependencies

No new dependencies. No package additions.

## Interfaces

### 1. Middleware context accessor

```go
// BEFORE
func GetTeamID(c *gin.Context) uint

// AFTER
func GetTeamBizKey(c *gin.Context) int64
```

`TeamScopeMiddleware` continues to call `teamRepo.FindByBizKey` to validate the team exists, but no longer resolves to `team.ID`. It injects `teamBizKey int64` directly:

```go
// BEFORE
c.Set("teamID", team.ID)   // uint

// AFTER
c.Set("teamBizKey", teamBizKey)  // int64
```

### 2. ProgressService

```go
// BEFORE
Append(ctx context.Context, teamID, authorID, subItemID uint, ...) (*model.ProgressRecord, error)
CorrectCompletion(ctx context.Context, teamID, recordID uint, completion float64) error
List(ctx context.Context, teamID, subItemID uint) ([]model.ProgressRecord, error)

// AFTER
Append(ctx context.Context, teamBizKey int64, authorID, subItemID uint, ...) (*model.ProgressRecord, error)
CorrectCompletion(ctx context.Context, teamBizKey int64, recordID uint, completion float64) error
List(ctx context.Context, teamBizKey int64, subItemID uint) ([]model.ProgressRecord, error)
```

Bug fix in `Append` implementation:
```go
// BEFORE (wrong — stores internal uint ID as team_key)
TeamKey: int64(teamID),

// AFTER (correct — stores the snowflake bizKey)
TeamKey: teamBizKey,
```

### 3. TeamService

```go
// BEFORE
UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID, roleID uint) error
isPMRole(ctx context.Context, bizKey uint) bool  // private

// AFTER
UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID uint, roleBizKey int64) error
isPMRole(ctx context.Context, bizKey int64) bool  // private
```

`InviteMember` fix — remove `uint(roleID)` cast:
```go
// BEFORE
if roleID, err := pkg.ParseID(req.RoleKey); err == nil && s.isPMRole(ctx, uint(roleID)) {

// AFTER
if roleID, err := pkg.ParseID(req.RoleKey); err == nil && s.isPMRole(ctx, roleID) {
```

`UpdateMemberRole` fix — remove `int64(roleID)` cast:
```go
// BEFORE
roleKey := int64(roleID)
member.RoleKey = &roleKey

// AFTER
member.RoleKey = &roleBizKey
```

### 4. ViewService

```go
// BEFORE
WeeklyComparison(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResponse, error)
GanttView(ctx context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error)
TableView(ctx context.Context, teamID uint, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error)
TableExportCSV(ctx context.Context, teamID uint, filter dto.TableFilter) ([]byte, error)

// AFTER
WeeklyComparison(ctx context.Context, teamBizKey int64, weekStart time.Time) (*dto.WeeklyViewResponse, error)
GanttView(ctx context.Context, teamBizKey int64, filter dto.GanttFilter) (*dto.GanttResult, error)
TableView(ctx context.Context, teamBizKey int64, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error)
TableExportCSV(ctx context.Context, teamBizKey int64, filter dto.TableFilter) ([]byte, error)
```

### 5. ReportService

```go
// BEFORE
Preview(ctx context.Context, teamID uint, weekStart time.Time) (*dto.ReportPreview, error)
ExportMarkdown(ctx context.Context, teamID uint, weekStart time.Time) ([]byte, error)

// AFTER
Preview(ctx context.Context, teamBizKey int64, weekStart time.Time) (*dto.ReportPreview, error)
ExportMarkdown(ctx context.Context, teamBizKey int64, weekStart time.Time) ([]byte, error)
```

### 6. MainItemService, SubItemService, ItemPoolService

All `teamID uint` parameters that originate from the URL-scoped team context change to `teamBizKey int64`. The `callerID`, `pmID`, `itemID`, `assigneeID` parameters that are internal IDs remain `uint`.

```go
// MainItemService
// BEFORE
Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)

// AFTER
Create(ctx context.Context, teamBizKey int64, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
List(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)

// SubItemService
// BEFORE
Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error)
Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error)
List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)

// AFTER
Create(ctx context.Context, teamBizKey int64, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error)
Get(ctx context.Context, teamBizKey int64, itemID uint) (*model.SubItem, error)
List(ctx context.Context, teamBizKey int64, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)

// ItemPoolService
// BEFORE
Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error)

// AFTER
Submit(ctx context.Context, teamBizKey int64, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
List(ctx context.Context, teamBizKey int64, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
Get(ctx context.Context, teamBizKey int64, poolItemID uint) (*model.ItemPool, error)
```

### 7. Handler call sites (mechanical updates)

All 7 handler files replace:
```go
// BEFORE
teamID := middleware.GetTeamID(c)
svc.SomeMethod(ctx, teamID, ...)

// AFTER
teamBizKey := middleware.GetTeamBizKey(c)
svc.SomeMethod(ctx, teamBizKey, ...)
```

`team_handler.go` additionally passes `roleBizKey int64` directly from the parsed request body to `UpdateMemberRole` and `InviteMember`.

### Complete change surface

| File | Change type |
|------|-------------|
| `middleware/team_scope.go` | Inject `teamBizKey int64`; remove `team.ID` resolution |
| `handler/progress_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/main_item_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/sub_item_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/item_pool_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/view_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/report_handler.go` | Replace `GetTeamID` → `GetTeamBizKey` |
| `handler/team_handler.go` | Replace `GetTeamID` → `GetTeamBizKey`; pass `roleBizKey int64` to `UpdateMemberRole` / `InviteMember` |
| `service/progress_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64`; fix `TeamKey` assignment |
| `service/main_item_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64` |
| `service/sub_item_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64` |
| `service/item_pool_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64` |
| `service/view_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64` |
| `service/report_service.go` | Interface + impl: `teamID uint` → `teamBizKey int64` |
| `service/team_service.go` | `roleID uint` → `roleBizKey int64`; fix `isPMRole` cast |

## Data Models

No model changes. `ProgressRecord.TeamKey` is already `int64` — the bug was in the assignment, not the field type.

## Error Handling

No new error types are introduced. The relevant existing error types and their HTTP mappings are:

### Service → Handler → HTTP status map

| Error | Source | Handler response | HTTP status |
|-------|--------|-----------------|-------------|
| `apperrors.ErrCannotAssignPMRole` | `TeamService.UpdateMemberRole` / `InviteMember` | `{"error": "cannot assign pm role"}` | 403 Forbidden |
| `gorm.ErrRecordNotFound` (from `FindByBizKey`) | any service doing a bizKey lookup | `{"error": "not found"}` | 404 Not Found |
| `FindByBizKey(0)` → `gorm.ErrRecordNotFound` | routing misconfiguration (middleware not registered) | `{"error": "not found"}` | 404 Not Found |
| DB error (non-not-found) | any repo call | `{"error": "internal error"}` | 500 Internal Server Error |

These errors are translated by the existing handler error-mapping helper (`pkg/apperrors` → HTTP status). This refactor does not change that mapping — it only changes the type of the `teamBizKey` parameter flowing into the service calls that may produce these errors.

### GetTeamBizKey failure behavior

`GetTeamBizKey` uses the comma-ok type assertion to avoid a panic on missing or wrong-type context key:

```go
func GetTeamBizKey(c *gin.Context) int64 {
    val, exists := c.Get("teamBizKey")
    if !exists {
        return 0
    }
    bizKey, ok := val.(int64)
    if !ok {
        return 0
    }
    return bizKey
}
```

Returning `0` on a missing/wrong-type key is safe because `TeamScopeMiddleware` always runs before any handler that calls `GetTeamBizKey` — if the middleware aborts, the handler never executes. A `0` reaching a service call indicates a routing misconfiguration (middleware not registered), not a user error; the service will return a not-found error on the downstream `FindByBizKey(0)` call.

### Middleware failure → HTTP status map

| Failure condition | HTTP status | Abort body |
|-------------------|-------------|------------|
| `:teamId` param missing or empty | 400 Bad Request | `{"error": "invalid team id"}` |
| `pkg.ParseID` fails (non-numeric string) | 400 Bad Request | `{"error": "invalid team id"}` |
| `teamRepo.FindByBizKey` returns not found | 404 Not Found | `{"error": "team not found"}` |
| `teamRepo.FindByBizKey` returns DB error | 500 Internal Server Error | `{"error": "internal error"}` |

These are the existing middleware abort paths — this refactor does not add new ones. The only change is that the middleware no longer resolves `team.ID uint` after a successful lookup; it injects `teamBizKey int64` directly and aborts on the same conditions as before.

The `isPMRole` helper returns `false` on `FindByBizKey` error (role not found) — this behavior is preserved.

## Cross-Layer Data Map

| Field | URL Param | Middleware context | Handler local var | Service param | Repository |
|-------|-----------|-------------------|-------------------|---------------|------------|
| team identifier | `:teamId` string | `teamBizKey int64` | `teamBizKey int64` | `teamBizKey int64` | `FindByBizKey(int64)` internally resolves to `uint` for FK ops |
| role identifier | request body `roleKey` string | — | `roleBizKey int64` (parsed) | `roleBizKey int64` | `FindByBizKey(int64)` |
| user/caller identifier | JWT claim | `userID uint` | `callerID/pmID uint` | `callerID/pmID uint` | `FindByID(uint)` |

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | What to Test | Files |
|-------|-----------|--------------|-------|
| Middleware | Unit | `GetTeamBizKey` returns `int64`; `GetTeamID` removed | `team_scope_test.go` |
| Service | Unit | All mock interfaces updated to `int64` params; `TeamKey` value in progress record equals input `teamBizKey` | 7 service test files |
| Handler | Unit | `GetTeamBizKey()` called instead of `GetTeamID()`; correct type passed to mock service | 10 handler test files |
| Integration | Integration | Progress record `team_key` matches team bizKey; role assignment rejects PM role correctly | `views_reports_test.go`, `helpers.go` |

### Key Test Scenarios

1. **TeamKey correctness**: `progressService.Append` with `teamBizKey = 123456789012345678` → stored record has `TeamKey == 123456789012345678`
2. **isPMRole with int64**: `UpdateMemberRole` with PM role's bizKey → returns `ErrCannotAssignPMRole`
3. **Compile gate**: `go build ./...` passes with zero errors — compiler enforces no `uint(bizKey)` casts remain
4. **Middleware context type**: `GetTeamBizKey(c)` returns `int64`; calling `GetTeamID` is a compile error

### Test Tooling

| Tool | Purpose |
|------|---------|
| `go test -race` | All unit and integration tests; detects data races |
| `github.com/stretchr/testify/assert` | Non-fatal assertions |
| `github.com/stretchr/testify/require` | Fatal assertions (stops test on first failure) |
| `net/http/httptest` | Handler tests via `httptest.NewRecorder()` |
| `gin.CreateTestContext` | Gin context construction in handler unit tests |

### Overall Coverage Target

Maintain ≥ 75% statement coverage across `internal/service` and `internal/handler` packages (current baseline). No regression permitted — CI runs `go test -race -coverprofile=coverage.out ./internal/...` and fails if coverage drops below this threshold.

## Security Considerations

`bizKey` is already the externally-exposed identifier. Internal `uint` IDs remain hidden — this change reinforces that boundary.

### Threat Model

**Threat 1: Cross-team access via crafted bizKey**

A caller presents a valid bizKey for team A while their JWT belongs to team B.

`TeamScopeMiddleware` calls `teamRepo.FindByBizKey(bizKey)` to confirm the team exists, then aborts with 404 if not found. However, team membership authorization (does this user belong to this team?) is enforced by a separate RBAC permission check downstream — that check is on `callerID uint` from the JWT, which is unaffected by this refactor. The type change `uint → int64` for `teamBizKey` does not alter the middleware's abort logic or the downstream RBAC lookup path. No new cross-team access vector is introduced.

**Threat 2: Privilege escalation via isPMRole fix**

Before this fix, `isPMRole(ctx, uint(roleID))` silently truncated a snowflake int64 to uint, meaning a PM role's bizKey could fail to match and the PM-role guard would be bypassed. The fix passes the full int64 bizKey to `FindByBizKey`, making the guard more restrictive, not less. A role that previously slipped past the truncation check will now be correctly identified as PM and rejected. No privilege escalation risk.

**Threat 3: Zero bizKey reaching service layer**

As documented in Error Handling, `GetTeamBizKey` returns `0` only if the middleware was not registered (routing misconfiguration). In that case no team membership check has run. This is a deployment/configuration concern, not a new attack surface introduced by this refactor — the same gap existed with `GetTeamID` returning `0`.

## PRD Coverage Map

| PRD Acceptance Criterion | Design Component | Interface / Model |
|--------------------------|------------------|-------------------|
| Progress record `team_key` equals snowflake bizKey | `ProgressService.Append` implementation | `ProgressRecord.TeamKey = teamBizKey` |
| `isPMRole` correctly identifies PM role via bizKey | `teamService.isPMRole(int64)` | `roleRepo.FindByBizKey(bizKey)` |
| Compiler rejects `uint` where `int64` expected | All service interface signatures | Go type system |
| `GetTeamBizKey() int64` replaces `GetTeamID()` | `middleware/team_scope.go` | `c.Get("teamBizKey").(int64)` |
| Zero `uint(bizKey)` casts in service/handler | All 7 handler files + 8 service files | Verified by `grep` |
| `UpdateMemberRole` and `InviteMember` use `int64` for role | `TeamService` interface | `roleBizKey int64` param |

## Open Questions

- [x] Should `subItemID`, `poolItemID`, `itemID` also migrate to `int64`? → **Out of scope**, addressed separately per PRD
- [x] Should `pmID`/`callerID`/`authorID` (from JWT) also migrate? → **Out of scope**, these are internal IDs from auth context

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Targeted: fix only 3 known bug sites | Minimal diff | Doesn't prevent recurrence; next developer repeats the mistake | Treats symptoms, not cause |
| Remove `FindByID` entirely | Simplest contract | Breaks internal FK joins, hurts performance | Over-correction |
| New wrapper type `type BizKey int64` | Stronger type safety | Requires changes to all model fields and GORM tags | Scope too large; int64 is sufficient |
