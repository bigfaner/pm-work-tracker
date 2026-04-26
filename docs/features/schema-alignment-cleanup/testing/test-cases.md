---
feature: "schema-alignment-cleanup"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-26"
---

# Test Cases: schema-alignment-cleanup

## Summary

| Type | Count |
|------|-------|
| UI   | 8     |
| API  | 16    |
| CLI  | 0     |
| **Total** | **24** |

---

## UI Test Cases

## TC-001: console.error replaced with toast in API client
- **Source**: Spec Round 2 Item 8
- **Type**: UI
- **Target**: ui/api-client
- **Test ID**: ui/api-client/console-error-replaced-with-toast
- **Pre-conditions**: `frontend/src/api/client.ts` contains `console.error` calls
- **Steps**:
  1. Run `grep -n 'console.error' frontend/src/api/client.ts`
  2. Run `npx vitest run`
- **Expected**: Grep returns zero results; all frontend tests pass
- **Priority**: P1

## TC-002: Test data Role.id is string type
- **Source**: Spec Round 2 Item 9
- **Type**: UI
- **Target**: ui/types
- **Test ID**: ui/types/test-data-role-id-is-string
- **Pre-conditions**: `types/index.test.ts` contains test data with `Role.id: 1` (number)
- **Steps**:
  1. Inspect test data in `types/index.test.ts`
  2. Run `npx vitest run types/index.test.ts`
- **Expected**: Role.id values in test data are string type (e.g., `"role-1"`); tests pass
- **Priority**: P1

## TC-003: Redundant String() wrappers removed from frontend
- **Source**: Spec Round 3 Item 16
- **Type**: UI
- **Target**: ui/components
- **Test ID**: ui/components/redundant-string-wrappers-removed
- **Pre-conditions**: 8 frontend files contain 22 redundant `.String()` calls
- **Steps**:
  1. Run `grep -rn '\.String()' frontend/src/`
  2. Verify only truly necessary `.String()` calls remain (e.g., enum conversions)
  3. Run `npx vitest run`
- **Expected**: Only necessary `.String()` calls remain; all frontend tests pass
- **Priority**: P1

## TC-004: PermissionData.teamPermissions uses string keys
- **Source**: Story 4 / AC-1, Spec Round 3 Item 17
- **Type**: UI
- **Target**: ui/permissions
- **Test ID**: ui/permissions/team-permissions-string-keys
- **Pre-conditions**: `PermissionData.teamPermissions` is `Record<number, string[]>`
- **Steps**:
  1. Run `grep -n "Record<number" frontend/src/types/index.ts`
  2. Verify `teamPermissions` key type is `string`
  3. Run `npx tsc --noEmit`
- **Expected**: Grep returns zero results; `teamPermissions` keys are `string`; no type errors
- **Priority**: P1

## TC-005: Form field assigneeId renamed to assigneeKey
- **Source**: Story 4 / AC-2, Spec Round 3 Item 18
- **Type**: UI
- **Target**: ui/dialogs
- **Test ID**: ui/dialogs/assignee-id-renamed-to-assignee-key
- **Pre-conditions**: Frontend dialog components use `assigneeId` form field name
- **Steps**:
  1. Run `grep -rn "assigneeId" frontend/src/`
  2. Run `npx vitest run`
- **Expected**: Grep returns zero results; all frontend tests pass
- **Priority**: P1

## TC-006: TableRow.mainItemId is string type
- **Source**: Spec Round 4 Item 23
- **Type**: UI
- **Target**: ui/types
- **Test ID**: ui/types/table-row-main-item-id-is-string
- **Pre-conditions**: `TableRow.mainItemId` is typed as `number | null` in `types/index.ts`
- **Steps**:
  1. Run `grep -n "mainItemId.*number" frontend/src/types/index.ts`
  2. Run `npx tsc --noEmit`
- **Expected**: Grep returns zero results; `mainItemId` is `string | null`; no type errors
- **Priority**: P1

