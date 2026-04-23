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
- **N+1 queries**: 9 locations where list endpoints make per-item DB calls — current p95 response time for `ItemPoolHandler.List` with ~200 records is 1.4s; `ProgressHandler.List` p95 is 1.1s with ~150 records; both exceed the 200ms target by 5x
- **In-memory filtering**: 4 service methods load all records then filter/paginate in Go, instead of pushing work to SQL
- **Code duplication**: ~600 lines of duplicated status dropdown code across 3 frontend files; identical pagination logic in 8 backend repositories; repeated filter functions, date formatters, and member name resolvers
- **Monolithic files**: `ItemViewPage.tsx` (1462 lines), `MainItemDetailPage.tsx` (928 lines) each contain 5+ inline components and 7+ dialog modals

This debt slows development velocity, makes bugs harder to find, and will cause performance degradation as data volume grows. **Why now**: the team is onboarding 2 new developers next sprint, and the current codebase's dead code, contract mismatches, and duplicated patterns would significantly extend their ramp-up time. Additionally, a recent user report flagged slow list-page loads on datasets exceeding 100 items, making performance remediation a stakeholder-visible priority.

## Proposed Solution

A four-phase progressive cleanup that proceeds from low-risk dead code removal to higher-impact structural refactoring. Each phase produces a shippable PR. Estimated total effort: 2-3 weeks for a single developer (or 1-2 weeks with two developers working backend/frontend in parallel). Phase estimates assume the existing codebase familiarity and no concurrent feature work.

**Phase dependencies and parallelization**: Phase 2 requires Phase 1 backend work complete (N+1 fixes touch the same service files as dead code removal). Phase 3 backend deduplication requires Phase 2 complete (shared helpers extract patterns established in Phase 2). Phase 3 frontend deduplication can begin after Phase 1 frontend work ships (it depends only on the contract-aligned types from Phase 1, not on Phase 2 backend changes). Phase 4 can begin as soon as Phase 3 for the same layer is complete. With two developers, the backend dev leads Phase 1 backend → Phase 2 → Phase 3 backend → Phase 4 backend, while the frontend dev starts Phase 1 frontend concurrently, then proceeds to Phase 3 frontend → Phase 4 frontend (the frontend dev has a 3-4 day gap while waiting for Phase 1 backend to complete before Phase 3 frontend can start).

**User-facing behavior:** End users will observe no visual or behavioral changes. All UI interactions, layouts, and navigation remain identical. The only observable difference will be faster list view page loads — the p95 response time target translates to sub-200ms API responses, which users on typical connections will experience as near-instantaneous data rendering on list pages compared to the current perceptible loading delay on larger datasets.

| Phase | Focus | Estimated Duration |
|-------|-------|--------------------|
| Phase 1 | Dead code & contract fixes | 2-3 days |
| Phase 2 | N+1 queries & performance | 3-5 days |
| Phase 3 | Shared components & deduplication | 3-4 days |
| Phase 4 | File splits & readability | 2-3 days |

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
- No `as any` type escapes remain in the frontend codebase

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
- List API p95 response time drops below 200ms for datasets with > 100 records (or achieves at least 40% reduction from baseline)

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
- No React anti-patterns remain (`useMemo` for side effects, data fetching outside React Query)
- Consistent constructor pattern across all backend handlers

## Alternatives Considered

| Approach | Effort | Coverage | Pros | Cons | Verdict |
|----------|--------|----------|------|------|---------|
| Progressive 4-phase cleanup | 2-3 weeks (1 dev) | ~100% of identified issues | Each phase is shippable independently; low-risk first; incremental value | Takes longer to complete all phases; some dependencies between phases | Selected |
| All-at-once cleanup | 2 weeks (1 dev) | ~100% | Single context switch; completes faster wall-clock | Very large diff; harder to review; higher regression risk | Rejected |
| Frontend-only cleanup | 1-1.5 weeks (1 dev) | ~40% (dead code removal, file splits, frontend dedup) | Immediate UI dev velocity improvement; lower regression surface | Leaves backend N+1 and dead code; contract mismatches remain; no performance improvement | Rejected |
| Backend-only cleanup | 1-1.5 weeks (1 dev) | ~50% (N+1 fixes, dead code, in-memory filtering) | Solves performance issues at the source; addresses user-reported slowness | Leaves frontend monolith and duplication untouched; new developers still face 1462-line files | Rejected |
| No cleanup (status quo) | 0 | 0% | Zero effort | Debt grows; performance degrades with scale; dev velocity decreases; onboarding friction | Rejected |

