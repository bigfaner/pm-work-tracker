---
iteration: 2
scorer: adversary
total_score: 880
pass: true
date: "2026-06-04"
---

# Journey Evaluation: team-and-progress-visibility (Iteration 2)

## Dimension Scores

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| D1: Completeness | 180/200 | 120 | Yes |
| D2: Semantic Purity | 175/200 | 120 | Yes |
| D3: Precondition Exclusivity | 135/150 | 90 | Yes |
| D4: Fact Alignment | 130/150 | 90 | Yes |
| D5: Surface Fitness | 140/150 | 90 | Yes |
| D6: Internal Consistency | 120/150 | 90 | Yes |
| **Total** | **880/1000** | **850** | **Yes** |

## Per-Dimension Analysis

### D1: Completeness (180/200)

**Journey metadata (45/50):**
- Name is kebab-case: `team-and-progress-visibility`. Correct.
- Risk level Medium with justification referencing permission boundary enforcement and data leakage risk. Justified.
- Surface types declared: `["web", "api"]`. Correct per sources.
- Sources cited: references PRD stories, prd-spec, and api-handbook. Covers the relevant documents.
- Minor deduction: sources list does not call out the specific api-handbook section for Enhanced List Teams (`GET /v1/teams`) explicitly; it says "Enhanced Gantt View" which is a different endpoint. The team listing endpoint is the relevant one for this journey.

**Steps complete (80/80):**
- Step 1 (web team selector): action + expected result present, covers Story 11 AC1 (team selector).
- Step 2 (web weekly progress mixed): action + expected result present, covers Story 11 AC2 and prd-spec #16.
- Step 3 (web weekly progress no active terminal): action + expected result present, covers prd-spec #16 edge case.
- Step 4 (api list teams): action + expected result present, covers API surface for team filtering.
- Edge cases E1-E9: all have action + expected result, coherent sequence.
- Both PRD stories (10 and 11) are covered. Note: Story numbering in the journey's sources says "Story 10, Story 11" but the PRD file has these as Story 10 (gantt) and Story 11 (team selector + weekly progress). The journey correctly covers the content of PRD Story 11.

**Outcomes coverage (55/70):**
- Mandatory derived outcomes present: session-expired (E4), validation-error web (E5), unauthorized api (E6, E7). Good.
- Boundary outcomes: not-found (E8), validation-error api (E9). Good.
- Deduction: E6 precondition says "authenticated user without team listing permission" but the api-handbook shows `GET /v1/teams` requires "JWT only" auth -- there is no separate "team listing permission." This is a minor factual misalignment in the outcome setup. The unauthorized scenario should focus on unauthenticated access (E7 already covers this), making E6 somewhat redundant or its precondition questionable.
- Deduction: No outcome for the SuperAdmin edge case mentioned in the api-handbook ("SuperAdmin returns all teams"). This is a relevant boundary case for the team filtering story that is not covered.

### D2: Semantic Purity (175/200)

**Natural language outcomes (75/80):**
- Most expected results use user-observable language: "Only teams the user has permission to access are displayed," "The user is redirected to the login page," "An error message is displayed." Good.
- Minor issue: E6 says "The API returns an authorization error" and E7 says "The API returns an authentication error" -- these are close to implementation language but still describe observable behavior. Acceptable for API surface steps where the user is a developer consuming the API.
- E9 says "The API returns a validation error response describing the invalid parameter" -- acceptable but borderline on implementation coupling.

**Declarative preconditions (55/60):**
- Most preconditions are declarative state descriptions: "User has permission to access Team A and Team B only," "A completed main item had a sub-item update in the current week." Good.
- E4 precondition: "The user was previously authenticated and the session has since expired while viewing the weekly progress page" -- the "while viewing" clause is borderline procedural (describes a temporal sequence rather than a state). Minor.
- E5 precondition: "The team selector URL is manipulated to contain an invalid team identifier" -- "is manipulated" is somewhat procedural. Could be rephrased as "The team selector URL contains an invalid team identifier."

**No implementation coupling (45/60):**
- No HTTP status codes in web steps. Good.
- No CSS/XPath selectors. Good.
- However, API steps (E6-E9) reference "API returns an authorization error" / "authentication error" / "validation error response" / "not found error." While these describe the API contract rather than internal implementation, they are closer to implementation language than pure user-observable outcomes. For API surface types this is arguably appropriate, but the rubric asks for natural language outcomes. A 15-point deduction applies.
- No framework or component names leaked. Good.

