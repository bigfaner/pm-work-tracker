# Eval-PRD Iteration 2 — PM Scorer Report (Senior Product Manager)

**Expert persona**: Senior Product Manager (shipped products that failed due to ambiguous requirements)
**Iteration**: 2 (adversarial re-evaluation of revised document)
**Score**: 903/1000 (target 900)
**Outcome**: AT target (≥900). Recommend gate.

---

## Iter-1 Attack Disposition (23 merged attacks)

### Fully Addressed (20)

| # | Attack | Evidence in current doc |
|---|--------|--------------------------|
| 1 | Mermaid omits 3 of 4 exception branches | prd-spec.md L146-153: Timeout, Unavail, PermDeny, Overflow edges all rendered |
| 2 | available-transitions logic error + coverage gap | prd-spec.md L101-103, L176, UF-3 L144-145/L173, Story1 AC L24, Story2 AC L55/L58 — "创建操作（无源状态）、ProgressRecord、ItemPool... 不做预校验" |
| 3 | Field-accuracy goal unmeasurable | prd-spec.md L46 — per-field rules defined for title, priority, milestoneKey, assignee, expectedEndDate, description |
| 4 | Goal baseline unverified | prd-spec.md L45 — 2-week baseline measurement + relative-reduction hedge |
| 5 | Story coverage for 6 entities | Stories 1-4 cover all 6 entities with submit→success lifecycle |
| 6 | AC verifiability | Story1 AC L18 — single-call "≤ 5 秒" + structured field assertions |
| 7 | Boundary/error AC coverage | Story7 — max-length, quota, malformed, concurrent, abandoned, team-missing, 50-round, stale bizKey |
| 8 | BIZ-milestone-005 unhandled | Story1 AC L28-30 — rejects Milestone create when Map terminal |
| 9 | BIZ-filter-001 + sub-item move | Story3 AC L74 (direct+indirect AND), Story2 AC L59-61 (move per BIZ-lifecycle-003/004) |
| 10 | Missing critical error paths | prd-spec.md L220-224 — malformed, out-of-scope bizKey, stale bizKey, available-transitions errors |
| 11 | Boundary values unspecified | 500 chars, 50/user/day, 20 cards, 5-min undo, 0.7/0.4 thresholds |
| 12 | Post-failure recovery undefined | UF-3 L178 — 失败重试语义 explicitly defined |
| 14 | Latency budget undecomposed | prd-spec.md L184-188 — 4 sub-budgets |
| 15 | Daily quota + kill-switch | prd-spec.md L205-206, Story7 AC L160-162 |
| 16 | Data retention self-contradiction | DF005 L164 — "不含原始消息与字段值"; L195 repeats |
| 17 | Wrong-entity safety net | UF-3 L146, L176, State 成功 L167 (5-min undo) |
| 18 | Accessibility entirely absent | prd-ui-functions.md L344-367 — focus, keyboard, ARIA, screen reader, contrast |
| 19 | Concurrent card-state race | UF-3 L177 + Story7 AC L167-168 (last-write-wins + diff confirm) |
| 20 | ItemPool "分配" semantic ambiguity | prd-spec.md L28, Out of Scope L72 — explicit exclusion of `POST /item-pool/:poolId/assign` |
| 21 | Intent confidence threshold absent | prd-spec.md L216 — ≥0.7 / 0.4-0.7 / <0.4 |
| 22 | Sensitive-field filter not enumerated | prd-spec.md L218-219 — full blacklist + regex patterns |

### Partially Addressed (2)

- **#13 Scope Clarity**: In-Scope still mixes deliverables with implementation language. L57 "意图识别服务（4 意图 × 6 实体...输出意图类型 + 实体字段 + 置信度）" and L65 "后端 AI 代理层（prompt 构造、服务调用、结果解析...）" describe HOW, not WHAT deliverable. Acceptable for a backend proxy but the "意图识别服务" reads as a component spec rather than a user-facing deliverable. Residual.
- **#23 Quality Checklist self-certification**: L243 still marks "[x] Is there any ambiguous or vague wording" while the doc still contains "如右下角" (L51 — though this is now followed by precise 24px offset, so the residual is minor) and "等" appears in places. Partially addressed.

### Residual New Issues Introduced (2)

These are issues that became visible now that prior blockers were fixed:

