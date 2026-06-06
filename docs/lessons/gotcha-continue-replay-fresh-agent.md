---
created: "2026-06-04"
tags: [architecture, testing]
---

# CONTINUE Action Dispatches Fresh Agent That Re-Executes Completed Work

## Problem

Task dispatcher resumes an in-progress task via `ACTION: CONTINUE`, spawning a brand-new subagent. When the task's code changes were already completed in a previous (interrupted) run, the fresh agent has no memory of prior work. It re-executes the entire task from scratch, causing:
- 50+ minute hang on "Searching for 1 pattern" as the agent tries to understand existing state
- Redundant code changes that may conflict with already-completed work
- User frustration and wasted time

## Root Cause

1. **Surface cause**: Subagent stuck on grep/search for 50+ minutes during task re-execution.
2. **Direct cause**: Fresh subagent has zero context from the previous interrupted run. It reads the task file and starts implementation from scratch, trying to understand what code exists.
3. **Structural cause**: The `forge task claim` → `CONTINUE` flow does not distinguish between "task was interrupted mid-implementation" and "task code is already done, just needs record submission". Both cases dispatch a full execution agent.

## Solution

Before dispatching for CONTINUE, the dispatcher should check if the task's code is already complete:

1. Run `go test ./...` (or equivalent) to verify compilation + tests pass
2. If tests pass: submit the record directly or dispatch a minimal "submit-only" agent
3. If tests fail: dispatch the full execution agent as usual

This avoids re-executing completed work and eliminates the 50-minute search hang.

## Reusable Pattern

When the dispatcher receives `ACTION: CONTINUE`:
1. **Quick-verify first**: Run the test suite to check if work is already done
2. **Branch on result**: Pass → submit record directly. Fail → dispatch full agent
3. **Never assume the agent knows what happened before**: Each subagent invocation is stateless

This pattern applies to any task resume scenario where the previous run may have completed work but failed to submit its record.

## Related Files

- `docs/lessons/gotcha-interface-blast-radius-dispatcher.md` — related: interface changes cause apparent "stuck" behavior
