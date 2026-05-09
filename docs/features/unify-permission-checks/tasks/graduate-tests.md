---
id: "T-test-4"
title: "Graduate Test Scripts"
priority: "P1"
estimated_time: "30min"
dependencies: ["T-test-3"]
status: pending
---

# Graduate Test Scripts

## Description

Call `/graduate-tests` skill to migrate feature test scripts from `tests/e2e/features/unify-permission-checks/` to the project-wide regression suite.

## Reference Files

- `tests/e2e/features/unify-permission-checks/results/latest.md` — Must show PASS
- `tests/e2e/features/unify-permission-checks/` — Source scripts
- `tests/e2e/` — Destination regression suite

## Acceptance Criteria

- [ ] `tests/e2e/features/unify-permission-checks/results/latest.md` shows status = PASS
- [ ] `tests/e2e/.graduated/unify-permission-checks` marker exists
- [ ] Spec files present in `tests/e2e/<module>/`

## User Stories

No direct user story mapping. This is a standard test graduation task.

## Implementation Notes

1. Verify e2e passed
2. Run `/graduate-tests` skill
3. Mark completed
