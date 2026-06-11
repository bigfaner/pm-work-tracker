# Journey Eval: milestone-map-visualization — Iteration 1

**Date**: 2026-06-08
**Evaluator**: Senior QA Adversary
**Document**: `docs/features/milestone-map/testing/milestone-map-visualization/journey.md`

---

## Reasoning Audit

The journey covers read-only visualization: list view, timeline view, filters, search, zoom, tooltips, and navigation. It is scoped to Story 9 (PM views progress) and parts of Story 14 (read-only access). The document contains 10 happy-path steps, 13 edge-case steps, and 5 journey invariants. The surface is web-only, matching the PRD scope.

Key reasoning paths verified:
1. Does the journey cover all user-facing behaviors from Story 9? Largely yes, with gaps noted below.
2. Are mandatory surface outcomes (session-expired, validation-error, server-error) present? session-expired and server-error are present; validation-error is explicitly waived via a comment — the waiver is defensible since the journey is read-only with no free-text validation.
3. Are preconditions distinct and sufficient? Some preconditions are missing or ambiguous.
4. Do facts trace correctly to source documents? Generally good, with a few misalignments.

---

## Dimension Scoring

### 1. Completeness — 200 pts (metadata 50, steps 80, outcomes 70)

**Metadata (50 pts): Score 45/50**
- feature, journey, risk_level, surface_types, surface_keys, sources, generated are all present and correct.
- Risk level "Medium" is appropriate for a read-only workflow.
- Deduction: sources list only two files; prd-ui-functions.md is heavily used (especially for delete constraint BR-4, delete visibility, permission details) but is not listed in sources. (-5)

**Steps (80 pts): Score 65/80**
- 10 happy-path steps + 13 edge-case steps = 23 total steps, providing solid coverage of the visualization workflow.
- Happy path covers: list load, hover, navigate to timeline, node hover, search, status filter, reset, zoom, breadcrumb back, click MI item. This is comprehensive for Story 9's visualization scope.
- Edge cases cover: empty state, server error, status filter, owner filter, search, refresh, loading state, empty timeline, data load failure, description tooltip, horizontal scroll, delete hidden, access denied, session expired.
- Deductions:
  - Step 10 (Click MI item) has no corresponding edge case for what happens when the MI has been deleted or the MI link is invalid. (-3)
  - No step covers keyboard accessibility (e.g., Tab navigation through cards, Enter to activate). PRD Story 9 does not mention this explicitly, but web surface convention expects basic keyboard coverage. (-2)
  - No step covers the "milestone node thumbnails" at the bottom of list cards mentioned in Story 9 acceptance criteria. The happy path Step 1 mentions them in Expected Result but no step verifies their behavior (clicking, tooltip, etc.). (-3)
  - Step 3 says "edit/delete controls" in Expected Result but neither edit nor delete is tested as a step. While this journey is scoped to visualization only, the presence of controls should at least have a step verifying their visibility/hidden state based on permissions beyond Step 3f (delete hidden). No step covers the edit button visibility based on milestone:update permission. (-4)
  - No step covers the "create button" in the filter bar mentioned in Step 3 Expected Result. (-3)

**Outcomes (70 pts): Score 58/70**
- Expected results are generally clear and specific. Many cite concrete UI elements (breadcrumb, status badge, progress, tooltips).
- Deductions:
  - Step 3 Expected Result lists many components in a single dense sentence — "breadcrumb, detail title area (name, status badge, edit/delete controls), basic info card (owner, plan dates, progress, description), filter bar (search, status filter, reset, refresh, create button, zoom controls), and horizontal timeline." This is a verification checklist crammed into one outcome. It should be separate verifiable assertions. (-4)
  - Step 8 Expected Result says "Milestone node positions remain consistent across zoom levels" — "consistent" is ambiguous. Does it mean pixel-identical x-coordinates? The PRD says "节点位置不变" and specifies transition 200ms, which the journey omits. (-3)
  - Step 3g (access denied) says "Page displays a forbidden access message" but does not specify what "forbidden access message" looks like. PRD Story 14 says "页面返回 403 提示". The journey should reference the 403 status explicitly. (-2)
  - Step 1c (server error) and Step 3d (data load failure) both say "retryable error message" / "retry control" but neither specifies the exact UI treatment. PRD says "加载失败，请重试" — the journey should include this text or at minimum reference the PRD. (-3)

