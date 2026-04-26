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

Exit verification gate for Phase 1 (P0 Bug Fixes). Confirms that both bug fixes are correct, tests pass, and no regressions before dead code removal begins.

## Verification Checklist

1. [ ] `go build ./...` compiles without errors
2. [ ] `go test ./internal/service/ -run TestSubItem` passes (Item 1)
3. [ ] `go test ./internal/repository/gorm/ -run TestFilter` passes (Item 2)
4. [ ] `grep -rn "assignee_id" backend/internal/service/sub_item_service.go` returns zero
5. [ ] Filter behavior verified: valid bizKey returns correct subset, invalid returns empty
6. [ ] No deviations from design spec (or deviations are documented as decisions)

## Reference Files

- `docs/features/schema-alignment-cleanup/design/tech-design.md` — Items 1-2 specification
- This phase's task records — `records/1.*.md`
- This phase's summary — `records/1-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If issues are found:
1. Fix inline if trivial
2. Document non-trivial issues as decisions in the record
3. Set status to `blocked` if a blocking issue cannot be resolved
