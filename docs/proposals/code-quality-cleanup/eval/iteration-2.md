# Proposal Evaluation — Iteration 2

**Date**: 2026-04-22
**Score**: 78/100

## DIMENSIONS

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 16 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 11 | 15 |
| Success Criteria | 10 | 15 |

## ATTACKS

The top 3 weaknesses to fix:

1. **Risk mitigations still lack depth and risk identification is incomplete**: The "Phase dependencies cause rework" risk (line 174) has the mitigation "Phase 1 is prerequisite for Phase 3/4" — this restates the dependency, not a mitigation for what goes wrong if rework occurs. Missing risks include: partial completion (only Phase 1-2 ship before feature work resumes), merge conflicts with concurrent feature branches, and the risk that the p95 200ms target is unachievable without schema changes that are out of scope. The "Low/Low" risks for "Large file splits cause import path breakage" and "Phase dependencies cause rework" beg the question of why they are listed at all if both likelihood and impact are low.

2. **Success criteria still have measurement and coverage gaps**: "No duplicated component logic > 10 lines" (line 184) remains subjective — what constitutes "duplicated"? Is copy-paste of 9 lines acceptable? Is structural similarity with different variable names "duplication"? The criterion "Backend repository layer has zero duplicated boilerplate" (line 182) is not objectively testable without defining "boilerplate." Additionally, the per-phase success criteria for Phase 1 (lines 58-61) include "No `as any` type escapes remain for fields that the backend actually returns" — the qualifier "that the backend actually returns" creates a loophole where `as any` casts for other reasons are acceptable, but the final checklist (line 179) says "no `as any` escapes" unconditionally. This inconsistency between per-phase and final criteria remains.

3. **User-facing behavior is still not described**: The proposal never states what the end user will experience. Will page loads feel faster? By how much? Will any UI behavior change (e.g., dropdown rendering after component extraction)? Will there be any visual difference at all? A single sentence like "End users will observe no behavioral changes except faster list view page loads" would close this gap. Without it, the "observable behavior" dimension remains unfilled for a proposal that explicitly affects API response times and frontend component structure.

## Detailed Scores

### Problem Definition (15/20)

**Problem stated clearly (6/7):** The problem is well-structured with six concrete categories: dead code, contract mismatches, N+1 queries, in-memory filtering, code duplication, and monolithic files. Each category includes specific counts and file names (e.g., "18+ unused functions," "9 locations," "~600 lines across 3 frontend files," "ItemViewPage.tsx (1462 lines)"). Minor deduction: line 20 claims the debt "slows development velocity" — this is stated as fact without describing the specific developer friction (e.g., longer onboarding, longer PR review times, fear of touching certain files).

**Evidence provided (4/7):** Strong quantitative evidence for code-level problems (file sizes, function counts, duplication line counts). However, no performance data is cited — no current API response times, no data volume numbers, no profiling results. The N+1 claim on line 15 ("9 locations where list endpoints make per-item DB calls, degrading API response time as data grows") would be far stronger with a single measured data point showing actual degradation. No user feedback or bug reports are referenced. The claim on line 20 that this "will cause performance degradation as data volume grows" remains speculative without current volume or growth projections.

**Urgency justified (5/6):** Substantially improved from iteration 1. The per-phase estimates (line 24: "2-3 weeks for a single developer") and the effort table with duration columns provide time-bounding context that helps answer "why now." The problem describes compounding debt (performance degrades with scale) which provides implicit urgency. Minor gap: no triggering incident or stakeholder demand is cited — the urgency is entirely technical, with no business context for timing.

### Solution Clarity (16/20)

**Approach is concrete (7/7):** Exceptionally concrete. Every phase names specific files, functions, and transformations. Examples: "Delete unused functions: `parseItemIDAsUint` (sub_item_handler.go)" (line 37), "batch-fetch by `WHERE id IN (...)` instead of per-ID lookups" (line 68). A developer could start working from this document immediately.

**User-facing behavior described (3/7):** The proposal still does not describe what end users will observe. There is no statement about whether UI behavior changes, whether pages load faster, or whether the experience is identical. The p95 response time target (line 92) implies faster loads but does not translate this into user-observable terms. For a proposal that restructures frontend components and changes API response times, the absence of any user-facing description is a notable gap.

**Distinguishes from alternatives (6/6):** Improved from iteration 1. The phase ordering is now explained: "proceeds from low-risk dead code removal to higher-impact structural refactoring" (line 24). The phase dependency is explicit in the risk table (line 174). The alternatives table clearly shows why progressive cleanup was chosen over all-at-once and partial-scope approaches.

### Alternatives Analysis (12/15)

**At least 2 alternatives listed (5/5):** Five alternatives are listed including "No cleanup (status quo)." This covers full-scope (all-at-once), partial-scope (frontend-only, backend-only), progressive, and do-nothing options.

