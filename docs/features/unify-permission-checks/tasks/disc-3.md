---
id: "disc-3"
title: "Fix: RequirePermission middleware missing SuperAdmin bypass for non-team routes"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix: RequirePermission middleware missing SuperAdmin bypass for non-team routes

## Root Cause

TC-004 fails because setupRbacFixtures() calls GET /v1/admin/roles which requires user:manage_role. The RequirePermission middleware for non-team-scoped routes (/admin/*) only checks roleRepo.HasPermission() without checking isSuperAdmin flag. TeamScopeMiddleware has the SuperAdmin bypass but it only runs for /teams/:teamId/* routes. Fix: add isSuperAdmin check to RequirePermission before querying DB, similar to how TeamScopeMiddleware handles it.

## Reference Files

- Source: backend/internal/middleware/permission.go
- Test script: tests/e2e/features/unify-permission-checks/api.spec.ts
- Test results: tests/e2e/features/unify-permission-checks/results/latest.md

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task T-test-3 is automatically restored to pending if all its dependencies are completed.
