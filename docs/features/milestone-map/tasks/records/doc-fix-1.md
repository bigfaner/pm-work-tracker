---
status: "completed"
started: "2026-06-08 04:29"
completed: "2026-06-08 04:37"
time_spent: "~8m"
---

# Task Record: doc-fix-1 Fix: Journey eval failures — missing mandatory outcomes and traceability

## Summary
Fixed all 7 journey files addressing 6 systemic eval failures: added session-expired boundary outcomes (7/7), added unauthorized API boundary outcomes for dual-surface journeys (2/2), added fact_id traceability with source:inferred annotations (80 fact annotations, 31 inferred annotations across all files), fixed surface_types to match actual content (5 journeys changed from [web,api] to [web]), removed CSS implementation coupling (rounded-full, w-32, opacity-50, text-tertiary, translate-x 300ms, debounce timings), and added validation-error boundary outcome annotations where forms exist.

## Changes

### Files Created
无

### Files Modified
- docs/features/milestone-map/testing/milestone-map-lifecycle/journey.md
- docs/features/milestone-map/testing/milestone-lifecycle/journey.md
- docs/features/milestone-map/testing/milestone-item-management/journey.md
- docs/features/milestone-map/testing/item-milestone-binding/journey.md
- docs/features/milestone-map/testing/milestone-map-visualization/journey.md
- docs/features/milestone-map/testing/item-list-milestone-integration/journey.md
- docs/features/milestone-map/testing/read-only-milestone-access/journey.md

### Key Decisions
无

## Document Metrics
session-expired: 7/7 journeys, unauthorized: 2/2 API-surface journeys, fact annotations: 80, inferred annotations: 31, CSS coupling removed: 0 remaining

## Referenced Documents
- docs/features/milestone-map/testing/journeys/.eval-report.md
- docs/features/milestone-map/testing/milestone-map-lifecycle/eval/iteration-1.md
- docs/features/milestone-map/testing/milestone-lifecycle/eval/iteration-1.md
- docs/features/milestone-map/prd/prd-spec.md
- docs/features/system-ux-optimization/testing/sub-item-management/journey.md

## Review Status
final

## Acceptance Criteria
- [x] All 7 journeys have session-expired boundary outcome
- [x] API-surface journeys have unauthorized boundary outcome or surface_types corrected to web-only
- [x] All derived outcomes have source:inferred annotations
- [x] All factual claims have fact traceability annotations
- [x] No CSS class or timing implementation coupling remains
- [x] surface_types matches actual content for all journeys

## Notes
Also fixed incorrect delete constraint in milestone-map-lifecycle (changed from planning-only to planning/reviewed/ready). Fixed milestone-lifecycle invariant to include completed->in_progress rollback and cancelled as terminal state. Changed milestone-item-management and item-milestone-binding surface_types from [web,api] to [web] since they had no API-level steps.
