# Journey Evaluation: item-list-milestone-integration

**Iteration**: 1
**Date**: 2026-06-08
**Evaluator**: Senior QA Engineer (adversary mode)

---

## Reasoning Audit

### Surface Classification

The journey declares `surface_types: ["web"]` with `surface_keys: ["frontend"]`. This is correct: the item-list page and table view are purely frontend interactions. The PRD (prd-spec Story 12 and Story 13) describes these as frontend-only features. No backend API surface is independently exercised; all API calls are implicit side-effects of frontend actions.

### Source Traceability

Sources listed: `prd-user-stories.md` (Stories 12, 13) and `prd-spec.md`. Every Step's `<!-- fact: ... -->` annotation correctly references either Story 12, Story 13, or a specific PRD clause. Verified against the actual document content.

---

## Rubric Scoring

### 1. Completeness — 200 pts (metadata 50, steps 80, outcomes 70)

**Metadata (50 pts): 45/50**

- feature, journey, risk_level, surface_types, surface_keys, sources, generated date all present. Risk level "Low" justified with inline comment ("Workflow is read-only or purely observational"). This is correct: the journey covers only filtering, viewing, and sorting — no mutation.
- Deduction: The `sources` field references `prd-spec.md` generically but the fact annotations reference "prd-spec Story 12, Story 13" — Stories are in `prd-user-stories.md`, not `prd-spec.md`. The source list is slightly misleading. (-5)

**Steps (80 pts): 65/80**

- Happy path covers 8 steps (Steps 1-8): view filter, filter by milestone, filter by unassigned, filter by all, view badge, view column, sort asc, sort desc. This maps well to the core flow from the PRD.
- Edge cases cover 8 additional scenarios (Steps 2b, 5b, 6b, 6c, 6d, 6e, 7b, 7c, E1). Good coverage of error and boundary conditions.
- Deduction: Step numbering is inconsistent — edge case steps use suffixes (2b, 5b, 6b-6e, 7b-7c) but the suffix logic is opaque. Step 5b ("Filter with invalid bizKey") is about the list filter but is numbered after Step 5 (view badge), creating a confusing reading order. (-3)
- Deduction: Missing step for the PRD Story 12 acceptance criterion: "MI milestone_key points to a cancelled milestone — what happens in the list view?" The journey covers cancelled milestones not appearing in the dropdown (Step 2b) but never tests what happens when an already-assigned MI's milestone is later cancelled. The list badge should still show or not? Not tested. (-5)
- Deduction: No step covers "团队下没有创建任何里程碑" (team has zero milestones) — Story 12 specifies "下拉框仅显示'未分配'选项" for this case. This is a distinct precondition not covered. (-4)
- Deduction: No step tests pagination / large data set behavior. While this may be a web-level concern, the PRD mentions "超出 200 个 MI 时启用分页加载" — no step validates filter behavior with paginated results. (-3)

**Outcomes (70 pts): 55/70**

- Most expected results are clear and testable.
- Deduction: Step 6d says "Milestone column defaults to ascending order" — but the PRD Story 13 says "默认升序" applies only to the milestone column's own default, not necessarily the table's overall default sort. The journey implies the entire table is sorted by milestone on first load, which may conflict with existing default sort. The outcome is ambiguous. (-3)
- Deduction: Step 5b ("Filter with invalid bizKey — Filter falls back to 'All' without producing an error") — "without producing an error" is vague. No error toast? No console error? No network request? The expected behavior should specify UX feedback explicitly. (-3)
- Deduction: Step 6c outcome says "Milestone column shows a fallback value for all rows" — but does not specify what the fallback value is. PRD Story 13 says show "--". The journey should match. (-3)
- Deduction: Step 6b outcome says "placeholder indicating the milestone is no longer available" — PRD Story 13 says display "--". The journey uses vaguer language. (-3)
- Deduction: No negative outcome validation. Steps only describe what should happen on success. For example, if the user lacks `milestone:read` permission, what happens? Not covered in this journey (only the happy-path permission `milestone:read` is in setup). (-3)

**Completeness Total: 45 + 65 + 55 = 165/200**

---

### 2. Semantic Purity — 200 pts (natural language 80, declarative preconditions 60, no impl coupling 60)

**Natural Language (80 pts): 70/80**

