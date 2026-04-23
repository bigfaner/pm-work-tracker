---
proposal: docs/proposals/weekly-stats-optimization/proposal.md
iteration: 1
evaluator: zcode-proposal-scorer
date: 2026-04-23
score: 60/100
---

# Proposal Evaluation Report — 每周进展统计优化

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 8 | 20 |
| Solution Clarity | 14 | 20 |
| Alternatives Analysis | 10 | 15 |
| Scope Definition | 13 | 15 |
| Risk Assessment | 8 | 15 |
| Success Criteria | 7 | 15 |
| **Total** | **60** | **100** |

---

## Dimension Breakdown

### 1. Problem Definition — 8/20

**Problem stated clearly (5/7)**

Three issues are named and distinguishable: opaque rules, incomplete status coverage, and an undocumented overlap between `justCompleted` and `activeSubItems`. The framing is clear enough that two readers would reach the same interpretation. Deduction: "容易对数字产生疑惑" is asserted without grounding — it describes a hypothetical user reaction, not an observed one.

**Evidence provided (2/7)**

No data, no user feedback, no support tickets, no screenshots. The entire problem section is assertion. "用户无法直接看到" and "容易对数字产生疑惑" are claims without backing. A proposal that opens with "we think users are confused" and provides zero evidence loses most of its evidence score.

**Urgency justified (1/6)**

No urgency argument exists. Why solve this now rather than next quarter? What is the cost of deferral? The proposal is silent. One point awarded only because the known bug (`justCompleted` double-counted without documentation) implies some latent confusion risk.

---

### 2. Solution Clarity — 14/20

**Approach is concrete (6/7)**

The 7-card table with explicit filter rules is the strongest part of the proposal. Each card has a name, a rule, and a tooltip string. A developer can implement directly from this table. Minor deduction: the "本周活跃" base condition uses two sub-conditions joined by OR, but the interaction between them (e.g., an item with a progress record that also ended before weekStart) is not resolved.

**User-facing behavior described (5/7)**

The tooltip strings are specified. However, there is no description of the visual layout — how do 7 cards fit in the existing bar? What happens on hover (delay, position, style)? The observable interaction is partially described but the layout change from 4 to 7 columns is mentioned only as a con in the alternatives table, not specified as a deliverable behavior.

**Distinguishes from alternatives (3/6)**

The chosen approach is labeled "采纳" but the rationale is a single phrase: "覆盖关键状态，规则透明." This does not argue why 7 cards is the right number, why `closed` and `completed` are excluded from individual cards, or why tooltip-only (option 2) is insufficient beyond restating the problem.

---

### 3. Alternatives Analysis — 10/15

**At least 2 alternatives listed (5/5)**

Three alternatives are listed, including a minimal-change option (keep 4 cards, add tooltip) that functions as the "do nothing meaningful" baseline. Full marks.

**Pros/cons for each (3/5)**

The table entries are thin. Option 1's con ("卡片过多，completed/closed 数字意义不大") is asserted without argument — why is `completed` count not meaningful? Option 2's con is a restatement of the problem, not a trade-off analysis. Option 3's con ("布局从 4 列变 7 列，需调整 UI") is accurate but understates the complexity.

**Rationale for chosen approach (2/5)**

No explicit reasoning connects the problem statement to the selection of option 3. The verdict column says "采纳" but does not explain why the pros of option 3 outweigh its cons relative to the alternatives. A reader cannot reconstruct the decision logic from what is written.

---

### 4. Scope Definition — 13/15

**In-scope items are concrete (5/5)**

Each item names a specific artifact: `WeeklyStats` DTO fields, `buildWeeklyGroups` logic, `StatsBar` component, tooltip addition, documentation. These are deliverables, not areas. Full marks.

**Out-of-scope explicitly listed (5/5)**

Four deferred items are named explicitly. Full marks.

**Scope is bounded (3/5)**

