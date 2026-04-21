# Design: Status Flow Optimization

## Overview

Follow the existing layered architecture (Router → Handler → Service → Repository → Model). Add a shared `pkg/status` package for enumeration and transition logic, a `StatusHistory` model/repo/service for audit logging, and modify `MainItemService`/`SubItemService` for the state machine and linkage.

```
pkg/status/           NEW — status enum + transition maps (pure data, no DB dependency)
model/status_history  NEW — GORM model
repository/           NEW — StatusHistoryRepo interface
repository/gorm/      NEW — GORM impl
service/              NEW — StatusHistoryService
service/              MOD — MainItemService (ChangeStatus, EvaluateLinkage, AvailableTransitions)
service/              MOD — SubItemService (updated ChangeStatus, trigger linkage on create/delete)
handler/              MOD — new endpoints + updated handlers
dto/                  MOD — remove Status from MainItemUpdateReq
frontend/src/lib/     NEW — status.ts (code-name mapping)
frontend/src/api/     MOD — new API calls
frontend/src/components/shared/ MOD — StatusBadge, StatusDropdown
```

## New Package: `pkg/status`

Pure data package with no dependencies on model/service layers. Both services and handlers import from here.

**`backend/internal/pkg/status/status.go`** — Status definition and lookup:

```go
package status

type StatusDef struct {
    Code     string // "pending", "progressing", ...
    Name     string // "待开始", "进行中", ...
    Terminal bool   // completed, closed
}

var (
    MainItemStatuses = map[string]StatusDef{...} // 7 entries
    SubItemStatuses  = map[string]StatusDef{...} // 6 entries

    MainTerminalStatuses = []string{"completed", "closed"}
    SubTerminalStatuses  = []string{"completed", "closed"}
)

func GetMainItemStatus(code string) (StatusDef, bool)
func GetSubItemStatus(code string) (StatusDef, bool)
func IsMainTerminal(code string) bool
func IsSubTerminal(code string) bool
```

**`backend/internal/pkg/status/transition.go`** — Transition rules:

```go
var (
    MainItemTransitions = map[string][]string{
        "pending":     {"progressing", "closed"},
        "progressing": {"blocking", "pausing", "reviewing", "closed"},
        "blocking":    {"progressing"},
        "pausing":     {"progressing", "closed"},
        "reviewing":   {"completed", "progressing"}, // PM-only enforced in handler
    }
    SubItemTransitions = map[string][]string{
        "pending":     {"progressing", "closed"},
        "progressing": {"blocking", "pausing", "completed", "closed"},
        "blocking":    {"progressing"},
        "pausing":     {"progressing", "closed"},
    }
)

func IsValidTransition(transitions map[string][]string, from, to string) bool
func GetAvailableTransitions(transitions map[string][]string, current string) []string
```

**Why a separate package**: Status codes are used by models (default values), services (transition validation), handlers (available-transitions API), and frontend (via API). A shared package avoids circular dependencies and centralizes the single source of truth.

## New Model: `StatusHistory`

**`backend/internal/model/status_history.go`**:

```go
type StatusHistory struct {
    ID          uint       `gorm:"primarykey" json:"id"`
    ItemType    string     `gorm:"type:varchar(20);not null;index:idx_item" json:"itemType"`
    ItemID      uint       `gorm:"not null;index:idx_item" json:"itemId"`
    FromStatus  string     `gorm:"type:varchar(20);not null" json:"fromStatus"`
    ToStatus    string     `gorm:"type:varchar(20);not null" json:"toStatus"`
    ChangedBy   uint       `gorm:"not null" json:"changedBy"`
    IsAuto      bool       `gorm:"not null;default:false" json:"isAuto"`
    Remark      string     `gorm:"type:varchar(500)" json:"remark"`
    CompletedAt *time.Time `json:"completedAt"`
    CreatedAt   time.Time  `json:"createdAt"`
}
```

Does NOT embed `BaseModel` — no `UpdatedAt` or soft delete needed (append-only log).

## New Repository: `StatusHistoryRepo`

**`backend/internal/repository/status_history_repo.go`** (interface):

```go
type StatusHistoryRepo interface {
    Create(ctx context.Context, record *model.StatusHistory) error
    ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error)
}
```

**`backend/internal/repository/gorm/status_history_repo.go`** (GORM impl) — follows existing constructor pattern `NewGormStatusHistoryRepo(db) StatusHistoryRepo`.