### D3: Precondition Exclusivity (135/150)

**Preconditions distinct across outcomes (55/60):**
- Each edge case has a distinct precondition. E1 (one team), E2 (current week activity), E3 (previous week activity), E4 (session expired), E5 (invalid URL), E6 (no team listing perm), E7 (no auth token), E8 (non-existent team), E9 (invalid team ID format). All distinct.
- Minor overlap: E6 and E7 both deal with authorization/authentication failure. E6 specifies "without team listing permission" and E7 specifies "without valid credentials." These are distinguishable (authenticated-but-unauthorized vs unauthenticated), but as noted in D1, E6's precondition may not be realizable given the actual API design (JWT-only endpoint).

**Preconditions sufficient to uniquely select outcome (45/50):**
- Most preconditions fully specify the trigger state. For E2 and E3, the preconditions clearly differentiate current-week-only vs previous-week-only activity.
- E5 precondition ("URL is manipulated to contain an invalid team identifier") is sufficient to trigger a validation error, but it is not clear what constitutes "invalid" -- non-existent? non-numeric? malformed? The expected result says "invalid team identifier" which is circular.
- Deduction: E8 and E9 both involve an invalid/non-existent team ID but their preconditions differentiate by "does not exist in the database" vs "non-numeric value." These are sufficient.

**No missing preconditions for error/boundary (35/40):**
- All error outcomes have concrete setup preconditions. Good.
- Deduction: The SuperAdmin edge case (all teams returned) from the api-handbook has no corresponding outcome or precondition. This is a missing boundary case.

### D4: Fact Alignment (130/150)

**Factual claims traceable (50/60):**
- Fact annotations present on Steps 1-3, E2, E3 preconditions. Good practice.
- Step 4 references "api-handbook team endpoints" which is vague -- should reference the specific `GET /v1/teams` endpoint section.
- E6 references "api-handbook" but the api-handbook shows `GET /v1/teams` auth is "JWT only" with no separate permission check. The claim that a user could be authenticated but lack "team listing permission" is not traceable to any source document. This is a factual error.
- E8 references "not-found" as a "common boundary outcome" but does not cite the api-handbook which actually documents 404 NOT_FOUND error responses.

**Inferred claims have source (45/50):**
- E4, E5, E6, E7, E8, E9 all have `source: inferred` annotations with reasoning. Good.
- E6's inference is flawed: it derives from "API surface unauthorized mandatory outcome" but the actual API design (JWT-only auth on `GET /v1/teams`) does not support an authenticated-but-unauthorized scenario for team listing. The inference does not match the source material.
- E5's inference ("URL is manipulated") is reasonable as a web validation-error scenario.

**No hallucinated claims (35/40):**
- Most claims align with source material. No fabricated PRD content.
- E6's "without team listing permission" precondition appears to be fabricated -- the api-handbook specifies JWT-only auth for `GET /v1/teams`, not a specific permission code. The only authorization distinction mentioned is SuperAdmin vs regular user.
- The invariant "Activity for terminal items includes: status changes, sub-item creation or updates, and main item progress updates" correctly matches prd-spec #16 and Story 11 AC2.

### D5: Surface Fitness (140/150)

**Mandatory derived outcomes present (55/60):**
- Web mandatory outcomes: session-expired (E4), validation-error (E5). Both present.
- API mandatory outcomes: unauthorized (E6, E7). Both present.
- Deduction: E6's scenario may not be realizable per the actual API design, which weakens the unauthorized coverage. E7 (unauthenticated) is the more realistic unauthorized scenario.

**Test strategy proportions (45/50):**
- Web steps: Step 1, 2, 3, E1, E2, E3, E4, E5 = 8 steps
- API steps: Step 4, E6, E7, E8, E9 = 5 steps
- Ratio is roughly 60/40 web/api. Acceptable balance given the feature is primarily web-facing with API support.
- Deduction: API coverage is thinner -- no API step for weekly progress happy path (only the edge case E7/E8/E9 touch weekly progress via API). The journey declares the weekly progress API endpoint in sources but does not have a corresponding API happy-path step.