No timeframe is defined anywhere in the proposal. More critically, testing is entirely absent from scope — no unit tests for the new counting logic, no e2e tests for tooltip behavior. Given the project's TDD conventions (CLAUDE.md, `.claude/rules/testing.md`), omitting tests from scope is a meaningful gap that leaves the scope unbounded in practice.

---

### 5. Risk Assessment — 8/15

**Risks identified (2/5)**

Only 2 risks are listed. The rubric requires at least 3 meaningful risks. Missing risks include:

- **DTO backward compatibility**: adding fields to `WeeklyStats` may break existing API consumers or serialization contracts.
- **Overdue calculation performance**: `expectedEndDate < today` requires a date comparison on every active sub-item; at scale this could affect query performance.
- **Test coverage gap**: the new counting logic has no test coverage specified, creating regression risk.

**Likelihood + impact rated (3/5)**

Both listed risks have likelihood and impact ratings. The ratings are plausible (layout overflow: medium/low; definition inconsistency: low/medium). No inflation of all risks to "high impact."

**Mitigations are actionable (3/5)**

Both mitigations are actionable: `flex-wrap/grid` for layout, reuse `isOverdue` from `lib/status.ts` for consistency. However, only 2 mitigations exist because only 2 risks were identified.

---

### 6. Success Criteria — 7/15

**Criteria are measurable (3/5)**

- "统计栏显示 7 个卡片" — measurable.
- "每个卡片 hover 时显示 tooltip" — measurable (presence only, not content).
- "pending/pausing/overdue 三个新字段在后端**正确计算**" — **vague language penalty applied (-2)**. "正确计算" is not objectively verifiable without test cases or example inputs/outputs.
- "现有 4 个统计的数字与优化前保持一致" — measurable.
- "前端布局在 1280px 宽度下不溢出" — measurable.

**Coverage is complete (2/5)**

Criteria do not cover:
- Tooltip content accuracy (only presence is checked, not whether the text matches the defined rules).
- The mathematical relationships stated in the solution (cards 3–6 sum ≤ card 1; card 7 may overlap cards 3–6).
- Behavior at screen widths other than 1280px (the risk section mentions small screens but no criterion covers it).
- Accessibility of tooltips (keyboard/screen reader).

**Criteria are testable (2/5)**

"数字与定义规则一致" cannot be tested without example fixtures. No test inputs or expected outputs are provided. "正确计算" has the same problem. The criteria read as acceptance statements, not as testable specifications.

---

## Top Attacks

### Attack 1 — Problem Definition: Zero evidence for the core claim

> "用户无法直接看到每个数字的统计口径，容易对数字产生疑惑"

This is the entire evidence base for the problem. No user feedback, no support tickets, no observed confusion events, no metrics. The proposal asks for a non-trivial UI and backend change on the basis of a hypothetical. Must improve: cite at least one concrete instance — a user question, a Slack message, a bug report — or quantify the confusion (e.g., "3 of 5 team members could not explain what 'active' means without reading the code").

### Attack 2 — Risk Assessment: Only 2 risks, missing DTO compatibility and test coverage

> "Key Risks" table has 2 rows.

The rubric requires at least 3 meaningful risks. The proposal misses DTO backward compatibility (adding fields to `WeeklyStats` may silently break existing callers), the absence of test coverage for new counting logic (a regression risk given the project's TDD mandate), and potential overdue query performance at scale. Must improve: add at minimum one risk covering DTO contract changes and one covering the test coverage gap, each with likelihood/impact ratings and actionable mitigations.

### Attack 3 — Success Criteria: "正确计算" is untestable; tooltip content not verified

> "pending/pausing/overdue 三个新字段在后端正确计算"
> "每个卡片 hover 时显示 tooltip，说明统计口径"

"正确计算" triggers the vague language penalty (-2 pts) and cannot be verified without example inputs and expected outputs. The tooltip criterion checks only presence, not content — a tooltip showing "hello" would pass. Must improve: replace "正确计算" with specific test fixtures (e.g., "given sub-items A, B, C with states X, Y, Z and dates P, Q, R, overdue count = 2"), and add a criterion that tooltip text matches the strings defined in the solution table.
