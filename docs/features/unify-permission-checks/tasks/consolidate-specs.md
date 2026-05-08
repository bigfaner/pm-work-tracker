---
id: "T-test-5"
title: "Consolidate Specs"
priority: "P2"
estimated_time: "20min"
dependencies: ["T-test-4.5"]
status: pending
---

# Consolidate Specs

## Description

Call `/consolidate-specs` skill to extract business rules from PRD and technical specifications from design into `specs/` directory. Present preview to user for review before integrating to project-level shared directories.

## Reference Files

- `docs/features/unify-permission-checks/prd/prd-spec.md` — Source for business rules
- `docs/features/unify-permission-checks/design/tech-design.md` — Source for technical specs
- `docs/features/unify-permission-checks/design/api-handbook.md` — Source for API contracts

## Acceptance Criteria

- [ ] `docs/features/unify-permission-checks/specs/biz-specs.md` exists
- [ ] `docs/features/unify-permission-checks/specs/tech-specs.md` exists
- [ ] If any `[CROSS]` items exist: `docs/features/unify-permission-checks/specs/review-choices.md` exists
- [ ] `docs/features/unify-permission-checks/specs/.integrated` marker exists

## User Stories

No direct user story mapping. This is a standard knowledge consolidation task.

## Implementation Notes

1. Verify prerequisites exist
2. Run `/consolidate-specs` skill
3. If ALL items are LOCAL, skip integration and record with `coverage: -1.0`
