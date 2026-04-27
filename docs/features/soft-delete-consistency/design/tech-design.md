---
created: 2026-04-27
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Soft-Delete Consistency Fix

## Overview

Add `NotDeleted` scope to all repository query methods for entities embedding `BaseModel`, fix SubItem's broken soft-delete, unify TeamMember to use soft-delete, and align the `pmw_sub_items` unique index.

Single-layer change: Repository layer only. No handler, service, or frontend changes needed.

## Architecture

### Layer Placement

Repository layer (`internal/repository/gorm/`) + generic helpers (`internal/pkg/repo/helpers.go`). The fix is transparent to upper layers — API contracts don't change.

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│              pkg/repo/helpers.go                  │
│  FindByID[T] ← adds isSoftDeletable[T]() check  │
│  FindByIDs[T] ← adds isSoftDeletable[T]() check │
└──────────────────────┬──────────────────────────┘
                       │ delegates
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
┌────────────┐  ┌────────────┐  ┌────────────┐
│ user_repo  │  │ team_repo  │  │ main_item  │ ...
│ 5 methods  │  │ 8 methods  │  │ 6 methods  │
└────────────┘  └────────────┘  └────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌───────────┐ ┌──────────┐ ┌──────────┐
   │ sub_item  │ │item_pool │ │ role_repo│
   │ fix Delete│ │ 2 methods│ │ 3 joins  │
   └───────────┘ └──────────┘ └──────────┘
```

### Dependencies

No new external dependencies. All changes use existing `NotDeleted` / `NotDeletedTable` scopes defined in `internal/repository/gorm/scopes.go`:

```go
// NotDeleted filters records where deleted_flag = 0.
func NotDeleted(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_flag = 0")
}

// NotDeletedTable returns a scope for use in multi-table queries (joins).
func NotDeletedTable(table string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where(table + ".deleted_flag = 0")
    }
}
```

Existing internal modules used: `internal/model` (BaseModel), `internal/pkg/repo` (helpers), `gorm.io/gorm`.

## Key Design Decisions

### Decision 1: Runtime type switch in generic helpers

**Problem**: `FindByID[T]` and `FindByIDs[T]` are shared by both soft-deletable types (User, MainItem, etc.) and non-soft-deletable types (ProgressRecord, StatusHistory). Adding `NotDeleted` blindly would cause SQL errors on tables without `deleted_flag`.

**Options considered**:

| Approach | Pros | Cons |
|----------|------|------|
| A. Split `identifiable` into two type sets, two function names | Compile-time safety | Function name proliferation, caller changes |
| B. Runtime type switch `isSoftDeletable[T]()` | Zero caller changes, single function | Runtime dispatch (negligible cost) |
| C. Override at repo level (custom FindByID per repo) | No shared code changes | Code duplication, FindByIDs logic repeated |

**Chosen**: **B — Runtime type switch**. Minimal change, zero caller impact, `isSoftDeletable[T]()` compiles to a constant per instantiation.

```go
// isSoftDeletable returns true for model types that embed BaseModel (have deleted_flag).
func isSoftDeletable[T any]() bool {
    switch any(new(T)).(type) {
    case *model.ProgressRecord, *model.StatusHistory:
        return false
    default:
        return true
    }
}
```

### Decision 2: Unify TeamMember to soft-delete

**Problem**: `RemoveMember` uses hard delete (`db.Delete`), but TeamMember embeds BaseModel with `deleted_flag`. This is inconsistent.

**Change**: Replace `db.Delete(&model.TeamMember{})` with `Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()})`.

**Impact**:
- UK `uk_team_user_deleted(team_key, user_key, deleted_flag, deleted_time)` already supports soft-delete
- Re-adding a removed member creates a new row (deleted_flag=0) instead of conflicting
- `FindMember` needs `NotDeleted` so soft-deleted members are not found (allowing re-add)

## Interfaces

### Modified: `FindByID[T]` (pkg/repo/helpers.go)

```go
func FindByID[T any](db *gormlib.DB, ctx context.Context, id uint) (*T, error) {
    var item T
    query := db.WithContext(ctx)
    if isSoftDeletable[T]() {
        query = query.Scopes(NotDeleted)
    }
    err := query.First(&item, id).Error
    if err != nil {
        if errors.Is(err, gormlib.ErrRecordNotFound) {
            return nil, apperrors.ErrNotFound
        }
        return nil, err
    }
    return &item, nil
}
```

### Modified: `FindByIDs[T]` (pkg/repo/helpers.go)

```go
func FindByIDs[T identifiable](db *gormlib.DB, ctx context.Context, ids []uint) (map[uint]*T, error) {
    result := make(map[uint]*T)
    if len(ids) == 0 {
        return result, nil
    }
    query := db.WithContext(ctx)
    if isSoftDeletable[T]() {
        query = query.Scopes(NotDeleted)
    }
    var items []*T
    if err := query.Where("id IN ?", ids).Find(&items).Error; err != nil {
        return nil, err
    }
    for _, item := range items {
        result[getID(item)] = item
    }
    return result, nil
}
```

### Modified: Repo methods (per-repo changes)

Each repo method gets `.Scopes(NotDeleted)` added to its query chain. For join queries, use `NotDeletedTable(table)`. Two representative examples follow.

**Example 1 — Simple query: `user_repo.FindByBizKey`**

All repo files import `errors` from `"pm-work-tracker/backend/internal/pkg/errors"` (the same package aliased as `apperrors` in service/handler code).

```go
// BEFORE — no NotDeleted filter
func (r *userRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&user).Error
    if err != nil {
        if stderrors.Is(err, gormlib.ErrRecordNotFound) {
            return nil, errors.ErrNotFound
        }
        return nil, err
    }
    return &user, nil
}

