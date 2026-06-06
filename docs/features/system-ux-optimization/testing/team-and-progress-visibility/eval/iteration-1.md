---
iteration: 1
scorer: adversary
total_score: 790
pass: false
date: "2026-06-04"
---

# Journey Eval: team-and-progress-visibility — Iteration 1

## Dimension Scores

| Dimension                        | Score | Threshold | Pass |
|----------------------------------|-------|-----------|------|
| D1: Completeness                 | 150   | 120       | Yes  |
| D2: Semantic Purity              | 160   | 120       | Yes  |
| D3: Precondition Exclusivity     | 80    | 90        | No   |
| D4: Fact Alignment               | 105   | 90        | Yes  |
| D5: Surface Fitness              | 95    | 90        | Yes  |
| D6: Internal Consistency         | 130   | 90        | Yes  |
| **Total**                        | **720** | **850** | **No** |

## Per-Dimension Analysis

### D1: Completeness — 150 / 200

**Journey metadata (35/50):**
- Name is correctly kebab-case (`team-and-progress-visibility`).
- Risk level is stated as "Low" with an inline justification comment, which is good.
- Surface types include both `web` and `api`, which is correct.
- Sources are cited but use a single combined reference rather than linking individual facts to specific PRD sections. The citation format is adequate but not precise.

**Steps complete (65/80):**
- Happy path steps have actions and expected results with coherent sequencing.
- Story 10 (team selector filtering) is covered by Steps 1 and 1b.
- Story 11 (weekly progress hiding) is covered by Steps 2, 2b, 3, and 3b.
- Deduction: Steps 2 and 3 combine the user action with precondition-like language in the "User Action" field (e.g., "non-terminal main items exist with activity this week"). This conflation makes it harder to separate action from setup. Minor deduction for impure action descriptions.
- Missing: No step covers the boundary of a terminal item that transitions to terminal status within the current/previous week — does the status change itself count as activity? This is partially addressed in Step 3b but not explicitly for the transition scenario.

**Outcomes coverage (50/70):**
- Missing mandatory derived outcome `validation-error` for the web surface. While this is a read-only journey, the rubric requires boundary outcomes like `not-found` and `validation-error` for web surface types. There is no step exploring what happens with invalid team selection or a nonexistent team ID.
- Missing `unauthorized` outcome for the api surface. No step tests what happens when an unauthenticated or unauthorized user calls the API endpoints.
- Missing `not-found` boundary outcome (e.g., navigating to a weekly progress page for a team that does not exist).
- The edge cases present are good (single team, activity in current week only, activity in last week only) but do not cover the full set of mandatory derived outcomes.

### D2: Semantic Purity — 160 / 200

**Natural language outcomes (70/80):**
- Expected results are generally user-observable: "Only teams the user has permission to access are displayed", "The completed main item is displayed because it has activity in the current week". Good.
- Minor deduction: Step 3b mentions "status_history record within range" which is a data model / implementation term rather than a user-observable description. A user does not see "status_history records" — they see a status change. Should say something like "the item's status was changed last week".

**Declarative preconditions (50/60):**
- The Setup section uses declarative state ("User belongs to Team A and Team B, but not Team C"). Good.
- Edge case preconditions are declarative: "User has permission to exactly one team", "A completed main item had a sub-item updated this week but no activity last week". Good.
- Deduction: The preconditions for Steps 2 and 3 are embedded in the User Action field rather than stated separately as declarative preconditions. This mixes procedural action with state setup.

**No implementation coupling (40/60):**
- No HTTP status codes, CSS selectors, or XPath selectors appear. Good.
- Deduction: The term "status_history" in Step 3b and the invariants section is an implementation-level concept (database field name). The invariants section uses implementation language: "status_history changes", "sub-item creation/update". While invariants can reference domain concepts more precisely, "status_history" is a schema-level term, not a domain term visible to users.
- "Team dropdown selector" in Step 1 is borderline — it references a UI component type but is arguably a domain-level description of the interface element.

