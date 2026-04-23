---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 6
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 91/100
previous-score: 87/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 6)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 5 |
|-----------|-------|-----|-------------|
| Problem Definition | 16 | 20 | +1 |
| Solution Clarity | 18 | 20 | 0 |
| Alternatives Analysis | 14 | 15 | 0 |
| Scope Definition | 15 | 15 | +2 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 13 | 15 | +1 |
| **Total** | **91** | **100** | **+4** |

---

## What Changed Since Iteration 5

All three attacks from iteration 5 were raised. All three were addressed in this iteration.

- **Attack 1 (priority rule not fixture-tested)**: Addressed. Item F added to fixture: `completed + actualEndDate < weekStart + 本周有进展记录`. Expected output updated to active=6. The fixture now explicitly asserts that F counts toward active but not toward 新完成, directly exercising the condition-1-overrides-condition-2 rule. Score moves from 3/5 to 4/5 on testability. Success Criteria moves from 12/15 to 13/15.

- **Attack 2 (evidence is a single event)**: Addressed. "近 5 次 sprint review 中有 3 次出现团队成员无法解释'活跃'数字构成的情况（典型反应：'新完成的事项为什么也算活跃？'）" replaces the single code-review anecdote. This is a frequency claim with a quoted reaction — specific and falsifiable. Evidence score moves from 4/7 to 5/7. Problem Definition moves from 15/20 to 16/20.

- **Attack 3 (no timeframe, frontend test missing)**: Addressed. "预计交付：1 个 sprint（约 5 个工作日）" added. "前端：`StatsBar` tooltip 渲染单元测试，覆盖 7 个卡片的 tooltip 文本与方案表格字符串完全一致" added to in-scope items. Scope Definition moves from 13/15 to 15/15.

---

## Dimension Breakdown

### 1. Problem Definition — 16/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete, named instance with a quoted team reaction. "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (5/7)**

"近 5 次 sprint review 中有 3 次出现团队成员无法解释'活跃'数字构成的情况" is a meaningful improvement over a single event. It is a frequency claim (3/5 sprint reviews) with a specific quoted reaction. Deduction retained: this is still self-reported by the proposal author. No external validation — no support tickets, no user feedback channel, no independent observer. A reader cannot distinguish "I counted carefully" from "I estimated." The 3–5 minute urgency estimate remains labeled "约" without a measurement source.

**Urgency justified (5/6)**

The frequency data ("3 of 5 sprint reviews") now anchors the urgency argument. The cost estimate — "约 20 个活跃子事项，每次周会前手动逐项检查约需 3–5 分钟" — is specific and falsifiable. Deduction retained: "约" signals an estimate, not a measurement; the growth claim ("子事项数量每翻倍，该成本等比增长") is still an assertion about behavior, not an observed trend.

---

### 2. Solution Clarity — 18/20

**Approach is concrete (7/7)**

The "条件优先级" note is unambiguous: condition 1 (progress record this week) overrides condition 2 regardless of `actualEndDate`. The rule table covers all 7 cards with explicit formulas. A developer can implement `buildWeeklyGroups` without undocumented judgment calls. Full marks.

**User-facing behavior described (5/7)**

Tooltip strings are specified and the string-exact match criterion is in success criteria. Still no description of visual layout: card order, card width, hover interaction (delay, position, style). The layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior. A frontend developer must make undocumented choices about card ordering and hover UX. Deduction unchanged.

**Distinguishes from alternatives (6/6)**

The "方案选择说明" paragraph explicitly explains why option 1's all-time completed count is inferior to card 2 ("本周新完成"), and rebuts the "7 cards dilute attention" argument with an information-density argument. Full marks.

---

### 3. Alternatives Analysis — 14/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option. Full marks.

**Pros/cons for each (4/5)**

Option 1's con is now substantively argued. Option 2's con is well-supported. Minor deduction: option 3's con ("布局从 4 列变 7 列，需调整 UI；后端 DTO 新增 3 个字段") still understates the work — no mention of responsive behavior testing, tooltip accessibility, or the frontend test effort.

**Rationale for chosen approach (5/5)**

The "方案选择说明" paragraph addresses all three options with explicit reasoning. Full marks.

---

### 4. Scope Definition — 15/15

**In-scope items are concrete (5/5)**

Each item names a specific artifact. The frontend tooltip unit test is now explicitly listed as a deliverable. Full marks.

**Out-of-scope explicitly listed (5/5)**

Four deferred items are named explicitly. Full marks.

**Scope is bounded (5/5)**

"预计交付：1 个 sprint（约 5 个工作日）" provides a timeframe. The frontend tooltip test is now in scope, closing the gap between the success criterion and the deliverable list. Full marks.

---

### 5. Risk Assessment — 15/15

Four meaningful risks with likelihood/impact ratings and actionable mitigations. Full marks. Unchanged from iteration 4.

---

### 6. Success Criteria — 13/15

**Criteria are measurable (5/5)**

All criteria are objectively verifiable. The math invariant ("进行中(1) + 阻塞中(1) + 未开始(1) + 暂停中(1) = 4 ≤ 活跃(6)") is now an explicit criterion. Full marks.

**Coverage is complete (4/5)**

Item F closes the priority-rule gap. Remaining gap: tooltip accessibility (keyboard navigation, screen reader) is not covered by any criterion. The proposal adds tooltips to 7 cards but does not specify whether they are keyboard-accessible or screen-reader-friendly.

**Criteria are testable (4/5)**

The fixture now has 6 items (A–F) and item F directly exercises the condition-1-overrides-condition-2 rule. One testability gap remains: the `closed` status boundary. The "本周活跃" definition excludes items "未在本周开始前完成/关闭（actualEndDate < weekStart）" — the `closed` exclusion path is never tested. No fixture item has `status=closed + actualEndDate < weekStart`. A developer could implement `closed` incorrectly (e.g., treating it like `completed`) and all six fixture assertions would still pass.

---

## Top Attacks

### Attack 1 — Problem Definition: Evidence is self-reported frequency without external validation

> "近 5 次 sprint review 中有 3 次出现团队成员无法解释'活跃'数字构成的情况"

This is a meaningful improvement over a single event, but the count is self-reported by the proposal author. A reader cannot verify it. There are no support tickets, no Slack threads, no independent observer cited. The 3–5 minute urgency estimate is still labeled "约" with no measurement source. Must improve: cite a verifiable artifact (e.g., a specific sprint review date, a Slack thread, a ticket) so a reader can independently confirm the frequency claim.

### Attack 2 — Solution Clarity: Visual layout and hover UX are unspecified

> "布局从 4 列变 7 列，需调整 UI"

The proposal names 7 cards and their tooltip strings but does not specify: card display order, card width/sizing, hover trigger delay, tooltip position (above/below/auto), or what happens on mobile where hover is unavailable. A frontend developer must make all of these choices without guidance. Must improve: add a UI behavior section or reference a design spec that defines card order, responsive breakpoints, and hover interaction details.

### Attack 3 — Success Criteria: `closed` status boundary is not fixture-tested

> "未在本周开始前完成/关闭（actualEndDate < weekStart）"

The "本周活跃" definition explicitly excludes `closed` items with `actualEndDate < weekStart`, but no fixture item tests this exclusion. Items A–F cover `progressing`, `pending`, `pausing`, `completed`, and `blocking` — `closed` is absent. A developer could treat `closed` identically to `completed` (or omit it entirely from the exclusion logic) and all six fixture assertions would still pass. Must improve: add a fixture item G — e.g., `closed + actualEndDate < weekStart + 无进展记录` — and assert it does NOT count toward active, with expected active remaining at 6.
