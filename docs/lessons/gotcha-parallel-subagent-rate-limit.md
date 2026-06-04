---
created: "2026-06-04"
tags: [local-dev-deployment, testing]
---

# Parallel Subagent Launch Triggers API Rate Limit

## Problem

Launched 7 scorer subagents simultaneously to evaluate 7 journey documents. All 7 hit the API at the same time, triggering 529 rate limit errors. 5 returned 529 errors, 2 were killed. Zero scoring work completed — complete waste of API quota and time.

## Root Cause

1. **Surface cause**: 7 concurrent API requests exceeded rate limit (529 errors).
2. **Direct cause**: All 7 subagents were launched in a single message with `run_in_background: true`, causing simultaneous API connections.
3. **Structural cause**: No consideration of API concurrency limits when parallelizing subagent work. The eval loop's "spawn in parallel" instruction was interpreted as "all at once" rather than "in controlled batches".

## Solution

Batch subagent launches to stay within API rate limits:

- **Max 2-3 concurrent subagents** for heavy operations (scoring, revising, searching)
- **Sequential batching**: launch batch 1 (2-3 agents), wait for completion, then launch batch 2
- For lightweight operations (file reads, simple lookups), higher concurrency is safe

## Reusable Pattern

When parallelizing subagent work:
1. **Estimate API load** — each subagent makes multiple tool calls (reads, writes, searches)
2. **Batch size = min(3, total_tasks)** — never launch more than 3 heavy subagents at once
3. **Wait for batch completion** before launching the next batch
4. **Monitor for 529 errors** — if any batch hits rate limits, reduce batch size for next round

Pattern for N tasks with batch size B:
```
for i in range(0, N, B):
    launch agents for tasks[i:i+B] in parallel
    wait for all to complete
    process results
```

## Related Files

- `docs/lessons/gotcha-continue-replay-fresh-agent.md` — related: dispatcher subagent issues
