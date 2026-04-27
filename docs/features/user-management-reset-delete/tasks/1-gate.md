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

Exit verification gate for Phase 1. Confirms that error types, DTOs, repository methods, and auth checks are complete and consistent before the API layer (Phase 2) begins.

## Verification Checklist

1. [ ] All new error constants compile and are mapped to correct HTTP status codes
2. [ ] `ResetPasswordReq` and `ResetPasswordResp` DTOs match `design/tech-design.md` Data Models section
3. [ ] `SoftDelete` repository method compiles and tests pass
4. [ ] `NotDeleted` scope applied to `ListFiltered` and `FindByBizKey`
5. [ ] Login rejects deleted users (unit test passes)
6. [ ] Auth middleware rejects deleted users' JWTs (unit test passes)
7. [ ] Project builds successfully (`cd backend && go build ./...`)
8. [ ] All existing tests pass
9. [ ] No deviations from design spec (or deviations are documented as decisions)

## Reference Files

- `docs/features/user-management-reset-delete/design/tech-design.md` — Cross-Layer Data Map, Error Handling
- `docs/features/user-management-reset-delete/design/api-handbook.md` — Error Codes, Data Contracts
- Phase 1 task records: `records/1.*.md`
- Phase 1 summary: `records/1-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence
