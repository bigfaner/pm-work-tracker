# UI Design Evaluation Report -- Milestone Map

**Iteration**: 2
**Date**: 2026-05-12
**Document**: `docs/features/milestone-map/ui/ui-design.md`
**PRD Reference**: `docs/features/milestone-map/prd/prd-ui-functions.md`

---

## Overall Score: 76/100

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 22 | 25 |
| User Experience (End User) | 19 | 25 |
| Design Integrity (Designer) | 18 | 25 |
| Implementability (Developer) | 17 | 25 |

---

## Previous Issues Check

| # | Iteration-1 Issue | Status | Evidence |
|---|-------------------|--------|----------|
| 1 | No keyboard navigation specified | **Resolved** | New "Keyboard Navigation" table (lines 119-131) with Tab, Arrow, Enter, Escape, Space mappings. Keyboard drag-and-drop mode with aria-live announcements added. |
| 2 | No ARIA roles or labels | **Resolved** | New "ARIA Roles" table (lines 133-144) covering timeline container, nodes, MI rows, dropdowns, panels, modals, drag state, and search box. |
| 3 | Contrast violation for not_started status color | **Resolved** | Changed from `#94a3b8` (2.9:1) to `#64748b` (4.6:1 on white) with WCAG AA notation in the node card spec (line 81). |
| 4 | No screen reader loading announcement | **Resolved** | Loading state now includes `aria-busy="true"` + `sr-only` hidden text "Loading milestone data" (line 99). |
| 5 | No error states for UF-4, UF-5, UF-6 | **Resolved** | UF-4 Error: "Milestone: Load failed" (text-error), disabled selection (line 327). UF-5 Error: "Load failed" with preserved binding (line 379). UF-6 Error: cell shows "--" + tooltip (line 425). |
| 6 | No timeline positioning algorithm | **Resolved** | New "Timeline Layout Algorithm" section (lines 54-70) with date-to-x formula, pxPerDay per zoom level table, and overlap handling logic. |
| 7 | Undefined design tokens (accent-light, elevation-2, elevation-3, accent-ring) | **Partially resolved** | Tokens are still used without hex definitions (lines 82-84: "accent-light", "elevation-2", "accent-ring", line 261: "elevation-3"). The spec references them but never defines their values. |
| 8 | No undo for drag-and-drop | **Not resolved** | Drag-and-drop reassignment remains immediate ("Call API to update milestone_key", line 112). No undo toast, no confirmation, no reversible action mechanism added. |
| 9 | "Blank area" for drag-to-unbind is undefined | **Not resolved** | Line 113 still says "Drag MI to blank area --> Call API to clear milestone_key" with no visual drop zone, instructional text, or spatial definition of "blank area." |
| 10 | Delete confirmation dialog not designed | **Not resolved** | UF-3 interactions (line 281) still reference "Pop up confirmation dialog" but the confirmation dialog has no layout, states, or component spec. |
| 11 | No transition descriptions between states | **Not resolved** | State tables remain flat listings. No narrative or diagram showing Loading -> Populated, Error -> retry -> Loading -> Populated flows. |
| 12 | Orphan UI elements (refresh button, search box had no interaction/binding) | **Partially resolved** | Refresh button now has interaction entry (line 116) and data binding (line 159). Search box now has interaction entry (line 117) and data binding (line 160). However, the search box interaction says "Filter milestone nodes" with "debounce 300ms, non-matching nodes hidden, matching nodes highlighted" -- this is a richer behavior than initially specified but still lacks detail on what "highlight" means visually. |
| 13 | Navigation position label mismatch | **Not resolved** | PRD says "Between 'Items List' and 'Gantt Chart'" but design still says "Between 'Items List' and 'Overall Progress'" (line 24). |
| 14 | UF-5 interaction chain incomplete | **Not resolved** | UF-5 interaction table (lines 384-386) still says "Save and update milestone_key" as action, which is not a trigger-action-feedback chain. The full chain (select dropdown -> update form state -> user clicks save -> validate -> API call -> feedback) is still absent. |

