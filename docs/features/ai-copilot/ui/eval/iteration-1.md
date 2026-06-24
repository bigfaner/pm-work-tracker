# UI Design Evaluation — Iteration 1

- **Document**: `docs/features/ai-copilot/ui/ui-design.md`
- **Source of truth**: `docs/features/ai-copilot/prd/prd-ui-functions.md`
- **Rubric**: `ui-web.md` (1000 pts, target 950)
- **Expert persona**: Senior UX Engineer
- **Iteration**: 1

---

## Phase 1 — PRD Coverage Trace (UF → Component)

| PRD UI Function | Design Component | Coverage | Notes |
|-----------------|------------------|----------|-------|
| UF-1 浮动气泡 | Component 1 | Partial | 4 PRD states covered (默认/首次引导/AI不可用/隐藏); AI不可用 click behavior missing from Interactions table |
| UF-2 聊天面板 | Component 2 | Partial | 6 PRD states covered; **2 PRD Validation Rules missing**: Team 上下文缺失阻止发送 + 50 轮上限阻止发送 have no state/interaction/binding |
| UF-3 预填表单卡片 | Component 3 | Partial | States rich (incl. 不可逆成功/撤回成功/撤回窗口过期); **missing 权限不足 state** ("权限不足时不渲染可提交卡片，改为权限提示文字"); **missing retry interaction row** |
| UF-4 查询结果卡片 | Component 4 | Covered | 有结果/无结果/截断 all present; route set matches sitemap |
| UF-5 歧义消解卡片 | Component 5 | Covered | 待选择/已选择/已确认; ≥2 candidate rule implied |
| UF-6 降级与超时提示 | Component 6 | Covered | timeout/unavailable; route mapping present |
| UF-7 首次引导卡片 | Component 7 | Covered | First-show / 已引导; chip fill-no-send rule explicit |

**Navigation Architecture**:
- "No new pages, global overlay" — respected (all 7 components Mode=existing-page, Target=global overlay). ✓
- Secondary page jumps (UF-4 → `/items/:id`, `/items/:id/sub/:id`, `/milestones/:mapId`; UF-6 → `/item-pool`, `/items`) — all 4 routes referenced in Components 4 & 6. ✓
- "页面导航前若有未提交卡片弹出离开确认" — referenced in Component 2 Interactions but **no Dialog layout spec, no state row, no data binding for hasUncommittedCards**.

---

## Phase 2 — Rubric Scoring

### 1. Requirement Coverage (PM) — 175 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| UI function coverage (80) | 50 | All 7 UF have a component, but Component 3 is missing the **权限不足** state required by PRD UF-3 Validation Rule ("权限不足时不渲染可提交卡片，改为权限提示文字") — no row in States table, no data binding. Component 3 also lacks an explicit **retry** interaction row though PRD demands "失败后重试必须重新运行 available-transitions 预校验". -30 (PRD state gap within a UF). |
| Navigation Architecture coverage (40) | 30 | No new pages — correct. Secondary-page routes referenced. But the "页面导航前离开确认 Dialog" navigation rule (PRD Navigation Rules line 41) has no Dialog component spec — a navigation rule without its supporting UI. -10. |
| State requirement coverage (80) | 55 | Most PRD states covered, but two PRD UF-2 Validation Rules produce user-facing states that are absent from Component 2's States/Interactions/Data Binding: (a) "Team 上下文缺失时阻止发送并提示" — no state row, no blocked-send interaction, no data binding for `teamCtxMissing`; (b) "会话达 50 轮上限时阻止发送并提示用户开启新会话" — Data Binding mentions "≤50 轮" but no UI feedback when cap hit. -25. |
| Edge case handling (50) | 40 | 500-char truncation ✓, 20-card truncation ✓, undo expiry ✓, long-text max-w 85%/90% ✓. Missing: slow-network/streaming-interruption handling beyond "错误"; concurrent-edit diff overlay referenced in Component 3 ("展示 diff 浮层供确认") but diff 浮层 has no layout/state spec — referenced but undefined. -10. |

**Deductions applied**: PRD state gap within UF-3 (-30 already inside UI function coverage); UF-2 validation-rule states missing (-25 inside state coverage).

