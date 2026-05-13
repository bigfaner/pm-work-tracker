---
id: "disc-1"
title: "Generate UI Milestones page test scripts (TC-001~TC-030)"
priority: "P1"
estimated_time: "1h"
dependencies: ["T-test-1b"]
status: pending
breaking: false
---

# Generate UI Milestones page test scripts (TC-001~TC-030)

Split from T-test-2 per gotcha-large-output-stall-subagent.md to avoid subagent output stall.

## Description

Generate `tests/e2e/features/milestone-map/milestones-page.spec.ts` for 30 UI test cases (TC-001~TC-030, ~36KB). These cover milestone map CRUD, timeline view, milestone CRUD, and detail panel.

## Reference Files

- `testing/test-cases.md` — Test case document (read TC-001 through TC-030 only)
- `docs/sitemap/sitemap.json` — Page element locators
- `tests/e2e/config.yaml` — Test environment config

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/milestones-page.spec.ts` exists
- [ ] All 30 UI test cases (TC-001~TC-030) are covered
- [ ] Each test() includes traceability comment (TC ID + PRD source)
- [ ] TypeScript compilation passes: `cd tests/e2e && npx tsc --noEmit`

## Implementation Notes

1. Run `/gen-test-scripts` skill, scoped to milestone page test cases only (TC-001~TC-030)
2. If the skill doesn't support scoping, manually generate the spec file following patterns from existing `tests/e2e/features/` specs
3. Verify with `cd tests/e2e && npx tsc --noEmit`
