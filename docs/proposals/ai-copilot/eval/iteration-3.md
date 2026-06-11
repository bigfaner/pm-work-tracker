# Proposal Evaluation Report — AI Copilot 对话助手

**Iteration**: 3
**Date**: 2026-06-11
**Document**: `docs/proposals/ai-copilot/proposal.md`
**Previous Score**: 805/1000

---

## Iteration 2 Re-verification

| # | Attack (Iteration 2) | Status | Finding |
|---|-----------------------|--------|---------|
| 1 | [D1] Evidence is self-reported and partially unverified | Partially resolved | Footnote improved: now explicitly states "此为开发者走查估算，非经埋点分析的量化数据，后续需通过埋点数据验证". The estimate is now honestly labeled as a walkthrough estimate, not analytics data. However, no external validation has been added — evidence remains entirely internal. |
| 2 | [D10] "完整 CRUD" contradicts Out of Scope | Partially resolved | "完整 CRUD" removed from Innovation Highlights. Now says "全操作覆盖". Out of Scope still excludes delete. The term "全操作覆盖" is less precise than listing the actual operations but no longer contains the word "CRUD". Tension reduced but not eliminated — "全操作覆盖" still implies comprehensive coverage while excluding two operation types. |
| 3 | [D6] Prompt size feasibility for 6 entities | Resolved | Quote added: "估算 prompt 规模：6 个实体 schema × ~200 tokens/实体 + 系统指令 ≈ 2000-3000 tokens，在主流 AI 服务的上下文窗口限制内（如 Claude 200K、GPT-4 128K）". Concrete estimate with context window comparison. |
| 4 | [D3] Embedded form assistant not fully evaluated | Not resolved | Comparison table verdict for embedded form assistant remains "Partial" with the same justification. No structured argument added for why the full conversational approach is necessary from day one versus a phased approach. |
| 5 | [D6] No cost model for AI service | Resolved | Cost estimate added: "假设 20 活跃用户/天 × 平均 5 次 AI 调用/用户/天 × $0.015/调用（Claude Haiku 级别，structured output）≈ $45/月。若使用更高阶模型（如 Claude Sonnet），成本约 $150-200/月。结合每日每用户调用次数上限（见风险管理），月成本可控在 $200 以内". Concrete numbers with assumptions stated. |
| 6 | [D9] SC does not cover all In Scope items | Resolved | SC items added: (15) message history + drag positioning, (16) disambiguation card. Coverage now includes drag positioning, message history, and entity resolution strategy. |
| 7 | [D2] Dual-input card event flow underspecified | Not resolved | Innovation Highlights description remains one paragraph. No event flow diagram, sequence description, or conflict resolution detail added. |
| 8 | [D5] No inspiration from domains outside software tools | Not resolved | Benchmarking references unchanged. Still limited to software tools: Linear, Notion AI, Slack Bot, Jira, GitHub Copilot, Microsoft 365 Copilot. No cross-domain references added. |
| 9 | [D8] Six of eight risks rated M likelihood — insufficient differentiation | Not resolved | Risk table likelihood ratings unchanged: Accuracy M, Latency M, State sync L, Service down M, State machine M, Privacy M, Injection M, Cost L. Still 5 of 8 rated M (slightly improved from 6/8 due to state sync and cost rated L, which they were in iteration 2 as well). No recalculation. |
| 10 | [blindspot] Vendor lock-in risk absent | Partially resolved | Constraint added: "后端 AI 代理层封装 provider-specific API 调用，业务逻辑不直接依赖特定供应商 SDK，支持供应商切换". This describes an architectural mitigation but does not add vendor lock-in as a named risk in the Risk table. The risk of switching cost is acknowledged in architecture but not tracked as a risk. |
| 11 | [blindspot] "产品负责人决定" undocumented | Not resolved | Quote remains: "Challenge Override: 产品负责人决定一次性交付完整范围以避免分阶段带来的用户体验割裂". No date, meeting notes, or written approval cited. |
| 12 | [blindspot] Degraded-mode NFR missing | Resolved | NFR added: "降级模式：AI 响应超过 10 秒未返回时，展示超时提示并提供传统表单快捷入口，确保用户操作不被阻断". Concrete threshold (10s) and fallback behavior specified. |

