---
id: "disc-2"
title: "Fix: e2e test scripts use wrong cwd 'ui' instead of 'frontend'"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix: e2e test scripts use wrong cwd 'ui' instead of 'frontend'

## Root Cause

TC-001 and TC-002 fail because runCli('npx tsc --noEmit', 'ui') passes 'ui' as cwd but the frontend directory is 'frontend'. Fix: change 'ui' to 'frontend' in ui.spec.ts lines 8 and 15.

## Reference Files

- Source: tests/e2e/features/unify-permission-checks/ui.spec.ts
- Test script: tests/e2e/features/unify-permission-checks/ui.spec.ts
- Test results: tests/e2e/features/unify-permission-checks/results/latest.md

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task T-test-3 is automatically restored to pending if all its dependencies are completed.
