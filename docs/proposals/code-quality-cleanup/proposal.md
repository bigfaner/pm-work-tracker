---
created: 2026-04-22
author: "faner"
status: Draft
---

# Proposal: Deep Code Quality Cleanup

## Problem

After multiple feature iterations, the codebase has accumulated significant technical debt across both backend and frontend:

- **Dead code**: 18+ unused functions, components, types, and API endpoints that increase cognitive load and maintenance burden
- **Contract mismatches**: Frontend types declare fields the backend never sends (`delayCount`, nested `pm` object), and backend response fields are never consumed by the UI (`proposerId`, `lesson`, `isPMCorrect`)
- **N+1 queries**: 9 locations where list endpoints make per-item DB calls, degrading API response time as data grows
- **In-memory filtering**: 4 service methods load all records then filter/paginate in Go, instead of pushing work to SQL
- **Code duplication**: ~600 lines of duplicated status dropdown code across 3 frontend files; identical pagination logic in 8 backend repositories; repeated filter functions, date formatters, and member name resolvers
- **Monolithic files**: `ItemViewPage.tsx` (1462 lines), `MainItemDetailPage.tsx` (928 lines) each contain 5+ inline components and 7+ dialog modals

This debt slows development velocity, makes bugs harder to find, and will cause performance degradation as data volume grows.

## Proposed Solution

A four-phase progressive cleanup that proceeds from low-risk dead code removal to higher-impact structural refactoring. Each phase produces a shippable PR.

### Phase 1: Remove Dead Code & Fix Contract Mismatches

**Backend dead code removal:**
- Delete unused functions: `parseItemIDAsUint` (sub_item_handler.go), `paginateUsers` (admin_handler.go), `derefStr` (admin_service.go), `Filename` (report/renderer.go)
- Delete entire unused file: `vo/team_vo.go` (`TeamMemberVO`, `NewTeamMemberVO`, `NewTeamMemberVOs` — never called)
- Delete empty file: `middleware/rbac.go` (only contains a comment)
- Remove unused middleware helpers: `GetCallerTeamRole` (team_scope.go), `GetUsername` (auth.go)
- Remove deprecated `WeeklyView` method and its exclusive helpers (`toSubItemWeekDTO` if only used by WeeklyView) from view_service.go
- Remove dead `codeMap` assignment in role_service.go `codesToItems`

**Frontend dead code removal:**
- Delete unused components: `AdminRoute.tsx`, `AdminPage.tsx`, `ItemFilters.tsx`
- Delete unused types: `WeeklyViewResp`, `WeeklyGroup`, `SubItemWithProgress`, `TeamMember` from types/index.ts
- Delete unused API functions: `archiveMainItemApi`, `assignSubItemApi`, `getItemPoolApi`, `correctCompletionApi`
- Remove unused `addToast` destructuring in ItemViewPage and MainItemDetailPage
- Remove unused exports: `ApiSuccessEnvelope`, `ApiErrorEnvelope` from client.ts

**Contract mismatch fixes:**
- Remove `delayCount` field from frontend `MainItem` and `SubItem` types (backend never sends it)
- Add `statusName` to backend `SubItemVO` for parity with `MainItemVO`
- Align `TeamMemberResp`: change frontend to use flat `role` string (matching backend), or extend backend to return `roleId`/`roleName`
- Align `AdminTeam.pm`: change frontend to use flat `pmDisplayName` (matching backend), or extend backend to return nested object
- Fix `changeMainItemStatusApi` frontend return type (currently declares `{ status }` but backend returns full `MainItemVO`)
- Fix `setAuth` redundant expression: `user?.isSuperAdmin ?? user?.isSuperAdmin` → `user?.isSuperAdmin`

**Success criteria:**
- No compile errors, all existing tests pass
- Frontend types accurately reflect backend response shapes
- No `as any` type escapes remain for fields that the backend actually returns

### Phase 2: Fix N+1 Queries & Performance

**Backend N+1 elimination:**
- `ItemPoolHandler.List`: batch-resolve submitter names and main item codes (currently 2 queries per item)
- `ProgressHandler.List`: batch-resolve author names (currently 1 query per record)
- `view_service.resolveAssigneeNames`: batch-fetch by `WHERE id IN (...)` instead of per-ID lookups
- `view_service.WeeklyComparison`: same batch pattern for assignee resolution

**Backend in-memory filtering → SQL:**
- `AdminService.ListUsers`: push search filter and pagination to SQL query
- `TeamService.SearchAvailableUsers`: use `NOT IN` subquery for exclusion, `LIKE` for search at DB level
- `view_service.TableView`: push filtering, sorting, and pagination to SQL
- `view_service.GanttView`: add status filter as SQL `WHERE` clause

**Backend minor performance:**
- `TeamService.GetTeamDetail`: use `COUNT(*)` for member count instead of loading full member list
- Add composite index on `sub_items(team_id, main_item_id)` if query analysis supports it
- Bound `linkageMuMap` growth (add cleanup or use `sync.Map`)

**Frontend performance:**
- Fix `useMemo` misuse in `MainItemDetailPage` (side effects inside `useMemo`) → change to `useEffect`
- Replace "fetch all pages" pattern in `ItemViewPage` detail view with server-side pagination
- Replace imperative sub-item fetching with React Query in `ItemViewPage`
- Fix `user` parameter type in `store/auth.ts`: `user: any` → `user: User`

