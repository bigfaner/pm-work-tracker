---
id: "T-test-1"
title: "Generate e2e Test Cases"
priority: "P1"
estimated_time: "1-2h"
dependencies: ["3.gate"]
status: pending
noTest: true
mainSession: false
---

# Generate e2e Test Cases

## Description

Call `/gen-test-cases` skill to generate structured test case documentation from PRD acceptance criteria.

## Reference Files

- `prd/prd-spec.md` — PRD specification
- `prd/prd-user-stories.md` — User stories (with Given/When/Then acceptance criteria)
- `prd/prd-ui-functions.md` — UI function requirements

## Acceptance Criteria

- [ ] `testing/test-cases.md` file created
- [ ] Each test case includes Target and Test ID fields
- [ ] All test cases traceable to PRD acceptance criteria
- [ ] Test cases grouped by type (UI -> API -> CLI)

## User Stories

No direct user story mapping. This is a standard test generation task.

## Implementation Notes

1. `docs/sitemap/sitemap.json` exists — skip `/gen-sitemap`
2. Run `/gen-test-cases` skill
3. Verify generated `testing/test-cases.md` contains Target and Test ID fields