### 2. User Experience (End User) — 185 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Information hierarchy (80) | 55 | Chat panel header (title 14px + Team badge) is clear. But in Component 3 the **target-entity confirmation area** ("`认证模块 · MI-0023`") is rendered at "text-secondary 12px" — the same weight/size as casual meta — even though PRD UF-3 calls this a high-impact **二次确认** element. Title and bizCode are not visually separated, weakening the anti-wrong-entity guard. Also Component 4 摘要行 and 结果列表 use nearly identical typography rhythm (both 13–14px), so a P0 count summary doesn't visually dominate the list beneath it. -25. |
| Interaction intuitiveness (80) | 65 | Conventional patterns (click bubble → panel, Enter send, chip fill). But the **"AI 不可用" bubble still opens the panel** (Component 1 States: "点击仍可展开面板") yet the Interactions table has no row describing what happens on click when AI is unavailable — a user clicking a greyed/disabled-looking control and getting a panel is unintuitive and undocumented. -10. The 撤回 countdown mm:ss with no spec on tick source is also a usability unknown. -5. |
| Accessibility (90) | 65 | Good intent: role=log, role=dialog, aria-live polite, role=alert for errors, aria-label on bubble. But the design **asserts** "对比度：正文 ≥4.5:1" without verifying Component 6's palette: warning-text `#d97706` on warning-bg `#fffbeb` is ~3.9:1 (below 4.5:1 for 14px text) — the "AI 响应超时" title at 14px 500 in that color likely fails. Also `border warning #d97706/30` (30% opacity border, Component 6) reduces border perceptibility. No skip-link, no reduced-motion spec for the 三点跳动 / slide animations. -25. |

### 3. Design Integrity (Designer) — 180 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Design system adherence (80) | 60 | Tokens inlined correctly (blue=success, 13px density, border-delineated cards, level-1/2/3 shadow mapping). But Component 2 textarea spec "min-h 40px、max-h 120px" **contradicts the project Frontend Component Convention** (`min-h-[120px] resize-y` base for textareas). Even if intentional for chat, the design does not call out the deliberate deviation. Component 6 introduces "warning-bg `#fffbeb`" and "border warning `#d97706`/30" and "error-bg `#fef2f2`" / "error-text `#991b1b`" — **these tokens are NOT in the inlined Design System color table** (which only defines Warning `#d97706` and Error `#dc2626`, no -bg/-text variants). The design uses tokens it never defined. -20. |
| Visual coherence (90) | 70 | All 7 cards share surface/border/rounded-xl/shadow-1 — coherent. But **z-index conflict**: Component 1 bubble is `z-index: 50` at right:24px/bottom:24px; Component 2 panel is `z-index: 50`, right:0, width:420px. When the panel opens it visually covers the bubble's position, yet Component 1 Interactions says "面板展开时点击气泡 / 按 Esc | 收起面板" — implying the bubble remains clickable. The design never states whether the bubble hides, repositions, or sits above the panel. Cross-component visual/interaction inconsistency. -20. |
| State completeness (80) | 50 | Each component has a States table, and transitions are mostly described. But several state **transitions are undefined**: Component 2 has no transition from 思考中 → 错误 vs 思考中 → 超时 (both possible); Component 3 has no transition into 权限不足 (state missing entirely); Component 5 已确认 → "推送后续写/查卡片" — but no description of how the disambiguation card visually hands off to the next card (replace? stack?). Component 2 also references "页面导航前有未提交卡片 → 弹离开确认 Dialog" as an interaction but the Dialog has **no state row** in any component. Happy-path bias on the leave-confirmation flow. -30. |

**Deductions**: state-transition gaps (-30 inside State completeness).

### 4. Implementability (Developer) — 165 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Layout specificity (80) | 55 | Most components give concrete widths/spacing (panel 420px, bubble 56×56, paddings p-3/p-4). But: (a) the **离开确认 Dialog** ("弹离开确认（Dialog sm）") has no width, no button labels beyond "当前有未提交的操作，确定离开？", no cancel/confirm button spec — a developer cannot build it; (b) Component 3 "diff 浮层" referenced ("展示 diff 浮层供确认") with zero layout; (c) Component 4 progress bar "accent-light 填充" with no track color, no height; (d) Component 3 撤回倒计时 mm:ss has no positioning or tick-source spec. -25. |
| Data binding explicit (80) | 60 | Every component has a Data Binding table — good baseline. Orphan/missing bindings: (a) **离开确认 Dialog trigger** has no `hasUncommittedCards` field anywhere; (b) Component 3 **countdown timer** (`undoDeadline` is bound, but the rendered mm:ss ticking value has no field/source); (c) Component 2 **char-count truncation toast** ("已截断至 500 字符") — the toast itself is an orphan UI element with no data binding; (d) Component 1 **AI 不可用 click** routes to a 降级 panel state but no field binds the "panel-opened-from-disabled-bubble" vs normal-open distinction. -20 (orphans). |
| Interaction unambiguity (90) | 50 | Most trigger→action→feedback chains are explicit. Gaps: (a) Component 1 has **no row for clicking the bubble in AI 不可用 state** — the States table says "点击仍可展开面板" but Interactions only lists the default-click row; (b) Component 3 has **no retry interaction row** — 失败 state says "+ 重试" but no trigger→action→feedback; (c) Component 2 "页面导航前有未提交卡片" trigger is vague — what counts as "未提交"? UF-3 draft? UF-5 pending selection? Not specified; (d) Component 5 "继续后续流程（推送 UF-3/4）" — action is "推送" but no explicit event/handoff mechanism. -40. |

