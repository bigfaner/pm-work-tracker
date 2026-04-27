# Evaluation Report: Integration Test Coverage Proposal

**Document:** `docs/proposals/integration-test-coverage/proposal.md`
**Date:** 2026-04-27
**Iteration:** 2

---

## Changes Since Iteration 1

All three iteration-1 attacks were addressed:

1. **Alternatives Analysis:** "Do nothing" now lists honest pros. All alternatives include effort estimates. The verdict is explicit ("We chose flow-based integration tests because...").
2. **Risk Assessment:** Test brittleness risk added. Data isolation mitigation commits to transaction rollback (not "or"). Severity threshold for exposed bugs is now defined.
3. **Problem Definition / Urgency:** Concrete triggering incident added: commit `1883499` timezone bug in `view_handler.go`, with reference to `docs/lessons/weekly-view-bug-fixes.md` (Bugs 2 and 3).

Score is reassessed from scratch below. No credit is given for "improvement" -- only for the current state of the text.

---

## Dimension 1: Problem Definition (18/20)

### Problem stated clearly: 6/7

The core problem is quantified and unambiguous: "54 API endpoints but only 18 (33%) have integration tests." The domain-by-domain gap table (8 rows with total/tested/gap columns) leaves no room for misinterpretation. Two readers would arrive at the same understanding.

Deduction (-1): The problem still conflates two distinct concerns -- integration test gaps and unit test gaps -- in a single sentence: "Several unit test gaps also exist (`permission_handler.go` completely untested, `ConvertToMain`/`UpdateTeam`/`GetByBizKey` methods lack coverage)." The word "also" tacks on unit test gaps as an afterthought without framing their relationship to the integration test problem. Are these equally urgent? A prerequisite? A stretch goal? The reader must infer the prioritization.

### Evidence provided: 6/7

Strong quantitative evidence. The domain table is specific, verifiable, and actionable. Unit test gaps name exact files and methods. The triggering incident in urgency (commit hash, bug descriptions, cross-reference to `docs/lessons/weekly-view-bug-fixes.md`) is concrete and traceable.

Deduction (-1): The source of the endpoint count (54 total, 18 tested) is still not cited. Was this derived from a coverage tool? A manual `grep` of route registrations? An audit of `*_test.go` files? Without provenance, the reader cannot independently verify or reproduce the baseline. A single sentence like "Endpoint counts derived from `grep -r 'r.GET\|r.POST\|r.PUT\|r.DELETE\|r.PATCH' backend/internal/handler/`" would resolve this.

### Urgency justified: 6/6

The urgency section is now anchored in a concrete incident: commit `1883499` introduced a timezone bug in `view_handler.go` and a filtering logic bug in `view_service.go`, both of which escaped to manual testing because the Views domain had zero integration tests. The cross-reference to `docs/lessons/weekly-view-bug-fixes.md` (Bugs 2 and 3) is specific and verifiable. The cost of inaction is quantified: "each escaped regression costs ~2-4 hours of manual diagnosis."

No deduction. This is strong.

---

## Dimension 2: Solution Clarity (16/20)

### Approach is concrete: 7/7

The user-flow organization is well-explained. Six test files are named and specified at the individual endpoint and test-case level. The tables for F1-F6 list every endpoint, every test scenario (happy path, validation errors, permission denied, cascading effects), and every expected HTTP status code. A developer could start implementing immediately from these tables.

The "shared test helpers" from iteration 1 is now in scope as a deliverable ("Shared test helpers extracted from existing integration tests for reuse"). While still not described in detail, the rest of the solution is sufficiently concrete to stand on its own.

No deduction.

### User-facing behavior described: 5/7

The test descriptions specify HTTP status codes (201, 200, 403, 404, 422) and cascading behaviors (progress append to completion rollup, status change to sub-item cascade). These are observable, testable behaviors.

Deduction (-1): The proposal describes test cases, not user-facing behavior. For each flow, the reader sees a list of endpoints and assertions but no narrative of the user experience. A brief sentence per flow like "A PM creates a main item, breaks it into sub-items, assigns them to team members, and tracks completion through progress updates" would connect the technical test cases to the user workflow they protect.

Deduction (-1): F5 (Views & Reports) still does not define correctness. "Weekly stats (NEW/completed/in-progress/overdue counts), comparison with previous week, delta badges, empty data" lists what the response contains but not what correct means. What is the expected weekly stat for a team with 3 items created and 1 completed? What does "comparison with previous week" assert? The F1-F4 flows benefit from clear business logic (create item -> 201, duplicate -> 409) but F5 lacks similar precision.

### Distinguishes from alternatives: 4/6

The flow-based approach is clearly differentiated from endpoint-isolated testing (Alternative A). The inter-endpoint bug argument is now supported by a specific example: "a `ConvertToMain` flow spans pool submission, sub-item creation, and pool status update -- isolated tests for each endpoint would all pass even if the handoff between them breaks."

