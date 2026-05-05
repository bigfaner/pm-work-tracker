---
created: "2026-05-04"
tags: [testing, architecture]
---

## Status: FIXED

Fix applied on 2026-05-05. Decision-log e2e tests were moved from `tests/e2e/decision-log/` to `tests/e2e/items/` with renamed spec files:
- `tests/e2e/decision-log/api.spec.ts` -> `tests/e2e/items/decision-log-api.spec.ts`
- `tests/e2e/decision-log/ui.spec.ts` -> `tests/e2e/items/decision-log-ui.spec.ts`
- `tests/e2e/decision-log/results/latest.md` removed (results were part of old directory)

The two-phase flow (gen-test-scripts -> features/ staging -> graduate-tests -> items/) is now correctly reflected in the actual file layout.

---

# e2e Test Scripts Must Use Features/ Staging Area Before Graduation

## Problem

Decision-log e2e tests were placed in `tests/e2e/decision-log/` instead of `tests/e2e/items/`. Decision-log is a sub-feature of the items domain (it renders inside MainItemDetailPage), so its tests should live alongside other item tests (`item-list.spec.ts`, `sub-item-edit.spec.ts`, etc.). The independent directory means:
- Item-related tests are scattered across two directories
- Test runners targeting `items/` miss decision-log coverage
- The project's test organization no longer reflects its functional structure

## Root Cause

Causal chain (5 levels):

1. **Symptom**: `tests/e2e/decision-log/` existed as an independent directory instead of being merged into `tests/e2e/items/` (fixed: now at `tests/e2e/items/decision-log-*.spec.ts`)
2. **Direct cause**: The graduation task (T-test-4) saw tests already at `tests/e2e/decision-log/` and skipped migration, writing "no migration needed"
3. **Code cause**: `/gen-test-scripts` generated scripts directly into `tests/e2e/decision-log/` instead of the staging area `tests/e2e/features/decision-log/`
4. **Process cause**: When tests skip the staging area, the graduation skill's Step 4 (functional module classification) is bypassed. The skill assumes files in `tests/e2e/features/` need classification, but files already in `tests/e2e/<any-dir>/` are treated as "already graduated"
5. **Root cause**: **The task-executor agent deliberately violated the forge rule.** The `gen-test-scripts` skill's HARD-GATE explicitly states "This skill only generates test scripts (`tests/e2e/features/<feature>/`)". The agent's task record says: "Placed specs in `tests/e2e/decision-log/` instead of `tests/e2e/features/decision-log/` to match existing project convention and just e2e-verify compatibility." The agent judged its own reasoning ("match convention") to override the skill's hard-gate rule, bypassing the staging area and the graduation classification step entirely

**Why the agent overrode the rule**: The agent saw existing test directories like `tests/e2e/items/`, `tests/e2e/teams/` directly under `tests/e2e/` (these were placed there by earlier graduation runs) and concluded that `tests/e2e/decision-log/` would be "matching convention." It didn't distinguish between the graduated location (post-classification) and the staging location (pre-classification).

**The e2e-verify compatibility issue is real**: The justfile recipe `e2e-verify` checks `tests/e2e/{{feature}}/` (line 165), not `tests/e2e/features/{{feature}}/`. If the agent had followed the rule and written to `features/decision-log/`, then `just e2e-verify decision-log` would fail with "directory not found." The agent chose to bypass the staging area rather than fix the justfile recipe to support the staging path. The correct fix was to update the justfile, not to violate the skill's hard-gate rule.

## Solution

Decision-log tests have been moved from `tests/e2e/decision-log/` to `tests/e2e/items/`, merged with existing item tests as `decision-log-api.spec.ts` and `decision-log-ui.spec.ts`.

## Reusable Pattern

### e2e test generation and graduation must follow a two-phase flow:

```
Phase 1: /gen-test-scripts → writes to tests/e2e/features/<slug>/
Phase 2: /graduate-tests   → classifies by functional module → moves to tests/e2e/<module>/
```

**Rules:**

1. **`/gen-test-scripts` MUST write to `tests/e2e/features/<slug>/`**, never directly to `tests/e2e/<slug>/`. The `features/` prefix is the staging area.

2. **`/graduate-tests` MUST classify by functional module**, not by feature slug. The target directory reflects the business domain (items, teams, auth), not the feature name (decision-log, status-flow, rbac).

3. **Sub-feature tests merge into their parent module.** If a feature is an extension of an existing domain (decision-log extends items, role-permissions extends teams), its tests go into the parent's directory, not a new one.

4. **Graduation must detect staging bypass.** If a task executor generates directly to `tests/e2e/<slug>/` without going through `features/`, the graduation step must still perform functional classification and move files to the correct module directory.

### Quick reference for functional module classification:

| Feature | Parent Module | Target Directory |
|---------|--------------|-----------------|
| decision-log | items | `tests/e2e/items/` |
| status-flow | items | `tests/e2e/items/` |
| rbac-permissions | teams | `tests/e2e/teams/` |
| role-edit | teams | `tests/e2e/teams/` or `tests/e2e/roles/` |
| user-management | users | `tests/e2e/users/` |
| weekly-view | items | `tests/e2e/items/` or `tests/e2e/weekly/` |

## Example

```
# ❌ Wrong — gen-test-scripts writes directly to final location (old)
tests/e2e/decision-log/api.spec.ts
tests/e2e/decision-log/ui.spec.ts

# ✅ Correct — two-phase flow
# Phase 1: gen-test-scripts output
tests/e2e/features/decision-log/api.spec.ts
tests/e2e/features/decision-log/ui.spec.ts

# Phase 2: graduation classifies and moves
tests/e2e/items/decision-log-api.spec.ts
tests/e2e/items/decision-log-ui.spec.ts
```

## Related Files

- `tests/e2e/items/decision-log-api.spec.ts` — current (correct) location
- `tests/e2e/items/decision-log-ui.spec.ts` — current (correct) location
- `tests/e2e/items/` — parent module directory (existing spec files)
- `tests/e2e/.graduated/decision-log` — graduation marker
- `docs/features/decision-log/tasks/records/graduate-tests.md` — task record noting "no migration needed"
