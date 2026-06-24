# Eval-PRD Iteration 3 — PM Scorer Report (Senior Product Manager)

**Iteration**: 3
**Score**: 890/1000 (target 900)
**Mode**: A (feature with UI)
**Outcome**: Target NOT reached — 10 points short.

This is the final iteration. The reviser fixed all three Tier-1 attacks from iteration 2, but several Tier-2 issues carried forward and the scope-vs-evidence gap on the 4×6 coverage claim was only partially mitigated.

---

## Tier-1 Attack Verification

### T1-1: 5-min undo window under-specified → **ADDRESSED**

Story 8 now carries 11 acceptance criteria that fully operationalize the undo window. Cross-references are consistent across three documents:

- UF-3 Data Requirements adds `undoAvailable`, `undoDeadline` (= 提交成功时间 + 5min), and `previousValue` (Quote prd-ui-functions.md L158-160).
- UF-3 Validation Rules L178-185 specifies: reverse-op semantics (assignee restore / status restore), re-check available-transitions before restore, window boundary (`undoDeadline` = 提交成功时间 + 5 分钟), session scope (current session only; browser close/logout invalidates), AI-independence (undo calls existing entity API), uniqueness (one undo per op).
- Story 8 (prd-user-stories.md L202-235) provides 11 G/W/T ACs: success card with countdown, expiry ("撤回窗口已过期"), panel close/navigation persists, browser-close session-loss invalidation, AI-down-during-window still succeeds, one-undo-only, irreversible (terminal) tagged "该操作不可撤回", and the available-transitions re-check failure case.

Quote: "撤回为状态变更类，恢复原状态前重新调用 available-transitions 校验，当前状态已不允许回到原状态 → 撤回失败，卡片显示'当前状态无法回到原状态，合法目标状态为：{validTransitions}'，不执行反向操作" — objectively testable.

Verdict: Fully addressed. No residual.

### T1-2: Confidence-threshold bands not in diagram → **ADDRESSED**

The M1 diamond in the Business Flow Diagram now branches on three explicit bands (Quote prd-spec.md L125-128):

```
M1 -->|≥ 0.7 高置信| Resolve[...]
M1 -->|0.4 ≤ c < 0.7 中置信| Clarify[...]
M1 -->|< 0.4 低置信| Guide[...]
```

Boundary ACs added in Story 7 at all four edges (0.7 / 0.69 / 0.4 / 0.39):

- "置信度 = 0.7 → 高置信，直接推送预填卡片" (L187-189)
- "置信度 = 0.69 → 中置信，返回引导文字 + 候选意图列表" (L190-192)
- "置信度 = 0.4 → 中置信" (L193-195)
- "置信度 = 0.39 → 低置信，返回'无法理解'" (L196-198)

Verdict: Fully addressed. No residual.

### T1-3: Scope 24-combo claim vs query evidenced only for MainItem → **PARTIALLY ADDRESSED**

The Goals row and In-Scope bullet still claim "4 意图 × 6 实体 = 24" coverage (Quote prd-spec.md L49, L57). The reviser's mitigation is a single assertion in Story 3 AC: "查询支持全部 6 实体，MainItem 查询为本故事的参考 AC，其余实体走同一查询处理器" (prd-user-stories.md L83). Milestone and ItemPool queries also received explicit ACs (L78-83).

However:
- **Modify** op evidenced only for SubItem move and MainItem state-change (Story 2). No modify ACs for Milestone, MilestoneMap, ProgressRecord, ItemPool, SubItem (non-move).
- **Assign** op evidenced only for MainItem (Story 3 "把用户认证模块分配给李四"). No assign ACs for SubItem/Milestone/etc.
- The "统一查询处理器" hand-wave is an architectural assertion, not user-facing evidence. QA cannot verify SubItem/MilestoneMap/ProgressRecord query behavior from any AC.

