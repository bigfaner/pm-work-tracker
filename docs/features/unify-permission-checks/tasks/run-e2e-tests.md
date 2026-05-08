---
id: "T-test-3"
title: "Run e2e Tests"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["T-test-2"]
status: pending
---

# Run e2e Tests

## Description

Call `/run-e2e-tests` skill to execute the generated test scripts and produce a results report.

## Reference Files

- `tests/e2e/features/unify-permission-checks/` — Test scripts
- `tests/e2e/features/unify-permission-checks/results/latest.md` — Output

## Acceptance Criteria

- [ ] `tests/e2e/features/unify-permission-checks/results/latest.md` exists
- [ ] All tests pass (status = PASS in latest.md)

## User Stories

No direct user story mapping. This is a standard test execution task.

## Implementation Notes

1. Run `/run-e2e-tests` skill
2. Read results to determine outcome
3. On failure: create fix tasks, mark this task blocked
