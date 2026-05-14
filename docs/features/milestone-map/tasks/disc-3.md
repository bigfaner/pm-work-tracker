---
id: "disc-3"
title: "Verify all e2e scripts compile and run"
priority: "P1"
estimated_time: "15min"
dependencies: ["fix-2", "disc-1", "disc-2"]
status: pending
breaking: false
noTest: true
---

# Verify all e2e scripts compile and run

Final verification after all 3 split sub-tasks complete.

## Description

Verify that all generated e2e test scripts compile and have no unresolved markers.

## Acceptance Criteria

- [ ] `cd tests/e2e && npx tsc --noEmit` passes
- [ ] `just e2e-verify --feature milestone-map` shows no unresolved markers
- [ ] All 3 spec files exist under `tests/e2e/features/milestone-map/`

## Implementation Notes

1. Run `cd tests/e2e && npx tsc --noEmit` to verify TypeScript compilation
2. Run `just e2e-verify --feature milestone-map` to check for unresolved markers
3. If any issues found, fix inline