- Steps are written in plain language: "PM loads the items list page", "PM opens the milestone dropdown and selects a specific milestone." Good.
- Deduction: Step 5b says "An invalid bizKey is passed to the milestone filter" — this is implementation-coupled language. A user-facing description would be "PM manipulates the URL or filter parameter to contain an invalid value." The word "bizKey" is an internal concept. (-5)
- Deduction: Step 6b references "soft-deleted milestone" which is an implementation concept. From the user's perspective, the milestone "has been removed" or "no longer exists." (-3)
- Deduction: Step 6c and 7c both reference "API call fails" — this is implementation-coupled. A user-facing description would be "the server is unavailable or returns an error." (-2)

**Declarative Preconditions (60 pts): 50/60**

- Setup section uses declarative style: "Multiple MainItems exist", "User has milestone:read permission."
- Edge case preconditions are generally declarative.
- Deduction: Step 5b precondition "An invalid bizKey is passed to the milestone filter" is imperative/ambiguous — who passes it? How? The action "Filter is triggered with an invalid value" is also vague about the mechanism. (-5)
- Deduction: Step 6c precondition "Milestone data API call fails" — not declarative from a state perspective. Should be "The server returns an error when loading milestone data." (-3)
- Deduction: Step 7c precondition "Milestone options API call fails" — same issue. (-2)

**No Impl Coupling (60 pts): 45/60**

- Most steps avoid coupling to specific UI components or API endpoints.
- Deduction: Step 5b references "bizKey" directly — internal identifier leaking into journey. (-5)
- Deduction: Step 6b references "soft-deleted" — database-level concept. (-3)
- Deduction: Steps 6c and 7c reference "API call" — backend implementation detail. (-3)
- Deduction: Step 6d outcome "Milestone column defaults to ascending order" — implies a specific column-level default rather than describing the user-observed behavior. Not a severe coupling but slightly prescriptive. (-2)
- Deduction: The journey references "milestone_key" in edge case preconditions (Step 6b: "milestone assignment points to a milestone that has been soft-deleted") — this is the DB column name, not a user-visible concept. (-2)

**Semantic Purity Total: 70 + 50 + 45 = 165/200**

---

### 3. Precondition Exclusivity — 150 pts (distinct 60, sufficient 50, missing for errors 40)

**Distinct Preconditions (60 pts): 50/60**

- Setup establishes a valid baseline. Edge case preconditions are generally distinct from each other and from the happy path.
- Deduction: Steps 6b, 6c, 6d, 6e all share the same base state ("table view is loaded with milestone column visible") but don't explicitly state this as a precondition — they rely on implicit context from the happy path. This creates ambiguity about whether they are independent or sequential. (-5)
- Deduction: Step E1 (session expired) shares its precondition with the happy path setup but doesn't explicitly state what prior state must exist (e.g., "PM has already loaded the items list and the session expires mid-interaction"). (-5)

**Sufficient Preconditions (50 pts): 38/50**

- Happy path setup is sufficient for the core flow.
- Deduction: Step 6e ("Column header filter by milestone") has precondition "Table view is loaded with milestone column visible" — but doesn't state that the column header has a filter mechanism. This is an assumption about UI capability not stated in the preconditions. (-4)
- Deduction: Step 7b ("Switch team resets milestone filter") — no precondition states that the user belongs to multiple teams or that team switching is possible. (-3)
- Deduction: No precondition states the number or variety of milestones needed. "At least one milestone exists in the team in a non-cancelled state" is minimal — but for testing sorting (Steps 7, 8), multiple milestones with different names are needed to verify alphabetical order. (-3)
- Deduction: Step 2b requires "Team has milestones in cancelled status" but the setup says "At least one milestone exists in a non-cancelled state." These are compatible but the journey doesn't state both conditions can coexist in the same test data set. (-2)

**Missing for Error Paths (40 pts): 30/40**

- Step 5b (invalid bizKey) — precondition is stated but vague.
- Steps 6c, 7c (API failures) — preconditions use implementation language.
- Step E1 (session expired) — precondition stated.
- Deduction: No precondition for what happens when the user lacks `milestone:read` permission. The setup only covers the positive case. (-5)
- Deduction: No precondition for the "zero milestones" state mentioned in PRD Story 12 (dropdown shows only "Unassigned"). (-3)
- Deduction: Step 6b (soft-deleted milestone) precondition doesn't explain how this state is reached — was the milestone deleted while the MI was already assigned? The lifecycle context is missing. (-2)

