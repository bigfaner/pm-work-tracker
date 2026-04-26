# Phase Gate Tasks Must Be Modeled Explicitly

## Problem

When breaking down tasks from a PRD that defines multi-phase execution with quality gates between phases, the task decomposition produced only business implementation tasks (12) + standard test tasks (2). The round-level verification gates defined in the PRD flow diagram (e.g., `R1Test{Round 1 测试全部通过?}`) were not modeled as explicit tasks.

This means the `task claim` dispatcher has no checkpoint between phases — it could claim a Round 3 task before Round 2's dead code removal is fully verified, risking build failures on stale assumptions.

## Root Cause

The breakdown-tasks skill's Element Mapping table only covers design-level artifacts:

| Design Element       | Source         | Task Type                |
| -------------------- | -------------- | ------------------------ |
| Interface definition | tech-design.md | Interface task           |
| Data model           | tech-design.md | Model task               |
| Backend component    | tech-design.md | Implementation (Backend) |

PRD flow diagram nodes (especially diamond decision nodes representing quality gates) are not in this mapping. The mapper treated the PRD flow as descriptive context rather than prescriptive requirements.

**Causal chain**: No mapping rule for flow gates → mapper ignored diamond nodes → only business tasks produced → `task claim` has no phase boundaries.

## Solution

After generating all business tasks per phase, add a gate task for each phase boundary. A gate task:

- Depends on ALL business tasks in the preceding phase
- Is a dependency for ALL tasks in the next phase
- Runs verification commands (e.g., `go build ./...`, `go test ./...`, `npx vitest run`)
- Has breaking: true in index.json to trigger full test suite

Example task structure for a 4-round cleanup:

```
1.x  Round 1 tasks (P0 bug fixes)
1.G  Round 1 Gate: verify bug fixes + all tests pass
2.x  Round 2 tasks (dead code) — all depend on 1.G
2.G  Round 2 Gate: verify build + tests pass
3.x  Round 3 tasks (pattern unification) — all depend on 2.G
3.G  Round 3 Gate: verify build + tests pass
4.x  Round 4 tasks (architecture) — all depend on 3.G
4.G  Round 4 Gate: verify full test suite
T-test-1, T-test-2 — depend on 4.G
```

## Key Takeaway

**When the PRD flow diagram contains diamond decision nodes (quality gates, phase transitions), each gate must become an explicit verification task in the task breakdown.** These gates enforce phase boundaries that the `task claim` dispatcher respects.

Check: after generating business tasks, compare the task dependency graph against the PRD flow diagram. Every diamond node should have a corresponding task.
