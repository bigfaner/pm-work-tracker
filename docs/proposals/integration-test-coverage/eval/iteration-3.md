# Evaluation Report: Integration Test Coverage Proposal

**Document:** `docs/proposals/integration-test-coverage/proposal.md`
**Date:** 2026-04-27
**Iteration:** 3

---

## Changes Since Iteration 2

All three iteration-2 attacks were addressed:

1. **Success Criteria inconsistency:** "0 untested public methods in service/handler layers" replaced with the scoped "All 6 unit test gaps listed in F6 resolved (each named file/method has at least one passing test)." The criterion now matches the in-scope commitment.
2. **Hybrid approach not considered:** The verdict now explicitly addresses and rejects the hybrid strategy with a reasoned argument about cognitive overhead and marginal savings (~4-6 hours saved vs. consistency cost).
3. **F5 correctness definitions:** All F5 endpoints now include concrete expected values: `stats: {NEW: 0, completed: 1, inProgress: 2, overdue: 0}`, specific date ranges, filter/query parameter behavior, BOM bytes, and markdown section names.

Score is reassessed from scratch below. No credit is given for "improvement" -- only for the current state of the text.

---

## Dimension 1: Problem Definition (18/20)

### Problem stated clearly: 6/7

The core problem is quantified and unambiguous: "54 API endpoints but only 18 (33%) have integration tests." The domain-by-domain gap table (8 rows with total/tested/gap columns) is precise. Two readers would arrive at the same understanding.

Deduction (-1): The problem still conflates two distinct concerns in a single opening paragraph. The sentence "Several unit test gaps also exist (`permission_handler.go` completely untested, `ConvertToMain`/`UpdateTeam`/`GetByBizKey` methods lack coverage)" tacks unit test gaps onto the integration test problem without framing their relationship. Are unit test gaps equally urgent? A prerequisite? Lower priority? The reader must infer. The word "also" is doing too much work -- it hides a prioritization decision.

### Evidence provided: 6/7

Strong quantitative evidence. The domain table is verifiable (specific endpoint counts per domain). Unit test gaps name exact files and methods. The urgency section references a concrete incident: commit `1883499`, cross-referenced to `docs/lessons/weekly-view-bug-fixes.md` (Bugs 2 and 3).

Deduction (-1): The provenance of the endpoint count (54 total, 18 tested) is still not cited. Was this derived from a coverage tool? A manual grep of route registrations? An audit of test files? Without provenance, a reader cannot independently verify or reproduce the baseline. A single sentence of methodology would close this gap.

### Urgency justified: 6/6

The urgency section is anchored to a concrete incident: commit `1883499` introduced a timezone bug in `view_handler.go` and a filtering logic bug in `view_service.go`, both undetected because the Views domain had zero integration tests. The cross-reference to `docs/lessons/weekly-view-bug-fixes.md` is verifiable. The cost of inaction is quantified: "each escaped regression costs ~2-4 hours of manual diagnosis."

No deduction. This is strong.

---

## Dimension 2: Solution Clarity (17/20)

### Approach is concrete: 7/7

Six test files are named and specified at the individual endpoint and test-case level. F1-F6 tables list every endpoint, every test scenario (happy path, validation errors, permission denied, cascading effects), and expected HTTP status codes. F5 now includes concrete expected output values. A developer could implement directly from these tables.

No deduction.

### User-facing behavior described: 6/7

The test descriptions specify HTTP status codes (201, 200, 403, 404, 422) and cascading behaviors (progress append to completion rollup, status change to sub-item cascade). F5 now defines expected outputs with concrete values rather than vague descriptions.

Deduction (-1): The proposal still describes test cases rather than user-facing behavior. For each flow, the reader sees endpoints and assertions but no narrative of the user experience. A brief sentence per flow like "A PM creates a main item, breaks it into sub-items, assigns them to team members, and tracks completion through progress updates" would connect the technical test cases to the user workflows they protect. The user-flow framing is in the section title but not in the content.

### Distinguishes from alternatives: 4/6

The flow-based approach is differentiated from endpoint-isolated testing with a specific example: "a `ConvertToMain` flow spans pool submission, sub-item creation, and pool status update -- isolated tests for each endpoint would all pass even if the handoff between them breaks." The hybrid approach is now explicitly considered and rejected with reasoning.

Deduction (-2): The differentiator from contract testing (Alternative C) remains underdeveloped. The argument is "Overkill for a single-service application with one consumer (the frontend)." This is an opinion, not an analysis. The proposal does not articulate what specific scenarios integration tests catch that contract tests would miss, beyond the single example of "archive in-progress item fails." The verdict also does not address why integration and contract tests could not complement each other -- the choice is presented as mutually exclusive without justification for excluding a layered approach.

