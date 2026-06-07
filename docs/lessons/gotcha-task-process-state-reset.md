---
created: "2026-06-07"
tags: [local-dev-deployment, architecture]
---

# Resetting Task State Requires Clearing process/state.json, Not Just index.json

## Problem

When manually resetting a task's status (e.g., from `in_progress` back to `pending`), editing `index.json` alone is insufficient. `forge task claim` continues to report data integrity errors:

```
ERROR: Task 'Generate Test Journeys' has unexpected status: pending
```

## Root Cause

1. **Surface cause**: `forge task claim` rejects claim after index.json was updated to reset task status.
2. **Direct cause**: Forge stores claimed task state in TWO places:
   - `tasks/index.json` — task status field (`pending`, `in_progress`, `completed`, etc.)
   - `tasks/process/state.json` — currently claimed task's full snapshot including `startedTime`
3. **Structural cause**: When `forge task claim` runs, it cross-checks both files. If `process/state.json` says a task is started (has `startedTime`) but `index.json` says it's `pending`, the CLI detects a conflict and refuses to proceed. The `forge cleanup` command does NOT clear `process/state.json`.

## Solution

When manually resetting task state, delete BOTH:
1. Reset status in `index.json` (or use `forge task transition`)
2. Delete `docs/features/<slug>/tasks/process/state.json`

Only then will `forge task claim` work correctly.

## Reusable Pattern

When needing to reset/unblock a task that was claimed:
1. `forge task transition <id> pending --reason "..."`
2. Delete `docs/features/<slug>/tasks/process/state.json`
3. Now `forge task claim` will work

**File location**: `docs/features/<feature-slug>/tasks/process/state.json` — NOT in `.forge/state.json` (which is the worktree-level feature selector, not task claim state).

## Example

```bash
# Reset a claimed task
forge task transition T-test-gen-journeys pending --reason "resetting"
rm docs/features/milestone-map/tasks/process/state.json
forge task claim  # now works
```

## Related Files

- `docs/lessons/gotcha-claim-feature-vs-pipeline.md` — related: claim picks pipeline over feature tasks
- `docs/lessons/gotcha-task-claim-priority-skip.md` — related: fix-task auto-unblock timing
