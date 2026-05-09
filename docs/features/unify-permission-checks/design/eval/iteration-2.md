---
date: "2026-05-08"
doc_dir: "docs/features/unify-permission-checks/design/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 86/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │   17     │  20      │ ⚠️         │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   6/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │   17     │  20      │ ⚠️         │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   6/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │   13     │  15      │ ✅         │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   4/5    │          │            │
│    HTTP status codes mapped  │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │   13     │  15      │ ✅         │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │   18     │  20      │ ✅         │
│    Components enumerable     │   6/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   8      │  10      │ ⚠️         │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │   86     │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 18/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:75-77 Dependencies | Only 1 dependency listed (`pkg/permissions`). The design introduces a new dependency: `teamRepo` (Interface 1, line 120: "roleService needs teamRepo"). This new dependency is not listed in the Dependencies table. Also missing: `apperrors` package (referenced in propagation strategy), `userRepo`, `roleRepo` (all used in Interface 1 code). | -2 pts (Dependencies) |
| tech-design.md:108 `teamRepo.ListTeamBizKeys(ctx)` | Return type is not specified. The AFTER code assigns `teams := teamRepo.ListTeamBizKeys(ctx)` and then iterates `for _, bizKey := range teams` assigning to `map[int64][]string` keys. Developer must infer the return type is `[]int64`. The interface signature is incomplete. | -1 pt (Interface signatures typed) |
| tech-design.md:144-148 Interface 2 notes | `isPMRole()` helper: line 145 says "Remove `isPMRole()` helper (no longer needed)" but line 147 says "`InviteMember`: keep `isPMRole` check for preventing direct PM role assignment". The note contradicts itself within 3 lines. A developer needs to know which specific callsites keep vs remove this helper. | -1 pt (Models concrete) |
| tech-design.md:157-159 Interface 3 | `Append` has 9 positional parameters including `skipRegressionCheck bool` as the last one. The design says "sourced from `sub_item:assign` permCode in the handler" but does not show the handler-level wiring code. Developer must infer how to extract this from the context and determine default behavior when the handler does NOT have `sub_item:assign`. | -1 pt (Directly implementable) |
| tech-design.md:200 Propagation Strategy | "Unchanged. Errors flow from repository -> service -> handler -> response via `apperrors`." The Error Path Migration table (lines 204-248) documents removed checks well, but the propagation strategy statement is still prose-only with no concrete mapping of how the shifted error origin (service -> middleware) changes the error propagation path. Middleware errors go through `c.AbortWithStatusJSON()` directly, not through the repository->service->handler chain described. | -1 pt (Propagation strategy clear) |
| api-handbook.md:88-92 Error Codes | Only "(existing) \| 403" is listed. The API handbook should map all error codes for the modified endpoints but provides a single generic row. Missing: specific error codes/messages for each modified endpoint (e.g., what exact error body does `RequirePermission` return on denial? What does the client see differently?). | -1 pt (HTTP status codes mapped) |
| tech-design.md:316-345 Testing Strategy | Test tooling column lists "testify/mock" for backend and "vitest + testing-library" for frontend. Missing: the backend test runner (`go test`) is not named. The RBAC integration tests say "httptest" but do not clarify whether they run against a real DB (critical given project convention of integration tests hitting real DB per project memory). | -2 pts (Test tooling named) |
| tech-design.md:362-393 PRD Coverage Map | PRD scope items for "Delete `User.IsSuperAdmin` field", "Remove `RequirePermission` short-circuit", and "Remove `AuthMiddleware` isSuperAdmin context" are marked "Intentionally kept" with a Scope Divergence rationale. This is an explicit documented divergence, not a gap. However, the PRD still formally lists these as in-scope items. Until the PRD is updated to reflect the re-scoped decision, the coverage is incomplete from the PRD's perspective. | -1 pt (PRD AC coverage) |
| tech-design.md:352 Threat table row 1 | Mitigation for "Custom role gains unintended access" still says "Removing assignee check only broadens access for users already holding the permission code -- which is the intended fix". This restates the change rather than providing a concrete verification mechanism (e.g., "integration test matrix validates no permission escalation beyond granted codes"). | -1 pt (Mitigations concrete) |
| tech-design.md:356-360 Mitigations section | Three bullet mitigations are listed but they do not map 1:1 to the three threats in the table above. "seedPresetRoles is additive" addresses the SuperAdmin DB column threat. "API response sanitized" addresses the frontend UI threat. "No new attack surface" is a general statement not paired to a specific threat. The mapping is implicit. | -1 pt (Mitigations concrete) |
| tech-design.md:267-311 Frontend enumeration | Frontend files are enumerated, but the "Types & Store" table (line 271) says `types/index.ts` has 4 refs at "lines 4, 70, 433, 494" -- line numbers in source files drift and become stale. The design should identify the 4 interfaces by name rather than by line number for durability. | -1 pt (Components enumerable) |

---

## Attack Points

### Attack 1: Security — Mitigations Are Descriptions, Not Countermeasures

