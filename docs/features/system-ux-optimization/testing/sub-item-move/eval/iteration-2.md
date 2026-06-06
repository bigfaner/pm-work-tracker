---
iteration: 2
total_score: 810
pass: false
dimensions:
  completeness: 160
  semantic_purity: 175
  precondition_exclusivity: 120
  fact_alignment: 120
  surface_fitness: 120
  internal_consistency: 115
---

# Journey Eval: sub-item-move — Iteration 2

**Date**: 2026-06-04

## Score Summary

| Dimension | Score | Min Threshold | Status |
|-----------|-------|---------------|--------|
| Completeness | 160/200 | 120 | PASS |
| Semantic Purity | 175/200 | — | PASS |
| Precondition Exclusivity | 120/150 | 90 | PASS |
| Fact Alignment | 120/150 | 90 | PASS |
| Surface Fitness | 120/150 | 90 | PASS |
| Internal Consistency | 115/150 | 90 | PASS |
| **Total** | **810/1000** | **≥850** | **FAIL** |

**Overall: FAIL** — All dimensions pass threshold, but total score 810 < 850 target.

## Dimension Details

### 1. Completeness — 160/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata complete | 45/50 | Name `sub-item-move` in kebab-case. `risk_level: High` justified by state mutation and irreversible parent change. Sources list PRD user stories, PRD spec, and api-handbook. Generated date present. Slight deduction: `fact_table` reference still absent — no explicit fact ID citations in any step. |
| Steps complete with required fields | 70/80 | Happy path Steps 1-3 each have Precondition, User Action, and Expected Result — well-structured. Edge cases E1-E8 each have all three fields. Steps form a coherent ordered sequence. Deduction: Steps still lack numbered outcome identifiers (e.g., Outcome 1.1, 1.2) which would improve traceability for test generation, but this is less critical given the single-outcome-per-step structure. |
| Outcomes cover happy path + required derived scenarios | 45/70 | Happy path fully covered (Steps 1-3). Edge cases cover: closed target (E1), same-parent (E2), concurrent move (E3), deleted source (E4), unauthorized API (E5), validation-error web (E6), session-expired web (E7), API validation-error (E8). Web mandatory `validation-error` is present (E6). Web mandatory `session-expired` is present (E7). API mandatory `unauthorized` is present (E5). API common `validation-error` is present (E8). **Gaps**: No `not-found` outcome for API (api-handbook lists 404 NOT_FOUND for both sub-item and target main item — no step covers moving a non-existent sub-item or targeting a non-existent main item via API). No `network-error` web outcome. These are not mandatory per surface rules but are common and documented in the api-handbook error responses, reducing completeness score. |

### 2. Semantic Purity — 175/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 78/80 | Excellent natural language throughout. Outcomes describe what the user/system observes: "A target main item selector is displayed, listing available main items", "The moved sub-item appears in the sub-item list". Step E5 references the API endpoint path in the User Action ("sends PUT /api/v1/teams/:teamId/sub-items/:subId/move") which is borderline acceptable for an API surface — it describes the action, not an assertion. No regex, CSS selectors, or framework assertions found. |
| Preconditions are declarative statements | 58/60 | Preconditions are properly declarative: "PM user is viewing a sub-item detail page that belongs to main item A", "A main item C is in terminal status". Step E6's precondition "The move target selector is open but no target has been selected" is declarative. Step E3's precondition "Two PM users have the same sub-item's move dialog open simultaneously" is declarative and states the condition, not the setup procedure. |
| No implementation coupling in Step descriptions | 39/60 | Step E5's User Action contains the full API endpoint path `PUT /api/v1/teams/:teamId/sub-items/:subId/move` — this is implementation coupling (HTTP method + path). For API surface, some endpoint reference is expected, but the full path with method is more appropriate for a Contract test than a Journey description. Steps 1-3 and E1-E4, E6-E7 are clean of implementation coupling. Step E8 also contains the endpoint reference. The Step descriptions for web steps are entirely user-level. The api-handbook is cited as a source, which partially justifies the endpoint reference, but the rubric asks for "user-level or system-level actions, not internal function calls, database queries, or API endpoint details." |

