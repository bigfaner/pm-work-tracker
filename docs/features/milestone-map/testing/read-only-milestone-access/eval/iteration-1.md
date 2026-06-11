---
iteration: 1
evaluator: QA-Adversary
date: "2026-06-08"
total_score: 720
---

# Journey Eval: read-only-milestone-access — Iteration 1

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| 1. Completeness | 155/200 | 120 | PASS |
| 2. Semantic Purity | 160/200 | 120 | PASS |
| 3. Precondition Exclusivity | 100/150 | 90 | PASS |
| 4. Fact Alignment | 110/150 | 90 | PASS |
| 5. Surface Fitness | 100/150 | 90 | PASS |
| 6. Internal Consistency | 95/150 | 90 | PASS |

---

## 1. Completeness — 155/200

### Metadata (35/50)

- journey name, risk_level, surface_types, surface_keys, sources, generated date: all present
- surface_types lists only `["web"]` — this journey is frontend-only, which is correct for a read-only viewing journey
- Deduction: no surface_type for backend/API. While the journey is web-only by intent, the PRD explicitly mentions backend returns 403 (prd-spec line 228: "milestone:read 缺失时 /milestones 页面返回 403"). No backend contract test surface is declared. (-10)
- Deduction: sources list prd-spec and prd-user-stories but not prd-ui-functions.md, which contains the detailed RBAC visibility rules. (-5)

### Steps (65/80)

- Happy path covers 6 steps: list view, empty state, timeline view, detail panel, tooltip/hover, MI navigation. Solid coverage of the read-only viewing workflow.
- Edge cases cover 7 additional scenarios: API error retry, access denied, read-only info display, tooltip overflow, filters, MI navigation from panel, session expired.
- Deduction: Step 3 (timeline view) does not verify breadcrumb navigation back to list. This is a distinct read-only behavior (prd-user-stories Story 14, final AC: clicking breadcrumb returns to list). (-5)
- Deduction: Step 4 (detail panel) does not verify panel open/close mechanics for read-only users. Story 14 AC says panel opens, but the journey does not test overlay-click-to-close or Escape key in read-only mode. (-5)
- Deduction: No step verifies the milestone map card hover effect (border highlight + shadow) in read-only mode. Story 9 AC mentions this explicitly. (-5)

### Outcomes (55/70)

- Each step has clear Expected Result sections describing what is visible and what is NOT visible.
- Deduction: Step 5 (tooltip/hover) expected result says "All hover interactions work normally" without specifying what "normally" means. Compare with PRD Story 9 which specifies "Tooltip（'X 个事项，Y 已完成'），节点背景高亮". The expected result should repeat the specific tooltip content. (-5)
- Deduction: Step E1 (session expired) says "user is redirected to the login page" but does not specify the post-login behavior precisely enough. PRD specifies the user must re-navigate. The step says "must re-navigate" but does not verify what happens if the user tries to use browser back button or bookmark — this is a realistic web scenario. (-5)
- Deduction: Edge Case Step 1b (API error) says "retryable error message with a retry control" but does not specify the exact behavior after retry succeeds. Does the page reload fully? Does it restore filter state? (-5)

---

## 2. Semantic Purity — 160/200

### Natural Language (70/80)

- Steps use user-facing language: "navigates to", "views", "clicks", "hovers". Good declarative style.
- Deduction: Step 1b precondition says "Backend returns an error or times out" — this is implementation language describing backend behavior, not a user-observable precondition. A more natural phrasing: "The server is unavailable or unresponsive." (-5)
- Deduction: Step E1 precondition: "The user was previously authenticated and the session has expired" — acceptable but mixes technical (session) with user-facing. (-5)

### Declarative Preconditions (50/60)

