---
id: "doc-fix-1"
title: "Fix: Journey eval failures — missing mandatory outcomes and traceability"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: false
type: "doc.fix"
---

# Fix: Journey eval failures — missing mandatory outcomes and traceability

## Root Cause

All 7 journeys failed eval against 850/1000 target. Priority fixes needed:
1. Add session-expired boundary outcome to all 7 web-surface journeys
2. Add unauthorized boundary outcome to 5 API-surface journeys (or remove api from surface_types)
3. Add fact_id traceability and source:inferred annotations to all derived outcomes
4. Add validation-error boundary outcomes where applicable
5. Remove implementation coupling (CSS classes, timings) from outcome descriptions
6. Restrict surface_types to [web] for purely UI journeys or add API-level steps
See testing/journeys/.eval-report.md for per-journey scores and attacks.

## Reference Files

- Source: docs/features/milestone-map/testing/*/journey.md
- Error details: All 7 journeys scored below 850/1000 (avg 687). Systemic issues: missing session-expired (7/7), unauthorized (5/7), no fact traceability (7/7), API surface declared but empty (5/7).

## Content Fix Guidance

When fixing documentation failures, observe these boundaries:

**Scope:**
- Fix only the markdown/content issues identified in the root cause
- Do not modify source code files — this is a documentation-only fix
- Do not run code quality gates (lint, compile, test) — they are irrelevant for doc fixes

**Correct workflow:**
1. Read the failing document and understand the reported issue
2. Identify the specific content problem (broken links, missing sections, incorrect terminology, formatting errors)
3. Apply the minimal fix to resolve the issue
4. Verify the document renders correctly and internal references are valid

When this task is recorded as completed via `task record`, the source task T-eval-journey is automatically restored to pending if all its dependencies are completed.
