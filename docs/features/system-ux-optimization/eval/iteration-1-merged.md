# Eval Iteration 1 — Merged Report

**PM Score**: 833/1000
**QA Score**: 770/1000
**Average**: 802/1000

## Merged Attack Points

1. [Background & Goals]: Baseline metrics lack sourcing and are untestable — "30% failure rate", "每周约 15 次" manual operations, "40% filter coverage", and "60% visual noise reduction" are presented as measured facts but no analytics source is cited and no measurement method is defined — must add source attribution, reframe as estimates, or replace with verifiable targets that include measurement methods.
2. [Flow Diagrams]: Delete flow has no error branch — the mermaid diagram shows only the happy path with no failure branch for transaction errors, and the prd-spec says "单个事务中软删除" but no scenario covers what the user sees if the transaction fails — must add a failure branch to the diagram with user-facing feedback and recovery path.
3. [Flow Diagrams]: Status transition error branch has no recovery path — diagram ends at "展示行内错误消息" with no subsequent action for the user — must extend the diagram to show what the user can do after seeing the error.
4. [Flow Diagrams]: Three items (#4, #6, #7 转换表单) are described in text but have no nodes in the Mermaid diagram — the diagram covers only status transition, delete, move, and filter flows — must add the conversion form flow to the diagram.
5. [Scenario Completeness]: "终态" is never defined — Stories 7, 8, 9 and UI Functions 6, 7, 8 reference "终态" for sorting, filtering, and blocking operations but no document enumerates which status values are terminal — must add an explicit terminal state definition.
6. [Scenario Completeness]: Form submission failure vs. form reset ambiguity — Story 4 says fields clear on form close, prd-spec says fields clear after "提交成功或关闭表单后", but failed submission is not addressed — must clarify whether fields persist or clear on submission failure.
7. [Scenario Completeness]: No end-to-end scenario for member permission fix (#8) — the blocking bug has no flow description showing the fixed member login experience — must add a complete user flow for the permission fix.
8. [User Stories]: Story 6 AC2 uses a SQL query as acceptance criteria — "SELECT COUNT(*) FROM pmw_team_members WHERE role_key IS NULL THEN 返回 0" — this is a database assertion, not a user-observable behavior; belongs in technical design, not PRD.
9. [User Stories]: Story 5 verifies implementation not behavior — "子事项按 id 倒序排列（最新在前）" — should verify user-visible ordering by creation time, not database id.
10. [User Stories]: Story 8 embeds non-functional performance requirement in AC — "数据量为 1000 主事项 + 5000 子事项时响应时间 ≤ 500ms" cannot be verified via Given/When/Then interaction; belongs in Performance Requirements.
11. [Functional Specs]: UF-7 (过滤穿透) does not specify how the frontend distinguishes direct-match from sub-item-match — Data Requirements lists "匹配标识" as type "badge" and source "过滤结果" but the backend response schema that carries this distinction is unspecified — must define the API response structure for penetration results.
12. [Functional Specs]: UF-2 (子事项开始时间) has zero validation rules — "无特殊校验" does not address whether start date can be after end date, in the past, or conflict with parent item dates — must specify all date constraint business rules.
13. [Functional Specs]: UF-6 (子事项移动) Data Requirements lacks API specification — "目标主事项" source is "主事项列表" with no endpoint or query parameters defined — must define the API contract for the move operation.
14. [Edge Case Coverage]: No zero-result empty state is described for any filter or search scenario — across 11 stories and 10 UI functions, none address what the user sees when filters return no results — must add empty state handling.
15. [Edge Case Coverage]: Concurrent operations are not addressed — two users moving the same sub-item, "NextSubCode" renumbering during move has no concurrency protection, or one deleting a main item while another moves a sub-item from it — must document expected behavior for concurrent access.
16. [Scope Clarity]: Scope item #6 "所有新增/转换表单关闭时清空字段" is broader than Story 4 AC "再次打开同一表单" — scope says all forms but AC only covers re-opening — must align scope statement with story acceptance criteria or expand the AC.
17. [blindspot]: Alert dismiss behavior unspecified — prd-ui-functions.md shows error state as "Alert 行内消息" but never specifies when/how it dismisses or if user can manually close it — must define alert lifecycle and user dismissal mechanism.
18. [blindspot]: Filter penetration logic underspecified for combined filters — UF-7 says "状态和负责人过滤器可独立使用或组合使用" but does not define whether combined mode requires AND or OR matching, or how penetration interacts with status filter — must specify the boolean logic for combined filter mode.
19. [blindspot]: "活跃事项" definition for weekly progress page (#16) is critical and undefined — Story 11 references "状态变更/子事项新增编辑/进度更新" but does not specify which database fields or timestamps constitute "active" — must provide a precise, testable definition.
20. [blindspot]: No recovery path for accidentally deleted items — soft delete is chosen but no mechanism exists for PMs to view or recover deleted items — must define whether a restore flow is in scope and, if not, explicitly document the exclusion.