---

## Dimension 1: Requirement Coverage -- 22/25 (PM Perspective)

### UI Function Coverage: 8/8

All 6 UI functions from the PRD have corresponding design sections:
- UF-1 (Timeline page) -> "Component: Milestone Timeline Page (UF-1)"
- UF-2 (Create/Edit modal) -> "Component: Create/Edit Milestone Modal (UF-2)"
- UF-3 (Detail panel) -> "Component: Milestone Detail Panel (UF-3)"
- UF-4 (Items filter) -> "Component: Items List Page Milestone Filter (UF-4)"
- UF-5 (MI edit selector) -> "Component: Main Item Edit Modal Milestone Selector (UF-5)"
- UF-6 (Table column) -> "Component: Table View Milestone Column (UF-6)"

Each section contains placement, layout, states, interactions, and data binding. Full coverage.

### Navigation Architecture Coverage: 3/4

PRD specifies 1 primary nav entry and 2 secondary pages. Design covers:
- Primary: "Milestone Map" at /milestones. Present (line 24).
- Secondary: Milestone detail panel as overlay. Present (UF-3 slide-over).
- Secondary: Main item detail from MI click. Present (UF-1 interaction, line 115).

**Deduction (-1):**
- PRD specifies navigation position as "Between 'Items List' and 'Gantt Chart'" but the design says "Between 'Items List' and 'Overall Progress'" (line 24). This position mismatch was flagged in iteration 1 and remains unresolved. The developer may place the nav entry incorrectly.
- No icon specification for the nav entry beyond what the PRD provides. The design does not specify which icon renders in the sidebar.

### State Requirement Coverage: 7/8

All 18 PRD-defined states are present in the design, plus additional error states for UF-4, UF-5, UF-6 added since iteration 1:

| Component | PRD States | Design States | Match |
|-----------|-----------|---------------|-------|
| UF-1 | Loading, Empty, Populated, No Permission, Error | All 5 | Yes |
| UF-2 | Create Mode, Edit Mode, Submitting, Validation Error | All 4 | Yes |
| UF-3 | Loading, Populated, Cancelled | All 3 | Yes |
| UF-4 | Default, Filtered | All 2 + Error (new) | Yes |
| UF-5 | Default, Empty | All 2 + No Milestones + Error (new) | Yes |
| UF-6 | Assigned, Unassigned | All 2 + Error (new) | Yes |

**Deduction (-1):**
- UF-3 PRD specifies validation rules for status transitions with 4 state groups (not_started, in_progress, completed, cancelled) and explicit allowed/disallowed transitions with tooltip messages. The design's state table (lines 269-273) only has Loading, Populated, and Cancelled states. The interaction table mentions "Validate transition legality" (line 280) but does not enumerate the transition rules or their UI behavior (which buttons show, which are disabled, what tooltip text). The PRD is quite specific about this -- the design defers it to a vague "validate" step.

### Edge Case Handling: 4/5

**Covered since iteration 1:**
- Empty state (all components)
- Permission denied (UF-1)
- Error states (all components now -- UF-4, UF-5, UF-6 added)
- Loading states (all components)

**Still missing (-1):**
- **Long text overflow**: Milestone names up to 100 characters in w-40 (~160px) cards. No truncation rule (text-overflow ellipsis, max-lines, line-clamp). MI titles in h-8 compact rows have no overflow handling. The design does not specify what happens when "Milestone Release -- Phase 2 Backend Infrastructure Optimization and Performance Tuning Sprint" exceeds the card width.
- **Concurrent actions**: Drag-and-drop reassignment has no conflict resolution. Two users dragging the same MI simultaneously -- no optimistic locking, no error state for 409 Conflict.
- **Slow network / timeout**: No timeout or progressive loading guidance. The Loading state shows a skeleton but no maximum wait time or fallback behavior.

---

## Dimension 2: User Experience -- 19/25 (End User Perspective)

### Information Hierarchy: 7/8

