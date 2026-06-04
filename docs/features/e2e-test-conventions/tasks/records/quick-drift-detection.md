---
status: "completed"
started: "2026-06-03 21:40"
completed: "2026-06-03 21:50"
time_spent: "~10m"
---

# Task Record: T-quick-doc-drift Detect Spec Drift

## Summary
Detected and auto-fixed spec drift in testing conventions. API testing spec (testing/api/core.md) drifted from Go-based to TypeScript/Vitest-based implementation. Web E2E testing spec (testing/web/core.md) had inaccurate file naming, directory structure, assertion library, and helper descriptions. Testing index (testing/index.md) updated with correct framework and file locations.

## Changes

### Files Created
- docs/.vocabulary.md

### Files Modified
- docs/conventions/testing/api/core.md
- docs/conventions/testing/web/core.md
- docs/conventions/testing/index.md

### Key Decisions
无

## Document Metrics
3 spec files drift-fixed, 0 orphaned rules, 0 new implicit rules, 1 vocabulary index generated

## Referenced Documents
- docs/conventions/testing/api/core.md
- docs/conventions/testing/web/core.md
- docs/conventions/testing/index.md
- docs/conventions/testing/unit.md
- docs/conventions/api-boundary.md
- docs/conventions/frontend-architecture.md
- docs/conventions/lint-enforcement.md
- docs/conventions/data-model.md
- docs/conventions/naming.md
- docs/conventions/development-workflow.md

## Review Status
final

## Acceptance Criteria
- [x] No spec drift detected between updated convention files and current codebase, or all detected drifts auto-fixed

## Notes
Drift detected in API testing spec (Go -> TypeScript/Vitest) and Web E2E testing spec (inaccurate file naming, assertion library, helpers). All drifts auto-fixed. No orphaned rules found. Other spec files (api-boundary, frontend-architecture, lint-enforcement, data-model, naming, development-workflow) verified as current. Committed with [auto-specs] tag for traceability.
