---
iteration: 2
score: 79
date: 2026-04-23
doc: docs/features/weekly-stats-optimization/ui/ui-design.md
---

# UI Design Evaluation — Iteration 2

**Total Score: 79 / 100**

---

## Dimension Scores

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 21 | 25 |
| User Experience (End User) | 20 | 25 |
| Design Integrity (Designer) | 17 | 25 |
| Implementability (Developer) | 21 | 25 |

---

## 1. Requirement Coverage — 21/25

### UI Function Coverage — 7/8

All three PRD UI functions are represented: the 7-card StatsBar, the Tooltip component, and the Responsive Layout component. Data fields map 1:1 to `prd-ui-functions.md`. The PRD validation rule "卡片顺序固定：本周活跃 → 本周新完成 → 进行中 → 阻塞中 → 未开始 → 暂停中 → 逾期中" is implied by table row order in the design but never stated as an explicit constraint. A developer reading the design alone has no ordering guarantee — the table could be reordered without violating any stated rule.

### State Requirement Coverage — 8/8

All four StatsBar states (Loading, Loaded, Empty/0, Error) are addressed with concrete visuals. The Loading state now correctly wraps the bar in `aria-live="polite" aria-busy="true"`. Tooltip states (Hidden, Visible) match the PRD. Responsive breakpoints match the three PRD states exactly. Full marks.

### Edge Case Handling — 6/9

Covered: long number overflow (`9999+`), zero data, slow network (skeleton), label truncation.

Missing:
- **Permission denied**: No state defined for when the user lacks access to the weekly stats endpoint. Neither the PRD nor the design addresses it, but a complete design must.
- **Concurrent week switching**: If the user changes the selected week while a request is in-flight, the design gives no guidance — cancel the previous request? Show stale data? This is a documented race condition in the existing weekly view (`docs/lessons/weekly-view-bug-fixes.md`) and the design repeats the omission.
- **`9999+` tooltip data source**: The No-truncation Rule states "display as `9999+` with a tooltip showing the exact value" but does not specify the data source for that tooltip. Is it the same `stats.*` field? A separate API call? Unspecified.

---

## 2. User Experience — 20/25

### Information Hierarchy — 6/8

The stat number (28px bold) vs label (12px uppercase) contrast is correct. The left-border color accents now provide meaningful urgency differentiation: `逾期中` uses red-500, `阻塞中` uses amber-600 — a PM can visually distinguish risk cards from neutral ones. However, all 7 cards remain the same physical size and weight. There is no mechanism to make critical cards (overdue, blocked) more prominent than neutral ones (activeSubItems, pending) beyond color alone. Color-blind users receive no additional urgency signal.

### Interaction Intuitiveness — 7/8

All patterns are conventional: hover delay, click-to-toggle on mobile, Tab/Escape keyboard support. The 300ms delay is explicitly specified. One gap remains: the design does not specify behavior when the user hovers a new card while a tooltip from a previous card is still fading out — is the outgoing animation cancelled? Does the new tooltip wait for the fade-out to complete? Not addressed.

### Accessibility — 7/9

Improvements from iteration 1:
- Error banner now carries `role="alert"` (implies `aria-live="assertive"`) — screen readers will announce it. Fixed.
- Loading wrapper now uses `aria-live="polite" aria-busy="true/false"` — screen readers will announce the loaded transition. Fixed.

Remaining gap:
- **Contrast not verified**: `slate-400 (#94a3b8)` is used for placeholder/disabled text on white. This is approximately 2.85:1 — fails WCAG AA (4.5:1 required for small text). The design lists it as a token without flagging the contrast risk. The label text uses `slate-600 (#475569)` which passes at ~5.9:1, but any element rendered in `slate-400` (e.g., skeleton placeholder context, disabled state) will fail.

---

## 3. Design Integrity — 17/25

### Design System Adherence — 5/8

The amber/emerald token contradiction from iteration 1 was fixed. However, iteration 2 introduces three new token violations in the Card Color Accents table:

1. **`逾期中` uses `red-500 (#ef4444)`** — Core Tokens defines `Red 600 (#dc2626)`. These are different values. The design contradicts its own token table.
2. **`本周活跃` uses `primary-400 (#818cf8)`** — Core Tokens defines only `Primary 600 (#4f46e5)` and `Primary 50 (#eef2ff)`. `primary-400` is not a defined token. A developer has no authoritative hex value to use.
3. **`暂停中` uses `slate-500 (#64748b)`** — Core Tokens defines `slate-900`, `slate-600`, `slate-400`, `slate-200`. `slate-500` is not a defined token.