## New Service: `StatusHistoryService`

**`backend/internal/service/status_history_service.go`**:

```go
type StatusHistoryService interface {
    Record(ctx context.Context, record *model.StatusHistory) error
    ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error)
}
```

Simple wrapper — no business logic. Constructor takes `StatusHistoryRepo`, validates non-nil with panic pattern.

## Modified Service: `MainItemService`

### New Method: `ChangeStatus`

```go
ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error)
```

Flow:
1. Fetch item via repo, verify team ownership
2. Self-transition check: `newStatus == item.Status` → return `ErrInvalidStatus`
3. Validate transition via `status.IsValidTransition(status.MainItemTransitions, item.Status, newStatus)`
4. **PM-only check**: if `from == "reviewing"` → verify `callerID == item.PmID` (or caller is SuperAdmin)
5. Build update fields map: `{"status": newStatus}`
6. Terminal side effects: if `status.IsMainTerminal(newStatus)` → add `{"completion": 100, "actual_end_date": time.Now()}`
7. Update via repo
8. Record to StatusHistory (non-auto)
9. Return updated item

**PM authorization**: Check `callerID == item.PmID` directly on the model. SuperAdmin bypass handled by middleware. No new permission code needed for the PM-only check — the existing `main_item:change_status` permission gates the endpoint, and the service enforces the per-item PM constraint for reviewing transitions.

### New Method: `EvaluateLinkage`

```go
EvaluateLinkage(ctx context.Context, mainItemID uint, changedBy uint) (*LinkageResult, error)
```

Called by SubItem service after status changes, creates, and deletes. Runs under the per-MainItem mutex. Returns `LinkageResult` so callers (SubItemService) can propagate linkage warnings to the handler.

Flow:
1. Acquire mutex for `mainItemID`
2. Fetch MainItem and all its SubItems
3. If no SubItems → return nil (no linkage)
4. Evaluate 5-level priority rules (see PRD R3)
5. If target status matches current → return nil
6. If transition not valid → record to StatusHistory (`is_auto=true`, remark="linkage failed: X→Y not allowed"), return nil
7. Build update fields, apply terminal side effects if needed
8. Update MainItem via repo
9. Record to StatusHistory (`is_auto=true`)

### New Method: `AvailableTransitions`

```go
AvailableTransitions(ctx context.Context, teamID, callerID, itemID uint) ([]string, error)
```

1. Fetch item, verify team ownership
2. Get base transitions from `status.GetAvailableTransitions(status.MainItemTransitions, item.Status)`
3. **PM-only filter**: if `callerID != item.PmID` → remove `completed` and `progressing` from results when `item.Status == "reviewing"`
4. Return filtered list

### Modified Method: `Update`

Remove the `if req.Status != nil` block (lines 88-89). Status changes must go through `ChangeStatus`.

### Modified Method: `Archive`

Change `"已完成"` / `"已关闭"` to `"completed"` / `"closed"`.

### Modified Method: `Create`

Change default `Status: "待开始"` to `Status: "pending"`.

### New Permission Code

Add `main_item:change_status` to `permissions/codes.go` Registry under the `main_item` resource block.

## LinkageResult Struct

Defined in `service/` — shared by MainItemService.EvaluateLinkage return value and SubItemService.ChangeStatus:

```go
type LinkageResult struct {
    Triggered    bool   // whether linkage was attempted (had sub-items)
    Success      bool   // whether the transition succeeded
    TargetStatus string // the intended target status
    Remark       string // failure reason if not success
}

func (r *LinkageResult) Warning() string {
    if r.Triggered && !r.Success {
        return fmt.Sprintf("主事项状态联动失败：%s", r.Remark)
    }
    return ""
}
```

## Modified Service: `SubItemService`

### Updated `ChangeStatus`

**Updated signature** (changed from returning `error` to returning `ChangeStatusResult`):

```go
type SubItemChangeResult struct {
    SubItem       *model.SubItem
    LinkageResult *LinkageResult // nil if no linkage was triggered
}

ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*SubItemChangeResult, error)
```

Flow:
1. Replace `allowedTransitions` map with `status.SubItemTransitions` from `pkg/status`
2. Replace `isValidTransition` call with `status.IsValidTransition`
3. **Remove** `newStatus == "已延期"` block (DelayCount increment, priority upgrade)
4. **Update** `newStatus == "completed"` block: use code `"completed"`, set `completion=100` + `actual_end_date` in same update fields map
5. After successful status update:
   a. If `newStatus == "completed"` → call `RecalcCompletion` (existing)
   b. Call `mainItemSvc.EvaluateLinkage(ctx, item.MainItemID, callerID)` (new) — capture LinkageResult
