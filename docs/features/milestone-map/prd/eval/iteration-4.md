---
date: "2026-05-12"
doc_dir: "docs/features/milestone-map/prd/"
iteration: 4
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval -- Iteration 4

**Score: 91/100** (target: 90)
**Mode**: A (with UI)

| Dimension | Score | Max | Verdict |
|-----------|-------|-----|---------|
| Background & Goals | 14 | 15 | Strong. All three elements present, goals quantified. Baseline precision remains minor. |
| Flow Diagrams | 19 | 20 | Good. Diamonds and error branches present. State-transition decision node still absent. |
| Functional Specs | 19 | 20 | UF-3 validation rules now fully enumerated with button-disable rules. Minor gap on delete confirmation detail. |
| User Stories | 25 | 30 | Story 1 properly split into 1a/1b/1c/1d. Stories 4/5/6 gained ACs. Remaining gaps in Story 4 edge cases and Story 3 drag-on-timeline specifics. |
| Scope Clarity | 14 | 15 | Concrete deliverables, explicit out-of-scope, consistent. List API still untested in stories. |

---

## Previous Issues Check

### Attack 1 (from iteration 3): Story 1 still compounds create, edit, and delete into one story

**Status**: ADDRESSED. Story 1 has been split into four stories:
- Story 1a: "PM 创建里程碑" -- "I want to 创建里程碑（名称+计划日期）"
- Story 1b: "PM 编辑里程碑信息" -- "I want to 编辑里程碑的名称和计划日期"
- Story 1c: "PM 删除里程碑" -- "I want to 删除不再需要的里程碑"
- Story 1d: "PM 切换里程碑状态" -- "I want to 切换里程碑状态（not_started/in_progress/completed/cancelled）"

Each story now has a single concrete action, distinct permissions, and its own ACs. Story 1a has 4 ACs (valid create, 100/101 char boundary, empty name, API 500). Story 1b has 2 ACs (edit save, concurrency conflict). Story 1c has 1 AC (delete + transaction unbind). Story 1d has 4 ACs (valid transition, completed->in_progress rejection, cancelled rejection, cancelled auto-unbind). Fully addressed.

### Attack 2 (from iteration 3): Stories 4 and 5 have thin AC coverage with missing edge cases

**Status**: PARTIALLY ADDRESSED. Story 4 now has 5 ACs (up from 3 in iteration 3): AC1 (read-only view), AC2 (disabled create button), AC3 (API timeout), AC4 (empty state for non-PM -- "我没有 milestone:create 权限且团队有 0 个里程碑"), AC5 (cancelled milestone display -- "已取消的里程碑节点以灰色样式显示"). The empty state and cancelled display gaps from iteration 3 are now covered.

Story 5 now has 7 ACs (up from 5). New additions: AC6 (ascending default sort -- "已分配里程碑的 MI 按里程碑名称升序（默认）排列，未分配的 MI 排在末尾"), AC7 (filter+sort interaction -- "表格视图已按里程碑筛选为 M1 且按里程碑列排序为降序，When 结果显示，Then 仅 M1 下的 MI 按名称降序排列"). These address the ascending-sort and filter+sort interaction gaps.

Remaining gap: Story 5 still lacks an AC for interaction between milestone filter and existing non-milestone filters in the table view (e.g., what happens when both status filter and milestone filter are active). However, this is a minor gap -- the filter+sort AC covers the core interaction pattern.

### Attack 3 (from iteration 3): UF-3 validation rules are thin and not actionable

**Status**: ADDRESSED. UF-3 validation rules now enumerate explicit transitions per state with button-disable rules:
- `not_started` -> `in_progress` (enabled) / `cancelled` (enabled) / `completed` (disabled, tooltip: "未开始的里程碑不可直接标记为已完成")
- `in_progress` -> `completed` (enabled) / `cancelled` (enabled) / `not_started` (disabled, tooltip: "进行中不可回退为未开始")
- `completed` -> `cancelled` (enabled) / `in_progress` (disabled) / `not_started` (disabled, tooltip: "已完成状态不可回退")
- `cancelled` -> all disabled, tooltip: "已取消的里程碑不可恢复"