**Where**: tech-design.md:350-360. The threat model identifies "Custom role gains unintended access via assignee check removal" and the mitigation is "Removing assignee check only broadens access for users already holding the permission code -- which is the intended fix". The three bullet mitigations (lines 356-360) are general properties of the design ("no new attack surface", "seedPresetRoles is additive", "API response sanitized"), not specific countermeasures.

**Why it's weak**: A mitigation must describe an active verification or prevention mechanism, not restate the design decision. A proper mitigation would be: "Integration test scenario 2 validates that a user with only `sub_item:view` cannot update a sub-item (403), confirming no permission escalation beyond granted codes." The current text reads as "we removed a check, and that's fine" -- which may be correct, but it is not a security mitigation. The mitigations also do not map 1:1 to threats: the threat model has 3 rows, the mitigations section has 3 bullets, but the correspondence is implicit and "No new attack surface" does not pair with any specific threat.

**What must improve**: Replace each mitigation with a concrete verification mechanism (test, audit, or design constraint). Map each mitigation explicitly to a threat (e.g., "Threat 1 -> Mitigation: integration test matrix validates permission boundaries"). Add a statement about how the RBAC integration tests in the Testing Strategy serve as the security regression gate.

### Attack 2: Testing Strategy — Test Tooling Incomplete for Implementation

**Where**: tech-design.md:316-345. The Per-Layer Test Plan table lists "testify/mock" and "httptest" for backend but does not name `go test` as the runner. The RBAC integration tests say "httptest" but do not specify whether they are in-memory or against a real database. The project convention (per CLAUDE.md) says `cd backend && go test ./...` and the project memory notes that integration tests should hit a real DB.

**Why it's weak**: A developer reading this table cannot determine how to run the tests. The RBAC integration tests are the most critical test layer (they validate the entire permission matrix), and the design does not clarify whether `httptest` means "mock HTTP server" or "real DB + HTTP test server". This distinction determines whether the tests catch real permission-check regressions or only mock-level issues. The frontend side is slightly better (names "vitest" and "testing-library") but still omits the project's actual runner command (`npx vitest run` per CLAUDE.md).

**What must improve**: Add `go test` as the backend test runner. Specify whether RBAC integration tests use an in-memory DB, a test SQLite file, or a real MySQL instance. Name the specific testify packages (`testify/assert`, `testify/mock`, `testify/suite` if applicable). Reference the project's test runner conventions from CLAUDE.md.

### Attack 3: Interface Definitions — `teamRepo` Dependency Untyped, `isPMRole` Helper Contradiction

**Where**: tech-design.md:108 (`teamRepo.ListTeamBizKeys(ctx)` has no return type) and tech-design.md:144-148 (`isPMRole()` is both "removed" and "kept" within 3 lines).

**Why it's weak**: Interface 1 introduces a new `teamRepo` dependency with an untyped `ListTeamBizKeys` method. The developer must infer the return type from the surrounding code (`for _, bizKey := range teams` assigned to `map[int64][]string` keys means `[]int64`). This is exactly the kind of ambiguity the "Directly implementable" criterion should catch. For `isPMRole`, lines 145-147 state "Remove `isPMRole()` helper" then "keep `isPMRole` check for preventing direct PM role assignment". A developer needs a definitive answer: is the function deleted entirely with the check inlined at `InviteMember`, or is it kept as a local helper only called from one place?

**What must improve**: Add the full typed signature for `teamRepo.ListTeamBizKeys(ctx context.Context) ([]int64, error)`. Resolve the `isPMRole` contradiction: either state "remove `isPMRole()` and inline the role-name check at `InviteMember`" or "keep `isPMRole()` as a private helper, called only from `InviteMember`".

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 1): PRD-Design Scope Divergence Blocks Task Derivation | ✅ Addressed | New "Scope Divergence from PRD" section (tech-design.md:362-373) explicitly documents all 3 divergent scope items with rationale. PRD Coverage Map (lines 374-393) marks them as "Intentionally kept" with cross-references. Task breakdown section (line 372) states "No migration task to drop the column, no task to remove the middleware short-circuit." |
| Attack 2 (iter 1): Frontend Change Scope is Not Enumerable | ✅ Addressed | New "Frontend Change File Enumeration" section (tech-design.md:263-311) lists every file with module grouping, source/test classification, reference counts, and specific change descriptions. Task Group summary (lines 304-311) provides decomposition guidance. |
| Attack 3 (iter 1): Missing Error Paths for Modified Handlers | ✅ Addressed | New "Error Path Migration" section (tech-design.md:204-248) maps every removed handler-level and service-level error check to its replacement. Includes old/new behavior, user-facing change, and PRD AC cross-reference. Edge case for empty teams documented (lines 247-249). |

---

## Verdict

- **Score**: 86/100
- **Target**: 90/100
- **Gap**: 4 points
- **Breakdown-Readiness**: 18/20 -- **can proceed to /breakdown-tasks** (meets 18/20 gate)
- **Action**: Continue to iteration 3. Priority fixes: (1) Replace security mitigations with concrete verification mechanisms and map them 1:1 to threats, (2) Complete test tooling with runner names and DB strategy for integration tests, (3) Type the `teamRepo.ListTeamBizKeys` return signature and resolve the `isPMRole` contradiction.