**Completeness Total: 168/200**

---

### 2. Semantic Purity — 200 pts (natural language 80, declarative preconditions 60, no impl coupling 60)

**Natural Language (80 pts): Score 72/80**
- Steps use user-action language: "PM navigates to", "PM clicks", "PM hovers", "PM types". This is good.
- Deductions:
  - Step 3 Expected Result uses implementation-style enumeration: "breadcrumb, detail title area (name, status badge, edit/delete controls), basic info card..." — this reads as a component checklist rather than a user-observable outcome. (-4)
  - Step 5 Expected Result: "Only milestone nodes whose names match the keyword are shown. Non-matching nodes are hidden." This couples to a specific filter mechanism (hide/show) rather than stating the user-observable outcome ("The timeline displays only milestones whose names contain the keyword"). (-4)

**Declarative Preconditions (60 pts): Score 48/60**
- The Setup section uses bullet-list preconditions — acceptable but not fully declarative Given/When format.
- Edge cases use "Precondition" headers which is good.
- Deductions:
  - Step 1b: "The team has 0 milestone maps" — this is a database state, not declarative in terms of user actions or system state setup. (-2)
  - Step 1c: "Backend returns an error" — this is implementation-coupled. A declarative form would be "Given the server is unavailable." (-3)
  - Step 3f: "Milestone map status is not in a deletable state ('executing', 'completed', 'cancelled')" — this lists specific enum values, which is fine, but the quote from the journey incorrectly states the deletable constraint. Per BR-4, only `planning`/`reviewed`/`ready` are deletable, meaning `executing`, `completed`, and `cancelled` are NOT deletable. The precondition is correct but phrased as a negative with parenthetical exceptions — a declarative form would be clearer: "Given the milestone map status is 'executing'." (-3)
  - Step 3g: "User does not have milestone:read permission" — this is good, declarative. (-0)
  - Step E1: "The user was previously authenticated and the session has expired" — good, declarative. (-0)
  - Step 3e: "Detail title area description text overflows the display area" — this is a UI state that would be hard to set up declaratively without specifying data. (-2)
  - Step 1d and 1e preconditions ("List has milestone maps with various statuses" / "List has milestone maps with different owners") are vague — "various" and "different" are not precise. (-2)

**No Impl Coupling (60 pts): Score 52/60**
- Most steps describe user-observable behavior without coupling to implementation.
- Deductions:
  - Step 1 Expected Result: "A loading placeholder appears during data fetch" — "data fetch" is implementation language. User-observable: "A loading indicator (skeleton) appears while the page loads." (-3)
  - Step 3b Expected Result: "A loading placeholder is shown during data fetch. The placeholder is replaced with actual content once data arrives." — "data fetch" and "data arrives" are implementation terms. (-3)
  - Invariant 3: "All filter operations (search, status, owner) are client-side or debounced to avoid excessive API calls" — "client-side or debounced" is implementation detail. The user-observable behavior is that filter results appear without perceptible delay. (-2)

**Semantic Purity Total: 172/200**

---

### 3. Precondition Exclusivity — 150 pts (distinct 60, sufficient 50, missing for errors 40)

**Distinct (60 pts): Score 50/60**
- Setup preconditions are global and shared across all steps. Edge-case preconditions are mostly distinct.
- Deductions:
  - Steps 1d, 1e, 1f (list filter edge cases) all require "List has milestone maps with various/different/multiple" — these preconditions overlap significantly and could be a single "List has milestone maps with varied statuses, owners, and names" setup. (-3)
  - Steps 1b (empty) and 1c (server error) share no preconditions with the happy path but Step 1b's "0 milestone maps" is mutually exclusive with Setup's "at least 3 milestone maps" — this is correct exclusivity. (-0)
  - Steps 3b (loading), 3c (empty), 3d (data load failure) all have different preconditions — good. (-0)
  - Steps 5 and 6 (search and status filter) in the happy path have no explicit preconditions beyond Setup, but both assume specific data exists (milestones with varied names and statuses). The Setup section says "At least one milestone map has milestones with associated MainItems" but doesn't guarantee name/status diversity. (-7)