**Precondition Exclusivity Total: 50 + 38 + 30 = 118/150**

---

### 4. Fact Alignment — 150 pts (traceable 60, inferred with rules 50, no hallucinated 40)

**Traceable Facts (60 pts): 52/60**

- Step 1 references "prd-spec Story 12 — milestone filter on items list" — verified: Story 12 first acceptance criterion confirms "筛选栏在现有'负责人'筛选器右侧显示'里程碑'下拉筛选器".
- Step 3 references "prd-spec — filter by unassigned MIs" — verified: Story 12 says "列表仅显示 milestone_key 为空的 MI".
- Step 6 references "prd-spec Story 13 — milestone column in table" — verified: Story 13 first acceptance criterion confirms the column.
- Step 2b references "prd-spec — cancelled milestones excluded" — verified: Story 12 says "cancelled 状态的里程碑不在选项中出现".
- Step 6b references "prd-spec — soft-deleted milestone data" — verified: Story 13 says "MI 的 milestone_key 指向已被软删除的里程碑...显示'--'".
- Deduction: Step 6d says "Milestone column defaults to ascending order" — but the PRD Story 13 says "页面首次加载...里程碑列默认升序". The journey implies this is an edge case, but the PRD frames it as the default behavior. The fact alignment is slightly off in framing. (-3)
- Deduction: Step 6e ("Column header filter by milestone") has no fact annotation. Story 13 does say "我使用列头筛选选择某里程碑，Then 仅显示对应里程碑的 MI" — so the fact exists but is not cited. (-3)
- Deduction: Steps 7 and 8 (sort ascending/descending) have no fact annotations, though Story 13 clearly specifies both sort directions. (-2)

**Inferred Facts with Rules (50 pts): 42/50**

- Steps 6c and 7c use "inferred — derived from Web surface server-error boundary outcome" — this is a valid inference rule. When a web surface makes server calls, server errors are a standard boundary condition.
- Step E1 uses "inferred — derived from Web surface session-expired mandatory outcome" — valid inference.
- Deduction: Step 6d ("Default sort on first load") has no inference annotation but is not explicitly stated in the PRD as an edge case — it is the PRD's default behavior. This should either be a fact citation or an inference. (-3)
- Deduction: Step 7b ("Switch team resets milestone filter") is inferred but has no inference rule annotation. Story 12 does say "切换团队...里程碑筛选器重置为'全部'" — so this is actually a direct fact, not an inference. The annotation is missing. (-3)
- Deduction: Step 5b (invalid bizKey) — inferred from Story 12's "筛选值传入非法 bizKey...回退到'全部'" but has no fact or inference annotation. (-2)

**No Hallucinated Facts (40 pts): 35/40**

- No facts appear to be fabricated. All scenarios correspond to real PRD requirements.
- Deduction: Step 6d says "Milestone column defaults to ascending order. Unassigned MIs appear at the bottom." The PRD Story 13 confirms the ascending default and unassigned at bottom, but says "未分配显示'-'" — the journey doesn't mention the display value for unassigned in the sort context. Minor inconsistency. (-3)
- Deduction: Step 7c outcome "Dropdown shows a load failure state and is disabled" — the PRD Story 12 says "显示'加载失败'且下拉框禁用". The journey says "load failure state" which is vaguer but not hallucinated. (-2)

**Fact Alignment Total: 52 + 42 + 35 = 129/150**

---

### 5. Surface Fitness — 150 pts (mandatory outcomes 60, strategy proportions 50, realistic assumptions 40)

**Mandatory Outcomes (60 pts): 50/60**

- Web surface mandatory outcomes: `validation-error` and `session-expired`.
- Step E1 covers `session-expired` — present and correct.
- `validation-error` is not explicitly covered as a named scenario. The closest is Step 5b (invalid bizKey) but the outcome says "falls back to 'All' without producing an error" — so there is no validation error shown to the user.
- Deduction: No step explicitly covers a `validation-error` scenario for the web surface. The PRD Story 12 does not describe frontend validation errors for the filter (it describes backend fallback). However, the rubric mandates a validation-error outcome for web surfaces. The journey should have at least one step where invalid input produces a visible validation error. (-7)
- Deduction: No `unauthorized` outcome is covered. The journey has `milestone:read` in setup but never tests what happens without it. While the surface is web (not API), a web surface can still encounter 403 responses. (-3)

