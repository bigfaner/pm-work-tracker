# Proposal Evaluation Report — AI Copilot 对话助手

**Iteration**: 2
**Date**: 2026-06-11
**Document**: `docs/proposals/ai-copilot/proposal.md`
**Previous Score**: 736/1000

---

## Iteration 1 Re-verification

| # | Attack (Iteration 1) | Status | Finding |
|---|----------------------|--------|---------|
| 1 | D9: MilestoneMap/ItemPool missing from SC | Resolved | SC items 7 and 8 now cover ItemPool and MilestoneMap |
| 2 | D1: No external evidence | Partially resolved | Footnote added with team feedback + click-count metric, but still no support tickets, analytics tool output, or third-party data |
| 3 | D3: Chosen approach not justified | Partially resolved | Verdict cell now includes architectural reasoning (single binary, no Slack integration), but no structured argument against the lower-risk alternative |
| 4 | D6: No PoC validation | Partially resolved | PoC recommendation added (1 day, 20-30 samples), but not yet executed — feasibility still unvalidated |
| 5 | D2: Dual-input card underspecified | Partially resolved | Innovation Highlights expanded to describe single-source-of-truth model, but still one paragraph with no event flow or conflict resolution detail |
| 6 | D9: Field accuracy ambiguous | Resolved | Per-field accuracy definition added with date tolerance and assignee fuzzy-match rules |
| 7 | D8: External service downtime rated too low | Resolved | Re-rated from L to M |
| 8 | D4: Conversation context management missing | Resolved | Constraint added: max 50 turns per session, cleared on navigation |
| 9 | D10: "全操作覆盖" vs CRUD reality | Partially resolved | Footnote clarifies intent taxonomy, but "完整 CRUD" phrasing still implies Delete which has no SC |
| 10 | D7: Mobile unaddressed | Resolved | Explicitly listed in Out of Scope |
| 11 | [blindspot] No rollback plan | Resolved | Feature flag constraint added |
| 12 | [blindspot] Batch operations silent | Resolved | Explicitly listed in Out of Scope |
| 13 | [blindspot] Undocumented user decision | Partially resolved | Attributed to "产品负责人" but still no documented source or date |

---

## Phase 1: Reasoning Audit

### Problem -> Solution Trace

| Problem Statement | Solution Element | Trace Quality |
|---|---|---|
| 创建 MainItem 需要手动填写 10+ 个字段 | AI 解析意图后推送预填表单卡片 | Direct mapping. Solved. |
| 用户需要理解状态机规则和权限体系才能正确操作 | AI 自动校验状态机合法性 + 错误提示合法目标状态 | Partially addressed. Reactive only — prevents errors but does not proactively guide users toward valid operations. |
| 非技术人员填写结构化字段困难 | 对话输入 + AI 推断字段值 | Direct mapping. Solved. |

**Verdict**: Solution directly addresses stated problems. The RBAC-understanding gap remains: a user who does not know what "进行中" means still does not know, even if the AI rejects an invalid transition. The solution reduces failure cost but not cognitive burden.

### Solution -> Evidence Trace

- Evidence footnote added: *"团队内部反馈（多名成员反映表单字段多、操作路径长）及用户操作时长分析（创建一个完整 MainItem 平均需 8-12 次点击/输入交互）"*
- The footnote adds qualitative user feedback and a quantitative metric (8-12 clicks). This is an improvement over iteration 1.
- However, "多名成员" is vague (how many?), "用户操作时长分析" references an analysis without citing its source or methodology, and no external validation (support tickets, analytics platform screenshots, user interviews) is provided.

**Verdict**: Evidence is improved but still entirely internal/self-reported. No external validation.

### Evidence -> Success Criteria Trace

- SC now covers all 6 entities (MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool) — coverage gap from iteration 1 is resolved.
- SC field accuracy definition (per-field matching, date tolerance +/-1 day, assignee fuzzy match) maps directly to the "10+ fields" problem.

### Self-Contradiction Check

- The `consistency_check_result` comment block confirms reconciliation of 4 prior attacks.
- No new contradictions introduced in this iteration.
- Remaining tension: "完整 CRUD 操作" in Innovation Highlights implies Delete, but Out of Scope explicitly excludes "删除操作（第一版不支持通过 AI 执行删除）". This is a minor wording contradiction — "完整 CRUD" should be changed to "完整操作（创建、查询、修改、分配）" or similar.