### 3. Precondition Exclusivity — 120/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes | 55/60 | Iteration 2 significantly improves over iteration 1. Step E3 now has two distinct Expected Results (A and B) with clearly differentiated outcomes — first transaction succeeds, second detects conflict. This resolves the iteration-1 ambiguity. Most steps have single outcomes, so no overlap. Deduction: Step E3's two outcomes (A and B) still share the same precondition ("Two PM users have the same sub-item's move dialog open simultaneously"). The precondition does not distinguish which user's transaction succeeds vs. fails — timing/ordering is the differentiator but is not stated as a precondition. This is inherent to the concurrency scenario and partially excusable. (-5) |
| Preconditions sufficient to uniquely select Outcome | 40/50 | Each step has a single expected outcome (except E3 which has two branches). For the happy path and most edge cases, the precondition uniquely determines the outcome. Step E3's two outcomes are differentiated by "first" vs "second" transaction order, which is implicit rather than explicitly stated in preconditions. For test generation, a tester would need to know to test both branches separately. (-10) |
| No missing Preconditions for error/boundary Outcomes | 25/40 | Step E5's precondition states "A user without the sub_item:update permission sends a move API request" — improved from iteration 1 which only said "member-role user." However, Step E6's precondition "The move target selector is open but no target has been selected" is good. Step E7's precondition "The user's session has expired while the move target selector is open" is good. Step E4's precondition "Main item A has been soft-deleted by another user" is good. **Remaining gaps**: Step E1's precondition says "A main item C is in terminal status" but does not state that the user is attempting the move — the action says "PM user attempts to select main item C" but the expected result covers two sub-cases (disabled in selector OR rejected if somehow selected) without distinct preconditions. Step E8's precondition "missing or invalid fields (e.g., empty target main item ID)" is slightly vague — it should specify the exact invalid state. (-15) |

### 4. Fact Alignment — 120/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 40/60 | Significant improvement over iteration 1 (which scored 20/60). The Journey now includes `source: inferred` annotations on Steps E5, E6, E7, E8 — each citing the mandatory outcome rule that triggered derivation (e.g., "derived from API surface `unauthorized` mandatory outcome", "derived from Web surface `validation-error` mandatory outcome"). The `sources` frontmatter references prd-user-stories.md (Story 7), prd-spec.md, and api-handbook.md. However, no explicit `fact_id` references exist anywhere — factual claims like "code number is automatically regenerated under the target" (which matches api-handbook's `newSubCode` field) and "status and assignee remain unchanged" (from PRD Story 7 acceptance criteria) are not individually traced. The Journey Invariants section makes strong factual claims ("move and code re-generation always execute within a single database transaction") without traceability markers. (-20 for missing fact_id traceability) |
| Inferred claims have required_outcomes rule support and source: inferred | 45/50 | All four inferred steps (E5, E6, E7, E8) properly include `source: inferred` annotations with the triggering mandatory outcome rule. E5: "derived from API surface `unauthorized` mandatory outcome". E6: "derived from Web surface `validation-error` mandatory outcome". E7: "derived from Web surface `session-expired` mandatory outcome". E8: "derived from API surface `validation-error` outcome". Edge cases E1 (closed target) and E2 (same parent) are derived from PRD Story 7 acceptance criteria but do not have `source: inferred` annotations — they could be considered factual claims from the PRD, which is a gray area. Step E3 (concurrent move) and E4 (deleted source) are inferred from general system design but lack annotations. (-5 for incomplete inference annotations on non-mandatory edge cases) |
| No hallucinated claims without classification | 35/40 | Major improvement from iteration 1 which had a -30 deduction for incorrect API endpoint. Iteration 2 correctly states `PUT /api/v1/teams/:teamId/sub-items/:subId/move` matching the api-handbook exactly. No factual errors detected in endpoint methods, paths, or error codes. The claim in E3 about "conflict error" aligns with the api-handbook's CONFLICT 409 error code. One minor concern: Step E1's expected result says "The closed main item is shown as disabled or non-selectable in the selector" — this is a UI behavior claim not traceable to any source document (the PRD says the operation is "rejected" but does not specify whether the UI prevents selection or the backend rejects). This is an unclassified inference. (-5) |