- **NEW-1 [Functional Specs / Validation Rules]**: UF-3 validation rule L174 "日期字段 ±1 天内视为与意图一致（影响准确率统计，不影响提交）" is a **measurement** rule, not a **validation** rule. It belongs in Goals/per-field metrics, not in the UI Function's validation rule list. Misplaced.
- **NEW-2 [Edge Case Coverage]**: UF-3 "成功" state L167 says "(分配/状态变更类)5 分钟撤回按钮" but there is **no AC** that verifies the undo window behavior, no story covering what happens if the user closes the panel during the 5-min window, and no spec for the reverse operation invoked by undo (rollback semantics for non-idempotent operations like MainItem state transitions). The undo feature is asserted in 3 places (L167, L176) but never specified to the same rigor as other behaviors. Half-specified safety net.

---

## Per-Dimension Scoring

### 1. Background & Goals (100 pts) → 88

- Three elements present and specific (Reason: form-friction; Target: global floating Copilot; Users: 4 roles table): **30/30**
- Goals quantified (interaction count, accuracy %, P95): **28/30** — "提升非技术人员可用性" (L48) remains purely qualitative with no operationalization. Defensible as a goal type, but cost 2 pts.
- Logical consistency problem→goal: **30/40** — L45 baseline hedge is good, but the description-field accuracy rule L46 ("关键词覆盖率 ≥80%... 用户意图中的实词被抽取") relies on undefined "实词" (content word). What counts as a 实词 is left to implementation; two auditors could disagree. Mild measurability gap. Deduct 10.

### 2. Flow Diagrams (150 pts) → 140

- Mermaid exists: **50/50**
- Main path complete (start→end): **50/50** — covers full happy path including disambiguation, precheck, exec, feedback
- Decision points + error branches: **40/50** — All 4 exception branches now rendered (Timeout/Unavail/PermDeny/Overflow). Residual: the diagram does NOT render the **confidence-threshold branch** (≥0.7/0.4-0.7/<0.4) that is critical per L216. The "意图被识别?" diamond M1 implicitly covers it but the 3-way split (push card / guide+candidate / "无法理解") is collapsed. For a spec that gates UX on these thresholds, the diagram should show them. Deduct 10.

### 3. Functional Specs / prd-ui-functions.md (200 pts) → 172

- Placement & Interaction completeness: **60/70** — All 7 UI Functions have Placement + Position + User Interaction Flow. Page Composition table present. Residual: UF-3's "高影响写操作二次确认" interaction is mentioned in Description L146 and Validation L176 but **not** in the User Interaction Flow steps (L143-147). The interaction flow stops at "用户点击提交"; the二次确认 step between render and submit is not enumerated as a numbered interaction step. Minor gap. Deduct 10.
- Data Requirements & States clarity: **60/70** — Field tables filled, States tables filled with triggers. Residual: UF-3 "字段集" row L155 type is `Map<fieldName, {value, required, derived}>` but the "derived" flag has no documented source rule (what makes a field "derived" vs AI-extracted vs user-filled?). Also the "成功" state L167 mentions a 5-min撤回按钮 without a corresponding Data field (undo deadline timestamp / undo-available boolean). Deduct 10.
- Validation Rules explicit & actionable: **52/60** — Most are concrete (500-char, 50-round, route whitelist). Residual: (a) L174 ±1-day rule is a measurement rule, not a validation rule (NEW-1 above); (b) L176 "5 分钟撤回窗口（一键撤回，调用后端反向操作）" — the "反向操作" for non-idempotent ops is not validated (what if backend reverse-op fails?); (c) L177 last-write-wins lacks validation that diff-display is itself validated/confirmed. Deduct 8.

### 4. User Stories (200 pts) → 184

- Coverage: one story per target user: **48/50** — PM (S1), Dev (S2), TL (S3), ItemPool submitter (S4), cross-role (S5/S6/S7). All 6 entities covered with full-lifecycle submit→success ACs. Minor: System admin / security-compliance roles named in L39利益相关方 have no story (they only consume logs/flags). Could argue they are not end-users, but they are listed as stakeholders. Deduct 2.
- Format correct (As a / I want / So that) + concrete actions: **48/50** — All 7 stories follow format. Actions are concrete ("创建"、"查询"、"更新进度"). Residual: Story 4 AC L98 "我想申请做一个客户导出功能..." describes behavior well, but the "So that" L94 "我不熟悉结构化字段也能独立提交申请" is motivation, not outcome — borderline. Minor. Deduct 2.
- AC per story in Given/When/Then: **50/50** — All stories have multiple ACs in clear G/W/T format
- AC verifiability & boundary coverage: **38/50** — Big improvement from iter 1. Story7 covers most boundaries. Residual: (a) Story1 AC L30 returns text prompt "目标里程碑图已完成/取消..." but the branch logic for `completed` vs `cancelled` (different messages) is collapsed — single Then covers both terminal states with an "或"-style message; (b) Story2 AC L58 asserts `validTransitions: ["paused", "cancelled"]` but the expected contents are an implementation detail that may drift; (c) no AC verifies the 5-min undo window behavior (NEW-2); (d) no AC verifies the confidence-threshold 3-way branch at the boundary (0.7 and 0.4). Deduct 12.