- Preconditions exist for Steps 2, 1b, 2b, 3b, 4b, 5b, 6b, E1. Steps 1, 3, 4, 5, 6 in Happy Path lack explicit preconditions (they rely on the global Setup section).
- Deduction: Step 1 (list page) has no precondition distinguishing it from Step 2 (empty state). The implicit precondition is "at least one milestone map exists" but it is not stated. (-5)
- Deduction: Step 3b and Step 3 are nearly identical scenarios — both have a user with milestone:read only on timeline — but the precondition in 3b ("User has milestone:read only") restates the global Setup, making it redundant rather than additive. (-5)

### No Impl Coupling (40/60)

- The journey avoids coupling to specific API endpoints, DOM selectors, or implementation details.
- Deduction: Step 1b references "Backend returns an error" which couples the test to backend error semantics rather than describing what the user observes (error message on screen). (-5)
- Deduction: Step E1 references "session has expired" which is an implementation detail. A user-observable description would be "the user's login has timed out." (-5)
- Deduction: The note at line 144 references "validation-error" outcome by name, which is a test framework concept leaking into the journey definition. (-5)
- Deduction: Several steps reference specific controls by their UI implementation names ("breadcrumb", "filter bar", "zoom controls") which is acceptable for web surface but borders on impl coupling when the PRD defines these as functional behaviors rather than widget names. (-5)

---

## 3. Precondition Exclusivity — 100/150

### Distinct (40/60)

- Most preconditions are unique and test different scenarios.
- Deduction: Setup precondition "User has milestone:read permission but no milestone:create, milestone:update, or milestone:delete permissions" is the default for ALL happy path steps. Steps 3b, 5b restate this same condition explicitly, creating redundancy without adding exclusivity. (-5)
- Deduction: Step 2b precondition "User does not have milestone:read permission" directly contradicts the Setup section. This is intentional for negative testing, but there is no explicit step to transition from the Setup state to this negative state. An agent executing this step needs to understand it is a separate test run with different credentials. (-5)
- Deduction: Step 2 (empty state) precondition "Team has 0 milestone maps" and Step 1 (implicit "team has milestone maps") are mutually exclusive but Step 1 never states its implicit precondition explicitly. (-5)
- Deduction: No precondition distinguishes between a user who has SOME mutation permissions (e.g., milestone:create but not milestone:update) vs. a user who has NONE. The journey only tests the all-or-nothing case. (-5)

### Sufficient (35/50)

- Preconditions cover the main scenarios but miss important variations.
- Deduction: No precondition for a milestone map in terminal state (completed/cancelled). The PRD business rules (BIZ-milestone-005: "Parent-terminal blocks child operations") suggest that viewing a terminal milestone map may have different read-only behavior (e.g., cancelled map shows grey styling per Story 8). (-5)
- Deduction: No precondition for a milestone with cancelled status. The PRD specifies cancelled milestones display grey styling, empty MI list, and no add button. This is a distinct read-only viewing scenario not covered. (-5)
- Deduction: No precondition for different zoom levels (week/month/quarter). Step 5b mentions "zoom controls" but does not precondition a specific zoom state. (-5)

### Missing for Errors (25/40)

- Error preconditions exist for API error (Step 1b), access denied (Step 2b), session expired (Step E1).
- Deduction: No precondition/step for network connectivity loss during read-only browsing (distinct from server error). (-5)
- Deduction: No precondition/step for milestone map data that is partially loaded (some milestones render, others fail). (-5)
- Deduction: No precondition for concurrent data changes — another user deletes a milestone map while the read-only user is viewing it. (-5)

---

## 4. Fact Alignment — 110/150

### Traceable (45/60)

