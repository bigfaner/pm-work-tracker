---
date: "2026-04-27"
doc_dir: "docs/proposals/soft-delete-consistency/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval -- Iteration 2

**Score: 86/100** (target: 90)

```
+-----------------------------------------------+----------+----------+------------+
|                PROPOSAL QUALITY SCORECARD      |          |          |            |
+-----------------------------------------------+----------+----------+------------+
| Dimension                    | Score    | Max      | Status     |
+-----------------------------------------------+----------+----------+------------+
| 1. Problem Definition        |  18      |  20      | OK         |
|    Problem clarity           |  7/7     |          |            |
|    Evidence provided         |  7/7     |          |            |
|    Urgency justified         |  4/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 2. Solution Clarity          |  18      |  20      | OK         |
|    Approach concrete         |  6/7     |          |            |
|    User-facing behavior      |  7/7     |          |            |
|    Differentiated            |  5/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 3. Alternatives Analysis     |  14      |  15      | OK         |
|    Alternatives listed (>=2) |  5/5     |          |            |
|    Pros/cons honest          |  4/5     |          |            |
|    Rationale justified       |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 4. Scope Definition          |  15      |  15      | PASS       |
|    In-scope concrete         |  5/5     |          |            |
|    Out-of-scope explicit     |  5/5     |          |            |
|    Scope bounded             |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 5. Risk Assessment           |  12      |  15      | OK         |
|    Risks identified (>=3)    |  5/5     |          |            |
|    Likelihood + impact rated |  3/5     |          |            |
|    Mitigations actionable    |  4/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 6. Success Criteria          |  12      |  15      | OK         |
|    Measurable                |  4/5     |          |            |
|    Coverage complete         |  4/5     |          |            |
|    Testable                  |  4/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| TOTAL                        |  86*     |  100     |            |
+-----------------------------------------------+----------+----------+------------+

* Includes -3 inconsistency deduction (Success Criteria do not verify non-soft-deletable types work after generic constraint split)
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Urgency section | Two sentences only: "Data integrity issue. Users see phantom records and get confusing errors." No deadline, no user impact quantification, no triage context | Urgency: -2 |
| Solution paragraph | Solution summary is one sentence; real concreteness lives in Scope section, not Solution itself. Reader must cross-reference to understand the approach | Approach: -1 |
| Alternatives tables | "Effort" and "Blast radius" rows largely restate Pros/Cons content rather than adding independent analysis dimensions | Pros/cons: -1 |
| Risk ratings | Likelihood/Impact ratings are single words with no justification. Risk 1 is "High/High" -- why? Risk 3 is "Low/Medium" -- based on what? | Likelihood+impact: -2 |
| Risk 2 mitigation | "Run full test suite before merge" is standard practice, not a targeted mitigation specific to this risk | Mitigations: -1 |
| Success criteria vs Scope | In-scope item 1 explicitly calls out the generic constraint split challenge for ProgressRecord/StatusHistory, but no success criterion verifies non-soft-deletable types still return correct results after the refactor | Inconsistency: -3 |
| Success criteria | Criterion 4 "All existing tests pass" is tautological -- it applies to any change and does not define proposal-specific success | Measurable: -1 |
| Success criteria coverage | No criterion explicitly tests that ProgressRecord and StatusHistory queries remain unaffected after the generic helper split | Coverage: -1 |
| Success criteria testability | Criterion 5 is partially testable but does not enumerate which entity types must be tested, leaving coverage ambiguous | Testable: -1 |

---

## Attack Points

### Attack 1: Risk Assessment -- Likelihood/impact ratings lack justification

**Where**: "FindByID[T] generic can't blindly add NotDeleted (ProgressRecord/StatusHistory lack deleted_flag) | High | High"
**Why it's weak**: The ratings are bare labels without supporting reasoning. Why is Risk 1 "High" likelihood? Is it because the generic helper is used by N callers? Is it because the constraint split is architecturally complex? Why is Risk 3 "Low" likelihood for existing GORM-style soft-delete data inconsistency? The reader has no basis to agree or disagree with the ratings. Compare this to the Alternatives rationale, which is specific and persuasive -- the risk section should match that standard.
**What must improve**: Add 1-2 sentences of justification per risk. For example: "Likelihood: High -- the generic helper is used by 6 entity types, 2 of which lack deleted_flag, making a naive NotDeleted addition guaranteed to fail at compile time or silently filter wrong records."

### Attack 2: Success Criteria -- Missing verification for non-soft-deletable entity types

**Where**: Scope item 1 says "Challenge: ProgressRecord and StatusHistory don't have deleted_flag; need separate generic constraints" but no success criterion covers this.
**Why it's weak**: This is the most technically risky part of the entire proposal (rated High/High in Risks), yet the success criteria silently skip it. The six criteria verify that soft-deleted records are filtered out, but none verify that ProgressRecord and StatusHistory still return all records (including any that might accidentally match a filter condition) after the generic constraint split. A reviewer cannot confirm this critical path is covered.
**What must improve**: Add a success criterion: "FindByID[T] and FindByIDs[T] return all records (unfiltered by NotDeleted) when T is a non-soft-deletable type (ProgressRecord, StatusHistory). Write explicit test cases for both entity types."

### Attack 3: Problem Definition -- Urgency section remains thin

**Where**: "Urgency: Data integrity issue. Users see phantom records and get confusing errors."
**Why it's weak**: This is identical in substance to iteration 1. No triage context (P0/P1/P2?), no user impact scope (how many users? which teams?), no deadline, no consequences of delay. The rest of the proposal is now detailed and persuasive, making this two-sentence urgency section feel like an afterthought by comparison. A PM reading this cannot prioritize this against other work.
**What must improve**: Expand urgency to 3-4 sentences. Quantify impact: "Affects all teams using role management and sub-item features." Add triage context: "P1 -- user-reported data integrity bug." State consequences: "Without fix, deleted roles leak into permission checks, potentially granting or denying access incorrectly."

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Solution Clarity -- No user-facing behavior described | YES | New "User-Facing Impact" subsection with three Before/After scenarios covering Team/Role, SubItem, and User/MainItem/ItemPool |
| Attack 2: Risk Assessment -- Zero likelihood or impact ratings | PARTIAL | Likelihood and Impact columns added with ratings (High/High, Medium/High, Low/Medium, Medium/Low), but ratings lack justification |
| Attack 3: Alternatives Analysis -- Superficial pros/cons | YES | Each alternative now has a structured table with Pros, Cons, Effort, Blast radius, and Maintenance rows. Alternative 1 acknowledges wider regression surface. Alternative 2 gets honest pros. |

---

## Verdict

- **Score**: 86/100
- **Target**: 90/100
- **Gap**: 4 points
- **Action**: Continue to iteration 3 -- biggest gains available in Risk Assessment (+3 possible via rating justifications) and Success Criteria (+3 possible via covering non-soft-deletable types and removing tautological criterion). Closing the urgency gap (+2) would bring the proposal to target.
