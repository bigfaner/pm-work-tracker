---
iteration: 1
scorer: qa-journey-eval
date: "2026-06-04"
journey: item-deletion
total_score: 590
verdict: FAIL
---

# Journey Evaluation: item-deletion — Iteration 1

## Scores

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| 1. Completeness | 105/200 | 120 | BELOW |
| 2. Semantic Purity | 155/200 | 120 | PASS |
| 3. Precondition Exclusivity | 80/150 | 90 | BELOW |
| 4. Fact Alignment | 80/150 | 90 | BELOW |
| 5. Surface Fitness | 80/150 | 90 | BELOW |
| 6. Internal Consistency | 90/150 | 90 | PASS |
| **Total** | **590/1000** | **850** | **FAIL** |

---

## 1. Completeness (105/200)

### Journey metadata — 35/50

- Name `item-deletion` is valid kebab-case.
- Risk level `High` is present and justified by the comment (destructive operation with cascading deletes).
- Missing: no explicit `surface_types` mapping to individual steps. The journey mixes web and API steps without labeling which surface each step targets.

### Steps are complete — 45/80

- Steps 1, 2, 3 cover happy path for main item deletion, sub-item deletion, and permission-based visibility.
- Steps 1b through 5b cover important edge cases.
- Deficiencies:
  - No step covers deletion of an already-deleted item (re-soft-delete idempotency).
  - No step covers deleting a non-existent item (404 scenario).
  - Step numbering is inconsistent: Happy Path uses Step 1/2/3, Edge Cases uses Step 1b/2b/3b/4b/5b. Step 3 in Happy Path ("Non-PM user sees no delete button") is arguably an edge/boundary case, not happy path.
  - No explicit step for the session-expired scenario (mandatory for web surface).

### Outcomes cover happy + required derived scenarios — 25/70

- Happy path outcomes are present for the primary flows.
- Missing mandatory derived outcomes:
  - **validation-error** (web mandatory): No step tests form validation errors. While deletion is a confirm/cancel flow rather than a form, there is no scenario testing what happens if an invalid item ID is submitted via API or if the confirmation dialog receives unexpected input.
  - **session-expired** (web mandatory): Completely absent. No step covers a user whose session expires while the confirmation dialog is open or during the deletion request.
  - **unauthorized** (API mandatory): Step 4b covers "403 Forbidden" which is authorization (not authentication). The API `unauthorized` outcome requires a scenario with missing/invalid/expired credentials returning 401. Step 4b tests insufficient permissions (403), not missing authentication (401).
  - **not-found** (implied for both surfaces): No step tests deleting a non-existent item.

---

## 2. Semantic Purity (155/200)

### Outcome descriptions use natural language — 70/80

- All outcomes are in natural language describing what the user/system observes.
- Minor issue: Step 4b mentions "API returns 403 Forbidden" which is a status code reference. Acceptable for API surface but borderline for semantic purity.
- No regex patterns, CSS selectors, XPath, or framework assertions detected.

### Preconditions are declarative — 50/60

- Preconditions in edge cases are declarative ("Main item with sub-items is displayed", "Another PM user is moving a sub-item...").
- Setup section is declarative.
- Minor deduction: Step 2b precondition describes a concurrent action ("Another PM user is moving a sub-item out of the main item") which borders on procedural setup rather than pure state declaration. A more declarative form would be: "A sub-item move operation is in progress for the main item being deleted."

### No implementation coupling in Step descriptions — 35/60

- Steps describe user-level actions well ("clicks the delete button", "clicks cancel").
- Step 4b contains API endpoint detail: "Member user sends DELETE /api/main-items/:id" — this is implementation coupling (specific HTTP method and URL path in a Journey step).
- Step 2b mentions "transactions" and "source main item does not exist error" which leak implementation details about transaction handling.

---

## 3. Precondition Exclusivity (80/150)

### Preconditions distinct across Outcomes — 35/60