**Sufficient (50 pts): Score 38/50**
- Happy-path steps largely rely on the global Setup, which is minimal but mostly sufficient.
- Deductions:
  - Step 8 (zoom) has no precondition about the timeline having sufficient date range to demonstrate zoom differences. If all milestones share the same date, zoom would be meaningless. (-5)
  - Step 8b (horizontal scroll) says "Milestone nodes are densely packed" — no precondition sets up this density (how many nodes, what date range). (-3)
  - Step 10 (click MI item) has no precondition that a milestone has associated MainItems visible in the timeline. Setup says "At least one milestone map has milestones with associated MainItems" but the step does not reference this or specify which map. (-4)

**Missing for Errors (40 pts): Score 30/40**
- Error preconditions are present for: server error (Step 1c), data load failure (Step 3d), access denied (Step 3g), session expired (Step E1).
- Deductions:
  - No error case for search returning no results. Step 5 tests "types a keyword" and expects matching nodes shown, but no step tests what happens when no nodes match. (-5)
  - No error case for status filter returning empty results. (-3)
  - Step 1c precondition "Backend returns an error" does not specify what kind of error (500, timeout, network error). Different errors may produce different UI states. (-2)

**Precondition Exclusivity Total: 118/150**

---

### 4. Fact Alignment — 150 pts (traceable 60, inferred with rules 50, no hallucinated 40)

**Traceable (60 pts): Score 54/60**
- Most steps have `<!-- fact: prd-spec Story 9 -->` annotations tracing to the source.
- Deductions:
  - Step 3 Expected Result mentions "edit/delete controls" but Story 9 acceptance criteria say "详情标题区（名称+可点击状态 Badge+编辑/删除按钮）" — the edit/delete buttons are part of Story 2 and Story 3 (edit/delete milestone map), not Story 9. The journey for visualization includes write-operation controls that belong to different stories. (-3)
  - Step 3f claims delete is hidden for "executing", "completed", "cancelled" states. The fact annotation says `<!-- fact: prd-spec — only planning, reviewed, ready are deletable -->`. However, per prd-spec BR-4, the deletable states are `planning`/`reviewed`/`ready`. The journey Step 3f lists the non-deletable states but the PRD Story 3 says: "仅 planning 状态可删除，BR-4". Wait — checking more carefully: Story 3 says "仅 planning 状态可删除（BR-4）" but BR-4 in prd-ui-functions says "仅 `planning`/`reviewed`/`ready` 状态的里程碑图可删除". There is an inconsistency between Story 3 (only planning) and BR-4 (planning/reviewed/ready). The journey follows BR-4. This is a source inconsistency, not a journey error. However, the journey's precondition lists 'executing', 'completed', 'cancelled' as non-deletable, which misses that `reviewed` and `ready` ARE deletable per BR-4. The step says "status is not in a deletable state" with a parenthetical, so the examples are correct. No deduction for this.
  - Step 3 mentions "create button" in the filter bar. Story 9 says "创建里程碑按钮" in the filter bar. This is traceable. (-0)
  - Step 2 (hover card) has no fact annotation. It corresponds to Story 9 "Given 我悬停某张卡片，Then 卡片边框高亮，出现阴影效果". The fact tag is missing. (-3)