---

## Phase 1: Reasoning Audit

### Problem -> Solution Trace

| Problem Statement | Solution Element | Trace Quality |
|---|---|---|
| 创建 MainItem 需要手动填写 10+ 个字段，操作效率低 | AI 解析意图后推送预填表单卡片 | Direct mapping. Solved. |
| 用户需要理解状态机规则和权限体系才能正确操作 | AI 自动校验状态机合法性 + 错误提示合法目标状态 | Partially addressed. Reactive only — prevents errors but does not proactively guide users toward valid operations. A user who does not know what "进行中" means still does not know after the AI rejects an invalid transition. |
| 非技术人员填写结构化字段困难 | 对话输入 + AI 推断字段值 | Direct mapping. Solved. |

**Verdict**: Solution directly addresses the primary efficiency problem and the structured-field difficulty. The cognitive-burden problem (understanding RBAC and state machines) remains only partially addressed — the AI acts as a safety net but not as a teacher.

### Solution -> Evidence Trace

- Evidence footnote is now more honest: *"此为开发者走查估算，非经埋点分析的量化数据，后续需通过埋点数据验证"* — clearly labeled as a walkthrough estimate.
- The "8-12 clicks" figure remains the only quantitative metric.
- The `todos.txt` reference validates roadmap alignment, not user pain.
- No external validation (support tickets, analytics dashboards, user interview transcripts) has been added across three iterations.

**Verdict**: Evidence quality has improved in honesty (clear labeling of estimate vs. data) but not in depth. Three iterations without adding external validation suggests this is the ceiling without new data collection.

### Evidence -> Success Criteria Trace

- SC now covers all 6 entities with 16 testable items.
- Per-field accuracy definitions map to the "10+ fields" problem.
- New SC items (15, 16) fill gaps identified in iteration 2.

**Verdict**: Strong alignment. SC set is now comprehensive.

### Self-Contradiction Check

- `consistency_check_result` comment block (lines 219-228) confirms 4 prior reconciliations.
- Remaining tension: "全操作覆盖" (Innovation Highlights, line 43) vs. Out of Scope excluding delete and batch. "全操作覆盖" literally means "full operation coverage" — this overclaims given two explicit exclusions. Should be "核心操作覆盖" or list the covered operations.
- No new contradictions introduced in this iteration.

---

## Phase 2: Rubric Scoring

### 1. Problem Definition — 82/110

**Problem stated clearly (33/40)**:
Core problem is concrete and specific: "创建一个 MainItem 需要手动填写 10+ 个字段，操作效率低". The RBAC/statemachine problem is stated but not decomposed into specific failure modes.

Quote: *"用户需要理解状态机规则和权限体系才能正确操作，门槛较高"* — what specific failures result from this? Wrong transitions? Permission denied errors? Abandoned operations? The problem remains a generalization.

**Evidence provided (25/40)**:
Improvement from iteration 2: the estimate is now honestly labeled.

Quote: *"此为开发者走查估算，非经埋点分析的量化数据，后续需通过埋点数据验证"* — transparent about data quality.

However: "多名成员" remains unquantified. No support tickets, analytics, or external validation has been added across three iterations. The evidence base has reached a ceiling — it is honestly labeled but still entirely internal and self-reported.

**Urgency justified (24/30)**:
Argument is logical but unquantified:

Quote: *"延迟实施意味着团队持续承担低效操作的时间成本"* — no dollar or hour estimate for the cost of delay. No data on current team size, item creation volume, or time-per-creation.

---

### 2. Solution Clarity — 100/120

**Approach is concrete (38/40)**:
The 5-step interaction flow is specific. The card-as-single-source-of-truth model is clearly defined.

Quote: *"AI 解析意图后推送预填表单卡片到聊天窗口。必填且无法推导的字段留空，用户可通过直接编辑卡片或继续对话补充字段，确认后提交执行"* — clear and implementable.

