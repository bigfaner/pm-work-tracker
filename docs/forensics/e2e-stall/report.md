# Forensic Report: Agent Stalled During E2E Test Execution

**Date:** 2026-05-14
**Session:** `8e0e2b62-ffb2-4a98-89bc-8159d91add95`
**Feature:** milestone-map
**Skill:** `/run-tasks` dispatcher → `forge:task-executor` subagent

## Summary

Two consecutive task-executor agents stalled while executing T-test-3 (Run e2e Tests). Cross-referencing with earlier sessions revealed this is a **recurring issue**, not a one-time failure. The true root cause: **Playwright's stdout output floods the agent's stream, preventing "progress" signals (thinking/tool_use) from being emitted. The 600s stream watchdog sees no progress and kills the agent.** An earlier session's agent worked around this by redirecting output to a file and polling — confirming the diagnosis.

## Historical Evidence (This Problem Occurred Before)

Three T-test-3 agents across two sessions, same task, different outcomes:

| Session | Agent | Duration | Strategy | Result |
|---------|-------|----------|----------|--------|
| e68fd6ae (prev) | a6b4a38d | **28.1min** | 3× `npx playwright test` hit 300s timeout → switched to `> /tmp/log 2>&1` + `sleep` polling | **Completed** (workaround) |
| e68fd6ae (prev) | aa5c6f0 | 4.5min | Learned from prior attempt, used `--reporter=list` directly | **Completed** |
| 8e0e2b62 (current) | ad6db974 | **Stalled** | Used `just test-e2e` directly → output flooded stream → watchdog killed agent | **Failed** |

The 28.1-minute agent's top slowest actions reveal the struggle:

| Duration | Command |
|----------|---------|
| 300.4s | `npx playwright test features/milestone-map/ --reporter=json,list` (1st attempt, hit timeout) |
| 300.4s | `npx playwright test features/milestone-map/` (2nd attempt, hit timeout) |
| 300.4s | `E2E_FEATURE=1 npx playwright test ... \| tail -100` (3rd attempt, hit timeout) |
| 180.1s | `sleep 180 && tail -15 /tmp/e2e-test-output.log` (workaround: redirect to file + poll) |
| 120.1s | `sleep 120 && tail -50 /tmp/e2e-test-output.log` (polling) |
| 120.1s | `sleep 120 && wc -l /tmp/e2e-test-output.log` (polling) |

The agent eventually discovered that **redirecting Playwright output to a file** prevents stream flooding, and used `sleep` + `tail` to poll for completion.

## Root Cause Analysis

### The Real Problem: Playwright Output Floods Agent Stream

Claude Code's stream watchdog monitors for "progress signals" — thinking blocks and tool_use calls. When an agent runs a long Bash command, the command's stdout is streamed back. Playwright's `list` reporter outputs a line for every test start, pass, and fail. With 18 tests, this produces a continuous stream of text output.

**The stream watchdog appears to treat raw Bash output differently from thinking/tool_use signals.** When the Bash command is actively producing output but no thinking or tool_use events are emitted, the watchdog may still trigger because it only counts structured agent actions as "progress" — not raw tool output.

This explains why the earlier agent that redirected to a file succeeded: by removing Playwright output from the stream, the agent could emit progress signals (new Bash calls for `sleep && tail`) that kept the watchdog happy.

### Causal Chain (Revised)

1. **Symptom:** Agent killed by 600s stream watchdog while running `just test-e2e`
2. **Direct cause:** `just test-e2e` runs Playwright which produces continuous stdout over 10+ minutes. During this time, no new thinking or tool_use events are emitted. The stream watchdog sees 600s of no "progress" and kills the agent.
3. **Root cause (`pipeline-gap`):** The `test-e2e` recipe pipes Playwright output directly to stdout. In agent context, this floods the stream and starves the watchdog of progress signals. The recipe needs to buffer output for agent consumption.

### Why Playwright's 30s Timeout Is Irrelevant

Playwright's `timeout: 30_000` (per-test) works correctly — each test times out at 30s. But 18 tests × 30s = 540s total runtime. The issue isn't that tests hang; it's that the **total run time exceeds the watchdog's 600s window**, and during that entire window the agent emits zero progress signals because it's waiting for the Bash command to finish.

## Timeout Layer Analysis

| Layer | Mechanism | Value | Enforced? | What Happened |
|-------|-----------|-------|-----------|---------------|
| 1. Playwright test timeout | `timeout: 30_000` in config | 30s/test | **Yes** | Working correctly; not the issue |
| 2. Bash tool timeout | Agent set `timeout: 300000` | 5min | **No** | Process ran 1116s — timeout not enforced |
| 3. Stream watchdog | Claude Code runtime | 600s no-progress | **Yes** | Killed agent at 600s; orphaned the process |
| 4. test-e2e recipe | None | N/A | N/A | No output buffering, no global timeout |

The watchdog is doing its job — the agent genuinely produced no progress for 600s. The fix is to ensure the test-e2e recipe doesn't cause that situation.

## Fix

### test-e2e recipe: redirect output to file, cat after completion

When running in a non-TTY/agent context, Playwright output should be buffered to a file rather than streamed directly. This prevents stream flooding and allows the agent to retrieve results after the command completes.

**Before:**
```just
test-e2e feature="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ "{{feature}}" != "" ]; then
        cd tests/e2e && npx playwright test {{feature}}/
    else
        [ ! -d tests/e2e/node_modules ] && npm install --prefix tests/e2e
        cd tests/e2e && npx playwright test
    fi
```

**After:**
```just
test-e2e feature="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ "{{feature}}" != "" ]; then
        cd tests/e2e && npx playwright test {{feature}}/ > /tmp/e2e-output.log 2>&1; cat /tmp/e2e-output.log
    else
        [ ! -d tests/e2e/node_modules ] && npm install --prefix tests/e2e
        cd tests/e2e && npx playwright test > /tmp/e2e-output.log 2>&1; cat /tmp/e2e-output.log
    fi
```

By redirecting to a file during execution and only printing after completion, the Bash command's total output arrives as a single burst at the end. The stream watchdog won't see continuous output (which it ignores for progress counting), and the agent doesn't get killed mid-run.
