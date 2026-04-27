# Proposal: Soft-Delete Consistency Fix

## Problem

Soft-deleted records (deleted_flag=1) leak into API responses across multiple entities. The root cause is systemic: most repository query methods skip the `NotDeleted` scope, and one unique index (`pmw_sub_items`) doesn't account for soft-deleted rows.

**Evidence:**
- Role deletion bug reported by user: deleted roles still appear in list (role_repo.List missing NotDeleted)
- SubItem `SoftDelete` uses `db.Delete()` (GORM built-in) which doesn't set `deleted_flag`/`deleted_time`
- `pmw_sub_items.uk_sub_items_main_code` index blocks re-creating sub-items with same code after soft-delete
- `FindByID[T]` / `FindByIDs[T]` generic helpers return soft-deleted records for all entities

**Urgency:** P1 user-reported data integrity bug. Affects all teams using role management and sub-item features across 7 repository files (~25 query methods). Deleted roles leak into permission checks, potentially granting or denying access incorrectly. Deleted sub-items remain visible and block re-creation with the same code due to a unique constraint collision. There is no workaround -- deleted data is indistinguishable from active data in API responses because `FindByID[T]` and `FindByIDs[T]` lack the `NotDeleted` scope.

## Solution

Systematically add `NotDeleted` scope to all repository query methods for entities with `BaseModel`, fix SubItem soft-delete implementation, and align the sub_items unique index.

### User-Facing Impact

**Team / Role:**
- Before: After deleting a role, the deleted role still appears in the team role list and permission assignment dropdowns. Deleting a team leaves it visible in the team selector.
- After: Deleted roles and teams are excluded from all list responses and dropdowns immediately.

