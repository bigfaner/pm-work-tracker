---
id: "disc-2"
title: "Generate UI Existing pages + Integration test scripts (TC-031~TC-052)"
priority: "P1"
estimated_time: "1h"
dependencies: ["T-test-1b"]
status: pending
breaking: false
---

# Generate UI Existing pages + Integration test scripts (TC-031~TC-052)

Split from T-test-2 per gotcha-large-output-stall-subagent.md to avoid subagent output stall.

## Description

Generate `tests/e2e/features/milestone-map/existing-pages.spec.ts` for 22 test cases (TC-031~TC-052, ~26KB). These cover item edit milestone selector, items page filter, table view column, permission views, navigation, and cross-interface integration.

## Reference Files

- `testing/test-cases.md` — Test case document (read TC-031 through TC-052 only)
- `docs/sitemap/sitemap.json` — Page element locators
- `tests/e2e/config.yaml` — Test environment config

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/existing-pages.spec.ts` exists
- [ ] All 22 test cases (TC-031~TC-052) are covered
- [ ] Each test() includes traceability comment (TC ID + PRD source)
- [ ] TypeScript compilation passes: `cd tests/e2e && npx tsc --noEmit`

## Implementation Notes

1. Run `/gen-test-scripts` skill, scoped to existing page test cases only (TC-031~TC-052)
2. If the skill doesn't support scoping, manually generate the spec file following patterns from existing `tests/e2e/features/` specs
3. Verify with `cd tests/e2e && npx tsc --noEmit`
