---
created: "2026-06-11"
tags: [architecture, testing]
---

# Task dependency declared in prose but not frontmatter causes 16-hour agent hang

## Problem

The `/run-tasks` dispatcher dispatched fix-6 ("E2E: item-milestone-binding 6 tests still failing") while fix-5 ("E2E: item-milestone-binding baseUrl import missing") was still blocked. The fix-6 agent ran for 16 hours in an unbounded retry loop — tests couldn't pass without the `baseUrl` import that fix-5 was supposed to add.

## Root Cause

Causal chain (5 levels):

1. **Agent ran 16 hours**: The fix-6 agent kept retrying tests that couldn't pass without `baseUrl` import, entering an unbounded edit→run→fail→edit loop.
2. **Prerequisite not met**: fix-5 (add `baseUrl` import) was blocked/incomplete when fix-6 was dispatched. fix-6's tests all use `${baseUrl}` in `page.goto()` calls.
3. **Dependencies field empty**: Both fix-5 and fix-6 had `dependencies: []` in their frontmatter. fix-6's *description text* says "After baseUrl import fix, 6 tests still fail" — but the dispatcher only reads the `dependencies` field, not prose.
4. **Dispatcher's 30-minute timeout not enforced**: The `/run-tasks` skill specifies "30-minute timeout per task" but relies on the Agent tool's default behavior, which has no timeout enforcement for long-running agents.
5. **No bounded retry in task executor**: The forge:task-executor has no built-in limit on how many fix→verify cycles it attempts before giving up.

## Solution

1. **Always declare dependencies in frontmatter `dependencies` field** — the dispatcher reads this field exclusively. Prose descriptions mentioning prerequisites are invisible to the claim logic.
2. **Enforce agent timeout**: The dispatcher must set a hard timeout (30 min) when dispatching agents. If the agent doesn't return within the window, terminate and create a fix task.
3. **Add bounded retry to task executor**: Cap fix→verify cycles at 3 attempts. After 3 failures, report what was tried and exit — let the dispatcher create a new fix task.

## Reusable Pattern

When creating forge tasks with `forge task add`:

- **Use `--block-source`** to declare that a fix task blocks the source task
- **Use explicit dependency declarations** when a task depends on another fix completing first
- **Never rely on prose descriptions** for dependency information — the dispatcher is a machine reader

When writing the dispatcher or task executor:

- **Enforce a hard wall-clock timeout** per agent dispatch (30 min)
- **Cap retry loops** — 3 consecutive failures → STOP (already in dispatcher spec, but task executor also needs its own bound)

## Example

```yaml
# WRONG: dependency only in prose
---
id: "fix-6"
title: "E2E: 6 tests still failing"
dependencies: []
---
# After baseUrl import fix, 6 tests still fail...

# CORRECT: dependency in frontmatter
---
id: "fix-6"
title: "E2E: 6 tests still failing"
dependencies: ["fix-5"]
---
# 6 tests still fail after fix-5 adds baseUrl import...
```

## Related Files

- `docs/features/milestone-map/tasks/fix-5.md` (blocked, never completed)
- `docs/features/milestone-map/tasks/fix-6.md` (16-hour hang)
