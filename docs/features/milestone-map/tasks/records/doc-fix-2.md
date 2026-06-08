---
status: "completed"
started: "2026-06-08 04:54"
completed: "2026-06-08 05:01"
time_spent: "~7m"
---

# Task Record: doc-fix-2 Fix: Remaining journey eval failures — validation-error, fact annotations, API depth

## Summary
Fixed remaining 5 journey eval failures below 850: added cancelled transition step and milestone setup to milestone-map-lifecycle (690), added API not-found/duplicate-name/BR-5 terminal guard steps to milestone-lifecycle (775), fixed API bypass language and bizKey terminology in item-milestone-binding (770), fixed hallucinated redirect claim and added permission-denied scenario to milestone-map-visualization (810), fixed hallucinated redirect claim and added fact annotations to read-only-milestone-access (775). Added validation-error justification notes to read-only-only journeys. Added fact traceability annotations across all 5 journeys.

## Changes

### Files Created
无

### Files Modified
- docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
- docs/features/milestone-map/testing/milestone-lifecycle/journey.md
- docs/features/milestone-map/testing/item-milestone-binding/journey.md
- docs/features/milestone-map/testing/milestone-map-visualization/journey.md
- docs/features/milestone-map/testing/read-only-milestone-access/journey.md

### Key Decisions
无

## Document Metrics
5 journeys fixed: +cancelled cascade step, +3 API steps (not-found, duplicate-name, BR-5 guard), +validation-error outcome, +permission-denied scenario, 2 hallucinated claims corrected, 15+ fact annotations added

## Referenced Documents
- docs/features/milestone-map/testing/journeys/.eval-report.md
- docs/features/milestone-map/testing/milestone-map-lifecycle/eval/iteration-1.md
- docs/features/milestone-map/testing/milestone-lifecycle/eval/iteration-1.md
- docs/features/milestone-map/prd/prd-spec.md

## Review Status
final

## Acceptance Criteria
- [x] milestone-map-lifecycle: cancelled transition step added
- [x] milestone-map-lifecycle: milestone creation in setup
- [x] milestone-map-lifecycle: API not-found and validation-error steps added
- [x] milestone-lifecycle: not-found and duplicate-name outcomes added
- [x] milestone-lifecycle: BR-5 parent-terminal guard step added
- [x] milestone-lifecycle: invariant contradiction fixed
- [x] item-milestone-binding: validation-error outcome added
- [x] item-milestone-binding: Step 2d API bypass language removed
- [x] item-milestone-binding: Step 5b bizKey language fixed
- [x] milestone-map-visualization: hallucinated redirect claim fixed
- [x] milestone-map-visualization: permission-denied scenario added
- [x] milestone-map-visualization: Step 3b fixed as non-user-action
- [x] read-only-milestone-access: hallucinated redirect claim fixed
- [x] read-only-milestone-access: fact annotations added to unannotated steps
- [x] read-only-only journeys: validation-error justification notes added

## Notes
This was a fix-record-missed recovery task. The original doc-fix-2 execution was claimed but never implemented. This execution performed the full implementation. The Content Fix Guidance in the task file correctly states that code quality gates (lint, compile, test) are irrelevant for doc fixes -- only the 5 journey markdown files were modified.
