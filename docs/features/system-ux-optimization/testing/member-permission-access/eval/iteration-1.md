---
iteration: 1
scorer: adversary
total_score: 620
pass: false
date: "2026-06-04"
---

# Journey Eval: member-permission-access (Iteration 1)

## Dimension Scores

| Dimension | Score | Threshold | Status |
|---|---|---|---|
| 1. Completeness | 120/200 | 120 | PASS |
| 2. Semantic Purity | 155/200 | 120 | PASS |
| 3. Precondition Exclusivity | 75/150 | 90 | FAIL |
| 4. Fact Alignment | 85/150 | 90 | FAIL |
| 5. Surface Fitness | 80/150 | 90 | FAIL |
| 6. Internal Consistency | 105/150 | 90 | PASS |

## Pass/Fail Status

| Criterion | Required | Actual | Status |
|---|---|---|---|
| All dimensions pass threshold | Yes | 3 of 6 pass | FAIL |
| Total score >= 850 | 850 | 620 | FAIL |
| **Overall** | | | **FAIL** |

---

## Dimension 1: Completeness (120/200)

### Journey metadata (30/50)

- Name `member-permission-access` is kebab-case: PASS.
- `risk_level: Medium` is stated but the justification comment is weak. The comment says "multi-step interaction without irreversible side effects" which is generic and does not reference the specific risk of nil RoleKey being an auth regression that silently breaks access for an entire role class. A Medium risk justification should tie to the actual content.
- `surface_types: ["web", "api"]` is correct given the PRD covers both UI menu visibility and API permission enforcement.
- Sources cite two PRD documents. However, the PRD spec section reference is imprecise -- it should cite a specific section or requirement number (e.g., "#6: Fix nil RoleKey permission query bug").

**Deduction: -20** for weak risk justification and imprecise source citation.

### Steps complete (55/80)

- Happy path Steps 1-4 form a coherent sequence: login -> menu check -> item listing -> submission. This covers the core PRD Story 6 flow.
- Each step has an action and expected result.
- However, coverage gaps exist:
  - No step explicitly verifies that the nil RoleKey fallback actually fires (e.g., a step that confirms a user with empty `role_key` column specifically gets the default permission set). The happy path just says "member role" generically. The core bug scenario -- a user with nil RoleKey -- is never exercised in the happy path.
  - No step tests that a member-role user can perform `item_pool:submit` and `main_item:list` via API (only the web UI is tested in happy path, API only appears in edge cases for rejection scenarios).
  - The PRD states "Login succeeds for member-role users without permission errors" but Step 1 only checks "no permission errors during login" -- it does not check that the login response itself does not contain permission-related error fields.

**Deduction: -25** for missing nil RoleKey happy-path step and incomplete API happy-path coverage.

### Outcomes coverage (35/70)

Mandatory derived outcomes for declared surface types:

- **web -- validation-error**: Not present. No step tests form validation or input error on the web surface. For example, no step tests what happens if login credentials are malformed.
- **web -- session-expired**: Present in Step 4b. Covers redirect and no partial data. PASS.
- **api -- unauthorized**: Present in Step 2b (401). PASS.
- **not-found**: Not present. No step tests accessing a non-existent resource.
- **validation-error (api)**: Not present. No step tests sending malformed input to an API endpoint.

Only 2 of 5 expected boundary outcomes are covered.

**Deduction: -35** for missing validation-error (web), not-found, and validation-error (api) outcomes.

---

## Dimension 2: Semantic Purity (155/200)

### Natural language outcomes (60/80)