6. Record to StatusHistory (non-auto)
7. Return `SubItemChangeResult` with updated SubItem + LinkageResult

### Updated `Create`

Change default `Status: "待开始"` to `Status: "pending"`. After creating, call `mainItemSvc.EvaluateLinkage`.

### Updated `Delete`

After deleting a SubItem, call `mainItemSvc.EvaluateLinkage`.

### New Method: `AvailableTransitions`

```go
AvailableTransitions(ctx context.Context, teamID, subID uint) ([]string, error)
```

Fetches item, returns `status.GetAvailableTransitions(status.SubItemTransitions, item.Status)`. No PM filtering needed for SubItem.

## Concurrency: Per-MainItem Mutex

**`backend/internal/service/main_item_service.go`**:

```go
import "sync"

var (
    linkageMuMap = make(map[uint]*sync.Mutex)
    linkageMapMu sync.Mutex // protects linkageMuMap
)

func getLinkageMutex(mainItemID uint) *sync.Mutex {
    linkageMapMu.Lock()
    defer linkageMapMu.Unlock()
    if mu, ok := linkageMuMap[mainItemID]; ok {
        return mu
    }
    mu := &sync.Mutex{}
    linkageMuMap[mainItemID] = mu
    return mu
}
```

`EvaluateLinkage` acquires `getLinkageMutex(mainItemID)` at the start, releases on return. This serializes all linkage evaluations for the same MainItem, preventing race conditions when multiple SubItems change simultaneously.

No cleanup of mutex map needed — the number of MainItems is bounded by usage, and each mutex is a small allocation.

## Security Considerations

The security surface for this feature is the PM-only constraint on reviewing transitions (PRD AC-20). Three layers enforce this:

| Layer | Mechanism | Scope |
|-------|-----------|-------|
| **Endpoint** | Permission code `main_item:change_status` gates `PUT /main-items/:itemId/status` via middleware | Any authenticated user with this permission can call the endpoint |
| **Service** | `ChangeStatus` checks `callerID == item.PmID` when `from == "reviewing"` | Only the PM assigned to this specific MainItem (or SuperAdmin) can transition out of reviewing |
| **API** | `AvailableTransitions` removes `completed`/`progressing` from results when `callerID != item.PmID` and `item.Status == "reviewing"` | Non-PM users never see reviewing exit options in the dropdown |
| **Frontend** | PM-only UI visibility for reviewing transitions | Defense in depth — not the sole enforcement point |

SubItem has no PM-only transitions, so the existing `sub_item:change_status` permission plus assignee check (already implemented) is sufficient.

No other security concerns: status is not sensitive data, all endpoints require team membership (existing middleware), and status history is append-only (no delete/update).

## Modified DTO

**`backend/internal/dto/item_dto.go`**:

- Remove `Status *string` from `MainItemUpdateReq` (line 96)
- No change to SubItem DTOs (SubItemUpdateReq already has no Status field)

## Handler & Router Changes

### New Routes

```
PUT  /teams/:teamId/main-items/:itemId/status                    → MainItem.ChangeStatus      (perm: main_item:change_status)
GET  /teams/:teamId/main-items/:itemId/available-transitions     → MainItem.AvailableTransitions (perm: main_item:read)
GET  /teams/:teamId/sub-items/:subId/available-transitions       → SubItem.AvailableTransitions  (perm: sub_item:read)
```

Permission `main_item:change_status` is new. The available-transitions endpoints use existing read permissions since they don't modify data.

### Modified Handlers

**`MainItemHandler`** — add constructor dependency: `StatusHistoryService`. New methods:
- `ChangeStatus(c *gin.Context)` — binds `{status: string}`, calls `svc.ChangeStatus`, returns VO
- `AvailableTransitions(c *gin.Context)` — calls `svc.AvailableTransitions`, returns `{"transitions": [...]}`

**`SubItemHandler`** — add constructor dependency: `StatusHistoryService`. New method:
- `AvailableTransitions(c *gin.Context)` — calls `svc.AvailableTransitions`, returns `{"transitions": [...]}`

### Dependencies Struct Update

