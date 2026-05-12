---
date: "2026-05-12"
doc_dir: "docs/features/milestone-map/prd/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval -- Iteration 3

**Score: 86/100** (target: 90)
**Mode**: A (with UI)

| Dimension | Score | Max | Verdict |
|-----------|-------|-----|---------|
| Background & Goals | 14 | 15 | Strong. Three elements present, goals quantified, minor baseline precision issue persists. |
| Flow Diagrams | 19 | 20 | Good. Diamonds and error branches present. Missing state-transition validation decision node. |
| Functional Specs | 18 | 20 | Improved. UF-4 and UF-6 validation rules now partially actionable. Remaining gap: UF-3 validation rules are thin. |
| User Stories | 21 | 30 | Story 1 split partially addressed (1b extracted), but Story 1 still compounds create+edit+delete. Stories 3/5/6 gained some error ACs but gaps remain. |
| Scope Clarity | 14 | 15 | Concrete deliverables, explicit out-of-scope, consistent cross-section. Minor list API coverage gap. |

---

## Previous Issues Check

### Attack 1 (from iteration 2): UF-4 and UF-6 validation rules are behavioral descriptions, not validation rules

**Status**: ADDRESSED. UF-4 Validation Rules now reads: "筛选值必须是当前团队内已存在的里程碑 bizKey 或 `all`（全部）或 `unassigned`（未分配），其他值不产生筛选效果并回退到 `all`" and "切换团队时筛选器重置为 `all`，下拉选项刷新为新团队的里程碑列表" and "下拉选项排除 `cancelled` 状态的里程碑". These are concrete, actionable validation rules with accepted values and rejection behavior. UF-6 Validation Rules now reads: "排序方向仅接受 `asc`（升序，默认）或 `desc`（降序），按里程碑名称字母序排列" and "筛选值必须是有效里程碑 bizKey 或 `all`（全部）或 `unassigned`（未分配）" and "未分配里程碑的记录（milestone_key 为空）在筛选 `unassigned` 时显示，排序时按空值处理（排末尾）". These are proper validation rules. Fully addressed.

### Attack 2 (from iteration 2): Story 1 is a compound story with 4 distinct actions

**Status**: PARTIALLY ADDRESSED. Story 1 now reads: "PM 创建、编辑和删除里程碑" and the state switching has been extracted into Story 1b: "PM 切换里程碑状态". However, Story 1 still combines three actions: create, edit, and delete. Each has different permissions (milestone:create, milestone:update, milestone:delete), different validation rules, and different error paths. The "I want to" clause reads: "创建里程碑（名称+计划日期）、编辑里程碑信息、删除里程碑" -- three distinct verbs. These should be three separate stories. Partial credit for extracting state switching.

### Attack 3 (from iteration 2): Stories 3, 5, 6 lack error-case and boundary ACs

**Status**: PARTIALLY ADDRESSED. Story 3 now has AC6 (0 milestones empty state), AC7 (API 500 error with retry button), AC8 (in_progress milestone with 0 associated MI shows completion 0). These are good additions. Story 5 now has AC3 (data load failure shows "--"), AC4 (orphaned milestone_key to deleted milestone shows "--"), AC5 (sort direction descending, unassigned at end). Story 4 already had AC3 (timeout). Story 6 now has AC4 (empty state), AC5 (API timeout/500 with retry). Significant improvement. Remaining gap: Story 6 still lacks an AC for what management sees when switching teams (multi-team scenario). Story 3 still lacks an AC for drag-and-drop failure during timeline interaction. But these are now minor gaps compared to iteration 2.

---

## Dimension-by-Dimension Scoring

### 1. Background & Goals (14/15)

**Three elements (Reason/Target/Users): 5/5** -- All three present and specific. Reason quantifies the problem (30-50 MainItems, 20-30 min per project, 8 hours/month wasted). Target lists 5 concrete deliverables. Users table covers 3 roles (PM, team member, management) with distinct usage scenarios.

**Goals quantified: 4/4** -- Numeric targets present: "从 25 分钟/项目降至 5 分钟/项目" (80% efficiency), "3 个现有页面增加里程碑维度". Goals 1 and 3 lack independent numeric targets but the overall set is quantified. Sufficient.

