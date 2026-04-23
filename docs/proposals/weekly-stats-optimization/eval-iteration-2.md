---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 2
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 75/100
previous-score: 60/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 2)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 1 |
|-----------|-------|-----|-------------|
| Problem Definition | 11 | 20 | +3 |
| Solution Clarity | 14 | 20 | 0 |
| Alternatives Analysis | 10 | 15 | 0 |
| Scope Definition | 13 | 15 | 0 |
| Risk Assessment | 15 | 15 | +7 |
| Success Criteria | 12 | 15 | +5 |
| **Total** | **75** | **100** | **+15** |

---

## What Changed Since Iteration 1

Three attacks were raised in iteration 1. Here is what was addressed:

- **Attack 1 (zero evidence)**: Partially addressed. A concrete code-review instance was added: team members could not explain the "active" count without reading source code. This is a real observed event, not a hypothetical. Urgency remains absent.
- **Attack 2 (only 2 risks)**: Fully addressed. Two new risks added — DTO backward compatibility and test coverage gap — each with likelihood/impact ratings and actionable mitigations. Risk Assessment now scores full marks.
- **Attack 3 ("正确计算" untestable; tooltip content not verified)**: Fully addressed. A specific 5-item fixture with expected outputs replaces the vague criterion. Tooltip criterion now requires string-exact match against the solution table.

---

## Dimension Breakdown

### 1. Problem Definition — 11/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable: opaque counting rules and incomplete status coverage. The `justCompleted`/`activeSubItems` overlap is a concrete, named example. Two readers would reach the same interpretation. Minor deduction: "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation.

**Evidence provided (4/7)**

The code-review instance is a genuine improvement: "团队成员在不阅读源码的情况下无法解释'活跃'数字的构成" is an observed event, not a hypothetical. However, it is a single anecdote from one review session. No user feedback, no support tickets, no frequency data. One concrete instance raises the score from 2 to 4 but does not reach the threshold for "data or concrete examples" at full marks.

**Urgency justified (1/6)**

No urgency argument exists in iteration 2, same as iteration 1. Why solve this now rather than next quarter? What is the cost of deferral? The proposal is silent. One point awarded only because the undocumented overlap implies latent regression risk.

---

### 2. Solution Clarity — 14/20

**Approach is concrete (6/7)**

The 7-card table with explicit filter rules remains the strongest part of the proposal. Each card has a name, a rule, and a tooltip string. A developer can implement directly from this table. The "本周活跃" base condition is defined with two OR sub-conditions, but the edge case where an item has a progress record AND ended before weekStart is still unresolved — does it count as active?

**User-facing behavior described (5/7)**

Tooltip strings are specified and the string-exact match criterion is now in success criteria. However, there is still no description of the visual layout: how do 7 cards fit in the existing bar? What is the hover interaction (delay, position, style)? The layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior.

**Distinguishes from alternatives (3/6)**

The chosen approach verdict now reads "覆盖所有决策相关状态，tooltip 使规则自解释，解决已知混淆实例" — marginally more specific than iteration 1. Still does not argue why 7 is the right number, why `completed` and `closed` are excluded from individual status cards, or why tooltip-only (option 2) is insufficient beyond restating the problem.

---

### 3. Alternatives Analysis — 10/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option that functions as the "do nothing meaningful" baseline. Full marks.

**Pros/cons for each (3/5)**

Option 1's con ("completed/closed 的历史计数对当周决策无意义") is still asserted without argument — why is a completed count not meaningful for weekly review? Option 2's con is a restatement of the problem, not a trade-off analysis. Option 3's con accurately names the UI and DTO changes but understates the complexity of the layout work.

**Rationale for chosen approach (2/5)**

No explicit reasoning connects the problem statement to the selection of option 3. The verdict column says "采纳" but does not explain why the pros of option 3 outweigh its cons relative to the alternatives. A reader cannot reconstruct the decision logic from what is written.

---

### 4. Scope Definition — 13/15

**In-scope items are concrete (5/5)**

Each item names a specific artifact: `WeeklyStats` DTO fields, `buildWeeklyGroups` logic and unit tests, `StatsBar` component, tooltip addition, documentation. Tests are now explicitly in scope, addressing the gap from iteration 1. Full marks.

**Out-of-scope explicitly listed (5/5)**

Four deferred items are named explicitly. Full marks.