**`handler/router.go`**: `Dependencies` struct gains `StatusHistoryService` and `StatusHistoryRepo` fields. Wiring in `SetupRouter` initializes these.

## Model Changes

**`model/main_item.go`**:
- Remove `DelayCount` field (line 21)
- Change default status from `'待开始'` to `'pending'`

**`model/sub_item.go`**:
- Remove `DelayCount` field (line 23)
- Change default status from `'待开始'` to `'pending'`
- Remove the 8-status comment block (lines 7-8)

**`model/status_history.go`** — new file, auto-migrated via `migration/runner.go`.

## VO Changes

**`vo/item_vo.go`** — `MainItemVO.Status` remains `string`. The API returns the English code. Frontend maps code→name for display.

Alternatively, return both code and name:
```go
type MainItemVO struct {
    // ...
    Status     string `json:"status"`     // code: "progressing"
    StatusName string `json:"statusName"` // name: "进行中"
}
```

**Decision**: Return both. The VO constructor looks up `status.GetMainItemStatus(m.Status)` to populate `StatusName`. This lets the frontend render without maintaining a parallel code→name map for basic display, while still using the code for API calls.

Same for `SubItemSummaryVO`.

## Frontend Changes

### New File: `frontend/src/lib/status.ts`

```ts
export const MAIN_ITEM_STATUSES = {
  pending:     { name: '待开始', variant: 'planning',    terminal: false },
  progressing: { name: '进行中', variant: 'in-progress', terminal: false },
  blocking:    { name: '阻塞中', variant: 'overdue',     terminal: false },
  pausing:     { name: '已暂停', variant: 'on-hold',     terminal: false },
  reviewing:   { name: '待验收', variant: 'pending',     terminal: false },
  completed:   { name: '已完成', variant: 'completed',   terminal: true },
  closed:      { name: '已关闭', variant: 'cancelled',   terminal: true },
} as const

export const SUB_ITEM_STATUSES = {
  pending:     { name: '待开始', variant: 'planning',    terminal: false },
  progressing: { name: '进行中', variant: 'in-progress', terminal: false },
  blocking:    { name: '阻塞中', variant: 'overdue',     terminal: false },
  pausing:     { name: '已暂停', variant: 'on-hold',     terminal: false },
  completed:   { name: '已完成', variant: 'completed',   terminal: true },
  closed:      { name: '已关闭', variant: 'cancelled',   terminal: true },
} as const

export function isOverdue(expectedEndDate?: string, status?: string): boolean {
  if (!expectedEndDate || !status) return false
  const isTerminal = MAIN_ITEM_STATUSES[status]?.terminal || SUB_ITEM_STATUSES[status]?.terminal
  if (isTerminal) return false
  return new Date(expectedEndDate) < new Date()
}
```

### Modified: `StatusBadge`

Replace hardcoded Chinese `statusVariantMap` with lookup from `lib/status.ts`:

```tsx
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from '@/lib/status'

// Variant comes from the status definition, not a Chinese string key
const variant = (MAIN_ITEM_STATUSES[status] || SUB_ITEM_STATUSES[status])?.variant ?? 'default'
```

Display text uses `statusName` from API response (or falls back to the map).

### Modified: `StatusDropdown` (in ItemViewPage, MainItemDetailPage)

- Fetch available transitions via API: `GET /teams/:teamId/main-items/:itemId/available-transitions`
- Render only returned transitions as dropdown items
- `onSelect` calls `changeMainItemStatusApi` or `changeSubItemStatusApi`
- For terminal transitions: show confirmation dialog before calling API
- PM-only transitions are already filtered server-side

### New API Calls

**`frontend/src/api/mainItems.ts`**:

```ts
export async function changeMainItemStatusApi(teamId: number, itemId: number, req: ChangeStatusReq) {
  return client.put(`/teams/${teamId}/main-items/${itemId}/status`, req)
}

export async function getMainItemTransitionsApi(teamId: number, itemId: number) {
  return client.get(`/teams/${teamId}/main-items/${itemId}/available-transitions`)
}
```

**`frontend/src/api/subItems.ts`**:

```ts
export async function getSubItemTransitionsApi(teamId: number, subId: number) {
  return client.get(`/teams/${teamId}/sub-items/${subId}/available-transitions`)
}
```

### Modified: Frontend Types

