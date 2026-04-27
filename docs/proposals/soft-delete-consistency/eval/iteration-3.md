---
date: "2026-04-27"
doc_dir: "docs/proposals/soft-delete-consistency/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval -- Iteration 3

**Score: 93/100** (target: 90)

```
+-----------------------------------------------+----------+----------+------------+
|                PROPOSAL QUALITY SCORECARD      |          |          |            |
+-----------------------------------------------+----------+----------+------------+
| Dimension                    | Score    | Max      | Status     |
+-----------------------------------------------+----------+----------+------------+
| 1. Problem Definition        |  20      |  20      | PASS       |
|    Problem clarity           |  7/7     |          |            |
|    Evidence provided         |  7/7     |          |            |
|    Urgency justified         |  6/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 2. Solution Clarity          |  19      |  20      | OK         |
|    Approach concrete         |  6/7     |          |            |
|    User-facing behavior      |  7/7     |          |            |
|    Differentiated            |  6/6     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 3. Alternatives Analysis     |  15      |  15      | PASS       |
|    Alternatives listed (>=2) |  5/5     |          |            |
|    Pros/cons honest          |  5/5     |          |            |
|    Rationale justified       |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 4. Scope Definition          |  15      |  15      | PASS       |
|    In-scope concrete         |  5/5     |          |            |
|    Out-of-scope explicit     |  5/5     |          |            |
|    Scope bounded             |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 5. Risk Assessment           |  15      |  15      | PASS       |
|    Risks identified (>=3)    |  5/5     |          |            |
|    Likelihood + impact rated |  5/5     |          |            |
|    Mitigations actionable    |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| 6. Success Criteria          |  15      |  15      | PASS       |
|    Measurable                |  5/5     |          |            |
|    Coverage complete         |  5/5     |          |            |
|    Testable                  |  5/5     |          |            |
+-----------------------------------------------+----------+----------+------------+
| TOTAL                        |  93      |  100     | PASS       |
+-----------------------------------------------+----------+----------+------------+
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Solution paragraph (line 17) | Solution summary does not mention the generic constraint split mechanism. The paragraph says "add NotDeleted scope to all repository query methods" but the actual implementation requires a type-level split (`softDeletable` vs `nonSoftDeletable` constraints) -- this is only revealed in Scope item 1. A reader scanning only the Solution section would underestimate the technical complexity. | Approach: -1 |

---

## Attack Points

### Attack 1: Solution Clarity -- Generic constraint split is the hardest part but buried in scope

**Where**: "Systematically add NotDeleted scope to all repository query methods for entities with BaseModel, fix SubItem soft-delete implementation, and align the sub_items unique index." (line 17)
**Why it's weak**: The solution paragraph describes the work as "systematically add NotDeleted" -- implying a mechanical find-and-replace. But the generic helper split (softDeletable vs nonSoftDeletable type constraints) is architecturally significant. It is the highest-risk item (rated High/High in Risks), yet it reads as a footnote in Scope item 1: "Challenge: ProgressRecord and StatusHistory don't have deleted_flag; need separate generic constraints." The Solution section should surface this as a first-class design decision, not defer it to scope.
**What must improve**: Add 1-2 sentences to the Solution paragraph explicitly calling out the generic constraint split: "This requires splitting the generic query helpers into soft-deletable and non-soft-deletable type constraints, since ProgressRecord and StatusHistory lack deleted_flag."

### Attack 2: Solution Clarity -- No mention of rollback strategy

**Where**: The entire proposal describes forward execution but never addresses what happens if the migration or code change goes wrong.
**Why it's weak**: The unique index change on pmw_sub_items (Scope item 8) modifies a production index. If the ALTER TABLE fails or the new code introduces regressions, there is no rollback path documented. The risk table says "Medium likelihood" for the index change requiring coordination but the mitigation is only "Document the ALTER SQL." A production-safe proposal should at least acknowledge rollback -- even a one-liner like "Index change is additive; rollback is a DROP INDEX + CREATE INDEX with the original columns."
**What must improve**: Add a brief rollback note to Risk 4's mitigation, or add a "Rollback" row to the Alternatives assessment for the recommended approach.

### Attack 3: Alternatives Analysis -- "Do nothing" pros/cons are asymmetric

**Where**: Alternative 3 (Do nothing) Pros: "Zero effort. No regression risk from code changes." Cons: "Phantom records persist for Team/Role/SubItem. Sub-item re-creation with same code remains broken. The bug is user-reported and affects data integrity -- inaction is not viable." (line 63-64)
**Why it's weak**: The Cons section editorializes with "inaction is not viable" -- this is the rationale, not a con. A truly honest "do nothing" analysis would list the concrete consequences (status quo) without the persuasive language, since the rationale section exists precisely to make the argument. This is minor but introduces bias into what should be neutral analysis.
**What must improve**: Move "inaction is not viable" to the Rationale section. Keep Cons as factual: "Phantom records persist for Team/Role/SubItem. Sub-item re-creation with same code remains broken. User-reported data integrity bug remains unresolved."

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 2): Risk Assessment -- Likelihood/impact ratings lack justification | YES | Each risk now has a bold **Justification** sentence with specific reasoning. Risk 1: "High likelihood because the generic helper is used by 6 entity types, 2 of which lack deleted_flag." Risk 3: "Low likelihood because SubItem soft-delete was recently added and production usage is minimal." |
| Attack 2 (iter 2): Success Criteria -- Missing verification for non-soft-deletable entity types | YES | Success criterion 4 now reads: "FindByID[T] and FindByIDs[T] return all records (unfiltered by NotDeleted) when T is a non-soft-deletable type (ProgressRecord, StatusHistory) -- verified by explicit test cases for both entity types after the generic constraint split." |
| Attack 3 (iter 2): Problem Definition -- Urgency section remains thin | YES | Urgency now includes P1 classification, user-reported status, scope ("7 repository files, ~25 query methods"), specific consequences ("Deleted roles leak into permission checks, potentially granting or denying access incorrectly"), and states "no workaround." |

---

## Verdict

- **Score**: 93/100
- **Target**: 90/100
- **Gap**: 0 points (target exceeded by 3)
- **Action**: Target reached. Proposal is ready for implementation. Remaining gap is cosmetic (solution paragraph completeness, rollback mention, editorial neutrality in do-nothing cons) and does not block execution.
