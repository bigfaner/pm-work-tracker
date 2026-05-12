# UI Design Evaluation Report -- Milestone Map

**Iteration**: 3
**Date**: 2026-05-12
**Document**: `docs/features/milestone-map/ui/ui-design.md`
**PRD Reference**: `docs/features/milestone-map/prd/prd-ui-functions.md`

---

## Overall Score: 81/100

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 23 | 25 |
| User Experience (End User) | 20 | 25 |
| Design Integrity (Designer) | 20 | 25 |
| Implementability (Developer) | 18 | 25 |

---

## Previous Issues Check

| # | Iteration-2 Issue | Status | Evidence |
|---|-------------------|--------|----------|
| 1 | Undefined design tokens (accent-light, elevation-2, elevation-3, accent-ring) | **Resolved** | Token Reference table (lines 19-27) now defines: accent-light=`#3b82f6`, accent-bg=`#eff6ff`, accent-ring=`#bfdbfe`, elevation-2 and elevation-3 with full box-shadow values. All five previously undefined tokens now have explicit CSS values. |
| 2 | No undo for drag-and-drop | **Resolved** | Interaction table (line 122) now specifies: completion animation + undo toast "MI-XXXX 已移至 [里程碑名]" with "撤销" button (5s auto-dismiss), clicking undo calls API to restore original milestone_key. |
| 3 | Delete confirmation dialog not designed | **Resolved** | New "Delete Confirmation Dialog" section (lines 295-337) with ASCII layout, states (Default/Submitting/Delete Error), interactions (confirm/cancel/overlay), and data binding. Danger variant styling specified. |
| 4 | Navigation position label mismatch | **Resolved** | Design line 34 now says "侧边栏导航'事项清单'和'甘特图'之间" matching PRD "独立页面，主导航'事项清单'和'甘特图'之间". |
| 5 | Orphan UI elements (refresh button, search box) | **Resolved** | Refresh button has interaction entry (line 126) and data binding (line 169). Search box has interaction entry (line 127) with debounce 300ms detail and data binding (line 170). |
| 6 | "Blank area" for drag-to-unbind is undefined | **Not resolved** | Line 123 still says "拖拽 MI 到空白区域 → 调用 API 置空 milestone_key". No visual drop zone, instructional text, or spatial definition of "blank area" has been added. A user cannot discover this action; a developer must guess what constitutes "blank area" in code. |
| 7 | No transition descriptions between states | **Not resolved** | State tables remain flat listings with no narrative or diagram for Loading -> Populated, Error -> retry -> Loading -> Populated flows. |
| 8 | UF-5 interaction chain incomplete | **Not resolved** | UF-5 interaction table (line 438) still says "选择里程碑 → 保存后更新 milestone_key" as a single entry. The actual chain (select dropdown -> form state update -> user clicks parent save -> validate -> API -> feedback) is not decomposed. This is misleading because "保存后" implies the API call is deferred but the trigger column presents it as immediate. |
| 9 | UF-3 status transition rules not enumerated from PRD | **Not resolved** | PRD specifies 4 status groups with explicit allowed/disallowed transitions and tooltip messages. Design interaction (line 290) says "校验转换合法性" generically but does not enumerate the per-state rules or tooltip text from the PRD. |
| 10 | Tooltip mouse-only, no keyboard trigger | **Not resolved** | Tooltip is triggered by "悬停里程碑节点" (line 120). No keyboard-accessible tooltip mechanism specified for users navigating via Tab + Arrow keys. |
| 11 | Focus management not specified for dialogs | **Not resolved** | ARIA table defines `role="dialog"` but no focus trap, initial focus placement, or programmatic focus restoration spec beyond "焦点回到触发的里程碑节点" (line 137). |
| 12 | UF-4/UF-5/UF-6 ARIA roles missing | **Not resolved** | ARIA table (lines 143-154) only covers UF-1 elements. UF-4 dropdown, UF-5 selector, and UF-6 sortable column header have no ARIA specifications. |

---

## Dimension 1: Requirement Coverage -- 23/25 (PM Perspective)

### UI Function Coverage: 8/8

All 6 UI functions from the PRD have corresponding design sections with placement, layout, states, interactions, and data binding:

