---
created: 2026-04-22
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Code Quality Cleanup

## Overview

A four-phase progressive cleanup of the PM Work Tracker codebase. Each phase produces a shippable PR with no behavioral changes to end users.

**Technical approach**: Delete dead code and fix contracts first (lowest risk), then fix performance (highest impact), then deduplicate (maintainability), then split files (readability). The hybrid SQL strategy uses SQL-level filtering for single-entity queries (AdminListUsers, SearchAvailableUsers, GanttView) while keeping Go-level filtering for cross-entity queries (TableView) with batch name resolution to eliminate N+1.

## Architecture

### Layer Placement

No new layers. Changes touch existing layers within their defined responsibilities:

| Phase | Backend Layers | Frontend Layers |
|-------|---------------|-----------------|
| 1 | Model, VO, Handler, Service, Middleware | Types, API, Pages, Components, Store |
| 2 | Repository, Service, Handler | Pages, Hooks, Store |
| 3 | Repository/gorm, Service, pkg/ | Components/shared, Hooks, Lib |
| 4 | Service, Handler, DTO | Pages (file splits only) |

### Component Diagram

```
Phase 1: Delete dead code + fix contracts
┌─────────────────────────────────────────────────────────┐
│ Backend                                                 │
│  Delete: middleware/rbac.go, vo/team_vo.go              │
│  Delete: parseItemIDAsUint, paginateUsers, derefStr,    │
│          Filename, GetCallerTeamRole, GetUsername,       │
│          WeeklyView, codeMap                             │
│  Fix: SubItemVO +statusName, remove contract mismatches │
├─────────────────────────────────────────────────────────┤
│ Frontend                                                │
│  Delete: AdminRoute.tsx, AdminPage.tsx, ItemFilters.tsx │
│  Delete: WeeklyViewResp, WeeklyGroup, SubItemWithProgress│
│  Delete: archiveMainItemApi, assignSubItemApi,          │
│          getItemPoolApi, correctCompletionApi            │
│  Delete: ApiSuccessEnvelope, ApiErrorEnvelope (client.ts)│
│  Delete: TeamMember type (types/index.ts)                │
│  Fix: Remove delayCount, isKeyItem, align               │
│       TeamMemberResp/AdminTeam                           │
│  Fix: setAuth redundant expression (auth.ts)            │
└─────────────────────────────────────────────────────────┘

Phase 2: N+1 + performance
┌─────────────────────────────────────────────────────────┐
│ Backend                                                 │
│  New repo methods: FindByIDs, ListFiltered,             │
│                    ListByTeamAndStatus                   │
│  Batch resolution: resolveAssigneeNames → map lookup    │
│  SQL pushdown: AdminListUsers, SearchAvailableUsers,    │
│                GanttView                                │
│  Hybrid: TableView keeps Go filter + batch names        │
├─────────────────────────────────────────────────────────┤
│ Frontend                                                │
│  Fix: useMemo→useEffect, user:any→User                  │
│  Fix: imperative fetch → React Query                    │
└─────────────────────────────────────────────────────────┘

Phase 3: Dedup + extraction
┌─────────────────────────────────────────────────────────┐
│ Backend                                                 │
│  New: pkg/repo (FindByID[T], FindByIDs[T],              │
│                      UpdateFields[T])                   │
│  New: shared applyItemFilter in repo layer              │
│  Extract: formatTimePtr → pkg/dates.FormatTimePtr       │
│  Merge: DBTransactor interfaces                         │
├─────────────────────────────────────────────────────────┤
│ Frontend                                                │
│  New: components/shared/StatusTransitionDropdown.tsx    │
│  New: hooks/useMemberName.ts                            │
│  New: lib/format.ts (unified formatDate)                │
│  New: components/shared/MemberSelect.tsx                │
│  New: lib/status.ts (terminal status constants)         │
└─────────────────────────────────────────────────────────┘

Phase 4: File splits + readability
┌─────────────────────────────────────────────────────────┐
│ Backend                                                 │
│  Decompose: WeeklyComparison into smaller functions     │
│  Standardize: all handlers → panic-on-nil constructors │
│  Move: inline request structs → dto package             │
├─────────────────────────────────────────────────────────┤
│ Frontend                                                │
│  Split: ItemViewPage (1459→~8 files)                    │
│  Split: MainItemDetailPage (927→~5 files)               │
│  Target: no page file > 300 lines                       │
└─────────────────────────────────────────────────────────┘
```

