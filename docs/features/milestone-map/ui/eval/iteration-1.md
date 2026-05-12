# UI Design Evaluation Report — 里程碑图

**Iteration**: 1
**Date**: 2026-05-12
**Document**: `docs/features/milestone-map/ui/ui-design.md`
**PRD Reference**: `docs/features/milestone-map/prd/prd-ui-functions.md`

---

## Overall Score: 65/100

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 21 | 25 |
| User Experience (End User) | 13 | 25 |
| Design Integrity (Designer) | 16 | 25 |
| Implementability (Developer) | 15 | 25 |

---

## Dimension 1: Requirement Coverage — 21/25 (PM Perspective)

### UI Function Coverage: 8/8

All 6 UI functions from the PRD have corresponding design sections:
- UF-1 (Timeline page) → "里程碑时间线页面（UF-1）"
- UF-2 (Create/Edit modal) → "创建/编辑里程碑弹窗（UF-2）"
- UF-3 (Detail panel) → "里程碑详情面板（UF-3）"
- UF-4 (Items filter) → "事项清单页里程碑筛选（UF-4）"
- UF-5 (MI edit selector) → "主事项编辑弹窗里程碑选择器（UF-5）"
- UF-6 (Table column) → "表格视图里程碑列（UF-6）"

Each section includes placement, layout, states, interactions, and data binding. Full coverage.

### Navigation Architecture Coverage: 3/4

PRD specifies 1 primary nav entry and 2 secondary pages. The design covers all three targets:
- Primary: "里程碑图" at /milestones. Present.
- Secondary: 里程碑详情面板 as overlay. Present (UF-3 slide-over panel).
- Secondary: 主事项详情 from MI click. Present (UF-1 interaction table, "跳转主事项详情").

**Deduction (-1):**
- PRD specifies icon keyword "milestone/timeline" for the nav entry. The design does not specify which icon to use in the sidebar.
- PRD says position is "事项清单"和"甘特图"之间, but design says "事项清单"和"整体进度"之间 — positioning label mismatch.

### State Requirement Coverage: 8/8

All 18 PRD-defined states are present in the design:

| Component | PRD States | Design States | Match |
|-----------|-----------|---------------|-------|
| UF-1 | Loading, Empty, Populated, No Permission, Error | All 5 | Yes |
| UF-2 | Create Mode, Edit Mode, Submitting, Validation Error | All 4 | Yes |
| UF-3 | Loading, Populated, Cancelled | All 3 | Yes |
| UF-4 | Default, Filtered | All 2 | Yes |
| UF-5 | Default, Empty | All 2 (+ extra No Milestones) | Yes |
| UF-6 | Assigned, Unassigned | All 2 | Yes |

### Edge Case Handling: 2/5

**Covered:**
- Empty state (all components)
- Permission denied (UF-1 only)

**Missing:**
- **Long text overflow**: Milestone names up to 100 characters in w-40 (~160px) cards. No truncation rule (ellipsis, max lines). MI titles in h-8 compact rows — no overflow handling.
- **Slow network**: No timeout, progressive loading, or skeleton-to-content transition timing.
- **Concurrent actions**: Drag-and-drop reassignment has no conflict resolution. Two users dragging the same MI simultaneously — no optimistic locking, no error handling.
- **Large datasets**: 50+ milestones on the timeline. No virtualization, no pagination, no performance guidance.
- **Permission for UF-4/5/6**: Only UF-1 has a No Permission state. The other components assume milestones are always readable — no permission-denied handling for the filter, selector, or table column.

---

## Dimension 2: User Experience — 13/25 (End User Perspective)

### Information Hierarchy: 6/8

**Strengths:**
- Clear page title (h1, 18px, font-semibold) as primary heading.
- Toolbar as secondary level with filters/search/zoom is logical.
- Node card shows structured info (name + completion + date + count) at appropriate density.

**Deductions (-2):**
- Milestone node cards and MI rows are the two most important visual elements, but their visual differentiation relies only on size (w-40 vs h-8) and shape (card vs row). There is no heading hierarchy within the timeline — the time axis labels ("6月", "7月") compete visually with the node content for attention, and no weight distinction is specified.
- The detail panel mixes information sections, action buttons, and danger zone without clear visual section dividers. The layout uses "──" lines in the ASCII art but no spacing, heading, or background differentiation is specified.

### Interaction Intuitiveness: 5/8

**Strengths:**
- Click-to-open patterns (node → panel, create → modal) are conventional.
- Zoom control button group (周/月/季) is a standard pattern.
- Dropdown filters and selectors are conventional.