**Scope is bounded (3/5)**

No timeframe is defined anywhere in the proposal. E2E tests for tooltip behavior are absent from scope — the unit tests cover counting logic but no test covers the frontend tooltip rendering or the string-match criterion from success criteria. Given the project's TDD conventions, this is a gap.

---

### 5. Risk Assessment — 15/15

**Risks identified (5/5)**

Four meaningful risks are now listed: layout overflow, overdue definition inconsistency, DTO backward compatibility, and test coverage gap. All four are non-trivial and directly relevant to the proposed change. Full marks.

**Likelihood + impact rated (5/5)**

All four risks have likelihood and impact ratings. The ratings are plausible and not uniformly inflated. Full marks.

**Mitigations are actionable (5/5)**

- Layout: `flex-wrap` or `grid` — actionable.
- Overdue: reuse `lib/status.ts` `isOverdue` — actionable and names the specific file.
- DTO: grep all `WeeklyStats` references before release — actionable.
- Test coverage: TDD, write tests first per project convention — actionable.

Full marks.

---

### 6. Success Criteria — 12/15

**Criteria are measurable (5/5)**

- "统计栏显示 7 个卡片，布局在 1280px 宽度下不溢出" — measurable.
- Fixture test with 5 named sub-items and exact expected outputs — measurable.
- "tooltip 文本与方案表格'Tooltip 说明'列的字符串完全一致（字符串匹配，非仅检查存在性）" — measurable.
- "现有 4 个统计的数字与优化前保持一致（无回归）" — measurable.

Full marks.

**Coverage is complete (3/5)**

Criteria do not cover:
- The mathematical relationships stated in the solution: cards 3–6 sum ≤ card 1; card 7 may overlap cards 3–6. These are stated as invariants but no criterion verifies them.
- Behavior at screen widths other than 1280px — the risk section mentions small screens but no criterion covers it.
- Accessibility of tooltips (keyboard navigation, screen reader).
- Frontend tooltip rendering — the string-match criterion is stated but no test type (unit/e2e) is specified for it.

**Criteria are testable (4/5)**

The fixture test case is specific and directly testable. The tooltip string-match criterion is testable. The layout criterion is testable. The regression criterion is testable. Minor deduction: the fixture covers 5 items but does not include an item that is `closed` — the boundary between `closed` and `completed` in the "本周活跃" definition is untested.

---

## Top Attacks

### Attack 1 — Problem Definition: Urgency is completely absent

> (no urgency section exists in the proposal)

The proposal provides zero argument for why this work should happen now. What is the cost of deferring one quarter? What decision is being blocked by the missing visibility? The code-review anecdote establishes that confusion exists, but not that it is causing harm at a rate that justifies prioritization over other work. Must improve: add one sentence explaining the consequence of not solving this — e.g., "PM cannot identify stalled or overdue items from the stats bar alone, requiring manual sub-item inspection each week."

### Attack 2 — Alternatives Analysis: Rationale for chosen approach is unargued

> "覆盖所有决策相关状态，tooltip 使规则自解释，解决已知混淆实例" — Verdict: 采纳

This is a label, not a rationale. The proposal does not explain why option 3's cons (7-column layout, DTO change) are acceptable given the alternatives. Option 2 (keep 4 cards, add tooltip) addresses the transparency problem with zero backend change — the proposal dismisses it by restating the problem ("仍缺少 pending/pausing/overdue 可见性") without arguing why that visibility is necessary for the decision-making use case. Must improve: write 2–3 sentences explaining why the status visibility gap in option 2 is a blocking deficiency, not a nice-to-have.

### Attack 3 — Success Criteria: Mathematical invariants and multi-width layout are untested

> "卡片 3-6 之和 ≤ 卡片 1（活跃总数）" and "7 列布局在小屏幕上溢出" (from Risk section)

The solution section states two invariants (cards 3–6 sum ≤ card 1; card 7 may overlap cards 3–6) but no success criterion verifies them. The risk section identifies small-screen overflow as a risk but the only layout criterion is "1280px 宽度下不溢出" — a single breakpoint. Must improve: add a criterion that verifies the sum invariant via the fixture (e.g., "given the 5-item fixture, progressing(1) + blocking(1) + pending(1) + pausing(1) = 4 ≤ active(5)"), and add a layout criterion for at least one smaller breakpoint (e.g., 768px).
