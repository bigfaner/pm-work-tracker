---
created: "2026-06-08"
tags: [testing, architecture]
---

# Eval Tasks Are Iterative — Don't Count Progress as Failure

## Problem

The `/run-tasks` dispatcher stopped after 3 "consecutive failures" during the journey eval pipeline (T-eval-journey). But the eval was actually making progress: iteration 1 scored 687 avg, iteration 2 scored 798 avg, with 2/7 journeys passing after fixes. The dispatcher's 3-strikes rule killed a converging improvement cycle.

## Root Cause

**Level 1 — Dispatcher stopped**: The `consecutive_failures` counter reached 3 because each blocked eval created a fix task, which incremented the counter.

**Level 2 — Blocked eval counted as failure**: T-eval-journey has ACs like "All Journeys scored >= 850/1000". When ACs aren't met, the task is `blocked`. The dispatcher treats `blocked → fix-task created` as a failure event.

**Level 3 — Iterative quality gate ≠ task failure**: The eval task is fundamentally an iterative quality improvement loop (score → identify gaps → fix → re-score). Each cycle improves quality but may not cross the threshold. The dispatcher's binary "completed or failed" model doesn't match this pattern. A task that goes from 687→798→820 is converging, not failing.

**Level 4 — Architecture mismatch**: The dispatcher's `consecutive_failures` counter was designed for coding tasks where "blocked" means a real blocker (compiler error, test failure). For eval tasks, "blocked" means "not yet good enough" — a different semantic that requires a different control flow.

## Solution

For eval/journey tasks, the eval loop should be treated as an internal iteration cycle, not as dispatcher-level failures. Options:

1. **Reset failure counter on progress**: If the eval score improves between iterations, reset `consecutive_failures` to 0 (the fix genuinely helped).
2. **Max-iterations gate**: Use a separate `max_eval_iterations` counter instead of `consecutive_failures` for eval tasks. Allow 3-5 eval iterations before stopping.
3. **MAIN_SESSION evals bypass failure counter**: Since MAIN_SESSION evals are orchestrated by the dispatcher itself, don't count them toward the failure limit. Only count subagent dispatch failures.

## Reusable Pattern

When a task type involves iterative quality improvement (eval, review, optimization), distinguish between:
- **Convergence failure** (score not improving) → count as failure
- **Progress toward threshold** (score improving but not yet passing) → do NOT count as failure

For the `/run-tasks` dispatcher specifically: eval tasks (type `eval.*`) should use a separate iteration limit rather than the generic `consecutive_failures` counter.

## Related Files

- `docs/features/milestone-map/testing/journeys/.eval-report.md` — eval score progression
- `docs/features/milestone-map/tasks/eval-journey.md` — task definition with ACs
