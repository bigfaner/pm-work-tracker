---
id: "1.gate"
title: "Phase 1 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["1.summary"]
status: pending
breaking: true
---

# 1.gate: Phase 1 Exit Gate

## Description

Exit verification gate for Phase 1 (Middleware & Seed Foundation). Confirms that SuperAdmin loads all 29 codes via TeamScopeMiddleware, RequirePermission has no SuperAdmin bypass, and seed data has 29 codes for superadmin role.

## Verification Checklist

1. [ ] Backend compiles: `cd backend && go build ./...`
2. [ ] All existing tests pass: `cd backend && go test ./...`
3. [ ] No SuperAdmin bypass in `RequirePermission` — verify no `IsSuperAdmin` call in `middleware/permission.go`
4. [ ] TeamScopeMiddleware injects all 29 codes for SuperAdmin — verify `permissions.AllCodeStrings()` call in `middleware/team_scope.go`
5. [ ] `seedPresetRoles` seeds superadmin with all 29 codes — verify `permissions.AllCodeStrings()` call in `migration/rbac.go`
6. [ ] No deviations from design spec (or deviations are documented as decisions)

## Reference Files

- `design/tech-design.md` — Interface 1 (RequirePermission), Interface 2 (TeamScopeMiddleware), Interface 6 (seedPresetRoles)
- Phase 1 task records: `records/1.*.md`
- Phase 1 summary: `records/1-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