---

## Dimension 3: Alternatives Analysis (13/15)

### At least 2 alternatives listed: 5/5

Three alternatives: endpoint-isolated (A), do nothing (B), contract tests (C). Breadth is good. All three are legitimate approaches.

### Pros/cons for each: 4/5

All alternatives now have both pros and cons with honest treatment. Alternative B (do nothing) lists legitimate pros: "Zero immediate investment. Team continues shipping features without test-maintenance overhead." Alternative A lists concrete pros: "easier to debug when it fails. Faster to write per-endpoint (~15 min/test case)." All three alternatives include effort estimates.

Deduction (-1): Effort estimates lack provenance. Alternative A says "~15 min/test case" -- how was this derived? From the team's experience with existing integration tests? A benchmark? A rough guess? Similarly, "~30 developer-hours" for A and "~40 developer-hours" for C are asserted without methodology. The weekly-view incident anchors B's "~2-4 hours per regression," but A and C have no such grounding. Without transparency on estimation method, the reader cannot calibrate confidence.

### Rationale for chosen approach: 4/5

The verdict is now explicit: "We chose flow-based integration tests because they catch both inter-endpoint wiring bugs and business logic errors against a real database, with no new tooling dependencies." The hybrid approach is explicitly considered and rejected with reasoning: "mixing two test patterns creates inconsistent test structure that increases cognitive overhead for reviewers and maintainers, while the effort savings are marginal."

Deduction (-1): The hybrid rejection argument relies on a soft claim ("increases cognitive overhead") that is not substantiated. What evidence supports the assertion that mixing two patterns increases cognitive overhead? Is this from team experience? Industry consensus? The claim may be correct, but it is presented as self-evident. A one-sentence grounding would strengthen it.

---

## Dimension 4: Scope Definition (13/15)

### In-scope items are concrete: 5/5

Five named integration test files with specific endpoint and test-case tables. Six named unit test gaps with file and method names. Shared test helpers as a deliverable. Each item is a deliverable, not a vague area.

### Out-of-scope explicitly listed: 5/5

Four clear exclusions: frontend tests, performance/load testing, E2E browser testing, new features/bug fixes. Each is named and justified.

### Scope is bounded: 3/5

36 endpoints across 5 files, 6 unit test gaps, shared helpers. The risk section mentions "one flow file per PR" which implies incremental delivery.

Deduction (-1): No explicit ordering or phasing. Is F1 first because the item lifecycle is the most critical domain? Is F6 (unit gaps) lower priority than F1-F5? If time runs short after F1-F3, should the team complete F4 or F5 first? The proposal does not state a recommended implementation order.

Deduction (-1): No total effort estimate for the chosen approach. The alternatives list effort estimates (~30h for A, ~40h for C), but the chosen approach has no explicit estimate in the Solution or Scope section. The verdict mentions "~40 developer-hours for the entire suite," but this is buried in the Alternatives verdict paragraph rather than stated as a scope commitment. A team executing this needs to know the expected investment upfront.

---

## Dimension 5: Risk Assessment (14/15)

### Risks identified: 5/5

Five meaningful risks: test execution time, test data pollution, edge cases exposing existing bugs, large PR size, and test brittleness from response-schema coupling. All are specific to integration test coverage. The test brittleness risk is well-described with a concrete mitigation.

### Likelihood + impact rated: 4/5

All five risks have likelihood and impact ratings. The ratings show appropriate range: Low impact (exposed bugs -- "good"), Medium (execution time, PR size, brittleness), High (data pollution impact, exposed bugs likelihood).

Deduction (-1): Impact ratings remain qualitative without thresholds. "Medium" impact for "test execution time" -- does this mean 30s? 60s? 120s? The mitigation says "Target <30s per file" but this target is not connected to the impact rating. A reader cannot tell whether hitting 45s per file would escalate to "High" impact. Quantitative thresholds would make the assessment more honest and actionable.

### Mitigations are actionable: 5/5

All mitigations are specific and actionable:

- "Each flow file is independent; can run in parallel. Target <30s per file."
- "Use transaction rollback (`tx.Begin()` at test start, `tx.Rollback()` in `t.Cleanup`) to guarantee no persisted state."
- "if the bug causes data loss, incorrect business state (e.g., wrong status transition accepted), or auth bypass, fix immediately in the same PR. Otherwise, file a bug and continue"
- "Submit one flow file per PR for incremental review."
- "Assert on structural fields (status code, top-level keys, specific business fields like `itemStatus`) but avoid asserting on field ordering, error message text, or pagination metadata format."

