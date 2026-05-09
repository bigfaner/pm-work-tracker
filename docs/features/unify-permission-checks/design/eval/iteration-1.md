---
date: "2026-05-08"
doc_dir: "docs/features/unify-permission-checks/design/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 68/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │   16     │  20      │ ⚠️         │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   6/7    │          │            │
│    Dependencies listed       │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │   16     │  20      │ ⚠️         │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   5/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │    8     │  15      │ ⚠️         │
│    Error types defined       │   3/5    │          │            │
│    Propagation strategy clear│   3/5    │          │            │
│    HTTP status codes mapped  │   2/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │   13     │  15      │ ✅         │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │   10     │  20      │ ❌         │
│    Components enumerable     │   5/7    │          │            │
│    Tasks derivable           │   2/7    │          │            │
│    PRD AC coverage           │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │    5     │  10      │ ⚠️         │
│    Threat model present      │   3/5    │          │            │
│    Mitigations concrete      │   2/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │   68     │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 18/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:77 Dependencies | Only 1 dependency listed (`pkg/permissions`). Missing: test framework imports, `roleRepo` interface, `teamRepo` interface (newly introduced), `userRepo`, `apperrors` package, frontend `hasPermission` module. The design introduces a new `teamRepo` dependency (Interface 1) but does not list it in the Dependencies table. | -3 pts (Dependencies) |
| tech-design.md:13 Overview | Cross-section inconsistency: PRD states "delete `User.IsSuperAdmin` field" and "remove `is_super_admin` column" (prd-spec.md line 23, scope item 1) but the design explicitly states the column is "kept as an internal optimization". This is a scope divergence between PRD and design. | -3 pts (Breakdown-Readiness: PRD AC coverage) |
| tech-design.md:99-117 Interface 1 | `GetUserPermissions` AFTER code does not compile -- `teamRepo.ListTeamBizKeys(ctx)` returns `[]int64` presumably but the return type is not typed. The `map[int64][]string` key type is `int64` but `ListTeamBizKeys` return type is unspecified. Developer must guess. | -1 pt (Interface signatures typed) |
| tech-design.md:119 Interface 1 note | "**New dependency**: `roleService` needs `teamRepo` (or a `ListTeamBizKeys` method)" -- the "or a" phrasing leaves two options open. A developer cannot implement without resolving this ambiguity first. | -1 pt (Directly implementable) |
| tech-design.md:144-148 Interface 2 note | "Remove `isPMRole()` helper (no longer needed for PM identity verification)" contradicts line 148: "`InviteMember`: keep `isPMRole` check for preventing direct PM role assignment". The helper is both "removed" and "kept" without specifying which callsites keep it vs remove it. | -1 pt (Models concrete) |
| tech-design.md:157-159 Interface 3 | `Append` signature has 8 parameters with no struct wrapper. The `skipRegressionCheck bool` is the last positional parameter. The design does not specify what happens when handler does NOT have `sub_item:assign` -- the boolean defaults to `false` implicitly but this is never stated. | -1 pt (Interface signatures typed) |
| tech-design.md:186-197 Error Handling | "No new error types. Existing errors preserved." -- but the design removes assignee checks and PM identity checks, which changes which errors are triggered and when. The table only lists 3 error codes. What error does `RequirePermission` return when denied? Is it `ErrForbidden`? Not stated. | -2 pts (Error types defined) |
| tech-design.md:200 Propagation | "Unchanged. Errors flow from repository -> service -> handler -> response via `apperrors`." -- prose-only, no diagram or concrete mapping of how the newly-simplified handler code surfaces errors differently. The design removes handler-level error paths (assignee check errors) without specifying what replaces them. | -2 pts (Propagation strategy clear) |
| api-handbook.md:90-92 Error Codes | "No new error codes" and only references "(existing)" and 403. The API handbook should map all error scenarios for the modified endpoints but only mentions 403 generically. Missing: what happens when SuperAdmin user queries `/me/permissions` and no teams exist? What does `ListTeamBizKeys` return? | -3 pts (HTTP status codes mapped) |
| tech-design.md:265-279 PRD Coverage Map | PRD scope item "Remove `RequirePermission` middleware's SuperAdmin short-circuit" (prd-spec.md line 51) is NOT addressed in the design. The design explicitly keeps tier-1 SuperAdmin bypass in `RequirePermission`. This is a major PRD-design divergence. | -3 pts (PRD AC coverage) |
| tech-design.md:20-33 Layer table | Layer table lists 13 layers with change types but no "Components enumerable" count. A task breakdown needs to know exact file counts. E.g., "Frontend components" says "Remove all `isSuperAdmin` references" but which 25 files? PRD says 25 files, design does not enumerate them. | -2 pts (Components enumerable) |
| tech-design.md:123-141 Interface 2 | Simplified signatures are shown as BEFORE/AFTER Go code, but no handler-level wiring is specified. How does the handler call `UpdateTeam(ctx, teamBizKey, req)` when the handler receives `(c *gin.Context)` and needs to extract `teamBizKey` from URL params? The mapping from HTTP input to service call is missing. | -2 pts (Tasks derivable) |
| tech-design.md:217-229 Testing | Frontend test tooling mentions "vitest + testing-library" but the project's actual test runner is `npx vitest run` per CLAUDE.md. The backend test tool says "testify/mock" but does not name the test runner (`go test`). Missing: the RBAC integration test approach says "httptest" but does not specify whether these are in-memory or against a real DB (critical given the project convention of integration tests hitting real DB per memory feedback). | -2 pts (Test tooling named) |
| tech-design.md:254-262 Security | Threat "Custom role gains unintended access via assignee check removal" is marked "Low" risk with mitigation "Removing assignee check only broadens access for users already holding the permission code -- which is the intended fix". This is not a mitigation; it is restating the change. A real mitigation would specify how to verify no unintended permission escalation occurs (e.g., audit log, permission matrix review). | -2 pts (Mitigations concrete) |
| tech-design.md:260-262 Mitigations | "seedPresetRoles is additive" and "API response sanitized" are properties, not countermeasures paired to specific threats. The threat model has 3 rows; the mitigations section has 3 bullets but they do not map 1:1 to threats. | -1 pt (Mitigations concrete) |

