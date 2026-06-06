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

Extract and consolidate business rules and tech specs from the system-ux-optimization feature.

## Feature Context


## Discovery Strategy
1. Scan docs/features/system-ux-optimization/ for all feature documents (PRD, design, task records)
2. Scan docs/proposals/system-ux-optimization/ for proposal
3. Extract rules and specs from discovered documents
4. Compare against existing specs in docs/business-rules/ and docs/conventions/

Run in non-interactive mode: auto-integrate all CROSS items. Commit with [auto-specs] tag.

## Acceptance Criteria
- [ ] 1. Task completes without error
