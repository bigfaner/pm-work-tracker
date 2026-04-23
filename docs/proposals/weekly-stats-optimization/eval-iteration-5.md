---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 5
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 87/100
previous-score: 83/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 5)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 4 |
|-----------|-------|-----|-------------|
| Problem Definition | 15 | 20 | +1 |
| Solution Clarity | 18 | 20 | +2 |
| Alternatives Analysis | 14 | 15 | +2 |
| Scope Definition | 13 | 15 | 0 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 12 | 15 | -1 |
| **Total** | **87** | **100** | **+4** |

---

## What Changed Since Iteration 4

All three attacks from iteration 4 were raised. All three were addressed in this iteration.

- **Attack 1 (urgency unquantified)**: Addressed. Added concrete data point: "以当前团队规模（约 20 个活跃子事项）为例，每次周会前手动逐项检查约需 3–5 分钟；子事项数量每翻倍，该成本等比增长。" Partial credit — the estimate is labeled "约" and is not a measured value, but it is a specific, falsifiable claim. Score moves from 4/6 to 5/6.

- **Attack 2 ("本周活跃" edge case)**: Addressed. Added explicit "条件优先级" note with a typical scenario: "若子事项本周内有进展记录（满足条件 1），则无论其 `actualEndDate` 是否早于 `weekStart`，该事项均计为本周活跃。" The developer ambiguity is resolved. Score moves from 6/7 to 7/7.

- **Attack 3 (option 1 rejection unargued)**: Addressed. The "方案选择说明" paragraph now explicitly distinguishes card 2 ("本周新完成") from option 1's all-time completed count, and rebuts the "7 cards dilute attention" argument. Score moves from 5/6 to 6/6 and rationale from 4/5 to 5/5.

One regression introduced: the new "条件优先级" rule is not exercised by the fixture in success criteria, creating a new testability gap. Success Criteria drops from 13/15 to 12/15.

---

## Dimension Breakdown

### 1. Problem Definition — 15/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete, named instance with a quoted team reaction. "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (4/7)**

One code-review instance remains the sole evidence. "在代码审查中，团队成员在不阅读源码的情况下无法解释'活跃'数字的构成" is a single observed event from a single review session. No user feedback, no support tickets, no frequency data. A single event is thin support for a product change.

**Urgency justified (5/6)**

The new data point — "约 20 个活跃子事项，每次周会前手动逐项检查约需 3–5 分钟" — is a specific, falsifiable claim that a reader can evaluate. This is a meaningful improvement. Deduction retained: "约" signals an estimate, not a measurement; the growth claim ("子事项数量每翻倍，该成本等比增长") is still an assertion about behavior, not an observed trend.

---

### 2. Solution Clarity — 18/20

**Approach is concrete (7/7)**

The "条件优先级" note fully resolves the edge case from iteration 4. The rule is unambiguous: condition 1 (progress record this week) overrides condition 2 regardless of `actualEndDate`. A developer can implement `buildWeeklyGroups` without making an undocumented judgment call. Full marks.

**User-facing behavior described (5/7)**

Tooltip strings are specified and the string-exact match criterion is in success criteria. Still no description of visual layout: card order, card width, hover interaction (delay, position, style). The layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior. Deduction unchanged.

**Distinguishes from alternatives (6/6)**

The "方案选择说明" paragraph now explicitly explains why option 1's all-time completed count is inferior to card 2 ("本周新完成"): "方案 1 的 completed/closed 卡片显示的是全量历史累计数，该数字在每次周会时只会单调增长，无法回答'本周完成了多少'这一核心问题。" The "7 cards dilute attention" counter-argument is also rebutted with an information-density argument. Full marks.

---

### 3. Alternatives Analysis — 14/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option. Full marks.

**Pros/cons for each (4/5)**

Option 1's con is now substantively argued — the distinction between all-time count and this-week count is explicit. Option 2's con is well-supported. Minor deduction: option 3's con ("布局从 4 列变 7 列，需调整 UI；后端 DTO 新增 3 个字段") still understates the layout work — no mention of responsive behavior testing, tooltip accessibility, or the frontend test gap.

**Rationale for chosen approach (5/5)**

The "方案选择说明" paragraph addresses all three options with explicit reasoning. Full marks.

---

### 4. Scope Definition — 13/15

**In-scope items are concrete (5/5)**

Each item names a specific artifact. Full marks.

**Out-of-scope explicitly listed (5/5)**

Four deferred items are named explicitly. Full marks.

**Scope is bounded (3/5)**

No timeframe is defined anywhere in the proposal. Frontend tooltip test is absent from in-scope deliverables — the success criteria require string-exact tooltip match but no test artifact (unit test, e2e spec) is listed as a deliverable. Given the project's TDD conventions (CLAUDE.md), this is a structural gap: the test should be named in scope before the criterion can be considered executable.

---

### 5. Risk Assessment — 15/15

Four meaningful risks with likelihood/impact ratings and actionable mitigations. Full marks. Unchanged from iteration 4.

---

### 6. Success Criteria — 12/15

**Criteria are measurable (5/5)**

All four criteria are objectively verifiable. Full marks.

**Coverage is complete (4/5)**

Remaining gaps:
- Accessibility of tooltips (keyboard navigation, screen reader) — not covered.
- Frontend tooltip rendering test type is unspecified — the string-match criterion exists but no test artifact is listed in scope to verify it.

**Criteria are testable (3/5)**

The fixture test is specific and directly testable for the 5 named items. Two testability gaps:

1. The new "条件优先级" rule — an item with `actualEndDate < weekStart` but a progress record this week should count as active — is not exercised by any fixture item. The fixture has item D (completed + actualEndDate in this week) but no item with actualEndDate *before* weekStart and a progress record this week. The rule is specified but untested.

2. The `closed` boundary gap from iteration 4 remains: no fixture item tests the "未在本周开始前完成/关闭" condition for `closed` status specifically.

---

## Top Attacks

### Attack 1 — Success Criteria: Priority rule is specified but not fixture-tested

> "条件优先级：条件 1 优先于条件 2。若子事项本周内有进展记录（满足条件 1），则无论其 `actualEndDate` 是否早于 `weekStart`，该事项均计为本周活跃。"

The priority rule is the key new addition in this iteration. It resolves the developer ambiguity — but the fixture does not include an item that exercises it. Item D (completed + actualEndDate ∈ [weekStart, weekEnd]) does not test the override case. A developer could implement the rule incorrectly and all five fixture assertions would still pass. Must improve: add a fixture item F — e.g., `completed + actualEndDate < weekStart + 本周有进展记录` — and assert it counts toward active=6, with the expected output updated accordingly.

### Attack 2 — Problem Definition: Evidence is a single event, urgency estimate is unverified

> "在代码审查中，团队成员在不阅读源码的情况下无法解释'活跃'数字的构成"
> "以当前团队规模（约 20 个活跃子事项）为例，每次周会前手动逐项检查约需 3–5 分钟"

The urgency section improved but the evidence base did not. One code-review event remains the sole evidence for the rule-opacity problem. The 3–5 minute estimate is labeled "约" — it is an estimate, not a measurement. A reader cannot distinguish "we timed this once" from "we made this up." Must improve: note how frequently the code-review confusion has occurred (e.g., "observed in 3 of the last 5 sprint reviews"), or replace the estimate with a measured value.

### Attack 3 — Scope Definition: No timeframe, frontend test deliverable missing

The proposal has no timeframe anywhere — not even a rough estimate ("1 sprint", "2 days"). The success criteria require string-exact tooltip match, but no frontend test (unit or e2e) is listed as an in-scope deliverable. The project's TDD conventions (CLAUDE.md) require tests to be written before implementation. Without naming the test artifact in scope, the criterion is unexecutable as written. Must improve: add an estimated delivery timeframe and add a frontend tooltip test (e.g., "前端：StatsBar tooltip 渲染单元测试，覆盖 7 个卡片的 tooltip 文本") to in-scope items.
