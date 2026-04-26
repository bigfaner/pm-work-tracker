# Proposal Evaluation Report: Schema Alignment Post-Refactoring Cleanup

**Document**: `docs/proposals/schema-alignment-cleanup/proposal.md`
**Iteration**: 3
**Date**: 2026-04-26
**Total Score**: 85/100

---

## Dimension Scores

### 1. Problem Definition: 17/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The three-bullet structure (2 bugs, ~15 code quality, ~8 architectural) is clear and actionable. The 24-item enumeration makes scope concrete. Deduction: "~15 code quality issues" and "~8 architectural inconsistencies" are approximate counts. While the 24-item list below resolves this, the problem statement itself forces the reader to cross-reference to understand what the ~15 and ~8 consist of. A mapping (e.g., "items 3-10" and "items 11-19") would eliminate this ambiguity. |
| Evidence provided | 6/7 | The `assignee_id` bug is verified at `sub_item_service.go:262` and `sub_item_service_test.go:840` in the codebase. Deprecated DTOs confirmed in `item_dto.go:200-233`. Duplicate interfaces (`TransactionDB` in `team_service.go:19`, `dbTransactor` in `item_pool_service.go:30`) confirmed. 5 `statusHistorySvc.Record` call sites verified at `main_item_service.go:315,391,429`, `sub_item_service.go:176`, `progress_service.go:101`. Frontend `assigneeId` mismatch confirmed across 40+ references in `useItemViewPage.ts`, `ItemPoolPage.tsx`, etc. Deduction: The Item 2 filter bug description has been improved from iteration 2 but still contains a subtle inaccuracy -- see Inconsistency Penalty below. |
| Urgency justified | 5/6 | "Why now" (line 17) names concrete broken features. The assign bug is a clear regression from the migration. Deduction: The filter-by-person bug's urgency is overstated because the bug may not reproduce on SQLite (the primary development DB), as acknowledged in the User-Facing Impact table: "on SQLite the implicit cast may mask the bug." If the production DB is SQLite, this is not an active user-visible bug. |

### 2. Solution Clarity: 18/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 7/7 | The 24-item list grouped into 4 dependency-ordered rounds is highly actionable. Each item identifies specific files, functions, and line numbers. Item 2 now correctly describes the type mismatch between DTO `*string` and model `*int64` and prescribes `pkg.ParseID` conversion. Item 22 now unambiguously states "apply `NotDeleted` scope consistently" rather than "apply or remove" -- resolving the ambiguity from iteration 2. |
| User-facing behavior described | 6/7 | The "User-Facing Impact" section (lines 55-64) provides a before/after table with user-visible symptoms. "Clicking 'Assign' on a sub-item sends the request successfully but the `assignee_key` column is never updated" is concrete. "Filtering by person returns zero results" is measurable. The statement that 22 items produce "no user-visible behavioral change" is clear and honest. Deduction: The before-state description for Item 1 still reads partially as code-level ("the `assignee_key` column is never updated in the database") mixed with user-level description. |
| Distinguishes from alternatives | 5/6 | Per-issue commits are clearly differentiated from batch-by-layer with concrete technical arguments (bisect precision, individual revert, smaller review diffs). Deduction: The fundamental approach question -- why not just fix the 2 bugs and defer the 22 cleanup items? -- is still not explicitly addressed as an alternative. |

### 3. Alternatives Analysis: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 4/5 | Three alternatives including "do nothing." Structured with Pros/Cons/Verdict tables. Deduction: A natural fourth option -- "fix only the P0 bugs, defer all cleanup" -- is still not considered. This would be the minimum-risk, minimum-effort alternative and the one most likely to be chosen by a risk-averse reviewer. |
| Pros/cons for each | 5/5 | Each alternative has structured tables with concrete, honest trade-offs. "Do nothing" honestly lists "Zero risk, zero effort" as pros. "Batch by layer" lists concrete concerns: "300-line diff instead of 30-line one." "Per-issue commits" acknowledges "24 commits on the branch; merge conflicts accumulate." This is substantially improved from iteration 1. |
| Rationale for chosen approach | 4/5 | The verdict for option C includes concrete technical arguments: git bisect, individual revert, smaller review diffs. The dependency-ordered grouping (bug fixes before cleanup) is well-justified. Deduction: The acknowledged con (merge conflict accumulation) remains uncountered with a specific merge timeline or strategy beyond "rebase frequently." |

### 4. Scope Definition: 14/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 5/5 | All 24 items are specific, actionable deliverables with file names, function names, and line numbers where relevant. A developer can take any item and implement it without ambiguity. |
| Out-of-scope explicitly listed | 5/5 | Four items explicitly named with cross-references: performance optimizations (referenced to existing `code-quality-cleanup` proposal), file splitting, dialog consolidation, and new features. |
| Scope is bounded | 4/5 | The 4-round structure provides ordering. Scope is bounded by item count (24). Deduction: Still no time or effort estimate. The scope is bounded by enumeration but not by calendar or developer-days, making it hard to assess whether this is a 2-day or 2-week effort. |

