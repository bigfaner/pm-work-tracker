---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 3
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 83/100
previous-score: 75/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 3)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 2 |
|-----------|-------|-----|-------------|
| Problem Definition | 14 | 20 | +3 |
| Solution Clarity | 16 | 20 | +2 |
| Alternatives Analysis | 12 | 15 | +2 |
| Scope Definition | 13 | 15 | 0 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 13 | 15 | +1 |
| **Total** | **83** | **100** | **+8** |

---

## What Changed Since Iteration 2

Three attacks were raised in iteration 2. Here is what was addressed:

- **Attack 1 (urgency absent)**: Addressed. A "推迟代价" section was added naming the concrete workflow impact: PM must expand sub-items one by one each week, and the cost scales linearly with team size. This is a real improvement. Deduction remains because the impact is still unquantified — no data on how many PMs are affected, how much time per week, or at what team size the cost becomes critical.
- **Attack 2 (rationale unargued)**: Fully addressed. A "方案选择说明" paragraph now explains why option 2's gap is blocking (not nice-to-have) and why option 3's costs are bounded and acceptable. The reasoning is explicit and reconstructable.
- **Attack 3 (invariants and multi-width untested)**: Fully addressed. The fixture criterion now explicitly verifies the sum invariant (progressing(1) + blocking(1) + pending(1) + pausing(1) = 4 ≤ active(5)), and the layout criterion now covers both 1280px and 768px breakpoints.

---

## Dimension Breakdown

### 1. Problem Definition — 14/20

**Problem stated clearly (6/7)**

Two problems remain clearly named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete, named example. "容易对数字产生疑惑" is still an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (4/7)**

The code-review instance remains the sole evidence: one observed event from one review session. No user feedback, no support tickets, no frequency data. Score unchanged from iteration 2.

**Urgency justified (4/6)**

The "推迟代价" section is a genuine improvement. "PM 每周须逐一展开子事项才能判断是否存在未启动或逾期事项" names a concrete workflow cost. "随团队规模增长，手动检查成本线性放大" adds a scalability argument. However, both claims remain unquantified: how many PMs are currently affected? How much time per week does manual inspection take? What team size triggers the problem? Without a number, the urgency argument is still an assertion, not evidence.

---

### 2. Solution Clarity — 16/20

**Approach is concrete (6/7)**

The 7-card table with explicit filter rules is the strongest part of the proposal. Each card has a name, a rule, and a tooltip string. One edge case remains unresolved: an item that has a progress record this week but whose `actualEndDate` falls before `weekStart` — the "本周活跃" definition's two OR sub-conditions would include it (condition 1 matches), but condition 2 would exclude it (ended before weekStart). The interaction between the two conditions is not specified for this overlap case.

**User-facing behavior described (5/7)**

Tooltip strings are specified and the string-exact match criterion is in success criteria. Still no description of the visual layout: card order, card width, hover interaction (delay, position, style). The layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior.

**Distinguishes from alternatives (5/6)**

The "方案选择说明" paragraph now provides explicit reasoning: option 2's gap is blocking because the PM's core use case (detecting unstated or overdue items without expanding sub-items) cannot be satisfied by tooltip alone. Option 3's costs are argued to be bounded and acceptable. Minor deduction: the proposal still does not explain why 7 is the right number — why not 8 (adding a `closed` card) or 6 (merging two cards)?

---

### 3. Alternatives Analysis — 12/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option. Full marks.

**Pros/cons for each (3/5)**

Option 1's con ("completed/closed 的历史计数对当周决策无意义") is still asserted without argument. Why is a completed-all-time count not meaningful for weekly review? The proposal already includes a "本周新完成" card — the distinction between that and a total-completed card is not explained. Option 2's con is now better supported by the rationale paragraph. Option 3's con still understates the layout work.

**Rationale for chosen approach (4/5)**

The "方案选择说明" paragraph is a clear improvement. It explains why option 2 is blocking, not nice-to-have, and why option 3's costs are acceptable. Minor deduction: the rejection of option 1 is still not argued — the paragraph focuses entirely on option 2 vs. option 3.

---

### 4. Scope Definition — 13/15

**In-scope items are concrete (5/5)**

Each item names a specific artifact. Full marks.

**Out-of-scope explicitly listed (5/5)**

Four deferred items are named explicitly. Full marks.

**Scope is bounded (3/5)**

No timeframe is defined anywhere in the proposal. E2E tests for tooltip rendering are absent from scope — the success criteria require string-exact tooltip match but no test type (unit/e2e) is specified, and no frontend test is listed as an in-scope deliverable. Given the project's TDD conventions, this is a gap.

---

### 5. Risk Assessment — 15/15

Four meaningful risks with likelihood/impact ratings and actionable mitigations. Unchanged from iteration 2. Full marks.

---

### 6. Success Criteria — 13/15

**Criteria are measurable (5/5)**

All four criteria are objectively verifiable. Full marks.

**Coverage is complete (4/5)**

The mathematical invariant is now covered by the fixture criterion. The 768px breakpoint is now covered. Remaining gaps:
- Accessibility of tooltips (keyboard navigation, screen reader) — not covered.
- Frontend tooltip rendering test type is unspecified — the string-match criterion exists but no test artifact (unit test, e2e spec) is listed in scope to verify it.

**Criteria are testable (4/5)**

The fixture test is specific and directly testable. The layout criteria now cover two breakpoints. Minor deduction: the fixture still does not include a `closed` item — the boundary between `closed` and `completed` in the "本周活跃" definition (condition 2: "未在本周开始前完成/关闭") is untested.

---

## Top Attacks

### Attack 1 — Problem Definition: Urgency is asserted, not evidenced

> "随团队规模增长，手动检查成本线性放大"

The "推迟代价" section names the right workflow impact but provides no numbers. How many PMs are currently doing this manual check? How long does it take per week? At what team size does the cost become unacceptable? "线性放大" is a claim about growth behavior, not a measurement. A reader cannot use this to prioritize the work against other items in the backlog. Must improve: add one concrete data point — e.g., "a team with 20 sub-items requires ~5 minutes of manual inspection per weekly review" — or cite the frequency of the code-review confusion event.

### Attack 2 — Solution Clarity: "本周活跃" edge case is unresolved

> "本周活跃 = 子事项满足以下任一条件：1. 本周内有进展记录 ... 2. 日期范围与本周重叠：创建于本周结束前，且未在本周开始前完成/关闭"

An item can satisfy condition 1 (has a progress record this week) while failing condition 2 (actualEndDate < weekStart, meaning it ended before the week started). The two OR sub-conditions are not mutually exclusive, but their interaction for this case is unspecified. A developer implementing `buildWeeklyGroups` would need to decide: does a progress record this week override an end date before the week? The proposal does not answer this. Must improve: add a note clarifying that condition 1 takes precedence, or add a fixture item that exercises this overlap.

### Attack 3 — Alternatives Analysis: Option 1 rejection is unargued

> "completed/closed 的历史计数对当周决策无意义，卡片过多稀释注意力" — Verdict: 否决

The proposal already includes a "本周新完成" card (card 2) that counts completed items. The distinction between card 2 and a hypothetical "all-time completed" card from option 1 is never explained. Why is a total-completed count meaningless for weekly review while a this-week-completed count is meaningful? The "卡片过多稀释注意力" argument applies equally to the chosen 7-card solution. Must improve: explain what specific information option 1 would add that is not decision-relevant, and why 7 cards does not itself dilute attention.
