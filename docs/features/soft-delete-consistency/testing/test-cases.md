---
feature: "soft-delete-consistency"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-27"
---

# Test Cases: soft-delete-consistency

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 21    |
| CLI  | 0     |
| **Total** | **21** |

---

## API Test Cases

### TC-001: Deleted role excluded from role list API response

- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/deleted-role-excluded-from-list
- **Pre-conditions**: A role exists in the database with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Seed a role record with `deleted_flag=1` in the database
  2. Call the role list API endpoint
  3. Inspect the response payload
- **Expected**: The deleted role does not appear in the API response list
- **Priority**: P0

### TC-002: Deleted role returns 404 when accessed by bizKey

- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/deleted-role-returns-404-on-bizkey
- **Pre-conditions**: A role exists with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Seed a role record with `deleted_flag=1`
  2. Request the role by its `bizKey` via the role detail API
- **Expected**: The API returns HTTP 404 (Not Found)
- **Priority**: P0

### TC-003: Soft-deleted sub-item disappears from sub-item list

- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/soft-deleted-subitem-excluded-from-list
- **Pre-conditions**: A sub-item exists under a main item with `item_code` "P001-01"
- **Route**: N/A
- **Steps**:
  1. Soft-delete the sub-item (set `deleted_flag=1`, `deleted_time` to current time)
  2. Call the sub-item list API for the parent main item
  3. Inspect the response
- **Expected**: The deleted sub-item does not appear in the response; `deleted_flag=1` and `deleted_time` are set on the record
- **Priority**: P0

### TC-004: Re-create sub-item with same item_code after soft-delete succeeds

- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/recreate-same-code-after-soft-delete
- **Pre-conditions**: Sub-item "P001-01" has been soft-deleted (`deleted_flag=1`)
- **Route**: N/A
- **Steps**:
  1. Create a new sub-item under the same main item with `item_code` "P001-01"
- **Expected**: Creation succeeds without unique constraint error; the new sub-item appears in the list with `deleted_flag=0`
- **Priority**: P0

### TC-005: FindByID returns NotFound for soft-deleted User

- **Source**: Story 3 / AC-1
- **Type**: API
- **Target**: api/generic-helpers
- **Test ID**: api/generic-helpers/findbyid-softdeleted-user-returns-notfound
- **Pre-conditions**: A User record exists with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Set a User record's `deleted_flag=1`
  2. Call `FindByID[User]` with that user's ID
- **Expected**: Returns `ErrNotFound` / GORM `ErrRecordNotFound`
- **Priority**: P0

### TC-006: FindByID returns record for non-soft-deletable ProgressRecord

- **Source**: Story 3 / AC-2
- **Type**: API
- **Target**: api/generic-helpers
- **Test ID**: api/generic-helpers/findbyid-nonsoftdeletable-returns-record
- **Pre-conditions**: A ProgressRecord exists in the database (no `deleted_flag` column)
- **Route**: N/A
- **Steps**:
  1. Call `FindByID[ProgressRecord]` with an existing ProgressRecord ID
- **Expected**: Returns the record without SQL error (no `deleted_flag` column referenced)
- **Priority**: P0

### TC-007: HasPermission returns false for soft-deleted team member

- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/permissions
- **Test ID**: api/permissions/haspermission-excludes-deleted-member
- **Pre-conditions**: A team member has been soft-deleted (`deleted_flag=1`) and their role includes a specific permission code
- **Route**: N/A
- **Steps**:
  1. Soft-delete a team member record (`deleted_flag=1`)
  2. Call `HasPermission(userID, permissionCode)`
- **Expected**: Returns `false` -- the deleted member's permissions are excluded
- **Priority**: P0

### TC-008: GetUserTeamPermissions excludes permissions from team where member is deleted

- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/permissions
- **Test ID**: api/permissions/getuserteampermissions-excludes-deleted-team
- **Pre-conditions**: User A is a member of Team T1 (role has `permission_x`) and Team T2 (role has `permission_y`). The TeamMember record for T1 has `deleted_flag=1`.
- **Route**: N/A
- **Steps**:
  1. Soft-delete User A's membership in Team T1
  2. Call `GetUserTeamPermissions(userID=A)`
