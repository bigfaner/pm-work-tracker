# Evaluation Report: Integration Test Coverage Proposal

**Document:** `docs/proposals/integration-test-coverage/proposal.md`
**Date:** 2026-04-27
**Iteration:** 1

---

## Dimension 1: Problem Definition (16/20)

### Problem stated clearly: 6/7

The core problem is clear: 54 endpoints, only 18 (33%) tested. The table breaks gaps down by domain with specific counts. Two readers would arrive at the same understanding of the problem.

Deduction: The problem conflates two distinct issues -- integration test gaps AND unit test gaps -- without distinguishing their relative priority or whether they are equally urgent. The sentence "Several unit test gaps also exist" tacks on a separate concern mid-paragraph without framing how it relates to the integration test problem.

### Evidence provided: 6/7

Strong quantitative evidence in the domain-by-domain table. The specific file and method names for unit test gaps (`permission_handler.go`, `ConvertToMain`, `GetByBizKey`) are concrete and verifiable.

Deduction: The source of the endpoint count (54 total, 18 tested) is not cited. Is this from a coverage tool? A manual audit? `grep`? Without provenance, the reader cannot verify or reproduce the baseline.

### Urgency justified: 4/6

The urgency section makes a reasonable argument about regressions being caught only manually and unit tests missing handler-to-repository wiring errors.

Deduction (-1): The urgency is asserted but not demonstrated with a concrete incident. No example of a regression that slipped through, no recent bug report, no production issue. For a test coverage proposal, one past incident would be far more compelling than the abstract warning.

Deduction (-1): The phrase "only caught manually" is vague -- does the team currently do manual QA? Is there no staging environment? The urgency assumes context the reader may not have.

---

## Dimension 2: Solution Clarity (15/20)

### Approach is concrete: 6/7

The user-flow organization strategy is well-explained. The 6 test files are named and their contents are specified at the individual endpoint and test-case level. This is highly concrete -- a developer could start implementing immediately from F1-F6 tables.

Deduction: The proposal says "shared test helpers extracted from existing integration tests for reuse" but never shows or describes what these helpers are, what patterns they abstract, or how they would be structured.

### User-facing behavior described: 5/7

The test descriptions focus on HTTP status codes and endpoint behaviors (201, 403, 404, 422), which are observable behaviors. The cascading effects (progress append -> completion rollup, status change -> sub-item cascade) describe real user-facing workflows.

Deduction (-1): For the "end user" criterion, the proposal describes test cases, not user-facing behavior. The reader must infer the user experience from test case names. The proposal would be stronger with a brief narrative of what each user flow looks like from the user's perspective before diving into test tables.

Deduction (-1): The Views & Reports section (F5) describes expected output content ("BOM", "correct headers", "markdown export") but does not specify what correctness means. What are the "expected sections" in a weekly report? What headers should the CSV have?

### Distinguishes from alternatives: 4/6

The flow-based organization is clearly differentiated from the endpoint-isolated pattern in Alternative A. The rationale (catches inter-endpoint bugs) is specific and defensible.

Deduction: The distinction from endpoint-isolated tests is stated in one sentence: "Simpler to write but misses inter-endpoint bugs." No evidence or example is given of a bug that endpoint-isolated tests would miss but flow-based tests would catch. The differentiator is asserted, not demonstrated.

---

## Dimension 3: Alternatives Analysis (9/15)

### At least 2 alternatives listed: 4/5

Three alternatives are listed: endpoint-isolated (A), do nothing (B), and contract tests (C). Good breadth.

Deduction: A code-generation approach (generating test scaffolding from OpenAPI specs or route definitions) is a viable alternative that was not considered. For 36 endpoints, automated test scaffolding could reduce manual effort significantly.

### Pros/cons for each: 3/5

Alternative A has a pro ("simpler to write") and a con ("misses inter-endpoint bugs"). Alternative B has only risks listed (no pros). Alternative C has only a con ("overkill").

Deduction (-1): Alternative B (do nothing) lists only risks, no pros. There are legitimate arguments for the status quo: no test maintenance burden, no increase in CI time, team can focus on feature work. A straw-man "do nothing" weakens the analysis.

Deduction (-1): No quantitative comparison. How much longer do flow-based tests take to write vs. endpoint-isolated? How many more lines of code? How much more CI time? For a test coverage proposal, these numbers matter.

### Rationale for chosen approach: 2/5

The proposal chose flow-based tests, and the primary rationale is catching inter-endpoint bugs. This is reasonable but thin.

Deduction: The verdict is never explicitly stated. The reader must infer the choice from the fact that Solution describes flow-based tests while Alternatives lists endpoint-isolated as "rejected." There is no "We chose X because..." sentence.

Deduction: No consideration of a hybrid approach (flow-based for critical paths, endpoint-isolated for simple CRUD). The choice is binary when a blended strategy might be optimal.

---

## Dimension 4: Scope Definition (13/15)

### In-scope items are concrete: 4/5

Five named test files with specific endpoint and test case tables. Six named unit test gaps with file and method names. Shared test helpers. All are deliverables.

Deduction: "Exhaustive edge cases per endpoint" is not a concrete deliverable. "Exhaustive" is subjective -- what counts as exhaustive? The test tables that follow are concrete, but this bullet uses unquantified language.

### Out-of-scope explicitly listed: 5/5

Four clear exclusions: frontend tests, performance/load testing, E2E browser testing, and new features/bug fixes. Each is named, not implied.

### Scope is bounded: 4/5

The scope is well-bounded: 5 files, 6 unit test gaps, 36 endpoints. A team can execute this.

