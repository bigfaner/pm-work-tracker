---
scope: backend
source: feature/soft-delete-consistency, feature/schema-alignment-cleanup
---

# Repository Helper Conventions

Shared generic helpers in `pkg/repo/` and transaction abstractions used across the repository layer.

## RH-001: FindByID with Automatic NotDeleted

**Rule**: `FindByID[T]` automatically applies the `NotDeleted` scope for soft-deletable types. Callers never need to add the scope manually.

**Why**: Without automatic filtering, every caller of `FindByID` on a soft-deletable entity would need to remember to add `.Scopes(NotDeleted)`. Forgetting this is the root cause of "ghost data" bugs where deleted records appear in API responses.

**Signature**:
```go
func FindByID[T any](db *gorm.DB, ctx context.Context, id uint) (*T, error)
```

**Behavior by type**:

| Type category | NotDeleted applied? | Example types |
|---------------|---------------------|---------------|
| Soft-deletable (embeds BaseModel) | Yes | User, Team, MainItem, SubItem, ItemPool, Role, TeamMember |
| Non-soft-deletable | No | ProgressRecord, StatusHistory |

**Error mapping**: `gorm.ErrRecordNotFound` is converted to `apperrors.ErrNotFound`. Soft-deleted records also return `ErrNotFound` (indistinguishable from never-existed).

## RH-002: FindByIDs Returns Map with Partial Results

**Rule**: `FindByIDs[T]` returns `map[uint]*T`. If some IDs are not found (including soft-deleted), the map contains only the found entries. No error is returned for partial results.

**Why**: Batch lookups are typically used for display (e.g., resolving user names for a list of items). Failing the entire batch because one ID is missing would break the UI for all items.

**Signature**:
```go
func FindByIDs[T identifiable](db *gorm.DB, ctx context.Context, ids []uint) (map[uint]*T, error)
```

**Behavior**:
- Empty `ids` slice: returns empty map, no query executed.
- Some IDs not found: map contains only found entries, no error.
- NotDeleted scope applied automatically for soft-deletable types (same as FindByID).
- Map key is the record's internal `uint` ID.

## RH-003: UpdateFields with ColumnExpr Type Safety

**Rule**: Use `UpdateFields` or GORM's `Updates(map[string]any{...})` for partial column updates. Column names in the map must match the database column name (snake_case), not the Go field name.

**Why**: GORM's `Updates` with a map only updates the specified columns. Using a struct would zero out unset fields. Column names must be exact -- a typo (e.g., `assignee_id` instead of `assignee_key`) silently writes to a non-existent column with no error.

**Pattern**:
```go
err := r.db.WithContext(ctx).
    Model(&model.SubItem{}).
    Where("id = ?", id).
    Updates(map[string]any{
        "assignee_key": assigneeBizKey,
        "item_status":  newStatus,
    }).Error
```

**Common pitfalls**:
- `assignee_id` is wrong; the column is `assignee_key` (stores bizKey, not internal ID).
- `status` is wrong; the column is `item_status` (entity-prefixed per DM-004).
- Always verify column names against the schema DDL, not Go struct field names.

## RH-004: DBTransactor Interface for Transactions

**Rule**: Services that need database transactions depend on the `DBTransactor` interface, not on `*gorm.DB` directly.

**Why**: Two services (`team_service`, `item_pool_service`) each defined their own local transaction interface. Consolidating to a single shared interface eliminates duplication and ensures all services use the same transaction contract.

**Interface**:
```go
// pkg/repo/transactor.go
type DBTransactor interface {
    Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}
```

**Usage**:
```go
type teamService struct {
    teamRepo   repository.TeamRepo
    transactor repo.DBTransactor
    // ...
}

func (s *teamService) InviteMember(ctx context.Context, ...) error {
    return s.transactor.Transaction(func(tx *gorm.DB) error {
        // all operations within tx
    })
}
```

**Constraint**: Never import `*gorm.DB` directly in service constructors. Depend on the `DBTransactor` interface for testability and consistency.

## RH-005: Shared Interface Change Protocol

**Rule**: When adding methods to a shared repository interface (any interface with >3 test consumers), the interface change MUST be done as a separate step before the feature implementation task.

**Why**: Adding one method to a shared interface (e.g., `MainItemRepo` with 17 methods) triggers O(consumers) mock updates. Each of the 9 mock types must implement all 17+ methods. An agent tasked with "implement feature X" will spend its entire budget fixing cascading mock compilation errors instead of writing feature logic, causing stalls.

**Protocol**:

1. **Interface update task** — Add new method(s) to the interface AND all existing mock/stub types. No business logic. Verify all tests compile and pass.
2. **Feature implementation task** — Implement the actual feature using the already-updated interface. Mocks already compile, agent focuses on business logic.

**Detection**: Applies when a task modifies any of these high-consumer interfaces:
- `MainItemRepo` (9 test consumers)
- `SubItemRepo` (multiple consumers)
- Any interface with >5 methods

**Example**:
```go
// Step 1 (interface update task): Add method to MainItemRepo
type MainItemRepo interface {
    // ... existing 16 methods ...
    CalcCompletionByMap(ctx context.Context, mapBizKey int64) (float64, error) // NEW
}

// Fix ALL mock types immediately:
// - mockMainItemRepo (main_item_service_test.go)
// - mockMainItemRepoForPool (item_pool_service_test.go, item_pool_handler_test.go)
// - mockViewMainItemRepo (view_service_test.go)
// - msMockMainItemRepo (milestone_service_test.go)
// - mmMockMainItemRepo (milestone_map_service_test.go)
// - trackingMainItemRepo (item_pool_handler_test.go)
// - StubRouterRepoMainItem (router_test_stubs.go)

// Step 2 (feature task): Use the method in service logic
func (s *milestoneMapService) enrichComputedFields(ctx context.Context, m *model.MilestoneMap) {
    avg, _ := s.mainItemRepo.CalcCompletionByMap(ctx, m.BizKey)
    m.OverallProgress = avg
}
```
