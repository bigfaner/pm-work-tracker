---
date: "2026-04-30"
doc_dir: "docs/proposals/e2e-test-scripts-rebuild/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 72/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  14      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   5/7    │          │            │
│    Urgency justified         │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  14      │  20      │ ⚠️          │
│    Approach concrete         │   6/7    │          │            │
│    User-facing behavior      │   4/7    │          │            │
│    Differentiated            │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  11      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   3/5    │          │            │
│    Rationale justified       │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  12      │  15      │ ⚠️          │
│    In-scope concrete         │   4/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  11      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   3/5    │          │            │
│    Mitigations actionable    │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  10      │  15      │ ⚠️          │
│    Measurable                │   3/5    │          │            │
│    Coverage complete         │   3/5    │          │            │
│    Testable                  │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  72      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Problem §Evidence | "scripts may not conform" — conformance gap is speculative; no audit performed | -2 pts (evidence) |
| Problem §Urgency | Generic drift argument with no concrete example of a missed bug or measured staleness | -3 pts (urgency) |
| Solution §User-facing | End state for the developer running regression is never described | -3 pts (user-facing behavior) |
| Alternatives §Pros/cons | "Do nothing" pros listed as only "No effort"; ignores real pro of zero regression-suite risk | -2 pts (honest trade-offs) |
| Alternatives §Rationale | Proposed approach cons: "More effort" — no estimate, no ROI argument | -2 pts (rationale) |
| Scope §Bounded | No timeline, no effort estimate, "in batches" undefined | -2 pts (bounded) |
| Risk §Likelihood | Likelihood/impact scale undefined; "Medium"/"Low" are unanchored | -2 pts (rating honesty) |
| Risk §Mitigation | "Review generated scripts against current codebase before graduating" — no criteria, no owner | -2 pts (actionability) |
| Success §Measurable | "conforming to forge standards" — no linter, no checklist, not objectively verifiable | -2 pts (measurable) |
| Success §Coverage | No criterion verifying all test-cases.md entries are represented in generated scripts | -2 pts (coverage) |

---

## Attack Points

### Attack 1: Problem Definition — urgency is asserted, not demonstrated

**Where**: "The longer graduation is deferred, the more the scripts drift from the current codebase."

**Why it's weak**: This is a generic truism that applies to any deferred work. There is no concrete example of a bug that slipped through because these scripts weren't in regression, no measurement of how stale the existing scripts already are, and no data on how long they have been stranded. The urgency section reads as filler rather than a case for acting now versus next quarter.

**What must improve**: Provide at least one concrete data point — e.g., a specific bug that was caught late because these tests weren't running, or a diff showing how many test-cases.md entries reference APIs that have already changed. If no such data exists, state the risk quantitatively: "X of 11 features have had backend changes since their scripts were written."

---

### Attack 2: Solution Clarity — end state for the developer is never described

**Where**: "Graduate scripts using `/graduate-tests` to move them into `tests/e2e/<target>/` with proper imports and graduation markers."

**Why it's weak**: The proposal describes the mechanical steps (regenerate, graduate, markers) but never answers: what does a developer experience after this work is done? How many tests will be in `tests/e2e/`? What command runs them? What does a passing regression run look like? The "user" of this work is the developer running CI or the regression suite, and their experience is completely absent.

**What must improve**: Add a concrete end-state description: the command to run the full regression suite, the expected test count range, and what a clean run output looks like. Even one sentence — "After graduation, `node tests/e2e/run-all.js` will execute ~N tests covering these 11 feature areas" — would anchor the proposal.

---

### Attack 3: Risk Assessment — highest-impact mitigation is not actionable

**Where**: "Some test cases reference APIs or UI flows that no longer exist | Medium | Scripts have dead assertions | Review generated scripts against current codebase before graduating"

**Why it's weak**: This is the second-highest-impact risk (Medium likelihood, dead assertions entering the regression suite and producing permanent noise). The mitigation "Review generated scripts against current codebase before graduating" has no criteria for what passes review, no owner, no time estimate, and no definition of what "current codebase" means in practice. It is indistinguishable from "we'll handle it."

**What must improve**: Make the mitigation concrete: specify who reviews (the author, a second engineer), what they check (each assertion maps to an existing API endpoint or UI element in sitemap.json), and what the exit criterion is (zero dead-endpoint assertions, or a documented exception list). Alternatively, propose an automated check — e.g., run `/run-e2e-tests` on each feature's scripts before graduating and require a minimum pass rate.

---

### Attack 4: Alternatives Analysis — pros/cons are straw-man thin

**Where**: "Do nothing | No effort | 11 features unprotected by regression; scripts rot further | Rejected"

**Why it's weak**: The "do nothing" pros column lists only "No effort" — this ignores the real advantage of do-nothing: zero risk of introducing non-conformant or broken tests into the regression suite. The proposed approach cons column says only "More effort" with no estimate of how much effort. These are not honest trade-offs; they are constructed to make the chosen approach look obviously correct.

**What must improve**: For "do nothing," acknowledge the real pro: no regression-suite pollution risk. For the proposed approach, quantify the effort cost — even a rough estimate ("~2 hours per feature × 11 features = ~22 hours") makes the ROI argument credible.

---

### Attack 5: Success Criteria — "conforming to forge standards" is not verifiable

**Where**: "All 11 features have scripts regenerated via `/gen-test-scripts` conforming to forge standards"

**Why it's weak**: "Conforming to forge standards" cannot be objectively verified. There is no linter, no automated check, and no checklist defined in the proposal. Two engineers could disagree on whether a script conforms. This criterion is the first one listed and the most important — it is the entire justification for regenerating rather than graduating as-is — yet it has no verification mechanism.

**What must improve**: Define what "conforming to forge standards" means in checkable terms: e.g., "uses Playwright for UI tests, `fetch` for API tests, `child_process` for CLI tests; imports only from `node:test` and `node:assert`; no `describe`/`it` from external frameworks." These checks can be automated with a grep or a linter rule.

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 72/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Iteration 1 complete. Key weaknesses: urgency lacks evidence, end-state behavior is absent, highest-impact risk mitigation is vague, alternatives trade-offs are thin, and the primary success criterion is not objectively verifiable.
