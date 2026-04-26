# Proposal Evaluation Report: Schema Alignment Post-Refactoring Cleanup

**Document**: `docs/proposals/schema-alignment-cleanup/proposal.md`
**Iteration**: 4
**Date**: 2026-04-26
**Total Score**: 84/100

---

## Dimension Scores

### 1. Problem Definition: 17/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The three-bullet structure (2 bugs, ~15 code quality, ~8 architectural) with a 24-item enumeration is clear. Deduction: "~15 code quality issues" and "~8 architectural inconsistencies" remain approximate counts that require cross-referencing to the 24-item list. The mapping (items 3-10 = code quality, items 11-19 = pattern + architecture) is implicit rather than explicit. |
| Evidence provided | 6/7 | Bug 1 is verified: `sub_item_service.go:262` uses `"assignee_id"` as the column name, confirmed in source code. Bug 2 is verified: `filter_helpers.go:14-15` passes `*string` (dereferenced `*assigneeKey`) to `WHERE assignee_key = ?`, but the DB column `model.SubItem.AssigneeKey` is `*int64`. Deprecated DTOs (`WeeklyViewResult`, `WeeklyGroupDTO`, `SubItemWeekDTO`, `SubItemSummaryDTO`) confirmed at `item_dto.go:200-240`. Duplicate interfaces confirmed: `TransactionDB` at `team_service.go:19`, `dbTransactor` at `item_pool_service.go:30`. 5 `statusHistorySvc.Record` call sites confirmed at exact line numbers. Frontend `assigneeId` mismatch confirmed across 40+ references. Deduction: The `assigneeKey` in the frontend `MainItem` and `SubItem` types is `string | null` (confirmed at `types/index.ts:95,114`), NOT `*int64`. This means the frontend is already aligned -- the type mismatch is purely a backend concern. The problem statement's characterization of "~8 architectural inconsistencies: systemic int64/uint casting on *Key fields" conflates backend and frontend scope. |
| Urgency justified | 5/6 | "Why now" names concrete broken features. The assign bug is a clear regression. Deduction: The filter-by-person bug's urgency depends on the database engine -- the proposal acknowledges "on SQLite the implicit cast may mask the bug." If production uses SQLite, this is not user-visible now. |

### 2. Solution Clarity: 17/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 6/7 | The 24-item list grouped into 4 dependency-ordered rounds is highly actionable. Each item identifies specific files, functions, and line numbers. Item 2 correctly describes the type mismatch and prescribes `pkg.ParseID`. Deduction: Item 22 has a subtle scope issue. It says "Apply `NotDeleted` scope consistently across all repositories: replace inlined `deleted_flag = 0` checks in `team_repo.go` (lines 37, 222, 225) and `role_repo.go` (lines 32, 44)." But Item 4 in Round 2 says "Remove unused `NotDeleted` scope from `scopes.go`." These two items directly contradict: Item 4 removes the scope, Item 22 applies it. If Item 4 executes first (Round 2 before Round 4), the scope is gone before Item 22 tries to use it. This is a dependency ordering error within the proposal's own structure. |
| User-facing behavior described | 6/7 | The "User-Facing Impact" table provides concrete before/after descriptions. "Clicking 'Assign' on a sub-item sends the request successfully but the `assignee_key` column is never updated" is specific. "Filtering by person returns zero results" is measurable. Deduction: The before-state for Item 1 still mixes code-level and user-level language: "the `assignee_key` column is never updated in the database" is code-level, while "The sub-item remains unassigned after page refresh" is user-level. |
| Distinguishes from alternatives | 5/6 | Per-issue commits are clearly differentiated from batch-by-layer with concrete technical arguments. Deduction: The fundamental question -- why bundle urgent P0 bugs with 22 low-priority cleanups instead of fixing bugs first? -- remains unanswered as an explicit alternative. |

### 3. Alternatives Analysis: 12/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 4/5 | Three alternatives including "do nothing." Structured with Pros/Cons/Verdict. Deduction: The most pragmatic middle-ground alternative -- "fix only the 2 P0 bugs, defer all 22 cleanup items" -- is still absent. This is the option most likely to be chosen by a risk-averse reviewer who agrees the bugs must be fixed but questions coupling them with refactoring. |
| Pros/cons for each | 4/5 | Each alternative has structured tables with honest trade-offs. "Batch by layer" lists concrete concerns about 300-line diffs. "Per-issue commits" acknowledges merge conflict accumulation. Deduction: The con for option C ("24 commits on the branch; if the branch lives long, merge conflicts accumulate") is the strongest objection to the chosen approach, yet the verdict dismisses it with "the trade-off favors reviewability" without quantitative evidence or a concrete merge timeline. |
| Rationale for chosen approach | 4/5 | Technical arguments for per-issue commits (bisect precision, individual revert, smaller review diffs) are concrete. The 4-round dependency ordering is well-justified. Deduction: The Round 2/Round 4 dependency conflict (Item 4 removes `NotDeleted`, Item 22 applies it) undermines the dependency ordering claim. If the rounds are truly dependency-ordered, this internal contradiction should not exist. |