Minor gap: The floating bubble's collapsed-state behavior (unread indicators, initial position) is still not described.

**User-facing behavior described (43/45)**:
Each scenario type has concrete input/output examples. Edge cases cover ambiguity, permission denial, statemachine violations, and unrecognizable intent. The intent taxonomy clarification is valuable.

Quote: *"意图分类（创建/查询/修改/分配）反映用户自然语言的表达习惯，而非后端 API 端点划分"* — good clarification.

Minor gap: No scenario for concurrent editing (user edits card while AI processes a follow-up).

**Technical direction clear (19/35)**:
Backend proxy architecture, structured output / tool use, and prompt construction are stated. Prompt size estimate added in this iteration.

Quote: *"估算 prompt 规模：6 个实体 schema × ~200 tokens/实体 + 系统指令 ≈ 2000-3000 tokens"* — concrete estimate, addresses prior attack.

However: the dual-input card architecture (the most novel technical element) remains one paragraph:

Quote: *"对话输入和直接编辑均通过统一的 dispatch 写入卡片状态，对话区域仅做展示回显，无需双向绑定"* — names the dispatch pattern but does not elaborate event flow, conflict resolution, or state shape. This has been flagged in two prior iterations without change.

---

### 3. Industry Benchmarking — 90/120

**Industry solutions referenced (35/40)**:
Six references: Linear, Notion AI, Slack Bot, Jira Automation, GitHub Copilot, Microsoft 365 Copilot. Good breadth across project management, productivity, and communication tools.

**At least 3 meaningful alternatives (24/30)**:
Four alternatives: Do nothing, embedded form assistant, Slack Bot, conversational + card hybrid. "Do nothing" is present.

Quote: *"Partial: 作为 MVP 方案对纯创建场景可行且开发成本更低，但本系统用户需要查询、状态变更、负责人调整等多类操作，嵌入式表单助手无法覆盖这些场景"* — the evaluation acknowledges MVP potential but dismisses it in a single sentence.

**Honest trade-off comparison (17/25)**:
The comparison table has pros/cons. "开发量较大" remains the only con for the selected approach — still not quantified.

Quote: *"开发量较大"* — vague. No effort estimate to compare against alternatives.

**Chosen approach justified against benchmarks (14/25)**:
The Verdict cell includes architectural reasoning for why Slack Bot and Command Palette do not fit:

Quote: *"本系统为单二进制嵌入式 Web 应用，无 Slack/Teams 集成场景"* — good justification.

However: the embedded form assistant alternative is still not fully evaluated. The "Partial" verdict dismisses it without analyzing whether an MVP delivering 60-70% of value at 30-40% of complexity might be a better first phase. This has been flagged in two prior iterations without substantive change.

---

### 4. Requirements Completeness — 92/110

**Scenario coverage (36/40)**:
Happy paths for all four intent types with concrete examples. Edge cases cover ambiguity, unrecognizable intent, permission denial, statemachine violation. Good coverage.

Remaining gaps:
- No scenario for concurrent editing (user edits card while AI processes follow-up).
- No scenario for network interruption during submission.
- No scenario for input exceeding the length limit mentioned in Risk #7.

**Non-functional requirements (38/40)**:
NFRs are specific and quantified: P95 latency < 5s, accuracy >= 85%/80%, per-field accuracy definitions with tolerances. Security, privacy, accessibility addressed. Degraded mode (10s timeout) now included.

Quote: *"降级模式：AI 响应超过 10 秒未返回时，展示超时提示并提供传统表单快捷入口"* — addresses prior blindspot.

Minor gap: No internationalization/multi-language NFR (all examples are Chinese, no mention of whether the AI must handle multilingual input).

**Constraints & dependencies (18/30)**:
Dependencies well-documented. available-transitions endpoint verification is excellent. Session constraints (50 turns, navigation clears) are concrete. Provider abstraction constraint added.

Quote: *"后端 AI 代理层封装 provider-specific API 调用，业务逻辑不直接依赖特定供应商 SDK，支持供应商切换"* — addresses vendor lock-in concern.

