---
date: "2026-04-30"
doc_dir: "docs/proposals/e2e-test-scripts-rebuild/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 88/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  17      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   6/7    │          │            │
│    Urgency justified         │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  17      │  20      │ ⚠️          │
│    Approach concrete         │   6/7    │          │            │
│    User-facing behavior      │   6/7    │          │            │
│    Differentiated            │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  13      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  14      │  15      │ ⚠️          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  13      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   4/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  14      │  15      │ ⚠️          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   4/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  88      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Problem §Evidence | "scripts may not conform to current forge gen-test-scripts standards" — conformance gap is still speculative; no audit performed to confirm which scripts are non-conformant | -1 pt (evidence) |
| Problem §Urgency | Broken scripts documented but no cost-of-delay argument — why act now vs. next sprint? The scripts being broken is a static fact that was also true last month | -1 pt (urgency) |
| Solution §Approach | "in batches" appears in Next Steps but is never defined — batch size, sequencing, and gate criteria are absent | -1 pt (approach concrete) |
| Solution §User-facing | "approximately 11–22 additional spec files" — 2x range is unexplained; which of the 11 features have CLI coverage is never stated, making the count knowable but left unresolved | -1 pt (user-facing behavior) |
| Alternatives §Pros/cons | "Graduate as-is" cons include "others may have stale locators" — speculative, no evidence; proposed approach pros acknowledge no non-effort downside (e.g., risk that /gen-test-scripts produces worse output than manually-authored scripts) | -1 pt (honest trade-offs) |
| Alternatives §Rationale | "22 hrs of rebuild vs. ongoing manual verification" — the ongoing manual verification cost is asserted but never quantified; ROI denominator is missing | -1 pt (rationale) |
| Scope §Bounded | Effort estimate (~22 hrs) appears only in the alternatives table, not the scope section; no timeline stated | -1 pt (bounded) |
| Risk §Likelihood | Risk 3 and Risk 4 carry bare "Low" with no qualifying evidence; likelihood scale is still undefined across the table | -1 pt (rating honesty) |
| Risk §Mitigation | Risk 3 mitigation: "`/graduate-tests` agent handles deduplication via split/merge logic" — delegates entirely to the tool with no fallback if the agent fails or produces incorrect deduplication | -1 pt (actionability) |
| Success §Coverage | No criterion verifying sitemap.json was updated before script generation — it is explicitly in-scope but absent from success criteria | -1 pt (coverage) |

---

## Attack Points

### Attack 1: Problem Definition — urgency proves scripts are broken but not why to act now

**Where**: "Running these scripts against the current backend fails immediately on any assertion that uses a numeric ID as a URL path segment (e.g., `rbac-permissions` has 10 such references, `soft-delete-consistency` has 19). These scripts are not just unrun — they are broken against the current API and will produce false failures if graduated as-is."

**Why it's weak**: The proposal proves the scripts are broken, which is a genuine improvement over iteration 1. But it does not answer the decision-relevant question: why fix this now rather than next sprint or next quarter? The scripts being broken is a static fact — they were equally broken last month. What changes if this is deferred? Is there a CI integration planned that would surface these failures? A release requiring regression coverage? The urgency section establishes severity but not time-sensitivity.

**What must improve**: Add one sentence explaining the trigger or cost of delay. For example: "The team is planning to enable CI regression runs in sprint N; graduating broken scripts before that date prevents false-failure noise from day one." Or: "Each additional sprint of backend changes increases the diff between test-cases.md and the current API, making regeneration more expensive — the 7 broken scripts already require fixing 10–19 assertions each." Either framing makes the now-vs-later case.

---

### Attack 2: Solution Clarity — spec file count is knowable but left as a 2x range

**Where**: "The 11 features are expected to contribute approximately 11–22 additional spec files (each feature has at least `api.spec.ts`; features with CLI coverage add `cli.spec.ts`)."

**Why it's weak**: The range 11–22 is a 2x spread. The explanation given implies the answer is deterministic right now — just count which of the 11 features have CLI test cases in their test-cases.md. Instead the proposal leaves it as a range, which means the developer setting up CI cannot verify the end state or plan for it. The success criterion says "at least one graduated spec file per feature" — if the actual output is 11 files, that criterion is met; if it is 22, the regression suite is twice as large. These are materially different outcomes.

**What must improve**: Audit the 11 features' test-cases.md files and state the actual count: "X of 11 features have CLI test cases, so the expected output is Y spec files." This takes 5 minutes and eliminates the ambiguity entirely. If the count is genuinely unknown until generation, say so explicitly and explain why.

---

### Attack 3: Risk Assessment — duplicate coverage mitigation has no fallback

**Where**: "Duplicate coverage with already-graduated tests causes noise | Low | Redundant test failures | `/graduate-tests` agent handles deduplication via split/merge logic"

**Why it's weak**: This mitigation delegates entirely to an automated agent with no human review step and no exit criterion. If the agent's split/merge logic fails — or produces incorrect deduplication — the proposal has no fallback. "Low" likelihood does not mean zero likelihood, and the impact (redundant test failures polluting the regression suite permanently) is non-trivial. The mitigation is indistinguishable from "the tool will handle it."

**What must improve**: Add a verification step after graduation: compare test IDs in newly graduated spec files against those already in `tests/e2e/` and flag any duplicates for manual review. Alternatively, define an acceptance criterion: "no test ID appears in more than one spec file across the full `tests/e2e/` directory after graduation." This converts a tool-trust claim into a verifiable check.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Urgency lacks concrete evidence — generic drift argument | ✅ | Added specific counts: "7 of the 11 stranded scripts reference resources by internal numeric `.id`... `rbac-permissions` has 10 such references, `soft-delete-consistency` has 19" |
| Attack 2: End state for developer never described | ✅ | Added: "`npm test` run from `tests/e2e/` executes the full regression suite... approximately 11–22 additional spec files... A clean run exits with code 0 and prints a per-file summary of passed/skipped/failed test counts" |
| Attack 3: Highest-impact risk mitigation not actionable | ✅ | Replaced vague "Review generated scripts" with: "run `/run-e2e-tests`; require ≥80% pass rate; failing assertions must be fixed or documented in `KNOWN_FAILURES.md`; author is responsible; no graduation proceeds without resolution" |
| Attack 4: Alternatives pros/cons are straw-man thin | ✅ | "Do nothing" pros now include "zero risk of introducing broken tests into the regression suite"; proposed approach cons now include "~2 hrs per feature × 11 features ≈ 22 hrs total effort" with explicit ROI framing |
| Attack 5: "Conforming to forge standards" not verifiable | ✅ | Replaced with checkable specifics: "using `node:test` + `node:assert`, Playwright for UI, `fetch` for API, `child_process` for CLI, and no imports from external test frameworks" |

All 5 iteration-1 attacks were addressed. The score improvement from 72 to 88 reflects genuine substantive changes.

---

## Verdict

- **Score**: 88/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Iteration 2 complete. All five iteration-1 attacks were addressed. Remaining weaknesses are narrow: urgency lacks a cost-of-delay argument, the spec file count range (11–22) is knowable but unresolved, the duplicate-coverage mitigation has no fallback, and the sitemap.json update is in-scope but absent from success criteria. The proposal is substantially stronger and ready for execution with minor clarifications.
