---
id: "3.summary"
title: "Phase 3 Summary"
priority: "P0"
estimated_time: "15min"
dependencies: ["3.x"]
status: pending
---

# 3.summary: Phase 3 Summary

## Description

Generate a structured summary of all completed tasks in this phase. This summary is read by T-test tasks to maintain cross-phase consistency.

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/unify-permission-checks/tasks/records/` whose filename starts with `3.` and does NOT contain `.summary`.

### Step 2: Extract structured data into the summary field

## Reference Files

- All phase task records: `docs/features/unify-permission-checks/tasks/records/3.*.md`
- Design reference: `docs/features/unify-permission-checks/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template
- [ ] Types & Interfaces Changed table is populated
- [ ] Record created via `/record-task` with `coverage: -1.0`

## Implementation Notes

This is a documentation-only task. No code should be written.