### 5. Surface Fitness — 120/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 50/60 | **Web mandatory `validation-error`**: Present in Step E6. Covers user submitting without target selection. Aligns with surface-web.md guidance ("required field left empty"). **Web mandatory `session-expired`**: Present in Step E7. Covers session expiry during move workflow. Aligns with surface-web.md guidance. **API mandatory `unauthorized`**: Present in Step E5. Covers request without sub_item:update permission. Aligns with surface-api.md guidance. All three mandatory outcomes are present — major improvement from iteration 1 which scored 25/60. Deduction: Step E6 covers only one validation scenario (no target selected). The surface-web.md guidance lists multiple validation types (required field empty, format invalid, non-numeric input). The journey could benefit from at least one more validation edge case (e.g., selecting an invalid/non-existent target). (-10) |
| Test strategy proportions match surface guidance | 35/50 | Web and API both require balanced 50/50 (Contract/Journey). The journey covers the full user workflow (Steps 1-3 happy path) for Journey-level tests and individual edge cases (E1-E8) for Contract-level tests. The happy path is a proper Journey sequence. Edge cases are more Contract-level (individual behavior verification). However, the Journey does not explicitly distinguish which steps are intended for Contract vs Journey test levels, and no guidance is provided on how to split testing effort. The surface annotations (`<!-- surface: web -->`, `<!-- surface: api -->`) correctly tag steps, which helps test generation. (-15 for missing explicit test strategy guidance) |
| Surface-specific environment assumptions realistic | 35/40 | Web steps correctly assume browser interactions: selector display, navigation, page updates. API steps correctly assume HTTP request/response: PUT method, endpoint path, error responses. The concurrency scenario (E3) is realistic for both web and api surfaces (tagged `<!-- surface: web, api -->`). Session expiry (E7) correctly assumes browser redirect behavior. The validation-error (E6) correctly assumes client-side form validation ("confirm button is disabled or a validation error message is displayed"). Deduction: Step E6's expected result is ambiguous about whether validation is client-side or server-side — "no API request is sent" implies client-side, but the precondition does not explicitly state this is client-side validation. For a web surface, this distinction matters for test implementation. (-5) |

### 6. Internal Consistency — 115/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 45/60 | Five journey invariants are declared. Invariant 1 ("Sub-item code number is always automatically regenerated under the target main item") holds in Steps 2, 3, E3. Invariant 2 ("Status and assignee are never changed by the move operation") holds in Step 2 ("status and assignee remain unchanged"). Invariant 3 ("Move and code re-generation always execute within a single database transaction") — this is a strong claim. In Step E3, the concurrency scenario, the expected result says "The first transaction succeeds" and "The second transaction detects the sub-item's parent has changed and returns a conflict error." This is consistent — both transactions execute, and the second detects the conflict. Invariant 4 ("Closed/completed main items are never valid targets") holds in Step E1. Invariant 5 ("Moving to same parent is always rejected") holds in Step E2. **Concern**: Step E4 (move from deleted main item) — the expected result says "An error message is displayed indicating the source main item no longer exists; no move is performed." This is consistent with invariants but raises a question: if the source main item is soft-deleted, can the sub-item still be moved elsewhere? The invariant does not address this case. This is a minor gap, not a violation. (-15 for incomplete invariant coverage of edge cases) |
| Cross-Step references consistent | 40/50 | Happy path references are consistent: Step 1 establishes the sub-item under main item A, Step 2 moves it to B, Step 3 verifies under B. Setup section now includes "At least two PM users have access to the team (for concurrent operations)" — resolving iteration 1's complaint about E3's "two PM users" not being established. Setup also includes main item C in terminal status, established for E1. Step E4 references "Main item A has been soft-deleted by another user" — this is a new condition introduced during the scenario, not a setup precondition, which is acceptable for edge cases. Deduction: Step E3 references "Both users confirm the move to different targets" but the setup only establishes that two users exist — it does not establish what "different targets" means or how they are selected. The cross-reference is clear enough but could be more specific. (-10) |
| Risk level consistent with content | 30/40 | Risk level "High" is justified by: (1) State mutation — moving changes the parent relationship. (2) Code number regeneration — the sub-item gets a new identifier under the target. (3) Concurrent operation risk — two users moving the same sub-item could cause data conflicts. The risk classification comment mentions "data loss risk" — but moving a sub-item does not cause data loss (it can be moved back). The "irreversible" claim is also debatable. However, the concurrency risk and state mutation are sufficient to justify High. The risk level is defensible but the justification comment slightly overstates the risk. (-10 for imprecise risk justification) |