**Surface-specific environment (40/40):**
- Web steps assume browser interaction (login, page navigation, selectors). Realistic.
- API steps assume authenticated/unauthenticated HTTP requests. Realistic.
- No environment assumptions that contradict the declared surface types.

### D6: Internal Consistency (120/150)

**Invariants hold in every step (50/60):**
- Invariant 1 (team selector filters by permission): upheld in Steps 1, E1, E4, E5. Step 4 (API) also upholds it.
- Invariant 2 (non-terminal always displayed): upheld in Steps 2, 3.
- Invariant 3 (terminal hidden only when no activity in both weeks): upheld in Steps 2, 3, E2, E3.
- Invariant 4 (activity definition): consistent with Steps 2, 3, E2, E3 preconditions.
- Invariant 5 (Monday-Sunday boundaries): stated but never tested with a boundary condition (e.g., activity on Sunday vs Monday). Minor gap.
- Deduction: E6 is inconsistent with the invariant -- the invariant says "team selector always filters to only teams the current user has permission to access" but E6 describes a user who cannot list any teams, which contradicts the implicit assumption that authenticated users have at least one team.

**Cross-step references consistent (40/50):**
- Setup declares "User belongs to Team A and Team B, but not Team C." Steps 1 and 4 reference these consistently.
- Setup declares "At least one main item is in terminal status." Steps 2 and 3 reference terminal items consistently.
- E1 introduces a different user ("exactly one team") which is fine as an edge case.
- Deduction: E6 introduces yet another user profile ("without team listing permission") that is neither in the setup nor traceable to the source material. This creates a dangling reference.
- E8 references a "team ID that does not exist in the database" but setup only mentions Teams A, B, C. The "does not exist" ID is not established in setup.

**Risk level consistent with content (30/40):**
- Risk level Medium with "permission boundary enforcement with data leakage risk" justification. Appropriate.
- The content complexity is moderate: permission filtering + time-based visibility rules. Medium is a fair classification.
- Deduction: The journey covers both permission boundaries (team selector) and temporal filtering (weekly progress), which are two distinct concern areas. A journey covering two separate concerns might warrant either a higher risk level or being split into two journeys. The risk justification focuses only on the permission aspect and does not address the temporal filtering complexity.

## Critical Issues

None blocking. All dimensions pass their thresholds.

### Minor Issues (non-blocking, for future improvement)

1. **E6 precondition may not be realizable**: `GET /v1/teams` uses "JWT only" auth per api-handbook, meaning any authenticated user can list their teams. There is no separate "team listing permission" to lack. This step tests an impossible scenario. Consider replacing with a more realistic unauthorized scenario (e.g., a user who is not a member of any team, or removing this step in favor of E7).

2. **Missing SuperAdmin boundary case**: The api-handbook explicitly states "SuperAdmin returns all teams." No step or edge case covers this behavior, which is a notable boundary for the team filtering story.

3. **Missing API happy path for weekly progress**: The journey lists `GET /teams/:teamId/views/weekly` in sources but has no API happy-path step for it. Only edge cases E7-E9 touch weekly progress via API.

4. **Source annotation mismatch**: The sources list references "Enhanced Gantt View" section of api-handbook, but the relevant section for team filtering is "Enhanced List Teams."

5. **Week boundary not tested at edge**: Invariant 5 declares Monday-Sunday boundaries but no step tests activity at the boundary (e.g., activity on Sunday 23:59 vs Monday 00:00).

## Pass/Fail Status

| Criterion | Required | Actual | Status |
|-----------|----------|--------|--------|
| D1: Completeness | 120 | 180 | PASS |
| D2: Semantic Purity | 120 | 175 | PASS |
| D3: Precondition Exclusivity | 90 | 135 | PASS |
| D4: Fact Alignment | 90 | 130 | PASS |
| D5: Surface Fitness | 90 | 140 | PASS |
| D6: Internal Consistency | 90 | 120 | PASS |
| **Total** | **850** | **880** | **PASS** |

The journey passes all dimension thresholds and the total target of 850. The iteration-1 critical issues have been substantially addressed: API surface steps added, mandatory derived outcomes present, preconditions extracted, implementation coupling removed from invariants, and fact annotations added throughout. The remaining issues are minor and non-blocking.