### D3: Precondition Exclusivity — 80 / 150

**Preconditions distinct across outcomes (35/60):**
- The edge cases do have distinct preconditions: single-team user vs. multi-team user; activity this-week-only vs. last-week-only vs. no activity. These are reasonably distinct.
- However, Steps 2 and 3 in the Happy Path do not have explicit, separate preconditions. Their setup conditions are embedded in the User Action field. This makes it impossible to formally evaluate distinctness.
- The preconditions for Steps 2b and 3b are similar enough to risk ambiguity: both describe "a terminal item with activity in exactly one of the two weeks." They differ only in which week and the type of activity, which is sufficient but borderline.

**Preconditions sufficient to uniquely select outcome (25/50):**
- The preconditions for edge cases are mostly sufficient. Knowing "activity in current week only" leads to "item is displayed"; knowing "no activity in either week" leads to "item is hidden."
- Missing: No precondition setup for what constitutes "activity" at the data level in a way that would let a test engineer deterministically set up the state. For example, Step 2b says "a sub-item updated this week" — but what kind of update? Any field change? A status change specifically? The PRD spec defines activity more broadly (status_history changes, sub-item creation/update, main item progress updates), but the journey does not set up enough detail to be deterministic.
- Missing: No precondition for the API surface type. All steps implicitly assume the web UI. There are no steps that set up API-specific preconditions (e.g., authentication token state, request headers).

**No missing preconditions for error/boundary (20/40):**
- There are zero error/boundary outcomes in this journey. No step explores unauthorized access, invalid inputs, or not-found scenarios. This is a significant gap since both surface types (web and api) are declared and the rubric requires error outcomes.
- Deduction is heavy here because the journey completely omits the error category.

### D4: Fact Alignment — 105 / 150

**Factual claims traceable (40/60):**
- The journey references PRD Stories 10 and 11 in the sources, and the content aligns with those stories.
- However, individual claims within steps are not annotated with fact references (e.g., `[fact: prd-spec#16]`). The rubric expects traceability at the claim level, not just at the document level.
- The PRD spec #16 about activity definition and week boundaries is reflected in the invariants but not explicitly cited.

**Inferred claims have source (35/50):**
- Several claims are inferred rather than directly stated in the PRD:
  - "page layout is clean without gaps" (Step 3) — not in the PRD, and not marked as inferred.
  - "the selector may still function as a dropdown for consistency" (Step 1b) — reasonable inference but not marked.
  - The specific behavior of "displayed because it has activity in the current week" is a logical inference from the spec but not explicitly marked as such.
- The journey does not use any `inferred` or `derived` annotations to distinguish these.

**No hallucinated claims (30/40):**
- No clearly fabricated content detected. All claims are reasonable inferences from the PRD.
- Minor concern: "page layout is clean without gaps" is a UX assertion that goes beyond what the PRD specifies. It is not hallucinated per se, but it is an unsubstantiated claim presented as an expected result.

### D5: Surface Fitness — 95 / 150

**Mandatory derived outcomes present (30/60):**
- Web surface: Missing `validation-error` and `session-expired` outcomes. The journey covers only happy-path and mild edge cases. For a web surface type, the rubric requires at minimum `validation-error` and `session-expired` derived outcomes.
- API surface: Missing `unauthorized` outcome. The journey declares `api` as a surface type but includes zero API-specific steps or outcomes.
- No steps exercise API endpoints (GET /teams, GET /teams/:teamId/views/weekly) despite these being documented in the API handbook.

**Test strategy proportions (35/50):**
- Coverage is skewed entirely toward the web surface. All steps describe web UI interactions (opening dropdowns, viewing pages).
- The API surface is declared but has zero steps exercising it. This is a significant imbalance.
- The web coverage itself is reasonably balanced across the two stories (team selector vs. weekly progress).

