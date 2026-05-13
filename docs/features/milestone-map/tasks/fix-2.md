---
id: "fix-2"
title: "Generate API e2e test scripts (TC-053~TC-070)"
priority: "P0"
estimated_time: "30min"
dependencies: ["T-test-1b"]
status: pending
breaking: false
---

# Generate API e2e test scripts (TC-053~TC-070)

Split from T-test-2 per gotcha-large-output-stall-subagent.md to avoid subagent output stall.

## Description

Generate `tests/e2e/features/milestone-map/api.spec.ts` for 18 API test cases (TC-053~TC-070, ~23KB).

## Reference Files

- `testing/test-cases.md` — Test case document (read TC-053 through TC-070 only)
- `docs/sitemap/sitemap.json` — Page element locators
- `tests/e2e/config.yaml` — Test environment config

## Acceptance Criteria

- [ ] `tests/e2e/features/milestone-map/api.spec.ts` exists
- [ ] All 18 API test cases (TC-053~TC-070) are covered
- [ ] Each test() includes traceability comment (TC ID + PRD source)
- [ ] TypeScript compilation passes: `cd tests/e2e && npx tsc --noEmit`

## Implementation Notes

1. Run `/gen-test-scripts` skill, scoped to API test cases only (TC-053~TC-070)
2. If the skill doesn't support scoping, manually generate the spec file following patterns from existing `tests/e2e/features/` specs
3. Verify with `cd tests/e2e && npx tsc --noEmit`
