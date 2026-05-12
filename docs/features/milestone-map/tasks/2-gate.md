---
id: "2.gate"
title: "Phase 2 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
noTest: false
mainSession: false
---

# 2.gate: Phase 2 Exit Gate

## Description

Exit verification gate for Phase 2. Confirms that all backend APIs are functional, MainItem integration works, and computed fields return correct values.

## Verification Checklist

1. [ ] All MilestoneMap endpoints return correct responses (Create, List, Get, Update, Delete, ChangeStatus, AvailableTransitions)
2. [ ] All Milestone endpoints return correct responses
3. [ ] MainItem create/update accepts milestoneKey
4. [ ] MainItem list returns milestoneName enrichment
5. [ ] Soft-deleted milestones show "--" in MainItem VO
6. [ ] Status transitions validated (illegal transitions return 422)
7. [ ] Delete cascade: milestone delete → MI milestone_key nullified; map delete → all milestones + MIs cleaned
8. [ ] Computed fields (completion, overallProgress) return correct values
9. [ ] `just compile` passes
10. [ ] `just test` passes
11. [ ] No deviations from design spec

## Reference Files

- `design/api-handbook.md` — Endpoint contracts
- `design/tech-design.md` — Cross-Layer Data Map, Computed Fields
- Phase 2 task records: `records/2.*.md`
- Phase 2 summary: `records/2-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence
