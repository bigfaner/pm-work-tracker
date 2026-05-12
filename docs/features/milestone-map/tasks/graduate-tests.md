---
id: "T-test-4"
title: "Graduate Test Scripts"
priority: "P1"
estimated_time: "30min"
dependencies: ["T-test-3"]
status: pending
noTest: false
mainSession: false
---

# Graduate Test Scripts

## Description

Call `/graduate-tests` skill to migrate feature test scripts to the project-wide regression suite.

## Reference Files

- `tests/e2e/features/milestone-map/results/latest.md` — Must show PASS
- `tests/e2e/features/milestone-map/` — Source scripts

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/results/latest.md` shows status = PASS
- [ ] `tests/e2e/.graduated/milestone-map` marker exists
- [ ] Spec files present in `tests/e2e/<module>/`

## User Stories

No direct user story mapping.

## Implementation Notes

1. Verify latest.md shows PASS
2. Run `/graduate-tests` skill
3. Record task completed
