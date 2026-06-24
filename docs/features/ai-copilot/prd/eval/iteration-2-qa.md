# Eval-PRD Iteration 2 — QA Scorer Report (Senior QA Engineer)

**Expert persona**: Senior QA Engineer (caught production bugs that test plans missed)
**Iteration**: 2 (adversarial re-evaluation of the revised document)
**Score**: 872/1000 (target 900)
**Outcome**: Below target by 28 points. Recommend revise.

---

## Iter-1 Attack Disposition

### Fully Addressed (most of the 23)

The revision genuinely closed the major logic and coverage holes:

- **#1 Mermaid branches**: prd-spec.md L146-153 now renders Timeout, Unavail, PermDeny, and Overflow edges. Addressed.
- **#2 available-transitions logic + coverage gap**: prd-spec.md L101-103, L176; UF-3 L144-145/L173; Story 1 AC L24; Story 2 AC L55/L58 consistently state "创建操作（无源状态）、ProgressRecord、ItemPool... 不做预校验". Addressed.
- **#3 Field-accuracy rules**: prd-spec.md L46 defines per-field matching for title/priority/milestoneKey/assignee/expectedEndDate/description. Addressed.
- **#4 Baseline**: prd-spec.md L45 — 2-week baseline measurement + relative-reduction hedge. Addressed.
- **#5 Story coverage of 6 entities**: Stories 1-4 now cover all 6 entities with submit→success lifecycle. Addressed.
- **#7 Boundary/error AC coverage**: Story 7 enumerates max-length, quota, malformed, concurrent, abandoned, team-missing, 50-round, stale bizKey. Addressed.
- **#8 BIZ-milestone-005**: Story 1 AC L28-30 rejects Milestone create when Map is terminal. Addressed.
- **#9 BIZ-filter-001 + move**: Story 3 AC L74 (direct+indirect AND); Story 2 AC L59-61 (move per BIZ-lifecycle-003/004 happy path). Addressed at the happy-path level (see NEW residual below for boundary cases).
- **#10 Error paths**: prd-spec.md L220-224 — malformed, out-of-scope bizKey, stale bizKey, available-transitions errors enumerated. Addressed (but see NEW residual on missing ACs).
- **#11 Boundary numbers**: 500 chars, 50/user/day, 20 cards, 5-min undo, 0.7/0.4 thresholds — all concrete. Addressed.
- **#12 Post-failure recovery**: UF-3 L178 — retry re-runs pre-check, card stays editable, idempotent/transactional assumption stated. Addressed.
- **#14 Latency budget**: prd-spec.md L184-188 — 4 sub-budgets. Addressed.
- **#15 Quota + kill-switch**: prd-spec.md L205-206, Story 7 AC L160-162. Addressed (counting semantics undefined — see below).
- **#16 Retention contradiction**: DF005 L164 now says "**不含**原始消息与字段值"; L195 repeats the constraint. Reconciled.
- **#17 Wrong-entity safety net**: UF-3 L146, L176, State 成功 L167 (5-min undo + title/bizCode confirmation). Asserted — but see NEW-1 below for the spec gap.
- **#18 Accessibility**: prd-ui-functions.md L344-367 — focus, keyboard, ARIA, screen reader, contrast. Addressed.
- **#19 Concurrent card-state race**: UF-3 L177 + Story 7 AC L167-168 (last-write-wins + diff confirm for AI-origin). Addressed (see residual on determinism).
- **#20 ItemPool assign semantics**: prd-spec.md L28, Out of Scope L72 explicitly excludes `POST /item-pool/:poolId/assign`. Addressed.
- **#21 Confidence threshold**: prd-spec.md L216 — ≥0.7 / 0.4-0.7 / <0.4. Addressed.
- **#22 Sensitive-field filter**: prd-spec.md L218-219 — full blacklist + regex. Addressed.

### Partially Addressed / Residual