Deduction: No time estimate or milestone breakdown. The proposal says "one flow file per PR" in the risk section, which implies incremental delivery, but there is no explicit phasing, ordering, or timeline. Is F1 first because it is the most critical? Is F6 last because unit gaps are lower priority?

---

## Dimension 5: Risk Assessment (10/15)

### Risks identified: 4/5

Four risks: test execution time, test data pollution, edge cases exposing bugs, and large PR size. All are meaningful for a test coverage effort.

Deduction: Missing risk: test brittleness from coupling to implementation details (e.g., specific error messages, response field ordering). Integration tests that assert on exact response bodies become expensive to maintain when APIs evolve. This is a well-known risk for integration test suites.

### Likelihood + impact rated: 3/5

Ratings are provided for all four risks. The third risk (edge cases exposing bugs) is honestly rated as High likelihood, Low impact, with the note "(good)" -- this is candid and appropriate.

Deduction (-1): All likelihood ratings are Medium or High. No risk is rated Low likelihood. This suggests either the risk landscape is uniformly dire (unlikely) or the assessment lacks range. For instance, contract test tooling breaking would be a Low likelihood risk worth mentioning.

Deduction (-1): The impact ratings are single words without explanation. What does "Medium" impact for "test execution time" mean in minutes? What does "High" impact for "test data pollution" look like in practice (flaky tests? false positives? CI failures?)? Without thresholds, the ratings are subjective.

### Mitigations are actionable: 3/5

Mitigations are specific: "Target <30s per file," "Transaction rollback or cleanup in t.Cleanup," "Submit one flow file per PR."

Deduction: "Transaction rollback or cleanup in t.Cleanup" presents two options without choosing one. The mitigation is undecided. Which strategy will be used? Transaction rollback and per-test cleanup have different trade-offs (speed vs. isolation).

Deduction: The "file bugs or fix immediately depending on severity" mitigation for exposed bugs is vague. What severity threshold triggers an immediate fix vs. filing a bug? This should be defined.

---

## Dimension 6: Success Criteria (12/15)

### Criteria are measurable: 4/5

"All 54 API endpoints have at least one integration test" -- measurable. "Total integration test count >= 150 new test cases" -- measurable with a threshold. "go test ./tests/integration/... passes with 0 failures" -- binary and verifiable.

Deduction: "Cascading effects are verified" is not independently measurable. What counts as verified? An assertion on a specific response field? A database state check? The criterion is a category, not a measurement.

### Coverage is complete: 4/5

The criteria cover: endpoint count, per-endpoint status code coverage, status transitions, cascading effects, unit test gaps, total test count, and CI pass.

Deduction: No criterion for test execution time. The risk section mentions "Target <30s per file" but this target does not appear in Success Criteria. If 30s per file is important enough to be a risk mitigation, it should be a success criterion.

### Criteria are testable: 4/5

Most criteria are testable: you can count endpoints, count test cases, run the test suite, check for public methods without tests.

Deduction: "All unit test gaps closed (0 untested public methods in service/handler layers)" is broad. Does this mean ALL public methods across the entire codebase, or only the 6 named gaps? If the former, it is scope creep beyond what is in scope. If the latter, the criterion should reference the specific list.

---

## Vague Language Penalty

Instances of unquantified language:

1. "Exhaustive edge cases per endpoint" -- what is "exhaustive"? (-2)
2. "comprehensive coverage" (in out-of-scope: "frontend already has comprehensive coverage") -- not quantified (-2)

**Total vague language penalty: -4**

---

## Inconsistency Check

The in-scope section says "5 integration test files covering 36 untested endpoints." The Solution section describes 6 files (5 flow + 1 unit gap). This is fine -- F6 is unit tests, not integration tests. No material inconsistency.

The success criteria say "All 54 API endpoints" but the solution only covers 36 untested ones. This is consistent (36 new + 18 existing = 54 total), but the proposal never makes this arithmetic explicit. No penalty.

**No inconsistency penalty applied.**

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 16 | 20 |
| Solution Clarity | 15 | 20 |
| Alternatives Analysis | 9 | 15 |
| Scope Definition | 13 | 15 |
| Risk Assessment | 10 | 15 |
| Success Criteria | 12 | 15 |
| Vague language penalty | -4 | — |
| **Total** | **71** | **100** |

---

SCORE: 71/100

DIMENSIONS:
- Problem Definition: 16/20
- Solution Clarity: 15/20
- Alternatives Analysis: 9/15
- Scope Definition: 13/15
- Risk Assessment: 10/15
- Success Criteria: 12/15

ATTACKS:
1. Alternatives Analysis: Straw-man treatment of "do nothing" and absent quantitative trade-off comparison -- "Current 33% endpoint coverage. Risks: regressions in untested flows go unnoticed" -- the do-nothing alternative lists only risks with zero pros, and no alternative includes effort estimates (test files to write, CI minutes added, maintenance cost), making the comparison ungrounded. The section needs honest pros for each alternative and at least rough effort/cost numbers.
2. Risk Assessment: Missing test brittleness risk and undecided mitigation strategy -- "Transaction rollback or cleanup in t.Cleanup" presents two options without committing to one, and no risk addresses the well-known problem of integration tests becoming tightly coupled to response schemas, making them expensive to maintain when APIs evolve. Add a brittleness risk, pick ONE data isolation strategy, and define the severity threshold for "fix immediately vs. file bug."
3. Problem Definition: Urgency lacks a concrete triggering incident -- "regressions in API behavior ... are only caught manually" -- this abstract warning has no anchor in a real event. Cite one recent regression that escaped to production or staging because the relevant endpoint lacked an integration test. Without it, the urgency reads as generic test-coverage advocacy rather than a response to a felt pain.
