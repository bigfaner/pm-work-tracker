---
date: "2026-04-27"
doc_dir: "docs/features/decision-log/ui/"
iteration: 2
target_score: 80
evaluator: Claude (automated, adversarial)
---

# UI Design Eval — Iteration 2

**Score: 84/100** (target: 80)

```
┌─────────────────────────────────────────────────────────────────┐
│                    UI DESIGN QUALITY SCORECARD                   │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension / Perspective      │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 1. Requirement Coverage (PM) │  23      │  25      │ ✅         │
│    UI function coverage      │  8/8     │          │            │
│    State requirement coverage│  8/8     │          │            │
│    Edge case handling        │  7/9     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 2. User Experience (User)    │  20      │  25      │ ⚠️         │
│    Information hierarchy     │  7/8     │          │            │
│    Interaction intuitiveness │  7/8     │          │            │
│    Accessibility             │  6/9     │          │            │
├──────────────────────────────┼──────────┬────────────┤
│ 3. Design Integrity (Design) │  20      │  25      │ ⚠️         │
│    Design system adherence   │  7/8     │          │            │
│    Visual coherence          │  7/9     │          │            │
│    State completeness        │  6/8     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. Implementability (Dev)    │  21      │  25      │ ✅         │
│    Layout specificity        │  7/8     │          │            │
│    Data binding explicit     │  7/8     │          │            │
│    Interaction unambiguity   │  7/9     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  84      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Edge cases (PM) | No handling for concurrent actions: another user publishes or deletes a draft while the current user is editing it. The draft-not-found case is covered but race conditions during active editing are not. | -2 pts (Edge cases) |
| Content expand (User) | Click-to-expand on timeline content still lacks a visible affordance beyond "Cursor pointer, hover text-primary." No "展开" link, chevron, or indicator that truncated text is interactive. | -1 pt (Interaction intuitiveness) |
| Accessibility (User) | Tertiary text `#94a3b8` on `#ffffff` is ~3.0:1 contrast ratio, failing WCAG AA (4.5:1 for normal text). No reduced-motion alternative for skeleton `animate-pulse`. No `aria-label` specified for the [编辑] ghost button. | -3 pts (Accessibility) |
| Timeline item layout (User) | Row 3 combines Tags + Creator + Edit on one line without visual separation between three different information types. | -1 pt (Information hierarchy) |
| Badge variant mapping (Design) | Category badge "default (slate)" is not mapped to a specific design system badge token. The design system defines status badge variants (success/warning/error) but no "slate" or "default" variant. | -1 pt (Design system adherence) |
| Skeleton radius (Design) | Skeleton rows use `rounded-lg` (8px) inside a `rounded-xl` (12px) card, creating an inconsistency in border-radius rhythm within the same component. | -1 pt (Visual coherence) |
| Timeline rail (Design) | The dot-and-line rail is a new visual element with no precedent in the existing design system. No guidance on how it integrates with the page's existing border-delineated card pattern. | -1 pt (Visual coherence) |
| Success feedback (Design) | No "Success" state or feedback after save/publish completes. Interactions say "close dialog, refresh timeline" but no toast, success message, or visual confirmation is specified. | -2 pts (State completeness) |
| Expanded content height (Dev) | No max-height or scroll behavior specified for expanded content. A 2000-character decision could render as 40+ lines with no constraint. | -1 pt (Layout specificity) |
| Recent tags data source (Dev) | The recent tags dropdown is described in layout and interactions but has no entry in the DecisionFormDialog Data Binding table. The data source (API? localStorage?) is unspecified. | -1 pt (Data binding explicit) |
| Dirty check mechanism (Dev) | "Close dialog (no confirmation if unchanged)" on line 255 does not define how "unchanged" is determined — dirty check against initial form state? Deep comparison? Developer must guess. | -1 pt (Interaction unambiguity) |
| Form submission note (Dev) | "Tag text input: Enter | Add tag (prevent form submission)" but the dialog is not a `<form>` element — the note about preventing form submission could mislead the developer into unnecessary `preventDefault` logic. | -1 pt (Interaction unambiguity) |

---

## Attack Points

### Attack 1: User — Accessibility contrast and motion gaps remain

**Where**: Design System section lists "Text: tertiary `#94a3b8`" and DecisionTimeline States specify skeleton rows with "animate-pulse". DecisionTimeline layout says "Edit: [编辑] ghost btn (draft only)".
**Why it's weak**: Despite the major accessibility improvements in iteration 2 (ARIA, keyboard nav, focus trap, live regions), two concrete gaps persist. First, tertiary text color `#94a3b8` on white `#ffffff` yields approximately 3.0:1 contrast ratio, which fails WCAG AA's 4.5:1 requirement for normal text. This color is used extensively: timeline item time text (12px), footer count text, tag badge text, creator text, form counter text. Second, the skeleton loading state uses `animate-pulse` with no `prefers-reduced-motion` media query alternative — users with vestibular disorders will see constant animation. Third, the edit button has no `aria-label` specified in the accessibility section, unlike the draft badge and category badge which both have `aria-label` entries. These are not theoretical issues — they affect real users daily.
**What must improve**: (1) Audit all text-on-surface contrast ratios against WCAG AA. If tertiary text must stay at `#94a3b8`, restrict it to decorative or large-text-only use (WCAG AA for large text is 3:1). For small body text, use a darker secondary variant. (2) Add `@media (prefers-reduced-motion: reduce)` guidance: replace `animate-pulse` skeletons with static gray blocks, replace `animate-spin` spinner with a static "loading" text. (3) Add `aria-label="编辑草稿"` to the edit button in the Accessibility section.

