---
created: 2026-04-26
author: "faner"
status: Draft
---

# Proposal: Schema Alignment Post-Refactoring Cleanup

## Problem

The `jlc-schema-alignment` refactoring (30+ commits) introduced `bizKey`-based navigation, renamed columns, and migrated IDs from `int64` to `string` across the stack. While the feature works, the migration left behind:

- **2 silent bugs** that cause functionality to fail without errors (assign feature broken, filter-by-assignee returns nothing)
- **~15 code quality issues**: deprecated types not cleaned up, duplicate patterns introduced as workarounds during migration, naming inconsistencies between form fields and API fields
- **~8 architectural inconsistencies**: systemic `int64`/`uint` casting on `*Key` fields, duplicate interfaces, inconsistent soft-delete handling, two constructors for optional dependencies

**Why now**: The bugs are user-visible (assign and filter-by-person are broken). The accumulated cruft makes the codebase harder to maintain -- every developer must mentally track which layer uses which ID type and which `String()` wraps are necessary vs artifacts.

## Solution

Fix each issue individually with its own commit. Group issues into 4 rounds ordered by dependency:

### Round 1: Bug Fixes (P0)
1. Fix `SubItem.Assign()` column name: `assignee_id` → `assignee_key`
2. Fix `assignee_key` filter type mismatch: `filter_helpers.applyItemFilter` receives `*string` from DTO (`MainItemFilter.AssigneeKey *string`) but the DB column is `*int64` (`model.MainItem.AssigneeKey *int64`, `model.SubItem.AssigneeKey *int64`). The string bizKey is passed directly to `WHERE assignee_key = ?` without conversion, causing a type mismatch against the int64 column. Fix: parse the string bizKey to `int64` via `pkg.ParseID` before passing to the query in `filter_helpers.go`.

### Round 2: Dead Code Removal (P1, low risk)
3. Remove deprecated DTOs (`WeeklyViewResult`, `WeeklyGroupDTO`, `SubItemWeekDTO`, `SubItemSummaryDTO`) from `item_dto.go`
4. Remove dead `_ = team.PmKey` assignment in `team_service.go`
5. Remove dead nil-check after panic-on-nil constructors (`item_pool_handler.go`, `progress_handler.go`)
6. Remove dead nil-slice initialization in `team_service.go`
7. Remove redundant `column:` GORM tags in `role_repo.go` scan struct
8. Frontend: resolve TODO in `client.ts` -- wire up toast instead of console.error
9. Frontend: fix stale test data (`Role` with `id: 1` instead of `bizKey`)

### Round 3: Pattern Unification (P1, medium risk)
10. Consolidate duplicate `TransactionDB` / `dbTransactor` interfaces into one shared interface
11. Replace `admin_handler.parsePagination` and `view_handler.Table` manual pagination with `dto.ApplyPaginationDefaults`
12. Extract shared `resolveBizKey` helper from the 7 duplicate resolve/parse functions across handlers
13. Replace `team_handler.teamToDTO` untyped `gin.H` with typed VO struct
14. Extract shared `userToDTO` base conversion from `auth_handler.userToDTO` and `admin_service.modelToAdminUserDTO` -- both construct user DTOs from `*model.User` with identical field mapping (BizKey, Username, DisplayName, Email, Status, IsSuperAdmin)
15. Extract status-history recording helper from 5 duplicate call sites: `main_item_service` (lines 315, 391, 429), `sub_item_service` (line 176), `progress_service` (line 101) -- all construct `&model.StatusHistory{...}` and call `s.statusHistorySvc.Record(ctx, ...)` with identical structure
16. Frontend: remove 22 truly redundant `String()` wraps where the value is already `string` (e.g., `String(item.bizKey)` where `bizKey: string`, `String(params.id)` in MSW handlers where route params are always `string`, `String(x.assigneeKey)` after truthiness guard already narrows away `null`). Retain 13 necessary conversions where `number`→`string` is intentional (e.g., `String(inviteRoleId!)` for Radix Select `value`, `String(roleId)` for API `roleKey` param)
17. Frontend: align `PermissionData.teamPermissions` from `Record<number, ...>` to `Record<string, ...>` and update `usePermission`/`PermissionGuard` signatures
18. Frontend: rename form field `assigneeId` → `assigneeKey` to match API naming