- Step 2b expected result says "API returns 401 Unauthorized" -- this is an HTTP status code embedded in a web/API step description. While this is an API surface step (not a web-only step), the rubric states "no HTTP status codes in web steps." Step 2b is API-surface, so the status code is borderline acceptable. However, Step 3b says "API returns 403 Forbidden" which is similarly an API step. For the web-surface steps (4b), the outcome is properly declarative: "redirected to login page or shown a session-expired message." This is acceptable.
- Step 1b expected result "No team member records exist with empty role_key" is a declarative observation. Good.
- Step 2 "menus outside the member role's permissions (e.g., delete, admin functions) are not visible" -- this is user-observable language. Good.
- However, the expected result in Step 3 references "items within the user's team scope" which introduces the concept of "team scope" without prior definition. This is a minor semantic impurity -- it assumes implementation knowledge.

**Deduction: -20** for "team scope" assumption and minor coupling in outcome descriptions.

### Declarative preconditions (55/60)

- Step 1b: "A user exists in the team member list with an empty role_key" -- declarative state. Good.
- Step 2b: "No authentication token is provided" -- declarative. Good.
- Step 3b: "Member user is logged in" -- declarative. Good.
- Step 4b: "Member user's session has expired" -- declarative. Good.
- Minor issue: Step 3b precondition ("Member user is logged in") is ambiguous about which user -- the nil RoleKey user from Step 1b or a generic member user. This is more of an exclusivity issue than a declarative issue.

**Deduction: -5** for minor ambiguity in Step 3b precondition reference.

### No implementation coupling (40/60)

- Step 2b action says "GET /api/menus" -- this is a specific API endpoint path. While this is necessary for API surface journeys, the exact path could change. A more resilient approach would reference the API operation semantically (e.g., "requests the menu list API endpoint").
- Step 3b action says "DELETE request via API to an item endpoint" -- this uses the HTTP method as a verb in the action description. Acceptable for API surface but borderline.
- No CSS selectors or XPath expressions. Good.
- No framework or component names in outcomes. Good.

**Deduction: -20** for hardcoded API paths that couple to implementation details.

---

## Dimension 3: Precondition Exclusivity (75/150)

### Preconditions distinct across outcomes (25/60)

- Step 2b ("No authentication token is provided") and Step 4b ("Member user's session has expired") produce different outcomes (401 vs. redirect-to-login) but the preconditions overlap semantically: both describe an unauthenticated state. The distinction is that one never had a token and the other had an expired one. This is borderline acceptable but could be clearer.
- Step 3b ("Member user is logged in") is identical to the implicit precondition of Steps 2-4 in the happy path. The only differentiator is the action (DELETE vs. GET). The precondition itself does not distinguish the outcome -- the action does. This means the preconditions are not sufficient to uniquely select the outcome by themselves.

**Deduction: -35** for overlapping preconditions between Steps 2b/4b and indistinct preconditions between Step 3b and happy path.

### Preconditions sufficient to uniquely select outcome (20/50)

- Step 1b preconditions are insufficient for the stated outcome. The precondition is "A user exists in the team member list with an empty role_key" but the expected result is "No team member records exist with empty role_key." This is contradictory -- the precondition sets up a state that the expected result says should not exist. The step seems to be testing that the system prevents nil RoleKey, but the PRD says the fix is to handle nil RoleKey gracefully, not to prevent it. This is a fundamental precondition-outcome mismatch.
- Step 3b does not specify which API endpoint or which resource the DELETE targets. Different resources may have different permission checks. The precondition is insufficient to uniquely determine the 403 outcome.

**Deduction: -30** for contradictory Step 1b preconditions and insufficient Step 3b specificity.

### No missing preconditions for error/boundary (30/40)

- Step 2b has clear preconditions. Good.
- Step 3b lacks a precondition specifying that the target resource exists and that the member user's role does not include delete permission. The latter is implied by "member role default permissions" in setup but should be explicit in the step precondition.
- Step 4b lacks a precondition specifying how long sessions last or how expiry is triggered. For reproducibility, a concrete setup (e.g., "session TTL has elapsed" or "session token is manually invalidated") is needed.

**Deduction: -10** for missing concrete preconditions in Steps 3b and 4b.

---

## Dimension 4: Fact Alignment (85/150)

### Factual claims traceable (25/60)