The delete rule now reads: "点击删除按钮后弹出确认弹窗（非 alert），用户确认后执行删除，取消则关闭弹窗不做操作" -- specifies confirmation dialog type and both paths. This is actionable. Fully addressed.

---

## Dimension-by-Dimension Scoring

### 1. Background & Goals (14/15)

**Three elements (Reason/Target/Users): 5/5** -- All three present and specific. Reason quantifies the problem (30-50 MainItems, 20-30 min per project, ~8 hours/month wasted). Target lists 5 concrete deliverables (Milestone entity, timeline page, four-state machine, auto-calculated completion, existing page integration). Users table covers 3 roles with distinct usage scenarios.

**Goals quantified: 4/4** -- Numeric targets: "从 25 分钟/项目降至 5 分钟/项目" (80% efficiency), "3 个现有页面增加里程碑维度". Goal 1 lacks an independent numeric metric ("支持创建里程碑并关联 MainItem" is binary), but the overall goal set is sufficiently quantified with the efficiency metric as the headline number.

**Logical consistency: 5/6** -- Goals logically follow from the stated problem. One persistent issue across all iterations: background states "每次跨阶段状态盘点需逐条扫描甘特图或列表视图，单个项目每周约需 20-30 分钟人工对齐进度" but the goal metric uses "25 分钟/项目" as baseline. The 25 is the midpoint of 20-30, which is a reasonable estimate but not a measured baseline. This is a minor precision issue that has persisted since iteration 1. Deduction: -1.

### 2. Flow Diagrams (19/20)

**Mermaid diagram exists: 7/7** -- Complete Mermaid flowchart with 16+ nodes, styled with color-coded states and error nodes. Well-structured.

**Main path complete: 7/7** -- Full lifecycle: Create -> NotStarted -> BindMI -> InProgress -> Completed -> Cancelled, with UnbindAll at terminal state. All four states represented. CalcCompletion flows correctly.

**Decision points + error branches: 5/6** -- Three diamond decision nodes: `Validate` (field validation), `APICheck` (API response), `BindCheck` (team membership). Three error branches: `ValidationErr`, `CreateErr`, `BindErr`. The diagram does not include a decision node for state-transition validation. The flow shows `InProgress --> AllDone{所有关联 MI 均为 completed?}` and `InProgress -->|PM 手动标记| Completed` but no diamond checks whether the transition is valid per the state machine (e.g., what blocks `not_started -> completed`). The state machine table in prd-spec.md defines these rules, but the diagram does not encode them as decision logic. This gap persists from iteration 3. Deduction: -1.

### 3. Functional Specs (19/20)

**Placement & Interaction completeness: 7/7** -- All 6 UI Functions have explicit Placement (mode, target page, position). User Interaction Flows are numbered step-by-step and cover the full path. Navigation Architecture provides page composition and rules. No gaps found.

**Data Requirements & States clarity: 6/7** -- UF-1: 8 fields (complete for a timeline page). UF-2: 2 fields (appropriate for a simple create/edit modal). UF-3: 5 fields (appropriate for a detail panel). UF-4: 2 fields (milestone list + MI milestone name). UF-5: 2 fields (appropriate for a dropdown selector). UF-6: 1 field (milestone name). State tables are filled for all 6 functions with explicit Display and Trigger columns. Deduction: -1 for UF-4 and UF-6 data tables being thin. UF-4 is missing MI status/completion fields that would be needed for a complete list display context. UF-6 has only one field. While these are functionally adequate, they are not comprehensive compared to UF-1 through UF-3.

**Validation Rules explicit: 6/6** -- UF-1: name required (1-100 chars), date required, drag target must be same team. UF-2: name required (1-100 chars), date required. UF-3: fully enumerated state transitions per state with button-disable rules and tooltip messages. Delete rule specifies confirmation dialog type (non-alert) and both confirm/cancel paths. UF-4: accepted values (bizKey, all, unassigned), rejection behavior (fallback to all), team-switch reset, cancelled exclusion. UF-5: optional field, cancelled exclusion. UF-6: accepted sort values (asc/desc), accepted filter values, null handling for unassigned. All validation rules are actionable. Full marks.

### 4. User Stories (25/30)

