# PRD Evaluation Report — Iteration 1 (PM Persona)

**Evaluator:** Senior PM (adversarial stance)
**Date:** 2026-06-02
**Documents evaluated:**
- prd-spec.md
- prd-user-stories.md
- prd-ui-functions.md

---

## Phase 1: Reasoning Audit (Pre-Score Anchors)

### Problem -> Solution Alignment

The document defines 16 numbered items across two phases. The solution maps 1:1 to each item. This is a collection of fixes and small enhancements, not a speculative feature. The alignment is strong: every item traces to a concrete, observable behavior change.

### Solution -> Evidence

The Goals table provides numeric baselines: "30 seconds guessing time", "30% form failure rate", "15 manual workarounds per week", "60% visual noise reduction". These numbers appear precise but lack sourcing. No user research, analytics dump, or support ticket analysis is cited. The numbers are plausible but unverified. This weakens the evidentiary chain for anyone reviewing this PRD in isolation.

### Evidence -> Success Criteria

User stories provide Given/When/Then acceptance criteria. Most criteria are testable. However, several success criteria only test the happy path. For example, Story 7 (move sub-item) covers "move to closed main item" and "move to same main item" but does not cover concurrent move, or what happens if the target main item is deleted mid-operation.

### Self-Contradiction Check

No direct contradictions found. The document is internally consistent across spec, stories, and UI functions. One mild tension: prd-spec says "当前不分页（全量返回）" under Performance Requirements, but Story 8 AC states "响应时间 <= 500ms" for 1000+5000 records. If the dataset grows, the no-pagination stance becomes a liability, but the document does acknowledge this ("若未来引入分页需重新评估穿透逻辑").

---

## Phase 2: Rubric Scoring

### Dimension 1: Background & Goals (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Background has three elements (Reason/Target/Users) | 28/30 | All three present. Reason identifies blocking bug + UX defects. Target describes two phases. Users table maps roles to affected items. Deduction: the Users table lists "PM 用户（5名）" and "Member 用户（3名）" but does not explain how these numbers were derived. The "macOS 用户（3人）" role overlaps entirely with PM/Member users and is more of a platform constraint than a user type. |
| Goals are quantified | 27/30 | Six goals with numeric targets. Strong. Deduction: the baseline numbers (30s, 30%, 15x/week, 60%) have no source attribution. "从每次 30 秒降至 0 秒" is suspiciously round. A PM shipping products that failed from ambiguous requirements would demand: where did this number come from? |
| Background and goals are logically consistent | 36/40 | Each goal traces to specific numbered items. Logical chain holds. Deduction: Goal "减少页面干扰信息" claims "减少约 60% 视觉干扰" but there is no definition of what constitutes "visual interference" or how 60% is measured. Is it item count? Screen area? Cognitive load? |

**Dimension 1 Total: 91/100**

---