**Pros/cons for each (3/5):** The pros/cons remain shallow. "Frontend-only cleanup" has pro "Smaller scope" and con "Leaves backend N+1 and dead code; contract mismatches remain" — these are self-evident restatements of scope, not analytical trade-offs. No effort estimates are provided for alternatives (e.g., how long would frontend-only take vs. the full 2-3 weeks?). The con "Takes longer to complete all phases" for the progressive approach is now partially mitigated by the per-phase estimates, but no comparison estimates exist for alternatives.

**Rationale for chosen approach (4/5):** The verdict column is clear and consistent. The progressive approach is justified by its incremental-shippable property. A brief narrative paragraph beyond the tabular format explaining the decision logic would strengthen the rationale.

### Scope Definition (14/15)

**In-scope items are concrete (5/5):** Improved from iteration 1. "Type safety improvements" has been replaced with "Type safety fixes: remove all `as any` casts, fix `user: any` to `user: User`" (line 156), which is specific and measurable. All other in-scope items name deliverable categories.

**Out-of-scope explicitly listed (5/5):** Clear and specific. Six distinct exclusions prevent scope creep. The parenthetical "(only adding indexes)" on line 161 is a useful nuance.

**Scope is bounded (4/5):** Substantially improved. The effort estimates (line 24: "2-3 weeks for a single developer") and per-phase durations (2-3 days, 3-5 days, etc.) provide concrete time-bounding. The assumption "existing codebase familiarity and no concurrent feature work" (line 24) is a useful caveat. Minor gap: no milestone dates or sequencing constraints beyond the phase order (e.g., can Phase 2 and 3 overlap?).

### Risk Assessment (11/15)

**Risks identified (4/5):** Five risks are listed. The mitigations are improved from iteration 1 — the contradictory "keep existing filter logic" is gone. However, notable risks are still missing: partial completion risk (what if only Phase 1-2 complete before feature work resumes?), merge conflict risk with concurrent feature branches, and the risk that the 200ms p95 target is unachievable without schema changes that are explicitly out of scope.

**Likelihood + impact rated (3/5):** Ratings are present but two risks are rated "Low/Low" (lines 173-174), which raises the question of whether they warrant inclusion. The SQL filtering risk (line 171) at "Medium/Medium" may understate impact — converting 4+ working service methods from in-memory to SQL filtering has significant regression surface. The "Medium/High" for contract mismatch fixes (line 170) seems appropriately rated.

**Mitigations are actionable (4/5):** Improved from iteration 1. "Write integration tests for each converted endpoint covering edge cases (empty results, null fields, multi-page results) before removing in-memory logic" (line 171) is specific and actionable. "Run `tsc --noEmit` and `go vet ./...` before each commit" (line 173) is concrete. However, line 174's mitigation ("Phase 1 is prerequisite for Phase 3/4") restates the dependency rather than mitigating the risk of rework — if Phase 3 requires changes to code that Phase 1 already touched, what is the specific action to prevent rework?

### Success Criteria (10/15)

**Criteria are measurable (4/5):** Improved from iteration 1. The p95 200ms target with "or at least 40% reduction from baseline" (line 187) is now quantified. "Zero unused exports," "No page exceeds 300 lines," and "O(1) association queries" are measurable. However, "No duplicated component logic > 10 lines" (line 184) remains subjective — what counts as "duplicated"? Identical copy-paste? Structural similarity with renamed variables? "Zero duplicated boilerplate" (line 182) similarly lacks a precise definition.

**Coverage is complete (3/5):** Improved from iteration 1 — the final checklist now includes items previously only in per-phase criteria (e.g., "All in-memory filtered endpoints use SQL-level filtering," "Backend repository layer has zero duplicated boilerplate," "Consistent constructor pattern"). However, some per-phase criteria still diverge from the final checklist. Phase 1's "No `as any` type escapes remain for fields that the backend actually returns" (line 61) has a qualifier ("that the backend actually returns") that creates a loophole absent from the unconditional final checklist item (line 179). Phase 2's "No React anti-patterns (`useMemo` for side effects, fetching outside React Query)" (line 91) is not reflected in the final checklist at all.

**Criteria are testable (3/5):** Most criteria are testable: "All existing tests pass" (test runner), "No page exceeds 300 lines" (script), "Zero unused exports" (linter). However, "No duplicated component logic > 10 lines" requires subjective judgment about what constitutes duplication. "Backend repository layer has zero duplicated boilerplate" similarly requires defining "boilerplate" before it can be tested. The p95 response time criterion is testable but requires establishing a baseline measurement first, which is not mentioned as a prerequisite.

## Deductions

- **Vague language penalty (-2 x 1 = -2):**
  - Line 184: "No duplicated component logic > 10 lines" — "duplicated" is not defined precisely enough for objective measurement. What similarity threshold qualifies as duplication?