**Logical consistency: 5/6** -- Goals logically follow from the stated problem. One persistent issue: background states "每次跨阶段状态盘点需逐条扫描甘特图或列表视图，单个项目每周约需 20-30 分钟人工对齐进度" but the goal metric uses "25 分钟/项目" as baseline. The 25 is the midpoint of 20-30, which is a reasonable estimate but not a measured baseline. This was flagged in iterations 1 and 2 and remains unchanged. Deduction: -1.

### 2. Flow Diagrams (19/20)

**Mermaid diagram exists: 7/7** -- Complete Mermaid flowchart with 16+ nodes. Well-structured with styling.

**Main path complete: 7/7** -- Full lifecycle covered: Create -> NotStarted -> BindMI -> InProgress -> Completed -> Cancelled, with UnbindAll at terminal state. All four states represented.

**Decision points + error branches: 5/6** -- Three diamond decision nodes: `Validate` (field validation), `APICheck` (API response), `BindCheck` (team membership). Three error branches: `ValidationErr`, `CreateErr`, `BindErr`. The diagram covers the create-and-bind happy path well. However, the state machine table in prd-spec.md defines transition rules (e.g., cancelled is terminal, completed cannot revert to in_progress) but the diagram does not include a decision node for state transition validation. The flow goes `InProgress --> AllDone{...}` and then directly to `Completed` without checking whether the transition is valid (what if PM tries to mark a not_started milestone as completed?). Similarly, the `Cancelled` state shows "PM 取消" from multiple states but no decision node checks whether the current state allows cancellation (the state machine says all states except cancelled can transition to cancelled, but this validation is not depicted). Deduction: -1.

### 3. Functional Specs (18/20)

**Placement & Interaction completeness: 7/7** -- All 6 UI Functions have explicit Placement (mode, target page, position). User Interaction Flows are numbered and cover the full path. Navigation Architecture section provides page composition and rules. No gaps.

**Data Requirements & States clarity: 6/7** -- UF-1 through UF-3 have complete data tables with Field/Type/Source/Notes columns (UF-1: 8 fields, UF-2: 2 fields appropriate for its scope, UF-3: 5 fields). UF-4 has 2 fields (milestone list, MI milestone name) -- missing MI status/completion data that the filter would need to function alongside the list display. UF-5 has 2 fields (appropriate for a dropdown selector). UF-6 has 1 field (milestone name) -- thin but adequate for a read-only column. State tables are filled for all 6 functions with explicit triggers. Deduction: -1 for UF-4 data table being thin (missing fields like MI status that are needed when the filter interacts with the list), and UF-6 being minimal.

**Validation Rules explicit: 5/6** -- Significant improvement from iteration 2. UF-4 now has proper validation rules: accepted values (bizKey, all, unassigned), rejection behavior (fallback to all), team-switch reset, cancelled exclusion. UF-6 now has proper rules: accepted sort values (asc/desc), accepted filter values, null handling for unassigned. UF-1 and UF-2 have clear character limits and required-field rules. UF-5 has proper optional-field and exclusion rules. Remaining gap: UF-3 (里程碑详情面板) validation rules are thin: "状态切换：只能按状态机允许的方向切换" and "删除：需二次确认". The state machine rule references the state machine table but does not specify which transitions are valid in the UI context (e.g., the UI should disable buttons for invalid transitions, not just show an error after clicking). "需二次确认" is a UX requirement, not a validation rule. Deduction: -1.

### 4. User Stories (21/30)

**Coverage: one story per target user: 6/7** -- All 3 user types from the background have stories: PM (Stories 1, 1b, 2, 3, 5), team member (Story 4), management (Story 6). The management user story covers read-only timeline access with permission checks. Deduction: -1 because the background says management wants to "快速了解项目整体进度和阶段分布" which implies a summary/overview capability beyond the standard timeline view (e.g., aggregated metrics, cross-project comparison), but Story 6 only covers the same timeline view that PM and team members see, just in read-only mode. The management persona's unique need (quick overview without detail) is not differentiated from the team member's view.

