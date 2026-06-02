# PRD Evaluation Report — Iteration 1

**Feature**: System UX Optimization Batch
**Evaluator**: Senior QA Engineer (Adversary)
**Date**: 2026-06-02
**Total Score**: 770/1000

---

## Phase 1: Reasoning Audit

### Argument Chain Trace

1. **Problem → Solution**: The PRD identifies 16 items (1 blocking bug + 15 UX defects). Each item maps to a specific fix or enhancement. The chain is sound — solutions directly address stated problems.

2. **Solution → Evidence**: Goals include quantified metrics (30s→0s, 30%→<5%, 15→0 manual workarounds). These are reasonable proxy measures but some metrics (40% filter coverage, 60% visual noise reduction) lack traceable baselines or measurement methods.

3. **Evidence → Success Criteria**: User stories provide Given/When/Then acceptance criteria for most items. However, several stories verify implementation (e.g., "ORDER BY id DESC") rather than user-visible behavior, and some success criteria are not independently testable without subjective judgment.

4. **Self-contradiction**: No critical contradictions found. Minor tension: Story 6 AC references specific permission codes (`item_pool:submit`, `main_item:list`) not mentioned in the In Scope list for #8, which only says "fix member role bug."

### SC Consistency Deep-Dive

| Goal | SC Coverage | Status |
|------|-------------|--------|
| Member login works | Story 6 | Satisfied |
| Reduce transition failure guess time | Story 1 | Satisfied |
| Reduce form submission failure | Story 4 | Satisfied |
| Reduce manual workarounds | Stories 2, 3, 5, 9 | Satisfied |
| Improve filter efficiency | Stories 7, 8, 9 | Partially — 40% metric not verifiable |
| Reduce visual noise | Stories 9, 11 | Satisfied |

---

## Phase 2: Rubric Scoring

### 1. Background & Goals — 82/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Background three elements | 28/30 | All three present. Reason clearly identifies blocking bug + 15 UX defects. Target specifies phased rollout. Users table covers PM, Member, macOS users with affected items. Deduction: Users table gives "Member 用户（3名）" and "PM 用户（5名）" — these are specific counts, but there is no source for these numbers (are these current team sizes? projected?). Minor ambiguity. |
| Goals quantified | 27/30 | Six goals, all with metrics: time (30s→0s), percentage (30%→<5%), count (15→0), coverage (40%), reduction (60%). Deduction: "覆盖 40% 的过滤需求" and "减少约 60% 视觉干扰" are stated as metrics but have no defined measurement method. How does one verify "40% of filter needs" or "60% visual noise"? These are estimates, not testable targets. |
| Logical consistency | 27/40 | Goals broadly follow from the stated problems. **Deduction 1 (-8)**: Goal "减少手动绕路操作" claims "#3, #5, #6, #9" as sources. Item #3 is delete, #5 is sort, #6 is form clear, #9 is move. The claim that sorting or form clearing eliminates "绕路操作" (workaround operations) is asserted without evidence — why would users currently perform workarounds for these? The "每周约 15 次" baseline is not sourced. **Deduction 2 (-5)**: The metric "从每次 30 秒降至 0 秒" for status transition is odd — it implies zero time to understand the error, but the user still needs to read the message. "0 seconds" is aspirational, not measurable. |

---

### 2. Flow Diagrams — 120/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid diagram exists | 50/50 | A single comprehensive Mermaid flowchart exists covering status transition, delete, move sub-item, and filter operations. |
| Main path complete | 35/50 | The diagram covers happy paths for all four major flows. **Deduction (-10)**: The filter flow ends at "展示结果" without showing what happens next (e.g., user can clear filter, modify filter). The diagram treats filtering as a one-shot action. **Deduction (-5)**: The status transition flow shows "执行流转" but does not show the post-success state update or UI refresh. |
| Decision points + error branches | 35/50 | Decision diamonds exist for "可流转?", "是否终态?", "目标有效?", "因子事项匹配?". **Deduction (-10)**: Missing error branches for: (a) delete operation — no error branch shown (what if delete fails?); (b) concurrent access during move — the move flow shows only happy path and "目标已关闭/同一主事项" rejection. No branch for server error during move execution. (c) filter flow — no branch for empty results or error responses. **Deduction (-5)**: The status transition error branch ends at "展示行内错误消息" with no recovery path shown (what does the user do next?). |

