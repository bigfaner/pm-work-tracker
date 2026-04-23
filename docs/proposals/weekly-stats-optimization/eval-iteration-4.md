---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 4
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 83/100
previous-score: 83/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 4)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 3 |
|-----------|-------|-----|-------------|
| Problem Definition | 14 | 20 | 0 |
| Solution Clarity | 16 | 20 | 0 |
| Alternatives Analysis | 12 | 15 | 0 |
| Scope Definition | 13 | 15 | 0 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 13 | 15 | 0 |
| **Total** | **83** | **100** | **0** |

---

## What Changed Since Iteration 3

All three attacks from iteration 3 were raised. None were addressed in this iteration. The proposal text is unchanged from iteration 3. Scores are held constant — no credit for prior improvement, no penalty for stagnation; the score reflects only what is on the page.

- **Attack 1 (urgency unquantified)**: Not addressed. "随团队规模增长，手动检查成本线性放大" remains an unquantified assertion. No data point added.
- **Attack 2 ("本周活跃" edge case)**: Not addressed. The OR-condition interaction for an item with a progress record this week but `actualEndDate < weekStart` is still unspecified.
- **Attack 3 (option 1 rejection unargued)**: Not addressed. The distinction between card 2 ("本周新完成") and option 1's all-time completed count is still unexplained.

---

## Dimension Breakdown

### 1. Problem Definition — 14/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete, named instance. "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (4/7)**

One code-review instance remains the sole evidence. No user feedback, no support tickets, no frequency data. A single observed event from a single review session is thin support for a product change.

**Urgency justified (4/6)**

"PM 每周须逐一展开子事项才能判断是否存在未启动或逾期事项" names a real workflow cost. "随团队规模增长，手动检查成本线性放大" adds a scalability argument. Both remain unquantified: how many PMs are affected, how long does manual inspection take per week, at what team size does the cost become unacceptable? Without a number, this is an assertion, not evidence.

---

### 2. Solution Clarity — 16/20

**Approach is concrete (6/7)**

The 7-card table with explicit filter rules is the strongest part of the proposal. Each card has a name, a rule, and a tooltip string. The edge case remains unresolved: an item that satisfies condition 1 (has a progress record this week) while failing condition 2 (`actualEndDate < weekStart`). The two OR sub-conditions are not mutually exclusive, and their interaction for this overlap is unspecified. A developer implementing `buildWeeklyGroups` must decide whether a progress record overrides an end date before the week — the proposal does not answer this.

**User-facing behavior described (5/7)**

Tooltip strings are specified and the string-exact match criterion is in success criteria. Still no description of visual layout: card order, card width, hover interaction (delay, position, style). The layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior.

**Distinguishes from alternatives (5/6)**

The "方案选择说明" paragraph provides explicit reasoning for why option 2 is blocking and why option 3's costs are bounded. Minor deduction: the proposal still does not explain why 7 is the right number — why not 8 (adding a `closed` card) or 6 (merging two cards)?

---

### 3. Alternatives Analysis — 12/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option. Full marks.

**Pros/cons for each (3/5)**

Option 1's con ("completed/closed 的历史计数对当周决策无意义") is still asserted without argument. The proposal already includes a "本周新完成" card — the distinction between that and a total-completed card from option 1 is never explained. Option 2's con is well-supported by the rationale paragraph. Option 3's con still understates the layout work.

**Rationale for chosen approach (4/5)**

The "方案选择说明" paragraph explains why option 2 is blocking and why option 3's costs are acceptable. Minor deduction: the rejection of option 1 is still not argued — the paragraph focuses entirely on option 2 vs. option 3.

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

Four meaningful risks with likelihood/impact ratings and actionable mitigations. Full marks. Unchanged from iteration 3.

---

### 6. Success Criteria — 13/15

**Criteria are measurable (5/5)**

All four criteria are objectively verifiable. Full marks.

**Coverage is complete (4/5)**

Remaining gaps:
- Accessibility of tooltips (keyboard navigation, screen reader) — not covered.
- Frontend tooltip rendering test type is unspecified — the string-match criterion exists but no test artifact (unit test, e2e spec) is listed in scope to verify it.

**Criteria are testable (4/5)**

The fixture test is specific and directly testable. The layout criteria cover two breakpoints. Minor deduction: the fixture still does not include a `closed` item — the boundary between `closed` and `completed` in the "本周活跃" definition (condition 2: "未在本周开始前完成/关闭") is untested.

---

## Top Attacks

### Attack 1 — Problem Definition: Urgency is asserted, not evidenced

> "随团队规模增长，手动检查成本线性放大"

The "推迟代价" section names the right workflow impact but provides no numbers. How many PMs are currently doing this manual check? How long does it take per week? At what team size does the cost become unacceptable? "线性放大" is a claim about growth behavior, not a measurement. A reader cannot use this to prioritize the work against other backlog items. Must improve: add one concrete data point — e.g., "a team with 20 sub-items requires ~5 minutes of manual inspection per weekly review" — or cite the frequency of the code-review confusion event.

### Attack 2 — Solution Clarity: "本周活跃" edge case is unresolved

> "本周活跃 = 子事项满足以下任一条件：1. 本周内有进展记录 ... 2. 日期范围与本周重叠：创建于本周结束前，且未在本周开始前完成/关闭"

An item can satisfy condition 1 (has a progress record this week) while failing condition 2 (`actualEndDate < weekStart`, meaning it ended before the week started). The two OR sub-conditions are not mutually exclusive, but their interaction for this case is unspecified. A developer implementing `buildWeeklyGroups` would need to decide: does a progress record this week override an end date before the week? The proposal does not answer this. Must improve: add a note clarifying that condition 1 takes precedence, or add a fixture item that exercises this overlap.

### Attack 3 — Alternatives Analysis: Option 1 rejection is unargued

> "completed/closed 的历史计数对当周决策无意义，卡片过多稀释注意力" — Verdict: 否决

The proposal already includes a "本周新完成" card (card 2) that counts completed items. The distinction between card 2 and a hypothetical "all-time completed" card from option 1 is never explained. Why is a total-completed count meaningless for weekly review while a this-week-completed count is meaningful? The "卡片过多稀释注意力" argument applies equally to the chosen 7-card solution. Must improve: explain what specific information option 1 would add that is not decision-relevant, and why 7 cards does not itself dilute attention.
