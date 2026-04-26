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

Generate a structured summary of all completed tasks in Phase 3 (Pattern Unification). This summary is read by subsequent phase tasks to maintain cross-phase consistency.

## Instructions

### Step 1: Read all phase task records

Read each record file from `docs/features/schema-alignment-cleanup/tasks/records/` whose filename starts with `3.` and does NOT contain `summary` (e.g., `3.1-*.md`, `3.2-*.md`, `3.3-*.md`, `3.4-*.md`).

### Step 2: Extract structured data into the summary field

<HARD-RULE>
The `summary` field in `record.json` MUST follow this exact template:
</HARD-RULE>

```
## Tasks Completed
- 3.1: {{one-line summary}}
- 3.2: {{one-line summary}}
- 3.3: {{one-line summary}}
- 3.4: {{one-line summary}}

## Key Decisions
- {{each keyDecision from all records, prefixed with task ID}}

## Types & Interfaces Changed
| Name | Change | Affects |
|------|--------|---------|
| {{name}} | {{added/modified/removed}} | {{which subsequent tasks care}} |

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
  "summary": "<filled from Step 2>",
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

- All phase task records: `docs/features/schema-alignment-cleanup/tasks/records/3.*.md`
- Design reference: `docs/features/schema-alignment-cleanup/design/tech-design.md`

## Acceptance Criteria

- [ ] All phase task records have been read
- [ ] Summary follows the exact 5-section template above
- [ ] Types & Interfaces Changed table is populated (or "None" if no changes)
- [ ] Record created via `/record-task` with `coverage: -1.0`

## Implementation Notes

This is a documentation-only task. No code should be written.
- Set `coverage: -1.0` in the record to indicate no tests expected
- The summary MUST be structured — subsequent phase tasks depend on parsing it
