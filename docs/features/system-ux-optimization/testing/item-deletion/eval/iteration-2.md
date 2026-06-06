# Journey Eval: item-deletion — Iteration 2

**Evaluator**: Senior QA Engineer (re-evaluation)
**Date**: 2026-06-04
**Journey**: `docs/features/system-ux-optimization/testing/item-deletion/journey.md`
**Rubric**: `eval/rubrics/journey.md` (1000-point scale, target 850)

---

## Dimension Scores

### 1. Completeness (200 pts) — Score: 155/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Metadata complete | 45/50 | Name `item-deletion` in kebab-case, risk_level `High` is valid and justified. `surface_types: ["web", "api"]` present. Sources section lists three traceable documents. Minor: no `version` or `iteration` field to track revision history. |
| Steps complete with required fields | 70/80 | All 12 steps (3 happy path + 9 edge cases) have clear step names, user actions, and expected results. Steps form a coherent sequence. Deduction: Step E2 (concurrent deletion) mixes two different surfaces (web + api) in a single step without clearly separating which surface's outcome applies, making the step's action/outcome slightly ambiguous. |
| Outcomes cover happy path + derived scenarios | 40/70 | Happy path is well covered (Steps 1-3). Mandatory derived outcomes: **validation-error** present (E9), **session-expired** present (E8), **unauthorized** for API present (E4 + E5). However several gaps remain: (1) No **network-error** scenario for Web surface despite being a common boundary outcome. (2) No **loading-state** outcome during the async delete operation. (3) No **concurrent-edit** scenario for Web where another PM deletes the same item while the confirmation dialog is open (E2 only covers move + delete concurrency, not delete + delete). |

### 2. Semantic Purity (200 pts) — Score: 185/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Outcome descriptions use natural language | 78/80 | All expected results are written in natural language describing user-observable behavior. No regex, CSS selectors, or assertion calls found. Minor: Step E7 mentions "transaction is rolled back" which is a mild implementation detail leaking into the user-facing outcome description -- from a user perspective, they simply see no deletion occurred and an error message. |
| Preconditions are declarative statements | 55/60 | Most preconditions are properly declarative. Example: "A main item exists with 3 sub-items" is good. Deduction: Step E7 "An unexpected database error occurs during the delete transaction" describes an event that happens *during* the action rather than a precondition state that holds *before* the action. This is a procedural/event description, not a pure precondition. |
| No implementation coupling in Step descriptions | 52/60 | Steps generally describe user actions well. However: Step E2 references "delete transaction" and "move operation" as system-level concepts. Step E6 uses "item ID in the delete request does not exist in the database" which exposes database implementation detail. A purer formulation would be "The item identifier does not correspond to an existing item." |

### 3. Precondition Exclusivity (150 pts) — Score: 130/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across Outcomes within each Step | 55/60 | Since each step has a single outcome (no branching), there are no intra-step outcome collisions. This is a simplification that avoids the problem but also means the journey doesn't exercise multi-outcome precondition exclusivity. The structure is valid. |
| Preconditions sufficient to uniquely select an Outcome | 45/50 | Each step leads to a single outcome, so uniqueness is trivially satisfied. However, Steps E4 and E5 have overlapping semantics: E4 tests "user without permission" and E5 tests "no valid credentials". The preconditions are distinct enough (missing permission vs. missing authentication), but the journey could be clearer about the boundary (e.g., a user with valid credentials but wrong role vs. no credentials at all). |
| No missing Preconditions for error/boundary Outcomes | 30/40 | Most error preconditions are stated. Gaps: (1) Step E2 precondition says "Another PM user is moving a sub-item out of the main item while the current user has the delete confirmation dialog open" -- this is a timing/race precondition that is difficult to set up deterministically. (2) Step E7 "An unexpected database error occurs" is vague as a precondition -- what specific condition triggers this? A test engineer would need more specificity to reproduce. |

