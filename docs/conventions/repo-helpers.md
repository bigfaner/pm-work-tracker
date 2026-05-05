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
