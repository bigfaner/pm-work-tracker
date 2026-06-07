---
id: "T-review-doc"
title: "Review Documentation Quality"
priority: "P1"
estimated_time: "30min"
dependencies: ["1.2", "1.1"]
type: "doc.review"
surface-key: ""
surface-type: ""
---

Review documentation quality for the milestone-map feature (breakdown mode).

## Acceptance Criteria

- [ ] All doc task acceptance criteria verified against actual deliverables
- [ ] MySQL and SQLite DDL for new tables + ALTER are syntactically correct
- [ ] Indexes and unique constraints match design specs
- [ ] Business rules extracted to `docs/business-rules/`
- [ ] Tech specs extracted to `docs/conventions/`
- [ ] No drift between feature docs and consolidated specs

Extract and consolidate business rules and tech specs from the milestone-map feature.


## Discovery Strategy

Scan ONLY the following allowlist of directories for target documents:
- docs/features/milestone-map/ (prd/, design/, testing/, and any subdirectories)
- docs/proposals/milestone-map/

EXCLUDE the following from scanning — do NOT read or process these:
- tasks/ directory (task definitions are not deliverables)
- tasks/records/ directory (execution records are not deliverables)
- manifest.md (build artifact)
- index.json (build artifact)

Only .md files under the allowlist directories are target deliverables.
