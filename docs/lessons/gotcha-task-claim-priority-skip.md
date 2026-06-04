---
created: "2026-06-04"
tags: [architecture, testing]
---

# fix-task Completion Does Not Auto-Unblock Source Task in Time

## Problem

After fix-1 completed, `forge task claim` selected T-test-gen-journeys instead of the now-unblocked task 2.1. All Phase 2 coding tasks (2.1–2.8) were skipped.

## Root Cause

1. **Surface cause**: Claim picked T-test-gen-journeys over 2.1 after fix-1 completed.
2. **Direct cause**: Task 2.1 was still in "blocked" status when the claim ran. The fix-1 file promises "source task 2.1 is automatically restored to pending", but the auto-unblock did not take effect before the next claim.
3. **Structural cause**: The dispatcher's `/run-tasks` loop runs `forge task claim` immediately after a subagent returns. If the fix-task's submit doesn't synchronously unblock the source task (or if the unblock is deferred), the next claim sees the source task as still blocked and skips it.

## Solution

The dispatcher must explicitly check and restore blocked source tasks after a fix-task completes:
1. After fix-task agent returns, run `forge task status <source-task-id>`
2. If status is still "blocked", manually restore it: `forge task transition <source-task-id> pending --reason "fix-task completed"`
3. Then run the normal claim loop

This prevents coding tasks from being starved by test pipeline tasks in the claim queue.

## Reusable Pattern

When a fix-task completes and the dispatcher expects the source task to be unblocked:
- Don't assume auto-unblock happened — verify explicitly
- If still blocked, manually transition to pending before claiming
- The dispatcher should track source-task-id from fix-task creation and verify after completion

## Related Files

- `docs/lessons/gotcha-continue-replay-fresh-agent.md` — related: CONTINUE dispatch issue
