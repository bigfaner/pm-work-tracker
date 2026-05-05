---
id: "T-test-5"
title: "Consolidate Specs"
priority: "P2"
estimated_time: "20min"
dependencies: ["T-test-4"]
status: pending
---

# Consolidate Specs

## Description

Call `/consolidate-specs` skill to extract business rules from PRD and technical specifications from design into `specs/` directory. Present preview to user for review before integrating to project-level shared directories.

## Reference Files

- `docs/features/decision-log/prd/prd-spec.md` — Source for business rules
- `docs/features/decision-log/prd/prd-user-stories.md` — Source for business context
- `docs/features/decision-log/design/tech-design.md` — Source for technical specs
- `docs/features/decision-log/design/api-handbook.md` — Source for API contracts

## Acceptance Criteria

- [ ] `docs/features/decision-log/specs/biz-specs.md` exists with extracted business rules
- [ ] `docs/features/decision-log/specs/tech-specs.md` exists with extracted technical specs
- [ ] If any `[CROSS]` items exist: `docs/features/decision-log/specs/review-choices.md` exists with user's approved/rejected items
- [ ] If integration occurred: only items marked "approved" in review-choices.md were integrated to project-level dirs
- [ ] `docs/features/decision-log/specs/.integrated` marker exists

## Skip Conditions

If ALL extracted items are `[LOCAL]` (no cross-cutting candidates), generate preview files only and mark task completed with `coverage: -1.0`. No integration step needed.

If no extractable rules found in PRD/design, mark task completed with `coverage: -1.0`.

If running under `/run-tasks` (non-interactive session) and CROSS items exist, write preview files and mark task as `blocked` with note "User review required for integration." Do NOT auto-integrate.

## User Stories

No direct user story mapping. This is a standard knowledge consolidation task.

## Implementation Notes

**Step 1: Verify prerequisites**

Confirm feature documents exist. If missing, mark task `blocked` and stop.

Check idempotency: if `docs/features/decision-log/specs/.integrated` exists, skip.

**Step 2: Extract and classify**

Run `/consolidate-specs` skill.

**Step 3: Early exit or user review**

If ALL items are `[LOCAL]`, skip to Step 5.

Otherwise, present preview files to the user for review.

**Step 4: Integrate approved items**

For each approved item, append to project-level file.

**Step 5: Record**

Record task via `/record-task` skill (set `coverage: -1.0`).