### 4. Scope Definition: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 4/5 | All 24 items are specific with file names, function names, and line numbers. Deduction: Item 17 claims "22 truly redundant `String()` wraps" but the grep results show `String(item.assigneeKey)` at `useItemViewPage.ts:118` where `assigneeKey` is `string | null` (confirmed at `types/index.ts:95`). On a `string | null` value, `String()` converts `null` to `"null"` (the string), which is a semantic change vs. leaving it as `null`. The proposal says to "retain 13 necessary conversions" but does not explain how the 22 vs 13 split was determined. This is a scope accuracy issue. |
| Out-of-scope explicitly listed | 5/5 | Four items explicitly named with cross-references: performance optimizations (referenced to existing `code-quality-cleanup` proposal), file splitting, dialog consolidation, and new features. |
| Scope is bounded | 4/5 | The 4-round structure provides ordering. Scope is bounded by item count (24). Deduction: No time or effort estimate. The scope is bounded by enumeration but not by calendar or developer-days, making it impossible to assess whether this is a 2-day or 2-week effort. |

### 5. Risk Assessment: 10/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 3/5 | Seven risks listed. Deduction: Two significant risks are still missing: (1) The Item 4/Item 22 dependency conflict -- Item 4 removes `NotDeleted` scope in Round 2, but Item 22 in Round 4 needs it. This is not a hypothetical risk; it is a concrete scheduling contradiction within the proposal. (2) Item 17's `String()` removal on `string | null` values (like `assigneeKey`) changes semantics -- `String(null)` produces `"null"` (the string), so removing it could alter filter behavior. This conflation risk remains unlisted. |
| Likelihood + impact rated | 3/5 | Ratings are present for all 7 risks. Deduction: Risk 4 ("Frontend type changes cause runtime errors") is rated Low/Medium. However, changing `Record<number, string[]>` to `Record<string, string[]>` at `types/index.ts:5` would break all existing permission lookups that use numeric keys (confirmed at `types/index.ts:5`: `teamPermissions: Record<number, string[]>`). This is at least Medium likelihood given the number of call sites. Risk 5 (Item 2 fix location) contains a self-contradiction: the solution says fix in `filter_helpers.go` but the risk says the fix location may be misidentified, which means either the solution is wrong or the risk is overstated. |
| Mitigations are actionable | 4/5 | Risk 5 has a specific, testable mitigation. Risk 7's mitigation ("verify `NotDeleted` scope produces identical SQL") is concrete. Risk 6's mitigation ("rebase frequently") is standard but not quantified. Deduction: Risk 4's mitigation ("TypeScript compiler catches mismatches at build time") is overconfident -- TypeScript does not catch runtime type issues with dynamic key access on `Record` types (e.g., `teamPermissions[someNumber]` when the key type changes from `number` to `string`). |

### 6. Success Criteria: 15/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 5/5 | 22 specific criteria, most grep-verifiable (e.g., criterion 4: `grep -r "WeeklyViewResult\|..." backend/` returns zero; criterion 8: `grep -r "dbTransactor" backend/` returns zero). All criteria specify concrete verification methods. |
| Coverage is complete | 5/5 | All 4 rounds have corresponding criteria. All 24 items are individually covered. Criterion 12 covers Item 15 (`userToDTO`). Criterion 13 covers Item 16 (`statusHistorySvc.Record`). Items 9-10 are covered by criterion 7. |
| Criteria are testable | 5/5 | Most criteria are grep-verifiable or test-verifiable. Criteria 1-2 reference specific test assertions. Criterion 19 uses a grep pattern. Deduction: The `formatDate` grep pattern (`function formatDate\|const formatDate`) could produce false positives with `formatDateInput` at `GanttViewPage.tsx:64` and `formatDate` wrappers at `RoleManagementPage.tsx:139`, `TeamDetailPage.tsx:53`, `TeamManagementPage.tsx:42`. However, since the proposal specifically identifies these as targets for extraction, the false positives are actually the intended matches -- the criterion verifies consolidation, not absence. |

---

## Vague Language Penalty

