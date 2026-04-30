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

Call `/graduate-tests` skill to migrate feature test scripts from `tests/e2e/e2e-test-scripts-rebuild/` to the project-wide regression suite.

## Reference Files

- `testing/results/latest.md` — Must show status = PASS before graduating
- `tests/e2e/e2e-test-scripts-rebuild/` — Source scripts to migrate
- `tests/e2e/` — Destination regression suite

## Acceptance Criteria

- [ ] `testing/results/latest.md` shows status = PASS
- [ ] `tests/e2e/.graduated/e2e-test-scripts-rebuild` marker exists
- [ ] Spec files present in `tests/e2e/<category>/`

## User Stories

No direct user story mapping. This is a standard test graduation task.

## Implementation Notes

**Step 1: Verify e2e passed**

Read `testing/results/latest.md`. Check status field.
- Status = PASS → proceed to Step 2
- Status = FAIL → mark task `blocked` and stop

**Step 2: Graduate**

Run `/graduate-tests` skill.

**Step 3: Record**

Mark task completed.
