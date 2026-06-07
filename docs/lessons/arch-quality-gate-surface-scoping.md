---
created: "2026-06-07"
tags: [architecture, testing]
---

# Quality gate must scope verification to task's surface

## Problem

Task 2.3 (MainItem repository milestone methods) is a pure backend task (`surface-key: backend`, `surface-type: api`). After execution, the quality gate ran and **blocked** the task — not because of a backend issue, but because frontend lint failed to install `@stylistic/eslint-plugin-ts` due to npm network timeout. A fix task (fix-1) was spawned to resolve this, but the fix itself was also a backend-only task (`report_service_test.go`) that got stuck on the same frontend lint gate.

**Symptom:** Backend task stuck in `blocked` status due to an unrelated frontend infrastructure issue.

## Root Cause

Causal chain (3 levels deep):

1. **Surface level:** Quality gate blocked task 2.3, reporting "frontend lint cannot run" as the reason.
2. **Process level:** The quality gate runs verification for **all configured surfaces** (`backend` + `frontend`) regardless of which surface the task belongs to. It does not filter by the task's `surface-key` field.
3. **Design level:** The gate was designed as a full-stack safety net — verify everything that could break. This is correct for integration/gate tasks but wrong for scoped feature tasks where the agent only touched one surface's code.

## Solution

The quality gate should respect the task's `surface-key` and only run verification for the relevant surface. For tasks without a surface (doc, gate, summary types), running all surfaces is acceptable since no code was changed.

**Immediate workaround:** Manually unblock the task with `forge task transition 2.3 pending --reason "backend-only task blocked by unrelated frontend issue"`.

## Reusable Pattern

When a scoped task (with `surface-key` set) gets blocked by a quality gate failure on a **different surface**, the block is likely spurious. Check:
1. Does the task's `surface-key` match the surface that failed?
2. Did the agent modify files in the failing surface's codebase?
3. If both answers are "no", the gate over-scoped — unblock and consider fixing the gate to filter by surface.

**Rule:** Quality gate verification should be scoped to the task's `surface-key`. Full-stack verification is appropriate only for tasks that touch multiple surfaces (e.g., integration tests, gate tasks).

## Related Files

- `docs/features/milestone-map/tasks/index.json` — task 2.3 and fix-1 entries
- `.forge/config.yaml` — `surfaces` configuration (backend + frontend)

## References

- Related: [[gotcha-hook-stop-e2e-blocking]] — stop hook fires e2e check on unrelated work
- Related: [[arch-task-failure-recovery-loop]] — task executor encounters test failures
