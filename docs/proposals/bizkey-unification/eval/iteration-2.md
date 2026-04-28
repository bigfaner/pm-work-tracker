---
date: "2026-04-28"
doc_dir: "docs/proposals/bizkey-unification/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 90/100** (target: N/A)

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
│ 2. Solution Clarity          │  17      │  20      │ ⚠️          │
│    Approach concrete         │   6/7    │          │            │
│    User-facing behavior      │   6/7    │          │            │
│    Differentiated            │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  13      │  15      │ ✅          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  15      │  15      │ ✅          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  14      │  15      │ ✅          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   5/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  15      │  15      │ ✅          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   5/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  92      │  100     │            │
│ Deductions                   │  -2      │          │            │
│ TOTAL                        │  90      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Urgency section | Vague statistic carried over from iteration 1: "~50% chance of introducing the same class of bug" — still no empirical or analytical basis given | -2 pts |

---

## Attack Points

### Attack 1: Problem Definition — urgency statistic still unsubstantiated

**Where**: "Each new feature that touches team/user/role IDs has a ~50% chance of introducing the same class of bug because the type system provides no guidance."

**Why it's weak**: This was flagged in iteration 1 and is unchanged. The "~50%" figure is presented as a probability but has no derivation — not from historical bug rate, not from a count of call sites vs. correct usages, not from any model. It reads as a rhetorical flourish. The two preceding urgency points (active data corruption, silent truncation on 64-bit systems) are concrete and sufficient; this third point undermines them by introducing an unverifiable claim.

**What must improve**: Either remove the statistic and replace it with a count-based argument ("X of Y call sites in the current codebase already have this bug, with Z more sites at risk"), or drop the third urgency point entirely — the first two are strong enough to justify action.

---

### Attack 2: Solution Clarity — "where those values originate from external input" is an unresolved judgment call

**Where**: "Service interfaces replace all `teamID uint`, `userID uint`, `roleID uint` parameters with `int64` bizKey equivalents where those values originate from external input (URL params, request bodies)."

**Why it's weak**: The qualifier "where those values originate from external input" is the hardest part of the implementation decision, and the proposal leaves it as a judgment call for the implementer. There is no decision rule for ambiguous cases — e.g., a service method called both from a handler (external input) and from another service (internal). The scope section lists 8 service files but does not enumerate which specific parameters in each file are in-scope vs. out-of-scope. A developer reading this could make different calls on the same method and both be "following the proposal."

**What must improve**: Add a decision rule: e.g., "If the value enters the system via an HTTP request (URL param, path param, or request body), it must be `int64` bizKey at the service boundary. If the value is derived entirely within the repository layer (FK lookup result), it stays `uint`." Then apply this rule to enumerate the specific parameters being changed in each of the 8 service files, or at minimum the ambiguous cases.

---

### Attack 3: Alternatives Analysis — cons for chosen approach are not calibrated to the now-known scope

**Where**: "Full: service layer uses int64 bizKey (this proposal) | ... | Cons: Larger change, touches many files"

**Why it's weak**: The scope section now quantifies the change as 35 files (8 service + 7 handler + 20 test). The alternatives table was not updated to reflect this. "Touches many files" is the same vague language that was in the original; a reader comparing alternatives cannot assess whether 35 files is a large or small risk relative to the team's velocity or test coverage. The cons column should be calibrated: e.g., "35 files changed; all are covered by existing tests, so regression risk is low" — or the opposite if coverage is thin.

**What must improve**: Update the cons for the chosen approach to reference the actual scope magnitude and pair it with the coverage/risk context: "35 files changed across service, handler, and test layers; compile-time enforcement and existing test coverage reduce regression risk to low."

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Solution Clarity — user-facing behavior absent | ✅ | Solution section now opens with: "After this change, progress records will contain correct snowflake `team_key` values and role permission checks will resolve against the correct bizKey — eliminating the silent data corruption currently affecting every progress write." |
| Success Criteria — coverage gap on UpdateMemberRole, InviteMember | ✅ | Added explicit criterion: "`UpdateMemberRole` and `InviteMember` signatures use `int64` for all ID/key params that originate from external input (verified by grep on `team_service.go` interface definition)" and the zero-uint grep criterion covering all service interfaces. |
| Scope Definition — "all" items unbounded | ✅ | Scope now enumerates file counts: "8 files", "7 files", "20 files: 10 handler tests, 7 service tests, `team_scope_test.go`, `views_reports_test.go`, `helpers.go`". Effort is now estimable. |

---

## Verdict

- **Score**: 90/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: All three iteration-1 attacks were addressed. Remaining weaknesses are the unsubstantiated urgency statistic (cosmetic but penalized), an ambiguous decision rule for "external input" origin (implementation risk), and thin cons in the alternatives table (presentation gap). None block execution — the proposal is ready to proceed to PRD.
