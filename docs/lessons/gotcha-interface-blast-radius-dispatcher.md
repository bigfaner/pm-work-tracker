---
created: "2026-06-04"
tags: [testing, architecture, interface]
---

# Interface Method Addition Causes Large Blast Radius in Task Execution

## Problem

Task dispatcher agent appeared "stuck" when executing a task that added new methods to shared Go interfaces (`MainItemRepo`, `MainItemService`). The agent was interrupted by the user twice. VS Code diagnostics showed 20+ compilation errors across many test files during intermediate states.

## Root Cause

1. **Surface cause**: Agent was working slowly through many files, user saw persistent compilation errors and interrupted.
2. **Direct cause**: Adding `SoftDelete`, `CascadeSoftDelete`, `Delete` to shared interfaces required updating ALL mock/stub implementations across 10+ test files (handler tests, service tests, constructor tests, permission matrix tests).
3. **Structural cause**: Go interfaces are implicitly satisfied — adding a method breaks every mock/stub at compile time. There is no incremental way to add methods; all implementors must be updated simultaneously. The agent's file-by-file approach creates a window where diagnostics show errors in files not yet updated.

## Solution

The agent's approach was actually correct — it was making progress through all affected files. The issue was a false perception of "stuck" caused by:
- VS Code diagnostics updating in real-time as files are saved
- Each intermediate file save triggering errors in other files that haven't been updated yet
- The user seeing accumulated errors and interpreting them as the agent failing

## Reusable Pattern

When a task requires adding methods to shared interfaces:
1. **Expect large blast radius** — count all mock/stub implementations before estimating effort
2. **Don't interrupt based on intermediate diagnostics** — the agent needs to update all implementors sequentially; errors between steps are expected
3. **Consider task splitting** — if an interface change touches 5+ files, the task may be too large for a single agent run and should be pre-split into: (a) interface + production implementation, (b) test stub updates, (c) new feature tests
4. **Trust `go build ./...` over VS Code diagnostics** — run a full build to check actual state rather than relying on LSP diagnostics that refresh per-file

## Related Files

- `backend/internal/repository/main_item_repo.go` — interface definition
- `backend/internal/service/main_item_service.go` — interface definition
- `backend/internal/handler/router_test_stubs.go` — centralized stubs