**Inferred with Rules (50 pts): Score 42/50**
- Steps 1c, 3d, E1 are marked `<!-- source: inferred — derived from Web surface ... -->`. These inferences are reasonable.
- Deductions:
  - Step 3b (Timeline loading state) has no source annotation. It is inferred from the PRD ("显示骨架屏") but is not tagged. (-3)
  - Step 3c (Timeline empty state) has no source annotation. It is inferred from Story 9 ("显示'暂无里程碑'空状态和创建按钮") but is not tagged. (-3)
  - Step 3e (description tooltip) has no source annotation. It is inferred from Story 9 ("描述最多三行+Tooltip") but is not tagged. (-2)

**No Hallucinated (40 pts): Score 34/40**
- No steps appear entirely fabricated.
- Deductions:
  - Step 8 Expected Result says "Milestone node positions remain consistent across zoom levels" — the PRD says "里程碑节点位置不变" which is the same. However, the journey omits the "transition 200ms" detail from the PRD. This is an omission, not a hallucination. (-0)
  - Step 1g (refresh list) shows "refresh button shows loading state during fetch" — the PRD says "刷新按钮显示 loading 状态". Aligned. (-0)
  - Step 3g (access denied) says "Page displays a forbidden access message." The PRD says "页面返回 403 提示". The journey uses vaguer language but does not contradict. (-0)
  - Step 1b (empty state) says "No milestone maps" message — the PRD says "暂无里程碑图". The journey does not specify the exact text, which is fine for a journey (not a contract), but the English text "No milestone maps" is a paraphrase, not a hallucination. (-0)
  - Step 1b Expected Result: "and a create button (if PM has milestone:create permission)." This conditional is not in Story 9's empty state acceptance criteria. Story 9 says "显示空状态提示'暂无里程碑图'及创建按钮" unconditionally. Story 14 says for users with milestone:read but no milestone:create, the create button should not show. The journey adds a permission check that is only in Story 14. This is correct cross-referencing, not hallucination, but the fact tag is missing. (-3)
  - Step 3f precondition: "Milestone map status is not in a deletable state ('executing', 'completed', 'cancelled')" — checking against BR-4 more carefully, the step says status is NOT deletable and lists executing/completed/cancelled. But `reviewed` and `ready` ARE deletable per BR-4, and the step does not test those. This is a coverage gap, not a hallucination. However, the step's fact tag says `<!-- fact: prd-spec — only planning, reviewed, ready are deletable -->`. This fact tag actually contradicts the step's precondition — the tag says planning/reviewed/ready ARE deletable, but the step tests a non-deletable state. The tag is accurate; the step's choice of non-deletable examples is valid. (-3)

**Fact Alignment Total: 130/150**

---

### 5. Surface Fitness — 150 pts (mandatory outcomes 60, strategy proportions 50, realistic assumptions 40)

**Mandatory Outcomes (60 pts): Score 48/60**
- Surface is web. Mandatory outcomes for web: validation-error, session-expired, server-error.
  - session-expired: Present (Step E1). The step covers the mandatory outcome. (+full)
  - server-error: Present (Steps 1c, 3d). (+full)
  - validation-error: Explicitly waived with a comment: "This journey covers read-only visualization... The validation-error mandatory outcome is not applicable to this journey's scope." The waiver is defensible — the only user inputs are search keywords and predefined filter selections. (-0)
- Deductions:
  - Step E1 (session expired) does not specify the re-authentication flow in detail. It says "After re-authenticating, the user must re-navigate to the desired page" — this is adequate but does not test whether unsaved state is lost (not applicable for read-only). (-0)
  - Step 1c (server error) says "retryable error message with a retry control" — this covers the server-error outcome. (-0)
  - Step 3d (data load failure) similarly covers server-error for the timeline view. (-0)
  - Step 3g (access denied) covers the unauthorized/forbidden case. While not listed as "mandatory" in the rubric for web, it is important for security. The step is present. (-0)
  - Deduction: The journey has no step for network timeout specifically. Steps 1c and 3d cover "server error" but do not distinguish timeout from 500 error. PRD Story 14 mentions "后端 API 超时或返回 500" as a single case, so a single step is acceptable. (-0)
  - Deduction: The session-expired step (E1) has a precondition "The user was previously authenticated and the session has expired" but does not specify HOW the session expires (idle timeout, token expiry, server restart). For a journey, the mechanism does not matter — only the state matters. (-0)
  - Deduction: The mandatory outcome for "unauthorized" (403) is present in Step 3g, but the rubric says "backend=api (unauthorized mandatory)" — this journey is web-only, not api. The web mandatory outcomes are validation-error, session-expired, and server-error. Step 3g covers 403 which goes beyond mandatory. (-0)
  - Actual deduction: Step 3g covers access-denied for the list page, but there is no step for access-denied on the timeline view (/milestones/:mapId). A user without milestone:read could directly navigate to a timeline URL. (-6)
  - Step 1c says "No blank page is shown" — this is a good negative assertion. (-0)
  - The invariants section correctly states "Error states always provide a retry option, never a blank page." (-0)