---

## Phase 2: Rubric Scoring

### 1. Problem Definition — 80/110

**Problem stated clearly (32/40)**:
Core problem remains concrete: "创建一个 MainItem 需要手动填写 10+ 个字段，操作效率低". The RBAC/statemachine problem is stated but conflated with form-filling — these are distinct problems (efficiency vs. comprehension) that deserve separate articulation.

Quote: *"用户需要理解状态机规则和权限体系才能正确操作，门槛较高"* — still not broken down into specific failure modes (wrong status transitions? permission denied errors? abandoned operations?).

**Evidence provided (24/40)**:
Improvement from iteration 1. The footnote adds: *"多名成员反映表单字段多、操作路径长"* and *"创建一个完整 MainItem 平均需 8-12 次点击/输入交互"*.

However: "多名成员" is unquantified. "用户操作时长分析" references an analysis without methodology. No support tickets, analytics dashboards, or interview transcripts are cited. The `todos.txt` reference remains — this validates roadmap intent, not user pain.

Quote: *"后续可通过埋点数据量化验证"* — this acknowledges the evidence gap but does not fill it.

**Urgency justified (24/30)**:
Argument is logical but unquantified: "随着团队和事项数量增长，表单交互的效率瓶颈会加剧." No data on current team size, item creation volume, or time-per-creation to quantify the cost of delay.

Quote: *"延迟实施意味着团队持续承担低效操作的时间成本"* — directionally correct but no dollar or hour estimate.

---

### 2. Solution Clarity — 98/120

**Approach is concrete (37/40)**:
The 5-step interaction flow is specific. The card-as-single-source-of-truth model is now more clearly defined. A reader can explain back the complete interaction model.

Quote: *"AI 解析意图后推送预填表单卡片到聊天窗口。必填且无法推导的字段留空，用户可通过直接编辑卡片或继续对话补充字段，确认后提交执行"* — clear and specific.

Minor gap: The floating bubble's collapsed-state behavior (how it indicates unread messages, where it positions on first load) is still not described.

**User-facing behavior described (42/45)**:
Each scenario type has concrete input/output examples. Edge cases cover ambiguity, permission denial, statemachine violations, and unrecognizable intent. The note clarifying that intent taxonomy reflects user language (not API endpoints) is a valuable clarification.

Quote: *"意图分类（创建/查询/修改/分配）反映用户自然语言的表达习惯，而非后端 API 端点划分"* — good clarification.

**Technical direction clear (19/35)**:
Backend proxy architecture is stated. Structured output / tool use is mentioned. Prompt construction strategy is referenced but underspecified.

Quote: *"后端负责 prompt 构造（动态组装当前 Team 的实体 schema、用户权限范围、状态机规则）"* — this describes what the prompt contains but not how it fits within context window limits, how schema changes are reflected, or how large the prompt would be for 6 entities.

The dual-input card architecture (the most novel element) is described in one paragraph: *"对话输入和直接编辑均通过统一的 dispatch 写入卡片状态，对话区域仅做展示回显，无需双向绑定"* — the dispatch pattern is named but the event flow, conflict resolution, and state shape are not elaborated.

---

### 3. Industry Benchmarking — 90/120

**Industry solutions referenced (34/40)**:
Six references: Linear, Notion AI, Slack Bot, Jira Automation, GitHub Copilot, Microsoft 365 Copilot. Good breadth.

**At least 3 meaningful alternatives (24/30)**:
Four alternatives: Do nothing, embedded form assistant, Slack Bot, conversational + card hybrid. "Do nothing" is present.

Remaining issue: The "embedded form assistant" (内嵌式 AI 助手) is evaluated as "Partial" rather than as a standalone alternative. It is described through the lens of the selected approach's scope ("查询/修改仍需独立入口") rather than evaluated on its own merits as a viable first phase.

**Honest trade-off comparison (18/25)**:
The comparison table has pros/cons. The selected approach's cons are: "AI 准确性依赖外部服务，开发量较大". "开发量较大" remains vague — no effort estimate to compare against alternatives.

