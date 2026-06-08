---
id: "doc-fix-2"
title: "Fix: Remaining journey eval failures — validation-error, fact annotations, API depth"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: false
type: "doc.fix"
---

# Fix: Remaining journey eval failures — validation-error, fact annotations, API depth

## Root Cause

5 journeys remain below 850 after doc-fix-1. Focused fixes needed per journey:
- milestone-map-lifecycle (690): Add cancelled transition step, add milestone creation to setup, fix Step 2c annotation (server-error→concurrent-edit), add API not-found/validation-error steps
- milestone-lifecycle (775): Add not-found and duplicate-name outcomes, add BR-5 parent-terminal guard, fix invariant contradiction on cancellation scope, add API depth
- item-milestone-binding (770): Add validation-error outcome, remove Step 2d API bypass, fix Step 5b bizKey language
- milestone-map-visualization (810): Add validation-error justification note, add permission-denied scenario, fix Step 3b 'Data is loading' as non-user-action
- read-only-milestone-access (775): Fix Step E1 hallucinated redirect claim, add fact annotations to unannotated steps, add validation-error justification

## Reference Files

- Source: docs/features/milestone-map/testing/*/journey.md
- Error details: 5/7 journeys below 850 after iteration 2. Remaining gaps: missing validation-error (5 journeys), thin API surface (2), incomplete fact annotations (3), missing cancellation steps (1), hallucinated claim (1).

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