### Round 4: Architecture Alignment (P2, higher risk)
19. Rename `roles` → `pmw_roles` and `role_permissions` → `pmw_role_permissions` to align with the `pmw_` prefix convention used by all other tables (`pmw_users`, `pmw_teams`, `pmw_main_items`, etc.): update `model/role.go` `TableName()` methods, both schema files, `role_test.go` assertions, and `role_test.go` raw SQL DDL
20. Consolidate `ViewService`'s two constructors into one with optional `userRepo`
21. Unify `itemPoolToVO`+`itemPoolsToVOs` and `progressRecordToVO`+`progressRecordsToVOs` into single batch functions
22. Apply `NotDeleted` scope consistently across all repositories: replace inlined `deleted_flag = 0` checks in `team_repo.go` (lines 37, 222, 225) and `role_repo.go` (lines 32, 44) with the `NotDeleted` scope from `scopes.go`
23. Frontend: align `TableRow.mainItemId` type from `number | null` to `string | null`
24. Frontend: extract shared `formatDate` from 3 duplicate local definitions in team pages

### User-Facing Impact

The 22 internal refactoring items (items 3-24) produce no user-visible behavioral change -- they restructure code without altering API contracts or UI behavior.

The 2 P0 bug fixes are user-visible:

| Issue | Before (broken) | After (fixed) |
|-------|-----------------|---------------|
| **Item 1**: `SubItem.Assign()` uses wrong column name | Clicking "Assign" on a sub-item sends the request successfully but the `assignee_key` column is never updated in the database. The sub-item remains unassigned after page refresh. | Assigning a sub-item persists the assignee. On reload, the assigned person's name appears on the sub-item. |
| **Item 2**: Filter-by-assignee passes string bizKey to int64 DB column | The "Filter by Person" dropdown on the item list sends the selected assignee bizKey as a string but the query compares it against an int64 column. On MySQL this returns zero results (type mismatch); on SQLite the implicit cast may mask the bug. | Filtering by person returns only items assigned to the selected person on both database engines. |

## Alternatives

### A. Do nothing (status quo)

| Aspect | Assessment |
|--------|------------|
| **Pros** | Zero risk of regression; zero engineering effort; no merge-conflict surface on the branch |
| **Cons** | P0 bugs remain broken (assign feature non-functional); each new feature PR must work around the `int64`/`uint` casting, duplicate interfaces, and deprecated DTOs; onboarding cost increases as new developers encounter two conflicting patterns |
| **Verdict** | Rejected. The assign bug is a regression from the schema-alignment migration -- leaving it broken means the migration is incomplete. |

### B. Batch changes by layer (all model changes, then all service changes, etc.)

| Aspect | Assessment |
|--------|------------|
| **Pros** | Fewer commits (4-6 instead of 24); each commit is a self-contained layer; atomic within that layer |
| **Cons** | Each commit touches 5-10 files across unrelated features (e.g., one commit changes both team and progress models); harder to review because the diff mixes unrelated domains; harder to bisect -- a regression introduced in the "all service changes" commit requires reading a 300-line diff instead of a 30-line one; rollback is all-or-nothing per layer |
| **Verdict** | Rejected. Per-issue commits produce smaller diffs that are faster to review and can be individually reverted without affecting unrelated fixes. |

### C. Per-issue commits, grouped by dependency round (chosen)

| Aspect | Assessment |
|--------|------------|
| **Pros** | Each commit addresses exactly one issue (avg. 1-3 files changed); `git bisect` pinpoints regressions to a single change; individual commits can be reverted without touching other fixes; the 4-round grouping ensures dependencies are resolved first (e.g., bug fixes before the code that references the buggy path) |
| **Cons** | 24 commits on the branch; if the branch lives long, merge conflicts accumulate; reviewers see many small PRs or one large PR with many commits |
| **Verdict** | Chosen. The trade-off favors reviewability and rollback precision over commit count. The 24 commits are ordered so each is independently testable. |

## Scope

### In-scope
- All 24 issues listed above
- Each fix committed individually
- Run related tests after each fix