Quote: *"开发量较大"* — this is the only con for the most complex approach. No quantification.

**Chosen approach justified against benchmarks (14/25)**:
The Verdict cell for the selected approach now includes architectural reasoning:

Quote: *"本系统为单二进制嵌入式 Web 应用，无 Slack/Teams 集成场景，Slack Bot 模式不适用；Command Palette 仅支持单次操作，无法处理多轮对话补充字段的场景。对话 + 卡片混合模式在现有 Web 应用内原生嵌入..."*

This is an improvement. The justification explains why Slack Bot and Command Palette don't fit. However, it does not address why the embedded form assistant (lower risk, faster delivery) is insufficient as a first phase. The "Partial" verdict on the embedded approach effectively dismisses it without full evaluation of its MVP potential.

---

### 4. Requirements Completeness — 90/110

**Scenario coverage (35/40)**:
Happy paths for all four intent types. Edge cases: ambiguity, unrecognizable intent, permission denial, statemachine violation. Good coverage.

Remaining gaps:
- No scenario for concurrent editing (user edits card while AI is processing a follow-up message).
- No scenario for network interruption during submission.
- No scenario for a user who sends a very long message exceeding the input length limit mentioned in Risk #7.

**Non-functional requirements (37/40)**:
NFRs are specific and quantified: P95 latency < 5s, accuracy >= 85%/80%, per-field accuracy definitions. Security, privacy, and accessibility addressed. Session constraint (50 turns max) added.

Quote: *"字段准确率按 per-field 匹配率计算：AI 提取值与用户意图一致则计为正确；日期字段在 ±1 天范围内计为正确；assignee 字段在 Team 成员模糊匹配范围内计为正确"* — precise and testable.

Minor gaps:
- No internationalization/multi-language NFR (all examples are Chinese).
- No degraded-mode NFR (slow but not down AI service).

**Constraints & dependencies (18/30)**:
Dependencies on existing API endpoints are well-documented. available-transitions endpoint verification is excellent. Session constraints (50 turns, navigation clears) are new and valuable.

Quote: *"对话上下文窗口仅限当前会话，单会话最大 50 轮对话，页面导航时清除会话"* — concrete constraint.

Remaining gaps:
- No constraint on prompt/context size limits (6 entities with full schema could produce a very large prompt).
- No constraint on supported browsers for the floating bubble.
- No mention of AI service rate limits or quota constraints from the provider side.

---

### 5. Solution Creativity — 65/100

**Novelty over industry baseline (25/40)**:
The card-as-single-source-of-truth hybrid model is a genuine innovation over pure chatbot or pure form-fill. However, the proposal acknowledges inspiration from Slack Bot + Notion AI patterns, and the "context awareness" (page-level team detection) is standard practice.

Quote: *"不同于纯对话机器人或纯表单填充，采用卡片为中心的混合模式——卡片是唯一数据源（single source of truth）"* — clear differentiation.

The novelty is in the interaction model combination, not fundamental technical innovation. The "why didn't I think of that" quality is moderate.

**Cross-domain inspiration (22/35)**:
References span project management, communication, and productivity tools. No inspiration drawn from domains outside software tools (customer service chatbots, voice assistants, game UI patterns, etc.).

**Simplicity of insight (18/25)**:
Core insight ("use cards for structured data, conversation for unstructured input") is clean and understandable. Implementation complexity (dual-input state management, prompt construction for 6 entity types, fallback mechanisms) may undermine the simplicity in practice.

---

### 6. Feasibility — 78/100

**Technical feasibility (34/40)**:
Strong case: existing APIs verified, tech stack is standard, task bounded (4 intents x 6 entities). AI service evaluation is reasonable.

The PoC recommendation is a positive addition:

Quote: *"建议用 1 个工作日进行 PoC 验证——构造 20-30 条典型用户输入样本，通过目标 AI 服务的 structured output / function calling 接口测试意图识别和字段提取准确率"* — concrete and actionable.

However, the PoC has not been executed. The feasibility claim ("预期准确率可达目标") remains an assertion. The prompt construction complexity for 6 entities with dynamic schema assembly is still understated.

