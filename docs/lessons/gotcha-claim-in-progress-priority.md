---
created: "2026-06-04"
tags: [architecture, testing]
---

# forge task claim 优先恢复 in_progress 任务

## Problem

Dispatcher 在 fix-1 完成后期望领取 2.1 等 pending 任务，但 `forge task claim` 持续返回 `in_progress` 的 T-eval-journey（MAIN_SESSION 任务），阻塞了整个调度循环。所有 scorer subagent 因 API 529 限流失败，任务无法完成。

## Root Cause

1. **Surface cause**: `forge task claim` 总是返回 T-eval-journey，跳过 2.x pending 任务。
2. **Direct cause**: `forge task claim` 的优先级是 CONTINUE (in_progress) > CLAIMED (pending)。任何 `in_progress` 任务都会被优先恢复。
3. **Structural cause**: MAIN_SESSION 任务的执行依赖外部 subagent（scorer），当这些 subagent 因基础设施问题（API 限流）失败时，任务卡在 `in_progress` 无法推进，同时阻塞了整个 claim 队列。

## Solution

当 MAIN_SESSION 任务因外部原因无法完成时：
1. 使用 `forge task transition <id> suspended --reason "..."` 挂起任务
2. 完成其他 pending 任务后再恢复：`forge task transition <id> pending --reason "ready to retry"`
3. 或者对于非关键的 MAIN_SESSION 任务，直接 skip

## Reusable Pattern

Dispatcher 在收到 CONTINUE 时应检查：该任务是否真的可以推进？如果外部依赖不可用（API 限流、网络问题），应主动挂起而不是反复重试阻塞队列。

```
if ACTION == CONTINUE and MAIN_SESSION:
    check if execution is feasible (no rate limits, deps available)
    if not feasible:
        forge task transition <id> suspended --reason "external dependency unavailable"
        continue loop (claim next task)
```

## Related Files

- `docs/lessons/gotcha-continue-replay-fresh-agent.md` — related: CONTINUE dispatches fresh agent