**Strategy Proportions (50 pts): Score 42/50**
- The journey has 23 steps: 10 happy path, 13 edge cases. This is a good ratio.
- The happy path covers the primary workflow (list -> timeline -> filter -> zoom -> navigate back).
- Edge cases cover loading, error, empty, permission, and interaction variations.
- Deductions:
  - The ratio of happy-path to edge-case steps is approximately 43%/57%, which is appropriate for a medium-risk read-only journey. (-0)
  - However, the edge cases are heavily concentrated on the list view (Steps 1b-1g = 6 edge cases for list, Steps 3b-3g = 6 edge cases for timeline, Step 8b = 1 for zoom, Step E1 = 1 for session). The timeline view edge cases could benefit from a few more: e.g., timeline with milestones in different statuses, timeline with cancelled milestones. (-5)
  - No edge case for concurrent data changes (e.g., another user deletes a milestone while PM is viewing timeline). (-3)

**Realistic Assumptions (40 pts): Score 34/40**
- Setup assumptions are realistic: 3+ milestone maps, various statuses/owners, at least one with milestones and items.
- Deductions:
  - Setup says "At least 3 milestone maps exist with various statuses and owners" — "various statuses" should specify at least one non-planning status to make status filter meaningful. (-3)
  - Setup says "At least one milestone map has milestones with associated MainItems" — but Step 8 (zoom) and Step 8b (scroll) require specific data distributions not guaranteed by this setup. (-3)

**Surface Fitness Total: 124/150**

---

### 6. Internal Consistency — 150 pts (invariants 60, cross-Step refs 50, risk level 40)

**Invariants (60 pts): Score 48/60**
- 5 journey invariants are listed at the bottom.
- Deductions:
  - Invariant 1: "List view always shows a loading placeholder during data fetch, never a blank page." Step 1 (happy path) does not explicitly verify the loading placeholder — it jumps straight to "Page loads showing all milestone map cards." The loading state is only tested in Step 3b for the timeline. The invariant is stated but not verified by a step for the list view. (-5)
  - Invariant 2: "Timeline node positions are determined by plan completion date and do not change when zoom scale changes." Step 8 verifies positions "remain consistent" but does not specify "plan completion date" as the determinant. (-2)
  - Invariant 3: "All filter operations (search, status, owner) are client-side or debounced to avoid excessive API calls." This is an implementation invariant, not a user-observable one. It also cannot be verified through a journey test. (-3)
  - Invariant 4: "Breadcrumb navigation provides a consistent way to return from timeline to list view." Step 9 verifies breadcrumb, but no step tests breadcrumb from a filtered state or after interaction. (-2)

**Cross-Step References (50 pts): Score 40/50**
- Steps reference each other implicitly (Step 3 navigates from Step 1, Step 9 returns to Step 1 context).
- Deductions:
  - Step 1d (filter list by status), 1e (filter by owner), 1f (search list by name) all modify list state but none reference what happens when transitioning to Step 3 from a filtered list. Does clicking a card while filters are active navigate correctly? (-4)
  - Step 7 (reset filters) resets to "default" but does not reference what "default" means. Step 5 and 6 set filters; Step 7 resets. No step verifies that reset restores the state from before Steps 5/6 or from the initial load. (-3)
  - Step 10 (click MI item) navigates away from the timeline but no step covers returning to the timeline after navigation. Step 9 covers breadcrumb from timeline to list, but not from MI detail back to timeline. (-3)

