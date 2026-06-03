---
status: "completed"
started: "2026-06-03 22:30"
completed: "2026-06-03 22:32"
time_spent: "~2m"
---

# Task Record: 7 更新 API 测试实践规范（api/core.md）

## Summary
Updated api/core.md with 7 practice rules (E-100~E-106), 8-item anti-pattern table, and updated domains frontmatter

## Changes

### Files Created
无

### Files Modified
- docs/conventions/testing/api/core.md

### Key Decisions
无

## Document Metrics
7 practice rules (E-100~E-106), 8 anti-patterns, 125 total lines

## Referenced Documents
- docs/proposals/e2e-test-conventions/proposal.md

## Review Status
final

## Acceptance Criteria
- [x] api/core.md contains 7 rules (E-100~E-106) with ID, rule name, correct example, incorrect example
- [x] api/core.md contains 8 anti-patterns with description and alternative
- [x] api/core.md domains frontmatter updated with testing-api, vitest, http-client

## Notes
Replaced existing 5-item anti-pattern table with 8-item table from proposal. Appended practice rules section after auto-generated content without conflicting with forge:test-guide marker. Content is tool/method-focused, no overlap with strategy sections.