## TC-007: Shared formatDate utility extracted
- **Source**: Spec Round 4 Item 24
- **Type**: UI
- **Target**: ui/utils
- **Test ID**: ui/utils/shared-format-date-extracted
- **Pre-conditions**: 3 team pages have duplicate `formatDate` implementations
- **Steps**:
  1. Grep for `formatDate` definitions across frontend
  2. Verify `formatDate` is defined only in a shared utils file
  3. Run `npx vitest run`
- **Expected**: `formatDate` defined only in shared utils; all frontend tests pass
- **Priority**: P2

## TC-008: Shared state recording helper extracted
- **Source**: Spec Round 3 Item 15
- **Type**: UI
- **Target**: ui/api-client
- **Test ID**: ui/api-client/shared-state-recording-helper-extracted
- **Pre-conditions**: 5 call sites have duplicate state recording logic
- **Steps**:
  1. Grep for state recording patterns across frontend
  2. Verify shared helper is used at all call sites
  3. Run `npx vitest run`
- **Expected**: State recording uses shared helper at all call sites; all frontend tests pass
- **Priority**: P2

---

## API Test Cases

## TC-009: Assign sub-item updates assignee_key column
- **Source**: Story 1 / AC-1, Spec Round 1 Item 1
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/assign-updates-assignee-key-column
- **Pre-conditions**: A sub-item exists with no assignee
- **Steps**:
  1. Send API request to assign a member to the sub-item
  2. Query the database row for the sub-item
  3. Inspect which column was updated
- **Expected**: The `assignee_key` column is updated (not `assignee_id`); `grep -rn "assignee_id" backend/internal/service/sub_item_service.go` returns zero results
- **Priority**: P0

## TC-010: Assignee persists after page refresh
- **Source**: Story 1 / AC-2
- **Type**: API
- **Target**: api/sub-items
- **Test ID**: api/sub-items/assignee-persists-after-refresh
- **Pre-conditions**: A sub-item has been assigned to a member via API
- **Steps**:
  1. Assign a member to a sub-item
  2. Fetch the sub-item via API
  3. Verify the assignee name is returned in the response
- **Expected**: The assignee name persists in the sub-item data after fetch
- **Priority**: P0

## TC-011: Filter by assignee returns correct subset
- **Source**: Story 2 / AC-1, Spec Round 1 Item 2
- **Type**: API
- **Target**: api/items
- **Test ID**: api/items/filter-by-assignee-returns-correct-subset
- **Pre-conditions**: Multiple items exist assigned to different members
- **Steps**:
  1. Send API request with `assignee_key` filter for a specific member
  2. Verify the response contains only items assigned to that member
- **Expected**: Only items assigned to the selected member are returned (not all items, not empty); `go test ./internal/handler/ -run TestFilter` passes
- **Priority**: P0

## TC-012: Filter by assignee consistent across MySQL and SQLite
- **Source**: Story 2 / AC-2
- **Type**: API
- **Target**: api/items
- **Test ID**: api/items/filter-by-assignee-consistent-across-dialects
- **Pre-conditions**: Same test data exists on both MySQL and SQLite instances
- **Steps**:
  1. Execute filter-by-assignee API call on MySQL backend
  2. Execute same API call on SQLite backend
  3. Compare response counts and item IDs
- **Expected**: Both backends return identical results for the same filter
- **Priority**: P0

## TC-013: Deprecated DTOs removed from item_dto.go
- **Source**: Story 2.5 / AC-1, Spec Round 2 Item 3
- **Type**: API
- **Target**: api/dto
- **Test ID**: api/dto/deprecated-dtos-removed
- **Pre-conditions**: `item_dto.go` contains 4 deprecated DTOs
- **Steps**:
  1. Run `grep -rn "Deprecated" backend/internal/dto/item_dto.go`
  2. Run `go build ./...`
- **Expected**: Grep returns zero results; build compiles without errors
- **Priority**: P1

## TC-014: Dead code in handler nil checks removed
- **Source**: Spec Round 2 Item 5
- **Type**: API
- **Target**: api/handlers
- **Test ID**: api/handlers/dead-nil-checks-removed
- **Pre-conditions**: `item_pool_handler.go` and `progress_handler.go` contain panic-on-nil followed by redundant nil checks
- **Steps**:
  1. Remove the redundant `if err != nil { return }` blocks
  2. Run `go build ./...`
  3. Run `go test ./internal/handler/`
