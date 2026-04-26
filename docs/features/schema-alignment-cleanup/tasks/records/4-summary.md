---
status: "completed"
started: "2026-04-26 22:19"
completed: "2026-04-26 22:21"
time_spent: "~2m"
---

# Task Record: 4.summary Phase 4 Summary

## Summary
## Tasks Completed
- 4.1: Renamed roles to pmw_roles and role_permissions to pmw_role_permissions across model TableName methods, both schema files, migration DDL in rbac.go, and test assertions
- 4.2: Merged NewViewService/NewViewServiceWithUserRepo into single NewViewService with variadic userRepo; merged itemPoolToVO+itemPoolsToVOs into buildItemPoolVOs and progressRecordToVO+progressRecordsToVOs into buildProgressRecordVOs with batch lookup logic
- 4.3: Replaced inline deleted_flag=0 with NotDeleted/NotDeletedTable scopes in team_repo and role_repo; changed TableRow.mainItemId type from number|null to string|null; added formatDateOnly utility to lib/format.ts replacing 3 local date wrapper functions

## Key Decisions
- 4.1: Updated migration DDL in rbac.go to use pmw_ prefix so fresh installs create correctly-named tables; updated tableExists checks to look for pmw_roles and pmw_role_permissions
- 4.2: Used variadic userRepo ...repository.UserRepo for optional parameter instead of separate constructor; single-item callers wrap item in 1-element slice to reuse batch path avoiding N+1; buildItemPoolVOs and buildProgressRecordVOs always use batch FindByIDs
- 4.3: Added NotDeletedTable(table) scope for multi-table queries where bare deleted_flag column is ambiguous; kept pmw_main_items.deleted_flag=0 inside ListAllTeams correlated subquery as raw SQL since GORM scope cannot apply inside SELECT clause; removed local formatDate wrappers entirely from RoleManagementPage using formatDateOnly directly

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|--------|
| Role.TableName | modified (returns pmw_roles) | migration, schema files |
| RolePermission.TableName | modified (returns pmw_role_permissions) | migration, schema files |
| NewViewService signature | modified (variadic userRepo) | view_service, handlers |
| buildItemPoolVOs | added (replaces itemPoolToVO + itemPoolsToVOs) | view_service |
| buildProgressRecordVOs | added (replaces progressRecordToVO + progressRecordsToVOs) | view_service |
| NotDeleted / NotDeletedTable | added (scopes in scopes.go) | team_repo, role_repo |
| TableRow.mainItemId | modified (number|null to string|null) | frontend types |
| formatDateOnly | added (lib/format.ts) | TeamDetailPage, TeamManagementPage, RoleManagementPage |

## Conventions Established
- 4.1: All new tables must use pmw_ prefix; existing tables updated when touched
- 4.2: Batch VO builders accept slices and use FindByIDs for batch DB lookups, never individual FindByID in a loop
- 4.3: Soft-delete filtering uses NotDeleted scope for single-table queries, NotDeletedTable(table) for multi-table joins

## Deviations from Design
- None

## Changes

### Files Created
无

### Files Modified
无

### Key Decisions
- Summary task - no code changes; aggregated decisions from 4.1, 4.2, 4.3

## Test Results
- **Passed**: 0
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] All phase task records read and analyzed
- [x] Summary follows the exact template with all 5 sections
- [x] Types & Interfaces table lists every changed type

## Notes
无