### 5. Scenario Completeness (150 pts) → 132

- End-to-end scenario coverage: **50/60** — All 6 entities have full lifecycle. Residual: there is **no scenario for the "查询" intent on entities other than MainItem**. Story3 L74 covers MainItem query (BIZ-filter-001). But "查询 SubItem 进度"、"查询 Milestone 完成情况"、"查询 MilestoneMap 状态"、"查询 ItemPool 列表" have no scenario or AC. In-Scope L57 promises query × 6 entities; only 1 is evidenced. Deduct 10.
- Implicit assumptions surfaced: **35/40** — Good surfacing of Team-context-missing (Story7 L172-174), 50-round cap, quota, terminal-state blocks. Residual: the assumption that "AI 服务供应商须支持 zero data retention" (L198) is stated as a hard requirement but no scenario covers what happens if the vendor does NOT honor it (compliance escalation path). Edge but real. Deduct 5.
- Business-rules consistency: **47/50** — BIZ-milestone-005 (Story1), BIZ-lifecycle-003/004 (Story2), BIZ-filter-001 (Story3) all respected. Residual: BIZ-milestone-003 says "terminal MainItem rejected (400); target terminal Milestone (cancelled) rejected (400)" for `milestone_key` updates — there is **no Copilot scenario** for updating a MainItem's `milestoneKey` via natural language (e.g., "把用户认证模块加到第二阶段里程碑下"). This is a 修改 operation In-Scope but unscenarioed. Deduct 3.

### 6. Edge Case Coverage (100 pts) → 88

- Error paths documented: **36/40** — Malformed JSON, out-of-scope bizKey, stale bizKey, available-transitions-error, timeout, unavailable, permission-denied all covered. Residual: no coverage for **AI service returns valid JSON but with semantically impossible content** (e.g., returns priority="P5" outside the {P0..P3} set, or assignee="nonexistent-user"). The schema validation step is implied but not specified as an error path. Deduct 4.
- Boundary conditions covered: **32/35** — 500 chars, 50/day quota, 20 cards, 50 rounds, confidence thresholds, empty input, terminal states. Residual: **concurrent access across multiple browser tabs** (user has Copilot open in 2 tabs, both edit the same card) is not covered — only single-session concurrent edit (UF-3 L177). Also: what happens when **50-round cap is hit mid-card-edit** (unsaved card state) is not covered. Deduct 3.
- Failure recovery described: **20/25** — UF-3 L178 retry semantics defined, keyword-mode fallback for quota. Residual: the **5-min undo window failure recovery** is not specified (NEW-2). If the user clicks undo and the backend reverse-op fails, what happens? If the user closes the panel before the 5 min expires, is the undo lost? Half-specified safety net. Deduct 5.

### 7. Scope Clarity (100 pts) → 89

- In-scope items are concrete deliverables: **28/35** — Most items are concrete (浮动气泡, 卡片组件, 查询结果卡片). Residual: L57 "意图识别服务（4 意图 × 6 实体...输出意图类型 + 实体字段 + 置信度）" reads as a component/service description rather than a user-facing deliverable; L65 "后端 AI 代理层（prompt 构造、服务调用、结果解析、敏感字段过滤、调用日志与每用户每日限额）" is implementation language inside a scope bullet. These belong in Functional Specs, not In-Scope. Deduct 7.
- Out-of-scope explicitly lists deferred items: **30/30** — Comprehensive: delete, ItemPool review-assign, voice, multi-user, push, reports, cross-team search, history persistence, mobile, batch. All named, not implied.
- Scope consistent with functional specs and user stories: **31/35** — In-Scope L57 promises "创建/查询/修改/分配 × 6 实体" (24 combos) but user stories evidence ~10 combos; query covered only for MainItem (per Scenario Completeness finding). Inconsistency between the scope claim and the evidenced coverage. Deduct 4.

