---
id: "T-specs-consolidate"
title: "Consolidate Specs"
priority: "P2"
estimated_time: "20min"
dependencies: ["T-test-run-frontend"]
type: "doc.consolidate"
surface-key: ""
surface-type: ""
---

## Acceptance Criteria
- [ ] New business rules extracted to docs/business-rules/
- [ ] New tech specs extracted to docs/conventions/
- [ ] No drift between feature docs and consolidated specs

Extract and consolidate business rules and tech specs from the milestone-map feature.

## Feature Context


## Discovery Strategy
1. Scan docs/features/milestone-map/ for all feature documents (PRD, design, task records)
2. Scan docs/proposals/milestone-map/ for proposal
3. Extract rules and specs from discovered documents
4. Compare against existing specs in docs/business-rules/ and docs/conventions/

Run in non-interactive mode: auto-integrate all CROSS items. Commit with [auto-specs] tag.
