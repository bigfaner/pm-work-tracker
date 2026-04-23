---
date: "2026-04-22"
doc_dir: "docs/features/code-quality-cleanup/prd"
iteration: "2"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 94/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅         │
│    Background three elements │  7/7     │          │            │
│    Goals quantified          │  7/7     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 2. Flow Diagrams             │  18      │  20      │ ⚠️         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  7/7     │          │            │
│    Decision + error branches │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 3. Functional Specs          │  19      │  20      │ ⚠️         │
│    Tables complete           │  7/7     │          │            │
│    Field descriptions clear  │  7/7     │          │            │
│    Validation rules explicit │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. User Stories              │  20      │  20      │ ✅         │
│    Coverage per user type    │  7/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story              │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Scope Clarity             │  17      │  20      │ ⚠️         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  94      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md:196 | Vague language: "`linkageMuMap` unbounded growth: Add cleanup or use `sync.Map`" — two alternatives with no decision rule for choosing between them | -1 pt (from Func Specs validation rules) |
| prd-spec.md:169 vs :261 | Inconsistency: Phase 1 AC states "No `as any` type escapes remain in the codebase" but Phase 4 frontend table lists "`as any` type escapes: Multiple instances: Fix all by extending TypeScript types" — contradictory phase ownership | -3 pts (from Scope Clarity consistency) |

---

## Attack Points

### Attack 1: Scope Clarity — `as any` removal is double-booked across Phase 1 and Phase 4

**Where**: prd-spec.md line 169 (Phase 1 AC: "No `as any` type escapes remain in the codebase") vs line 261 (Phase 4 frontend: "`as any` type escapes | Multiple instances | Fix all by extending TypeScript types to match API responses")

**Why it's weak**: Phase 1 acceptance criteria explicitly claims it will eliminate ALL `as any` escapes. If Phase 1 succeeds, there are zero `as any` instances left. Yet Phase 4 lists `as any` removal as a deliverable. This is a direct contradiction: either Phase 1 removes them all (making Phase 4's item redundant), or Phase 1 does not remove them all (making its AC dishonest). A developer cannot determine which phase owns which `as any` instances. The in-scope item "Type safety fixes: remove all `as any` casts, fix `user: any` to `user: User`" also does not specify a phase boundary for this split.

**What must improve**: Either (a) scope Phase 1's `as any` removal to specific categories (e.g., "remove `as any` in API response handling and store types") and Phase 4 to the remaining cases, or (b) move all `as any` removal to a single phase and remove the duplicated entry. The acceptance criteria for each phase must be independently verifiable without contradiction.

### Attack 2: Flow Diagrams — decision branches are still only rework loops with no escalation

**Where**: prd-spec.md lines 89-121 (Mermaid diagram) — all four "No" branches: "Fix errors", "Optimize queries", "Extract remaining" (x2)

**Why it's weak**: Every "No" branch in the diagram is an optimistic rework loop: fail gate check -> fix -> recheck. There are no branches for: what happens if a phase fails repeatedly (rollback? escalate? defer?), what happens if the p95 target cannot be met after multiple optimization attempts, or what happens if Phase 1 frontend dead code removal breaks an undocumented dependency. The iteration 1 report flagged this exact issue and it remains unchanged. For a multi-phase cleanup that takes 2-3 weeks, the absence of any failure-mode handling is a real gap — the team has no documented plan for when things go wrong.

**What must improve**: Add at least one escalation branch, e.g., after P2Test "No" → a decision node: "Attempts > 3?" → Yes → "Escalate: defer optimization, document as known issue" → Done. Or add a note that rework loops have a maximum iteration count before the item is deferred. This does not require restructuring the entire diagram — just adding one realistic failure path.

### Attack 3: Functional Specs — `linkageMuMap` fix remains a vague choice without criteria

**Where**: prd-spec.md line 196: "`linkageMuMap` unbounded growth | Add cleanup or use `sync.Map`"

**Why it's weak**: This item presents two implementation alternatives ("Add cleanup or use `sync.Map`") with no decision criteria. The iteration 1 attack 2 specifically called this out: "For `linkageMuMap`, pick one approach and justify it, or specify the decision criteria." The composite index item on the same table was fixed with concrete criteria ("if `EXPLAIN` shows a full table scan and benchmark shows >20ms improvement"), but this item was left unchanged. Without criteria, one developer might add manual cleanup, another might switch to `sync.Map`, and neither can be held accountable during acceptance because both satisfy the vague "fix it" requirement.

**What must improve**: Either (a) pick one approach: "Replace `linkageMuMap` with `sync.Map` for concurrent-safe access without manual cleanup" or (b) add criteria: "Use `sync.Map` if `go test -race` shows data races on `linkageMuMap`; otherwise add TTL-based cleanup with a background goroutine."

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Flow diagram missing parallelization | Partially | The diagram already showed a fork from Start to P1B and P1F, but lacks swim-lane notation or synchronization points. The text at lines 79-86 describes parallelization in detail. However, no escalation/failure branches were added. The core critique (no error handling beyond rework loops) remains. |
| Attack 2: Vague conditional — composite index criteria | Yes | Line 195 now reads: "Add composite index on `(team_id, main_item_id)` if `EXPLAIN` shows a full table scan and benchmark shows >20ms improvement on a 200-record dataset" — concrete and measurable. However, the `linkageMuMap` item on line 196 was not fixed. |
| Attack 3: Missing frontend developer story | Yes | Story 4 "Navigable Frontend After Restructuring" was added to prd-user-stories.md with 4 Given/When/Then ACs covering `StatusTransitionDropdown`, `formatDate`, file line counts, and dialog extraction. |

---

## Verdict

- **Score**: 94/100
- **Target**: 90/100
- **Gap**: 0 points (target met, exceeded by 4)
- **Action**: Target reached. Score of 94 exceeds the 90-point threshold. Two of three iteration-1 attacks were fully or partially addressed; the remaining issues (flow diagram error branches, `linkageMuMap` vagueness) are minor. The newly discovered Phase 1/Phase 4 `as any` inconsistency is the most impactful item to resolve if further refinement is desired.