- Most facts reference prd-spec Story 14 and general prd-spec RBAC rules.
- Deduction: Step 1 Expected Result references "RBAC controls visibility, milestone:create required for create buttons" but this fact traces to a general PRD statement, not a specific Story 14 AC. The PRD Story 14 AC line 292 explicitly says "页面不显示'+ 创建里程碑图'按钮和虚线创建卡片" — the journey correctly captures this but the fact annotation is imprecise. (-5)
- Deduction: Step 3 Expected Result mentions "breadcrumb" but Story 14 AC does not mention breadcrumb behavior specifically. The breadcrumb is mentioned in Story 9 AC. The fact annotation should trace to Story 9, not Story 14. (-5)
- Deduction: Step 5 Expected Result mentions "node tooltips with item summaries" but the fact annotation is missing. Story 9 AC specifies "Tooltip（'X 个事项，Y 已完成'）" — this is a specific format that should be traced. (-5)

### Inferred with Rules (35/50)

- Step 1b (API error) is marked as "inferred — derived from Web surface server-error boundary outcome." This is a valid inference from the Web surface testing strategy.
- Step E1 (session expired) is marked as "inferred — derived from Web surface session-expired mandatory outcome." Valid inference.
- Deduction: Step 3b (read-only info displays) has no source annotation. It appears to be inferred from Story 14 AC but is not marked as such. (-5)
- Deduction: Step 4b (tooltip overflow) has no source annotation. The PRD Story 9 AC mentions description overflow with Tooltip. Should be annotated. (-5)
- Deduction: Step 5b (filters in read-only) claims filters work "identically to users with edit permissions" but this is not explicitly stated in any PRD AC. This is an inference that should be marked and justified. (-5)

### No Hallucinated (30/40)

- Most facts are grounded in PRD source material.
- Deduction: Step 3 mentions "info card" as part of timeline rendering. The PRD Story 9 AC calls it "基本信息卡片" (basic info card). The journey says "info card" which is close but a slight simplification that could cause confusion. (-2)
- Deduction: Step 4 mentions "add control" as something NOT shown in the detail panel. The PRD Story 10 AC refers to "+ 添加" button. "Add control" is ambiguous — it could be interpreted as any add mechanism rather than the specific "+ 添加" button. (-3)
- Deduction: Step 5b mentions "refresh" as a read-only interactive control. The PRD does not explicitly distinguish refresh as a read-only vs. mutation control. Including it without source justification risks testing implementation assumptions. (-5)

---

## 5. Surface Fitness — 100/150

### Mandatory Outcomes (45/60)

- Web surface mandatory outcomes: validation-error, session-expired, unauthorized.
- Session-expired: Covered in Step E1. (pass)
- Unauthorized: Covered in Step 2b (access denied without milestone:read). (pass)
- Validation-error: The journey note at line 144 explicitly states "validation-error outcomes are not applicable" because "no forms or mutation controls are available." This is a reasonable justification for a read-only journey.
- Deduction: While validation-error on forms is not applicable, there is no step testing what happens if a user manipulates URL parameters (e.g., /milestones/invalid-id) in read-only mode. This is a web-specific validation scenario. (-5)
- Deduction: No step tests the unauthorized response for timeline view specifically. Step 2b only tests the list page (/milestones). What happens when a user without milestone:read navigates directly to /milestones/:mapId? (-5)
- Deduction: Step E1 says session expired applies when "navigating to a new page or interacting with the timeline" but does not test session expiry specifically during data-heavy operations like timeline rendering with many milestones. (-5)

### Strategy Proportions (30/50)

- The journey has 6 happy path steps + 7 edge cases + 1 session-expired = 14 total steps.
- All steps are web surface. No API/backend surface steps despite the journey involving API calls.
- Deduction: The journey declares `surface_types: ["web"]` and `surface_keys: ["frontend"]` only. The PRD Story 14 AC line 296 specifies "页面返回 403 提示" which implies a backend API response. A backend/API surface with unauthorized outcome testing would strengthen coverage. (-5)
- Deduction: No step tests backend API behavior directly (e.g., calling GET /api/milestone-maps with read-only token and verifying the response does not include mutation-capability fields). (-5)
- Deduction: Proportion of positive (6 steps) to negative (7 edge cases) is close to 1:1.2, which is good. However, the negative cases are thin — many just verify "same as for users with edit permission" without specifying what that means. (-5)
- Deduction: No step tests the breadcrumb navigation in read-only mode. This is a core navigation surface behavior. (-5)

