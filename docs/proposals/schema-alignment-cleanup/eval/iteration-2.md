# Proposal Evaluation Report: Schema Alignment Post-Refactoring Cleanup

**Document**: `docs/proposals/schema-alignment-cleanup/proposal.md`
**Iteration**: 2
**Date**: 2026-04-26
**Total Score**: 78/100

---

## Dimension Scores

### 1. Problem Definition: 16/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The three-bullet structure (bugs, code quality, architectural) is clear. The 24-item enumeration makes the problem concrete. Deduction: "~15 code quality issues" and "~8 architectural inconsistencies" (lines 14-15) are approximate counts that don't map cleanly to the 24-item list below. Line 14 says "naming inconsistencies between form fields and API fields" without specifying which form fields or API fields. |
| Evidence provided | 5/7 | The `assignee_id` bug is confirmed at `sub_item_service.go:262` (verified in codebase). Deprecated DTOs are verifiable in `item_dto.go`. The "~15" and "~8" counts remain unbacked by direct enumeration mapping to the 24 items. The second bug claim ("filter-by-assignee returns nothing") points to `filter_helpers.go` but the code there is correct: `assigneeKey *string` correctly queries `assignee_key = ?`. The actual issue, if it exists, is the `*int64` model type vs `*string` DTO type mismatch -- not a bug in `filter_helpers.go` itself. |
| Urgency justified | 5/6 | The "Why now" section (line 17) names concrete broken features. The assign bug is a clear regression. Deduction: The filter-by-person claim still lacks precise failure location. Saying "filter type mismatch in `filter_helpers.go`" when the filter helper code is correct undermines credibility of the urgency claim. |

### 2. Solution Clarity: 16/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 6/7 | The 24-item list grouped into 4 dependency-ordered rounds is actionable. Each item identifies a specific change. The factual error from iteration 1 (wrong function names in item 21) is now corrected to `itemPoolToVO`/`itemPoolsToVOs` and `progressRecordToVO`/`progressRecordsToVOs`. Deduction: Item 22 says "apply `NotDeleted` scope consistently or remove it entirely" -- these are two different solutions with different scopes and risks. The author has not decided which path to take, leaving ambiguity in what will actually be built. |
| User-facing behavior described | 5/7 | Significantly improved from iteration 1. The new "User-Facing Impact" section (lines 55-64) provides a before/after table for both P0 bugs. The statement that the 22 internal items produce "no user-visible behavioral change" is clear and honest. Deduction: The before/after table still reads like code-level descriptions ("the `assignee_key` column is never updated in the database") rather than user-visible behavior. Better: "User clicks Assign, selects a person, clicks Save. After page refresh, the assignment is lost." |
| Distinguishes from alternatives | 5/6 | The per-issue approach is clearly differentiated from batch-by-layer with concrete technical arguments: precise bisect, individual revert, smaller review diffs. This is a genuine technical advantage. |

### 3. Alternatives Analysis: 12/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 4/5 | Three alternatives including "do nothing." Structured with Pros/Cons/Verdict tables. Deduction: All alternatives focus on commit granularity. A natural fourth option -- "fix only the P0 bugs, defer all cleanup" -- is not considered. This would be the minimum-risk, minimum-effort alternative. |
| Pros/cons for each | 4/5 | Improved from iteration 1. Each alternative has a structured table. "Do nothing" honestly lists "Zero risk, zero effort" as pros. "Batch by layer" lists concrete trade-offs. Deduction: The con for option C ("24 commits on the branch; if the branch lives long, merge conflicts accumulate") raises a real concern that is not addressed in the risk section or mitigation plan. |
| Rationale for chosen approach | 4/5 | The verdict for option C now includes concrete technical arguments: git bisect, individual revert, smaller review diffs. No longer just an appeal to authority. Deduction: The acknowledged con (merge conflict accumulation for long-lived branches) is not countered -- how long will this branch live? What's the merge window? |

### 4. Scope Definition: 14/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 5/5 | All 24 items are specific, actionable deliverables. A developer can take any item and implement it. |
| Out-of-scope explicitly listed | 5/5 | Four items explicitly named with cross-references to other proposals. Well done. |
| Scope is bounded | 4/5 | The 4-round structure provides ordering. Still no time or effort estimate. The scope is bounded by item count (24) but not by calendar or developer-days. |

### 5. Risk Assessment: 9/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 3/5 | Four risks listed, up from generic ones in iteration 1. Missing risks: (1) Item 2 mislocalizes the filter bug to `filter_helpers.go` when the code there is correct -- a developer may "fix" the wrong file. (2) Item 22 is unresolved about which approach (apply consistently vs. remove entirely) -- design ambiguity is an execution risk. (3) The 24-commit merge conflict risk is mentioned as a "con" in alternatives but not tracked as a risk. |
| Likelihood + impact rated | 3/5 | Ratings are present for all four risks. Deduction: "Frontend type changes cause runtime errors" is rated Low likelihood, but changing `Record<number, string[]>` to `Record<string, string[]>` breaks all existing permission lookups using numeric keys -- this is at least Medium. The "P0 fixes reveal hidden dependents" risk is rated Low/Medium but the P0 bugs have been in the codebase for 30+ commits, making hidden dependents more likely than Low. |
| Mitigations are actionable | 3/5 | "Grep codebase for `assignee_id` references" is specific and actionable. "TypeScript compiler catches mismatches at build time" is overconfident -- TypeScript does not catch runtime type issues with dynamic key access on Record types. "Run full test suite after each P2 change" is generic but acceptable for a lower-priority round. |

