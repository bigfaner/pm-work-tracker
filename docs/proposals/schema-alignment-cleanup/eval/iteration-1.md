# Proposal Evaluation Report: Schema Alignment Post-Refactoring Cleanup

**Document**: `docs/proposals/schema-alignment-cleanup/proposal.md`
**Iteration**: 1
**Date**: 2026-04-26
**Total Score**: 62/100

---

## Dimension Scores

### 1. Problem Definition: 15/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The problem is largely clear: post-refactoring bugs and code quality issues. However, the claim of "2 silent bugs" is slightly misleading. The `assignee_id` bug in `SubItem.Assign()` is confirmed (line 262 uses `"assignee_id"` instead of `"assignee_key"`), but the "filter-by-assignee returns nothing" bug is not actually a bug. Reading `filter_helpers.go`, the `applyItemFilter` function correctly filters by `assignee_key` using a `*string` parameter. The proposal says "filter type mismatch" but the code shows the filter accepts `*string` and queries `assignee_key = ?` -- this appears correct. The "2 silent bugs" framing inflates the urgency. |
| Evidence provided | 5/7 | The proposal references specific files and patterns. The `assignee_id` column bug is verifiable in `sub_item_service.go:262`. The deprecated DTOs are confirmed in `item_dto.go` with `Deprecated:` comment on line 201. However, the "~15 code quality issues" and "~8 architectural inconsistencies" counts are not backed by an enumeration that maps to the 24-item list -- the numbers appear invented for rhetorical effect. |
| Urgency justified | 4/6 | The claim that "assign and filter-by-person are broken" provides urgency. However, as noted, the filter claim is unsubstantiated by the code. Only one bug (assignee_id column name) is clearly user-visible. The urgency is overstated by bundling a real bug with a questionable one. |

**Deductions**: None beyond scoring above.

### 2. Solution Clarity: 12/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 5/7 | The 24-item list grouped into 4 rounds is concrete and actionable. A reader could explain back what will be done. However, some items are vague: "Extract shared `resolveBizKey` helper from the 7 duplicate resolve/parse functions" -- a grep shows only `parseBizKey` in handlers and `ParseID` in pkg, which are not the same pattern being duplicated 7 times. Item 21 says "Unify `SingleItemToVO` + `BatchToVOs`" but these function names do not exist in the codebase. The actual names are `itemPoolToVO`/`itemPoolsToVOs` and `progressRecordToVO`/`progressRecordsToVOs`. This is a factual error. |
| User-facing behavior described | 2/7 | This is a significant gap. The proposal is almost entirely about internal code changes. There is no description of what the end user will experience differently. The only user-facing items are fixing broken assign and filter features, but even those are described in code terms. No screenshots, no user stories, no before/after behavioral descriptions. |
| Distinguishes from alternatives | 5/6 | The two alternatives (do nothing, batch by layer) are clear. The rationale for per-issue commits is stated: "User preferred per-issue commits." This is a valid differentiator, though it relies on a single stakeholder preference rather than a technical argument. |

**Deductions**: -2 for "user-facing behavior described" vagueness -- the proposal uses "better" and "improved" code quality without quantification in multiple places (e.g., "code stays messy", "naming inconsistencies", "accumulated cruft").

### 3. Alternatives Analysis: 8/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 3/5 | Two alternatives are listed (do nothing, batch by layer). This meets the minimum. However, no alternatives are considered for individual issues. For example, the "consolidate two constructors" could also be solved by keeping both but documenting them. The alternatives are at a high level only. |
| Pros/cons for each | 2/5 | The "do nothing" alternative gets a one-liner dismissal: "The bugs remain broken, code stays messy." No honest assessment of its pros (zero risk, zero effort). The "batch by layer" gets one trade-off: "fewer commits but each commit touches more files." This is thin. Neither alternative gets a structured pros/cons analysis. |
| Rationale for chosen approach | 3/5 | The verdict is justified ("User preferred per-issue commits") but the rationale is a single sentence appeal to authority, not a technical argument. Why are per-issue commits better for review? For rollback? For bisecting? These arguments exist but are not made. |

### 4. Scope Definition: 14/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 5/5 | Each of the 24 items is a specific, identifiable deliverable. This is the strongest aspect of the proposal. A developer can take any single item and implement it. |
| Out-of-scope explicitly listed | 5/5 | Four explicit out-of-scope items are named with cross-references to other proposals. This is well done. |
| Scope is bounded | 4/5 | The 4-round structure provides some ordering, but there is no time estimate or deadline. "Run related tests after each fix" is the only completion criterion. No estimate of how long 24 individual commits will take. |