### Realistic Assumptions (25/40)

- The journey assumes a straightforward user flow through list → timeline → detail panel.
- Deduction: Assumes the user arrives at /milestones via direct navigation, but does not test navigation from other pages (e.g., clicking milestone badge in /items list). (-5)
- Deduction: Assumes milestone:read permission exists as an independent permission. The PRD confirms this (RBAC line 53), so this is valid. No deduction.
- Deduction: Assumes that "read-only" means all mutation controls are hidden (not disabled). The PRD Story 14 AC line 294 mentions "+ 创建里程碑" button could be "不显示或显示为禁用态" (hidden OR disabled). The journey assumes hidden only, which is a stronger assumption than the PRD allows. (-5)
- Deduction: Does not test the scenario where a user has milestone:read + milestone:create but NOT milestone:update. This partial-permission scenario is realistic (e.g., a junior PM who can create but not edit) and is not covered. (-5)

---

## 6. Internal Consistency — 95/150

### Invariants (40/60)

- 6 journey invariants are declared. Most are consistent with the steps.
- Deduction: Invariant "All read-only interactions (hover, tooltips, navigation, filters, zoom) work identically regardless of mutation permissions" — this is stated as an invariant but Step 5b is the only step that tests it (partially). No step tests zoom specifically. The invariant is asserted more broadly than it is verified. (-5)
- Deduction: Invariant "API errors always display a retry option, never a blank page" — Step 1b verifies this for the list page but not for the timeline view. The invariant claims "always" but coverage is partial. (-5)
- Deduction: Invariant "Users without milestone:read receive a forbidden error and cannot access any milestone pages" — Step 2b tests /milestones but not /milestones/:mapId. The invariant says "any milestone pages" but only verifies the list page. (-5)
- Deduction: Missing invariant about read-only users being able to navigate to MI detail pages. Step 6 covers this but no invariant captures it as a universal rule. (-5)

### Cross-Step References (30/50)

- Steps generally do not reference each other, which is good for independence.
- Deduction: Step 6b references "Detail panel is open in read-only mode" as a precondition, which depends on Step 4 being completed first. However, the journey does not explicitly state this dependency or how to set up the panel. (-5)
- Deduction: Step 3b (read-only info displays) overlaps significantly with Step 3 (timeline view without edit controls). Both test the same view with the same permissions. The distinction is unclear — Step 3b adds verification of "hover interactions" but Step 5 also covers hover. (-5)
- Deduction: Step 5b references "read-only mode on the timeline view" without specifying how the user got there. It should reference arriving via Step 3. (-5)
- Deduction: No cross-reference between Step E1 (session expired) and the other steps that trigger API calls (Steps 1, 3). Session could expire during any of these. (-5)

### Risk Level (25/40)

- Risk level is declared as "Low" with justification "Workflow is read-only or purely observational."
- Deduction: Low risk is reasonable for read-only access, but the journey includes error handling (Step 1b), access control (Step 2b), and session management (Step E1). These are security-adjacent concerns that might warrant at least "Medium" risk. (-5)
- Deduction: The PRD security requirements (prd-spec lines 228-229) specify "milestone:read 缺失时 /milestones 页面返回 403" and "无权限操作返回 403，前端按钮禁用态." The second point (disabled buttons for unauthorized) contradicts the journey's assumption that buttons are hidden. This inconsistency with the PRD is not addressed in the risk assessment. (-5)
- Deduction: No risk consideration for the fact that read-only users can still access MI detail pages (Step 6), potentially exposing data beyond milestone scope. The risk level should at least acknowledge this cross-boundary navigation. (-5)

---

## Blindspot Attacks