---

### 3. Functional Specs (prd-ui-functions.md) — 155/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction completeness | 55/70 | All 10 UI Functions have Placement (mode, target page, position). Page Composition table maps functions to pages. **Deduction (-10)**: UF-2 (子事项编辑弹窗开始时间) placement says "现有字段之间" — which specific field position? Before or after which existing field? This is ambiguous for implementation. **Deduction (-5)**: UF-6 (子事项移动) placement lists both "子事项详情页 / 子事项编辑界面" but the User Interaction Flow only describes one flow — it is unclear if the interaction differs between these two contexts. |
| Data Requirements & States clarity | 55/70 | Field tables and state tables exist for all UI Functions. Sources are identified. **Deduction (-10)**: UF-6 Data Requirements table lists "目标主事项" source as "主事项列表 — 排除已关闭和当前主事项" but does not specify the API endpoint or query parameters for fetching the filtered list. For a downstream implementer, this is insufficient. **Deduction (-5)**: UF-7 (过滤穿透) Data Requirements lists "匹配标识" as type "badge" with source "过滤结果" — this is a frontend-computed value, not a data source. The table entry adds no implementation value. |
| Validation Rules explicit | 45/60 | Each UI Function has validation rules. **Deduction (-10)**: UF-2 validation says "无特殊校验，日期格式由日期选择器组件保证" — this is a cop-out. Should start date have any relationship to end date? Can it be in the past? In the future? After the end date? No validation rules are specified. **Deduction (-5)**: UF-10 (团队选择器) validation says "无特殊校验" — but what if the user has zero teams? What happens then? This is an unhandled edge case. |

---

