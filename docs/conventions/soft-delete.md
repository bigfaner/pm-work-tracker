---
scope: backend
source: feature/soft-delete-consistency, feature/jlc-schema-alignment
verified: "2026-05-04"
---

# Soft-Delete Conventions

Rules for implementing soft-delete across all repository layers. Every new repo method and schema change must follow these rules.

## SD-001: Soft-Deletable vs Non-Soft-Deletable Entities

**Rule**: An entity is soft-deletable if and only if it embeds `BaseModel` (which provides `deleted_flag` and `deleted_time`).

**Classification**:

| Category | Entities | Reason |
|----------|----------|--------|
| Soft-deletable | User, Team, MainItem, SubItem, ItemPool, Role, TeamMember, MilestoneMap, Milestone | Embed BaseModel; can be logically removed without data loss |
| Non-soft-deletable | ProgressRecord, StatusHistory, DecisionLog | Append-only tables; no biz_key, no deleted_flag. Records are never removed. |

**Why**: ProgressRecord and StatusHistory are audit/append-only tables -- deleting records would break historical integrity. TeamMember is soft-deletable because members can be removed from a team but the relationship record must be preserved for auditing and re-add scenarios.

## SD-002: NotDeleted Scope

**Rule**: All queries on soft-deletable entities must use the `NotDeleted` scope. Never use inline `WHERE deleted_flag = 0`.

**Two scope variants**:

| Scope | Signature | Generated SQL | When to use |
|-------|-----------|---------------|-------------|
| `NotDeleted` | `func(db *gorm.DB) *gorm.DB` | `WHERE deleted_flag = 0` | Single-table queries |
| `NotDeletedTable(table)` | `func(table string) func(db *gorm.DB) *gorm.DB` | `WHERE {table}.deleted_flag = 0` | Multi-table JOINs (avoids column ambiguity) |

**Why**: Centralizing the filter condition in a scope prevents bugs from forgotten WHERE clauses and makes it easy to grep for violations.

**Example -- single table**:
```go
// All Find/First/Count/List calls on soft-deletable entities
err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&item).Error
```

**Example -- JOIN query**:
```go
// Filter on a specific joined table
err := r.db.WithContext(ctx).
    Table("pmw_team_members").
    Scopes(NotDeletedTable("pmw_team_members")).
    Joins("JOIN pmw_role_permissions ON ...").
    Where(...).
    Count(&count).Error
```

**Violation check**:
```bash
# Should return zero results -- all inline deleted_flag checks replaced by scope
grep -rn "deleted_flag.*=.*0" backend/internal/repository/ | grep -v '_test.go'
```

## SD-003: SoftDelete Method Pattern

**Rule**: Each soft-deletable repo exposes a `SoftDelete(ctx, id uint) error` method. Never use `db.Delete()` outside the SoftDelete method. The SoftDelete method is the only place that sets `deleted_flag = 1`.

**Why**: Encapsulating the update logic prevents callers from accidentally hard-deleting or from forgetting to set `deleted_time`. The repo interface does not expose a hard-delete method.

**Implementation pattern**:
```go
func (r *subItemRepo) SoftDelete(ctx context.Context, id uint) error {
    result := r.db.WithContext(ctx).
        Model(&model.SubItem{}).
        Where("id = ?", id).
        Updates(map[string]any{
            "deleted_flag": 1,
            "deleted_time": time.Now(),
        })
    if result.Error != nil {
        return result.Error
    }
    return nil
}
```

**Constraints**:
- Always use `Updates(map[string]any{...})`, never `db.Delete()`.
- The WHERE clause must include `deleted_flag = 0` guard (or rely on NotDeleted scope) to prevent double-soft-delete.
- `RowsAffected == 0` handling is repo-specific: some repos return `ErrNotFound`, others silently succeed (idempotent).

## SD-004: Generic Helpers with isSoftDeletable

**Rule**: Generic repo helpers (`FindByID[T]`, `FindByIDs[T]`) use a runtime type switch to automatically apply NotDeleted for soft-deletable types.