| PRD Function | Design Section | Coverage |
|-------------|---------------|----------|
| UF-1 (Timeline page) | Component: 里程碑时间线页面（UF-1） | Full |
| UF-2 (Create/Edit modal) | Component: 创建/编辑里程碑弹窗（UF-2） | Full |
| UF-3 (Detail panel) | Component: 里程碑详情面板（UF-3） | Full |
| UF-4 (Items filter) | Component: 事项清单页里程碑筛选（UF-4） | Full |
| UF-5 (MI edit selector) | Component: 主事项编辑弹窗里程碑选择器（UF-5） | Full |
| UF-6 (Table column) | Component: 表格视图里程碑列（UF-6） | Full |

No gaps.

### Navigation Architecture Coverage: 4/4

PRD specifies 1 primary nav entry and 2 secondary pages. All covered:

- Primary: "里程碑图" at /milestones (line 34). Position now matches PRD: "事项清单"和"甘特图"之间. Resolved from iteration 2.
- Secondary: Milestone detail panel as overlay (UF-3, slide-over). Present.
- Secondary: Main item detail from MI click. Present (UF-1 interaction line 125).

### State Requirement Coverage: 7/8

All 18 PRD-defined states are present in design state tables:

| Component | PRD States | Design States | Delta |
|-----------|-----------|---------------|-------|
| UF-1 | Loading, Empty, Populated, No Permission, Error | All 5 | +0 |
| UF-2 | Create Mode, Edit Mode, Submitting, Validation Error | All 4 | +0 |
| UF-3 | Loading, Populated, Cancelled | All 3 | +0 |
| UF-4 | Default, Filtered | All 2 + Error | +1 |
| UF-5 | Default, Empty | All 2 + No Milestones + Error | +2 |
| UF-6 | Assigned, Unassigned | All 2 + Error | +1 |

**Deduction (-1):**
- UF-3 PRD specifies detailed status transition validation rules (lines 175-179 of PRD): 4 state groups with explicit allowed/disallowed transitions and specific tooltip messages like "未开始的里程碑不可直接标记为已完成". The design's interaction table (line 290) collapses this to "校验转换合法性 → 调用 API" with "非法：按钮禁用 + tooltip 说明原因". This is a generic description that does not enumerate the PRD's specific transition rules or carry forward the tooltip text. A developer implementing this would need to cross-reference the PRD manually rather than having the design serve as the authoritative specification.

### Edge Case Handling: 4/5

**Covered:**
- Empty states (all components)
- Permission denied (UF-1 No Permission state)
- Error states (all components including UF-4/5/6 added since iteration 1)
- Loading states with skeleton screens (all components)

**Still missing (-1):**
- **Long text overflow**: Milestone names can be up to 100 characters (PRD validation rule). The w-40 (~160px) node card has no truncation rule. No `text-overflow: ellipsis`, `line-clamp`, or `max-lines` specification. MI titles in h-8 compact rows also lack overflow handling. What happens when "里程碑发布 -- 第二阶段后端基础设施优化与性能调优冲刺" exceeds the card width is undefined.
- **Concurrent actions**: Two users dragging the same MI simultaneously has no conflict resolution. No optimistic locking, no 409 Conflict error state.
- **Slow network / timeout**: No maximum wait time, no progressive loading guidance, no timeout fallback beyond the static skeleton.

---

## Dimension 2: User Experience -- 20/25 (End User Perspective)

### Information Hierarchy: 7/8

**Strengths:**
- Clear page title (h1, 18px, font-semibold) as primary visual anchor.
- Toolbar (filters/search/zoom) as secondary level is logical and conventional.
- Node card information density is well-structured: status dot + name + completion, then date, then MI count.
- UF-3 detail panel has clear content sections: metadata cards, progress bar, status transition controls, MI list, danger zone.

**Deduction (-1):**
- UF-3 panel sections use ASCII-art "──" dividers (lines 258-267) but no spacing values, heading hierarchy, or background differentiation is specified. The section labels "状态切换", "关联事项 (3)", and "危险操作" carry no font-weight or heading-level distinction. A designer implementing this would not know whether these are h3, h4, or simply bold labels, or what vertical spacing separates them.