**Format correct: 5/7** -- All stories follow As a / I want / So that format. Stories 1b, 2, 3, 4, 5, 6 have concrete, single-action "I want" clauses. Deduction: -2 because Story 1 still compounds three actions: "创建里程碑（名称+计划日期）、编辑里程碑信息、删除里程碑". The "I want to" clause contains three distinct verbs with three different permission requirements (milestone:create, milestone:update, milestone:delete). This was flagged in iterations 1 and 2 and has not been fully split. The extraction of state switching into Story 1b was a partial fix, but create/edit/delete remain fused.

**AC per story (Given/When/Then): 6/6** -- All stories have ACs in Given/When/Then format. Story 1 has 7 ACs, Story 1b has 4 ACs, Story 2 has 6 ACs, Story 3 has 8 ACs, Story 4 has 3 ACs, Story 5 has 5 ACs, Story 6 has 5 ACs. Well done.

**AC verifiability & boundary coverage: 4/10** -- Stories 1 and 1b have strong AC coverage with boundary values (100/101 chars), error cases (API 500, concurrency conflict), and empty input. Story 2 has network error rollback and empty-team edge case. Story 3 gained error ACs (API 500 at AC7, empty state at AC6, boundary at AC8 for 0 MI with in_progress). Story 5 gained orphaned reference AC (AC4: deleted milestone shows "--") and sort direction AC (AC5). Story 6 gained API timeout AC (AC5) and empty state AC (AC4). These are improvements.

However, significant gaps remain:

1. **Story 3**: No AC for drag-and-drop failure during timeline interaction. UF-1 step 6 describes drag-and-drop for MI reassignment, but no AC tests the drag-and-drop operation failing (e.g., API error during reassignment via drag). The drag failure is covered in Story 2 AC5 but from the MainItem editing perspective, not from the timeline interaction perspective. Additionally, AC2 tests zoom (week/month/quarter) but does not verify that MI positions recalculate correctly -- "位置重新排列" is subjective without defining what correct recalculation means.

2. **Story 4**: Only 3 ACs. AC1 tests read-only view. AC2 tests disabled create button. AC3 tests API timeout. No AC for what the team member sees when milestones have cancelled status -- cancelled milestones per the state machine should have no associated MI (they are unbound), but the timeline might still show them. No AC for the team member viewing a team with 0 milestones (empty state for non-PM users). Story 6 AC4 covers empty state for management but Story 4 does not cover it for team members.

3. **Story 5**: AC5 tests sort direction but only tests descending order and unassigned-at-end behavior. No AC testing ascending order (default). No AC for filter interaction -- what happens when filtering by a specific milestone and then sorting. No AC for the interaction between milestone filter and existing filters in the table view.

4. **Story 6**: AC1 tests read-only with no create/edit/delete buttons. AC2 tests zoom. AC3 tests 403. AC4 tests empty state. AC5 tests API failure. But no AC for what happens when management switches between teams -- the background says management wants "快速了解项目整体进度和阶段分布" which implies potentially viewing multiple teams, but no AC covers team switching.

Deduction: -6 total across these gaps.

### 5. Scope Clarity (14/15)

**In-scope items are concrete deliverables: 5/5** -- All 9 in-scope items are specific: table name (pmw_milestones), API operations (CRUD), permission codes (milestone:create/update/delete/read), page URL (/milestones), page modifications with specific locations. No vague areas.

**Out-of-scope explicitly lists deferred items: 4/4** -- 5 items explicitly named: milestone reports/export, notifications, Gantt integration, status history, progress records. Each is a concrete feature, not a vague area.

**Scope consistent with functional specs and user stories: 5/6** -- In-scope items align with UI Functions (UF-1 through UF-6) and user stories (Stories 1-6). The CRUD API is covered by stories, timeline by UF-1/Story 3, existing page integration by UF-4/5/6 and Stories 2/4/5. Deduction: -1 because the in-scope "CRUD API" includes a "list" operation but no user story explicitly tests the milestone list API's behavior (pagination, sorting, filtering milestones themselves). Stories cover viewing milestones on the timeline (Story 3) and in dropdowns (Story 2) but not the raw list endpoint's behavior such as response format, pagination boundaries, or sort order.

