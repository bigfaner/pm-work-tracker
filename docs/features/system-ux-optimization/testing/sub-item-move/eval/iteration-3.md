---
iteration: 3
total_score: 885
pass: true
dimensions:
  completeness: 175
  semantic_purity: 180
  precondition_exclusivity: 130
  fact_alignment: 135
  surface_fitness: 140
  internal_consistency: 125
---

# Journey Eval: sub-item-move — Iteration 3

**Date**: 2026-06-04

## Score Summary

| Dimension | Score | Min Threshold | Status |
|-----------|-------|---------------|--------|
| Completeness | 175/200 | 120 | PASS |
| Semantic Purity | 180/200 | — | PASS |
| Precondition Exclusivity | 130/150 | 90 | PASS |
| Fact Alignment | 135/150 | 90 | PASS |
| Surface Fitness | 140/150 | 90 | PASS |
| Internal Consistency | 125/150 | 90 | PASS |
| **Total** | **885/1000** | **≥850** | **PASS** |

**Overall: PASS** — Total score 885 >= 850 and all dimensions above min threshold.

## Dimension Details

### 1. Completeness — 175/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata complete | 48/50 | Name `sub-item-move` in kebab-case. `risk_level: High` justified. Sources list three documents. Generated date present. Near-perfect. Minor: `feature` field matches manifest slug. |
| Steps complete with required fields | 75/80 | Happy path Steps 1-3 each have Precondition, User Action, and Expected Result. Edge cases E1-E10 all have the three required fields. Steps form a coherent ordered sequence. The addition of E8 (move non-existent sub-item), E9 (move to non-existent target), and E10 (API validation error) resolves the major iteration-2 gap. Minor deduction: Steps still lack numbered outcome identifiers for test traceability, though the single-outcome-per-step structure mitigates this. |
| Outcomes cover happy path + required derived scenarios | 52/70 | Happy path fully covered (Steps 1-3). **Web mandatory**: `validation-error` present in E6, `session-expired` present in E7. **API mandatory**: `unauthorized` present in E5. **New in iteration 3**: `not-found` API outcomes now covered by E8 (non-existent sub-item) and E9 (non-existent target main item). API `validation-error` covered by E10. This is a substantial improvement. Remaining gaps: No `network-error` web outcome (not mandatory but common). No `conflict` step outside the concurrent scenario (the api-handbook lists CONFLICT 409 as a general error code). The happy path does not cover the full API response contract (e.g., verifying `newSubCode` and `mainItemBizKey` fields are returned). These are minor. |

### 2. Semantic Purity — 180/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 78/80 | All expected results are in natural language. No regex, CSS selectors, XPath, or framework assertions. Examples: "A target main item selector is displayed, listing available (non-closed, non-completed) main items excluding the current parent" and "The API returns a 'not found' error; no data is modified." Clean throughout. |
| Preconditions are declarative statements | 55/60 | Preconditions are properly declarative throughout. E8: "The sub-item ID in the request does not exist in the database." E9: "The target main item ID in the request does not exist in the database." E10: "A move API request is sent with missing or invalid fields (e.g., empty target main item ID)." All state conditions, not setup procedures. Minor: E10's precondition uses the passive construction "is sent with" which blends a state description with an action — it describes what the request looks like rather than a pure state, but this is acceptable for API surface validation scenarios. |
| No implementation coupling in Step descriptions | 47/60 | Web steps (1-3, E1-E4, E6-E7) are entirely user-level and clean of implementation coupling. API steps (E5, E8-E10) use action descriptions like "sends a move API request" and "An API request is sent to the move endpoint" which avoid the explicit HTTP method + path coupling that was flagged in iteration 2. This is a clear improvement. E5's action is now "The user sends a move API request without proper authorization" rather than the full endpoint path. Minor deduction: E10's action says "An API request is sent to the move endpoint with an invalid or missing target ID" — "the move endpoint" is a slight endpoint reference, but it is generic enough. E8/E9 actions say "A move API request is sent" which is clean. The remaining coupling is minimal. |