**SubItem:**
- Before: Soft-deleting a sub-item does not set `deleted_flag`/`deleted_time` (GORM's `db.Delete()` uses a different column). The sub-item remains visible. Attempting to create a new sub-item with the same `item_code` under the same main item triggers a unique constraint error (`uk_sub_items_main_code`).
- After: Soft-deleted sub-items disappear from all list/get responses. A new sub-item can be created with the same `item_code` after deletion without constraint errors.

**User / MainItem / ItemPool:**
- Before: No delete feature exists yet, but `FindByID[T]` and `FindByIDs[T]` generic helpers would return soft-deleted rows if delete were added, because they lack the `NotDeleted` scope today.
- After: The generic helpers filter out soft-deleted records. When delete features are added later, no further repo-level filtering changes are needed — the defense is already in place.

## Alternatives

### 1. Fix all repos (Recommended)

Apply `NotDeleted` to every repository method for all entities with `BaseModel`, including entities that don't yet have delete flows.

| Aspect | Assessment |
|--------|-----------|
| **Pros** | Defense in depth — prevents future regressions when delete features are added to User/MainItem/ItemPool. Single PR covers everything; no follow-up needed. Consistent behavior across all entities reduces cognitive load for future contributors. |
| **Cons** | Touches 7 repo files + 2 generic helpers (~15 methods total). Larger diff means more review time and wider regression surface. Generic helper split (`softDeletable` vs `nonSoftDeletable`) adds type-level complexity. |
| **Effort** | ~2-3 days: 1 day for code changes, 1 day for tests, 0.5 day for schema migration, 0.5 day buffer for generic constraint refactor. |
| **Blast radius** | Medium — every Find/List method changes, but each change is mechanical (add one scope call). Regression risk contained by per-method tests. |
| **Maintenance** | Low — future delete features require zero repo-level filtering changes. |

### 2. Fix only repos with active delete flows

Apply `NotDeleted` only to Team, Role, and SubItem repos (entities where delete is already implemented).

| Aspect | Assessment |
|--------|-----------|
| **Pros** | Smaller diff (~6 methods). Lower regression risk in the short term. Ships faster (1-2 days). No generic constraint split needed since ProgressRecord/StatusHistory repos are untouched. |
| **Cons** | User/MainItem/ItemPool repos will return soft-deleted records the moment delete features are added — a latent bug that will be missed unless the contributor knows to add `NotDeleted`. The generic helpers (`FindByID[T]`, `FindByIDs[T]`) remain unsafe. Requires a follow-up PR later. |
| **Effort** | ~1-2 days: surgical changes to 3 repo files + SubItem schema fix. |
| **Blast radius** | Low — changes are scoped to actively-used delete paths only. |
| **Maintenance** | High — every future delete feature must remember to add `NotDeleted` to that entity's repo methods. Easy to forget, no compile-time safety. |

### 3. Do nothing

| Aspect | Assessment |
|--------|-----------|
| **Pros** | Zero effort. No regression risk from code changes. |
| **Cons** | Phantom records persist for Team/Role/SubItem. Sub-item re-creation with same code remains broken. The bug is user-reported and affects data integrity — inaction is not viable. |

**Rationale for Alternative 1:** The incremental cost over Alternative 2 is ~1 extra day, but it eliminates an entire class of future bugs. The generic helper split is a one-time complexity cost that pays off permanently. Alternative 2's "fix it later" approach relies on future contributors remembering to add `NotDeleted`, which is exactly how the current bug was created.

## Scope

### In-Scope

1. **Generic helpers** (`pkg/repo/helpers.go`)
   - `FindByID[T]` — add NotDeleted for soft-deletable types
   - `FindByIDs[T]` — add NotDeleted for soft-deletable types
   - Challenge: ProgressRecord and StatusHistory don't have deleted_flag; need separate generic constraints

2. **User repo** (`user_repo.go`)
   - `FindByBizKey`, `FindByUsername`, `List`, `ListFiltered`, `SearchAvailable`

3. **Team repo** (`team_repo.go`)
   - `List`, `ListFiltered`, `FindByBizKey`, `ListMembers`, `FindMember`, `CountMembers`, `FindPMMembers`, `FindTeamsByUserIDs`
   - Note: `FindByID` and `ListAllTeams` already have NotDeleted

4. **MainItem repo** (`main_item_repo.go`)
   - `FindByBizKey`, `List`, `CountByTeam`, `ListNonArchivedByTeam`, `FindByBizKeys`, `ListByTeamAndStatus`

5. **SubItem repo** (`sub_item_repo.go`)
   - Fix `SoftDelete` to use `Updates(map[string]any{"deleted_flag": 1, "deleted_time": now()})` instead of `db.Delete()`
   - `FindByBizKey`, `List`, `ListByMainItem`, `ListByTeam`

6. **ItemPool repo** (`item_pool_repo.go`)
   - `FindByBizKey`, `List`

7. **Role repo** (`role_repo.go`) — already partially fixed
   - `HasPermission`, `GetUserTeamPermissions`, `CountMembersByRoleID` — add NotDeletedTable("pmw_team_members")

8. **Schema fix** — `pmw_sub_items`
   - Change `uk_sub_items_main_code(main_item_key, item_code)` to `(main_item_key, item_code, deleted_flag, deleted_time)` in both MySQL and SQLite schema files

9. **Tests** — write test cases FIRST for each fix category

### Out-of-Scope

- GORM model tag alignment (P2, cosmetic — AutoMigrate doesn't modify existing indexes)
- Adding delete features to User/MainItem/ItemPool (future work)
- Frontend changes (backend fix only)

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `FindByID[T]` generic can't blindly add NotDeleted (ProgressRecord/StatusHistory lack deleted_flag) | High | High | Split generic constraint: `softDeletable` vs `nonSoftDeletable`, or override at repo level. Write compile-time tests to verify both paths. **Justification:** High likelihood because the generic helper is used by 6 entity types, 2 of which (ProgressRecord, StatusHistory) lack `deleted_flag` -- a naive `NotDeleted` addition will fail at compile time. High impact because `FindByID[T]` is the primary lookup path for every entity. |
| Changing generic helpers affects all callers -- unintended filtering of valid records | Medium | High | Write tests first for each entity type. After split, run targeted tests per entity to catch empty-result regressions. **Justification:** Medium likelihood because the constraint split makes the change compile-time safe, but a mislabeled constraint could silently filter valid records. High impact because any false filtering breaks core lookups (role permissions, team membership, item visibility). |
| SubItem SoftDelete change may break existing data with GORM-style soft-delete rows | Low | Medium | Query production for rows where GORM's `deleted_at` is set but `deleted_flag=0`. If found, write a one-time migration to sync flags. If zero rows, no migration needed. **Justification:** Low likelihood because SubItem soft-delete was recently added and production usage is minimal. Medium impact because mismatched rows would cause some deleted sub-items to reappear. |
| Unique index change on pmw_sub_items requires manual ALTER TABLE | Medium | Low | Document the ALTER SQL in migration README. Index change is additive (adds columns to existing index) -- no data loss. Test on SQLite and MySQL. **Justification:** Medium likelihood because ALTER TABLE on an active table requires coordination, but the change is additive. Low impact because it only affects the sub-item re-creation path, and the old index remains valid until migration runs. |

## Success Criteria

1. All repo List/Find methods for soft-deletable entities filter out deleted_flag=1
2. SubItem SoftDelete correctly sets deleted_flag=1 and deleted_time
3. Sub-items can be re-created with same code after soft-delete
4. `FindByID[T]` and `FindByIDs[T]` return all records (unfiltered by NotDeleted) when T is a non-soft-deletable type (ProgressRecord, StatusHistory) -- verified by explicit test cases for both entity types after the generic constraint split
5. New test cases cover: soft-deleted entity excluded from List, Find, and bulk lookups (enumerated per entity: Team, Role, SubItem, User, MainItem, ItemPool)
6. Role repo methods (`HasPermission`, `GetUserTeamPermissions`, `CountMembersByRoleID`) exclude soft-deleted team members from permission checks
