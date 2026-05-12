---
id: "3.summary"
title: "Phase 3 Summary"
priority: "P0"
estimated_time: "15min"
dependencies: ["3.x"]
status: pending
noTest: true
mainSession: false
---

# 3.summary: Phase 3 Summary

## Description

Generate a structured summary of all completed tasks in Phase 3.

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/milestone-map/tasks/records/` whose filename starts with `3.` and does NOT contain `.summary`.

### Step 2: Extract structured data

```
## Tasks Completed
- 3.1: {{one-line summary}}
- 3.2: {{one-line summary}}
- 3.3: {{one-line summary}}
- 3.4: {{one-line summary}}
- 3.5: {{one-line summary}}
- 3.6: {{one-line summary}}

## Key Decisions
- {{each keyDecision}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|

## Conventions Established
- {{each convention or pattern}}

## Deviations from Design
- {{each deviation, or "None"}}
```

## Reference Files

- Phase 3 task records: `docs/features/milestone-map/tasks/records/3.*.md`
- Design reference: `docs/features/milestone-map/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template
- [ ] Record created via `/record-task`
