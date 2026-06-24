# UI Design Evaluation — Iteration 3

- **Document**: `docs/features/ai-copilot/ui/ui-design.md`
- **Source of truth**: `docs/features/ai-copilot/prd/prd-ui-functions.md`
- **Rubric**: `ui-web.md` (1000 pts, target 950)
- **Expert persona**: Senior UX Engineer
- **Iteration**: 3 (adversarial re-score of the revised document)

---

## Phase 0 — Iteration-2 Attacks: Addressed vs Residual

| # | Iteration-2 Attack | Status | Evidence in current doc |
|---|--------------------|--------|-------------------------|
| 1 | Component 3 **diff 浮层** had no layout/state/data binding | **Addressed** | New `#### diff 确认浮层` subsection (lines 295–304): position absolute anchored to card (`top:8px;right:8px`, z-60), max-w 320px, surface/border accent/rounded-lg/shadow level-2/p-3, content spec (title + diff rows with strikethrough old → new), buttons (应用 Primary / 丢弃 Ghost), keyboard (focus to 应用, Enter=应用, Esc=丢弃, Tab trapped). Data binding derived from existing `fieldSet` + pending增量. Cross-overlay precedence defined against leave-confirm Dialog (z-70 > z-60). |
| 2 | **Concurrent in-flight cards** blindspot | **Addressed** | New `#### 并发卡片策略` subsection (lines 219–223): multi-card coexistence, serial AI request (queued when thinking), write-submit independent across cards, optimistic lock defers to backend. |
| 3 | **WCAG warning title mis-classification** (14px 500 not large-text) | **Addressed** | Component 6 title now uses warning-text `#92400e` (line 409: "14px/500 非大字，需 ≥4.5:1，故用 warning-text `#92400e` 而非 warning-title `#d97706`"); Accessibility section (line 485) explicitly states "**14px/500 不属大字**" and "warning-title `#d97706` 仅用于图标（图形非文字，3:1 即可）". |
| 4 | **reduced-motion** missing for slide/skeleton/跳动 | **Addressed** | New `动效兜底（reduced-motion）` bullet (line 486): `@media (prefers-reduced-motion: reduce)` collapses all listed animations (panel slide, 三点跳动, skeleton fill, progress width, card state transitions) to instant change; state info does not depend on motion (思考态 has "AI 思考中…" text). |
| 5 | **teamChangedMidFlight** in Data Binding + Interactions but missing from States table | **Addressed** | New States row "Team 切换（in-flight）" (line 176) with visual (opacity-60 + lock icon, inline warning notice) and trigger condition (`teamCtx.bizKey` changes with pending/in-flight cards). |
| 6 | Component 2 **textarea** deviates from `min-h-[120px]` without note | **Addressed** | New 约定偏离说明 (line 159): explicit callout that `min-h 40px / max-h 120px` deliberately deviates from `frontend-components.md` TECH-frontend-001, with rationale (single-line growing input vs descriptive field) and scope (Copilot chat input only). |
| 7 | **diff-overlay vs leave-confirm stacking precedence** blindspot | **Addressed** | Line 304: "若 diff 浮层与离开确认 Dialog 同时存在，Dialog（z-70）压在 diff 浮层（z-60）之上，用户须先处理 Dialog；Dialog「确定离开」会连带丢弃未应用的 diff 增量". Explicit stacking + side-effect rule. |
| 8 | Component 4 **progress bar** has no track/height/radius | **Addressed** | Line 320: track bg `#e2e8f0`, h-1.5 (6px), rounded-full, w-full; fill bg accent-light `#3b82f6`, width=`{percent}%`, transition-width 300ms. All four properties specified. |
| 9 | Component 2 inline warning bar — no width/placement | **Residual (minor)** | The Team-blocked/round-blocked/team-change inline notices are described only as "inline notice warning-text" (lines 173–176). Width, padding, placement above textarea (full-width banner vs floating chip), and icon presence are not stated. Diff 浮层 got full spec; inline notices did not. |
| 10 | **Disambiguation push mechanism** — "推送" undefined | **Addressed** | Line 386: "本卡片折叠为'已选择：{title}'摘要行…后续写/查卡片（UF-3/UF-4）作为新的 AI 消息 **append 到消息列表末尾**（与普通 AI 回复同一消息流，非替换本卡片），`pendingIntent` + 所选 `candidate.bizKey` 注入新卡片作为预填来源". Explicit append + injection semantics. |
| 11 | **Stream-interrupt cleanup** undefined for UF-4/UF-5 | **Addressed** | Line 335 note: "查询结果（UF-4）与歧义消解（UF-5）均**原子返回**…因此不适用 UF-3 的「流式填充中断」清理规则；UF-2 的「流式中断」态仅针对 UF-3 写卡片的骨架增量填充场景". Explicitly scoped out with rationale. |
| 12 | **Assignee restore / reverse API** for undo implicit | **Partially Addressed** | Interaction (line 277) + Data Binding (line 289) document `previousValue` carries assignee/status; 撤销态持久化 block (line 293) documents sessionStorage persistence. However the **API endpoint name** for the reverse operation is still not bound (no "undoReverseApi" field, no source row). The semantics ("调后端反向操作") are documented per PRD, but the API source remains implicit. Treated as addressed at design-level (design is not API spec); residual minor. |

