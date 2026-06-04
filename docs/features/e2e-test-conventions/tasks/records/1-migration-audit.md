---
status: "completed"
started: "2026-06-03 19:59"
completed: "2026-06-03 20:06"
time_spent: "~7m"
---

# Task Record: 1 迁移前审计：分类 tests/e2e/ 文件到 surface

## Summary
Created audit-manifest.md classifying all 52 spec files in tests/e2e/ into surface categories (22 api, 20 web, 7 infra, 3 shared non-spec). Marked 3 request.newContext files, 4 .serial files, 13 waitForTimeout files (166 occurrences), and classified all 20 no-suffix files by rule application. Cross-validated against proposal's Infra file migration mapping (12/12 MATCH) and target directory structure.

## Changes

### Files Created
- docs/features/e2e-test-conventions/tasks/audit-manifest.md

### Files Modified
无

### Key Decisions
无

## Document Metrics
52 spec files classified, 12 infra mappings verified, 20 no-suffix files classified by rule, 3 request.newContext flagged, 4 .serial flagged, 166 waitForTimeout occurrences across 13 files

## Referenced Documents
- docs/proposals/e2e-test-conventions/proposal.md

## Review Status
final

## Acceptance Criteria
- [x] Produce audit-manifest.md listing all .spec.ts files with surface classification (api/web/infra/shared)
- [x] Mark request.newContext files, .serial tests, and no-suffix files needing human judgment
- [x] Cross-validate against proposal target directory structure and Infra file migration mapping with no omissions

## Notes
Proposal's target tree shows representative examples, not exhaustive lists. Extra files in web/ and api/ are valid per proposal's own classification rules. roles/rbac-cli.spec.ts has -cli suffix but no runCli usage — classified as API (RBAC data migration test).