### 3. Precondition Exclusivity — 130/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes | 58/60 | Iteration 3 retains the E3 dual-outcome structure from iteration 2 with distinct Expected Results A and B. All other steps have single outcomes. The preconditions for E8 ("sub-item ID does not exist"), E9 ("target main item ID does not exist"), and E10 ("missing or invalid fields") are clearly distinct from each other and from all other steps. Minor: E3's two outcomes still share the same precondition text (ordering is the implicit differentiator), but this was noted in iteration 2 as inherent to concurrency scenarios. |
| Preconditions sufficient to uniquely select Outcome | 45/50 | Each step (except E3) has a single outcome. E3's two outcomes are differentiated by "first request arrives" vs "second request arrives after first transaction committed" — the iteration-3 revision added "after first transaction committed" to Outcome B, improving distinguishability. For test generation, a tester can now clearly construct two test cases. Minor deduction: the precondition text for E3 still does not split into two variants; the differentiation lives in the expected results, not the preconditions. |
| No missing Preconditions for error/boundary Outcomes | 27/40 | E8, E9, E10 each have clear preconditions stating what triggers the error. E8: "The sub-item ID in the request does not exist in the database." E9: "The target main item ID in the request does not exist in the database." E10: "missing or invalid fields (e.g., empty target main item ID)." E5: "A user without the sub_item:update permission." E4: "Main item A has been soft-deleted." These are strong. Remaining gaps: E1's expected result still covers two sub-cases ("shown as disabled or non-selectable" AND "if somehow selected, the operation is rejected") without distinct preconditions. E10's precondition "missing or invalid fields" is slightly vague — it covers both missing and invalid but these could have different behaviors. E3's precondition does not establish which user's request arrives first. (-13) |

### 4. Fact Alignment — 135/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 50/60 | Iteration 3 adds source annotations to E3's concurrent outcome: `<!-- source: inferred — derived from concurrent edit pattern in prd-spec -->`. E5 through E10 all carry `source: inferred` annotations citing the specific surface mandatory outcome that triggered derivation. This is a clear improvement. The Journey frontmatter `sources` list provides document-level traceability. Factual claims from the PRD (code number regeneration, status/assignee preservation) are present in Steps 2 and 3 with `fact: prd-spec Story 7 AC1` annotation on Step 1. Remaining gap: No explicit fact_id references for api-handbook claims (e.g., 404 NOT_FOUND error codes in E8/E9 are documented in api-handbook but not individually traced). The Journey Invariants section makes factual claims without traceability. (-10) |
| Inferred claims have required_outcomes rule support and source: inferred | 48/50 | All mandatory-derived steps (E5-E10) have proper `source: inferred` annotations with the triggering rule. E5: "derived from API surface `unauthorized` mandatory outcome." E6: "derived from Web surface `validation-error` mandatory outcome." E7: "derived from Web surface `session-expired` mandatory outcome." E8: "derived from API surface `not-found` common boundary outcome." E9: "derived from API surface `not-found` common boundary outcome." E10: "derived from API surface `validation-error` outcome." E3: "derived from concurrent edit pattern in prd-spec." This is thorough. Minor: E1 (closed target) and E2 (same parent) are derived from PRD Story 7 acceptance criteria but still lack explicit source annotations — these are factual rather than inferred, which is defensible, but the boundary is not always clear. (-2) |
| No hallucinated claims without classification | 37/40 | All significant claims are either traced to PRD/api-handbook or annotated as inferred. No factual errors detected. The api-handbook documents 400 BAD_REQUEST for closed target and same-parent, 404 NOT_FOUND for sub-item and target main item, and 409 CONFLICT — all correctly reflected in the journey. One minor concern: E1's expected result says "The closed main item is shown as disabled or non-selectable in the selector" — this UI-level behavior is not explicitly specified in any source document (the PRD says "rejected" but does not specify whether prevention is UI-side or server-side). This is an unclassified inference, but it is a reasonable assumption for a web surface. (-3) |

### 5. Surface Fitness — 140/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 58/60 | **Web mandatory `validation-error`**: Present in E6. **Web mandatory `session-expired`**: Present in E7. **API mandatory `unauthorized`**: Present in E5. All three mandatory outcomes are present and well-specified. **New**: API `not-found` (E8, E9) and API `validation-error` (E10) now present as additional boundary outcomes. This is comprehensive. Minor deduction: E6 covers only the "no target selected" validation case. Surface-web.md mentions multiple validation types (format invalid, non-numeric input). Adding at least one more validation variant would strengthen completeness. (-2) |
| Test strategy proportions match surface guidance | 42/50 | Web and API both require balanced 50/50. Steps 1-3 form the Journey smoke test (end-to-end web workflow). Edge cases E1-E4, E6-E7 are web Contract-level. E5, E8-E10 are API Contract-level. E3 spans both. The surface annotations correctly tag each step. The distribution is roughly balanced between web and API steps (7 web-tagged, 4 API-tagged, 1 dual). However, there is still no explicit test strategy section delineating which steps are Contract vs Journey test level. The implicit split is clear to a human reader but could be more explicit for automated test generation. (-8) |
| Surface-specific environment assumptions realistic | 40/40 | Web steps correctly assume browser interactions. API steps correctly assume HTTP request/response patterns. E7 (session-expired) correctly assumes browser redirect to login. E6 (validation-error) correctly assumes client-side validation ("no API request is sent"). E8/E9/E10 (API errors) correctly assume server-side response codes. The concurrency scenario (E3) realistically describes both web and API perspectives. All environment assumptions are realistic for their respective surfaces. Full marks. |