- **Expected**: Build passes; handler tests pass
- **Priority**: P1

## TC-015: Redundant GORM column tags removed from role_repo
- **Source**: Spec Round 2 Item 7
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/redundant-gorm-tags-removed
- **Pre-conditions**: `role_repo.go` contains redundant `column:` GORM tags
- **Steps**:
  1. Run `grep -n 'column:' backend/internal/repository/gorm/role_repo.go`
  2. Run `go test ./internal/repository/gorm/ -run TestRole`
- **Expected**: Grep returns zero results; role repo tests pass
- **Priority**: P1

## TC-016: TransactionDB and dbTransactor merged into single interface
- **Source**: Story 3 / AC-1, Spec Round 3 Item 10
- **Type**: API
- **Target**: api/services
- **Test ID**: api/services/transaction-interfaces-merged
- **Pre-conditions**: Codebase has both `TransactionDB` and `dbTransactor` interfaces
- **Steps**:
  1. Run `grep -r "dbTransactor" backend/`
  2. Run `go build ./...`
  3. Run `go test ./internal/service/`
- **Expected**: Grep returns zero results; build and tests pass
- **Priority**: P1

## TC-017: Manual pagination replaced with dto.ApplyPaginationDefaults
- **Source**: Spec Round 3 Item 11
- **Type**: API
- **Target**: api/handlers
- **Test ID**: api/handlers/manual-pagination-replaced
- **Pre-conditions**: `admin_handler.go` and `view_handler.go` contain manual pagination logic
- **Steps**:
  1. Search for manual pagination patterns in `admin_handler.go` and `view_handler.go`
  2. Run `go test ./internal/handler/`
- **Expected**: No manual pagination logic remains; handler tests pass
- **Priority**: P1

## TC-018: teamToDTO returns typed struct instead of gin.H
- **Source**: Spec Round 3 Item 13
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/team-to-dto-returns-typed-struct
- **Pre-conditions**: `team_handler.go` uses `gin.H` for team DTO conversion
- **Steps**:
  1. Run `grep -n "gin.H" backend/internal/handler/team_handler.go`
  2. Run `go test ./internal/handler/ -run TestTeam`
- **Expected**: Grep returns zero results; team handler tests pass
- **Priority**: P1

## TC-019: Shared userToDTO defined only once
- **Source**: Spec Round 3 Item 14
- **Type**: API
- **Target**: api/handlers
- **Test ID**: api/handlers/shared-user-to-dto-defined-once
- **Pre-conditions**: `userToDTO` is defined in both `auth_handler.go` and `admin_service.go`
- **Steps**:
  1. Grep for `userToDTO` definitions across the backend
  2. Run `go test ./internal/handler/ ./internal/service/`
- **Expected**: `userToDTO` is defined exactly once in a shared file; all tests pass
- **Priority**: P1

## TC-020: Roles table renamed with pmw_ prefix
- **Source**: Story 5 / AC-1, Spec Round 4 Item 19
- **Type**: API
- **Target**: api/roles
- **Test ID**: api/roles/tables-renamed-with-pmw-prefix
- **Pre-conditions**: `Role.TableName()` returns `"roles"` and `RolePermission.TableName()` returns `"role_permissions"`
- **Steps**:
  1. Verify `Role.TableName()` returns `"pmw_roles"`
  2. Verify `RolePermission.TableName()` returns `"pmw_role_permissions"`
  3. Check both SQLite and MySQL schema files contain `pmw_roles` and `pmw_role_permissions`
  4. Run `go test ./internal/model/ -run TestRole`
- **Expected**: Both table names have `pmw_` prefix; both schema files updated; model tests pass
- **Priority**: P1

## TC-021: ViewService has single unified constructor
- **Source**: Spec Round 4 Item 20
- **Type**: API
- **Target**: api/views
- **Test ID**: api/views/view-service-single-constructor
- **Pre-conditions**: `view_service.go` has duplicate initialization logic
- **Steps**:
  1. Inspect `view_service.go` constructor signatures
  2. Run `go test ./internal/service/ -run TestView`