**Coverage: one story per target user: 6/7** -- PM: Stories 1a, 1b, 1c, 1d, 2, 3, 5. Team member: Story 4. Management: Story 6. All three user types have stories. Deduction: -1 because the background says management wants to "快速了解项目整体进度和阶段分布" but Story 6 only provides the same timeline view in read-only mode, without any unique management capability (e.g., aggregated metrics, cross-project summary). The management persona is not differentiated from the team member persona beyond permission levels.

**Format correct: 7/7** -- All stories follow As a / I want / So that format. Each story now has a single concrete action verb: "创建里程碑", "编辑里程碑的名称和计划日期", "删除不再需要的里程碑", "切换里程碑状态", "通过时间线拖拽或主事项编辑页将 MainItem 绑定到里程碑", "在时间线图页面查看所有里程碑", "查看团队的里程碑时间线", "在表格视图中看到每个 MainItem 的里程碑归属列", "通过时间线图以只读方式查看项目里程碑". All "I want" clauses are concrete and single-action. Full marks -- the Story 1 split resolved this completely.

**AC per story (Given/When/Then): 6/6** -- All stories have ACs in Given/When/Then format. Story 1a: 4 ACs, Story 1b: 2 ACs, Story 1c: 1 AC, Story 1d: 4 ACs, Story 2: 6 ACs, Story 3: 8 ACs, Story 4: 5 ACs, Story 5: 7 ACs, Story 6: 5 ACs. All ACs follow the Given/When/Then pattern. Well done.

**AC verifiability & boundary coverage: 6/10** -- Significant improvement from iteration 3. Stories 1a-1d have strong boundary coverage (100/101 char boundary, empty name, API 500, concurrency conflict, state transition rejection). Story 2 has network error rollback and empty-team edge case. Story 3 has error ACs (API 500, empty state, 0 MI with in_progress), zoom, and 200-MI pagination. Story 4 now has empty state (AC4) and cancelled display (AC5). Story 5 now has ascending sort default (AC6) and filter+sort interaction (AC7). Story 6 has empty state (AC4) and API timeout (AC5).

Remaining gaps:

1. **Story 3 (8 ACs)**: No AC for drag-and-drop failure on the timeline itself. UF-1 step 6 describes drag-and-drop for MI reassignment on the timeline. Story 2 AC5 covers drag failure but from a data-perspective ("当我将 MainItem 拖拽至目标里程碑，When 后端 API 返回网络错误"), not from the timeline UI perspective (e.g., visual feedback, snap-back animation). Additionally, AC2 tests zoom switching ("时间轴刻度标签对应变化，里程碑和事项位置重新排列") but "位置重新排列" is not objectively verifiable -- what defines correct repositioning? No expected positions or layout rules are specified.

2. **Story 4 (5 ACs)**: AC4 tests empty state ("我没有 milestone:create 权限且团队有 0 个里程碑"). AC5 tests cancelled milestone display. But no AC tests what the team member sees when they lack milestone:read permission (the UF-1 States table includes a "No Permission" state: "403 提示页 | 用户缺少 milestone:read 权限"). Story 4 assumes milestone:read but never explicitly tests the absence of it. Story 6 AC3 tests 403 for management but Story 4 does not test 403 for team members.

3. **Story 5 (7 ACs)**: AC3 tests data load failure. AC4 tests orphaned reference. AC5 tests descending sort. AC6 tests ascending sort (default). AC7 tests filter+sort interaction. However, no AC tests what happens when a milestone is created/deleted while the table view is open (stale data scenario). Also no AC tests the interaction between the milestone filter and existing table filters (e.g., filtering by status AND milestone simultaneously).

Deduction: -4 total across these gaps (Story 3 drag-on-timeline feedback gap: -1, Story 3 zoom verifiability gap: -1, Story 4 missing 403 AC: -1, Story 5 cross-filter interaction gap: -1).

### 5. Scope Clarity (14/15)

**In-scope items are concrete deliverables: 5/5** -- All 9 in-scope items are specific: table name (pmw_milestones), API operations (CRUD), permission codes (milestone:create/update/delete/read), page URL (/milestones), page modifications with specific locations, completion calculation method, state machine definition, soft-delete behavior, database migration targets. No vague areas.

**Out-of-scope explicitly lists deferred items: 4/4** -- 5 items explicitly named: milestone reports/export (CSV, PDF), notifications (milestone due/overdue), Gantt integration, status change history, milestone-level progress records. Each is a concrete feature, not a vague area.