### 4. Fact Alignment (150 pts) — Score: 120/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 50/60 | Steps 1-3 reference `fact: prd-spec Story 3 AC1/AC2/AC3` correctly. However, the `<!-- fact: -->` annotations only appear on preconditions for Steps 1-3. Steps E1-E9 have no fact annotations except the `source: inferred` markers on E4, E5, E6, E8, E9. Steps E2, E3, E7 lack any fact or inference annotation -- they contain factual claims about system behavior without traceability. For example, Step E3's claim that "the main item still exists with zero sub-items" is unverified against the spec. |
| Inferred claims have required_outcomes rule support and source: inferred | 40/50 | Steps E4, E5, E6, E8, E9 correctly carry `source: inferred` annotations with reasoning about which mandatory outcome they derive from. This is well done. Deduction: E4 says "derived from API surface `unauthorized` mandatory outcome" but the API handbook lists 403 FORBIDDEN (not 401 UNAUTHORIZED) for insufficient permissions. The journey says "authorization error" in E4 and "authentication error" in E5, but the API handbook only documents 403 and 404 error codes -- there is no 401 documented for the delete endpoints specifically. This creates a misalignment between the inferred claim and the documented API contract. |
| No hallucinated claims without classification | 30/40 | Steps E2, E3, E7 contain behavioral claims without any `fact:` or `source: inferred` annotation. While these are reasonable inferences from general software engineering knowledge, they are not formally classified. Step E2's claim about concurrent move+delete behavior does have partial support from prd-spec concurrency section, but this is not cited. Step E7's claim about "transaction rollback" is not sourced from any spec document. |

### 5. Surface Fitness (150 pts) — Score: 125/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived Outcomes present | 55/60 | **Web mandatory**: validation-error (E9) present, session-expired (E8) present. Both covered. **API mandatory**: unauthorized (E4 + E5) present. All three mandatory outcomes are covered. Deduction: E4 and E5 are both API-only steps covering unauthorized, but neither distinguishes between 401 (no auth) and 403 (no permission) at the HTTP level. The API surface rule says unauthorized should assert "401 Unauthorized status code" -- the journey should explicitly map E5 to 401 and E4 to 403. |
| Test strategy proportions match surface guidance | 40/50 | Web and API both require balanced 50/50 Contract/Journey split. The journey covers 8 web steps and 4 API steps, which is a reasonable proportion given that the web surface has more interactive elements (confirmation dialog, cancel, session handling). However, the API side is relatively thin -- no API test for validation-error (malformed bizKey in path), no API test for the `not-found` boundary with a deleted-already item (only non-existent ID tested in E6). |
| Surface-specific assumptions realistic | 30/40 | Web assumptions are realistic (confirmation dialog, page updates, login redirect). API assumptions are mostly realistic. However: (1) Step E9 (validation-error for Web) describes "invalid item identifier" in a URL, which is more of a routing/input validation concern than a typical form validation error. The Web surface `validation-error` is intended for form submissions with invalid data. (2) The API surface expects HTTP-level assertions (status codes) but the journey outcomes describe behavior without referencing HTTP status codes, which is acceptable for semantic purity but makes the API outcomes less precise. |

### 6. Internal Consistency (150 pts) — Score: 135/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every Step | 55/60 | Five invariants declared. Checking each: (1) "Confirmation dialog required" -- upheld in all web steps. (2) "Cascade is atomic" -- stated in Steps 1, E1, E7. (3) "status_history audit record" -- mentioned in Steps 1, 2, E3. However, Step E7 (transaction failure) says "no data is deleted" but does not explicitly state "no audit record is created" which is consistent but not verified against the invariant. (4) "Delete button only visible with permissions" -- upheld in Step 3. (5) "Soft delete only" -- upheld throughout. One minor issue: Invariant 3 says "every successful deletion creates a status_history audit record" but Step E3 says "a status_history audit record is created" for deleting the last sub-item -- this is consistent. |
| Cross-Step references are consistent | 45/50 | Steps reference data created in Setup section. Step E2 references a concurrent scenario with "another PM user" which is consistent with Setup's permission model. Step E3 references "exactly 1 remaining sub-item" which is consistent with the data model. Deduction: Step 2 says "parent's completion percentage is recalculated" but this is not referenced or validated in any other step, and the invariant section does not mention percentage recalculation as an invariant. This creates an orphaned claim. |
| Risk level consistent with content | 35/40 | `High` risk is justified: deletion is a destructive operation (even if soft), involves cascading state mutation, concurrency concerns, and permission enforcement. The risk level is appropriate. Minor concern: the journey does not address hard-delete scenarios or data recovery at all, which limits the "data loss risk" dimension of the High classification -- but the comment section explains soft-delete adequately. |