**`frontend/src/types/index.ts`**:
- Remove `status?: string` from `UpdateMainItemReq`
- Add `statusName?: string` to `MainItem` and `SubItem` interfaces (for API response)

### Overdue Badge

Wherever an item is displayed with a status badge, conditionally render an overdue indicator:

```tsx
{isOverdue(item.expectedEndDate, item.status) && (
  <Badge variant="error" size="sm">延期</Badge>
)}
```

Co-locate with the `StatusBadge` — either as a wrapper component or inline in page components.

### Linkage Failure Toast

The `ChangeStatus` API response includes linkage result info:

```json
{"code": 0, "data": {"status": "completed", "linkageWarning": "主事项状态联动失败：reviewing→progressing 不允许"}}
```

SubItem's `ChangeStatus` returns `SubItemChangeResult` containing `LinkageResult` (see LinkageResult Struct section). The handler checks `result.LinkageResult.Warning()` and includes it in the API response. The frontend checks for this field and shows a toast with `addToast(linkageWarning, 'warning')`.

## Execution Flow: SubItem Status Change

```
1. Handler receives PUT /sub-items/:subId/status {status: "completed"}
2. Handler calls SubItemService.ChangeStatus(ctx, teamID, callerID, subID, "completed")
3. Service validates transition (SubItemTransitions)
4. Service builds update fields: {status: "completed", completion: 100, actual_end_date: now}
5. Service calls repo.Update
6. Service records to StatusHistory (isAuto=false)
7. Service calls MainItemService.RecalcCompletion (updates MainItem.completion)
8. Service calls MainItemService.EvaluateLinkage (under mutex)
   a. EvaluateLinkage fetches all SubItems
   b. Evaluates priority rules
   c. If all completed/closed → target = reviewing
   d. If transition valid → updates MainItem, records to StatusHistory (isAuto=true)
   e. If transition invalid → records intent to StatusHistory, returns LinkageResult with warning
9. Service returns updated SubItem + LinkageResult
10. Handler returns VO + linkageWarning if present
11. Frontend shows toast if linkageWarning exists
```

## Testing Strategy

Coverage target: **90% for new code**. Follow project conventions in `.claude/rules/testing.md`: table-driven tests, mock repos/services via interfaces, co-located test files.

### `pkg/status` — Unit Tests

Table-driven tests for pure functions:

| Test | Pattern | Cases |
|------|---------|-------|
| `TestIsValidTransition` | table-driven | All 10 MainItem + 9 SubItem legal transitions pass; illegal ones return false; self-transitions return false; unknown status codes return false |
| `TestGetAvailableTransitions` | table-driven | Each MainItem/SubItem status returns correct target list; terminal states return empty; unknown status returns empty |
| `TestGetMainItemStatus` / `TestGetSubItemStatus` | table-driven | All 7/6 codes resolve; unknown code returns false |
| `TestIsMainTerminal` / `TestIsSubTerminal` | table-driven | completed/closed return true; others false |

### Service Layer — Unit Tests

Mock repos via interfaces. Mock `StatusHistoryService` for callers.

**`main_item_service_test.go`** — new tests:

| Test | Description |
|------|-------------|
| `TestChangeStatus_ValidTransitions` | Table-driven: 10 legal paths, each succeeds |
| `TestChangeStatus_InvalidTransitions` | Table-driven: illegal paths return `ErrInvalidStatus` |
| `TestChangeStatus_SelfTransition` | Same status returns error |
| `TestChangeStatus_TerminalSideEffects` | completed/closed sets completion=100 + actual_end_date |
| `TestChangeStatus_PMOnlyReviewing` | Non-PM caller on reviewing→completed returns 403 |
| `TestEvaluateLinkage_*` | Table-driven: 5 priority levels (AC-8), AC-9 (reviewing+new subitem), AC-10 (delete triggers), AC-11 (no subitems=no-op), AC-12 (invalid target logs intent) |
| `TestAvailableTransitions_*` | PM sees reviewing exits; non-PM does not; non-reviewing status shows full list |

**`sub_item_service_test.go`** — update existing tests:

| Change | Description |
|--------|-------------|
| Replace all Chinese status strings | `"待开始"` → `"pending"`, etc. |
| Remove | `TestChangeStatus_已延期_*` (3 tests), `TestChangeStatus_待验收_*` (3 tests), `TestChangeStatus_挂起_*` (2 tests) |
| Add | `TestChangeStatus_TerminalSideEffects`, `TestChangeStatus_TriggersLinkage`, `TestChangeStatus_ReturnsLinkageResult` |
| Update | `TestCreate_TriggersLinkage`, `TestDelete_TriggersLinkage` |

