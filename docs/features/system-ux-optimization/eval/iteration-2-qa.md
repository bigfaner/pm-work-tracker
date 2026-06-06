# QA Eval Iteration 2 — PRD

**Evaluator**: QA (Senior QA Engineer, adversary role)
**Target**: 900/1000
**Result**: 915/1000

## Iteration 1 Resolution Summary

Of 20 iteration-1 attack points, **18 are resolved**, **2 are partially resolved**:

| # | Attack | Status | Evidence |
|---|--------|--------|----------|
| 1 | Baseline metrics lack sourcing | Resolved | All estimates now tagged "(PM 团队反馈估算)" |
| 2 | Delete flow no error branch | Resolved | Mermaid has J1→J2→J3 recovery path |
| 3 | Status error no recovery path | Resolved | Mermaid has E→E1 with retry/abort |
| 4 | Conversion form missing from diagram | Resolved | CF1–CF9 nodes added |
| 5 | "终态" undefined | Resolved | Explicit 终态定义 section added |
| 6 | Form failure vs reset ambiguity | Resolved | prd-spec line 104-105 + UF-4 Validation Rules clarified |
| 7 | No member permission E2E scenario | Resolved | 4-step verification flow added |
| 8 | SQL query as AC | Resolved | Rewritten as behavioral assertion |
| 9 | Story 5 verifies implementation | Partial | Story AC now says "创建时间" but UF-5 still says "ORDER BY id DESC" — cross-doc mismatch |
| 10 | NFR in Story 8 AC | Resolved | Moved to Performance Requirements |
| 11 | UF-7 API response unspecified | Resolved | matchType + matchedSubItemIds defined |
| 12 | UF-2 zero validation rules | Resolved | 4 validation rules added |
| 13 | UF-6 API unspecified | Resolved | Full API spec with method/body/response codes |
| 14 | No zero-result empty state | Resolved | Empty State Handling section added |
| 15 | Concurrent ops not addressed | Resolved | 3 concurrency scenarios with optimistic locking |
| 16 | Scope #6 vs Story 4 misalignment | Resolved | Story 4 now says "任意" not "同一" |
| 17 | Alert dismiss unspecified | Resolved | Alert 生命周期 section with 3 dismiss triggers |
| 18 | Combined filter logic | Resolved | AND logic with 3 sub-rules specified |
| 19 | "活跃事项" undefined | Resolved | 3-condition definition with db field references |
| 20 | No deleted-item recovery path | Resolved | Explicitly listed in Out of Scope with explanation |

## Dimension Scores

### 1. Background & Goals: 91/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Three elements present | 28/30 | Reason/Target/Users all present. Users table has 3 roles with counts and affected items. Minor: macOS users (3) are a subset of PM/Member — overlap is noted but not explicitly acknowledged. |
| Goals quantified | 27/30 | All 6 goals have numeric targets. Estimates now attributed as "(PM 团队反馈估算)". Deduction: "预估减少约 60% 视觉干扰" — "视觉干扰" has no objective measurement method; "预估覆盖 40% 的过滤需求" — 40% of what total universe is undefined. |
| Logical consistency | 36/40 | Goals follow from stated problems. Deduction: "40% filter coverage" figure lacks a defined denominator — it's unclear what constitutes 100% of filter needs. The reasoning chain is otherwise solid. |

### 2. Flow Diagrams: 142/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Mermaid diagram exists | 48/50 | Comprehensive single flowchart covering 5 operation types. |
| Main path complete | 48/50 | Status transition, conversion form, delete, move, and filter flows all trace start→end. |
| Decision points + error branches | 46/50 | 8 diamond nodes, 3 error branches with recovery paths. Deduction: the filter flow (R→S→T→W) has no error branch — what happens if the backend filter query fails or times out? |

