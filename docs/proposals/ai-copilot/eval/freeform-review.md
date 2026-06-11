---
reviewer: "AI Conversational PM Tool Architect"
doc: "docs/proposals/ai-copilot/proposal.md"
date: "2026-06-11"
type: freeform
---

# Freeform Expert Review: AI Copilot 对话助手 Proposal

## Background Assessment

This proposal introduces a conversational AI assistant embedded into an existing project-management web application. The core idea is sound and well-grounded: the current form-based interaction model requires users to navigate 10+ fields to create a MainItem, and a natural-language interface combined with structured pre-filled cards could meaningfully lower the operational barrier. The hybrid chat-plus-card pattern is a thoughtful choice -- it avoids the unreliability of pure conversational agents by grounding AI output in editable, structured form cards that the user must explicitly confirm before any mutation is executed.

The proposal demonstrates awareness of the most critical architectural constraint: the AI layer must sit on top of existing API endpoints, RBAC permission checks, and state-machine validation rules, never bypassing them. The six entity types (MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool) and four intent categories (create, query, modify, assign) are a reasonable scope boundary for an initial AI integration. The industry benchmarking section correctly identifies that no existing tool offers the exact combination of natural-language input plus structured-card confirmation for domain-entity operations, which strengthens the case for building rather than adopting an off-the-shelf solution.

That said, the proposal operates at a high level of abstraction that leaves several critical design questions unanswered. As someone who has built similar systems and lived through the failure modes of AI-assisted CRUD operations, I see significant gaps in the proposal's treatment of the intent-to-API mapping, the card-dialog synchronization model, the AI service integration boundary, and the failure-degradation design. These gaps are not fatal, but they must be resolved before this proposal matures into a technical design.

## Key Risks

### Intent-to-API Mapping Gaps

The proposal defines four intent types (create, query, modify, assign) and six entity types, but it does not address how these compose when a single user utterance maps to multiple entities or multi-step operations.

风险：The scenario "在XX事项下加一个子任务" requires the AI to first resolve a parent MainItem reference (which could be ambiguous -- by title? by bizKey? by partial match?), then construct a SubItem creation request with the correct parent reference. The proposal states the AI needs to "识别 parent MainItem" but provides no detail on how entity resolution works when the user's reference is imprecise. In my experience, entity resolution is where conversational AI assistants fail most often -- users say "那个认证模块的事项" and the system cannot uniquely identify which MainItem they mean. The proposal should define a resolution strategy: exact match by bizKey, fuzzy match by title within Team scope, or interactive disambiguation when multiple candidates exist.

问题：The proposal lists "assign" as a separate intent type, but in the existing system, assigning a person to a MainItem is a field update on the entity (changing the `assignee` field), not a distinct API operation. Similarly, "把用户认证模块的状态改为进行中" is classified under "修改场景" but actually requires a dedicated status-transition API call that must pass through the state-machine validation defined in `status-machine.md`. The intent taxonomy conflates field-level updates with domain-specific operations that have their own validation rules. This conflation risks the AI layer generating the wrong API call structure for operations that look similar at the natural-language level but invoke fundamentally different backend endpoints.

### Card-Dialog Synchronization

The "Innovation Highlights" section and the Key Scenarios both describe a bidirectional editing model, but the actual design guidance in the Risk table contradicts this.

问题：The proposal states as a key innovation: "对话 + 卡片双向编辑：不同于纯对话机器人或纯表单填充，采用混合模式——AI 推送结构化卡片，用户可以对话补充字段也可以直接编辑卡片，卡片实时同步更新." But then in the risk table, the mitigation for "卡片 + 对话双向编辑状态同步复杂" is: "设计明确的单向数据流：对话修改 → 更新卡片状态；卡片编辑 → 同步到对话显示。避免双向绑定的复杂性." These two statements contradict each other. The innovation claims bidirectional real-time sync; the risk mitigation admits it should be unidirectional to avoid complexity. Which is the actual design intent? If the answer is a single-source-of-truth model (e.g., card state is truth, dialog is display-only for history), then the "innovation highlight" is misleading and should be reworded to reflect the actual pattern: card-centric editing with dialog-driven field updates that flow one way into the card.

风险：Without resolving this ambiguity before tech design, the implementation team is likely to build a two-way binding system that produces state conflicts. Consider the scenario: user edits the `priority` field directly on the card from P1 to P2, then immediately types in the chat "优先级改成 P0." Which value wins? If the card is truth, the dialog should read P2 before applying the P0 change. If they are truly bidirectional, a race condition exists. The proposal must specify the conflict-resolution semantics before any code is written.

