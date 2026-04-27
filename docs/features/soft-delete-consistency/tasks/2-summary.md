---
id: "2.summary"
title: "Phase 2 Summary"
priority: "P0"
estimated_time: "15min"
dependencies: ["2.x"]
status: pending
---

# 2.summary: Phase 2 Summary

## Description

Generate a structured summary of all completed tasks in Phase 2.

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/soft-delete-consistency/tasks/records/` whose filename starts with `2.` and does NOT contain `.summary`.

### Step 2: Extract structured data into the summary field

```
## Tasks Completed
- 2.1: {{summary}}
- 2.2: {{summary}}
- 2.3: {{summary}}
- 2.4: {{summary}}
- 2.5: {{summary}}

## Key Decisions
- {{decisions}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|

## Conventions Established
- {{conventions}}

## Deviations from Design
- {{deviations}}
```

### Step 3: Populate record.json

Set `coverage: -1.0`.

## Reference Files

- Phase 2 task records: `records/2.*.md`
- Design: `docs/features/soft-delete-consistency/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records read
- [ ] Summary follows exact template
- [ ] Record created via `/record-task` with `coverage: -1.0`