### Dimension 2: Flow Diagrams (150 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid diagram exists | 50/50 | Single comprehensive Mermaid flowchart present covering four operation types. |
| Main path complete (start -> end) | 42/50 | Four flows covered: status transition, delete, move sub-item, filter penetration. Main paths all reach terminal states. Deduction: the flow for "转换表单" (items #4, #6, #7) is described in text but has NO node in the Mermaid diagram. This is a significant omission -- three of the 16 items are absent from the diagram. |
| Decision points + error branches covered | 40/50 | Diamond nodes exist for: can-transition? is-terminal? target-valid?. Error branches: "展示行内错误消息", "提示拒绝原因". Deduction: the delete flow has no error branch. What happens if the delete API fails? What if the database transaction partially rolls back? The diagram jumps from confirmation directly to "事务内级联软删除" with no failure path. |

**Dimension 2 Total: 132/150**

---

### Dimension 3: Functional Specs (200 pts) — evaluates prd-ui-functions.md

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction completeness | 58/70 | All 10 UI Functions have Placement sections with Mode, Target Page, Position. Page Composition table maps pages to UI Functions. User Interaction Flows describe sequential steps. Deduction: UF-6 (子事项移动) Placement says "子事项详情页 / 子事项编辑界面" but the interaction flow describes a "目标主事项选择器（搜索或列表）" without specifying whether this is a new modal, dropdown, or drawer. The placement of this selector within the page is ambiguous. UF-10 (团队选择器过滤) says "若仅有一个团队，可省略下拉或固定展示" -- "可" (can/may) is ambiguous. Is this a design decision left to implementation, or a requirement? |
| Data Requirements & States clarity | 60/70 | All UI Functions have Data Requirements tables with Field/Type/Source/Notes. States tables cover typical lifecycle. Deduction: UF-3 (删除) Data Requirements lists "子事项数量" but does not specify the API endpoint or how the count is fetched. Is it part of the main item detail response, or a separate query? UF-7 (过滤穿透) Data Requirements lists "匹配标识" as type "badge" and source "过滤结果" -- this is vague. What data triggers the badge? How does the frontend know a main item was included due to sub-item match? The backend response schema for this is unspecified. |
| Validation Rules explicit | 48/60 | Most UI Functions have validation rules. Deduction: UF-2 (开始时间) says "无特殊校验，日期格式由日期选择器组件保证" -- but what about business rules? Can the start date be after the end date? Can it be in the past? Can it be changed to a date that conflicts with parent item dates? UF-5 (排序) says "排序在服务端完成" -- this is an implementation detail, not a validation rule. UF-10 (团队选择器过滤) says "无特殊校验" -- but there should be a validation rule that a user cannot select a team they lack permissions for, even if the list is pre-filtered. UF-9 (甘特图修复) Validation Rules are "时间范围自动计算，不留空白" and "滚动条至少在 hover 时显示" -- these describe behavior, not validation rules. |

**Dimension 3 Total: 166/200**

---

### Dimension 4: User Stories (200 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Coverage: one story per target user | 45/50 | Background identifies 3 user types: PM (5), Member (3), macOS users (3). Story 1-5, 7-10 cover PM users. Story 6 covers Member users. Story 10 (team selector) covers "所有用户". Deduction: macOS-specific concerns are only covered in Story 10 (甘特图), but the Background identifies macOS users as a distinct affected group. The macOS scroll bar fix (#14) is folded into Story 10 rather than having its own story, which is acceptable but borderline. |
| Format correct (As a / I want / So that) | 48/50 | All 11 stories follow As a / I want / So that format. Actions are concrete ("看到具体的错误原因", "直接修改开始时间", "删除误建的主事项和子事项"). Deduction: Story 11 says "I want to 团队选择器只展示我有权限的团队" -- this describes a system behavior the user wants, not an action the user takes. It's close but reads more like a constraint than a user desire. |
| AC per story (Given/When/Then) | 42/50 | All stories have Given/When/Then format. Deduction: Story 2 (开始时间) has a single AC that is trivially testable but thin. It does not verify: what happens if the date is cleared? What if the save fails? Story 5 (排序) has a single AC. What if there are zero sub-items? Story 10 (甘特图) has two ACs but neither addresses what happens when there is only one main item (single data point time range). |
| AC verifiability & boundary coverage | 36/50 | Most ACs are objectively testable. Deduction: Story 8 AC4 says "响应时间 <= 500ms" for "1000 主事项 + 5000 子事项" -- this is a performance test, not a user-observable acceptance criterion. How is this verified in acceptance testing? Story 6 AC2 uses a SQL query as an acceptance criterion ("SELECT COUNT(*) FROM pmw_team_members WHERE role_key IS NULL THEN 返回 0") -- this is a database assertion, not a user-observable behavior. It belongs in technical design, not PRD. Story 11 AC2 (每周进展过滤) is extremely dense: "不展示处于终态且本周（自然周，周一至周日）和上周都没有活跃事项（状态变更/子事项新增编辑/进度更新）的主事项" -- this is testable but the definition of "活跃" spans three event types and two time windows. The AC should be decomposed or accompanied by examples. |

**Dimension 4 Total: 171/200**

---

### Dimension 5: Scenario Completeness (150 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| End-to-end scenario coverage | 48/60 | Status transition flow: covers error display, terminal state confirmation. Delete flow: covers confirmation, cascade, history. Move flow: covers target validation, renumbering. Filter flow: covers multi-select, penetration, badge display. Deduction: The "转换表单" scenario only covers the happy path in the flow description. What happens after a conversion succeeds? Does the item disappear from the todo list? Does the user get feedback? The flow description says "提交成功或关闭表单后，所有字段自动清空" but does not describe what the user sees after successful conversion. The "团队选择器过滤" scenario (UF-10) has a note "若仅有一个团队，可省略下拉或固定展示" but no scenario covers the single-team case end-to-end. |
| Implicit assumptions surfaced | 30/40 | The document states "当前不分页（全量返回）" which surfaces an important assumption. It also notes "若未来引入分页需重新评估穿透逻辑". Deduction: Several assumptions are not surfaced: (1) The move sub-item flow assumes the target main item list is small enough to render in a selector. What if there are hundreds of main items? (2) The delete flow assumes soft delete is sufficient for all use cases, but does not state whether deleted items are visible anywhere (admin view? audit log?). (3) The filter penetration flow assumes the backend can distinguish between "main item matched directly" and "main item matched via sub-item" in the API response, but this distinction is never specified as a backend API requirement. (4) The conversion form reset assumes "所有字段" means all fields including those pre-populated from the todo item -- but if description is disabled (grayed out), does clearing it mean the pre-populated value is lost? |
| Business-rules consistency | 42/50 | No contradictions found between business rules across documents. The permission model is consistent (PM can delete, Member cannot). The status model is consistent (terminal states handled uniformly). Deduction: The document does not define what constitutes a "终态" (terminal state). Multiple stories and UI functions reference "终态" but no story or spec defines the complete list of terminal states. Is "已关闭" the only terminal state? What about "已完成"? "已取消"? This is a critical business rule that is assumed but never stated. |

**Dimension 5 Total: 120/150**

---

### Dimension 6: Edge Case Coverage (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Error paths documented | 30/40 | Status transition: error message from backend documented. Move sub-item: closed target and same-item rejection documented. Delete: confirmation dialog documented. Deduction: (1) Delete API failure is not documented anywhere. What happens if the soft delete fails mid-transaction? The prd-spec says "级联删除须在单个事务内执行" but no error path for the user is described. (2) Filter penetration: what happens if the API call times out or returns partial data? (3) Team selector: what if the user has zero teams? The UI function mentions "单团队" and "多团队" states but not "零团队". |
| Boundary conditions covered | 22/35 | Performance boundary is stated: "1000 主事项 + 5000 子事项, <= 500ms". Empty filter state is covered: "未选择任何过滤器时展示全部事项". Deduction: (1) No coverage for empty sub-item list during move operation. (2) No coverage for maximum field lengths in forms. (3) No coverage for concurrent operations: two users trying to move the same sub-item simultaneously, or one user deleting a main item while another moves a sub-item out of it. (4) Gantt chart with zero items or a single item -- what is the time range? (5) The "因子事项匹配" badge: what if ALL main items are shown due to sub-item match? Does the badge lose meaning? |
| Failure recovery described | 12/25 | The document describes inline error messages for status transition failures. Deduction: (1) For delete failures, no recovery path is described. Does the user retry? Is there a different message? (2) For move failures, the rejection is shown but no recovery guidance. (3) For form submission failures, the form should presumably retain data for correction, but this is not stated. Story 4 says fields clear on form close, but what about failed submission -- do fields persist or clear? |

**Dimension 6 Total: 64/100**

---

### Dimension 7: Scope Clarity (100 pts)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete deliverables | 33/35 | Each in-scope item is a specific numbered feature with a clear description. Items are action-oriented: "展示错误消息", "添加开始时间字段", "倒序排列". Deduction: #10 ("状态过滤器改为多选 + 负责人过滤穿透子事项") bundles two changes into one item. These could be independently delivered, but the document does not clarify whether they must ship together. |
| Out-of-scope explicitly lists deferred items | 26/30 | Seven out-of-scope items listed: hard delete, audit UI, bulk operations, move history, drag sorting, seed data updates beyond new permission codes. Deduction: "新增权限码之外的其他 seed 数据更新（由 RBAC 提案覆盖）" references an external "RBAC 提案" but does not link to it or describe its scope. If the RBAC proposal doesn't exist yet, this is a dangling dependency. |
| Scope consistent with functional specs and user stories | 30/35 | In-scope items map to all 11 user stories and all 10 UI Functions. No extra items in stories/functions that are out of scope. Deduction: Story 11 combines two items (#15 team selector, #16 weekly progress filter) but the scope lists them separately. Story 10 (#13, #14) similarly bundles two scope items. This is not a consistency error but creates ambiguity about whether both items in a bundled story must be delivered together. |

**Dimension 7 Total: 89/100**

---

### Cross-Dimension Coherence Check

1. **Story 4 (转换表单) across dimensions:** The scope item #4 says "描述字段添加置灰样式", #6 says "关闭时清空字段", #7 says "必填校验". These are separate scope items but one story and one UI Function. The prd-spec Flow Description bundles #4, #6, #7 into one flow. This is coherent but creates a delivery risk: if one of the three is deferred, the story and UI Function must be split.

2. **"终态" across dimensions:** Stories 7, 8, 9, and UI Functions 6, 7, 8 all reference "终态" but no document defines the complete set of terminal states. This is a cross-cutting ambiguity.

3. **Performance requirement placement:** The "<= 500ms" requirement appears in Story 8 AC, prd-spec Performance Requirements, and is implicitly referenced by the Mermaid diagram node "后端内存过滤+穿透". The repeated mention is consistent, but Story 8 AC4 is testing infrastructure performance, not user behavior.

---

## Phase 3: Blindspot Hunt

### [blindspot-1] No definition of "终态" (terminal state)
Multiple items reference terminal states for sorting, filtering, and blocking operations, but the complete set of terminal status values is never defined. This is a business rule that should be explicit in the PRD. Without it, implementers may guess, and testers cannot verify correctness.

### [blindspot-2] Sub-item move: no consideration of status incompatibility
Story 7 says "状态和负责人保持不变" when moving a sub-item. But what if the sub-item's current status is not valid in the target main item's workflow? The document assumes all sub-items use the same status set, but this is never stated. If sub-items inherit workflow from their parent main item, moving could create an invalid state.

### [blindspot-3] Form reset vs. form submission failure
Story 4 AC says "用户关闭任何新增/转换表单（含取消）...所有字段为空". But what about when a form submission fails? Does the form reset (losing user input) or retain values (for correction)? The prd-spec says "提交成功或关闭表单后，所有字段自动清空" but does not address failed submission. This is a real user experience gap.

### [blindspot-4] Delete: no mention of undo or recovery path
The scope explicitly excludes hard delete and audit UI, but there is no mention of any recovery mechanism for accidental deletion. The confirmation dialog is the only safeguard. For a system used by 5 PMs managing real project data, a single confirmation dialog for cascade-deleting a main item with all its sub-items is a thin safety net. At minimum, the document should acknowledge this risk in a "Known Limitations" section.

### [blindspot-5] Weekly progress filter (#16): "活跃" definition is complex and untested
Story 11 AC2 defines "活跃事项" as items with "状态变更/子事项新增编辑/进度更新" in "本周（自然周，周一至周日）和上周". This is a compound condition across three event types and two time windows. The document does not specify: (1) where these events are tracked (status_history table?), (2) whether "编辑" includes any field change or specific fields, (3) whether "进度更新" is a specific event type or inferred from field changes.

### [blindspot-6] No consideration of zero-result states
Across all filter and search scenarios, no story or UI function addresses what the user sees when filters return zero results. An empty state message ("没有匹配的事项") is a basic UX requirement that is completely absent.

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Background & Goals | 91 | 100 |
| 2. Flow Diagrams | 132 | 150 |
| 3. Functional Specs | 166 | 200 |
| 4. User Stories | 171 | 200 |
| 5. Scenario Completeness | 120 | 150 |
| 6. Edge Case Coverage | 64 | 100 |
| 7. Scope Clarity | 89 | 100 |
| **Total** | **833** | **1000** |

---

## Detailed Attack List

1. **Edge Case Coverage**: Delete API failure has no user-facing error path — prd-spec says "级联删除须在单个事务内执行" but no story or UI function describes what happens when the transaction fails — must add a failure scenario with user-facing feedback.

2. **Scenario Completeness**: "终态" is never defined — Stories 7, 8, 9 and UI Functions 6, 7, 8 reference "终态" for sorting, filtering, and blocking operations but no document enumerates which status values are terminal — must add an explicit terminal state definition.

3. **User Stories**: Story 11 AC2 uses a SQL query as acceptance criteria — "SELECT COUNT(*) FROM pmw_team_members WHERE role_key IS NULL THEN 返回 0" — this is a database assertion, not a user-observable behavior; belongs in technical design, not PRD.

4. **Functional Specs**: UF-7 (过滤穿透) does not specify how the frontend distinguishes direct-match from sub-item-match — Data Requirements lists "匹配标识" as type "badge" and source "过滤结果" but the backend response schema that carries this distinction is unspecified — must define the API response structure for penetration results.

5. **Flow Diagrams**: Three items (#4, #6, #7 转换表单) are described in text but have no nodes in the Mermaid diagram — the diagram covers only status transition, delete, move, and filter flows — must add the conversion form flow to the diagram.

6. **Edge Case Coverage**: No zero-result empty state is described for any filter or search scenario — across 11 stories and 10 UI functions, none address what the user sees when filters return no results — must add empty state handling.

7. **Functional Specs**: UF-2 (开始时间) validation rules say "无特殊校验" — but does not address: can start date be after end date? Can it be in the past? Can it conflict with parent item dates? — must add business rule validation.

8. **Edge Case Coverage**: Concurrent operations are not addressed — two users moving the same sub-item, or one deleting a main item while another moves a sub-item from it — must document expected behavior for concurrent access.

9. **Scenario Completeness**: Form submission failure vs. form reset ambiguity — Story 4 says fields clear on form close, prd-spec says fields clear after "提交成功或关闭表单后", but failed submission is not addressed — must clarify whether fields persist or clear on submission failure.

10. **Background & Goals**: Baseline metrics lack sourcing — "30% failure rate", "15 manual operations per week", "60% visual noise" are presented as measured facts but no analytics source is cited — must add source attribution or reframe as estimates.
