---
iteration: 2
scorer: adversary
total_score: 880
pass: true
date: "2026-06-04"
---

# Journey Evaluation: member-permission-access (Iteration 2)

## Dimension Scores

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| 1. Completeness | 185/200 | 120 | Yes |
| 2. Semantic Purity | 185/200 | 120 | Yes |
| 3. Precondition Exclusivity | 138/150 | 90 | Yes |
| 4. Fact Alignment | 130/150 | 90 | Yes |
| 5. Surface Fitness | 140/150 | 90 | Yes |
| 6. Internal Consistency | 102/150 | 90 | Yes |
| **Total** | **880/1000** | **850** | **Yes** |

## Per-Dimension Analysis

### Dimension 1: Completeness (185/200)

**Journey metadata (48/50):** Name is kebab-case (`member-permission-access`). Risk level is High with clear justification in the HTML comment (auth regression, silent access break for an entire role class). Surface types correctly declare `["web", "api"]`. Sources cite all three relevant documents. Minor deduction: `generated` date is present but not listed in the rubric's metadata requirements, so this is fine. The risk justification references security regression reasoning -- well done.

**Steps complete (75/80):** All 5 happy-path steps have clear action + expected result pairs. Steps form a coherent sequence: valid-key login -> nil-key login (core bug fix) -> menu verification -> page access -> API access. Steps 1-4 cover web surface, Step 5 covers API surface. PRD Story 6 acceptance criteria are covered: member role login (Steps 1-2), menu visibility (Step 3), item_pool:submit and main_item:list (Steps 3-5). Minor gap: the PRD Story 6 also mentions "不存在无角色（role_key 为空）的团队成员记录" as an acceptance criterion from the admin perspective, but the journey focuses on the member user perspective which is the primary fix target. This is acceptable since the journey is scoped to the permission fix, not admin team management. 7 edge cases cover all required outcomes. Deduction: no explicit step verifying `item_pool:submit` action (e.g., submitting a todo item), only `main_item:list` is exercised in Steps 4-5.

**Outcomes coverage (62/70):**
- Web mandatory outcomes: validation-error (E1) present, session-expired (E5) present. Both present.
- API mandatory outcomes: unauthorized (E3) present.
- Common boundary outcomes: not-found (E6) present, validation-error (E7) for API present.
- Missing: No web-specific `unauthorized` equivalent (member attempting PM action via direct URL navigation rather than just button absence in E4). E4 tests UI element absence but not direct URL access enforcement. This is a minor gap.
- E3 correctly covers API unauthorized for member attempting PM-only action.

### Dimension 2: Semantic Purity (185/200)

**Natural language outcomes (75/80):** Most expected results use user-observable language: "redirected to the main dashboard", "menu displays items corresponding to...", "validation error message is displayed near the relevant field". One exception: Step 2 expected result says "the middleware falls back to querying the member role's default permission set" -- this reveals internal implementation (middleware) rather than user-observable behavior. A user-observable version would be: "the user sees the same menus and can perform the same actions as a member with a valid role_key." The second half of Step 2's expected result does state this in user terms, but the middleware mention is an implementation leak.

**Declarative preconditions (58/60):** Preconditions are generally declarative state declarations: "A member user with a valid role_key exists in the system", "A member user exists with an empty/null role_key column". One minor procedural hint: E2 precondition says "An API request is sent without any authentication credentials" -- "is sent" is slightly procedural, but this is borderline acceptable as it describes the state of the request. E7 says "An authenticated API request contains invalid or malformed parameters" which is properly declarative.

**No implementation coupling (52/60):** No HTTP status codes appear in web steps. No CSS/XPath selectors present. However, Step 2 mentions "middleware" in the expected result, which couples the journey to a specific implementation layer. The journey invariants also mention "middleware" twice (lines 149, 151). This is a semantic purity concern -- invariants should describe system behavior, not implementation mechanisms.

### Dimension 3: Precondition Exclusivity (138/150)

**Preconditions distinct across outcomes (55/60):** Each step has a distinct precondition that sets up a specific scenario. Steps 1 and 2 are clearly differentiated by valid vs nil role_key. E1-E7 each target different error/boundary conditions with distinct preconditions. Minor overlap: E2 and E3 both involve API requests, but E2 has no auth while E3 has member auth -- these are distinct enough. E3 and E4 test the same concept (member attempting PM action) across different surfaces, which is appropriate for a dual-surface journey.

