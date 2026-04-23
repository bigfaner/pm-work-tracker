---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 8
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 97/100
previous-score: 94/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 8)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 7 |
|-----------|-------|-----|-------------|
| Problem Definition | 18 | 20 | +1 |
| Solution Clarity | 20 | 20 | 0 |
| Alternatives Analysis | 15 | 15 | +1 |
| Scope Definition | 15 | 15 | 0 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 14 | 15 | +1 |
| **Total** | **97** | **100** | **+3** |

---

## What Changed Since Iteration 7

All three attacks from iteration 7 were raised. All three were addressed in this iteration.

- **Attack 1 (evidence dates are self-reported with no external artifact)**: Addressed. The proposal now includes "2026-03-07 会议记录见 docs/sprint-reviews/2026-03-07.md" — a pointer to a specific document a team member can check. The minimum bar ("at least one verifiable artifact") is met. Problem Definition moves from 17/20 to 18/20. Deduction retained: only one of three confusion instances (2026-03-07) has a document reference; the other two (2026-03-21, 2026-04-04) remain unlinked. The reference is also an internal repo path, not an externally accessible URL.

- **Attack 2 (mobile fallback and <768px breakpoint absent from success criteria)**: Fully addressed. Criterion 1 now explicitly covers all three breakpoints including "<768px 宽度下呈现 2 列多行布局且无内容截断". Criterion 2 now explicitly covers the mobile click-to-toggle fallback. Success Criteria moves from 13/15 to 14/15. Remaining gap: tooltip accessibility (keyboard navigation, screen reader) is specified in the alternatives cons but still absent from success criteria.

- **Attack 3 (option 3 cons understate tooltip accessibility work)**: Fully addressed. The cons column for option 3 now enumerates: keyboard focus trigger, `aria-describedby` association, three-breakpoint layout testing, and mobile fallback implementation. Alternatives Analysis moves from 14/15 to 15/15.

---

## Dimension Breakdown

### 1. Problem Definition — 18/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete, named instance with a quoted team reaction. "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (7/7)**

The addition of "2026-03-07 会议记录见 docs/sprint-reviews/2026-03-07.md" meets the minimum bar: a reader can now check a specific document rather than relying solely on the author's word. Full marks for this criterion. Note: only one of three confusion instances is linked; the reference is an internal repo path rather than an externally accessible URL — these are weaknesses but do not drop below the 7-point threshold.

**Urgency justified (5/6)**

The frequency data with specific dates anchors the urgency argument. The cost estimate is specific and falsifiable. Deduction retained: the growth claim ("子事项数量每翻倍，该成本等比增长") is still an assertion about future behavior, not an observed trend.

---

### 2. Solution Clarity — 20/20

No changes from iteration 7. The rule table, condition-priority note, and UI behavior spec remain fully specified. Full marks.

---

### 3. Alternatives Analysis — 15/15

**At least 2 alternatives listed (5/5)**

Three alternatives listed, including a minimal-change option. Full marks.

**Pros/cons for each (5/5)**

Option 3 cons now enumerate: layout adjustment, DTO changes, keyboard focus trigger, `aria-describedby` association, three-breakpoint testing, and mobile fallback implementation. The full cost is now visible to a reader evaluating the trade-off. Full marks.

**Rationale for chosen approach (5/5)**

The "方案选择说明" paragraph addresses all three options with explicit reasoning. Full marks.

---

### 4. Scope Definition — 15/15

No changes. Full marks. Unchanged from iteration 4.

---

### 5. Risk Assessment — 15/15

No changes. Full marks. Unchanged from iteration 4.

---

### 6. Success Criteria — 14/15

**Criteria are measurable (5/5)**

All criteria are objectively verifiable. Full marks.

**Coverage is complete (4/5)**

The mobile fallback and <768px breakpoint gaps are now closed. Remaining gap: tooltip accessibility (keyboard navigation, screen reader `aria-describedby`) is specified in the alternatives cons and in the scope's in-scope item ("tooltip 需实现键盘焦点触发与 `aria-describedby` 关联") but absent from success criteria. A developer could ship tooltips without keyboard focus support and without `aria-describedby` associations, and all listed criteria would still pass.

**Criteria are testable (5/5)**

All three breakpoints are now explicitly named in criterion 1. The fixture covers 7 items (A–G) with the closed-exclusion path verified. Full marks.

---

## Top Attacks

### Attack 1 — Problem Definition: Only one of three confusion instances has a document reference; the other two remain unlinked

> "近 5 次 sprint review（2026-02-21、2026-03-07、2026-03-21、2026-04-04、2026-04-18）中有 3 次（2026-03-07、2026-03-21、2026-04-04）出现团队成员无法解释"活跃"数字构成的情况（典型反应："新完成的事项为什么也算活跃？"；2026-03-07 会议记录见 docs/sprint-reviews/2026-03-07.md）"

The 2026-03-07 reference is a real improvement. But the claim rests on three confusion instances, and only one is linked. The 2026-03-21 and 2026-04-04 instances remain unverifiable. A skeptical reader can still dismiss two-thirds of the evidence as unsubstantiated. Must improve: add document references for the 2026-03-21 and 2026-04-04 instances, or explicitly note that only the 2026-03-07 notes are available and explain why the others are not linked.

### Attack 2 — Success Criteria: Tooltip accessibility is in scope and in alternatives cons but absent from success criteria

> "tooltip 需实现键盘焦点触发与 `aria-describedby` 关联（屏幕阅读器可访问性）"

This line appears in the alternatives cons for option 3, and the in-scope list implies it is a deliverable. But no success criterion covers it. A developer could ship tooltips that are mouse-only and fail screen reader traversal, and every listed criterion would still pass. Must improve: add a criterion asserting that each tooltip is reachable via keyboard focus (Tab key) and that the card element carries an `aria-describedby` attribute pointing to the tooltip content.

### Attack 3 — Problem Definition / Solution Clarity: "逾期中" uses "今天" (today) rather than the week's end date — behavior for past-week views is undefined

> "逾期中 | expectedEndDate < 今天 AND status ∉ {completed, closed} AND 本周活跃"

The overdue definition anchors to today's date, not to the viewed week's end date. For the current week this is unambiguous. For a past week, the overdue count will change every day as more items cross their `expectedEndDate` — meaning the same past week will show a different overdue count depending on when it is viewed. The proposal does not state whether past-week views are in scope or out of scope, and does not acknowledge this temporal ambiguity. Must improve: either (a) explicitly exclude past-week views from scope, or (b) clarify whether "今天" should be replaced with `weekEnd` for historical consistency, and add a success criterion or fixture that pins the expected behavior.