### Dependencies

**Existing dependencies used** (no new dependencies added):
- `github.com/glebarez/sqlite` — already used for storage
- `@tanstack/react-query` ^5.40.0 — already used across pages
- `zustand` — already used for stores
- `vitest` + `@testing-library/react` — already used for tests

**New internal packages**:
- `backend/internal/pkg/repo` — generic repository helpers
- `frontend/src/lib/format.ts` — unified date formatting
- `frontend/src/lib/status.ts` — shared status constants

## Interfaces

### Phase 2: New Repository Methods

```go
// repository/user_repo.go — add batch method
type UserRepo interface {
    // ... existing methods ...
    FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.User, error)
}

// repository/main_item_repo.go — add filtered list + batch
type MainItemRepo interface {
    // ... existing methods ...
    FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error)
    ListByTeamAndStatus(ctx context.Context, teamID uint, status string) ([]model.MainItem, error)
}

// repository/user_repo.go — SQL-level filtering
type UserRepo interface {
    // ... existing methods ...
    ListFiltered(ctx context.Context, search string, offset, limit int) ([]model.User, int64, error)
    SearchAvailable(ctx context.Context, teamID uint, search string, limit int) ([]model.User, error)
}
```

### Phase 3: Generic Repository Helpers

```go
// pkg/repo/helpers.go
func FindByID[T any](db *gorm.DB, ctx context.Context, id uint) (*T, error)
func FindByIDs[T any](db *gorm.DB, ctx context.Context, ids []uint) (map[uint]*T, error)
// UpdateFields updates only the named fields on an existing model instance.
// Field keys must match GORM column names (snake_case). Accepted keys per entity:
//
//   MainItem: "title", "priority", "status", "assignee_id"
//   SubItem:  "title", "status", "assignee_id"
//   User:     "display_name"
//
// Any key not in the entity's accepted set causes the method to return
// apperrors.ErrInvalidField before executing the update.
func UpdateFields[T any](db *gorm.DB, ctx context.Context, item *T, teamID uint, fields map[string]any) error
```

### Phase 3: Frontend Shared Components

```tsx
// components/shared/StatusTransitionDropdown.tsx
interface StatusTransitionDropdownProps {
  currentStatus: string
  itemType: 'main' | 'sub'
  teamId: number
  itemId: number
  onStatusChanged: (newStatus: string) => void
  /** Optional callback before terminal status transition (e.g., achievement dialog for sub-items) */
  onBeforeTerminalStatus?: (newStatus: string) => Promise<boolean>
  disabled?: boolean
}

// hooks/useMemberName.ts
// Returns a lookup function: pass an assigneeId, get the display name or "Unassigned"
function useMemberName(members: TeamMember[]): (assigneeId: number | null) => string {
  return useCallback((assigneeId: number | null): string => {
    if (assigneeId == null) return 'Unassigned'
    const member = members.find((m) => m.userId === assigneeId)
    return member?.displayName ?? 'Unknown'
  }, [members])
}

// lib/format.ts
function formatDate(date: string | null | undefined): string
```

## Data Models

### Phase 1: Contract Fixes

```go
// vo/item_vo.go — add statusName to SubItemVO (parity with MainItemVO)
type SubItemVO struct {
    ID              uint    `json:"id"`
    TeamID          uint    `json:"teamId"`
    MainItemID      uint    `json:"mainItemId"`
    Title           string  `json:"title"`
    Description     string  `json:"description"`
    Priority        string  `json:"priority"`
    AssigneeID      *uint   `json:"assigneeId"`
    StartDate       *string `json:"startDate"`
    ExpectedEndDate *string `json:"expectedEndDate"`
    ActualEndDate   *string `json:"actualEndDate"`
    Status          string  `json:"status"`
    StatusName      string  `json:"statusName"` // NEW: parity with MainItemVO
    Completion      float64 `json:"completion"`
    IsKeyItem       bool    `json:"isKeyItem"`
    Weight          float64 `json:"weight"`
    CreatedAt       string  `json:"createdAt"`
    UpdatedAt       string  `json:"updatedAt"`
}

// MainItemVO already has StatusName — shown for reference (no changes)
type MainItemVO struct {
    ID              uint    `json:"id"`
    TeamID          uint    `json:"teamId"`
    Code            string  `json:"code"`
    Title           string  `json:"title"`
    Priority        string  `json:"priority"`
    ProposerID      uint    `json:"proposerId"`
    AssigneeID      *uint   `json:"assigneeId"`
    StartDate       *string `json:"startDate"`
    ExpectedEndDate *string `json:"expectedEndDate"`
    ActualEndDate   *string `json:"actualEndDate"`
    Status          string  `json:"status"`
    StatusName      string  `json:"statusName"`
    Completion      float64 `json:"completion"`
    IsKeyItem       bool    `json:"isKeyItem"`
    ArchivedAt      *string `json:"archivedAt"`
    CreatedAt       string  `json:"createdAt"`
    UpdatedAt       string  `json:"updatedAt"`
}
```

