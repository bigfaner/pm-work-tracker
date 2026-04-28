---
id: "2.gate"
title: "Phase 2 Exit Gate: Full Build + Test Pass + Zero Casts"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
---

# 2.gate: Phase 2 Exit Gate: Full Build + Test Pass + Zero Casts

## Description

Final exit gate. Confirms the entire refactor is complete, all tests pass, and no forbidden cast patterns remain.

## Verification Checklist

1. [ ] `go build ./...` passes with zero errors
2. [ ] `grep -rn "GetTeamID" backend/internal/` returns zero results
3. [ ] `grep -rn "uint(.*[Bb]iz[Kk]ey\|int64(teamID)" backend/internal/service/ backend/internal/handler/` returns zero results
4. [ ] `progress_service.go`: `TeamKey` field assigned from `teamBizKey` (not `int64(teamID)`)
5. [ ] `go test -race ./internal/...` passes with zero failures
6. [ ] Coverage ≥ 75% on `internal/service` and `internal/handler` (`go test -race -coverprofile=coverage.out ./internal/...`)
7. [ ] `team_service.go`: `isPMRole` accepts `int64`; `UpdateMemberRole` last param is `int64`

## Reference Files

- `design/tech-design.md` — PRD Coverage Map section
- `records/2-summary.md` — Phase 2 summary

## Acceptance Criteria

- [ ] All 7 checklist items pass
- [ ] Any deviations from design are documented in the record
- [ ] Record created via `task record` with test evidence (testsPassed > 0, coverage ≥ 75)

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If grep finds remaining casts, fix them inline before marking this gate complete.