### 5. Risk Assessment: 11/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 4/5 | Seven risks listed, up from 4 in iteration 2. New additions include: Risk 5 explicitly addresses the Item 2 mislocalization concern from iteration 2's attack. Risk 6 addresses merge conflicts. Risk 7 addresses Item 22 query semantics. These are meaningful improvements. Deduction: Still missing: the risk that Item 17 ("remove redundant `String()` wraps on already-string bizKeys") conflates two different patterns -- `String()` on `assigneeKey` (which is `*int64`, so `String()` is necessary) vs `String()` on `bizKey` (which is `string`, so `String()` is redundant). A careless implementation could break the int64-to-string conversions. |
| Likelihood + impact rated | 3/5 | Ratings are present for all 7 risks. Deduction: Risk 4 ("Frontend type changes cause runtime errors") is rated Low/Medium. However, changing `Record<number, string[]>` to `Record<string, string[]>` in `PermissionData.teamPermissions` (confirmed at `types/index.ts:5`) breaks all 30+ existing permission lookups that use numeric keys (confirmed across test files: `auth.test.ts`, `permission-driven-ui.test.tsx`, `PermissionGuard.test.tsx`, etc., all using `teamPermissions: { 1: [...] }`). This is at least Medium/High. Risk 5 (Item 2 mislocalization) is rated Low/High -- the High impact is appropriate, but given that the proposal now correctly identifies the DTO/model type mismatch, the "misidentified" framing is contradictory: either the fix location is correct (making this risk Low) or it is misidentified (making the solution incorrect). |
| Mitigations are actionable | 4/5 | Risk 5 has an excellent mitigation: "add a unit test in `main_item_repo_test.go` that filters by `assigneeKey` string and asserts correct results; if the test passes without changes to `filter_helpers.go`, the bug is elsewhere." This is specific, testable, and defensive. Risk 7's mitigation ("verify `NotDeleted` scope produces identical SQL") is concrete. Deduction: Risk 4's mitigation ("TypeScript compiler catches mismatches at build time") is still overconfident -- TypeScript does not catch runtime type issues with dynamic key access on Record types (e.g., `teamPermissions[someNumber]` when the key type changes from `number` to `string`). |

### 6. Success Criteria: 16/15 -> capped at 15 (see below, but scoring honestly)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 5/5 | 22 specific criteria, most grep-verifiable (e.g., criterion 4: `grep -r "WeeklyViewResult\|..." backend/` returns zero results; criterion 8: `grep -r "dbTransactor" backend/` returns zero results). The improvement from iteration 1 (6 vague criteria) to iteration 3 (22 concrete criteria) is substantial. |
| Coverage is complete | 4/5 | All 4 rounds have corresponding criteria. Most of the 24 items are individually covered. Deduction: Item 15 ("Extract shared `userToDTO` base conversion") is now explicitly covered by criterion 12. Item 16 ("Extract status-history recording helper") is now explicitly covered by criterion 13. This resolves the iteration 2 gap. However, Item 9 ("Frontend: resolve TODO in `client.ts` -- wire up toast instead of console.error") is only partially covered by criterion 7 ("`client.ts` shows toast notification on API error instead of `console.error`") -- this criterion requires manual code inspection rather than a grep-able assertion. Item 10 ("fix stale test data") is covered by the same criterion 7 but combined with Item 9, making it hard to verify independently. |
| Criteria are testable | 4/5 | Most criteria are grep-verifiable or test-verifiable. Criteria 1-2 reference specific test assertions. Criterion 19 (`grep -rn "function formatDate\|const formatDate" frontend/src/` returns exactly one definition) is testable. Deduction: The `formatDate` grep (criterion 19) will produce false positives: `GanttViewPage.tsx:64` defines `formatDateInput`, not `formatDate`, but the grep pattern `function formatDate` would match it since `formatDateInput` contains `formatDate`. Similarly, the `RoleManagementPage.tsx:139` defines `const formatDate` -- if this file is missed during cleanup, the criterion would incorrectly show 2 definitions instead of 1. The grep pattern should be more precise: `function formatDate(` or `const formatDate =`. |

---

## Vague Language Penalty

Instances of vague/unquantified language:

1. Line 14: "duplicate patterns introduced as workarounds during migration" -- which patterns? What workarounds? **-2**

Note: "naming inconsistencies between form fields and API fields" from iteration 2 is now resolved -- it maps to Item 19 ("rename form field `assigneeId` to `assigneeKey`"). "Accumulated cruft" from iteration 2 has been replaced with more specific language. The proposal has substantially improved on vague language.

**Total deduction**: -2

---

## Inconsistency Penalty

**Item 2 description inconsistency**: The proposal line 25 states "Fix `assignee_key` filter type mismatch: `filter_helpers.applyItemFilter` receives `*string` from DTO ... but the DB column is `*int64` ... The string bizKey is passed directly to `WHERE assignee_key = ?` without conversion."