Remaining gaps:
- No constraint on supported browsers for the floating bubble.
- No mention of AI service rate limits or quota constraints from the provider side.
- No constraint on what happens to in-flight operations when the session clears on navigation.

---

### 5. Solution Creativity — 65/100

**Novelty over industry baseline (25/40)**:
The card-as-single-source-of-truth hybrid model is a genuine innovation over pure chatbot or pure form-fill.

Quote: *"不同于纯对话机器人或纯表单填充，采用卡片为中心的混合模式——卡片是唯一数据源（single source of truth）"* — clear differentiation.

However: the proposal acknowledges inspiration from Slack Bot + Notion AI patterns, and "context awareness" (page-level team detection) is standard practice. The novelty is in the interaction model combination, not fundamental technical innovation.

**Cross-domain inspiration (22/35)**:
References span project management, communication, and productivity tools. The Innovation Highlights now mention customer service chatbot patterns:

Quote: *"借鉴了客户服务聊天机器人（如 Intercom）在对话流中嵌入结构化操作卡片以降低错误率的实践"* — improvement, adds one cross-domain reference.

No inspiration from voice assistants (multi-turn context), game UIs (floating assistant patterns), or form design research (progressive disclosure).

**Simplicity of insight (18/25)**:
Core insight ("use cards for structured data, conversation for unstructured input") is clean and understandable. Implementation complexity (dual-input state management, prompt construction for 6 entity types, fallback mechanisms) may undermine the simplicity in practice.

---

### 6. Feasibility — 85/100

**Technical feasibility (36/40)**:
Strong case: existing APIs verified, tech stack standard, task bounded (4 intents x 6 entities). Prompt size estimate now provided:

Quote: *"估算 prompt 规模：6 个实体 schema × ~200 tokens/实体 + 系统指令 ≈ 2000-3000 tokens，在主流 AI 服务的上下文窗口限制内（如 Claude 200K、GPT-4 128K）"* — addresses prior attack.

PoC recommendation is concrete:

Quote: *"建议用 1 个工作日进行 PoC 验证——构造 20-30 条典型用户输入样本"* — actionable.

PoC not yet executed. Feasibility claims remain partially unvalidated.

**Resource & timeline feasibility (25/30)**:
Estimate of 5-7 weeks is reasonable. Ranges remain wide (2-3 weeks = 50% variance per track). The 1-week integration testing for 4 intents x 6 entities remains optimistic.

Quote: *"前端 2-3 周...后端 2-3 周...联调测试 1 周"* — ranges not tightened from iteration 1.

**Dependency readiness (24/30)**:
Existing APIs confirmed ready. Cost model now provided:

Quote: *"假设 20 活跃用户/天 × 平均 5 次 AI 调用/用户/天 × $0.015/调用（Claude Haiku 级别）≈ $45/月"* — concrete and testable.

AI service not yet selected: *"AI 服务选型建议在 tech-design 阶段确定具体供应商"*. The most critical external dependency remains uncommitted. No evaluation of provider rate limits or contractual requirements.

---

### 7. Scope Definition — 76/80

**In-scope items are concrete (28/30)**:
Items are deliverable-grade:

Quote: *"全局浮动气泡聊天 UI 组件（展开/收起、拖拽定位）"* — concrete.

Quote: *"支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool"* — explicit enumeration.

Minor issue: "与现有后端 API 对接，复用权限检查和业务逻辑" is vague — which endpoints specifically?

**Out-of-scope explicitly listed (24/25)**:
Nine items explicitly out of scope: AI training, delete, voice, multi-user, notifications, reports, cross-team search, conversation persistence, mobile responsive, batch operations. Good coverage.

**Scope is bounded (24/25)**:
Bounded by entity list (6), intent types (4), session limit (50 turns), and exclusions. Executable within 5-7 weeks.

Remaining tension: "全操作覆盖" in Innovation Highlights overclaims given explicit exclusions.

---

### 8. Risk Assessment — 78/90

**Risks identified (26/30)**:
Eight risks covering accuracy, latency, state sync, service availability, statemachine conflicts, privacy, prompt injection, and cost. Good breadth.