### 6. Internal Consistency — 125/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 50/60 | All five invariants are preserved across all steps. Invariant 1 (code regeneration) holds in Steps 2, 3, E3A. Invariant 2 (status/assignee unchanged) holds in Step 2. Invariant 3 (atomic operation) — revised to "Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)" — this is better than iteration 2's "single database transaction" wording. Holds in E3 (first succeeds atomically, second is rejected). Invariant 4 (closed targets invalid) holds in E1. Invariant 5 (same parent rejected) holds in E2. E4 (deleted source) is consistent — no move is performed. E8-E10 (not-found, validation) result in no data modification, consistent with all invariants. Minor: Invariant 3's "atomically" wording is slightly ambiguous — does "atomically" mean "in a single transaction" or "idempotent"? The parenthetical clarifies, but "fully completes or fully rolls back" is transactional language, not idempotency language. This is acceptable. (-10 for slight ambiguity in invariant wording) |
| Cross-Step references consistent | 42/50 | Setup now establishes: main item A (source), main item B (valid target), main item C (terminal status), two PM users. E3 references "different targets" — clear enough. E4 references "Main item A has been soft-deleted by another user" — introduces a new condition mid-scenario, which is acceptable for edge cases. E8/E9 introduce non-existent IDs which are not setup entities but boundary conditions. E10 introduces invalid request payloads. All cross-references are consistent. Minor deduction: The Setup says "A sub-item exists under main item A" but E8 tests a non-existent sub-item — there is no explicit statement that E8 operates outside the setup state (using a fabricated ID). This is implicit and generally understood for negative testing, but could be more explicit. (-8) |
| Risk level consistent with content | 33/40 | Risk level "High" is justified: (1) State mutation (parent relationship change). (2) Code number regeneration (identifier change). (3) Concurrent operation risk. (4) Multiple edge cases involving data integrity. The risk justification comment mentions "data loss risk" — moving a sub-item does not destroy data, but the concurrent and deleted-source scenarios could lead to user confusion or unintended state, which is a soft form of data integrity risk. The "irreversible" claim in the justification comment is slightly overstated (a sub-item can be moved back). However, the overall High classification is defensible. (-7 for imprecise risk justification comment) |

## Attack Points (Prioritized Fixes)

1. **Precondition Exclusivity**: Step E1 covers two distinct behaviors (disabled-in-selector AND server-rejection-if-selected) under a single outcome with a single precondition. Quote: `Expected Result: The closed main item is shown as disabled or non-selectable in the selector; if somehow selected, the operation is rejected with a message indicating the target is closed`. Fix: Split into two outcomes with distinct preconditions — Outcome A: "The closed main item is listed but displayed as disabled/non-selectable in the target selector" (UI prevention), Outcome B: "If a closed main item ID is submitted to the API, the operation is rejected with a 400 BAD_REQUEST response" (API rejection).

2. **Completeness**: Step E10's precondition "missing or invalid fields" conflates two different validation scenarios — missing target ID vs invalid target ID format. These could trigger different validation logic. Quote: `Precondition: A move API request is sent with missing or invalid fields (e.g., empty target main item ID)`. Fix: Split into two steps or two outcomes — one for missing target ID (empty/null) and one for invalid format (non-bizKey string), as these may produce different error response structures.

3. **Internal Consistency**: Invariant 3 uses "atomically" with parenthetical clarification "the operation either fully completes or fully rolls back." This is transactional language but does not specify the isolation level or concurrency guarantee. Quote: `Move and code re-generation always execute atomically (the operation either fully completes or fully rolls back)`. Fix: Replace with "Move and code re-generation are performed as a single atomic unit — if any part fails, the entire operation is rolled back and no partial state is persisted."

4. **Fact Alignment**: Edge cases E1 and E2 are derived from PRD Story 7 acceptance criteria but lack explicit source annotations. Quote from E1: no `<!-- source: ... -->` comment. Fix: Add `<!-- source: derived from PRD Story 7 AC ("用户尝试移动子事项到已关闭状态的主事项") -->` to E1 and `<!-- source: derived from PRD Story 7 AC ("用户尝试移动子事项到同一主事项") -->` to E2.

5. **Precondition Exclusivity**: Step E3's two outcomes (A and B) share identical precondition text. The temporal ordering differentiator ("first request arrives" vs "second request arrives after first transaction committed") appears only in the expected results. Quote: Both share precondition "Two PM users have the same sub-item's move dialog open simultaneously." Fix: Add precondition variants — Outcome A: precondition includes "User 1's request is processed first by the server." Outcome B: precondition includes "User 2's request arrives after User 1's transaction has committed."
