---
iteration: 1
score: 78
date: 2026-04-23
doc: docs/features/weekly-stats-optimization/ui/ui-design.md
---

# UI Design Evaluation — Iteration 1

**Total Score: 78 / 100**

---

## Dimension Scores

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 21 | 25 |
| User Experience (End User) | 19 | 25 |
| Design Integrity (Designer) | 18 | 25 |
| Implementability (Developer) | 20 | 25 |

---

## 1. Requirement Coverage — 21/25

### UI Function Coverage — 7/8

All three PRD UI functions are represented in the design: the 7-card StatsBar, the Tooltip component, and the Responsive Layout component. Data fields map 1:1 to `prd-ui-functions.md`. However, the PRD explicitly states "卡片顺序固定" as a validation rule, and the design never restates this constraint — it is only implied by table row order. A developer reading the design alone has no explicit ordering guarantee.

### State Requirement Coverage — 8/8

All four states from UI Function 1 (Loading, Loaded, Empty/0, Error) are addressed with concrete visuals. Tooltip states (Hidden, Visible) match the PRD. Responsive breakpoints match the three PRD states exactly (≥1280px, 768–1279px, <768px). Full marks.

### Edge Case Handling — 6/9

Covered: long number overflow (`9999+`), zero data, slow network (skeleton), label truncation.

Missing:
- **Permission denied**: No state defined for when the user lacks access to the weekly stats endpoint. The PRD does not mention it either, but a complete design must account for it.
- **Concurrent week switching**: If the user changes the selected week while a request is in-flight, the design gives no guidance — cancel the previous request? Show stale data? This is a real race condition in the existing weekly view.

---

## 2. User Experience — 19/25

### Information hierarchy — 6/8

The stat number (28px bold) vs label (12px uppercase) contrast is correct. The left-border color accents help differentiate cards. However, all 7 cards carry equal visual weight — there is no mechanism to surface the most decision-critical stats (overdue, blocked) more prominently than neutral ones (activeSubItems). A PM scanning the bar cannot immediately identify "something needs attention" without reading every card. The design shows the happy path of equal prominence; it does not address urgency signaling.

### Interaction Intuitiveness — 7/8

All patterns are conventional: hover delay, click-to-toggle on mobile, Tab/Escape keyboard support. The 300ms delay is explicitly specified. One minor gap: the design does not specify what happens if the user hovers a card while a tooltip from a previous card is still fading out — is there a race condition in the animation? Not addressed.

### Accessibility — 6/9

Positives: `aria-describedby`, `role="tooltip"`, `aria-hidden`, `focus-visible:ring-2` are all specified.

Gaps:
- **No `aria-live` region for the error banner**: "加载失败，请刷新" is a dynamic state change. Screen readers will not announce it without `role="alert"` or `aria-live="assertive"`.
- **No `aria-live` for loading→loaded transition**: A screen reader user has no way to know when the skeleton resolves to real data.
- **Contrast not verified**: `slate-400 (#94a3b8)` is used for placeholder/disabled text on white. This is 2.85:1 — fails WCAG AA (4.5:1 required for small text). The design lists it as a token without flagging the contrast risk.

---

## 3. Design Integrity — 18/25

### Design system adherence — 6/8

The Core Tokens table defines `Amber 600 (#d97706)` and `Emerald 600 (#059669)`, but the Card Color Accents table uses `amber-500 (#f59e0b)` and `emerald-500 (#10b981)` respectively. These are different values — the design contradicts its own token definitions. A designer or developer cannot know which is authoritative.

Additionally, the Layout Structure pseudocode describes `<StatsBar>` as a `flex container`, but the Responsive Grid table uses `grid grid-cols-{n}`. These are mutually exclusive layout modes. The design never resolves this contradiction.

### Visual Coherence — 6/9

Within each component the styling is consistent. The tooltip (slate-900 bg, white text, rounded-lg, shadow-lg) is coherent. The card structure is uniform. However, the token inconsistency noted above (amber-500 vs amber-600, emerald-500 vs emerald-600) means the rendered colors will not match the defined palette — a subtle but real coherence break that would be caught in design review.

### State Completeness — 6/8

The four StatsBar states are defined. Tooltip has two states. Transitions are described (opacity fade, 150ms ease-out).

Gap: the Loading → Error transition is not described. If the skeleton is visible and the API returns an error, does the skeleton immediately swap to the error banner? Is there an intermediate state? The design only defines the end states, not the transition path between them.

---

## 4. Implementability — 20/25

### Layout Specificity — 6/8

Breakpoints, gap, padding, and border-radius are all specified. The tooltip positioning formula (`bottom: calc(100% + 8px)`, `translateX(-50%)`) is precise.

The flex vs grid contradiction (Layout Structure pseudocode says `flex container`; Responsive Grid table says `grid grid-cols-7`) forces the developer to make a judgment call. This is a spec ambiguity that will produce inconsistent implementations across developers.

No explicit width or height is given for the card itself — `flex-1` with `min-w-0` is specified but no minimum card height, which means cards with a single-digit number vs a 4-digit number will have different heights unless `h-full` or `items-stretch` is applied. Not addressed.

### Data Binding — 7/8

All 7 stat numbers are mapped to `stats.*` fields. Tooltip text is mapped to static config with exact strings. Error state display (`—`) is specified.

Minor gap: the `9999+` overflow display is defined, but the design states "a tooltip showing the exact value" without specifying the data source for that tooltip. Is it the same `stats.*` field? A separate API call? Unspecified.

### Interaction Unambiguity — 7/9

The interactions table is the strongest part of the document. Every trigger has an explicit action and feedback. Animation durations (150ms ease-out) are specified. Escape key is covered.

Two vague entries:
- "Click outside any card (mobile)" — no implementation hint for how to detect "outside" (document-level click listener? pointer capture?). A developer will implement this differently without guidance.
- Tooltip flip condition: "card's top offset < tooltip height + 8px" — `card's top offset` is ambiguous. Top offset relative to what? The viewport? The scroll container? The page? This will produce bugs at different scroll positions.

---

## Top Attacks

1. **Design Integrity / Token Contradiction**: Core Tokens defines `Amber 600 (#d97706)` but Card Color Accents uses `amber-500 (#f59e0b)`; same mismatch for Emerald. The design contradicts its own token table — one of them is wrong and a developer cannot know which.

2. **Accessibility / Missing aria-live**: The error banner "加载失败，请刷新" and the loading→loaded transition have no `role="alert"` or `aria-live` annotation. Screen reader users will not be notified of these dynamic state changes.

3. **Implementability / flex vs grid contradiction**: Layout Structure pseudocode declares `flex container` while the Responsive Grid table uses `grid grid-cols-{n}`. These are incompatible — the spec must pick one and remove the other.

4. **Requirement Coverage / No concurrent-request handling**: The design has no guidance for the race condition when a user switches weeks while a request is in-flight. This is a known issue in the existing weekly view (see `docs/lessons/weekly-view-bug-fixes.md`) and the design repeats the omission.

5. **User Experience / No urgency signaling**: All 7 cards are visually equal weight. The design provides no mechanism for surfacing critical states (overdue, blocked) more prominently — a PM cannot scan the bar and immediately know "action required."

6. **Implementability / Tooltip flip reference frame**: "card's top offset < tooltip height + 8px" does not specify what the offset is measured against (viewport? scroll container? document?). This will produce incorrect flip behavior at non-zero scroll positions.
