---
id: "4.gate"
title: "Phase 4 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["4.summary"]
status: pending
breaking: true
---

# 4.gate: Phase 4 Exit Gate

## Description

Final exit verification gate for Phase 4 (Architecture Alignment). Confirms full test suite passes across all layers after all 24 items are complete. This is the last gate before e2e test generation.

## Verification Checklist

1. [ ] `go build ./...` compiles without errors
2. [ ] `go test ./...` passes (full backend suite)
3. [ ] `npx vitest run` passes (full frontend suite)
4. [ ] `npx tsc --noEmit` passes
5. [ ] Both SQLite and MySQL schema files in sync
6. [ ] No type mismatches between layers
7. [ ] Coverage floors met: backend >= 78%, frontend >= 90%
8. [ ] No deviations from design spec

## Reference Files

- `docs/features/schema-alignment-cleanup/design/tech-design.md` — full specification
- This phase's task records — `records/4.*.md`
- This phase's summary — `records/4-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Full test suite green (backend + frontend)
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If issues are found:
1. Fix inline if trivial
2. Document non-trivial issues as decisions in the record
3. Set status to `blocked` if a blocking issue cannot be resolved
