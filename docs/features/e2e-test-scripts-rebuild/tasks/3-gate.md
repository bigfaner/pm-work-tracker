---
id: "3.gate"
title: "Phase 3 Exit Gate"
priority: "P0"
estimated_time: "30min"
dependencies: ["3.summary"]
status: pending
breaking: true
---

# 3.gate: Phase 3 Exit Gate

## Description

Exit verification gate for Phase 3. Confirms the regression suite is complete and `npm test` runs cleanly before the standard test tasks begin.

## Verification Checklist

1. [ ] `tests/e2e/package.json` `test:api` includes all graduated API spec paths
2. [ ] `tests/e2e/package.json` `test:ui` exists and includes all graduated UI spec paths (if any)
3. [ ] `tests/e2e/package.json` `test:cli` includes all graduated CLI spec paths
4. [ ] `npm test` in `tests/e2e/` exits with code 0 (or all failures are in KNOWN_FAILURES.md)
5. [ ] No spec path in package.json points to a non-existent file
6. [ ] No deviations from design spec (or deviations documented as decisions)

## Reference Files

- `design/tech-design.md` — Interface 3: package.json Updater
- `tests/e2e/package.json`
- `tests/e2e/KNOWN_FAILURES.md`
- `tasks/records/3-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations from design are documented as decisions in the record
- [ ] Record created via `task record` with test evidence

## Implementation Notes

Verification-only task. No new feature code should be written.
Run `npm test` from `tests/e2e/` and capture exit code. If exit code is non-zero, check whether all failures are documented in KNOWN_FAILURES.md before marking as passed.