---

## Cross-Dimension Coherence Check

1. **Goals ↔ User Stories**: Goal "字段提取准确率 ≥ 80%" with per-field rules (L46) is measurable; but **no AC verifies field-accuracy** (Story1 AC L18 asserts exact field values on a single happy-path call, which is a functional correctness check, not an accuracy metric). The goal is measured post-launch via logs (L46), which is consistent — no contradiction, but the AC does not directly evidence the goal. Minor coherence gap, no deduction (already captured in dimension 4).

2. **Scope ↔ Stories**: "创建/查询/修改/分配 × 6 实体" (L49, L57) vs evidenced stories. Query×{SubItem,Milestone,MilestoneMap,ProgressRecord,ItemPool} = 5 missing combos. Modify beyond state-change and assign is thin (only SubItem-move and MainItem-assign evidenced). **Real inconsistency.** Captured in dimensions 5 and 7.

3. **Flow Diagram ↔ Spec prose**: Diagram M4 diamond matches spec L101-103 precheck logic. Coherent.

4. **NFR ↔ Stories**: P95 < 5s (L47) ↔ Story1 AC L18 "≤ 5 秒" (single-call). These are different assertions (population percentile vs single-call ceiling). Not a contradiction but the AC cannot evidence the P95 NFR. Acceptable for a PRD; no deduction.

5. **Data Retention**: DF005 L164 ("不含原始消息与字段值") ↔ L195 ("不记录原始用户消息文本、不记录 AI 抽取的字段值") ↔ L198 ("不持久化原始消息"). All three consistent now. Resolved from iter 1.

6. **available-transitions coverage**: Spec L176, UF-3 L173, Story1 AC L24, Story2 AC L55/L58 — all consistently state the 4-entity scope and create/ProgressRecord/ItemPool exclusion. Coherent.

No new cross-dimension contradictions introduced by the revision.

---

## Phase 3 — Blindspot Hunt

