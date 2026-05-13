---
id: "T-test-3"
title: "Run e2e Tests"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["disc-3"]
status: pending
noTest: false
mainSession: false
---

# Run e2e Tests

## Description

Execute the generated e2e test scripts and produce a results report.

## Reference Files

- `tests/e2e/features/milestone-map/` — Test scripts
- `tests/e2e/features/milestone-map/results/latest.md` — Output

## Hard Rules

- MUST use `just e2e-setup` then `just test-e2e --feature milestone-map` to run tests
- MUST NOT use `npx playwright test` directly — `just test-e2e` handles server lifecycle, env vars, and test discovery
- If tests fail: create fix tasks per root cause using `task add --template fix-task --block-source`

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/results/latest.md` exists
- [ ] All tests pass (status = PASS)

## Implementation Notes

1. Run `just e2e-setup` to ensure dependencies are ready
2. Run `just e2e-verify --feature milestone-map` to check for unresolved markers
3. Run `just test-e2e --feature milestone-map` to execute tests
4. Collect results and write report to `results/latest.md`
5. If tests fail: analyze failures, create fix tasks, do NOT mark this task completed