The Goals/Scope claim "4×6 = 24 全量交付" is not substantiated by the ACs. The reviser narrowed the gap from "5 entities missing" (iter 2) to "modify/assign for 3-4 entity types missing" (iter 3) but did not reconcile the headline claim. Either the Goals row should narrow to "create (6) + query (6) + modify/assign for MainItem/SubItem" or additional ACs are needed.

Verdict: Partially addressed. Residual — reflected in deductions across Goals, User Stories, Scenario Completeness, and Scope Clarity.

---

## Per-Dimension Scoring

### 1. Background & Goals — 88/100

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| Three elements (Reason/Target/Users) | 30 | 30 | Reason: 10+ fields, 8-12 clicks. Target: floating bubble + card hybrid. Users: 4-role table. Specific. |
| Goals quantified | 30 | 30 | ≤3 interactions, ≥85% accuracy, P95<5s, 50/day quota — all numeric. Baseline mitigation ("上线前 2 周在对照组...埋点测量真实基线并固化为 `baseline_interactions`") is sound. |
| Logical consistency | 40 | 28 | Two deductions: (a) Goals row "覆盖核心操作 \| 支持创建/查询/修改/分配四类操作 × 6 个实体" over-claims — modify and assign are evidenced for at most 2 entities, not 6. (b) Per-field metric for description: "用户意图中的实词被抽取到 description 字段" — "实词" is a linguistic term, not a deterministic tokenization rule; cannot be objectively verified (carried Tier-2 #13). |

### 2. Flow Diagrams — 146/150

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| Mermaid exists | 50 | 50 | Single comprehensive flowchart. |
| Main path complete | 50 | 48 | Start→Open→Input→Proxy→Parse→M1→Resolve→M2→M3→Card→Edit→M4→M5→Confirm→Exec→Feedback→End covers happy path. Minor: post-success undo subflow not visualized (arguably out-of-diagram scope). |
| Decision + error branches | 50 | 48 | Five diamonds (M1-M5), four error branches (timeout, unavailable, permission, overflow). M1 now properly decomposes confidence bands. Clean. |

### 3. Functional Specs (prd-ui-functions.md) — 180/200

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| Placement & Interaction | 70 | 60 | All 7 UI Functions have Placement/Interaction. Deduction: UF-1 Placement (L50) lists `/users`, `/roles` as bubble routes, but UF-2 Validation (L124) blanket-blocks send on Team-missing, and Story 7 blocks writes on `/users`. Are queries allowed on `/users`/`/roles`? Can the user even open the panel there to do anything useful? Tier-2 #9 contradiction unresolved. |
| Data Requirements & States | 70 | 65 | Field/state tables complete for all 7 functions. UF-3 now includes undoAvailable/undoDeadline/previousValue (Tier-2 #5 fixed). Deduction: UF-3 lacks a Data field for the available-transitions-fallback state (spec L227 says "降级为提交后端校验" but no field captures this fallback mode in UF-3 Data/States). |
| Validation Rules | 60 | 55 | UF-3 undo validation is now excellent (L178-185). Concurrent-edit merge (L184) and retry semantics (L185) clear. Deductions: (a) available-transitions endpoint failure UX undefined — silent fallback to post-submit? (Tier-2 #11); (b) Reverse-op failure recovery undefined — if the undo API itself returns 500, what does the card show? |

### 4. User Stories — 183/200

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| Coverage (one per user) | 50 | 45 | 4 roles + 4 cross-role stories. Deduction: query coverage only MainItem/Milestone/ItemPool (3/6); modify only SubItem move + MainItem state; assign only MainItem. The 4×6 scope claim is under-evidenced (T1-3 residual). |
| Format correct | 50 | 50 | All stories use As a / I want / So that. Actions concrete (not "manage"). |
| AC format | 50 | 50 | All stories have Given/When/Then. Story 8 has 11 ACs. |
| AC verifiability & boundary | 50 | 38 | Verifiability is strong overall (concrete values, boundary points). Deductions: (a) Sub-item move rejection cases (BIZ-lifecycle-004: terminal source, cross-team target, same-source target) have no Copilot ACs — only happy path in Story 2 (Tier-2 #12); (b) MainItem milestoneKey modify (BIZ-milestone-003: terminal MainItem rejected, terminal Milestone rejected, order MainItem-then-Milestone) has no Copilot AC (Tier-2 #6); (c) Concurrent-edit clock granularity (sub-second direct-edit-vs-AI, two near-simultaneous direct edits) undefined (Tier-2 #20). |

### 5. Scenario Completeness — 125/150

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| End-to-end coverage | 60 | 48 | Full submit→success lifecycles evidenced for MainItem (S1), SubItem (S2), Milestone (S1), MilestoneMap (S3), ProgressRecord (S2), ItemPool (S4) on create. Coverage matrix at bottom of prd-user-stories.md is helpful. Deductions: (a) No end-to-end scenario for modify (non-move) or assign on any entity other than MainItem; (b) No query scenarios for SubItem/MilestoneMap/ProgressRecord (only architectural assertion). |
| Implicit assumptions surfaced | 40 | 32 | Team context, parent entity resolution, bizKey precision surfaced. Deduction: Team-context ambiguity on `/users`/`/roles` (can you query? can you open the panel? is send blocked entirely or only writes?) — Tier-2 #9 unresolved. Empty-Team edge cases not covered. |
| Business-rules consistency | 50 | 45 | BIZ-milestone-005 (terminal Map blocks child) ✓ in Story 1 AC. BIZ-filter-001 (direct + indirect assignee, AND with status) ✓ in Story 3 AC. BIZ-lifecycle-003/004 (move preserves status+assignee, only main_item_key+item_code change) ✓ in Story 2 AC. Deduction: BIZ-milestone-003 (MainItem milestoneKey modify: check MainItem status first then Milestone) not exercised by any scenario. |

### 6. Edge Case Coverage — 82/100

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| Error paths documented | 40 | 35 | Timeout, unavailable, permission, malformed AI, stale bizKey, overflow, quota-fallback — all covered with ACs. Deduction: AI semantic-invalid output (priority=P5, past expectedEndDate, non-team assignee, bizKey outside Team) — spec L225 covers "bizKey 不属于当前用户 Team 或权限范围" but not impossible field values (Tier-2 #7). |
| Boundary conditions | 35 | 25 | 500-char, 50-round cap, 0.7/0.69/0.4/0.39 confidence, undoDeadline+5min all addressed. Deductions: (a) Quota counting semantics undefined — does the counter increment on request, response, or success? Does a timeout/malformed response consume a slot? Day-boundary timezone? Atomicity under multi-tab? No ACs at 49/50/51 boundary (Tier-2 #8); (b) Multi-tab concurrent Copilot sessions undefined — does each tab have its own 50-round counter? Same session? (Tier-2 #10). |
| Failure recovery | 25 | 22 | Retry reruns precheck (UF-3 L185), timeout/unavailable fallback to traditional form, undo as recovery, abandon-card handling. Deduction: recovery path when undo's available-transitions re-check fails — Story 8 AC says "不执行反向操作" but no recovery guidance for the user (e.g., "go to traditional form to manually restore"). |

### 7. Scope Clarity — 86/100

| Criterion | Points | Score | Notes |
|-----------|--------|-------|-------|
| In-scope concrete deliverables | 35 | 28 | 12 items, mostly concrete. Deduction: still mixes deliverables with implementation language — "意图识别服务（4 意图...输出意图类型 + 实体字段 + 置信度）" and "后端 AI 代理层（prompt 构造、服务调用、结果解析、敏感字段过滤...）" describe HOW not WHAT (Tier-2 #19). A PRD In-Scope should name capabilities, not internal service composition. |
| Out-of-scope named | 30 | 30 | 10 named exclusions including ItemPool review/assign workflow, delete, voice, batch, cross-session persistence. Strong. |
| Scope consistent with specs/stories | 35 | 28 | Goals row "4 操作 × 6 实体" + In-Scope bullet "4 意图 × 6 实体" over-claim — modify/assign evidenced for ≤2 entities, query evidenced for 3/6. Cross-section inconsistency (T1-3 residual). |

---

## Tier-2 Residual Tracking

| # | Iter-2 Issue | Iter-3 Status |
|---|--------------|---------------|
| 4 | ±1-day rule misplaced | **Fixed** — now in Goals per-field metrics (L46). |
| 5 | UF-3 undo Data fields missing | **Fixed** — L158-160. |
| 6 | MainItem milestoneKey modify missing | Residual. |
| 7 | AI semantic-invalid output (P5, past date) | Residual. |
| 8 | Quota counting semantics | Residual. |
| 9 | Team-context contradictions (/users, /roles) | Residual. |
| 10 | Multi-tab concurrent Copilot | Residual. |
| 11 | available-transitions fallback UX | Residual. |
| 12 | Sub-item move rejection cases | Residual. |
| 13 | "实词" undefined | Residual. |
| 14 | Late-arriving AI response after timeout | Residual (blindspot). |
| 15 | Multi-intent utterances | Residual (blindspot). |
| 16 | Sensitive-field regex false-positive | Residual (blindspot). |
| 17 | 50-round cap ∩ unsubmitted card | Residual (blindspot). |
| 18 | Team-context staleness during in-flight call | Residual (blindspot). |
| 19 | In-Scope mixes impl language | Residual. |
| 20 | Concurrent-edit clock granularity | Residual. |

---

## Blindspots (outside rubric dimensions)

### [blindspot] B-1: Quota race in multi-tab

Quote prd-user-stories.md L168: "Given 用户当日 AI 调用次数已达 50 次 When 用户发送新指令 Then 系统返回降级提示'今日 AI 调用已达上限（50 次）'".

Two browser tabs both at count 49 fire simultaneously. Client-side check passes for both; server-side counter may increment to 51 before either sees the limit. Atomicity of the counter (compare-and-swap? server-side hard reject at 51?) is not specified. The AC is not testable in the concurrent case.

**Must improve**: Specify server-side atomic enforcement (e.g., "server-side counter is the source of truth; requests that push the count past 50 are rejected with the quota-fallback response, even if the client-side check passed"). Add an AC for the 49→50 and 50→51 boundary under concurrent submission.

### [blindspot] B-2: Late-arriving AI response after timeout fallback (carried from iter-2)

Quote prd-spec.md L149: "Proxy -.->|超时 >10s| Timeout[展示超时提示 + 传统表单快捷入口]".

If the user clicks the fallback, navigates to the traditional form, and submits — and the late AI response arrives 2 seconds later with a pre-filled card — what happens? Does the late card render? Is it deduplicated against the form submission? Is the form submission idempotent against a later card submission of the same intent? Undefined.

**Must improve**: Specify that a late-arriving AI response after timeout is discarded (or rendered read-only with a notice "this arrived after your fallback action"), and that the traditional-form submission and any in-flight AI card are deduplicated by intent+entity.

### [blindspot] B-3: Multi-intent utterances

Quote prd-spec.md L89: "用户输入自然语言指令（如'创建一个 P1 事项，标题完成用户认证模块，分配给张三，下周五截止'）".

All examples and ACs assume a single intent per utterance. What about "创建 P1 事项并分配给张三然后查询我的 P0 事项"? Two write intents across two entities + one query. Does the system reject, serialize (do them in order), or pick the highest-confidence? The spec is silent. This is a likely real-user behavior.

**Must improve**: Add an explicit rule (reject-with-clarify, or serialize-with-confirmation) and at least one AC for a two-intent utterance.

### [blindspot] B-4: Undo failure recovery path

Quote prd-user-stories.md L215-217: "撤回失败，卡片显示'当前状态无法回到原状态，合法目标状态为：{validTransitions}'，不执行反向操作".

After this failure, the undo button disappears (the operation is "spent"), and the original success card is now in a half-undone state. There is no recovery guidance for the user — no "go to traditional form to manually transition to X" link, no escalation path. The user is left holding a state they did not intend.

**Must improve**: Add a recovery affordance on undo failure — e.g., a link to the entity's detail page or to the traditional state-transition UI pre-populated with the valid transitions.

### [blindspot] B-5: Sensitive-field regex false-positive (carried from iter-2)

Quote prd-spec.md L222: "命中策略：基于字段名黑名单 + 正则匹配（API key/token 模式：`(?i)(sk-[a-z0-9]{20,}|ghp_[a-z0-9]{36}|eyJ[a-z0-9_-]+\.eyJ[a-z0-9_-]+)` 等），命中则替换为占位符 `[REDACTED]` 后再发送".

If a user legitimately types a GitHub-shaped token in `description` (e.g., documenting an integration), the AI receives `[REDACTED]` in its place. Intent recognition may break or extract a wrong title, and the user has no visible notice that their input was altered. The pre-filled card will contain `[REDACTED]` verbatim, which the user may submit unknowingly.

**Must improve**: Specify a user-visible notice when redaction occurs ("输入中含有疑似敏感字段，已替换为 [REDACTED] 后发送给 AI"), and a rule that `[REDACTED]` must not appear in the pre-filled card fields (fall back to empty + highlight).

### [blindspot] B-6: 50-round cap ∩ unsubmitted card (carried from iter-2)

Quote prd-user-stories.md L181-183: "Given 会话已达到 50 轮上限 When 用户尝试发送第 51 轮指令 Then 发送被阻止".

If round 49 produced an unsubmitted UF-3 card and the user has not yet submitted it when they attempt round 50/51, can they still submit the round-49 card? The cap blocks *sending new instructions*, but card submission is a separate action. The interaction between the cap and pending cards is undefined.

**Must improve**: Clarify that card submission (a write-op execution) is independent of the round counter and remains available regardless of cap state. Add an AC.

### [blindspot] B-7: Team-context staleness during in-flight AI call (carried from iter-2)

Quote prd-ui-functions.md L105: "当前 Team 上下文 | Team 对象（bizKey + 名称） | 当前页面路由/全局状态 | 明确展示给用户，所有操作在此 Team 范围内".

If the user sends an instruction on Team A, the AI is still thinking, and the user navigates to Team B's page (context switches to B) before the card renders — which Team is the card scoped to? The Team-at-send or the Team-at-render? The pre-filled card's entity resolution, RBAC scope, and disambiguation all depend on this. Undefined.

**Must improve**: Pin Team-at-send as the authoritative scope for the in-flight request; render the card with the send-time Team badge regardless of current page. Add an AC.

---

## Final Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Background & Goals | 88 | 100 |
| Flow Diagrams | 146 | 150 |
| Functional Specs | 180 | 200 |
| User Stories | 183 | 200 |
| Scenario Completeness | 125 | 150 |
| Edge Case Coverage | 82 | 100 |
| Scope Clarity | 86 | 100 |
| **Total** | **890** | **1000** |

**Vs iteration 2 (PM): 903 → 890 (-13).** The Tier-1 fixes (undo, confidence bands, partial query coverage) added real rigor, but the scoring got stricter this iteration because the 4×6 over-claim is now a cross-dimension coherence violation (it persists in Goals + In-Scope after the reviser's partial fix), and several Tier-2 items that were "potential" in iter 2 are now "confirmed residual" in iter 3 because the reviser did not address them. Target (900) not reached.

**Highest-leverage fixes for a hypothetical iteration 4:**

1. Reconcile the 4×6 coverage claim: either narrow Goals/In-Scope to match evidence (e.g., "create 6 + query 6 + modify/assign on MainItem/SubItem"), or add the missing modify/assign/query ACs. (+8-10 points across 4 dimensions.)
2. Resolve Team-context contradictions on `/users`/`/roles`. (+4 points.)
3. Define quota counting semantics + multi-tab atomicity. (+4 points.)
4. Add Sub-item move rejection ACs and MainItem milestoneKey modify AC. (+4 points.)