**Strategy Proportions (50 pts): 40/50**

- 16 steps total: 8 happy path + 8 edge cases. Roughly 50/50 split.
- For a "Low" risk journey focused on read-only operations, the proportion is reasonable.
- Deduction: All edge cases are in a flat "Edge Cases" section. There is no distinction between boundary conditions, error paths, and negative tests. A clearer categorization would improve strategy coverage. (-5)
- Deduction: No explicit negative test for the web surface (e.g., attempting to filter when data is empty, or when the user has no permissions). (-5)

**Realistic Assumptions (40 pts): 35/40**

- The journey assumes the items list page and table view exist and function independently of the milestone feature. This is realistic.
- Assumptions about team switching, session expiration, and API failures are all realistic web application scenarios.
- Deduction: The journey assumes "Milestone column defaults to ascending order" on first load (Step 6d), but this may conflict with the table view's existing default sort (by item title or creation date). The journey doesn't acknowledge this potential conflict. (-3)
- Deduction: The journey doesn't state an assumption about whether the milestone filter and the table view's milestone column share state (e.g., if you filter by "Milestone A" in the list, does it persist when you switch to table view?). (-2)

**Surface Fitness Total: 50 + 40 + 35 = 125/150**

---

### 6. Internal Consistency — 150 pts (invariants 60, cross-Step refs 50, risk level 40)

**Invariants (60 pts): 52/60**

- Five invariants declared at the bottom. They are clear and testable.
- Invariant 1: "Cancelled milestones never appear in dropdown options or filter selections." — consistent with Steps 2b and throughout.
- Invariant 2: "Unassigned MIs consistently display without a badge or with a placeholder, and always sort to the bottom." — consistent with Steps 5, 7, 8, 6d.
- Invariant 3: "Milestone-related failures never block rendering." — consistent with Steps 6c, 7c.
- Invariant 4: "Switching teams always resets milestone filter to All." — consistent with Step 7b.
- Invariant 5: "Invalid filter values gracefully fall back to All." — consistent with Step 5b.
- Deduction: Invariant 2 says "without a badge or with a placeholder" — but the expected results in Steps 5, 7, 8 only mention "no badge" for unassigned. The "placeholder" part is only for the table column (Step 6). The invariant conflates two different display contexts (list vs table). (-4)
- Deduction: No invariant covers the behavior when milestone data is stale or outdated (e.g., milestone name changes but the list still shows old name). Not a critical gap but a consistency concern. (-4)

**Cross-Step References (50 pts): 40/50**

- Steps are generally self-contained, which is good for independent execution.
- Deduction: Step 6d ("Default sort on first load") and Step 7 (sort ascending) have overlapping expected results. Step 6d says "Milestone column defaults to ascending order" and Step 7 says "Assigned MIs are sorted by milestone name alphabetically ascending." These should reference each other to avoid redundancy. (-4)
- Deduction: Step 5b (invalid bizKey on list filter) and Step 6b (invalid filter on table column) test the same concept in different views but don't cross-reference each other. A test executor might not realize they're testing the same business rule. (-3)
- Deduction: Steps 6b and 6c both deal with milestone column data issues but don't establish whether they're mutually exclusive states or can co-occur. (-3)

**Risk Level (40 pts): 35/40**

- Risk level "Low" is justified: the journey covers read-only filtering, viewing, and sorting operations. No mutations, no state transitions, no data creation.
- The inline comment explains the classification criteria.
- Deduction: While the operations are read-only, Steps 5b (invalid bizKey) and E1 (session expired) introduce conditions that could cause unexpected behavior. The risk classification is still appropriate but the journey does touch on error handling, which slightly exceeds "purely observational." (-3)
- Deduction: The step count (16 steps including edge cases) is substantial for a "Low" risk journey. This suggests the coverage is thorough for a low-risk area, which is fine but could indicate the risk level might be "Medium" if error paths are significant. (-2)

**Internal Consistency Total: 52 + 40 + 35 = 127/150**

---

## Blindspot Hunt

