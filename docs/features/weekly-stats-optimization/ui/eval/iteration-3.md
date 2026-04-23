---
iteration: 3
score: 83
date: 2026-04-23
doc: docs/features/weekly-stats-optimization/ui/ui-design.md
---

# UI Design Evaluation — Iteration 3

**Total Score: 83 / 100**

---

## Dimension Scores

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage (PM) | 22 | 25 |
| User Experience (End User) | 20 | 25 |
| Design Integrity (Designer) | 19 | 25 |
| Implementability (Developer) | 22 | 25 |

---

## 1. Requirement Coverage — 22/25

### UI Function Coverage — 7/8

All three PRD UI functions are represented: the 7-card StatsBar, the Tooltip component, and the Responsive Layout. Data fields map 1:1 to `prd-ui-functions.md`. However, the PRD validation rule "卡片顺序固定：本周活跃 → 本周新完成 → 进行中 → 阻塞中 → 未开始 → 暂停中 → 逾期中" is still only implied by table row order in the design — never stated as an explicit ordering constraint. A developer reading the design alone has no guarantee the order is fixed; the table could be reordered without violating any stated rule. This was flagged in iteration 2 and remains unaddressed.

### State Requirement Coverage — 8/8

All four StatsBar states (Loading, Loaded, Empty/0, Error) are addressed with concrete visuals and ARIA annotations. Tooltip states (Hidden, Visible) match the PRD. Responsive breakpoints match the three PRD states exactly. Full marks.

### Edge Case Handling — 7/9

Covered: long number overflow (`9999+`), zero data, slow network (skeleton), label truncation, concurrent week switching (now handled via AbortController — fixed from iteration 2).

Missing:
- **Permission denied**: No state defined for when the user lacks access to the weekly stats endpoint. Neither the PRD nor the design addresses it, but a complete design must cover authorization failure as a distinct error path.
- **`9999+` tooltip data source**: The No-truncation Rule states "display as `9999+` with a tooltip showing the exact value" but does not specify the data source for that tooltip value. Is it the same `stats.*` field the component already holds? A separate API call? Two developers will implement this differently. Carried over from iteration 2 without fix.

---

## 2. User Experience — 20/25

### Information Hierarchy — 6/8

The stat number (28px bold) vs label (12px uppercase) contrast is correct. The left-border color accents provide meaningful urgency differentiation — `逾期中` uses red, `阻塞中` uses amber. However, all 7 cards remain the same physical size and weight. There is no mechanism to make critical cards (overdue, blocked) more visually prominent than neutral ones (activeSubItems, pending) beyond color alone. Color-blind users receive no additional urgency signal — no icon, no size difference, no positional priority.

### Interaction Intuitiveness — 7/8

All patterns are conventional: hover delay, click-to-toggle on mobile, Tab/Escape keyboard support, 300ms delay explicitly specified. One gap remains: the design does not specify behavior when the user hovers a new card while a tooltip from a previous card is still fading out. Is the outgoing animation cancelled immediately? Does the new tooltip wait for the fade-out to complete? Not addressed.

### Accessibility — 7/9

Strengths: error banner carries `role="alert"`, loading wrapper uses `aria-live="polite" aria-busy="true/false"`, tooltip uses `aria-describedby` / `role="tooltip"` / `aria-hidden`, focus ring is specified.

Remaining gap:
- **Contrast not verified**: `slate-400 (#94a3b8)` is listed as a Core Token for "Placeholder / disabled" use on white backgrounds. This color pair is approximately 2.85:1 — it fails WCAG AA (4.5:1 required for small text). The design lists it as a token without flagging the contrast risk. Any element rendered in `slate-400` (disabled state, placeholder context) will fail accessibility requirements. Carried over from iteration 2 without fix.

---

## 3. Design Integrity — 19/25

### Design System Adherence — 6/8

The three token violations from iteration 2 were resolved by adding `Primary 400 (#818cf8)` and `Slate 500 (#64748b)` to Core Tokens, and correcting `逾期中` to use `Red 600 (#dc2626)`. The token table is now internally consistent.