- Happy Path Steps 1, 2, 3 have implicit preconditions that are mostly distinct.
- Step 4b precondition ("A member-role user sends a DELETE request") overlaps with Step 3 precondition conceptually (both test non-PM access). Step 3 tests UI visibility, Step 4b tests API rejection — they are on different surfaces, but the journey does not explicitly separate them by surface.
- Step 1b and Step 1 share the same starting precondition ("main item with sub-items") without clearly stating what differentiates the paths (user clicks cancel vs. user clicks confirm).

### Preconditions sufficient to uniquely select Outcome — 25/50

- Happy Path steps do not have explicit precondition blocks. The preconditions are embedded in the user action descriptions or assumed from Setup. This makes it ambiguous what state is required for each step.
- Step 3 ("Non-PM user sees no delete button") does not state its precondition explicitly — it is only implied by the action description mentioning "member-role user."
- Step 3b ("Delete last sub-item") precondition is clear ("main item has exactly 1 remaining sub-item").

### No missing Preconditions for error/boundary Outcomes — 20/40

- Step 5b ("Delete transaction failure") precondition says "Database constraint or unexpected error occurs during the delete transaction" — this is the trigger itself, not a precondition describing the system state. A proper precondition would describe what exists before the action, and the error trigger would be part of the outcome.
- Step 2b precondition describes a concurrent action scenario but does not specify what data state must exist (e.g., how many sub-items, which specific sub-item is being moved).

---

## 4. Fact Alignment (80/150)

### Factual claims traceable to fact_id or marked UNKNOWN — 25/60

- No fact_id references exist anywhere in the journey.
- Multiple factual claims are made without traceability:
  - "status_history records the deletion event" (Steps 1, 2, 3b)
  - "soft-deleted in a single transaction" (Step 1)
  - "API returns 403 Forbidden" (Step 4b)
  - "deletion audit trail is not created" (Step 4b)
- None of these are marked UNKNOWN or marked as inferred.
- Sources are listed in frontmatter (prd-user-stories.md, prd-spec.md) but individual claims are not traced to specific source sections.

### Inferred claims have required_outcomes rule support — 25/50

- Step 4b (unauthorized via API) appears to be a derived outcome based on the API `unauthorized` required_outcomes rule, but it is not annotated with `source: inferred`.
- The concurrent deletion scenario (Step 2b) is an inferred boundary case with no `source: inferred` annotation.
- Step 5b (transaction failure) is an inferred boundary case with no annotation.

### No hallucinated claims — 30/40

- No obviously fabricated claims detected.
- All claims are plausible given the domain context.
- However, several claims lack classification entirely (neither factual with traceability nor inferred with rule support). The claim "deletion audit trail is not created" for a 403 response is stated as fact without evidence.

---

## 5. Surface Fitness (80/150)

### Mandatory derived Outcomes present — 20/60

**Web mandatory outcomes:**
- **validation-error**: ABSENT. No step covers submitting invalid data in a deletion context.
- **session-expired**: ABSENT. No step covers session expiration during the deletion workflow.

**API mandatory outcomes:**
- **unauthorized**: PARTIALLY PRESENT. Step 4b covers 403 Forbidden (authorization failure), but does not cover 401 Unauthorized (authentication failure — missing/invalid token). These are distinct scenarios per the API surface rules which explicitly require testing "Request to an authenticated endpoint without valid credentials" with 401 status.

Score: 0 for validation-error absence, 0 for session-expired absence, partial credit for unauthorized covering 403 but missing 401.

### Test strategy proportions — 30/50

- The journey lists `surface_types: ["web", "api"]` implying balanced coverage.
- 8 steps total: Steps 1, 2, 3, 1b, 2b, 3b are web-oriented; Step 4b is API-oriented; Step 5b is ambiguous (could be either).
- Only 1 step (Step 4b) explicitly tests the API surface. This is disproportionate for a dual-surface journey.
- Contract vs. Journey balance is not explicitly addressed.

### Surface-specific environment realistic — 30/40

- Web steps use realistic browser interactions (clicking buttons, confirmation dialogs).
- API step (4b) uses realistic HTTP method and endpoint pattern.
- Step 2b mixes web and API concerns (concurrent operations across sessions) without clarifying the execution model.

---

## 6. Internal Consistency (90/150)

### Invariants hold in every Step — 40/60