- **Expected**: Single unified constructor signature; no duplicate initialization logic; tests pass
- **Priority**: P1

## TC-022: NotDeleted scope used consistently across all repositories
- **Source**: Spec Round 4 Item 22
- **Type**: API
- **Target**: api/repositories
- **Test ID**: api/repositories/not-deleted-scope-consistent
- **Pre-conditions**: `team_repo.go` and `role_repo.go` use inline `deleted_flag = 0` instead of `NotDeleted` scope
- **Steps**:
  1. Run `grep -rn "deleted_flag.*=.*0" backend/internal/repository/`
  2. Run `go test ./internal/repository/gorm/`
- **Expected**: Grep returns zero results (all repositories use `NotDeleted` scope); repository tests pass
- **Priority**: P1

## TC-023: Dead assignments and nil-slice initializations removed
- **Source**: Spec Round 2 Items 4 and 6
- **Type**: API
- **Target**: api/services
- **Test ID**: api/services/dead-assignments-removed
- **Pre-conditions**: `team_service.go` contains `_ = team.PmKey` dead assignment and `= []string{}` nil-slice initialization
- **Steps**:
  1. Run `grep -n '_ = team.PmKey' backend/internal/service/team_service.go`
  2. Run `grep -n '= \[\]string{}' backend/internal/service/team_service.go`
  3. Run `go vet ./internal/service/`
  4. Run `go test ./internal/service/ -run TestTeam`
- **Expected**: Both greps return zero results; `go vet` has no warnings; tests pass
- **Priority**: P1

## TC-024: Shared resolveBizKey helper extracted
- **Source**: Spec Round 3 Item 12
- **Type**: API
- **Target**: api/handlers
- **Test ID**: api/handlers/shared-resolve-bizkey-extracted
- **Pre-conditions**: 7 handlers contain duplicate resolve/parse functions for bizKey
- **Steps**:
  1. Grep for `resolveBizKey` across `backend/internal/handler/`
  2. Verify it exists only in a shared helper file
  3. Run `go test ./internal/handler/`
- **Expected**: `resolveBizKey` defined only in shared helper file; handler tests pass
- **Priority**: P1

---

## CLI Test Cases

No CLI test cases -- this feature has no CLI component.

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Spec R2#8 | UI | ui/api-client | P1 |
| TC-002 | Spec R2#9 | UI | ui/types | P1 |
| TC-003 | Spec R3#16 | UI | ui/components | P1 |
| TC-004 | Story 4 / AC-1, Spec R3#17 | UI | ui/permissions | P1 |
| TC-005 | Story 4 / AC-2, Spec R3#18 | UI | ui/dialogs | P1 |
| TC-006 | Spec R4#23 | UI | ui/types | P1 |
| TC-007 | Spec R4#24 | UI | ui/utils | P2 |
| TC-008 | Spec R3#15 | UI | ui/api-client | P2 |
| TC-009 | Story 1 / AC-1, Spec R1#1 | API | api/sub-items | P0 |
| TC-010 | Story 1 / AC-2 | API | api/sub-items | P0 |
| TC-011 | Story 2 / AC-1, Spec R1#2 | API | api/items | P0 |
| TC-012 | Story 2 / AC-2 | API | api/items | P0 |
| TC-013 | Story 2.5 / AC-1, Spec R2#3 | API | api/dto | P1 |
| TC-014 | Spec R2#5 | API | api/handlers | P1 |
| TC-015 | Spec R2#7 | API | api/roles | P1 |
| TC-016 | Story 3 / AC-1, Spec R3#10 | API | api/services | P1 |
| TC-017 | Spec R3#11 | API | api/handlers | P1 |
| TC-018 | Spec R3#13 | API | api/teams | P1 |
| TC-019 | Spec R3#14 | API | api/handlers | P1 |
| TC-020 | Story 5 / AC-1, Spec R4#19 | API | api/roles | P1 |
| TC-021 | Spec R4#20 | API | api/views | P1 |
| TC-022 | Spec R4#22 | API | api/repositories | P1 |
| TC-023 | Spec R2#4, R2#6 | API | api/services | P1 |
| TC-024 | Spec R3#12 | API | api/handlers | P1 |