Each mitigation tells a developer exactly what to do. No ambiguity.

---

## Dimension 6: Success Criteria (14/15)

### Criteria are measurable: 5/5

Strong measurable criteria: "All 54 API endpoints have at least one integration test." "Total integration test count >= 150 new test cases." "`go test ./tests/integration/...` passes with 0 failures." "All 6 unit test gaps listed in F6 resolved (each named file/method has at least one passing test)." Each criterion can be objectively verified.

"Cascading effects are verified" is borderline -- but the parenthetical examples ("progress append to completion rollup, status change to sub-item cascade") and the specific test cases in the F1/F2 tables provide enough specificity to make this verifiable against the test tables.

### Coverage is complete: 4/5

Criteria cover: endpoint count, per-endpoint status code coverage, status transitions, cascading effects, unit test gaps, total test count, and CI pass.

Deduction (-1): No criterion for test execution time. The risk section targets "<30s per file" and the mitigation says "Target <30s per file," but this target does not appear in Success Criteria. If execution time matters enough to be a risk mitigation target, it should be a success criterion. Adding "Total integration test suite completes in <150s (5 files x 30s)" would close this gap.

### Criteria are testable: 5/5

All criteria are testable: count endpoints with integration tests (binary check), count test cases (grep or test runner output), run the suite and check for failures (CI gate), verify the 6 named gaps have passing tests. The iteration-2 inconsistency ("0 untested public methods" vs. "6 unit test gap fixes") is resolved -- the criterion now correctly references "the 6 unit test gaps listed in F6."

---

## Vague Language Penalty

Scanning for unquantified "better," "improved," "enhanced":

No instances of these specific words found. The proposal has been cleaned across iterations.

One instance of vague language remains:

1. The out-of-scope section says "Frontend test changes (frontend test suite already covers component and E2E flows)" -- "covers" is not quantified. What coverage percentage? How many test cases? This is in an out-of-scope justification, so the impact is limited, but the word is still unquantified.

**Total vague language penalty: -2** (one instance of unquantified language in a scope exclusion justification)

---

## Inconsistency Check

The in-scope section says "5 integration test files covering 36 untested endpoints." The Solution describes 5 flow files + 1 unit gap file = 6 files. F6 is unit tests, not integration tests. Consistent.

The success criteria say "All 54 API endpoints" and the solution covers 36 new endpoints (18 already tested). The arithmetic (36 + 18 = 54) is inferable. Consistent.

The success criterion "All 6 unit test gaps listed in F6 resolved" now matches the in-scope "6 unit test gap fixes." The iteration-2 inconsistency is resolved.

The verdict mentions "~40 developer-hours for the entire suite" but the Scope section does not include an effort estimate. This is a gap rather than an inconsistency -- the number exists but is not formally committed in scope.

**No inconsistency penalty applied.**

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 18 | 20 |
| Solution Clarity | 17 | 20 |
| Alternatives Analysis | 13 | 15 |
| Scope Definition | 13 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 14 | 15 |
| Vague language penalty | -2 | -- |
| **Total** | **87** | **100** |

---

SCORE: 87/100

DIMENSIONS:
- Problem Definition: 18/20
- Solution Clarity: 17/20
- Alternatives Analysis: 13/15
- Scope Definition: 13/15
- Risk Assessment: 14/15
- Success Criteria: 14/15

ATTACKS:
1. Scope Definition: No effort estimate for the chosen approach in the Scope section -- the verdict paragraph in Alternatives mentions "~40 developer-hours for the entire suite" but the Scope section never commits to this number, and there is no phasing or ordering guidance (which flow first, which can be deferred). A team executing this proposal needs an explicit time commitment and priority order in the Scope section, not buried in the Alternatives verdict.
2. Solution Clarity: Differentiator from contract testing remains opinion-based -- "Overkill for a single-service application with one consumer (the frontend)" is a judgment call, not an analysis. The proposal does not explain what specific bug classes integration tests catch that contract tests miss (beyond one example), nor why a layered approach (contract tests for shape + integration tests for logic) was excluded. The verdict frames the choice as mutually exclusive without justification.
3. Risk Assessment / Success Criteria: Test execution time target exists only in risk mitigation, not in success criteria -- the mitigation says "Target <30s per file" but no success criterion enforces this. If execution time matters enough to appear in risk assessment, it should be a measurable success criterion (e.g., "Total integration test suite completes in <150s").