- **Expected**: Result contains only T2's `permission_y`; T1's `permission_x` is absent
- **Priority**: P0

### TC-009: CountMembersByRoleID excludes soft-deleted members

- **Source**: Story 4 / AC-3
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/countmembers-excludes-single-deleted
- **Pre-conditions**: A role has 3 team members, 1 of which is soft-deleted (`deleted_flag=1`)
- **Route**: N/A
- **Steps**:
  1. Soft-delete one of the team member records under the role
  2. Call `CountMembersByRoleID(roleID)`
- **Expected**: Returns `2`, not `3`
- **Priority**: P1

### TC-010: CountMembersByRoleID with multiple soft-deleted members

- **Source**: Story 4 / AC-4
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/countmembers-excludes-multiple-deleted
- **Pre-conditions**: A role has 5 team members, 3 of which are soft-deleted (`deleted_flag=1`)
- **Route**: N/A
- **Steps**:
  1. Soft-delete three of the team member records under the role
  2. Call `CountMembersByRoleID(roleID)`
- **Expected**: Returns `2`
- **Priority**: P2

### TC-011: GetUserTeamPermissions returns empty map when user deleted from all teams

- **Source**: Story 4 / AC-5
- **Type**: API
- **Target**: api/permissions
- **Test ID**: api/permissions/getuserteampermissions-empty-when-all-deleted
- **Pre-conditions**: User A is a member of Team T1 and Team T2. Both TeamMember records are soft-deleted (`deleted_flag=1`).
- **Route**: N/A
- **Steps**:
  1. Soft-delete User A's memberships in both T1 and T2
  2. Call `GetUserTeamPermissions(userID=A)`
- **Expected**: Returns an empty map with no team permissions
- **Priority**: P1

### TC-012: User repo FindByBizKey excludes soft-deleted users

- **Source**: Spec Module 2 / FindByBizKey
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/findbybizkey-excludes-deleted
- **Pre-conditions**: A User record exists with `deleted_flag=1` and a known `bizKey`
- **Route**: N/A
- **Steps**:
  1. Soft-delete a user record
  2. Call `FindByBizKey(bizKey)`
- **Expected**: Returns `ErrNotFound`
- **Priority**: P0

### TC-013: User repo List excludes soft-deleted users

- **Source**: Spec Module 2 / List
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/list-excludes-deleted
- **Pre-conditions**: Database contains both active and soft-deleted users
- **Route**: N/A
- **Steps**:
  1. Ensure at least one user has `deleted_flag=1`
  2. Call `List()`
- **Expected**: Only active users (`deleted_flag=0`) appear in the result
- **Priority**: P0

### TC-014: User repo ListFiltered excludes soft-deleted users

- **Source**: Spec Module 2 / ListFiltered
- **Type**: API
- **Target**: api/users
- **Test ID**: api/users/listfiltered-excludes-deleted
- **Pre-conditions**: Database contains soft-deleted users matching the search term
- **Route**: N/A
- **Steps**:
  1. Soft-delete a user whose username matches a search term
  2. Call `ListFiltered(search, offset, limit)` with that search term
- **Expected**: The deleted user is excluded from results; total count reflects only active users
- **Priority**: P1

### TC-015: Team repo List excludes soft-deleted teams

- **Source**: Spec Module 3 / List
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/list-excludes-deleted
- **Pre-conditions**: Database contains teams with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Soft-delete a team record
  2. Call team `List()`
- **Expected**: Deleted teams are excluded from the result
- **Priority**: P0

### TC-016: Team repo ListMembers excludes soft-deleted users from member list

- **Source**: Spec Module 3 / ListMembers
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/listmembers-excludes-deleted-users
- **Pre-conditions**: A team has members, one of which is a soft-deleted user (`deleted_flag=1` in pmw_users)
- **Route**: N/A
- **Steps**:
  1. Soft-delete a user who is a member of a team
  2. Call `ListMembers(teamID)`
- **Expected**: The deleted user does not appear in the team member list
- **Priority**: P0

### TC-017: MainItem repo List excludes soft-deleted items

- **Source**: Spec Module 4 / List
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/list-excludes-deleted
- **Pre-conditions**: A team has main items, some with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Soft-delete one or more main items
  2. Call main item `List(teamID, filter, page)`
