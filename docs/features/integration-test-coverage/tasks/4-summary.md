---
id: "4.summary"
title: "Phase 4 Summary"
priority: "P0"
estimated_time: "15min"
dependencies: ["4.x"]
status: pending
---

# 4.summary: Phase 4 Summary

## Description

Generate a structured summary of all completed tasks in Phase 4 (Views & Reports Tests).

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/integration-test-coverage/tasks/records/` whose filename starts with `4.` and does NOT contain `.summary`.

### Step 2: Extract structured data into the summary field

```
## Tasks Completed
- 4.1: {{one-line summary}}

## Key Decisions
- {{each keyDecision from all records, prefixed with task ID}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|

## Conventions Established
- {{each convention or pattern, prefixed with task ID}}

## Deviations from Design
- {{each deviation from tech-design.md, or "None"}}
```

### Step 3: Populate remaining record.json fields

```json
{
  "taskId": "4.summary",
  "status": "completed",
  "summary": "<filled from Step 2 template above>",
  "filesCreated": [],
  "filesModified": [],
  "keyDecisions": [],
  "testsPassed": 0,
  "testsFailed": 0,
  "coverage": -1.0,
  "acceptanceCriteria": [
    {"criterion": "All phase task records read and analyzed", "met": true},
    {"criterion": "Summary follows the exact template with all 5 sections", "met": true},
    {"criterion": "Types & Interfaces table lists every changed type", "met": true}
  ]
}
```

## Reference Files

- All phase task records: `docs/features/integration-test-coverage/tasks/records/4.*.md`
- Design reference: `docs/features/integration-test-coverage/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template above
- [ ] Record created via `task record` with `coverage: -1.0`

## Implementation Notes

This is a documentation-only task. No code should be written.