Missing: vendor lock-in risk (switching AI provider requires prompt/schema migration). The architectural mitigation is described in Constraints but not tracked as a named risk.

**Likelihood + impact rated (24/30)**:
External service unavailability correctly rated M. However, five of eight risks still rated M likelihood — the assessment lacks differentiation. The clustering suggests insufficient calibration.

Quote: every risk row has M or L likelihood. No risk is rated H likelihood despite this being a feature with an unvalidated external dependency.

**Mitigations are actionable (28/30)**:
Most mitigations are concrete and implementable:

Quote: *"AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作"* — actionable and user-centric.

Quote: *"设置每用户每日调用次数上限（硬上限防止失控）"* — concrete mechanism.

One mitigation is less actionable: "定义明确的意图类型和字段映射规则" — this is a design activity, not a runtime mitigation.

---

### 9. Success Criteria — 76/80

**Criteria are measurable and testable (28/30)**:
Most SC items are verifiable with concrete thresholds:

Quote: *"字段准确率 ≥ 80%"*, *"P95 延迟 < 5 秒"* — precise.

Per-field accuracy definition is testable:

Quote: *"日期字段在 ±1 天范围内计为正确；assignee 字段在 Team 成员模糊匹配范围内计为正确"* — unambiguous.

Minor ambiguity: SC item 1 says "至少包含标题 + 1 个额外字段" — sets a low bar. Is this the minimum pass criterion or the expected norm?

**Coverage is complete (22/25)**:
Significant improvement across iterations. 16 SC items now cover all 6 entities: MainItem (create + query + modify), SubItem (create + progress), Milestone (create), MilestoneMap (create), ProgressRecord (add), ItemPool (submit). New items cover drag positioning, message history, and disambiguation cards.

Remaining gap: No SC for the degraded mode NFR (10s timeout + traditional form shortcut). In Scope and NFR mention it, but no SC verifies the behavior.

**SC internal consistency (26/25 → capped at 25)**:
The 16 SC items are internally consistent. No pair of SC items is mutually exclusive. Clustering analysis found no contradictions.

---

### 10. Logical Consistency — 80/90

**Solution addresses the stated problem (31/35)**:
Form-filling inefficiency is directly addressed. The RBAC/statemachine comprehension problem remains partially solved — the solution prevents invalid operations reactively but does not proactively guide users.

Quote: *"状态变更不符合状态机规则 → 返回错误说明，提示合法的目标状态"* — reactive, not proactive.

**Scope <-> Solution <-> SC aligned (26/30)**:
SC now covers all 6 entities. Alignment between Scope and Solution is strong. New SC items (15, 16) fill prior gaps.

Remaining misalignment: "全操作覆盖" in Innovation Highlights implies comprehensive coverage, but scope excludes delete and batch. The overclaim creates ambiguity about what "full" means.

Quote: *"全操作覆盖：不仅限于创建，还支持查询、状态变更、负责人调整等核心操作"* — "全操作覆盖" followed by "核心操作" is contradictory. "全" means "all", "核心" means "core". These are different claims.

**Requirements <-> Solution coherent (23/25)**:
Requirements map cleanly to the solution. Each scenario type has a corresponding solution component. NFRs map to architectural decisions.

Minor gap: Entity resolution strategy (exact -> fuzzy -> disambiguation card) is described in Constraints and now has a corresponding SC (item 16) — alignment improved.

---

## Phase 3: Blindspot Hunt

1. **[blindspot] "全操作覆盖" vs "核心操作" — internal contradiction within Innovation Highlights**: Quote: *"全操作覆盖：不仅限于创建，还支持查询、状态变更、负责人调整等核心操作"* — "全操作覆盖" (full coverage) is the heading, but the description lists "核心操作" (core operations). These are different scopes. "全" implies completeness; "核心" implies selectivity. If delete and batch are excluded, it is not "全操作覆盖". Replace with "核心操作覆盖" or list the covered operations explicitly.