However, a new contradiction is present: the Typography spec states "Stat number: Inter 700, **28px**, slate-900" but the Layout Structure pseudocode uses `text-2xl font-bold`. Tailwind's `text-2xl` is **1.5rem = 24px** at the default 16px base — not 28px. A developer following the Tailwind class will render stat numbers 4px smaller than the typography spec requires. To achieve 28px, the class must be `text-[28px]` or a custom utility. The spec and the implementation class contradict each other.

### Visual Coherence — 7/9

Within each component the styling is internally consistent. The tooltip (slate-900 bg, white text, rounded-lg, shadow-lg) is coherent. The card accent colors are now all drawn from the defined token set. However, the 28px vs `text-2xl` (24px) discrepancy means the rendered stat numbers will not match the typography spec — the most prominent visual element on the bar will be the wrong size.

### State Completeness — 6/8

The four StatsBar states are defined. Tooltip has two states. Transitions are described (opacity fade, 150ms ease-out).

Gap: the Loading → Error transition path is still not described. If the skeleton is visible and the API returns an error, does the skeleton immediately swap to the error banner? Is there a brief intermediate state? The design defines the end states but not the transition between them. Carried over from iteration 2 without fix.

---

## 4. Implementability — 22/25

### Layout Specificity — 7/8

Breakpoints, gap, padding, border-radius, and tooltip positioning formula are all specified. The tooltip flip condition is now clarified as relative to the viewport (`element.getBoundingClientRect().top` relative to the viewport) — fixed from iteration 2.

Remaining gap: no explicit card height is given. `min-w-0` prevents overflow but does not enforce uniform card height. Cards containing a single-digit number vs a 4-digit number will render at different heights unless `items-stretch` or a fixed `h-*` is applied to the grid. Not addressed.

### Data Binding — 7/8

All 7 stat numbers are mapped to `stats.*` fields. Tooltip text is mapped to static config with exact strings. Error state display (`—`) is specified.

Remaining gap: the `9999+` overflow display is defined, but "a tooltip showing the exact value" does not specify the data source. The component already holds the `stats.*` field — if that is the intended source, the design should state it explicitly. Without this, two developers will implement it differently.

### Interaction Unambiguity — 8/9

The interactions table is now comprehensive. Concurrent week switching is explicitly handled (AbortController, discard stale response, show Loading immediately) — fixed from iteration 2. Tooltip flip reference frame is now clarified as viewport-relative — fixed from iteration 2. Animation durations (150ms ease-out) are specified. Escape key is covered.

One vague entry remains:
- **"Click outside any card (mobile)"** — no implementation hint for detecting "outside" (document-level click listener? pointer capture? `useClickOutside` hook?). Developers will implement this inconsistently.

---

## Top Attacks

1. **Design Integrity / Typography Contradiction**: Typography spec states "Stat number: Inter 700, 28px" but Layout Structure uses `text-2xl font-bold` — Tailwind's `text-2xl` is 24px, not 28px. A developer following the Tailwind class will render the most prominent visual element on the bar 4px smaller than specified. To achieve 28px, the class must be `text-[28px]`.

2. **Requirement Coverage / Card Ordering Not Explicit**: PRD validation rule states "卡片顺序固定：本周活跃 → 本周新完成 → 进行中 → 阻塞中 → 未开始 → 暂停中 → 逾期中" but the design never states this as an explicit ordering constraint — only implies it through table row order. A developer has no stated guarantee the order is fixed.

3. **Accessibility / Unverified Contrast**: `slate-400 (#94a3b8)` on white is ~2.85:1, failing WCAG AA (4.5:1 required for small text). The token is listed for "Placeholder / disabled" use without any contrast warning. Carried over from iteration 2 without fix.

4. **Design Integrity / Loading→Error Transition Undefined**: The design defines Loading and Error as separate end states but never describes the transition between them. When the skeleton is visible and the API returns an error, the visual path to the error banner is unspecified.

5. **Implementability / Card Height Unspecified**: No minimum or fixed card height is given. Cards with single-digit vs 4-digit numbers will render at different heights without `items-stretch` or a fixed `h-*` class — the design does not address this.

6. **Requirement Coverage / Permission Denied State Missing**: No state is defined for when the user lacks access to the weekly stats endpoint. Authorization failure is a distinct error path from a generic API error and requires its own handling.
