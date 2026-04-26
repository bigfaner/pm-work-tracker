---
created: 2026-04-26
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Schema Alignment Post-Refactoring Cleanup

## Overview

This is a code cleanup initiative — no new features. The design specifies where to place shared helpers, how to consolidate duplicates, and the exact file changes for each of the 24 items. All changes fit within the existing layer architecture (Model → Repository → Service → Handler → VO/DTO).

## Architecture

### Layer Placement

All changes are confined to existing layers. No new layers or packages are needed. Shared helpers go into existing shared locations:

| New Helper | Location | Rationale |
|------------|----------|-----------|
| `DBTransactor` interface | `backend/internal/pkg/repo/transactor.go` | Adjacent to existing `helpers.go` in `pkg/repo/` |
| `ParseBizKeyParam` helper | `backend/internal/pkg/handler/bizkey.go` (new file) | New `pkg/handler/` sub-package for handler utilities |
| `FormatDateOnly` function | `frontend/src/lib/format.ts` | Extend existing shared format module |
| `RecordStatusChange` helper | `backend/internal/service/status_history_helper.go` (new file) | Service-layer helper, same package as callers |
| `UserToDTO` conversion | `backend/internal/vo/user_vo.go` (new file) | Follow existing VO pattern in `vo/` package |
| `TeamVO` struct + constructor | `backend/internal/vo/team_vo.go` (new file) | Follow existing VO pattern |

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│                   Handler Layer                  │
│  ┌────────────┐  ┌──────────────────────────┐   │
│  │ handler/*  │──│ pkg/handler.ParseBizKey  │   │
│  └────────────┘  └──────────────────────────┘   │
│          │                                       │
│          ▼                                       │
│  ┌──────────────────────────────────────────┐    │
│  │              VO Layer (vo/)               │    │
│  │  user_vo.go  team_vo.go  item_vo.go      │    │
│  └──────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│                  Service Layer                   │
│  ┌────────────────┐  ┌───────────────────────┐  │
│  │ service/*      │──│ status_history_helper  │  │
│  └────────────────┘  └───────────────────────┘  │
│          │                                       │
│          ▼                                       │
│  ┌──────────────────────────────────────────┐    │
│  │         pkg/repo (shared repo helpers)    │    │
│  │  helpers.go     transactor.go             │    │
│  └──────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│               Repository Layer                   │
│  ┌──────────────────────────────────────────┐    │
│  │  filter_helpers.go (fix type mismatch)    │    │
│  │  scopes.go (NotDeleted scope)             │    │
│  │  role_repo.go (remove column: tags)       │    │
│  └──────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│                 Model Layer                      │
│  ┌──────────────────────────────────────────┐    │
│  │  role.go (TableName pmw_roles)            │    │
│  └──────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│                Frontend                          │
│  ┌───────────────┐  ┌────────────────────────┐  │
│  │ types/index.ts│  │ lib/format.ts (+format) │  │
│  │ hooks/*.ts    │  │ api/client.ts (+toast)  │  │
│  │ pages/*.tsx   │  └────────────────────────┘  │
│  └───────────────┘                               │
└─────────────────────────────────────────────────┘
```

## Interfaces

### Interface 1: `DBTransactor`

Replaces both `TransactionDB` (team_service.go:19) and `dbTransactor` (item_pool_service.go:31).

```go
// pkg/repo/transactor.go
type DBTransactor interface {
    Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}
```

Both `team_service` and `item_pool_service` will depend on `repo.DBTransactor` instead of their local definitions.

### Interface 2: `ParseBizKeyParam`

Shared handler helper replacing 4 duplicate parse functions.

```go
// pkg/handler/bizkey.go
// ParseBizKeyParam extracts and validates a path parameter as int64 bizKey.
// Responds with ErrValidation on failure. Returns (bizKey, true) on success, (0, false) on failure.
func ParseBizKeyParam(c *gin.Context, paramName string) (int64, bool)
```

### Interface 3: `ResolveBizKey`

For handlers that need to look up an entity by bizKey to get the internal uint ID.

```go
// pkg/handler/bizkey.go
// ResolveBizKey parses the bizKey param, looks up the entity, returns internal uint ID.
// On parse failure: responds with ErrValidation (HTTP 400), returns (0, false).
// On lookup failure: if lookupFn returns apperrors.ErrNotFound, responds HTTP 404;
// any other error is propagated via RespondError (defaults to HTTP 500). Returns (0, false).
func ResolveBizKey(c *gin.Context, paramName string, lookupFn func(ctx context.Context, bizKey int64) (uint, error)) (uint, bool)
```

### Interface 4: `RecordStatusChange`

Service-layer helper replacing 5 duplicate status-history recording call sites.

```go
// service/status_history_helper.go
// RecordStatusChange creates a StatusHistory record if the recorder is non-nil.
// Returns the error from the recorder so callers can propagate DB failures.
func RecordStatusChange(
    recorder StatusHistoryRecorder,
    ctx context.Context,
    itemType string,
    itemKey int64,
    fromStatus, toStatus string,
    changedBy uint,
    isAuto int,
    remark string,
) error
```

Where `StatusHistoryRecorder` is:
```go
type StatusHistoryRecorder interface {
    Record(ctx context.Context, history *model.StatusHistory) error
}
```

When `recorder` is nil, returns nil (no-op). Otherwise returns the error from `recorder.Record`. All 5 call sites must check the returned error and propagate it up to the handler layer.

## Data Models

### Model 1: `TeamVO`

Replaces `team_handler.teamToDTO`'s untyped `gin.H`.

```go
// vo/team_vo.go
type TeamVO struct {
    BizKey      string  `json:"bizKey"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Code        string  `json:"code"`
    PmKey       string  `json:"pmKey"`
    CreatedAt   string  `json:"createdAt"`     // RFC3339 formatted
    UpdatedAt   string  `json:"updatedAt"`     // RFC3339 formatted
}

func NewTeamVO(t *model.Team) TeamVO {
    return TeamVO{
        BizKey:      pkg.FormatID(t.BizKey),
        Name:        t.TeamName,
        Description: t.TeamDesc,
        Code:        t.Code,
        PmKey:       pkg.FormatID(t.PmKey),
        CreatedAt:   t.CreateTime.Format(time.RFC3339),
        UpdatedAt:   t.DbUpdateTime.Format(time.RFC3339),
    }
}
```

### Model 2: `UserVO`

Shared base conversion for `auth_handler.userToDTO` and `admin_service.modelToAdminUserDTO`.

```go
// vo/user_vo.go
type UserVO struct {
    BizKey       string `json:"bizKey"`
    Username     string `json:"username"`
    DisplayName  string `json:"displayName"`
    Email        string `json:"email"`
    Status       string `json:"status"`
    IsSuperAdmin bool   `json:"isSuperAdmin"`
}

func NewUserVO(u *model.User) UserVO {
    return UserVO{
        BizKey:       pkg.FormatID(u.BizKey),
        Username:     u.Username,
        DisplayName:  u.DisplayName,
        Email:        u.Email,
        Status:       u.UserStatus,
        IsSuperAdmin: u.IsSuperAdmin,
    }
}
```

### Model 3: `Role` TableName update

```go
// model/role.go — only TableName methods change
func (Role) TableName() string         { return "pmw_roles" }
func (RolePermission) TableName() string { return "pmw_role_permissions" }
```

### Model 4: Frontend type updates

```ts
// types/index.ts
interface PermissionData {
  isSuperAdmin: boolean
  teamPermissions: Record<string, string[]>  // was Record<number, ...>
}

interface TableRow {
  mainItemId?: string | null  // was number | null
}
```

## Error Handling

No new error types. All changes use existing `apperrors` sentinel errors (`ErrValidation`, `ErrNotFound`).

The `ParseBizKeyParam` helper standardizes error handling: all parse failures respond with `apperrors.ErrValidation` (HTTP 400). Currently `parseUserBizKey` in admin_handler silently returns without responding — this will be fixed by using the shared helper.

### `ResolveBizKey` error contract

The `lookupFn` callback passed to `ResolveBizKey` must follow this contract:

| Condition | `lookupFn` returns | `ResolveBizKey` responds with | HTTP status |
|-----------|--------------------|-------------------------------|-------------|
| Entity not found | `apperrors.ErrNotFound` (or any error wrapping it) | `apperrors.ErrNotFound` | 404 |
| Other DB/service failure | Any other `error` | Propagated via `apperrors.RespondError` | 500 (default) |

This mirrors the existing pattern in `resolvePoolID`, `resolveSubID`, and `resolveUserBizKey` where service-layer `GetByBizKey` returns `apperrors.ErrNotFound` for missing entities and `RespondError` maps it to HTTP 404. Any unexpected error (connection failure, timeout) falls through to the default 500 response.

### RecordStatusChange error propagation

`RecordStatusChange` returns `error` instead of being void. The previous design silently swallowed DB write failures in status-history recording. Now each call site must check the returned error:

```go
if err := status_history.RecordStatusChange(recorder, ctx, ...); err != nil {
    return fmt.Errorf("record status change: %w", err)
}
```

This ensures a failed status-history insert does not go undetected. The service methods that call `RecordStatusChange` already return `error`, so propagation is a one-line addition at each of the 5 call sites.

### Filter parse failure handling (Item 2)

The `assigneeKey` filter in `filter_helpers.go` receives a `*string` from the DTO layer but compares against a `BIGINT` column. The current code passes the raw string into SQL without conversion. The fix must not silently skip the filter on parse failure — doing so returns all items instead of the filtered subset (an authorization bypass). On `strconv.ParseInt` failure, the query applies `WHERE 1 = 0` to guarantee zero results.

## Cross-Layer Data Map

Only the two P0 bugs involve cross-layer data changes:

| Field | Storage (DB) | Model (Go) | API/DTO | Frontend | Fix |
|-------|-------------|------------|---------|----------|-----|
| `assignee_key` in SubItem.Assign | `assignee_key BIGINT` | `*int64` | — | — | Column name fix: `"assignee_id"` → `"assignee_key"` |
| `assignee_key` in filter | `assignee_key BIGINT` | — | `*string` (DTO filter) | `string` (query param) | Type conversion: `pkg.ParseID(*string)` before SQL WHERE |

## Item-by-Item Implementation Guide

### Round 1: Bug Fixes

**Item 1** — `sub_item_service.go:262`: Change `"assignee_id"` to `"assignee_key"` in the `Assign` method's `UpdateFields` call map.

**Item 2** — `filter_helpers.go:15`: Add `strconv.ParseInt` conversion before the SQL query. On parse failure, apply a `WHERE 1=0` clause to return zero results instead of silently skipping the filter (which would return all items — an authorization bypass):
```go
if assigneeKey != nil && *assigneeKey != "" {
    ak, err := strconv.ParseInt(*assigneeKey, 10, 64)
    if err != nil {
        // Invalid assigneeKey: return empty result, never all items
        query = query.Where("1 = 0")
    } else {
        query = query.Where("assignee_key = ?", ak)
    }
}
```

### Round 2: Dead Code Removal

**Item 3** — `item_dto.go`: Delete `WeeklyViewResult`, `WeeklyGroupDTO`, `SubItemWeekDTO`, `SubItemSummaryDTO` structs and all their methods. Keep `MainItemSummaryDTO` (used by `report_service.go`).

**Item 4** — `team_service.go:181`: Delete `_ = team.PmKey` line and its comment.

**Item 5** — `item_pool_handler.go` and `progress_handler.go`: Remove `if userRepo != nil` guards around user lookups (constructors panic on nil, so these checks are dead code).

**Item 6** — `team_service.go:343-345`: Remove the `if result == nil { result = []*dto.UserSearchDTO{} }` block (make on line 335 already returns non-nil).

**Item 7** — `role_repo.go:129-132`: Remove `gorm:"column:..."` tags from `teamPermRow` struct fields (GORM naming strategy handles the mapping).

**Item 8** — `frontend/src/api/client.ts`: Replace the TODO console.error block with toast notifications. `client.ts` is a non-React module (it runs inside an Axios interceptor), so `useToast` (a React hook) cannot be used here. Instead, create a new `frontend/src/lib/toast.ts` shim that exports a standalone `showToast(message: string, variant: 'success' | 'error')` function. The shim stores a module-level reference to the toast context, set by `ToastProvider` on mount. The Axios interceptor calls `showToast(message, 'error')` directly. This avoids importing React hooks into a non-React module while keeping the toast state inside the existing React component tree.

**Item 9** — `frontend/src/types/index.test.ts`: Change `id: 1` to `bizKey: "1"` in the Role test data to match the updated `Role` interface.

### Round 3: Pattern Unification

**Item 10** — Create `pkg/repo/transactor.go` with `DBTransactor` interface. Update `team_service.go` and `item_pool_service.go` to import from `pkg/repo`. Delete local interface definitions.

**Item 11** — `admin_handler.go:159-175`: Replace `parsePagination` with `dto.ApplyPaginationDefaults`. `view_handler.go:97-107`: Replace manual offset calculation with `dto.ApplyPaginationDefaults`.

**Item 12** — Create `pkg/handler/bizkey.go` with `ParseBizKeyParam` and `ResolveBizKey`. Update 7 call sites across handlers:
- `main_item_handler.go`: replace `parseBizKey`
- `sub_item_handler.go`: replace `parseSubBizKey` and `resolveSubID`
- `admin_handler.go`: replace `parseUserBizKey`
- `role_handler.go`: replace `parseRoleBizKey`
- `team_handler.go`: replace `resolveUserBizKey`
- `progress_handler.go`: replace `resolveSubID`
- `item_pool_handler.go`: replace `resolvePoolID`

**Item 13** — Create `vo/team_vo.go` with `TeamVO` struct and `NewTeamVO` constructor. Update `team_handler.go:teamToDTO` to return `NewTeamVO(team)` instead of `gin.H`. Fix `createdAt`/`updatedAt` formatting from raw `time.Time` to RFC3339.

**Item 14** — Create `vo/user_vo.go` with `UserVO` and `NewUserVO`. Update `auth_handler.go:userToDTO` to call `vo.NewUserVO`. Update `admin_service.go:modelToAdminUserDTO` to use `vo.NewUserVO` as base, then add team-specific fields.

**Item 15** — Create `service/status_history_helper.go` with `RecordStatusChange`. Update 5 call sites in `main_item_service.go` (3), `sub_item_service.go` (1), `progress_service.go` (1).

**Item 16** — Remove 22 `String()` calls where the value is already `string`:
- `String(xxx.bizKey)` where `bizKey: string` — just use `xxx.bizKey`
- `String(params.xxxId)` in MSW handlers — route params are always `string`
- `String(x.assigneeKey)` after truthiness guard — already narrowed to `string`
- Keep 13 necessary conversions: `String(inviteRoleId!)`, `String(roleId)` etc. where source is `number`

**Item 17** — `types/index.ts`: Change `Record<number, string[]>` to `Record<string, string[]>`. Update `usePermission.ts` and `PermissionGuard.tsx` signatures to accept `string` teamId. Update all call sites to pass team bizKey string instead of number.

**Item 18** — Rename `assigneeId` to `assigneeKey` in all form state types and usage across dialog components:
- `EditMainItemFormState.assigneeId` → `assigneeKey`
- `CreateMainItemFormState.assigneeId` → `assigneeKey`
- `CreateSubItemFormState.assigneeId` → `assigneeKey`
- `EditSubItemFormState.assigneeId` → `assigneeKey` (item-view version only)
- `toMainForm.assigneeId` / `toSubForm.assigneeId` in ItemPoolPage
- Remove corresponding `assigneeId:` / `assigneeKey:` mapping in API calls

### Round 4: Architecture Alignment

**Item 19** — `model/role.go`: Change `TableName()` returns from `"roles"` to `"pmw_roles"`, from `"role_permissions"` to `"pmw_role_permissions"`. Update both schema files (`SQLite-schema.sql`, `MySql-schema.sql`) with `RENAME TABLE`. Update `role_test.go` assertions and raw SQL DDL.

**Item 20** — `view_service.go`: Remove `NewViewServiceWithUserRepo`. Merge into `NewViewService` by adding `userRepo` as an optional parameter with nil meaning "no user lookups":

```go
func NewViewService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo, userRepo ...repository.UserRepo) ViewService {
    var ur repository.UserRepo
    if len(userRepo) > 0 {
        ur = userRepo[0]
    }
    return &viewService{
        mainItemRepo: mainItemRepo,
        subItemRepo:  subItemRepo,
        progressRepo: progressRepo,
        userRepo:     ur,
    }
}
```

Update `router.go`: replace both `NewViewService(...)` and `NewViewServiceWithUserRepo(...)` calls with `NewViewService(mainItemRepo, subItemRepo, progressRepo)` and `NewViewService(mainItemRepo, subItemRepo, progressRepo, userRepo)` respectively.

**Item 21** — `item_pool_handler.go`: Replace `itemPoolToVO` + `itemPoolsToVOs` with a single `buildItemPoolVOs` function that handles both single and batch by accepting a slice:

```go
func buildItemPoolVOs(items []model.ItemPool, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, c *gin.Context) []vo.ItemPoolVO {
    if len(items) == 0 {
        return []vo.ItemPoolVO{}
    }
    // Batch-resolve submitter names and main-item codes (from itemPoolsToVOs logic)
    // Single-item callers pass a 1-element slice; the batch path has no N+1 overhead.
}
```

Same pattern for `progress_handler.go`: merge `progressRecordToVO` + `progressRecordsToVOs` into:

```go
func buildProgressRecordVOs(records []model.ProgressRecord, userRepo repository.UserRepo, c *gin.Context) []vo.ProgressRecordVO {
    if len(records) == 0 {
        return []vo.ProgressRecordVO{}
    }
    // Batch-resolve author names (from progressRecordsToVOs logic)
}
```

Single-item callers wrap the item in a 1-element slice and index `[0]` of the result. Delete `itemPoolToVO`, `itemPoolsToVOs`, `progressRecordToVO`, and `progressRecordsToVOs`.

**Item 22** — `scopes.go`: Keep `NotDeleted` scope (was going to be deleted, now will be used). Replace inline `WHERE deleted_flag = 0` in `team_repo.go` (lines 37, 222, 225) and `role_repo.go` (lines 32, 44) with `.Scopes(NotDeleted)`.

**Item 23** — `types/index.ts`: Change `TableRow.mainItemId` from `number | null` to `string | null`. Update `TableViewPage.tsx` usage and test data.

**Item 24** — `lib/format.ts`: Add `formatDateOnly(dateStr?: string): string` function (truncate ISO to date, delegate to existing `formatDate`). Update `TeamDetailPage.tsx`, `TeamManagementPage.tsx`, `RoleManagementPage.tsx` to import from `@/lib/format` instead of defining local wrappers. Skip `GanttViewPage.tsx` (different purpose — Date-to-ISO for input elements).

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test |
|-------|-----------|------|--------------|
| Model | Unit | go test | `TestRole_TableName` returns `pmw_roles`, `TestRolePermission_TableName` returns `pmw_role_permissions` |
| Repository | Unit | go test + gorm sqlite | Filter by assignee returns correct subset (Item 2); `NotDeleted` scope applied correctly (Item 22) |
| Service | Unit | go test + mocks | `SubItem.Assign` passes `"assignee_key"` not `"assignee_id"` (Item 1); status history helper records correctly (Item 15) |
| Handler | Unit | go test + httptest | `ParseBizKeyParam` validates and parses; `ResolveBizKey` handles not-found |
| VO | Unit | go test | `NewTeamVO` formats dates as RFC3339; `NewUserVO` maps all fields |
| Frontend | Unit | vitest | Types compile; `formatDateOnly` works; `PermissionData` keyed by string; no `assigneeId` remaining |
| Frontend | Integration | MSW | Filter dropdown returns filtered results |

### Key Test Scenarios

1. **Item 1**: `TestSubItemService_Assign` asserts `fields["assignee_key"]` exists, `fields["assignee_id"]` does not
2. **Item 2**: `TestApplyItemFilter_AssigneeKey` passes string bizKey, asserts correct SQL WHERE with int64 value; test with invalid string asserts `WHERE 1 = 0` is applied (zero results, not all items)
3. **Item 10**: `go build ./...` passes after removing local transaction interfaces
4. **Item 16**: `grep -rn "String(" frontend/src/pages/ | grep -v "node_modules"` shows only necessary conversions
5. **Item 17**: TypeScript compiler (`npx tsc --noEmit`) passes with `Record<string, string[]>`
6. **Item 19**: `go test ./internal/model/ -run TestRole` passes with new table names

### Overall Coverage Target

Current baselines (measured 2026-04-26):
- **Backend:** 80.1% statement coverage
- **Frontend:** 92.5% statement coverage

Floor: coverage must not drop below 78% (backend) or 90% (frontend) after all 24 items land. These floors allow small variance from test refactoring while preventing regressions. Measured via `go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | tail -1` (backend) and `npx vitest run --coverage` (frontend).

## Security Considerations

### Threat Model

The cleanup fixes two pre-existing security-relevant bugs:

**Item 2 — Assignee filter authorization bypass (pre-existing, HIGH impact):**
- **Vulnerability:** `filter_helpers.go` passes a raw `*string` directly into `WHERE assignee_key = ?`. The column is `BIGINT`, so the string-to-int mismatch causes the filter to match zero rows in MySQL (or be silently ignored depending on the driver), returning all items instead of the filtered subset. A user filtering by assignee sees items belonging to other assignees.
- **Impact:** Horizontal privilege escalation within a team. Users with "view filtered items" permission can see items assigned to any team member, not just their own.
- **Mitigation:** The fix parses the string to `int64` before the SQL query. On parse failure, the query applies `WHERE 1 = 0` (returns empty) rather than silently skipping the filter.
- **Residual risk:** During deployment, if the filter receives a non-numeric value from a malformed request, users will see zero results instead of all results — a fail-closed posture.

**Item 1 — Silent data corruption (pre-existing, MEDIUM impact):**
- **Vulnerability:** `sub_item_service.Assign` writes to column `"assignee_id"` which does not exist. The assignment silently fails, leaving `assignee_key` unchanged.
- **Impact:** Item assignment operations appear to succeed but do not persist.
- **Mitigation:** Column name fix to `"assignee_key"`. Existing test assertion `fields["assignee_key"]` validates the fix.

### Recommendations

- **Audit log:** Consider adding an audit log entry when `filter_helpers` receives a non-numeric `assigneeKey` value. This would indicate either a client bug or a probing attempt.
- **Input validation at DTO layer:** The `assigneeKey` filter field in the DTO should be validated as numeric before reaching the repository layer. This defense-in-depth measure is out of scope for this cleanup but should be tracked as a follow-up.

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|--------|------------------|-------------------|
| Story 1: 指派子项恢复 | `sub_item_service.Assign` | Column name fix to `assignee_key` |
| Story 2: 按人筛选恢复 | `filter_helpers.applyItemFilter` | `strconv.ParseInt` conversion |
| Story 2.5: 清理死代码 | Items 3-9 deletions | No new interfaces |
| Story 3: 代码库可维护性 | `DBTransactor`, `ParseBizKeyParam`, `RecordStatusChange`, `NewUserVO` | Items 10, 12, 14, 15 |
| Story 4: 前端类型一致性 | `PermissionData`, `TableRow`, form renames | Items 16, 17, 18, 23 |
| Story 5: 表名规范统一 | `Role.TableName()`, `RolePermission.TableName()` | Item 19 |

## Open Questions

None. All items are well-defined with clear file locations and changes.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Put `DBTransactor` in `dto/` package | DTO already has shared types | `dto/` is for request/response types, not database abstractions | Violates layer responsibility |
| Put `ParseBizKeyParam` in `handler/` package directly | No new sub-package | `handler/` already has domain-specific files; adding utilities mixes concerns | Cleaner in `pkg/handler/` |
| Keep `NotDeleted` scope but also inline in some repos | Less churn | Defeats the purpose of unification | Inconsistent |
| Create `frontend/src/lib/date.ts` for `formatDateOnly` | New dedicated module | Only one function — overkill | Extend existing `format.ts` |