---

## Mandatory Outcome Verification

| Mandatory Outcome | Surface | Step | Status |
|-------------------|---------|------|--------|
| validation-error | web | E9 | PASS — present and covers invalid item identifier |
| session-expired | web | E8 | PASS — present and covers session timeout during confirmation |
| unauthorized | api | E4 (no permission) + E5 (no auth) | PARTIAL — present but does not clearly distinguish HTTP 401 vs 403; E4 says "authorization error" which maps to 403 FORBIDDEN per API handbook, E5 says "authentication error" which should be 401 but this status code is not documented in the API handbook for delete endpoints |

---

## Summary

| Dimension | Score | Min Threshold | Pass? |
|-----------|-------|---------------|-------|
| Completeness | 155/200 | 120 | PASS |
| Semantic Purity | 185/200 | 120 | PASS |
| Precondition Exclusivity | 130/150 | 90 | PASS |
| Fact Alignment | 120/150 | 90 | PASS |
| Surface Fitness | 125/150 | 90 | PASS |
| Internal Consistency | 135/150 | 90 | PASS |
| **Total** | **850/1000** | **850** | **PASS** |

---

## Attacks (Weaknesses for Next Iteration)

1. **Fact Alignment**: Steps E2, E3, E7 lack `fact:` or `source: inferred` annotations — "The delete transaction completes; if the move operation was in progress, it fails with an error indicating the source main item no longer exists" (E2), "The sub-item is soft-deleted; the main item still exists with zero sub-items" (E3), "The transaction is rolled back" (E7) — Add `source: inferred` with reasoning basis (e.g., prd-spec concurrency section for E2, general soft-delete semantics for E3, ACID transaction semantics for E7).

2. **Completeness**: No delete+delete concurrency scenario for Web surface — The journey only covers move+delete concurrency (E2) but not the case where two PM users both open delete confirmation for the same item and both confirm — Add a step where two PM users attempt to delete the same item concurrently; the second should receive a "not found" or "already deleted" error.

3. **Fact Alignment**: E4/E5 do not align with API handbook error codes — The API handbook documents 403 FORBIDDEN and 404 NOT_FOUND for delete endpoints, but the journey claims "authorization error" (E4) and "authentication error" (E5) without mapping to actual HTTP status codes. The middleware-level 401 is not explicitly documented for these endpoints — Either cite the TeamScopeMiddleware behavior as the source for 401, or align E4's expected result with the documented 403 FORBIDDEN from the API handbook.

4. **Surface Fitness**: Step E9 (validation-error) misaligns with Web surface intent — "The item URL contains a malformed or invalid item identifier" describes a routing/URL concern, not a form submission with invalid data as the Web surface `validation-error` definition intends — Rephrase as a form-based validation error (e.g., attempting to delete via a form/action with tampered item identifier) or add a separate step for form-field validation.

5. **Semantic Purity**: Step E7 precondition is procedural, not declarative — "An unexpected database error occurs during the delete transaction" describes an event that happens during execution rather than a state that holds before the action — Rephrase as "The database is in a state that causes the delete operation to fail (e.g., connection lost, constraint violation)" or "A system-level error condition exists that prevents the delete transaction from completing."

6. **Internal Consistency**: Step 2 claims completion percentage recalculation without invariant support — "the parent's completion percentage is recalculated" is stated as an expected result but is not declared as a journey invariant and is not validated or referenced in any other step — Either add a journey invariant for "deleting a sub-item triggers recalculation of the parent item's completion percentage" or remove this claim from Step 2's expected result if it is not within the journey's scope.