**Success criteria:**
- List API endpoints make O(1) DB queries for association resolution (not O(N))
- In-memory filtered endpoints become SQL-level filtered
- No React anti-patterns (`useMemo` for side effects, fetching outside React Query)
- Measurable improvement in list API response time for datasets > 100 records

### Phase 3: Extract Shared Components & Eliminate Duplication

**Backend deduplication:**
- Extract generic `FindByID[T]` helper function for the identical 8-repo pattern
- Unify pagination logic: repos receive pre-computed offset/limit instead of re-applying `ApplyPaginationDefaults`
- Extract shared filter-application function for `MainItem`/`SubItem` (same Status, Priority, AssigneeID, IsKeyItem filters)
- Unify transaction interface: single `DBTransactor` interface shared by `item_pool_service` and `team_service`
- Extract `formatDatePtr` to shared utility (currently duplicated in `vo/item_vo.go` and `view_service.go`)
- Consolidate `containsIgnoreCase` / `derefStr` helpers
- Extract shared table data-fetching logic from `TableView` and `TableExportCSV`

**Frontend deduplication:**
- Extract `StatusTransitionDropdown` shared component (~600 lines eliminated across ItemViewPage, MainItemDetailPage, SubItemDetailPage)
- Extract `useMemberName` hook (currently copy-pasted in 3 pages)
- Unify `formatDate` utility (currently 4+ independent versions)
- Create `MemberSelect` component (repeated in 8+ dialogs across pages)
- Extract terminal status constants to `lib/status.ts`

**Success criteria:**
- No copy-pasted component logic > 10 lines
- Backend repository layer has zero duplicated boilerplate
- Shared utilities are in dedicated modules, not inline

### Phase 4: Split Monolithic Files & Improve Readability

**Backend readability:**
- Decompose `WeeklyComparison` (288 lines) into smaller functions: progress splitting, activity tracking, delta computation, stats aggregation
- Standardize handler constructors: all use panic-on-nil pattern, remove "stub" constructors
- Move inline anonymous request structs to `dto` package
- Unify role request types between `role_handler` and `role_service`
- Fix confusing `TeamMember.Role` `gorm:"-"` field (document or remove write-path usage)

**Frontend readability:**
- Split `ItemViewPage.tsx` (1462 lines) into: `ItemViewPage.tsx`, `ItemSummaryView.tsx`, `ItemDetailView.tsx`, `CreateMainItemDialog.tsx`, `EditMainItemDialog.tsx`, `CreateSubItemDialog.tsx`, `EditSubItemDialog.tsx`, `AppendProgressDialog.tsx`
- Split `MainItemDetailPage.tsx` (928 lines) similarly
- Fix all `as any` type escapes by extending TypeScript types to match actual API responses
- Group related `useState` in large pages or convert to `useReducer`

**Success criteria:**
- No page component file exceeds 300 lines
- No `as any` type escapes
- Consistent constructor pattern across all backend handlers

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Progressive 4-phase cleanup | Each phase is shippable independently; low-risk first; incremental value | Takes longer to complete all phases; some dependencies between phases | Selected |
| All-at-once cleanup | Faster to complete; single context switch | Very large diff; harder to review; higher regression risk | Rejected |
| Frontend-only cleanup | Smaller scope; immediate UI dev velocity improvement | Leaves backend N+1 and dead code; contract mismatches remain | Rejected |
| Backend-only cleanup | Solves performance issues at the source | Leaves frontend monolith and duplication untouched | Rejected |
| No cleanup (status quo) | Zero effort | Debt grows; performance degrades with scale; dev velocity decreases | Rejected |

## Scope

### In Scope
- Dead code removal (backend functions, frontend components/types/APIs)
- Frontend-backend contract alignment
- N+1 query elimination and batch query patterns
- In-memory filtering replaced with SQL-level filtering
- Shared component/hook/utility extraction
- Large file decomposition
- Type safety improvements (remove `as any`, fix `user: any`)

### Out of Scope
- UI/UX visual changes
- New features or new API endpoints
- Database schema changes (only adding indexes)
- Test coverage expansion (existing tests must pass, but no new test files)
- Third-party dependency upgrades
- API versioning or backwards-compatibility shims

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Contract mismatch fixes break frontend | Medium | High | Run all frontend + backend tests after Phase 1; verify API responses match updated types |
| SQL-level filtering changes introduce subtle query bugs | Medium | Medium | Keep existing service-level filter logic as fallback; test with edge cases (empty results, null fields) |
| Extracting shared components changes component behavior | Low | Medium | Visual regression test each page after Phase 3; compare rendered output |
| Large file splits cause import path breakage | Low | Low | TypeScript compiler catches all import errors; fix before commit |
| Phase dependencies cause rework | Low | Low | Phase 1 is prerequisite for Phase 3/4 (dead code must be removed before refactoring around it) |

## Success Criteria

- [ ] All dead code removed: zero unused exports, functions, components, or types remain
- [ ] Frontend types accurately reflect backend response shapes with no `as any` escapes
- [ ] List API endpoints make O(1) association queries (not O(N))
- [ ] No page component file exceeds 300 lines
- [ ] No duplicated component logic > 10 lines
- [ ] All existing tests pass with no regressions
- [ ] Measurable improvement in list API response time for datasets > 100 records

## Next Steps

- Proceed to `/write-prd` to formalize requirements per phase
