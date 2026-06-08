---
status: "completed"
started: "2026-06-08 13:17"
completed: "2026-06-08 13:23"
time_spent: "~6m"
---

# Task Record: T-specs-consolidate Consolidate Specs

## Summary
Consolidated milestone-map specs: integrated 3 CROSS tech specs (performance targets, completion calculation, entity hierarchy) into project-level dirs. Fixed 1 drift (phantom button-icons.md reference in INDEX). All 6 business rules were already consolidated in prior iterations.

## Changes

### Files Created
- docs/features/milestone-map/specs/biz-specs.md
- docs/features/milestone-map/specs/tech-specs.md
- docs/features/milestone-map/specs/review-choices.md
- docs/features/milestone-map/specs/.integrated

### Files Modified
- docs/conventions/performance-targets.md
- docs/conventions/data-model.md
- docs/conventions/INDEX.md
- docs/features/milestone-map/manifest.md

### Key Decisions
无

## Document Metrics
3 CROSS tech specs integrated, 6 BIZ rules pre-existing, 1 drift fixed, 0 orphaned specs

## Referenced Documents
- docs/features/milestone-map/prd/prd-spec.md
- docs/features/milestone-map/design/tech-design.md
- docs/features/milestone-map/design/api-handbook.md
- docs/business-rules/milestone-state-constraints.md
- docs/conventions/status-machine.md
- docs/conventions/permission-codes.md
- docs/conventions/soft-delete.md
- docs/conventions/error-codes.md

## Review Status
final

## Acceptance Criteria
- [x] New business rules extracted to docs/business-rules/
- [x] New tech specs extracted to docs/conventions/
- [x] No drift between feature docs and consolidated specs

## Notes
Non-interactive mode: auto-integrated all CROSS items. Most milestone rules were already consolidated during prior feature work. Drift scan found button-icons.md referenced in INDEX but file did not exist.
