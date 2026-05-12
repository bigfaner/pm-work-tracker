---
id: "T-test-5"
title: "Consolidate Specs"
priority: "P2"
estimated_time: "20min"
dependencies: ["T-test-4.5"]
status: pending
noTest: true
mainSession: false
---

# Consolidate Specs

## Description

Call `/consolidate-specs` skill to extract business rules and technical specifications from PRD and design documents.

## Reference Files

- `docs/features/milestone-map/prd/prd-spec.md`
- `docs/features/milestone-map/prd/prd-user-stories.md`
- `docs/features/milestone-map/design/tech-design.md`
- `docs/features/milestone-map/design/api-handbook.md`

## Acceptance Criteria

- [ ] `docs/features/milestone-map/specs/biz-specs.md` exists
- [ ] `docs/features/milestone-map/specs/tech-specs.md` exists
- [ ] If CROSS items exist: `docs/features/milestone-map/specs/review-choices.md` exists
- [ ] `docs/features/milestone-map/specs/.integrated` marker exists

## User Stories

No direct user story mapping.

## Implementation Notes

1. Verify prerequisites exist
2. Run `/consolidate-specs` skill
3. If all LOCAL: early exit, mark completed
4. If CROSS items: present to user for review before integration
