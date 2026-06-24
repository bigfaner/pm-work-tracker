# PRD Eval — Iteration 1 (Senior QA Engineer)

Iteration: 1
Expert persona: Senior QA Engineer
Mode: A (prd-ui-functions.md present)
Target score: 900/1000

Documents reviewed:
- `docs/features/ai-copilot/prd/prd-spec.md`
- `docs/features/ai-copilot/prd/prd-user-stories.md`
- `docs/features/ai-copilot/prd/prd-ui-functions.md`

Reality-check sources consulted:
- `backend/internal/handler/router.go` (route existence for `available-transitions`)
- `frontend/src/App.tsx` (route existence for sitemap claims)

---

## Phase 1 — Reasoning Audit

### Problem → Solution → Evidence → Success Criteria trace

- **Problem** is concrete and quantified (8-12 clicks for MainItem create, 10+ fields, complex state machine + RBAC). Strong.
- **Solution** (conversational + card hybrid, card as single source of truth) addresses the problem directly. Internally consistent.
- **Evidence** is weak: every quantitative baseline is hedged — "（开发者走查估算，待埋点验证）", "（设计假设，待上线后校准）". The Goal metrics (≤3 interactions, P95 < 5s, ≥85% accuracy) cannot be falsified today because baselines are not measured. Acceptable for a Draft, but it weakens "logical consistency" between Goals and Evidence.
- **Success Criteria** mapping: In-Scope items map cleanly to UI Functions (UF-1..UF-7) and to user stories. Good vertical traceability.

### Self-contradiction scan

