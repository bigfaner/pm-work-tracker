# Proposal Evaluation Report — AI Copilot 对话助手

**Iteration**: 1
**Date**: 2026-06-11
**Document**: `docs/proposals/ai-copilot/proposal.md`

---

## Phase 1: Reasoning Audit

### Problem -> Solution Trace

| Problem Statement | Solution Element | Trace Quality |
|---|---|---|
| 创建 MainItem 需要手动填写 10+ 个字段 | AI 解析意图后推送预填表单卡片 | Direct mapping. Solved. |
| 用户需要理解状态机规则和权限体系 | AI 自动校验状态机合法性 | Partially addressed. Only covers status transitions, not RBAC understanding. |
| 非技术人员填写结构化字段困难 | 对话输入降低门槛 | Direct mapping. Solved. |

**Verdict**: Solution directly addresses stated problems. The RBAC-understanding problem is not fully solved — the proposal relies on the backend rejecting unauthorized operations, but does not reduce the user's need to understand permissions.

### Solution -> Evidence Trace

- "系统已有复杂的状态机和 RBAC 权限体系" — evidence is architectural assertion, not user pain data.
- "`todos.txt` 第 39 条" — references a roadmap item, validates that AI integration is planned, but does not prove current user pain.
- No user interviews, support tickets, or usage analytics are cited.

**Verdict**: Evidence is entirely introspective (code structure, roadmap). No external validation.

### Evidence -> Success Criteria Trace

- SC mentions field accuracy >= 80% — maps to the "10+ fields" problem.
- SC mentions latency P95 < 5s — no evidence in the Problem section justifies this specific target.
- SC covers MainItem, SubItem, Milestone, but In Scope lists 6 entities. Coverage gap for MilestoneMap, ProgressRecord, ItemPool in SC.

### Self-Contradiction Check

- No direct contradictions found. The `consistency_check_result` comment block confirms prior reconciliation.
- One tension: "对话驱动的卡片编辑" claims card is single source of truth, but Risk #3 describes "卡片 + 对话编辑状态同步复杂" — the risk acknowledges the complexity the innovation claims to solve. This is honest but worth noting.

### SC Consistency Deep-Dive

**Cluster 1: Entity CRUD (SC 1-6, InScope item 3-4)**
- SC covers MainItem, SubItem, Milestone creation, MainItem query, MainItem status change, ProgressRecord creation.
- InScope lists 6 entities: MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool.
- Missing from SC: MilestoneMap CRUD, ItemPool CRUD, any modification of SubItem/Milestone beyond status, any delete operation.
- Satisfiable as a set: Yes. Coverage incomplete but no contradictions.

**Cluster 2: Card/Chat Interaction (SC 7, InScope item 5-6)**
- SC 7: Card supports both dialog and direct edit, no data conflict.
- InScope: "对话输入和直接编辑均写入同一卡片状态".
- Satisfiable: Yes, aligned.

**Cluster 3: Infrastructure (SC 8-12, InScope item 1-2, 7-9)**
- SC 8: Bubble visible on all pages, no blocking. InScope: "全局浮动气泡".
- SC 9: Team context displayed. InScope: "Team 上下文自动检测".
- SC 10: All writes require confirmation. InScope: "所有写操作经卡片确认后执行".
- SC 11: Latency < 5s. NFR states same target. Aligned.
- SC 12: Permission error. InScope: "复用权限检查".
- Satisfiable: Yes.

**Result**: 2 coverage gaps (MilestoneMap/ItemPool in SC), no logical contradictions.

---

## Phase 2: Rubric Scoring

### 1. Problem Definition — 72/110

**Problem stated clearly (30/40)**:
The core problem (form-based interaction is inefficient, 10+ fields per MainItem) is concrete and unambiguous. One reader could interpret "操作效率低" as either "time-consuming" or "error-prone" — the proposal leans toward the former but does not cleanly separate the two. The RBAC/statemachine problem is stated but conflated with the form-filling problem.

Quote: *"创建一个 MainItem 需要手动填写 10+ 个字段，操作效率低"* — clear and specific.
Quote: *"用户需要理解状态机规则和权限体系才能正确操作，门槛较高"* — stated but not broken down into specific failure modes.