### AI Service Integration Boundary

The proposal acknowledges dependency on an external AI service but treats this as a black box with no architectural specificity.

风险：The constraint section states "依赖外部 AI 服务进行自然语言理解和意图识别" and the Feasibility Assessment notes "意图识别和实体抽取是成熟 AI 能力，风险可控." This understates the engineering challenge. The proposal does not specify whether the AI service is called from the frontend (direct client-side API call) or proxied through the backend. If called from the frontend, the AI service API key is exposed in the browser. If proxied through the backend, the backend must now handle streaming responses, manage AI-service authentication tokens, and absorb the latency of an additional network hop. The proposal should explicitly commit to a backend-proxy architecture and define the API contract between the frontend chat component and the backend AI-orchestration layer.

问题：The proposal does not address the prompt-engineering boundary. Who constructs the system prompt that instructs the AI model on the available intents, entity schemas, and field constraints? Is this prompt static (hardcoded at deploy time) or dynamic (assembled per-request from the current Team's schema and the user's permissions)? A static prompt will drift out of sync with backend schema changes. A dynamic prompt adds latency and complexity but stays accurate. This decision fundamentally shapes the architecture and should be surfaced in the proposal.

### Confirmation Gate Integrity

The proposal correctly requires user confirmation for all write operations, but the confirmation semantics are underspecified.

问题：The success criteria state: "所有写操作均需用户在卡片上确认后提交，无绕过确认的直接执行路径." This is the right principle, but the proposal does not define what "confirm" means architectically. Is there a single "Submit" button on the card? Does each field have individual confirm/cancel actions? When the user edits a card and then types additional dialog commands that modify the same card, is a single confirmation sufficient, or does each modification cycle require re-confirmation? The proposal should define the confirmation UX as a card-level action with a clear submit/cancel lifecycle.

风险：The scenario "操作涉及权限不足 → 返回权限提示，不推送卡片" suggests the AI layer performs RBAC checks before rendering the card. But if the AI service is external (as implied), it has no access to the RBAC layer. This means the backend must pre-validate the operation before the AI generates the card, or the card must include a server-side validation step before rendering. The proposal should specify whether permission checks happen at card-generation time (pre-validation) or at card-submission time (post-validation). Pre-validation prevents users from seeing cards they cannot submit, which is a better UX; post-validation is architecturally simpler but produces frustrated users who fill out cards only to have them rejected.

### State-Machine Validation Timing

问题：The risk table mitigation for "状态机规则与 AI 推理冲突" states: "AI 推送卡片前先调用 available-transitions API 校验合法性，不合法操作直接拒绝并提示." This is the correct approach -- pre-validate before rendering the confirmation card. But this requires the backend to expose an `available-transitions` endpoint for each entity type. The proposal does not confirm whether such endpoints currently exist in the API surface. If they do not, the proposal has introduced an unscoped backend requirement (building new introspection endpoints) that is not reflected in the feasibility assessment or resource estimates.

风险：The status-machine.md convention documents complex inter-entity rules: Milestone transitions are blocked when the parent MilestoneMap is terminal (BR-5), and Milestone completion requires all related MainItems to be terminal (BR-1). If the AI suggests a Milestone status change without checking the parent MilestoneMap's state, the pre-validation API call will reject it, but the rejection message must be specific enough to explain why ("cannot transition this Milestone because its parent MilestoneMap is already completed"). The proposal's error UX only specifies "返回错误说明，提示合法的目标状态" which is insufficient for these multi-entity business rules.

### Non-Functional Realism

风险：The proposal sets two latency targets: "AI 响应（从用户发送到卡片推送）应在 3 秒内完成" and success criteria "AI 响应（从发送到卡片推送）P95 延迟 < 5 秒." These targets are inconsistent -- the NFR says 3 seconds, the success criteria says P95 under 5 seconds. More importantly, neither target accounts for the full round-trip chain: frontend sends user message to backend, backend constructs prompt, backend calls external AI service, AI service returns intent+entities, backend calls available-transitions API for pre-validation, backend calls RBAC check, backend returns card payload to frontend, frontend renders card. Each of these steps adds latency. The external AI service call alone can easily take 2-4 seconds for a moderately complex prompt. If the backend must also call available-transitions and permission-check endpoints before returning the card, the 3-second target is unrealistic without aggressive parallelization or caching strategies that the proposal does not mention.

问题：The accuracy targets ("AI 意图识别准确率 ≥ 85%，字段提取准确率 ≥ 80%") are stated without defining how accuracy is measured. Is this per-utterance accuracy (did the AI pick the right intent) or per-field accuracy (did it extract the correct value for each field)? How will the system measure this in production? Without a measurement methodology, these targets are aspirational claims that cannot be validated or iterated upon.

### Degradation Strategy

风险：The risk table mitigation for external AI service unavailability is: "设计降级策略：AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作." This is the minimum viable degradation, but it wastes a valuable UI surface. When the AI service is down, the chat panel could still serve as a shortcut interface: the user types "创建事项" and the system opens the traditional creation form pre-positioned in the chat panel, without AI involvement. This provides partial utility during outages rather than rendering the entire chat panel useless. The proposal should consider a tiered degradation model rather than a binary available/unavailable switch.

### Scope and Entity Complexity

问题：The In Scope section lists six entity types: "MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool." Each entity has a different schema, different required fields, different state machines, and different RBAC rules. The proposal treats this as a single scope item, but the AI must be trained (via prompts or fine-tuning) to understand the schema of each entity. This multiplies the prompt complexity and the validation surface area. The proposal should acknowledge that supporting six entity types in the initial release is ambitious and consider whether a phased rollout (e.g., MainItem + SubItem first, then Milestones, then the rest) would reduce risk.

## Improvement Suggestions

建议：The proposal should resolve the bidirectional vs. unidirectional card-dialog sync contradiction by committing to a single-source-of-truth model. Specifically, state that the card is the source of truth for field values, dialog-driven field updates write into the card state, and the chat history displays what was changed but is not itself an editable data source. This eliminates the ambiguity in the current text where the innovation section claims "卡片实时同步更新" without specifying the direction.

建议：Add an explicit architecture decision: all AI service calls must be proxied through the backend, never called directly from the frontend. This protects the AI service API key, enables server-side prompt construction that can incorporate RBAC and state-machine context, and allows the backend to log all AI interactions for debugging and accuracy measurement. The proposal should state this as a constraint under "Constraints & Dependencies."

建议：Define the entity-resolution strategy for ambiguous references. When a user says "把认证模块的状态改了," the system must decide: is there exactly one MainItem whose title contains "认证模块"? If multiple matches exist, does the AI return a disambiguation card listing the candidates? If no match exists, does it ask the user to clarify? This strategy should be documented in the proposal as part of the core interaction flow, not deferred to tech design.

建议：Clarify the confirmation gate architecture: permission checks and state-machine validation should happen at card-generation time (before the card is shown to the user), not at card-submission time. This means the backend must call the relevant validation APIs during the AI-response round-trip. Acknowledge that this adds latency but prevents the UX anti-pattern of users filling out cards they cannot submit. The latency budget should be revised to account for these validation calls.

建议：Unify the two latency targets. Choose a single target (e.g., "P95 < 5 seconds end-to-end, including pre-validation calls") and decompose it into a latency budget: external AI service call < 3s, backend pre-validation < 500ms, network overhead < 500ms, frontend rendering < 500ms, buffer < 500ms. This makes the target verifiable and gives the implementation team concrete sub-targets to design against.

建议：Define how accuracy will be measured in production. At minimum, log each AI interaction with the user's original utterance, the AI's predicted intent and extracted fields, and whether the user confirmed, edited, or abandoned the card. An "edited" card indicates the AI's initial extraction was partially wrong. An "abandoned" card indicates it was completely wrong. This logging enables continuous accuracy measurement without manual labeling.

建议：Consider a tiered degradation model for AI service unavailability: Tier 1 (full AI) when the service is available; Tier 2 (template mode) when the AI service is down but the backend can still match utterances against keyword patterns to open the correct form type; Tier 3 (message-only) when even the backend is degraded. This provides a smoother user experience than the current binary available/dead model.

建议：The proposal should scope the first release to fewer entity types. Start with MainItem and SubItem (the two most commonly used entities), validate the full intent-to-card-to-submission pipeline end-to-end, and then expand to Milestone, MilestoneMap, ProgressRecord, and ItemPool in subsequent iterations. This reduces the prompt-engineering surface area and the number of state-machine rules the AI must handle correctly at launch.