---

## Summary

| Dimension | Score | Max | Key Issue |
|-----------|-------|-----|-----------|
| Background & Goals | 14 | 15 | Baseline precision (20-30 range vs 25 midpoint) |
| Flow Diagrams | 19 | 20 | No decision node for state-transition validation |
| Functional Specs | 18 | 20 | UF-3 validation rules are thin (state machine reference + UX requirement) |
| User Stories | 21 | 30 | Story 1 still compounds 3 actions; Stories 4/5/6 AC gaps |
| Scope Clarity | 14 | 15 | List API not explicitly tested in stories |
| **Total** | **86** | **100** | |

---

## Attack Points

### Attack 1: User Stories -- Story 1 still compounds create, edit, and delete into one story

**Where**: prd-user-stories.md Story 1: "I want to 创建里程碑（名称+计划日期）、编辑里程碑信息、删除里程碑"

**Why it's weak**: Story 1 contains three distinct actions (create, edit, delete), each requiring different permissions (milestone:create, milestone:update, milestone:delete), each with different validation rules and error paths. The ACs span 7 criteria covering three different operations, making it impossible to independently track or prioritize. A developer completing "create" but not "delete" would leave this story in an ambiguous state. This was flagged in iterations 1 and 2. Iteration 2 extracted state switching into Story 1b (good), but create/edit/delete remain fused. Three actions means three stories.

**What must improve**: Split Story 1 into three separate stories: "PM 创建里程碑" (with ACs 1-4 covering create validation, boundary, error), "PM 编辑里程碑信息" (with AC 5 covering edit + AC 6 covering concurrency), "PM 删除里程碑" (with AC 7 covering delete + transaction unbind).

### Attack 2: User Stories -- Stories 4 and 5 have thin AC coverage with missing edge cases

**Where**: prd-user-stories.md Story 4 has only 3 ACs; Story 5 lacks interaction testing ACs.

**Why it's weak**: Story 4 (team member view) has only 3 ACs covering read-only view, disabled button, and API timeout. Missing: (a) empty state for non-PM users (Story 6 AC4 covers this for management but Story 4 does not), (b) what the team member sees for cancelled milestones on the timeline (the state machine says cancelled milestones have no associated MI, but should cancelled milestones still appear on the timeline?). Story 5 (table view) AC5 only tests descending sort and unassigned-at-end. Missing: (a) ascending sort verification, (b) filter-then-sort interaction, (c) interaction between milestone filter and existing table filters.

**What must improve**: Add at least 2 ACs to Story 4 (empty state for team member, cancelled milestone display). Add at least 1 AC to Story 5 verifying default ascending sort and filter+sort interaction.

### Attack 3: Functional Specs -- UF-3 validation rules are thin and not actionable

**Where**: prd-ui-functions.md UF-3 Validation Rules: "状态切换：只能按状态机允许的方向切换" and "删除：需二次确认".

**Why it's weak**: "只能按状态机允许的方向切换" is a reference to another document, not an actionable validation rule. A developer reading this cannot determine which UI buttons should be disabled for which states. For example: should the "mark as completed" button be disabled when the milestone is in not_started state? The state machine table in prd-spec.md says not_started can only go to in_progress or cancelled, but the UF-3 validation rule does not encode this. "需二次确认" is a UX requirement (confirmation dialog), not an input validation rule. Compare with UF-4 which now has specific accepted values and rejection conditions. UF-3 should enumerate: "not_started 状态下只允许切换到 in_progress 或 cancelled；in_progress 状态下只允许切换到 completed 或 cancelled；completed 状态下只允许切换到 cancelled；cancelled 状态下不可切换" and specify that invalid transition buttons should be disabled.

**What must improve**: Replace the state machine reference with an explicit enumeration of valid transitions per state. Specify that UI buttons for invalid transitions should be disabled (not just rejected on click). Replace "需二次确认" with a concrete validation rule about what input triggers deletion.
