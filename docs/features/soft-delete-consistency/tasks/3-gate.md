---
id: "3.gate"
title: "Phase 3 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["3.summary"]
status: pending
breaking: true
---

# 3.gate: Phase 3 Exit Gate

## Description

Final exit gate. Confirms SubItem SoftDelete, TeamMember RemoveMember, and schema migration are correct. Verifies end-to-end soft-delete flow.

## Verification Checklist

1. [ ] SubItem SoftDelete sets `deleted_flag=1` and `deleted_time`
2. [ ] TeamMember RemoveMember soft-deletes (record persists with `deleted_flag=1`)
3. [ ] Re-adding removed member succeeds (new row with `deleted_flag=0`)
4. [ ] Schema files are in sync (MySQL + SQLite both updated)
5. [ ] Project builds successfully
6. [ ] All tests pass (`cd backend && go test ./internal/...`)
7. [ ] SubItem re-create with same item_code after soft-delete succeeds

## Reference Files

- Phase 3 task records: `records/3.*.md`
- `docs/features/soft-delete-consistency/design/tech-design.md` — Security Threat T3, T5

## Acceptance Criteria

- [ ] All verification checklist items pass
- [ ] Deviations documented as decisions
- [ ] Record created with test evidence

## Implementation Notes

This is the final verification gate. After this, T-test-1 and T-test-2 generate e2e test cases and scripts.