Three out of seven card accents reference values outside the defined token set. The design system table is not self-consistent.

### Visual Coherence — 6/9

Within each component the styling is internally consistent. The tooltip (slate-900 bg, white text, rounded-lg, shadow-lg) is coherent. However, the three undefined tokens noted above (primary-400, slate-500, red-500 vs Red 600) mean the rendered colors will not match the defined palette. A designer reviewing the implementation against the spec will find three cards rendering off-palette colors with no documented justification.

### State Completeness — 6/8

The four StatsBar states are defined. Tooltip has two states. Transitions are described (opacity fade, 150ms ease-out).

Gap: the Loading → Error transition path is still not described. If the skeleton is visible and the API returns an error, does the skeleton immediately swap to the error banner? Is there a brief intermediate state? The design defines the end states but not the transition between them.

---

## 4. Implementability — 21/25

### Layout Specificity — 7/8

The flex vs grid contradiction from iteration 1 is fixed — the Layout Structure pseudocode now correctly labels `<StatsBar>` as a `grid container`. Breakpoints, gap, padding, and border-radius are all specified. The tooltip positioning formula is precise.

Remaining gap: no explicit card height is given. `min-w-0` prevents overflow but does not enforce uniform card height. Cards containing a single-digit number vs a 4-digit number will render at different heights unless `items-stretch` or a fixed `h-*` is applied to the grid. Not addressed.

### Data Binding — 7/8

All 7 stat numbers are mapped to `stats.*` fields. Tooltip text is mapped to static config with exact strings. Error state display (`—`) is specified.

Remaining gap: the `9999+` overflow display is defined, but "a tooltip showing the exact value" does not specify the data source. Is it the same `stats.*` field (which the component already has)? A separate API call? Without this, two developers will implement it differently.

### Interaction Unambiguity — 7/9

The interactions table is comprehensive. Every trigger has an explicit action and feedback. Animation durations (150ms ease-out) are specified. Escape key is covered.

Two vague entries remain from iteration 1:
- **"Click outside any card (mobile)"** — no implementation hint for detecting "outside" (document-level click listener? pointer capture? `useClickOutside` hook?). Developers will implement this inconsistently.
- **Tooltip flip condition**: "card's top offset < tooltip height + 8px" — `card's top offset` is still ambiguous. Top offset relative to what? The viewport? The scroll container? The document? This will produce incorrect flip behavior at non-zero scroll positions.

---

## Top Attacks

1. **Design Integrity / New Token Violations**: Iteration 2 fixed the amber/emerald contradiction but introduced three new ones — `逾期中` uses `red-500 (#ef4444)` while Core Tokens defines `Red 600 (#dc2626)`; `本周活跃` uses `primary-400 (#818cf8)` which is not in Core Tokens at all; `暂停中` uses `slate-500 (#64748b)` which is not in Core Tokens. Three of seven card accents are off-palette.

2. **Implementability / Tooltip Flip Reference Frame**: "card's top offset < tooltip height + 8px" does not specify what the offset is measured against (viewport? scroll container? document?). This will produce incorrect flip behavior at non-zero scroll positions. Carried over from iteration 1 without fix.

3. **Requirement Coverage / No Concurrent-Request Handling**: The design has no guidance for the race condition when a user switches weeks while a request is in-flight. This is a documented issue in `docs/lessons/weekly-view-bug-fixes.md` and the design repeats the omission for the second iteration.

4. **Accessibility / Unverified Contrast**: `slate-400 (#94a3b8)` on white is ~2.85:1, failing WCAG AA (4.5:1). The token is used for placeholder/disabled states without any contrast warning in the design.

5. **Implementability / Undefined Card Height**: No minimum or fixed card height is specified. Cards with single-digit vs 4-digit numbers will render at different heights without `items-stretch` or a fixed `h-*` class — the design does not address this.

6. **Requirement Coverage / `9999+` Tooltip Data Source**: The No-truncation Rule specifies "a tooltip showing the exact value" for numbers exceeding 4 digits but does not identify the data source for that tooltip value, leaving the implementation undefined.