Deduction (-2): The differentiator from contract testing (Alternative C) is weak: "Overkill for a single-service application." This is an opinion, not an analysis. The proposal does not explain why contract tests cannot complement integration tests, or what specific scenarios integration tests catch that contract tests would miss (or vice versa). The "single consumer" argument is valid but underdeveloped.

---

## Dimension 3: Alternatives Analysis (12/15)

### At least 2 alternatives listed: 5/5

Three alternatives: endpoint-isolated (A), do nothing (B), contract tests (C). Breadth is good. All three are legitimate approaches to the test coverage problem.

### Pros/cons for each: 4/5

All alternatives now have both pros and cons. Alternative B (do nothing) lists honest pros: "Zero immediate investment. Team continues shipping features without test-maintenance overhead. CI time stays the same. No risk of flaky integration tests blocking PRs." This is a fair treatment, not a straw man.

Deduction (-1): The pros/cons for Alternative A (endpoint-isolated) are not fully grounded. It says "~15 min/test case because no setup chaining is needed" -- but this estimate is asserted without justification. How was 15 minutes derived? Is it from the team's experience with the existing 18 integration tests? A benchmark from another project? A rough guess? Effort estimates for B and C have similar provenance issues, though B's "~2-4 hours per escaped regression" is anchored to the weekly-view incident timeline.

### Rationale for chosen approach: 3/5

The verdict is now explicit: "We chose flow-based integration tests because they catch both inter-endpoint wiring bugs and business logic errors against a real database, with no new tooling dependencies."

Deduction (-2): The verdict still does not address a hybrid approach. The choice is presented as binary (flow-based vs. endpoint-isolated) when a blended strategy -- flow-based for multi-step critical paths (item lifecycle, item pool) and endpoint-isolated for simple CRUD (team management, admin users) -- could be more efficient. The proposal estimates ~40 developer-hours for flow-based. A hybrid might reduce this. The analysis does not consider this middle ground, and the verdict does not justify why pure flow-based is superior to a mix.

---

## Dimension 4: Scope Definition (14/15)

### In-scope items are concrete: 5/5

Five named integration test files with specific endpoint and test-case tables. Six named unit test gaps with file and method names. Shared test helpers as a deliverable. Each item is a deliverable, not a vague area. The "exhaustive edge cases" language from iteration 1 has been replaced with "Edge cases per endpoint: happy path, validation errors, permission denied, not found, cascading effects (see test tables for specific cases per endpoint)" -- which is concrete and points to the detailed tables.

### Out-of-scope explicitly listed: 5/5

Four clear exclusions: frontend tests, performance/load testing, E2E browser testing, new features/bug fixes. Each is named and justified.

### Scope is bounded: 4/5

36 endpoints across 5 files, 6 unit test gaps, shared helpers. A team can execute this. The risk section mentions "one flow file per PR" which implies incremental delivery.

Deduction (-1): No explicit ordering or phasing. Is F1 first because the item lifecycle is the most critical domain? Is F6 (unit gaps) lower priority than F1-F5? The proposal does not state a recommended implementation order. For a team executing this, phasing guidance would help prioritize if time runs short.

---

## Dimension 5: Risk Assessment (14/15)

### Risks identified: 5/5

Five meaningful risks: test execution time, test data pollution, edge cases exposing existing bugs, large PR size, and test brittleness from response-schema coupling. All are specific to integration test coverage efforts. The test brittleness risk from iteration 1 is now present and well-described.

### Likelihood + impact rated: 4/5

All five risks have likelihood and impact ratings. The ratings now show range: Low (impact for exposed bugs), Medium (execution time, PR size, brittleness), High (likelihood for exposed bugs, impact for data pollution). The "all Medium/High" problem from iteration 1 is resolved.

Deduction (-1): Impact ratings are still single words without quantitative thresholds. "Medium" impact for "test execution time" -- does this mean 30s? 60s? 120s? The mitigation says "Target <30s per file" but this target is not connected to the impact rating. A reader cannot tell whether hitting 45s per file would be a "High" or "Medium" impact. One or two sentences mapping ratings to concrete thresholds would make the assessment more honest.

### Mitigations are actionable: 5/5

All mitigations are now specific and actionable:

- "Each flow file is independent; can run in parallel. Target <30s per file."
- "Use transaction rollback (`tx.Begin()` at test start, `tx.Rollback()` in `t.Cleanup`) to guarantee no persisted state. This is preferred over manual cleanup because it is automatic and cannot be forgotten."
- "if the bug causes data loss, incorrect business state (e.g., wrong status transition accepted), or auth bypass, fix immediately in the same PR. Otherwise, file a bug and continue"
- "Submit one flow file per PR for incremental review."
- "Assert on structural fields (status code, top-level keys, specific business fields like `itemStatus`) but avoid asserting on field ordering, error message text, or pagination metadata format."