**Strengths:**
- Clear page title (h1, 18px, font-semibold) as primary heading.
- Toolbar as secondary level with filters/search/zoom is logical.
- Node card shows structured info (name + completion + date + count) at appropriate density.
- UF-3 detail panel has clear sections (metadata cards, progress bar, status transition, MI list, danger zone).

**Deduction (-1):**
- The detail panel (UF-3) mixes information sections, action buttons, and danger zone with ASCII-art "──" lines but no spacing, heading hierarchy, or background differentiation is specified. The sections "Status Switching", "Associated Items (3)", and "Dangerous Operations" use the same visual weight (lines 248-257). There is no heading level or font-weight distinction among these section labels.

### Interaction Intuitiveness: 6/8

**Strengths:**
- Click-to-open patterns (node -> panel, create -> modal) are conventional.
- Zoom control button group (Week/Month/Quarter) is a standard pattern.
- Dropdown filters and selectors are conventional.
- Keyboard drag-and-drop mode added (Space to enter, arrows to select target, Enter to confirm, Escape to cancel) -- a significant improvement.

**Deductions (-2):**
- **Drag-to-unbind remains non-obvious**: "Drag MI to blank area to unbind" (line 113) still has no visual drop zone, no "trash" or "unassign" target, no instructional text. A user would not discover this without being told. No affordance cue exists. This was flagged in iteration 1 and not addressed.
- **No undo mechanism**: Drag-and-drop reassignment is immediate ("Call API to update milestone_key", line 112). No undo toast, no confirmation, no reversible action. An accidental drag has immediate permanent consequence. This was flagged in iteration 1 and not addressed.

### Accessibility: 6/9

**Improvements since iteration 1:**
- Keyboard navigation table added (lines 119-131) with Tab, Arrow, Enter, Escape, Space mappings.
- ARIA roles table added (lines 133-144) covering all major components.
- Keyboard drag-and-drop mode with aria-live announcements added (line 131).
- Loading state now has `aria-busy="true"` + `sr-only` text (line 99).
- not_started color fixed to `#64748b` (4.6:1 on white, WCAG AA) (line 81).

**Remaining deductions (-3):**
- **Focus management incomplete**: The ARIA table defines `role="dialog"` for the panel and modal, but there is no spec for focus trap implementation, initial focus placement on open, or focus restoration on close. The keyboard table says "Escape closes panel, focus returns to triggering milestone node" (line 127) but does not specify how focus is programmatically moved (`.focus()`, `autofocus`, etc.).
- **Tooltip still mouse-only**: "12px, dark bg" tooltip (line 110) has no keyboard trigger. A keyboard user navigating to a milestone node and pressing... what? There is no keyboard-triggered tooltip mechanism specified.
- **Drag-and-drop live region is incomplete**: The aria-live region for drag state (line 143) says "broadcasts drag source, target, result" but the Keyboard Navigation section only specifies announcements for the keyboard drag mode. Mouse-based drag-and-drop (line 112) has no aria-live announcement. A screen reader user observing a mouse user's drag action would not be notified.
- **UF-4/UF-5/UF-6 ARIA not specified**: The ARIA roles table covers UF-1 components but does not mention UF-4's dropdown, UF-5's selector, or UF-6's sortable column header. These are smaller components but still need `role="listbox"` / `aria-sort` semantics.

---

## Dimension 3: Design Integrity -- 18/25 (Designer Perspective)

### Design System Adherence: 5/8

**Strengths:**
- Correctly references design system stack (Tailwind v4 + Radix UI + CVA).
- Blue accent `#2563eb`, 13px body text, white card + `#f8fafc` background match system.
- Border separation (not shadow) for cards.
- Border-radius hierarchy: rounded-xl (dialog) > rounded-lg (button) > rounded-md (input).

