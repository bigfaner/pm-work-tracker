---
id: "disc-4"
title: "Fix: CLI test cwd path and re-run e2e"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix: CLI test cwd path and re-run e2e

## Root Cause

CLI tests use 'frontend' as cwd but Playwright runs from tests/e2e/ so it resolves to tests/e2e/frontend/ which doesn't exist. Fix to use '../../frontend'. Also re-verify API tests pass after disc-3 fix.

## Reference Files

- Source: tests/e2e/features/unify-permission-checks/ui.spec.ts
- Test script: tests/e2e/features/unify-permission-checks/ui.spec.ts
- Test results: tests/e2e/features/unify-permission-checks/results/latest.md

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task T-test-3 is automatically restored to pending if all its dependencies are completed.