**Deductions (-3):**
- **Drag-to-unbind is non-obvious**: The design says "拖拽 MI 到空白区域 → 解绑 MI" but there is no visual drop zone, no "trash" or "unassign" target, no instructional text. A user would not discover this without being told. No affordance cue exists.
- **No undo mechanism**: Drag-and-drop reassignment is immediate ("调用 API 更新 milestone_key"). No undo toast, no confirmation, no reversible action. An accidental drag has immediate permanent consequence.
- **Drag feedback is minimal**: "opacity-50 + 移动跟随鼠标" and "目标节点 border-dashed accent" are specified, but no snapping animation, no position indicator, and no "drop to reassign" tooltip during drag.

### Accessibility: 2/9

**Critical gaps:**

1. **No keyboard navigation specified**: Zero mention of keyboard support. No focus management for the slide-over panel (how does focus trap work? where does focus land on open/close?). No tab order for timeline nodes. No Escape key behavior for modals or panels. No keyboard alternative for drag-and-drop.

2. **No ARIA roles or labels**: No aria-label, aria-labelledby, aria-describedby, or role attributes mentioned for any component. The modal, panel, dropdown, and timeline have no screen reader semantics.

3. **Contrast violations**: The design specifies `not_started=#94a3b8` for status dots. #94a3b8 on white ≈ 2.9:1 contrast ratio — fails WCAG AA (requires 4.5:1 for text, 3:1 for large text/UI components). Status dots at small sizes would be difficult to distinguish.

4. **Tooltip inaccessible**: "12px, dark bg" tooltip is mouse-only. No keyboard trigger specified.

5. **No live regions**: Drag-and-drop feedback ("两个里程碑完成度动画更新") is purely visual. No aria-live announcement for state changes (status change, deletion, reassignment).

6. **Loading states**: Skeleton screens are visual-only. No text equivalent for screen readers ("Loading milestones...").

---

## Dimension 3: Design Integrity — 16/25 (Designer Perspective)

### Design System Adherence: 6/8

**Strengths:**
- Correctly references the design system stack (Tailwind v4 + Radix UI + CVA).
- Blue accent `#2563eb`, 13px body text, white card + `#f8fafc` background — all match system description.
- Border separation (not shadow) for cards.
- Border-radius hierarchy: rounded-xl (dialog) > rounded-lg (button) > rounded-md (input).

**Deductions (-2):**
- **Undefined tokens**: "accent-light" is used for progress bar background ("bg accent-light") and "accent-bg" / "text-accent" for UF-4 badge, but these are not defined hex values. "shadow elevation-2" and "shadow-elevation-3" are not standard Tailwind classes and are not defined. "accent-ring" is used but not defined.
- **Color mapping inconsistency**: Design system section says "success=蓝, warning=橙, error=红" but the actual status colors are "not_started=#94a3b8, in_progress=#3b82f6, completed=#1d4ed8, cancelled=#cbd5e1". These are 4 distinct colors that don't cleanly map to the 3-category system. Is "completed" success (blue) or is it a different blue?

### Visual Coherence: 7/9

**Strengths:**
- Consistent border-radius usage across all 6 components.
- Consistent font size (13px) and text hierarchy.
- White background + border separation pattern is uniform.
- Filter dropdowns in UF-4 and selector in UF-5 follow same style.

**Deductions (-2):**
- **Badge shape inconsistency**: UF-4 milestone badge uses "rounded-full" (pill shape) while milestone node cards use "rounded-xl" and MI rows use "rounded-md". The badge shape difference is not documented as intentional.
- **Elevation system inconsistency**: The detail panel (UF-3) uses "shadow-elevation-3" for depth, but the modal (UF-2) achieves separation through a backdrop overlay. When the edit modal is opened from within the detail panel, the z-index layering (modal z-50 vs panel z-40) is specified, but the visual stacking behavior (does the panel dim? does it stay interactive?) is not described.

### State Completeness: 3/8

**Missing error states (-5):**
- UF-4 (Items filter): No error state. What if the milestones API fails? The dropdown shows nothing? Shows an error? Falls back to "全部"?
- UF-5 (MI edit selector): No error state. Same issue — if milestones fail to load, the dropdown is empty with no explanation.
- UF-6 (Table column): No error state. If milestone name lookup fails, what shows in the cell?

This is a **happy-path only design** for UF-4/5/6 — the rubric deducts 5 points.

**Missing transition descriptions:**
- No state transition diagram or narrative exists for any component. The design lists states as a flat table but never describes how the UI moves between them (e.g., Loading → Populated, Error → retry → Loading → Populated).
- UF-3: Delete action has no intermediate "Deleting" state (spinner while API call is in flight).

---

## Dimension 4: Implementability — 15/25 (Developer Perspective)

### Layout Specificity: 5/8

