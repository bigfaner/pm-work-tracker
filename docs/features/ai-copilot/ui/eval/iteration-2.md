# UI Design Evaluation — Iteration 2

- **Document**: `docs/features/ai-copilot/ui/ui-design.md`
- **Source of truth**: `docs/features/ai-copilot/prd/prd-ui-functions.md`
- **Rubric**: `ui-web.md` (1000 pts, target 950)
- **Expert persona**: Senior UX Engineer
- **Iteration**: 2 (adversarial re-score of the revised document)

---

## Phase 0 — Iteration-1 Attacks: Addressed vs Residual

| # | Iteration-1 Attack | Status | Evidence in current doc |
|---|--------------------|--------|-------------------------|
| 1 | Component 3 missing **权限不足** state | **Addressed** | States table row "权限不足 \| 不渲染字段区与提交按钮；卡片体替换为权限提示" + 渲染规则 block (`canSubmit=false`) + Data Binding `canSubmit boolean` (lines 231, 259, 283) |
| 2 | Component 3 missing **retry interaction row** | **Addressed** | Interactions row "点重试（失败态） \| 状态变更类先重跑 available-transitions 预校验 → 通过则重新调提交 API" (line 270) |
| 3 | UF-2 missing **Team 缺失阻断** state | **Addressed** | State row "发送阻断（Team 缺失）" + Interaction row + Data Binding `teamCtxMissing` (lines 172, 183, 200) |
| 4 | UF-2 missing **50 轮上限阻断** state | **Addressed** | State row "发送阻断（轮次上限）" + Interaction row + Data Binding `sessionRoundCount` (lines 173, 184, 201) |
| 5 | **离开确认 Dialog** had no layout | **Addressed** | Full spec under "离开确认 Dialog（路由守卫）" — `max-w-sm`/overlay/surface/buttons/focus/Tab order (lines 207–215); Data Binding `hasUncommittedCards` added (line 203) |
| 6 | Component 3 **diff 浮层** referenced undefined | **Residual (partial)** | Trigger row still says "展示 diff 浮层供确认" (line 267) but the **diff overlay itself has no layout spec, no states, no data binding**. The leave-confirm Dialog got a full spec; the diff overlay did not. Referenced-but-undefined persists. |
| 7 | Component 3 **撤回倒计时 mm:ss** orphan source | **Addressed** | Data Binding row "撤回倒计时显示 \| undoCountdown 'mm:ss' \| 客户端 ticker：由 undoDeadline 倒推（每秒 tick），deadline 到达触发「撤回窗口过期」" (line 282) |
| 8 | Component 2 **char-truncation toast** orphan | **Addressed** | Data Binding `truncationNotice` (line 202); Interactions row ties it to a 3s auto-reset (line 182) |
| 9 | Component 1 **AI 不可用 click** no interaction row | **Addressed** | Interactions row "点击气泡（AI 不可用态）" (line 121) |
| 10 | Component 6 **undefined tokens** (warning-bg/error-bg/error-text/warning-text/border-warning-30) | **Addressed** | New "Status variants" table declares Success/Warning/Warning Title/Error BG/Error Text (lines 41–50) |
| 11 | **z-index/position conflict** bubble vs panel | **Addressed** | "面板展开时气泡隐藏（display:none，非 opacity）…面板展开期间无需点击气泡收起" (line 105) |
| 12 | **Accessibility contrast** (warning #d97706 on #fffbeb ~3.9:1) | **Addressed (with a residual technical inaccuracy)** | The 14px body text was moved to warning-text `#92400e` (verified 6.84:1 — passes). BUT the doc still claims "warning-title #d97706 仅用于图标与 14px 500 标题（large-text ≥3:1）" (line 464). Verified ratio 3.07:1 — meets ≥3:1 numerically, **but 14px 500 is NOT WCAG "large text"** (WCAG large = ≥18pt regular = 24px, or ≥14pt bold = 18.66px). So a 14px 500 title at 3.07:1 actually **fails** WCAG AA for its size. Minor residual. |

**Count**: 10 of 12 fully addressed; 1 partial (diff overlay), 1 with a residual technical inaccuracy (large-text classification).

---

## Phase 1 — PRD Coverage Trace (UF → Component)

| PRD UI Function | Design Component | Coverage | Notes |
|-----------------|------------------|----------|-------|
| UF-1 浮动气泡 | Component 1 | Covered | All 4 states + drag/click/AI-unavailable interactions bound |
| UF-2 聊天面板 | Component 2 | Covered | All 6 PRD states + 3 new blocking states + leave-confirm Dialog fully spec'd |
| UF-3 预填表单卡片 | Component 3 | Covered (one residual) | 权限不足 + 不可逆成功 + 撤回 states present; retry interaction present; **diff 浮层 referenced but still undefined** |
| UF-4 查询结果卡片 | Component 4 | Covered | Route set + 20-card cap + progress bar (no track color/height — minor) |
| UF-5 歧义消解卡片 | Component 5 | Covered | ≥2 candidate rule implied; 已确认 → fold+push handoff described in words |
| UF-6 降级与超时提示 | Component 6 | Covered | tokens now defined; route mapping present |
| UF-7 首次引导卡片 | Component 7 | Covered | Fill-no-send rule explicit |

Navigation architecture: still respected (no new pages; all overlay). Leave-confirm Dialog now fully specified.

---

## Phase 2 — Rubric Scoring

### 1. Requirement Coverage (PM) — 224 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| UI function coverage (80) | 74 | All 7 UF mapped to components. UF-3 diff 浮层 still referenced without a layout spec — a sub-element of UF-3 is undefined. -6. |
| Navigation Architecture coverage (40) | 38 | No new pages — correct. Secondary-page routes all referenced. Leave-confirm Dialog now spec'd. Minor: Component 4 progress bar `accent-light 填充` (line 301) gives no track color or height — but this is a layout concern counted under Implementability. -2 here for the progress-stat data binding being optional/marked "（里程碑场景）" without specifying when it applies. |
| State requirement coverage (80) | 72 | All PRD states + new blocking/stream-interrupt/team-switch states covered. Residual: PRD UF-3 requires the diff preview before "对话补充" applies (line 184 of PRD: "对话补充产生的增量变更在应用到卡片前先展示 diff 供用户确认"); the design references the diff 浮层 but provides no state/layout for it, so the "confirm diff" state is unspecified. -8. |
| Edge case handling (50) | 40 | 500-char truncation ✓, 20-card cap ✓, undo expiry ✓, long-text max-w 85%/90% ✓, 权限不足 ✓, Team missing ✓, 50-round cap ✓, stream-interrupt ✓, Team switch mid-flight ✓, undo persistence across overlay remount ✓. Missing: concurrent / multi-card-in-flight rules within a session (still unaddressed — see blindspot); slow-network beyond "错误/超时" still not differentiated. -10. |

### 2. User Experience (End User) — 220 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Information hierarchy (80) | 72 | Component 3 high-impact confirmation row now properly weighted ("实体标题 13px 500 text-primary（强权重），bizCode 12px text-tertiary（次权重，视觉分离）", line 240). Component 4 summary 14px 500 vs list 13px — adequate. Residual: in Component 2 the team-blocked and round-cap blocked notices both render as inline warning notice but there is no visual distinction between "blocking warning" (cannot send) and a normal transient notice — a user may miss that send is disabled. Quote: "发送阻断（Team 缺失） \| … inline notice warning-text #92400e". -8. |
| Interaction intuitiveness (80) | 72 | Bubble-hides-on-panel-open removes the contradictory click-to-close. AI-unavailable click now documented. Residual: the diff 浮层 appears on "继续对话补充" with no spec on how it dismisses (confirm button? click-outside? Esc?), so the user's path back to the card is unspecified. Quote: "继续对话补充 \| 后端解析增量 → diff 确认 → 更新卡片 state \| 展示 diff 浮层供确认，不静默覆盖". -8. |
| Accessibility (90) | 76 | role=log/role=dialog/role=alert/aria-live all present; contrast verified for body text. **Residual**: the doc claims "warning-title #d97706 仅用于图标与 14px 500 标题（large-text ≥3:1）" (line 464). Verified 3.07:1. WCAG "large text" = ≥24px regular or ≥18.66px bold. 14px 500 is **neither**; it is normal text requiring ≥4.5:1. So the 14px 500 警告标题 at 3.07:1 actually **fails WCAG AA** for its text size — the doc's "large-text" classification is incorrect. Component 6 titles ("AI 响应超时"/"AI 暂不可用") at 14px 500 warning-title `#d97706` are non-compliant. -10. No reduced-motion spec for 三点跳动 / slide / skeleton-stream animations (still missing). -4. |

### 3. Design Integrity (Designer) — 222 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Design system adherence (80) | 72 | Status-variant tokens now declared in the Design System table before use (lines 41–50) — the undefined-token problem is fixed. Component 2 textarea still spec'd `min-h 40px、max-h 120px` (line 158), contradicting project convention `min-h-[120px] resize-y`; the doc does not explicitly flag this as a deliberate deviation. -8 (deviation not called out). |
| Visual coherence (90) | 80 | All cards share surface/border/rounded-xl/shadow-1. Bubble/panel z-index conflict resolved via hide-on-open. Residual: the diff 浮层 (Component 3) and the leave-confirm Dialog (Component 2) are both "level-3 shadow" overlays, but the design never states stacking order when both could appear (e.g., user mid-对话补充 diff confirmation, then attempts route navigation). Cross-overlay precedence undefined. -5. Component 4 progress bar visual treatment (track color, height, radius) is undefined, so its coherence with other progress bars in the app (e.g., milestone progress) is unverified. -5. |
| State completeness (80) | 70 | Most states and transitions described. Residuals: (a) the diff 浮层 confirm step has no state (shown/confirmed/cancelled); (b) Component 2 "Team 切换 (teamChangedMidFlight)" state exists in Data Binding but **not in the States table** — the state row "发送阻断" covers Team missing, but mid-flight Team change is only an interaction, not a listed state, so its visual is unspecified beyond an inline warning notice. -10 (state missing from States table while present in Data Binding — cross-section inconsistency). |

### 4. Implementability (Developer) — 218 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Layout specificity (80) | 66 | Leave-confirm Dialog now fully spec'd (`max-w-sm`/p-4/buttons/focus/Tab). Component 3 undoCountdown source bound. **Residual**: (a) Component 3 **diff 浮层** ("展示 diff 浮层供确认", line 267) has **zero layout** — no width, no position, no fields, no buttons; a developer cannot build it. -8. (b) Component 4 progress bar "accent-light 填充" (line 301) — no track color, no height, no radius. -3. (c) Component 3 必填空字段 warning 竖条 ("左侧 warning 竖条", line 238) — no width/placement spec. -3. |
| Data binding explicit (80) | 72 | hasUncommittedCards, truncationNotice, undoCountdown, teamCtxMissing, sessionRoundCount, streamInterrupted, teamChangedMidFlight all bound. **Residual**: (a) the **diff 浮层** has no data binding (no `fieldDiff[]` field, no source); (b) Component 3 `previousValue` is bound but the "reverse operation" semantics (assignee restore / status restore) are described only in PRD — the design Data Binding does not document which API the reverse call uses, so the "撤回" action's API source is implicit. -8 (diff overlay orphan). |
| Interaction unambiguity (90) | 80 | Trigger→action→feedback chains mostly explicit; retry, AI-unavailable-click, truncation, send-block all complete. **Residual**: (a) Component 3 "继续对话补充" → "展示 diff 浮层供确认，不静默覆盖" — the action ("diff 确认") is vague: what does the user click in the diff overlay to accept/reject? No trigger in the diff overlay is defined (the overlay itself is undefined). -6. (b) Component 5 "点确认 \| 折叠本卡片，推送后续写/查卡片" — "推送" is not an event spec; no mechanism (message append? card insert at index?). -4. |

---

## Phase 3 — Cross-Dimension Coherence Check

- **Component 3 diff 浮层** is referenced by Interactions ("展示 diff 浮层供确认") and required by PRD UF-3 Validation Rule (line 184 of PRD), but appears in **no** layout, **no** state table, **no** data binding, **no** interaction-of-its-own. This is a cross-section gap touching Requirement Coverage, User Experience, Design Integrity, and Implementability simultaneously. The single largest residual flaw.
- **Component 2 Team-change-mid-flight** appears in Data Binding (`teamChangedMidFlight`) and Interactions but is **absent from the States table** — cross-section inconsistency between Data Binding and States of the same component.
- **Component 6 warning-title at 14px 500** declared as "large-text" in the Design System Accessibility section (line 464), but WCAG large-text definition excludes 14px 500 — internal specification inaccuracy that yields a real accessibility failure for the timeout/unavailable card titles.
- **Component 2 textarea min-h 40px** vs project convention `min-h-[120px]` — deviation not flagged.

---

## Phase 4 — Blindspot Hunt

1. **[blindspot] Concurrent in-flight cards within a single session** — The design now describes per-card states (预填/编辑中/校验失败/提交中) and a single `hasUncommittedCards` aggregate, but it does **not** define what happens when a user issues a second instruction that spawns a new card while a previous UF-3 card is still in 校验失败 or 提交中. Do two live UF-3 cards coexist? Does the second one block until the first resolves? Quote: "未提交卡片存在 \| hasUncommittedCards \| 任一 UF-3 处于 预填/编辑中/校验失败/提交中，或 UF-5 处于 待选择/已选择未确认 时为 true". Must improve: define concurrency rules — allow N concurrent uncommitted cards vs serialize — and the leave-confirm behavior when multiple are pending.

2. **[blindspot] Diff-overlay vs leave-confirm stacking precedence** — When the user is mid-对话补充 (diff 浮层 open on a UF-3 card) and attempts route navigation, both `hasUncommittedCards=true` (the UF-3 card) and the diff overlay are live. The design specifies neither whether the diff overlay blocks the leave-confirm Dialog nor which z-index wins. Quote: "继续对话补充 \| … \| 展示 diff 浮层供确认". Must improve: define stacking/precedence between transient overlays (diff) and route-guard dialogs.

3. **[blindspot] Reduced-motion for skeleton streaming / bubble hover / panel slide** — The doc mandates 150ms/200ms transitions (lines 100, 143) and a 三点跳动 "AI 思考中…" indicator, but provides no `prefers-reduced-motion` fallback. PRD Accessibility Requirements (line 374) covers contrast but not motion. Quote: "面板从右滑入；焦点移入输入框" (Component 1) and "切思考态" 三点跳动 (Component 2). Must improve: add reduced-motion spec (e.g., replace 三点跳动 with static "AI 思考中…" text; collapse slide to instant show).

4. **[blindspot] Streamed partial-card cleanup across UF-3/UF-4/UF-5** — Component 2 "流式中断" state says "半填充骨架卡丢弃" for the streaming bubble, and Component 3 has a corresponding "流式填充中断" state. But Component 4 (查询结果) and Component 5 (歧义消解) also render via the streaming path and have **no interrupt state**. If the stream drops mid-UF-4 or mid-UF-5, cleanup behavior is undefined. Quote: "流式返回 \| 卡片骨架（border-dashed 占位）先显示，字段增量填充". Must improve: extend stream-interrupt cleanup rule to all streamed card types, or state explicitly that only UF-3 streams incrementally.

---

## Score Summary

| Dimension | Score |
|-----------|-------|
| Requirement Coverage | 224 / 250 |
| User Experience | 220 / 250 |
| Design Integrity | 222 / 250 |
| Implementability | 218 / 250 |
| **Total** | **884 / 1000** |

Below target (950). The revision closed 10 of 12 attacks cleanly (权限不足 state, retry row, Team-missing block, 50-round block, leave-confirm Dialog, undoCountdown source, truncationNotice binding, AI-unavailable click, status-variant tokens, bubble z-index). Three residual drivers hold the score down:

1. **Component 3 diff 浮层** — referenced by Interactions and required by PRD, but has zero layout / state / data binding / interaction-of-its-own. Touches all four dimensions.
2. **Component 6 warning-title contrast mis-classification** — 14px 500 at 3.07:1 is treated as WCAG "large text" but is not; real AA failure for timeout/unavailable titles.
3. **Component 2 Team-change-mid-flight** — present in Data Binding and Interactions but absent from the States table — cross-section inconsistency.

Plus blindspots on concurrent in-flight cards, overlay stacking precedence, reduced-motion, and stream-interrupt cleanup for UF-4/UF-5.
