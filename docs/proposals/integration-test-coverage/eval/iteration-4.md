# Evaluation Report: Integration Test Coverage Proposal

**Document:** `docs/proposals/integration-test-coverage/proposal.md`
**Date:** 2026-04-27
**Iteration:** 4

---

## Changes Since Iteration 3

All three iteration-3 attacks were addressed:

1. **Effort estimate and ordering in Scope section:** Line 149 now states "Estimated effort: ~40 developer-hours for the entire suite." Line 151 adds "Recommended execution order: F1 -> F2 -> F3 -> F4 -> F5 -> F6" with rationale for each position.
2. **Contract testing differentiator strengthened:** Alternative C's con now lists three specific bug classes with concrete examples: middleware chain bugs (RequireTeamMember failing to inject teamID), GORM scope errors (zero teamID returning all rows), and status machine violations (archive in-progress succeeding against DB). The verdict also explains why a layered approach is redundant.
3. **Execution time in success criteria:** Line 178 adds "Total integration test suite completes in <150s (5 flow files x 30s target per file)."

Score is reassessed from scratch below. No credit is given for "improvement" -- only for the current state of the text.

---

## Dimension 1: Problem Definition (18/20)

### Problem stated clearly: 6/7

The core problem is quantified and precise: "54 API endpoints but only 18 (33%) have integration tests." The domain-by-domain gap table with exact counts per domain is unambiguous. Two readers would arrive at the same understanding.

Deduction (-1): The opening paragraph still conflates two distinct concerns. "Several unit test gaps also exist (`permission_handler.go` completely untested, `ConvertToMain`/`UpdateTeam`/`GetByBizKey` methods lack coverage)" tacks unit test gaps onto the integration test problem without framing their relationship. Are these equally urgent? A prerequisite for the integration tests? A lower-priority stretch goal? The word "also" is doing structural work without earning it -- it hides a prioritization decision that the reader must infer.

### Evidence provided: 6/7

Strong quantitative evidence throughout. The domain table (8 rows, total/tested/gap columns) is precise and verifiable in structure. Unit test gaps name exact files and methods. The urgency section references a concrete incident: commit `1883499`, with cross-references to `docs/lessons/weekly-view-bug-fixes.md` (Bugs 2 and 3). The cost of inaction is grounded: "each escaped regression costs ~2-4 hours of manual diagnosis."

Deduction (-1): The provenance of the endpoint baseline (54 total, 18 tested) is still not cited. Was this derived from a coverage tool? A manual grep of route registrations? An audit of test files? A single sentence of methodology would allow independent verification. Without it, the entire problem statement rests on an unattributed count.

### Urgency justified: 6/6

The urgency section is anchored to a specific, verifiable incident: commit `1883499` introduced a timezone bug in `view_handler.go` and a filtering logic bug in `view_service.go`, both undetected because the Views domain had zero integration tests. The cross-reference to `docs/lessons/weekly-view-bug-fixes.md` is concrete. The "do nothing" alternative quantifies ongoing cost: "~2-4 hours per regression" based on the weekly-view incident timeline.

No deduction. This is well-grounded.

---

## Dimension 2: Solution Clarity (18/20)

### Approach is concrete: 7/7

Six test files are named and specified at the individual endpoint and test-case level. F1-F6 tables list every endpoint, every test scenario (happy path, validation errors, permission denied, cascading effects), and expected HTTP status codes. F5 includes concrete expected output values (specific stats objects, date ranges, BOM bytes, markdown section names). A developer could implement directly from these tables without ambiguity.

No deduction.

### User-facing behavior described: 6/7

The test descriptions specify HTTP status codes (201, 200, 403, 404, 422) and cascading behaviors (progress append to completion rollup, status change to sub-item cascade). F5 defines expected outputs with concrete values rather than vague descriptions. The section headers frame by "user flow" (Item Lifecycle, Item Pool Flow, Team Management, Admin User Management, Views & Reports).

Deduction (-1): The proposal describes test cases but not user-facing behavior narratives. For each flow, the reader sees endpoints and assertions but no sentence connecting the technical test cases to the user workflow they protect. For example, F1's header says "Item Lifecycle" but the content jumps straight into endpoint tables without a one-sentence narrative: "A PM creates a main item, breaks it into sub-items, assigns them to team members, and tracks completion through progress updates." The user-flow framing in the section title is not carried through to the content.

### Distinguishes from alternatives: 5/6

The flow-based approach is well-differentiated from endpoint-isolated testing with a specific example: "a `ConvertToMain` flow spans pool submission, sub-item creation, and pool status update -- isolated tests for each endpoint would all pass even if the handoff between them breaks."

The differentiator from contract testing is now substantially improved. Alternative C lists three specific bug classes that contract tests miss: (1) middleware chain bugs (RequireTeamMember failing to inject teamID causes 500, response shape still valid), (2) GORM scope errors (zero teamID returns all rows, correct-shaped but wrong data), (3) status machine violations (archive in-progress succeeds against DB when it should 422). The verdict also addresses the layered approach: "the integration tests already assert response structure in every test case, making a separate contract layer redundant."

