---
iteration: 1
total_score: 620
pass: false
dimensions:
  completeness: 140
  semantic_purity: 165
  precondition_exclusivity: 95
  fact_alignment: 75
  surface_fitness: 75
  internal_consistency: 70
---

# Journey Eval: sub-item-move — Iteration 1

**Date**: 2026-06-04

## Score Summary

| Dimension | Score | Min Threshold | Status |
|-----------|-------|---------------|--------|
| Completeness | 140/200 | 120 | PASS |
| Semantic Purity | 165/200 | — | PASS |
| Precondition Exclusivity | 95/150 | 90 | PASS |
| Fact Alignment | 75/150 | 90 | **FAIL** |
| Surface Fitness | 75/150 | 90 | **FAIL** |
| Internal Consistency | 70/150 | 90 | **FAIL** |
| **Total** | **620/1000** | **≥850** | **FAIL** |

**Overall: FAIL** — 3 dimensions below threshold.

## Dimension Details

### 1. Completeness — 140/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata complete | 45/50 | Name in kebab-case, risk_level High justified (state mutation + irreversible parent change). Sources listed. Missing `fact_table` reference — no explicit fact IDs cited anywhere. |
| Steps complete with required fields | 65/80 | Happy path Steps 1-3 have clear actions and outcomes. Edge case Steps 1b-5b each have preconditions, actions, and expected results. However, Steps lack numbered outcome identifiers (e.g., Outcome 1.1, 1.2), making traceability difficult for test generation. |
| Outcomes cover happy path + required derived scenarios | 30/70 | Happy path covered. Edge cases cover closed target, same-parent, concurrent move, deleted source, and unauthorized. However, **Web mandatory outcomes `validation-error` and `session-expired` are completely absent**. API mandatory `unauthorized` is present (Step 5b) but `validation-error` is not. This is a significant gap. |

### 2. Semantic Purity — 165/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 75/80 | Outcomes are generally well-written in natural language describing what the user/system observes. One minor violation: Step 5b mentions "POST /api/sub-items/:id/move" — an implementation-level API detail rather than a user-observable behavior description. |
| Preconditions are declarative statements | 55/60 | Preconditions are declarative: "A main item C is in terminal status", "Sub-item currently belongs to main item A". Good. Minor: Step 3b's precondition "Two PM users simultaneously attempt to move" is borderline procedural. |
| No implementation coupling in Step descriptions | 35/60 | Step 5b contains "POST /api/sub-items/:id/move" — an API endpoint detail. Step 2 mentions "NextSubCode" which is an implementation mechanism. Step 3b mentions "NextSubCode increments per request" — internal database behavior. These couple the Journey to implementation specifics. |

### 3. Precondition Exclusivity — 95/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes | 45/60 | Within individual Steps, Outcomes are distinct. However, Step 3b (concurrent move) has an ambiguous expected result — "first transaction succeeds; second either succeeds or fails gracefully" — which means two mutually exclusive outcomes are lumped into one expected result without distinct preconditions to differentiate them. (-20 deduction for ambiguous pair) |
| Preconditions sufficient to uniquely select Outcome | 35/50 | Steps 1, 2, 3 in the happy path have single outcomes, so no ambiguity. But Step 3b combines two possible outcomes (success or failure) without preconditions to determine which applies. The Journey does not specify what determines whether the second transaction succeeds or fails. |
| No missing Preconditions for error/boundary Outcomes | 15/40 | Step 5b's precondition is "A member-role user sends a move request via API" — but the precondition should specify what makes the member-role user unauthorized (e.g., "member-role user does not have sub_item:update permission"). The actual precondition for the 403 is implicit. Step 4b is good: "Main item A has been soft-deleted." |

### 4. Fact Alignment — 75/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 20/60 | Zero fact_id references exist in the entire Journey. No UNKNOWN markings. Claims like "code number is auto-regenerated using target's NextSubCode" and "move executes in a single transaction" are factual claims that need traceability. The PRD spec (Section "移动子事项 #9") confirms the transaction requirement and NextSubCode behavior, but the Journey does not cite this. |
| Inferred claims have required_outcomes rule support and source: inferred | 30/50 | Edge cases (closed target, same-parent) appear to be derived from PRD requirements but are not annotated with `source: inferred`. The concurrent move scenario is inferred from PRD concurrency requirements but lacks annotation. Step 5b (unauthorized) appears to be a required_outcomes derivation but has no `source: inferred` annotation. |
| No hallucinated claims without classification | 25/40 | Step 5b states "POST /api/sub-items/:id/move" — the actual endpoint is `PUT /api/v1/teams/:teamId/sub-items/:subId/move` per api-handbook.md. This is a **factual error**: wrong HTTP method (POST vs PUT) and wrong path structure. The claim about "403 Forbidden" for member-role users may also be inaccurate — the auth permission is `sub_item:update`, and whether a member has this permission depends on seed data, not role alone. (-30 for hallucinated incorrect endpoint claim) |