- The journey claims "default permissions including item_pool:submit and main_item:list" (Setup). This is traceable to PRD Spec #6. However, no explicit fact annotations (e.g., `[fact: prd-spec#6]`) are used. The traceability is implicit at best.
- The journey claims "the system ensures all members have a valid role assigned" (Step 1b expected result). This is NOT traceable to the PRD. The PRD says the fix handles nil RoleKey by falling back to default permissions -- it does NOT say the system ensures all members have a valid role. This appears to be an inferred or fabricated claim.
- The journey claims "Menu displays at minimum: todo item submission (item_pool:submit) and item listing (main_item:list)" (Step 2). This is partially traceable to the PRD but the specific menu labels ("todo item submission", "item listing") are not in the PRD. These are inferred labels.

**Deduction: -35** for missing fact annotations and untraceable claims in Step 1b and Step 2.

### Inferred claims have source: inferred (35/50)

- The menu labels in Step 2 are inferred but not marked as `[inferred]`. They should be annotated.
- Step 1b expected result ("the system ensures all members have a valid role assigned") contradicts the PRD and appears to be a hallucinated inference. It is not marked as inferred.
- The invariant "Users with empty/null role_key are never left without permissions" aligns with the PRD fix description and is a reasonable inference, but it is not marked as `[inferred]`.

**Deduction: -15** for missing inferred annotations on multiple claims.

### No hallucinated claims (25/40)

- Step 1b is the most concerning. The PRD explicitly states the fix is to fall back to default permissions when RoleKey is nil. The journey Step 1b instead claims "No team member records exist with empty role_key; the system ensures all members have a valid role assigned." This contradicts the PRD. The PRD does not say the system prevents nil RoleKey -- it says the system handles it gracefully. This is a hallucinated requirement.
- Step 2b claims "no sensitive information is leaked" in the 401 response. While this is a good security practice, it is not stated in the PRD. This is an unclassified inferred claim.

**Deduction: -15** for Step 1b hallucinated requirement contradicting the PRD.

---

## Dimension 5: Surface Fitness (80/150)

### Mandatory derived outcomes present (25/60)

- **web -- validation-error**: MISSING. No step tests form validation on the web surface.
- **web -- session-expired**: PRESENT (Step 4b).
- **api -- unauthorized**: PRESENT (Step 2b).
- For a dual-surface journey (web + api), 5 mandatory outcomes should be present (validation-error-web, session-expired-web, unauthorized-api, not-found, validation-error-api). Only 2 are present.

**Deduction: -35** for 3 missing mandatory derived outcomes.

### Test strategy proportions (30/50)

- Happy path: 4 steps, all web-surface.
- Edge cases: 4 steps, 3 are API-surface (2b, 3b), 1 is mixed (4b -- web redirect but session management is cross-cutting), 1 is ambiguous (1b -- admin views team list, web-surface).
- The API surface is only tested in rejection scenarios (401, 403). No API happy-path step exists (e.g., "API request to GET /api/items returns 200 with item list for member role user"). This creates an imbalanced test strategy where API is only tested for failure and web is only tested for success.
- The happy path should include at least one API success scenario to validate that the nil RoleKey fix works through the API surface as well.

**Deduction: -20** for imbalanced web/API coverage in happy path.

### Surface-specific environment (25/40)

- The setup section does not specify environment details for either surface. For web: no mention of browser assumptions, cookie handling, or redirect behavior. For API: no mention of token format (Bearer, JWT, etc.), header requirements, or response format.
- Step 4b (session-expired) is a web-surface outcome but sessions are also relevant to API. No step tests API token expiry (e.g., expired JWT returning 401). The session-expired outcome is only tested on the web surface.

**Deduction: -15** for missing surface-specific environment details and incomplete session coverage.

---

## Dimension 6: Internal Consistency (105/150)

### Invariants hold in every step (50/60)