Instances of vague/unquantified language:

1. Line 14: "duplicate patterns introduced as workarounds during migration" -- which patterns? What workarounds? The 24-item list resolves this partially (items 11-19 describe specific patterns), but the problem statement itself uses this phrase without enumeration. **-2**

No other instances of "better", "improved", "enhanced" without quantification found.

**Total deduction**: -2

---

## Inconsistency Penalty

**Item 4 vs Item 22 dependency conflict**: Item 4 (Round 2, line 29) says "Remove unused `NotDeleted` scope from `scopes.go`." Item 22 (Round 4, line 51) says "Apply `NotDeleted` scope consistently across all repositories: replace inlined `deleted_flag = 0` checks in `team_repo.go`... and `role_repo.go`... with the `NotDeleted` scope from `scopes.go`."

These two items directly contradict each other. Item 4 removes the scope that Item 22 needs. Round 2 executes before Round 4, so Item 4 would run first. The proposal's own 4-round dependency ordering is internally inconsistent. A developer following this proposal sequentially would delete the scope, then attempt to use it two rounds later. **-3**

**Total deduction**: -3

---

## Previous Attack Resolution

| Attack (Iteration 3) | Status | Evidence |
|--------|--------|----------|
| **ATK 1**: Risk Assessment missing key risks (String() conflation on Item 17) | **NOT resolved** | Item 17 still claims "22 truly redundant String() wraps" without acknowledging that `String(null)` on `string | null` types produces `"null"` (the string), which is a semantic change. The risk of breaking filter behavior at `useItemViewPage.ts:118` is unlisted. |
| **ATK 2**: Internal contradiction between Solution and Risks on Item 2 fix location | **NOT resolved** | Item 2 solution (line 25) says fix in `filter_helpers.go` with `pkg.ParseID`. Risk 5 (line 113) says the fix location may be misidentified. Both statements cannot be simultaneously true. |
| **ATK 3**: Missing "P0-only" alternative | **NOT resolved** | The three alternatives remain: do nothing, batch by layer, per-issue commits. The pragmatic middle ground (fix only the 2 bugs) is still absent. |

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 17 | 20 |
| Solution Clarity | 17 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 13 | 15 |
| Risk Assessment | 10 | 15 |
| Success Criteria | 15 | 15 |
| Vague Language Penalty | -2 | -- |
| Inconsistency Penalty | -3 | -- |
| **Total** | **84** | **100** |

---

## Top 3 Attack Points

### ATTACK 1: Internal Item 4/Item 22 dependency contradiction [Solution Clarity + Scope]
The proposal's Round 2 Item 4 (line 29) says "Remove unused `NotDeleted` scope from `scopes.go`." Round 4 Item 22 (line 51) says "Apply `NotDeleted` scope consistently across all repositories... with the `NotDeleted` scope from `scopes.go`." These items directly contradict: Item 4 removes the scope that Item 22 needs. The 4-round dependency ordering is the proposal's core organizational principle, yet it contains a scheduling impossibility. The scope is currently defined at `scopes.go:7` and only used there (confirmed by grep: no repo files import `NotDeleted`). Item 22 wants to expand its usage, but Item 4 deletes it first. This must be resolved by either removing Item 4 or reordering it after Item 22.

### ATTACK 2: Risk Assessment underrates frontend type-change risk and misses String() semantic risk [Risk Assessment: 10/15]
Risk 4 rates frontend type changes as Low/Medium. But changing `Record<number, string[]>` to `Record<string, string[]>` (Item 18) at `types/index.ts:5` breaks all existing permission lookups that use numeric team keys -- TypeScript's type system does not catch runtime `Record` key mismatches. Additionally, Item 17's claim of "22 truly redundant `String()` wraps" does not account for semantic changes: at `useItemViewPage.ts:118`, `String(item.assigneeKey)` on a `string | null` value converts `null` to `"null"` (the string literal). Removing this wrap changes the filter's null-handling behavior. Neither risk is identified.

### ATTACK 3: Alternatives missing minimum-effort option [Alternatives Analysis: 12/15]
The three alternatives (do nothing, batch by layer, per-issue commits) are all-or-nothing on the 24-item scope. The most pragmatic alternative -- "fix only the 2 P0 bugs, defer all 22 cleanup items" -- is not considered. The "do nothing" alternative's own con ("P0 bugs remain broken") already concedes that fixing the bugs alone would be valuable, yet this middle-ground option is absent. A reviewer who agrees the bugs must be fixed but questions coupling them with 22 low-priority cleanups has no alternative to choose from. This gap weakens the proposal's ability to gain incremental approval.
