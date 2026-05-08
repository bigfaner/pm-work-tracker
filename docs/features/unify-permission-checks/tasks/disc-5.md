---
id: "disc-5"
title: "Fix: SuperAdmin bypass missing in RequirePermission middleware for admin routes"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix: SuperAdmin bypass missing in RequirePermission middleware for admin routes

## Root Cause

The RequirePermission middleware on /admin/* routes does not check isSuperAdmin flag. It only checks roleRepo.HasPermission(). When setupRbacFixtures() calls GET /v1/admin/roles with the admin (superadmin) token, it gets 403 because the admin role lacks user:manage_role permission. This causes cascade failure: TC-004 fails and TC-005..TC-039 are all skipped. Fix: add isSuperAdmin bypass check in RequirePermission middleware, OR grant user:manage_role to the admin role in seed data.

## Reference Files

- Source: backend/internal/middleware/permission.go
- Test script: tests/e2e/features/unify-permission-checks/api.spec.ts
- Test results: tests/e2e/features/unify-permission-checks/results/latest.md

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task T-test-3 is automatically restored to pending if all its dependencies are completed.