```typescript
// types/index.ts — remove fields backend never sends
interface MainItem {
  id: number
  teamId: number
  code: string
  title: string
  description?: string
  priority: string
  proposerId: number
  assigneeId: number | null
  startDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  status: string
  statusName?: string
  completion: number
  isKeyItem: boolean
  // REMOVED: delayCount: number — backend never sends this field
  // REMOVED: archivedAt: string | null — backend never sends this field
  createdAt: string
  updatedAt: string
}

interface SubItem {
  id: number
  teamId: number
  mainItemId: number
  code: string
  title: string
  description: string
  priority: string
  assigneeId: number | null
  startDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  status: string
  statusName?: string
  completion: number
  isKeyItem: boolean
  // REMOVED: delayCount: number — backend never sends this field
  weight: number
  createdAt: string
  updatedAt: string
}

// REMOVED types: WeeklyViewResp, WeeklyGroup, SubItemWithProgress, TeamMember
```

```typescript
// store/auth.ts — fix redundant expression
// Before: isSuperAdmin: (user?.isSuperAdmin ?? user?.isSuperAdmin) === true
// After:
isSuperAdmin: user?.isSuperAdmin === true
```

### Phase 2: New DTOs

```go
// dto/item_dto.go — move inline request structs here
type ChangeStatusReq struct {
    Status string `json:"status" binding:"required"`
}

type AssignSubItemReq struct {
    AssigneeID uint `json:"assigneeId" binding:"required"`
}
```

### Phase 3: Shared Filter Helper

```go
// repository/gorm/filter_helpers.go
// Backend-only: isKeyItem filter operates on the DB column; frontend type
// does not expose this field (removed in Phase 1).
func applyItemFilter(db *gorm.DB, status, priority string, assigneeID *uint, isKeyItem *bool) *gorm.DB {
    if status != "" {
        db = db.Where("status = ?", status)
    }
    if priority != "" {
        db = db.Where("priority = ?", priority)
    }
    if assigneeID != nil {
        db = db.Where("assignee_id = ?", *assigneeID)
    }
    if isKeyItem != nil {
        db = db.Where("is_key_item = ?", *isKeyItem)
    }
    return db
}
```

## Error Handling

No new error types. Existing error patterns remain:
- `pkg/errors.MapNotFound` for repository not-found mapping
- `apperrors.RespondError` / `RespondOK` for handler responses
- Generic helpers in Phase 3 use the same `MapNotFound` pattern internally

Phase 2 risk mitigation: If a batch query returns partial results (some IDs not found), the service layer treats missing IDs as "unknown" rather than failing — matching current behavior where individual lookups silently handle not-found.

### Batch Query Contract

The `FindByIDs[T]` generic helper in `pkg/repo` and all Phase 2 batch methods (`UserRepo.FindByIDs`, `MainItemRepo.FindByIDs`) follow this contract:

**Return type**: `map[uint]*T`

| Condition | Behavior |
|-----------|----------|
| SQL/database error | Return `(nil, error)` — propagate as-is via `pkg/errors.MapNotFound` |
| Empty input (`len(ids) == 0`) | Return `map[uint]*T{}` (empty map), no query executed |
| Zero results | Return `map[uint]*T{}` (empty map), `nil` error |
| Partial results | Return map with found entries only; absent keys = not found, no error |

Callers detect missing entries with `entry, ok := result[id]`. This avoids sentinel values and lets the caller decide how to represent "unknown" (e.g., `"Unknown"` string for display names).

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tooling | Scope |
|-------|-----------|---------|-------|
| Backend Handler | Unit: `httptest` + mock service | `testing`, `testify/assert` | Request parsing, response codes, auth context |
| Backend Service | Unit: mock repository via interface | `testing`, `testify/assert` | Business logic, error mapping |
| Backend Repository | Integration: real SQLite test DB | `testing`, `gorm` | Query correctness, filter behavior |
| Frontend Component | Unit: render + interaction | `vitest`, `@testing-library/react` | Props rendering, event handling |
| Frontend Hook | Unit: `renderHook` | `vitest`, `@testing-library/react` | Return values, callback invocations |
| Frontend API Module | Unit: mock `client` module | `vitest`, `vi.mock` | Endpoint paths, request shapes |

### Regression Target

**100% of existing test cases must pass with zero test file modifications.** This is the hard gate for every phase. The test suite count (number of test cases) must not decrease — any accidental removal is a failure.

### Per-Phase Testing Gates

| Phase | Gate | Command |
|-------|------|---------|
| All | All existing tests pass (100%, zero modifications) | `cd backend && go test ./...` + `npm test` |
| All | Type checking clean | `go vet ./...` + `npx tsc --noEmit` |
| 1 | No unused imports/vars | `go vet ./...` passes after deletions |
| 2 | Performance benchmark | `go test -bench=. -benchmem` for affected service methods |
| 3 | Duplication check | `jscpd --threshold 10` (frontend), `dupl` linter (backend) |
| 4 | File split verification | `npx tsc --noEmit` (import resolution) + `npm test` (runtime smoke) |
| 4 | Line count check | Script to verify no page file > 300 lines |

### Phase 2: Performance Benchmarks

Add benchmark functions for the fixed N+1 methods:

```go
func BenchmarkItemPoolHandler_List(b *testing.B) {
    // Setup: seed 200 items
    for i := 0; i < b.N; i++ {
        svc.List(ctx, teamID, filter, page)
    }
}
```

Target: p95 < 200ms or ≥40% reduction for datasets > 100 records.

## Security Considerations

### Threat Model

No new attack surface. All changes are internal refactoring:
- No new endpoints
- No schema changes (only adding indexes)
- No new dependencies
- No changes to auth/permission logic

### Mitigations

- Phase 1 contract fixes improve security by eliminating `as any` type escapes that could mask data shape issues
- Phase 2 SQL pushdown must use parameterized queries via GORM (not string concatenation) — existing pattern, no change needed
- Phase 3 generic helpers inherit the same parameterized query pattern from GORM

## Open Questions

- [x] TableView SQL strategy → **Hybrid**: SQL for single-entity, Go filter + batch names for cross-entity (resolved with user)
- [x] SubItem status dropdown "achievement" dialog → **Optional `onBeforeTerminalStatus` callback prop** on `StatusTransitionDropdown`. Rationale: a single component with an optional callback avoids duplicating the dropdown logic; callers that need the achievement dialog pass the callback, others omit it. This matches the interface defined in Phase 3 shared components.
- [x] `linkageMuMap` LRU capacity → **1000 entries**. Rationale: the map caches team-linkage mutexes; at any given time the active concurrency is bounded by the number of teams. 1000 entries covers teams with heavy linkage usage with comfortable headroom, and eviction is a simple delete-oldest when `len(map) >= cap`.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Full SQL for TableView | Maximum performance | Complex JOINs across main+sub items, fragile filter logic, hard to maintain | Cross-entity filtering in SQL is complex; hybrid achieves performance target with lower risk |
| New dependency for LRU cache | Battle-tested implementation | Adds dependency, PRD scope says no new deps | Simple map with capacity check is sufficient |
| Generic repository base struct | Reduces boilerplate most | Couples repos to GORM, breaks interface-driven pattern per DECISIONS.md | Generic helper functions in pkg/repo preserve interface pattern while reducing duplication |
| Separate StatusDropdown per item type | Type safety, no conditional logic | 3 components with shared logic, contradicts dedup goal | Single component with optional callback is simpler and meets dedup requirement |

### References

- `docs/ARCHITECTURE.md` — layer constraints
- `docs/DECISIONS.md` — existing technical decisions (constructor DI, interface-driven repos, VO layer)
- `docs/features/code-quality-cleanup/prd/prd-spec.md` — requirements source