**Deductions (-3):**
- **Undefined tokens persist**: "accent-light" (line 82, 264), "accent-bg" / "text-accent" (line 312), "shadow elevation-2" (line 83), "shadow-elevation-3" (line 261), "accent-ring" (line 84) are still used without hex or CSS definitions. Flagged in iteration 1, partially addressed but not resolved. A developer or designer implementing this must guess at these values. (-2 for vague language)
- **Status color mapping inconsistency**: Design system section says "success=blue, warning=orange, error=red" (line 14) but status colors are "not_started=#64748b, in_progress=#3b82f6, completed=#1d4ed8, cancelled=#cbd5e1" (line 81). These are 4 distinct colors that do not map cleanly to the 3-category system. Is "completed" a "success" (which would be the primary blue `#2563eb`)? Or is it `#1d4ed8` which is a darker blue? The mapping is ambiguous. (-1)

### Visual Coherence: 7/9

**Strengths:**
- Consistent border-radius usage across all 6 components.
- Consistent font size (13px) and text hierarchy.
- White background + border separation pattern is uniform.
- Filter dropdowns in UF-4 and selector in UF-5 follow same style.

**Deductions (-2):**
- **Badge shape inconsistency**: UF-4 milestone badge uses "rounded-full" (pill shape, line 312) while milestone node cards use "rounded-xl" and MI rows use "rounded-md". This shape difference is not documented as intentional. A designer implementing the system would not know whether this is a deliberate design choice or an oversight.
- **Elevation system inconsistency**: The detail panel (UF-3) uses "shadow-elevation-3" for depth (line 261), but the modal (UF-2) achieves separation through a backdrop overlay (line 170). When the edit modal is opened from within the detail panel, the z-index layering (modal z-50 vs panel z-40) is specified, but the visual stacking behavior (does the panel dim? does it stay interactive?) is not described.

### State Completeness: 6/8

**Improvements since iteration 1:**
- Error states added for UF-4, UF-5, UF-6 (happy-path-only deduction resolved).

**Remaining deductions (-2):**
- **No state transition descriptions**: State tables remain flat listings. No narrative or diagram exists showing how the UI moves between states (e.g., Loading -> Populated, Error -> retry -> Loading -> Populated). The developer must infer transition logic from the behavior column alone.
- **UF-3 delete action has no intermediate state**: "Pop up confirmation dialog -> Confirm and delete -> Panel closes + Timeline refreshes" (line 281). There is no "Deleting" state (spinner while API call is in flight). What happens if the delete API fails? No error state is defined for the delete action.
- **UF-2 missing network error state**: The "Submitting" state shows a spinner, but the interaction table says "Failure: Display error" (line 211) without defining what that error display looks like. Is it an inline error below the form? A toast? A dialog? The state table does not include a "Network Error" state.

---

## Dimension 4: Implementability -- 17/25 (Developer Perspective)

### Layout Specificity: 6/8

**Strengths:**
- Key dimensions specified: w-40 nodes, h-8 MI rows, min-height 400px timeline, 400px dialog, w-[360px] panel, w-32 table column.
- ASCII art diagrams give clear layout concepts.
- Timeline layout algorithm now specified (lines 54-70) with date-to-x formula, pxPerDay per zoom level, and overlap handling.

**Deductions (-2):**
- **Responsive behavior missing**: No mention of behavior at different viewport widths. Does the timeline scroll horizontally on mobile? Does the 360px panel push content or overlay? Does the 240px sidebar + timeline fit on small screens? No breakpoint behavior defined.
- **UF-4 dropdown width unspecified**: "h-10 rounded-md" gives height but not width. Developer must guess or inspect the existing filter width.
- **UF-5 relies on existing code**: "Same style as existing 'Assignee' dropdown" (line 369) -- the developer must find and match an existing component rather than having a self-contained spec. This is a reference to code not in the document.

### Data Binding Explicit: 6/8

**Strengths:**
- UF-1, UF-2, UF-3 have complete data binding tables with UI element -> data field -> source mappings.
- Data transformations are noted (e.g., "API -> color mapping", "API -> AVG(main_items.completion)").
- Refresh button and search box now have data binding entries (lines 159-160), resolving the iteration-1 orphan issue.