// AFTER — adds .Scopes(NotDeleted) before .Where
func (r *userRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&user).Error
    if err != nil {
        if stderrors.Is(err, gormlib.ErrRecordNotFound) {
            return nil, errors.ErrNotFound
        }
        return nil, err
    }
    return &user, nil
}
```

All simple query methods (`FindByUsername`, `List`, `ListFiltered`, `FindByBizKey`, etc.) follow this same one-line addition pattern: insert `.Scopes(NotDeleted)` after `.WithContext(ctx)`.

**Example 2 — Join query: `role_repo.HasPermission`**

Join queries use `NotDeletedTable` with the table alias to avoid ambiguity.

```go
// BEFORE — no filter on pmw_team_members.deleted_flag
func (r *roleRepo) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Table("pmw_team_members").
        Joins("JOIN pmw_role_permissions ON pmw_role_permissions.role_id = pmw_team_members.role_key").
        Where("pmw_team_members.user_key = ? AND pmw_role_permissions.permission_code = ?", userID, code).
        Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

// AFTER — adds .Scopes(NotDeletedTable("pmw_team_members")) after .Table(...)
func (r *roleRepo) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Table("pmw_team_members").
        Scopes(NotDeletedTable("pmw_team_members")).
        Joins("JOIN pmw_role_permissions ON pmw_role_permissions.role_id = pmw_team_members.role_key").
        Where("pmw_team_members.user_key = ? AND pmw_role_permissions.permission_code = ?", userID, code).
        Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}
