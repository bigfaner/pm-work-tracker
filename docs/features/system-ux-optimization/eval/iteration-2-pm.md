# Eval Iteration 2 — PM Report

**Evaluator**: Senior PM (Adversary)
**Date**: 2026-06-02
**Mode**: A (Feature WITH UI)

---

## Iteration 1 Status

All 20 attack points from iteration 1 have been addressed. This is a strong revision.

---

## Phase 1: Reasoning Audit

### Problem → Solution

The problem statement identifies two categories: (1) a blocking permission bug for member users, and (2) 15 UX defects across status transitions, forms, filtering, sorting, and navigation. The 16 in-scope items map directly to these problems with no speculative additions. **Chain holds.**

### Solution → Evidence

Metrics are attributed to "PM 团队反馈估算" and "基于 PM 团队日常使用场景分析" rather than analytics data. This is an improvement from iteration 1 (unattributed), but the estimates remain unverifiable. For a batch UX fix targeting internal PM tool users, this is acceptable as directional targets, but the "from ~30% to <5%" and "from ~15 to 0" metrics cannot be measured without first establishing a measurement method. **Chain weakened by lack of measurement method.**

### Evidence → Success Criteria

The goals have metrics but no measurement mechanism. For example, "减少表单提交失败率" targets <5% but does not specify how this will be measured (frontend error tracking? backend logs? manual testing?). The success criteria in user stories are user-observable behaviors, which is correct, but the goals-level metrics are disconnected from how they would be verified. **Partial gap.**

### Self-Contradiction Check

- Terminal state definition in prd-spec ("已关闭" or "已完成") is consistent across UF-8, Story 9, and scope items.
- Form clear-on-close behavior is consistent between Story 4 AC, UF-4 states, and prd-spec flow.
- Delete cascade behavior is consistent between Story 3, UF-3, and prd-spec flow.
- Scope item #6 now aligns with Story 4 AC (both say "所有新增/转换表单").
- Concurrency requirements in prd-spec are consistent with move and delete flows.

No contradictions found.

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals (100 pts)

**Background has three elements (0-30): 28/30**
- Reason: Clearly articulated two categories of problems with reference to issue numbers.
- Target: Two-phase implementation of 16 items, well-structured.
- Users: Table with 3 user types, counts, and affected items.
- Deduction (-2): User counts ("5名", "3名", "3人") are presented without explaining the source -- are these all current users? Are there more?

**Goals are quantified (0-30): 24/30**
- All 6 goals have numeric targets.
- Deduction (-6): Three metrics use "PM 团队反馈估算" as the sole source, with no measurement method. "从每次 30 秒降至 0 秒" is directional but not instrumented. "约 30% 降至 <5%" requires defining what constitutes a "submission failure." "每周约 15 次降至 0" is a strong claim -- 0 is a hard number that requires a clear definition of what counts as a "manual workaround."

**Background and goals are logically consistent (0-40): 38/40**
- Every goal maps to specific issue numbers.
- The blocking bug goal correctly targets member users.
- Deduction (-2): The "提升过滤效率" goal claims "预估覆盖 40% 的过滤需求" but this is ambiguous -- 40% of what? All filter operations? All user sessions? Without defining the denominator, this metric cannot be evaluated.

**Dimension Score: 90/100**

---

### Dimension 2: Flow Diagrams (150 pts)

**Mermaid diagram exists (0-50): 50/50**
- Single comprehensive mermaid flowchart covers all major operations.