- Invariant "Deletion always requires a confirmation dialog" — Step 4b (API deletion) has no confirmation dialog. This could be a violation or an exception for API surface, but the invariant does not specify surface applicability.
- Invariant "Main item deletion always cascades to all sub-items" — Step 5b (transaction failure) confirms rollback, consistent.
- Invariant "Every successful deletion creates a status_history record" — consistent across steps.
- Invariant "Delete button is only visible to PM users" — Step 3 is consistent but Step 4b tests API access, not button visibility. No contradiction but the invariant wording assumes web surface only.

### Cross-Step references consistent — 30/50

- Steps are mostly self-contained with minimal cross-references.
- Step 2b references concurrent operations between "current user" and "another PM user" but does not reference any specific prior step's data.
- Setup section establishes preconditions for all steps, but individual steps do not reference each other's outputs (e.g., Step 2 does not say "using the main item from Step 1").

### Risk level consistent — 20/40

- Risk level `High` is justified by the destructive nature of deletion.
- However, the journey also includes non-destructive read-only scenarios (Step 3: "Non-PM user sees no delete button") which are more Medium/Low risk. The High classification is appropriate for the primary flow but the journey mixes risk levels without differentiation.

---

## Attack Vectors

### ATTACK 1: Completeness: Missing web mandatory outcome session-expired — "No step covers a user whose session expires while the confirmation dialog is open or during the deletion request." — Add a Step 6b with precondition "PM user's session has expired" where the user confirms deletion and the system responds with a session-expired redirect or message, preserving the dialog state for re-authentication.

### ATTACK 2: Completeness: Missing web mandatory outcome validation-error — "No step tests form validation errors" — Add a Step 7b testing deletion with an invalid or tampered item identifier (e.g., non-existent ID, malformed ID), expecting a validation error response with user-facing error message.

### ATTACK 3: Surface Fitness: API unauthorized covers 403 but not 401 — Step 4b says "API returns 403 Forbidden" but the API surface rules require 401 Unauthorized for "Request to an authenticated endpoint without valid credentials." — Add a separate step covering an unauthenticated DELETE request (no Bearer token or expired token) expecting 401 Unauthorized.

### ATTACK 4: Precondition Exclusivity: Happy Path steps lack explicit precondition blocks — Steps 1, 2, 3 embed preconditions in action descriptions instead of separate precondition fields, e.g. Step 1 says "PM user clicks the delete button on a main item that has 3 sub-items" — Extract preconditions into explicit `**Precondition:**` blocks for every step, matching the pattern used in Edge Cases steps.

### ATTACK 5: Fact Alignment: Zero fact_id references in entire journey — "status_history records the deletion event", "soft-deleted in a single transaction", "API returns 403 Forbidden" are all stated as facts without traceability — Annotate each factual claim with either a fact_id from the source documents or mark as UNKNOWN, and annotate derived boundary outcomes with `source: inferred`.

### ATTACK 6: Internal Consistency: Invariant "Deletion always requires a confirmation dialog" violated by API step — Step 4b has no confirmation dialog for API deletion, yet the invariant states "Deletion always requires a confirmation dialog before executing" — Either update the invariant to scope it to web surface only ("For web surface, deletion always requires a confirmation dialog"), or acknowledge the API exception explicitly.

### ATTACK 7: Precondition Exclusivity: Step 5b confuses trigger with precondition — "Database constraint or unexpected error occurs during the delete transaction" is the error trigger, not a precondition describing the initial system state — Rewrite precondition as the normal state ("Main item with sub-items exists; PM user has confirmed deletion") and move the database error to the outcome trigger.

---

## Summary

The journey covers the core deletion happy path well with clear natural language descriptions. However, it fails four of six dimension thresholds primarily due to:
1. **Missing mandatory derived outcomes** for both surfaces (validation-error, session-expired for web; 401 unauthorized for API)
2. **No fact traceability** — zero fact_id references or UNKNOWN/inferred annotations
3. **Weak precondition discipline** — happy path steps lack explicit precondition blocks, and one step confuses a trigger with a precondition
4. **Surface confusion** — a dual-surface journey with only 1 API-oriented step out of 8, and invariant wording that does not account for surface differences
