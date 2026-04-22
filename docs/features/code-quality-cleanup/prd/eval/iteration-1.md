---
date: "2026-04-22"
doc_dir: "docs/features/code-quality-cleanup/prd"
iteration: "1"
target: "90"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

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
│ 3. Functional Specs          │  18      │  20      │ ⚠️         │
│    Tables complete           │  7/7     │          │            │
│    Field descriptions clear  │  7/7     │          │            │
│    Validation rules explicit │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. User Stories              │  19      │  20      │ ⚠️         │
│    Coverage per user type    │  6/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story              │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Scope Clarity             │  19      │  20      │ ⚠️         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  94      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md:185 | Vague language: "Add composite index on `(team_id, main_item_id)` if analysis supports" — no criteria given for when to proceed vs skip | -2 pts (from Func Specs) |

---

## Attack Points

### Attack 1: Flow Diagrams — missing parallelization and true error handling

**Where**: prd-spec.md lines 89-111 (flow diagram) vs lines 79-86 (parallelization text)

**Why it's weak**: The text explicitly describes concurrent execution: *"Backend dev leads Phase 1 backend -> Phase 2 -> Phase 3 backend -> Phase 4 backend. Frontend dev starts Phase 1 frontend concurrently, then Phase 3 frontend -> Phase 4 frontend."* Yet the Mermaid diagram shows only a single-threaded sequential flow with no parallel branches. Additionally, the four "No" branches are all optimistic rework loops (fix -> recheck). There are no true error/exception branches — no handling for scenarios like "tests fail after merge," "performance regression in unrelated endpoint," or "scope creep identified." The diagram presents the happy path only with gate checks, not a realistic execution flow with failure modes.

**What must improve**: Either (a) add a parallel swim-lane diagram showing backend and frontend workstreams with synchronization points, or (b) add error branches that address what happens when a phase gate fails repeatedly (rollback? skip? escalate?). The text claims concurrent execution is possible but the diagram contradicts this.

### Attack 2: Functional Specs — vague conditional without decision criteria

**Where**: prd-spec.md line 185: "Add composite index on `(team_id, main_item_id)` if analysis supports"

**Why it's weak**: The phrase "if analysis supports" has no defined success criteria. What analysis? What metric determines "supports"? Is it query plan analysis showing a full table scan? Is it a benchmark showing >50ms improvement? Without criteria, this item is unimplementable — one developer might add the index, another might skip it, and neither can be held accountable during acceptance. Similarly, line 186 ("`linkageMuMap` unbounded growth: Add cleanup or use `sync.Map`") presents two alternatives with no decision rule for choosing between them.

**What must improve**: Replace "if analysis supports" with a concrete condition, e.g., "Add composite index on `(team_id, main_item_id)` if `EXPLAIN` shows full table scan on sub_items during list queries (benchmark: >20ms improvement on 200-record dataset)." For `linkageMuMap`, pick one approach and justify it, or specify the decision criteria (e.g., "use `sync.Map` if concurrent access patterns are confirmed by race detector").

### Attack 3: User Stories — incomplete coverage of frontend developer persona

**Where**: prd-user-stories.md — only 3 stories total

**Why it's weak**: The background section identifies "Development team (primary): existing and onboarding developers" as users. Story 2 covers an onboarding developer reading code. Story 3 covers a backend developer adding features. But no story covers an existing frontend developer modifying the codebase — the PRD includes significant frontend refactoring (Phase 3 frontend dedup: `StatusTransitionDropdown`, `useMemberName`, `MemberSelect`; Phase 4 frontend file splits of `ItemViewPage.tsx` and `MainItemDetailPage.tsx`). The frontend developer persona is implicitly addressed only through Story 2's reading-focused AC, not through a modification-focused story. This matters because the frontend refactoring (file splits, component extraction, React Query migration) is qualitatively different from the backend work and has distinct acceptance criteria.

**What must improve**: Add a Story 4 for the frontend developer persona: "As a frontend developer, I want page components decomposed into focused files with shared hooks, so that I can locate and modify specific UI behaviors without scrolling through 1000+ line files." This would cover Phase 3 frontend dedup and Phase 4 file splits with appropriate ACs (e.g., "Given I need to modify the status transition UI, when I search for it, then I find one `StatusTransitionDropdown` component instead of 3 copies").

---

## Verdict

- **Score**: 94/100
- **Target**: 90/100
- **Gap**: 0 points (target met, exceeded by 4)
- **Action**: Target reached. Score of 94 exceeds the 90-point threshold.
