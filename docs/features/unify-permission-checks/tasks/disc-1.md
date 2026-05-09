---
id: "disc-1"
title: "Fix: TableViewPage date styling test (pre-existing)"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix: TableViewPage date styling test (pre-existing)

## Root Cause

Pre-existing test failure: expected-date-4 should not have text-error class. Fails on both current and stashed state — unrelated to Phase 1 backend changes.

## Reference Files

- Source: frontend/src/pages/TableViewPage.test.tsx,frontend/src/pages/TableViewPage.tsx
- Test script: frontend/src/pages/TableViewPage.test.tsx
- Test results: console output

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task 1.gate is automatically restored to pending if all its dependencies are completed.
