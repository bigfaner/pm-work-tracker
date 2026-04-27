---
date: "2026-04-27"
doc_dir: "docs/proposals/soft-delete-consistency/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 62/100** (target: 90)

```
+-----------------------------------------------+----------+----------+------------+
|                PROPOSAL QUALITY SCORECARD      |          |          |            |
+-----------------------------------------------+----------+----------+------------+
| Dimension                    | Score    | Max      | Status     |
+-----------------------------------------------+----------+----------+------------+
| 1. Problem Definition        |  15      |  20      | :warning:  |
|    Problem clarity           |  6/7     |          |            |
|    Evidence provided         |  6/7     |          |            |
|    Urgency justified         |  3/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 2. Solution Clarity          |  9       |  20      | :x:        |
|    Approach concrete         |  4/7     |          |            |
|    User-facing behavior      |  2/7     |          |            |
|    Differentiated            |  3/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 3. Alternatives Analysis     |  10      |  15      | :warning:  |
|    Alternatives listed (>=2) |  5/5     |          |            |
|    Pros/cons honest          |  2/5     |          |            |
|    Rationale justified       |  3/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 4. Scope Definition          |  13      |  15      | :white_check_mark: |
|    In-scope concrete         |  5/5     |          |            |
|    Out-of-scope explicit     |  4/5     |          |            |
|    Scope bounded             |  4/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 5. Risk Assessment           |  7       |  15      | :x:        |
|    Risks identified (>=3)    |  4/5     |          |            |
|    Likelihood + impact rated |  0/5     |          |            |
|    Mitigations actionable    |  3/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 6. Success Criteria          |  11      |  15      | :warning:  |
|    Measurable                |  4/5     |          |            |
|    Coverage complete         |  3/5     |          |            |
|    Testable                  |  4/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| TOTAL                        |  62*     |  100     |            |
+-----------------------------------------------+----------+----------+------------+

* Includes -3 inconsistency deduction (Success Criteria don't cover Role repo in-scope item 7)
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Solution (entire section) | Solution section is a single vague paragraph; concrete details are only in Scope | Approach: -3 (partial) |
| Solution | No user-facing behavior described anywhere | User-facing behavior: -5 (partial) |
| Urgency | Single sentence: "Data integrity issue. Users see phantom records and get confusing errors." | Urgency: -3 (partial) |
| Risks (entire table) | No likelihood or impact ratings on any risk | Likelihood+impact: -5 (full) |
| Alternatives | No honest pros AND cons for each alternative; straw-man arguments | Pros/cons: -3 (partial) |
| Success Criteria vs Scope | In-scope item 7 (Role repo) has no corresponding success criterion | Inconsistency: -3 |

---

## Attack Points

### Attack 1: Solution Clarity — No user-facing behavior described

**Where**: The entire proposal lacks any description of what the end user experiences. The closest is in Urgency: "Users see phantom records and get confusing errors."
**Why it's weak**: A proposal reader cannot understand the before/after user experience. What does a "phantom record" look like in the UI? Which screens are affected? What error does the user see when re-creating a sub-item with a soft-deleted code? Without this, reviewers cannot assess whether the fix actually solves the user-facing problem.
**What must improve**: Add a "User-Facing Impact" subsection under Solution describing concrete before/after scenarios. For example: "Before: deleted roles appear in the role dropdown on the team settings page. After: deleted roles are filtered from all dropdowns and list views."

### Attack 2: Risk Assessment — Zero likelihood or impact ratings

**Where**: The Risks table contains four risks but the format is `| Risk | Mitigation |` with no columns for likelihood or impact.
**Why it's weak**: The rubric requires "Is the assessment honest? Not all 'low likelihood, high impact'?" — but the proposal doesn't rate anything at all. A reader cannot prioritize mitigation efforts. Risk 1 (generic constraint split) could be high-impact and high-likelihood since it touches every entity, or it could be trivial — the proposal gives no signal.
**What must improve**: Add Likelihood (High/Medium/Low) and Impact (High/Medium/Low) columns to the risk table. For example: "| FindByID[T] generic can't blindly add NotDeleted | Likelihood: High | Impact: High | Split generic constraint... |"

### Attack 3: Alternatives Analysis — Superficial pros/cons with no honest trade-offs

**Where**: "Fix only repos with active delete flows — Less code, but User/MainItem/ItemPool will break when delete features are added later."
**Why it's weak**: Alternative 2 gets a single con and no pros. Alternative 1 gets a single pro and no cons. This is a straw-man setup: the recommended option looks obviously superior because its drawbacks are never stated. Honest trade-offs would acknowledge: Alternative 1 touches more files (higher regression risk, larger PR, more review time) vs. Alternative 2 being surgical but incomplete. The cost difference in code changes, testing effort, and deployment risk is entirely absent.
**What must improve**: Provide a structured pros/cons table for each alternative. For Alternative 1, acknowledge cons: "Touches 7 repo files + generic helpers; larger blast radius; requires comprehensive regression testing." For Alternative 2, acknowledge pros: "Minimal diff; lower regression risk; faster to ship."

---

## Previous Issues Check

<!-- Only for iteration > 1 -->

N/A — Iteration 1.

---

## Verdict

- **Score**: 62/100
- **Target**: 90/100
- **Gap**: 28 points
- **Action**: Continue to iteration 2 — biggest gains available in Solution Clarity (+11 possible), Risk Assessment (+8 possible), and Alternatives Analysis (+5 possible)