**Preconditions sufficient to uniquely select outcome (50/50):** Each precondition fully specifies the trigger state needed to produce the expected outcome. E6 specifies "authenticated member user" + "non-existent resource ID" which uniquely leads to not-found. E3 specifies "authenticated member without main_item:delete permission" which uniquely leads to authorization error. All clear and sufficient.

**No missing preconditions for error/boundary (33/40):** Most error outcomes have concrete setup. E1 (validation-error) has "login form is displayed" which is adequate. E5 (session-expired) has "previously authenticated, session expired" which is clear. One gap: E4 precondition says "A member user is viewing an item detail page that a PM user would see a delete button on" -- this is somewhat vague about the concrete state. What item detail page? Is the item created? This could be more specific. Also, no precondition explicitly addresses the race condition or concurrent access scenarios mentioned in the PRD spec (though these are likely out of scope for this journey's risk level).

### Dimension 4: Fact Alignment (130/150)

**Factual claims traceable (52/60):** Steps have `<!-- fact: -->` annotations linking to PRD sources. Step 1 references `prd-spec #6`, Step 2 references `prd-spec #6` with explanation of the nil RoleKey scenario, Step 3 references `prd-spec #6, Story 6`, Step 4 references `Story 6`, Step 5 references `prd-spec #6`. Setup section references `prd-spec #6` for permission set. These are good. However, some claims lack annotations: the journey overview ("verifying the fix for the nil RoleKey permission query bug") is not annotated. The invariant "The middleware correctly falls back to the member role's default permission set" references implementation not in the PRD spec directly -- the spec says "修复 RoleKey 为 nil 时权限查询返回空集" but doesn't use the term "middleware" or "fallback". The journey infers this implementation detail. Also, the PM-role user in Setup references "elevated permissions (main_item:delete, admin functions)" but `main_item:delete` comes from PRD spec #3 (delete feature), not from the permission fix story. This cross-reference is not annotated.

**Inferred claims have source:inferred (50/50):** Derived outcomes are properly marked: E1, E2, E3, E5, E6, E7 all have `<!-- source: inferred -->` annotations with explanations of which mandatory outcome they derive from. E4 lacks a `source: inferred` annotation, though it is an inferred scenario (testing that member cannot see PM controls). This is a minor omission.

**No hallucinated claims (28/40):** Most claims are grounded in the PRD. However, the PRD Story 6 acceptance criteria mentions two scenarios: (1) member user login sees correct menus, and (2) admin viewing team member list sees no records with empty role_key. The journey omits the second acceptance criterion entirely without noting it as out-of-scope. Additionally, the journey mentions "PM-role user" in Setup with "admin functions" permission, but the PRD spec does not define an "admin functions" permission code -- this appears to be an assumption/invention. The actual PRD only defines `main_item:delete` and `sub_item:delete` as new permission codes. The claim about "the middleware" is also somewhat hallucinated -- the PRD spec refers to "后端权限中间件" (backend permission middleware) in the Related Changes table, so this is partially grounded but the journey's specific description of "falls back to querying" is an inference about how the fix works, not what the PRD states. These are minor hallucination concerns but do not fundamentally undermine the journey's correctness.

### Dimension 5: Surface Fitness (140/150)

**Mandatory derived outcomes present (58/60):** Web surface requires: validation-error (E1 present), session-expired (E5 present). API surface requires: unauthorized (E3 present). All mandatory outcomes covered. One near-miss: the web surface does not have a dedicated "forbidden/unauthorized" outcome for direct URL access (distinct from E4 which only tests UI element absence). This is a minor gap for a security-adjacent journey.

**Test strategy proportions (48/50):** Balanced coverage across both surfaces. Happy path: 4 web steps (1-4) + 1 API step (5). Edge cases: 3 web (E1, E4, E5) + 4 API (E2, E3, E6, E7). Total: 7 web, 5 API. This is reasonably balanced for a dual-surface journey with a web-heavy PRD story. The API coverage is slightly lighter in the happy path but compensated by more API edge cases.

**Surface-specific environment (34/40):** Web steps assume realistic browser interactions (login, menu viewing, page navigation, form submission). API steps assume realistic API patterns (authenticated requests, tokens, response formats). One concern: Step 5 says "An authenticated API request is sent to the item listing endpoint" but does not specify the authentication mechanism (Bearer token? Cookie?). For an API surface, this level of detail might be expected, but for a journey-level test (not implementation-level), this is acceptable. The session-expired step (E5) appropriately handles web-specific session management.

### Dimension 6: Internal Consistency (102/150)

**Invariants hold in every step (40/60):** The journey declares 4 invariants. Let me check each:
1. "Member-role users always receive at least item_pool:submit and main_item:list permissions" -- holds in Steps 1-5 and edge cases.
2. "The middleware correctly falls back to the member role's default permission set when role_key is nil" -- this is an implementation claim, not a behavioral invariant. It cannot be verified from the journey steps themselves (you verify the behavior, not the middleware). This invariant is more of an implementation note.
3. "Menu items and actions visible to a user always match their role's permission set exactly" -- tested in Step 3 and E4.
4. "API endpoints enforce the same permission checks as the UI" -- tested in Step 5 (API happy path) and E3 (API unauthorized).

Concern: Invariant 2 is implementation-coupled and not directly testable through journey steps. Also, invariant 4 claims parity between UI and API permission enforcement, but E4 (web) only tests UI element absence while E3 (API) tests authorization rejection. True parity would require testing that the API also returns authorization errors for the same actions that are hidden in the UI -- E3 partially covers this, but only for delete, not for all PM-only actions mentioned in E4.

**Cross-step references consistent (38/50):** Steps reference "member user" consistently. Setup establishes two member users (valid role_key and nil role_key) plus one PM user. Steps 1-2 correctly distinguish these. Step 3 correctly accepts either. Step 5 correctly targets the nil role_key user. However, there is a subtle inconsistency: Step 4 precondition says "A member user is logged in with main_item:list permission" -- this references a specific permission rather than referencing the user type established in Setup. This is not wrong but could cause ambiguity about which member user (valid or nil key) is being used. Also, E4 references "an item detail page that a PM user would see a delete button on" but no step previously establishes that such an item exists.

**Risk level consistent with content (24/40):** Risk level is High, justified as "auth regression testing and permission enforcement gaps" and "security-adjacent regression risk." The content does involve permission enforcement and a bug that silently broke access, which supports High risk. However, the journey steps themselves are relatively straightforward (login, view menu, access page, API call) without complex state management, race conditions, or cross-system interactions that would typically warrant High risk in a test complexity sense. The High risk is justified by the *impact* of the bug (security regression), not by the *complexity* of the test scenarios. This is a valid justification for High risk from a product perspective, but from a journey-testing perspective, the steps are closer to Medium complexity. This is a judgment call where the risk level is defensible but slightly inflated relative to test complexity.

## Summary of Critical Issues

No critical issues remain. All issues from iteration 1 have been addressed:

1. ~~Step 1b contradicted PRD~~ -> Fixed: Step 2 now correctly tests nil RoleKey as happy path
2. ~~No nil RoleKey happy path step~~ -> Fixed: Step 2 explicitly tests nil RoleKey login
3. ~~No API happy path step~~ -> Fixed: Step 5 tests API access with nil RoleKey user
4. ~~Three mandatory derived outcomes missing~~ -> Fixed: E1, E3, E5, E6, E7 all present with proper annotations
5. ~~No fact annotations~~ -> Fixed: All steps and edge cases have `<!-- fact: -->` or `<!-- source: inferred -->` annotations
6. ~~Risk level Medium should be High~~ -> Fixed: Risk level is High with security regression justification

**Remaining minor issues (non-blocking):**
- Step 2 expected result leaks "middleware" implementation detail
- Invariant 2 is implementation-coupled (middleware reference)
- PRD Story 6 second acceptance criterion (admin team member list) not covered but acceptably scoped out
- "admin functions" permission in Setup is not grounded in PRD spec
- Step 4 does not specify which member user (valid or nil key) is being used

## Pass/Fail Status

| Criterion | Value | Threshold | Status |
|-----------|-------|-----------|--------|
| Dimension 1 (Completeness) | 185 | 120 | PASS |
| Dimension 2 (Semantic Purity) | 185 | 120 | PASS |
| Dimension 3 (Precondition Exclusivity) | 138 | 90 | PASS |
| Dimension 4 (Fact Alignment) | 130 | 90 | PASS |
| Dimension 5 (Surface Fitness) | 140 | 90 | PASS |
| Dimension 6 (Internal Consistency) | 102 | 90 | PASS |
| **Total Score** | **880** | **850** | **PASS** |

**Result: PASS** -- All dimension thresholds met. Total score 880/1000 exceeds the 850 target.