---

## Phase 3 — Cross-Dimension Coherence Check

- **Component 2 vs Component 1 z-index/position**: both `z-index: 50`; panel covers bubble area; bubble-click-to-close interaction contradicts panel-cover geometry. Cross-section inconsistency (-30 already counted in Design Integrity Visual coherence).
- **Component 6 uses undefined tokens** (`warning-bg #fffbeb`, `error-bg #fef2f2`, `error-text #991b1b`, `border warning /30`) that are absent from the inlined Design System color table. Cross-section inconsistency between "Design System" section and Component 6 (-20 counted in Design Integrity adherence).
- **Component 2 textarea** spec contradicts project textarea convention (min-h 40px vs `min-h-[120px]`). (-20 counted in Design Integrity adherence).
- **Component 3 retry** appears in States (失败 "+ 重试") but absent from Interactions — cross-section inconsistency between States and Interactions of the same component.

---

## Phase 3 — Blindspot Hunt

1. **[blindspot] Streaming-interruption / partial-card cleanup** — Component 2 "流式返回" state says "卡片骨架（border-dashed 占位）先显示，字段增量填充" but there is **no state or interaction for streaming interrupted mid-fill** (network drop during streaming). Does the partial card persist? Get discarded? Show a resume? PRD UF-2 states only enumerate idle/thinking/streaming/error/timeout — error covers it generically, but the design's streaming card is a new visual artifact that needs its own teardown rule. Quote: "流式返回 | 卡片骨架（border-dashed 占位）先显示，字段增量填充 | AI 流式返回". Must improve: add a 流式中断 state with explicit partial-card cleanup behavior.

2. **[blindspot] Bubble drag vs panel-open conflict** — Component 1 makes the bubble draggable ("cursor: grab … 限制在视口范围内"), and Component 1 Interactions says clicking the bubble while the panel is open closes the panel. But drag-vs-click disambiguation (mousedown → move threshold) is unspecified, and if the panel covers the bubble (see z-index issue), the user cannot reach the bubble to drag or click it. Quote: "拖拽气泡 | 改变位置 | 会话内保持新位置". Must improve: define drag threshold, define bubble visibility/position when panel is open.

3. **[blindspot] Multi-card concurrency in single message stream** — The design describes cards as messages in a single stream, but PRD allows "继续对话补充" to update an existing UF-3 card (Component 3 "继续对话补充 | 后端解析增量 → diff 确认 → 更新卡片 state"). If the user issues a second instruction that spawns a **new** card while a previous UF-3 card is still in 提交中 or 校验失败 state, the design does not define whether multiple in-flight cards coexist, block each other, or stack. Quote: "继续对话补充 | 后端解析增量 → diff 确认 → 更新卡片 state | 展示 diff 浮层供确认，不静默覆盖". Must improve: define concurrency rules across multiple live cards in one session.

4. **[blindspot] Team context switch mid-session** — Component 2 Team badge binds `teamCtx` from "当前页面路由/全局状态". If the user navigates the underlying page (the panel is a non-blocking overlay) to a different Team's page mid-conversation, the Team context changes but the existing message history and any in-flight UF-3 card were authored under the prior Team. Quote: "Team 徽章 | teamCtx {bizKey,name} | 当前页面路由/全局状态". Must improve: define behavior on Team switch mid-session (warn? freeze cards? invalidate?).

5. **[blindspot] Undo across page navigation** — PRD UF-3 says the undo window survives "同会话页面导航" but the panel is a global overlay mounted on the page; if the user navigates and the overlay remounts, the in-memory `undoDeadline` and `previousValue` may be lost. The design's Data Binding source is "提交后状态" (in-memory). Quote: "撤回按钮 | undoAvailable, undoDeadline, previousValue | 提交后状态". Must improve: specify persistence mechanism so undo survives overlay remount on navigation.

---

## Score Summary

| Dimension | Score |
|-----------|-------|
| Requirement Coverage | 175 / 250 |
| User Experience | 185 / 250 |
| Design Integrity | 180 / 250 |
| Implementability | 165 / 250 |
| **Total** | **705 / 1000** |

Below target (950). Major drivers: missing 权限不足 state, missing UF-2 validation-rule states (Team missing / 50-round cap), undefined Dialog/diff-overlay layouts, orphan data bindings (countdown timer, truncation toast, hasUncommittedCards), undefined tokens in Component 6, z-index/position conflict between bubble and panel, and absence of retry / AI-unavailable-click interaction rows.