1. **[blindspot] Completeness**: No step tests a cancelled milestone in read-only mode. The PRD Story 8 and Story 10 AC specify that cancelled milestones display grey styling, empty MI list, and no add button. A read-only user viewing a cancelled milestone is a distinct scenario. — Quote from journey: Step 4 Expected Result says "Detail panel opens showing all read-only information" but does not distinguish between active and cancelled milestones. — Must add a step with precondition "Milestone is in cancelled status" verifying grey styling and empty MI list.

2. **[blindspot] Precondition Exclusivity**: No step tests a user with partial permissions (e.g., milestone:read + milestone:create but not milestone:update). The journey only tests the binary case: all mutation permissions or none. — Quote from journey Setup: "User has milestone:read permission but no milestone:create, milestone:update, or milestone:delete permissions" — this is the only permission profile tested. — Must add at least one step testing a mixed permission scenario.

3. **[blindspot] Completeness**: No step tests direct URL access to /milestones/:mapId timeline view in read-only mode. Step 2b only tests /milestones (list page) without read permission. — Quote from journey: Step 2b User Action is "User attempts to access /milestones" — only the list page is tested for access denial. — Must add a step testing /milestones/:mapId access without milestone:read.

4. **[blindspot] Fact Alignment**: Step 3 Expected Result mentions "breadcrumb" but the fact annotation traces to "prd-spec — timeline view for read-only" without citing a specific AC. The breadcrumb behavior is defined in Story 9 AC, not Story 14. — Quote from journey Step 3: "Timeline renders fully (breadcrumb, title, info card, filter bar, timeline)" with fact annotation "<!-- fact: prd-spec — timeline view for read-only -->". — Must update fact annotation to cite Story 9 AC for breadcrumb and verify breadcrumb "click" returns to list works in read-only mode.

5. **[blindspot] Surface Fitness**: The note at line 144 declares validation-error "not applicable" because "no forms or mutation controls are available." However, the URL /milestones/:mapId with an invalid mapId should trigger a validation/404 error. This is a web-specific validation scenario that is dismissed without justification. — Quote from journey: "This journey covers read-only access only. No forms or mutation controls are available to the user, so validation-error outcomes are not applicable." — Must add a step for invalid URL parameter handling or justify its exclusion.

6. **[blindspot] Completeness**: No step tests the breadcrumb click in read-only mode. The PRD Story 9 AC line 193 says "Given 我点击面包屑'里程碑图'，Then 返回列表视图（路由跳转至 /milestones）." This navigation behavior is core to the read-only experience. — Quote from journey: Step 3 mentions breadcrumb is rendered but no step verifies clicking it. — Must add a step verifying breadcrumb navigation from timeline to list in read-only mode.

7. **[blindspot] Internal Consistency**: Invariant "Users without milestone:create never see create buttons" is tested by Steps 1 and 2 but the PRD Story 14 AC line 294 says the create button could be "不显示或显示为禁用态" (hidden OR disabled). The invariant and steps assume hidden only, creating a mismatch with the PRD. — Quote from journey invariant: "Users without milestone:create never see create buttons" — this asserts "never see" which excludes the "disabled" option the PRD allows. — Must either update the invariant to match PRD ambiguity or add a step that verifies the specific implementation choice.

8. **[blindspot] Precondition Exclusivity**: Step 3b and Step 3 test the same view with the same user permissions. Step 3b's Expected Result says "All read-only information displays correctly: name, status, progress, description, associated MI list. All hover interactions work" while Step 3 already verifies "Timeline renders fully (breadcrumb, title, info card, filter bar, timeline)" with edit/delete controls hidden. The overlap creates ambiguity about what each step uniquely verifies. — Quote from journey Step 3b: "Precondition: User has milestone:read only" and Step 3: "User Action: User clicks a milestone map card to enter the timeline view" — both arrive at the same view with the same permissions. — Must differentiate by giving Step 3b a distinct precondition (e.g., specific data state like "Milestone map has 0 milestones") or merge with Step 3.
