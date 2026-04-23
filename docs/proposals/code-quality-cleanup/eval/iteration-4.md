---
date: "2026-04-22"
doc_dir: "Z:\\project\\ai\\coding-harness\\pm-work-tracker\\docs\\proposals\\code-quality-cleanup"
iteration: "4"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 4

**Score: 94/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  19      │  20      │ ✅         │
│    Problem clarity           │  7/7     │          │            │
│    Evidence provided         │  6/7     │          │            │
│    Urgency justified         │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  20      │  20      │ ✅         │
│    Approach concrete         │  7/7     │          │            │
│    User-facing behavior      │  7/7     │          │            │
│    Differentiated            │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  14      │  15      │ ✅         │
│    Alternatives listed (≥2)  │  5/5     │          │            │
│    Pros/cons honest          │  4/5     │          │            │
│    Rationale justified       │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  14      │  15      │ ✅         │
│    In-scope concrete         │  5/5     │          │            │
│    Out-of-scope explicit     │  5/5     │          │            │
│    Scope bounded             │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ 5. Risk Assessment           │  15      │  15      │ ✅         │
│    Risks identified (≥3)     │  5/5     │          │            │
│    Likelihood + impact rated │  5/5     │          │            │
│    Mitigations actionable    │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ 6. Success Criteria          │  15      │  15      │ ✅         │
│    Measurable                │  5/5     │          │            │
│    Coverage complete         │  5/5     │          │            │
│    Testable                  │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ TOTAL                        │  97-3    │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Lines 169 vs 178 | **Inconsistency**: Scope explicitly states "Test coverage expansion (existing tests must pass, but no new test files)" as out-of-scope, but Risk Assessment mitigation for Phase 2 says "Write integration tests for each converted endpoint covering edge cases" — these directly contradict | -3 pts |

---

## Attack Points

### Attack 1: Scope/Risk Inconsistency — "no new test files" contradicts "write integration tests"

**Where**: Line 169: "Test coverage expansion (existing tests must pass, but no new test files)" vs. Line 178: "Write integration tests for each converted endpoint covering edge cases (empty results, null fields, multi-page results) before removing in-memory logic"
**Why it's weak**: The scope section explicitly excludes writing any new test files. Yet the most important risk mitigation for Phase 2 (SQL-level filtering bugs) depends on writing integration tests. A developer executing this proposal cannot simultaneously follow both instructions. Either the scope must allow targeted integration tests for converted endpoints, or the risk mitigation must be achievable without new test files (e.g., "manually verify edge cases against staging data"). This contradiction creates ambiguity that could lead to either skipped testing or scope creep disputes.
**What must improve**: Resolve the contradiction. Either (a) revise the out-of-scope to allow integration tests specifically for endpoints whose filtering logic changes in Phase 2, with a qualifier like "no new test files except integration tests for Phase 2 SQL conversion endpoints," or (b) replace the risk mitigation with a verification strategy that does not require new test files (e.g., manual test plan, staging environment checklist).

### Attack 2: Alternatives Analysis — Pros/cons remain partially tautological despite coverage column addition

**Where**: Lines 144-150: The alternatives table. For example, "Frontend-only cleanup" pros: "Immediate UI dev velocity improvement; lower regression surface" and cons: "Leaves backend N+1 and dead code; contract mismatches remain; no performance improvement"
**Why it's weak**: While the addition of an Effort column and Coverage percentages is a genuine improvement, the Pros and Cons columns for several alternatives still describe the consequences of scope inclusion/exclusion rather than analytical trade-offs. "Lower regression surface" for frontend-only is the obvious result of touching fewer files. "Leaves backend N+1 and dead code" is just describing what frontend-only means. Compare to the all-at-once row where "Single context switch; completes faster wall-clock" and "Very large diff; harder to review; higher regression risk" — these are genuine analytical observations about *why* the approach has trade-offs, not restatements of its scope. The frontend-only and backend-only rows lack this analytical depth.
**What must improve**: Rewrite pros/cons for at least 2 alternatives to include non-obvious trade-offs. For frontend-only: what *would* it enable (e.g., "could ship before the new devs arrive if backend work is deferred to next quarter")? For backend-only: what specific risk does it reduce (e.g., "avoids the frontend contract renegotiation coordination cost")? Each pro/con should pass the "would a naive reader already know this from the approach name?" test.

### Attack 3: Problem Definition — "slows development velocity" claim still lacks developer evidence

**Where**: Line 20: "This debt slows development velocity, makes bugs harder to find, and will cause performance degradation as data volume grows"
**Why it's weak**: The performance claim is now well-supported with p95 measurements. The urgency is well-supported with onboarding and user reports. But the core claim about developer velocity remains unsubstantiated. "Slows development velocity" and "makes bugs harder to find" are asserted without evidence. No PR review duration trends, no developer complaints, no measured time-to-fix metrics, no examples of bugs that were hard to find because of the debt. The reader has to take this on faith, while every other claim in the problem section is now concretely evidenced. The asymmetry is noticeable.
**What must improve**: Add one specific piece of evidence for the velocity claim. Even a single example suffices: "A recent PR to fix the status transition bug required touching 3 files and reviewing 900+ lines in ItemViewPage.tsx to locate the relevant logic" or "Average PR review time for files in the monolithic pages has increased from 1 day to 3 days over the past 2 months." Without this, the velocity claim reads as padding around an otherwise well-evidenced problem statement.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Problem Definition - Evidence lacks performance data and urgency lacks triggering event | ✅ Yes | Line 15 now includes measured p95 times: "current p95 response time for `ItemPoolHandler.List` with ~200 records is 1.4s; `ProgressHandler.List` p95 is 1.1s with ~150 records; both exceed the 200ms target by 5x." Line 20 adds two urgency triggers: "the team is onboarding 2 new developers next sprint" and "a recent user report flagged slow list-page loads on datasets exceeding 100 items." |
| Alternatives Analysis - Pros/cons are tautological, not analytical | Partially | Effort and Coverage columns added (lines 144-150). Rationale paragraph added (line 152). However, individual pros/cons for frontend-only and backend-only rows remain largely tautological. Score improved from 3/5 to 4/5. |
| Scope Definition - No sequencing constraints or parallelization guidance | ✅ Yes | Line 26 now contains a full paragraph on phase dependencies and parallelization: "Phase 2 requires Phase 1 backend work complete... Phase 3 frontend deduplication can begin after Phase 1 frontend work ships... With two developers, the backend dev leads Phase 1 backend → Phase 2 → Phase 3 backend → Phase 4 backend, while the frontend dev starts Phase 1 frontend concurrently..." The gap for the frontend dev is quantified ("3-4 day gap"). Score improved from 4/5 to 4/5 on boundedness (still no calendar milestones). |

---

## Verdict

- **Score**: 94/100
- **Target**: 90/100
- **Gap**: -4 (exceeds target)
- **Action**: Target reached. The proposal exceeds the 90-point threshold. The remaining weaknesses (scope/risk inconsistency, partially tautological alternatives, unsupported velocity claim) are minor and do not prevent the proposal from being actionable.