### 4. User Stories — 145/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Coverage: one story per target user | 40/50 | PM users: Stories 1-5, 7-10 (covered). Member users: Story 6 (covered). macOS users: Story 10 covers macOS scrollbar (covered). **Deduction (-10)**: There is no story for the "所有用户" persona introduced in Story 11 ("As a 所有用户"). The Background only defines PM, Member, and macOS user roles. "所有用户" is a new user type introduced without definition in the Background. |
| Format correct | 45/50 | All 11 stories follow As a / I want / So that format. **Deduction (-5)**: Story 5's "I want" says "按最新创建的排在前面" — this is an implementation detail (sort by creation time), not a user need. The real need might be "quickly find recently added sub-items." Similarly, Story 6's AC references SQL query `SELECT COUNT(*) FROM pmw_team_members WHERE role_key IS NULL` — this is a database verification step, not a user story acceptance criterion. |
| AC per story (Given/When/Then) | 30/50 | All stories have at least one AC in Given/When/Then. **Deduction (-20)**: Multiple stories have incomplete AC coverage: (a) Story 1 — no AC for what happens when the error is resolved (does the message disappear?). (b) Story 2 — no AC for what happens when saving fails. (c) Story 4 — AC says "提交按钮禁用" but no AC for the error message shown when attempting to submit (or if the button is simply disabled, no feedback is given about WHY it's disabled — only that labels show *). (d) Story 10 — AC for macOS scrollbar says "hover 容器或常态下" but the Then only says "底部水平滚动条可见" without specifying whether it's always-visible or hover-triggered. (e) Story 11 — the second AC about weekly progress page filtering is extremely long and complex (covering terminal state + activity detection for this week and last week), making it difficult to verify. |
| AC verifiability & boundary coverage | 30/50 | **Deduction (-10)**: Story 5's Then says "子事项按 id 倒序排列（最新在前）" — this verifies implementation (id ordering), not behavior. What if two sub-items have the same creation time? What if ids are non-sequential? The user-visible behavior is "most recent first," which should be verified by creation timestamp, not id. **Deduction (-10)**: Story 8 (过滤穿透) AC includes a performance requirement: "数据量为 1000 主事项 + 5000 子事项时响应时间 ≤ 500ms". This is a non-functional requirement embedded in a user story AC. It cannot be verified through a Given/When/Then interaction — it requires a load test. This belongs in Performance Requirements, not in a user story. Story 7 has the same issue — "因子事项匹配" visual indicator is stated but not specified (color? icon? text?). |

---

### 5. Scenario Completeness — 100/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| End-to-end scenario coverage | 40/60 | Flow descriptions cover happy paths for status transition, delete, form conversion, move, and filter. **Deduction (-10)**: No end-to-end scenario for the member permission fix (#8). This is listed as a blocking bug but the Flow Description section does not describe the member login flow or what the fixed behavior looks like. **Deduction (-10)**: No scenario for team selector filtering (#15). The UI function describes it, but there's no business flow description for what happens when a user has zero teams, one team, or multiple teams. |
| Implicit assumptions surfaced | 25/40 | **Deduction (-10)**: Assumption: "后端已支持" (UF-2 Data Requirements) — the start date field is assumed to be already supported by the backend API (`SubItemUpdateReq.StartDate`). If this assumption is wrong, the entire story is blocked. This dependency is not flagged. **Deduction (-5)**: Assumption: "NextSubCode" for sub-item renumbering during move is assumed to exist and be thread-safe. No mention of concurrent move operations or race conditions on the same target main item. **Deduction**: No explicit pre-conditions or environmental dependencies section. |
| Business-rules consistency | 35/50 | Scenarios generally respect domain rules. **Deduction (-10)**: The delete flow says "在 status_history 表中记录删除事件" but no scenario describes what the status_history entry looks like (what status value? "deleted"? what from_status/to_status?). This is a data integrity concern. **Deduction (-5)**: The filter penetration logic says "未选择任何过滤器时展示全部事项" but the progress page says "首次加载'进行中'复选框默认选中" — these are two different default behaviors for similar-looking filter states, which could confuse users. |

---

### 6. Edge Case Coverage — 80/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Error paths documented | 30/40 | Error paths exist for: status transition failure (inline Alert), move to closed target (rejection), move to same target (rejection), form validation (disabled submit). **Deduction (-5)**: No error path for delete failure — what if the soft delete transaction fails? What does the user see? **Deduction (-5)**: No error path for network failures during any operation. The document assumes all API calls succeed or fail with business logic errors, not transport errors. |
| Boundary conditions covered | 28/35 | Boundary: Performance threshold documented (500ms at 1000+5000 scale). **Deduction (-4)**: No boundary for: (a) main item with zero sub-items — does the delete confirmation still show "将同时删除 0 个子事项"? (b) sub-item move when target has maximum sub-items — is there a limit? (c) filter with zero results — what does the user see? (d) form clear — does it clear validation errors too, or just field values? **Deduction (-3)**: Gantt chart edge case: what if all visible items have the same start/end date? The "前1天/后1天" rule would create a 3-day range for what might be a single-day item — is this the intended behavior? |
| Failure recovery described | 22/25 | Most operations have confirmation dialogs (delete, move, terminal state transition). **Deduction (-3)**: The status transition error message has no described recovery — after seeing the error, what does the user do? The flow just ends at "展示行内错误消息." There's no "dismiss" or "retry" action described. |

---

### 7. Scope Clarity — 88/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 30/35 | All 16 items are specific features or fixes with clear descriptions. Each is a checkbox item. **Deduction (-5)**: Item #8 "修复 member 角色用户登录后获取不到权限的 bug" is specific as a bug but does not specify what the fix looks like. The "Related Changes" table mentions "修复 RoleKey 为 nil 时权限查询返回空集" but this is implementation-level — the scope item itself is vague about expected behavior. |
| Out-of-scope explicitly lists | 28/30 | Seven items explicitly listed as out of scope. **Deduction (-2)**: "独立的删除审计 UI 或通知推送" is listed as out of scope, but status_history deletion recording IS in scope. The boundary between "recording in status_history" and "audit UI" is reasonable but could be clearer — what if someone needs to query deletion history? |
| Scope consistent with specs and stories | 30/35 | In-scope items generally match user stories and UI functions. **Deduction (-5)**: Scope item #6 says "所有新增/转换表单关闭时清空字段" but Story 4 AC says "再次打开同一表单" — the scope says "所有" but the AC only covers re-opening. What about first-open after a different user's session? What about browser refresh? The scope is broader than the AC. |

---

## Cross-Dimension Coherence Check

1. **Scope ↔ Stories**: 16 scope items vs. 11 stories. Items #4 (description disabled) and #6 (form clear) are combined into Story 4. Items #11 and #12 are combined into Story 9. Item #13 and #14 are combined into Story 10. Item #15 and #16 are combined into Story 11. The mapping is reasonable but not explicitly documented — a traceability matrix is missing.

2. **UI Functions ↔ Stories**: 10 UI Functions vs. 11 stories. UF-8 covers both Stories 9 and 12 (scope items #11 and #12). UF-4 covers scope items #4, #6, #7. Some UI functions cover multiple scope items but this is not always clear.

3. **Goals ↔ Metrics**: The "40% filter coverage" goal has no corresponding metric in any story or UI function. It is untraceable.

---

## Phase 3: Blindspot Hunt

**[blindspot-1] No error/dismiss behavior for inline Alert messages.** prd-spec says "前端展示行内错误消息（Alert 组件）" and prd-ui-functions.md States table shows "错误 | Alert 行内消息 | 后端返回流转失败" but nowhere does it specify: (a) when does the Alert dismiss? (b) can the user manually dismiss it? (c) does it persist across re-attempts? This is a critical UX gap — a stale error message could confuse users into thinking their subsequent attempt also failed.

**[blindspot-2] Move sub-item: no story for the TARGET main item's perspective.** When a sub-item moves to a new main item, the target main item gains a sub-item. Is there any notification? Does the target's sub-item count update in real-time? Does the target main item's PM get any indication? The PRD treats the move as purely affecting the source sub-item, ignoring the target's perspective.

**[blindspot-3] Delete: no story for restoring/recovering deleted items.** Soft delete is explicitly chosen (scope says "硬删除" is out of scope), but no mechanism is described for viewing or recovering soft-deleted items. If a PM accidentally deletes a main item, there is no documented recovery path. The PRD says "独立的删除审计 UI" is out of scope, but recovery is different from audit.

**[blindspot-4] Filter penetration: unclear behavior when both status and assignee filters are active.** prd-ui-functions.md UF-7 validation says "状态和负责人过滤器可独立使用或组合使用" but the interaction flow only describes them independently. When combined, does an item need to match BOTH filters? Or either? Does the penetration apply only to the assignee dimension while status applies to main items only? The logic is underspecified.

**[blindspot-5] Weekly progress page filter (#16) has no dedicated story AC for edge cases.** The filter logic "终态且本周/上周都没有活跃事项" is complex — it requires checking activity across two time windows for two categories (status change, sub-item add/edit, progress update). Story 11's AC includes this but it's a single monolithic Then clause. What counts as "活跃"? Is it any status_history entry in the time window? Any sub-item created_at or updated_at? Any progress update? The definition of "active" is critical and unstated.

**[blindspot-6] No accessibility requirements.** For a batch of UX improvements, no mention of keyboard navigation, screen reader support, or ARIA attributes for new interactive elements (inline alerts, multi-select checkboxes, target selectors).

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 82 | 100 |
| Flow Diagrams | 120 | 150 |
| Functional Specs | 155 | 200 |
| User Stories | 145 | 200 |
| Scenario Completeness | 100 | 150 |
| Edge Case Coverage | 80 | 100 |
| Scope Clarity | 88 | 100 |
| **Total** | **770** | **1000** |