### Attack 2: Design — Missing success feedback and undefined badge variant

**Where**: DecisionFormDialog Interactions, rows: "Click [保存草稿]" and "Click [发布]" both specify "Button loading → on success: close dialog, refresh timeline" with no intermediate success message. DecisionTimeline Category badge colors table lists "default (slate)" as the variant for 4 of 6 categories.
**Why it's weak**: The design specifies detailed loading, error, and validation states but completely omits success feedback. When a user clicks "发布" (publish), the button shows a spinner, then the dialog simply closes and the timeline refreshes. There is no toast, no success badge, no brief confirmation message. In a PM tool where decisions are important records, the user needs reassurance that the publish action succeeded — especially since the timeline refresh is asynchronous and the new item may not appear instantly at the top (it could be at the bottom of the existing list or require a scroll). Additionally, the category badge variant "default (slate)" does not correspond to any defined badge variant in the referenced DESIGN.md. The design system defines success/warning/error badge variants mapped to semantic colors, but has no "slate" or "default" badge specification. A developer implementing this would need to invent the slate badge style, leading to inconsistency.
**What must improve**: (1) Add a success toast notification: "决策已发布" / "草稿已保存" that appears for 3 seconds after the API call succeeds, before or concurrent with dialog close. (2) Define the "default (slate)" badge variant explicitly: specify its bg color, text color, and border (e.g., `bg: #f8fafc, text: #475569, border: #e2e8f0`) or map it to the existing badge system's base style. (3) Consider whether the newly created/published item should be highlighted or scrolled to in the timeline after the dialog closes.

### Attack 3: Dev — Recent tags data source orphan and dirty check ambiguity

**Where**: DecisionFormDialog Layout Structure describes "Recent tags dropdown: absolute, w-full, bg-surface, shadow-lg, rounded-md, z-10 └── Tag items: px-3 py-2, hover bg-bg-alt" and Interactions specify "Type in tag input | Filter recent tags dropdown." DecisionFormDialog Data Binding table has no entry for recent tags. Interactions specify "Click [取消] / [×] / overlay | Close dialog (no confirmation if unchanged)."
**Why it's weak**: The recent tags dropdown is a significant UI element — it appears when the tag input is focused, shows up to 10 suggestions, and filters based on current input. Yet it has no data binding entry. The developer cannot know: Is there a `GET /api/v1/teams/:teamId/recent-tags` endpoint? Is it derived from existing decisions' tags? Is it stored in localStorage? Is it a prop passed from the parent? This is an orphan UI element with no data source, which the rubric penalizes at -3 pts per element. Separately, the dirty check for the "unchanged" determination in the close-without-confirmation behavior is unspecified. "Unchanged" could mean: (a) no fields were touched (pristine check), (b) current values equal initial values (deep comparison), or (c) only non-whitespace content differs. Each approach has different edge cases — e.g., if a user adds a tag then removes it, is the form "changed"? If they type in the content field then delete back to the original, is it "changed"? The developer needs explicit guidance.
**What must improve**: (1) Add a row to DecisionFormDialog Data Binding: "Recent tags dropdown items | `recentTags[]` | Source: [specify API endpoint or prop or localStorage]." (2) Clarify the dirty check: "A form is considered 'unchanged' if all field values equal their initial state (deep comparison). Adding and removing a tag returns to 'unchanged' if the final tag set matches the initial tag set." (3) Specify whether the expanded content area in the timeline should have a max-height with overflow-y scroll (e.g., `max-h-[200px] overflow-y-auto`) to prevent very long decisions from creating an excessively tall timeline item.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Zero accessibility specification | ✅ Fully addressed | Both components now have dedicated Accessibility sections with ARIA attributes (role="feed", role="article", role="dialog", aria-modal, aria-expanded, aria-controls, aria-live), keyboard navigation tables, focus trap specification, focus return, and screen reader announcements. This was the single largest improvement from iteration 1. |
| Attack 2: Missing pagination error and stale data edge cases | ✅ Fully addressed | "Pagination Error" state added (line 95): "Inline error below last loaded item: '加载更多失败' text + [重试] ghost button." "Draft not found on edit" interaction added (line 105): "Show toast error: '该草稿已被删除或不存在', remove the item from timeline list." Tag overflow addressed (line 84): "max-width 120px with ellipsis" and "+N overflow badge." |
| Attack 3: Ambiguous data loading for edit mode | ✅ Fully addressed | Interactions table now specifies (line 246): "Parent passes `bizKey` as prop; dialog calls `GET /api/v1/teams/:teamId/main-items/:itemId/decisions/:bizKey` to fetch fresh draft data." "Loading Draft" state added (line 235). Creator binding clarified (line 118): "server embeds `createdByName` in decision list response; no client-side lookup." |

---

## Verdict

- **Score**: 84/100
- **Target**: 80/100
- **Gap**: 0 points (target reached)
- **Action**: Target reached. The document has addressed all three critical issues from iteration 1 (accessibility, pagination error, data loading ambiguity). Remaining deductions are moderate: accessibility contrast ratios, missing success feedback, and an orphaned recent-tags data source. These are real issues but do not block implementation. The document is production-ready for a development handoff with the understanding that the developer will need to make small decisions about recent-tags sourcing and success feedback patterns.