- **#6 AC verifiability**: Per-call latency ("≤ 5 秒") is now single-call verifiable (Story 1 AC L18). But "用户当日 AI 调用次数已达 50 次" (Story 7 AC L160) is not deterministically settable by a test runner — counting semantics undefined. Residual.
- **#13 Scope Clarity**: "意图识别服务（4 意图 × 6 实体...输出意图类型 + 实体字段 + 置信度）" (L57) and "后端 AI 代理层（prompt 构造...）" (L65) still read as implementation components, not deliverables. Residual, minor.
- **#23 Quality Checklist self-certification**: L243 marks "[x] Is there any ambiguous or vague wording" while the doc still contains "等" (e.g., L219 regex examples "等"), "等" in L57, and "如右下角" softened to 24px (mostly fixed). Residual.

---

## NEW Issues Introduced / Surfaced by Revision

### NEW-1 [Edge Case / Undo safety net] — Half-specified

UF-3 State "成功" L167: "(分配/状态变更类)5 分钟撤回按钮" and L176: "提供 5 分钟撤回窗口（一键撤回，调用后端反向操作）".

The undo button is asserted in 3 places but never specified to the same rigor as other behaviors. Specifically missing:
- No AC verifies the undo button appears, is clickable within 5 minutes, and disappears after.
- No story covers what "调用后端反向操作" means for non-idempotent transitions (e.g., MainItem `in_progress → paused`: is the undo `paused → in_progress`? That itself must pass the state machine — what if the machine forbids it?).
- No AC covers: user closes the panel during the 5-min window (does the undo survive? per session-state UF-2 says messages are not cross-session persistent), user navigates away, AI service is down when undo is clicked.
- No AC covers: concurrent undo (user clicks undo twice rapidly), undo after the target entity was edited by another user.

For a safety net that the document itself positions as "防止错实体", leaving it half-specified is a QA test-gap.

### NEW-2 [Edge Case / available-transitions fallback] — Specified in spec, no AC

prd-spec.md L224: "available-transitions 端点本身返回错误 → 降级为提交后端校验（依赖后端状态机校验），不阻断用户提交."

This fallback has no AC in any story. A test runner has no way to verify the user-facing behavior: does the card show a spinner, silently submit, show a banner? The fallback is asserted but not behaviorally defined or testable.

### NEW-3 [Scenario Completeness / Team-context contradiction] — Cross-doc inconsistency

- UF-1 L50 lists `/users`, `/roles` among authenticated routes where the bubble renders.
- Story 7 AC L172-174: "用户当前不在任何 Team 上下文页面（如 `/users` 全局页面） → When 用户尝试在 Copilot 发送写操作指令 → Then 发送被阻止".

These are consistent IF interpreted as "bubble shows on `/users`, write ops blocked, query ops allowed". But:
- The doc never states whether **query operations** are allowed on `/users`/`/roles` (Team-less context). "我的 P0 事项" requires a Team scope per BIZ-filter-001.
- UF-2 Validation L124 only says "Team 上下文缺失时阻止发送" — this would block ALL sends (including queries) on `/users`, contradicting the design intent of a global assistant.
- No AC covers the `/users`-page query scenario.

Ambiguous behavior at a boundary the document itself surfaces.

### NEW-4 [Scenario Completeness / Empty-Team edge cases] — Missing

- Story 1 AC L26-27: Milestone create when "Team 下仅一个 MilestoneMap" → auto-attach; "多个" → highlight required. **Missing**: Team has ZERO MilestoneMaps. The AI cannot infer a parent; what does the card show? No AC, no flow-diagram branch.
- Story 3 AC for "把用户认证模块分配给李四": assumes the entity resolves. No AC for the wrong-entity protection (no verification that the title+bizCode confirmation actually appears pre-submit per UF-3 L146).

### NEW-5 [Edge Case / Sub-item move boundary cases] — Missing ACs for BIZ-lifecycle-004 violations

