---
id: "T-validate-code"
title: "Validate Code Quality"
priority: "P2"
estimated_time: "15min"
dependencies: ["T-test-run-frontend"]
type: "validation.code"
surface-key: ""
surface-type: ""
---

Validate code quality for the system-ux-optimization feature.

## Validation Criteria
- [ ] All acceptance criteria met

## Additional Checks
- Check docs/conventions/ for project-specific quality standards (read each file's `domains` frontmatter to determine relevance)
- Run the quality gate: just compile → just fmt → just lint → just test

## Acceptance Criteria
- [ ] 1. Task completes without error
