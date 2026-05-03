---
id: "2.gate"
title: "Phase 2 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
---

# 2.gate: Phase 2 Exit Gate

## Description

Exit verification gate for Phase 2. Confirms that all 11 features have been processed — either graduated or documented as blocked — before Phase 3 finalizes the regression suite.

## Verification Checklist

1. [ ] All 11 features have been attempted (none silently skipped)
2. [ ] Each graduated feature has a marker in `tests/e2e/.graduated/<slug>`
3. [ ] Each blocked feature has an entry in `tests/e2e/KNOWN_FAILURES.md` with slug, test ID, reason, and owner
4. [ ] No graduated spec file imports from `testing/scripts/` paths (run `validateSpec()` spot-check)
5. [ ] Graduated specs are organized under `tests/e2e/api/`, `tests/e2e/ui/`, `tests/e2e/cli/` as appropriate
6. [ ] No deviations from design spec (or deviations documented as decisions)

## Reference Files

- `design/tech-design.md` — Model 3 Feature Inventory
- `tests/e2e/.graduated/` — graduation markers
- `tests/e2e/KNOWN_FAILURES.md`
- `tasks/records/2-summary.md`

## Acceptance Criteria

- [ ] All 11 features accounted for (graduated or blocked with documented reason)
- [ ] Graduation marker count + blocked count = 11
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `task record` with test evidence

## Implementation Notes

Verification-only task. No new feature code should be written.
Count graduation markers: `ls tests/e2e/.graduated/ | wc -l` should show ≥11 (includes pre-existing markers).