```

The other two join methods (`GetUserTeamPermissions`, `CountMembersByRoleID`) follow the same pattern: add `.Scopes(NotDeletedTable("pmw_team_members"))` after `.Table("pmw_team_members")`.

### Modified: `RemoveMember` (team_repo.go)

```go
func (r *teamRepo) RemoveMember(ctx context.Context, teamID, userID uint) error {
    result := r.db.WithContext(ctx).
        Model(&model.TeamMember{}).
        Where("team_key = ? AND user_key = ?", teamID, userID).
        Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()})
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.ErrNotFound
    }
    return nil
}
```

### Modified: `SoftDelete` (sub_item_repo.go)

```go
func (r *subItemRepo) SoftDelete(ctx context.Context, id uint) error {
    result := r.db.WithContext(ctx).
        Model(&model.SubItem{}).
        Where("id = ?", id).
        Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()})
    if result.Error != nil {
        return result.Error
    }
    // RowsAffected == 0 is silently ignored — idempotent delete.
    return nil
}
```

## Data Models

### BaseModel (central to this fix)

Defined in `internal/model/base.go`. Every entity that embeds `BaseModel` is soft-deletable via `deleted_flag`.

```go
type BaseModel struct {
    ID           uint      `gorm:"primarykey;autoIncrement" json:"-"`
    BizKey       int64     `gorm:"not null" json:"bizKey"`
    CreateTime   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
    DbUpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"dbUpdateTime"`
    DeletedFlag  int       `gorm:"not null;default:0;index" json:"-"`
    DeletedTime  time.Time `gorm:"not null;default:'1970-01-01 08:00:00'" json:"-"`
}
```

Fields relevant to this fix:
- `DeletedFlag` — `0` = active, `1` = soft-deleted. The `NotDeleted` scope filters on `deleted_flag = 0`.
- `DeletedTime` — set to `time.Now()` on soft-delete. Included in unique indexes to allow re-creation after deletion.

Entities embedding BaseModel: User, Team, MainItem, SubItem, ItemPool, Role, TeamMember.
Non-soft-deletable entities (no BaseModel): ProgressRecord, StatusHistory.

No new models. One schema change:

### Schema: `pmw_sub_items` unique index

```sql
-- Before
UNIQUE KEY uk_sub_items_main_code (main_item_key, item_code)

-- After
UNIQUE KEY uk_sub_items_main_code (main_item_key, item_code, deleted_flag, deleted_time)
```

Applied to both `MySql-schema.sql` and `SQLite-schema.sql`. Manual ALTER TABLE required for existing databases.

## Error Handling

### Error Types

No new error types. The existing `ErrNotFound` is the primary error produced by this change:

```go
// Defined in internal/pkg/errors/errors.go
var ErrNotFound = &AppError{Code: "NOT_FOUND", Status: 404, Message: "resource not found"}
```

**Package alias clarification**: Repo files import this package as `errors` (`"pm-work-tracker/backend/internal/pkg/errors"`). Service and handler files import the same package as `apperrors`. They are the same package — `errors.ErrNotFound` and `apperrors.ErrNotFound` are the same value. The design uses both aliases to match each layer's actual import convention.

Domain-specific not-found errors also exist: `ErrItemNotFound`, `ErrRoleNotFound`, `ErrTeamNotFound`, `ErrUserNotFound` — all with HTTP 404 status.

### Error Propagation

The `MapNotFound` helper in `internal/pkg/errors/errors.go` converts generic `ErrNotFound` or GORM's `gorm.ErrRecordNotFound` into domain-specific errors:

```go
func MapNotFound(err error, domainErr *AppError) error {
    if err == nil { return nil }
    if stderrors.Is(err, gorm.ErrRecordNotFound) || stderrors.Is(err, ErrNotFound) {
        return domainErr
    }
    return err
}
```

### Cross-Layer Error Trace

Although no handler or service code changes, the design must confirm that soft-deleted records produce the correct HTTP response. Two concrete traces:

**Trace 1: SubItem GetByBizKey (PRD Story 2 — deleted sub-item returns 404)**

```
1. Handler: SubItemHandler.Get
   → pkgHandler.ResolveBizKey(c, "subId", lookupFn)
   → lookupFn calls svc.GetByBizKey(ctx, bizKey)

2. Service: subItemService.GetByBizKey
   → s.subItemRepo.FindByBizKey(ctx, bizKey)
   → repo.FindByBizKey queries: SELECT * FROM pmw_sub_items WHERE biz_key = ? AND deleted_flag = 0
   → returns gorm.ErrRecordNotFound for soft-deleted record

3. Service: apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
   → converts gorm.ErrRecordNotFound → ErrItemNotFound{Code:"ITEM_NOT_FOUND", Status:404}

4. Handler: ResolveBizKey calls apperrors.RespondError(c, err)
   → RespondError extracts AppError.Status = 404
   → HTTP response: 404 {"code":"ITEM_NOT_FOUND","message":"item not found"}
```

**Trace 2: Role GetRole (PRD Story 1 — deleted role bizKey returns 404)**

```
1. Handler: RoleHandler.GetRole
   → pkgHandler.ParseBizKeyParam(c, "id") → bizKey
   → svc.GetRole(ctx, bizKey)

2. Service: roleService.GetRole
   → s.roleRepo.FindByBizKey(ctx, roleBizKey)
   → repo.FindByBizKey queries: SELECT * FROM pmw_roles WHERE biz_key = ? AND deleted_flag = 0
   → returns ErrNotFound for soft-deleted role

3. Service: apperrors.MapNotFound(err, ErrRoleNotFound)
   → converts ErrNotFound → ErrRoleNotFound{Code:"ERR_ROLE_NOT_FOUND", Status:404}

4. Handler: apperrors.RespondError(c, err)
   → RespondError extracts AppError.Status = 404
   → HTTP response: 404 {"code":"ERR_ROLE_NOT_FOUND","message":"角色不存在"}
```

Both traces confirm: no handler or service changes needed. The existing propagation path correctly maps repository-layer `ErrNotFound` to the appropriate HTTP 404 response.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Repository | Unit | SQLite in-memory + testify/assert + testify/require | NotDeleted filtering per method | 100% of changed methods |

### Test Pattern

Each affected repo gets a `TestNotDeleted_*` test following this pattern:

```go
func TestNotDeleted_<Repo>_<Method>(t *testing.T) {
    db := setupTestDB(t)  // SQLite in-memory + AutoMigrate
    repo := NewGorm<Repo>(db)
    ctx := context.Background()

    // Create a record
    record := seed<Record>(t, db, ...)

    // Soft-delete it
    db.Model(record).Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()})

    // Verify method excludes the deleted record
    _, err := repo.<Method>(ctx, ...)
    assert.ErrorIs(t, err, ErrNotFound)  // or not in list results
}
```

### Key Test Scenarios

1. **FindByID with soft-deleted record** → returns ErrNotFound for User, MainItem, SubItem, ItemPool
2. **FindByID for non-soft-deletable types** → returns record normally for ProgressRecord, StatusHistory
3. **List methods exclude soft-deleted** → List, ListFiltered, SearchAvailable for each repo
4. **SubItem SoftDelete** → sets deleted_flag=1, item disappears from List
5. **SubItem re-create after soft-delete** → same item_code succeeds
6. **TeamMember RemoveMember** → soft-deletes, FindMember returns ErrNotFound, re-add succeeds
7. **Role repo join queries** → HasPermission/GetUserTeamPermissions/CountMembersByRoleID exclude deleted members

### Overall Coverage Target

100% of modified methods (each changed method has at least one test verifying NotDeleted behavior).

## Security Considerations

### Threat Model

| # | Threat | Impact | Likelihood | Mitigation | Verification Test |
|---|--------|--------|------------|------------|-------------------|
| T1 | Deleted role still granting permissions via `HasPermission` join query | A deleted role's permission codes remain in `pmw_role_permissions`. If `HasPermission` does not exclude soft-deleted team members from its join on `pmw_team_members`, a removed member retains their permissions. | Medium — `HasPermission` currently has no `NotDeleted` filter on `pmw_team_members`. | Add `NotDeletedTable("pmw_team_members")` to `HasPermission` query. The role's `pmw_role_permissions` rows are not a concern because `role_repo.FindByBizKey` already excludes deleted roles from listing/editing; the join is on member, not role. | `TestNotDeleted_Role_HasPermission`: create member with role, soft-delete the member via `RemoveMember`, call `HasPermission(userID, code)`, assert returns `false`. |
| T2 | Soft-deleted team member still counted in `CountMembersByRoleID` | Admin UI shows inflated member counts for roles, and `DeleteRole` refuses deletion (`ErrRoleInUse`) even when all members are soft-deleted. | Medium — `CountMembersByRoleID` currently has no `NotDeleted` filter. | Add `NotDeletedTable("pmw_team_members")` to `CountMembersByRoleID` query. | `TestNotDeleted_Role_CountMembersByRoleID`: create member with role, soft-delete the member, call `CountMembersByRoleID(roleID)`, assert returns `0`. |
| T3 | Deleted SubItem data leaks via direct bizKey lookup | User bookmarks a sub-item URL or shares a direct link. If `FindByBizKey` does not filter by `deleted_flag`, the deleted sub-item's data (title, description, assignee) is returned. | Low — `FindByBizKey` in `sub_item_repo.go` currently has no `NotDeleted` scope. | Add `.Scopes(NotDeleted)` to `subItemRepo.FindByBizKey`. | `TestNotDeleted_SubItem_FindByBizKey`: create sub-item, soft-delete it, call `FindByBizKey(bizKey)`, assert returns `ErrNotFound`. |
| T4 | Migration window: duplicate rows during index rebuild | MySQL ALTER TABLE causes implicit commits — `BEGIN; DROP INDEX; ADD INDEX; COMMIT;` does NOT wrap atomically. Between the two ALTER statements, concurrent INSERTs could create duplicate `(main_item_key, item_code)` rows that violate the old index. | Medium — the window is small but non-zero on busy tables. | Use a single `ALTER TABLE` statement with both operations: `ALTER TABLE pmw_sub_items DROP INDEX uk_sub_items_main_code, ADD UNIQUE KEY uk_sub_items_main_code (main_item_key, item_code, deleted_flag, deleted_time);` — MySQL executes compound ALTER atomically per table. Apply during low-traffic window. For SQLite (dev/test), DDL is instantaneous. | Integration test: run ALTER in a test, verify no duplicate rows possible by attempting a conflicting insert. |
| T5 | Concurrent re-create race after soft-delete | Two requests simultaneously re-create a SubItem with the same `item_code` after soft-deletion. Both see `deleted_flag=1` for the old row and attempt INSERT. | Low — GORM's `Create` uses a single INSERT, and the new unique index `(main_item_key, item_code, deleted_flag, deleted_time)` includes `deleted_time` which differs per request, so both rows are distinct under the index. | No code change needed — the unique index with `deleted_time` naturally prevents collision. MySQL's default isolation level (REPEATABLE READ) ensures consistent reads. If a collision did occur (e.g., same `deleted_time` down to nanosecond), the unique index rejects the second INSERT, returning a MySQL duplicate-key error that maps to `ErrAlreadyExists` via the existing error handler. | `TestNotDeleted_SubItem_ReCreate`: soft-delete a sub-item, re-create with same `item_code` (succeeds), soft-delete again, re-create again (succeeds). Verify two distinct rows exist with different `deleted_time` values. |

### Defense-in-Depth: Preventing Future Regressions

The root cause of this bug is repo methods added without `NotDeleted`. To prevent recurrence:

1. **Convention enforcement**: All new repo methods on soft-deletable entities must include `.Scopes(NotDeleted)`. Documented in `scopes.go` godoc.
2. **Test pattern mandatory**: Every repo method test suite must include a `TestNotDeleted_*` case that soft-deletes a record and verifies exclusion. The 7 test scenarios above serve as the template.
3. **CI grep check**: A lightweight enforcement script runs in CI to catch repo methods that query soft-deletable tables without `NotDeleted`:
   ```bash
   # Fails if any repo file has a .Find( or .First( call without NotDeleted in the same function
   ! grep -rn '\.Find(\|\.First(' backend/internal/repository/gorm/*.go \
       | grep -v 'NotDeleted' \
      | grep -v '_test.go'
   ```
   This catches the most common regression pattern — adding a new query method without the scope. It is a heuristic, not a guarantee, but covers the root cause of this bug.

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|--------|------------------|-------------------|
| Story 1 AC: deleted role not in list | role_repo.List (already has NotDeleted) | `List()` + NotDeleted |
| Story 1 AC: deleted role bizKey returns 404 | role_repo.FindByBizKey (already has NotDeleted) | `FindByBizKey()` + NotDeleted → HTTP 404 via MapNotFound → ErrRoleNotFound |
| Story 2 AC: deleted sub-item disappears | sub_item_repo.SoftDelete + List | `SoftDelete()` fix + `List()` + NotDeleted |
| Story 2 AC: re-create with same item_code | Schema fix uk_sub_items_main_code | Index includes deleted_flag, deleted_time |
| Story 3 AC: FindByID filters soft-deleted User | helpers.go FindByID[T] | `isSoftDeletable[T]()` + NotDeleted |
| Story 3 AC: FindByID works for ProgressRecord | helpers.go FindByID[T] | `isSoftDeletable[T]()` returns false |
| Story 4 AC1: HasPermission excludes deleted member | role_repo.HasPermission | `NotDeletedTable("pmw_team_members")` |
| Story 4 AC2: GetUserTeamPermissions excludes | role_repo.GetUserTeamPermissions | `NotDeletedTable("pmw_team_members")` |
| Story 4 AC3: CountMembersByRoleID excludes | role_repo.CountMembersByRoleID | `NotDeletedTable("pmw_team_members")` |
| Story 4 AC4: multiple deleted members | Same as AC3 | Count returns correct value |
| Story 4 AC5: deleted member's all-team perms excluded | Same as AC2 | Empty map returned |
| TeamMember: FindMember excludes deleted | team_repo.FindMember | `.Scopes(NotDeleted)` — soft-deleted member not found, allowing re-add |
| TeamMember: CountMembers excludes deleted | team_repo.CountMembers | `.Scopes(NotDeleted)` — count only active members |
| TeamMember: ListMembers excludes deleted users | team_repo.ListMembers | `NotDeletedTable("pmw_users")` — deleted users not in member list |
| TeamMember: RemoveMember uses soft-delete | team_repo.RemoveMember | `Updates(deleted_flag=1)` instead of hard DELETE |

## Open Questions

- [x] TeamMember: hard-delete vs soft-delete → Decided: unify to soft-delete
- [x] Unique index format → Decided: always include (deleted_flag, deleted_time)

## Appendix

### Implementation Phase Order

Phases must execute in order due to dependencies:

| Phase | Files | Dependency | Rationale |
|-------|-------|------------|-----------|
| **P1: Generic helpers** | `pkg/repo/helpers.go` | None | Foundation — `FindByID[T]`/`FindByIDs[T]` fix propagates to all repos that delegate to them |
| **P2: Repo methods** | `user_repo.go`, `team_repo.go`, `main_item_repo.go`, `sub_item_repo.go`, `item_pool_repo.go`, `role_repo.go` | P1 | Each repo's custom methods (FindByBizKey, List, etc.) need the `NotDeleted` scope. Repos using generic helpers get P1's fix automatically |
| **P3: SubItem SoftDelete fix** | `sub_item_repo.go` | P2 | SoftDelete implementation change; must run after List/FindByBizKey have NotDeleted so deleted sub-items are properly excluded |
| **P4: TeamMember soft-delete** | `team_repo.go` (RemoveMember) | P2 | Unifies RemoveMember to soft-delete; FindMember/CountMembers must already have NotDeleted from P2 |
| **P5: Schema migration** | `migrations/MySql-schema.sql`, `migrations/SQLite-schema.sql` | P3 | Index change depends on SubItem soft-delete being correct; apply during low-traffic window |

Within each phase, files can be changed in any order (no cross-file dependencies within a phase).

### Complete Change List

| File | Change | Methods |
|------|--------|---------|
| `pkg/repo/helpers.go` | Add `isSoftDeletable[T]()`, apply in `FindByID`, `FindByIDs` | 2 methods + 1 helper |
| `gorm/user_repo.go` | Add `.Scopes(NotDeleted)` | FindByBizKey, FindByUsername, List, ListFiltered, SearchAvailable |
| `gorm/team_repo.go` | Add `.Scopes(NotDeleted)` or `NotDeletedTable`; fix RemoveMember | List, ListFiltered, FindByBizKey, ListMembers, FindMember, CountMembers, FindPMMembers, FindTeamsByUserIDs, RemoveMember |
| `gorm/main_item_repo.go` | Add `.Scopes(NotDeleted)` | FindByBizKey, List, CountByTeam, ListNonArchivedByTeam, FindByBizKeys, ListByTeamAndStatus |
| `gorm/sub_item_repo.go` | Fix SoftDelete; add `.Scopes(NotDeleted)` | SoftDelete, FindByBizKey, List, ListByMainItem, ListByTeam |
| `gorm/item_pool_repo.go` | Add `.Scopes(NotDeleted)` | FindByBizKey, List |
| `gorm/role_repo.go` | Add `.Scopes(NotDeletedTable("pmw_team_members"))` | HasPermission, GetUserTeamPermissions, CountMembersByRoleID |
| `migrations/MySql-schema.sql` | Update uk_sub_items_main_code | DDL |
| `migrations/SQLite-schema.sql` | Update uk_sub_items_main_code | DDL |

### Migration SQL

```sql
-- MySQL — single atomic ALTER (avoids implicit commit window between DROP and ADD)
ALTER TABLE pmw_sub_items
  DROP INDEX uk_sub_items_main_code,
  ADD UNIQUE KEY uk_sub_items_main_code (main_item_key, item_code, deleted_flag, deleted_time);
```
