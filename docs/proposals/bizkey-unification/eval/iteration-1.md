---
date: "2026-04-28"
doc_dir: "docs/proposals/bizkey-unification/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 78/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  18      │  20      │ ✅          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   7/7    │          │            │
│    Urgency justified         │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  14      │  20      │ ⚠️          │
│    Approach concrete         │   6/7    │          │            │
│    User-facing behavior      │   3/7    │          │            │
│    Differentiated            │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  13      │  15      │ ✅          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  13      │  15      │ ⚠️          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  12      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   4/5    │          │            │
│    Mitigations actionable    │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  13      │  15      │ ⚠️          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   3/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  83      │  100     │            │
│ Deductions                   │  -5      │          │            │
│ TOTAL                        │  78      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Urgency section | Vague statistic: "~50% chance of introducing the same class of bug" — no basis given | -2 pts |
| Scope vs. Success Criteria | Scope says "All service interfaces: replace uint ID params" but success criteria only verify 2 specific functions (isPMRole, GetTeamBizKey). UpdateMemberRole, InviteMember, and other methods have no coverage criterion. | -3 pts |

---

## Attack Points

### Attack 1: Solution Clarity — user-facing behavior absent

**Where**: The entire "Proposed Solution" section describes only internal type changes. The closest thing to observable behavior is buried in the Problem section: "every progress record written to the database has a wrong team_key value."

**Why it's weak**: The rubric requires describing what the end user experiences — the observable outcome, not the internals. This proposal never states it. A reader cannot answer: "after this change, what is different from a user's perspective?" The data-correctness fix (progress records now have valid team_key) is the actual user-observable outcome and it goes unnamed in the solution.

**What must improve**: Add an explicit "Observable outcome" statement in the solution: e.g., "Progress records will contain correct snowflake team_key values; queries filtering by team_key will return accurate results." Even for internal refactors, the observable contract must be stated.

---

### Attack 2: Success Criteria — coverage gap on "all service interfaces"

**Where**: Scope states "All service interfaces: replace all `teamID uint`, `userID uint`, `roleID uint` parameters with `int64` bizKey equivalents where those values originate from external input." Success criteria verify only: `isPMRole` accepts `int64`, `GetTeamID` is removed, `progress_service.go` TeamKey fix.

**Why it's weak**: `UpdateMemberRole(ctx, pmID, teamID, targetUserID, roleID uint)` is called out by name in the evidence section as a bug site, yet there is no success criterion verifying it was fixed. Same for `InviteMember`. The criteria cover 3 of the ~6 named bug sites and say nothing about the broader "all service interfaces" scope item. A reviewer cannot confirm the work is done.

**What must improve**: Add a criterion that covers the full scope: e.g., "Zero occurrences of `uint` typed parameters named `*ID` or `*Key` in service interface definitions where the value originates from external input (verified by grep or go vet custom analyzer)." Or enumerate each affected method explicitly.

---

### Attack 3: Scope Definition — "all" items are unbounded

**Where**: In-scope lists "All service interfaces," "All handler call sites," and "All unit and integration tests that mock or assert on these signatures."

**Why it's weak**: None of these items have a count or file list. A team cannot estimate effort, cannot know when they are done, and cannot detect if they missed something. "All handler call sites" could mean 3 files or 15. The rubric criterion is whether "a team can execute this in a defined timeframe" — they cannot, because the scope has no boundary on the most expensive items.

**What must improve**: Replace "all" with an enumerated list or a grep-derived count. E.g., "Handler call sites: 4 files — team_handler.go, progress_handler.go, view_handler.go, report_handler.go (confirmed by `grep -r GetTeamID`)." This makes the scope verifiable and effort estimable.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 78/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Strongest areas are problem definition and alternatives analysis. The three weakest areas — user-facing behavior, success criteria coverage, and unbounded scope items — are all fixable with targeted additions rather than rewrites.