Deduction (-1): The argument against the layered approach, while improved, still relies on a factual claim that is not substantiated: "integration tests already assert response structure in every test case." This claim could be verified by inspecting the proposed test tables, but the reader must do that work themselves. A single sentence like "Each of the 150+ test cases in the F1-F5 tables asserts on status code, response body keys, and business field values" would make the claim self-evident rather than requiring cross-referencing.

---

## Dimension 3: Alternatives Analysis (14/15)

### At least 2 alternatives listed: 5/5

Three alternatives: endpoint-isolated (A), do nothing (B), contract tests (C). All three are legitimate approaches with genuine trade-offs.

### Pros/cons for each: 5/5

All alternatives have balanced, honest treatment. Alternative A's pros are concrete: "Each test is self-contained, easier to debug when it fails. Faster to write per-endpoint (~15 min/test case) because no setup chaining is needed. Lower test brittleness." Alternative B's pros are legitimate: "Zero immediate investment. Team continues shipping features without test-maintenance overhead. CI time stays the same." Alternative C's pro is specific: "Catches schema regressions (field renames, type changes) -- would have caught the int64->string snowflake ID migration." All three alternatives include effort estimates with honest timeframes.

No deduction. The treatment is thorough and balanced.

### Rationale for chosen approach: 4/5

The verdict is explicit and well-reasoned: "We chose flow-based integration tests because they catch both inter-endpoint wiring bugs and business logic errors against a real database, with no new tooling dependencies." The hybrid rejection is argued: "mixing two test patterns creates inconsistent test structure that increases cognitive overhead for reviewers and maintainers, while the effort savings are marginal."

Deduction (-1): The hybrid rejection argument still relies on "increases cognitive overhead" as a self-evident claim. This is a reasonable engineering judgment, but it is not grounded. A single grounding sentence would strengthen it: e.g., "In our experience with the existing unit test suite, tests that follow different patterns (table-driven vs. single-case) generate the most review comments and PR back-and-forth" -- or acknowledge it as a team preference. As stated, the reader must accept this judgment on faith.

---

## Dimension 4: Scope Definition (14/15)

### In-scope items are concrete: 5/5

Five named integration test files with specific endpoint and test-case tables. Six named unit test gaps with file and method names. Shared test helpers as a deliverable. Estimated effort now explicitly stated: "~40 developer-hours for the entire suite." Recommended execution order: F1 -> F2 -> F3 -> F4 -> F5 -> F6 with rationale.

No deduction.

### Out-of-scope explicitly listed: 5/5

Four clear exclusions: frontend tests ("frontend test suite already covers component and E2E flows"), performance/load testing, E2E browser testing ("separate workflow"), and new features/bug fixes ("this is purely test coverage"). Each is named with parenthetical justification.

### Scope is bounded: 4/5

The effort estimate (40 developer-hours) is now in scope. The recommended execution order provides phasing. Each flow is sized (F1 is the largest at 17 endpoints, F5 depends on seeded data from F1-F4). The risk section mentions "one flow file per PR" for incremental delivery.

Deduction (-1): No explicit timeline or milestone structure. The recommended execution order says F1 first because "it covers the largest surface area and establishes shared helpers," but there is no estimate per flow. If the team has 2 weeks, which flows fit? If F1 takes 15 hours (largest), does F5 (dependent) get deferred? Per-flow effort estimates or a time-boxed phasing plan would make the scope more executable. The 40-hour total is useful but not decomposed.

---

## Dimension 5: Risk Assessment (14/15)

### Risks identified: 5/5

Five meaningful risks, all specific to integration test coverage: test execution time, test data pollution, edge cases exposing existing bugs, large PR size, and test brittleness from response-schema coupling. None are trivial.

### Likelihood + impact rated: 4/5

All five risks have likelihood and impact ratings showing appropriate range: Low impact for exposed bugs ("good"), Medium for execution time/PR size/brittleness, High for data pollution impact and exposed bugs likelihood.

Deduction (-1): Impact ratings remain qualitative without quantitative thresholds. "Medium" impact for "test execution time" -- does this mean 30s? 60s? 120s? The mitigation says "Target <30s per file" but this target is not linked to the impact rating. A reader cannot tell whether hitting 45s per file would escalate to "High" impact. Quantitative thresholds would make the assessment more honest.

### Mitigations are actionable: 5/5

All five mitigations are specific and actionable:

- "Each flow file is independent; can run in parallel. Target <30s per file."
- "Use transaction rollback (`tx.Begin()` at test start, `tx.Rollback()` in `t.Cleanup`) to guarantee no persisted state. This is preferred over manual cleanup because it is automatic and cannot be forgotten."
- "if the bug causes data loss, incorrect business state (e.g., wrong status transition accepted), or auth bypass, fix immediately in the same PR. Otherwise, file a bug and continue -- do not block the test coverage PR on non-critical fixes."
- "Submit one flow file per PR for incremental review."
- "Assert on structural fields (status code, top-level keys, specific business fields like `itemStatus`) but avoid asserting on field ordering, error message text, or pagination metadata format. When a test fails due to a legitimate API change, update the test -- do not suppress failures."

Each tells a developer exactly what to do. No ambiguity.

---

## Dimension 6: Success Criteria (15/15)

### Criteria are measurable: 5/5

All criteria are quantified and objectively verifiable: "All 54 API endpoints have at least one integration test." "Total integration test count >= 150 new test cases." "`go test ./tests/integration/...` passes with 0 failures." "Total integration test suite completes in <150s (5 flow files x 30s target per file)." Each can be checked with a single command or query.

### Coverage is complete: 5/5

Criteria now cover: endpoint count (all 54), per-endpoint status code coverage (happy path, validation, permission denied, not found), status transitions (all valid + at least one invalid), cascading effects, unit test gaps (6 named), total test count (150+), CI pass (0 failures), and execution time (<150s). The iteration-3 gap (execution time missing from success criteria) is closed. Every in-scope item has a corresponding success criterion.

### Criteria are testable: 5/5

All criteria are testable: count endpoints with integration tests (binary check), count test cases (test runner output), run the suite and check for failures (CI gate), verify 6 named gaps have passing tests, measure suite execution time. The iteration-2 inconsistency (overbroad "0 untested public methods" vs. scoped "6 unit test gaps") remains resolved.

No deduction. This is strong.

---

## Vague Language Penalty

Scanning for unquantified "better," "improved," "enhanced":

No instances of these specific words found.

Scanning for other vague language:

1. Line 153: "Frontend test changes (frontend test suite already covers component and E2E flows)" -- "covers" is not quantified. What coverage percentage? How many test cases? This is in an out-of-scope justification, so the impact on the proposal's substance is limited, but the word is unquantified.

This is a minor instance in a scope exclusion. The core proposal language is precise throughout.

**Total vague language penalty: 0** (the instance is minor and in an out-of-scope justification, not a substantive claim).

---

## Inconsistency Check

1. In-scope says "5 integration test files covering 36 untested endpoints." Solution describes 5 flow files + 1 unit gap file (F6). F6 is unit tests, not integration. Consistent.
2. Success criteria say "All 54 API endpoints" and the solution covers 36 new endpoints (18 already tested). 36 + 18 = 54. Consistent.
3. Success criterion "All 6 unit test gaps listed in F6 resolved" matches in-scope "6 unit test gap fixes." Consistent.
4. Scope section now explicitly states "~40 developer-hours" matching the Alternatives verdict. Consistent.
5. Success criterion "Total integration test suite completes in <150s (5 flow files x 30s target per file)" matches risk mitigation "Target <30s per file." Consistent.
6. Execution order in Scope (F1 first) aligns with Solution rationale (F1 establishes shared helpers). Consistent.

**No inconsistency penalty applied.**

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 18 | 20 |
| Solution Clarity | 18 | 20 |
| Alternatives Analysis | 14 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 15 | 15 |
| Vague language penalty | 0 | -- |
| **Total** | **93** | **100** |

---

SCORE: 93/100

DIMENSIONS:
- Problem Definition: 18/20
- Solution Clarity: 18/20
- Alternatives Analysis: 14/15
- Scope Definition: 14/15
- Risk Assessment: 14/15
- Success Criteria: 15/15

ATTACKS:
1. Problem Definition: Unit test gaps and integration test gaps are conflated in the opening paragraph without framing their relationship -- "Several unit test gaps also exist (`permission_handler.go` completely untested, `ConvertToMain`/`UpdateTeam`/`GetByBizKey` methods lack coverage)" -- the word "also" hides a prioritization decision; the reader cannot tell whether unit test gaps are equally urgent, a prerequisite, or lower priority than the integration test coverage that is the proposal's primary subject.
2. Scope Definition: Per-flow effort estimates are missing, making the scope less executable -- the total "~40 developer-hours" is stated but not decomposed; if F1 (17 endpoints) takes proportionally more time than F4 (6 endpoints), a team cannot plan sprint capacity without per-flow estimates. Add rough hour estimates per flow file (e.g., "F1: ~12h, F2: ~7h, F3: ~8h, F4: ~5h, F5: ~5h, F6: ~3h").
3. Alternatives Analysis: The hybrid rejection argument relies on the unsubstantiated claim "mixing two test patterns creates inconsistent test structure that increases cognitive overhead for reviewers and maintainers" -- this is a reasonable engineering judgment presented as self-evident; ground it with team experience (e.g., "In our PR reviews, mixed-pattern test files generate 2x more review comments") or acknowledge it as a stated team preference.
