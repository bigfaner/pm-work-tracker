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

Generate a structured summary of all completed tasks in Phase 3 (Core Business Flow Tests: F2, F3, F4).

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/integration-test-coverage/tasks/records/` whose filename starts with `3.` and does NOT contain `.summary`.

### Step 2: Extract structured data into the summary field

```
## Tasks Completed
- 3.1: {{one-line summary}}
- 3.2: {{one-line summary}}
- 3.3: {{one-line summary}}

## Key Decisions
- {{each keyDecision from all records, prefixed with task ID}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|
| {{type/interface name}} | {{added/modified/removed: brief description}} | {{which subsequent tasks care}} |

## Conventions Established
- {{each convention or pattern, prefixed with task ID}}

## Deviations from Design
- {{each deviation from tech-design.md, or "None"}}
```

### Step 3: Populate remaining record.json fields

```json
{
  "taskId": "3.summary",
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

- All phase task records: `docs/features/integration-test-coverage/tasks/records/3.*.md`
- Design reference: `docs/features/integration-test-coverage/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template above
- [ ] Types & Interfaces Changed table is populated
- [ ] Record created via `task record` with `coverage: -1.0`

## Implementation Notes

This is a documentation-only task. No code should be written.