Found one material inconsistency between spec and reality (see Blindspot B1): spec claims `available-transitions` is the universal pre-validation mechanism ("该端点已存在于 MainItem/SubItem/MilestoneMap/Milestone，无需新增" — Related Changes #3) while In-Scope and Stories cover `ProgressRecord` and `ItemPool` writes that have no such endpoint. This is an internal contradiction between In-Scope coverage list and Related Changes capability claim.

### SC Consistency Deep-Dive

Clustering Success Criteria + In Scope by affected area:

| Area | Success Criteria / In Scope entries | Coherence check |
|------|------|------|
| Write pre-validation | In-Scope "状态机预校验：写操作前校验合法性", Story 2 AC#4, UF-3 Validation "状态变更类操作必须先通过 available-transitions 预校验才允许提交" | Bidirectional derivation OK for MainItem/SubItem/Milestone/MilestoneMap; FAILS for ProgressRecord and ItemPool which lack the endpoint. |
| Timeout & fallback | In-Scope "降级模式：AI 超 10 秒未返回", Story 6 AC#1, UF-6 | Consistent. |
| Disambiguation | In-Scope "歧义消解", Story 5, UF-5 | Consistent. |
| Card as source of truth | In-Scope "预填表单卡片组件…唯一数据源", Story 1 AC#2, UF-3 Description | Consistent wording. |
| Permission | In-Scope "复用现有 API 权限检查", Story 4 AC#4, UF-3 "权限不足时不渲染可提交卡片" | Consistent — but no story covers permission failure on a modify/assign op, only on ItemPool create. Asymmetric coverage. |

Attack point 1: ProgressRecord/ItemPool pre-validation gap.
Attack point 2: Asymmetric permission-failure coverage (only one entity, one role).

---

## Phase 2 — Rubric Scoring

### 1. Background & Goals (100 pts) → 80/100

| Criterion | Score | Notes |
|-----------|-------|-------|
| Three elements present & specific | 28/30 | Reason / Target / Users all present and specific. Users table has 4 roles with concrete high-frequency ops. Minor: "利益相关方" introduces system admin & security/compliance roles that have NO user story coverage later — light inconsistency but within this dimension it's just completeness. |
| Goals quantified | 22/30 | Numeric targets exist (≤3 clicks, ≥85% / ≥80% accuracy, P95 <5s). BUT three of five metrics are explicitly hedged with placeholders: "（通过对话一句指令 + 卡片确认；待埋点验证基线）", "（设计假设，待上线后校准）". Rubric deducts -20 per vague/unquantified instance. Two baseline-pending metrics cost ~8 pts. |
| Background ↔ Goals logically consistent | 30/40 | Goals follow from the stated problem (interaction cost → ≤3 clicks). The deduction is for the gap between "提升非技术人员可用性 … 定性目标，通过用户反馈验证" (qualitative, unverifiable) and the otherwise quantitative table — and for "覆盖核心操作 … 6 个实体" being listed as a Goal when it is actually a Scope item (a coverage target is a deliverable, not a measurable goal). |

### 2. Flow Diagrams (150 pts) → 115/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mermaid diagram exists | 50/50 | Present, syntactically valid. |
| Main path complete | 35/50 | Start → Input → Proxy → Parse → Disambig → Card → Precheck → Submit → Feedback covers the happy path. But the diagram's Resolve node says "精确匹配 bizKey → 模糊匹配标题" — there is NO branch for the case where **neither** matches (entity not found). The flowchart silently assumes at least one candidate always exists. Also, the Query branch (`M3 →|否 查询| Query`) does not show what happens when a query returns zero results — UF-4 has an "无结果" state but the diagram omits it. |
| Decision points + error branches | 30/50 | Four decision diamonds present (M1-M4). Error branches: only the timeout dashed edge and the `M4 →|否| Err → Edit` retry loop. MISSING branches: (a) AI service unavailable (mentioned in 异常分支 text but NOT in the mermaid), (b) permission denied (mentioned in 异常分支 text but NOT in mermaid), (c) input exceeds max length (mentioned in 异常分支 text but NOT in mermaid), (d) per-user daily quota exceeded (mentioned in Security but neither in 异常分支 text nor in mermaid). The prose lists 4异常 branches; the diagram implements 1. -20. |

### 3. Functional Specs (200 pts, evaluates prd-ui-functions.md) → 145/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Placement & Interaction completeness | 50/70 | Every UF has Placement (good). Interaction flows exist for all 7 UFs. Deductions: UF-3 "User Interaction Flow" describes a pre-check + submit sequence but does not state what happens when the user edits a field AFTER a previous pre-check failure and the new value is still invalid — is the card locked, re-validated live, or re-checked on submit? UF-5 interaction flow is silent on what happens if the user closes/abandons the disambiguation card without choosing (no "cancel" state in States table). UF-6 interaction flow does not specify whether the in-flight AI request is cancelled when the user clicks the fallback entry or whether it can still return later and confuse the panel state. |
| Data Requirements & States clarity | 55/70 | Field tables and state tables are populated with explicit sources/triggers. Deductions: UF-2 "发送状态" enum includes `streaming` but no State row in the States table is labeled `streaming` consistently — there is a "流式返回" state, so naming is inconsistent across the two tables for the same concept (could confuse implementers). UF-3 "字段集" type is `Map<fieldName, {value, required, derived}>` — `derived` boolean is never defined (derived from what? page context? AI? both?). UF-4 "实体卡片列表" has no upper bound on count; Validation says "结果过多时分页或限制展示数量" — "过多" and "限制" are not quantified. |
| Validation Rules explicit | 40/60 | Most rules are actionable ("feature flag 关闭时气泡完全不渲染", "必填字段为空时提交按钮禁用并提示"). Deductions: UF-1 "气泡不得遮挡页面主操作按钮（需与现有页面元素做避让）" — "主操作按钮" and "避让" are not actionable without a definition. UF-2 "单次输入超过最大长度时截断并提示" — the maximum length value is never given anywhere in the document. UF-3 "日期字段 ±1 天内视为与意图一致（影响准确率统计，不影响提交）" — interaction between accuracy stats and submit is interesting but the rule does not say whether the AI-pre-filled date that is >1 day off is treated as a validation failure or silently accepted. UF-5 "候选列表限定在当前 Team 范围内" — does not state the cap on candidate count when many matches exist. |

### 4. User Stories (200 pts) → 135/200

| Criterion | Score | Notes |
|-----------|-------|-------|
| Coverage: one story per target user | 40/50 | All 4 background roles have at least one story (PM S1, Dev S2, TL S3, ItemPool S4). Cross-role stories S5/S6 cover shared concerns. Deduction: "系统管理员（配置 AI 服务与灰度开关）" and "安全/合规（数据隐私审计）" are listed as stakeholders in Background but have NO user story. -10. |
| Format correct (As a / I want / So that) | 45/50 | All stories follow format. Actions are mostly concrete. Minor: Story 3 "I want to 通过自然语言查询事项进度、创建里程碑图、调整负责人分配" bundles three operations in one story — weak single-responsibility, harder to verify independently. |
| AC per story in Given/When/Then | 45/50 | Every story has 3-4 ACs in correct format. Deduction: Story 1 AC#4 ("Given PM 输入'加个里程碑…'") is structurally Given/When/Then but the "Given" is actually a new trigger, not a precondition carrying state from the previous AC — the AC chain reads as separate scenarios glued together, slightly muddying test traceability. |
| AC verifiability & boundary coverage | 5/50 | MAJOR deductions as QA. Many "Then" clauses are not objectively verifiable or do not cover boundaries: <br>• Story 1 AC#1 Then "系统在 P95 < 5 秒内推送预填卡片" — P95 is a population statistic, a single observation cannot verify P95. Should be "≤ 5 秒" for an individual call, with P95 measured at population level elsewhere. <br>• Story 1 AC#1 "milestoneKey=留空高亮" — "高亮" is a UI detail, not objectively verifiable without a CSS spec; in UI Functions layer this is fine, but at the Story/AC layer it is subjective. <br>• Story 2 AC#4 Then "返回错误说明并列出合法目标状态" — does not specify the format/structure that QA can assert on. <br>• Story 5 AC#1 Then "推送歧义消解卡片列出所有候选实体（标题 + 编号）" — "所有候选" has no upper bound; if 200 items match, is the card still "列出所有"? Untestable boundary. <br>• Boundary coverage is thin: NO AC covers (a) max input length exceeded, (b) per-user daily quota hit, (c) confidence-score-below-threshold behavior, (d) AI returning malformed/incomplete extraction, (e) concurrent edits on the same card from dialog and direct edit, (f) user abandoning a card mid-edit, (g) Team context missing (only mentioned in UF-2 Validation, no story), (h) prompt-injection input blocked. These are exactly the failure paths a Senior QA would file bugs against. -45. |

### 5. Scenario Completeness (150 pts) → 95/150

| Criterion | Score | Notes |
|-----------|-------|-------|
| End-to-end scenario coverage | 35/60 | Write happy path is covered end-to-end (Story 1). Query is covered (Story 3). Modify/assign partially covered (Story 3 AC#4 for assign; Story 2 AC#4 for state-change failure). MISSING end-to-end scenarios: <br>• ProgressRecord update from creation → submit → success (Story 2 AC#3 stops at the card push, no submit-then-success AC). <br>• ItemPool submission end-to-end success (Story 4 stops at "两种方式均写入同一卡片状态" — no submit-and-confirm AC). <br>• Milestone / MilestoneMap creation submit + result feedback (Story 1 AC#4 ends at "推送 Milestone 创建卡片" — no confirmation AC). <br>So 3 of the 6 supported entities have NO end-to-end "submit → success" scenario. -25. |
| Implicit assumptions surfaced | 30/40 | Team context is surfaced ("Team 上下文自动检测（跟随当前页面）并明确展示给用户"). Deductions: <br>• "用户权限范围" is passed to the AI prompt (DF002) but no scenario describes behavior when AI proposes an entity the user cannot see/modify (cross-team leak). <br>• "单会话最大 50 轮" is enforced (UF-2 Validation) but no scenario describes what the user sees when they hit 50 mid-task with an unsubmitted card. <br>• Confidence threshold for "意图是否被识别" (decision M1) is never quantified — implicit assumption that there is a binary cutoff. -10. |
| Business-rules consistency | 30/50 | Loaded business rules from injected context. Issues: <br>• BIZ-milestone-001/002 (terminal items sort/filter behavior) — not referenced anywhere; if AI proposes a milestone→completed transition on a Milestone whose MainItems are not all terminal, the spec's pre-check ("available-transitions") is correct but no scenario or AC explicitly demonstrates this guard. Silent. <br>• BIZ-milestone-003 (MainItem milestone_key update checks MainItem status FIRST then Milestone status) — the spec only describes a generic "状态机预校验" without acknowledging this two-step ordering; QA cannot tell whether the error message will reveal which check failed. <br>• BIZ-lifecycle-004 (sub-item move target must be same team, non-terminal, different main item; 400 on violation) — there is NO user story covering "move sub-item" via Copilot, yet "修改" is one of the four supported operations. This is a coverage gap relative to the stated operation set. <br>• BIZ-filter-001 (assignee filter returns direct + indirect matches) — relevant to Story 3 query scenarios; the spec does not say whether Copilot queries honor the same direct+indirect semantics, risking inconsistent results between Copilot and the existing UI. -20. |

### 6. Edge Case Coverage (100 pts) → 50/100

| Criterion | Score | Notes |
|-----------|-------|-------|
| Error paths documented | 20/40 | Documented: timeout (>10s), AI unavailable, permission denied (one story), state-machine violation (one story), max input length (mentioned, value missing). UNDOCUMENTED error paths: <br>• AI returns malformed JSON / unrecognized intent structure (the spec assumes AI always returns a parseable intent or "not recognized"; no handling for parse failure). <br>• AI proposes an entity bizKey that does not exist or belongs to another Team (cross-Team data leak via prompt). <br>• Per-user daily quota exceeded (mentioned in Security but no UX behavior defined — does the user see an error? Silent degrade to keyword match? The spec says "降级为关键词匹配模式" under Monitoring but that's a backend mode, not user-facing). <br>• Available-transitions endpoint itself fails (network error, 500) — pre-check is now a single point of failure for the whole write flow. <br>• Existing CRUD API failure on submit (UF-3 has a "失败" state but no story / AC covers what the user does next). <br>-20. |
| Boundary conditions covered | 15/35 | Covered: 50-round session cap, max input length (value missing). MISSING: <br>• Empty input (user sends whitespace or only punctuation) — no behavior defined. <br>• Very long single message (just under max) — performance/UX behavior undefined. <br>• Candidate list size in disambiguation (no cap). <br>• Query result size (UF-4 says "限制展示数量" without a number). <br>• Concurrent edits to the same card (direct edit + dialog supplement arriving simultaneously — race on card state). <br>• Two cards pending at once (can the user have multiple unsubmitted UF-3 cards in one session? Undefined). <br>• 0 entities in Team (fresh team, nothing to disambiguate against, nothing to query). <br>-20. |
| Failure recovery described | 15/25 | Recovery for AI timeout/unavailable is well-described (UF-6 + Story 6). Deductions: <br>• No recovery path for "API submit failed" — UF-3 has 失败 state with 重试 but no AC or scenario describes whether retry re-runs pre-check, whether the card stays editable, or whether partial server-side effects are possible. <br>• No recovery path for "user abandoned a card then returns" — is the card preserved, discarded? <br>• No recovery path for "page navigation blocked" confirmation — what happens if the user confirms leave with unsubmitted card (lost? saved as draft?). -10. |

### 7. Scope Clarity (100 pts) → 78/100

| Criterion | Score | Notes |
|-----------|-------|-------|
| In-scope items are concrete deliverables | 30/35 | Most items are concrete UI components or backend modules. Deduction: "AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）" describes a capability, not a deliverable — it could mean a service, a prompt template, a model selection. -5. |
| Out-of-scope explicitly lists deferred items | 25/30 | 10 explicit Out-of-Scope items, each named. Deduction: bulk operations are excluded ("批量操作（如'把所有 P0 事项分配给张三'）") but "查询多个" (e.g., "列出我所有 P0 和 P1") is not addressed — queries can inherently return multiple, so the boundary between allowed multi-result query and disallowed bulk write is implicit. -5. |
| Scope consistent with functional specs and user stories | 23/35 | Inconsistencies: <br>• In-Scope lists "状态机预校验：写操作前校验合法性，不合法直接拒绝并提示" — but ProgressRecord and ItemPool have no `available-transitions` endpoint (see Blindspot B1), so the in-scope promise cannot be honored for 2 of 6 entities. <br>• In-Scope lists "支持的实体：MainItem、SubItem、Milestone、MilestoneMap、ProgressRecord、ItemPool" but Story 2 AC#3 ("系统推送 ProgressRecord 进度更新卡片") has no matching end-to-end submit scenario, and there is NO story covering Milestone or MilestoneMap modify/assign — only create (Story 1 AC#4, Story 3 AC#3). So "修改" and "分配" operation coverage is partial across entities. <br>• Functional Specs (Related Changes #3) says the available-transitions endpoint exists on "MainItem/SubItem/MilestoneMap/Milestone" — directly contradicting the In-Scope claim of 6-entity coverage. -12. |

---

## Phase 3 — Blindspot Hunt

### B1 [blindspot] Pre-validation mechanism does not exist for 2 of 6 supported entities
- Quote: prd-spec Related Changes #3 — "复用 available-transitions 端点做写操作预校验 … 该端点已存在于 MainItem/SubItem/MilestoneMap/Milestone，无需新增"
- Reality: `backend/internal/handler/router.go` lines 123/135/159/171 confirm `available-transitions` is registered ONLY for main-items, sub-items, milestone-maps, milestones. There is NO such endpoint for `progress-records` or `item-pool`. Yet In-Scope lists ProgressRecord and ItemPool as supported write entities, and Story 2 AC#3 / Story 4 both imply write flows.
- Must improve: Either (a) explicitly state that ProgressRecord and ItemPool writes skip pre-validation and rely solely on submit-time API errors, with UX for that path, or (b) add new pre-validation endpoints in Related Changes. The current spec is internally contradictory on this point.

### B2 [blindspot] "Assign" operation semantic ambiguity
- Quote: prd-spec Background "其中'分配'映射到实体的 assignee 字段更新"
- Issue: ItemPool has its own `POST /item-pool/:poolId/assign` route (router.go line 149) which is a *review* action (assigning a pool item to a MainItem / reviewer), not an assignee-field update. SubItem assignee update is a normal field update, but MainItem assignee update is also a field update. The PRD collapses "分配" into "assignee field update" but does not clarify whether Copilot can trigger the ItemPool-specific assign action (which is `item_pool:review` permission, not `item_pool:submit`). This is a permission boundary that a QA would file as a P1 ambiguity.
- Must improve: Distinguish field-level assignee update (MainItem/SubItem) from the ItemPool assign workflow action, and state Copilot coverage for each.

### B3 [blindspot] Card single-source-of-truth vs. concurrent dialog updates — race undefined
- Quote: UF-3 "用户可直接编辑卡片字段（onChange 更新卡片 state）或对话补充（后端解析增量变更后更新卡片 state）"
- Issue: If the user is typing in a field (onChange firing rapidly) while a dialog-driven update arrives asynchronously from the backend, the merge order is undefined. Two writes to the same field can produce a flaky UI where the user's keystrokes are overwritten by a stale AI update. This is exactly the kind of bug that escapes happy-path testing.
- Must improve: Define merge semantics (last-write-wins with timestamps? dialog updates paused while field is focused? optimistic local lock?) — and add an AC asserting the merge behavior.

### B4 [blindspot] No definition of "intent recognition confidence threshold"
- Quote: Flow Diagram decision M1 "意图是否被识别?" and mermaid "M1{意图被识别?}"
- Issue: There is no confidence threshold defined anywhere. The Goal "AI 意图识别准确率 ≥ 85%" is a population metric, not a decision threshold. Without a threshold, QA cannot test the boundary between "返回引导文字" and "进入实体解析" — every test outcome becomes tester-discretion.
- Must improve: Define a numeric confidence threshold (or document an explicit fallback rule), and add boundary ACs at the threshold edge.

### B5 [blindspot] "Sensitive field filtering" is asserted but not enumerated
- Quote: DF002 "敏感字段（密码、token）已在代理层过滤" and Security "禁止将密码/token 等敏感字段发送至 AI 服务"
- Issue: The filter is asserted as already implemented ("已在代理层过滤") but the only examples are "密码、token". A real prompt-injection / data-leak test plan needs an enumerated list (API keys, JWTs, PII fields, internal bizKeys of other teams, etc.) and the filter mechanism (regex? field-name allowlist? schema-aware?). Without enumeration, the security-review test plan is unfalsifiable.
- Must improve: Provide an enumerated sensitive-field list and the filter strategy, or move this from "已实现" assertion to "to-be-defined" with a concrete spec.

### B6 [blindspot] "每用户每日调用上限" value never specified
- Quote: Security "速率限制：每用户每日 AI 调用次数硬上限，防止成本失控" and Monitoring "每用户每日调用次数达上限时降级为关键词匹配模式"
- Issue: The quota NUMBER is never given. Two different consequences are also described inconsistently: Security frames it as a hard ceiling for cost control, Monitoring frames the threshold-crossing behavior as "降级为关键词匹配模式" — but the user-facing UX for hitting the ceiling is never specified in any UI Function or Story. A QA cannot write a test for "user hits daily quota" without a number and a UX.
- Must improve: Specify the numeric quota and add a Story/AC for the quota-exceeded user experience.

---

## Cross-Dimension Coherence Notes

1. **Coverage asymmetry between operations and entities.** The Goals table promises "支持创建/查询/修改/分配四类操作 × 6 个实体" (24 cells). The user stories actually cover roughly: create (4 entities), query (MainItem only), modify (MainItem state change + ProgressRecord card-push only), assign (MainItem only). That is ~7 of 24 cells with end-to-end coverage. This is the single biggest QA risk: the Goals row sets an expectation the Stories do not deliver evidence for.

2. **"P95" vs single-call ambiguity.** P95 is used both as a per-call SLA (Story 1 AC#1 "系统在 P95 < 5 秒内推送预填卡片") and as a population metric (Performance Requirements, Monitoring). These are different test artifacts. Story-level ACs should use per-call thresholds; population P95 belongs in NFRs.

3. **Inconsistent max-length treatment.** "限制单次最大长度" / "限制单次输入最大长度" / "单次输入超过最大长度时截断并提示" appear in 3 places; the number is never given. Single source of truth missing for an input-validation rule.

4. **Page navigation guard scope mismatch.** prd-spec 异常分支 says "页面导航前若有未提交卡片"; UF-2 Navigation Rules says "未提交卡片（UF-3/UF-5）"; Story 6 AC#3 says "未提交的卡片". UF-5 (disambiguation) is a selection card, not a write card — is an unanswered disambiguation card also blocking? The three statements are not clearly reconciled.

---

## Score Summary

| Dimension | Score | Max |
|-----------|------:|----:|
| Background & Goals | 80 | 100 |
| Flow Diagrams | 115 | 150 |
| Functional Specs | 145 | 200 |
| User Stories | 135 | 200 |
| Scenario Completeness | 95 | 150 |
| Edge Case Coverage | 50 | 100 |
| Scope Clarity | 78 | 100 |
| **Total** | **698** | **1000** |

Below target (900) by 202 points. Primary drivers of the gap:
- Edge Case Coverage (50/100) — broad missing error paths and boundary conditions
- User Stories AC verifiability (5/50 within that dimension) — non-objective "Then" clauses and missing boundary ACs
- Scenario Completeness end-to-end coverage — 3 of 6 entities lack submit→success scenarios
- Scope/Functional Specs contradiction over `available-transitions` coverage (Blindspot B1)