**Rationale**: The progressive approach is selected because it maximizes risk control while still achieving full coverage. Frontend-only and backend-only alternatives each address less than half the identified problems, leaving either the user-reported performance issue or the onboarding friction unresolved. The all-at-once approach achieves full coverage faster but concentrates regression risk in a single large diff — unacceptable given that this cleanup must not break existing behavior. The phased structure ensures that if work pauses after any phase, the completed work still delivers standalone value (Phase 1 reduces cognitive load, Phase 2 fixes performance, Phase 3 reduces duplication, Phase 4 improves readability).

## Scope

### In Scope
- Dead code removal (backend functions, frontend components/types/APIs)
- Frontend-backend contract alignment
- N+1 query elimination and batch query patterns
- In-memory filtering replaced with SQL-level filtering
- Shared component/hook/utility extraction
- Large file decomposition
- Type safety fixes: remove all `as any` casts, fix `user: any` to `user: User`

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
| SQL-level filtering changes introduce subtle query bugs | Medium | Medium | Write integration tests for each converted endpoint covering edge cases (empty results, null fields, multi-page results) before removing in-memory logic |
| Extracting shared components changes component behavior | Low | Medium | Run existing frontend tests after each extraction; manually verify each affected page's dropdown behavior before merge |
| Large file splits cause import path breakage | Low | Low | Run `tsc --noEmit` and `go vet ./...` before each commit; fix all import errors before opening PR |
| Phase dependencies cause rework | Medium | Medium | When Phase 3/4 refactor code that Phase 1 already touched, diff against the Phase 1 commit to verify the dead-code removal did not alter the surrounding logic being refactored. Tag Phase 1 PRs with `cleanup-phase-1` so rework scope is traceable |
| Partial completion — only Phase 1-2 ship before feature work resumes | Medium | Medium | Each phase produces a standalone shippable PR. If work pauses after Phase 2, the remaining phases can resume later without redoing Phase 1-2 work, since later phases add new structure rather than modifying Phase 1-2 output |
| Merge conflicts with concurrent feature branches | Medium | Medium | Keep each phase PR open for no more than 2 days; rebase onto `main` before merge. Coordinate with feature branch authors when touching shared files (e.g., `ItemViewPage.tsx`, `view_service.go`) |
| p95 200ms target unachievable without schema changes | Low | Medium | Measure baseline p95 before starting Phase 2. If the 40% reduction target is met but absolute p95 remains above 200ms due to schema constraints, document the bottleneck and propose a follow-up scope addition for the required schema change rather than blocking this cleanup |

## Success Criteria

- [ ] All dead code removed: zero unused exports, functions, components, or types remain
- [ ] Frontend types accurately reflect backend response shapes with no `as any` escapes anywhere in the codebase
- [ ] List API endpoints make O(1) association queries (not O(N))
- [ ] All in-memory filtered endpoints use SQL-level filtering instead
- [ ] Backend repository layer has zero duplicated boilerplate — defined as: identical repository method bodies (e.g., the `FindByID` pattern repeated across 8 repos, identical pagination offset/limit computation) verifiable by counting repos that duplicate the same 5+ line pattern
- [ ] No page component file exceeds 300 lines
- [ ] No duplicated component logic > 10 lines (duplication defined as: two or more code blocks in different files with identical structure and logic that differ only in variable names, verifiable by `jscpd` with a 10-line threshold)
- [ ] No React anti-patterns remain (`useMemo` for side effects, data fetching outside React Query)
- [ ] Consistent constructor pattern across all backend handlers
- [ ] All existing tests pass with no regressions
- [ ] List API p95 response time under 200ms for datasets with > 100 records (or at least 40% reduction from pre-cleanup baseline)

## Next Steps

- Proceed to `/write-prd` to formalize requirements per phase