### 5. Surface Fitness — 75/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 25/60 | The Journey declares `surface_types: ["web", "api"]`. **Web mandatory `validation-error`: ABSENT**. **Web mandatory `session-expired`: ABSENT**. API mandatory `unauthorized`: present (Step 5b). API common `validation-error`: absent. Missing two mandatory Web outcomes is a critical gap. |
| Test strategy proportions match surface guidance | 25/50 | Web and API both require balanced 50/50 (Contract/Journey). The Journey leans heavily into happy path + edge cases without distinguishing between Contract-level individual behavior tests and Journey-level workflow tests. No guidance on how to split testing effort. |
| Surface-specific environment assumptions realistic | 25/40 | Web steps (1, 2, 3) correctly assume browser interactions (selector, navigation, detail page). API step (5b) has wrong HTTP method and path. Step 3b's concurrency scenario is realistic for both surfaces. However, the Journey does not clearly separate Web-flow steps from API-flow steps — it mixes them in a single sequence, making test generation ambiguous. |

### 6. Internal Consistency — 70/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 35/60 | Most invariants hold. However, Invariant 4 ("Closed/completed main items are never valid targets") is somewhat contradicted by Step 3b's "or fails gracefully" — if the second concurrent transaction could succeed moving to a target that becomes closed mid-transaction, this invariant could be violated. Invariant 5 ("Moving to same parent is always rejected") is consistent with Step 2b. Invariant 1 ("code number auto-regenerated using NextSubCode") is consistent across all steps. |
| Cross-Step references consistent | 20/50 | Step 3 references "the moved sub-item" and "main item B" which are established in Steps 1-2. However, Step 4b references "Main item A" being soft-deleted, but the setup section only mentions "A sub-item exists under main item A" — Step 4b's scenario introduces a new condition not established in setup or prior steps. Step 3b introduces "two PM users" but the setup only establishes one PM user. Cross-step consistency is weak for edge cases. |
| Risk level consistent with content | 15/40 | Risk level is "High" which is justified by the move being a state-mutation operation that changes parent relationships. However, the Journey does not clearly call out data loss risk or irreversibility in its risk classification comment — it mentions "data loss risk" but the actual operation is reversible (you could move the sub-item back). The risk level may be slightly overclassified but is defensible. |

## Attack Points (Prioritized Fixes)

1. **Surface Fitness**: Missing Web mandatory outcome `validation-error` — No step covers what happens when the move form is submitted with invalid data (e.g., empty target selection). Add a step: "User submits move without selecting a target" → "Error message displayed, form not submitted."

2. **Surface Fitness**: Missing Web mandatory outcome `session-expired` — No step covers session expiry during the move workflow. Add a step: "User's session expires while the move confirmation dialog is open" → "Appropriate redirect or message, unsaved state preserved or warning shown."

3. **Fact Alignment**: Step 5b contains incorrect API endpoint — "User sends POST /api/sub-items/:id/move" is wrong; the actual endpoint per api-handbook.md is `PUT /api/v1/teams/:teamId/sub-items/:subId/move`. Fix the HTTP method to PUT and correct the path.

4. **Fact Alignment**: No fact traceability — Add `fact_id` references or `source: inferred` annotations to outcomes. At minimum, cite the PRD spec section for the move workflow requirements and the api-handbook.md for endpoint details.

5. **Internal Consistency**: Step 3b ambiguous outcome — "first transaction succeeds; second either succeeds or fails gracefully" combines two possible outcomes. Split into separate outcomes with distinct preconditions: (a) "Second transaction succeeds — sub-item is re-numbered under second target" vs (b) "Second transaction fails — error returned because sub-item's parent has changed."

6. **Completeness**: Steps lack outcome identifiers — Add numbered Outcome identifiers (e.g., "Outcome 1.1", "Outcome 2.1", "Outcome 2.2") to each expected result block for traceability in test generation.

7. **Semantic Purity**: Implementation coupling in descriptions — Remove "NextSubCode" references from user-facing step descriptions. Replace with "code number is automatically regenerated." Keep implementation details in a separate technical notes section.

8. **Precondition Exclusivity**: Step 5b precondition insufficient — "A member-role user sends a move request" should specify "A user without sub_item:update permission" to make the precondition explicit and testable.

9. **Internal Consistency**: Setup does not establish all edge case preconditions — Add "Multiple PM users exist" to setup for Step 3b, and clarify that Step 4b's "soft-deleted" condition occurs during the active session.
