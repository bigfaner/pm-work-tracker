---
status: "completed"
started: "2026-04-28 00:49"
completed: "2026-04-28 00:51"
time_spent: "~2m"
---

# Task Record: 2.summary Phase 2 Summary

## Summary
## Tasks Completed
- 2.1: Added .Scopes(NotDeleted) to FindByUsername, List, and SearchAvailable methods in user_repo.go. Added 3 new tests verifying soft-deleted users are excluded from query results.
- 2.2: Added NotDeleted/NotDeletedTable scopes to 8 team_repo methods (List, ListFiltered, FindByBizKey, FindMember, ListMembers, CountMembers, FindPMMembers, FindTeamsByUserIDs). Added 4 new tests verifying soft-deleted teams and members are excluded.
- 2.3: Added .Scopes(NotDeleted) to 6 query methods in main_item_repo.go (FindByBizKey, List, CountByTeam, ListNonArchivedByTeam, FindByBizKeys, ListByTeamAndStatus). Added 6 new tests verifying soft-deleted main items are excluded.
- 2.4: Added .Scopes(NotDeleted) to all query methods in sub_item_repo (FindByBizKey, List, ListByMainItem, ListByTeam) and item_pool_repo (FindByBizKey, List). Added 6 new tests verifying soft-deleted records are excluded.
- 2.5: Added NotDeletedTable scope to HasPermission and GetUserTeamPermissions join queries, and NotDeleted scope to CountMembersByRoleID in role_repo.go. Added 3 new tests verifying soft-deleted members are excluded from permission checks and member counts.

## Key Decisions
- 2.1: Inserted .Scopes(NotDeleted) after .WithContext(ctx) and before .Where() in all methods, consistent with existing patterns
- 2.2: Used NotDeletedTable("pmw_users") for ListMembers and FindPMMembers which join pmw_users; used NotDeletedTable("pmw_teams") for FindTeamsByUserIDs which joins pmw_teams
- 2.3: CountByTeam test uses separate team to avoid DB state leakage from sibling sub-tests
- 2.4: Tests use direct db.Model().Update("deleted_flag", 1) to simulate soft-delete without calling SoftDelete method
- 2.5: Used NotDeletedTable("pmw_team_members") for HasPermission and GetUserTeamPermissions because they use raw Table() joins where deleted_flag is ambiguous

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|----------|
| user_repo (FindByUsername, List, SearchAvailable) | Added .Scopes(NotDeleted) | User queries in service layer |
| team_repo (List, ListFiltered, FindByBizKey, FindMember, ListMembers, CountMembers, FindPMMembers, FindTeamsByUserIDs) | Added .Scopes(NotDeleted) / NotDeletedTable | Team and member queries |
| main_item_repo (FindByBizKey, List, CountByTeam, ListNonArchivedByTeam, FindByBizKeys, ListByTeamAndStatus) | Added .Scopes(NotDeleted) | Main item queries |
| sub_item_repo (FindByBizKey, List, ListByMainItem, ListByTeam) | Added .Scopes(NotDeleted) | Sub-item queries |
| item_pool_repo (FindByBizKey, List) | Added .Scopes(NotDeleted) | Item pool queries |
| role_repo (HasPermission, GetUserTeamPermissions, CountMembersByRoleID) | Added .Scopes(NotDeletedTable) / NotDeleted | Permission checks and role member counts |

## Conventions Established
- All repo query methods consistently use .Scopes(NotDeleted) for single-table queries and .Scopes(NotDeletedTable("table")) for join queries where deleted_flag column is ambiguous
- Scope insertion point: after .WithContext(ctx) and Joins, before .Where() — consistent across all repos

## Deviations from Design
- None

## Changes

### Files Created
无

### Files Modified
无

### Key Decisions
- Summary-only task: aggregates phase 2 task records into structured cross-phase summary

## Test Results
- **Passed**: 0
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] All phase task records read
- [x] Summary follows exact 5-section template
- [x] Types & Interfaces Changed table is populated
- [x] Record created via /record-task with coverage: -1.0

## Notes
无