**Strengths:**
- Key dimensions specified: w-40 nodes, h-8 MI rows, min-height 400px timeline, 400px dialog, w-[360px] panel, w-32 table column.
- ASCII art diagrams give a clear layout concept for each component.

**Deductions (-3):**
- **Timeline layout algorithm undefined**: The ASCII art shows the concept but the actual positioning logic is not specified. How does a developer map `planned_date` to x-position on the time axis? What is the pixel-per-day ratio for each zoom level? How are overlapping milestone nodes handled? How are connector lines routed when MI lists under adjacent milestones overlap vertically?
- **Responsive behavior missing**: No mention of behavior at different viewport widths. Does the timeline scroll horizontally on mobile? Does the 360px panel push content or overlay? Does the 240px sidebar + timeline fit on small screens?
- **UF-4 dropdown width**: No width specified. "h-10 rounded-md" gives height but not width.
- **UF-5 relies on existing code**: "与现有'负责人'下拉框样式一致" — the developer must find and match an existing component rather than having a self-contained spec.

### Data Binding Explicit: 5/8

**Strengths:**
- UF-1, UF-2, UF-3 have complete data binding tables with UI element → data field → source mappings.
- Data transformations are noted (e.g., "API → 颜色映射", "API → AVG(main_items.completion)").

**Deductions (-3):**
- **No API endpoints specified**: Every source column says "API" but no endpoint paths, HTTP methods, request parameters, or response shapes are defined. The developer must guess or ask what API to call for each data need.
- **Join path ambiguous**: UF-4 says "main_item → milestone" for the milestone name on MI rows. Is this a joined response from a single endpoint, or does the frontend need to call a milestones endpoint and join client-side? Not specified.
- **Orphan elements**: The "刷新" (refresh) button in UF-1 toolbar appears in the layout diagram but has no interaction table entry and no data binding. The "搜索里程碑..." search box appears in the layout but has no interaction, no data binding, and no search behavior specification. These are UI elements with no defined behavior.

### Interaction Unambiguity: 5/9

**Strengths:**
- UF-1 interaction table covers 7 triggers with explicit action and feedback columns.
- UF-2 and UF-3 have clear trigger → action → feedback chains.
- UF-4 and UF-6 interactions are straightforward and adequately described.

**Deductions (-4):**
- **UF-5 interaction chain is incomplete**: The trigger is "选择里程碑" but the actual API call happens on form save, not on dropdown selection. The interaction table does not describe the full chain: select dropdown → update form state → user clicks save → validate → API call → success feedback / error feedback. The entry "保存后更新 milestone_key" is not a trigger-action pair, it's a side effect. (-2)
- **Delete confirmation dialog not designed**: UF-3 says "弹出确认弹窗" but this confirmation dialog is never designed — no layout, no states, no component spec. The developer must invent it. (-1)
- **"空白区域" undefined**: "拖拽 MI 到空白区域" — the developer must define what "blank area" means visually and in code. Is it any space outside a milestone node? Is there a dedicated drop zone? Is the entire timeline background a drop target? (-1)

---

## Summary of Top Issues

1. **Accessibility is critically lacking** (UX: 2/9). No keyboard navigation, no ARIA semantics, contrast violations, no screen reader support. This is not a minor gap — it makes the feature unusable for users with disabilities and potentially non-compliant.

2. **Happy-path only for UF-4/5/6** (Design Integrity: -5 deduction). Three components have zero error handling. If the milestones API fails, these components have no defined fallback behavior.

3. **Timeline layout algorithm is undefined** (Implementability: 5/8). The core visual component of this feature — the horizontal timeline with positioned nodes and connector lines — has no positioning logic specification. A developer cannot build this from the spec alone.

---

## Improvement Priorities

| Priority | Dimension | Action |
|----------|-----------|--------|
| P0 | Accessibility | Add keyboard navigation spec (focus trap, tab order, Escape behavior, drag-and-drop keyboard alternative). Define ARIA roles. Fix contrast for not_started status color. |
| P0 | Design Integrity | Add error states for UF-4, UF-5, UF-6. Add state transition descriptions for all components. |
| P1 | Implementability | Define timeline positioning algorithm (date-to-x mapping, zoom level pixel ratios, overlap handling). Specify API endpoints or at minimum the data contract. |
| P1 | UX | Replace "drag to blank area" with a visible unassign drop zone. Add undo mechanism for drag-and-drop reassignment. |
| P2 | Requirement Coverage | Add edge case handling for long text (truncation rules), slow networks (timeout behavior), and concurrent modifications (conflict resolution). |
| P2 | Design Integrity | Define undefined design tokens (accent-light, accent-bg, accent-ring, elevation-2, elevation-3). Resolve status color mapping with the 3-category badge system. |
