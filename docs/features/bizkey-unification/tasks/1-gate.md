---
id: "1.gate"
title: "Phase 1 Exit Gate: Service Interfaces Compile Clean"
priority: "P0"
estimated_time: "1h"
dependencies: ["1.summary"]
status: pending
breaking: true
---

# 1.gate: Phase 1 Exit Gate: Service Interfaces Compile Clean

## Description

Exit verification gate for Phase 1. Confirms all service interface changes are complete, internally consistent, and the project builds before handler updates begin.

## Verification Checklist

1. [ ] `go build ./...` passes with zero errors
2. [ ] `GetTeamID` is removed; `GetTeamBizKey() int64` exists in `middleware/team_scope.go`
3. [ ] `grep -rn "GetTeamID" backend/internal/` returns zero results
4. [ ] `grep -rn "uint(.*bizKey\|int64(teamID)" backend/internal/service/` returns zero results
5. [ ] `ProgressRecord.TeamKey` assignment in `progress_service.go` uses `teamBizKey` (not `int64(teamID)`)
6. [ ] `go test -race ./internal/middleware/... ./internal/service/...` passes

## Reference Files

- `design/tech-design.md` — Complete change surface table (§7)
- `records/1-summary.md` — Phase 1 summary

## Acceptance Criteria

- [ ] All 6 checklist items pass
- [ ] Any deviations from design are documented in the record
- [ ] Record created via `task record` with test evidence (testsPassed > 0)

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If `go build` fails due to handler files referencing the removed `GetTeamID`, that is expected — handlers are updated in Phase 2. The gate only verifies the service layer is internally consistent.