### Interaction Intuitiveness: 7/8

**Strengths:**
- Click-to-open patterns (node -> panel, create button -> modal) are conventional and well-understood.
- Zoom control button group (Week/Month/Quarter) is a standard pattern for timeline views.
- Dropdown filters and selectors are conventional.
- Drag-and-drop with undo toast (line 122) is now present, providing recoverability for accidental drags.
- Keyboard drag-and-drop mode (Space + arrows + Enter) is a strong accessibility feature.

**Deduction (-1):**
- **Drag-to-unbind has no discoverable affordance**: Line 123 specifies "拖拽 MI 到空白区域 → 调用 API 置空 milestone_key" but no visual drop zone, instructional text, hover hint, or "trash/unassign" target is defined. A user cannot discover this feature through visual cues alone. The drag-and-drop interaction for reassignment (line 122) has clear feedback (opacity change, border-dashed target, undo toast) but the unbind variant provides none of these affordances.

### Accessibility: 6/9

**Improvements since iteration 2:**
- Undo toast for drag-and-drop now includes screen-reader-accessible feedback via the existing aria-live region.

**Remaining deductions (-3):**

1. **Focus management incomplete (-1)**: ARIA table defines `role="dialog"` for the panel and modal, but no focus trap implementation is specified. The keyboard table says "Escape closes panel, focus returns to triggering milestone node" (line 137) but does not specify initial focus placement on panel open (which element receives focus first), focus trap boundaries (Tab wraps within panel while open), or programmatic focus restoration mechanism. WCAG 2.1 SC 2.4.3 (Focus Order) requires this.

2. **Tooltip is mouse-only (-1)**: The tooltip trigger is "悬停里程碑节点" (line 120) with no keyboard-accessible trigger mechanism. A keyboard user navigating via Tab/Arrow to a milestone node receives no equivalent of the hover tooltip. No mechanism (e.g., focus-triggered tooltip, keyboard shortcut, or aria-describedby) is specified.

3. **UF-4/UF-5/UF-6 ARIA roles missing (-1)**: The ARIA roles table (lines 143-154) covers only UF-1 elements. UF-4's milestone filter dropdown needs `role="listbox"` with `aria-label`. UF-5's milestone selector needs `role="combobox"` or `role="listbox"`. UF-6's sortable column header needs `aria-sort` attributes. Without these, the smaller components are invisible to assistive technology.

---

## Dimension 3: Design Integrity -- 20/25 (Designer Perspective)

### Design System Adherence: 7/8

**Strengths:**
- Correctly references design system stack (Tailwind v4 + Radix UI + CVA).
- Blue accent `#2563eb`, 13px body text, white card + `#f8fafc` background match the described system.
- Border separation pattern (not shadow) is consistently applied.
- Border-radius hierarchy documented: rounded-xl (dialog) > rounded-lg (button) > rounded-md (input).
- Token Reference table (lines 19-27) now provides explicit hex/shadow values for all tokens.

**Deduction (-1):**
- **Status color mapping ambiguity**: Design system section states "状态徽章系统：success=蓝, warning=橙, error=红" (line 14). But milestone status colors are defined as not_started=`#64748b`, in_progress=`#3b82f6`, completed=`#1d4ed8`, cancelled=`#cbd5e1` (line 91). The mapping between PRD status values and the design system's 3-category system (success/warning/error) is never made explicit. Is "completed" a "success" (which maps to the system blue `#2563eb`)? If so, why is it `#1d4ed8` instead of the accent `#2563eb`? Is "in_progress" also "success" (blue)? Two different shades of blue for what might be the same semantic category creates ambiguity. A designer implementing this cannot determine intent.

### Visual Coherence: 7/9

**Strengths:**
- Consistent border-radius usage across all 6 components (rounded-xl cards, rounded-md inputs, rounded-lg buttons).
- Consistent 13px font size and text hierarchy throughout.
- White background + border separation pattern is uniform.
- Filter dropdowns in UF-4 and selector in UF-5 reference the same style.
- Progress bar (h-2 rounded-full) consistent between UF-1 nodes and UF-3 panel.