### Out-of-scope
- Performance optimizations (N+1 queries, in-memory filtering) -- covered by existing `code-quality-cleanup` proposal
- Monolithic file splitting (ItemViewPage, MainItemDetailPage)
- Frontend dialog component consolidation (duplicated across item-view/ and main-item-detail/)
- New features or behavioral changes

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| P0 fixes reveal hidden dependents on the broken behavior | Low | Medium | Grep codebase for `assignee_id` references before fixing; review all callers of filter endpoint |
| P1 dead code removal breaks something that imports it | Low | Low | Each removal verified by `go build` + related tests |
| P2 architecture changes introduce regressions | Medium | Medium | Run full test suite after each P2 change; P2 changes are last so they can be batch-reverted |
| Frontend type changes cause runtime errors | Low | Medium | TypeScript compiler catches mismatches at build time |
| Item 17 `PermissionData.teamPermissions` key change breaks runtime lookups -- changing `Record<number, string[]>` to `Record<string, string[]>` at `types/index.ts:5` invalidates all call sites that index by numeric team ID | Medium | High | Grep all `teamPermissions[` and `hasPermission(` call sites; update each to pass string team bizKey instead of number; run `npx vitest run` to verify |
| Item 16 `String()` removal at `useItemViewPage.ts:118` changes filter semantics -- current code uses `String(item.assigneeKey)` which converts `null` to `"null"` literal for comparison against the filter dropdown value | Low | Medium | Verify the filter dropdown never emits `"null"` as a value; if it does, the comparison was already broken and this removal is safe. Add a unit test for the filter logic before removing |
| Merge conflicts accumulate on long-lived branch (24 commits) | Medium | Medium | Rebase frequently on `main`; each commit is small and touches different files, reducing conflict surface per rebase |
| Item 21 (`NotDeleted` scope) changes query semantics in `team_repo` and `role_repo` that currently inline `deleted_flag = 0` | Low | Medium | Verify `NotDeleted` scope produces identical SQL (`WHERE deleted_flag = 0`) before replacing; run existing integration tests for team and role repos |

## Success Criteria

### Round 1: Bug Fixes (items 1-2)
1. `SubItem.Assign()` writes to `assignee_key` column: unit test passes with `fields["assignee_key"]` assertion (fixes `sub_item_service.go:262` using wrong column name `"assignee_id"`)
2. Filter-by-assignee returns only matching items when `assigneeKey` query param is provided: handler test confirms filter value reaches `applyItemFilter` as `*string` without truncation; repository test confirms the string bizKey is converted to `int64` before the SQL `WHERE` clause
3. All existing tests pass after both fixes (`go test` + `npx vitest run`)

### Round 2: Dead Code Removal (items 3-9)
4. `grep -r "WeeklyViewResult\|WeeklyGroupDTO\|SubItemWeekDTO\|SubItemSummaryDTO" backend/` returns zero results
5. `go build ./...` succeeds after each removal
6. Frontend test data uses string `bizKey` for `Role` instead of numeric `id: 1`; `client.ts` shows toast notification on API error instead of `console.error`

### Round 3: Pattern Unification (items 10-18)
7. Single `TransactionDB` interface exists; `grep -r "dbTransactor" backend/` returns zero results
8. `admin_handler` and `view_handler.Table` both call `dto.ApplyPaginationDefaults`; no manual offset/page arithmetic in either handler
9. Single `resolveBizKey` helper exists; all handler `parseBizKey`/`pkg.ParseID` call sites for path-param extraction use it
10. `team_handler.teamToDTO` returns a typed struct, not `gin.H`
11. Single shared `userToDTO` function exists in `pkg/` or `handler/` utility; `grep -rn "func userToDTO\|func modelToAdminUserDTO" backend/` returns exactly one definition; both `auth_handler` and `admin_service` call it
12. Single shared `recordStatusChange` helper exists; `grep -rn "statusHistorySvc.Record" backend/internal/service/` shows all 5 call sites use the helper (no inline `&model.StatusHistory{}` construction at call sites)
13. Zero redundant `String()` wraps on already-string values in frontend: `String(xxx.bizKey)`, `String(params.xxxId)` in MSW, `String(x.assigneeKey)` after truthiness guards all removed; necessary `number`→`string` conversions (inviteRoleId, roleId for Radix Select/API) remain
14. `PermissionData.teamPermissions` typed as `Record<string, ...>` throughout; `usePermission` accepts `string` key, not `number`
15. Frontend form field name matches API field name: `grep -rn "assigneeId" frontend/src/` returns zero results (all renamed to `assigneeKey`)

### Round 4: Architecture Alignment (items 19-24)
16. `Role.TableName()` returns `"pmw_roles"`, `RolePermission.TableName()` returns `"pmw_role_permissions"`; both schema files and test assertions updated; `go test ./internal/model/ -run TestRole` passes
17. `ViewService` has single constructor `NewViewService(..., userRepo repository.UserRepo)`; second constructor removed
18. Handler-layer VO conversion uses unified batch functions (one for item_pool, one for progress) instead of separate single/batch pairs
19. `NotDeleted` scope applied consistently in all repositories that query `deleted_flag` tables: `grep -rn "deleted_flag = 0" backend/internal/repository/gorm/` returns zero results (all replaced by scope); `grep -rn "NotDeleted" backend/internal/repository/gorm/` returns all repo files that query soft-delete tables
20. `TableRow.mainItemId` typed as `string | null` in frontend; no `number` references in `TableRow`-related code
21. Single shared `formatDate` utility; `grep -rn "function formatDate\|const formatDate" frontend/src/` returns exactly one definition
22. All existing tests pass after Round 4 changes
