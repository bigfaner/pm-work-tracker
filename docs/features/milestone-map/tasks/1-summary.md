---
id: "1.summary"
title: "Phase 1 Summary"
priority: "P0"
estimated_time: "15min"
dependencies: ["1.x"]
status: pending
noTest: true
mainSession: false
---

# 1.summary: Phase 1 Summary

## Description

Generate a structured summary of all completed tasks in this phase. This summary is read by subsequent phase tasks to maintain cross-phase consistency.

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/milestone-map/tasks/records/` whose filename starts with `1.` and does NOT contain `.summary`.

### Step 2: Extract structured data into the summary field

```
## Tasks Completed
- 1.1: {{one-line summary}}
- 1.2: {{one-line summary}}

## Key Decisions
- {{each keyDecision from all records}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|

## Conventions Established
- {{each convention or pattern}}

## Deviations from Design
- {{each deviation, or "None"}}
```

### Step 3: Populate record.json

## Reference Files

- Phase task records: `docs/features/milestone-map/tasks/records/1.*.md`
- Design reference: `docs/features/milestone-map/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template
- [ ] Types & Interfaces Changed table is populated
- [ ] Record created via `/record-task`
