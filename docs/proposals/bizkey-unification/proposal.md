---
created: 2026-04-28
author: fanhuifeng
status: Draft
---

# Proposal: BizKey Unification — Eliminate uint/int64 ID Confusion Across Service Layer

## Problem

Internal auto-increment `uint` IDs and external snowflake `int64` bizKeys are silently mixed throughout the service layer, causing data mismatches that are invisible at compile time.

### Evidence

Concrete bugs found in the current codebase:

1. **`progress_service.go:57`** — `TeamKey: int64(teamID)` casts the internal `uint` auto-increment ID into the `TeamKey` field, which is supposed to hold a snowflake bizKey. The stored value is wrong.

2. **`team_service.go:175`** — `s.isPMRole(ctx, uint(roleID))` where `roleID` is parsed from `req.RoleKey` (a string bizKey) via `pkg.ParseID` returning `int64`. The cast `uint(roleID)` silently truncates the int64 snowflake value.

3. **`team_service.go:305`** — `isPMRole(ctx context.Context, bizKey uint)` — the parameter is named `bizKey` but typed `uint`. A snowflake ID is `int64`; this signature is a lie.

4. **`team_service.go:281`** — `UpdateMemberRole(ctx, pmID, teamID, targetUserID, roleID uint)` — `roleID` here is the internal auto-increment ID, but the handler resolves it from a bizKey string. The two representations are conflated.

5. **`middleware/team_scope.go:35,39,75`** — Middleware resolves `teamBizKey → team.ID (uint)` and injects `teamID uint` into context. Every downstream service then receives an internal ID, not a bizKey, making it impossible to distinguish the two at call sites.

### Urgency

- The `TeamKey` bug in `progress_service.go` means every progress record written to the database has a wrong `team_key` value. Queries filtering by `team_key` will silently return wrong results.
- The `uint(roleID)` cast will silently corrupt role lookups on any system where snowflake IDs exceed `uint` range (which they do on 64-bit systems when the high bit is set).
- Each new feature that touches team/user/role IDs has a ~50% chance of introducing the same class of bug because the type system provides no guidance.

## Proposed Solution

Make `int64` bizKey the **only** identifier that crosses layer boundaries above the repository. After this change, progress records will contain correct snowflake `team_key` values and role permission checks will resolve against the correct bizKey — eliminating the silent data corruption currently affecting every progress write. Specifically:

1. **Middleware** injects `teamBizKey int64` into context instead of `teamID uint`. `GetTeamID` is replaced by `GetTeamBizKey`.
2. **Service interfaces** replace all `teamID uint`, `userID uint`, `roleID uint` parameters with `int64` bizKey equivalents where those values originate from external input (URL params, request bodies).
3. **Internal-only lookups** (FK joins, repo-to-repo calls within a single transaction) continue to use `uint` IDs via `FindByID` — this stays inside the repository layer and never surfaces to services.
4. **`isPMRole`** is retyped to accept `int64` and renamed to make the contract explicit.
5. **Forbidden pattern enforced**: no `uint(bizKey)` or `int(bizKey)` casts anywhere in service or handler code. The compiler will catch violations once types are correct.

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Do nothing | Zero effort | `TeamKey` bug corrupts data silently; every new feature risks the same class of bug | Rejected: active data corruption |
| Targeted: fix only the known bugs (3–4 sites) | Small diff, low risk | Doesn't prevent recurrence; next developer will make the same mistake | Rejected: treats symptoms not cause |
| Full: service layer uses int64 bizKey (this proposal) | Compiler enforces correctness; eliminates entire class of bugs | Larger change, touches many files | Accepted |
| Remove FindByID entirely | Simplest contract | Breaks internal FK joins, hurts performance | Rejected: over-correction |

## Scope

### In Scope

- `middleware/team_scope.go`: inject `teamBizKey int64` instead of `teamID uint`; update `GetTeamID` → `GetTeamBizKey`
- Service interfaces (8 files: `item_pool_service.go`, `main_item_service.go`, `progress_service.go`, `report_service.go`, `role_service.go`, `sub_item_service.go`, `team_service.go`, `view_service.go`): replace `uint` ID params with `int64` bizKey where the value originates from external input
- `team_service.go`: fix `isPMRole` signature, fix `UpdateMemberRole` roleID type, fix `InviteMember` roleID cast
- `progress_service.go`: fix `TeamKey` assignment bug
- Handler call sites (7 files: `item_pool_handler.go`, `main_item_handler.go`, `progress_handler.go`, `report_handler.go`, `sub_item_handler.go`, `team_handler.go`, `view_handler.go`): pass bizKey from context/request instead of resolved uint IDs
- Unit and integration tests that mock or assert on these signatures (20 files: 10 handler tests, 7 service tests, `team_scope_test.go`, `views_reports_test.go`, `helpers.go`)
- `middleware/team_scope_test.go`: update `capturedTeamContext` and assertions

### Out of Scope

- Repository interface signatures (`FindByID`, `FindByBizKey`) — these stay as-is
- Frontend changes — API contract (JSON field names) is unchanged
- Database schema — no DDL changes
- Non-team entities (main items, sub items) unless they have the same uint/int64 confusion — addressed separately

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Missing a call site causes a compile error | High | Low — compile errors are caught before merge | Run `go build ./...` as gate; CI will catch it |
| Test mocks use hardcoded uint values that need updating | High | Low | Update mocks alongside interface changes; table-driven tests make this mechanical |
| `GetTeamBizKey` rename breaks handler code not covered by tests | Medium | Medium | Grep all usages of `GetTeamID` before starting; verify zero remaining references after |
| Integration tests rely on internal IDs in assertions | Low | Medium | Review integration test helpers; replace any `uint` team ID assertions with bizKey equivalents |

## Success Criteria

- [ ] `go build ./...` passes with zero errors after all changes
- [ ] `go test ./internal/service/... ./internal/middleware/... ./internal/handler/...` passes with zero failures
- [ ] Zero occurrences of `uint(.*bizKey)`, `int(.*bizKey)`, or `int64(teamID)` patterns in `internal/service/` and `internal/handler/` (verified by grep)
- [ ] `progress_service.go` `TeamKey` field is assigned from a snowflake `int64` source, not from a `uint` internal ID
- [ ] `isPMRole` (or its replacement) accepts `int64`, not `uint`
- [ ] `UpdateMemberRole` and `InviteMember` signatures use `int64` for all ID/key params that originate from external input (verified by grep on `team_service.go` interface definition)
- [ ] Zero occurrences of `uint` typed parameters named `*ID` or `*Key` in service interface definitions where the value originates from external input — confirmed by `grep -n "uint" internal/service/*.go` returning no service method signatures
- [ ] `GetTeamID` is removed; `GetTeamBizKey() int64` is the only team context accessor

## Next Steps

- Proceed to `/write-prd` to formalize requirements
