---
status: "completed"
started: "2026-06-05 00:06"
completed: "2026-06-05 00:18"
time_spent: "~12m"
---

# Task Record: T-specs-consolidate Consolidate Specs

## Summary
Extracted and consolidated 19 CROSS specs (12 business rules + 7 tech specs) from system-ux-optimization feature documents into project-level dirs. Created 2 new business-rules files (item-lifecycle.md, filtering.md), appended to 9 existing convention files. Fixed 1 drift issue (error code naming). Generated vocabulary index.

## Changes

### Files Created
- docs/business-rules/item-lifecycle.md
- docs/business-rules/filtering.md
- docs/features/system-ux-optimization/specs/biz-specs.md
- docs/features/system-ux-optimization/specs/tech-specs.md
- docs/features/system-ux-optimization/specs/review-choices.md
- docs/features/system-ux-optimization/specs/.integrated

### Files Modified
- docs/conventions/soft-delete.md
- docs/conventions/item-codes.md
- docs/conventions/data-model.md
- docs/conventions/permission-codes.md
- docs/conventions/error-codes.md
- docs/conventions/security.md
- docs/conventions/authorization.md
- docs/conventions/frontend-ux.md
- docs/conventions/performance-targets.md
- docs/conventions/data-validation.md
- docs/conventions/api-boundary.md
- docs/features/system-ux-optimization/manifest.md
- docs/.vocabulary.md

### Key Decisions
无

## Document Metrics
12 biz rules extracted (4 LOCAL, 8 CROSS integrated), 7 tech specs extracted (5 LOCAL, 2 CROSS integrated), 9 convention files extended, 2 new business-rule files, 1 drift fix

## Referenced Documents
- docs/features/system-ux-optimization/prd/prd-spec.md
- docs/features/system-ux-optimization/design/tech-design.md
- docs/conventions/soft-delete.md
- docs/conventions/permission-codes.md
- docs/conventions/error-codes.md
- docs/conventions/api-boundary.md
- docs/conventions/status-machine.md
- docs/conventions/data-model.md
- docs/conventions/item-codes.md
- docs/conventions/data-validation.md
- docs/conventions/security.md
- docs/conventions/authorization.md
- docs/conventions/frontend-ux.md
- docs/conventions/performance-targets.md

## Review Status
final

## Acceptance Criteria
- [x] Task completes without error

## Notes
Non-interactive mode: all CROSS items auto-integrated with [auto-specs] commit tag. Drift detection found 1 issue: error-codes.md documented ErrTargetClosed/ErrSameMainItem with invented unique codes (TARGET_CLOSED, SAME_MAIN_ITEM) but code uses BAD_REQUEST -- corrected doc to match code. Permission code count updated from 29 to 31.