**Evidence provided (18/40)**:
Evidence is entirely internal/structural: form field counts, a roadmap reference (`todos.txt`), and architectural assertions. No user feedback, no support tickets, no usage analytics, no competitive analysis of user drop-off rates. The `todos.txt` reference is the closest to external validation but only proves planning intent, not user pain.

Quote: *"MainItem 创建表单包含 title、description、priority、assignee、planStartDate、expectedEndDate、milestoneKey 等必填/选填字段"* — this is code inspection, not user evidence.

**Urgency justified (24/30)**:
The urgency argument is logically sound ("越早引入，用户越早受益") but relies on a scaling assumption ("随着团队和事项数量增长") without quantification. No data on current team size, current creation volume, or time-per-creation to justify the cost of delay.

Quote: *"延迟实施意味着团队持续承担低效操作的时间成本"* — directionally correct but unquantified.

---

### 2. Solution Clarity — 95/120

**Approach is concrete (36/40)**:
The 5-step interaction flow is specific. A reader can explain back: "floating bubble -> chat panel -> natural language -> AI pushes pre-filled card -> user confirms." The card-as-single-source-of-truth model is well-defined.

Quote: *"AI 解析意图后推送预填表单卡片到聊天窗口。必填且无法推导的字段留空，用户可通过直接编辑卡片或继续对话补充字段"* — very clear.

**User-facing behavior described (40/45)**:
Each scenario type (create, query, modify, assign) has concrete examples with input/output pairs. Edge cases cover ambiguity, permission denial, statemachine violations. The query results behavior ("摘要文字 + 可点击卡片") is well-specified.

Minor gap: No description of what the floating bubble looks like in collapsed state, how it indicates unread messages, or how it behaves on mobile.

**Technical direction clear (19/35)**:
The proposal states "AI 服务调用经后端代理" and mentions structured output / tool use. But critical architectural decisions are deferred:

Quote: *"AI 服务选型建议在 tech-design 阶段确定具体供应商"* — acceptable for a proposal, but the prompt construction strategy ("动态组装当前 Team 的实体 schema") is mentioned without detail on prompt size limits, context window management, or how schema changes are reflected in prompts.

The front-end architecture for the card component's dual-input (chat + direct edit) is the most technically novel part but receives only one sentence of explanation.

---

### 3. Industry Benchmarking — 82/120

**Industry solutions referenced (32/40)**:
Six references: Linear, Notion AI, Slack Bot, Jira Automation, GitHub Copilot, Microsoft 365 Copilot. Good breadth across embedded AI, command palette, and bot patterns. Each is briefly described.

Quote: *"Linear：提供 Command Palette（⌘K）快速操作，部分支持自然语言搜索"* — concrete.

**At least 3 meaningful alternatives (22/30)**:
Four alternatives are presented: Do nothing, embedded form assistant, Slack Bot, conversational + card hybrid. "Do nothing" is present. However, two issues:

1. The "embedded form assistant" is marked "Partial" rather than treated as a full alternative — it is described as a subset of the selected approach rather than evaluated independently.
2. Missing: A "command palette" approach (like Linear's ⌘K) is mentioned in industry solutions but not presented as a standalone alternative. This is a meaningful option that sits between "do nothing" and "full conversational AI."

Quote: *"Partial: 适合创建场景，但查询/修改仍需独立入口"* — this evaluates the alternative against the selected solution's scope rather than on its own merits.

**Honest trade-off comparison (18/25)**:
The comparison table has pros/cons for each alternative. However, the cons of the selected approach ("AI 准确性依赖外部服务，开发量较大") are understated relative to the pros. The "开发量较大" con has no estimate to compare against other alternatives.

Quote: *"| **对话 + 卡片混合模式** | ... | 自然语言低门槛，卡片保证结构化准确性，对话 + 直接编辑灵活 | AI 准确性依赖外部服务，开发量较大 |"* — "开发量较大" is vague. Per rubric: vague language without quantification = -20 per instance, but this is in a comparison table where relative assessment is acceptable. One vague instance noted.

**Chosen approach justified against benchmarks (10/25)**:
The proposal states the selected approach is "兼顾易用性和可靠性，与用户需求最匹配" but does not provide a structured argument for why this beats the embedded form assistant (which would be lower-risk and faster to ship) for the initial use case. The "Partial" verdict on the embedded approach effectively dismisses it without full evaluation.

Quote: *"Selected: 兼顾易用性和可靠性，与用户需求最匹配"* — assertion without supporting argument.

---

### 4. Requirements Completeness — 82/110

**Scenario coverage (32/40)**:
Happy paths for all four intent types are covered with examples. Edge cases include: ambiguous input, unrecognizable intent, permission denial, statemachine violation. Good coverage.

Gaps: No scenario for concurrent editing (user edits card while AI is processing a follow-up message), no scenario for very long conversations that exceed context window, no scenario for network interruption during submission.

**Non-functional requirements (32/40)**:
NFRs are specific and quantified where it matters: P95 latency < 5s, accuracy >= 85%/80%. Security, privacy, and accessibility are addressed.

Quote: *"AI 服务供应商必须支持数据不用于训练（zero data retention）"* — specific and actionable.

Gaps:
- No mention of internationalization / multi-language support (the examples are all in Chinese).
- No offline / degraded-mode NFR (what happens when AI service is slow but not down?).
- No maximum conversation length or session timeout NFR.

**Constraints & dependencies (18/30)**:
Dependencies on existing API endpoints are well-documented. The available-transitions endpoint verification is excellent.

Quote: *"状态机预校验依赖 available-transitions 端点，该端点已存在于 MainItem、SubItem、MilestoneMap、Milestone 四个实体路由中（见 router.go），无需新增"* — very concrete.

Gaps:
- No constraint on prompt/context size limits (6 entities with full schema could be large).
- No constraint on supported browsers or screen sizes for the floating bubble.
- No mention of AI service rate limits or quota constraints.
- No dependency on specific React versions or UI library versions.

---

### 5. Solution Creativity — 62/100

**Novelty over industry baseline (24/40)**:
The card-as-single-source-of-truth hybrid model is a genuine innovation over pure chatbot or pure form-fill approaches. However, the proposal acknowledges this is inspired by Slack Bot + Notion AI patterns.

Quote: *"不同于纯对话机器人或纯表单填充，采用卡片为中心的混合模式——卡片是唯一数据源（single source of truth）"* — clear differentiation.

The novelty is primarily in the interaction model combination rather than in any fundamental technical innovation. The "context awareness" (page-level team detection) is standard practice, not novel.

**Cross-domain inspiration (22/35)**:
References span project management (Linear, Jira), communication (Slack), and productivity (Notion, Microsoft 365, GitHub). The combination is well-assembled. However, no inspiration is drawn from domains outside software tools (e.g., customer service chatbots, voice assistants, game UI patterns).

**Simplicity of insight (16/25)**:
The core insight ("use cards for structured data, conversation for unstructured input") is clean and understandable. The implementation complexity (dual-input to same state, prompt construction for 6 entity types, fallback mechanisms) may undermine the simplicity. The "why didn't I think of that" quality is moderate — the approach is sensible but not surprising.

---

### 6. Feasibility — 72/100

**Technical feasibility (32/40)**:
Strong case: existing APIs are verified, tech stack is standard React, the task is bounded (4 intent types x 6 entities). The AI service evaluation is reasonable.

Quote: *"意图识别和实体抽取对结构化输入（4 意图类型 × 6 实体，有限字段集）属于高确定性任务，预期准确率可达目标"* — good justification.

Concern: The prompt construction complexity is understated. Dynamically assembling entity schemas, user permissions, and statemachine rules for 6 entities into a prompt that reliably produces structured output is non-trivial. No prototype or PoC result is cited.

**Resource & timeline feasibility (22/30)**:
The estimate of 5-7 weeks total is reasonable for the scope described. However:

Quote: *"前端 2-3 周（气泡 UI + 聊天面板 + 卡片组件），后端 2-3 周（AI 代理层 + prompt 管理 + 日志），联调测试 1 周"* — the ranges are wide (2-3 weeks = 50% variance). No breakdown of what is included in each phase, and the 1-week integration testing seems optimistic for a system with 4 intent types x 6 entities.

**Dependency readiness (18/30)**:
Quote: *"AI 服务选型建议在 tech-design 阶段确定具体供应商"* — the most critical external dependency (AI service) is not yet evaluated with a PoC. This is a significant risk for feasibility assessment.

No evaluation of AI service pricing, rate limits, or contractual requirements. The proposal assumes a suitable service exists with the right features (structured output, zero data retention, acceptable latency) but has not validated this.

---

### 7. Scope Definition — 65/80

**In-scope items are concrete (24/30)**:
Most items are deliverable-grade: "全局浮动气泡聊天 UI 组件（展开/收起、拖拽定位）", "AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）". These can be traced to implementation tasks.

Quote: *"支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool"* — explicit and bounded.

Minor issue: "与现有后端 API 对接，复用权限检查和业务逻辑" is somewhat vague — which APIs specifically? All CRUD endpoints for all 6 entities?

**Out-of-scope explicitly listed (20/25)**:
Seven items are explicitly out of scope, covering AI training, voice, multi-user, notifications, reports, cross-team search, and conversation persistence. Good coverage.

Missing from out-of-scope: Multi-language support, mobile responsive design, dark mode for chat panel, accessibility beyond keyboard shortcuts.

**Scope is bounded (21/25)**:
The scope is bounded by entity list (6), intent types (4), and the "no training, no persistence" exclusions. The proposal is executable within the stated 5-7 week timeline.

However, In Scope claims "全操作覆盖" (full CRUD for all entities) which is ambitious. Combined with the "user chose to proceed with full scope" note in Assumptions Challenged, there is a risk of scope creep if the full scope proves too large.

---

### 8. Risk Assessment — 74/90

**Risks identified (26/30)**:
Eight risks covering AI accuracy, latency, state sync, service availability, statemachine conflicts, data privacy, prompt injection, and cost control. Good breadth. Missing: vendor lock-in risk, prompt versioning/migration risk (when entity schemas change), and user adoption risk (users may not trust AI-generated suggestions).

**Likelihood + impact rated (22/30)**:
Ratings are generally reasonable. Concern: Six out of eight risks are rated "M" (Medium) likelihood. This clustering suggests the assessment may not be differentiating enough. The "AI 服务不可用" risk is rated "L" likelihood — for an external API dependency, this is optimistic.

Quote: *"外部 AI 服务不可用导致功能完全失效 | L | H"* — Low likelihood for an external service dependency is debatable. Any cloud service has non-trivial downtime.

**Mitigations are actionable (26/30)**:
Most mitigations are concrete and implementable: "调用 available-transitions API 校验", "设置每用户每日调用次数上限", "后端对用户输入做基础清洗". These can be translated to tasks.

Quote: *"AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作"* — actionable and user-centric.

One mitigation is less actionable: "定义明确的意图类型和字段映射规则" — this is a design activity, not a runtime mitigation.

---

### 9. Success Criteria — 60/80

**Criteria are measurable and testable (22/30)**:
Most SC items are verifiable: "字段准确率 >= 80%" (measurable from logs), "P95 延迟 < 5 秒" (measurable from telemetry), "无权限用户执行操作时收到明确的权限不足提示" (testable).

Some items are less testable:
- Quote: *"AI 成功推送预填卡片且字段准确率 >= 80%"* — "准确率" requires defining what counts as "accurate" for each field type. Is a date that is off by one day "accurate"? Is a fuzzy-matched assignee "accurate"?
- Quote: *"AI 正确解析父子关系并推送预填卡片"* — "正确解析" is binary but the definition of correctness for fuzzy matching is unclear.

**Coverage is complete (14/25)**:
Significant coverage gaps:
- In Scope lists 6 entities (MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool). SC covers MainItem (create + query + modify), SubItem (create + progress), Milestone (create), ProgressRecord (add). Missing: **MilestoneMap CRUD**, **ItemPool CRUD**, **SubItem status change**, **Milestone status change**, **any delete operation**.
- In Scope lists "全局浮动气泡...拖拽定位" — no SC for drag positioning.
- In Scope lists "聊天消息界面（会话内消息历史、气泡消息、系统消息）" — no SC for message history display or system messages.

**SC internal consistency (24/25)**:
The 12 SC items are internally consistent. No pair of SC items is mutually exclusive. The clustering analysis (above) found no contradictions, only coverage gaps.

One ambiguity: SC 1 says "至少包含标题 + 1 个额外字段" which sets a low bar for the "full CRUD for all entities" scope. This may be intentional as a minimum viable check, but it could also be interpreted as the only requirement.

---

### 10. Logical Consistency — 72/90

**Solution addresses the stated problem (28/35)**:
The form-filling inefficiency is directly addressed by AI-powered pre-filling and card-based interaction. The statemachine/RBAC complexity is partially addressed (AI validates statemachine rules, but users still need to know what they want to do).

Gap: The problem states "用户需要理解状态机规则和权限体系才能正确操作" but the solution does not help users understand what is possible — it only prevents invalid operations. A user who does not know what "进行中" means still does not know, even if the AI rejects an invalid transition.

Quote: *"状态变更不符合状态机规则 → 返回错误说明，提示合法的目标状态"* — this helps after failure, but does not proactively guide.

**Scope <-> Solution <-> SC aligned (22/30)**:
Alignment between Scope and Solution is strong. SC alignment has gaps (see D9 coverage analysis).

The most notable misalignment: In Scope claims "全操作覆盖" for 6 entities, but SC only tests 4 of 6 entities. The gap for MilestoneMap and ItemPool is unexplained.

Quote: *"不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作"* — "完整 CRUD" includes Delete, which has no SC, no scenario, and no risk assessment.

**Requirements <-> Solution coherent (22/25)**:
Requirements map cleanly to the solution. Each scenario type has a corresponding solution component. The NFRs (latency, accuracy, security) map to specific architectural decisions (backend proxy, structured output, confirmation cards).

Minor gap: The "entity resolution strategy" (exact -> fuzzy -> disambiguation card) is described in Constraints but has no corresponding SC to verify it works.

---

## Phase 3: Blindspot Hunt

1. **[blindspot] No rollback plan**: The proposal introduces a global floating bubble UI component and an AI proxy layer in the backend. If the feature needs to be rolled back, there is no discussion of feature flags, gradual rollout, or how to cleanly remove the components. For an infrastructure-level change (new backend proxy, new AI service dependency), this is a significant omission.

2. **[blindspot] Conversation context management is unaddressed**: The proposal describes multi-turn interactions ("对话补充字段") but does not address conversation state management. How many previous messages are included in each AI call? What happens when a conversation exceeds the context window? Is there a conversation reset mechanism? This is architecturally critical for the backend proxy design.

3. **[blindspot] Cost estimation absent**: While risk #8 mentions "AI API 调用成本失控", the proposal provides no cost estimate. With "每用户每日调用次数上限" as the mitigation, what is the expected per-user cost at that limit? For a small team tool, AI API costs could exceed infrastructure costs. The feasibility assessment is incomplete without this.

4. **[blindspot] Assumption "user chose to proceed with full scope" is unreferenced**: Quote: *"Challenge Override: user chose to proceed with full scope. Reason: 希望一次到位"*. This references a user conversation that is not documented. The proposal's scope justification relies on an undocumented decision. If this is the product owner's call, it should be attributed; if it is a hypothetical user, it should be validated.

5. **[blindspot] No mobile/responsive strategy**: The floating bubble + chat panel pattern works well on desktop but is problematic on mobile screens. The proposal does not address responsive behavior, despite the constraint "前端技术栈为 React + TypeScript + Tailwind CSS + Radix UI" which is typically a responsive stack. Is mobile support in scope or out of scope? This ambiguity could lead to scope disputes.

6. **[blindspot] Multi-entity operations are not addressed**: Users may naturally say "把所有P0事项的状态改为进行中" or "创建3个子任务". The proposal's examples are all single-entity operations. Batch operations would stress the card model (single card per operation? multiple cards?) and are not discussed as either in-scope or out-of-scope.

---

## Summary

| # | Dimension | Score | Max |
|---|-----------|-------|-----|
| 1 | Problem Definition | 72 | 110 |
| 2 | Solution Clarity | 95 | 120 |
| 3 | Industry Benchmarking | 82 | 120 |
| 4 | Requirements Completeness | 82 | 110 |
| 5 | Solution Creativity | 62 | 100 |
| 6 | Feasibility | 72 | 100 |
| 7 | Scope Definition | 65 | 80 |
| 8 | Risk Assessment | 74 | 90 |
| 9 | Success Criteria | 60 | 80 |
| 10 | Logical Consistency | 72 | 90 |
| | **Total** | **736** | **1000** |

---

## ATTACKS (Prioritized)

1. **[D9: Success Criteria] Coverage gap for 2 of 6 entities**: Quote: *"支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool"* (In Scope) — but MilestoneMap and ItemPool have zero SC entries. Either remove them from scope or add verifiable criteria. This inflates scope without accountability.

2. **[D1: Problem Definition] No external evidence of user pain**: Quote: *"MainItem 创建表单包含 title、description、priority、assignee..."* — this is code inspection, not user feedback. The entire evidence section is introspective. Add at least one data point from user research, support tickets, or usage analytics.

3. **[D3: Industry Benchmarking] Chosen approach not justified against benchmarks**: Quote: *"Selected: 兼顾易用性和可靠性，与用户需求最匹配"* — this is an assertion, not an argument. Explain why the lower-risk embedded form assistant is insufficient for an MVP, or why the Slack Bot pattern's context-switching problem makes it inferior despite faster delivery.

4. **[D6: Feasibility] Critical dependency (AI service) has no PoC validation**: Quote: *"AI 服务选型建议在 tech-design 阶段确定具体供应商"* — the most important technical risk (can an AI service reliably parse 4 intent types x 6 entities with >= 85% accuracy?) is deferred without evidence. A 1-day PoC with Claude API or GPT API would significantly strengthen feasibility claims.

5. **[D2: Solution Clarity] Technical direction for dual-input card is underspecified**: Quote: *"对话输入和直接编辑均通过统一的 dispatch 写入卡片状态"* — this is the single most novel technical element, but receives one sentence. How does the dispatch resolve conflicts? What is the event flow? This needs a sequence diagram or at least a paragraph in the proposal.

6. **[D9: Success Criteria] Field accuracy definition is ambiguous**: Quote: *"AI 成功推送预填卡片且字段准确率 >= 80%"* — is a fuzzy-matched assignee (user says "张三", system matches "张三丰") counted as accurate? Is a date parsed as Friday when the user said "下周五" but meant "this Friday" accurate? Define accuracy per field type.

7. **[D8: Risk Assessment] External service downtime rated too low**: Quote: *"外部 AI 服务不可用导致功能完全失效 | L"* — for any cloud API, medium likelihood is more realistic. Every major AI provider has had outages. Re-rate to M and validate the degradation UX.

8. **[D4: Requirements Completeness] Conversation context management is a missing requirement**: Multi-turn interactions are central to the proposal ("继续对话补充字段"), but no requirement addresses context window limits, conversation length, or session management. Add NFR or constraint for maximum conversation turns.

9. **[D10: Logical Consistency] "全操作覆盖" claim vs. scope reality**: Quote: *"不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作"* — "完整 CRUD" implies Delete, which has no scenario, no SC, and no risk. Either add delete support with corresponding SC, or change "完整 CRUD" to the actual supported operations.

10. **[D7: Scope Definition] Mobile/responsive behavior is unaddressed**: The floating bubble + chat panel is a desktop-first pattern. No mention of mobile behavior exists in scope, out-of-scope, or NFRs. Explicitly declare mobile in or out of scope.

11. **[blindspot] No rollback/feature-flag strategy**: For an infrastructure-level change (new AI proxy layer, global UI component), there is no discussion of how to safely deploy, gradually roll out, or roll back. Add a rollout strategy to Feasibility or Risk.

12. **[blindspot] Batch/multi-entity operations are silent**: Users will naturally say things like "把所有P0事项分配给张三". The proposal only addresses single-entity operations. Declare batch operations as in-scope or out-of-scope.

13. **[blindspot] Undocumented user decision drives scope**: Quote: *"Challenge Override: user chose to proceed with full scope"* — this references an undocumented conversation. Attribute the decision to a named stakeholder or validate it through user research.
