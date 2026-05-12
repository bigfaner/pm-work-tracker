---
id: "T-test-2"
title: "Generate e2e Test Scripts"
priority: "P1"
estimated_time: "1-2h"
dependencies: ["T-test-1b"]
status: pending
noTest: false
mainSession: false
---

# Generate e2e Test Scripts

## Description

Call `/gen-test-scripts` skill to generate executable TypeScript e2e test scripts from test cases.

## Reference Files

- `testing/test-cases.md` — Test case document
- `docs/sitemap/sitemap.json` — Page element locators
- `tests/e2e/config.yaml` — Test environment config

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/` contains at least one spec file
- [ ] NO spec files exist directly at `tests/e2e/milestone-map/`
- [ ] `tests/e2e/helpers.ts` exists
- [ ] Each test() includes traceability comment

## User Stories

No direct user story mapping.

## Implementation Notes

1. Run `/gen-test-scripts` skill
2. Verify spec files under `tests/e2e/features/milestone-map/`
3. Run `just e2e-verify --feature milestone-map` to check for unresolved markers
4. Run TypeScript compilation check: `cd tests/e2e && npx tsc --noEmit`