**Risk Level (40 pts): Score 34/40**
- Risk level "Medium" is justified: "Workflow involves multi-step interaction without irreversible side effects."
- The journey is read-only, which supports Medium (not High).
- Deductions:
  - The risk justification does not mention permission-based access control (Step 3g), which adds security risk. However, Medium is still appropriate. (-0)
  - Step 3f (delete hidden) and Step 3g (access denied) verify security constraints, which is appropriate for Medium risk. (-0)
  - The risk classification does not account for the data volume assumptions (200+ MI per PRD performance requirements). A timeline with 200 MI items could have performance implications not covered by any step. (-4)
  - No step covers the "skeleton screen" loading state for the list view (only the timeline loading is in Step 3b). Step 1's Expected Result mentions "loading placeholder" but no step verifies it as a separate precondition+action+result. (-2)

**Internal Consistency Total: 122/150**

---

## Blindspot Hunt

1. **[blindspot] No search-with-no-results edge case.** Step 5 tests search with matching results, but no step tests entering a keyword that matches no milestones. Expected behavior (empty timeline? message?) is undefined. — Quote: Step 5 Expected Result: "Only milestone nodes whose names match the keyword are shown. Non-matching nodes are hidden." — What must improve: Add Step 5b with precondition that no milestone names contain the search keyword, verifying appropriate empty/no-results feedback.

2. **[blindspot] No status-filter-with-no-results edge case.** Same as above for Step 6. Filtering by a status that no milestone has should show an empty state or message. — Quote: Step 6 Expected Result: "Only milestone nodes matching the selected statuses are shown." — What must improve: Add Step 6b covering the no-match scenario.

3. **[blindspot] Access denied for timeline view URL is missing.** Step 3g covers /milestones list page access denied, but a user without milestone:read could directly navigate to /milestones/:mapId. — Quote: Step 3g Precondition: "User does not have milestone:read permission." Step 3g User Action: "User attempts to navigate to /milestones." — What must improve: Add a step for direct navigation to a timeline URL without permission.

4. **[blindspot] Concurrent data modification not covered.** If another user deletes a milestone map while the PM is viewing the timeline, or adds/removes milestones, the current PM's view could become stale. No step covers refresh-after-stale-data scenarios beyond manual refresh. — Quote: Step 1g covers manual refresh but no step covers stale data detection or automatic refresh. — What must improve: Add a step or edge case for viewing a timeline when the underlying milestone map has been deleted by another user.

5. **[blindspot] Step 3f precondition vs BR-4 alignment gap.** Step 3f tests delete hidden for executing/completed/cancelled states but does not verify delete IS visible for planning/reviewed/ready states. Only the negative case is tested, not the positive. — Quote: Step 3f Precondition: "Milestone map status is not in a deletable state ('executing', 'completed', 'cancelled')." — What must improve: Add a complementary step verifying delete button IS displayed for planning (or reviewed/ready) status.

6. **[blindspot] Edit button visibility based on permission not covered.** Step 3 mentions "edit/delete controls" in Expected Result, Step 3f covers delete hidden for non-deletable states, but no step covers the edit button being hidden when the user lacks milestone:update permission. PRD Story 14 says users with only milestone:read see no edit/delete buttons. — Quote: Step 3 Expected Result includes "edit/delete controls" without any permission precondition. — What must improve: Add Step 3h covering the read-only user scenario on the timeline view (edit/delete hidden, create hidden).

7. **[blindspot] Cancelled milestone map timeline view not covered.** No step tests what happens when navigating to a milestone map in cancelled status. PRD BR-5 says terminal-state maps hide create/edit/delete controls, and BR-6 says cancellation cascades. — Quote: No step or edge case references a cancelled milestone map. — What must improve: Add Step 3i for viewing a cancelled milestone map timeline (create/edit/delete hidden, milestones in cancelled state).