### 3. Functional Specs: 182/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Placement & Interaction | 65/70 | All 10 UFs have complete Placement and Interaction Flow. Deduction: UF-10 step 3 uses "可省略下拉或固定展示" — "可" (may) is ambiguous for a downstream agent; is this required or optional? |
| Data Requirements & States | 65/70 | Field tables and state tables are complete. UF-6 and UF-7 now have explicit API specs. Deduction: UF-3 (delete) lacks API spec despite being a new feature with new permission codes — inconsistent with UF-6/UF-7 treatment. UF-10 also lacks API/response structure. |
| Validation Rules | 52/60 | UF-1 (2), UF-2 (4), UF-3 (2), UF-4 (4), UF-5 (1), UF-6 (2), UF-7 (3), UF-8 (2), UF-9 (2), UF-10 ("无特殊校验"). UF-5 "排序在服务端完成" is an implementation note, not a validation rule. UF-10 has no validation despite needing to handle the case where a user has zero teams. |

### 4. User Stories: 182/200

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Coverage per target user | 45/50 | PM: Stories 1-5, 7-10. Member: Story 6. macOS: Story 10 (Gantt). Deduction: macOS users are a distinct role in Background but have no dedicated story — Story 10 covers them as a PM sub-concern, not as a standalone persona. |
| Format correct | 48/50 | All 11 stories follow As a/I want/So that with concrete actions. Minor: Story 6 "我权限范围内的菜单和功能" is vague until clarified by AC. |
| AC per story (Given/When/Then) | 47/50 | All stories have Given/When/Then ACs. Deduction: Story 2 (start time) and Story 5 (sorting) have only 1 AC each — no error or edge case coverage at the story level. |
| AC verifiability & boundary | 42/50 | Most ACs are objectively testable. Deductions: Story 1 has no AC for successful transition or terminal-state confirmation dialog. Story 3 has no AC for delete transaction failure (the flow covers it, but the story doesn't). Story 4 AC3 "再次打开任意新增/转换表单" — "任意" could mean cross-form-type clearing which may not be the intended behavior. |

### 5. Scenario Completeness: 135/150

| Criterion | Score | Justification |
|-----------|-------|---------------|
| E2E scenario coverage | 55/60 | 6 flow descriptions cover the major features. Deduction: Gantt chart fix (#13/#14) and team selector (#15) have no flow description in prd-spec — they exist only as UI functions. |
| Implicit assumptions | 35/40 | Terminal states, "活跃事项", concurrency, and empty states are now explicit. Deduction: "NextSubCode" renumbering assumes the counter never overflows — no boundary addressed. UF-5 assumes id ordering equals creation time ordering — unproven if snowflake IDs are used internally. |
| Business-rules consistency | 45/50 | Most rules are consistent across documents. Deduction: UF-5 says "按 id 倒序排列 (ORDER BY id DESC)" but Story 5 says "按创建时间倒序排列" — these are semantically different fields. Per project memory, the system uses snowflake IDs; internal auto-increment id may correlate with creation time for single-insert scenarios but this is an implementation assumption baked into the PRD. |

### 6. Edge Case Coverage: 88/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Error paths | 36/40 | Status transition error, delete failure, form submission failure, move validation all covered with user-facing feedback and recovery. Deduction: No error path for filter query failure/timeout, or team selector API failure. |
| Boundary conditions | 30/35 | Empty state, concurrent access (3 scenarios), past dates, closed-target blocking all covered. Deduction: No boundary for user with zero teams (team selector empty state). No boundary for NextSubCode overflow. |
| Failure recovery | 22/25 | Delete retry/abort, form field retention on failure, optimistic lock with refresh guidance. Deduction: Filter query failure has no recovery path — user sees what? |

### 7. Scope Clarity: 95/100

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope concrete | 34/35 | 16 specific items with checkbox format. Clear deliverables. |
| Out-of-scope explicit | 29/30 | 7 deferred items named: hard delete, restore, audit UI, batch ops, move history, drag sort, other seed data. Explicit explanation for restore exclusion. |
| Scope consistency | 32/35 | Scope items align with UI functions and stories. Deduction: Scope #5 "倒序排列" matches UF-5 "ORDER BY id DESC" but contradicts Story 5 "按创建时间倒序" — the ordering criterion is inconsistent. |

## Score Summary

| Dimension | Score |
|-----------|-------|
| Background & Goals | 91/100 |
| Flow Diagrams | 142/150 |
| Functional Specs | 182/200 |
| User Stories | 182/200 |
| Scenario Completeness | 135/150 |
| Edge Case Coverage | 88/100 |
| Scope Clarity | 95/100 |
| **Total** | **915/1000** |

## Attack Points for Iteration 3

1. [Scenario Completeness / Scope Clarity]: UF-5 ordering criterion contradicts Story 5 — UF-5 says "按 id 倒序排列 (ORDER BY id DESC)" but Story 5 AC says "按创建时间倒序排列（最新创建的显示在列表最前）". Per project memory, bizKey uses snowflake strings; internal auto-increment id may not guarantee creation-time ordering in all scenarios. Must align both documents to use the same ordering criterion — preferably "创建时间" to match user-facing behavior.

2. [Functional Specs]: UF-3 (delete) lacks API specification — UF-6 and UF-7 now have explicit API specs (method, body, response codes) but UF-3, a new feature with new permission codes and cascade logic, has none. Must add delete API spec with method, request/response, and error codes for consistency.

3. [Functional Specs]: UF-10 step 3 uses ambiguous language — "若仅有一个团队，可省略下拉或固定展示" — "可" (may) is not actionable for a downstream agent. Must change to a definitive requirement: either "省略下拉并固定展示团队名" or "保持下拉样式不变".

4. [User Stories]: Story 1 (status transition error) has no AC for success case or terminal-state confirmation dialog — the flow description mentions both "执行流转" (success) and "确认对话框" (terminal state), but the story's only AC covers the error scenario. Must add ACs for successful transition and terminal-state confirmation.

5. [User Stories]: Story 3 (delete) has no AC for delete transaction failure — prd-spec flow describes J1→J2 (error)→J3 (retry/abort) but the story has no AC covering what the user sees when the delete transaction fails. Must add an AC for transaction failure.

6. [Edge Case Coverage]: Filter query failure has no error path or recovery — the Mermaid diagram's filter flow (R→S→T→W) shows only success paths. Must document what the user sees if the backend filter query times out or fails, and what recovery actions are available.

7. [Edge Case Coverage]: Zero-team empty state for team selector — UF-10 covers single-team and multi-team states but never addresses what happens if a user has zero teams (e.g., after being removed from all teams). Must define the empty state display.

8. [Scenario Completeness]: NextSubCode overflow not addressed — the move sub-item flow says "目标父项的 NextSubCode 获得新编号" but no boundary condition or error handling exists for counter exhaustion. Must define behavior if NextSubCode reaches its maximum value.

9. [Scenario Completeness]: Empty state section claims applicability to "每周进展页面" but that page has no user-facing filters — prd-ui-functions.md states "每周进展页面的后端过滤（无 UI 变更）" and the UF scope says "无 UI 变更". The "清除过滤条件" button cannot exist on a page with no filter UI. Must either remove weekly progress from empty-state scope or explain how backend filtering interacts with the empty state.

10. [blindspot]: Story 4 AC3 "再次打开任意新增/转换表单" — "任意" implies cross-form-type clearing. If a user opens a todo→subitem form, fills it, closes it, then opens a todo→main-item form, does the second form also reset? The scope item "所有新增/转换表单关闭时清空字段" supports this interpretation, but "任意" could also mean "any form the user subsequently opens" which is ambiguous about whether form state is shared or per-form-type. Must clarify whether clearing is per-form-instance, per-form-type, or global.
