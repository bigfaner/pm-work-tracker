# Subagent Partial Commits: Implementation Code Left Uncommitted

## Problem

After a full `/run-tasks` session completes, 75+ implementation files remain uncommitted. `git status` shows widespread changes across models, handlers, services, repos — but `git log` shows commits only touching `docs/` records and a handful of files.

## Root Cause

`forge:task-executor` subagents make code changes, then call `task record` CLI to create the execution record. The `task record` CLI only stages and commits the record file (`docs/features/.../records/*.md`) and updates `index.json`. The subagent treats this as "done" and exits — leaving all implementation code changes unstaged.

Causal chain:
1. Subagent writes code changes across many files
2. Subagent calls `task record --data process/record.json`
3. CLI commits only the record file + index.json
4. Subagent sees "commit done" and stops
5. All implementation files remain in working tree, never staged

## Solution

In the task executor prompt (or the task file itself), explicitly require the subagent to commit implementation changes **before** calling `task record`:

```bash
# Stage all implementation changes first
git add backend/... frontend/...
git commit -m "feat(...): implement task X.Y"

# Then record the task
task record X.Y --data docs/features/.../tasks/process/record.json
```

Or: after `task record`, run a final `git add -p` / `git status` check and commit any remaining changes.

## Key Takeaway

`task record` CLI only commits the record file — it does NOT commit implementation code. Subagents must explicitly stage and commit their code changes separately, before or after calling `task record`. Add this as an explicit step in task executor prompts: "commit all changed implementation files, then record the task."
