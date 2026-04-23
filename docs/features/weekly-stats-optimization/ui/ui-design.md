---
created: 2026-04-23
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: 每周进展统计优化

## Design System

**Style: Tailwind UI** — Professional warmth, indigo accent.

### Core Tokens

| Token | Value | Usage |
|-------|-------|-------|
| Primary 600 | #4f46e5 | Accent, active states |
| Primary 50 | #eef2ff | Accent bg |
| Slate 900 | #0f172a | Stat numbers |
| Slate 600 | #475569 | Labels, secondary text |
| Slate 400 | #94a3b8 | Placeholder / disabled |
| Slate 200 | #e2e8f0 | Card borders |
| White | #ffffff | Card surface |
| Primary 400 | #818cf8 | Accent (activeSubItems card border) |
| Emerald 600 | #059669 | Success (newlyCompleted) |
| Red 600 | #dc2626 | Error / overdue |
| Amber 600 | #d97706 | Warning (blocked) |
| Slate 500 | #64748b | Paused card border |

**Typography:**
- Stat number: Inter 700, 28px, slate-900
- Card label: Inter 500, 12px, uppercase, letter-spacing 0.05em, slate-600
- Tooltip text: Inter 400, 13px, line-height 1.5

**Card:**
- bg white, border 1px solid slate-200, rounded-xl (12px)
- shadow: 0 1px 3px rgba(0,0,0,0.1)
- padding: 16px (p-4)

---

## Component: StatsBar

### Layout Structure

```
<StatsBar>                          ← grid container, gap-3, w-full (grid-cols-{n} per breakpoint)
  <StatCard × 7>                    ← min-w-0
    <div.card-inner>                ← flex col, gap-1, p-4
      <span.stat-number>            ← text-2xl font-bold
      <span.stat-label>             ← text-xs font-medium uppercase tracking-wide
    <Tooltip>                       ← absolute positioned, z-50
```

**Responsive grid:**

| Breakpoint | Layout | Tailwind |
|------------|--------|----------|
| ≥1280px (xl) | 7 columns, single row | `grid grid-cols-7 gap-3` |
| 768–1279px (md–lg) | 4+3 two rows, flex-wrap | `grid grid-cols-4 gap-3` (last 3 span naturally) |
| <768px (sm) | 2 columns, multi-row | `grid grid-cols-2 gap-3` |

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Skeleton pulse: number area replaced by `w-12 h-7 bg-slate-200 rounded animate-pulse`; label replaced by `w-16 h-3 bg-slate-100 rounded animate-pulse`. Wrap StatsBar in `<div aria-live="polite" aria-busy="true">` while loading. | Shown while API request is in-flight |
| Loaded | Actual number + label. Set `aria-busy="false"` on the wrapper — screen readers announce the region has updated. | After successful API response |
| Empty (0) | Number shows `0`, label normal | When count is zero for that state |
| Error | All numbers show `—` (em dash, slate-400); error banner above bar: `text-sm text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2` with text "加载失败，请刷新". Banner element must carry `role="alert"` (implies `aria-live="assertive"`) so screen readers announce it immediately. | API returns non-2xx or request timeout |

### Card Color Accents

Each card has a subtle left-border accent (4px) to aid quick scanning:

| Card | Accent Color | Rationale |
|------|-------------|-----------|
| 本周活跃 | Primary 400 (#818cf8) | Neutral highlight |
| 本周新完成 | emerald-600 (#059669) | Positive / success |
| 进行中 | primary-600 (#4f46e5) | Active work |
| 阻塞中 | amber-600 (#d97706) | Warning |
| 未开始 | slate-400 (#94a3b8) | Neutral / inactive |
| 暂停中 | Slate 500 (#64748b) | Paused |
| 逾期中 | Red 600 (#dc2626) | Risk / overdue |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Mouse enter card (desktop) | Start 300ms timer | — |
| 300ms elapsed | Show tooltip | Tooltip fades in (opacity 0→1, 150ms ease-out) |
| Mouse leave card | Cancel timer / hide tooltip | Tooltip fades out (150ms) |
| Click card (mobile, touch device) | Toggle tooltip open/closed | Tooltip appears immediately |
| Click outside any card (mobile) | Close active tooltip | Tooltip fades out |
| Tab focus on card (keyboard) | Show tooltip immediately (no delay) | Tooltip visible |
| Tab away from card | Hide tooltip | Tooltip fades out |
| Escape key | Close active tooltip | Tooltip fades out |
| Week selector changes while request in-flight | Cancel the in-flight request (AbortController); discard any response that arrives after cancellation; show Loading state for the new week | Skeleton replaces current numbers immediately; no stale data from the previous week is rendered |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 本周活跃 number | `stats.activeSubItems` | — |
| 本周新完成 number | `stats.newlyCompleted` | — |
| 进行中 number | `stats.inProgress` | — |
| 阻塞中 number | `stats.blocked` | — |
| 未开始 number | `stats.pending` | — |
| 暂停中 number | `stats.pausing` | — |
| 逾期中 number | `stats.overdue` | — |

---

## Component: StatCard Tooltip

### Layout Structure

```
<div.tooltip-wrapper>               ← position: relative, inline-block
  <div.card-trigger>                ← tabIndex=0, aria-describedby="tooltip-{id}"
    ... card content ...
  <div.tooltip-content>             ← id="tooltip-{id}", role="tooltip"
                                       position: absolute, bottom: calc(100% + 8px)
                                       left: 50%, transform: translateX(-50%)
                                       max-w-xs, bg-slate-900, text-white
                                       rounded-lg, px-3 py-2, text-xs
                                       shadow-lg, z-50
    <span> tooltip text </span>
    <div.tooltip-arrow>             ← CSS triangle, centered bottom
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Hidden | `opacity-0 pointer-events-none` | Default |
| Visible | `opacity-100` | After trigger condition met |

### Tooltip Placement

- Default: above card (`bottom: calc(100% + 8px)`)
- Flip to below when: `element.getBoundingClientRect().top` (relative to the viewport) < tooltip height + 8px (i.e., insufficient space above the visible viewport top)
- When flipped: `top: calc(100% + 8px)`, arrow points upward

### Tooltip Content (static config)

| Card | `id` | Text |
|------|------|------|
| 本周活跃 | `tooltip-active` | 本周有进展记录，或计划周期与本周重叠的子事项总数 |
| 本周新完成 | `tooltip-completed` | 本周内实际完成（actualEndDate 落在本周）的子事项数 |
| 进行中 | `tooltip-in-progress` | 状态为"进行中"且本周活跃的子事项数 |
| 阻塞中 | `tooltip-blocked` | 状态为"阻塞中"且本周活跃的子事项数 |
| 未开始 | `tooltip-pending` | 已创建但尚未启动（状态为 pending）且本周活跃的子事项数 |
| 暂停中 | `tooltip-pausing` | 状态为"暂停中"且本周活跃的子事项数 |
| 逾期中 | `tooltip-overdue` | 计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数 |

### Accessibility

- Card trigger element: `tabIndex={0}`, `aria-describedby="tooltip-{id}"`
- Tooltip node: `id="tooltip-{id}"`, `role="tooltip"`
- Tooltip hidden state uses `aria-hidden="true"` when not visible
- Focus ring on card: `focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2`

---

## Component: Responsive Layout

### Breakpoint Behavior

| Viewport | Grid | Card min-width |
|----------|------|----------------|
| ≥1280px | `grid-cols-7` | auto |
| 768–1279px | `grid-cols-4` (7 items → row 1: 4, row 2: 3) | auto |
| <768px | `grid-cols-2` | auto |

Implementation note: use CSS Grid with `grid-cols-{n}` at each breakpoint. Cards use `min-w-0` to prevent overflow. No horizontal scroll at any breakpoint.

### No-truncation Rule

Card inner content uses `truncate` only on the label if needed, but the number must always be fully visible. If a number exceeds 4 digits (e.g., 10000+), display as `9999+` with a tooltip showing the exact value.
