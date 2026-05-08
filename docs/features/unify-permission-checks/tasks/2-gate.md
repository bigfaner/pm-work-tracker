---
id: "2.gate"
title: "Phase 2 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
---

# 2.gate: Phase 2 Exit Gate

## Description

Exit verification gate for Phase 2 (Handler & Service Refactoring). Confirms all handler-level bypass removed, service signatures simplified, VO/DTO clean, and GetUserPermissions returns all 29 codes for SuperAdmin.

## Verification Checklist

1. [ ] Backend compiles: `cd backend && go build ./...`
2. [ ] All existing tests pass: `cd backend && go test ./...`
3. [ ] No `isPMOrSuperAdmin()` function exists — `grep -r "isPMOrSuperAdmin" backend/internal/` returns 0
4. [ ] No assignee ownership checks in sub_item_handler — no `AssigneeKey != callerBizKey` pattern
5. [ ] No PM BizKey substitution in team_handler — no `pmBizKey = team.PmKey` pattern
6. [ ] TeamService methods have no `pmBizKey`/`callerBizKey` parameters
7. [ ] No `IsSuperAdmin` field in any response DTO (`vo/`, `dto/`)
8. [ ] `GetUserPermissions` returns all 29 codes when `user.IsSuperAdmin == true`
9. [ ] ProgressService.Append uses `skipRegressionCheck bool` parameter
10. [ ] No deviations from design spec (or deviations are documented as decisions)

## Reference Files

- `design/tech-design.md` — Interfaces 3-5, Error Path Migration, Cross-Layer Data Map
- Phase 2 task records: `records/2.*.md`
- Phase 2 summary: `records/2-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