**Deductions (-2):**
- **Badge shape inconsistency (-1)**: UF-4 milestone badge uses "rounded-full" (pill shape, line 364) while milestone node cards use "rounded-xl" and MI rows use "rounded-md". Three different border-radius values for what are essentially the same semantic element (a milestone identity indicator). No design rationale is documented.
- **Elevation stacking behavior undefined (-1)**: The detail panel (UF-3) uses z-40 and shadow-elevation-3 (line 271). The modal (UF-2) uses z-50 with overlay (line 170). When the edit modal opens from within the detail panel, the spec does not describe whether the panel dims, stays interactive, or is visually suppressed. The developer must infer the visual behavior of stacked overlays.

### State Completeness: 6/8

**Improvements since iteration 2:**
- Delete confirmation dialog now has full state coverage: Default, Submitting (spinner + disabled cancel), Delete Error (error text, button re-enabled) (lines 316-318). This was a significant gap in iteration 2.

**Remaining deductions (-2):**
- **No state transition narratives (-1)**: State tables remain flat listings. No narrative or diagram exists showing how the UI transitions between states (e.g., Loading -> Populated, Error -> user clicks retry -> Loading -> Populated). The behavior column describes when a state occurs but not what triggers the transition out of it. This is a persistent gap across all 3 iterations.
- **UF-2 missing network error state (-1)**: The Submitting state (line 214) shows a spinner and disables inputs. The interaction table says "失败：显示错误" (line 221) but the state table does not include a "Network Error" or "Submit Error" state. What does the error display look like? An inline error below the form? A toast? A replaced button state? The design does not specify, forcing the developer to invent the error presentation.

---

## Dimension 4: Implementability -- 18/25 (Developer Perspective)

### Layout Specificity: 6/8

**Strengths:**
- Key dimensions specified: w-40 nodes, h-8 MI rows, min-height 400px timeline, 400px dialog, w-[360px] panel, w-32 table column.
- ASCII art diagrams give clear layout concepts for all 6 components.
- Timeline layout algorithm (lines 64-80) with date-to-x formula, pxPerDay per zoom level, and overlap handling gives precise implementation guidance.
- Panel position (fixed right-0 top-0 h-full) and animation (translate-x 300ms ease-out) are specified.

**Deductions (-2):**
- **Responsive behavior missing (-1)**: No viewport breakpoint behavior is defined anywhere. Does the timeline scroll horizontally on narrow screens? Does the 360px panel push content or overlay on tablet widths? How does the 240px sidebar + timeline + 360px panel fit on screens narrower than ~1000px? No mobile or tablet adaptation is specified.
- **UF-4 dropdown width unspecified (-0.5) + UF-5 references existing code (-0.5)**: UF-4 says "h-10 rounded-md border-border-dark" (line 362) giving height but not width. UF-5 says "与现有'负责人'下拉框样式一致" (line 423) which is a reference to code not in this document. The developer must inspect existing components rather than having a self-contained spec. Combined: -1.

### Data Binding Explicit: 6/8

**Strengths:**
- All 6 components have data binding tables with UI Element -> Data Field -> Source mappings.
- Data transformations are noted (e.g., "API -> 颜色映射", "API -> Badge 组件").
- Delete confirmation dialog has its own data binding table (lines 333-336).
- Refresh button and search box now have binding entries (lines 169-170), resolving the iteration-1 orphan issue.

**Deductions (-2):**
- **No API endpoints or data contracts (-1)**: Every source column says "API" but no endpoint paths, HTTP methods, request parameters, or response shapes are defined. The developer must guess what API to call, what parameters to send, and what shape to expect. This might be acceptable if a separate API design doc is referenced, but no such reference exists.
- **Join path ambiguous for UF-4 and UF-6 (-1)**: UF-4 binding says "main_item → milestone" for the milestone name (line 395). UF-6 says "main_item → milestone.name" (line 491). In both cases, the join strategy is unspecified: is the milestone name embedded in the list/table API response, or does the frontend call a separate milestones endpoint and join client-side? This architectural decision affects implementation significantly.

### Interaction Unambiguity: 6/9

