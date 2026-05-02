# Piped output buffering hides E2E test progress for 50+ minutes

## Problem

Running `npx playwright test` (373 tests) appeared to hang for 50+ minutes with no output, repeatedly across multiple invocations. Each time I killed the process, restarted, and it appeared to hang again.

## Root Cause

Causal chain (4 levels deep):

1. **Symptom**: `npx playwright test --reporter=line 2>&1 | tail -40` shows no output for 50+ minutes
2. **Direct cause**: `tail -40` buffers ALL stdout until the upstream process completes. No lines are shown until Playwright finishes all 373 tests
3. **Root cause**: I consistently piped test output through `tail -N` or `tee file | tail -N`, which is a Unix buffering pattern that defeats Playwright's streaming `--reporter=line` output
4. **Trigger condition**: Using `tail -N` as the final pipe stage for any long-running streaming process

This was compounded by the original `agent-browser` issue (133 tests × 60s timeout each), but even after fixing that, the `tail` buffering made the 15-minute normal run appear frozen.

## Solution

**For running test suites, never pipe through `tail` or `head`.** Instead:

```bash
# BAD — buffers all output until completion
npx playwright test --reporter=line 2>&1 | tail -40

# GOOD — shows streaming output in real-time
npx playwright test --reporter=line 2>&1

# GOOD — save to file without buffering
npx playwright test --reporter=line 2>&1 | tee /tmp/results.log

# GOOD — filter only summary line
npx playwright test --reporter=line 2>&1 | grep -E "passed|failed|skipped"
```

If you need to limit output in Claude Code's Bash tool, use `grep` (streaming filter) instead of `tail`/`head` (buffering filter).

## Key Takeaway

`tail -N` is a **buffering** pipe — it must read the entire stream before showing the last N lines. For long-running processes with streaming output (test runners, build tools, servers), use `grep` or `tee` instead. This applies to any `tail`/`head` usage on streaming process output.