**Surface-specific environment (30/40):**
- The web environment assumptions are realistic (team dropdown, weekly progress page).
- No API environment is described at all: no mention of authentication tokens, request/response formats, or API error handling.
- The journey would benefit from at least acknowledging the API context even if web is the primary surface.

### D6: Internal Consistency — 130 / 150

**Invariants hold in every step (50/60):**
- The declared invariants are consistent with the steps. The team filtering invariant holds across Steps 1 and 1b. The weekly progress hiding invariant holds across Steps 2, 2b, 3, and 3b.
- Minor issue: The invariant "Non-terminal main items are always displayed" is stated but never explicitly tested by a step that has ONLY non-terminal items with no activity. Step 3 comes close but does not set up a scenario where a non-terminal item has no activity — it just says "all terminal items had no activity." The invariant should be tested more explicitly.

**Cross-step references consistent (45/50):**
- No dangling or ambiguous references. Steps refer to consistent entities (Team A, Team B, Team C, terminal/non-terminal items).
- The setup section provides the shared context and all steps reference it consistently.

**Risk level consistent with content (35/40):**
- "Low" risk classification is reasonable: both workflows are read-only display operations as justified in the comment.
- However, the absence of error/boundary testing could be considered a risk understatement if permission filtering is security-sensitive. If a user could see teams they should not, that would be a data leakage issue. The risk level might more accurately be "Medium" given the permission boundary. This is a judgment call and the deduction is minor.

## Critical Issues

1. **Missing mandatory derived outcomes (D1, D5):** The journey declares both `web` and `api` surface types but includes zero API-specific steps. Missing `unauthorized` for API, `validation-error` and `session-expired` for web. This is a structural gap that causes failures across two dimensions.

2. **No error/boundary outcomes (D3):** Every error outcome category is empty. No unauthorized access, no invalid team ID, no not-found scenario, no expired session. The journey covers only the happy path and mild edge variations. This makes the precondition exclusivity dimension impossible to satisfy for error states.

3. **Implementation coupling in invariants (D2):** The term `status_history` is a database schema concept leaking into the journey specification. Invariants should express domain-level guarantees, not reference internal data structures.

4. **Claim-level fact traceability absent (D4):** The journey cites source documents at the top level but does not annotate individual claims with fact references. The rubric expects claims to be traceable to specific PRD sections.

5. **Preconditions embedded in User Action (D1, D3):** Steps 2 and 3 mix setup state with the user action, making it impossible to evaluate precondition distinctness and sufficiency formally.

## Pass/Fail Status

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| D1: Completeness | 150 | 120 | PASS |
| D2: Semantic Purity | 160 | 120 | PASS |
| D3: Precondition Exclusivity | 80 | 90 | **FAIL** |
| D4: Fact Alignment | 105 | 90 | PASS |
| D5: Surface Fitness | 95 | 90 | PASS |
| D6: Internal Consistency | 130 | 90 | PASS |
| **Total** | **720** | **850** | **FAIL** |

**Overall: FAIL** — Total score 720 is below the 850 target. D3 (Precondition Exclusivity) fails its dimension threshold at 80/90. The primary drivers are the complete absence of error/boundary outcomes, missing API surface steps, and lack of explicit precondition separation.

## Recommendations for Iteration 2

1. Add API-specific steps: at minimum, test GET /teams (authorized vs. unauthorized) and GET /teams/:teamId/views/weekly with various auth states.
2. Add error/boundary outcomes: `unauthorized` (api), `session-expired` (web), `validation-error` (web), `not-found` (both surfaces).
3. Extract preconditions from User Action fields into separate, declarative precondition blocks for each step.
4. Replace `status_history` with user-facing language like "status change recorded" in outcomes and invariants.
5. Add `fact:` annotations to key claims, and mark inferred claims with `inferred:` labels.
6. Consider whether "Low" risk is accurate given the permission boundary security implications.