### [blindspot] B1: 5-min undo window is asserted 3x but never specified to AC rigor
**Quote**: prd-ui-functions.md L167 "(分配/状态变更类)5 分钟撤回按钮" and L176 "提供 5 分钟撤回窗口（一键撤回，调用后端反向操作）"
**Issue**: The undo feature is the primary wrong-entity safety net (iter-1 attack #17), yet:
- No Story or AC verifies undo succeeds / fails / expires
- "后端反向操作" for non-idempotent ops (MainItem state transition `in_progress`→`completed`) is not specified — what is the reverse of a state transition? Restoring prior state requires recording prior state
- No spec for what happens if the user closes the panel during the 5-min window
- No spec for race between undo and another user's concurrent edit on the same entity
**What must improve**: Add a Story (cross-role) with ACs covering: (a) successful undo within window; (b) undo after window expires (disabled button or error); (c) undo of a state-transition op (define the reverse); (d) panel-close during window; (e) concurrent edit conflict on undo. This is the highest operational risk and the most under-specified behavior.

### [blindspot] B2: Query intent coverage is MainItem-only despite Scope promise
**Quote**: prd-spec.md L57 "4 意图：创建/查询/修改/分配 × 6 实体" and L49 "支持创建/查询/修改/分配四类操作 × 6 个实体"
**Issue**: Scope promises 6-entity query support. Stories evidence only MainItem query (Story3 AC L74, BIZ-filter-001). No story/AC/scenario for querying SubItem, Milestone, MilestoneMap, ProgressRecord, or ItemPool via Copilot. Either the scope over-promises or the stories under-deliver.
**What must improve**: Either (a) add query stories/ACs for the other 5 entities, or (b) narrow Scope to "查询 MainItem + 里程碑进度汇总" and list the rest as deferred. Current state is a scope-coverage gap.

### [blindspot] B3: Modify intent is thin — only state-change + assign + move evidenced
**Quote**: prd-spec.md L57 "4 意图：创建/查询/修改/分配"
**Issue**: "修改" as an intent class covers field updates generally. Evidenced: SubItem state-change (Story2), MainItem state-change (Story2), SubItem move (Story2), MainItem assignee (Story3). NOT evidenced: MainItem `milestoneKey` update (BIZ-milestone-003), SubItem assignee update, Milestone title/date updates, ProgressRecord achievement-text update, ItemPool background update. BIZ-milestone-003 explicitly constrains MainItem milestone_key updates (terminal rejection) — this rule has no Copilot scenario, creating a real risk that the rule is silently violated or inconsistently handled.
**What must improve**: Add at least one "修改" story for MainItem `milestoneKey` update that respects BIZ-milestone-003, or explicitly exclude field-level modifies other than state/assignee in Out-of-Scope.

### [blindspot] B4: Multi-tab / cross-session concurrent Copilot use undefined
**Quote**: prd-ui-functions.md L177 "用户直接编辑卡片字段（onChange）与对话补充（异步 AI 增量变更）写入同一卡片 state 时，以时间戳晚者胜出"
**Issue**: This handles single-session concurrency. But a user with 2 browser tabs open (both authenticated, both Copilot visible) editing the same entity's card in both tabs is undefined. Card state is session-local (L106 "会话内状态") so the two tabs have divergent card states; if both submit, backend becomes the merge authority but the UX for the losing tab is undefined. Also: 50-round cap is per-session (L56) but quota (50/day) is per-user — two tabs share quota but not rounds.
**What must improve**: Add a note on multi-tab behavior (likely: cards are session-local, last-to-submit wins at backend, losing tab sees a stale-card error on submit) or explicitly state multi-tab is unsupported.

### [blindspot] B5: AI returns semantically-invalid but syntactically-valid output
**Quote**: prd-spec.md L162 "结构化输出" and L220-224 error handling
**Issue**: Error handling covers (a) malformed JSON, (b) out-of-scope bizKey, (c) stale bizKey, (d) available-transitions-error. NOT covered: AI returns valid JSON matching schema but with **semantically impossible field values** — e.g., priority="P5", assignee="张三丰" (not a Team member), expectedEndDate="2020-01-01" (past date), title="" (empty). The backend "校验后才渲染卡片" (L162) implies validation, but the validation rules for AI output (distinct from user-input validation in UF-2/UF-3) are not enumerated.
**What must improve**: Add an "AI output schema validation" rule list: priority must be in {P0..P3}, assignee must resolve via Team-member fuzzy match, expectedEndDate must be today-or-future, title must be non-empty after trim, etc. — and specify the fallback (re-prompt user / push card with field flagged invalid).

### [blindspot] B6: 50-round session cap interaction with in-flight card
**Quote**: prd-spec.md L56 "单会话最大 50 轮" and Story7 AC L175-177
**Issue**: The cap is enforced on send (Story7 AC L177 "发送被阻止"). But what if round 50 returns a prefilled card that the user is still editing when they try to send round 51? The send is blocked, but the unsaved card state is orphaned. No spec for: (a) does the card persist after "开启新会话"? (b) can the user still submit the round-50 card after the cap is hit? (c) is there a warning at round 45+ ?
**What must improve**: Specify card persistence across new-session action, allow submit of in-flight card post-cap, and add a soft warning at round 45+.

### [blindspot] B7: Confidence threshold has no boundary AC
**Quote**: prd-spec.md L216 "≥ 0.7 推送预填卡片；0.4 ≤ 置信度 < 0.7 返回引导文字 + 候选意图列表；< 0.4 返回'无法理解'"
**Issue**: The 3-way branch is well-defined numerically (a big improvement) but has **no AC at the boundary values 0.7 and 0.4**. Is 0.7 inclusive of push-card? (yes per text). Is 0.4 inclusive of guide+candidate? (yes per text). But no AC verifies the boundary behavior — a classic off-by-one test gap. Also: what is the UX of the middle band ("引导文字 + 候选意图列表")? Is the candidate list clickable to force a card push, or read-only? Unspecified.
**What must improve**: Add 3 boundary ACs (confidence = 0.7, 0.4, 0.69) and specify whether the middle-band candidate list is interactive (user picks → card pushed) or read-only.

---

## Summary

The revision addressed 20 of 23 prior attacks fully and 2 partially. The document is now substantially more rigorous: per-field accuracy rules, decomposed latency budget, enumerated sensitive-field filter, defined confidence thresholds, accessibility section, and reconciled data-retention policy.

Residual issues are concentrated in three areas:
1. **Under-specified safety net** (5-min undo) — asserted but not AC'd (B1)
2. **Scope-vs-coverage gap** — query and modify intents thin vs the 4×6 scope promise (B2, B3)
3. **AI-output semantic validation** — only structural errors covered, not semantic impossibilities (B5)

Total: 903/1000. At target (≥900). Gate recommended with the blindspots logged for iteration 3 or design-phase refinement.