**Strengths:**
- UF-1 interaction table covers 9 triggers with explicit Action and Feedback columns (lines 119-127).
- UF-2 has clear trigger -> action -> feedback chains for confirm, cancel, and overlay click.
- UF-3 has complete chains including the now-designed delete confirmation dialog.
- UF-4 and UF-6 interactions are straightforward.

**Deductions (-3):**
- **UF-5 interaction chain is still incomplete (-2)**: Line 438 says "选择里程碑 → 保存后更新 milestone_key" as the action. This is misleading: "保存后" means the API call happens when the parent form's "Save" button is clicked, not when the dropdown selection changes. The actual chain is: select dropdown -> update form state -> user clicks parent dialog "Save" -> validate -> API call -> success/error feedback. None of these steps are decomposed. The developer must infer that the dropdown is a controlled form field within a larger form, and that its value is only persisted on parent-form save. This was flagged in iterations 1 and 2 and remains unaddressed.
- **"Blank area" spatially undefined (-1)**: Line 123 specifies "拖拽 MI 到空白区域 → 调用 API 置空 milestone_key" but "空白区域" is never defined spatially or visually. Is it any space outside a milestone node? The entire timeline background? A dedicated drop zone with visual indication? The developer must define this boundary condition in code without guidance. Flagged in all 3 iterations.

---

## Summary of Top Issues

1. **UF-5 interaction chain is fundamentally misleading** (Implementability: -2). The trigger "Select milestone" with action "Save and update milestone_key" conflates dropdown selection with API persistence. These are two different user actions at two different times. Three iterations without resolution suggests the spec author may not recognize this as a defect. What must improve: Decompose into "Select dropdown -> update form state" and separately document that the parent form's Save button triggers the API call. Add explicit note that this field is part of the parent form submission flow.

2. **Drag-to-unbind has zero discoverable affordance** (UX: -1, Implementability: -1). "Blank area" is undefined spatially, visually, and semantically across 3 iterations. Users cannot discover the unbind action. Developers cannot code a "blank area" drop target. What must improve: Define a visible unassign drop zone (e.g., a trash icon zone at timeline bottom or a "Release from milestone" contextual action), specify its visual appearance in all drag states, and add it to the data binding table.

3. **State transition narratives remain absent across 3 iterations** (Design Integrity: -1). All state tables are flat with no directional flow between states. A developer cannot determine what triggers Loading -> Populated, or Error -> retry -> Loading -> Populated from the state table alone. What must improve: Add a transition column or a separate "State Machine" section showing directional edges between states with trigger conditions.

---

## Improvement Priorities

| Priority | Dimension | Action |
|----------|-----------|--------|
| P0 | Implementability | Fix UF-5 interaction table: decompose "select dropdown" (updates form state) from "parent form save" (triggers API). Add note that milestone_key is part of the parent form submission. |
| P0 | UX + Implementability | Define drag-to-unbind drop zone: add a visible "unassign" target area (icon + label), specify its appearance during drag hover, and add it to the interaction table and data binding table. |
| P1 | Design Integrity | Add state transition flows: for each component, add a transition column or narrative showing directional edges (Loading -> Populated on success, Error -> Loading on retry, etc.). |
| P1 | Design Integrity | Add UF-2 submit error state to the state table: define visual treatment (inline error below form vs toast) and interaction (retry button behavior). |
| P1 | Requirement Coverage | Enumerate UF-3 status transition rules in the design. Carry forward the PRD's per-state allowed/disallowed transitions and tooltip messages into the interaction table or a dedicated validation section. |
| P2 | Accessibility | Add focus trap spec for dialog/panel: initial focus target, Tab wrap boundaries, focus restoration on close. |
| P2 | Accessibility | Add keyboard-triggered tooltip mechanism (e.g., focus triggers tooltip after 500ms delay, or Enter on node shows tooltip). |
| P2 | Accessibility | Add ARIA roles for UF-4 (listbox), UF-5 (combobox/listbox), UF-6 (aria-sort on column header). |
| P2 | Implementability | Add responsive breakpoint behavior: minimum viewport width, horizontal scroll behavior for timeline, panel behavior on narrow screens. |
| P2 | Requirement Coverage | Add text overflow rules: truncation strategy for milestone names in w-40 nodes and MI titles in h-8 rows. |