### 6. Success Criteria: 11/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 4/5 | Massively improved from iteration 1. 20 specific criteria, many grep-verifiable (e.g., criterion 4: `grep -r "WeeklyViewResult\|..." backend/` returns zero results). Deduction: Criterion 17 "Soft-delete scope applied consistently across all repositories or removed entirely" is not measurable -- "applied consistently" is a judgment, and the "or" creates two possible outcomes with no way to distinguish success. |
| Coverage is complete | 4/5 | All 4 rounds have corresponding criteria. Most of the 24 items are individually covered. Deduction: Item 15 ("Extract shared `userToDTO` base conversion from `auth_handler` and `admin_service`") has no explicit criterion. Item 16 ("Extract status-history recording helper from 4 duplicate call sites") has no explicit criterion. These are subsumed under the general Round 3 header but not individually verified. |
| Criteria are testable | 3/5 | Most criteria are grep-verifiable or test-verifiable. Criteria 1-2 reference specific test assertions. Deduction: Criterion 11 ("`team_handler.teamToDTO` returns a typed struct, not `gin.H`") requires manual code inspection. Criterion 17 is untestable as stated. Criterion 19 ("`grep -rn "function formatDate\|const formatDate" frontend/src/` returns exactly one definition") could produce false positives if the pattern appears in comments or imports. |

---

## Vague Language Penalty

Instances of vague/unquantified language:

1. Line 14: "naming inconsistencies between form fields and API fields" -- which form fields? Which API fields? Not specified. **-2**
2. Line 17: "accumulated cruft" -- unquantified, rhetorical rather than factual. **-2**
3. Line 14: "duplicate patterns introduced as workarounds during migration" -- which patterns? What workarounds? **-2**

**Total deduction**: -6

---

## Inconsistency Penalty

- Item 2 claims "Fix `assignee_key` filter type mismatch in `filter_helpers.go`" but `filter_helpers.go:15` correctly queries `assignee_key = ?` with `*string`. The model stores `*int64` (see `main_item.go:13`, `sub_item.go:12`). If a bug exists, it is the `*string` vs `*int64` type mismatch between DTO and model, not a bug in `filter_helpers.go`. This mislocalization could lead a developer to modify working code. **-3**

**Total deduction**: -3

---

## Previous Attack Resolution

| Attack | Status | Evidence |
|--------|--------|----------|
| Success Criteria: only 6 criteria for 24 items | **Resolved** | 20 explicit criteria now cover all 4 rounds. Grep-verifiable criteria are specific and testable. |
| Solution Clarity: no user-facing behavior | **Mostly resolved** | New "User-Facing Impact" section with before/after table. Still reads as code-level description rather than user experience. |
| Alternatives Analysis: weak pros/cons, appeal to authority | **Resolved** | Structured Pros/Cons/Verdict tables for all alternatives. Technical arguments for chosen approach (bisect, revert, review diffs). |

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 16 | 20 |
| Solution Clarity | 16 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 9 | 15 |
| Success Criteria | 11 | 15 |
| Vague Language Penalty | -6 | -- |
| Inconsistency Penalty | -3 | -- |
| **Total** | **78** | **100** |

---

## Top 3 Attack Points

### ATTACK 1: Risk Assessment [9/15]
The risk section has 4 risks but misses key risks: (1) Item 2 mislocalizes the filter bug to `filter_helpers.go` when the code there is correct -- a developer could "fix" working code and introduce a real bug. (2) Item 22 is undecided between "apply consistently" and "remove entirely" -- this design ambiguity is an execution risk not captured. (3) The merge-conflict risk from 24 commits on a long-lived branch is acknowledged as a "con" in the alternatives section but not tracked as a risk with likelihood/impact/mitigation. Quote from proposal: "24 commits on the branch; if the branch lives long, merge conflicts accumulate" (line 89, Alternatives section) -- this risk exists but is absent from the Risk table.

### ATTACK 2: Success Criteria coverage gaps [11/15]
While improved from 6 to 20 criteria, gaps remain. Item 15 ("Extract shared `userToDTO` base conversion") and Item 16 ("Extract status-history recording helper") have no explicit success criterion. Criterion 17 ("Soft-delete scope applied consistently or removed entirely") is unmeasurable due to the "or" creating two undefined outcomes. Quote: Criterion 17 -- "applied consistently across all repositories or removed entirely (no mixed usage)" -- "applied consistently" is a judgment call, not a binary check.

### ATTACK 3: Factual mislocalization of bug [Problem Definition + Solution]
Item 2 says "Fix `assignee_key` filter type mismatch in `filter_helpers.go`" but `filter_helpers.go:7` takes `assigneeKey *string` and line 15 correctly queries `WHERE assignee_key = ?`. The actual type mismatch is between the DTO (`*string`) and the model (`*int64` at `main_item.go:13`, `sub_item.go:12`). The proposal directs effort at the wrong file. This inflates Problem Definition (questionable second P0 bug) and weakens Solution Clarity (fix targets wrong location). Quote from proposal: "Fix `assignee_key` filter type mismatch in `filter_helpers.go`" (line 25) vs actual code at `filter_helpers.go:15`: `query = query.Where("assignee_key = ?", *assigneeKey)` which is correct.