**Why**: These helpers serve both soft-deletable (User, MainItem) and non-soft-deletable (ProgressRecord, StatusHistory, DecisionLog) types. Adding NotDeleted blindly would cause SQL errors on tables without `deleted_flag`. A runtime type switch adds NotDeleted only when appropriate, with zero caller-side changes.

**Implementation**:
```go
func isSoftDeletable[T any]() bool {
    switch any(new(T)).(type) {
    case *model.ProgressRecord, *model.StatusHistory, *model.DecisionLog:
        return false
    default:
        return true
    }
}

func FindByID[T any](db *gorm.DB, ctx context.Context, id uint) (*T, error) {
    var item T
    query := db.WithContext(ctx)
    if isSoftDeletable[T]() {
        query = query.Where("deleted_flag = 0")
    }
    err := query.First(&item, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrNotFound
        }
        return nil, err
    }
    return &item, nil
}
```

**When adding a new non-soft-deletable type**: Add it to the `case` list in `isSoftDeletable`. All soft-deletable types are handled by the `default` branch automatically.

## SD-005: Unique Index Includes deleted_flag and deleted_time

**Rule**: Business unique indexes on soft-deletable tables must include `deleted_flag` and `deleted_time` as trailing columns. This allows re-creating a record with the same business key after soft-deletion.

**Why**: Without `deleted_flag` in the unique index, soft-deleting a record blocks creating a new record with the same business key. Including `deleted_time` ensures multiple soft-deleted records with the same key are distinct.

**Example**:
```sql
-- Allows multiple soft-deleted rows with same (team_key, code),
-- but only one active row per (team_key, code)
UNIQUE KEY uk_teams_code_deleted (code, deleted_flag, deleted_time)
UNIQUE KEY uk_main_items_team_code_deleted (team_key, code, deleted_flag, deleted_time)
UNIQUE KEY uk_sub_items_main_code (main_item_key, code, deleted_flag, deleted_time)
UNIQUE KEY uk_team_user_deleted (team_key, user_key, deleted_flag, deleted_time)
```

## SD-006: Cascade Soft-Delete for Main Items

**Rule**: Deleting a main item soft-deletes all its sub-items within a single database transaction. The transaction sequence is:
1. Fetch main item by bizKey
2. List all sub-items under the main item
3. Batch `SoftDelete` all sub-items
4. `SoftDelete` the main item
5. Insert `status_histories` audit entries for each deleted item (`from_status=current`, `to_status="deleted"`)
6. Recalculate parent completion percentages if needed

The `item_status` field is NOT modified during deletion -- only `deleted_flag` and `deleted_time` are set.

**Why**: Ensures data consistency by preventing orphaned sub-items. Single transaction guarantees atomicity. Audit trail in `status_histories` provides deletion traceability without modifying the item's status field.

**Source**: feature/system-ux-optimization BIZ-003, BIZ-004

## SD-007: Cascade Soft-Delete for MilestoneMap

**Rule**: Deleting a MilestoneMap soft-deletes all its Milestones within a single database transaction. The transaction also clears `milestone_key` on all related MainItems (sets to NULL). This cascade bypasses individual Milestone deletion constraints (only `not_started`/`cancelled` Milestones can be deleted individually, but map-level deletion ignores this).

**Transaction sequence**:
1. List all Milestones under the MilestoneMap
2. Batch clear `milestone_key` on all MainItems associated with those Milestones
3. Batch `SoftDelete` all Milestones
4. `SoftDelete` the MilestoneMap

**Why**: Prevents orphaned Milestones and dangling `milestone_key` references. Single transaction guarantees atomicity.

**Source**: feature/milestone-map BR-4, BR-6

## SD-008: Milestone Soft-Delete Unbinds Related Items

**Rule**: Deleting a Milestone sets all related MainItems' `milestone_key` to NULL within the same transaction.

**Transaction sequence**:
1. Clear `milestone_key` on all MainItems where `milestone_key = milestone.BizKey`
2. `SoftDelete` the Milestone

**Why**: Prevents dangling foreign key references to deleted Milestones.

**Source**: feature/milestone-map BR-4