Story 2 AC L59-61 covers the **happy** move (non-terminal source, same-team, different MainItem, non-terminal target). BIZ-lifecycle-004 explicitly returns 400 for: (a) target terminal, (b) cross-team, (c) same MainItem as source. None of these rejection paths have a Copilot AC. A user saying "把子任务X移到事项Y" where Y is terminal should produce a rejection card — but no AC verifies this end-to-end through the Copilot path.

### NEW-6 [Edge Case / Quota counting semantics] — Ambiguous

"每用户每日 AI 调用上限 50 次/用户/日" (prd-spec L205, L215) — undefined:
- Does the counter increment on (a) request sent, (b) AI response received, (c) successful response only?
- Does a timeout (>10s) count?
- Does the keyword-mode fallback count as a "call"?
- Does a malformed-AI-response count?
- Server-clock-day vs user-local-day for the boundary?

Story 7 AC L160 says "用户当日 AI 调用次数已达 50 次" — a test runner cannot deterministically reproduce this state without counting rules. Boundary value (49 → 50 → 51) is untestable.

### NEW-7 [User Stories / Concurrent-edit AC non-determinism] — Partial

Story 7 AC L167-168: "系统以**时间戳晚者胜出**合并...对话补充产生的增量变更在应用到卡片前先展示 diff".

But "直接编辑卡片字段（onChange）" is a synchronous local event — it has no network "timestamp". The AC conflates two clocks (keystroke time vs AI-response-arrival time). For two direct edits in rapid succession the merge is undefined. Test determinism at the millisecond boundary is not achievable as written.

### NEW-8 [Validation Rules / misplaced measurement rule]

UF-3 Validation Rules L174: "日期字段 ±1 天内视为与意图一致（影响准确率统计，不影响提交）".

This is a **measurement** rule (already in Goals L46), not a **validation** rule (which should describe accept/reject behavior on the card). Placing it under "Validation Rules" misleads a test author into writing a card-reject test that does not exist. Mislocated.

---

## Per-Dimension Scoring

### 1. Background & Goals (100 pts) — **88/100**