**`status_history_service_test.go`** — new tests:

| Test | Description |
|------|-------------|
| `TestRecord` | Creates record, verifies fields |
| `TestListByItem` | Pagination, filtering by itemType+itemID |

### Handler Layer — Unit Tests

Use `httptest` + Gin test context. Mock services.

| Test File | New Tests |
|-----------|-----------|
| `main_item_handler_test.go` | `TestChangeStatus_Success`, `TestChangeStatus_InvalidTransition` (422), `TestChangeStatus_PMOnly` (403), `TestAvailableTransitions` |
| `sub_item_handler_test.go` | `TestAvailableTransitions`, updated `TestChangeStatus_*` for new return format |

### Frontend — Component Tests

`@testing-library/react` + `vitest`. Mock API calls.

| Test File | Tests |
|-----------|-------|
| `StatusBadge.test.tsx` | Renders Chinese name from code; applies correct variant class; unknown code falls back to default |
| `status.test.ts` | `isOverdue` returns true for past date + non-terminal; false for terminal; false for no date |

### Integration Test (End-to-End Path)

One integration test covering the critical path (AC-13: execution order):

```
SubItem ChangeStatus("completed")
  → RecalcCompletion updates MainItem.completion
  → EvaluateLinkage updates MainItem.status to "reviewing"
  → status_histories records both (manual + auto)
```

Verify: MainItem.completion=100 AND MainItem.status="reviewing" after SubItem completion.

## File Change Summary

### New Files

| File | Purpose |
|------|---------|
| `backend/internal/pkg/status/status.go` | Status enum definitions |
| `backend/internal/pkg/status/transition.go` | Transition maps and helpers |
| `backend/internal/model/status_history.go` | StatusHistory GORM model |
| `backend/internal/repository/status_history_repo.go` | Repo interface |
| `backend/internal/repository/gorm/status_history_repo.go` | GORM implementation |
| `backend/internal/service/status_history_service.go` | Service interface + impl |
| `frontend/src/lib/status.ts` | Status code-name-variant mapping |

### Modified Files

| File | Changes |
|------|---------|
| `model/main_item.go` | Remove DelayCount, change default to "pending" |
| `model/sub_item.go` | Remove DelayCount, change default to "pending", remove status comment |
| `dto/item_dto.go` | Remove Status from MainItemUpdateReq |
| `service/main_item_service.go` | Add ChangeStatus, EvaluateLinkage, AvailableTransitions; modify Update, Archive, Create |
| `service/sub_item_service.go` | Replace transitions, remove delay logic, add linkage triggers, add AvailableTransitions |
| `handler/main_item_handler.go` | Add ChangeStatus, AvailableTransitions handlers; add StatusHistoryService dep |
| `handler/sub_item_handler.go` | Add AvailableTransitions handler; add StatusHistoryService dep |
| `handler/router.go` | Add 3 new routes, update Dependencies struct |
| `pkg/permissions/codes.go` | Add `main_item:change_status` |
| `migration/runner.go` | Add StatusHistory to auto-migration |
| `vo/item_vo.go` | Add StatusName field to MainItemVO, SubItemSummaryVO |
| `frontend/src/api/mainItems.ts` | Add changeMainItemStatusApi, getMainItemTransitionsApi |
| `frontend/src/api/subItems.ts` | Add getSubItemTransitionsApi |
| `frontend/src/types/index.ts` | Remove status from UpdateMainItemReq, add statusName |
| `frontend/src/components/shared/StatusBadge.tsx` | Use code-based lookup |
| `frontend/src/pages/ItemViewPage.tsx` | Update STATUS_OPTIONS, bind StatusDropdown, add overdue badge |
| `frontend/src/pages/MainItemDetailPage.tsx` | Update STATUS_OPTIONS, bind StatusDropdown |
| `frontend/src/pages/TableViewPage.tsx` | Update STATUS_OPTIONS |
| `frontend/src/mocks/handlers.ts` | Update status values to codes |

### Test Files to Update

All test files in backend that reference Chinese status strings must be updated to use English codes. The `pkg/status` package needs its own tests. New handler tests for the 3 new endpoints. Service tests for EvaluateLinkage covering all 5 priority levels.