**Deductions (-2):**
- **No API endpoints specified**: Every source column says "API" but no endpoint paths, HTTP methods, request parameters, or response shapes are defined. The developer must guess or ask what API to call for each data need. This is acceptable if there is a separate API design doc, but no reference to one is provided.
- **Join path ambiguous for UF-4**: "main_item -> milestone" for the milestone name on MI rows (line 342). Is this a joined response from a single endpoint, or does the frontend need to call a milestones endpoint and join client-side? Not specified.
- **Join path ambiguous for UF-6**: "main_item -> milestone.name" (line 438). Same issue -- is the milestone name embedded in the table response, or does the frontend need a separate lookup?

### Interaction Unambiguity: 5/9

**Strengths:**
- UF-1 interaction table covers 9 triggers with explicit action and feedback columns.
- UF-2 and UF-3 have clear trigger -> action -> feedback chains.
- UF-4 and UF-6 interactions are straightforward and adequately described.

**Deductions (-4):**
- **UF-5 interaction chain is still incomplete**: The trigger "Select milestone" has action "Save and update milestone_key" (line 385) but this is misleading -- the actual API call happens when the user clicks the parent form's "Save" button, not on dropdown selection. The interaction table does not describe the full chain: select dropdown -> update form state -> user clicks parent save -> validate -> API call -> success/error feedback. This was flagged in iteration 1 and not addressed. (-2)
- **Delete confirmation dialog still not designed**: UF-3 says "Pop up confirmation dialog" (line 281) but this confirmation dialog has no layout, no states, no component spec. The developer must invent it. Was flagged in iteration 1 and not addressed. (-1)
- **"Blank area" still undefined**: "Drag MI to blank area" (line 113) -- the developer must define what "blank area" means visually and in code. Is it any space outside a milestone node? Is there a dedicated drop zone? Is the entire timeline background a drop target? Was flagged in iteration 1 and not addressed. (-1)

---

## Summary of Top Issues

1. **Drag-to-unbind UX remains broken** (UX: -2, Implementability: -1). No visual drop zone, no instructional text, no affordance. Users cannot discover this feature. Developers cannot implement "blank area" without guessing.

2. **Delete confirmation dialog is an underspecified critical path** (Implementability: -1, Design Integrity: -1). Referenced but never designed. A delete action with confirmation is a standard pattern that needs layout, states, and interaction spec.

3. **Undefined design tokens create implementation ambiguity** (Design Integrity: -2). Five tokens (accent-light, accent-bg, accent-ring, elevation-2, elevation-3) are referenced but never given hex/CSS values. Designers and developers must guess.

---

## Improvement Priorities

| Priority | Dimension | Action |
|----------|-----------|--------|
| P0 | UX | Add visible "unassign" drop zone (trash icon or dedicated area) with tooltip "Release MI from milestone". Add undo toast for drag-and-drop reassignment (5-second timeout, "Undo" button). |
| P0 | Implementability | Design the delete confirmation dialog (layout, states, interactions). Define "blank area" spatially for drag-to-unbind. Fix UF-5 interaction chain to reflect that selection updates form state, not API. |
| P1 | Design Integrity | Define all undefined design tokens with hex values. Add state transition narratives for key flows (Loading -> Populated, Error -> retry). Add a "Deleting" intermediate state for UF-3. |
| P1 | Requirement Coverage | Fix navigation position label to match PRD ("Between 'Items List' and 'Gantt Chart'"). Add UF-3 status transition validation rules from PRD (per-state allowed/disallowed transitions with tooltip messages). |
| P2 | Accessibility | Add focus trap spec for modal/panel. Add keyboard-triggered tooltip mechanism. Add aria-live announcements for mouse-based drag. Add ARIA roles for UF-4/UF-5/UF-6. |
| P2 | Implementability | Add responsive behavior spec (mobile timeline, panel behavior). Add API endpoint references or data contract definitions. Specify join strategy for UF-4/UF-6. |