8. **[blindspot] Breadcrumb from filtered timeline state not tested.** Step 7 resets filters, Step 9 navigates back via breadcrumb, but no step tests breadcrumb navigation when filters are active (do filters persist when navigating back?). — Quote: Step 9 Expected Result: "Route navigates back to the list view." — No mention of filter state after navigation. — What must improve: Clarify whether breadcrumb navigation resets filter state or preserves it.

9. **[blindspot] Step 10 MI item navigation has no return path verification.** Step 10 navigates to /items/:mainItemId but no step verifies returning from MI detail to the milestone timeline. PRD mentions navigation rules for this return path. — Quote: Step 10 Expected Result: "Route navigates to the main item detail page." — What must improve: Either add a return step or document that return navigation is out of scope for this journey.

10. **[blindspot] Invariant 3 is untestable.** "All filter operations are client-side or debounced" cannot be verified through a journey test — it is an implementation constraint. — Quote: Invariant 3: "All filter operations (search, status, owner) are client-side or debounced to avoid excessive API calls." — What must improve: Replace with a testable invariant such as "Filter results update within 500ms of user input."

---

## Score Summary

| Dimension | Score | Min Threshold | Pass? |
|-----------|-------|---------------|-------|
| 1. Completeness | 168/200 | 120 | Yes |
| 2. Semantic Purity | 172/200 | 120 | Yes |
| 3. Precondition Exclusivity | 118/150 | 90 | Yes |
| 4. Fact Alignment | 130/150 | 90 | Yes |
| 5. Surface Fitness | 124/150 | 90 | Yes |
| 6. Internal Consistency | 122/150 | 90 | Yes |
| **Total** | **834/1000** | **600** | **Yes** |

---

## Top Attacks (Priority Order)

1. **[Completeness]: Missing search/status-filter no-results edge cases — Steps 5 and 6 test positive filter results but no step verifies behavior when no milestones match. Add Step 5b and 6b with preconditions that guarantee no matches, verifying appropriate empty-state feedback.**

2. **[Surface Fitness]: Access denied for timeline URL not covered — Step 3g only covers /milestones list page. A user without milestone:read could directly navigate to /milestones/:mapId. Add Step 3h verifying 403 for direct timeline URL access.**

3. **[Internal Consistency]: Invariant 1 (loading placeholder) not verified for list view — Step 1 jumps to loaded state without verifying the loading skeleton. Add a step or modify Step 1 to verify loading state before data arrives.**

4. **[Completeness]: Edit button permission visibility not tested — Step 3 Expected Result includes "edit/delete controls" but no step verifies edit button is hidden for users with only milestone:read. Add Step 3h for read-only user on timeline view.**

5. **[Completeness]: Cancelled milestone map timeline view not covered — No step tests navigating to a milestone map in cancelled status. Per BR-5, terminal-state maps should hide create/edit/delete controls. Add Step 3i.**

6. **[Precondition Exclusivity]: Happy path Steps 5/6/8 lack specific data preconditions — Steps assume milestone name/status diversity but Setup only guarantees "at least one milestone map has milestones with associated MainItems." Add explicit preconditions for data diversity.**

7. **[Internal Consistency]: Invariant 3 is untestable implementation detail — "client-side or debounced" cannot be verified via journey test. Replace with user-observable timing assertion.**

8. **[Fact Alignment]: Steps 3b, 3c, 3e missing source annotations — These steps are inferred from PRD Story 9 but lack `<!-- source: inferred -->` tags. Add annotations for traceability.**

9. **[Semantic Purity]: Step 3 Expected Result is an overstuffed component checklist — Lists 7+ UI components in a single sentence, making it difficult to verify individually. Split into separate verifiable assertions or add sub-steps.**

10. **[Surface Fitness]: No concurrent modification edge case — If another user deletes a milestone map while PM views the timeline, behavior is undefined. Add Step 3j for viewing a deleted milestone map.**
