---
status: "completed"
started: "2026-06-11 11:27"
completed: "2026-06-11 11:29"
time_spent: "~2m"
---

# Task Record: disc-1 Lesson: E2E test hardcoded IP/port anti-pattern

## Summary
Created lesson gotcha-hardcoded-url-e2e.md documenting the anti-pattern of hardcoding IP/port in E2E page.goto instead of importing baseUrl from helpers. Updated INDEX.md with new entry and added to E2E test generation trigger.

## Changes

### Files Created
- docs/lessons/gotcha-hardcoded-url-e2e.md

### Files Modified
- docs/lessons/INDEX.md

### Key Decisions
无

## Document Metrics
1 new lesson, 2 INDEX entries (category table + trigger row)

## Referenced Documents
无

## Review Status
final

## Acceptance Criteria
- [x] Lesson file created with Problem/Root Cause/Solution/Reusable Pattern sections
- [x] INDEX.md updated with new entry in gotcha category table
- [x] INDEX.md updated in E2E test generation trigger row
- [x] No duplicate of existing gotcha-e2e-script-generation lesson

## Notes
Existing gotcha-e2e-script-generation.md covers port guessing during generation; this new lesson specifically targets the hardcoded-URL fix pattern when resolving page.goto failures. Related via cross-link.