**Scope consistent with functional specs and user stories: 5/6** -- In-scope items align with UI Functions (UF-1 through UF-6) and user stories (Stories 1a-1d, 2, 3, 4, 5, 6). CRUD API is covered by Stories 1a-1d (create, update, delete) and Story 3 (read via timeline). Timeline is UF-1/Story 3. Existing page integration is UF-4/5/6 and Stories 2/4/5. Deduction: -1 because the in-scope "CRUD API" includes a "list" operation but no user story explicitly tests the milestone list API's behavior (pagination, sorting, filtering milestones themselves). The closest coverage is Story 3 which tests the timeline rendering, but the raw list endpoint's behavior (response format, pagination boundaries, sort order) is not covered by any AC. This gap persists from iteration 3.

---

## Summary

| Dimension | Score | Max | Key Issue |
|-----------|-------|-----|-----------|
| Background & Goals | 14 | 15 | Baseline precision (20-30 range vs 25 midpoint) |
| Flow Diagrams | 19 | 20 | No decision node for state-transition validation in diagram |
| Functional Specs | 19 | 20 | UF-4/UF-6 data tables thin |
| User Stories | 25 | 30 | Story 3 drag-on-timeline and zoom verifiability; Story 4 missing 403 AC; Story 5 cross-filter gap |
| Scope Clarity | 14 | 15 | List API not explicitly tested in stories |
| **Total** | **91** | **100** | |

---

## Attack Points

### Attack 1: User Stories -- Story 3 lacks verifiable zoom and drag-on-timeline ACs

**Where**: prd-user-stories.md Story 3 AC2: "Then 时间轴刻度标签对应变化，里程碑和事项位置重新排列"

**Why it's weak**: "位置重新排列" is subjective -- what defines correct repositioning after a zoom change? No expected layout rule or positioning constraint is specified. A developer could render items in any order and claim the AC passes. Additionally, Story 3 has no AC for drag-and-drop failure on the timeline itself (UF-1 step 6 describes this interaction). Story 2 AC5 covers drag failure from the data perspective but not the timeline UI perspective (snap-back animation, visual feedback during drag, ghost element positioning).

**What must improve**: Add an AC for zoom that specifies an objectively verifiable outcome (e.g., "week view shows daily tick marks, month view shows weekly tick marks, quarter view shows monthly tick marks"). Add an AC for drag-and-drop failure on the timeline that specifies visual feedback (e.g., "MI snaps back to original position, toast shows error message").

### Attack 2: User Stories -- Story 4 missing 403 permission AC for team members

**Where**: prd-user-stories.md Story 4 -- no AC tests the team member lacking milestone:read permission

**Why it's weak**: UF-1 States table includes "No Permission | 403 提示页 | 用户缺少 milestone:read 权限" but Story 4 only tests scenarios where the team member has milestone:read permission. Story 6 AC3 tests 403 for management ("我没有 milestone:read 权限，When 我访问 /milestones 页面，Then 页面返回 403 提示") but no corresponding AC exists for team members. This is a coverage gap -- the same permission gate should be tested for all user types.

**What must improve**: Add an AC to Story 4: "Given 我没有 milestone:read 权限，When 我访问 /milestones 页面，Then 页面返回 403 提示". This mirrors Story 6 AC3 and ensures the permission gate is verified for team members.

### Attack 3: User Stories -- Story 5 lacks cross-filter interaction AC

**Where**: prd-user-stories.md Story 5 -- no AC tests milestone filter combined with existing table filters

**Why it's weak**: Story 5 AC7 tests milestone filter + milestone sort interaction but no AC tests milestone filter combined with existing table filters (e.g., status filter + milestone filter active simultaneously). In a real table view, users commonly combine multiple filters. The absence of this AC means a bug where milestone filter overrides or conflicts with status/assignee filters could pass undetected. UF-6 validation rules specify accepted filter values for milestone column but do not address cross-filter interaction semantics.

**What must improve**: Add an AC: "Given 表格视图已按状态筛选为 '进行中' 且按里程碑筛选为 M1，When 结果显示，Then 仅显示状态为 '进行中' 且属于 M1 的 MI". This tests the conjunction of existing and new filters.