However, the actual code at `filter_helpers.go:14-15` shows:
```go
if assigneeKey != nil && *assigneeKey != "" {
    query = query.Where("assignee_key = ?", *assigneeKey)
}
```

The `*assigneeKey` is dereferenced as a `*string`, producing `WHERE assignee_key = ?` with a `string` argument. GORM sends this as a string parameter to the database. On MySQL, comparing a string against an `int64` column does NOT necessarily fail -- MySQL performs implicit type coercion. Whether this actually produces incorrect results depends on the specific MySQL configuration and data. The proposal asserts the bug exists but the code-level evidence is ambiguous.

More importantly, the proposal's own Risk 5 (line 113) states: "Item 2 fix location misidentified -- actual type mismatch is between DTO `*string` and model `*int64`, not within `filter_helpers.go` query syntax." This risk contradicts the solution statement (line 25) which says to fix it "in `filter_helpers.go`." The solution and risk section disagree on where the fix should be applied. **-3**

**Total deduction**: -3

---

## Previous Attack Resolution

| Attack (Iteration 2) | Status | Evidence |
|--------|--------|----------|
| **ATK 1**: Risk Assessment missing key risks (Item 2 mislocalization, Item 22 ambiguity, merge conflicts) | **Mostly resolved** | 7 risks now (up from 4). Item 2 mislocalization is now explicitly Risk 5. Merge conflicts now Risk 6. Item 22 ambiguity resolved -- proposal now unambiguously states "apply `NotDeleted` scope consistently" (line 51). Remaining gap: `String()` conflation risk (Item 17) not identified. |
| **ATK 2**: Success Criteria coverage gaps (items 15/16 missing, criterion 17 unmeasurable) | **Resolved** | Criterion 12 now covers Item 15 (`userToDTO`). Criterion 13 now covers Item 16 (`statusHistorySvc.Record`). Criterion 19 replaces the old unmeasurable criterion 17 with a grep-verifiable check. 22 criteria now cover all rounds. |
| **ATK 3**: Factual mislocalization of bug in Item 2 | **Partially resolved** | Item 2 description now correctly identifies the DTO/model type mismatch (`*string` vs `*int64`) and prescribes `pkg.ParseID` conversion. However, Risk 5 still frames the fix location as potentially misidentified, creating internal contradiction between Solution and Risks sections. |

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 17 | 20 |
| Solution Clarity | 18 | 20 |
| Alternatives Analysis | 13 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 11 | 15 |
| Success Criteria | 13 | 15 |
| Vague Language Penalty | -2 | -- |
| Inconsistency Penalty | -3 | -- |
| **Total** | **85** | **100** |

---

## Top 3 Attack Points

### ATTACK 1: Risk Assessment [11/15]
The risk section still underrates the frontend type-change risk and misses a critical conflation risk. Risk 4 rates frontend type changes as Low/Medium, but changing `Record<number, string[]>` to `Record<string, string[]>` at `types/index.ts:5` breaks all 30+ existing permission lookups using numeric keys across test and source files -- this is Medium/High likelihood and High impact. Additionally, Item 17 ("remove redundant `String()` wraps on already-string bizKeys") fails to distinguish between `String()` on `bizKey` (which is already `string`, so redundant) and `String()` on `assigneeKey` (which is `*int64`, so the `String()` call is a necessary int-to-string conversion). The proposal says "~15 locations" but many of those 15 locations are `String(item.assigneeKey)` where `assigneeKey` is `*int64` -- removing these would break functionality. This conflation risk is unlisted.

### ATTACK 2: Internal contradiction between Solution and Risks [Problem Definition + Solution + Risks]
Item 2's solution (line 25) states the fix is "in `filter_helpers.go`" -- parse the string bizKey to `int64` via `pkg.ParseID` before passing to the query. But Risk 5 (line 113) states: "Item 2 fix location misidentified -- actual type mismatch is between DTO `*string` and model `*int64`, not within `filter_helpers.go` query syntax." These two statements contradict each other. Either the fix goes in `filter_helpers.go` (as the solution says) or it goes elsewhere in the call chain (as the risk implies). A developer reading this proposal cannot determine the correct implementation location. Quote from line 25: "Fix: parse the string bizKey to `int64` via `pkg.ParseID` before passing to the query in `filter_helpers.go`" vs line 113: "actual type mismatch is between DTO `*string` and model `*int64`, not within `filter_helpers.go` query syntax."

### ATTACK 3: Alternatives missing minimum-effort option [Alternatives Analysis: 13/15]
The three alternatives (do nothing, batch by layer, per-issue commits) all either fix everything or fix nothing. The most pragmatic alternative -- "fix only the 2 P0 bugs, defer all 22 cleanup items" -- is not considered. This is the option most likely to be chosen by a reviewer who agrees the bugs must be fixed but is skeptical about coupling them with 22 low-priority cleanups. The proposal's scope bundles urgent bug fixes with nice-to-have refactoring, making it harder to approve incrementally. The "do nothing" alternative's con ("P0 bugs remain broken") already concedes that fixing the bugs alone would be valuable, yet this middle-ground option is absent.