- **Expected**: Deleted main items are excluded; total count reflects only active items
- **Priority**: P0

### TC-018: MainItem repo FindByBizKeys excludes soft-deleted items

- **Source**: Spec Module 4 / FindByBizKeys
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/findbybizkeys-excludes-deleted
- **Pre-conditions**: Main items exist, some with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Include bizKeys of both active and soft-deleted main items
  2. Call `FindByBizKeys(bizKeys)`
- **Expected**: Only active main items are returned; soft-deleted items are absent from the result map
- **Priority**: P1

### TC-019: SubItem repo SoftDelete sets deleted_flag and deleted_time

- **Source**: Spec Module 5 / SoftDelete
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/softdelete-sets-flag-and-time
- **Pre-conditions**: An active sub-item exists
- **Route**: N/A
- **Steps**:
  1. Call `SoftDelete(id)` on the sub-item
  2. Query the database directly for the sub-item record
- **Expected**: `deleted_flag=1`, `deleted_time` is set to a non-zero timestamp; the record still exists (not hard-deleted)
- **Priority**: P0

### TC-020: ItemPool repo List excludes soft-deleted pools

- **Source**: Spec Module 6 / List
- **Type**: API
- **Target**: api/item-pools
- **Test ID**: api/item-pools/list-excludes-deleted
- **Pre-conditions**: A team has item pools, some with `deleted_flag=1`
- **Route**: N/A
- **Steps**:
  1. Soft-delete one or more item pools
  2. Call item pool `List(teamID, filter, page)`
- **Expected**: Deleted item pools are excluded from results; total reflects only active pools
- **Priority**: P0

### TC-021: Sub-item unique index allows re-creation after soft-delete

- **Source**: Spec Module 8 / Schema change
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/unique-index-allows-recreate-after-softdelete
- **Pre-conditions**: The unique index `uk_sub_items_main_code` includes `(main_item_key, item_code, deleted_flag, deleted_time)`
- **Route**: N/A
- **Steps**:
  1. Create a sub-item with `item_code` "P001-01" under a main item
  2. Soft-delete that sub-item
  3. Create a new sub-item with the same `item_code` "P001-01" under the same main item
  4. Attempt to create yet another sub-item with the same `item_code` (active duplicate)
- **Expected**: Step 3 succeeds without unique constraint violation. Step 4 fails with a unique constraint error (prevents duplicate active sub-items)
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | API | api/roles | P0 |
| TC-002 | Story 1 / AC-2 | API | api/roles | P0 |
| TC-003 | Story 2 / AC-1 | API | api/sub-items | P0 |
| TC-004 | Story 2 / AC-2 | API | api/sub-items | P0 |
| TC-005 | Story 3 / AC-1 | API | api/generic-helpers | P0 |
| TC-006 | Story 3 / AC-2 | API | api/generic-helpers | P0 |
| TC-007 | Story 4 / AC-1 | API | api/permissions | P0 |
| TC-008 | Story 4 / AC-2 | API | api/permissions | P0 |
| TC-009 | Story 4 / AC-3 | API | api/roles | P1 |
| TC-010 | Story 4 / AC-4 | API | api/roles | P2 |
| TC-011 | Story 4 / AC-5 | API | api/permissions | P1 |
| TC-012 | Spec Module 2 / FindByBizKey | API | api/users | P0 |
| TC-013 | Spec Module 2 / List | API | api/users | P0 |
| TC-014 | Spec Module 2 / ListFiltered | API | api/users | P1 |
| TC-015 | Spec Module 3 / List | API | api/teams | P0 |
| TC-016 | Spec Module 3 / ListMembers | API | api/teams | P0 |
| TC-017 | Spec Module 4 / List | API | api/main-items | P0 |
| TC-018 | Spec Module 4 / FindByBizKeys | API | api/main-items | P1 |
| TC-019 | Spec Module 5 / SoftDelete | API | api/sub-items | P0 |
| TC-020 | Spec Module 6 / List | API | api/item-pools | P0 |
| TC-021 | Spec Module 8 / Schema | API | api/sub-items | P0 |