**Count**: 10 of 12 fully addressed; 2 with minor residuals (inline-notice layout #9, reverse-API source #12).

**No NEW contradictions introduced.** The revised entries are internally consistent: diff 浮层 z-60, leave-confirm Dialog z-70, panel z-50 — monotonic ordering; warning-text vs warning-title distinction applied uniformly; textarea deviation flagged without weakening global convention.

---

## Phase 1 — PRD Coverage Trace (UF → Component)

| PRD UI Function | Design Component | Coverage |
|-----------------|------------------|----------|
| UF-1 浮动气泡 | Component 1 | Full — 4 states, drag/click/AI-unavailable, hide-on-panel-open |
| UF-2 聊天面板 | Component 2 | Full — 9 states incl. all 3 blocking + team-change; leave-confirm Dialog; concurrency policy |
| UF-3 预填表单卡片 | Component 3 | Full — 权限不足 + 不可逆成功 + 撤回窗口过期 + 流式中断 + diff 浮层 |
| UF-4 查询结果卡片 | Component 4 | Full — summary + 20-card cap + progress bar + empty + truncation |
| UF-5 歧义消解卡片 | Component 5 | Full — 待选择/已选择/已确认 + append-push mechanism defined |
| UF-6 降级与超时提示 | Component 6 | Full — tokens, contrast-correct titles, route mapping |
| UF-7 首次引导卡片 | Component 7 | Full — fill-no-send rule explicit |

Navigation architecture: respected (no new pages; all overlays). Leave-confirm Dialog fully specified.

---

## Phase 2 — Rubric Scoring

### 1. Requirement Coverage (PM) — 240 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| UI function coverage (80) | 78 | All 7 UF mapped. Diff 浮层 now has full spec — closes the prior UF-3 sub-element gap. -2 for residual: inline-warning notice layout (width/placement/icon) unspecified for Team-blocked / round-blocked / Team-change states — the user-facing content is bound but its visual container is left to interpretation. Quote: "inline notice warning-text `#92400e`" (line 173) — no width, padding, banner-vs-chip form factor. |
| Navigation Architecture coverage (40) | 38 | All sitemap routes referenced; leave-confirm Dialog fully spec'd. -2: progress-bar milestone scenario marked "（里程碑场景）" (line 350) but trigger condition (when does a query produce progress{completed,total,percent} vs not?) is not specified — partial ambiguity on when the stat appears. |
| State requirement coverage (80) | 76 | All PRD states + blocking/stream-interrupt/team-change/undo-expiry/irreversible-success covered. -4: Component 5 has no 流式 / 中断 / 错误 state — the doc's atomic-return note (line 335) scopes out stream-interrupt, but Component 5 itself has no row for "AI returned disambiguation but the underlying entity list query failed" or "candidate selection timeout". Only 待选择/已选择/已确认. A network failure mid-disambiguation is unspecified. |
| Edge case handling (50) | 48 | 500-char truncation ✓, 20-card cap ✓, undo expiry + persistence ✓, long-text max-w 85%/90% ✓, 权限不足 ✓, Team missing ✓, 50-round cap ✓, stream-interrupt scoped ✓, Team switch mid-flight ✓, concurrency rules ✓, diff-overlay stacking ✓. -2: slow-network beyond 错误/超时 still not differentiated (the only network signal is the binary error/timeout). Minor. |

### 2. User Experience (End User) — 234 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Information hierarchy (80) | 76 | High-impact confirmation row weight explicit. -4: Team-blocked and round-blocked notices both render as "inline notice warning-text" (lines 173–174) with no visual weight distinction from the Team-change notice (also "inline warning notice", line 176). A user scanning cannot tell "send is disabled" (hard block) from "cards frozen, you may continue" (soft block). Same visual class for different severity. |
| Interaction intuitiveness (80) | 74 | Diff 浮层 dismiss semantics fully specified (Enter=应用, Esc=丢弃, Tab trapped). -6: the inline notice inside Component 2 for "上一条处理中…" (queued send, line 222) has no interaction model — does the user see the queued instruction? Can they cancel the queue? The new concurrency rule introduces a queued state with no UI affordance beyond an inline text hint. |
| Accessibility (90) | 84 | WCAG large-text definition now correctly cited; warning titles use warning-text; reduced-motion fallback complete. -6 residual: the diff 浮层 uses `position: absolute` anchored to card (line 299) — if the card scrolls within the message list while the overlay is open, the overlay detaches from its anchor with no `position: fixed` fallback or re-position rule documented. Screen-magnifier / low-vision users who scroll mid-diff lose the connection. Quote: "`position: absolute`，锚定到目标 UF-3 卡片右上角". |

### 3. Design Integrity (Designer) — 240 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Design system adherence (80) | 76 | Textarea deviation now explicitly called out with rationale and scope. Diff 浮层 uses border accent (not warning/error) — consistent with "incremental update" semantic. -4: the diff 浮层 uses `border accent #2563eb` (line 300) but Component 3's 校验错误条 uses `error-bg/error-text` and Component 6 uses `warning-bg`. The diff overlay is a "confirm before write" surface — using accent border conflates it visually with the Primary action color. There is no design-system rule justifying accent-border for an overlay (vs the pattern used for warning/error overlays). Minor visual-language inconsistency. |
| Visual coherence (90) | 84 | All cards share surface/border/rounded-xl/shadow-1. Diff 浮层 at z-60 + Dialog at z-70 + panel at z-50 = monotonic. -6: the inline notice element (used 3× in Component 2 for Team-blocked, round-blocked, Team-change) has no width spec — diff 浮层 got `max-w 320px` but the inline notice has no `max-w` / `w-full` / `inline-flex` declaration, so its rendered width is implementation-dependent. Inconsistent specificity for similar overlay-class elements. |
| State completeness (80) | 80 | All states and transitions described. Diff 浮层 has implicit shown/confirmed/cancelled transitions (via 应用/丢弃 buttons). teamChangedMidFlight now in States table. No state gaps. |

### 4. Implementability (Developer) — 238 / 250

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Layout specificity (80) | 76 | Diff 浮层 fully specified. Progress bar track/height/radius all set. 必填 warning 竖条 has `::before` absolute positioning spec. -4: the inline notice (Component 2, three uses) lacks `padding`, `rounded`, `bg`, `border`, `max-w` — a developer must guess whether it is a chip, a banner, or a tooltip. Compare diff 浮层 (`max-w 320px…p-3…rounded-lg`) which IS fully specified. Inconsistent treatment of similar elements. |
| Data binding explicit (80) | 76 | All overlay/card bindings present. Diff 浮层 derives from existing `fieldSet` + pending increment — no orphan. sessionStorage persistence for undo documented. -4: the queued-send inline notice ("上一条处理中…", line 222) has no Data Binding entry — its trigger (queue length > 0?) and dismissal (on prior request completion?) are not bound. New concurrency rule introduced a new UI element without a Data Binding row. |
| Interaction unambiguity (90) | 86 | Trigger→action→feedback chains explicit. Diff 浮层 has Enter/Esc/Tab/应用/丢弃 all defined. Push mechanism for UF-5 fully defined (append + injection). -4: Team-change Interactions row (line 187) lists two user options「按原 Team 继续」/「丢弃卡片」but neither has a Trigger column entry — where does the user click these? In the inline notice? As buttons? The action is documented but the trigger surface is unspecified. Quote: "用户可选「按原 Team 继续」（解冻卡片）/「丢弃卡片」（移除）" — no UI surface named. |

---

## Phase 3 — Cross-Dimension Coherence Check

- All 12 iter-2 attacks either fully resolved or reduced to minor residuals. No new cross-section contradictions.
- **z-index monotonicity** verified: panel z-50 < diff 浮层 z-60 < leave-confirm Dialog z-70. No conflict.
- **Warning color discipline** verified: warning-text `#92400e` for all 14px text; warning-title `#d97706` restricted to icon-only. Accessibility section explicitly states 14px/500 is not large-text.
- **Residual cluster**: the three "inline notice" elements in Component 2 (Team-blocked, round-blocked, Team-change) and the queued-send notice are the single largest remaining weak spot — they are text-content-bound but lack layout/width/padding specs, unlike the diff 浮层 which received full treatment. Touches Requirement Coverage, User Experience, Design Integrity, Implementability simultaneously.

---

## Phase 4 — Blindspot Hunt

1. **[blindspot] Diff-overlay scroll-detachment** — The diff 浮层 is `position: absolute` anchored to the UF-3 card (line 299: "`position: absolute`，锚定到目标 UF-3 卡片右上角"). The chat panel message区 is `overflow-y-auto` (line 152). If new messages arrive or the user scrolls while the diff 浮层 is open, the card moves but the spec does not state whether the overlay follows (via `position: fixed` on scroll), stays pinned to the last anchor position, or auto-dismisses. Must improve: specify scroll behavior — follow card via fixed-on-scroll, or dismiss on scroll with notice "diff 已暂存".

2. **[blindspot] Queued-send cancellation** — The concurrency policy (line 222) queues a new instruction when a prior one is thinking ("新指令入队等待（输入区 inline 提示'上一条处理中…'）"), but there is no UI to view the queue, cancel a queued instruction, or know how many are queued. Must improve: define queue visibility (count chip? expandable list?) and cancellation affordance.

3. **[blindspot] Diff-overlay multi-card conflict** — If the user is editing UF-3 card A (diff 浮层 open for A) and then issues a follow-up instruction that affects a different card B, does a second diff 浮层 appear on B simultaneously? The spec defines single-overlay vs Dialog precedence (line 304) but not multiple concurrent diff overlays. Must improve: state whether only one diff 浮层 may be open at a time, and what happens to an unconfirmed diff on A when a new diff on B arrives.

4. **[blindspot] Drag-bubble viewport edge on window resize** — Bubble position is stored in session state (line 104: "拖拽后位置存会话内 state，限制在视口范围内"). If the user resizes the browser smaller after dragging, the stored `{x,y}` may land outside the new viewport. No clamp-on-resize rule. Must improve: define resize handler that re-clamps `bubblePos` to the new viewport bounds (or hide bubble off-screen until re-dragged).

---

## Score Summary

| Dimension | Score |
|-----------|-------|
| Requirement Coverage | 240 / 250 |
| User Experience | 234 / 250 |
| Design Integrity | 240 / 250 |
| Implementability | 238 / 250 |
| **Total** | **952 / 1000** |

Above target (950). Iteration 3 closed 10 of 12 iter-2 attacks cleanly and reduced the remaining 2 to minor residuals. No new contradictions were introduced; z-index ordering and warning-color discipline are now internally consistent.

**Remaining residual cluster** (does not block target): the four "inline notice" elements in Component 2 (Team-blocked, round-blocked, Team-change, queued-send) lack the layout specificity that the diff 浮层 received — width, padding, banner-vs-chip form factor, and (for queued-send) Data Binding are implementation-dependent.

**Blindspots**: diff-overlay scroll-detachment, queued-send cancellation, multi-card concurrent diff overlays, drag-bubble resize clamp. None blocks implementation; all are edge-case refinements.