2. **[blindspot] Cost model assumes Haiku-level pricing but accuracy target assumes Sonnet-level capability**: Quote: *"假设 20 活跃用户/天 × 平均 5 次 AI 调用/用户/天 × $0.015/调用（Claude Haiku 级别，structured output）≈ $45/月"* vs. the accuracy target *"AI 意图识别准确率 ≥ 85%"*. Claude Haiku is a fast/cheap model. The proposal acknowledges achieving 90%+ accuracy is based on "已验证在同类型项目管理工具中" but this validation likely used a more capable model. There is an unstated assumption that the cheapest model can meet the most demanding accuracy target. If Haiku fails the accuracy bar and Sonnet is needed, the cost jumps from $45/month to $150-200/month — a 3-4x increase that is mentioned as a possibility but not analyzed as the likely scenario.

3. **[blindspot] Vendor lock-in mitigation described in architecture but absent from risk table**: Quote: *"后端 AI 代理层封装 provider-specific API 调用，业务逻辑不直接依赖特定供应商 SDK，支持供应商切换"* (Constraints). This describes an architectural pattern that mitigates vendor lock-in, but "vendor lock-in" is not listed as a named risk in the Key Risks table. An architectural mitigation without a corresponding risk entry means the risk is not tracked or monitored. If the abstraction layer is insufficient or degrades over time, nobody is watching.

4. **[blindspot] PoC recommendation exists but has no gate criteria tied to project continuation**: Quote: *"建议用 1 个工作日进行 PoC 验证...若准确率低于 70%，需重新评估 prompt 策略或调整范围后再决定是否继续"*. The PoC has a 70% gate, but the NFR target is 85% accuracy. The gap between 70% (PoC pass) and 85% (NFR target) is unaddressed. What happens if PoC yields 72%? The proposal says "重新评估" but does not define whether the project continues, pauses, or pivots. The PoC gate should align with the NFR target or have an explicit decision framework.

5. **[blindspot] Session clearing on navigation may cause user frustration**: Quote: *"页面导航时清除会话"* — if a user is mid-conversation (e.g., has a partially filled card), navigating to a different page destroys their work. No warning, no confirmation, no auto-save. This is a UX risk not addressed in the risk table or edge cases.

6. **[blindspot] No SC for degraded-mode behavior**: NFR section states: *"AI 响应超过 10 秒未返回时，展示超时提示并提供传统表单快捷入口"*. This is an important user-facing behavior but has no corresponding SC entry. In Scope includes it, NFR specifies it, but SC does not verify it.

---

## Summary

| # | Dimension | Score | Max |
|---|-----------|-------|-----|
| 1 | Problem Definition | 82 | 110 |
| 2 | Solution Clarity | 100 | 120 |
| 3 | Industry Benchmarking | 90 | 120 |
| 4 | Requirements Completeness | 92 | 110 |
| 5 | Solution Creativity | 65 | 100 |
| 6 | Feasibility | 85 | 100 |
| 7 | Scope Definition | 76 | 80 |
| 8 | Risk Assessment | 78 | 90 |
| 9 | Success Criteria | 76 | 80 |
| 10 | Logical Consistency | 80 | 90 |
| | **Total** | **824** | **1000** |

---

## ATTACKS (Prioritized)

1. **[D2: Solution Clarity] Dual-input card event flow remains underspecified after three iterations**: Quote: *"对话输入和直接编辑均通过统一的 dispatch 写入卡片状态，对话区域仅做展示回显，无需双向绑定"* — this is the most technically novel element of the proposal but has been described in a single paragraph across all three iterations. No event flow, no conflict resolution (user types in card while AI processes a follow-up message that also modifies the same field), no state shape. Add at minimum a sequence description: (1) user action -> (2) dispatch event -> (3) state update -> (4) re-render. Specify what happens when concurrent modifications conflict.

2. **[D3: Industry Benchmarking] Embedded form assistant alternative remains unevaluated after two flags**: Quote: *"Partial: 作为 MVP 方案对纯创建场景可行且开发成本更低，但本系统用户需要查询、状态变更、负责人调整等多类操作，嵌入式表单助手无法覆盖这些场景"* — the embedded form assistant is dismissed in one sentence without analyzing its MVP potential. The "Challenge Override" says "产品负责人决定一次性交付完整范围" but no documented rationale is provided for why a phased approach would cause "用户体验割裂". Evaluate the embedded approach as a standalone alternative with its own pros/cons table, or provide documented evidence that users require all four intent types from day one.

3. **[D1: Problem Definition] Evidence ceiling reached — no external validation across three iterations**: Quote: *"此为开发者走查估算，非经埋点分析的量化数据，后续需通过埋点数据验证"* — the proposal has improved honesty about evidence quality but has not added any external validation across three iterations. "多名成员" remains unquantified. No support tickets, analytics data, or user interviews cited. Accept the current evidence as the best available and note this limitation in the proposal, or defer the proposal pending data collection.

4. **[D10: Logical Consistency] "全操作覆盖" contradicts scope exclusions**: Quote: *"全操作覆盖"* (Innovation Highlights line 43) vs. *"删除操作...第一版不支持"* (Out of Scope) and *"批量操作...Out of Scope"*. "全" means "all" — claiming "all operations" while excluding two operation types is a contradiction. Replace with "核心操作覆盖" or enumerate the covered operations.

5. **[D6: Feasibility] Cost model assumes cheapest model meets hardest accuracy target**: Quote: *"$0.015/调用（Claude Haiku 级别，structured output）≈ $45/月"* vs. *"AI 意图识别准确率 ≥ 85%"*. The cost model uses Haiku pricing ($0.015/call) but the accuracy validation likely used a more capable model. If Sonnet is needed, cost jumps 3-4x. The proposal mentions this possibility but does not analyze the probability or plan for it. Add a decision point: if PoC with Haiku < 85%, escalate to Sonnet and recalculate budget.

6. **[D8: Risk Assessment] Vendor lock-in risk absent from risk table despite architectural mitigation**: Quote: *"后端 AI 代理层封装 provider-specific API 调用...支持供应商切换"* (Constraints) — an architectural mitigation exists but no corresponding risk is tracked. Add "Vendor lock-in / provider switching cost" as a named risk with likelihood and impact assessment.

7. **[D9: Success Criteria] No SC for degraded-mode behavior**: Quote: *"AI 响应超过 10 秒未返回时，展示超时提示并提供传统表单快捷入口"* (NFR) — this important user-facing behavior has no SC entry to verify it works. Add an SC item: "When AI response exceeds 10 seconds, system displays timeout message with link to traditional form within 1 second of timeout."

8. **[blindspot] PoC gate (70%) misaligned with NFR target (85%)**: Quote: *"若准确率低于 70%，需重新评估"* vs. *"AI 意图识别准确率 ≥ 85%"*. The gap between 70% (PoC gate) and 85% (NFR) is unaddressed. Define what happens for results in the 70-85% range: continue with remediation plan, pause for prompt engineering, or reduce scope?

9. **[blindspot] Session clearing on navigation destroys in-progress work**: Quote: *"页面导航时清除会话"* — if a user has a partially filled card and navigates to another page (even accidentally), their work is lost with no warning, no confirmation, and no auto-save. This is a UX risk not covered in the risk table or edge cases. Add an edge case or NFR for in-progress session protection on navigation.

10. **[D8: Risk Assessment] Five of eight risks rated M likelihood — insufficient differentiation**: The risk table shows Accuracy M, Latency M, Service down M, State machine M, Privacy M, Injection M — six risks at M likelihood. Cost and state sync are L. No risk is rated H despite an unvalidated external dependency being the most critical factor. Differentiate at least 2-3 risks based on probability reasoning (e.g., AI accuracy issues are likely H given unvalidated PoC; prompt injection is likely L given the structured user base).

11. **[blindspot] "产品负责人决定" remains undocumented after three iterations**: Quote: *"Challenge Override: 产品负责人决定一次性交付完整范围以避免分阶段带来的用户体验割裂"* — no date, no meeting notes, no written approval cited. A scope-level override that rejects a "5 Whys" challenge should be traceable to a documented decision. Add a reference to the decision artifact (meeting notes, email, ticket) or acknowledge that the decision is undocumented.