- Three elements present and specific (Reason: form inefficiency with 8-12 click evidence; Target: floating bubble + card hybrid; Users: 4-role table + stakeholders). Strong.
- Goals quantified: 6 metrics with numbers; field-accuracy per-field rules defined (L46); confidence threshold defined (L216). Strong.
- Logical consistency: solid, with one residual — Quality Checklist L243 self-certifies "no vague wording" while "等" (L57, L219) and the `/users` query-permission ambiguity persist. Minor inconsistency between self-audit and content.
- Deduction: -8 for residual vagueness in the self-certified checklist; -4 for the empty-Team edge (Goals promise "6 实体 × 4 操作" coverage but no empty-Team fallback exists in Goals' implied universality).

### 2. Flow Diagrams (150 pts) — **138/150**

- Mermaid exists with start→end main path complete.
- Decision points M1-M5 + 4 exception edges (Timeout, Unavail, PermDeny, Overflow). Strong.
- Deduction: -8 — the diagram still does not branch on confidence threshold (≥0.7 / 0.4-0.7 / <0.4), even though L216 defines it. The "M1{意图被识别?}" diamond conflates "not recognized" with "low confidence"; a tester cannot map the confidence bands to diagram nodes.
- Deduction: -4 — no diagram branch for empty-Team / Team-context-missing on write vs query (a real divergence in behavior).

### 3. Functional Specs (200 pts) — **172/200**

- Placement & Interaction completeness: every UF has Placement + Position + User Interaction Flow. UF-3 covers full card lifecycle. Strong.
- Data Requirements & States: field tables and state tables filled; sources/triggers explicit. UF-3 States L160-168 now include the undo, diff-confirm, retry semantics. Strong.
- Validation Rules: every UF has actionable rules EXCEPT the misplaced date-measurement rule (NEW-8) and UF-2 "Team 上下文缺失时阻止发送" is ambiguous on `/users` (NEW-3).
- Deduction: -10 (NEW-8 misplaced rule + UF-2 Team ambiguity); -10 (NEW-1 undo only half-specified at the functional-spec level — no rule defines what the "reverse operation" must do, no rule defines behavior when panel closes during the 5-min window); -8 (validation rule on quota counter is missing — the cap is asserted but the validation that increments the counter is unspecified).

### 4. User Stories (200 pts) — **160/200**

- Coverage: 4 roles + 3 cross-role stories. Each entity has submit→success lifecycle. Strong.
- Format: all "As a / I want / So that"; actions concrete.
- AC per story: every story has Given/When/Then ACs.
- AC verifiability & boundary coverage: largely verifiable, but with gaps.
- Deduction: -12 — Story 3 "分配给李四" AC (L84-86) does NOT verify the wrong-entity protection (no AC that the title+bizCode confirmation appears pre-submit; no AC for the 5-min undo). The highest-risk operation is happy-path-only.
- Deduction: -10 — Story 2 move AC (L59-61) is happy-path-only; missing rejection ACs for BIZ-lifecycle-004 violations (NEW-5).
- Deduction: -8 — Story 7 quota AC (L160) is not deterministically reproducible (NEW-6); concurrent-edit AC (L167-168) conflates clocks (NEW-7).
- Deduction: -10 — no AC verifies the available-transitions fallback (NEW-2); no AC verifies empty-Team milestone-create (NEW-4).

### 5. Scenario Completeness (150 pts) — **118/150**

- End-to-end coverage: main + cross-role + edge stories cover most lifecycles.
- Implicit assumptions: surfaced well (Team detection, feature flag, 50-round cap, sensitive filter).
- Business-rules consistency: BIZ-milestone-005 (Story 1 AC L28-30), BIZ-filter-001 (Story 3 AC L74), BIZ-lifecycle-003/004 (Story 2 AC L59-61) — all happy-path consistent.
- Deduction: -12 — Team-context query-permission on `/users`/`/roles` undefined (NEW-3) — implicit assumption not surfaced.
- Deduction: -10 — empty-Team edge cases (zero MilestoneMaps) not covered (NEW-4).
- Deduction: -10 — undo window lifecycle (panel close, session loss, AI-down-during-undo) not covered as a scenario (NEW-1).

### 6. Edge Case Coverage (100 pts) — **78/100**

- Error paths: malformed AI, out-of-scope bizKey, stale bizKey, available-transitions error, timeout, unavailable, permission-denied, quota — all listed in prd-spec L220-224. Strong at the prose level.
- Boundary conditions: 500 chars, 50/day, 20 cards, 0.7/0.4 confidence, 5-min undo, 50-round session. Concrete.
- Failure recovery: UF-3 L178 defines retry semantics.
- Deduction: -8 — quota counting semantics undefined (NEW-6) makes boundary testing impossible.
- Deduction: -7 — undo window recovery undefined (NEW-1) — what if undo itself fails?
- Deduction: -7 — available-transitions-fallback has no behavioral edge-case description (NEW-2) — only "降级为提交后端校验", no UX for the user.

### 7. Scope Clarity (100 pts) — **88/100**

- In-Scope items: mostly concrete (UI components, intent service, proxy layer, feature flag). Two items read as implementation rather than deliverables (L57 intent service, L65 proxy layer) — residual from iter-1 #13.
- Out-of-Scope: explicit and named (delete, ItemPool review-assign, voice, multi-user, push, reports, cross-team search, history persistence, mobile, batch).
- Consistency: scope matches functional specs and stories.
- Deduction: -7 (residual implementation-language In-Scope items); -5 (Quality Checklist self-certifies "no vague wording" while residual "等" remains).

---

## Blindspot Hunt

### [blindspot] BS-1: Quota counter atomicity under concurrent sends

Quote: "每用户每日 AI 调用上限 50 次/用户/日" (prd-spec.md L205).

A user with two browser tabs can issue the 50th and 51st call nearly simultaneously. The doc specifies the cap but not the counter's atomicity / consistency model. If both tabs read counter=49 and both proceed, the user exceeds 50. No AC, no spec on counter consistency. High-frequency test scenario.

### [blindspot] BS-2: Idempotency key for retried writes after timeout

Quote: "AI 响应（用户发送到卡片推送）P95 < 5 秒... 超 10 秒展示超时兜底" (prd-spec.md L184, L108).

When the AI call times out at 10s, the user gets a timeout prompt. But the AI request may still complete server-side after the 10s cutoff, push a card asynchronously, and the user — having used the fallback form — may now have two writes in flight. The doc says messages are session-scoped (not persisted), but does not define what happens to a late-arriving AI response after the user has navigated to the fallback. No dedup/idempotency-key strategy.

### [blindspot] BS-3: AI returning multiple intents in one utterance

Quote: "用户输入自然语言指令（如'创建一个 P1 事项，标题完成用户认证模块，分配给张三，下周五截止'）" (prd-spec.md L89).

The example bundles create + assign. What if a user says "创建 P1 事项 X 并把它的子任务 Y 移到事项 Z"? Two write intents, two entities, possibly conflicting. The intent-recognition spec (L57) assumes a single (intent, entity) per call. Multi-intent handling is undefined — does the system reject, serialize, or pick one? No AC, no diagram branch.

### [blindspot] BS-4: Sensitive-field regex false-positives blocking legitimate input

Quote: "正则匹配（API key/token 模式：`(?i)(sk-[a-z0-9]{20,}|ghp_[a-z0-9]{36}|eyJ[a-z0-9_-]+\.eyJ[a-z0-9_-]+)` 等），命中则替换为占位符 `[REDACTED]`" (prd-spec.md L219).

A legitimate user describing a GitHub integration might type "token starts with ghp_..." for context. The regex would REDACT it before the AI sees it, potentially breaking intent recognition ("token starts with [REDACTED]"). No false-positive handling, no user-visible notice that filtering occurred. QA test gap.

### [blindspot] BS-5: Session 50-round cap interaction with abandoned cards

Quote: "单会话最大 50 轮对话" (UF-2 L95) and "页面导航前若有未提交卡片（UF-3/UF-5），弹出离开确认提示" (UF-2 Navigation Rules L41).

If a user hits round 50 with an unsubmitted card, what happens? The send is blocked (Story 7 AC L176), but the existing unsubmitted card from round 49 — can the user still submit it? The two rules interact but no AC resolves the intersection. Boundary ambiguity at the cap.

### [blindspot] BS-6: Team-context indicator staleness after route change

Quote: "Team 上下文自动检测（跟随当前页面）" (prd-spec.md L59) and UF-2 "当前 Team 上下文 | Team 对象（bizKey + 名称） | 当前页面路由/全局状态".

If a user types a command, then navigates to a different Team's page while the AI is still thinking, which Team does the resulting card operate on? The Team context was captured at send time, but the indicator now shows a different Team. No AC for the race between route change and AI response.

---

## Summary

The revision closed the iter-1 logic holes (available-transitions, field-accuracy, story coverage, accessibility, sensitive-field enumeration, confidence threshold). However, the **undo safety net** (NEW-1), **available-transitions fallback** (NEW-2), and **Team-context-on-global-pages** (NEW-3) are asserted but not specified to testable rigor. Boundary ACs for **sub-item move rejections** (NEW-5) and **quota counting** (NEW-6) are missing. These are precisely the gaps a QA engineer would file as "test plan cannot be written" — the spec asserts behaviors but does not define their boundary or recovery semantics.

**Score: 872/1000 — below 900 target. Recommend revise focusing on NEW-1, NEW-2, NEW-5, NEW-6, and the misplaced validation rule (NEW-8).**
