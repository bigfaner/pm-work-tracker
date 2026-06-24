# PRD Evaluation — Iteration 1 (Senior PM Persona)

Evaluator: Senior Product Manager (adversarial)
Target: 900 / 1000
Scope: `prd-spec.md`, `prd-user-stories.md`, `prd-ui-functions.md` (Mode A)

---

## Phase 1 — Reasoning Audit

**Problem → Solution → Evidence → Success Criteria trace:**

- **Problem** is concrete and well-stated: form interactions cost ~8-12 clicks, RBAC/state machine is a learning barrier.
- **Solution** (floating bubble + prefill card as single source of truth) follows logically.
- **Evidence** is the weakest link. The very metric that justifies the headline goal ("8-12 次点击") is explicitly flagged as unverified: *"开发者走查估算，待埋点验证"*. The entire "reduce to ≤3 interactions" target rests on a baseline the team has not measured. Shipping decisions about success will be impossible to make objectively without first establishing baseline.
- **Success Criteria** (85%/80%/P95<5s) are numeric, but the per-field accuracy definition *"日期 ±1 天内计正确，assignee 在 Team 成员模糊匹配内计正确"* is reasonable yet leaves title/description/priority undefined — these are the fields most often wrong and least specified.

**Self-contradiction check:**
- In Scope claims *"创建/查询/修改/分配四类操作 × 6 个实体"* (24 op-entities combos). User stories cover ~12; coverage gaps in query/modify for Milestone, SubItem, ProgressRecord, ItemPool, MilestoneMap. Goal promise exceeds evidence in stories.
- Spec claims support for *"ProgressRecord"* as an entity; the only ProgressRecord story (S2) is creation-only; modify/query/delete of progress are absent though "modify" is a promised operation class.
- Story 1 says Milestone auto-associates "若 Team 下仅一个 MilestoneMap 自动关联，多个则留空高亮" — but BIZ-milestone-005 says when MilestoneMap is terminal, no Milestone can be created. No story/edge case handles this. Hard contradiction with business rules.

---

## Phase 2 — Rubric Scoring

### 1. Background & Goals — 70 / 100

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| Three elements present & specific | 30 | 24 | Reason/Target/Users all present and specific. Slight deduction: Users table mixes role with a tool surface ("ItemPool 提交者（含非技术人员）") — acceptable but the four roles then re-used in stories are well-mapped. Minor: no persona counts/sizes. |
| Goals quantified | 30 | 22 | All five goals numeric. But one is explicitly qualitative (*"提升非技术人员可用性 — 定性目标，通过用户反馈验证"*) and another rests on an unverified baseline (*"待埋点验证基线"*). The "覆盖核心操作" goal is binary (yes/no), not really a metric. |
| Logical consistency goal↔problem | 40 | 24 | Headline metric baseline is unverified ("8-12 次... 待埋点验证") → cannot determine if ≤3 is a real reduction or artifact. Per-field accuracy definition partial (only 2 of ~6 field types specified). |

Deductions:
- -6 (Goals quantified): "定性目标，通过用户反馈验证" is not a verifiable metric; rubric calls out qualitative goals as weakness.
- -10 (Logical consistency): baseline "待埋点验证" undermines the only hard interaction-count metric.
- -8 (Goals quantified): per-field accuracy rule covers only date/assignee; title/priority/description/milestoneKey unspecified → "≥80% 字段提取准确率" is ambiguous.

### 2. Flow Diagrams — 122 / 150

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| Mermaid exists | 50 | 50 | Present, syntactically parseable. |
| Main path complete | 50 | 42 | Happy path start→end covered. Deduction: the "权限不足" branch is mentioned in text异常分支 but only the timeout branch is drawn in Mermaid; "AI 服务不可用" branch and "权限不足" branch not in diagram. |
| Decision points + error branches | 50 | 30 | Four decision diamonds present. But Mermaid only renders ONE error branch (timeout). Spec's other three exception branches ("AI 服务不可用 → 降级提示", "权限不足 → 返回权限提示", "用户输入超出单次最大长度 → 截断并提示") are not on the diagram. |

