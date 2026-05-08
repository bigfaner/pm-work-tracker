---
date: "2026-05-08"
doc_dir: "docs/features/unify-permission-checks/design/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 96/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │   19     │  20      │ ✅         │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   7/7    │          │            │
│    Dependencies listed       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │   19     │  20      │ ✅         │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   7/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │   14     │  15      │ ✅         │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   5/5    │          │            │
│    HTTP status codes mapped  │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │   15     │  15      │ ✅         │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │   19     │  20      │ ✅         │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   10     │  10      │ ✅         │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │   96     │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 18/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:75-77 Dependencies table | Only 1 dependency listed (`pkg/permissions`). Interface 1 (line 120) explicitly introduces a new `teamRepo` dependency: "**New dependency**: `roleService` needs `teamRepo` for the SuperAdmin path." This new dependency is discussed inline but not reflected in the Dependencies table, which should be the canonical source. Also missing: `userRepo`, `roleRepo` (both used in Interface 1 code), and `apperrors` (referenced in propagation strategy, line 204). | -1 pt (Dependencies listed) |
| tech-design.md:163 Interface 3 | The `skipRegressionCheck bool` parameter is described as "sourced from `sub_item:assign` permCode in the handler" but no handler-level wiring code is shown. A developer must infer the extraction mechanism -- what function call checks for `sub_item:assign` from the Gin context, what package provides it, and what the default value is when the permCode is absent. The other three interfaces include concrete BEFORE/AFTER Go code; this one omits the handler bridge. | -1 pt (Directly implementable) |
| api-handbook.md:88-92 Error Codes | Single generic row: "(existing) \| 403 \| Permission denied — now also applies to handler-level checks (removed assignee bypass)". The handbook covers 8 modified endpoints (3 sub_item/progress + 5 team) but provides one undifferentiated error code entry. While the tech-design Error Path Migration table explains that all authorization failures produce the same `ErrForbidden`/403 response, the API contract document should confirm this explicitly per endpoint group, including the error response body schema, so API consumers don't have to cross-reference the tech design. | -1 pt (HTTP status codes mapped) |
| tech-design.md:366-396 Scope Divergence + PRD Coverage Map | PRD scope item 5.1 #3 (line 57) specifies: "DB migration：删除 `users.is_super_admin` 列，将原 SuperAdmin 用户绑定到 superadmin 角色" with two sub-items: (a) drop column, (b) bind existing SuperAdmin users to superadmin role via team_members records. The Scope Divergence table addresses sub-item (a) but does not explicitly address sub-item (b). The design achieves the functional equivalent (tier-1 bypass makes team_members unnecessary for SuperAdmin), but this implicit coverage is never stated. The PRD Coverage Map line 394 says "Column kept in DB and backend model" but doesn't mention the team_members binding at all. | -1 pt (PRD AC coverage) |

---

## Attack Points

### Attack 1: Interface & Model Defs — Interface 3 Handler Wiring Is Prose-Only

**Where**: tech-design.md:163 — "The `skipRegressionCheck` flag is sourced from `sub_item:assign` permCode in the handler."

**Why it's weak**: All other interfaces (1, 2, 4) include concrete Go code showing the full BEFORE/AFTER change. Interface 3 provides the function signature change but describes the handler-level extraction of `sub_item:assign` in a single prose sentence. A developer implementing this must determine: (1) what function extracts a permission code from the Gin context (is it a middleware helper, a service call, or a direct context lookup?), (2) what package it lives in, (3) what the default value of `skipRegressionCheck` should be when the user lacks `sub_item:assign`. This is the only interface where the implementer must guess at runtime wiring.

**What must improve**: Add 2-3 lines of handler-level code showing the extraction, e.g.: `skipRegression := middleware.HasPermissionCode(c, "sub_item:assign")` or equivalent. State the default value (presumably `false` for users without the permCode).

