---
created: "2026-06-07"
tags: [architecture, testing]
---

# forge task claim Does Not Prioritize Feature Tasks Over Pipeline Tasks

## Problem

`forge task claim` does not distinguish between feature coding tasks and test/evaluation pipeline tasks. When both are claimable (dependencies satisfied), claim picks based on internal ordering (likely key alphabetical), not task priority or phase.

**Observed instances:**
1. After fix-1 completed, claim selected T-test-gen-journeys instead of unblocked task 2.1.
2. With 2.4 (pending, all deps met) available, claim selected T-test-gen-journeys instead.
3. Result: test generation runs against incomplete features — generating tests for code that doesn't exist yet.

## Root Cause

1. **Surface cause**: Claim picks T-test-gen-journeys over available feature tasks (2.1, 2.4).
2. **Direct cause**: `forge task claim` treats all unblocked pending tasks equally — no priority ordering by task type, phase, or category. Feature tasks (2.1–2.9) and pipeline tasks (T-*) compete in the same queue.
3. **Structural cause**: The task index mixes two independent dependency chains:
   - Feature chain: 1.x → 2.x → 3.x → 4.x (sequential phases)
   - Pipeline chain: T-review-doc → T-test-gen-journeys → T-eval-journey → ... (parallel test workflow)

   These chains have no cross-dependencies, so pipeline tasks become claimable as soon as their own deps resolve — even if the feature is only 50% complete.

4. **Deeper issue**: When fix-1 blocks task 2.3, tasks 2.4+ remain claimable (2.4 doesn't depend on 2.3). But claim still skips them for pipeline tasks, suggesting the ordering issue is not just about blocking but about claim's internal sort order.

## Solution

**For dispatchers (/run-tasks):** Before claiming, check if feature coding tasks are still pending. If yes, prefer them over pipeline tasks:
1. Run `forge task list` to identify available feature tasks (type `coding.feature`, status `pending`, deps met).
2. If feature tasks exist, manually claim the lowest-ID one: identify the target task and dispatch directly.
3. Only fall through to pipeline tasks when no feature tasks remain.

**For task design:** Consider adding explicit dependencies from pipeline tasks to a phase gate (e.g., T-test-gen-journeys depends on 2.gate instead of just T-review-doc). This ensures test generation only starts after the feature's backend phase is complete.

## Reusable Pattern

When `forge task claim` selects a pipeline/test task over an available feature task:
- Don't blindly follow claim's choice — verify it makes sense contextually
- If feature tasks are still in-progress, manually prioritize them
- The dispatcher should implement a "feature-first" heuristic: claim feature tasks before pipeline tasks

## Related Files

- `docs/lessons/gotcha-task-claim-priority-skip.md` — related: fix-task auto-unblock timing
- `docs/lessons/gotcha-claim-in-progress-priority.md` — related: claim prioritizes in_progress over pending
- `docs/lessons/gotcha-continue-replay-fresh-agent.md` — related: CONTINUE dispatch issue