**Resource & timeline feasibility (24/30)**:
Estimate of 5-7 weeks is reasonable. Ranges are still wide (2-3 weeks = 50% variance). The 1-week integration testing for 4 intents x 6 entities remains optimistic.

Quote: *"前端 2-3 周（气泡 UI + 聊天面板 + 卡片组件），后端 2-3 周（AI 代理层 + prompt 管理 + 日志），联调测试 1 周"* — ranges not tightened from iteration 1.

**Dependency readiness (20/30)**:
Existing APIs confirmed ready. AI service not yet selected or validated:

Quote: *"AI 服务选型建议在 tech-design 阶段确定具体供应商"* — the most critical external dependency remains uncommitted.

No evaluation of AI service pricing, rate limits, or contractual requirements.

---

### 7. Scope Definition — 74/80

**In-scope items are concrete (27/30)**:
Items are deliverable-grade: "全局浮动气泡聊天 UI 组件（展开/收起、拖拽定位）", "AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）", "支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool".

Quote: *"预填表单卡片组件（对话输入和直接编辑均写入同一卡片状态，卡片为唯一数据源）"* — concrete and traceable.

Minor issue: "与现有后端 API 对接，复用权限检查和业务逻辑" is vague — which endpoints specifically?

**Out-of-scope explicitly listed (23/25)**:
Nine items explicitly out of scope: AI training, delete, voice, multi-user, notifications, reports, cross-team search, conversation persistence, mobile responsive, batch operations. Good coverage. Improved from iteration 1 (added mobile, batch operations).

**Scope is bounded (24/25)**:
Bounded by entity list (6), intent types (4), session limit (50 turns), and exclusions. The scope is executable within 5-7 weeks.

Remaining tension: Innovation Highlights claim "全操作覆盖" but scope explicitly excludes delete and batch. The "full coverage" claim should be qualified.

---

### 8. Risk Assessment — 80/90

**Risks identified (27/30)**:
Eight risks covering accuracy, latency, state sync, service availability, statemachine conflicts, privacy, prompt injection, and cost. Good breadth.

Missing: vendor lock-in risk (switching AI provider requires prompt migration), prompt versioning risk (schema changes require prompt updates), user adoption risk (users may not trust AI suggestions).

**Likelihood + impact rated (25/30)**:
Improvement from iteration 1: external service unavailability now rated M (was L). However, six of eight risks are still rated M likelihood — the assessment lacks differentiation. The clustering suggests insufficient calibration.

**Mitigations are actionable (28/30)**:
Most mitigations are concrete: "调用 available-transitions API 校验", "设置每用户每日调用次数上限", "后端对用户输入做基础清洗". These are implementable.

One mitigation is less actionable: "定义明确的意图类型和字段映射规则" — this is a design activity, not a runtime mitigation. It should be rephrased as a design constraint that is validated before implementation.

Quote: *"AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作"* — actionable and user-centric.

---

### 9. Success Criteria — 72/80

**Criteria are measurable and testable (27/30)**:
Most SC items are verifiable. The per-field accuracy definition is now precise:

Quote: *"字段准确率按 per-field 匩配率计算：AI 提取值与用户意图一致则计为正确；日期字段在 ±1 天范围内计为正确；assignee 字段在 Team 成员模糊匹配范围内计为正确"* — precise and testable.

Minor ambiguity: SC item 1 says "至少包含标题 + 1 个额外字段" which sets a low bar. Is this the minimum acceptable or the expected norm?

**Coverage is complete (21/25)**:
Significant improvement from iteration 1. SC now covers all 6 entities: MainItem (create + query + modify), SubItem (create + progress), Milestone (create), MilestoneMap (create), ProgressRecord (add), ItemPool (submit).

Remaining gaps:
- No SC for drag positioning of the floating bubble (In Scope lists "拖拽定位").
- No SC for message history display or system messages (In Scope lists "会话内消息历史、气泡消息、系统消息").
- No SC for entity resolution strategy (exact -> fuzzy -> disambiguation card) described in Constraints.

**SC internal consistency (24/25)**:
The 15 SC items are internally consistent. No pair of SC items is mutually exclusive. The clustering analysis found no contradictions.

---

### 10. Logical Consistency — 78/90