### Attack 2: Architecture Clarity — Dependencies Table Incomplete Despite Inline Discussion

**Where**: tech-design.md:75-77 — Dependencies table has one row: `pkg/permissions | AllCodeStrings() | Existing`. tech-design.md:120 — "**New dependency**: `roleService` needs `teamRepo` for the SuperAdmin path."

**Why it's weak**: The Dependencies section should be the canonical, complete list of all dependencies the design introduces or changes. Instead, the new `teamRepo` dependency is buried in the Interface 1 section and the Dependencies table is stale. Additionally, the implementation code in Interface 1 uses `userRepo.FindByID`, `roleRepo.GetUserTeamPermissions`, and `teamRepo.ListTeamBizKeys` -- three repository interfaces -- plus the error propagation relies on `apperrors`. None of these are in the Dependencies table. A developer scanning only the Dependencies section would miss the `teamRepo` addition entirely.

**What must improve**: Add rows for `teamRepo` (New), `userRepo` (Existing), `roleRepo` (Existing), and `apperrors` (Existing) to the Dependencies table. Mark `teamRepo` as "New" since it's a new dependency for the `roleService` constructor.

### Attack 3: Error Handling — API Handbook Error Codes Table Is Undifferentiated

**Where**: api-handbook.md:88-92 — Error Codes table: "(existing) | 403 | Permission denied — now also applies to handler-level checks (removed assignee bypass)"

**Why it's weak**: The API handbook is the contract document for API consumers. It covers 8 modified endpoints across two distinct authorization changes (assignee check removal for 3 endpoints, PM BizKey substitution removal for 5 endpoints). A single generic 403 row tells API consumers nothing about what changed in the error surface. While the tech design's Error Path Migration table (lines 206-253) thoroughly documents the migration, the API handbook should stand on its own. At minimum, it should confirm: (1) the error response body schema for 403 is unchanged from the existing `ErrForbidden` response, (2) which endpoints now have different authorization failure triggers (e.g., "sub_item Update: previously 403 for non-assignee non-PM users; now 403 only for users lacking `sub_item:update` permCode").

**What must improve**: Replace the single generic row with per-endpoint-group entries. Add a note confirming the error response body format is identical to the existing 403 response. Cross-reference the tech-design Error Path Migration table for implementation details.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 2): Security — Mitigations Are Descriptions, Not Countermeasures | ✅ Addressed | Threat table (tech-design.md:354-358) now maps each threat to a specific verification mechanism: "RBAC integration test scenario 2 asserts: user with only `sub_item:view` receives 403 on update." Mitigations section (lines 360-364) maps 1:1 to threats with explicit test names and runner commands. |
| Attack 2 (iter 2): Testing Strategy — Test Tooling Incomplete for Implementation | ✅ Addressed | Per-Layer Test Plan (tech-design.md:322-331) now names `go test` as runner with specific paths, `testify/mock` + `testify/assert` as libraries, `net/http/httptest` with "in-memory SQLite (via GORM `:memory:` DSN)" for integration tests, `vitest` with `npx vitest run` commands for frontend. |
| Attack 3 (iter 2): Interface Definitions — `teamRepo` Dependency Untyped, `isPMRole` Helper Contradiction | ✅ Addressed | `teamRepo.ListTeamBizKeys` now has full typed signature at line 124: `ListTeamBizKeys(ctx context.Context) ([]int64, error)`. The `isPMRole` contradiction is resolved at line 151: "Inline the role-name comparison directly (`role.Name == "pm"`) rather than calling the removed `isPMRole()` helper" — definitive, single answer. |

---

## Verdict

- **Score**: 96/100
- **Target**: 90/100
- **Gap**: +6 points (target exceeded)
- **Breakdown-Readiness**: 19/20 — **can proceed to /breakdown-tasks** (exceeds 18/20 gate)
- **Action**: Target reached. Design is ready for `/breakdown-tasks`. Remaining weaknesses are minor and non-blocking.