Deductions:
- -8: only 1 of 4 exception branches drawn.
- -20: diagram omits权限失败 and AI 不可用 paths even though text lists them.

### 3. Functional Specs (prd-ui-functions.md) — 152 / 200

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| Placement & Interaction completeness | 70 | 54 | All 7 UFs have Placement and User Interaction Flow. UF-3 interaction flow says *"前端调用 available-transitions 预校验"* — but spec PRD also says "复用 available-transitions 端点" — for create operations there is no transition to check (create is not a state transition). Calling available-transitions on a create card is undefined behavior. UF-1 placement enumerates routes but does not include `/item-pool` deep sub-paths consistently. |
| Data Requirements & States clarity | 70 | 54 | Field tables and state tables exist. Deductions: UF-3 字段集 type is `Map<fieldName, {value, required, derived}>` but does not specify how the client distinguishes required-by-AI-empty vs required-by-schema-missing; the "derived" flag has no semantics defined. UF-2 消息历史 type `List<{role, content, cardRef, ts}>` lacks max content length. UF-4 实体卡片列表 lacks max count even though Validation says "结果过多时分页或限制展示数量" without specifying the limit. |
| Validation Rules explicit | 60 | 44 | Most rules actionable. Deductions: UF-3 *"日期字段 ±1 天内视为与意图一致"* — this conflates accuracy metric with validation rule; the rule is non-actionable for the operator. UF-6 "快捷入口路由必须为已有页面路由" — no enumeration of which routes. UF-2 *"会话达 50 轮上限时提示用户开启新会话"* — no rule for what happens to in-flight cards at the cap. UF-5 "候选列表限定在当前 Team 范围内" — no rule when Team context missing (UF-2 says block send, but UF-5 doesn't reference). |

### 4. User Stories — 138 / 200

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| One story per target user | 50 | 38 | All 4 roles have a story (PM S1, Dev S2, TL S3, ItemPool S4). Deduction: 6 entities × 4 ops = 24 promised combos; stories cover only ~10. Query only appears for MainItem (S3) — query for SubItem/Milestone/MilestoneMap/ProgressRecord/ItemPool has no story. |
| Format correct | 50 | 44 | All stories use As a / I want / So that. Deduction: S3 "I want to 通过自然语言查询事项进度、创建里程碑图、调整负责人分配" bundles 3 distinct intents into one story — should be split. |
| AC in Given/When/Then | 50 | 44 | All ACs use GWT. Deduction: S1 second AC "Given 预填卡片已推送 / When PM 直接编辑...或继续对话...Then 两种方式均更新同一份卡片状态" — "或" makes the When non-deterministic; cannot test which branch. |
| AC verifiability & boundary coverage | 50 | 12 | Major deduction. Multiple unverifiable "Then": S3 *"推送 MilestoneMap 创建卡片，Team 字段自动预填当前 Team 上下文"* — "Team 字段" is not in any Data Requirements table. S1 *"milestoneKey=留空高亮"* — no rule for when AI should infer vs leave blank. S2 *"系统调用 available-transitions 预校验失败，返回错误说明并列出合法目标状态"* — for a SubItem creation card, available-transitions is meaningless. No AC covers boundary: max 50 turns reached, max input length, AI returns malformed JSON, AI returns an entity not in current Team, AI returns a bizKey that doesn't exist (stale ref). |

### 5. Scenario Completeness — 96 / 150

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| End-to-end scenario coverage | 60 | 34 | Many scenarios stop at "推送卡片". The full lifecycle (push → edit → submit → API call → success/failure UI → next action) is only described for MainItem create. Modify/assign/query/scenarios for the other 5 entities lack end-to-end. No scenario covers what happens after success (does card collapse? does a new card render?). |
| Implicit assumptions surfaced | 40 | 22 | Hidden assumptions: (1) AI service provider supports structured output — assumed but never validated; (2) "zero data retention" vendor requirement — no validation method; (3) bizKey exact-match resolution assumes user types bizKeys in natural language — implausible; (4) "Team 上下文缺失时阻止发送" but no rule for routes that are not Team-scoped (`/users`, `/roles`, `/report`); (5) assignee fuzzy match assumes Team member list is loaded — what if AI service doesn't have it? |
| Business-rules consistency | 50 | 40 | Mostly respects rules. Issues: (1) S1 "Milestone auto-associate to MilestoneMap" violates BIZ-milestone-005 if MilestoneMap is terminal — not addressed. (2) S2 SubItem move semantics (BIZ-lifecycle-003/004) not referenced — but SubItem move is out of AI scope implicitly; should be explicit. (3) BIZ-milestone-003: MainItem milestone_key update — AI "modify" intent should hit this guard; no scenario. (4) BIZ-filter-001: indirect assignee matches for queries like "我的 P0 事项" — not addressed; query result semantics ambiguous. |

### 6. Edge Case Coverage — 64 / 100

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| Error paths documented | 40 | 22 | Timeout, AI unavailable, permission denied covered. Missing: AI returns malformed/unparseable output; AI returns entity outside user's Team/permission scope; AI returns a bizKey that no longer exists; available-transitions endpoint itself errors/timeout; API call after confirm returns 4xx other than permission. |
| Boundary conditions covered | 35 | 22 | Input max length mentioned ("限制单次最大长度") but value unspecified. 50-turn cap mentioned, but behavior at cap unspecified (what about an in-flight card?). Large query result sets: "分页或限制展示数量" without a number. Token-usage cap mentioned ("每用户每日调用次数达上限") but the cap value and message-to-user behavior unspecified. Concurrent: no rule for user editing card while AI streams an update. |
| Failure recovery described | 25 | 20 | Recovery via "传统表单快捷入口" described for timeout/unavailable. But for "校验失败 → 用户修正后重试" the recovery is vague — does the card stay rendered with errors? Does the user retype? Not specified. No recovery for malformed AI output (no retry with different prompt, no re-ask). |

### 7. Scope Clarity — 80 / 100

| Criterion | Points | Awarded | Notes |
|---|---|---|---|
| In-scope concrete deliverables | 35 | 28 | Most items are concrete components/features. Deductions: "AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）" — implementation language, not a deliverable. "复用现有 API 权限检查与业务逻辑" — not a deliverable. |
| Out-of-scope explicit | 30 | 28 | Out of Scope is explicit and detailed. Minor: "批量操作" excluded but "查询" can return many results — boundary between batch and multi-result query unclear. |
| Scope consistent with specs/stories | 35 | 24 | In Scope promises 4 ops × 6 entities (24 combinations). User stories + UI functions only demonstrate ~10 combinations. Goal row 5 *"一次性全量交付，避免体验割裂"* sets expectation that all 24 work at launch, but stories/edge cases don't back this. Mismatch between Scope's promise and Stories' evidence. |

---

## Cross-Dimension Coherence Notes

1. **Metric definition asymmetry**: Goal "字段提取准确率 ≥ 80%" depends on per-field rules, but only date and assignee rules are given (prd-spec.md line 46). title/priority/milestoneKey/description have no rule → accuracy unmeasurable → unverifiable goal.
2. **available-transitions misuse**: PRD spec (line 64, 101) and UF-3 (line 143, 169) both call available-transitions for all write operations. But available-transitions is a state-machine endpoint — for a create operation there is no source state. This is a logic error repeated across spec + UI functions + story 2 AC4.
3. **"卡片为唯一数据源" vs editing after submit**: UF-3 success state says "成功反馈 + 实体跳转入口" but no rule for whether the card is then collapsed, retained, or editable. If user clicks edit again, what writes?
4. **Team context ambiguity**: UF-2 says "Team 上下文缺失时阻止发送"; but UF-1 enumerates non-Team-scoped routes (`/users`, `/roles`, `/report`) as showing Copilot. The bubble shows but the panel blocks send — inconsistent UX, not described.

---

## Phase 3 — Blindspot Hunt

`[blindspot]` items below are issues outside any single rubric dimension.

1. **[blindspot] Latency budget decomposition is missing** — *"AI 响应（用户发送到卡片推送）P95 < 5 秒"* gives one end-to-end budget. No breakdown for backend proxy latency, AI service round-trip, schema loading, fuzzy-match. Without breakdown, the team cannot diagnose P95 misses post-launch. Must improve: add a sub-second budget table.

2. **[blindspot] Cost ceiling / kill-switch thresholds are undefined** — *"每用户每日调用次数达上限时降级为关键词匹配模式"* and *"支持 20 活跃用户/天，平均 5 次 AI 调用/用户/天"* — but the actual per-user daily cap value is never specified. "关键词匹配模式" is mentioned once as a fallback with zero spec. If cost spikes, what triggers a global kill-switch? Not described. Must improve: define numeric caps and the fallback mode spec.

3. **[blindspot] Data retention / audit policy for AI call logs is incomplete** — *"不持久化原始消息"* (line 184) vs *"记录每次 AI 调用的意图识别结果、字段提取结果、用户后续行为"* (line 182) vs DF005 *"AI 调用记录（用户、时间、token 用量、意图）"*. The contradiction: if raw messages are not persisted, how are "字段提取结果" reconstructed for audit/compliance? Are extracted fields PII? Security/compliance stakeholders (line 39) get no data-classification policy. Must improve: a data-retention table per field type with TTL and access controls.

4. **[blindspot] Localization / language scope unstated** — All examples are Chinese. Does the AI service handle English/mixed input? Is the prompt Chinese-only? For multinational teams this is a launch-blocker. Must improve: explicit statement of supported input languages.

5. **[blindspot] Accessibility (a11y) is entirely absent** — The bubble, panel, cards have no keyboard-trap, screen-reader, focus-management, or ARIA rules. UF-2 says "Esc 收起" but tab-order, focus-return-to-trigger, and card-edit keyboard navigation are unspecified. For a feature whose goal is lowering barrier for non-technical users, this is a gap. Must improve: add a11y validation rules per UF.

6. **[blindspot] "Single source of truth" concurrency undefined** — *"两种方式均写入同一份卡片状态"* (line 26, 100, UF-3) but what if AI streams an update to a field the user is simultaneously editing? No conflict-resolution rule. Must improve: define edit-vs-stream precedence.

7. **[blindspot] No story/scenario for "AI confidently extracts wrong entity that happens to pass validation"** — The whole safety model rests on available-transitions pre-check. But e.g. "把张三的 P0 事项分配给李四" — if AI picks the wrong MainItem and that MainItem passes the assignee-update check, the wrong item is reassigned. No undo, no confirm-of-target beyond the card. This is the highest-severity operational risk and is unaddressed. Must improve: add an undo window or stronger target confirmation.

8. **[blindspot] Quality Checklist self-certifies "no ambiguous wording"** — *"Is there any ambiguous or vague wording [x] checked"* yet the document contains "等" (etc.) numerous times, "如右下角" (likely lower-right) placeholders, "结果过多时" thresholds unspecified. The checklist is a rubber stamp. Must improve: replace [x] with a verified-by evidence line per item.

---

## Score Summary

| Dimension | Score |
|---|---|
| Background & Goals | 70 / 100 |
| Flow Diagrams | 122 / 150 |
| Functional Specs | 152 / 200 |
| User Stories | 138 / 200 |
| Scenario Completeness | 96 / 150 |
| Edge Case Coverage | 64 / 100 |
| Scope Clarity | 80 / 100 |
| **Total** | **722 / 1000** |

Below target (900). Iteration 2 must focus on: (a) verified baseline metrics or hedged goal language, (b) full per-field accuracy rules, (c) complete error-branch diagram, (d) available-transitions applicability correction, (e) 4×6 coverage evidence, (f) numeric caps/limits, (g) data-retention policy, (h) wrong-entity undo/safety net.