---

## Attack Points

### Attack 1: Breakdown-Readiness -- PRD-Design Scope Divergence Blocks Task Derivation

**Where**: PRD prd-spec.md scope items state "delete `User.IsSuperAdmin` field" and "remove `is_super_admin` column" and "remove `RequirePermission` middleware's SuperAdmin short-circuit". The design (tech-design.md lines 11, 24-25) says: "The backend `is_super_admin` column and middleware-level SuperAdmin bypass are **kept** as an internal optimization."

**Why it's weak**: This is a fundamental scope conflict. The PRD defines the "what" -- full removal of `is_super_admin`. The design re-scopes to keep it, which means multiple PRD scope items (migration to drop column, `RequirePermission` short-circuit removal, `middleware/auth.go` changes) have no corresponding design task. A developer running `/breakdown-tasks` cannot reconcile which PRD items to implement vs skip. The PRD Coverage Map (tech-design.md:271) claims "AC3a-c: SuperAdmin all ops" maps to "Middleware tier-1 bypass (kept)" -- but the PRD explicitly requires *removing* that bypass, not keeping it.

**What must improve**: Either (a) update the PRD to formally re-scope and remove the items the design chooses not to implement, with a documented decision trace, or (b) update the design to match the PRD's full scope. Without this reconciliation, task breakdown will produce ambiguous tasks.

### Attack 2: Breakdown-Readiness -- Frontend Change Scope is Not Enumerable

**Where**: tech-design.md line 33 states "Frontend components | Remove all `isSuperAdmin` references". The PRD (prd-spec.md line 181) says "25 files". The design provides no file list.

**Why it's weak**: A task breakdown requires knowing exactly which files change. "Remove all isSuperAdmin references" is a grep-and-hope instruction, not a design. The design should enumerate the 25 files (or at minimum the modules/directories) so that tasks can be scoped, estimated, and parallelized. Without this, `/breakdown-tasks` can only produce one monolithic "remove isSuperAdmin from frontend" task, which is not decomposable into reviewable units.

**What must improve**: Add a file enumeration table for frontend changes, grouped by module (types, store, components, mocks). At minimum list the directories and file patterns so tasks can be split.

### Attack 3: Error Handling -- Missing Error Paths for Modified Handlers

**Where**: tech-design.md lines 186-200 state "No new error types. Existing errors preserved." and lists only 3 error codes: `ErrForbidden`, `ErrNotTeamMember`, `ErrProgressRegression`. The api-handbook.md (line 91) only mentions "(existing) | 403".

**Why it's weak**: The design removes handler-level assignee checks and PM identity checks. Previously, these checks returned specific error responses (assignee ownership failure, PM identity failure). The design does not specify what replaces these error paths. When a non-PM user calls `TransferPM` without `team:transfer` permission, does `RequirePermission` return `ErrForbidden` with a specific message? The handler no longer checks -- the middleware does. But the middleware error message is generic ("permission denied") vs the old handler-specific message ("not team PM"). This is a user-facing behavior change that is undocumented. Furthermore, the design introduces a new code path: `ListTeamBizKeys` returning no teams for SuperAdmin -- what error does `GetUserPermissions` return? `nil` with empty map? Not specified.

**What must improve**: Map every removed handler-level error check to its replacement middleware/service error response. Specify error messages and HTTP response bodies for all modified endpoints. Add the `ListTeamBizKeys` empty-result edge case to error handling.

---

## Previous Issues Check

<!-- Only for iteration > 1 -->

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| N/A (first iteration) | N/A | N/A |

---

## Verdict

- **Score**: 68/100
- **Target**: 90/100
- **Gap**: 22 points
- **Breakdown-Readiness**: 10/20 -- **cannot proceed to /breakdown-tasks** (below 18/20 gate)
- **Action**: Continue to iteration 2. Priority fixes: (1) Reconcile PRD-design scope divergence on `is_super_admin` column removal, (2) Enumerate frontend files for task decomposition, (3) Map all error path changes for modified handlers.