**Solution addresses the stated problem (30/35)**:
Form-filling inefficiency is directly addressed. The RBAC/statemachine comprehension problem remains only partially solved — the solution prevents invalid operations but does not proactively guide users toward understanding what is possible.

Quote: *"状态变更不符合状态机规则 → 返回错误说明，提示合法的目标状态"* — reactive, not proactive.

**Scope <-> Solution <-> SC aligned (25/30)**:
Improved from iteration 1. SC now covers all 6 entities. Alignment between Scope and Solution is strong.

Remaining misalignment: Innovation Highlights claim "全操作覆盖" but scope excludes delete and batch. "完整 CRUD 操作" implies Delete which has no SC, no scenario, and no risk.

Quote: *"不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作"* — "完整 CRUD" includes Delete. Either remove "完整 CRUD" or add delete support with SC.

**Requirements <-> Solution coherent (23/25)**:
Requirements map cleanly to the solution. Each scenario type has a corresponding solution component. NFRs map to architectural decisions.

Minor gap: Entity resolution strategy (exact -> fuzzy -> disambiguation card) is described in Constraints but has no corresponding SC to verify it works.

---

## Phase 3: Blindspot Hunt

1. **[blindspot] "完整 CRUD" wording contradiction with Out of Scope**: Quote: *"不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作"* (Innovation Highlights) vs. *"删除操作（第一版不支持通过 AI 执行删除，用户需通过传统界面操作）"* (Out of Scope). The Innovation Highlights claim "完整 CRUD" which includes Delete, but Out of Scope explicitly excludes it. This is a direct wording contradiction that could cause scope disputes during implementation.

2. **[blindspot] Evidence footnote uses circular reasoning**: Quote: *"以上痛点综合自团队内部反馈（多名成员反映表单字段多、操作路径长）及用户操作时长分析（创建一个完整 MainItem 平均需 8-12 次点击/输入交互）. 后续可通过埋点数据量化验证"* — the proposal acknowledges that quantitative validation is pending ("后续可通过埋点数据量化验证"), which means the "8-12 次点击" number is currently unverified. Presenting it as evidence while simultaneously saying it needs verification later is circular — the evidence is the claim.

3. **[blindspot] Prompt size feasibility for 6 entities**: Quote: *"后端负责 prompt 构造（动态组装当前 Team 的实体 schema、用户权限范围、状态机规则）"* — assembling schemas for 6 entities (each with multiple fields), permission scopes, and statemachine rules into a single prompt could easily exceed several thousand tokens. The proposal does not address prompt size limits, context window consumption, or the trade-off between complete schema inclusion and response latency/cost. This is a feasibility risk hidden in a single sentence.

4. **[blindspot] No cost model for AI service**: Risk #8 mentions "AI API 调用成本失控" and mitigation is "设置每用户每日调用次数上限". But no cost model is provided: what is the estimated per-request cost? What is the expected daily active user count? At the stated accuracy target (85%), 15% of requests may need follow-up (retries, corrections), multiplying costs. The feasibility section should include a rough cost estimate.

5. **[blindspot] "产品负责人决定" is still an undocumented appeal to authority**: Quote: *"Challenge Override: 产品负责人决定一次性交付完整范围"* — the phrase "产品负责人决定" is an improvement over "user chose", but still references an undocumented decision. No date, no meeting notes, no written approval is cited. A scope-level decision that overrides a "5 Whys" challenge should be traceable to a documented artifact.

6. **[blindspot] Vendor lock-in risk absent**: The proposal evaluates three AI service options (Claude, OpenAI, local) but does not discuss switching costs. The prompt construction strategy, structured output schemas, and error handling patterns will be tightly coupled to the chosen provider's API. If the provider raises prices, changes terms, or degrades quality, the switching cost is not assessed.

---

## Summary

| # | Dimension | Score | Max |
|---|-----------|-------|-----|
| 1 | Problem Definition | 80 | 110 |
| 2 | Solution Clarity | 98 | 120 |
| 3 | Industry Benchmarking | 90 | 120 |
| 4 | Requirements Completeness | 90 | 110 |
| 5 | Solution Creativity | 65 | 100 |
| 6 | Feasibility | 78 | 100 |
| 7 | Scope Definition | 74 | 80 |
| 8 | Risk Assessment | 80 | 90 |
| 9 | Success Criteria | 72 | 80 |
| 10 | Logical Consistency | 78 | 90 |
| | **Total** | **805** | **1000** |