1. **[blindspot] No test for zero-milestone team state.** PRD Story 12 explicitly says "团队下没有创建任何里程碑...下拉框仅显示'未分配'选项." The journey never tests this precondition. The setup says "At least one milestone exists" but no edge case inverts it. Quote from journey: "At least one milestone exists in the team in a non-cancelled state." What must improve: Add a step for "Team has zero milestones — dropdown shows only 'All' and 'Unassigned'."

2. **[blindspot] No test for MI assigned to a cancelled milestone in the list/table view.** Step 2b covers that cancelled milestones don't appear in the dropdown, but never tests what badge/column value appears for an MI that was previously assigned to a milestone that was later cancelled. PRD Story 8 says "关联 MI 自动解绑" on cancel, so the MI should become unassigned — but this cross-feature behavior is never verified. Quote from journey: "Cancelled milestones never appear in dropdown options or filter selections." This invariant doesn't address display of already-bound MIs. What must improve: Add a step for "MI was assigned to a milestone, that milestone is cancelled — MI badge/column shows unassigned."

3. **[blindspot] No validation-error mandatory outcome covered.** The web surface rubric requires a `validation-error` mandatory outcome. The journey has no step where user input produces a visible validation error. Step 5b's invalid bizKey produces a silent fallback, not a validation error. What must improve: Add a step that exercises frontend validation (e.g., manipulating filter state to an empty/null milestone value if the UI guards against it, or test that the list page properly validates filter parameters before submission).

4. **[blindspot] Missing precondition for multi-milestone sort testing.** Steps 7 and 8 test alphabetical sorting by milestone name, but the setup only guarantees "at least one milestone." Sorting requires at least 2+ milestones with distinct names to verify ordering. Quote from journey setup: "At least one milestone exists in the team in a non-cancelled state." What must improve: Setup should state "Multiple milestones exist with distinct names."

5. **[blindspot] No test for milestone filter persistence across page navigation.** The journey tests team-switch reset (Step 7b) but doesn't test whether the filter persists when navigating away and back (e.g., click an item detail then return to list). PRD doesn't explicitly require this, but it's a common UX expectation that should be addressed. What must improve: Either add a step or add an invariant stating filter persistence behavior.

6. **[blindspot] Step 6c outcome is vague about fallback value.** The journey says "Milestone column shows a fallback value for all rows" but doesn't specify what the fallback is. PRD Story 13 says display "--". Quote from journey: "Milestone column shows a fallback value for all rows." What must improve: Specify the exact fallback display value to match PRD.

7. **[blindspot] No test for permission-denied state.** The setup assumes `milestone:read` permission but never tests what happens without it. PRD Story 14 covers read-only roles and permission denial. The journey should include at least one step or edge case for unauthorized access. Quote from journey: "User has milestone:read permission." What must improve: Add a step for "User lacks milestone:read permission — items list page milestone filter is hidden or disabled."

8. **[blindspot] Invariant 2 conflates list badge and table placeholder.** The invariant says "Unassigned MIs consistently display without a badge or with a placeholder" but these are two different UI contexts with different behaviors. Quote from journey: "Unassigned MIs consistently display without a badge or with a placeholder, and always sort to the bottom of milestone column sorts." What must improve: Split into separate invariants for list view (no badge) and table view (placeholder value).

---

## Score Summary

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| 1. Completeness | 165/200 | 120 | Yes |
| 2. Semantic Purity | 165/200 | 120 | Yes |
| 3. Precondition Exclusivity | 118/150 | 90 | Yes |
| 4. Fact Alignment | 129/150 | 90 | Yes |
| 5. Surface Fitness | 125/150 | 90 | Yes |
| 6. Internal Consistency | 127/150 | 90 | Yes |
| **Total** | **829/1000** | | **Yes** |

**Overall: PASS** (all dimensions above minimum threshold)

---

## Improvement Priority

1. **[Critical] Add zero-milestone team state test** — PRD explicitly specifies this case.
2. **[Critical] Add validation-error mandatory outcome** — required by web surface rubric.
3. **[High] Fix missing fact annotations** on Steps 6e, 7, 8.
4. **[High] Add multi-milestone setup precondition** for sort steps.
5. **[Medium] Clarify Step 6c fallback value** to match PRD's "--".
6. **[Medium] Decouple implementation terms** (bizKey, soft-deleted, API call) from user-facing language.
7. **[Low] Add permission-denied edge case** for completeness.
8. **[Low] Split Invariant 2** into list-specific and table-specific invariants.
