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

Exit verification gate for Phase 2. Confirms all 4 incompatibility points are fixed, all tests pass, lint check is active, and the success criteria from the design are met.

## Verification Checklist

1. [ ] All modified files compile without errors: `go build ./cmd/server/`
2. [ ] Dialect module tests pass: `go test ./internal/pkg/dbutil/ -count=1`
3. [ ] Repo tests pass: `go test ./internal/repository/gorm/ -count=1`
4. [ ] Migration tests pass: `go test ./internal/migration/ -count=1`
5. [ ] All existing tests pass: `cd backend && go test ./...`
6. [ ] No deviations from design spec (or deviations are documented as decisions)

## Success Criteria

前提条件：`config.yaml` 配置的是 MySQL。

- [ ] 上述单元测试全部通过
- [ ] `frontend/__tests__/e2e/` 内测试通过率 ≥ 95%

## Reference Files

- `docs/features/db-dialect-compat/design/tech-design.md` — 成功标准 section
- Phase 2 task records: `docs/features/db-dialect-compat/tasks/records/2.*.md`
- Phase 2 summary: `docs/features/db-dialect-compat/tasks/records/2-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Success criteria verified (unit tests + e2e pass rate)
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.

For the e2e success criterion:
1. Configure `config.yaml` with MySQL connection
2. Start the backend server
3. Run `cd frontend && npx vitest run --config __tests__/e2e/...` or equivalent
4. Record pass rate in the task record

If e2e pass rate < 95%, document failures and set status to `blocked`.