## Attack Points (Prioritized Fixes)

1. **Completeness**: Missing `not-found` API outcomes — api-handbook.md documents 404 NOT_FOUND for both "sub-item not found" and "target main item not found" but no step covers these. Add an edge case: "API request with non-existent sub-item bizKey" and "API request with non-existent target main item bizKey." Quote from api-handbook.md: `| 404 | NOT_FOUND | 子事项不存在 |` and `| 404 | NOT_FOUND | 目标主事项不存在 |`. Fix: Add Step E9 covering 404 scenarios for the move endpoint.

2. **Semantic Purity**: API endpoint coupling in Step E5 User Action — "The user sends PUT /api/v1/teams/:teamId/sub-items/:subId/move" contains implementation-level details (HTTP method + path). Quote from Step E5: `User Action: The user sends PUT /api/v1/teams/:teamId/sub-items/:subId/move`. Fix: Replace with "The user sends an API request to move a sub-item without proper authorization" and move endpoint details to a technical note or Contract test specification.

3. **Completeness**: Steps lack outcome identifiers — no numbered outcomes (e.g., Outcome 1.1, Outcome 2.1) for traceability in test generation. Quote: Expected Results are plain paragraphs without identifiers. Fix: Add outcome identifiers to each Expected Result block, e.g., "Outcome E3.A: The first transaction succeeds..." and "Outcome E3.B: The second transaction detects..."

4. **Internal Consistency**: Invariant 3 ("single database transaction") is a strong technical claim that is not verifiable from the Journey steps alone. Quote: `Move and code re-generation always execute within a single database transaction`. Fix: Either add a technical note referencing the implementation that guarantees transactional behavior, or soften the invariant to "Move operations are atomic — either the entire move succeeds or no changes are persisted."

5. **Precondition Exclusivity**: Step E3's two outcomes share identical preconditions — timing/ordering differentiates them but is not stated. Quote: Both E3.A and E3.B share precondition "Two PM users have the same sub-item's move dialog open simultaneously." Fix: Add a precondition variant: "Second user's request arrives after the first user's transaction has committed" for Outcome E3.B.

6. **Fact Alignment**: Edge cases E1, E2, E3, E4 lack `source: inferred` annotations. Quote from E1: No source annotation present, though E1 is partially derived from PRD Story 7 acceptance criteria. Fix: Add `<!-- source: derived from PRD Story 7 acceptance criteria ("尝试移动子事项到已关闭状态的主事项") -->` and similar for E2, E3, E4.

7. **Surface Fitness**: No explicit test strategy guidance distinguishing Contract vs Journey test levels. The journey mixes both without demarcation. Fix: Add a "Test Strategy" section stating "Steps 1-3 form the Journey smoke test (end-to-end workflow). Steps E1-E8 represent Contract-level individual behavior tests."