---

## ATTACKS (Prioritized)

1. **[D1: Problem Definition] Evidence is self-reported and partially unverified**: Quote: *"创建一个完整 MainItem 平均需 8-12 次点击/输入交互。后续可通过埋点数据量化验证"* — the proposal presents the 8-12 click metric as evidence while simultaneously stating it needs future verification. Either verify the metric before presenting it as evidence, or present it as an estimate. Add at least one external data point (support ticket, analytics dashboard screenshot, user interview transcript).

2. **[D10: Logical Consistency] "完整 CRUD" contradicts Out of Scope exclusion of delete**: Quote: *"不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作"* (Innovation Highlights) vs. *"删除操作（第一版不支持通过 AI 执行删除）"* (Out of Scope). Replace "完整 CRUD 操作" with the actual supported operations: "创建、查询、修改、分配操作".

3. **[D6: Feasibility] Prompt size for 6 entities is an unaddressed feasibility risk**: Quote: *"后端负责 prompt 构造（动态组装当前 Team 的实体 schema、用户权限范围、状态机规则）"* — 6 entities with full schemas, permission scopes, and statemachine rules could produce prompts exceeding context window limits or inflating cost/latency. Add a constraint or analysis estimating prompt token count for the worst case.

4. **[D3: Industry Benchmarking] Embedded form assistant alternative is not fully evaluated**: Quote: *"Partial: 适合创建场景，但查询/修改仍需独立入口"* — the embedded form assistant is dismissed as "Partial" rather than evaluated as a standalone alternative. For an MVP, it could deliver 60-70% of the value at 30-40% of the complexity. Provide a structured argument for why the full conversational approach is necessary from day one.

5. **[D6: Feasibility] No cost model for AI service**: Risk #8 identifies cost overrun but no cost estimate exists. Add a rough cost model: estimated tokens per request, requests per user per day, cost per 1K tokens for candidate providers, and projected monthly cost at target user volume.

6. **[D9: Success Criteria] SC does not cover all In Scope items**: In Scope lists "拖拽定位", "会话内消息历史、气泡消息、系统消息", and entity resolution strategy — none have corresponding SC entries. Add SC for drag positioning, message history rendering, and disambiguation card behavior.

7. **[D2: Solution Clarity] Dual-input card event flow is still underspecified**: Quote: *"对话输入和直接编辑均通过统一的 dispatch 写入卡片状态，对话区域仅做展示回显，无需双向绑定"* — the dispatch pattern is named but its event flow, conflict resolution (simultaneous chat input + direct edit), and state shape are not elaborated. This is the most technically novel element and deserves at least a sequence-level description.

8. **[D5: Solution Creativity] No inspiration from domains outside software tools**: All benchmarking references are from software project management and productivity tools. Customer service chatbots (intent resolution + confirmation patterns), voice assistants (multi-turn context management), and game UIs (floating assistant patterns) could provide relevant inspiration and are not considered.

9. **[D8: Risk Assessment] Six of eight risks rated M likelihood — insufficient differentiation**: Quote: every row in the risk table has M likelihood. This clustering suggests the assessment was not calibrated with probability ranges or historical data. Differentiate at least 2-3 risks to H or L with supporting reasoning.

10. **[blindspot] Vendor lock-in risk is absent**: The proposal evaluates three AI providers but does not discuss switching costs. The prompt construction, structured output schemas, and error handling will be provider-specific. Add a risk for provider switching cost and a mitigation (e.g., provider-agnostic abstraction layer).

11. **[blindspot] "产品负责人决定" is an undocumented appeal to authority**: Quote: *"Challenge Override: 产品负责人决定一次性交付完整范围"* — no date, no meeting notes, no written approval. A scope-level override of a "5 Whys" challenge should be traceable to a documented decision artifact.

12. **[blindspot] Degraded-mode NFR is missing**: The proposal addresses "AI service down" (show message, use traditional forms) but not "AI service slow" (e.g., 8-12 second responses). What is the user experience when latency exceeds the 5s P95 target but the service is still responding? Add a degraded-mode NFR or behavior specification.