### 5. Risk Assessment: 7/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 2/5 | Only 4 risks are listed, and they are generic. Missing risks: (1) the `SingleItemToVO`/`BatchToVOs` claim is factually wrong, suggesting the proposal author may not have a full picture of the codebase; (2) removing deprecated DTOs could break imports that aren't caught by `go build` if reflection or marshaling is involved; (3) frontend `assigneeId` -> `assigneeKey` rename affects many files and could miss edge cases; (4) no risk around the 24-commit strategy creating merge conflicts on long-lived branches. |
| Likelihood + impact rated | 2/5 | The ratings skew toward "Low" likelihood. The P2 architecture changes are rated "Medium" for both -- but changes to shared interfaces like `TransactionDB` and `ViewService` constructors could break many callers. The "Frontend type changes cause runtime errors" risk is rated "Low" likelihood, but changing `Record<number, string[]>` to `Record<string, string[]>` is a breaking change for all existing permission lookups that use numeric keys -- this is at least Medium likelihood. |
| Mitigations are actionable | 3/5 | Some mitigations are specific ("Grep codebase for `assignee_id` references"), others are generic ("Run full test suite after each P2 change"). The "TypeScript compiler catches mismatches at build time" mitigation for frontend type changes is dangerous -- TypeScript does not catch all runtime type issues, especially with dynamic key access on `Record` types. |

### 6. Success Criteria: 6/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 2/5 | Criteria 1-3 are measurable (bugs fixed, zero deprecated code, all tests pass). Criteria 4-6 are measurable but narrowly scoped. Missing: no criterion for "filter-by-assignee works correctly" despite being listed as a P0 bug. No criterion for frontend form field rename (item 19). No criterion for the `TableRow.mainItemId` type change (item 23). |
| Coverage is complete | 2/5 | Only 6 success criteria for 24 in-scope items. Many items have no corresponding criterion. Items 5-10 (dead code removal) are only partially covered by criterion 2. Items 11-19 (pattern unification) are barely covered. Items 20-24 (architecture alignment) have criteria 4-6 but miss many sub-items. |
| Criteria are testable | 2/5 | Criteria 1 and 3 are testable. Criterion 2 ("Zero deprecated/dead code remaining in modified files") is partially testable -- how do you verify "dead code" is zero? Grep for `Deprecated` comments? Criterion 4 ("No `String()` wraps on values already typed as `string`") is testable via grep. Criteria 5-6 are testable. But the coverage gap makes many items unverified. |

---

## Vague Language Penalty

Instances of vague/unquantified language:
1. "code stays messy" (Alternative A) -- -2
2. "accumulated cruft" (Problem section) -- -2
3. "naming inconsistencies" (Problem section) -- -2

**Total deduction**: -6 (capped at the 3 most prominent instances)

---

## Inconsistency Penalty

- The proposal claims `SingleItemToVO` + `BatchToVOs` exist (item 21) but these functions do not exist in the codebase. The actual functions are `itemPoolToVO`/`itemPoolsToVOs` and `progressRecordToVO`/`progressRecordsToVOs`. **-3**
- The proposal lists "filter-by-assignee returns nothing" as a P0 bug but `filter_helpers.go` shows the filter correctly queries `assignee_key` with a `*string` parameter. The bug may exist elsewhere (e.g., in how the filter value is passed), but the proposal does not pinpoint the actual failure location. **-3** (because this inflates urgency and misleads about scope)

**Total deduction**: -6

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 12 | 20 |
| Alternatives Analysis | 8 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 7 | 15 |
| Success Criteria | 6 | 15 |
| Vague Language Penalty | -6 | — |
| Inconsistency Penalty | -6 | — |
| **Total** | **62** | **100** |

---

## Top 3 Attack Points

### ATTACK 1: Success Criteria [6/15]
Only 6 success criteria for 24 in-scope items. Items like frontend `assigneeId`->`assigneeKey` rename (item 19), `TableRow.mainItemId` type change (item 23), and `formatDate` deduplication (item 24) have no corresponding success criterion. The proposal cannot be verified as complete. Quote: "Success Criteria" section lists only 6 criteria while "In-scope" lists 24 items.

### ATTACK 2: Solution Clarity - User-Facing Behavior [2/7]
The proposal describes 24 code changes but never explains what the user will experience differently. There are no before/after behavioral descriptions, no user stories, and no screenshots. The only user-facing claim is "bugs are fixed" but even that is described in code terms. Quote: "Fix `SubItem.Assign()` column name: `assignee_id` → `assignee_key`" -- this is a code detail, not a user experience description.

### ATTACK 3: Alternatives Analysis [8/15]
The "do nothing" alternative is dismissed in one sentence. The "batch by layer" alternative gets one trade-off line. Neither receives honest pros/cons. The chosen approach is justified by "User preferred per-issue commits" -- an appeal to authority rather than a technical argument. Quote: "User preferred per-issue commits." No explanation of why per-issue is technically superior.
