---
iteration: 2
score: 908
target: 850
verdict: PASS
dimensions:
  completeness: 173
  semantic_purity: 194
  precondition_exclusivity: 144
  fact_alignment: 136
  surface_fitness: 135
  internal_consistency: 126
---

# Journey Eval: milestone-item-management (Iteration 2)

**Verdict: PASS** — 908/1000 (target: 850)

## Dimension Breakdown

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| Completeness | 173/200 | 120 | PASS |
| Semantic Purity | 194/200 | 120 | PASS |
| Precondition Exclusivity | 144/150 | 90 | PASS |
| Fact Alignment | 136/150 | 90 | PASS |
| Surface Fitness | 135/150 | 90 | PASS |
| Internal Consistency | 126/150 | 90 | PASS |

## Scoring Rationale

### Completeness: 173/200

- **Metadata (48/50):** Name valid kebab-case, risk_level "High" justified by mutation/irreversible operations, surface_types correct, sources listed. Complete and well-formed.
- **Steps (75/80):** 22 steps total (7 happy path + 15 edge cases). All have step name, user action, expected result. Coherent ordered sequence. Minor gap: Step 5 (Navigate to MI detail) is a single thin step with no edge cases.
- **Outcomes (50/70):** validation-error covered thoroughly (4b-4d), session-expired covered (E1), loading-state covered (1b, 4e). Missing: network-error boundary for mutation API calls (unbind failure, add creation failure, drag-rebind API error). Surface-web rules list network-error as an additional common boundary outcome.

### Semantic Purity: 194/200

- **Natural language (78/80):** All expected results describe user-observable behavior. No regex, CSS selectors, XPath, or framework assertions. Clean natural language throughout. Minor: "danger zone" is a component section name but acceptable as user-visible UI labeling.
- **Declarative preconditions (58/60):** All preconditions describe states, not setup procedures. One borderline case in Step 3b ("PM has just unbound a MainItem") references a prior action but describes the resulting state.
- **No implementation coupling (58/60):** Steps use user-level verbs (click, hover, drag, press). HTML comments reference implementation details (milestone_key) for traceability, but these are annotations outside the step narrative.

### Precondition Exclusivity: 144/150

- **Distinct preconditions (58/60):** Edge cases are properly separated into distinct steps with distinct preconditions. No two outcomes within the same step share identical preconditions.
- **Sufficient for unique selection (48/50):** Each step's precondition clearly differentiates it from the happy path and from other edge cases. No ambiguous scenarios.
- **Missing preconditions (38/40):** Error/boundary outcomes state their triggers. Steps 4b-4d share "Quick-add dialog is open" as precondition but differ by action; this is acceptable since the action is the differentiator.

### Fact Alignment: 136/150

- **Factual claims traceable (50/60):** Key steps have `<!-- fact: ... -->` annotations tracing to PRD. Gap: Step E1 claims "panel retains its original state" after re-authentication, which is not verified in any source document and not marked UNKNOWN. Many steps (1b, 1d, 1e, 2, 4e, 4f, 7b, 5b) lack fact annotations despite being directly traceable to PRD Story 10.
- **Inferred claims (48/50):** Steps 4b-4d and E1 have proper `source: inferred` annotations citing the correct surface rule (validation-error, session-expired).
- **No hallucinated claims (38/40):** All unverified claims are classified as inferred. The session recovery behavior in E1 is speculative but annotated.

### Surface Fitness: 135/150

- **Mandatory outcomes (55/60):** Both validation-error and session-expired mandatory outcomes are present. Minor gap: validation only covers quick-add form; no validation scenarios for other potential form interactions.
- **Test strategy proportions (42/50):** Approximately 55/45 contract/journey split, slightly overweight on contract. Steps 4b-4d test the same form validation from different angles as separate steps, inflating the contract count.
- **Realistic assumptions (38/40):** All browser interactions (click, hover, drag, keyboard) are Playwright-compatible. No DOM structure or CSS class assertions. Focus management assertions are realistic.

### Internal Consistency: 126/150

- **Invariants hold (40/60):** Invariants 1-4 are verified across all relevant steps. Invariant 5 ("All mutation operations require milestone:update permission; controls hidden without permission") is only partially verified: Step 1f checks edit control and status badge but does not verify that unbind buttons (Step 2), add button (Step 4), or drag-rebind (Step 6) are hidden without permission.
- **Cross-step references (48/50):** All references between steps are unambiguous and consistent. No dangling references.
- **Risk level consistent (38/40):** "High" is justified by mutation operations, irreversible state transitions (cancelled auto-unbind), and data loss risk.

## Attacks

1. **[Completeness] No network-error boundary for API mutations:** Steps 3 (unbind), 4 (quick-add), and 6 (drag-rebind) all make API calls but have no failure variant. The surface-web rules list network-error as "Additional common Web boundary Outcome": "API request fails due to network issues. Assert: error message displayed, retry option available, no data loss." Quote from Step 3: "The MI is removed from the list." — assumes success. Must add steps for API/server error during each mutation operation.

2. **[Internal Consistency] Invariant 5 under-verified:** The invariant claims "All mutation operations (unbind, add, drag-rebind) require milestone:update permission; edit/delete controls are hidden without permission." But Step 1f only verifies: "Edit control is not displayed. Status badge is not interactive." The unbind control (Step 2), add button (Step 4), and drag-rebind (Step 6) are never tested under the no-permission precondition. Must extend Step 1f or add separate no-permission edge cases for each mutation control.

3. **[Fact Alignment] Unverified session recovery claim:** Step E1 states: "After re-authenticating, the panel retains its original state." No source document describes session recovery behavior. The surface rule says "unsaved data is either preserved or user is warned about data loss" — about data, not panel state. This specific behavioral claim should be marked UNKNOWN or verified against actual application behavior.

4. **[Fact Alignment] Inconsistent fact annotations:** Steps 1b, 1d, 1e, 2, 4e, 4f, 7b, and 5b all describe behavior directly from PRD Story 10 but lack `<!-- fact: ... -->` annotations, while Steps 1, 3, 4, 5b, 6b, 6c, 6d, and 7b have them. This inconsistency weakens the fact traceability principle. Add fact annotations to all PRD-derived steps.

5. **[Completeness] Step 5 lacks depth:** Quote: "**Expected Result**: Route navigates to the main item detail page." — A single sentence for a navigation action. No coverage for: MI deleted between panel load and click, target page returning 404, or navigation failure. Contrast with the depth of Steps 3 and 4 which have multiple edge cases.

6. **[Surface Fitness] Contract-heavy step structure:** Steps 4b, 4c, 4d are three separate steps testing the same quick-add form validation from different field angles. Per the 50/50 Web test strategy, these could be consolidated into a single step with multiple outcomes to better reflect journey-level testing depth.
