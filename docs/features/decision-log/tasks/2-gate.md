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

Exit verification gate for Phase 2 (Backend Implementation). Confirms that all backend CRUD operations, permission checks, draft visibility filtering, and route registration are complete and working.

## Verification Checklist

1. [ ] All handler methods compile without errors
2. [ ] Data models match `design/tech-design.md` (no drift from Phase 1)
3. [ ] No type mismatches between Repository → Service → Handler → VO layers
4. [ ] Project builds successfully: `cd backend && go build ./...`
5. [ ] All existing tests pass: `cd backend && go test ./...`
6. [ ] Service unit tests cover: draft-only edit, owner-only access, status transitions
7. [ ] Handler unit tests cover: request binding, permission checks, response format
8. [ ] Route paths match api-handbook spec
9. [ ] No deviations from design spec (or deviations are documented as decisions)

## Reference Files

- `design/tech-design.md` — Cross-Layer Data Map, Security Considerations sections
- Phase 2 task records — `docs/features/decision-log/tasks/records/2.*.md`
- Phase 2 summary — `docs/features/decision-log/tasks/records/2-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If issues are found:
1. Fix inline if trivial (e.g., type mismatch in a single file)
2. Document non-trivial issues as decisions in the record
3. Set status to `blocked` if a blocking issue cannot be resolved
