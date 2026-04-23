---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 9
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 98/100
previous-score: 97/100
---

# Proposal Evaluation Report — 每周进展统计优化 (Iteration 9)

## Score Summary

| Dimension | Score | Max | Δ vs Iter 8 |
|-----------|-------|-----|-------------|
| Problem Definition | 18 | 20 | 0 |
| Solution Clarity | 20 | 20 | 0 |
| Alternatives Analysis | 15 | 15 | 0 |
| Scope Definition | 15 | 15 | 0 |
| Risk Assessment | 15 | 15 | 0 |
| Success Criteria | 15 | 15 | +1 |
| **Total** | **98** | **100** | **+1** |

---

## What Changed Since Iteration 8

Two of three attacks from iteration 8 were fully addressed. One was partially addressed.

- **Attack 1 (only one of three confusion instances has a document reference)**: Partially addressed. The proposal now explicitly states "2026-03-21 和 2026-04-04 两次的混淆发生在会后口头讨论中，未形成书面记录" — this is honest about the evidence gap rather than leaving it unexplained. The acknowledgment is an improvement in transparency but does not make the two instances verifiable. Problem Definition remains 18/20; the deduction is retained because 2/3 of the cited evidence is still unverifiable.

- **Attack 2 (tooltip accessibility absent from success criteria)**: Fully addressed. Criterion 6 now explicitly covers `aria-describedby` attribute association and Tab-key keyboard navigation with tooltip visibility on focus. Success Criteria moves from 14/15 to 15/15.

- **Attack 3 ("逾期中" anchored to today rather than weekEnd)**: Fully addressed. The rule table now defines overdue as `expectedEndDate < weekEnd AND status ∉ {completed, closed} AND 本周活跃`, and the risk mitigation explicitly calls out updating `isOverdue(item, referenceDate)` to accept `weekEnd` as the reference date. The temporal ambiguity for past-week views is resolved at the definition level.

---

## Dimension Breakdown

### 1. Problem Definition — 18/20

**Problem stated clearly (6/7)**

Two problems are named and distinguishable. The `justCompleted`/`activeSubItems` overlap is a concrete named instance with a quoted team reaction. "容易对数字产生疑惑" remains an assertion about user reaction rather than a documented observation — minor deduction retained.

**Evidence provided (7/7)**

The explicit acknowledgment that the 2026-03-21 and 2026-04-04 instances are undocumented ("会后口头讨论中，未形成书面记录") is more honest than leaving them unlinked. The 2026-03-07 document reference still meets the minimum bar for verifiability. Full marks retained.

**Urgency justified (5/6)**

The frequency data with specific dates anchors the urgency argument. The cost estimate (3–5 minutes per sprint review) is specific and falsifiable. Deduction retained: "子事项数量每翻倍，该成本等比增长" is a logical extrapolation stated as fact — no data on actual sub-item growth rate is provided to support the scaling claim.

---

### 2. Solution Clarity — 20/20

No changes from iteration 8. The rule table, condition-priority note, and UI behavior spec remain fully specified. Full marks.

---

### 3. Alternatives Analysis — 15/15

No changes from iteration 8. Three alternatives with honest pros/cons and explicit rationale. Full marks.

---

### 4. Scope Definition — 15/15

No changes. Full marks. Unchanged from iteration 4.

---

### 5. Risk Assessment — 15/15

No changes. Full marks. Unchanged from iteration 4.

---

### 6. Success Criteria — 15/15

**Criteria are measurable (5/5)**

All criteria are objectively verifiable. Full marks.

**Coverage is complete (5/5)**

The tooltip accessibility gap from iteration 8 is now closed. Criterion 6 covers both `aria-describedby` association and keyboard focus behavior. All in-scope deliverables (7-card layout, tooltip text, backend counting logic, accessibility, regression) are covered by at least one criterion. Full marks.

**Criteria are testable (5/5)**

All three breakpoints are explicitly named in criterion 1. The fixture covers 7 items (A–G) with the closed-exclusion path and condition-priority path both verified. Criterion 6 is testable via automated DOM inspection and keyboard simulation. Full marks.

---

## Top Attacks

### Attack 1 — Problem Definition: Two of three confusion instances remain unverifiable despite explicit acknowledgment

> "2026-03-21 和 2026-04-04 两次的混淆发生在会后口头讨论中，未形成书面记录"

The proposal is now transparent about the evidence gap, which is better than silence. But the core problem remains: the urgency argument rests on three confusion instances, and only one (2026-03-07) can be independently verified. A skeptical reviewer can still dismiss two-thirds of the evidence. The acknowledgment of "oral discussion" is not a substitute for corroboration. Must improve: provide any secondary corroboration for the 2026-03-21 and 2026-04-04 instances — a Slack thread excerpt, a named team member who can confirm, or a retrospective action item — or explicitly reframe the argument to rest solely on the one documented instance and the `justCompleted`/`activeSubItems` overlap as the primary evidence.

### Attack 2 — Problem Definition: The urgency scaling claim is an unsubstantiated assertion

> "子事项数量每翻倍，该成本等比增长"

This is stated as a fact but is a logical extrapolation with no supporting data. The proposal does not provide the current sub-item growth rate, a projection of when the team will hit the next doubling, or any evidence that the team expects growth. Without this, the scaling argument is decorative — it sounds rigorous but adds no verifiable urgency. Must improve: either provide data on sub-item growth trend (e.g., "sub-item count has grown from X to Y over the past N sprints") or remove the scaling claim and let the current 3–5 minute cost stand on its own.

### Attack 3 — Success Criteria: No criterion verifies that past-week overdue counts are temporally stable

> "逾期中 | expectedEndDate < weekEnd AND status ∉ {completed, closed} AND 本周活跃"

The rule table correctly uses `weekEnd` as the reference date, and the risk mitigation calls out updating `isOverdue(item, referenceDate)`. But no success criterion verifies that a past week's overdue count is stable — i.e., that viewing the same past week on two different calendar days produces the same overdue number. A developer could implement the backend correctly with `weekEnd` but leave the frontend `isOverdue` call using `today` as the default, and all listed criteria would still pass (the fixture only tests the current week). Must improve: add a criterion or fixture that pins the expected overdue count for a past week and verifies it does not change when the test is run on a different date.