**Main path complete (0-50): 46/50**
- Status transition: start → backend check → error/confirm → execute → page update. Complete.
- Conversion form: open → fill → submit → success/failure → clear. Complete.
- Delete: confirm → transaction → success/failure → retry/abandon. Complete.
- Move: select target → validate → confirm → execute. Complete.
- Filter: select conditions → backend filter → display results. Complete.
- Member permission fix: documented in text flow (4 steps), not in diagram.
- Deduction (-4): Member permission fix (#8) is the single blocking bug but has no representation in the mermaid diagram. While the text flow is adequate, the diagram explicitly covers other flows but not this critical one.

**Decision points + error branches (0-50): 48/50**
- Status transition: diamond for "可流转?" and "是否终态?" with error branch and recovery (retry/abandon).
- Conversion form: diamond for "必填项已填?" with disabled state loop; diamond for "后端校验" with failure branch.
- Delete: "删除事务" with success/failure branches, failure leads to retry/abandon.
- Move: "目标有效?" with rejection branch.
- Filter: "因子事项匹配?" with two display paths.
- Deduction (-2): The move flow's error path ("提示拒绝原因") is a terminal node with no recovery path back to target selection. User must re-initiate the entire flow.

**Dimension Score: 144/150**

---

### Dimension 3: Functional Specs (200 pts)

**Placement & Interaction completeness (0-70): 65/70**
- All 10 UI functions have explicit Placement sections with Mode, Target Page, and Position.
- User Interaction Flow covers full paths for all functions.
- Page Composition table maps all functions to pages.
- Deduction (-3): UF-10 (团队选择器过滤) has a weak interaction flow -- step 3 says "若仅有一个团队，可省略下拉或固定展示" with "可" (may), which is ambiguous. Is this a requirement or a suggestion?
- Deduction (-2): UF-5 (子事项排序) has only 2 interaction steps (open page → see sorted list). While the function is simple, no interaction is described for what happens when new sub-items are added during the session -- does the list auto-refresh? Does the user need to reload?

**Data Requirements & States clarity (0-70): 66/70**
- All UI functions have Data Requirements tables with Field, Type, Source, Notes.
- All UI functions have States tables with State, Display, Trigger.
- UF-7 has explicit API response structure (matchType, matchedSubItemIds).
- UF-6 has full API specification with request/response/error codes.
- Deduction (-2): UF-9 (甘特图修复) Data Requirements table has "时间范围" as a single field but does not specify the source API or data structure. What endpoint provides the "可见主事项数据"? Is this the same list endpoint or a dedicated gantt endpoint?
- Deduction (-2): UF-3 (删除) Data Requirements lists "子事项数量" and "权限码" but does not specify the API endpoint for delete operations (unlike UF-6 which has a full API spec). This inconsistency suggests the spec was addressed for UF-6 specifically but not uniformly applied.

**Validation Rules explicit (0-60): 55/60**
- UF-1: 2 rules, both actionable ("错误消息必须来自后端", "tooltip 被完全替代").
- UF-2: 4 rules including date constraints. Actionable.
- UF-3: 2 rules, actionable (role-based visibility, confirm dialog).
- UF-4: 4 rules, actionable (required fields, disabled state, clear behavior).
- UF-5: 1 rule ("排序在服务端完成"). This is implementation, not validation. Deduction.
- UF-6: 2 rules, actionable (target constraints).
- UF-7: 3 rules with detailed AND/OR logic. Strong.
- UF-8: 2 rules, actionable.
- UF-9: 2 rules, both actionable.
- UF-10: "无特殊校验" -- acceptable for a server-filtered dropdown.
- Deduction (-3): UF-5's validation rule "排序在服务端完成" describes implementation, not a testable validation rule. What should be validated? That items appear in creation-time order? This was partially fixed (Story 5 now says "创建时间倒序" instead of "id 倒序") but UF-5 still references "ORDER BY id DESC" in Data Requirements, which contradicts "创建时间倒序" if id and creation time are not guaranteed to correlate.
- Deduction (-2): UF-4 does not specify what happens when the user opens a conversion form, partially fills it, then switches to a different conversion form type. Does the clear-on-close apply when switching form types mid-session?

**Dimension Score: 186/200**

---

### Dimension 4: User Stories (200 pts)

**Coverage: one story per target user (0-50): 45/50**
- PM users: Stories 1-5, 7-10 (10 stories). Well covered.
- Member users: Story 6. One story, covers the blocking permission bug.
- macOS users: Story 10 covers scroll bar but this is a platform-specific issue, not a distinct user persona. No macOS-specific story for the scroll bar aspect exists separately -- it's bundled into Story 10 with the gantt time range fix.
- Deduction (-5): The Background identifies "macOS 用户（3人）" as a distinct user type, but Story 10 bundles gantt time range (all users) with macOS scroll bar (macOS users). These are different concerns for different audiences. The macOS user type from the Background has no dedicated story that describes their specific experience (scroll bar visibility).

**Format correct (0-50): 48/50**
- All 11 stories follow As a / I want / So that format.
- Actions are concrete ("看到具体的错误原因", "直接修改开始时间", "删除误建的主事项和子事项").
- Deduction (-2): Story 11's "So that" clause mixes two benefits: "团队选择器只展示我有权限的团队" is about the team selector, but the story also covers weekly progress page filtering. These are two distinct features in one story. The "So that" only addresses the team selector benefit, leaving the weekly progress filtering motivation unstated.

**AC per story in Given/When/Then (0-50): 48/50**
- All 11 stories have ACs in Given/When/Then format.
- Most stories have multiple ACs covering different scenarios.
- Deduction (-2): Story 2 has a single AC that is effectively a happy-path-only test: "弹窗渲染完成" → "弹窗中包含'开始时间'字段，修改后可成功保存". This AC combines rendering AND saving in one clause. If the save fails, the AC does not specify expected behavior. The validation rules in UF-2 cover date constraints, but the story AC does not reference them.

**AC verifiability & boundary coverage (0-50): 44/50**
- Most ACs are objectively testable.
- Story 3: 3 ACs covering PM delete main, PM delete sub, non-PM visibility. Good boundary coverage.
- Story 4: 3 ACs covering disabled state, required fields, form clear. Good.
- Story 7: 3 ACs covering valid move, closed target rejection, same-target rejection. Good.
- Story 8: 3 ACs covering filter by assignee, visual indicator, no-filter state. Good.
- Story 9: 3 ACs covering terminal sort, default filter, zero-filter state. Good.
- Deduction (-3): Story 6 AC1 says "用户能看到至少：待办事项提交(item_pool:submit)、事项查看(main_item:list)". The "至少" (at least) qualifier makes this non-deterministic -- how many more menu items are acceptable? The AC cannot objectively fail unless zero items are shown.
- Deduction (-3): Story 11 AC2 contains the "活跃事项定义" which is a complex multi-condition business rule embedded in an AC. The definition references database fields (`status_history`, `created_at`, `updated_at`) and conditions (a), (b), (c) with sub-conditions. While precise, this is a data-level specification, not a user-observable behavior. A tester cannot verify this through UI interaction alone -- they would need to inspect database state.

**Dimension Score: 185/200**

---

### Dimension 5: Scenario Completeness (150 pts)

**End-to-end scenario coverage (0-60): 55/60**
- Status transition: full lifecycle from click → error/confirm → execute → update.
- Delete: full lifecycle from click → confirm → transaction → record history → update.
- Conversion form: full lifecycle from open → fill → submit → success/failure → clear.
- Move: full lifecycle from select → validate → confirm → execute → update.
- Filter: full lifecycle from select → backend filter → display with indicators.
- Member permission: full lifecycle from fix → login → menu load → page access.
- Deduction (-3): Gantt chart fix (UF-9, Story 10) describes the fix at a technical level ("起始日期取最早开始时间的前1天") but does not describe a user scenario. What does the user see before and after the fix? What is the scrolling experience on macOS before and after? The scenario is incomplete from the user's perspective.
- Deduction (-2): Team selector filtering (UF-10, Story 11 partial) does not describe what happens when a user has zero team permissions. Is this possible? What does the UI show? The "single team" state is described but the "zero team" state is not.

**Implicit assumptions surfaced (0-40): 35/40**
- Performance assumption (1000 main + 5000 sub items) is stated explicitly.
- No pagination is assumed and called out as future risk.
- Soft delete chosen explicitly over hard delete.
- Terminal state values are enumerated.
- Deduction (-3): The "终态定义" section says "非终态状态包括：`待处理`、`进行中`、`已暂停` 等" -- the "等" (etc.) is ambiguous. Are these all non-terminal states, or are there others? If the system has a configurable status list, this should be stated. If these are fixed, "等" should be removed.
- Deduction (-2): Story 11's "活跃事项" definition references condition (c) "变更类型为进度更新" but there is no explanation of what constitutes a "进度更新" versus other update types. The `updated_at` field changes on any update; how does the system distinguish progress updates from other updates?

**Business-rules consistency (0-50): 48/50**
- Terminal state definition is consistent across all references.
- Delete cascade rules are consistent.
- Filter penetration logic is consistent between UF-7 and Story 8.
- Permission codes are consistently referenced.
- Deduction (-2): UF-5 Data Requirements says "ORDER BY id DESC" while Story 5 says "按创建时间倒序排列". If `id` is an auto-increment integer, these are equivalent. If `id` is a snowflake/UUID bizKey (as stated in the project's memory: "All IDs returned to frontend must use bizKey"), then `ORDER BY id DESC` may not produce creation-time ordering. This inconsistency needs resolution.

**Dimension Score: 138/150**

---

### Dimension 6: Edge Case Coverage (100 pts)

**Error paths documented (0-40): 36/40**
- Status transition failure: documented with specific error message flow.
- Delete transaction failure: documented with retry/abandon paths.
- Conversion form submission failure: documented with field retention.
- Move validation failure (closed/same target): documented.
- Concurrent access errors: documented with 3 specific scenarios.
- Deduction (-2): Move operation API returns 404 for "子事项或目标主事项不存在" but no user-facing scenario describes what happens when a sub-item is opened, then deleted by another user, and then the first user attempts to move it. The concurrent access section covers delete-during-move but not the reverse.
- Deduction (-2): The gantt chart fix does not describe error or edge cases. What happens with zero main items? What happens with a single main item spanning one day? The "前1天 / 后1天" calculation could produce a nonsensical range.

**Boundary conditions covered (0-35): 30/35**
- Empty state handling: documented with specific guidance for all filter/search scenarios.
- Concurrent access: documented for move, delete, and status transition.
- Zero filter selection: documented (shows all items).
- Deduction (-2): No maximum length or boundary specified for the error message string in UF-1. If the backend returns an extremely long error message, how should the UI handle it? Truncation? Wrapping? This is an edge case for the Alert component.
- Deduction (-3): The performance requirement states "当前不分页（全量返回）" but does not address what happens when data exceeds the 1000+5000 benchmark. The spec says "若未来引入分页需重新评估穿透逻辑" but does not define a threshold for when this becomes necessary. What happens at 2000+10000? Is there a degradation plan?

**Failure recovery described (0-25): 22/25**
- Delete failure: retry via confirmation dialog re-trigger.
- Status transition failure: retry by re-clicking after seeing error message.
- Conversion form failure: retry with retained field values.
- Move failure: user must re-initiate (no automatic recovery).
- Deduction (-3): The move flow's failure path at the mermaid diagram level shows "提示拒绝原因" as a terminal node. There is no explicit recovery path back to target selection -- the user must restart the entire move flow. While this is a reasonable UX choice, it should be documented as intentional rather than appearing as a gap.

**Dimension Score: 88/100**

---

### Dimension 7: Scope Clarity (100 pts)

**In-scope items are concrete deliverables (0-35): 34/35**
- All 16 items are specific features with clear descriptions.
- Phased implementation is well-defined (10 + 6).
- Checklist format makes tracking easy.
- Deduction (-1): Item #10 says "状态过滤器改为多选 + 负责人过滤穿透子事项（卡片视图和表格视图统一支持）" which combines two distinct features (multi-select status filter + assignee penetration) into one scope item. These could fail or succeed independently.

**Out-of-scope explicitly lists deferred items (0-30): 30/30**
- 7 items explicitly listed as out of scope.
- "已删除事项的恢复功能" is now explicitly excluded with explanation.
- "批量操作" and "子事项移动历史记录" are named, not implied.
- "新增权限码之外的其他 seed 数据更新（由 RBAC 提案覆盖）" provides context for deferral.

**Scope consistent with functional specs and user stories (0-35): 33/35**
- All 16 in-scope items map to UI functions and user stories.
- Out-of-scope items (hard delete, restore, batch ops) are not described in UI functions or stories.
- Deduction (-2): Scope item #14 "修复 macOS 下甘特图水平滚动条不显示问题" is implementation-specific (names the OS and the specific defect). This is a bug fix, not a feature. While legitimate to include, its presence in a PRD scope list alongside feature items creates a category mismatch that could confuse prioritization.

**Dimension Score: 97/100**

---

## Phase 3: Blindspot Hunt

**B1. [blindspot] Move operation leaves orphaned numbering gap** -- When a sub-item is moved from Main Item A to Main Item B, Main Item A's remaining sub-items retain their original codes (e.g., A-1, A-3 if A-2 was moved). The spec says the moved item gets a new code via NextSubCode, but does not address whether the source main item's sub-items are renumbered to close the gap. This could lead to confusing non-sequential numbering (A-1, A-3, A-5) that worsens over time. The spec should state explicitly whether source renumbering is in or out of scope.

**B2. [blindspot] "因子事项匹配" badge has no dismiss or acknowledge behavior** -- UF-7 specifies that indirect-match main items display a "因子事项匹配" badge, but does not specify whether this badge is informational-only or interactive. If the user clicks the badge, what happens? If the badge is persistent, it could become visual noise in heavily filtered views with many indirect matches.

**B3. [blindspot] Weekly progress page active-item filtering creates a discoverability problem** -- Story 11 hides main items that are terminal AND have no activity this week or last week. But if a PM wants to check on a completed item from two weeks ago, there is no mechanism described for revealing hidden items. UF-8 describes a checkbox-based filter for the overall progress page, but Story 11's weekly progress page has no equivalent "show hidden items" control. Users may assume deleted data rather than filtered data.

**B4. [blindspot] Delete confirmation dialog does not specify what happens with 0 sub-items** -- Story 3 AC says the dialog shows "将同时删除 N 个子事项". UF-3 Data Requirements says "子事项数量: number, 确认对话框中提示". When N=0, should the dialog still show "将同时删除 0 个子事项" or should it use different wording? This edge case is not addressed.

**B5. [blindspot] Filter penetration does not address the "all sub-items match" case** -- UF-7 says indirect-match main items show only matched sub-items. But what if ALL sub-items under a main item match the assignee filter? The main item would still be marked "因子事项匹配" (indirect) even though it could reasonably be considered a direct match. This creates a visual inconsistency where a main item with all-matching sub-items displays the same badge as one with partial matches.

---

## Final Score

| Dimension | Score |
|-----------|-------|
| 1. Background & Goals | 90/100 |
| 2. Flow Diagrams | 144/150 |
| 3. Functional Specs | 186/200 |
| 4. User Stories | 185/200 |
| 5. Scenario Completeness | 138/150 |
| 6. Edge Case Coverage | 88/100 |
| 7. Scope Clarity | 97/100 |
| **Total** | **928/1000** |

---

## Attacks Summary

1. [Background & Goals]: Metrics lack measurement methods -- "约 30% 降至 <5%" and "每周约 15 次降至 0" are attributed to PM team estimates but no measurement method is defined (what constitutes a "submission failure"? what counts as a "manual workaround"?) -- must add measurement method for each metric or explicitly label as uninstrumented directional targets.

2. [Background & Goals]: "覆盖 40% 的过滤需求" has undefined denominator -- the goal claims coverage percentage but does not define what constitutes the total universe of filter needs -- must either define the denominator or replace with a concrete, countable metric.

3. [Flow Diagrams]: Member permission fix (#8) has no mermaid representation -- the diagram covers status transition, conversion form, delete, move, and filter flows but the critical blocking bug fix has no diagram node -- must add a minimal member-login flow to the diagram.

4. [Flow Diagrams]: Move flow error path has no recovery -- "提示拒绝原因" is a terminal node with no path back to target re-selection -- must add a recovery arrow or document that restart is intentional.

5. [Functional Specs]: UF-5 "ORDER BY id DESC" contradicts Story 5 "按创建时间倒序" -- if id is a snowflake bizKey, these are not equivalent -- must resolve by specifying whether sorting is by creation timestamp or by id, and ensure UF-5 and Story 5 use the same language.

6. [Functional Specs]: UF-10 step 3 uses ambiguous "可" -- "若仅有一个团队，可省略下拉或固定展示" reads as optional, not required -- must use "应" (shall) or remove the conditional if it is a UX suggestion rather than a requirement.

7. [User Stories]: Story 6 AC1 "至少" qualifier is non-deterministic -- "能看到至少：待办事项提交...、事项查看..." cannot fail unless zero items are shown -- must specify the complete expected menu set or replace with a concrete list.

8. [User Stories]: Story 11 AC2 embeds database-level specification -- the "活跃事项定义" references `status_history`, `created_at`, `updated_at` database fields and multi-condition logic that cannot be verified through UI interaction -- must separate the business rule definition (in prd-spec) from the user-observable AC (in user stories).

9. [User Stories]: Story 11 merges two unrelated features -- team selector filtering and weekly progress page filtering serve different users and pages but share one story -- must split into two stories with independent ACs and distinct "So that" clauses.

10. [Scenario Completeness]: "非终态状态包括：`待处理`、`进行中`、`已暂停` 等" -- the trailing "等" leaves the non-terminal state list open-ended -- must enumerate all non-terminal states or state that the list is configurable and reference the configuration source.

11. [Scenario Completeness]: "活跃事项" condition (c) references "变更类型为进度更新" without defining how progress updates are distinguished from other updates in the data model -- must specify the field or flag that identifies a progress update.

12. [Edge Case Coverage]: No data volume degradation plan -- spec says "当前不分页" and "若未来引入分页需重新评估" but provides no threshold for when performance degrades or what the fallback behavior is -- must define a data volume threshold or document accepted performance degradation.

13. [Edge Case Coverage]: Gantt chart edge cases unaddressed -- zero main items, single-day main items, and the "前1天/后1天" boundary produce potentially invalid date ranges -- must specify behavior for these cases.

14. [blindspot]: Move operation leaves numbering gaps in source main item -- sub-items at source retain non-sequential codes (e.g., A-1, A-3, A-5) after A-2 is moved -- must explicitly state whether source renumbering is in or out of scope.

15. [blindspot]: Weekly progress page has no "show hidden items" control -- terminal + inactive items are hidden with no user-visible mechanism to reveal them, creating a discoverability problem -- must add a filter control or explicitly document the design decision with rationale.

16. [blindspot]: Delete confirmation with 0 sub-items is unspecified -- when N=0, the dialog text "将同时删除 N 个子事项" becomes misleading -- must specify alternate wording or conditional display.