Each mitigation tells a developer exactly what to do. The transaction rollback vs. cleanup ambiguity from iteration 1 is resolved. The severity threshold is defined.

---

## Dimension 6: Success Criteria (12/15)

### Criteria are measurable: 4/5

Strong measurable criteria: "All 54 API endpoints have at least one integration test." "Total integration test count >= 150 new test cases." "`go test ./tests/integration/...` passes with 0 failures."

Deduction (-1): "Cascading effects are verified" is still not independently measurable. What counts as "verified"? An assertion on a response field? A database state check? The criterion should specify what verification means, e.g., "Each cascading effect has at least one test that asserts on the downstream entity's state after the triggering action."

### Coverage is complete: 4/5

The criteria cover endpoint count, per-endpoint status code coverage, status transitions, cascading effects, unit test gaps, total test count, and CI pass.

Deduction (-1): No criterion for test execution time. The risk section targets "<30s per file" but this target does not appear in Success Criteria. If it matters enough to be a risk mitigation target, it should be a success criterion. The total test suite time is a deliverable quality metric.

### Criteria are testable: 4/5

Most criteria are testable: count endpoints, count test cases, run the suite, check for public methods.

Deduction (-1): "All unit test gaps closed (0 untested public methods in service/handler layers)" is still ambiguous about scope. Does this mean ALL public methods across the entire codebase, or only the 6 named gaps in F6? The scope section says "6 unit test gap fixes" but the success criterion says "0 untested public methods in service/handler layers" which is a much broader statement. If the criterion means "the 6 gaps listed in F6 are resolved," it should say so. If it means a full audit, it is scope creep beyond what is in scope.

---

## Vague Language Penalty

Scanning for unquantified "better," "improved," "enhanced":

No instances of these specific words found. The proposal has been cleaned up since iteration 1.

However, one instance of vague language remains:

1. The scope section says "frontend test suite already covers component and E2E flows" -- the word "covers" is not quantified. What coverage percentage? How many test cases? This is in an out-of-scope justification, so it is a minor issue, but it still uses unquantified language to justify an exclusion.

**Total vague language penalty: -2** (one instance of unquantified language in a scope justification)

---

## Inconsistency Check

The in-scope section says "5 integration test files covering 36 untested endpoints." The Solution section describes 5 flow files + 1 unit gap file = 6 files total. The scope bullet says "5 integration test files" which is correct (F6 is unit tests, not integration tests). Consistent.

The success criteria say "All 54 API endpoints" but the solution covers 36 new + 18 existing = 54 total. The proposal never explicitly states this arithmetic, but it is inferable. No material inconsistency.

The success criterion "0 untested public methods in service/handler layers" is broader than the in-scope "6 unit test gap fixes." This is a mild inconsistency -- the success criterion promises more than the scope delivers. However, since this was noted in iteration 1 and not addressed, applying penalty.

**Inconsistency penalty: -3** (success criterion "0 untested public methods" is broader than scope "6 unit test gap fixes")

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 18 | 20 |
| Solution Clarity | 16 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 12 | 15 |
| Vague language penalty | -2 | -- |
| Inconsistency penalty | -3 | -- |
| **Total** | **81** | **100** |

---

SCORE: 81/100

DIMENSIONS:
- Problem Definition: 18/20
- Solution Clarity: 16/20
- Alternatives Analysis: 12/15
- Scope Definition: 14/15
- Risk Assessment: 14/15
- Success Criteria: 12/15

ATTACKS:
1. Success Criteria: Criterion "0 untested public methods in service/handler layers" overpromises relative to scope -- the scope commits to "6 unit test gap fixes" but the success criterion promises zero untested public methods across the entire service and handler layers, which is a significantly broader promise. Either the scope must be expanded to include a full audit, or the criterion must be rewritten to reference the specific 6 gaps listed in F6.
2. Alternatives Analysis: Hybrid approach not considered -- the verdict says "We chose flow-based integration tests" as an all-or-nothing decision, but a blended strategy (flow-based for multi-step paths like item lifecycle and item pool, endpoint-isolated for simple CRUD like admin users) could reduce the ~40 developer-hour estimate. The analysis presents a false binary without justifying why pure flow-based is superior to a mix tailored to endpoint complexity.
3. Solution Clarity: F5 (Views & Reports) lacks correctness definitions -- "Weekly stats (NEW/completed/in-progress/overdue counts), comparison with previous week, delta badges, empty data" describes what the response contains but not what correct output looks like. Unlike F1-F4 where business logic is precise (create item -> 201, duplicate -> 409), F5 has no expected values or assertions specified, making it impossible for a developer to know what "correct" weekly stats or CSV export looks like.