- Invariant 1 ("Member-role users always receive at least item_pool:submit and main_item:list permissions after login") is validated by Steps 2-4. PASS.
- Invariant 2 ("Users with empty/null role_key are never left without permissions") is NOT directly validated by any step. No step sets up a user with nil RoleKey and then verifies they still get permissions. Step 1b claims the opposite (that no nil RoleKey users exist).
- Invariant 3 ("Menu items and actions visible to a user always match their role's permission set exactly") is validated by Step 2 (visible items match) and Step 3b (delete not visible). PASS.
- Invariant 4 ("API endpoints enforce the same permission checks as the UI") is partially validated by Step 3b (403 on DELETE) but there is no corresponding UI step showing the delete button is hidden for the same action. The cross-reference is implied but not explicit.

**Deduction: -10** for Invariant 2 not being validated by any step.

### Cross-step references consistent (30/50)

- Step 1b references "empty role_key" which contradicts the setup that says "the backend permission middleware has been fixed to handle nil RoleKey correctly." If the fix handles nil RoleKey, Step 1b should test that nil RoleKey users work, not that nil RoleKey users do not exist.
- The "member user" in Step 3b is ambiguous. Is this the same member user from the happy path? Is it a user with nil RoleKey? The reference is dangling.
- Step 4b references "protected page" without defining what a protected page is in the context of this journey. All pages accessed in Steps 2-4 are presumably protected, but this is not stated.

**Deduction: -20** for contradictory Step 1b reference and ambiguous user references.

### Risk level consistent with content (25/40)

- Risk level is Medium. The justification says "multi-step interaction without irreversible side effects." However, the journey is testing an auth regression fix -- a bug that could silently deny access to all member-role users. This is arguably higher risk than "Medium" suggests. The content involves permission enforcement, session management, and role-based access control. These are security-adjacent concerns that could warrant a High risk classification.
- Conversely, the journey steps themselves are straightforward (login, check menu, access page). No complex state transitions or error recovery is tested. Medium is defensible but the justification does not address the security dimension.

**Deduction: -15** for risk justification not addressing the security/regression dimension of the fix.

---

## Critical Issues

1. **Step 1b contradicts the PRD**: The PRD says the fix handles nil RoleKey by falling back to default permissions. Step 1b says "No team member records exist with empty role_key; the system ensures all members have a valid role assigned." This is the opposite of what the PRD specifies. The step should instead test that a user WITH nil RoleKey still gets default permissions and can log in successfully.

2. **Nil RoleKey scenario never tested in happy path**: The entire point of PRD Story 6 is fixing the nil RoleKey bug. The happy path never creates a user with nil RoleKey and verifies they can log in. This is the most critical gap -- the core bug scenario is untested.

3. **No API happy-path step**: The API surface is only tested for failure (401, 403). No step tests that a member-role user can successfully call API endpoints with their default permissions. For a dual-surface journey, this is a significant coverage gap.

4. **Three mandatory derived outcomes missing**: validation-error (web), not-found, and validation-error (api) are absent. For a dual-surface journey, these are required by the rubric.

5. **No fact annotations**: The journey makes factual claims traceable to the PRD but uses no `[fact: ...]` annotations, and several claims (menu labels, Step 1b outcome) are either inferred or hallucinated without classification.

---

## Recommendations for Iteration 2

1. **Replace Step 1b** with a happy-path step that explicitly creates a user with nil/empty `role_key` and verifies they can log in and receive default permissions. This is the core bug fix scenario.
2. **Add API happy-path step**: Test that a member-role user (with nil RoleKey) can successfully call `GET /api/items` and receives data, confirming the middleware fix works through the API surface.
3. **Add missing derived outcomes**: Add a web validation-error step (e.g., login with malformed email), an API not-found step, and an API validation-error step.
4. **Add `[fact: ...]` and `[inferred]` annotations** to all claims.
5. **Clarify risk justification** to address the security regression dimension.
6. **Make preconditions mutually exclusive**: Ensure each edge case step has a unique precondition that fully specifies the trigger state, without relying on the action to differentiate outcomes.
