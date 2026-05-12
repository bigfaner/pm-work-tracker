---
id: "T-test-3"
title: "Run e2e Tests"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["T-test-2"]
status: pending
noTest: false
mainSession: false
---

# Run e2e Tests

## Description

Call `/run-e2e-tests` skill to execute the generated test scripts and produce a results report.

## Reference Files

- `tests/e2e/features/milestone-map/` — Test scripts
- `tests/e2e/features/milestone-map/results/latest.md` — Output

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/results/latest.md` exists
- [ ] All tests pass (status = PASS)

## User Stories

No direct user story mapping.

## Implementation Notes

1. Run `/run-e2e-tests` skill
2. Run: `just test-e2e --feature milestone-map`
3. If tests fail: create fix tasks per root cause using `task add --template fix-task`
